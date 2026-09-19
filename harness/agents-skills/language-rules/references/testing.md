## Testing

> Tests are mandatory whenever a change alters behavior that has a useful automatic check. The bar is the **verifiable test gate** below — TDD makes the implementation sequence observable, while the gate keeps the result checkable.

### Does NOT apply to
- Repositories with no runtime-test target (CLAUDE.md hubs, IaC-only repos, documentation-only repos) are exempt from runtime-test requirements only; the verification matrix's review and readback gates still apply to behavior-bearing configuration. Each declares its scope in its own `CLAUDE.md > Rule Exclusions`.
- Disposable utilities under `_support/scripts/` that exist only for the current dev session.
- Passive documentation-only edits with no effect on runtime behavior, agent instructions, skill
  definitions, hooks or configuration — those use diff and consistency review, with runtime
  verification recorded as not-applicable. Agent instructions, skills, hooks and configuration
  belong to the standard-behavior review row when they can change behavior; use not-applicable
  only when no useful automatic check exists and record the reason.
- **Trivial changes**, as defined by the carve-out in `critical-thinking.md` (the canonical definition — not restated here). No new tests required, but existing tests must still pass.

### Test approach

Every new plan task declares exactly one `Test approach:` value. Direct work records the same
choice in its execution evidence. A direct workflow or a plan with `Session: no` may keep that
evidence in the current conversation, but must report it there and makes no durable cross-harness
resume claim:

| Approach | Use when | Required evidence |
|---|---|---|
| `tdd` | New behavior or a reproducible bug has a useful automatic check, including hooks, agent instructions, skills and executable configuration. | Write and run the check before the corresponding implementation, observe the expected RED failure, implement the smallest change, observe GREEN, and refactor only while the check remains green. |
| `characterization` | A pure refactor preserves behavior and has an existing or newly captured automatic check. | Establish the current behavior with a green baseline before editing, then run the affected check after the refactor. A behavior change moves the task to `tdd`. |
| `not-applicable` | Passive documentation, a purely visual change, or behavior-bearing instructions, hooks or configuration with no useful automatic check. | Record the concrete reason and use the applicable readback, consistency or visual review. This is not a silent exemption for a missing runner or test. |

`RED` means an observed, attributable failure of the intended check before the implementation. A
claim that the check would fail, a dependency or environment failure, a failed fix attempt, or a
remote operation failure is not RED. A normal RED→GREEN→refactor cycle does not require a new user
approval between phases; the approved plan covers the declared task.

If a `tdd` task's implementation already exists before RED was observed, preserve the work and
never invent retrospective evidence (moving the module aside to obtain "module not found" is
invention, not RED) or force a revert solely to recreate the sequence. The missing chronology is
agent-executable, never a question for the user: produce an isolated fail-to-pass comparison —
break what the check guards (reintroduce the defect, invert the branch) or run the check against
a pre-change baseline copy, observe the failure, restore, verify no residue — keep pass-to-pass
and independent checks, and label completion `fail-to-pass`, never strict TDD. Ask for a user
exception only when no such comparison can be produced (the behavior cannot be broken in
isolation and no baseline exists); until it is accepted the task cannot be reported complete,
advance to `built` or `verified`, or be published, and its completion is labeled
`exception-accepted`. RED is owed once per behavior, at the seam the plan names: a second check
over behavior already green (an integration spec after the unit spec that drove it) cannot RED
and is not missing one — declare it a second-layer check and evidence it by fail-to-pass. A
planned `not-applicable` classification is resolved at the plan gate.

### Coverage by change type

Outside the carve-outs above, tests are part of the implementation — never a follow-up task:
- New functions, utilities, pure logic → **unit tests**.
- New API endpoints, database operations, service interactions → **integration tests**.
- New user-facing flows spanning multiple components → **E2E** for the critical path only. Deferring a required E2E is a plan-gate decision — a user-approved plan that defers it IS the confirmation; deferred without that, the flow is reported **not-verified** at close (`reporting-integrity.md > Fix at the Root`). Never a mid-run stop to ask, never a silent skip.

### Verifiable test gate

> Canonical here; `agent-routing.md` (who verifies), `development-principles.md` (what violates it), and the flow references point here without restating the TDD mechanics.

For `tdd`, strict TDD clears the gate only when all three hold:
- **Observed RED-to-GREEN** — the intended check failed before the implementation and passes with it. A test that passes either way proves nothing.
- **Pass-to-pass** — every previously-passing test still passes. The change does not delete, weaken, or loosen existing tests to go green.
- **Verification runs the checks** — RED, GREEN and any required refactor check are actual runs with commands and results, never claims that they would pass.

**Two failure modes that read green through all three** — a passing suite and a rising test count are not coverage, and neither review nor CI separates them from the real thing:
- **A retired test's coverage is compared by FIXTURE, not by assertion.** Replacing a test means putting the old and new fixtures side by side and naming what precondition the substitute needed in order to pass. A new precondition in the fixture is a new condition in production: what the original covered unconditionally is now covered only under it.
- **A check counts as protection only once it has been seen to fail.** Before a new or repaired check is claimed to guard something, break what it guards — invert the precedence, revert the normalization, reintroduce the defect — confirm red, restore, and verify no residue. An assertion whose fixture cannot produce the failure its name claims is vacuous.

