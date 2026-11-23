# ARTEX service tool audit

> Remediation completed after this audit: see [NATIVE-TOOLS.md](NATIVE-TOOLS.md)
> for corrected command paths, installed tools, real browser tests, and selective
> BBOT dependency readiness. The observations below are the pre-remediation snapshot.

Audited 2026-11-16 using the running service mount namespace, uid `artex`,
HOME `/var/lib/artex`, PATH `/usr/local/bin:/usr/bin:/bin`, not Vlad's interactive
shell. CLI version/help checks only: no targets scanned, no optional dependencies
installed, no networking/service/prompt changes. Local application caches/config
were created by some tools' first-run/version checks.

## What ARTEX actually requires

The embedded Go binary and PostgreSQL are the core UI/API runtime. Both are active.
An LLM provider is required for autonomous work. Python is used by custom Python
tools; source builds additionally need Go and Node/npm. Scanner tools below are
the upstream image baseline, NOT all prerequisites for starting the native UI.
No Docker was run or installed; its Dockerfile was inspected only as a manifest.

## Scanner/browser baseline

| Tool | Image pin | Live service result |
|---|---|---|
| `nuclei` | 3.11.1 | 3.11.1; version probe passed |
| `httpx` | 1.12.0 (ProjectDiscovery) | WRONG CLI at /usr/bin/httpx (Python HTTPX). ProjectDiscovery 1.9.0 works as /usr/bin/httpx-toolkit. |
| `subfinder` | 2.17.0 | Not available to service. Version 2.14.0 exists at /home/vlad/go/bin/subfinder, behind private home permissions. |
| `katana` | 1.8.0 | Missing from service PATH. |
| `ffuf` | 2.3.0 | 2.1.0-dev; version probe passed, differs from image pin. |
| `dalfox` | 3.2.4 | Missing from service PATH. |
| `sqlmap` | 1.9.9 | 1.10.8#stable; version probe passed, differs from image pin. |
| `bbot` | 3.0.2 | 3.0.2; isolated /opt/bbot; CLI, module listing and Python dependency check passed. Optional module dependencies NOT preinstalled. |
| `playwright-cli` | 0.1.22 | Missing from service PATH and unavailable from offline npm cache. |
| `playwright` | 1.64.1 | No standalone command. Runtime 1.66.0-alpha-2026-11-09 exists transitively inside cached MCP; standalone offline npm invocation unavailable. |

## Core helpers found

All baseline helpers were found on the service PATH: `bash`, `python3`, `node`,
`npm`, `npx`, `curl`, `wget`, `vim`, `git`, `jq`, `unzip`, `xz`, `dig`, `host`,
`nslookup`, `ping`, `nc`, `telnet`, `whois`, `nmap`, `rg`. Python is 3.14.8;
BBOT uses isolated system Python 3.13.15. Node is 25.8.1; npm/npx 11.11.0.
Nmap reports 7.99. These differ from some image versions, not necessarily defects.
Only tools with recorded version/help checks were execution-tested; presence
alone does not prove all subcommands or scan privileges work.

## Browser and MCP

- Browser MCP is enabled, command `npx @playwright/mcp --headless`.
- Cached `@playwright/mcp` 0.0.85 responds to a correct offline version probe.
- Its transitive Playwright runtime is present; `playwright-cli` is a separate
  missing package. An npm cache is not the same as a pinned global installation.
- System Chromium 150.0.7871.181 and Chrome 156.0.8078.73 respond to version checks.
- No Playwright-managed browser cache exists under the service HOME. This is not
  proof that browser MCP is broken: it may use system Chrome. An actual browser
  session/sandbox/library smoke test was not performed.
- ScopeSentry MCP exists but is disabled. REA MCP is not configured.
- No non-system custom tool declarations were registered at audit time.

## Optional tools for installed red-team knowledge packs

These are not universal ARTEX requirements. PATH availability does not imply
credentials, network access, module readiness or elevated privileges.

