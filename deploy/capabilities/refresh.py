#!/usr/bin/python3
"""Root-owned inventory refresh. No scans, installation, or execution-policy changes."""
import argparse, datetime, fcntl, json, os, pathlib, subprocess, sys, tempfile

ROOT = pathlib.Path('/opt/artex/capabilities')
MARKER = '[ARTEX Kali inventory v1]'
TECH = ['worker', 'pentest', 'retester', 'auto']
WORKFLOW = ('Discover capabilities -> load the relevant Skill -> check prerequisites -> execute within task scope -> verify results. '
            'Before planning/delegating or using security commands, load Skill kali-capabilities. '
            'Read /opt/artex/capabilities/inventory.md for current verified commands, role guidance and limitations. '
            'A declaration is awareness, not permission or proof of credentials, module readiness or scan privileges.')
PROBE = '''import json,os,shutil,subprocess,sys
os.chdir('/var/lib/artex')
items=json.load(sys.stdin);out=[]
for item in items:
 r=dict(item);r['path']=shutil.which(item['name']);r['ready']=False
 if r['path']:
  try:
   p=subprocess.run([r['path']]+item['args'],capture_output=True,text=True,timeout=12,env=os.environ)
   r.update(exit_code=p.returncode,output=(p.stdout+p.stderr)[:3000])
   r['ready']=p.returncode in item.get('accept_codes',[0]) and bool(r['output'].strip()) and item.get('expect','') in r['output']
  except (OSError,subprocess.TimeoutExpired) as e:r['error']=type(e).__name__
 out.append(r)
print(json.dumps(out))
'''

def run(args, **kw):
    return subprocess.run(args, check=True, text=True, capture_output=True, **kw).stdout

def sql(s):
    return run(['/usr/sbin/runuser','-u','artex','--','/usr/bin/psql','-X','-q','-A','-t','-v','ON_ERROR_STOP=1','-h','/var/run/postgresql','-p','5434','-d','artex'],input=s)

def lit(s):
    # Dollar quoting prevents SQL/string escape ambiguity. Reject delimiter collisions.
    assert '$inventory$' not in s
    return '$inventory$'+s+'$inventory$'

def make_rows(results, agents):
    rows=[dict(key='bash',description=MARKER+' '+WORKFLOW,agents=agents,enabled=True)]
    for r in results:
        detail=f"{r['purpose']} Path {r['path'] or 'unavailable'}. Local help/version {'passed' if r['ready'] else 'not ready'}; functionality beyond that is not proven."
        rows.append(dict(key=r['name'],description=MARKER+' '+detail,agents=[a for a in TECH if a in agents],enabled=r['ready']))
    return rows

def write_file(path, content, mode=0o644):
    """Publish complete, explicitly readable snapshots independent of caller umask."""
    fd, name = tempfile.mkstemp(prefix='.'+path.name, dir=path.parent)
    try:
        with os.fdopen(fd, 'w') as out:
            out.write(content)
            os.fchmod(out.fileno(), mode)
        os.replace(name, path)
    finally:
        if os.path.exists(name): os.unlink(name)

