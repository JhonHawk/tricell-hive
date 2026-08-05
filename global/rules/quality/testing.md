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

### Coverage by change type

Outside the carve-outs above, tests are part of the implementation — never a follow-up task:
- New functions, utilities, pure logic → **unit tests**.
- New API endpoints, database operations, service interactions → **integration tests**.
- New user-facing flows spanning multiple components → **E2E** for the critical path only. Deferring a required E2E is a plan-gate decision — a user-approved plan that defers it IS the confirmation; deferred without that, the flow is reported **not-verified** at close (`development-principles.md > Fix at the Root`). Never a mid-run stop to ask, never a silent skip.

### Verifiable test gate

> Canonical here; `agent-routing.md` (who verifies) and `development-principles.md` (what violates it) reference this without restating it. The gate is a *checkable result*, not the test-first ritual — recommend writing the failing test first (it clarifies intent), require only the verifiable outcome.

Applies to every **behavior change** (non-trivial, per the carve-out above). A change clears the gate only when all three hold:
- **Fail-to-pass** — it ships with at least one test that fails *without* the change and passes *with* it. A test that passes either way proves nothing.
- **Pass-to-pass** — every previously-passing test still passes. The change does not delete, weaken, or loosen existing tests to go green.
- **Verification runs the tests** — the gate is cleared by an actual run (command + result), never by a claim that it would pass.

**Test-gaming is not clearing the gate.** Disabling, weakening, or bypassing tests to make a suite "pass" — `.skip`/`xit`, deleting assertions, `--no-verify`, editing the runner config or a `conftest.py` to force outcomes — is the agent's version of silencing a guardrail (`development-principles.md > Fix the cause, not the check`), not a solution. Fix the code until the unmodified tests pass.

### Execution Scope

> *What* to cover (above) is separate from *how much to run per iteration*: scope the run to the change; reserve the full suite for the merge boundary. Proportionality canon referenced by `CLAUDE.md > Build & Lint`.

- **The local loop runs the affected subset — selected by the tooling, not by eye:** `jest --findRelatedTests <files>` / `--changedSince`, vitest's `related` / watch mode. A guessed subset misses regressions; a dependency-graph-selected one covers the transitively-affected tests. A trivial or localized change does not earn the full suite per iteration.
- **The test/in-vivo run amortizes over the change-group, not per commit:** a cohesive group spanning several commits on a small surface runs the affected subset and the in-vivo gate ONCE at the group's close. **Exception — a commit that stands alone earns its own run:** ~15+ files, or core/risky logic even under that threshold. Deferring to group-close trades away per-commit bisect signal — acceptable for a cohesive group, not a sprawling one: split that group instead.
- **Full suite + E2E are the CI / pre-merge gate**, not a per-iteration step.
- **A green full gate is not re-earned per fix.** Once the merge-boundary gate (full suite, build, in-vivo/smoke) has passed, a follow-up fix re-runs its blast radius — the affected subset plus the specific check that failed — never the whole gate again. The full gate repeats only when the fix touches the build graph or a cross-cutting surface (config, deps, shared runtime).
- **Safeguard — the subset is only safe with a downstream gate.** Valid *only* if CI runs the full suite before merge. **Absent CI, run the full suite before marking the task done — nothing else will.** "I ran the related tests" never satisfies the verifiable test gate's pass-to-pass half, which *is* the full existing suite staying green.
- **Anti-pattern: fixtures coupled to the full run.** If one spec in isolation forces the whole suite to boot (DB prep living only inside the full gate), that coupling *is* the defect — make prep runnable once, independent of the runner, instead of defaulting to the full suite.
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
