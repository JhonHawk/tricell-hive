---
alwaysApply: true
---

## Project & Workspace Structure

The user's projects follow a canonical 3-level hierarchy. Understand this before creating files, proposing directories, or looking for documentation.

### Hierarchy
```
projects/<group>/<project>/
  _support/          ← workspace-level (shared material across repos)
    docs/  spec/  plan/  workspace/  evidence/
  <repo-a>/          ← one folder per codebase
    src/ or app/
    _support/        ← repo-level (material for this repo only)
      docs/  spec/  plan/  workspace/  evidence/
      scripts/  infrastructure/
  <repo-b>/
```
- `<group>` — client or domain in lowercase: `acme`, `initech`, `globex`, `tricell`
- `<project>` — business initiative in `kebab-case`
- `<repo>` — one concrete codebase per folder

### Scope rule: path determines level, name is uniform
- The folder is always called `_support/`. **The path decides the scope**:
  - `<project>/_support/` → **workspace-level** — shared material across repos of the same project
  - `<repo>/_support/` → **repo-level** — material specific to a single codebase
- When deciding where to place a new artifact: ask "does this apply to more than one repo in the project?". Yes → workspace-level. No → repo-level.

### File-routing rule: `_support` vs the specs repo

Projects following the flow pack have a versioned specs repo (`<project>-specs/`) that is
the project's durable memory. Where one exists, `_support/` is the **non-versioned layer**
(temporary reports, raw evidence, sensitive material, scratch) — it points and expires; the
specs repo preserves. Before writing any file, answer in order:

1. Must it have history/versioning? → `<project>-specs/`
2. Is it temporary, sensitive, raw evidence, or scratch? → `_support/`
3. Does it apply to a single repo? → `<repo>/_support/` or the repo's natural location
4. Did a temporary report produce a decision? → promote/summarize it into `<project>-specs/`

Workspaces WITHOUT a specs repo (pre-F3, or projects outside the flow pack) keep the
original semantics: durable material may live in workspace-level `_support/docs|spec/`
until a specs repo exists to absorb it.

### Infra repo split: `<project>-infra` vs app repos

