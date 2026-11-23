---
name: offensive-advanced-redteam
description: "Comprehensive red team operations methodology covering full engagement lifecycle from planning through reporting. Addresses engagement scoping and rules of engagement negotiation, multi-tier C2 infrastructure design with redirectors and domain fronting, malleable traffic profiles and beacon tradecraft, OPSEC discipline including attribution avoidance and indicator management, EDR and AMSI evasion techniques using direct syscalls and unhooking, data collection with chain-of-custody controls, and structured reporting with purple team debrief workflows. Covers assumed-breach, external-to-internal, insider threat, and hybrid physical-cyber engagement scenarios with MITRE ATT&CK mapping throughout. Targets operators planning or executing adversary simulation engagements against mature defenders. Use only for authorized assessments, defensive research, CTFs, or controlled labs."
---

# Advanced Red Team Operations

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

Category: `infrastructure`. Upstream skill: [Skills/infrastructure/offensive-advanced-redteam/SKILL.md](https://github.com/SnailSploit/Claude-Red/blob/739512a8588b28ff3b554e669391c22508602878/Skills/infrastructure/offensive-advanced-redteam/SKILL.md).

## Additional Claude-Red methodology

Load [references/claude-red-methodology.md](references/claude-red-methodology.md) for the independently retained current upstream methodology.
