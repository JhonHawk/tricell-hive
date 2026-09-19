## Test Gate

> How a change to behavior is tested, and what evidence the report has to carry.

### Test approach — one per task
- Use the approach the task declares; declare one yourself only when it names none, and never downgrade a declared `tdd`.
- Trivial change (typo, rename, one-line config, formatting): no new test; existing tests still pass.
- `tdd` — new behavior, or a reproducible bug with a useful automatic check: write the check, RUN it, observe RED, implement the smallest change, run it again for GREEN, refactor only while it stays green.
- `characterization` — a pure refactor that preserves behavior: run the existing (or newly captured) check green BEFORE editing, then again after. Behavior changed → the task is `tdd`.
- `not-applicable` — passive documentation, a purely visual change, or instructions, hooks, or configuration with no useful automatic check: record the concrete reason. Never for a missing runner, an unstarted stack, or a test that is merely inconvenient.
- Tests ship with the implementation, never as a follow-up task.

### RED
- RED is an OBSERVED, attributable failure of the intended check before the implementation exists — paste the command and the failing line into the report.
- A dependency error, an environment failure, a typo in the spec file, a failed fix attempt, or "it would fail" is not RED.
- Implementation already written before RED was observed → say so and replace the chronology with an isolated fail-to-pass comparison: break what the check guards (reintroduce the defect, invert the branch), observe the failure, restore, confirm no residue — or run the check against a pre-change baseline copy — and label the result `fail-to-pass`. Neither possible → report the task not-complete, exception needed.
- Never reconstruct a RED you did not see, never move a module aside to manufacture "module not found", never revert working code to replay the sequence.
- A new or repaired check counts as protection only once it has been SEEN to fail with the defect present; a fixture that cannot produce the failure its name claims is vacuous.
- RED is owed once per behavior: a second-layer check over already-green behavior (an integration spec after the unit spec that drove it) is declared as such and evidenced by fail-to-pass.

### Pass-to-pass
- Every previously passing test still passes, and that run is reported too.
- Never delete, weaken, or loosen an existing test to go green: no `.skip`/`xit`, no removed assertions, no `--no-verify`, no edits to the runner config, `conftest.py`, or an exclude list to change an outcome. Fix the code until the unmodified test passes.
- Replacing a test → put old and new fixtures side by side and name any precondition the substitute needs that the original did not; a new precondition in the fixture narrows what production is covered against.

### What to run
- Run the affected subset the tooling selects — `vitest related <files>`, `jest --findRelatedTests <files>`, `turbo --affected` — never a hand-picked subset, never the full suite per iteration, and the full suite only when the task asks for it or nothing downstream will run it.
- No graph-aware selector (Gradle, pytest) → run the touched module or package plus its direct dependents, and say the selection was manual.

### Coverage by change type
- New function, utility, or pure logic → unit test; new endpoint, database operation, or service interaction → integration test; new user-facing flow across components → E2E for the critical path only.
- Cover the happy path AND the error paths: null/undefined, empty collections, boundary values, special characters, async failures (timeouts, network errors, races).
- Test behavior, not implementation details; mock external dependencies, never internal logic; keep tests independent of each other. A bug a higher-level test caught alone gets the missing lower-level test added.

### Handoff
- Report per task: the declared approach, the test file paths, the exact command, and the actual RED and GREEN output lines — the lines themselves, not a summary.
- A check that never ran is reported as not-verified, never as passing.
