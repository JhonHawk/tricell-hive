
## Testing

> Tests are mandatory whenever a change alters behavior. The bar is the **verifiable test gate** below — not a ritual, a checkable result.

### Does NOT apply to
- Config-only repos with no runtime (CLAUDE.md hubs, IaC-only repos, documentation-only repos). Each declares its scope in its own `CLAUDE.md > Rule Exclusions`.
- Disposable utilities under `_support/scripts/` that exist only for the current dev session.
- Edits to agent prompts, skill definitions, or markdown rule files — those are reviewed by reading, not by running tests.
- **Trivial changes**, as defined by the carve-out in `quality/critical-thinking.md` (the canonical definition — not restated here). No new tests required, but existing tests must still pass.

### Coverage by change type

Outside the carve-outs above, tests are part of the implementation — never a follow-up task:
- New functions, utilities, pure logic → **unit tests**.
- New API endpoints, database operations, service interactions → **integration tests**.
- New user-facing flows spanning multiple components → **E2E** for the critical path only. Deferring a required E2E is a plan-gate decision — a user-approved plan that defers it IS the confirmation; deferred without that, the flow is reported **not-verified** at close (`quality/reporting-integrity.md > Fix at the Root`). Never a mid-run stop to ask, never a silent skip.

### Verifiable test gate

> Canonical here; `agent-routing.md` (who verifies) and `development-principles.md` (what violates it) reference this without restating it. The gate is a *checkable result*, not the test-first ritual — recommend writing the failing test first (it clarifies intent), require only the verifiable outcome.

Applies to every **behavior change** (non-trivial, per the carve-out above). A change clears the gate only when all three hold:
- **Fail-to-pass** — it ships with at least one test that fails *without* the change and passes *with* it. A test that passes either way proves nothing.
- **Pass-to-pass** — every previously-passing test still passes. The change does not delete, weaken, or loosen existing tests to go green.
- **Verification runs the tests** — the gate is cleared by an actual run (command + result), never by a claim that it would pass.

**Test-gaming is not clearing the gate.** Disabling, weakening, or bypassing tests to make a suite "pass" — `.skip`/`xit`, deleting assertions, `--no-verify`, editing the runner config or a `conftest.py` to force outcomes — is the agent's version of silencing a guardrail (`development-principles.md > Fix the cause, not the check`), not a solution. Fix the code until the unmodified tests pass.

### Execution Scope

