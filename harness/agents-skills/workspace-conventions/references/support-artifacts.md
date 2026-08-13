
## Support Artifacts — naming, grouping, retention & versioning

> Loads when the session touches `_support/**`. The pre-write essentials (date-first folders, retrieval-axis file names, one-deliverable-one-folder) live in `workflow/project-structure.md`; this rule owns the full conventions.

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
- Legacy → canonical: `manuals`/`reference` → `docs/`, `artifacts`/`bug-evidence` → `evidence/`, `context-ia` → `workspace/`, `_project/` → `_support/`; relocate `backup` → `_support/backup/`, `todos` → `_support/todos/`; `tmp` → delete (outside a `/flow-workspace` run the deletion is a destructive op — confirm per `CLAUDE.md > Destructive Operations`).
- **Retirement:** each legacy mapping lives only while some repo still uses the name; drop the entry when the last one migrates — this rule is not an archive.
