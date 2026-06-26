---
alwaysApply: true
---

## Debugging

> Diagnosis discipline — distinct from `development-principles.md` (how to write code) and `gap-resolution.md` (prerequisites). Corrects fix-by-guessing: changing things until the symptom disappears without knowing why.

> **Apply proportionally.** An obvious typo, a one-line error with a clear message, a known cause — fix directly. The discipline below activates when the root cause is NOT obvious: intermittent failures, multi-layer systems, errors that resist the first fix.

- **Root cause before fix.** Reproduce the failure, read the full error (not just the last line), and check what changed recently — before proposing a fix. A fix applied without a reproduction is a guess.
- **Establish authoritative ground truth before diagnosing.** When a diagnosis depends on external state — a remote git branch, the actual DB schema, deployed config, or a ledger's claim that a task is "done" — verify the live source first (`git fetch` then read `origin/<branch>`, query the DB, inspect the deployment), never a local ref, cache, or ledger entry. A multi-step diagnosis built on unverified state desyncs the moment the real state surfaces. Local refs and ledger entries are claims to verify, not ground truth (`gap-resolution.md > Divergence Between Sources`).
- **One hypothesis, one change.** Change a single variable at a time so the result tells you what actually worked. Stacking speculative fixes hides which one mattered and creates new failure modes.
- **Instrument boundaries in multi-layer systems.** When data crosses layers (UI → API → service → DB, or across services), log it in and out at each boundary, run once, and locate the failing layer from evidence — then investigate that layer. This replaces guessing layer by layer.
- **Three failed fixes = question the approach, not the hypothesis.** When the third attempted fix on the same symptom fails, stop: the pattern or architecture is now the suspect. Bring the symptom, the attempts, and the architecture question to the user before a fourth fix — don't accumulate patches on a wrong foundation (`development-principles.md > start fresh`).
- **Reproduce before claiming fixed.** A regression fix is verified by the red→green→red cycle: confirm the failing case fails, apply the fix, confirm it passes, and (where feasible) revert to confirm the test catches the regression. "It should work now" without a run is not verification (`testing.md`). After green, confirm it was *your* change that fixed it — revert just the fix and watch it fail again — to rule out a coincidental external change (Agans, Rule 9).
