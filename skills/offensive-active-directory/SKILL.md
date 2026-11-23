---
name: offensive-active-directory
description: "Active Directory attack methodology for internal network red team engagements. Covers reconnaissance (BloodHound, PowerView, ADExplorer), credential abuse (Kerberoasting, ASREProasting, NTLM relay, LLMNR/NBT-NS poisoning), privilege escalation (ACL abuse, GPO abuse, unconstrained/constrained delegation), lateral movement (Pass-the-Hash, Pass-the-Ticket, Overpass-the-Hash, WMI/WinRM/PsExec), persistence (Golden/Silver/Diamond Tickets, DCSync, DCShadow, AdminSDHolder, Skeleton Key), forest trust attacks, ADCS abuse (ESC1-ESC15), and modern MDI/Defender for Identity evasion. Use when assessing on-prem AD, hybrid AD/Entra ID environments, or ADCS deployments. Use only for authorized assessments, defensive research, CTFs, or controlled labs."
---

# Active Directory — Offensive Testing Methodology

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

Category: `active-directory`. Upstream skill: [Skills/active-directory/offensive-active-directory/SKILL.md](https://github.com/SnailSploit/Claude-Red/blob/739512a8588b28ff3b554e669391c22508602878/Skills/active-directory/offensive-active-directory/SKILL.md).
