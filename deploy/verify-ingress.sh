#!/bin/bash
# Creates and removes only this test's network namespace/veth pair.
set -euo pipefail
NS=artex-ingress-check
HOST_IF=artex-check-h
PEER_IF=artex-check-p
if ip netns list | grep -q "^${NS}\b" || ip link show "$HOST_IF" &>/dev/null; then
    echo 'Test resources already exist; refusing to reuse them.' >&2
    exit 1
fi
cleanup() {
    ip netns del "$NS" 2>/dev/null || true
    ip link del "$HOST_IF" 2>/dev/null || true
}
trap cleanup EXIT
ip netns add "$NS"
ip link add "$HOST_IF" type veth peer name "$PEER_IF"
ip link set "$PEER_IF" netns "$NS"
ip addr add 192.0.2.1/30 dev "$HOST_IF"
ip link set "$HOST_IF" up
ip -n "$NS" addr add 192.0.2.2/30 dev "$PEER_IF"
ip -n "$NS" link set "$PEER_IF" up
ip -n "$NS" link set lo up
ip -n "$NS" route add default via 192.0.2.1
counter() {
    nft -j list table inet artex | python3 -c 'import sys,json; d=json.load(sys.stdin); print(sum(e["counter"]["packets"] for x in d["nftables"] if "rule" in x for e in x["rule"]["expr"] if "counter" in e))'
}
before=$(counter)
curl --noproxy '*' --max-time 5 -fsS http://100.95.74.28:8787/api/health >/dev/null
for dst in 100.95.74.28 178.105.164.7 10.8.0.1; do
    for port in 8787 8788; do
        case_before=$(counter)
        code=0
        ip netns exec "$NS" curl --noproxy '*' --connect-timeout 2 --max-time 3 -fsS "http://${dst}:${port}/api/health" >/dev/null 2>&1 || code=$?
        case_after=$(counter)
        if [ "$code" -ne 28 ] || [ "$case_after" -le "$case_before" ]; then
            echo "FAIL: ${dst}:${port}: expected timeout and matching firewall drops; exit=${code}, drops=${case_before}->${case_after}"
            exit 1
        fi
        echo "PASS: non-Tailscale ingress blocked to ${dst}:${port} (curl exit ${code})"
    done
done
after=$(counter)
test "$after" -gt "$before"
echo "PASS: dedicated ARTEX drop counter increased ${before} -> ${after}"
