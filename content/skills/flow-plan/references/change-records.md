# Change records

Read when creating or updating a change folder, when closing one and you need the delta formats, or when locating the specs as the global guidance directs. `<specs>` below is the `openspec` directory named by the repository's `Specs` setting; resolve that setting under [project settings](project-settings.md). The layout follows OpenSpec file conventions without its CLI: the folder names, the archive naming, and the delta format match; `project.md` and the sections of `proposal.md` do not, so adopting that CLI later would need a migration.

## Location and lifecycle

Planned work and requirements live in the `openspec` directory that the `Specs` setting names. Without a specs repository it is `_support/openspec` in the repository; with one, `<specs-repo>/openspec`. It is always versioned: never place it in a workspace outside Git or in an ignored path.

**Condition:** resolving a relative `Specs` path.

**Actions:** resolve it from the root of the repository that declares it, not from the session's directory. When a task needs specs that are not at that path, or several repositories share a project without a specs repository, read [Locate the specs](#locate-the-specs) before acting; ask before cloning, and never create a specs repository to fill the gap. An unresolvable path without a URL is a missing value.

The directory holds `project.md` (the project ledger), `specs/<capability>/spec.md` (current requirements per capability; a product map, `product/<module>/<view>.md` beside the `openspec` directory, replaces it as the current-requirements home), `changes/<change-id>/` (one active planned change with its requirement deltas), and `changes/archive/YYYY-MM-DD-<change-id>/` (closed changes; an abandoned change keeps its never-applied deltas under `unapplied-specs/` or `unapplied-product/`, a Hive departure from OpenSpec).

- Every retained plan is a change folder, with or without requirement deltas; small understood edits need none. Add a delta file when the change adds, alters, or removes a requirement that a capability spec or product-map page states, or one the project would state there: an observable behavior, a contract, a permission, or stored data. A change that alters only implementation, tooling, wording, or the plan's own steps needs no delta. Before deciding which requirements a change alters, or that it needs no delta, and before applying the exemption above, search the current-requirements home, and any older document that [Areas still described by older documents](#areas-still-described-by-older-documents) keeps authoritative for the area, for each capability, behavior, command, view, setting, and integration the change adds or alters, each file a user, host, or other process reads or writes that it adds or alters, and the list, catalog, or report that an added item joins, since a requirement that enumerates a closed set never names the item being added. Read every requirement or product-map rule those searches find, including one that enumerates a closed set, such as a "limited to" list or the list of what a command reports, and read every prohibition, such as a SHALL NOT, in each capability spec or product-map page the searches touch, since a prohibition in general terms names neither the item nor its list: adding to that set or doing what it forbids alters it. Record in `proposal.md` the terms searched and the requirements read, or that nothing matched. Name `<change-id>` in English kebab-case, starting with the tracker key when one exists, such as `ark-642-portal-status-contrast`. Use [the plan format](plan-format.md).
- [Flow-close](../../flow-close/SKILL.md#close-the-change-record) archives an integrated or abandoned change, first merging an integrated change's deltas into the current-requirements home.
- Local or sensitive pointers stay in `_support/README.md`. Updating `project.md` or `openspec/` is a versioned change delivered under that repository's Git rules.
- Existing `sessions/` and ledgers stay as history; continuing work keeps its established home, and migration needs explicit authorization.

## Files

| File | Carries | From the [plan template](plan-format.md) |
| --- | --- | --- |
| `proposal.md` | The user's request and the named tickets' acceptance criteria verbatim, why the change exists, scope and exclusions, work not yet specified, constraints, numbered acceptance criteria (`AC<n>`), tracker issues, delivery decisions, the spec search that decided the deltas, current status | Control sheet, Request, Objective, Scope and acceptance, Delivery |
| `design.md` | Verified context, selected approach, contracts, UI view decisions, migration constraints | Verified context, Proposed design |
| `tasks.md` | The change's base commit once the build starts; tasks with their markers, `Closes:` lines, execution, locations, and verification; shared gates; review status, progress, and next step | Tasks, Verification and human review, Review status and progress |
| `research.md` | Retained investigation for this change, only when retention is authorized | — |
| `specs/<capability>/spec.md` | Requirement deltas against the current specs | — |

`proposal.md` and `tasks.md` are required; add `design.md` when the change has design decisions to record. A reader resuming the change starts at `proposal.md`, whose status row names the next step.

Keep in `<specs>/project.md` only what no other source states: ticket state comes from the tracker, delivery state from Git, and the handoff from the active change.

## Requirement deltas

Write each delta file for one capability, using these sections as needed. A delta that creates a capability's spec starts with a preamble, its `# <capability>` title and `## Purpose`, before the first section:

```markdown
## ADDED Requirements
### Requirement: <Requirement name>
The system SHALL <observable behavior>.
#### Scenario: <Scenario name>
- **WHEN** <condition or action>
- **THEN** <expected outcome>

## MODIFIED Requirements
### Requirement: <Existing requirement name, unchanged>
<Complete replacement text and its scenarios>

## REMOVED Requirements
### Requirement: <Existing requirement name>
<Reason for removal>
```

Name a capability after a product or system ability in English kebab-case, such as `template-submission` or `portal-status-contrast`, not after a ticket or epic. A MODIFIED or REMOVED requirement must match an existing requirement name in `<specs>/specs/`.

## Areas still described by older documents

Existing documents that still describe current behavior, such as epics, READMEs, or design notes that mix current rules with delivery history, remain authoritative for areas no capability spec or product map covers yet, until a change moves those rules into the current-requirements home.

When a change touches an area whose current rules live only in an older document, such as an epic that mixes current rules with delivery history, its deltas add those current rules as `ADDED` requirements (or a new product-map page) together with the change's own modifications. Cite the source document in `proposal.md`. At closure the area moves to the current-requirements home, and the older document stays as history for it. Do not rewrite older documents outside a change.

## Product map projects

When the project keeps a product map (`product/<module>/<view>.md` pages of business rules in force), it replaces `<specs>/specs/` as the current-requirements home. Keep the pages in their business voice and location; documentation sites and manuals reuse them.

- `proposal.md` lists the affected views with a one-line delta each.
- Write one delta file per affected view at `changes/<change-id>/product/<module>/<view>.md`, mirroring the page's section headings. Mark each rule `ADDED`, `MODIFIED` (quoting the current rule), or `REMOVED`, and add a new view page whole when the view does not exist yet.
- At closure, apply those rules to the page, add the change to the page's traceability line (for example `Influenciada por`), and leave the other sections untouched.

## Locate the specs

Read when a task needs specs that are not at the `Specs` path, or when several repositories share a project without a specs repository.

- When the specs are not at the path, use a sibling checkout whose `origin` matches the declared URL and report its path; otherwise ask before cloning the URL there.
- If the path holds another repository or the clone fails, pause only the work that needs specs and ask: grant access or correct the URL, move to `_support/openspec` (a contract change), or continue without specs. Never create a specs repository to fill the gap.
- When several repositories share a project without a specs repository, ask before the first change whether to create one (recommended when they share capabilities) or designate one repository, and point every `Specs` at that directory.

## Closing

The [flow-close](../../flow-close/SKILL.md#close-the-change-record) skill applies the deltas above, archives the folder, and versions the record once at close.
