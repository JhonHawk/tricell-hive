# Greenfield bootstrap playbook — standing a new project up

The procedure for taking a project from an idea or a bare codebase to a working workspace:
intake of the requirement, the workspace tree and its ledger, and the technical foundation
(naming table, repo matrix, base contracts, CI/CD). Runs conversationally — there is no
wizard command; each stage below is a checklist to work through, skipping what a given
project already has. Rubric for judging the intake document: `requirements-rubric.md`.



The idea/brief arrives in **any** form: raw notes, a conversation, a client email, a rough
document. This stage *produces* the requirements doc from that input — it never demands a
written one as a prerequisite (that demand is exactly what got the old step bypassed).

1. Dispatch **requirement-analyst** per the handoff protocol
   (`~/.claude/skills/flow-core/references/handoff-protocol.md`) with: the raw input,
   the rubric path (`${CLAUDE_SKILL_DIR}/references/requirements-rubric.md` — after deploy,
   `~/.claude/skills/flow-core/references/requirements-rubric.md`), whatever
   client context exists (sibling projects under the same `<group>/`), and the intent:
   "your improved document seeds the specs phase; your questions are what the user takes to
   the next client call — write them client-ready, in the client's language". Do not
   analyze the input yourself in the main thread.
2. From the analyst's report, write into `_support/docs/` (created in Stage B if not yet
   present — stage the files beside the source until then):
   - `requirements-v2.md` — the improved document, additions marked `[PROPUESTO]` so the
     client's words stay distinguishable from ours.
   - `open-questions.md` — blocking questions first, then nice-to-know, each with its
     proposed default assumption.
3. **Gate:** present the scorecard, the count of additions, and the **blocking questions
   verbatim**. The user takes them to the client now, or accepts the proposed defaults and
   proceeds at risk. **An accepted assumption is a decision** — it must survive to the specs
   repo later; record it at the top of `open-questions.md` as a promotion note.

Already have a satisfactory requirements doc → skip to Stage B, noting what you're reusing.


Stands up the workspace tree, the ledger, the ambient pair, and the memory/conventions layer.
Ledger already present → skip this stage; adopt the existing ledger and fill only missing
tracker/convention fields.

The copy-ready templates below live in this skill's `templates/` directory —
`~/.claude/skills/flow-core/templates/` after deploy.

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


This stage **authors** the technical foundation and provisions **NO environments** — the
first deploy of every environment is performed later by the pipeline, per
`~/.claude/skills/flow-core/references/promotion-playbook.md`. Foundation leaves everything
ready so that first deploy is a button-press, not a project. Proportional: a single-repo
project collapses the repo dialogue into one derivation + one dispatch.

1. **Naming table** — instantiate `<project>-specs/conventions/naming.md` from
   `workflow/infra-naming.md` and `~/.claude/skills/flow-core/references/naming-template.md`,
   using the project token recorded in the ledger. Every resource this stage creates gets its
   row BEFORE creation; client exceptions are documented with their reason and the user's
   sign-off. Include the
   **repo branch model** section (each repo's class and branch→environment mapping per
   `git-mechanics.md > Branching`) and the **code-layer conventions** section (boundary
   casing derived from the settled stacks).
2. **Repo matrix, asked per repo** — derive the repo list from the requirements/specs
   (mocks, specs, backend, frontend, transactional services): names per the naming table
   (`<project>-<component>`), stack, targets, repo class. Then, for each repo, present the
   creation plan as a recommendation the user approves — "how do we create this one?"
   (stack, class, branch model, where it lives) — folded into at most one consolidated
   question block. A clean derivation from the signed table needs only confirmation; gate
   harder only on a NEW naming exception, an unsettled repo split, or creation inside a
   client-owned org (outward-visible, expensive to rename). Repos already present → the
   matrix covers only the missing ones. A product whose runtime services share contracts
   and promote together may be born as ONE monorepo instead of N repos — decide it in
   this same question block; the format, layout, and done-criteria live in
   `/monorepo-cutover` (greenfield lane, skip the hoist).
3. **Parallel dispatch** — per the handoff protocol
   (`~/.claude/skills/flow-core/references/handoff-protocol.md`), in parallel where
   independent:
   - **system-designer** — base OpenAPI contracts into `<project>-specs/contracts/`, derived
     from the requirements/reviewed epics, honoring the naming table's code-layer conventions
     (API JSON casing; identifiers English). Every implementation agent later builds against
     these exactly.
   - **database-specialist** — initial schema + migration baseline in the backend repo(s),
     honoring DB naming from the table (including any invariant-name exception).
   - **devops-engineer** — per repo: scaffold (with a minimal repo `AGENTS.md` whose
     conventions block POINTS at the naming table, never copies it, and carries the
     project pointers — `Ledger: ../_support/PROJECT.md`, the workspace-root path, and the
     instruction to open them for project-scope tasks: child-repo sessions never auto-load
     the workspace files), the branch model per the
     repo's class (deployable multi-env: `development` default + `qa` + `production`,
     protections on `qa`/`production`; specs/mocks: single trunk), CI (lint → typecheck →
     build → test, fail fast) green from the first commit, the CD workflow parametrized by
     target environment, and scripted idempotent provisioning versioned in
     `<repo>/_support/infrastructure/` (written now, executed later per the promotion
     playbook). Every resource name comes from the naming table; every script step pairs its
     command with the expected output.
   Verify each agent's claims via Bash before accepting (contracts lint, migrations run
   locally, CI green on the near-empty repos).
