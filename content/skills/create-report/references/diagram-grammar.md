# Diagram grammar: inline SVG diagrams inside reports

> Adapted from [cathrynlavery/diagram-design](https://github.com/cathrynlavery/diagram-design)
> (v2.3.2, MIT) and re-skinned onto the create-report baseline tokens (`baseline.html`). It covers
> static inline SVG only: no Mermaid/draw.io import, animation, or PNG/SVG export.

Diagrams follow the report's own design system and are self-contained: inline SVG, system font stacks, no webfonts, no external requests. Sections §3, §4, §5 and §8 (skin, primitives, connector rules, accessible SVG) are in [diagram SVG construction](diagram-svg-construction.md); both files share one numbering.

## 1. When to draw

- Draw only when the reader learns more from the visual than from a paragraph or table; a one-shape "diagram" is a sentence.
- Deletion first: two nodes that always travel together are one node, and a connection that adjacency or alignment already shows is removed. The diagram is done when nothing can be removed.
- Target density 4/10; above 9 nodes, split into an overview and a detail diagram.

## 2. Type selection

Architecture (zones by tier or trust boundary, one primary flow direction), flowchart (diamonds for decisions, labeled YES/NO exits), sequence, state machine (transition labels carry event and guard), ER (entity as a mini table), timeline, swimlane, quadrant, layer stack, nested, tree, Gantt, bar chart (value labels at bar ends). When two types seem useful, pick the dominant axis. For a type outside this list (radar, venn, org chart), consult the upstream repo before improvising.

## 6. Layout and spacing

**4px grid:** coordinates, dimensions, gaps and padding divisible by 4. Node widths from {80, 96, 112, 128, 144, 160, 180, 200, 240, 320}, gaps {20, 24, 32, 40, 48}, box padding {8, 12, 16}. Exempt: stroke widths, opacities, mono micro-type (7–9px, §3), the radius {4, 6, 8}, the offsets and corner radius of the type tag that rides a node's top edge, the 6–10px gaps of arrow labels and their elevation of about 14px above the arrow, and text baselines.

**Complexity budget per diagram:** 9 nodes, 12 arrows, 2 accent elements, 3 zones, 2 annotation callouts; sequence 5 lifelines; swimlane 5 lanes; quadrant 12 items; ER 8 entities; nested 6 levels; tree depth 4; layer stack 6 layers; bar chart 8 bars; Gantt 12 tasks. Past a limit, split into overview and detail. The connector rules never relax.

## 7. Anti-patterns

Dark mode with cyan or purple glow; mono as a blanket "dev" font (names are sans); identical boxes for every node; accent on every "important" node (accent is 1–2 focals); a legend floating inside the diagram; an arrow label without a mask or touching its line; vertical `writing-mode` text; diagonal connectors; two connectors sharing a path or attach point; shadows on nodes; `rx` above 10; a bidirectional arrow where one direction is obvious; three identical summary cards under the diagram (vary widths, `1.1fr 1fr 0.9fr`).

## 9. Pre-output checklist

- [ ] A table or paragraph would not do the same job; right type (§2) within the budget (§6); nothing removable.
- [ ] Accent on at most 2 elements, and the legend covers exactly the treatments used, as a bottom strip with the `viewBox` expanded ~60px.
- [ ] Arrows drawn before boxes; every elbow orthogonal with `r=8`; no overlapping connectors, crossings bridged, shared edges fanned at least 12px.
- [ ] No connector transits behind a non-endpoint box (unless the dashed-transit exception applies); no label mask overlaps a node painted after it.
- [ ] Every arrow label masked with `d-mask` and a 6–10px gap off its line.
- [ ] Node names sans 12px/600, technical text mono, no webfont; nodes wireframe, masks only where a node sits on a connector.
- [ ] `role="img"`, first-child `<title>`, prefixed `<slug>-title`/`<slug>-desc` IDs; **zero color in presentation attributes**, every fill and stroke from a `d-*` class (§3 in diagram-svg-construction.md).
- [ ] Checked in both themes and in print preview.
- [ ] `python3 <skill-dir>/scripts/verify_geometry.py <file>` and `python3 <skill-dir>/scripts/self_check.py <file>` exit clean.

The complete MIT notice is in [diagram-design license](license-diagram-design.md).
