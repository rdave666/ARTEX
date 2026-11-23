#!/usr/bin/env python3
"""Initialize admin without writing tokens or passwords to command output."""
import json, os, pathlib, re, secrets, subprocess, urllib.request

base = 'http://100.95.74.28:8787'
http = urllib.request.build_opener(urllib.request.ProxyHandler({}))

def request(path, data=None, token=None):
    headers = {'Content-Type': 'application/json'}
    if token:
        headers['Authorization'] = 'Bearer ' + token
    req = urllib.request.Request(base + path, data=None if data is None else json.dumps(data).encode(), headers=headers)
    with http.open(req, timeout=15) as response:
        return json.load(response)

credentials = pathlib.Path('/home/vlad/.config/artex/admin-credentials.txt')
credentials.parent.mkdir(parents=True, exist_ok=True, mode=0o700)
os.chmod(credentials.parent, 0o700)
if request('/api/auth/status')['initialized']:
    raise SystemExit('Already initialized; no changes made.')
logs = subprocess.check_output(['sudo', '-n', 'journalctl', '-u', 'artex.service', '--no-pager', '-o', 'cat'], text=True)
setup_tokens = re.findall(r'Setup token for the /setup page: ([0-9a-f]+)', logs)
assert setup_tokens, 'No setup token found'
password = secrets.token_urlsafe(30)
# Save before initializing so a failed process cannot lose the admin password.
fd = os.open(credentials, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
with os.fdopen(fd, 'w') as output:
    output.write(f'URL: {base}/login/\nUsername: ARTEX\nPassword: {password}\n')
request('/api/auth/init', {'password': password, 'setup_token': setup_tokens[-1]})
login = request('/api/auth/login', {'username': 'ARTEX', 'password': password})
health = request('/api/health', token=login['token'])
assert health['service'] == 'artex' and health['version'] == '9f99b71', health
assert request('/api/auth/status')['initialized']
print(json.dumps({'admin_initialized': True, 'login_verified': True, 'authenticated_health': health, 'credential_file': str(credentials)}, indent=2))
