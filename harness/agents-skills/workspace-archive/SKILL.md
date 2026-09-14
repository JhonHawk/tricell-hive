---
name: workspace-archive
description: >
  Archive closed sessions out of a specs repo's `sessions/` so only active and recently
  closed work stays visible: `verified` plans older than 15 days move to `sessions/archived/`
  by `git mv`, index links rewritten. On request only (roughly monthly); never on a schedule
  or at session start. Subcommands: `run` (default) | `normalize` (one-time legacy Status
  cleanup). Triggers: "archiva las sesiones", "limpia sessions/", "workspace-archive".
disable-model-invocation: true
---

# /workspace-archive — keep `sessions/` navigable

Convention (`flow-core/references/specs-structure.md > Session & initiative conventions`):
`sessions/` holds what is active plus what closed within the last 15 days; everything closed
earlier lives one level down in `sessions/archived/<same-folder>/`. Closed means the plan's
`Status:` is `verified` — the only terminal state of `plan-format.md` — and the age is the
folder's last commit. A `planned`, `building` or `built` plan never archives by age: an old
one is a hygiene finding for the `workspace-hygiene-playbook`, not archive material.

The work is deterministic and lives in `scripts/sessions-archive.sh`; this skill only runs
it, shows you the manifest, and applies on your explicit ok. It spends no model reasoning on
the classification.

## `run` (default)

1. `bash $HOME/.agents/skills/workspace-archive/scripts/sessions-archive.sh [--days N] [--sessions <dir>]` — dry-run.
   Resolves the specs repo from the workspace ledger (`_support/PROJECT.md` above cwd) or from
   `./sessions/`; no `--sessions` guessing.
2. Present its blocks as they come, in the conversation language: archivables (folder, state,
   days since close, index signal), stale-not-archivable (`planned`/`building`/`built` past the
   threshold), needs-review (legacy `Status:` values, no-plan folders without an index close
   signal, an old `draft`), and the recent/active counts. Empty blocks are stated as empty. A
   `needs-review` entry with a legacy `Status:` means `normalize` has not run for this
   workspace — say so and offer it; do not classify by hand. Age is the folder's last commit
   (normalization commits excluded); `(mtime)` marks a folder git could not date.
3. Ask ONE question: apply the archivable block and commit (`chore(sessions): archive N closed
   session(s) older than 15 days`) — yes / no. Push follows the specs repo's declared workflow
   (direct-to-trunk repos push in the same step; otherwise the commit stays local and you say so).
4. On yes: re-run with `--apply` (it `git mv`s, rewrites `sessions/README.md` links, stages),
   then commit. On no: stop; the manifest is the deliverable.

Never touch a `stale` or `needs-review` folder here; never delete anything; never invoke
`--apply` without the yes.

## `normalize` (once per workspace)

Plans written before the portable format carry prose in `Status:` (`verified in QA`,
`completed-production`, `done (Fase 1)`). `run` reports them as needs-review until they are
canonical.

1. `bash $HOME/.agents/skills/workspace-archive/scripts/sessions-archive.sh --normalize [--sessions <dir>]` — dry-run:
   the deterministic map resolves most (`verified|done|completed|shipped|concluded|finaliz` →
   `verified`, `in-progress|building` → `building`); anything mentioning a phase, a partial
   or a deferral stays unresolved, and a `Status:` found inside the contract block is reported
   untouched.
2. Unresolved entries and no-plan folders with no index signal go to the `workspace-custodian`
   (read-only) in ONE dispatch with the list and the index rows; it returns a proposed state per
   folder with the evidence line (index row, last commit, ledger). Present the full manifest —
   mapped and proposed — and ask ONE question to apply it.
3. On yes: `--normalize --apply` for the mapped set (the original line is kept as a comment
   under the new `Status:`), the custodian's proposals applied by hand the same way, one commit
   with the line the script suggests (`chore(sessions): normalize N legacy plan status line(s)`
   — keep that prefix: the archiver ignores those commits when dating a folder). Changing
   `Status:` never touches a frozen contract: the line lives outside the contract block. Then
   run `run`.

## What this skill is not

- Not a scheduler: no SessionStart line, no standing job. The user invokes it.
- Not the hygiene audit: misplaced files, unpromoted decisions and broken ledger pointers are
  the `workspace-hygiene-playbook` with the `workspace-custodian`.
- Not a deleter: `_support/workspace/` scratch is reported by the script (age, concluded
  work), never removed here — deletion is confirm-gated (`CLAUDE.md > Destructive Operations`).

Stable path after deploy: `~/.agents/skills/workspace-archive/SKILL.md`; the script at
`~/.agents/skills/workspace-archive/scripts/sessions-archive.sh` (Codex/opencode/PI: the same
under `~/.agents/skills/`).
