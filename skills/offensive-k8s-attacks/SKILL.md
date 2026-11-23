---
name: offensive-k8s-attacks
description: "Kubernetes cluster attack techniques covering the full attack lifecycle from initial foothold in a pod to cluster-wide compromise. Covers service account token theft and impersonation, RBAC misconfiguration exploitation including wildcard permissions and privilege escalation via role binding, direct etcd access for secret extraction, kubelet API abuse on port 10250 and read-only port 10255, pod escape via hostPID hostNetwork and hostPath volume mounts, Kubernetes secrets enumeration and decoding, admission controller bypass techniques, network policy bypass and lateral movement, cloud metadata service access from pods for credential theft on AWS EKS GCP GKE and Azure AKS, CRD and operator abuse for persistence, and node compromise via DaemonSet deployment. Use only for authorized assessments, defensive research, CTFs, or controlled labs."
---

# Kubernetes Cluster Attacks

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

Category: `container`. Upstream skill: [Skills/container/offensive-k8s-attacks/SKILL.md](https://github.com/SnailSploit/Claude-Red/blob/739512a8588b28ff3b554e669391c22508602878/Skills/container/offensive-k8s-attacks/SKILL.md).
