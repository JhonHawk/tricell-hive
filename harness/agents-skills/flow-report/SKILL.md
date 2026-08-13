---
name: flow-report
description: >
  Render substantial multi-format agent output as single-file static HTML (plan, audit,
  research, review, comparison). Triggers: ≥~300 words AND 2+ info kinds (tables,
  diagrams, code, mockups). Not for short chat, handoffs, or live playgrounds.
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
7. **Legibility baseline & layout.** Follow `rules/quality/communication-format.md > Layout floor` (Codex/opencode: `references/communication-format.md`) — the shell is the measure, no prose caps, 18px base. Here the shell is ~1280px. Start from `references/baseline.html` — a concrete, validated skeleton (design-system tokens, type scale, status palette, masthead, TOC, stat strip, accent cards, clickable master table, reusable modal + JS, provenance footer) — and adapt it instead of improvising the CSS.
8. **Diagrams follow the grammar.** Any inline SVG diagram (architecture, flowchart, sequence, state, ER, timeline, swimlane, quadrant, layers, tree, Gantt, bar) is drawn per `references/diagram-grammar.md` — load it before drawing. Never improvise connector routing, arrow labels, or node styling; the grammar is skinned to the baseline tokens.

## Category patterns

Match the user request to a category and render its minimum structure.

### Plan / spec
Summary table (phases, milestones, owners) + data-flow or state diagram (SVG) + risk table (probability × impact) + inline mockups for UI-touching phases + critical code snippets with syntax highlighting.

### Code review / PR writeup
Diff rendered as 2-column with line numbers + inline margin annotations categorized `blocking` / `nit` / `nice` + file-list jump-links at top + TL;DR verdict at top.

### Audit / research / explainer
Flow diagram up top (one SVG, clickable nodes) + annotated code snippet beside prose + gotchas table at bottom (one row per edge case) + source list (links, files, commits referenced).

### Brainstorm / comparison
3-6 option cards in CSS grid. Each card: title, mockup or snippet, pros/cons, tradeoff tags. No pre-selected winner.

### Design tokens
Color swatches with hex + variable name + usage label. Typography scale rendered live. Spacing scale as horizontal bars. Component variants (primary/secondary/ghost) shown live with hover states.

### Interactive playgrounds, sliders, live editors — NOT this skill
If the output requires live interactive state (sliders that change preview in real time, knobs to tune values, editors that export state as JSON or natural-language prompt), delegate to the `playground` skill — that's its dedicated domain. `flow-report` covers static rich layout only. The boundary is: `playground` has live state binding, `flow-report` does not.

### Deck / presentation
16:9 slides, one idea per slide. Arrow-key navigation (`←` `→`). Press `P` for presenter notes. Press `Esc` to exit fullscreen. Include a contents slide as slide 1.

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

Report to the user:
- Full file path (so they can `open` it)
- Word count and rough section count
- Suggested next step: `open <path>` to view in browser, or destination for sharing (S3, Notion, Confluence)

If the user requests edits, prefer editing the HTML in place over regenerating from scratch — diff hygiene matters even when the format is HTML.
