---
name: flow-start
description: >
  Stand up a NEW client project workspace from scratch — the greenfield wizard. Use when
  beginning a project that has no workspace yet: raw notes or a brief in hand, a client
  just signed, "empecemos el proyecto de <cliente>". Walks three dialogue stages — intake
  (produce the requirements doc), bootstrap (workspace + ledger + memory), foundation
  (naming, repos, contracts, CI) — adapting or skipping any stage already done (retrofit).
---

# /flow-start — greenfield project wizard

An interactive, plan-like wizard that stands a project up end to end, fusing three steps
into one conversation: **intake → bootstrap → foundation**. Each stage is a dialogue with at
most one consolidated question block; present each stage's plan like something the user
approves, not a form you fill silently. Follow the flow contract
(`~/.agents/skills/flow-core/SKILL.md`) — but this is the entry that *creates* the ledger, so
the OPEN "ledger must exist" precondition is what the wizard satisfies, not what it assumes.

**Retrofit, don't refuse.** Inspect what already exists before each stage and adapt: a
requirements doc present → skip intake; a ledger present → skip bootstrap; some repos present
→ foundation covers only the gaps. Refuse only when the workspace is **fully established**
(ledger + specs repo + repos all present) — that is maintenance, not a start: point to
`/flow-hygiene` and stop.

## Resolve group & project

`$1` = group (client/domain, lowercase), `$2` = project (kebab-case). Resolve in order:
explicit arguments → infer from cwd when it matches `…/projects/<group>/<project>` (confirm
the inference in one line before creating anything) → ask for whichever token is missing.

---

## Stage A — Intake (produce the requirements doc)

The idea/brief arrives in **any** form: raw notes, a conversation, a client email, a rough
document. This stage *produces* the requirements doc from that input — it never demands a
written one as a prerequisite (that demand is exactly what got the old step bypassed).

1. Dispatch **requirement-analyst** per the handoff protocol
   (`~/.agents/skills/flow-core/references/handoff-protocol.md`) with: the raw input,
   the rubric path (`~/.agents/skills/flow-start/references/requirements-rubric.md`), whatever
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

---

## Stage B — Bootstrap (workspace, ledger, memory, conventions)

1. **Structure** — create the canonical tree per `workflow/project-structure.md` (cite it,
   don't restate): `projects/<group>/<project>/_support/{docs,spec,plan,workspace,evidence}`.
   Repos and the specs repo come in Stage C — no git repos here. If Stage A wrote files
   beside the source, move them into `_support/docs/` now.
2. **Ledger + ambient pair** — create `_support/PROJECT.md` from
   `~/.agents/skills/flow-core/references/ledger-template.md`. Record the **project token**
   (usually `<project>`; confirm if a shorter token is preferable — it seeds every infra
   name in Stage C). Then create the workspace ambient pair — `AGENTS.md` canonical (every
   AGENTS-compatible harness reads it natively), `CLAUDE.md` importing it — so all harnesses
   share one pointer with zero duplication:

   `AGENTS.md` (English; the conventions block is filled at step 4):
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
   Flow phase offering: offer the next `/flow-*` command per the ledger's
   `Current phase` / `Next suggested`; never execute a flow command uninvited,
   and never offer production deploys automatically. (Full directive lives in
   each harness's global layer.)

   <client conventions block, if any — see step 4>
   ```

   `CLAUDE.md`:
   ```markdown
   @AGENTS.md
   ```
   Claude-specific instructions, if ever needed, go below the import — never duplicated
   into both files.
3. **Pre-authorize the session lane (Claude Code)** — write the workspace
   `.claude/settings.json` allowing `Write`/`Edit` under any `sessions/**` path so plan
   capture and findings never hit a mid-answer permission prompt. Merge additively into an
   existing file — never overwrite user entries:
   ```json
   {
     "permissions": {
       "allow": ["Write(./**/sessions/**)", "Edit(./**/sessions/**)"]
     }
   }
   ```
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

Ledger already present → skip this stage; adopt the existing ledger and fill only missing
tracker/convention fields.

---

## Stage C — Foundation (naming, repos, contracts, CI)

This stage **authors** the technical foundation and provisions **NO environments** — the
first deploy of every environment is performed later by the pipeline, per
`~/.agents/skills/flow-core/references/promotion-playbook.md`. Foundation leaves everything
ready so that first deploy is a button-press, not a project. Proportional: a single-repo
project collapses the repo dialogue into one derivation + one dispatch.

1. **Naming table** — instantiate `<project>-specs/conventions/naming.md` from
   `workflow/infra-naming.md` and `~/.agents/skills/flow-core/references/naming-template.md`,
   using the project token. Every resource this stage creates gets its row BEFORE creation;
   client exceptions are documented with their reason and the user's sign-off. Include the
   **repo branch model** section (each repo's class and branch→environment mapping per
   `git-workflow.md > Branching`) and the **code-layer conventions** section (boundary
   casing derived from the settled stacks).
2. **Repo matrix, asked per repo** — derive the repo list from the requirements/specs
   (mocks, specs, backend, frontend, transactional services): names per the naming table
   (`<project>-<component>`), stack, targets, repo class. Then, for each repo, present the
   creation plan as a recommendation the user approves — "how do we create this one?"
   (stack, class, branch model, where it lives) — folded into at most one consolidated
   question block. A clean derivation from the signed table needs only confirmation; gate
   harder only on a NEW naming exception, an unsettled repo split, or creation inside a
   client-owned org (outward-visible, expensive to rename). Repos already present → the
   matrix covers only the missing ones.
3. **Parallel dispatch** — per the handoff protocol, in parallel where independent:
   - **system-designer** — base OpenAPI contracts into `<project>-specs/contracts/`, derived
     from the requirements/reviewed epics, honoring the naming table's code-layer conventions
     (API JSON casing; identifiers English). Every implementation agent later builds against
     these exactly.
   - **database-specialist** — initial schema + migration baseline in the backend repo(s),
     honoring DB naming from the table (including any invariant-name exception).
   - **devops-engineer** — per repo: scaffold (with a minimal repo `AGENTS.md` whose
     conventions block POINTS at the naming table, never copies it), the branch model per the
     repo's class (deployable multi-env: `development` default + `qa` + `production`,
     protections on `qa`/`production`; specs/mocks: single trunk), CI (lint → typecheck →
     build → test, fail fast) green from the first commit, the CD workflow parametrized by
     target environment, and scripted idempotent provisioning versioned in
     `<repo>/_support/infrastructure/` (written now, executed later per the promotion
     playbook). Every resource name comes from the naming table; every script step pairs its
     command with the expected output.
   Verify each agent's claims via Bash before accepting (contracts lint, migrations run
   locally, CI green on the near-empty repos).

---

## CLOSE

CLOSE per the flow contract: the ledger records what each stage produced (requirements doc,
tree, ledger, naming table path, repos, CI links; environments row = "pending — first deploy
runs via the promotion playbook when resources are confirmed"), the decisions taken, and the
`## Current handoff`. Report the created workspace and offer `/flow-specs init` as the next
step — the specs repo is where the accepted intake assumptions and the base contracts become
durable, versioned project memory.
