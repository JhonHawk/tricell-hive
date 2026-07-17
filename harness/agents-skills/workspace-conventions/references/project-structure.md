
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
- Support files live under `_support/`, never loose beside `src/` or `app/`.

### File-routing rule: `_support` vs the specs repo

Projects on the flow pack have a versioned specs repo (`<project>-specs/`) — the project's durable memory; `_support/` is the non-versioned layer that points and expires. Before writing any file, answer in order:

1. Must it have history/versioning? → `<project>-specs/`
2. Temporary, sensitive, raw evidence, or scratch? → `_support/`
3. Single-repo material? → `<repo>/_support/` or the repo's natural location
4. Did a temporary report produce a decision? → promote/summarize it into `<project>-specs/`

No specs repo yet (pre-specs-repo, or outside the flow pack) → durable material may live in workspace-level `_support/docs|spec/` until one exists to absorb it.

### Infra repo split: `<project>-infra` vs app repos

- **App-specific** (Dockerfile, dev compose, the service's CI/CD, its `.env.example`) → the app repo, in `<repo>/` or `<repo>/_support/infrastructure/`.
- **Foundational/shared** (Terraform/IaC, deployment orchestration, reverse-proxy config, DNS, runbooks, secret *templates*) → a dedicated `<project>-infra` repo, sibling to the app repos. Never in `<project>-specs` (contracts ≠ IaC), never buried in one app repo.
- **Routing test:** serves one service → app repo; 2+ or unowned (VPC, DNS, cloud account, IAM, shared cluster/buckets/DB) → `<project>-infra`. Create it when 2+ app repos share foundational infra or credentials need permission separation — never for symmetry. Name per `infra-naming.md` (`<project>-infra`, no env token).
- **Secrets:** commit only templates (`.env.example`, `*.tfvars.example`). Real secrets and Terraform state stay out of git — secret manager / remote backend.

### Canonical subfolder vocabulary

| Subfolder | Workspace | Repo | Purpose |
|---|---|---|---|
| `docs/` | ✓ | ✓ | Durable documentation (workspace-level: only until a specs repo absorbs it) |
| `spec/` | ✓ | ✓ | Specs, contracts, technical decisions (OpenAPI, schemas, ADRs); workspace-level: pre-specs-repo only |
| `plan/` | ✓ | ✓ | Implementation plans (`.md` or `.html` — `quality/communication-format.md`) |
| `workspace/` | ✓ | ✓ | **Ephemeral** scratch, AI notes — relocate or delete when work concludes; never let it accumulate |
| `evidence/` | ✓ | ✓ | Screenshots, bug evidence, validation artifacts (curated subset only — retention in `support-artifacts.md`) |
| `scripts/` | — | ✓ | Disposable dev-session utilities |
| `infrastructure/` | — | ✓ | App-specific IaC; shared/foundational → `<project>-infra` |
| `sessions/` | ✓ | ✓ | **Execution journal** (`YYYY-MM-DD-<slug>/`) — see Session capture layer below for where it lives |
| `archive/` | ✓ | ✓ | Non-reproducible material kept after work concludes (superseded reports). Dated names |

### Session capture layer

A **session** is one unit of real execution (dev session, sprint close, analysis pass) — distinct from the **intention** layer (business rules and prior analysis: `product/` (business rules in force), `decisions/`, `contracts/`, `epics/`, `conventions/`). Two axes that reference each other, never duplicate:

**Trigger — durable output on explicit signal, not flow membership.** A session folder is created or reused when execution produces a durable artifact on an **explicit signal**: the user asked for the analysis/report, or asks to keep a conclusion. `flow-plan`/`flow-build` create it as part of the plan/build stage, and on Claude Code an approved native plan-mode plan in a flow workspace is captured automatically (`flow-plan-capture` hook) unless the plan carries `Session: no`. Without an explicit signal, OFFER the artifact — don't write it. Boundary vs memory (`memory-routing.md`): `findings.md` is for conclusions a later session re-reads, with an Engram observation pointing at it; a conversational discovery goes to Engram alone. Trivial work with no durable artifact creates no session folder.

- **Execution (by time)** → `sessions/YYYY-MM-DD-<slug>/` — `<slug>-plan.md`, `<slug>-findings.md`, optional `analysis/`, `reports/`. Versioned; mutable during the session, immutable once concluded — a later correction supersedes with a new linked record, never an in-place edit.
- **Intention (by type)** → the durable by-type homes. Execution updates intention; it never replaces it. **Plans are execution:** a task-by-task implementation plan (the HOW) is a session's `plan.md`; a decision/ADR/design-spec (the WHAT and WHY) is intention. Classify by content, not filename.

**Where it lives:** sibling `<project>-specs/` exists → `<project>-specs/sessions/` (versioned; index at `sessions/README.md`, co-located — NOT the ledger, which is non-versioned and only points at the specs repo). Standalone repo → `<repo>/_support/sessions/` (committed). Workspace with no specs repo yet → `<project>/_support/sessions/` (staging).

**Raw stays out of git.** Logs, dumps, build output, raw screenshots, video → `_support/workspace|evidence/YYYY-MM-DD-<slug>/` (gitignored), under the SAME dated slug; the versioned session references them by path.

**Naming & lifecycle:** folders `YYYY-MM-DD-<kebab-slug>`; session top-level files carry the SLUG, not the date (`<slug>-plan.md`, `<slug>-findings.md` — self-identifying in basename-only surfaces). Structure is proportional (no fixed skeleton); a multi-session effort groups under an **initiative** folder. Lifecycle `in-progress → concluded → finalized`, tracked in the sessions index. **Full convention (naming rationale, initiative mechanics, back-references):** `flow-core/references/specs-structure.md > Session & initiative conventions`.

**Out of scope:** cross-cutting permanent docs (READMEs, conventions, guidelines) are not session artifacts; a session finding may be *promoted* into `conventions/` — the convention lives in the intention layer.

### Scripts: `_support/scripts/` vs project scripts

- **`_support/scripts/`** — disposable dev-session utilities (data exports, one-off migrations, scratch automation). Safe to delete after use.
- **Scripts that are part of the project** (report generation, DB migrations, CI helpers) live where the repo's convention puts them (`src/`, `scripts/`, the framework's expected home) — infer the target from existing scripts and the framework layout; ask only when no convention exists or two homes are genuinely plausible.

### Plans: `_support/plan/` vs `~/.claude/plans/`

- **`<scope>/_support/plan/`** — durable plan artifacts scoped to the project or repo. `.html` for substantial multi-modal plans, `.md` otherwise (`quality/communication-format.md`).
- **`~/.claude/plans/*.md`** — native Claude Code plan-mode mechanism (Shift+Tab); `.md` by convention.

### Artifact naming essentials (pre-write)

Dated folders are ALWAYS date-first (`YYYY-MM-DD-<slug>/`); loose files order by primary retrieval axis (chronology-primary → date prefix; subject-primary → subject first, date second); dates only on point-in-time snapshots, never on living documents edited in place. One deliverable = one folder, atomic artifact = loose file. Folders `kebab-case`, intention-revealing names — never generic. Full conventions (retention, versioning, evidence curation, legacy mappings): `workflow/support-artifacts.md` (path-scoped — loads on touching `_support/**`).
