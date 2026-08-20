---
name: flow-report
description: >
  Render an artifact that outlives the conversation — one the user keeps, shares, or
  returns to — as single-file static HTML in six archetypes (document, paper, explainer,
  review, comparison, deck), chosen by what is presented (audit, postmortem, design doc,
  research, PR review, option grid, pitch). `paper` is the formal, printable one: serif,
  numbered sections, footnotes, cover page. NOT for answering the user: an analysis, diagnosis, review verdict, status,
  or option comparison reported back in-thread stays prose however long or rich. Requires
  the artifact test in rules/quality/communication-format.md, plus ≥~300 words and 2+ info
  kinds (tables, diagrams, code, mockups) as necessary-not-sufficient conditions. Never for
  handoffs or live playgrounds.
---

# flow-report

Produces a single self-contained `.html` file under `_support/workspace/` or `_support/plan/`. Apply only when triggers in `rules/quality/communication-format.md` fire.

## Universal rules

Every output must:

1. **Be self-contained — with one host allowed and one declared exception.** All CSS in `<style>`, all JS in `<script>`, all imagery as inline SVG or base64 data URIs; the file opens correctly offline. **Google Fonts is the sole remote host allowed without ceremony** — it is the only one the Artifacts CSP lets through, and a failed font request degrades to the system stack instead of breaking the page. **Any other remote host** (a CDN for KaTeX, Mermaid, highlight.js) requires declaring `<meta name="flow-report-assets" content="external">`, which `self_check.py` enforces and reports. That declaration is a real trade, not a formality: the Artifacts CSP blocks every non-font host, so the report cannot be published as an Artifact, and an async CDN script renders **empty** if the reader prints before it loads. Never let layout or legibility depend on a remote request. Screenshots past a handful move to a sibling `images/` folder — and then the PDF, not the HTML, is what circulates.
2. **Render readable in first 5 seconds.** Title at top + 1-line TL;DR + table of contents if >5 sections.
3. **Use real layout, not stacked headings.** Comparisons → CSS grid columns. Timelines → horizontal axis. Hierarchy → indentation or boxes. If the structure would be invisible in Markdown, make it visible here.
4. **Be mobile-responsive.** Include `<meta name="viewport" content="width=device-width, initial-scale=1">`. Single-column collapse below 720px.
5. **Hold the system: engineering terminal, dual palette.** Mono is the grid — labels, data, IDs and chrome sit on a character cell; prose and headings are sans (never mono, never serif). Chrome is dashed hairlines and corner brackets: **no elevation anywhere** — no shadows, no gradient cards, no emoji headers. ONE accent (blue) carries focal emphasis and the informational state, leaving red/amber/green to mean status and nothing else. Every report ships both palettes from the same tokens with a persisted toggle, and **`@media print` forces light regardless of it** — these become PDFs. Print from the browser (`Cmd+P` → Save as PDF); a headless exporter that does not emulate print media keeps the dark palette.
6. **Include provenance footer.** Collapsible `<details>` at the bottom with: timestamp, source prompt (truncated), file paths referenced.
7. **Legibility baseline & layout.** Follow `rules/quality/communication-format.md > Layout floor` (Codex/opencode: `references/communication-format.md`) — the shell is the measure, no prose caps, 18px base; per-archetype interpretation below. Start from the archetype skeleton chosen in `Format selection` and adapt it instead of improvising the CSS — every skeleton ships the same validated design system (tokens, chrome layer, type scale, status palette, theme toggle, print rules, provenance footer). **Frame data, rail prose:** a dashed frame marks tables, stat strips, diagrams and code; prose takes a margin rail and no box. Framing prose too makes a 3,000-word audit read as noise where everything weighs the same. Draw both in CSS — never with text characters (`+ - - -`), which break reflow and pollute the accessibility tree.
8. **Diagrams follow the grammar.** Any inline SVG diagram (architecture, flowchart, sequence, state, ER, timeline, swimlane, quadrant, layers, tree, Gantt, bar) is drawn per `references/diagram-grammar.md` — load it before drawing. Never improvise connector routing, arrow labels, or node styling. **Color goes on `d-*` CSS classes, never in SVG presentation attributes** — a hard-coded `fill="#fff"` cannot follow the theme toggle and cannot invert for print, so it renders a white slab inside a dark report.

## Format selection

