# /flow-adopt — bring an existing project into the flow

The opt-in path for bringing an existing project (current OR past) into the flow — the broader
sibling of `/flow-workspace` `audit` + `apply`, adding specs-repo creation and active/archived
tiering. Mechanics live in `~/.claude/skills/flow-core/references/migration-playbook.md` —
**read it first; it is the source of truth.** Judge artifacts with
`~/.claude/skills/flow-core/references/judgment-criteria.md` (archive on doubt;
convention suggested, never enforced). Work by its principles, not a rigid checklist:

1. OPEN per the contract (`~/.claude/skills/flow-core/SKILL.md`). Detect **shape**
   (workspace/standalone; `<project>-specs/` present?) and
   **activity tier** by judgment — ACTIVE = full adoption; ARCHIVED/inactive = minimal structure +
   sweep loose docs into `sessions/previously/` (dated by mtime/git), no archaeology. Unsure → ask;
   default ARCHIVED.
2. Dispatch **workspace-custodian** (fresh, read-only) seeded with the playbook and the project
   root to produce a **migration manifest** (same simple shape: `action · source · destination ·
   why · risk`), covering specs-repo creation if needed, per-artifact routing, and index /
   back-reference creation. Save it → `<project>/_support/workspace/migration-<YYYY-MM-DD>.md`;
   render an HTML review via flow-report when substantial and open it. Already-conforming → say so
   and skip.
3. Gate via AskUserQuestion grouped by RISK: specs-repo creation, deletes, and reference-breaking
   moves per item; safe moves/renames batch. Deletions run only after typed `eliminar` + IDs.
4. Execute only what was approved: `git init <project>-specs` (if approved) → skeleton the ledger,
   the ambient pair, and the specs-repo structure guided by judgment (not a fixed bootstrap
   sequence); `git mv` inside a repo / `mv` across the `_support/`→specs boundary; write promoted
   files, indexes, and back-references; archive (never delete) on doubt; preserve content verbatim
   and the raw/curated split. **Ambient refresh:** a workspace with only a CLAUDE.md (pre-pack
   pattern) gets the inversion OFFERED as `/agents-md-primary <path>` (it owns the mechanics and
   is user-gated — never restate or auto-run them here); otherwise, if the workspace AGENTS.md
   predates the flow-start template
   (`~/.claude/skills/flow-start/templates/workspace-agents.md`), rewrite that
   block to it (a coherent revision, not a string swap) and add
   the `sessions/**` pre-authorization to `.claude/settings.json` if absent
   (`~/.claude/skills/flow-start/templates/sessions-permissions.json`).
5. CLOSE per the contract: update the ledger pointer (it does NOT index sessions — that stays in
   the co-located `sessions/README.md`), report executed vs skipped, and — if a specs repo was
   created — make its initial commit (this invocation authorizes it; no remote unless the user adds
   one). Re-running `/flow-adopt` is safe: already-conforming artifacts produce no findings.
