# Workspace hygiene playbook — audit and repair a workspace

Documentary hygiene of a client workspace: misplaced files, unpromoted decisions, stale
scratch, broken ledger pointers, specs-repo nonconformance. The `workspace-custodian`
agent performs the audit and proposes; it never executes. Risk judgment for what may be
applied automatically vs confirmed: `judgment-criteria.md`. A pre-pack project entering
the workspace convention for the first time uses `migration-playbook.md` instead — this
file is the lighter recurring pass.



Produces the actions manifest `apply` consumes. Judge with
`~/.claude/skills/flow-core/references/judgment-criteria.md`; propose, never execute.

1. OPEN per the contract (`~/.claude/skills/flow-core/SKILL.md`). If PROJECT.md or the workspace
   ambient pair (AGENTS.md canonical + CLAUDE.md importing it via `@AGENTS.md`) is missing or
   incomplete, those are the first proposals: reconstruct the ledger from the template
   (`~/.claude/skills/flow-core/references/ledger-template.md`; phase status inferred from
   observable state — specs repo, git history, the tracker), the ambient pair from the bootstrap
   bootstrap templates (`~/.claude/skills/flow-core/templates/workspace-agents.md` +
   `workspace-claude.md`). A workspace with only a CLAUDE.md (pre-pack pattern) gets an
   ambient-pair nonconformance proposal: OFFER `/agents-md-primary <path>` (it owns the
   inversion mechanics and is user-gated — never restate or auto-run them here).
2. Dispatch **workspace-custodian** (fresh context, read-only) with the workspace root, the specs
   repo path, and the intent: "your proposals will be presented verbatim to the user via this
   playbook's apply step — make every action executable as written". The custodian's fresh eyes are
   the point; don't audit from the main thread.
3. Save the **actions manifest** → `<project>/_support/workspace/hygiene-audit-<YYYY-MM-DD>.md`:
   one simple list, a row per proposed action — `action · path · why (one line) · risk
   (safe/destructive)`. This is the machine-readable source `apply` consumes; keep it free of
   prose. Give each row a short stable ID (any scheme) so `apply` can reference it days later in a
   session with no memory of this one.
4. When the audit is substantial (per `communication-format.md`), ALSO render a review report via
   the flow-report skill to `…/hygiene-audit-<YYYY-MM-DD>.html` and open it (`open <path>` on
   macOS); a small audit needs only the chat summary.
5. In chat: action counts by risk + the ID list and the manifest path. Clean → say so, skip the
   artifacts.


Consumes the manifest `audit` produced. Nothing executes without explicit approval.

1. Load the most recent **actions manifest** (`hygiene-audit-<date>.md`) — never re-read the HTML
   into context; it exists for the human. No manifest or it's stale → run `audit` first. Audit from
   a prior session → re-open the HTML (if any) so the user reviews first.
2. Gate via the question tool, **grouped by RISK, not by action type** — the question may only
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