Pick the archetype from what you are PRESENTING, not from who asked. The archetype supplies chrome and layout; when a flow skill or playbook prescribes a section contract (the audit-playbook per `flow-core/references/audit-playbook.md`, the spec-writing playbook's review findings by severity), that contract owns the outline and the archetype never overrides it.

| Presenting… | Archetype | Skeleton |
|---|---|---|
| Operational/status report, findings inventory, audit (the audit-playbook), spec/mock review (the spec-writing playbook), hygiene/migration review (the workspace-hygiene-playbook/migration-playbook), weekly dashboard, design tokens | **document** | `references/baseline.html` |
| Formal document that circulates and gets printed: postmortem, design doc, proposal, procedure, standalone guide, plan/spec meant to be signed off | **paper** | `references/skeleton-paper.html` |
| Code review, PR writeup, annotated diff | **review** | `references/skeleton-review.html` |
| Research synthesis, concept/subsystem explainer, module map, onboarding doc, FAQ | **explainer** | `references/skeleton-explainer.html` |
| Brainstorm, tech selection, A-vs-B, option grid | **comparison** | `references/skeleton-comparison.html` |
| Deck, pitch, guided walkthrough | **deck** | `references/skeleton-deck.html` |
| Live interactive state (sliders, live re-render, editors that export state) | — not this skill | `playground` skill |

- **document** — dense operational record: stat strip, master table + modal, accent cards. The default when no other row clearly wins.
- **paper** — a formal document read linearly and printed: serif, capped measure, hierarchical numbering, footnotes, status header and page-aware print. **Page 1 is a cover carrying the title block alone** (doctype, title, subtitle, metadata); abstract, contents and body all start on page 2 — `class="titleblock inline"` drops the cover for a short note. Carries the instruction layer (prerequisites, numbered steps, note/warning/caution callouts, `<kbd>`) that a procedure or guide needs. **Boundary with `starlight-docs-site`: what does the reader receive — one file, or one site?** A PDF attached to an email is this archetype; a manual that lives at a URL and is searched page by page is that skill.
- **explainer** — teach a concept: TOC, collapsible sections, tabbed code, gotchas table, FAQ/glossary; diagram up top.
- **review** — verdict over a code diff: severity-coded margin annotations (`blocking`/`nit`/`nice`), file jump-links, verdict TL;DR. Code diffs only — spec/mock reviews are document.
- **comparison** — undecided options: 3-6 cards (title, mockup/snippet, pros/cons, tradeoff tags), criteria matrix. Never a pre-selected winner.
- **deck** — one idea per 16:9 slide; `←`/`→` navigation, `P` presenter notes, `Esc` exit fullscreen, contents as slide 2. No-JS/print state = slides stacked full-width — the provenance footer lives there.

Content patterns (minimum structure, inside the chosen archetype):

- **Plan / spec** (document): phase/milestone summary table + data-flow or state SVG + risk table (probability × impact) + inline mockups for UI-touching phases + key code snippets.
- **Design tokens** (document): swatches with hex + variable name + usage label; type scale rendered live; spacing as horizontal bars; component variants live with hover states.
- **Postmortem** (paper): abstract + impact (users, duration, scope) + timeline (horizontal-axis SVG and minute-by-minute table) + root cause + detection + resolution + action items **with owner and date** + lessons.
- **Design doc** (paper): context and problem + goals and non-goals + proposed design + alternatives considered and why each was rejected + risks + rollout plan.
- **Procedure / guide** (paper): prerequisites + numbered steps (one action each, verification stated) + callouts at the failure points + what to do when it goes wrong.

### Interactivity: presentation only

Sanctioned JS reveals content already in the file: tabs, collapsibles, clickable rows/modals, deck keyboard nav, hover states, the theme toggle. Live-state binding — sliders that re-render a preview, knobs that tune values, editors that export state as JSON or prompt — is the `playground` skill's domain. The boundary is state, not visual richness: the theme toggle qualifies because it holds no content state.

**No scroll-in reveal animation ships with this system.** A stroke trace needs a dash length per shape and a `clip-path` wipe does not interpolate reliably; both failure modes render a *blank* diagram, which is worse than no animation. Adding motion later means opt-in from JS over an element that is already visible, honoring `prefers-reduced-motion` and `beforeprint`.

### Layout floor by archetype

`Layout floor` applies as: **document / explainer / comparison** — literal (~1280px shell, 18px base, the shell is the measure). **paper** — the ONE sanctioned exception, and it departs the allowed way: the *shell itself* is narrow (~68ch text column at 19px), never a cap on prose inside a wider container; the space left over carries hanging section numbers, and figures/tables opt into full width with `.bleed`. **review** — prose and annotations at the 18px floor; the diff is a data grid under the table clause (mono 14-15px in its own scroll container; commentary never inherits the code size). **deck** — the slide is the shell: each 16:9 slide fills edge-to-edge with on-slide type ≥24px; the stacked no-JS state reverts to the standard full-width shell.

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

- Every report: `python3 $HOME/.agents/skills/flow-report/scripts/self_check.py <file>` — self-containment (no external requests) + accessible-SVG contract.
- Reports embedding SVG diagrams: also `python3 $HOME/.agents/skills/flow-report/scripts/verify_geometry.py <file>` — label-mask/node overlap geometry.
- A multi-file deliverable runs the checks on every file individually.

Report to the user:
- Full file path (so they can `open` it)
- Archetype used
- Word count and rough section count
- Suggested next step: `open <path>` to view in browser, or destination for sharing (S3, Notion, Confluence)

If the user requests edits, prefer editing the HTML in place over regenerating from scratch — diff hygiene matters even when the format is HTML.
