---
globs:
  - "Dockerfile*"
  - "docker-compose*.{yml,yaml}"
  - "**/*.tf"
  - "**/*.tfvars"
  - ".github/workflows/**/*.{yml,yaml}"
  - "lefthook*.{yml,yaml}"
  - "turbo.json"
match: any
---

## DevOps Principles

> Universal principles — specific tooling varies per project. Read the project's CI config, Dockerfile, and deployment files before proposing changes.
>
> **Incident Response moved to `quality/debugging.md > Incident Response`** — it fires on a live-incident conversation, not on touching a file, so path-scoping it here would make it unreachable exactly when it matters.

- **Infrastructure as Code over manual changes.** If a resource is created via console/CLI, it will drift. Prefer Terraform, CloudFormation, CDK, or equivalent. Manual provisioning is acceptable only for one-off experiments, never production.
- **Secrets in deployment configs:** follow `security.md > Authentication & Secrets`. If a `.env.example` exists, reference it — never create `.env` with real values.
- **Every deployment needs a recovery path** — owned always-on by `CLAUDE.md > Destructive Operations`, since it fires on the act of deploying and not on touching a file here.
- **A path-based change selector excludes what cannot affect the build.** A selector that maps `apps/<name>/*` to "that app changed" fires on its `README.md`, its `AGENTS.md`, its ADRs — running the full app pipeline to prove that prose still compiles. Filter non-build inputs (`*.md` at minimum) BEFORE the path match, and cover it with a test case: the selector is itself code, and this defect is invisible until someone reads the CI bill. **This is the deterministic backstop for the documentation-only carve-out in `git-mechanics.md > PRs & promotion`** — that one is prompt-convention and depends on the agent choosing to skip the PR; this one holds even when the PR exists.
- **CI is staged by seam, not monolithic (env-branch/promotion topologies).** Quick gate on the PR into the integration branch: lint → typecheck → unit/affected tests — no build. Full suite post-merge on the integration branch, async: everything plus build, integration/e2e, scans. A promotion between environments is a release gate — guards plus deploy verification (a pre-merge build-only check only where the platform builds outside the pipeline, e.g. Amplify/Vercel); it never re-runs correctness suites. Where the trunk deploys production directly (platform-native repos), the full gate stays pre-merge on the PR. CI placement only — the agent-side task build gate (`CLAUDE.md > Build & Lint`) is unchanged. Within each stage fail fast: order steps by speed and likelihood of failure. Workflow-file patterns: `languages/iac-devops.md`.
- **Working in the tooling layer includes reporting its observable redundancy.** Two names for the same script, a check that runs in more than one place, independent steps serialized with no shared state, a path the deploy exercises that CI never does — name them when CI, build, or hook config is in play. This is duplication visible in config, not a performance conjecture (`development-principles.md > Don't guess performance`).
- **Docker images must be minimal.** Multi-stage builds for production. Final image should not contain build tools, dev dependencies, or source maps. Pin base images; never `latest` — pinning detail (tags vs SHA digests for production) per `languages/iac-devops.md`.
- **Health checks for long-running services.** Long-running services in container orchestrators (Kubernetes, Dokploy, ECS, Nomad) need liveness and readiness endpoints; CI should verify the endpoint responds after deploy. Static sites (Vercel, Netlify, S3+CDN), serverless functions, and short-lived workers are exempt — they have no place to hook a probe.
- **Destructive deployment operations** (deploys, restarts, scaling changes, database migrations on production, security group modifications, DNS changes) follow `CLAUDE.md > Destructive Operations`, including its non-prod carve-out — and its counterweight: production always confirms.
