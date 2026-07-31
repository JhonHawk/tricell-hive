
## Infrastructure Naming

> Always-on in Claude Code. Codex/opencode reach it through the `workspace-conventions` skill (`references/infra-naming.md`), which is why it is also in `SKILL_REFERENCE_INJECTIONS`.

> Generic layer for every client and provider (AWS, Hetzner, GitHub, DNS). Each project instantiates it into a concrete table at `<project>-specs/conventions/naming.md` (template: flow-core `naming-template.md`); project-specific exceptions are documented THERE with their reason, never improvised. No instantiated table → derive from this rule and say so.

### Golden rule

```text
<project>-<component>-<env>      kebab-case, env ALWAYS last
env ∈ { development, qa, production }
```

- **kebab-case** for every infra resource; the env token goes **last, complete, never abbreviated** (`development`, `qa`, `production` — never `dev`, `prod`, `staging`), written identically across every layer of the project (infra, apps, IAM, secrets, tags).
- **Length limits** (e.g. ALB/Target Groups: 32 chars): abbreviate the *component*, never the env — `internal-apps-bo-production-tg`, not `internal-apps-backoffice-prod-tg`.
- The golden rule covers the default case — ECS services/tasks (`acme-marketplace-api-production`), EC2/VPS `Name` tags, S3 buckets (`acme-fleet-media-qa`), ALB/TG/WAF/VPC/subnets, SNS/alarms, standalone Lambdas. Only the deviations below use a different shape.

### Deviations from the golden rule

| Resource | Template | Example | Why |
|---|---|---|---|
| ECS cluster | `<project>-<env>` | `acme-marketplace-qa` | the cluster IS the project scope — no component |
| Security group | `<project>-<role>-<env>-sg` | `acme-fleet-api-qa-sg` | type suffix after env |
| Secrets Manager | `<project>/<env>/<resource>` | `acme-fleet/qa/jwt` | slash hierarchy is the AWS idiom |
| Postgres/MySQL database | `<project>_<env>` | `internal_apps_qa` | hyphens force quoted identifiers — snake_case |
| GitHub repo / ECR image | `<project>-<component>` — no env | `acme-marketplace-backend` | env lives in branches/workflows and image tags |
| Git environment branches | `development` → `qa` → `production` | `production` deploys production | full env token as branch name; only deployable multi-env repos — repo classes in `git-workflow.md > Branching` |
| Non-prod subdomain | `<env-short>-<app>.<domain>` | `qa-api.example.com` | DNS-standard prefix position; human-typed public surface — the only place the short token is correct |
| Prod subdomain | `<app>.<domain>` — no token | `api.example.com` | the clean domain IS production |
| Serverless Framework | `<service>-<stage>-<fn>` | — | framework-imposed stage infix; only the stage vocabulary (`development\|qa\|production`) is ours |
| Env files | `.env.<env>` | `.env.development` | |

### Operational rules

- **Validate against the project's instantiated table before creating any resource.** A name conforming to the table or the golden rule → create it and add its row autonomously. A name that needs a NEW deviation → user sign-off, folded into the task's single question block — never a separate round-trip.
- **Never rename existing resources to "fix" their names on your own initiative** — renames imply recreate + migrate + repoint; they are planned migrations, not cleanups. Legacy names recorded as untouchable stay untouched.
- When requesting or scripting any new deployment, cite the project's naming table in the prompt/ticket.
