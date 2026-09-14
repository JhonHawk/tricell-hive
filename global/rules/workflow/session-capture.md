---
paths:
  - "**/_support/**"
  - "**/*-specs/**"
  - "**/sessions/**"
---

---
alwaysApply: true
---

## Session Capture & Support Vocabulary

> Path-scoped: loads when the session touches `_support/**`, a specs repo, or `sessions/**`. Codex/opencode/Grok/PI reach it through the `workspace-conventions` skill. Split out of `workflow/project-structure.md`, which owns the 3-level hierarchy and the `_support` vs specs-repo routing decision; this file owns where inside that structure a thing lands.

### Canonical subfolder vocabulary

| Subfolder | Workspace | Repo | Purpose |
|---|---|---|---|
| `docs/` | ✓ | ✓ | Durable documentation (workspace-level: only until a specs repo absorbs it) |
| `spec/` | ✓ | ✓ | Specs, contracts, technical decisions (OpenAPI, schemas, ADRs); workspace-level: pre-specs-repo only |
| `plan/` | ✓ | ✓ | Implementation plans (format per `quality/communication-format.md`) |
| `workspace/` | ✓ | ✓ | **Ephemeral** scratch, AI notes — relocate or delete when work concludes; never let it accumulate |
| `evidence/` | ✓ | ✓ | Screenshots, bug evidence, validation artifacts (curated subset only — retention in `support-artifacts.md`) |
| `scripts/` | — | ✓ | Disposable dev-session utilities |
| `infrastructure/` | — | ✓ | App-specific IaC; shared/foundational → `<project>-infra` |
| `sessions/` | ✓ | ✓ | **Execution journal** (`YYYY-MM-DD-<slug>/`) — placement below |
| `archive/` | ✓ | ✓ | Non-reproducible material kept after work concludes (superseded reports). Dated names |

### Session capture layer

A **session** is one unit of real execution (dev session, sprint close, analysis pass) — distinct from the **intention** layer (business rules and prior analysis: `product/` (business rules in force), `decisions/`, `contracts/`, `epics/`, `conventions/`). Two axes that reference each other, never duplicate:

**Trigger — durable output on explicit signal, not flow membership.** A session folder is created or reused when execution produces a durable artifact on an **explicit signal**: the user asked for the analysis/report, asks to keep a conclusion, or invokes `/flow-plan` for a durable plan. `/flow-plan` writes the portable plan directly; `/flow-build` reuses it for execution. Native harness Plan Mode is an optional drafting aid and never creates Hive authority. Without an explicit signal, OFFER the artifact — don't write it. Boundary vs memory (`memory-routing.md`): `findings.md` is for conclusions a later session re-reads, with an Engram observation pointing at it; a conversational discovery goes to Engram alone. Trivial work with no durable artifact creates no session folder.

**Portable plan approval.** `/flow-plan` normalizes the single Markdown plan and records the
contract digest plus explicit action/target grants in its Authorization section. Use the Flow
session home when a ledger exists, otherwise the standalone Git repository session home. Outside
both, report session-only execution; honor `Session: no` explicitly. A plan status or a native
Plan Mode approval is not consent. `/flow-build` may implement only after explicit implementation
authority is present, and it may reconcile commit/push/PR/merge/deploy only when those actions
were separately authorized and their mandatory gates pass. `verify` never publishes. A hash proves
content integrity, not consent provenance.

**A command's own working state is not a deliverable, and is written without asking.** A manifest one subcommand produces for another (`agents-md-primary audit` → `apply`, `memory-sync audit` → `apply`) lands in `_support/workspace/<name>-<YYYY-MM-DD>.md` — gitignored, dated, disposable — because it is how an expensive analysis survives the session that produced it, not something the user keeps. The explicit-signal rule above governs `sessions/`; this never lands there. Two limits: an analysis whose only consumer is the conversation stays in the conversation, and **live state is never persisted** (a board read, a check status, a deploy job) — a stale copy reads as current, which is worse than not having it.

- **Execution (by time)** → `sessions/YYYY-MM-DD-<slug>/` — `<slug>-plan.md`, `<slug>-findings.md`, optional `analysis/`, `reports/`. The plan contains its frozen contract, authorization and mutable execution evidence. Versioned; mutable during the session, immutable once concluded — a later correction supersedes with a new linked record, never an in-place edit.
- **Intention (by type)** → the durable by-type homes. Execution updates intention; it never replaces it. **Plans are execution:** a task-by-task implementation plan (the HOW) is a session's `plan.md`; a decision/ADR/design-spec (the WHAT and WHY) is intention. Classify by content, not filename.

**Where it lives:** sibling `<project>-specs/` exists → `<project>-specs/sessions/` (versioned; index at `sessions/README.md`, co-located — NOT the ledger, which is non-versioned and only points at the specs repo). Standalone repo → `<repo>/_support/sessions/` (committed). Workspace with no specs repo yet → `<project>/_support/sessions/` (staging).

**Raw stays out of git.** Logs, dumps, build output, raw screenshots, video → `_support/workspace|evidence/YYYY-MM-DD-<slug>/` (gitignored), under the SAME dated slug; the versioned session references them by path.

**Naming & lifecycle:** folders `YYYY-MM-DD-<kebab-slug>`; session top-level files carry the SLUG, not the date (`<slug>-plan.md`, `<slug>-findings.md` — self-identifying in basename-only surfaces). Structure is proportional (no fixed skeleton); a multi-session effort groups under an **initiative** folder. Lifecycle `in-progress → concluded → finalized`, tracked in the sessions index; `finalized` sessions older than 15 days move to `sessions/archived/` on request via `/workspace-archive` (never by age alone for a non-terminal plan). **Full convention (naming rationale, initiative mechanics, back-references):** `flow-core/references/specs-structure.md > Session & initiative conventions`.

**Out of scope:** cross-cutting permanent docs (READMEs, conventions, guidelines) are not session artifacts; a session finding may be *promoted* into `conventions/` — the convention lives in the intention layer.

### Scripts: `_support/scripts/` vs project scripts

- **`_support/scripts/`** — disposable dev-session utilities (data exports, one-off migrations, scratch automation). Safe to delete after use.
- **Scripts that are part of the project** (report generation, DB migrations, CI helpers) live where the repo's convention puts them (`src/`, `scripts/`, the framework's expected home) — infer the target from existing scripts and the framework layout; ask only when no convention exists or two homes are genuinely plausible.

### Plans: portable session plan vs native drafts

- **`<scope>/_support/plan/`** — durable plan artifacts scoped to the project or repo; format per `quality/communication-format.md`.
- **`sessions/YYYY-MM-DD-<slug>/<slug>-plan.md`** — canonical portable Hive plan: one Markdown file with frozen contract, authorization and execution evidence.
- **`~/.claude/plans/*.md`** — optional native Claude Code drafting surface (Shift+Tab); it must be normalized and explicitly approved through `/flow-plan` before `/flow-build` can use it.
