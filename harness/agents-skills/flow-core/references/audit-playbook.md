# Multi-lens audit playbook — the protocol an audit run follows

Preventive multi-specialist audit of a project's runtime repos: technical debt, antipatterns,
and architectural drift, captured BEFORE they compound across the next epics. Harness-agnostic
by design: each project instantiates this playbook as `<project>-specs/audit/README.md` (its
repos, doctrine paths, known debt), and any harness can run the instance as a plain document.

## When to run

- **Epic close** (the recommended baseline cadence) and **before a major architectural change**.
- **On smoke**: 2+ chained bugs from the same origin in one session (contract drift, wrong layer).
- NOT per sprint — ROI drops when the accumulated P2/P3 repeat the previous run's.

## Lenses

A run declares its lens set: `full` (all seven) or a scoped subset (`smells`, `security`, `db`,
`perf`, `arch`, `secrets`, `conventions`). **The lens→agent mapping resolves at RUN TIME against the current
roster** (`agent-routing.md`'s disambiguation table) — never frozen into the project instance;
a project README that hardcodes agent names drifts when the roster changes.

| Lens | Focus | Typical resolver (today) |
|---|---|---|
| `arch` | layers, module boundaries, circular deps, cross-service contracts, data ownership | code-reviewer (altitude) + system-designer doctrine |
| `smells` | naming, duplication, twin encodings of one rule, premature/missing abstractions | code-reviewer (per repo) + the stack specialist |
| `security` | OWASP, audience/cookies, input validation, CORS/CSRF | security-reviewer |
| `db` | schema, indexes, N+1, partition/pagination patterns | database-specialist |
| `perf` | sync I/O in handlers, sequential awaits, bundle size, caching | performance-engineer |
| `secrets` | gitleaks sweep, .env.example correctness, committed keys | secrets-auditor |
| `conventions` | 80/20 census: declared convention files (`<repo>/_support/docs/*-patterns.md`, `<project>-specs/conventions/`) vs code — per-file conformance list, drift since adoption, surfaces with no convention | code-scout (census) + code-reviewer (verdict) |

Per-agent finding cap (default 25) exists to bound the consolidated TOTAL to something
navigable (~175 on a 7-lens run); scale it down when lenses × repos grows. An agent that
finds more prioritizes by impact and says so in its scope notes.

## Severity scale

| Sev | Definition | Expected action |
|---|---|---|
| **P0** | Blocker — incident risk, data loss, exploitable NOW, or blocks the next epic | Fix before the next epic; high-priority tracker ticket |
| **P1** | Active debt — works today but compounds per feature; architectural one-way doors | Fix during the next epic (cherry-picked or dedicated pre-epic sprint) |
| **P2** | Backlog — real improvement, not urgent | Tracker tickets tagged tech-debt; groomed against features |
| **P3** | Informational — style, conscious decisions, far future | Workspace memory or report footnote; no ticket |

## Protocol

0. **Resolve the lens set and its agents.** Default to `full` for a first baseline, a scoped
   lens for follow-ups. Map lens→agent against the CURRENT `agent-routing.md` roster — never
   the instance README's remembered names, which go stale as the roster changes.
1. **Prepare.** Stable stack (no mid-merge). Collect: repo paths + HEAD SHAs, the canonical
   doctrine each reader must load (specs/architecture docs, workspace + repo AGENTS.md), and
   the known pre-existing debt list (tickets, memory) that readers must NOT re-report.
2. **Dispatch readers in parallel** (one message, N agents; no team overhead — findings are
   independent until consolidation). Every dispatch follows `references/handoff-protocol.md`
   (its four elements and the DONE/DONE_WITH_CONCERNS/NEEDS_CONTEXT/BLOCKED return contract —
   a lens that partially failed reports so, never a silent "0 findings"). Audit-specific
   additions per prompt: absolute repo paths, doctrine paths, the uniform finding format
   below, the severity criteria, the finding cap, the known-debt exclusion list, and the
   style: brief, actionable, identify — don't remediate.
3. **Uniform finding format** — one block per finding:
   `severity · repo · file:line · type · description (1-2 lines) · action (1 line, what not how)`.
   Each reader also returns: counts by severity/repo, its top-3 P0+P1, and scope notes (what
   was excluded and why).
4. **Dedup, then refute.** Merge findings across readers (same file:line + same defect = one).
   Dispatch finding-refuter over the deduped set — refuter count per
   `agent-routing.md > Verification runs in fresh context` (that rule owns the budget). Refuted
   findings drop to a "refuted" appendix; the inventory reports UNIQUE, surviving findings, and
   says so (raw counts stay in a per-agent section).
5. **Consolidate → report** via the flow-report skill, passing that destination EXPLICITLY —
   `<project>-specs/audit/reports/YYYY-MM-DD.html`, the versioned record; never let it fall
   back to `_support/workspace/`. No specs repo yet → the standard session-capture form
   (`<repo>/_support/sessions/YYYY-MM-DD-audit-<lens>/reports/`), flagged in the report as
   pre-specs-repo; absent an instance README, offer to bootstrap one from the instantiation
   section below (that write is part of the run). Structure: TL;DR → dashboard (severity
   counts + lens×repo matrix) → P0 prominent → health score per repo
   (`P0×4 + P1×2 + P2×1`) → findings per lens (collapsible) → tentative remediation plan
   (pre-epic / during / backlog) → baseline snapshot for the next run → refuted & excluded →
   provenance footer (date, prompt, lens set, agents, repo SHAs, doctrine consulted).
6. **Post-audit.** Present to the architect/tech lead; P0+P1 → tracker tickets (mechanics: `references/tracker-access.md`; the approved batch dispatches `state-fetcher`; batched to the
   close confirmation per `memory-routing.md > Tracker sync`); P2/P3 → backlog/memory; add the
   run's row to the instance README's history table. Re-run at the next epic close and compare
   baselines — only same-lens runs compare 1:1.

## What the audit is NOT

- Not per-PR code review (that stays with code-reviewer per PR), not lint/build (CI's job),
  not test execution (readers READ tests as audit material — test-shape smells count — but
  never run them), not runtime verification (static read-only; in-vivo checks are a separate
  flow), and never a re-report of known debt.

## Instantiating for a project

Copy this structure into `<project>-specs/audit/README.md` with: the concrete repo list, the
doctrine paths, the known-debt list location, any project severity conventions (ticket
prefixes), and an empty history table (`Date · Report · Closes · P0/P1/P2/P3 · Notes`).
Reports live in `<project>-specs/audit/reports/` — the declared carve-out from the flow
contract's Session reports rule (`flow-core/SKILL.md`), sanctioned in `specs-structure.md`:
baselines need one stable home to compare across runs. Raw evidence (screenshots, logs)
stays in `_support/` per `session-capture.md`. No specs repo → no carve-out: the report
routes through the standard session-capture form
(`<repo>/_support/sessions/YYYY-MM-DD-audit-<lens>/reports/`; a multi-repo audit with no
specs repo stages at the workspace level, `<project>/_support/sessions/…`, and `git mv`s
into the specs repo once it exists), flagged in the report as pre-specs-repo.
