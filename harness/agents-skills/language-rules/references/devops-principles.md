
## DevOps Principles

> Universal principles — specific tooling varies per project. Read the project's CI config, Dockerfile, and deployment files before proposing changes.
>
> **Incident Response moved to `quality/debugging.md > Incident Response`** — it fires on a live-incident conversation, not on touching a file, so path-scoping it here would make it unreachable exactly when it matters.

- **Infrastructure as Code over manual changes.** If a resource is created via console/CLI, it will drift. Prefer Terraform, CloudFormation, CDK, or equivalent. Manual provisioning is acceptable only for one-off experiments, never production.
- **Secrets in deployment configs:** follow `security.md > Authentication & Secrets`. If a `.env.example` exists, reference it — never create `.env` with real values.
- **Every deployment needs a recovery path** — owned always-on by `CLAUDE.md > Destructive Operations`, since it fires on the act of deploying and not on touching a file here.
- **CI pipelines fail fast.** Order stages by speed and likelihood of failure: lint → typecheck → build → test → deploy. Expensive steps (E2E, security scans) run last or in parallel.
- **Docker images must be minimal.** Multi-stage builds for production. Final image should not contain build tools, dev dependencies, or source maps. Pin base images; never `latest` — pinning detail (tags vs SHA digests for production) per `languages/iac-devops.md`.
- **Health checks for long-running services.** Long-running services in container orchestrators (Kubernetes, Dokploy, ECS, Nomad) need liveness and readiness endpoints; CI should verify the endpoint responds after deploy. Static sites (Vercel, Netlify, S3+CDN), serverless functions, and short-lived workers are exempt — they have no place to hook a probe.
- **Destructive deployment operations** (deploys, restarts, scaling changes, database migrations on production, security group modifications, DNS changes) follow `CLAUDE.md > Destructive Operations`, including its non-prod carve-out — and its counterweight: production always confirms.
