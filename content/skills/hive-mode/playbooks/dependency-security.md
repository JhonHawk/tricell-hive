# Playbook: dependency-security

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the user asks whether a published security advisory, CVE or package compromise affects their projects, or asks for an exposure sweep across repositories. This sweep replaces the router's steps 0 to 5 and ends as «After the report» sets.

**Stages.** `hive-mode` alone, read-only; the fixes run afterward in each affected project's own session, with the playbook of their kind. **Leaves.** [delegation](../../hive-core/delegation/delegation.md), [local write](../../hive-core/local/local-write.md), [security boundaries](../../hive-core/build/security-boundaries.md) for the fix.

**Reporting.** The sweep's report ends in the analysis row of [report endings](../../hive-core/messages/report-endings.md); a fix closes in its own project with the row of [close report](../../hive-core/close/close-report.md#variants) that the delivery mode in force picks.

Adding or updating a package or tool in a project is not this sweep: before it, run the package check of the global guidance's «Scope and authorization» (`check_package.py` of [flow-build](../../flow-build/SKILL.md)), then plan it as a [feature](feature.md) or [infra-ci](infra-ci.md) change. Reviewing a change's security boundaries is the `hive-review-security` role's, recommended as [delivery questions](../../hive-core/git/delivery-ask.md) sets.

Answer "does this advisory affect us, and where?" with evidence for each repository. The sweep is read-only: it changes no repository, dependency, tracker or deployment, except for the entries of the current project's `HIVE.local.md` and of the machine file that [local write](../../hive-core/local/local-write.md) authorizes. Fixes run afterward, in each project's own session.

## Frame the sweep

Before delegating, read [delegation](../../hive-core/delegation/delegation.md). Requests to advisory databases, registries and search engines follow the global rule on requests to third-party services: send advisory IDs and public package names, never the name of a repository, client or private package.

1. **Read each advisory from a primary source**: the vendor's or project's advisory, GHSA, OSV, NVD, or the CISA KEV catalog for exploitation in the wild. Record its ID, product and ecosystem, affected and fixed versions, any mitigation, the conditions it needs (platform, feature, configuration), its severity, and whether it is exploited. Mark a fact you found only in secondary reporting, and say so in any verdict that rests on it.
2. **Resolve the scope**: the repositories or folders the user names. When the user names none, ask. Reading other projects for this sweep is allowed, as the global rule on changing only the current project sets.

## Collect the evidence

3. **Inventory each repository.** Delegate the sweep to `hive-research` children run side by side, one per repository or group of repositories, launched and briefed as [delegation](../../hive-core/delegation/delegation.md) sets. Give the inventory, yours or each child's, these instructions together with the advisory facts from step 1:
   - For each repository and application, record the current branch, because only the working tree is read.
   - Record the version the lockfile resolves, not the manifest range, with its `path:line`.
   - When the advisory affects a runtime, record its pins: `.nvmrc`, `.node-version`, `engines`, Dockerfile base images, CI and hosting build settings.
   - Record whether the advisory's conditions appear, such as a configuration flag, a feature in use or the platform.
   - Run no install, audit, networked scan, Git fetch or checkout; web lookups carry only advisory IDs and public package names; never print a file that may hold secrets.
   - Report what cannot be seen from the repository, such as deployed versions, images behind a digest or hosting settings.
4. **Check the deciding claims.** Open the locators behind each affected or mitigated verdict, as [hive-mode step 4](../SKILL.md) sets.
5. **Assess reachability where it sets priority.** For an affected application whose exposure depends on how the code reaches the vulnerable path, mark that in the report and offer a reachability assessment by a `hive-review-security` child, which runs when the user asks for it or accepts it; it traces untrusted input to that path without exploiting anything, under the same limits on commands and web lookups as step 3, given the application, the advisory and its conditions. When the user asks for it or accepts it, run that child and update the application's verdict and priority in a follow-up report.

## Report

Lead with the affected repositories. Report critical or actively exploited exposure as soon as it is confirmed, not only at the end. For each repository and application, give:

- the verdict: affected, mitigated (and by what), not affected, or unknown (and what would settle it);
- the evidence locator;
- the fixed version or mitigation it needs, and anything in the repository that blocks it, such as a version cap or override;
- reachability, when the user asked for the assessment that step 5 offers.

State the coverage: repositories and branches read, sources consulted, and what was not checked.

## After the report

The fixes belong to each project's own session, under the global rule on changing only the current project: report, for each affected project, the fixed version or mitigation it needs as a finding for that session. When the current project is affected, end with the next-step block that [hive-mode step 5](../SKILL.md) sets, with the fix as one of its alternatives.
