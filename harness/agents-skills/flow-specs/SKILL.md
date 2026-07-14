---
name: flow-specs
description: >
  Manage the project's specs repo (F3 of the flow pack): create it with the canonical
  structure (init), draft a NEW epic or revise an EXISTING one with Gherkin tasks synced
  to the project's tracker (epic / revise), or run the pre-implementation quality gate
  with two independent reviewers (review). Use when creating, revising, or reviewing
  specs/épicas for a client project — including scope changes from stakeholder feedback,
  even after implementation started.
---

# /flow-specs — specs repo lifecycle

Follow the flow contract (`~/.agents/skills/flow-core/SKILL.md`). Parse `$1` as the
subcommand; `revise` is an alias of `epic` for an existing epic (same flow, same review
gate). With no argument, show the subcommands and the epics index status.

## `init` — create the specs repo

1. OPEN per the contract (ledger must exist; record the specs repo in it at CLOSE).
2. Create `<project>/<project>-specs/` following
   `~/.agents/skills/flow-core/references/specs-structure.md` exactly: README index,
   `product/` (map README seeded from the F1/F2 material: portals, actors/roles,
   end-to-end flow — module folders appear as epics define them), `conventions/`,
   `contracts/`, `decisions/`, `evidence/`, `epics/`.
3. **Astro Starlight presentation layer only if the user asked for one** (default:
   markdown-only — no question). When requested, scaffold via the `starlight-docs-site`
   skill's **`spec-site` profile** *over* the canonical structure: content stays at its
   flow-core paths as the source of truth — Starlight serves it, never reorganizes it
   into its own content tree. The sidebar follows the product topology and status rules
   in `specs-structure.md > Optional Astro Starlight presentation layer` — never the
   repo's delivery taxonomy. Query context7 for the installed Astro/Starlight version's
   config when scaffolding.
4. `git init` + initial commit. Suggest (never execute) creating the GitHub remote.
5. Link the tracker project per the ledger's `Tracker` / `Tracker access` fields (mcp →
   load tools via ToolSearch; cli → `acli` for Jira; api → env token; manual/none →
   skip, the specs repo is the task source): find or create the project, record the URL
   in README and PROJECT.md. Fields missing from the ledger → ask once and record them
   (kickoff normally sets them; pre-pack workspaces get them via `/flow-hygiene`).
6. If `_support/` already holds high-level spec material from F1/F2, propose the promotion
   plan (what moves into the repo, what stays as scratch) — file-routing rule question 4.

## `epic <name>` / `revise <epic-ref>` — draft a new epic or revise an existing one

`epic <name>` drafts a new epic; `revise <epic-ref>` is the same flow applied to an
existing one whose scope changed (stakeholder feedback, a new business rule, a direction
change — even after implementation started). Both end in the review gate; the difference
is only steps 1–2.