> *What* to cover (above) is separate from *how much to run per iteration*: scope the run to the change; reserve the full suite for the merge boundary — the change-group's close seam on the agent's side (build, in-vivo/smoke; the agent's own full-suite run is needed there only absent the covering CI — CI placement per the CI-gate bullet below). Proportionality canon referenced by `CLAUDE.md > Build & Lint`.

- **The local loop runs the affected subset — selected by the tooling, not by eye:** `jest --findRelatedTests <files>` / `--changedSince`, vitest's `related` / watch mode. A guessed subset misses regressions; a dependency-graph-selected one covers the transitively-affected tests. A trivial or localized change does not earn the full suite per iteration.
- **The test/in-vivo run amortizes over the change-group, not per commit:** a cohesive group spanning several commits on a small surface runs the affected subset and the in-vivo gate ONCE at the group's close. **Exception — a commit that stands alone earns its own run:** ~15+ files, or core/risky logic even under that threshold. Deferring to group-close trades away per-commit bisect signal — acceptable for a cohesive group, not a sprawling one: split that group instead.
- **The local gate closes BEFORE publishing.** Commit, push, PR, and merge are downstream of the change-group's local verification (affected subset, typecheck/build, the in-vivo/smoke pass), never concurrent with it. **Carve-out:** a local commit another step consumes (a SHA to vendor, a bisect point) may precede the gate; nothing leaves the machine until it is green. Prompt-convention; the `bash-policy` pre-push advisory reminds, it never blocks.
- **A verification waiting on the user blocks the publish, not just the check.** Whatever the run needs and only the user can grant — credentials, a confirmation, seed data — is asked BEFORE publishing, and the publish waits for the answer; advancing to push/PR/merge with the ask still open turns a blocking gate into a post-hoc one. Genuinely infeasible locally → name it and state the uncovered surface at close: "I'll run it after the push" is not that case.
- **Missing data or an unstarted stack is a resolvable blocker, never an infeasibility.** The role the dataset lacks → seed it; backend and frontend not up → start both (no commit required); a path that sends real email/SMS/webhooks → route it to a disposable channel (`+tag` alias, throwaway inbox, local mail catcher) instead of skipping the path. **An isolated local environment exists to be dirtied** — writing test data there IS the verification, not damage. Only a SHARED environment is protected, and the answer there is standing up the isolated one, not leaving the path unexercised.
- **Full suite + E2E are the CI gate — post-merge on the integration branch in env-branch/promotion topologies** (async; a merge queue collapses it pre-merge where volume warrants). Where the trunk deploys production directly — platform-native repos — the full gate stays pre-merge on the PR. Never a per-iteration step. A promotion between environments is a release gate (deploy verification), never a correctness re-run.
- **A green full gate is not re-earned per fix.** Once the merge-boundary gate (full suite, build, in-vivo/smoke) has passed, a follow-up fix re-runs its blast radius — the affected subset plus the specific check that failed — never the whole gate again. The full gate repeats only when the fix touches the build graph or a cross-cutting surface (config, deps, shared runtime).
- **Each verification layer runs at most once per change-group, at its boundary.** Subset per iteration, full gate at group close, one in-vivo pass, one independent review, one refuter pass per finding — the layers stack across the group, they do not repeat within it. The PR bug-hunt pass (`git-mechanics.md > PRs & promotion`, Phase B) is a release gate of the merge — analogous to CI, not counted against the one independent review. Verification effort scales with the new information a run produces, not with the confidence desired.
- **Re-running a green check requires a named nondeterminism trigger** (observed flake, race, external state) — never "to accumulate confidence". The repetition scopes to the flaky unit (the package, the spec), never the full gate; the full gate then runs ONCE at the end to close. Deterministic enforcement: the `post-tool-hub` hook flags repeated full-suite runs in a session — it matches the `pnpm`/`turbo` invocation shapes only, so a directly-invoked runner (`npx vitest`, `pytest`, `gradlew test`) is prompt-convention, not backstopped.
- **The stack scales by attendance mode.** Attended session (the user is present to validate): optimize for time-to-feedback — deliver the working increment after the affected subset + typecheck (the fail-to-pass test still ships with the change) and let the user's validation drive the next iteration; the full gate waits for the merge boundary. Unattended/delegated run: the agent is the only validator — full rigor before reporting done.
- **Safeguard — the subset is only safe with a downstream gate.** Valid *only* if CI runs the full suite — pre-merge on the PR or merge queue, or post-merge on every push to the integration branch (a red run is fixed or reverted immediately). **Absent that CI, run the full suite before marking the task done — nothing else will.** "I ran the related tests" never satisfies the verifiable test gate's pass-to-pass half, which *is* the full existing suite staying green.
- **Anti-pattern: fixtures coupled to the full run.** If one spec in isolation forces the whole suite to boot (DB prep living only inside the full gate), that coupling *is* the defect — make prep runnable once, independent of the runner, instead of defaulting to the full suite.
- **A diff touching a globally-registered pipeline component widens the in-vivo gate to every principal type.** Guard, interceptor, middleware, filter, or `common/**` consumed app-wide: the walk exercises one authenticated request per PRINCIPAL TYPE that traverses the component — never only the ticket's flow. The tooling-selected subset cannot see this class (the other principals live in untouched modules), and diff-scoped review cannot either; this is `critical-thinking.md > Count the instances` made an observable act (the edit to the global component is the trigger). Prompt-convention; the deterministic backstop is the project's e2e token×route matrix where one exists.
- **Independent checks dispatch in parallel.** Lint, typecheck, and the affected test subset share no state — run them concurrently and read the results together; serialize only real dependencies (build before start, migrate before seed).
- **Serialize only what shares state.** Restrict `--runInBand` (or any global serialization) to suites that truly share mutable state; let the rest parallelize.

### What to Test
- Happy path AND error paths; edge cases (null/undefined, empty arrays/strings, boundary values, special characters); async failures (network errors, timeouts, race conditions).
- Push coverage down the pyramid: when a higher-level test catches a bug and no lower-level test failed, add the missing lower-level test — don't re-assert the same behavior at every layer.

### Anti-patterns to Avoid
- Testing implementation details instead of behavior; tests that depend on each other or share mutable state; asserting too little; mocking everything — mock external dependencies, not internal logic.
- Treating green CI as proof of safety: passing tests verify the code matches the spec — not that the spec is correct, that assumptions hold at scale, or that implicit infrastructure constraints are respected.

### When to Update Tests
- If a change affects behavior, update or add tests — even if not explicitly requested. Fix implementation to match the spec, not the other way around; fix tests only when the spec changes.
