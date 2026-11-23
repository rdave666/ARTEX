---
name: offensive-windows-privesc
description: "Comprehensive Windows privilege escalation methodology for offensive security engagements. Covers the full attack surface from a standard user shell to NT AUTHORITY\\\\SYSTEM: token impersonation via SeImpersonate and SeAssignPrimaryToken privileges using JuicyPotato, PrintSpoofer, GodPotato, SweetPotato, and RoguePotato; service misconfigurations including unquoted service paths, weak service DACLs, writable service binaries, and insecure service creation permissions; AlwaysInstallElevated MSI exploitation; DLL hijacking through search order abuse, phantom DLL loading, and writable PATH directory injection; UAC bypass techniques via fodhelper.exe, eventvwr.exe, CMSTP, and environment variable manipulation; scheduled task abuse for writable task actions and new task creation;. Use only for authorized assessments, defensive research, CTFs, or controlled labs."
---

# Windows Privilege Escalation

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

Category: `privesc`. Upstream skill: [Skills/privesc/offensive-windows-privesc/SKILL.md](https://github.com/SnailSploit/Claude-Red/blob/739512a8588b28ff3b554e669391c22508602878/Skills/privesc/offensive-windows-privesc/SKILL.md).

## Additional Claude-Red methodology

Load [references/claude-red-methodology.md](references/claude-red-methodology.md) for the independently retained current upstream methodology.
