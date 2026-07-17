# session-hygiene-context

**Event:** `SessionStart`, no matcher (script filters to `startup|clear`). **Non-blocking** — always exits 0.

Deterministic backstop for the session-close ritual in `global/rules/workflow/git-workflow.md > Session close`. That ritual is "standing-authorized, automatic", but most session closes are silent (terminal closed, `/clear`, context exhausted) — no close-time signal exists for the model to react to, and a description-routed skill structurally cannot catch the absence of a signal. So the ceremony runs at the next deterministic seam: session start, where this hook injects pending-hygiene **facts** and the always-on rule drives the action. It injects state, never routing instructions (same contract as `flow-phase-context`).

## Mechanics

- Fires on `startup` and `clear` only (skips `resume` — the context already has it — and `compact` — mid-task).
- Local git queries only, no fetch/network: detects the integration branch (`development` if present, else `origin/HEAD`, else `main`/`master`), then gathers two objective prune signals:
  - local branches fully merged into the integration branch (excluding long-lived branches: `development|qa|production|main|master`), capped at 8;
  - local branches whose upstream is gone (`%(upstream:track)` = `[gone]`), capped at 8.
- Nothing pending → completely silent. Otherwise emits `hookSpecificOutput.additionalContext` with the facts and a pointer to the owning rule.
- Memory-side close (Engram session summary, `/memory-sync`) is deliberately out of scope — owned by the Engram plugin and `memory-routing.md`.

## Known limitations

- No `git fetch`, so "upstream gone" reflects the last fetch; a branch deleted on the remote minutes ago won't show until something fetches. Acceptable for an advisory hook — the alternative (network at session start) costs latency everywhere.
- Merged-detection is against the local integration ref; if it is behind its upstream, a branch merged remotely may not appear yet. Same trade as above.
- A checked-out merged branch is reported even though it can't be deleted until switching away — the fact is still actionable.
- The long-lived branch list is hardcoded to the repo classes in `git-workflow.md > Branching`; a repo with a nonstandard trunk name falls back to `origin/HEAD` detection.

## Smoke test

```bash
# In a repo with a merged or gone branch (create one if needed):
echo '{"source":"startup","cwd":"'"$PWD"'"}' | bash session-hygiene-context.sh
# expect additionalContext JSON listing the branch; empty output in a clean repo

echo '{"source":"resume","cwd":"'"$PWD"'"}' | bash session-hygiene-context.sh
# expect no output (resume filtered)
```
