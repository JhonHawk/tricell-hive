---
paths:
  - "Dockerfile*"
  - "docker-compose*.{yml,yaml}"
  - "**/*.tf"
  - "**/*.tfvars"
  - ".github/workflows/**/*.{yml,yaml}"
---

> Complements `workflow/devops-principles.md` (same path scope) with file-specific conventions. Read it first — don't duplicate.
>
> **Naming any resource these files create** (bucket, cluster, service, security group, DB, subdomain, branch) follows `workflow/infra-naming.md` — read it before inventing a name; a wrong name costs a recreate + migrate, not an edit.

## Infrastructure as Code

### Docker
- **Pin with SHA digest for production images** and any base image referenced from release artifacts: `FROM node:24-alpine@sha256:abc...`. Version tags drift. Skip SHA pinning for ephemeral images (local dev, CI scratch images) where the cost outweighs the reproducibility benefit.
- **Order `COPY` for cache efficiency:** dependency files first (`package.json`, `pnpm-lock.yaml`), then install, then source code. Invalidate only what changed.
- **Run as non-root user.** Add `USER node` (Node) or `USER 1001` after install steps. Never run production containers as root.
- **`.dockerignore` is mandatory.** At minimum: `node_modules`, `.git`, `.env*`, `dist/`, `*.md`, `_support/`.
- **`HEALTHCHECK` instruction** in Dockerfiles — complements the deployment-level health checks from `devops-principles.md`.
- **pnpm in images — never `corepack enable`/`corepack prepare`** (absent from Node 25+). Preferred (pnpm's official pattern): `FROM ghcr.io/pnpm/pnpm:<major>` + `pnpm runtime set node <ver> -g`. On an existing `node:*` base: `RUN npm i -g pnpm@<version>`, pinned to the `packageManager` field.

### Terraform
- **`required_providers` with version constraints** in every root module. Use `~>` for minor version flexibility: `version = "~> 6.0"`.
- **Remote state with locking.** Never use local state for shared infrastructure. S3 + DynamoDB or Terraform Cloud.
- **Naming:** `resource_type_purpose` in snake_case. Example: `aws_iam_role_lambda_execution`. Outputs: descriptive, not generic (`vpc_id`, not `output1`).
- **Modules for repeated patterns.** Extract when 2+ environments share the same resource set. Pin module sources with version tags.
- **Validate with the official HashiCorp binary** — `terraform fmt -check` and `terraform validate` — never by reading the HCL. Not installed → ask the user to approve the install (`CLAUDE.md > System Installations`); its absence never turns validation into a skipped step. **A green `validate` is not a green apply:** it checks syntax and schema, so a constraint the provider enforces at apply time (a rule requiring two statements, a name length, a regional restriction) passes it and still fails — never report validate as proof the config applies.
- **`terraform plan` before apply — always.** In CI, save the plan file and apply the exact plan.

### GitHub Actions
- **Pin third-party actions by full SHA.** `uses: third-party/action@<sha>` not `@v1`. Tags are mutable and supply-chain risk is real. First-party actions (`actions/*`, `github/*`, `aws-actions/*`, `docker/*`, `hashicorp/*`) may use a major-version tag (`@v4`) when Renovate or Dependabot is configured to auto-bump digests.
- **Minimal permissions.** Always declare `permissions:` at workflow or job level. Default to `contents: read`.
- **Cache dependencies** with `actions/cache` or built-in caching (setup-node, setup-python). Key on lockfile hash.
- **`fail-fast: false`** in matrices only when you need all results. Default `fail-fast: true` is correct for most CI.
- **Secrets: never echoed and never passed as CLI args visible in logs** — env-var or secret-manager injection only.
- **pnpm setup: the official action, chosen by major** — `pnpm/setup` for pnpm ≥11 (it requires 11+); `pnpm/action-setup` for pnpm ≤10. Version resolved from `packageManager`; never a `corepack enable` step. Non-GitHub pipelines (Amplify preBuild, GitLab, scripts): pnpm's standalone script or `npm i -g pnpm@<version>`.
- **Stage workflows by seam (env-branch repos).** One `ci.yml`, two jobs: `quick-checks` on `pull_request` into the integration branch (lint/typecheck/unit — no build), `full-suite` on `push` to the integration branch (adds build, integration/e2e) — gate each job with an `if:` on `github.base_ref` / `github.ref_name` so the wrong job skips. Promotion PRs (base `qa`/`production`) trigger only guard workflows, plus a build-only job where the platform builds out-of-band (Amplify/Vercel). Deploy workflows run no correctness suites — they build/publish the artifact and verify post-deploy health. Amplify's `preBuild` runs no tests either (same rule, platform-side).
- **Never pair `paths-ignore` with a required status check** — a docs-only PR never reports the check and hangs unmergeable; verify branch protection/rulesets whenever adding either side. When splitting or renaming jobs, keep required-check job names stable (or update the ruleset in the same change).
- **Contract gates are runnable locally.** A repo whose CI runs contract-compatibility probes (oasdiff/baseline) exposes the same comparison as a package script — `contracts:check` = fetch the target branch + oasdiff breaking — so promotion flows run it before opening the PR (promotion-playbook pre-gate) and contract-bearing repos may wire it as a lefthook pre-push job.
- **Trunk-red signal (repos with post-merge full-suite).** Each carries the trunk-red notifier: `on: workflow_run` of the CI workflow, failure on the integration branch → open/refresh one `trunk-red`-labeled issue; first green closes it. Session ground-truth checks include `gh issue list --label trunk-red --state open` before building on the integration branch; never close a session leaving a trunk-red your own push caused.
