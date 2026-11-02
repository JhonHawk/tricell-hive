# Diagram Grammar — inline SVG diagrams inside reports

> Adapted from [cathrynlavery/diagram-design](https://github.com/cathrynlavery/diagram-design)
> (v2.3.2, MIT) and re-skinned onto the flow-report baseline tokens (`baseline.html`). It covers
> static inline SVG only: no Mermaid/draw.io import, animation, or PNG/SVG export.

Diagrams follow the report's own design system — never the upstream default skin, never
webfonts. Everything is self-contained: inline SVG, system font stacks, no external requests.

Sections §3, §4, §5 and §8 (skin, primitives, connector rules, accessible SVG) are in
[diagram SVG construction](diagram-svg-construction.md); read it before writing or changing
SVG markup. Both files share one numbering.

## 1. When to draw

- Draw only when the reader learns more from the visual than from a paragraph or table. If a
  3-column table communicates the same thing, use the table. One-shape "diagrams" → write the
  sentence instead.
- **The highest-quality move is usually deletion.** Every node is a distinct idea — two nodes
  that always travel together are one node. Every connection carries information — if the
  relationship is obvious from layout, remove the line. The diagram is done when nothing can
  be removed, not when everything is added.
- Target density: 4/10. Above 9 nodes, split into an overview + a detail diagram.

## 2. Type selection

| If you're showing… | Type | Layout note |
|---|---|---|
| Components + connections in a system | **Architecture** | zones by tier/trust boundary; one primary flow direction (L→R or T→D) |
| Decision logic with branches | **Flowchart** | diamonds for decisions, labeled YES/NO exits |
| Time-ordered messages between actors | **Sequence** | vertical lifelines, activation bars; ≤5 lifelines |
| States + transitions + guards | **State machine** | rounded state boxes; transition labels carry the event/guard |
| Entities + fields + relationships | **ER / data model** | entity = mini table (name header + field rows); crow's-foot or 1/N labels |
| Events positioned in time | **Timeline** | single horizontal axis, ticks + event markers |
| Cross-functional process with handoffs | **Swimlane** | horizontal lanes, one per actor; ≤5 lanes |
| Two-axis positioning / prioritization | **Quadrant** | double-ended axes, quadrant labels as mono eyebrows |
| Stacked abstraction levels | **Layer stack** | full-width bars, one per layer; ≤6 layers |
| Hierarchy through containment / scope | **Nested** | concentric rounded rects; label each ring's edge |
| Parent → children relationships | **Tree** | top-down, orthogonal drops; depth ≤4 |
| Tasks and phases on a timeline | **Gantt** | rows + horizontal bars on a shared date axis; ≤12 tasks |
| Quantitative comparison across categories | **Bar chart** | ≤8 bars; value labels at bar ends, no gridline forest |

Rules of thumb: two types seem useful → pick the dominant axis. Over the complexity budget
(§6) → split into overview + detail. A type outside this table (radar, venn, org chart, …) →
consult the upstream repo before improvising.

## 6. Layout & spacing

**4px grid — coordinates, dimensions, gaps, padding all divisible by 4.** Node widths from
{80, 96, 112, 128, 144, 160, 180, 200, 240, 320}; gaps between nodes from {20, 24, 32, 40,
48}; box padding {8, 12, 16}; radius {4, 6, 8}. Exempt: stroke widths (0.8, 1, 1.2),
opacities, and mono micro-type (7–9px per the §3 scale in diagram-svg-construction.md). Quick check: a coordinate ending in
1, 2, 3, 5, 6, 7, 9 → fix it.

**Complexity budget (per diagram):**

| Limit | Value |
|---|---|
| Nodes | 9 |
| Arrows / transitions | 12 |
| Accent elements | 2 |
| Lifelines (sequence) | 5 |
| Lanes (swimlane) | 5 |
| Items (quadrant) | 12 |
| Entities (ER) | 8 |
| Nesting levels (nested) | 6 |
| Tree depth | 4 |
| Layers (layer stack) | 6 |
| Bars (bar chart) | 8 |
| Tasks (Gantt) | 12 |
| Zones | 3 |
| Annotation callouts | 2 |

Exceeded → split into overview + detail. The connector rules never relax.

## 7. Anti-patterns

| Anti-pattern | Why it fails |
|---|---|
| Dark mode + cyan/purple glow | looks "technical" without design decisions |
| Mono as blanket "dev" font | mono is for technical content; names are sans |
| Identical boxes for every node | erases hierarchy |
| Accent on every "important" node | accent is 1–2 editorial focals, not a signaling system |
| Legend floating inside the diagram | collides with nodes |
| Arrow label without mask, or touching its line | bleeds through / hides the connector |
| Vertical `writing-mode` text | unreadable |
| Diagonal / slanted connectors | orthogonal elbows are mandatory (§5.1 in diagram-svg-construction.md) |
| Two connectors sharing a path or attach point | each must be independently traceable |
| Shadows on nodes | borders are the elevation language here |
| `rx` > 10 on boxes | radius 4–8, or none |
| Bidirectional arrow when one direction is obvious | says nothing |
| 3 identical summary cards under the diagram | vary widths (`1.1fr 1fr 0.9fr`) |

## 9. Pre-output checklist

Run before shipping any diagram in a report:

- [ ] Would a table or paragraph do the same job? (If yes — don't draw.)
- [ ] Right type per §2? Within the §6 budget?
- [ ] Remove test passed: no removable node, mergeable pair, redundant arrow, or redundant label?
- [ ] Accent on ≤2 elements? Legend covers exactly the treatments used?
- [ ] Arrows drawn before boxes? All elbows orthogonal with `r=8`, no diagonals?
- [ ] Every arrow label masked with `d-mask` and a visible 6–10px gap off its line?
- [ ] No overlapping connectors; crossings bridged; shared edges fanned ≥12px?
- [ ] No connector transits behind a non-endpoint box (or the dashed-transit exception applies)?
- [ ] No label mask overlaps a node painted after it?
- [ ] Legend is a bottom strip; `viewBox` expanded ~60px for it?
- [ ] Coordinates/dimensions on the 4px grid?
- [ ] Node names sans 12px/600; technical text mono; no webfont anywhere in the file?
- [ ] Nodes wireframe (fill reserved for the 1–2 focal ones), masks only where a node sits on a connector?
- [ ] `role="img"`, first-child `<title>`, prefixed `<slug>-title`/`<slug>-desc` IDs?
- [ ] **Zero color in presentation attributes** — every fill/stroke comes from a `d-*` class (§3 in diagram-svg-construction.md)?
- [ ] Checked in BOTH themes and in print preview — no white slab, no invisible stroke?
- [ ] Deterministic checks clean: `python3 <skill-dir>/scripts/verify_geometry.py <file>` and
      `python3 <skill-dir>/scripts/self_check.py <file>` (both adapted from upstream; exit
      non-zero on failure)?

The complete MIT notice is in [diagram-design license](license-diagram-design.md).
