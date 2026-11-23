---
name: offensive-cloud
description: "Cloud security attack methodology covering AWS, Azure, and GCP. Includes credential harvesting (IMDS, ~/.aws, env vars, leaked CI secrets, instance roles), enumeration with cloud-specific tools (pacu, ScoutSuite, Prowler, ROADtools, gcp_enum), privilege escalation paths (IAM PassRole, AssumeRole chains, Lambda/Functions privilege flips, Azure Owner-on-self, GCP serviceAccountTokenCreator), persistence techniques (IAM user/key creation, AAD app registration, GCP svc account key creation, EventBridge/Logic Apps backdoors), data exfiltration (S3/Blob/GCS, snapshot share, RDS/CosmosDB/Cloud SQL exfil), cloud-native lateral movement (cross-account assume, Azure AD multi-tenant, GCP project hierarchy), serverless attacks (Lambda env vars, layer hijack, Step Functions), Kubernetes-on-cloud (EKS/AKS/GKE-specific. Use only for authorized assessments, defensive research, CTFs, or controlled labs."
---

# Cloud (AWS / Azure / GCP) — Offensive Testing Methodology

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

Category: `cloud`. Upstream skill: [Skills/cloud/offensive-cloud/SKILL.md](https://github.com/SnailSploit/Claude-Red/blob/739512a8588b28ff3b554e669391c22508602878/Skills/cloud/offensive-cloud/SKILL.md).
