---
name: skill-index
description: "Categorized ARTEX skill catalog and role routing. Read before selecting task-specific procedures; do not load all methodologies into one prompt."
---

# ARTEX Skill Index

Workflow: discover capabilities → load relevant Skill → check prerequisites → execute → verify results.

The Skill tool is the source of the currently assigned/visible skills. This catalog groups the installed knowledge packs; a name appearing here does not override agent visibility or prove a CLI exists.

## Role routing

- Planner, goal decomposer, main orchestrator: identify domain and delegate focused technical work to worker or pentest; use reporting/recon skills for synthesis.
- Worker, pentest, auto: load the matching technical skill, check tool/credential/environment prerequisites, execute, record evidence.
- Retester: load the finding class and reproduction skill; validate the specific fix and negative control.
- Reporter: load reporting and evidence skills; distinguish verified findings from hypotheses.

## recon

`offensive-osint`, `offensive-osint-methodology`

## web

`offensive-business-logic`, `offensive-deserialization`, `offensive-file-upload`, `offensive-graphql`, `offensive-idor`, `offensive-open-redirect`, `offensive-parameter-pollution`, `offensive-race-condition`, `offensive-rce`, `offensive-request-smuggling`, `offensive-sqli`, `offensive-ssrf`, `offensive-ssti`, `offensive-waf-bypass`, `offensive-xss`, `offensive-xxe`

## auth

`offensive-jwt`, `offensive-oauth`

## infrastructure

`offensive-advanced-redteam`, `offensive-edr-evasion`, `offensive-initial-access`, `offensive-keylogger-arch`, `offensive-shellcode`, `offensive-windows-boundaries`, `offensive-windows-mitigations`

## network

`offensive-network-attacks`

## privesc

`offensive-linux-privesc`, `offensive-windows-privesc`

## post-exploitation

`offensive-data-exfiltration`, `offensive-lateral-movement`, `offensive-persistence`

## exploit-dev

`offensive-basic-exploitation`, `offensive-crash-analysis`, `offensive-exploit-dev-course`, `offensive-exploit-development`, `offensive-mitigations`, `offensive-toctou`

## fuzzing

`offensive-bug-identification`, `offensive-fuzzing`, `offensive-fuzzing-course`, `offensive-vuln-classes`

## utility

`offensive-fast-checking`, `offensive-reporting`

## Other installed ARTEX skills

`api-recon`, `api-security`, `bbot-recon`, `codeaudit`, `edr-bypass-re`, `kali-capabilities`, `offensive-active-directory`, `offensive-ai-security`, `offensive-cicd-pipeline`, `offensive-cicd-secrets`, `offensive-cloud`, `offensive-k8s-attacks`, `offensive-netexec`, `offensive-supply-chain`, `patch-diff`, `php2shell`, `playwright-cli`, `rea`, `scopesentry`, `third-party-code-review`

For a skill not visible to your role, ask the orchestrator to assign the task to an agent that can load it; do not assume the methodology was loaded.
