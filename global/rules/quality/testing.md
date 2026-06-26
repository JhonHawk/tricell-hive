---
alwaysApply: true
---

## Testing

> Tests are mandatory whenever a change alters behavior. The bar is the **verifiable test gate** below — not a ritual, a checkable result.

### Does NOT apply to
- Config-only repos with no runtime (CLAUDE.md hubs, IaC-only repos, documentation-only repos). Each declares its scope in its own `CLAUDE.md > Rule Exclusions`.
- Disposable utilities under `_support/scripts/` that exist only for the current dev session.
- Edits to agent prompts, skill definitions, or markdown rule files — those are reviewed by reading, not by running tests.
- **Trivial changes**, as defined by the carve-out in `quality/critical-thinking.md` (the canonical definition — not restated here). No new tests required, but existing tests must still pass.

### Test Types
- **Unit tests**: isolated functions, utilities, pure logic. Always required for new code.
- **Integration tests**: API endpoints, database operations, service interactions. Required for backend features.
- **E2E tests**: critical user flows only. Required for user-facing features that span multiple components.

### Enforcement
Before considering any implementation complete, verify:
- New functions/utilities → unit tests exist and pass
- New API endpoints → integration tests exist and pass
- New user-facing flows → E2E tests exist and pass (or explicitly deferred with user confirmation)

If a change affects behavior, tests are part of the implementation — not a follow-up task.

### Verifiable test gate

> Canonical here; `agent-routing.md` (who verifies) and `development-principles.md` (what violates it) reference this without restating it. The gate is a *checkable result*, not the test-first ritual — recommend writing the failing test first (it clarifies intent), require only the verifiable outcome (the ritual without it is neutral at best).

Applies to every **behavior change** (non-trivial, per the `critical-thinking.md` trivial carve-out). A change clears the gate only when all three hold:
- **Fail-to-pass** — it ships with at least one test that fails *without* the change and passes *with* it. A test that passes either way proves nothing.
- **Pass-to-pass** — every previously-passing test still passes. The change does not delete, weaken, or loosen existing tests to go green.
- **Verification runs the tests** — the gate is cleared by an actual run (command + result), never by a claim that it would pass.

**Test-gaming is not clearing the gate.** Disabling, weakening, or bypassing tests to make a suite "pass" — `.skip`/`xit`, deleting assertions, `--no-verify`, editing the runner config or a `conftest.py` to force outcomes — is the agent's version of silencing a guardrail (`development-principles.md > Fix the cause, not the check`), not a solution. Fix the code until the unmodified tests pass.

### Execution Scope

> *What* to cover (above) is separate from *how much to run per iteration*. Modern test selection makes the affected subset cheap and fast — so the frequent runs a test-first loop needs are an enabler, not a bottleneck. Scope the run to the change; reserve the full suite for the merge boundary. This section is the proportionality canon referenced by `CLAUDE.md > Build & Lint`; "trivial" is defined in `quality/critical-thinking.md`.

- **Local loop runs the affected subset — selected by the tooling, not by eye.** Let the runner compute the impact graph: `jest --findRelatedTests <files>` / `--changedSince`, vitest's `related` / watch mode, test sharding. The subset an agent *guesses* is the source of missed regressions; the subset a dependency-graph tool selects covers the transitively-affected tests, so it is safe to trust. A trivial or localized change does not earn the full suite on every iteration.
- **The test/in-vivo run amortizes over the change-group, not per commit.** When one cohesive change-group spans several commits (e.g. issues split that way) but touches a small surface, run the affected subset and the in-vivo gate ONCE at the group's close — not N× per commit. Running 5 commits × the same ~10-file surface is 5× the cost for the same guarantee. **Exception — a commit large enough to stand alone earns its own run:** a single commit touching ~15+ files, or core/risky logic, gets the affected subset run at that commit, not deferred to the group's close. File count is a proxy — a commit that changes critical logic counts as "stand-alone" even under the threshold. Trade-off: deferring to group-close loses per-commit signal on which commit broke a test; acceptable for a cohesive group (cheap bisect), not for a sprawling one — split the group instead. This sets the *boundary* the affected-subset runs at; the full suite + E2E still run at the merge boundary below.
- **Full suite + E2E are the CI / pre-merge gate**, not a per-iteration step — run once before merge or deploy, where their cost buys the cross-cutting guarantee a subset can't.
- **Safeguard — the subset is only safe with a downstream gate.** Valid *only* if CI runs the full suite before merge. **Absent CI, run the full suite before marking the task done — nothing else will.** "I ran the related tests" never substitutes for coverage when no gate sits behind it — and it never satisfies the **verifiable test gate** above, whose pass-to-pass half *is* the full existing suite staying green.
- **Anti-pattern: fixtures coupled to the full run.** If running one spec in isolation forces the whole suite to boot — because DB prep (`migrate deploy` + seed) lives only inside the full gate — that coupling *is* the defect. Make prep runnable once, independent of the runner; until then the subset rule is inert, so fix the prep rather than defaulting to the full suite.
- **Serialize only what shares state.** Restrict `--runInBand` (or any global serialization) to suites that truly share mutable state; let the rest parallelize.

### What to Test
- Happy path AND error paths. A test suite that only covers the happy path is incomplete.
- Edge cases: null/undefined input, empty arrays/strings, boundary values, special characters.
- Async error handling: network failures, timeouts, race conditions.
- Push coverage down the pyramid: when a higher-level test catches a bug and no lower-level test failed, add the missing lower-level test — don't re-assert the same behavior at every layer.

### Anti-patterns to Avoid
- Testing implementation details (internal state) instead of behavior.
- Tests that depend on each other or share mutable state.
- Asserting too little — a test that always passes verifies nothing.
- Mocking everything — mock external dependencies, not internal logic.
- Treating green CI as proof of safety. Passing tests verify the code matches the spec — not that the spec is correct, that assumptions hold at scale, or that implicit infrastructure constraints are respected.

### When to Update Tests
- If a change affects behavior, update or add tests — even if not explicitly requested.
- Fix implementation to match the spec, not the other way around. Only fix tests when the spec changes.
