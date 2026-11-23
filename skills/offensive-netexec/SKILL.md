---
name: offensive-netexec
description: "Use this skill whenever the user asks about NetExec (nxc) — a network exploitation and post-exploitation tool for Active Directory environments. Triggers include: any mention of 'nxc', 'netexec', 'crackmapexec' successor questions, AD enumeration, SMB/LDAP/WinRM/MSSQL/SSH/RDP/VNC/WMI/FTP/NFS protocol attacks, password spraying, credential dumping (SAM, NTDS, LSASS, DPAPI), Kerberoasting, ASREPRoasting, lateral movement, BloodHound collection, module usage, or any pentest workflow involving Windows domain environments. This skill covers ALL protocols, ALL modules, and ALL core features of NetExec. Always provide full command examples with correct flags and options. Use only for authorized assessments, defensive research, CTFs, or controlled labs."
---

# NetExec (nxc) — Reference Skill

Use this skill for the task's target and objective. Preserve evidence and distinguish verified findings from hypotheses.

## Workflow

1. Identify the target, objective, and constraints supplied by the task.
2. Inspect the reference headings and load only the sections needed for the current task.
3. Adapt commands and payloads to the actual platform, versions, and available tools; check prerequisites.
4. Select a method that answers the testing question, using a reversible option when it proves the same result.
5. Record commands, timestamps, evidence, uncertainty, impact, and remediation-relevant observations.
6. If a tool, target, or assumption changes, diagnose and choose a viable next method rather than silently stopping.

## Reference

Read [references/methodology.md](references/methodology.md) before applying the methodology. For focused work, locate relevant sections first with:

```bash
rg -n '^#{1,4} |<keyword>' references/methodology.md
```

Category: `active-directory`. Upstream skill: [Skills/active-directory/offensive-netexec/SKILL.md](https://github.com/SnailSploit/Claude-Red/blob/739512a8588b28ff3b554e669391c22508602878/Skills/active-directory/offensive-netexec/SKILL.md).
