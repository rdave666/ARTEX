---
name: kali-capabilities
description: "Load before planning, delegating, or using local security tools. Discover verified ARTEX service commands, select relevant technical skills, check prerequisites, execute within task scope, and verify results. Covers web/API, recon, AD, binary analysis, packets and browser automation; includes dependency limitations."
---
# ARTEX Kali capability discovery

1. **Discover capabilities:** Read `/opt/artex/capabilities/inventory.md`. The adjacent `inventory.json` has timestamp, exact service paths, and help/version probe results. This curated inventory is not the whole Kali package list. If it is stale, check the specific command with `command -v` and local help; do not install packages automatically.
2. **Load relevant Skill:** Use the available Skill tool to load the relevant procedure before specialized work. Inventory/tool visibility alone does not load a methodology or provide an external integration. Consult a skill's referenced files as needed.
3. **Check prerequisites:** Verify current task intent/scope, required credentials, files/wordlists, service reachability and actual privileges. BBOT has only selected prepared module dependencies; use bbot-recon and --no-deps. Browser automation uses existing Chrome. REA and ScopeSentry require their configured/enabled integrations; do not assume they are operational.
4. **Execute:** Worker/pentest/retester use the relevant verified command through existing Bash/tool controls. Retesters reproduce a specific finding rather than expanding into broad recon. Planner/main use capability summaries to choose and delegate concrete intents; keep their existing role boundaries. Goals/reporter/model chat use the catalog for planning or evidence review, not as permission to take over assessment execution.
5. **Verify results:** Inspect exit status and output; distinguish missing prerequisites, unsupported privileges, inconclusive results and confirmed findings. Keep evidence/reproduction references and cleanup records. Do not infer success from tool presence or skill loading.

The inventory is root-managed; agents cannot refresh privileged state directly. Normal APT package transactions trigger refresh, with a daily fallback. An operator may run `sudo systemctl start artex-capabilities-refresh.service` after manual/Go/npm installs. No additional sudo, privileges, destinations or permissions are granted by this guide.

## Compact delegation map

| Work intent | Capability family | Load relevant skill(s) |
|---|---|---|
| Recon and attack-surface mapping | subfinder, BBOT, network discovery | api-recon, bbot-recon, vulnerability-scanner |
| Web/API checks and crawling | HTTP probes, crawlers, fuzzers, injection checks | api-security, sqlmap-database-pentesting, vulnerability-scanner |
| Windows/domain assessment | NetExec, Impacket, Certipy, LDAP/SMB | offensive-active-directory, offensive-netexec |
| Local source/binary analysis | source search, radare2, GDB | codeaudit, edr-bypass-re, patch-diff |
| Browser-driven reproduction | Playwright CLI/existing Chrome | playwright-cli |
| Packet evidence review | tshark/tcpdump offline analysis | vulnerability-scanner; verify live-capture privileges separately |

Execution roles receive the concrete declaration subset. Planner/main should load this guide, pick an intent, and delegate with inputs, limits, expected evidence and a stop condition. These examples are selection hints, not an instruction to run every listed tool.
