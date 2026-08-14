# CI Stage Placement Canon

> Where automated checks and code review (human or AI) run, for products using environment
> branches (`development → qa → production`, where a promotion PR moves the accumulated state
> of development into qa). Generic by design — no project-specific facts — so any workspace
> can adopt it verbatim.
>
> **Provenance:** adversarial research protocol (2026-08-14): 3 independent generators
> (canon-first / industry-practice / devil's advocate) → 1 adversarial cross-examiner that
> fetched every cited source. 61 claims: 49 confirmed, 9 weakened (fixes applied below),
> 1 refuted, 1 unverifiable, 1 empty slot (no claim asserted). Confidence: high — the corpus is unanimous on every directive
> below; each carries its verified citation.
>
> Stage 0 (local git hooks) was added after the protocol, grounded in the toolchain in force
> (lefthook) and the already-verified corpus — it did not go through the 61-claim cross-exam.

## The canon

### 0. Local git hooks (lefthook) front-run the presubmit — advisory, never the enforcement layer

Before anything reaches a PR, local hooks give the fastest possible feedback — but they are a
**convenience layer, not a gate**: any client-side hook is bypassable (`--no-verify`) and runs
on whatever the developer's machine happens to have, so the server-side PR CI remains the
enforcement layer and re-runs everything the hooks ran.

- **pre-commit** — deterministic, per-file, seconds-fast, autofixable checks on **staged files
  only**: formatter first, then linter with safe fixes, re-staging the fixes
  (`stage_fixed: true` in lefthook 2.x `jobs` syntax). Never whole-project work here.
- **pre-push** — checks that cannot scope to staged files but still run in well under a
  minute: the whole-project typecheck (the type graph can't be scoped per file).
- **Never in hooks:** builds, unit/integration tests, anything needing services or network —
  non-deterministic or slow checks train developers to bypass the hook, which then bypasses
  the deterministic ones too. They belong to stages 1–2.
- **Document the bypass in the hook config itself** (`git commit --no-verify`) — the audited
  escape valve principle from the operational corollaries applies at this layer too.
- Canonical grounding: DORA requires test/check feedback "in less than ten minutes **both on
  local workstations and** from the continuous integration system"
  ([test automation](https://dora.dev/capabilities/test-automation/)) — the local half of that
  sentence is this stage. Enforcement asymmetry is structural: Google's presubmit gates at
  *submit*, server-side; a client hook can never play that role.

### 1. Fast test subset → pre-merge, blocking, on the PR into `development`

Lint, typecheck, and the dependency-selected affected-test subset gate the feature PR.
Feedback SLA: **under 10 minutes** — above ~15, developers merge without waiting and the
signal erodes.

> **Portfolio refinement (2026-08-14): the build lives ONLY in the full suite (stage 2),
> never in the PR gate.** Build is typically the slowest step; typecheck in the quick gate
> catches most compile failures, and the residual (a non-building merge) is caught minutes
> later by the post-merge full suite under red-trunk discipline — the cost principle
> Google documents for keeping expensive checks off presubmit, applied to the build. Accepted consequence: a PR
> that fails to build can merge and briefly redden the trunk; the fix is revert-fast, not
> re-adding build to the PR gate.

- Google runs small tests presubmit: "the obvious ones to run as they tend to be the fastest
  and most reliable"; the submit is gated — "If the tests pass, the change is allowed into
  the codebase" ([SWE at Google ch. 23](https://abseil.io/resources/swe-book/html/ch23.html)).
- MinimumCD: "Work has automated testing **before** merge to trunk"
  ([minimumcd.org](https://minimumcd.org)).
- DORA: tests "must run in a few minutes or less"; feedback "in less than ten minutes"
  ([CI](https://dora.dev/capabilities/continuous-integration/),
  [test automation](https://dora.dev/capabilities/test-automation/)).
- Industry practice converges: Shopify ("If we run CI *before* merging to master, we ensure
  that only green changes merge"), Meta (diff-time selected tests + review-gated landing),
  Uber SubmitQueue (speculative pre-land validation of an always-green mainline).

### 2. Full/heavy suite → post-merge on `development`, asynchronous — never per-PR, never deferred to promotion

The full suite — **including the build** (per the refinement above) plus integration/e2e —
does NOT run on every PR; Google's stated reason: "The main reason is that it's too
expensive." It runs **after each merge** to the integration branch, because a class
of defects only exists post-integration:

- **Mid-air collisions:** two independently-green changes break the combination
  (SWE ch. 23). Empirically, dynamic semantic conflicts appear in 3–28% of merge scenarios
  per project (Brun et al., ESEC/FSE 2011).
- DORA requires both sides: "A set of automated tests is run **both before and after the
  merge**"; MinimumCD: "Work is tested with other work automatically on merge."
- Detecting this at the dev→qa promotion instead is too late to attribute: "a test failure
  can originate from many different changes, making it hard to debug" (DORA CI). GitHub's
  own 15-PR deploy trains "frequently derailed" for exactly this reason before it moved to
  merge queues.
- **Merge queues** (GitHub merge groups, GitLab merge trains) are the high-volume
  optimization that collapses both runs into one pre-merge test of the combined state —
  adopt one when merge volume makes post-merge breakage frequent; below that volume,
  post-merge async is the canon.

### 3. Code review — human and AI — attaches to the feature PR, in-team, in-platform

- DORA: "Use peer review to meet the goal of segregation of duties, with reviews, comments,
  and approvals captured in the team's development platform" — and found **no evidence**
  that a more formal, separated review stage lowers change-fail rates
  ([streamlining change approval](https://dora.dev/capabilities/streamlining-change-approval/)).
- Review effectiveness collapses with batch size: review "no more than 200 to 400 lines of
  code at a time"; 200–400 LOC over 60–90 min yields 70–90% defect discovery
  ([SmartBear/Cisco study](https://smartbear.com/learn/code-review/best-practices-for-peer-code-review/)).
  A promotion PR is typically far past that cliff — review placed there operates in the
  lowest-detection regime while re-reviewing already-approved commits.
- **AI reviewers are built for this attach point.** Anthropic's claude-code-action documents
  auto-review as `on: pull_request: [opened, synchronize]`; Cursor Bugbot "runs automatic
  reviews on every PR update" over PR diffs; Copilot code review is per-PR via rulesets;
  CodeRabbit's `base_branches` **defaults to `[]`** — auto-review fires only on PRs into the
  default branch, so promotion PRs into `qa`/`production` get no AI review unless someone
  opts in. Pointing an AI reviewer at a promotion PR is supported configuration but
  off-design use.
- **Blocking vs post-integration review is a risk-tiered decision, not dogma.** Fowler:
  pre-integration review "usually adds significant friction"; Ship/Show/Ask "encourages
  teams to use a blocking code review only when necessary, recognizing that post-integration
  review is often a better bet" ([Fowler CI](https://martinfowler.com/articles/continuousIntegration.html),
  [Ship/Show/Ask](https://martinfowler.com/articles/ship-show-ask.html)). Default: blocking
  review on risky surfaces (auth, payments, migrations, data integrity); Show-style
  merge-then-review is legitimate for low-risk changes. Note: DORA's finding is about
  internal-vs-external review, Fowler's about timing — orthogonal axes, both satisfied by
  "in-team, in-platform, risk-tiered blocking".

### 4. The promotion gate (dev→qa) is a release gate, not a defect-discovery or review venue

Its job: verify the **release candidate against the environment** —

- Deploy the **same immutable artifact** that already passed: "Only build packages once";
  "As an RC progresses through environments, its artifacts ideally should not be recompiled
  or rebuilt" ([Humble & Farley](https://continuousdelivery.com/implementing/patterns/),
  SWE ch. 23).
- Run environment-level verification: smoke/E2E against the deployed environment, migration
  checks, health/compliance signals, human release approval. Azure DevOps release gates are
  the clearest platform statement: every built-in gate **queries** signals (pass rate,
  incidents, alerts, policy) — none produces a test verdict
  ([Azure gates](https://learn.microsoft.com/en-us/azure/devops/pipelines/release/approvals/gates)).
- Google runs *larger/system-level* tests against the RC at each environment promotion, with
  stated purposes "as a sanity check", "for auditability", "to allow for cherry picks" — NOT
  as the primary correctness gate, which lives postsubmit on trunk (SWE ch. 23).
- Deploy the same way to every environment — promotion **rehearses** the production deploy;
  a per-environment bespoke process forfeits that rehearsal value (Humble & Farley).

**The one earned exception:** when the promoted aggregate does not equal any state trunk CI
already verified — partial promotion, cherry-picks, hotfix divergence — it earns a real
verification pass. Canonical support:
[trunkbaseddevelopment.com/branch-for-release](https://trunkbaseddevelopment.com/branch-for-release/):
QA teams "may still choose to QA work a second time on the release branch if commits between
cherry-picked changes were deliberately skipped."

### 5. Mapping: `development` is the trunk; `qa`/`production` are promotion-only refs

- `development` carries the trunk's obligations: daily integration, small-batch PRs, red
  build stops the line ("Nobody has a higher priority task than fixing the build" — Fowler;
  "All feature work stops when the main build is red" — MinimumCD), always releasable.
- `qa`/`production` follow release-branch flow rules: **no original commits, ever** — fixes
  land on `development` first and cherry-pick forward ("You should not fix bugs on the
  release branch in the expectation of cherry-picking them back to the trunk" —
  trunkbaseddevelopment.com; Fowler's branching-patterns Release Branch section agrees).
  A single original commit on `qa`/`production` degrades the topology into the
  multi-long-lived-branch model TBD declares incompatible.
- **Caveat carried honestly:** Fowler labels permanent environment branches "the classic
  example of an AntiPattern — something that looks appealing when you start, but soon leads
  to a world of misery" ([branching patterns](https://martinfowler.com/articles/branching-patterns.html));
  the harm mechanism is artifact divergence — merging between environment branches violates
  build-once, so "the code that will eventually run in production may not match 100% the
  code run during testing". This canon's mapping is **harm reduction** for an existing
  env-branch topology, not an endorsement; the cleaner end-state is artifact-based promotion
  (same binary, per-environment config), worth migrating toward when the cost is justified.

## Operational corollaries

- **Budget for flakiness structurally.** Google's data: 84% of pass→fail transitions come
  from flaky tests; ~16% of tests show some flakiness; "testing systems must be able to deal
  with a certain level of flakiness" (Micco, *State of CI Testing @Google*). A blocking gate
  whose failures are mostly false alarms trains developers to override it — keep the
  pre-merge gate to fast, reliable tests (DORA: "test failures should always indicate a real
  defect") and quarantine/deflake instead of widening retries.
- **Pair the deterministic gate with an audited escape valve.** Even Shopify's
  no-direct-merge setup keeps a documented emergency bypass — an explicit, logged exception
  beats pretending the gate is absolute.
- **Small batches apply doubly to AI-generated code.** DORA states it directly: "While AI
  excels at generating large blocks of code quickly, large changes are difficult to review,
  test, and integrate safely" ([small batches](https://dora.dev/capabilities/working-in-small-batches/)).

## Corrections the cross-exam forced (do not re-propagate these)

- ~~"The full correctness suite re-runs at promotion"~~ → Google runs larger/system tests
  against the RC for sanity/audit; the correctness suite lives postsubmit on trunk.
- ~~"GitHub merges 2,500 PRs/day through its merge queue"~~ → 2,500 PRs **per month**
  (500+ engineers, monorepo).
- ~~"DORA resolved Fowler's review-timing tension"~~ → orthogonal axes (internal-vs-external
  vs pre-vs-post-integration); reconciled by risk-tiered blocking, not by one superseding
  the other.
- ~~"A promotion PR is by construction past 400 LOC"~~ → typically, not necessarily.
- Vendor per-run AI-review pricing is volatile and was not vendor-verified — never cite it
  as a load-bearing argument; the batching-for-cost argument is negligible against engineer
  time anyway.

## Primary sources

[SWE at Google ch. 23](https://abseil.io/resources/swe-book/html/ch23.html) ·
[DORA CI](https://dora.dev/capabilities/continuous-integration/) ·
[DORA TBD](https://dora.dev/capabilities/trunk-based-development/) ·
[DORA small batches](https://dora.dev/capabilities/working-in-small-batches/) ·
[DORA test automation](https://dora.dev/capabilities/test-automation/) ·
[DORA change approval](https://dora.dev/capabilities/streamlining-change-approval/) ·
[Fowler CI](https://martinfowler.com/articles/continuousIntegration.html) ·
[Fowler branching patterns](https://martinfowler.com/articles/branching-patterns.html) ·
[Ship/Show/Ask](https://martinfowler.com/articles/ship-show-ask.html) ·
[Humble & Farley patterns](https://continuousdelivery.com/implementing/patterns/) ·
[MinimumCD](https://minimumcd.org) ·
[trunkbaseddevelopment.com](https://trunkbaseddevelopment.com) ·
[Micco, State of CI Testing @Google (PDF)](https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/45880.pdf) ·
[Brun et al. ESEC/FSE 2011](https://dl.acm.org/doi/10.1145/2025113.2025139) ·
[Bacchelli & Bird ICSE 2013](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/ICSE202013-codereview.pdf) ·
[SmartBear review study](https://smartbear.com/learn/code-review/best-practices-for-peer-code-review/) ·
[Azure release gates](https://learn.microsoft.com/en-us/azure/devops/pipelines/release/approvals/gates) ·
[GitLab merge trains](https://docs.gitlab.com/ci/pipelines/merge_trains/) ·
[GitHub merge queue](https://github.blog/engineering/engineering-principles/how-github-uses-merge-queue-to-ship-hundreds-of-changes-every-day/) ·
[Shopify merge queue](https://shopify.engineering/successfully-merging-work-1000-developers)