def main():
    ap=argparse.ArgumentParser();ap.add_argument('--dry-run',action='store_true');args=ap.parse_args()
    if os.geteuid()!=0:sys.exit('Run via sudo or the root-owned refresh service.')
    lock=open('/run/lock/artex-capabilities.lock','w');fcntl.flock(lock,fcntl.LOCK_EX)
    pid=run(['/usr/bin/systemctl','show','artex','-pMainPID','--value']).strip()
    if pid=='0':print('ARTEX is stopped; refresh deferred.');return
    manifest=json.loads((ROOT/'manifest.json').read_text())
    results=json.loads(run(['/usr/bin/nsenter','-t',pid,'-m','--','/usr/sbin/runuser','-u','artex','--','/usr/bin/env','HOME=/var/lib/artex','PATH=/usr/local/bin:/usr/bin:/bin','/usr/bin/python3','-c',PROBE],input=json.dumps(manifest)))
    agents=json.loads(sql("SELECT json_agg(key ORDER BY key) FROM agents WHERE enabled;"))
    existing=json.loads(sql("SELECT coalesce(json_agg(json_build_object('key',key,'system',system,'kind',kind,'description',description,'agents',agents,'enabled',enabled)),'[]') FROM tools;"))
    by={r['key']:r for r in existing}
    previous=json.loads((ROOT/'managed-state.json').read_text()) if (ROOT/'managed-state.json').exists() else {}
    desired=make_rows(results,agents);changes=[];conflicts=[]
    for row in desired:
        old=by.get(row['key'])
        if old:
            snapshot={k:old[k] for k in ('key','description','agents','enabled')}
            if old['system'] or old['kind']!='shell' or previous.get(row['key'])!=snapshot:
                conflicts.append(row['key']);continue
        changes.append(row)
    report={'utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'commands':results,'conflicts_preserved':conflicts,'declarations':desired}
    if args.dry_run:print(json.dumps(report,indent=2));return
    stmts=['BEGIN;']
    for row in changes:
        stmts.append("INSERT INTO tools(key,system,description,schema,agents,enabled,kind,exec,deferred) VALUES (%s,false,%s,'{}',%s::jsonb,%s,'shell','{}',false) ON CONFLICT(key) DO UPDATE SET description=excluded.description,agents=excluded.agents,enabled=excluded.enabled WHERE NOT tools.system AND tools.kind='shell' AND tools.description=%s AND tools.agents=%s::jsonb AND tools.enabled=%s;" % (lit(row['key']),lit(row['description']),lit(json.dumps(row['agents'])),'true' if row['enabled'] else 'false',lit(by.get(row['key'],{}).get('description','')),lit(json.dumps(by.get(row['key'],{}).get('agents',[]))),'true' if by.get(row['key'],{}).get('enabled',False) else 'false'))
    # Bind only this new guide; never change existing skill assignments or overrides.
    stmts.append("INSERT INTO agent_skill_visibility(agent_id,skill_name,enabled) SELECT id,'kali-capabilities',true FROM agents WHERE enabled ON CONFLICT DO NOTHING;")
    stmts.append('COMMIT;');sql('\n'.join(stmts))
    for row in changes:previous[row['key']]=row
    write_file(ROOT/'managed-state.json', json.dumps(previous,indent=2)+'\n', 0o600)
    write_file(ROOT/'inventory.json', json.dumps(report,indent=2)+'\n')
    lines=['# ARTEX service capabilities', '', 'Refreshed UTC: '+report['utc'], '', WORKFLOW, '',
        '## Delegation', 'Planner/main: select a relevant skill and delegate technical execution to worker; do not turn orchestration into broad scanning.',
        'Worker/pentest/retester: load the relevant technical skill, inspect local --help and prerequisites, execute only task-scoped work, then verify.',
        'Goals/reporter/model_chat: consult this inventory for planning/reporting; tool awareness does not change role or execution permission.', '',
        '## Verified help/version inventory', '| Command | Category | Local probe | Service path |', '|---|---|---|---|']
    lines += [f"| `{r['name']}` | {r['category']} | {'passed' if r['ready'] else 'unavailable/failed'} | `{r['path'] or 'not found'}` |" for r in results]
    lines += ['', 'These are help/version probes only, not target scans or complete capability tests. Missing commands are not installed automatically.',
        'BBOT: only prepared module subsets; browser: use existing Chrome. Credentials, wordlists, capture/ptrace/raw socket privileges, external MCPs and live API health are separate prerequisites.',
        'Existing skills that require unavailable integrations are not made operational by their visibility. Read Skill guidance and report unavailable prerequisites honestly.',
        'Operator-edited declarations are preserved; refresh conflicts: '+(', '.join(conflicts) or 'none'),
        'Refresh: sudo systemctl start artex-capabilities-refresh.service (also after APT package transactions; daily fallback).',
        'An in-progress agent may retain its old tool descriptions until the next tool assembly/session.']
    write_file(ROOT/'inventory.md', '\n'.join(lines)+'\n')
    print(json.dumps({'verified':sum(r['ready'] for r in results),'curated':len(results),'declarations_synced':len(changes),'conflicts_preserved':conflicts,'guide_bound_agents':len(agents)}))

if __name__=='__main__':main()
