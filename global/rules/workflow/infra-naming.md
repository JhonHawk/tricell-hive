---
alwaysApply: true
---

## Infrastructure Naming

> Generic layer, applies to every client and provider (AWS, Hetzner, GitHub, DNS). Each
> project instantiates it into a concrete table at `<project>-specs/conventions/naming.md`
> (template: flow-core `naming-template.md`); project-specific exceptions are documented
> THERE with their reason, never improvised. When asked to create any infra resource and
> no instantiated table exists, derive from this rule and say so explicitly.

### Golden rule

```text
<project>-<component>-<env>      kebab-case, env ALWAYS last
env ∈ { development, qa, production }
```

- **kebab-case** for every infra resource. Never camelCase or underscores (exceptions below).
- **The environment token goes last, complete, never abbreviated**: `development`, `qa`,
  `production`. Never `dev`, `prod`, `stage`, `staging`.
- The env token is written identically across every layer of the same project: infra,
  apps, IAM, secrets, tags.
- **Length limits (e.g. ALB/Target Groups: 32 chars):** abbreviate the *component*, never
  the env — `internal-apps-bo-production-tg`, not `internal-apps-backoffice-prod-tg`.

> **Provenance:** `env`-last and the full-form spelling are deliberate *local* conventions, not
> canon — Azure CAF leads with resource-type and abbreviates the env token (`prod`/`dev`/`qa`).
> We spell it out (kills `dev`/`develop`/`development` ambiguity) and pin it last (stable
> discriminating slot). Hold to it for internal consistency, not because a standard mandates it.

### Template by resource type

| Resource | Template | Example |
|---|---|---|
| ECS cluster | `<project>-<env>` | `acme-marketplace-qa` |
| ECS service / task definition | `<project>-<app>-<env>` | `acme-marketplace-api-production` |
| EC2 / VPS (tag `Name`) | `<project>-<role>-<env>` | `acme-fleet-api-development` |
| S3 bucket | `<project>-<content>-<env>` | `acme-fleet-media-qa` |
| Secrets Manager | `<project>/<env>/<resource>` | `acme-fleet/qa/jwt` |
| ALB / TG / WAF / VPC / subnets | `<project>-<resource>-<env>` | `acme-marketplace-alb-qa` |
| Security group | `<project>-<role>-<env>-sg` | `acme-fleet-api-qa-sg` |
| SNS / CloudWatch alarms | `<project>-<signal>-<env>` | `acme-marketplace-alerts-qa` |
| Standalone Lambda | `<project>-<action>-<env>` | `acme-fleet-reports-export-qa` |
| Postgres/MySQL database | `<project>_<env>` | `internal_apps_qa` |
| GitHub repo | `<project>-<component>` — **no env** | `acme-marketplace-backend` |
| Non-prod subdomain | `<env-short>-<app>.<domain>` | `qa-api.example.com` |
| Prod subdomain | `<app>.<domain>` — no token | `api.example.com` |
| Env files | `.env.<env>` | `.env.development` |

### Structural exceptions (rules, not inconsistencies)

1. **Subdomains prefix a SHORT env token** (`qa-`, `dev-`) — DNS standard position, one
   wildcard cert per domain, and URLs are a human-typed public interface. The only
   surface where the short token is correct.
2. **Production DNS carries no token** — the clean domain IS production.
3. **GitHub repos and ECR images carry no env** — environment lives in branches/workflows
   and image tags respectively.
4. **Serverless Framework infixes the stage** (`<service>-<stage>-<fn>`) — framework-
   imposed; only the stage vocabulary (`development|qa|production`) is ours to control.
5. **Postgres/MySQL names use snake_case** — hyphens force quoted identifiers; speak the
   engine's language.

### Operational rules

- **Validate against the project's instantiated table before creating any resource.**
  Resource not in the table → add the row first (user sign-off if it needs an exception),
  then create.
- **Never rename existing resources to "fix" their names on your own initiative** —
  renames imply recreate + migrate + repoint; they are planned migrations, not cleanups.
  Legacy names recorded as untouchable in a project's table stay untouched.
- When requesting or scripting any new deployment, cite the project's naming table in the
  prompt/ticket.
