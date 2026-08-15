# /flow-start — Stage B: Bootstrap (workspace, ledger, memory, conventions)

Stands up the workspace tree, the ledger, the ambient pair, and the memory/conventions layer.
Ledger already present → skip this stage; adopt the existing ledger and fill only missing
tracker/convention fields.

The copy-ready templates below live in this skill's `templates/` directory —
`${CLAUDE_SKILL_DIR}/templates/`, i.e. `~/.claude/skills/flow-start/templates/` after deploy.

1. **Structure** — create the canonical tree per `workflow/project-structure.md` (cite it,
   don't restate): `projects/<group>/<project>/_support/{docs,spec,plan,workspace,evidence}`.
   Repos and the specs repo come in Stage C — no git repos here. If Stage A wrote files
   beside the source, move them into `_support/docs/` now.
2. **Ledger + ambient pair** — create `_support/PROJECT.md` from
   `~/.claude/skills/flow-core/references/ledger-template.md`. Record the **project token**
   (usually `<project>`; confirm if a shorter token is preferable — it seeds every infra
   name in Stage C). Then create the workspace ambient pair — `AGENTS.md` canonical,
   `CLAUDE.md` importing it. Scope caveat: the workspace pair loads only in sessions opened
   AT the workspace root (every harness's discovery is git-root-bounded from a child repo;
   Claude Code's ancestor walk loads the CLAUDE.md but does not resolve its import) — the
   per-repo `AGENTS.md` that Stage C scaffolds is what child-repo sessions actually read:
   - `AGENTS.md` ← `templates/workspace-agents.md` (English; the conventions block at its
     end is filled at step 4).
   - `CLAUDE.md` ← `templates/workspace-claude.md` (a single `@AGENTS.md`
     import). Claude-specific instructions, if ever needed, go below the import — never
     duplicated into both files.
3. **Pre-authorize the session lane (Claude Code)** — write the workspace
   `.claude/settings.json` allowing `Write`/`Edit` under any `sessions/**` path so plan
   capture and findings never hit a mid-answer permission prompt, using
   `templates/sessions-permissions.json`. Merge additively into an
   existing file — never overwrite user entries.
4. **Memory + conventions** — run the `engram-init-workspace` skill so the multi-repo
   workspace shares one memory bucket from day zero. Then infer client conventions before
   asking: read a sibling project's PROJECT.md / AGENTS.md under the same `<group>/` and
   inherit what matches. Ask (one AskUserQuestion round) ONLY for the residue nothing
   answers — chiefly the **task tracker** (`Tracker` + `Tracker access`: linear / jira /
   none; mcp / cli / api / manual — required fields every later flow skill resolves "the
   tracker" through) and any sealed client conventions or cloud accounts known now.
   Inherited from a sibling → record, don't ask. Branching is NOT collected here —
   git-workflow decides it at first commit intent; only a client-imposed protected-branch
   scheme counts as a sealed convention worth recording. Tracker fields land in PROJECT.md;
   the rest goes as a short block in AGENTS.md. Nothing left to ask → skip the round and
   report what was inherited.
