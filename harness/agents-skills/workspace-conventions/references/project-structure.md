
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

No specs repo yet (pre-F3, or outside the flow pack) → durable material may live in workspace-level `_support/docs|spec/` until one exists to absorb it.

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
| `evidence/` | ✓ | ✓ | Screenshots, bug evidence, validation artifacts (curated subset only — see Retention) |
| `scripts/` | — | ✓ | Disposable dev-session utilities |
| `infrastructure/` | — | ✓ | App-specific IaC; shared/foundational → `<project>-infra` |
| `sessions/` | ✓ | ✓ | **Execution journal** (`YYYY-MM-DD-<slug>/`) — see Session capture layer below for where it lives |
| `archive/` | ✓ | ✓ | Non-reproducible material kept after work concludes (superseded reports). Dated names |

### Session capture layer

A **session** is one unit of real execution (dev session, sprint close, analysis pass) — distinct from the **intention** layer (business rules and prior analysis: `decisions/`, `contracts/`, `epics/`, `conventions/`). Two axes that reference each other, never duplicate:

**Trigger — durable output on explicit signal, not flow membership.** A session folder is created or reused when execution produces a durable artifact on an **explicit signal**: the user asked for the analysis/report, or asks to keep a conclusion. `flow-plan`/`flow-build` create it as part of F6, and on Claude Code an approved native plan-mode plan in a flow workspace is captured automatically (`flow-plan-capture` hook) unless the plan carries `Session: no`. Without an explicit signal, OFFER the artifact — don't write it. Boundary vs memory (`memory-routing.md`): `findings.md` is for conclusions a later session re-reads, with an Engram observation pointing at it; a conversational discovery goes to Engram alone. Trivial work with no durable artifact creates no session folder.

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

### Generated-artifact naming, grouping & retention

Applies to generated artifacts under `_support/workspace|evidence|archive|plan` — not to `src/` (the framework owns it) or versioned specs (own convention).

**Naming — order elements by primary retrieval axis, most significant first.**
- ISO dates `YYYY-MM-DD`, zero-padded — only then lexicographic = chronological. **Dated folders are ALWAYS date-first** (`2026-06-19-payment-audit/`) — one format across sessions, evidence, workspace runs, and archive: listings sort chronologically and cleanup ("purge everything before X") stays one glob; never `<slug>-YYYY-MM-DD/` for a folder. Loose files order by primary retrieval axis: chronology-primary → date prefix; subject-primary → subject first, date second (`in-vivo-fac-6-2026-07-01.md`).
- A date marks a point-in-time snapshot: immutable artifacts carry one; living documents edited in place do NOT.
- Intention-revealing names; never generic (`report`, `output`, `data`, `temp`, `analysis`).

**Grouping — one deliverable, one folder; atomic artifact, loose file.** 2+ files forming a single deliverable → folder `YYYY-MM-DD-{slug}/`, internal files unprefixed (the folder carries the date). A self-sufficient single artifact → loose file. The conceptual unit decides, not the file count.

**Retention — reproducible-from-source ⇒ ephemeral; non-reproducible ⇒ durable.**
- Raw run output (screenshot dumps, logs, intermediate dumps, build output) → `_support/workspace/YYYY-MM-DD-<run-slug>/`, gitignored, purged at task close. Never committed.
- Durable evidence is the curated subset only → `_support/evidence/YYYY-MM-DD-<slug>/`; curation at close is an explicit appraisal step — keep what documents an AC or bug, purge the rest.
- **Retained raster evidence → WebP lossless** (`cwebp -lossless <in> -o <out>.webp`): bit-exact, ~−75% on UI screenshots, renders natively in browsers and GitHub. Update report paths (`.png`→`.webp`) in the same step. Don't rewrite git history to shrink already-committed rasters.

**What gets versioned — text-that-interprets vs binary, NOT "is it evidence".** Retained ≠ versioned.
- Text that interprets or decides (report/findings markdown, ADR, contract, session index) → versioned always; it references binaries by path, never embeds the dump.
- A binary is versioned only as a non-reproducible source artifact (approved mockup, source diagram, brand asset). QA/in-vivo screenshots are reproducible by re-running the app: the *report* is durable, the screenshots are not.
- **Specs-repo projects (absolute):** only text reaches `<project>-specs/` — validation binaries NEVER do, curated or not; promotion promotes the decision/report, not the screenshots. A standalone repo MAY commit a curated subset into its versioned `_support/evidence/`.

**Discoverability:** naming carries it. A one-line README only for what structure can't encode (a deviation, an ownership boundary, a non-obvious invariant) — never a descriptive file listing that drifts.

### Naming & legacy mappings

- Folders: `kebab-case` for projects/repos, lowercase simple nouns for standard folders (`docs`, `spec`, `plan`, `scripts`).
- Legacy → canonical: `manuals`/`reference` → `docs/`, `artifacts`/`bug-evidence` → `evidence/`, `context-ia` → `workspace/`, `_project/` → `_support/`; relocate `backup` → `_support/backup/`, `todos` → `_support/todos/`; `tmp` → delete (outside a `/flow-hygiene` run, the deletion is a destructive op — confirm per `CLAUDE.md > Destructive Operations`).
- **Retirement:** each legacy mapping lives only while some repo still uses the name; drop the entry when the last one migrates — this rule is not an archive.