1. **Draft (`epic`)**: read the requirements doc and existing epics first; a new epic that
   contradicts or duplicates an existing one is a finding, not something to silently write
   around. **Revise (`revise`)**: resolve `<epic-ref>` to its folder (same resolution as
   `review`, below), read its current PRODUCT.md/tasks.md, and treat the change as a diff —
   what is added, dropped, or reshaped, and which already-implemented tasks the change
   invalidates (those become explicit findings for the close report and the next dev
   session's handoff, never silent rework).
2. **Reconcile with the product map first**: identify which `product/` vistas the epic
   creates or modifies (the mandatory `## Affected vistas` section). A vista/module the
   epic introduces gets its map entry drafted as part of this step — an epic may not
   reference a vista that exists nowhere. Business rules are drafted *for the vista
   pages*; PRODUCT.md carries only the delta (`specs-structure.md > product/`).
3. Create or amend `epics/<EPIC-ID>-<slug>/` with PRODUCT.md per the structure reference —
   **business only**: the delta plus its acceptance scenarios (Success Criteria /
   Validation, happy AND negative paths). NO tasks.md yet — task decomposition is
   delivery planning and waits for the gate (step 4). TECH.md comes later, once
   foundation exists — its Relevant Code section needs real paths. PRODUCT.md carries
   **no technical content**: a technical question that surfaces while drafting is parked
   in Open Questions (`technical — resolves in TECH.md`), never answered in the spec.
4. **The business gate is part of this subcommand, not optional**: run `review` on the
   epic — it gates PRODUCT.md + the vista drafts while no task exists to invalidate.
5. **Only on pass, derive delivery**: write tasks.md (Gherkin ACs per task) from the
   gated rules and sync to the declared tracker via its declared access: epic + one issue
   per task, ACs in the issue description. Keep IDs aligned both ways. `Tracker: none` or
   `access: manual` → tasks.md is the source of truth (self-assigned IDs); for `manual`,
   list the tracker updates the user must make in the close report. The tracker never
   sees pre-gate tasks. On revise, keep stable task IDs stable; only new tasks get new
   IDs — the diff from step 1 names the already-implemented tasks the change invalidates.
   Then mark the epic `reviewed` in the README index.

## `review <spec-ref>` — the BUSINESS gate

This gate closes *business* questions: rules, scope, actors, verifiability. It never
resolves technical ones — the technical gate is TECH.md at `flow-plan` (F6), once
foundation exists.

1. **Resolve `<spec-ref>` to an epic folder** (against the specs repo found at OPEN):
   - **Existing path** → a folder is the epic; a file inside it (`PRODUCT.md`/`TECH.md`/
     `tasks.md`) resolves to its containing folder.
   - **Not a path → treat as identifier**: match against `<project>-specs/epics/` by
     ID prefix (`E07`, `TRI-360`) or slug (`mensajeria`), case-insensitive and partial.
     One match → use it; several → AskUserQuestion with the candidates; none → stop with
     the available epics listed from the README index.
   - **No argument** → show the epics index and ask which to review; do not guess.
2. Dispatch BOTH reviewers in parallel (one message, two Agent calls), per the handoff
   protocol (`~/.agents/skills/flow-core/references/handoff-protocol.md`):
   - **spec-quality-reviewer** — pass: the epic path, the rubric path
     (`~/.agents/skills/flow-specs/references/spec-rubric.md`), pointers to sibling epics and
     contracts for implicit-rule hunting, and the intent: "findings feed a go/no-go gate
     before implementation; the user fixes the spec, not the client".
   - **product-critic** — pass: the epic path, the workspace layout (where the other
     epics, contracts, and runtime repos live), and its standing question: "challenge
     whether this should exist in this shape".
   Do NOT review the spec yourself — your job is dispatch and synthesis.
3. Synthesize: deduplicate; where the critic contradicts the reviewer, keep BOTH and mark
   the tension explicitly — resolving it is the user's call, not yours.
   Classify: `blocker` | `gap` | `rethink` | `polish` | `technical-parked`.
   **`technical-parked` is the mandatory classification for any finding whose resolution
   is technical** (schema shape, endpoint design, token/session mechanics, algorithm,
   library choice): it lands in the epic's Open Questions with an owner, marked
   `technical — resolves in TECH.md`. Resolving it by writing the answer into PRODUCT.md
   fails the gate — the rubric's leakage hard check catches exactly that.
4. Gate: if any `blocker` or `rethink` exists, present them via AskUserQuestion
   (fix spec first / proceed as-is / discuss). Do not soften the critic's findings.
5. Apply approved fixes to the epic files (Gherkin rewrites land verbatim). **On pass,
   apply the epic to the product layer**: create/update the `product/` vista pages with
   the decided rules and add the `Influenciada por` entries — the vista absorbs the
   business truth; the epic stays a delta. Update the README index status, mirror changes
   to the tracker (per the ledger's access fields).
6. Render the full report via the `flow-report` skill as
   `spec-review-<epic-slug>.html`, routed per flow-core's **Session reports** rule — the
   durable history of what the gate found and when (the epic files carry the applied
   outcome; this carries the review record). CLOSE per the contract.
