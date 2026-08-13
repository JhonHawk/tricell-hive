# /flow-workspace `apply` — execute the approved actions

Consumes the manifest `audit` produced. Nothing executes without explicit approval.

1. Load the most recent **actions manifest** (`hygiene-audit-<date>.md`) — never re-read the HTML
   into context; it exists for the human. No manifest or it's stale → run `audit` first. Audit from
   a prior session → re-open the HTML (if any) so the user reviews first.
2. Gate via AskUserQuestion, **grouped by RISK, not by action type** — the question may only
   reference IDs the user just saw. Approve everything `safe` in one option; `destructive` actions
   (expirations, moves that break references, legacy migrations) are reviewed per item or small
   batch, and deletions execute only after **typed confirmation** — the user writes the literal
   word `eliminar` plus the IDs (a click is not enough to destroy files; hesitation means archive
   instead). Nothing executes without explicit approval.
3. Execute only what was approved:
   - `move`: `git mv` inside a repo, plain `mv` otherwise
   - `promote`: write the summary into `<project>-specs/decisions/` (dated file), then point the
     source's ledger row at it
   - `expire`: delete only with typed approval; on hesitation, archive instead
   - `repair`: edit PROJECT.md rows as proposed
4. CLOSE per the contract (`~/.claude/skills/flow-core/SKILL.md`): update PROJECT.md (pruning rows
   the actions made stale), report executed
   vs skipped, and delete the consumed audit artifacts. If actions were skipped, keep them and note
   in the manifest which IDs remain open.
