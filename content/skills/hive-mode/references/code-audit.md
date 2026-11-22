# Code audit

Read from [hive-mode](../SKILL.md) when the user asks to audit a codebase, or to find what to improve in it, without naming one question to answer. The audit adds to the router's steps and replaces none: it sets step 1's aspects, the brief of step 4's children and the report of step 5. It changes no repository file.

## Scope

1. **Target.** Audit the repository that holds the working directory, or the one the user names; the global guidance's «Scope and authorization» sets what you may offer to change in another project. When the user names categories or an area, audit only those. Otherwise audit every category below, starting from the hotspots: the 20 files with the most commits in the last 90 days, counted per path from `git log --since=90.days --name-only --format=`, skipping empty lines and paths no longer in the tree, and the entry points of the feature the project exists for.
2. **Recon, before judging.** In the main thread, with filtered reads where they suffice, record:
   - the exact build, test, lint and type-check commands, taken from the CI workflow files or the project's scripts, and the conventions the code follows for errors, data access and naming, with one exemplar file each;
   - the decision and intent records that exist, such as `docs/adr/`, `docs/decisions/`, the `Specs` path of the `## Hive` section, `DESIGN.md`, `CONTEXT.md` or a product brief; a tradeoff one of them settles is not a finding;
   - the rejections in force, as [Rejections](#rejections) sets.

   When the repository has no command that tells whether the code works, such as no test runner and no build check, that is the first finding, ranked ahead of any change it would protect.
3. **Aspects.** List in the message, as the router's step 1 sets, one aspect per group of categories: correctness and security; performance and dependencies; tests and tooling; debt, architecture and docs; direction. An aspect goes to a `hive-research` child only when research step 4 sends it to one. A child's brief adds to [delegation](../../hive-core/delegation/delegation.md): this file's path with the headings of its categories and of [Finding format](#finding-format); the recon facts that bound the search, including the directories to skip; each tradeoff a decision record settles and each rejection in force, so the child does not report them; and that it returns at most 20 findings, strongest first, with the count of the rest per category, and edits nothing.
4. **Read-only checks.** Run only a type check and a linter in a form that writes nothing, such as `tsc --noEmit` or a linter without its fix option, in a repository of the current project; a project script that builds, writes or fixes is not run. Run the test suite (it can write snapshots or fixtures), a dependency audit command such as `npm audit` or `govulncheck` (it reaches an outside service and, for some, sends the dependency list there), or any code of a repository outside the current project, only when the user asks for it in this audit, and name each one skipped in the coverage. A check that spends money or quota, changes a shared resource or runs over 30 minutes follows [long checks](../../hive-core/build/checks.md#long-checks).

## Categories

A finding names a location and a consequence a reader can check; a suspicion without one is not a finding.

1. **Correctness.** Errors swallowed on a path that writes data or answers a user, unawaited asynchronous work, unchecked nulls, check-then-act on shared state, several writes without a transaction, a retried operation that is not idempotent, clusters of type escapes, handles never closed.
2. **Security.** Untrusted input reaching a query, shell, template, HTML or file path without parameters or validation; authorization checked only in the client, or a record read by its ID without its owner; production settings such as credentialed CORS for any origin, cookies without `HttpOnly`, `Secure` or `SameSite`, or debug output enabled; sensitive data in logs or error responses; critical or high advisories from the dependency audit, when it ran, that reach code the project runs or ships. A committed secret is reported by location and type only, as the global guidance's «Secrets» sets, with rotation in its fix. Describe the pattern and its remedy, never a runnable exploit.
3. **Performance.** A query or fetch per item inside a loop, expensive work repeated per request with no cache at a clear boundary, unbounded lists without pagination, slow synchronous work that belongs in a queue, a missing index only with the schema and the query as evidence. Claim no speed-up: a fix whose gain needs measuring goes to the `hive-tune-performance` role.
4. **Tests.** A critical path, such as money, permissions, data writes or the project's main feature, with no test or only tests that cannot fail by the filter of [Tests](../../hive-core/build/tests.md); a hotspot file with no test, which gets characterization tests before any refactor. Pruning an area's existing tests is [test-audit](../../test-audit/SKILL.md)'s work: name it as the finding's fix.
5. **Debt and architecture.** The same logic in three or more places or copies that drifted, a layer importing another's internals, an import cycle, dead code, three ways of doing one thing (naming the one the newest code uses), an abstraction with one implementation.
6. **Dependencies.** A major-version lag with a cost (end of life, a security cutoff), a deprecated API with an announced removal, an abandoned dependency on a critical path, two dependencies doing one job, a manifest and lockfile that disagree. Give the number of files that import each one as its blast radius.
7. **Tooling.** A missing type check, linter, formatter or hook; README setup steps that fail; environment variables the code reads that no example file or document names; a feedback loop measured in minutes. A missing or stale `AGENTS.md` is reported by name only: [harness-audit](../../harness-audit/SKILL.md) audits its content.
8. **Docs.** Only where the absence costs something: a published interface without reference documentation, documentation that is wrong, a decision nobody can reconstruct in an area still changing. A decision record the code no longer follows is a finding: report the drift, and never use the record to suppress it.
9. **Direction.** What the code is becoming, each suggestion citing the repository: TODO or FIXME clusters on one theme, flags never turned on, stubs, what the README, the specs or a product brief promise and the code lacks, one-way pairs such as export without import, manual work its documents or issues show. A suggestion that would fit any project of its kind is not a finding, and one a decision record rejected is reported as that contradiction instead.

## Finding format

Children number their findings `<category>-<nn>`; the main thread assigns `H<n>` after vetting. Each finding has:

- **Title**, in the imperative, and **Category**.
- **Evidence:** one to five `path:line` locations, each with one sentence on what is there, and "and about N similar sites" when it is widespread.
- **Impact:** high for data loss, a security exposure, an outage or a wrong result in the feature the project exists for; medium for a defect or cost met in normal use; low for one that is rare, cosmetic or internal. Add what goes wrong concretely, such as "each order list issues one query per item".
- **Effort:** S (hours), M (about a day) or L (several days), for the fix with its tests; **Fix risk:** low, medium or high, with what the fix could break.
- **Confidence:** high when the code was read and the consequence is certain; medium when the signal is strong and one named check would confirm it; low for a smell, whose outcome is an investigation, never a fix.
- **Fix sketch:** one to three sentences, enough to judge the effort.

For a direction finding, the impact says who wants it and why now, the confidence says how well the repository grounds it, and the effort is coarse.

## Vetting

Before reporting a finding, the main thread opens each location it cites, whoever found it: a child's line number and excerpt are leads. Drop or correct behavior by design (a platform convention, or a tradeoff a decision record settles, unless the implementation adds a risk the convention does not), evidence attributed to the wrong file or line, duplicates across children, and a finding that a rejection in force covers. Count what you dropped, by reason, for the coverage.

## Ranking

Put first, whatever their impact, the findings that unblock others, such as a verification baseline or the characterization tests a refactor needs. Order the rest by impact, then confidence, then smallest effort, then lowest fix risk, with high-confidence security findings at the front of their impact level. Name each dependency between findings, such as "H4 before H7". "Not worth doing" is a valid verdict: give it one line.

## Report

Write the report of the router's step 5 as:

1. A table of at most 15 findings, in the order [Ranking](#ranking) sets: ID, title, category, impact, effort, fix risk, confidence and first location; then the remaining findings as a count per category.
2. When the audit covers direction, up to four direction findings, apart from the table: options for the user to weigh, not problems ranked against defects. Report none when the repository grounds none.
3. The dependencies between findings.
4. The coverage: the categories and directories audited, each child with its status, the commands run, what was not audited, the findings dropped in vetting by reason, and the rejections in force that suppressed a finding.

End with the reply block of the analysis row of [report endings](../../hive-core/messages/report-endings.md). Its replies include: creating tickets for the findings the user picks; investigating or planning one of them; fixing now a confirmed defect, as that row sets; rejecting findings, which records them; and stopping. A finding gets a ticket when the ticket rule of [scope changes](../../hive-core/plan/scope-changes.md) covers it, and each one is a single ticket created as [create-issue](../../create-issue/SKILL.md) sets. Its body stands alone for someone who never saw this session: the finding's fields, its locations with the audited commit, and the commands from the recon that check the fix. Two exceptions to that skill's limits apply here: the locations go in the ticket's evidence, all of them, not in its hint; and a picked finding stays one ticket, never a parent with children, even when its locations span repositories. A P0 or security finding with a concrete failure scenario, or an exposed secret, is escalated at once, as «Scope and authorization» sets, not held for the table.

## Rejections

Findings the user rejects are recorded so later audits do not raise them again; findings the main thread drops in vetting are not.

- **Where.** One entry per repository in persistent memory, under the topic `code-audit/rejected/<repository>`, where `<repository>` is the folder name of the repository's main checkout, the path on the first `worktree` line of `git worktree list --porcelain`, so its worktrees share the entry; it lives in the project the global guidance's «Persistent memory» names for that repository. Each line: title · category · files, each with its lines · audited commit · date · reason.
- **Read.** During recon, search that topic and read the entry whole. A finding matches a line when they share the category, at least one cited file and the problem the title names; when several lines match, the newest governs. First run `git cat-file -e <commit>`: when it fails, the commit no longer exists, so report the finding again. Otherwise suppress the finding while `git diff --stat <commit> -- <files>` exits 0 and prints nothing, where `<files>` are the files the new finding cites, without line numbers; it compares the recorded commit with the working tree. When it prints a change or fails, report the finding again. A finding reported again says it was rejected before, with the reason.
- **Write.** When the user rejects findings, update the entry in the same turn, or create it when none exists: the user's decision is the write the global guidance's «Persistent memory» allows. Keep every line already there.
