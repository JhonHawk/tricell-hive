# /flow-start — Stage C: Foundation (naming, repos, contracts, CI)

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
