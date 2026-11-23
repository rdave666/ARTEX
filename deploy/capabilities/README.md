# ARTEX curated Kali capability declarations

## Installed result

- 32 command declarations (`kind=shell`, `system=false`), enabled only after a local help/version probe succeeds as the ARTEX service user inside its mount namespace. Nonzero help exits for masscan/Hydra/ldapsearch are explicitly accepted; masscan/Hydra also require an expected output signature.
- A compact `bash` workflow declaration points every enabled agent to the `kali-capabilities` Skill. Concrete command declarations bind only worker, pentest, retester and auto. They augment existing Bash descriptions; these are not independent callable tools or permission grants.
- The guide provides category-to-skill delegation summaries and the discover → Skill → prerequisites → execute → verify sequence. Existing skills and their assignments are preserved.
- `/opt/artex/capabilities/manifest.json` is the curated source; `inventory.md` / `inventory.json` provide current results. Root-managed script and state cannot be changed by execution agents. ARTEX can read the public inventory; existing `/opt/artex` access restrictions remain unchanged.

## Refresh

```
sudo systemctl start artex-capabilities-refresh.service
systemctl status artex-capabilities-refresh.service
systemctl list-timers artex-capabilities-refresh.timer
```

APT post-invoke starts the refresh asynchronously; daily timer is the fallback. After direct dpkg, npm, Go or manual binary changes, run the command explicitly. No automatic package installation or browser downloads occur. If ARTEX is stopped, refresh defers without starting it; manually refresh after starting it again. A timer failure is visible in the journal; inspect stale inventory timestamps.

The script uses a process lock and publishes complete inventory files atomically. Builtins are never modified. Previously operator-edited managed declarations are preserved as conflicts, including manual disabling and binding edits. Skill visibility inserts never overwrite existing rows. A service stop/restart during probing can fail the refresh; the existing inventory is retained and a later run can retry.

## Verification (2026-11-20)

- 32/32 local help/version probes passed; no target traffic or active assessment was requested.
- 4 Python unit tests passed, including file modes under restrictive umask.
- All 27 skill files parsed and loaded through the same Norma Skill implementation used by ARTEX.
- Read-only Go overlay test invoked the checkout's real `wireTools`/ToolResolve against current DB under the service mount namespace as artex: planner/main guide present, nmap detail absent; worker/pentest/retester/auto guide and nmap detail present. Seeding attempts were rejected by PostgreSQL read-only mode as intended; no live test writes. Test source is `assembly-readonly-test.go.txt`.
- Before/after comparison: all 59 existing builtin tool configurations and 136 existing skill bindings unchanged.
- ARTEX PID unchanged (266494); no restart or new listening ports. UI remains 100.95.74.28:8787, proxy 127.0.0.1:8788; capability bounding set unchanged (cap_net_raw).
- Refresh service exit 0; timer enabled; APT hook parsed by apt-config.

Evidence: `../evidence/capabilities/` (before/after snapshots, inventory, Skill loader output, read-only assembly test log).

Limitations: the running binary was not rebuilt and exact equality to the checkout is unproven; the assembly test validates checkout logic plus live DB, not an actual LLM response. New assembly/session picks up changes; an already-running agent can retain prior descriptions. Help/version success does not prove credentials, module dependencies, browser launch, scan privileges, packet capture, templates, target reachability or finding validity. Existing integration-dependent skills remain conditional on those integrations.

## Reinstallation

Install root-owned manifest/script under `/opt/artex/capabilities`, guide under `/opt/artex/skills/kali-capabilities`, unit/timer under `/etc/systemd/system`, APT hook at `/etc/apt/apt.conf.d/99artex-capabilities`, then daemon-reload, enable timer and start refresh. Source files in this directory and `../../skills/kali-capabilities/SKILL.md` are the maintained copies. Keep directory ownership root and existing ARTEX parent permissions.

## Rollback

Disable/stop `artex-capabilities-refresh.timer` and stop its oneshot first; remove only this task's APT hook and two units, then daemon-reload. Compare managed-state.json to current tools rows and delete only unedited, non-system shell rows owned by this inventory marker (including `bash`), preserving any operator edits. Delete only the new `kali-capabilities` visibility rows and guide directory if unchanged. Do not restore the entire before snapshot: that could erase later user changes. Remove only this task's `/opt/artex/capabilities` files after preserving desired audit evidence. No ARTEX restart or network-policy rollback is needed.
