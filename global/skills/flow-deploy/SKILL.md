---
name: flow-deploy
description: >
  Gated deployment for client projects (F7 of the flow pack): qa promotes through the
  pipeline with pre-gates and post-deploy verification, prod adds explicit confirmation
  and migration dry-run, verify runs the smoke + naming audit standalone, and setup env
  provisions a new environment (development/qa/production) end to end. Use on delivery
  days or whenever promoting, verifying, or standing up an environment.
argument-hint: "[qa | prod | verify | setup <env>]"
disable-model-invocation: true
---

# /flow-deploy — gated promotion

Follow the flow contract (`~/.claude/skills/flow-core/SKILL.md`). Deploys here are
**promotions through the pipeline, never direct pushes to servers** — if something only
works by SSH, that's a pipeline bug to fix first (F5's golden rule, still in force).
Production operations are destructive-class: explicit user confirmation per global rules,
always.

## `qa`

1. OPEN per the contract: the ledger states what is promotable and when the last
   promotion happened. **Promotion moves the source branch's whole current state, not
   just this session's changes** — the branch accumulates work across sessions; promote
   partially only when the user explicitly scopes it.
2. **Pre-gates** (all must pass before anything moves):
   - Pipeline green on the source branch; no failing checks waved through
   - Pending DB migrations identified and listed
   - **Rollback path stated** — per `devops-principles.md`, no rollback = not ready; stop
3. Deploy via the pipeline (trigger the workflow; never replicate its steps by hand).
   Watch the run to completion via `gh`.
4. **Post-deploy verification** — two layers, liveness then functional:
   - **Liveness (always runs)**: health endpoints + the project's smoke E2E suite against
     QA URLs. This proves the server responds — it does NOT prove the UI renders. An
     HTTP/curl smoke is liveness, never the functional validation below; treating the two as
     equivalent is the failure this split prevents.
   - **Functional validation — a handoff, not an auto-gate**: when the promotion touches UI,
     auth/session, or an integration, liveness alone does not prove it works (a logged-in
     screen full of untranslated keys still returns 200). This is QA's by default — most
     client projects have a human QA who owns the QA environment — so do NOT run it
     automatically and do NOT block the close on it. But never let it pass silently: at
     verification/close, state that functional QA is pending and OFFER to run the
     **in-vivo-qa-tester** agent against the QA URLs (authenticated real-user session walking
     the promoted ACs + the negative catalog, versioned report per `test-report-template.md`).
     The user decides — human QA covers it, or dispatch the agent. The failure this prevents
     is the silent skip: asserting "verified" when only liveness ran.
5. Release notes for the client from merged PRs + tracker states since the last
   promotion, per `flow-core/references/release-notes-template.md` — **versioned** (plain
   markdown, the durable record of what shipped, committed with the close): with a specs repo
   → `<project>-specs/releases/YYYY-MM-DD-<env>.md`; **no specs repo** (standalone single
   repo, or workspace pre-specs) → the deploy session's versioned location per the
   session-capture detection rule (`<repo>/_support/sessions/<slug>/release-notes-<env>.md`).
   NEVER `_support/workspace/` (gitignored). Per-project overrides (channel/language/tone) in
   the ledger's `Release notes` row. The user reviews before anything is sent — client-facing
   copy uses usted.
6. CLOSE per the contract: raw evidence (pipeline run links + smoke output) → ephemeral
   `_support/evidence/`; the **versioned report** (smoke results, any AC walk against QA
   URLs) → the versioned layer (specs repo → `<project>-specs/evidence/<epic>/`; no specs repo
   → `<repo>/_support/sessions/<slug>/reports/`) per `test-report-template.md`, one-line
   summary to the ledger. Purge the raw once synthesized; it is never versioned. If the
   promotion is tracked (release task/epic in the tracker), move its state and comment
   the evidence links — a deploy whose tracker still says "in progress" is not closed.

## `prod`

Everything `qa` does, plus:
- Explicit confirmation before starting, restating: what ships, what could break, the
  rollback path
- Migration dry-run against a production-like target before the real run
- Topology changes (DR, multi-region, new accounts) are NOT this skill's job — route to
  cloud-architect → devops-engineer per `agent-routing.md` before promoting
- **Post-deploy verification on prod is liveness-only** — health + smoke against
  production URLs, then a monitoring pause. The functional in-vivo walk does NOT run against
  production: functional validation already happened at QA (human QA or the
  in-vivo-qa-tester walk), and you do not run exploratory browser sessions on prod. "deployed" and "verified in prod" are
  different states in the ledger AND in the tracker: move the tracked release item to its
  final state only after the monitoring pause, not at pipeline green

## `setup <env>`

Stand up an environment that doesn't exist yet (`development` | `qa` | `production` — or
re-run idempotently to converge a drifted one). **This subcommand owns ALL environment
provisioning, including the project's first**: foundation authors the scripts and
pipelines but deliberately provisions nothing (client resources are rarely confirmed
that early); every environment — first QA, prod near delivery, a dev env mid-project —
is executed here, when resources are approved:

1. **Resource plan first (gate):** derive every resource the env needs from the naming
   table — adding the missing rows for this env BEFORE creating anything (naming-template
   rules apply: user sign-off for exceptions). Present the resource list + estimated
   blast radius (client account, provider, cost-bearing resources) via AskUserQuestion.
   One gate; after approval, execution is autonomous.
2. **Provision via the scripts, not by hand:** run/extend the idempotent provisioning in
   `<repo>/_support/infrastructure/` (dispatch devops-engineer per the handoff protocol).
   If setup requires a manual step, that's a provisioning bug — script it as part of this
   run so the next env converges from the repo.
3. **Secrets & config:** create the env's secret entries in the project's secret manager
   (Doppler/SSM/GitHub Secrets — whatever the project uses) and `.env.<env>` templates
   with placeholders. Never write real secret values to files; report which secrets the
   user must fill by hand.
4. **Wire the pipeline:** add the env as a CD target (workflow env, branch mapping per
   the project's branching model) so the FIRST deploy to the new env is performed by the
   pipeline — F5's golden rule applies to every environment, not just the first.
5. CLOSE per the contract: run `verify` against the new env + naming audit, and record
   the env in the ledger and the naming table (`status: exists`).

## `verify`

Standalone health pass, usable anytime:
1. Smoke suite + health endpoints against the target env
2. **Naming audit**: list live resources via `aws` / `hcloud` / `gh` CLIs and diff
   against `<project>-specs/conventions/naming.md`. Report drift (resource exists but
   isn't in the table / violates its row) — finding drift here is the cheap path; finding
   it in a client audit is the expensive one. Never rename anything as part of the audit;
   renames are planned migrations (per `infra-naming.md`).
3. Report + ledger update per the contract.
