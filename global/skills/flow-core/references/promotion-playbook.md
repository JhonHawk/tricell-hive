# Promotion & environment playbook

Procedure knowledge for promoting a client project through its pipeline, standing up a
new environment, and auditing live resources against the naming table. Consulted by
deploy sessions (git conventions drive the promotion), by `devops-engineer` dispatches,
and offered by the session hook after a deploy ("you deployed → run the QA walk?").

**This file owns procedure, not authorization.** The gates live elsewhere and are not
restated here:

- **Authorization & gates** — `workflow/git-workflow.md`. Promotion into a
  production-deploying branch always confirms; promotion into `qa` follows the non-prod
  carve-out (declared, not asked); push is confirm-gated. Promotion moves the source
  branch's whole current state, not just this session's changes.
- **Recovery path** — `workflow/devops-principles.md`. Every deployment needs a
  documented rollback or fix-forward path; no recovery path = not ready.

Promotions are **movements through the pipeline, never direct pushes to servers**. If
something only works by SSH, that is a pipeline bug to fix first — every environment's
first deploy included. Source/target branches resolve from the naming table's repo
branch model (canonical chain `development → qa → production`); platform-native repos
(Vercel/Netlify) promote via the platform's own flow. No branch-model table → derive
from `git-workflow.md > Branching` and say so.

## QA promotion walk

1. **Read promotable state.** The ledger states what is promotable and when the last
   promotion happened. Promote partially only when the user explicitly scopes it.
2. **Pre-gates — all pass before anything moves:**
   - Pipeline green on the source branch; no failing checks waved through.
   - Pending DB migrations identified and listed.
   - Rollback path stated (per `devops-principles.md`; no rollback → stop).
   - **No correctness re-runs at the promotion** — the source branch's post-merge CI
     already verified the promoted state; this gate is deploy verification (guards,
     health, smoke, the QA walk below). Exception: a promoted aggregate matching no
     CI-verified state (cherry-picks, hotfix divergence) earns a real verification pass.
   - **Contract-compatibility check runs locally BEFORE opening the promotion PR**
     (repos with a contract baseline): fetch the target branch, run the repo's
     `contracts:check`, and author any justified exception in the ignore file as part
     of the change — never in reaction to the CI probe after pushing.
3. **Deploy via the pipeline.** Trigger the workflow; never replicate its steps by hand.
   Watch the run to completion via `gh`.
4. **Post-deploy verification — two layers, liveness then functional:**
   - **Liveness (always runs):** health endpoints + the project's smoke E2E suite against
     the QA URLs. This proves the server responds; it does NOT prove the UI renders. An
     HTTP/curl smoke is liveness, never functional validation.
   - **Functional validation — a handoff, not an auto-gate:** when the promotion touches
     UI, auth/session, or an integration, liveness alone does not prove it works (a
     logged-in screen full of untranslated keys still returns 200). This is QA's by
     default — most client projects have a human QA who owns the QA environment — so do
     NOT run it automatically and do NOT block the close on it. Never let it pass
     silently: at close, state that functional QA is pending and OFFER the
     **in-vivo-qa-tester** agent against the QA URLs (authenticated real-user session
     walking the promoted ACs + the negative catalog, versioned report per
     `test-report-template.md`). The user decides — human QA covers it, or dispatch the
     agent.
5. **Release notes.** Draft client-facing notes from merged PRs + tracker states since the
   last promotion, per `release-notes-template.md`. Versioned (plain markdown, the durable
   record of what shipped): with a specs repo → `<project>-specs/releases/YYYY-MM-DD-<env>.md`;
   no specs repo → the deploy session's versioned location
   (`<repo>/_support/sessions/<slug>/release-notes-<env>.md`). Per-project overrides
   (channel/language/tone) live in the ledger's `Release notes` row. The user reviews
   before anything is sent — client-facing copy uses usted.
6. **Evidence routing.** Raw evidence (pipeline run links + smoke output) → ephemeral
   `_support/evidence/`. The versioned report (smoke results, any AC walk against QA URLs)
   → the versioned layer (specs repo → `<project>-specs/evidence/<epic>/`; no specs repo →
   `<repo>/_support/sessions/<slug>/reports/`) per `test-report-template.md`, one-line
   summary to the ledger. Purge the raw once synthesized; it is never versioned. If the
   promotion is tracked, move the tracker item's state and comment the evidence links — a
   deploy whose tracker still says "in progress" is not closed.

## Prod delta

Everything the QA walk does, plus:

- **Explicit user confirmation before starting** (always — `git-workflow.md`), restating
  what ships, what could break, and the rollback path.
- **Migration dry-run** against a production-like target before the real run.
- **Topology changes** (DR, multi-region, new accounts) are out of scope — route to
  `cloud-architect` → `devops-engineer` before promoting.
- **Liveness-only verification:** health + smoke against production URLs, then a monitoring
  pause. The functional in-vivo walk does NOT run against production — functional
  validation already happened at QA, and you do not run exploratory browser sessions on
  prod. "deployed" and "verified in prod" are distinct states in the ledger and the
  tracker: move the tracked release item to its final state only after the monitoring
  pause, not at pipeline green.

## Environment setup

Stand up an environment that does not exist yet (`development` | `qa` | `production`), or
re-run idempotently to converge a drifted one. This procedure owns ALL environment
provisioning, including the project's first: `/flow-start` authors the naming table and
the scripts/pipelines but provisions nothing (client resources are rarely confirmed that
early); every environment is executed here, when resources are approved.

1. **Resource plan first (gate):** derive every resource the env needs from the project's
   naming table (`<project>-specs/conventions/naming.md`, instantiated by `/flow-start`),
   adding the missing rows for this env BEFORE creating anything (naming-template rules
   apply: user sign-off for exceptions). Present the resource list + blast radius (client
   account, provider, cost-bearing resources) for approval. One gate; after it, execution
   is autonomous.
2. **Provision via the scripts, not by hand:** run/extend the idempotent provisioning in
   `<repo>/_support/infrastructure/` (dispatch `devops-engineer`). A required manual step
   is a provisioning bug — script it as part of this run so the next env converges from the
   repo.
3. **Secrets & config:** create the env's entries in the project's secret manager
   (Doppler/SSM/GitHub Secrets — whatever the project uses) and `.env.<env>` templates with
   placeholders. Never write real secret values to files; report which secrets the user
   must fill by hand.
4. **Wire the pipeline:** add the env as a CD target (workflow env, branch mapping per the
   naming table's repo branch model) so the FIRST deploy to the new env is performed by the
   pipeline — never by hand.
5. **Close:** run the verify pass below against the new env, and record the env in the
   ledger and the naming table (`status: exists`).

## Verify + naming audit

A standalone health pass, usable anytime:

1. **Smoke + health** against the target env.
2. **Naming audit:** list live resources via the provider CLIs (`aws` / `hcloud` / `gh`)
   and diff against `<project>-specs/conventions/naming.md`. Report drift — a resource that
   exists but is not in the table, or violates its row. Drift is **reported, never
   auto-renamed**; renames are planned migrations (per `infra-naming.md`), not audit
   side effects.
3. **Report + ledger update.**

---
