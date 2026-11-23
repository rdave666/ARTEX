---
name: offensive-cicd-pipeline
description: "Comprehensive CI/CD pipeline exploitation methodology covering GitHub Actions injection vectors (expression injection via PR titles and issue bodies, workflow_run event abuse, GITHUB_TOKEN over-scoping, composite action supply chain compromise), Jenkins attack paths (Groovy sandbox escapes, script console remote code execution, Java remoting deserialization, credential store dumping, shared library injection), GitLab CI exploitation (YAML anchor injection, runner registration token abuse, CI variable extraction, protected branch bypass via merge request pipelines), and Azure DevOps pipeline agent compromise with service connection theft. Use only for authorized assessments, defensive research, CTFs, or controlled labs."
---

# Offensive CI/CD Pipeline Exploitation

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

Category: `cicd`. Upstream skill: [Skills/cicd/offensive-cicd-pipeline/SKILL.md](https://github.com/SnailSploit/Claude-Red/blob/739512a8588b28ff3b554e669391c22508602878/Skills/cicd/offensive-cicd-pipeline/SKILL.md).
