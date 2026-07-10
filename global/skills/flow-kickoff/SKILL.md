---
name: flow-kickoff
description: >
  Bootstrap a new client project workspace (F2 of the flow pack): canonical 3-level
  structure, the PROJECT.md ledger, a workspace CLAUDE.md that makes the ledger ambient
  for every future session, unified Engram memory, and client conventions. Use once per
  new project, after intake.
argument-hint: "<group> <project>"
disable-model-invocation: true
---

# /flow-kickoff — workspace bootstrap

Creates the project container everything else assumes. `$1` = group (client/domain,
lowercase: `acme`, `globex`, …), `$2` = project (kebab-case).

Resolve group/project in this order:
1. **Explicit arguments** win.
2. **Infer from cwd** when it matches the canonical hierarchy
   (`…/projects/<group>/<project>`): both tokens come from the path. Confirm the
   inference in one line before creating anything — a wrong guess scaffolds a whole tree.
   If cwd is only the group folder (`…/projects/<group>`), infer the group and ask just
   for the project name.
3. Neither → ask for both.

If the target workspace already exists with content beyond an empty folder, stop and
report what's there — this skill bootstraps, it never overwrites an existing workspace
(`/flow-hygiene` handles repairing one). Standing inside the empty target folder is the
normal inference case, not a conflict.

## Phase 1 — Structure

Create per the global rule `workflow/project-structure.md` (cite it, don't restate it):

```
projects/<group>/<project>/
└── _support/
    ├── docs/  spec/  plan/  workspace/  evidence/
```

Repos and the specs repo come later (F3/F5) — kickoff creates no git repos.

## Phase 2 — Ledger + workspace CLAUDE.md

1. Create `_support/PROJECT.md` from
   `~/.claude/skills/flow-core/references/ledger-template.md`. Record the **project
   token** (usually `<project>`; confirm with the user if a shorter token is preferable —
   it seeds every infra name later, see the naming template).
2. Create the workspace ambient pair — `AGENTS.md` is canonical (Codex, opencode, and
   every AGENTS-compatible harness read it natively), `CLAUDE.md` imports it (per Claude
   Code's `@AGENTS.md` import) so all harnesses get the same pointer with zero
   duplication:

   `AGENTS.md` (English; the conventions block is filled in Phase 4):
   ```markdown
   # Workspace: <group>/<project>
   Read `_support/PROJECT.md` (project ledger) before working — current phase,
   artifact index, and open questions live there.
   File placement follows the flow-core file-routing rule: versioned material →
   `<project>-specs/`; temporary/sensitive/raw → `_support/`.
   Session artifacts (all harnesses): an approved plan lives at
   `sessions/YYYY-MM-DD-<slug>/<slug>-plan.md` (session-capture layout), first line
   `Status: planned`; a plan carrying `Session: no` means the user declined the
   session folder — don't create one. Investigation conclusions the user asks to
   keep go to `<slug>-findings.md` in the same layout. When session artifacts are
   produced, update the ledger's `## Current handoff` and commit them at close —
   standing-authorized, `chore(sessions): <slug>`. `/flow-build` adopts and
   executes any session plan. Trivial fixes, small commits, and investigations
   without kept artifacts proceed ad-hoc with no session machinery; epic-scoped
   formal work may still enter through `/flow-plan` (no command exposed → follow
   the skill files directly).

   <client conventions block, if any — see Phase 4>
   ```

   `CLAUDE.md`:
   ```markdown
   @AGENTS.md
   ```
   Claude-specific instructions, if ever needed, go below the import — never duplicated
   into both files.

3. Pre-authorize the session-artifact lane (Claude Code): write the workspace
   `.claude/settings.json` allowing `Write`/`Edit` under any `sessions/**` path so
   plan capture and findings never hit a mid-answer permission prompt:
   ```json
   {
     "permissions": {
       "allow": ["Write(./**/sessions/**)", "Edit(./**/sessions/**)"]
     }
   }
   ```
   Merge into an existing file additively — never overwrite user entries.

## Phase 3 — Absorb intake + memory

1. If `/flow-intake` produced artifacts outside the workspace, move them into
   `_support/docs/` and index them in the ledger.
2. Run the `engram-init-workspace` skill so the multi-repo workspace shares one memory
   bucket from day zero.

## Phase 4 — Client conventions

Infer before asking: sibling projects under the same `<group>/` usually share the
client's tracker, cloud accounts, and conventions — read a sibling's PROJECT.md /
AGENTS.md first and inherit what matches. Then ask (one AskUserQuestion round) ONLY for
the residue that nothing answers:

- **Task tracker** (`Tracker` + `Tracker access` per the ledger template: linear / jira /
  none; mcp / cli / api / manual) — required fields; every later flow skill resolves
  "the tracker" through them. Inherited from a sibling → record, don't ask.
- **Sealed client conventions** and cloud accounts actually known at kickoff — unknown
  is a valid answer; they land in the ledger when they surface.

Branching model is NOT collected here — git-workflow decides it at first commit intent.
Only a client-imposed protected-branch scheme (e.g. development→qa→master) counts as a
sealed convention worth recording; declarations in the workspace AGENTS.md are what
git-workflow's protected-branch gates read. Tracker fields land in PROJECT.md; the rest
goes as a short block in AGENTS.md (reaching Claude through the CLAUDE.md import).
Naming exceptions wait for the naming table (F5). Nothing left to ask → skip the round
and report what was inherited.

## Phase 5 — Close

CLOSE per the flow contract (`~/.claude/skills/flow-core/SKILL.md`): the ledger's phase
table marks F2 done with today's date, F3 pending. Report the created tree and suggest
`/flow-specs init` as the next step.
