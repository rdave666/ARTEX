---
name: offensive-supply-chain
description: "Comprehensive offensive methodology for software supply chain attacks covering the full kill chain from reconnaissance through exploitation. Addresses dependency confusion across npm, PyPI, and NuGet ecosystems where internal registry override allows an attacker to inject malicious packages that shadow private dependencies. Covers typosquatting techniques for popular packages, compromised package injection via maintainer account takeover or social engineering, and build system attacks through Makefile injection, setup.py install hooks, and npm postinstall scripts. Use only for authorized assessments, defensive research, CTFs, or controlled labs."
---

# Offensive Supply Chain Attacks

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

Category: `supply-chain`. Upstream skill: [Skills/supply-chain/offensive-supply-chain/SKILL.md](https://github.com/SnailSploit/Claude-Red/blob/739512a8588b28ff3b554e669391c22508602878/Skills/supply-chain/offensive-supply-chain/SKILL.md).
