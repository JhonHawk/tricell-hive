# Spec-writing playbook — creating the specs repo, drafting epics, and the business gate

The procedure behind the specs layer: standing the repo up, drafting or revising an epic,
and the MANDATORY business gate that must pass before delivery is derived. Artifact shapes
(repo layout, PRODUCT.md / TECH.md skeletons, tasks.md, Gherkin) live in
`specs-structure.md`; the quality bar the reviewers apply lives in `spec-rubric.md`.
Runs conversationally — the stages below are checklists, not a command.


## Stages

Follow the flow contract (`~/.claude/skills/flow-core/SKILL.md`). Work the stage that applies:
subcommand; `revise` is an alias of `epic` for an existing epic (same flow, same review
gate). With no argument, show the subcommands and the epics index status.

## `init` — create the specs repo

1. OPEN per the contract (ledger must exist; record the specs repo in it at CLOSE).
2. Create `<project>/<project>-specs/` following
   `~/.claude/skills/flow-core/references/specs-structure.md` exactly: README index,
   `product/` (map README seeded from the the greenfield bootstrap material — requirements + workspace
   bootstrap: portals, actors/roles,
   end-to-end flow — module folders appear as epics define them), `conventions/`,
   `contracts/`, `decisions/`, `evidence/`, `epics/`.
3. **Ask the presentation format once, at init** (a structured question, folded into the init
   question block): *markdown only* — the canonical structure with no site layer;
   *Astro Starlight* — scaffold via the `starlight-docs-site` skill's **`spec-site`
   profile** *over* the canonical structure (content stays at its flow-core paths as the
   source of truth — Starlight serves it, never reorganizes it into its own content
   tree; sidebar follows the product topology and status rules in
   `specs-structure.md > Optional Astro Starlight presentation layer`, never the repo's
   delivery taxonomy; query context7 for the installed Astro/Starlight version's config
   when scaffolding); or *other (specify)* — the user names the format and it is applied
   over the same canonical structure, which is non-negotiable as the source of truth.
   Recommend by context (large multi-consumer spec sets earn Starlight; otherwise
   markdown) and record the choice in the ledger at CLOSE.
4. `git init` + initial commit. Suggest (never execute) creating the GitHub remote.
5. Link the tracker project per the ledger's `Tracker` / `Tracker access` fields (mcp →
   load tools via ToolSearch; cli → `acli` for Jira; api → env token; manual/none →
   skip, the specs repo is the task source): find or create the project, record the URL
   in README and PROJECT.md. Fields missing from the ledger → ask once and record them
   (the bootstrap playbook normally sets them; pre-pack workspaces get them through the migration playbook).
6. If `_support/` already holds high-level spec material from the bootstrap stage, propose the promotion
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
3. Create or amend `epics/<EPIC-ID>-<slug>/` with PRODUCT.md per the structure reference.
   **Dispatch `sdd-spec-writer` (mode `spec`)** per the handoff protocol with the requirements
   doc, the sibling epics, the vista drafts from step 2 and the rubric path — it writes
   PRODUCT.md and returns its rubric self-score and Open Questions; a one-vista delta may
   stay inline. The content is **business only**: the delta plus its acceptance scenarios (Success Criteria /
   Validation, happy AND negative paths). NO tasks.md yet — task decomposition is
   delivery planning and waits for the gate (step 4). TECH.md comes later, once
   the bootstrap playbook's foundation stage ran — its Relevant Code section needs real paths. PRODUCT.md carries
   **no technical content**: a technical question that surfaces while drafting is parked
   in Open Questions (`technical — resolves in TECH.md`), never answered in the spec.
4. **The business gate is part of this subcommand, not optional**: run `review` on the
   epic — it gates PRODUCT.md + the vista drafts while no task exists to invalidate. On
   `revise`, the gate scopes to the delta: reviewers receive the changed sections plus the
   rules they impact and re-score only the impacted rubric dimensions (unchanged ones
   carry the prior score, marked as carried); a new epic gets the full pass.
5. **Only on pass, derive delivery**: write tasks.md (Gherkin ACs per task) from the
   gated rules and sync to the declared tracker via its declared access (mechanics:
   `flow-core/references/tracker-access.md`; a large sync dispatches `state-fetcher` with the
   approved task set): epic + one issue
   per task, ACs in the issue description. Keep IDs aligned both ways. `Tracker: none` or
   `access: manual` → tasks.md is the source of truth (self-assigned IDs); for `manual`,
   list the tracker updates the user must make in the close report. The tracker never
   sees pre-gate tasks. On revise, keep stable task IDs stable; only new tasks get new
   IDs — the diff from step 1 names the already-implemented tasks the change invalidates.
   Then mark the epic `reviewed` in the README index.

## `review <spec-ref>` — the BUSINESS gate

This gate closes *business* questions: rules, scope, actors, verifiability. It never
resolves technical ones — the technical gate is TECH.md at the `/flow-plan` approval gate,
once the bootstrap playbook's foundation stage ran.

1. **Resolve `<spec-ref>` to an epic folder** (against the specs repo found at OPEN):
   - **Existing path** → a folder is the epic; a file inside it (`PRODUCT.md`/`TECH.md`/
     `tasks.md`) resolves to its containing folder.
   - **Not a path → treat as identifier**: match against `<project>-specs/epics/` by
     ID prefix (`E07`, `TRI-360`) or slug (`mensajeria`), case-insensitive and partial.
     One match → use it; several → ask with the candidates; none → stop with
     the available epics listed from the README index.
   - **No argument** → show the epics index and ask which to review; do not guess.
2. Dispatch BOTH reviewers in parallel (one message, two Agent calls), per the handoff
   protocol (`~/.claude/skills/flow-core/references/handoff-protocol.md`):
   - **sdd-spec-reviewer** — pass: the epic path, the rubric path
     (`${CLAUDE_SKILL_DIR}/references/spec-rubric.md`), pointers to sibling epics and
     contracts for implicit-rule hunting, and the intent: "findings feed a go/no-go gate
     before implementation; the user fixes the spec, not the client".
   - **sdd-product-critic** — pass: the epic path, the workspace layout (where the other
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
4. Gate: if any `blocker` or `rethink` exists, present them via the question tool
   (fix spec first / proceed as-is / discuss). Do not soften the critic's findings.
5. Apply approved fixes to the epic files (Gherkin rewrites land verbatim). **On pass,
   apply the epic to the product layer**: create/update the `product/` vista pages with
   the decided rules and add the `Influenciada por` entries — the vista absorbs the
   business truth; the epic stays a delta. Update the README index status, mirror changes
   to the tracker (per the ledger's access fields).
6. Render the report via the `flow-report` skill as
   `spec-review-<epic-slug>.html`, routed per flow-core's **Session reports** rule — the
   durable history of what the gate found and when (the epic files carry the applied
   outcome; this carries the review record). A delta review below the
   `communication-format.md` trigger records its outcome in Markdown alongside the prior
   report instead of re-rendering HTML. CLOSE per the contract.
