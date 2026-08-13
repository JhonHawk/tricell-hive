# /flow-hygiene `audit` — the default subcommand

Produces the actions manifest `apply` consumes. Judge with
`~/.claude/skills/flow-core/references/judgment-criteria.md`; propose, never execute.

1. OPEN per the contract (`~/.claude/skills/flow-core/SKILL.md`). If PROJECT.md or the workspace
   ambient pair (AGENTS.md canonical + CLAUDE.md importing it via `@AGENTS.md`) is missing or
   incomplete, those are the first proposals: reconstruct the ledger from the template
   (`~/.claude/skills/flow-core/references/ledger-template.md`; phase status inferred from
   observable state — specs repo, git history, the tracker), the ambient pair from the flow-start
   bootstrap templates (`~/.claude/skills/flow-start/templates/workspace-agents.md` +
   `workspace-claude.md`). A workspace with only a CLAUDE.md (pre-pack pattern) gets an
   ambient-pair nonconformance proposal: OFFER `/agents-md-primary <path>` (it owns the
   inversion mechanics and is user-gated — never restate or auto-run them here).
2. Dispatch **workspace-custodian** (fresh context, read-only) with the workspace root, the specs
   repo path, and the intent: "your proposals will be presented verbatim to the user by
   flow-hygiene apply — make every action executable as written". The custodian's fresh eyes are
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
