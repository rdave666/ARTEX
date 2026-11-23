# Red-team skill import and agent assignments

Imported 13 standalone knowledge packs from Claude Red 0.3.1; original source
files and reference documents were copied byte-for-byte, without overwriting
existing packs. Live root: `/opt/artex/skills`; imported checkout copies: `skills/`.

## Agent roles

- **Auto** is the platform operations/chat agent, not automatic skill assignment.
- **Worker** executes a planner intent; technical procedures belong here.
- **Penetration Testing** runs a solo assessment and benefits from the full set.
- **Retester** checks an existing finding; broad reconnaissance was not added.
- **Planner / Goal Decomposer** get planning and prioritization guidance.
- **Main** gets coordination/review/report guidance.
- **Report writer** gets reporting and critical-thinking guidance.

All previously enabled bindings were retained. New desired bindings were added
transactionally, without changing MCP/tool permissions, prompts, model settings,
networking, or tasks. No service restart, scan, tool installation, or LLM request
was required. Files and visibility are consulted during agent tool assembly;
a run already in progress may keep its previous tool set until its next run.

## Verified complete assignment matrix

| Skill | Enabled agents |
|---|---|
| `api-recon` | `worker`, `auto`, `pentest` |
| `api-security` | `worker`, `auto`, `pentest`, `retester` |
| `bbot-recon` | `worker`, `auto`, `pentest` |
| `codeaudit` | `worker`, `auto`, `pentest`, `retester` |
| `critical-thinking` | `goals`, `planner`, `mainagent`, `worker`, `auto`, `pentest`, `reporter`, `retester` |
| `edr-bypass-re` | `worker`, `auto`, `pentest`, `retester` |
| `offensive-active-directory` | `worker`, `pentest`, `retester` |
| `offensive-advanced-redteam` | `goals`, `planner`, `mainagent`, `auto`, `pentest` |
| `offensive-ai-security` | `worker`, `pentest`, `retester` |
| `offensive-cicd-pipeline` | `worker`, `pentest`, `retester` |
| `offensive-cicd-secrets` | `worker`, `pentest`, `retester` |
| `offensive-cloud` | `worker`, `pentest`, `retester` |
| `offensive-k8s-attacks` | `worker`, `pentest`, `retester` |
| `offensive-lateral-movement` | `worker`, `pentest`, `retester` |
| `offensive-linux-privesc` | `worker`, `pentest`, `retester` |
| `offensive-netexec` | `worker`, `pentest`, `retester` |
| `offensive-reporting` | `planner`, `mainagent`, `worker`, `auto`, `pentest`, `reporter`, `retester` |
| `offensive-supply-chain` | `worker`, `pentest`, `retester` |
| `offensive-windows-privesc` | `worker`, `pentest`, `retester` |
| `patch-diff` | `worker`, `auto`, `pentest`, `retester` |
| `php2shell` | `worker`, `auto`, `pentest`, `retester` |
| `playwright-cli` | `worker`, `auto`, `pentest`, `retester` |
| `rea` | `worker`, `auto`, `pentest`, `retester` |
| `scopesentry` | `worker`, `auto`, `pentest` |
| `sqlmap-database-pentesting` | `worker`, `auto`, `pentest`, `retester` |
| `vulnerability-scanner` | `goals`, `planner`, `worker`, `auto`, `pentest`, `retester` |

## Validation and limitations

- All 26 skills parsed with the same Norma library used by ARTEX.
- The Skill meta-tool successfully loaded every skill without executing its
  procedures; all imported reference files are readable by the service user.
- All 26 imported files match source SHA-256 hashes in both deployment locations.
- All 26 installed skills have enabled agent assignments; existing bindings kept.
- Per-agent descriptions are bounded by ARTEX's skill-listing budget; beyond it
  the loader lists skill names only, rather than dropping skills.
- `rea` needs its MCP server, which is not configured. ScopeSentry exists but is
  disabled. Those integrations were not changed by importing or assigning skills.
- Knowledge availability does not establish CLI installation, external credentials,
  platform privileges, or successful target testing.
- Saved admin credentials no longer authenticate. They were neither reset nor
  altered: import used local filesystem access and the application's local
  database role. Authenticated HTTP UI discovery was not tested with a new login.

## Evidence and rollback

`deploy/evidence/redteam-import/` contains provenance/hashes, parser and invocation
results, before/after visibility, the exact additive SQL and rollback SQL.

To undo only this operation's visibility changes (not executed):

```sh
sudo -u artex psql -X -v ON_ERROR_STOP=1 -h /var/run/postgresql -p 5434 -d artex \
  < /home/vlad/ARTEX/deploy/evidence/redteam-import/rollback-bindings.sql
```

Rollback retains imported files. Review later user changes before rolling back.
Existing third-party skill content was not rewritten.
