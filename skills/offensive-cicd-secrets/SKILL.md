---
name: offensive-cicd-secrets
description: "Comprehensive secrets extraction methodology targeting CI/CD environments across all major platforms. Covers environment variable extraction from build contexts, exploitation of vault and secrets-manager misconfigurations (HashiCorp Vault, AWS Secrets Manager, Azure Key Vault, GCP Secret Manager), runner and agent token abuse for lateral movement, OIDC federation attacks exploiting trust relationships between CI/CD providers and cloud platforms, build log leakage analysis for inadvertently exposed credentials, cache poisoning techniques for credential exfiltration, platform-specific credential store exploitation (GitHub Actions secrets, GitLab CI variables, Jenkins credential providers), service connection and service account abuse in Azure DevOps and GCP, and Docker registry credential theft from build. Use only for authorized assessments, defensive research, CTFs, or controlled labs."
---

# Offensive CI/CD Secrets Extraction

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

Category: `cicd`. Upstream skill: [Skills/cicd/offensive-cicd-secrets/SKILL.md](https://github.com/SnailSploit/Claude-Red/blob/739512a8588b28ff3b554e669391c22508602878/Skills/cicd/offensive-cicd-secrets/SKILL.md).
