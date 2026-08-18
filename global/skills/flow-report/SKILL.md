---
name: flow-report
description: >
  Render an artifact that outlives the conversation — one the user keeps, shares, or
  returns to — as single-file static HTML in five archetypes (document, explainer, review,
  comparison, deck), chosen by what is presented (plan, audit, research, PR review, option
  grid, pitch). NOT for answering the user: an analysis, diagnosis, review verdict, status,
  or option comparison reported back in-thread stays prose however long or rich. Requires
  the artifact test in rules/quality/communication-format.md, plus ≥~300 words and 2+ info
  kinds (tables, diagrams, code, mockups) as necessary-not-sufficient conditions. Never for
  handoffs or live playgrounds.
allowed-tools: Read, Write, Edit, Glob, Grep, Bash(python3 ${CLAUDE_SKILL_DIR}/scripts/*)
---

# flow-report

Produces a single self-contained `.html` file under `_support/workspace/` or `_support/plan/`. Apply only when triggers in `rules/quality/communication-format.md` fire.

## Universal rules

Every output must:

1. **Be self-contained.** All CSS in `<style>`, all JS in `<script>`, all imagery as inline SVG or base64 data URIs. No CDN dependencies that may 404. The file opens correctly offline.
2. **Render readable in first 5 seconds.** Title at top + 1-line TL;DR + table of contents if >5 sections.
3. **Use real layout, not stacked headings.** Comparisons → CSS grid columns. Timelines → horizontal axis. Hierarchy → indentation or boxes. If the structure would be invisible in Markdown, make it visible here.
4. **Be mobile-responsive.** Include `<meta name="viewport" content="width=device-width, initial-scale=1">`. Single-column collapse below 720px.
5. **Avoid default-AI aesthetics.** No gradient cards with emoji headers, no purple-to-pink buttons, no `Inter` everywhere. Default to: serif body (Georgia, Charter, system serif), restrained palette (3-5 colors max), generous whitespace.
6. **Include provenance footer.** Collapsible `<details>` at the bottom with: timestamp, source prompt (truncated), file paths referenced.
7. **Legibility baseline & layout.** Follow `rules/quality/communication-format.md > Layout floor` (Codex/opencode: `references/communication-format.md`) — the shell is the measure, no prose caps, 18px base; per-archetype interpretation below. Start from the archetype skeleton chosen in `Format selection` and adapt it instead of improvising the CSS — every skeleton ships the same validated design system (tokens, type scale, status palette, provenance footer).
8. **Diagrams follow the grammar.** Any inline SVG diagram (architecture, flowchart, sequence, state, ER, timeline, swimlane, quadrant, layers, tree, Gantt, bar) is drawn per `references/diagram-grammar.md` — load it before drawing. Never improvise connector routing, arrow labels, or node styling; the grammar is skinned to the baseline tokens.

## Format selection

Pick the archetype from what you are PRESENTING, not from who asked. The archetype supplies chrome and layout; when a flow skill or playbook prescribes a section contract (the audit-playbook per `flow-core/references/audit-playbook.md`, the spec-writing playbook's review findings by severity), that contract owns the outline and the archetype never overrides it.

| Presenting… | Archetype | Skeleton |
|---|---|---|
| Operational/status report, findings inventory, audit (the audit-playbook), spec/mock review (the spec-writing playbook), hygiene/migration review (the workspace-hygiene-playbook/migration-playbook), weekly dashboard, incident timeline, plan/spec, design tokens | **document** | `references/baseline.html` |
| Code review, PR writeup, annotated diff | **review** | `references/skeleton-review.html` |
| Research synthesis, concept/subsystem explainer, module map, onboarding doc, FAQ | **explainer** | `references/skeleton-explainer.html` |
| Brainstorm, tech selection, A-vs-B, option grid | **comparison** | `references/skeleton-comparison.html` |
| Deck, pitch, guided walkthrough | **deck** | `references/skeleton-deck.html` |
| Live interactive state (sliders, live re-render, editors that export state) | — not this skill | `playground` skill |

- **document** — dense operational record: stat strip, master table + modal, accent cards. The default when no other row clearly wins.
- **explainer** — teach a concept: TOC, collapsible sections, tabbed code, gotchas table, FAQ/glossary; diagram up top.
- **review** — verdict over a code diff: severity-coded margin annotations (`blocking`/`nit`/`nice`), file jump-links, verdict TL;DR. Code diffs only — spec/mock reviews are document.
- **comparison** — undecided options: 3-6 cards (title, mockup/snippet, pros/cons, tradeoff tags), criteria matrix. Never a pre-selected winner.
- **deck** — one idea per 16:9 slide; `←`/`→` navigation, `P` presenter notes, `Esc` exit fullscreen, contents as slide 2. No-JS/print state = slides stacked full-width — the provenance footer lives there.

Content patterns (minimum structure, inside the chosen archetype):

- **Plan / spec** (document): phase/milestone summary table + data-flow or state SVG + risk table (probability × impact) + inline mockups for UI-touching phases + key code snippets.
- **Design tokens** (document): swatches with hex + variable name + usage label; type scale rendered live; spacing as horizontal bars; component variants live with hover states.
- **Incident timeline** (document): horizontal-axis timeline SVG + minute-by-minute table + action checklist.

### Interactivity: presentation only

Sanctioned JS reveals content already in the file: tabs, collapsibles, clickable rows/modals, deck keyboard nav, hover states. Live-state binding — sliders that re-render a preview, knobs that tune values, editors that export state as JSON or prompt — is the `playground` skill's domain. The boundary is state, not visual richness.

### Layout floor by archetype

`Layout floor` applies as: **document / explainer / comparison** — literal (~1280px shell, 18px base, the shell is the measure). **review** — prose and annotations at the 18px floor; the diff is a data grid under the table clause (mono 14-15px in its own scroll container; commentary never inherits the code size). **deck** — the slide is the shell: each 16:9 slide fills edge-to-edge with on-slide type ≥24px; the stacked no-JS state reverts to the standard full-width shell.

## Multi-file staged deliverables

Default: one deliverable = one file. A staged deliverable (exploration → mockups → plan) may split into sibling files under one dated folder (`_support/workspace/YYYY-MM-DD-<slug>/01-exploration.html`, `02-mockups.html`, …) with relative links between them. Each file is individually self-contained, carries its own provenance footer, and passes the deterministic checks on its own.

## Output location

- Ad-hoc reports, briefs, research (nothing a flow skill directs) → `_support/workspace/{kebab-slug}.html`
- Plans persisted in repo → `_support/plan/{kebab-slug}.html`
- **Caller-directed versioned record** (flow-pack reports — spec-review, mock-review, audit) →
  write to the exact versioned path the invoking flow skill names, not workspace. Known
  destinations: the session-capture `reports/` (specs repo →
  `<project>-specs/sessions/<slug>/reports/`, else standalone →
  `<repo>/_support/sessions/<slug>/reports/`) and the audit baseline home
  `<project>-specs/audit/reports/YYYY-MM-DD.html` (`flow-core/references/audit-playbook.md`).
  When the target is the specs repo, version no raster images: reference raw screenshots in
  `_support/evidence/<slug>/` by path, never embed them.
- Embedded images → sibling `images/` folder, referenced relatively (workspace/plan only)

## After writing

Run the deterministic checks (they exit non-zero on failure; fix and re-run until clean — never ship a failing report):

- Every report: `python3 ${CLAUDE_SKILL_DIR}/scripts/self_check.py <file>` — self-containment (no external requests) + accessible-SVG contract.
- Reports embedding SVG diagrams: also `python3 ${CLAUDE_SKILL_DIR}/scripts/verify_geometry.py <file>` — label-mask/node overlap geometry.
- A multi-file deliverable runs the checks on every file individually.

Report to the user:
- Full file path (so they can `open` it)
- Archetype used
- Word count and rough section count
- Suggested next step: `open <path>` to view in browser, or destination for sharing (S3, Notion, Confluence)

If the user requests edits, prefer editing the HTML in place over regenerating from scratch — diff hygiene matters even when the format is HTML.
