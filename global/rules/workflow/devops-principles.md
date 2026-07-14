---
alwaysApply: true
---

## DevOps Principles

> Universal principles — specific tooling varies per project. Read the project's CI config, Dockerfile, and deployment files before proposing changes.

- **Infrastructure as Code over manual changes.** If a resource is created via console/CLI, it will drift. Prefer Terraform, CloudFormation, CDK, or equivalent. Manual provisioning is acceptable only for one-off experiments, never production.
- **Secrets in deployment configs:** follow `security.md > Authentication & Secrets`. If a `.env.example` exists, reference it — never create `.env` with real values.
- **Every deployment needs a recovery path.** Before deploying, document or script how to recover — a rollback (blue-green swap, previous artifact) where the change is reversible, or a fix-forward path (feature flag, corrective release) where it is not (e.g. an applied schema migration that drops data). If there's no recovery strategy of either kind, the deployment isn't ready.
- **CI pipelines fail fast.** Order stages by speed and likelihood of failure: lint → typecheck → build → test → deploy. Expensive steps (E2E, security scans) run last or in parallel.
- **Docker images must be minimal.** Multi-stage builds for production. Final image should not contain build tools, dev dependencies, or source maps. Pin base images — version tags at minimum, SHA digests for production images (`languages/iac-devops.md`); never `latest`.
- **Health checks for long-running services.** Long-running services in container orchestrators (Kubernetes, Dokploy, ECS, Nomad) need liveness and readiness endpoints; CI should verify the endpoint responds after deploy. Static sites (Vercel, Netlify, S3+CDN), serverless functions, and short-lived workers are exempt — they have no place to hook a probe.
- **Destructive deployment operations** (deploys, restarts, scaling changes, database migrations on production, security group modifications, DNS changes) follow `CLAUDE.md > Destructive Operations` — including its carve-out: non-prod deploy ops with a documented rollback or under a declared flow/pipeline are standing-authorized; production always confirms.
