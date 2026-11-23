# ARTEX native deployment

## 2026-11-23 preservation update

Fork `rdave666/ARTEX` main and this native server were updated to commit
`cb1b8911c9cc7b78a7570bf753a1baaed24e3198` after a full Go test run,
static frontend build, isolated database-restore/migration test, and skill
parser check. Previous binary and skills remain at `/opt/artex/artex.pre-cb1b891`
and `/opt/artex/skills.pre-cb1b891`; the private database/runtime/Git recovery
snapshot is under `/home/vlad/.local/share/artex-recovery/2026-11-23`.
`MEMORY.md` records the operation and subsequent changes. Do not deploy from
the dirty `/home/vlad/ARTEX` checkout; use the reconciled fork/worktree and
verify hashes first.

The imported Skill Index and chat starter prompts are live. Runtime-only
prior skills are preserved by a **no-delete** skill sync. Post-update checks
showed the UI on the tailnet address, the proxy on loopback, and six synthetic
non-Tailscale ingress cases blocked by `inet artex`. A remote tailnet peer was
not available for an end-to-end access test.

The in-progress follow-on work adds a per-intent no-progress watcher. It
retains activity/transcript checkpoints and wakes the planner to diagnose an
unresponsive tool or changed target reachability; it must not blindly replay
an uncertain side effect. Verify and deploy that work separately before
describing it as live.

Installed 2026-11-16 from `https://github.com/BaaRai/ARTEX`, commit recorded in
`evidence/commit.txt`. Built from the checkout, no Docker or containers used.

## Access

- Connect your client to this tailnet, then open `http://100.95.74.28:8787/login/`.
- Username: `ARTEX`.
- Generated password: `/home/vlad/.config/artex/admin-credentials.txt` (0600).
- An admin password is already initialized; no setup token is needed.
- Add your chosen LLM provider in **System → LLM Configuration**. No provider,
  API credential, autonomous task, or target scan was configured or started.

## Runtime

- Service: `artex.service`, enabled at boot, automatic restart on failure.
- Ingress guard: `artex-network.service`, enabled and required before ARTEX.
- Binary and bundled skills: `/opt/artex/` (root-owned).
- Node/npm/npx runtime: `/opt/nodejs/artex-v25.8.1/`, root-owned copy of the
  existing Node v25.8.1 and npm/npx 11.11.0 installation, exposed through
  `/usr/local/bin/{node,npm,npx}`. This avoids private-home NVM links and works
  with `ProtectHome=true`; no service restart was needed. Verified as `artex`
  inside the running service's mount namespace, including offline npx execution.
- Database configuration: `/etc/artex/config.json`.
- Application data: `/var/lib/artex/data/`.
- JWT and encryption keys: `/var/lib/artex/keys/`. Back these up with the DB.
- Database: dedicated PostgreSQL cluster `18/artex`, DB/role `artex`, Unix
  socket `/var/run/postgresql`, logical socket port 5434, peer authentication.
  `listen_addresses=''`: no TCP database listener. Pre-existing clusters untouched.
- Main API/UI: **only** `100.95.74.28:8787` (IPv4 tailnet address).
- Recording proxy: **only** `127.0.0.1:8788`.
- Dedicated nftables table `inet artex` drops incoming TCP 8787/8788 on every
  interface except `tailscale0` and local `lo`. No output rules were added.
- No Tailscale Serve/Funnel or public reverse proxy was configured.

```sh
sudo systemctl status artex
sudo systemctl restart artex
sudo journalctl -u artex -f
sudo nft list table inet artex
sudo ss -lntp | grep -E '8787|8788|5434'
```

The bind address is explicit and fails closed if unavailable. A restart retries
every 10 seconds. The node's current Tailscale key expiry is 2026-11-26; renew
through your existing Tailscale administration before expiry if appropriate.
An address change requires updating `ExecStart` and re-running validation.
Never replace the address with `:8787` or `0.0.0.0:8787`.

## Validation

Evidence is under `deploy/evidence/` and `deploy/ingress-test.log`:

- Production static export and TypeScript build passed; 24 login page assets
  successfully fetched from the embedded binary.
- Config/entrypoint tests, focused auth/key tests, frontend polling tests passed.
- Admin initialization, login, authenticated health, restart persistence passed.
- Anonymous protected API returns 401; initialized-password reset returns 403.
- Public/WireGuard/loopback/IPv6 addresses have no UI socket.
- Synthetic non-Tailscale ingress is tested with a temporary network namespace,
  including packets addressed to the actual tailnet IP. Each case must timeout
  and increment this application's firewall drop counter. Resources are removed
  by an EXIT trap; the script refuses to reuse existing resources.
- Outgoing HTTPS as the service user returned HTTP 200 from the source repository.
- End-to-end access from a **different Tailscale peer was not tested**; same-host
  tailnet-IP requests route through loopback, not the encrypted tunnel.
- No full upstream suite, autonomous-agent run, or full scanner toolchain install
  was performed. Node/npm/npx are now available to the service; katana and dalfox
  are not on its PATH. Other optional tools may need setup.

Re-run the bounded ingress test:

```sh
sudo bash /home/vlad/ARTEX/deploy/verify-ingress.sh
```

## Stop / rollback

```sh
sudo systemctl disable --now artex.service
sudo systemctl disable --now artex-network.service
```

The firewall guard intentionally remains installed when its service stops.
After confirming ARTEX is stopped, remove only its table if uninstalling:

```sh
sudo nft delete table inet artex
```

Do not flush the host ruleset. The PostgreSQL cluster, database, keys, credentials
and data are intentionally retained. Stop the dedicated database only when wanted:
`sudo systemctl stop postgresql@18-artex.service`.
