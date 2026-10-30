---
name: security-advisory
description: Check whether a published security advisory, CVE or package compromise affects a set of repositories, and report each one's exposure with evidence, without changing them. Use when the user asks whether an advisory or a supply-chain incident affects their projects, or asks for an exposure sweep across repositories; not for reviewing a change's security boundaries or for checking a package before installing it.
---

# Security advisory exposure

Answer "does this advisory affect us, and where?" with evidence for each repository. The sweep is read-only: it changes no repository, dependency, tracker or deployment. Fixes run afterward, in each project's own session.

## Frame the sweep

Before delegating, read [delegation and waits](../flow-research/references/delegation.md). Before the report, read [reporting](../flow-close/references/reporting.md); before a user decision, read [questions](../hive-core/references/questions.md). Requests to advisory databases, registries and search engines follow the global rule on requests to third-party services: send advisory IDs and public package names, never the name of a repository, client or private package.

1. **Read each advisory from a primary source**: the vendor's or project's advisory, GHSA, OSV, NVD, or the CISA KEV catalog for exploitation in the wild. Record its ID, product and ecosystem, affected and fixed versions, any mitigation, the conditions it needs (platform, feature, configuration), its severity, and whether it is exploited. Mark a fact you found only in secondary reporting, and say so in any verdict that rests on it.
2. **Resolve the scope**: the repositories or folders the user names. When the user names none, ask. Reading other projects for this sweep is allowed, as the global rule on changing only the current project sets.

## Collect the evidence

3. **Inventory each repository.** Delegate the sweep to `hive-research` children by the criterion and split that [delegation and waits](../flow-research/references/delegation.md#split-work) sets. Give the inventory, yours or each child's, these instructions together with the advisory facts from step 1:
   - For each repository and application, record the current branch, because only the working tree is read.
   - Record the version the lockfile resolves, not the manifest range, with its `path:line`.
   - When the advisory affects a runtime, record its pins: `.nvmrc`, `.node-version`, `engines`, Dockerfile base images, CI and hosting build settings.
   - Record whether the advisory's conditions appear, such as a configuration flag, a feature in use or the platform.
   - Run no install, audit, networked scan, Git fetch or checkout; web lookups carry only advisory IDs and public package names; never print a file that may hold secrets.
   - Report what cannot be seen from the repository, such as deployed versions, images behind a digest or hosting settings.
4. **Check the deciding claims.** Open the locators behind each affected or mitigated verdict, as [flow-research](../flow-research/SKILL.md) sets for checking a child's claims.
5. **Assess reachability where it sets priority.** For an affected application whose exposure depends on how the code reaches the vulnerable path, mark that in the report and offer a reachability assessment by a `hive-review-security` child, which runs only if the user asks; it traces untrusted input to that path without exploiting anything, under the same limits on commands and web lookups as step 3, given the application, the advisory and its conditions. When the user asks for it, run that child as [security review](../flow-verify/references/security-review.md) sets for the brief and the report, and update the application's verdict and priority in a follow-up report.

## Report

Lead with the affected repositories. Report critical or actively exploited exposure as soon as it is confirmed, not only at the end. For each repository and application, give:

- the verdict: affected, mitigated (and by what), not affected, or unknown (and what would settle it);
- the evidence locator;
- the fixed version or mitigation it needs, and anything in the repository that blocks it, such as a version cap or override;
- reachability, when the user asked for the assessment that step 5 offers.

State the coverage: repositories and branches read, sources consulted, and what was not checked.

## After the report

The fixes belong to each project's own session, under the global rule on changing only the current project: report, for each affected project, the fixed version or mitigation it needs as a finding for that session. When the current project is affected, end with the route question that [flow-research](../flow-research/SKILL.md) sets, with the fix as one of its alternatives.
