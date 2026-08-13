# Diagram Grammar — inline SVG diagrams inside reports

> **Upstream source:** adapted from [cathrynlavery/diagram-design](https://github.com/cathrynlavery/diagram-design)
> (v2.3.2, MIT). Re-skinned onto the flow-report baseline tokens (`baseline.html`) and trimmed
> to report-relevant types. **Re-sync point:** to pull upstream updates, re-read that repo's
> `skills/diagram-design/SKILL.md` §4–§9 and its `references/type-*.md`, then re-map onto this
> file. Deliberately NOT adopted: Mermaid/draw.io import pipelines, animation, PNG/SVG export,
> brand onboarding, and long-tail types (medallion, radar, venn, DP matrices, loop, org chart).

Diagrams follow the report's own design system — never the upstream default skin, never
webfonts. Everything is self-contained: inline SVG, system font stacks, no external requests.

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

## 3. Skin — baseline tokens inside SVG

SVG presentation attributes take raw hex (`var()` does not work there). These values mirror
`baseline.html` `:root`; if the report swapped its accent family, swap it here identically.

| Role | Value | Use |
|---|---|---|
| paper | `#fffdf8` | SVG background rect + every node/label mask |
| ink | `#1b1916` | node names, primary strokes |
| ink wash | `rgba(27,25,22,.02–.10)` | zone / store / external fills |
| muted | `#6f675d` | default arrows, sublabels, secondary strokes |
| faint | `#938b7f` | arrow labels, legend text |
| rule | `rgba(27,25,22,.10)` | hairlines, legend separator |
| accent | `#1c5563` | 1–2 focal elements max |
| accent tint | `#e3edef` | focal node fill |
| link | `#274d78` | HTTP/API calls, external-system arrows (baseline `--blue`) |

**Focal rule:** accent goes on 1–2 elements max. Everything else is ink / muted / faint. If
you're tempted to accent 4 things, you haven't decided what's focal yet.

### Node type → treatment

| Type | Fill | Stroke |
|---|---|---|
| **Focal** (1–2 max) | `#e3edef` | `#1c5563` |
| **Backend / API / Step** | `#fffdf8` | `#1b1916` |
| **Store / State** | `rgba(27,25,22,.05)` | `#6f675d` |
| **External / Cloud** | `rgba(27,25,22,.03)` | `rgba(27,25,22,.30)` |
| **Input / User** | `rgba(111,103,93,.10)` | `#938b7f` |
| **Optional / Async** | `rgba(27,25,22,.02)` | `rgba(27,25,22,.20)` dashed `4,3` |
| **Security / Boundary** | `rgba(28,85,99,.05)` | `rgba(28,85,99,.50)` dashed `4,4` |

### Typography (system stacks only — no webfonts)

| Element | Family | Size / weight |
|---|---|---|
| Node name | `-apple-system,'Segoe UI',Helvetica,Arial,sans-serif` | 12px / 600 |
| Sublabel (ports, URLs, types) | `ui-monospace,'SF Mono',Menlo,monospace` | 9px / 400 |
| Eyebrow / type tag / axis label | same mono, uppercase, `letter-spacing .08–.14em` | 7–8px |
| Arrow label | same mono, all-caps, ≤14 chars | 8px |
| Editorial aside (callouts only) | `Charter,Georgia,serif` *italic* | 14px |

The diagram's title is NOT inside the SVG — the report's own heading (Charter serif) owns it.
Mono is for technical content; human-readable names are sans. Never mono as a blanket "dev"
font.

## 4. Core primitives

### Embedding in the report

```html
<figure class="diagram">
  <svg viewBox="0 0 960 520" role="img" aria-labelledby="slug-title slug-desc">…</svg>
  <figcaption>One-line reading of the diagram.</figcaption>
</figure>
```

```css
.diagram{background:var(--paper); border:1px solid var(--line); border-radius:var(--r);
  padding:20px; overflow-x:auto}
.diagram svg{display:block; width:100%; height:auto; min-width:640px}
.diagram figcaption{font-family:ui-monospace,Menlo,monospace; font-size:12px;
  color:var(--muted); margin-top:10px}
```

`min-width` keeps a wide diagram legible on mobile — it scrolls inside its own container, the
page body never scrolls horizontally. Background: a single `<rect width="100%" height="100%"
fill="#fffdf8"/>` — no dot patterns, no secondary container inside the SVG.

### Arrow markers (define all three, always)

```svg
<marker id="arrow" markerWidth="8" markerHeight="6" refX="7" refY="3" orient="auto">
  <polygon points="0 0, 8 3, 0 6" fill="#6f675d"/>
</marker>
<marker id="arrow-accent" markerWidth="8" markerHeight="6" refX="7" refY="3" orient="auto">
  <polygon points="0 0, 8 3, 0 6" fill="#1c5563"/>
</marker>
<marker id="arrow-link" markerWidth="8" markerHeight="6" refX="7" refY="3" orient="auto">
  <polygon points="0 0, 8 3, 0 6" fill="#274d78"/>
</marker>
```

Default arrows muted; accent for the 1–2 headline flows; link-blue for HTTP/API/external;
`stroke-dasharray="5,4"` + lighter weight (`stroke-width="1"`) for optional / passive /
return / async. **Draw arrows before boxes** so z-order puts lines behind nodes
(bg → zones → arrows → nodes).

### Node box — full pattern

```svg
<!-- 1. Opaque paper mask — prevents arrows bleeding through translucent fills -->
<rect x="X" y="Y" width="W" height="H" rx="6" fill="#fffdf8"/>
<!-- 2. Styled box -->
<rect x="X" y="Y" width="W" height="H" rx="6" fill="FILL" stroke="STROKE" stroke-width="1"/>
<!-- 3. Rectangular type tag (rx=2, NOT a pill) -->
<rect x="X+8" y="Y+6" width="28" height="12" rx="2" fill="transparent"
      stroke="STROKE" stroke-opacity="0.4" stroke-width="0.8"/>
<text x="X+22" y="Y+15" fill="STROKE" fill-opacity="0.8" font-size="7"
      font-family="ui-monospace,Menlo,monospace" text-anchor="middle"
      letter-spacing="0.08em">API</text>
<!-- 4. Node name (sans — human-readable) -->
<text x="CX" y="CY+2" fill="#1b1916" font-size="12" font-weight="600"
      font-family="-apple-system,'Segoe UI',sans-serif" text-anchor="middle">Node Name</text>
<!-- 5. Technical sublabel (mono) -->
<text x="CX" y="CY+18" fill="#6f675d" font-size="9"
      font-family="ui-monospace,Menlo,monospace" text-anchor="middle">tech:port</text>
```

### Arrow labels — always mask, always with margin

Every arrow label needs an opaque paper rect behind it, and the label sits with a visible gap
off the line — never on top of it.

```svg
<!-- Mask sits 14px above the arrow (8px text + 6px gap). Stroke is at ARROW_Y. -->
<rect x="MID_X-18" y="ARROW_Y-20" width="36" height="12" rx="2" fill="#fffdf8"/>
<text x="MID_X" y="ARROW_Y-11" fill="#938b7f" font-size="8"
      font-family="ui-monospace,Menlo,monospace" text-anchor="middle"
      letter-spacing="0.06em">WRITE</text>
```

≤14 characters, all-caps, centered on the segment midpoint. Never vertical `writing-mode`;
for vertical segments place the label beside the line with the same 6–10px horizontal gap.

### Legend — horizontal strip at the bottom

Never inside the diagram area. A hairline separator, then a horizontal row after all nodes;
expand the `viewBox` height ~60px to fit it. Legend covers every treatment used — and nothing
extra.

```svg
<line x1="30" y1="LEGEND_Y-8" x2="VIEWBOX_W-30" y2="LEGEND_Y-8"
      stroke="rgba(27,25,22,0.10)" stroke-width="0.8"/>
<text x="30" y="LEGEND_Y+8" fill="#6f675d" font-size="8"
      font-family="ui-monospace,Menlo,monospace" letter-spacing="0.14em">LEGEND</text>
```

### Zone grouping (architecture and friends)

Group 2+ nodes sharing a tier/trust boundary with a zone rect, drawn before arrows and nodes:
fill `rgba(27,25,22,.02)`, stroke `rgba(27,25,22,.10)` at 0.8, `rx=8`; mono eyebrow label on
a paper mask at the zone's top edge, ≥16px clear of the first enclosed node. Max 3 zones —
more reads as a swimlane (use that type).

## 5. Mandatory connector rules

Non-negotiable; verify with the §9 checklist before shipping any diagram.

1. **Orthogonal connectors only.** Never a diagonal `<line>` between nodes that don't share an
   x or y axis. Every bend is a quarter-arc, `r=8` (`r=6` minimum in tight layouts). Two-bend
   elbow, `mid = (x1+x2)/2`:

   ```svg
   <!-- right+down: (x1,y1) → (x2,y2) -->
   <path d="M x1,y1 H mid-8 Q mid,y1 mid,y1+8 V y2-8 Q mid,y2 mid+8,y2 H x2"
         fill="none" stroke="#6f675d" stroke-width="1.2" marker-end="url(#arrow)"/>
   ```

   Flip the vertical signs for right+up. Plain `<line>` only when endpoints share an axis.
   **Port selection:** a mainly-vertical path exits/enters top/bottom edges with a single-bend
   L-path — entering a node's side face on a vertical path looks like a puncture. Reserve
   left/right ports for mainly-horizontal travel.

2. **Label-to-connector margin: 6–10px visible gap, always.** The opaque mask stops
   bleed-through; the gap between mask edge and stroke keeps the connector traceable. A label
   that hides its own arrow is a hard fail.

3. **No overlapping connectors.** Two connectors never share a stroke path or run on top of
   each other. Crossings get a bridge/hop on the LESS important arrow (never both):

   ```svg
   <!-- horizontal hop over a vertical crossing at x=cx, line at y -->
   <path d="M x1,y H cx-8 a 8,8 0 0,1 16,0 H x2" fill="none"
         stroke="#6f675d" stroke-width="1.2" marker-end="url(#arrow)"/>
   ```

   (For a vertical hop over a horizontal: `a 8,8 0 0,0 0,16`.) Parallel runs stay ≥12px apart
   end-to-end. If connectors keep wanting to stack, the layout is over budget — redesign.

4. **Shared edge → fan the attach points.** N connectors on an edge of length L attach at
   `L·k/(N+1)`, k=1..N, ≥12px apart (8px floor for very small boxes) — no two connectors share
   a point on a box, and no connector hides another.

5. **No transit behind non-endpoint boxes.** Reroute around intervening boxes. Sole exception:
   a cross-cutting bar geometrically unavoidable on the only direct orthogonal path — then the
   stroke is dashed (`4,3`, "transit, not interaction"), the label sits at the visible end, and
   the arrowhead resolves only at the true destination. When in doubt, reroute.

6. **A label mask must not overlap a node drawn after it.** Nodes paint after labels: a mask
   partly inside a node gets clipped by the node fill and the text renders as a fragment on the
   border. Place labels on connector segments crossing open canvas. A mask fully inside a node
   is a badge chip (fine); over a zone is fine (zones paint first).

Dashed paths follow the same routing, port, and bridge rules — the dash communicates semantic
weight, not a different grammar. When a dashed and a solid path cross, bridge the dashed one.

## 6. Layout & spacing

**4px grid — coordinates, dimensions, gaps, padding all divisible by 4.** Node widths from
{80, 96, 112, 128, 144, 160, 180, 200, 240, 320}; gaps between nodes from {20, 24, 32, 40,
48}; box padding {8, 12, 16}; radius {4, 6, 8}. Exempt: stroke widths (0.8, 1, 1.2),
opacities, and mono micro-type (7–9px per the §3 scale). Quick check: a coordinate ending in
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
| Diagonal / slanted connectors | orthogonal elbows are mandatory (§5.1) |
| Two connectors sharing a path or attach point | each must be independently traceable |
| Shadows on nodes | borders are the elevation language here |
| `rx` > 10 on boxes | radius 4–8, or none |
| Bidirectional arrow when one direction is obvious | says nothing |
| 3 identical summary cards under the diagram | vary widths (`1.1fr 1fr 0.9fr`) |

## 8. Accessible SVG contract

1. `<svg>` carries `role="img"` and `aria-labelledby` naming its `<title>` and `<desc>`.
2. `<title>` is the FIRST child of `<svg>`, before `<defs>`; both `<title>` and `<desc>` filled.
3. IDs are prefixed per diagram (`<slug>-title` / `<slug>-desc`) — bare `title`/`desc` IDs
   collide when a report embeds two diagrams.
4. `<title>` ≈ the figure's short name (≤60 chars). `<desc>` is one sentence describing the
   CONTENT, not the geometry ("Architecture of X routing writes through Y", never "a box with
   five boxes below it").
5. Purely decorative SVG gets `aria-hidden="true"` instead.

## 9. Pre-output checklist

Run before shipping any diagram in a report:

- [ ] Would a table or paragraph do the same job? (If yes — don't draw.)
- [ ] Right type per §2? Within the §6 budget?
- [ ] Remove test passed: no removable node, mergeable pair, redundant arrow, or redundant label?
- [ ] Accent on ≤2 elements? Legend covers exactly the treatments used?
- [ ] Arrows drawn before boxes? All elbows orthogonal with `r=8`, no diagonals?
- [ ] Every arrow label masked in paper `#fffdf8` with a visible 6–10px gap off its line?
- [ ] No overlapping connectors; crossings bridged; shared edges fanned ≥12px?
- [ ] No connector transits behind a non-endpoint box (or the dashed-transit exception applies)?
- [ ] No label mask overlaps a node painted after it?
- [ ] Legend is a bottom strip; `viewBox` expanded ~60px for it?
- [ ] Coordinates/dimensions on the 4px grid?
- [ ] Node names sans 12px/600; technical text mono; no webfont anywhere in the file?
- [ ] `role="img"`, first-child `<title>`, prefixed `<slug>-title`/`<slug>-desc` IDs?
- [ ] Colors match §3 (report tokens) — no upstream tangerine, no foreign palette?
- [ ] Deterministic checks clean: `python3 <skill-dir>/scripts/verify_geometry.py <file>` and
      `python3 <skill-dir>/scripts/self_check.py <file>` (both adapted from upstream; exit
      non-zero on failure)?

The complete MIT notice is in [diagram-design license](license-diagram-design.md).
