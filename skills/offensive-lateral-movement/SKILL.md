---
name: offensive-lateral-movement
description: "Comprehensive lateral movement tradecraft for authorized red team engagements covering credential-based movement (pass-the-hash, pass-the-ticket, overpass-the-hash), NTLM relay attacks (ntlmrelayx with PetitPotam, DFSCoerce, PrinterBug coercion), remote execution protocols (WMI, WinRM, DCOM, PsExec and alternatives), RDP session hijacking, and network pivoting through tunneling tools (chisel, ligolo-ng, SSH tunnels, SOCKS proxies). Provides operator-ready command sequences for mimikatz, crackmapexec/netexec, impacket suite, and evil-winrm with emphasis on OPSEC considerations, SMB signing bypass, and detection evasion. Maps to MITRE ATT&CK T1021 (Remote Services), T1550 (Use Alternate Authentication Material), and sub-techniques. Use only for authorized assessments, defensive research, CTFs, or controlled labs."
---

# Offensive Lateral Movement

Use this skill for work that is explicitly authorized and in scope. Preserve evidence, avoid unnecessary impact, and separate verified findings from hypotheses.

## Workflow

1. Establish the target, authorization boundary, objective, and operational constraints.
2. Inspect the reference headings and load only the sections needed for the current task.
3. Adapt commands and payloads to the actual platform, versions, and available tools; do not assume prerequisites.
4. Prefer the least disruptive method that can answer the testing question.
5. Record commands, timestamps, evidence, uncertainty, impact, and remediation-relevant observations.
6. Stop or narrow the approach when scope, ownership, or safety constraints are unclear.

## Reference

Read [references/methodology.md](references/methodology.md) before applying the methodology. For focused work, locate relevant sections first with:

```bash
rg -n '^#{1,4} |<keyword>' references/methodology.md
```

Category: `post-exploitation`. Upstream skill: [Skills/post-exploitation/offensive-lateral-movement/SKILL.md](https://github.com/SnailSploit/Claude-Red/blob/739512a8588b28ff3b554e669391c22508602878/Skills/post-exploitation/offensive-lateral-movement/SKILL.md).

## Additional Claude-Red methodology

Load [references/claude-red-methodology.md](references/claude-red-methodology.md) for the independently retained current upstream methodology.