A project's infrastructure lives by responsibility axis, not all in one place:
- **App-specific** (Dockerfile, dev compose, the service's CI/CD, its `.env.example`) → with the app repo, in `<repo>/` or `<repo>/_support/infrastructure/`.
- **Foundational/shared** (Terraform/IaC, deployment orchestration, reverse-proxy config, DNS, operational runbooks, secret *templates*) → a dedicated `<project>-infra` repo, sibling to the app repos. Never in `<project>-specs` (contracts ≠ IaC), never buried in one app repo.

**Routing test:** does the resource serve one service, or 2+/none? One → app repo. 2+ or unowned (VPC, DNS, cloud account, IAM, shared cluster, buckets, shared DB) → `<project>-infra`.

**When to create it:** 2+ app repos share foundational infra, OR cloud credentials need permission separation. A single-app project with nothing shared keeps infra in its one repo — don't create the repo for symmetry. Name per `infra-naming.md` (`<project>-infra`, no env token).

**Secrets:** commit only templates (`.env.example`, `*.tfvars.example`). Real secrets and Terraform state stay out of git — secret manager / remote backend.

### Canonical subfolder vocabulary

| Subfolder | Workspace-level | Repo-level | Purpose |
|---|---|---|---|
| `docs/` | ✓ | ✓ | Durable documentation (workspace-level: only until a specs repo absorbs it — see file-routing rule) |
| `spec/` | ✓ | ✓ | Specs, contracts, technical decisions (OpenAPI, schemas, ADRs); workspace-level: pre-specs-repo only |
| `plan/` | ✓ | ✓ | Implementation plans (`.md` or `.html` — see `quality/communication-format.md`) |
| `workspace/` | ✓ | ✓ | **Ephemeral** scratch, AI notes — relocate to a durable folder or delete when work concludes; never let it accumulate |
| `evidence/` | ✓ | ✓ | Screenshots, bug evidence, validation artifacts |
| `scripts/` | — | ✓ | Disposable dev-session utilities (repo-level only) |
| `infrastructure/` | — | ✓ | App-specific Terraform/deploy manifests/IaC; shared/foundational infra → `<project>-infra` (see Infra repo split) |
| `sessions/` | ✓ | ✓ | **Execution journal** — one folder per work session (`YYYY-MM-DD-<slug>/`). In a project WITH a specs repo the versioned journal lives in `<project>-specs/sessions/` (see `flow-core/references/specs-structure.md`); `_support/sessions/` is for standalone repos (committed) and workspaces with no specs repo yet (staging, `git mv` later). See "Session capture layer" below |
| `archive/` | ✓ | ✓ | Relocation home for non-reproducible material kept after work concludes (superseded reports, session husks worth keeping). Dated names |

### Session capture layer

A **session** is one unit of real execution — a dev session, a sprint close, an analysis pass. It is distinct from the **intention** layer (business rules and prior technical analysis: `decisions/`, `contracts/`, `epics/`, `conventions/`), which stays intact. Two axes that reference each other, never duplicate:

- **Execution (by time)** → `sessions/YYYY-MM-DD-<slug>/` — `<slug>-plan.md`, `<slug>-findings.md`, optional `analysis/`, `reports/`. Versioned. Mutable during the session, immutable once concluded.
- **Intention (by type)** → the durable by-type homes above. The execution updates the intention; it does not replace it.
- **Plans are execution.** A task-by-task **implementation plan** (the HOW — `- [ ]` steps) is a session's `plan.md`, never a decision. A **decision/ADR or design-spec** (the WHAT and WHY) is intention (`decisions/`, `epics/`). Classify by content, not filename — a dated `*-plan.md` belongs in `sessions/`, not `decisions/`.

**Where it lives (detection rule)** — is there a sibling `<project>-specs/` repo?
- **Yes** → `<project>-specs/sessions/` (versioned, shared). Index: `<project>-specs/sessions/README.md`.
- **No (standalone repo)** → `<repo>/_support/sessions/` (committed; `_support/` is versioned except `workspace/`). Index: `<repo>/_support/sessions/README.md`.
- **Workspace, no specs repo yet** → `<project>/_support/sessions/` (unversioned staging; `git mv` into the specs repo when it exists).

**Raw stays out of git.** Logs, dumps, build output, raw screenshots, video → the relevant `_support/workspace|evidence/YYYY-MM-DD-<slug>/` (gitignored), under the SAME dated slug; the versioned session references them by path. Never put raw in the versioned layer.

**Naming.** Folder: `YYYY-MM-DD-<kebab-slug>` (ISO date prefix; lexicographic = chronological). Multiple sessions the same day → distinct intention-revealing slugs; a numeric tiebreaker (`-2`) only on a real slug collision. **Session top-level files carry the SLUG, not the date** — `<slug>-plan.md`, `<slug>-findings.md`, `<slug>-report.html`: the slug is the semantic handle that makes a hit self-identifying in basename-only surfaces (quick-open, editor tabs, filename/semantic search) where the folder path isn't shown; the date stays the folder's (a date on a living file asserts a fixity it doesn't have). Nested subfolder files (`analysis/…`, `reports/…`) stay short — their path is already specific. (This is a deliberate, scoped exception to "internal files unprefixed": sessions are high-volume and referenced individually; a one-off deliverable folder is not.)

**Structure is optional and proportional** — no fixed skeleton. A type-subfolder (`reports/`, `analysis/`, `internal/`) is created only when there are 2+ artifacts of that type (the one-deliverable-one-folder rule applied to types). A trivial session is just a `<slug>-plan.md` plus a couple of loose files. Homologate HOW artifacts are grouped, not WHAT files exist.

**Initiative (multi-session grouping).** When one effort is too large for a single session — a dense plan split into parts, executed across several sessions or days — group it under an **initiative** folder instead of scattering top-level dated sessions: `sessions/<start-date>-<slug>/` holding `README.md` (index), `findings/` (research), `plan/` (master + numbered parts `00-NN`), and the dated execution sub-sessions *inside* it (`YYYY-MM-DD-<sub>/`, each with its own `reports/`). The initiative is a container folder carrying its **start date** (immutable — preserves the index's chronological order; not a file, so the slug-not-date rule above does not apply); its sub-sessions carry their own dates, so a later day nests inside the initiative instead of fragmenting into a sibling folder. Each plan part carries its own `Status` (`flow-core/references/plan-format.md`); the master plan's part index lists them. **One-off work stays a flat session** — promote a flat session to an initiative only when it grows (the move is `/flow-hygiene`'s). Trigger: `flow-plan write` proposes the split when scope density warrants it.

**Lifecycle** (recorded in the sessions index, not the ledger): `in-progress` (active `<slug>-plan.md`) → `concluded` (work done, promotion pending — `/flow-hygiene` flags it) → `finalized` (durable outputs promoted, raw pruned). A concluded/finalized session is immutable — a later correction supersedes with a new linked record, never an in-place edit.

**Back-reference (by slug).** A session's `<slug>-plan.md` declares `Implements:` the intention it executes; the intention records the session that implemented or changed it (`Session:` in `tasks.md`, `Implementado en: sessions/<slug>` in decisions/epics). When execution diverges from the spec, update the spec (source of truth) and record the session slug as the origin — see `gap-resolution.md > Divergence Between Sources`.

**Index.** The sessions index is versioned and co-located with the sessions (`sessions/README.md`), NOT the ledger. The ledger (`_support/PROJECT.md`) is non-versioned and lives outside the specs repo, so it points to the specs repo and does not list sessions.

**Out of scope.** Cross-cutting permanent docs (READMEs, conventions, guidelines) are not session artifacts. A session *finding* may be promoted into the `conventions/` layer (that promotion is the bridge), but the convention itself lives in the intention layer.

### Repo internal structure
Support files live under `_support/`, never loose at the same level as `src/` or `app/`:
```
<repo>/
  src/ or app/            ← source root (respect the framework)
  workspace/              ← may exist here (legacy) or under _support/
  _support/
    docs/                 ← durable repo-specific documentation
    spec/                 ← specs, contracts, technical decisions
    plan/                 ← implementation plans (.md or .html)
    workspace/            ← ephemeral scratch — relocate or delete when done
    evidence/             ← screenshots, bug evidence, validation artifacts
    scripts/              ← disposable utility scripts (dev-session only, safe to delete)
    infrastructure/       ← app-specific IaC; shared → <project>-infra
```

### Scripts: `_support/scripts/` vs project scripts
- **`_support/scripts/`** — disposable utilities for the dev session (data exports, one-off migrations, scratch automation). Safe to delete after use.
- **Scripts that are part of the project** (report generation, DB migrations, CI helpers) live inside the project source (`src/`, `scripts/`, or wherever the framework expects them). Always confirm the target path with the user before creating a script.

### Plans: `_support/plan/` vs `~/.claude/plans/`
- **`<scope>/_support/plan/`** — durable plan artifacts scoped to the project or repo. Use `.html` when the plan is substantial multi-modal output (>300 words mixing prose with tables, diagrams, or code blocks); use `.md` otherwise. (Full trigger logic in `quality/communication-format.md`.) Agents may write here directly when persisting a plan alongside the codebase.
- **`~/.claude/plans/*.md`** — native Claude Code plan-mode mechanism (Shift+Tab). `.md` by Claude Code convention.

### Generated-artifact naming, grouping & retention

Applies to generated artifacts under `_support/workspace|evidence|archive|plan`. Not to `src/` (the framework owns it) or versioned specs (own convention).

**Naming — order elements by primary retrieval axis, most significant first.**
- Date format is ISO 8601 `YYYY-MM-DD`, zero-padded fixed-width (`2026-06-09`, never `2026-6-9`) — only then does lexicographic order equal chronological.
- Chronology-primary artifacts (snapshots, dated reports, daily exports) → date as **prefix** (`2026-06-19-payment-audit`).
- Subject/type-primary artifacts → lead with the subject, date as **secondary** element (`payment-audit-2026-06-19`) so they group by subject and sort by date within the group.
- A date marks a point-in-time snapshot: immutable artifacts (reports, audits, evidence) carry one; living documents edited in place (active plan, index) do NOT — a date there asserts a fixity they don't have.
- Names are intention-revealing; never generic (`report`, `output`, `data`, `temp`, `analysis`).

**Grouping — one deliverable, one folder; atomic artifact, loose file.**
- 2+ files that form a single deliverable → a folder `{slug}-YYYY-MM-DD/`, internal files unprefixed (the folder carries the date): `report.md`, `report.html`, `images/`.
- A self-sufficient single artifact (a script, one note) → loose file. Three independent scripts = three loose files; one report in three formats = one folder. The conceptual unit decides, not the file count.

**Retention — reproducible-from-source ⇒ ephemeral; original non-reproducible ⇒ durable.**
- Ephemeral (raw run output: the 50 screenshots of an in-vivo run, intermediate dumps, logs, build output) → `_support/workspace/<run-slug>/`, gitignored, purged at task close. Never committed.
- Durable artifacts are *retained* in `_support/evidence/<slug>-YYYY-MM-DD/` — only the curated subset, never the raw dump. Curation at close is an explicit appraisal step: keep the few that document an AC or bug, move them to `evidence/`, purge the rest.
- **Retained raster evidence → WebP lossless.** Convert the curated screenshot subset with `cwebp -lossless <in> -o <out>.webp` before retaining (or committing, in a standalone repo): it is **bit-exact** — no quality loss — yet ~−75% on UI screenshots, because their large flat areas and sharp text compress better losslessly than lossy; WebP renders natively in every browser and on GitHub. When a report references the images, update its paths (`.png`→`.webp`) in the same step so it keeps opening. Measure on one representative image first if unsure. Don't rewrite git history to shrink already-committed rasters — the win is the working tree and future clones, not the `.git` past.

**What gets versioned — the deciding axis is text-that-interprets vs binary, NOT "is it evidence".** Retained ≠ versioned: an artifact can be kept on disk yet never enter git.
- **Text that interprets or decides** (report/findings markdown, ADR, contract, session index) → versioned always. It references binaries by path; it never embeds the dump.
- **A binary is worth versioning only when it is a non-reproducible source artifact** (approved mockup, source diagram, brand asset). **Validation evidence — QA/in-vivo screenshots — is reproducible by re-running the app, so the *report* is the durable artifact and the screenshots are not**: reference them by path, don't version them.
- **Specs-repo projects (absolute):** `_support/` is the non-versioned layer (file-routing rule), so "retained" above means local disk, not git. Only text reaches `<project>-specs/` — validation binaries NEVER do, curated or not. Promoting to the specs repo promotes the *decision/report*, not the screenshots. A standalone repo (no specs repo) MAY commit a curated subset into its versioned `_support/evidence/` when convenient, but the default stays text-durable, binary-by-reference.

**Discoverability — naming carries it; README only for what structure can't express.**
- Default: a clear slug + the convention explain the folder — no README.
- Add a one-line README ONLY for knowledge the structure cannot encode: a deviation from convention, an ownership boundary, or a non-obvious contract/invariant. Scope it to that knowledge.
- Never a descriptive README (a file listing) — it drifts out of sync the moment a file changes. If it would only describe, improve the naming instead.

### Naming
- Folders: `kebab-case` for projects/repos, lowercase simple nouns for standard folders (`docs`, `spec`, `plan`, `scripts`)
- Legacy names → canonical: `manuals`/`reference` → `docs/`, `artifacts`/`bug-evidence` → `evidence/`, `context-ia` → `workspace/`, `_project/` → `_support/` (at workspace level, same name as repo-level)
- Legacy names → relocate: `backup` → `_support/backup/` (DB dumps, file copies, any temporary backup)
- Legacy names → relocate: `todos` → `_support/todos/` (.md files with pending items, review periodically)
- Legacy names → eliminate: `tmp` (delete)

**Retirement policy for legacy mappings.** An entry above lives only while at least one repo still uses the legacy name. When the last repo is migrated, drop the entry — this rule is not an archive.
