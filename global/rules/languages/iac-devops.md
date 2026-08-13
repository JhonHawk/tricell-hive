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
- **`terraform plan` before apply — always.** In CI, save the plan file and apply the exact plan.

### GitHub Actions
- **Pin third-party actions by full SHA.** `uses: third-party/action@<sha>` not `@v1`. Tags are mutable and supply-chain risk is real. First-party actions (`actions/*`, `github/*`, `aws-actions/*`, `docker/*`, `hashicorp/*`) may use a major-version tag (`@v4`) when Renovate or Dependabot is configured to auto-bump digests.
- **Minimal permissions.** Always declare `permissions:` at workflow or job level. Default to `contents: read`.
- **Cache dependencies** with `actions/cache` or built-in caching (setup-node, setup-python). Key on lockfile hash.
- **`fail-fast: false`** in matrices only when you need all results. Default `fail-fast: true` is correct for most CI.
- **Secrets: never echoed and never passed as CLI args visible in logs** — env-var or secret-manager injection only.
- **pnpm setup: the `pnpm/setup` action** (successor of `pnpm/action-setup`; version resolved from `packageManager`) — never a `corepack enable` step. Non-GitHub pipelines (Amplify preBuild, GitLab, scripts): pnpm's standalone script or `npm i -g pnpm@<version>`.
