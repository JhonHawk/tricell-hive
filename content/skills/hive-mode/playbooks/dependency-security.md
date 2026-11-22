# Playbook: dependency-security

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the user asks whether a published security advisory, CVE or package compromise affects their projects, or asks for an exposure sweep across repositories. This sweep replaces the router's steps 0 to 5 and ends as «After the report» sets.

**Stages.** `hive-mode` alone, read-only: it changes no repository, dependency, tracker or deployment, except for the writes [local write](../../hive-core/local/local-write.md) authorizes. The fixes run afterward in each affected project's own session, with the playbook of their kind; adding or updating a package is not this sweep, and its package check is the global guidance's «Scope and authorization» one (`check_package.py` of [flow-build](../../flow-build/SKILL.md)). Reviewing a change's security boundaries is the `hive-review-security` role's, recommended as [delivery questions](../../hive-core/git/delivery-ask.md) sets.

Answer "does this advisory affect us, and where?" with evidence for each repository.

## Frame the sweep

Requests to advisory databases, registries and search engines carry advisory IDs and public package names, never the name of a repository, client or private package.

1. **Read each advisory from a primary source**: the vendor's or project's advisory, GHSA, OSV, NVD, or the CISA KEV catalog for exploitation in the wild. Record its ID, product and ecosystem, affected and fixed versions, any mitigation, the conditions it needs (platform, feature, configuration), its severity, and whether it is exploited. Mark a fact found only in secondary reporting, and say so in any verdict that rests on it.
2. **Resolve the scope**: the repositories or folders the user names; when they name none, ask.

## Collect the evidence

3. **Inventory each repository.** Delegate to `hive-research` children side by side, one per repository or group, as [delegation](../../hive-core/delegation/delegation.md) sets. Give the inventory, yours or each child's, these instructions with the advisory facts from step 1:
   - Record the current branch (only the working tree is read), the version the lockfile resolves, not the manifest range, with its `path:line`, runtime pins when the advisory affects a runtime (`.nvmrc`, `engines`, Dockerfile base images, CI and hosting build settings), and whether the advisory's conditions appear.
   - Run no install, audit, networked scan, Git fetch or checkout; never print a file that may hold secrets.
   - Report what cannot be seen from the repository, such as deployed versions, images behind a digest or hosting settings.
4. **Check the deciding claims.** Open the locators behind each affected or mitigated verdict, as [hive-mode step 4](../SKILL.md) sets.
5. **Assess reachability where it sets priority.** For an affected application whose exposure depends on how the code reaches the vulnerable path, mark that in the report and offer a reachability assessment by a `hive-review-security` child, which runs when the user asks for it or accepts it; it traces untrusted input to that path without exploiting anything, under step 3's limits on commands and web lookups, and updates the application's verdict and priority in a follow-up report.

## Report

Lead with the affected repositories, and report critical or actively exploited exposure as soon as it is confirmed. For each repository and application, give:

- the verdict: affected, mitigated (and by what), not affected, or unknown (and what would settle it);
- the evidence locator;
- the fixed version or mitigation it needs, and anything in the repository that blocks it, such as a version cap or override;
- reachability, when the user asked for step 5's assessment.

State the coverage: repositories and branches read, sources consulted, and what was not checked.

## After the report

The fixes belong to each project's own session, under the global rule on changing only the current project: report, for each affected project, the fixed version or mitigation it needs as a finding for that session. When the current project is affected, end with the reply block of the analysis row of [report endings](../../hive-core/messages/report-endings.md), with the fix as one of its replies.