| Command | Service PATH result |
|---|---|
| `nxc` | /usr/bin/nxc |
| `netexec` | /usr/bin/netexec |
| `bloodhound-python` | /usr/bin/bloodhound-python |
| `impacket-GetUserSPNs` | /usr/bin/impacket-GetUserSPNs |
| `impacket-secretsdump` | /usr/bin/impacket-secretsdump |
| `certipy` | No certipy command; equivalent Kali certipy-ad exists and its help check passed. |
| `smbclient` | /usr/bin/smbclient |
| `rpcclient` | /usr/bin/rpcclient |
| `ldapsearch` | /usr/bin/ldapsearch |
| `enum4linux-ng` | /usr/bin/enum4linux-ng |
| `aws` | Not found on service PATH |
| `az` | Not found on service PATH |
| `gcloud` | Not found on service PATH |
| `kubectl` | Not found on service PATH |
| `helm` | Not found on service PATH |
| `prowler` | Not found on service PATH |
| `pacu` | Not found on service PATH |
| `roadrecon` | Not found on service PATH |
| `trivy` | Not found on service PATH |
| `gitleaks` | Not found on service PATH |
| `trufflehog` | Not found on service PATH |
| `semgrep` | Not found on service PATH |
| `ghidra` | No command and no launcher at /usr/share/ghidra/ghidraRun or /opt/ghidra/ghidraRun. |
| `r2` | /usr/bin/r2 |
| `gdb` | /usr/bin/gdb |
| `tshark` | /usr/bin/tshark |
| `tcpdump` | /usr/bin/tcpdump |
| `masscan` | /usr/bin/masscan |
| `naabu` | Not found on service PATH |
| `gobuster` | /usr/bin/gobuster |
| `feroxbuster` | /usr/bin/feroxbuster |
| `hydra` | /usr/bin/hydra |
| `nikto` | /usr/bin/nikto |
| `testssl` | Not found on service PATH |
| `wafw00f` | /usr/bin/wafw00f |

NetExec/nxc 1.5.1, BloodHound Python, an Impacket entrypoint, smbclient 4.24.6,
and certipy-ad passed local help/version checks. nxc --version exits 1 despite
printing the version; --help exits 0, so it is not classified as broken.
This optional sample is not an exhaustive dependency inventory of every command
in all 26 skill reference documents or all BBOT modules.

## BBOT installation completed

BBOT 3.0.2 is installed in `/opt/bbot`, exposed as `/usr/local/bin/bbot`, and
works inside the running ARTEX service environment without restarting ARTEX.
Python package consistency check passed. Exact packages: `evidence/bbot/`.

The shipped BBOT skill assumes image-style `--install-all-deps` setup. That was
NOT performed on this native installation. Several modules declare additional
pip/apt/Ansible/shell requirements; CLI/module listing does not verify them.
Its example `/app/data/tasks/` paths also refer to containers; native task data
lives under `/var/lib/artex/data/`. Existing skill content was not changed.

## Suggested order, not executed

1. Resolve ProjectDiscovery httpx command-name mismatch and expose subfinder
   without weakening private-home permissions.
2. Add katana, dalfox and Playwright CLI.
3. Verify the configured browser MCP can launch its selected browser.
4. Prepare only needed BBOT module dependencies. Available disk is about 3.5 GiB;
   do not blindly install every optional module/browser/toolchain.
5. Add cloud/Kubernetes/supply-chain tooling only for the tasks that require it.

## Sources and evidence

- `Dockerfile:57-64,88-112,142-181`: shipped binaries, helpers, BBOT and browser pins.
- `README.md:270-278,341-350`: native runtime versus build requirements.
- `server/customtool.go:214-243`: Python interpreter discovery.
- `db/db.go:266-270,314-329`: agent roles and seeded MCP defaults.
- `deploy/evidence/tool-audit/`: paths, versions, correct offline npm checks,
  cached packages and optional-tool smoke results.
- Tailscale isolation unchanged: UI `100.95.74.28:8787`, recording proxy
  `127.0.0.1:8788`, PostgreSQL socket-only.