Prompt-convention; a mutation-testing runner is the deterministic backstop where one is configured.

For `fail-to-pass` completions the gate substitutes the observed fail-to-pass comparison for the
chronological RED (a safe isolated pre-change baseline when the working tree already contains the
implementation) and keeps pass-to-pass evidence and the applicable independent review/checks. A
user-accepted exception (`exception-accepted`) waives chronology only and is reached solely when
that comparison could not be produced; neither is strict TDD.

For `characterization`, the gate requires an observed green baseline before the refactor and a
green affected check afterward, plus pass-to-pass evidence. For `not-applicable`, the gate
requires the recorded reason and the applicable readback, consistency or visual review; no
runtime test is claimed.

**Test-gaming is not clearing the gate.** Disabling, weakening, or bypassing tests to make a suite "pass" — `.skip`/`xit`, deleting assertions, `--no-verify`, editing the runner config or a `conftest.py` to force outcomes — is the agent's version of silencing a guardrail (`development-principles.md > Fix the cause, not the check`), not a solution. Fix the code until the unmodified tests pass.

### Execution Scope

> *What* to cover (above) is separate from *how much to run per iteration*: the TDD or characterization loop is part of implementation, while boundary verification is scoped to the change and the full suite is reserved for the merge boundary — the change-group's close seam on the agent's side (build, in-vivo/smoke; the agent's own full-suite run is needed there only absent the covering CI — CI placement per the CI-gate bullet below). Proportionality canon referenced by `CLAUDE.md > Build & Lint`.

- **The local loop runs the affected subset — selected by the tooling, not by eye:** `jest --findRelatedTests <files>` / `--changedSince`, vitest's `related` / watch mode. A guessed subset misses regressions; a dependency-graph-selected one covers the transitively-affected tests. A trivial or localized change does not earn the full suite per iteration.
- **The boundary test/in-vivo run amortizes over the change-group, not per commit:** a cohesive group spanning several commits on a small surface runs the affected subset and the in-vivo gate ONCE at the group's close. **Exception — a commit that stands alone earns its own run:** ~15+ files, or core/risky logic even under that threshold. This does not suppress the RED, GREEN or refactor runs required by the declared test approach.
- **The local gate closes BEFORE publishing.** Commit, push, PR, and merge are downstream of the change-group's local verification (affected subset, typecheck/build, the in-vivo/smoke pass), never concurrent with it. **Carve-out:** a local commit another step consumes (a SHA to vendor, a bisect point) may precede the gate; nothing leaves the machine until it is green. Prompt-convention; the `bash-policy` pre-push advisory reminds, it never blocks.
- **A verification waiting on the user blocks the publish, not just the check.** Whatever the run needs and only the user can grant — credentials, a confirmation, seed data — is asked BEFORE publishing, and the publish waits for the answer; advancing to push/PR/merge with the ask still open turns a blocking gate into a post-hoc one. Genuinely infeasible locally → name it and state the uncovered surface at close: "I'll run it after the push" is not that case.
- **Missing data or an unstarted stack is a resolvable blocker, never an infeasibility.** The role the dataset lacks → seed it; backend and frontend not up → start both (no commit required); a path that sends real email/SMS/webhooks → route it to a disposable channel (`+tag` alias, throwaway inbox, local mail catcher) instead of skipping the path. **An isolated local environment exists to be dirtied** — writing test data there IS the verification, not damage. Only a SHARED environment is protected, and the answer there is standing up the isolated one, not leaving the path unexercised.
- **Full suite + E2E are the CI gate — post-merge on the integration branch in env-branch/promotion topologies** (async; a merge queue collapses it pre-merge where volume warrants). Where the trunk deploys production directly — platform-native repos — the full gate stays pre-merge on the PR. Never a per-iteration step. A promotion between environments is a release gate (deploy verification), never a correctness re-run.
- **A green full gate is not re-earned per fix.** Once the merge-boundary gate (full suite, build, in-vivo/smoke) has passed, a follow-up fix re-runs its blast radius — the affected subset plus the specific check that failed — never the whole gate again. The full gate repeats only when the fix touches the build graph or a cross-cutting surface (config, deps, shared runtime).
- **Each boundary verification layer runs at most once per change-group, at its boundary.** The affected subset may run during the declared RED, GREEN and refactor phases because those runs establish the implementation evidence; the full gate runs at group close, with one in-vivo pass, one independent review and one refuter pass per finding. The PR bug-hunt pass (`git-mechanics.md > PRs & promotion`, Phase B) is a release gate of the merge — analogous to CI, not counted against the one independent review. Verification effort scales with the new information a run produces, not with the confidence desired.
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
- For `tdd`, write or update the check before the corresponding behavior and fix implementation to match the spec, not the other way around; fix tests only when the spec changes.
- For `characterization`, capture the existing behavior before a pure refactor and preserve it afterward. A new behavior requires `tdd`.
