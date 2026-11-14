# Code audit

Read from [flow-research](../SKILL.md) when the user asks to audit a codebase, or to find what to improve in it, without naming one question to answer. The audit adds to research's steps and replaces none: it sets step 1's aspects, the brief of step 4's children and the report of step 5. It changes no repository file.

## Scope

1. **Target.** Audit the repository that holds the working directory, or the one the user names; reading another project is allowed, and the global guidance's «Scope and authorization» sets what you may offer to change there. When the user names categories or an area, such as a package or a directory, audit only those. Otherwise audit every category below, starting from the hotspots: the 20 files with the most commits in the last 90 days (`git log --since=90.days --name-only --format= | sort | uniq -c | sort -rn | head -20`) and the entry points of the feature the project exists for.
2. **Recon, before judging.** In the main thread, with filtered reads where they suffice, record:
   - the languages, frameworks and package manager, and the exact build, test, lint and type-check commands, taken from the CI workflow files or the project's scripts, never guessed;
   - the conventions the code follows for errors, data access and naming, with one exemplar file each;
   - the decision and intent records that exist, such as `docs/adr/`, `docs/decisions/`, the `Specs` path of the `## Hive` section, `DESIGN.md`, `CONTEXT.md` or a product brief; a tradeoff one of them settles is not a finding;
   - the rejections in force, as [Rejections](#rejections) sets.

   When the repository has no command that tells whether the code works, such as no test runner and no build check, that is the first finding, ranked ahead of any change it would protect.
3. **Aspects.** List in the message, as research step 1 sets, one aspect per group of categories: correctness and security; performance and dependencies; tests and tooling; debt, architecture and docs; direction. An aspect that needs more than 5 large reads goes to a `hive-research` child as research step 4 sets. Its brief adds to [delegation](../../flow-build/references/delegation.md): this file's path with the headings of its categories and of [Finding format](#finding-format); the recon facts that bound the search, including the directories to skip; each tradeoff a decision record settles and each rejection in force, so the child does not report them; and that it returns at most 20 findings, strongest first, with the count of the rest per category, and edits nothing.
4. **Read-only checks.** Run only commands that change no tracked file and no shared state: a type check without output, a linter without fixes, the ecosystem's audit command, such as `npm audit`, `pip-audit`, `cargo audit` or `govulncheck`, and the test suite when the CI files show it needs no external service. Never install, format or write. A check that spends money or quota, changes a shared resource or runs over 30 minutes follows the global guidance's «Scope and authorization».

## Categories

A finding names a location and a consequence a reader can check; a suspicion without one is not a finding.

1. **Correctness.** An error swallowed on a path that writes data or answers a user; asynchronous work nobody awaits; a value that can be null used as if it cannot; empty collections, off-by-one limits and time zones; an enumeration with a branch that silently does nothing; check-then-act on shared state, several writes without a transaction, a retried operation that is not idempotent; clusters of type escapes, such as casts or ignore comments; handles, connections or subscriptions never closed.
2. **Security.** Untrusted input that reaches a query, a shell, a template, HTML or a file path without parameters or validation; authorization checked only in the client, or a record read by its ID without checking its owner; a request body trusted without a schema; production settings such as credentialed CORS for any origin, cookies without `HttpOnly`, `Secure` or `SameSite`, or debug output enabled; sensitive data in logs or error responses; critical or high advisories from the audit command that reach code the project runs or ships. A committed secret is reported by location and type only, as the global guidance's «Secrets» sets, with rotation in its fix. Describe the pattern and its remedy, never a runnable exploit.
3. **Performance.** A query or fetch per item inside a loop; nested scans where a keyed lookup belongs; the same expensive work repeated per request or render with no cache at a clear boundary; unbounded lists without pagination, or whole records fetched where an ID suffices; a heavy dependency loaded for a trivial use, or a rarely used route without code splitting; slow synchronous work that belongs in a queue; a missing index only with the schema and the query as evidence. Claim no speed-up: a fix whose gain needs measuring goes to the `hive-tune-performance` role.
4. **Tests.** A critical path, such as money, permissions, data writes or the project's main feature, with no test or only tests that cannot fail by the filter of flow-build's [Tests](../../flow-build/SKILL.md#tests); a file among the hotspots with no test, which gets characterization tests before any refactor. Pruning an area's existing tests is [test-audit](../../test-audit/SKILL.md)'s work: name it as the finding's fix.
5. **Debt and architecture.** The same logic in three or more places, or copies that drifted; a layer importing another's internals, or an import cycle; dead code, such as a module nothing imports or a flag already on everywhere; a file ten times larger than the repository's median that many files import; three ways of doing one thing, naming the one the newest code uses; an abstraction with one implementation, or a change that always touches the same set of files together.
6. **Dependencies.** A major-version lag with a cost, such as an end of life or a security cutoff; a deprecated API with an announced removal; an abandoned dependency on a critical path; two dependencies doing one job; a manifest and lockfile that disagree. Give the number of files that import each one as its blast radius.
7. **Tooling.** A missing type check, linter, formatter or hook; setup steps in the README that fail; environment variables the code reads that no example file or document names; a feedback loop measured in minutes. A missing or stale `AGENTS.md` is reported by name only: [harness-audit](../../harness-audit/SKILL.md) audits its content.
8. **Docs.** Only where the absence costs something: a published interface without reference documentation; documentation that is wrong, such as an example that no longer compiles; a decision nobody can reconstruct in an area still changing. A decision record the code no longer follows is a finding: report the drift, and never use the record to suppress it.
9. **Direction.** What the code is becoming, each suggestion citing the repository: TODO or FIXME clusters on one theme, flags never turned on, stubs; what the README, the specs or a product brief promise and the code lacks; one-way pairs, such as export without import, or create, read and update without delete; a capability the current architecture makes cheap; manual work around the project that its documents or issues show. A suggestion that would fit any project of its kind is not a finding, and one a decision record rejected is reported as that contradiction instead.

## Finding format

Children number their findings `<category>-<nn>`; the main thread assigns `H<n>` after vetting.

- **Title**, in the imperative.
- **Category.**
- **Evidence:** one to five `path:line` locations, each with one sentence on what is there, and "and about N similar sites" when it is widespread.
- **Impact:** what goes wrong or what it costs, concretely, such as "each order list issues one query per item".
- **Effort:** S (hours), M (about a day) or L (several days), for the fix with its tests.
- **Fix risk:** low, medium or high, with what the fix could break.
- **Confidence:** high when the code was read and the consequence is certain; medium when the signal is strong and one named check would confirm it; low for a smell, whose outcome is an investigation, never a fix.
- **Fix sketch:** one to three sentences, enough to judge the effort.

For a direction finding, the impact says who wants it and why now, the confidence says how well the repository grounds it, and the effort is coarse.

## Vetting

Before reporting a finding, the main thread opens each location it cites, whoever found it: a child's line number and excerpt are leads. Drop or correct:

- behavior by design: a platform convention, such as honoring proxy environment variables, or a tradeoff a decision record settles, unless the implementation adds a risk the convention does not;
- evidence attributed to the wrong file or line;
- duplicates across children;
- a finding that a rejection in force covers.

Count what you dropped, by reason, for the coverage.

## Ranking

Order the findings by leverage: impact divided by effort, lowered by low confidence and by fix risk. Among findings of similar leverage, put first the ones that unblock others, such as a verification baseline or the characterization tests a refactor needs, then high-confidence security findings, then the ones whose fix has a command that proves it. Name each dependency between findings, such as "H4 before H7". "Not worth doing" is a valid verdict: give it one line.

## Report

Write research step 5's report as:

1. A table of at most 15 findings, by leverage: ID, title, category, impact, effort, fix risk, confidence and first location; then the remaining findings as a count per category.
2. Two to four direction findings, apart from the table: options for the user to weigh, not problems ranked against defects.
3. The dependencies between findings.
4. The coverage: the categories and directories audited, each child with its status, the commands run, what was not audited, the findings dropped in vetting by reason, and the rejections in force that suppressed a finding.

End with research step 5's block for an analysis. Its replies include: creating tickets for the findings the user picks; investigating or planning one of them; fixing now a confirmed defect, as research step 5 sets; rejecting findings, which records them; and stopping. A ticket follows the global guidance's ticket rule in «Scope changes», including the search for an existing one, and its body stands alone for someone who never saw this session: the finding's fields, its locations with the audited commit, and the commands from the recon that check the fix. A P0 or security finding with a concrete failure scenario, or an exposed secret, is escalated at once, as «Scope and authorization» sets, not held for the table.

## Rejections

Findings the user rejects are recorded so later audits do not raise them again; findings the main thread drops in vetting are not.

- **Where.** One entry per repository in persistent memory, under the topic `code-audit/rejected/<repository>` of the project the global guidance's «Persistent memory» names. Each line: title · category · locations · audited commit · date · reason.
- **Read.** During recon, search that topic and read the entry whole. A finding matches a line when they share the category and a location. Suppress it while `git diff --stat <commit>..HEAD -- <locations>` shows no change; when the code changed, report it again and say it was rejected before, with the reason.
- **Write.** When the user rejects findings, update the entry in the same turn, or create it when none exists: the user's decision is the write the global guidance's «Persistent memory» allows. Keep every line already there.
