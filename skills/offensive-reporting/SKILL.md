---
name: offensive-reporting
description: "Penetration test and red team report writing methodology. Covers executive summary structuring (risk-led narrative for non-technical readers), technical finding format (title, severity, affected scope, narrative, reproduction steps, impact, remediation, references), CVSS v3.1 / v4.0 scoring with vector justification, OWASP risk rating, evidence hygiene (redacting credentials, hashing client data, time-stamping every action), screenshot and PoC artifact management, finding chain narratives, scope/limitations/assumptions documentation, retest evidence and remediation tracking, deliverable formats (PDF, DOCX, HTML, JSON for SIEM ingestion), client-customer-deliverable separation, and common report mistakes (over-CVSSing, undermining the triager, missing the 'so what'). Use only for authorized assessments, defensive research, CTFs, or controlled labs."
---

# Penetration Test Reporting — Professional Methodology

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

Category: `utility`. Upstream skill: [Skills/utility/offensive-reporting/SKILL.md](https://github.com/SnailSploit/Claude-Red/blob/739512a8588b28ff3b554e669391c22508602878/Skills/utility/offensive-reporting/SKILL.md).

## Additional Claude-Red methodology

Load [references/claude-red-methodology.md](references/claude-red-methodology.md) for the independently retained current upstream methodology.
