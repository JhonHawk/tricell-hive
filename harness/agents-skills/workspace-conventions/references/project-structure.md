
## Project & Workspace Structure

The user's projects follow a canonical 3-level hierarchy. Understand this before creating files, proposing directories, or looking for documentation.

### Hierarchy

```
projects/<group>/<project>/    ← group: client/domain, lowercase (acme, tricell); project: kebab-case
  _support/                    ← workspace-level (material shared across repos)
  <repo>/                      ← one folder per codebase
    src/ or app/               ← source root (respect the framework)
    _support/                  ← repo-level (material for this repo only)
```

- The support folder is always named `_support/`; **the path decides the scope**: `<project>/_support/` = shared across repos, `<repo>/_support/` = single codebase. New artifact → "does this apply to more than one repo?" Yes → workspace-level; no → repo-level.
- **A monorepo keeps ONE `_support/`, at the repo root — never one per package.** The hierarchy stays two levels; the package axis moves INSIDE the folder (`_support/docs/<app>/`, `_support/scripts/<app>/`), not into the path. The repo is the unit of versioning and nothing clones a single app, so a third level buys no isolation and costs coverage: every hook and CI glob written for `_support/**` silently misses `apps/*/_support/**`, and that miss stays invisible until someone audits why a directory is never linted. Monorepos are the default shape of new work, so this is the common case.
- **An established level is sticky.** An initiative that already keeps its dated `evidence/`/`sessions/` folders at one level puts new folders at that SAME level — the routing question is answered once per initiative, never re-derived per artifact; moving an initiative's material to the other level is a deliberate migration. The level also decides versioning (a workspace root is often not a git repo), so silent drift changes what gets committed.
- Support files live under `_support/`, never loose beside `src/` or `app/`.

### File-routing rule: `_support` vs the specs repo

Projects on the flow pack have a versioned specs repo (`<project>-specs/`) — the project's durable memory. **Workspace-level** `_support/` is the non-versioned layer that points and expires (a workspace root is typically not a git repo); **repo-level** `_support/` is committed with its repo, so durable repo-scoped material (e.g. a convention in `<repo>/_support/docs/`) is versioned there and travels with a single-repo clone. Before writing any file, answer in order:

1. Must it have history/versioning AND apply to more than one repo? → `<project>-specs/`
2. Temporary, sensitive, raw evidence, or scratch? → `_support/`
3. Single-repo material? → `<repo>/_support/` or the repo's natural location
4. Did a temporary report produce a decision? → promote/summarize it into `<project>-specs/`

No specs repo yet (pre-specs-repo, or outside the flow pack) → durable material may live in workspace-level `_support/docs|spec/` until one exists to absorb it.

### Infra repo split: `<project>-infra` vs app repos

- **App-specific** (Dockerfile, dev compose, the service's CI/CD, its `.env.example`) → the app repo, in `<repo>/` or `<repo>/_support/infrastructure/`.
- **Foundational/shared** (Terraform/IaC, deployment orchestration, reverse-proxy config, DNS, runbooks, secret *templates*) → `<project>-infra`, sibling to the app repos, **when that repo is warranted** (gate below). Never in `<project>-specs` — contracts ≠ IaC.
- **Creation gate first, routing second.** Create `<project>-infra` only when 2+ app repos share foundational infra, or IaC write access must be narrower than merge access to the app repo — never for symmetry, never merely to keep IaC out of the app repo.
- **Routing test:** serves one service → app repo; 2+ or unowned (VPC, DNS, cloud account, IAM, shared cluster/buckets/DB) → `<project>-infra` **where that repo exists**; where it does not, the single app repo, in a top-level `infrastructure/` or `terraform/`, applied by its own workflow with its own deploy role — the credential boundary is the workflow's role, not the repo boundary. Name per `infra-naming.md` (`<project>-infra`, no env token).
- **The single-repo choice carries an extraction condition** — written into the decision record, not left implicit: extract to `<project>-infra` when a second app repo appears, or when IaC write access must narrow. Its cost, stated once: anyone who can merge to the integration branch can edit the IaC.
- **Secrets:** commit only templates (`.env.example`, `*.tfvars.example`). Real secrets and Terraform state stay out of git — secret manager / remote backend.

### Where inside the structure (pre-write essentials)

- **Canonical subfolders** — `docs/` `spec/` `plan/` `workspace/` (ephemeral) `evidence/` `sessions/` `archive/`, plus repo-only `scripts/` and `infrastructure/`. Never invent a sibling.
- **Dated folders are ALWAYS date-first** (`YYYY-MM-DD-<slug>/`); dates only on point-in-time snapshots, never on living documents edited in place. One deliverable = one folder, atomic artifact = loose file.
- **Full conventions** — what each subfolder holds, and scripts/plans placement: `session-capture.md`, via the `workspace-conventions` skill. Retention, versioning, evidence curation, legacy mappings: `support-artifacts.md` (same skill; scoped to `_support/**`).

### Session capture layer

Execution artifacts land in `sessions/YYYY-MM-DD-<slug>/` (`<slug>-plan.md`, `<slug>-tasks.md`, `<slug>-findings.md`). Create or reuse the folder on an **explicit signal**, except for a direct-route task record when an existing dated session, a user request to preserve it, or reusable decisions/verified implementation evidence make the work worth resuming. This narrow exception does not turn conversational research or unsolicited reports into artifacts. **Moved, not deleted:** the full layer — execution-vs-intention split, where `sessions/` lives per repo shape, raw-stays-out-of-git, naming and lifecycle — is `session-capture.md`, loaded via the `workspace-conventions` skill. Read it before creating or placing a session folder.
