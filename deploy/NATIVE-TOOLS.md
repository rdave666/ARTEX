# Native tool remediation

Completed 2026-11-16, without Docker or ARTEX restart. Verification used the
running ARTEX mount namespace, user `artex`, HOME `/var/lib/artex`, and PATH
`/usr/local/bin:/usr/bin:/bin`. Browser tests additionally applied NoNewPrivileges
and empty capabilities. All evidence is in `evidence/tool-setup/`.

## Installed/exposed commands

| Command | Version | Service-accessible installation |
|---|---|---|
| `httpx` | ProjectDiscovery 1.9.0 | New `/usr/local/bin/httpx` symlink to existing `/usr/bin/httpx-toolkit` |
| `subfinder` | 2.14.0 | Root-owned copy under `/opt/artex-tools/subfinder-v2.14.0/`, new PATH symlink |
| `katana` | 1.8.0 | Root-owned `/opt/artex-tools/katana-v1.8.0/`, new PATH symlink |
| `dalfox` | 3.2.4 | Root-owned `/opt/artex-tools/dalfox-v3.2.4/`, new PATH symlink |
| `playwright-cli` | 0.1.22 | Root-owned `/opt/artex-playwright/`, new PATH symlink |

Version/help probes passed for all five. Katana/Dalfox release archives were
SHA-256 checked against their publishers' release checksum assets; source URLs
and binary hashes are recorded in `release-binaries.json`. These checks establish
download integrity against the same release channel, not independent publisher
authentication. Subfinder's copy matches the existing source binary's SHA-256.
Httpx/subfinder were reused, not upgraded to ARTEX's image pins. The unrelated
Python HTTPX command remains untouched at `/usr/bin/httpx`.

`/home/vlad` remains mode 0700. No symlink crosses into Vlad's private home.
Root-owned native binaries use 0755 permissions. Scanner installations occupy
122 MiB; Playwright CLI occupies 19 MiB. Download archives were temporary and
removed; no Go/Rust build toolchain or browser image was installed.

An initial Dalfox probe incorrectly used `dalfox version`; v3 treats bare words
as targets, so it tried to resolve `http://version/` and failed DNS resolution.
The correct `dalfox --version` and help probes passed. Initial output is preserved
in `scanner-smoke-initial.json`; corrected results are in `scanner-smoke.json`.
No external assessment target was intentionally scanned.

## Real browser verification

The existing configured MCP command `npx @playwright/mcp --headless` passed JSON-RPC
initialize, tools/list, localhost fixture navigation, snapshot marker verification,
evaluation, and browser_close. It selected existing system Chrome
`/opt/google/chrome/chrome` 156.0.8078.73. The installed CLI independently passed
open, snapshot, evaluation, and close with `--browser=chrome`.

No `--no-sandbox`, browser download, MCP configuration change, or weakened systemd
hardening was needed. Probes ran inside the service filesystem namespace and
privilege limits, but not inside its cgroup. Test processes/listeners were cleaned
up; normal small browser cache/profile files remain. Reproduction and exact
results: `evidence/tool-setup/browser-report.md`, `browser-mcp-hardened-smoke.json`,
and `browser-cli-hardened-smoke.json`.

## Selective BBOT preparation

Only `asyncpg==0.32.0` was added to `/opt/bbot`, for the passive `crt_db` module.
Prepared passive and lightweight active subsets passed no-target dry-run setup
with network isolation: 30/30 and 19/19 modules respectively, counting BBOT's
automatically loaded internal/output/dependency modules. No target scan was run
for these checks. Imported module/package-constraint checks also passed.

This does NOT prepare all optional BBOT modules or prove upstream API availability.
Heavy modules/toolchains remain deferred. The checkout and live `bbot-recon`
skill now document the prepared subset, native task path, `--no-deps`, and deferred
families instead of claiming image-wide dependency installation. Original vendor
sources remain unchanged. Full preparation details: `evidence/tool-setup/bbot-report.md`.

## Isolation and service checks

ARTEX, its firewall unit, and dedicated PostgreSQL cluster remain active.
`/api/health` returned `{"ok":true}`. Listeners remain Tailscale IPv4
`100.95.74.28:8787` and loopback `127.0.0.1:8788`, with no PostgreSQL TCP listener.
The existing six-case synthetic non-Tailscale ingress test passed for both ports
across Tailscale/public/WireGuard destination IPs, with dedicated nftables drop
counter increasing 37 to 49. Test namespace/veth were cleaned up.
See `ingress-recheck.txt`. No remote tailnet peer test was performed.

About 3.3 GiB free remained after installation. Do not install every optional
BBOT dependency or download redundant browsers on this filesystem.

## Rollback boundaries

Remove only newly added `/usr/local/bin/{httpx,subfinder,katana,dalfox,playwright-cli}`
symlinks after checking their targets. The newly created `/opt/artex-tools`
version directories and `/opt/artex-playwright` may then be removed. Do not remove
`/usr/bin/httpx`, `/usr/bin/httpx-toolkit`, the original private-home subfinder,
existing system Chrome, or ARTEX's existing npm cache. BBOT rollback/skill backups
are documented in its preparation report. No rollback was performed.
