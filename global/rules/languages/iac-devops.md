---
paths:
  - "Dockerfile*"
  - "docker-compose*.{yml,yaml}"
  - "**/*.tf"
  - "**/*.tfvars"
  - ".github/workflows/**/*.{yml,yaml}"
---

> Complements `devops-principles.md` (always-apply) with file-specific conventions. Read it first — don't duplicate.

## Infrastructure as Code

### Docker
- **Pin with SHA digest for production images** and any base image referenced from release artifacts: `FROM node:20-alpine@sha256:abc...`. Version tags drift. Skip SHA pinning for ephemeral images (local dev, CI scratch images) where the cost outweighs the reproducibility benefit.
- **Order `COPY` for cache efficiency:** dependency files first (`package.json`, `pnpm-lock.yaml`), then install, then source code. Invalidate only what changed.
- **Run as non-root user.** Add `USER node` (Node) or `USER 1001` after install steps. Never run production containers as root.
- **`.dockerignore` is mandatory.** At minimum: `node_modules`, `.git`, `.env*`, `dist/`, `*.md`, `_support/`.
- **`HEALTHCHECK` instruction** in Dockerfiles — complements the deployment-level health checks from `devops-principles.md`.

### Terraform
- **`required_providers` with version constraints** in every root module. Use `~>` for minor version flexibility: `version = "~> 5.0"`.
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
