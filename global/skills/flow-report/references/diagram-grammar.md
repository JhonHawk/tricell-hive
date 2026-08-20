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

## 3. Skin — color lives in CSS classes, never in attributes

**Never put a color in an SVG presentation attribute.** `fill="#fffdf8"` cannot follow the
report's theme toggle and cannot invert for print — a diagram written that way renders as a
white slab inside a dark page, and prints a black one. Put every color on a class defined in
the report's `<style>`, where `var()` resolves.

```svg
<rect class="d-node" x="20" y="62" width="170" height="66" rx="8"/>   <!-- yes -->
<rect fill="#fffdf8" stroke="#1b1916" .../>                          <!-- never -->
```

Geometry attributes (`x`, `width`, `rx`, `stroke-dasharray` used for *meaning*) stay on the
element. Only color moves to CSS.

| Class | Role |
|---|---|
| `d-bg` | the SVG's background rect — matches the page, so the figure has no edge of its own |
| `d-node` | default node: no fill, ink stroke |
| `d-focal` | the 1–2 focal nodes: accent tint fill, accent stroke |
| `d-store` | store / state: muted stroke |
| `d-ext` | external / third-party: faint stroke, `stroke-dasharray:4 3` |
| `d-arrow` / `d-arrow-accent` | connectors, default and headline |
| `d-label` | node name (sans) |
| `d-sub` | technical sublabel — ports, URLs, types (mono) |
| `d-alabel` | arrow label, legend text (mono, uppercase) |
| `d-mask` | opaque page-colored rect behind a label or under a node |
| `d-tag` / `d-tagtext` | the rectangular type tag on a node |
| `d-rule` / `d-legend-rule` | hairlines and the legend separator |

The class definitions ship in every archetype skeleton; copy them with the skeleton rather
than re-deriving them. **Focal rule:** accent goes on 1–2 elements max. Everything else is
ink / muted / faint. If you are tempted to accent four things, you have not decided what is
focal yet.

### Node type → treatment

This system draws nodes as **wireframe**: stroke carries the meaning, fill is reserved for the
focal node alone. A grid of filled boxes flattens the hierarchy the stroke was carrying.

| Type | Class | Treatment |
|---|---|---|
| **Focal** (1–2 max) | `d-focal` | accent tint fill + accent stroke |
| **Backend / API / Step** | `d-node` | no fill, ink stroke |
| **Store / State** | `d-store` | no fill, muted stroke |
| **External / Cloud** | `d-ext` | no fill, faint dashed stroke |
| **Optional / Async** | `d-ext` | same, on the connector as well |

A node that must sit on top of a connector gets a `d-mask` rect underneath it — that is what
the mask is for, now that nodes are unfilled.

### Typography (system stacks only — no webfonts)

Sizes stay on the class, not on the element. Node names are sans; anything technical is mono.
Never mono as a blanket "dev" font.

| Element | Class | Size / weight |
|---|---|---|
| Node name | `d-label` | 12px / 600, sans |
| Sublabel (ports, URLs, types) | `d-sub` | 9px, mono |
| Type tag on a node | `d-tagtext` | 7px, mono, uppercase, `letter-spacing .08em` |
| Arrow label · legend | `d-alabel` | 8px, mono, uppercase, ≤14 chars |

The diagram's title is NOT inside the SVG — the report's own heading owns it.

### Motion

**No scroll-in reveal ships with this system, and adding one is not a free win.** A stroke
trace needs a dash length per shape and a single value cannot fit both a 16px legend swatch
and a 600px connector; a `clip-path` wipe does not interpolate reliably between percentage
and unitless insets. Both failure modes render a *blank* diagram — strictly worse than no
animation. If motion is added later it must be opt-in from JS over a figure that is already
visible, and it must honor `prefers-reduced-motion` and `beforeprint`.

## 4. Core primitives

### Embedding in the report

The frame belongs to the report's chrome layer, and the scroll container must be an **inner**
element: a `.framed` that also carries `overflow` clips its own `[ TITLE ]` label.

```html
<figure class="diagram framed">
  <span class="frame-t">Flow</span>
  <div class="scroll-x">
    <svg viewBox="0 0 960 520" role="img" aria-labelledby="slug-title slug-desc">…</svg>
  </div>
  <figcaption>One-line reading of the diagram.</figcaption>
</figure>
```

`.diagram svg` keeps `min-width:640px` so a wide diagram stays legible on mobile — it scrolls
inside `.scroll-x`, and the page body never scrolls horizontally. Background: a single
`<rect class="d-bg" width="100%" height="100%"/>` — no dot patterns, no secondary container
inside the SVG.

### Arrow markers — one marker, `context-stroke`

Do not define one marker per color. `fill="context-stroke"` makes a single marker inherit the
stroke of whatever path uses it, so arrowheads follow the theme for free.

```svg
<marker id="ah" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7"
        orient="auto-start-reverse">
  <path d="M0,0 L10,5 L0,10 z" fill="context-stroke"/>
</marker>
```

Default arrows `d-arrow`; `d-arrow-accent` for the 1–2 headline flows; `stroke-dasharray="5,4"`
and a lighter weight for optional / passive / return / async. **Draw arrows before nodes** so
z-order puts lines behind boxes (bg → zones → arrows → nodes).

### Node box — full pattern

```svg
<!-- 1. Mask, ONLY where the node sits on top of a connector -->
<rect class="d-mask" x="X" y="Y" width="W" height="H" rx="8"/>
<!-- 2. The node itself (wireframe; d-focal for the 1-2 focal ones) -->
<rect class="d-node" x="X" y="Y" width="W" height="H" rx="8"/>
<!-- 3. Type tag STRADDLING the top edge, on its own mask (rx=2, NOT a pill) -->
<rect class="d-mask" x="X+5" y="Y-6" width="34" height="12"/>
<rect class="d-tag ink" x="X+8" y="Y-6" width="28" height="12" rx="2"/>
<text class="d-tagtext ink" x="X+22" y="Y+3" text-anchor="middle">API</text>
<!-- 4-5. Name + sublabel, optically centred on the WHOLE box -->
<text class="d-label" x="CX" y="CY-2" text-anchor="middle">Node Name</text>
<text class="d-sub"   x="CX" y="CY+13" text-anchor="middle">tech:port</text>
```

**The tag rides the top edge, it does not sit inside the box.** Placed inside, it steals the
top third and pushes the name/sublabel block below centre — the node then reads as
top-heavy even though the numbers look symmetric. Straddling the edge (the same move the
report's `[ FRAME TITLE ]` makes) frees the interior so the two text lines centre on the full
height: `CY-2` and `CY+13`.

Drop the `ink` modifier on the tag classes when the node is `d-focal` — the tag then picks up
the accent instead of the ink.

### Arrow labels — always mask, always with margin

Every arrow label needs an opaque `d-mask` rect behind it, and the label sits with a visible
gap off the line — never on top of it.

```svg
<!-- Mask sits 14px above the arrow (8px text + 6px gap). Stroke is at ARROW_Y. -->
<rect class="d-mask" x="MID_X-18" y="ARROW_Y-20" width="36" height="12" rx="2"/>
<text class="d-alabel" x="MID_X" y="ARROW_Y-11" text-anchor="middle">WRITE</text>
```

≤14 characters, all-caps, centered on the segment midpoint. Never vertical `writing-mode`;
for vertical segments place the label beside the line with the same 6–10px horizontal gap.

### Legend — horizontal strip at the bottom

Never inside the diagram area. A hairline separator, then a horizontal row after all nodes;
expand the `viewBox` height ~60px to fit it. Legend covers every treatment used — and nothing
extra.

```svg
<line class="d-legend-rule" x1="30" y1="LEGEND_Y-8" x2="VIEWBOX_W-30" y2="LEGEND_Y-8"/>
<text class="d-alabel" x="30" y="LEGEND_Y+8">LEGEND</text>
```

### Zone grouping (architecture and friends)

Group 2+ nodes sharing a tier/trust boundary with a `d-ext` zone rect, drawn before arrows and
nodes, `rx=8`; mono eyebrow label on a `d-mask` at the zone's top edge, ≥16px clear of the
first enclosed node. Max 3 zones — more reads as a swimlane (use that type).

## 5. Mandatory connector rules

Non-negotiable; verify with the §9 checklist before shipping any diagram.

1. **Orthogonal connectors only.** Never a diagonal `<line>` between nodes that don't share an
   x or y axis. Every bend is a quarter-arc, `r=8` (`r=6` minimum in tight layouts). Two-bend
   elbow, `mid = (x1+x2)/2`:

   ```svg
   <!-- right+down: (x1,y1) → (x2,y2) -->
   <path class="d-arrow" d="M x1,y1 H mid-8 Q mid,y1 mid,y1+8 V y2-8 Q mid,y2 mid+8,y2 H x2"
         marker-end="url(#ah)"/>
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
   <path class="d-arrow" d="M x1,y H cx-8 a 8,8 0 0,1 16,0 H x2" marker-end="url(#ah)"/>
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
- [ ] Every arrow label masked with `d-mask` and a visible 6–10px gap off its line?
- [ ] No overlapping connectors; crossings bridged; shared edges fanned ≥12px?
- [ ] No connector transits behind a non-endpoint box (or the dashed-transit exception applies)?
- [ ] No label mask overlaps a node painted after it?
- [ ] Legend is a bottom strip; `viewBox` expanded ~60px for it?
- [ ] Coordinates/dimensions on the 4px grid?
- [ ] Node names sans 12px/600; technical text mono; no webfont anywhere in the file?
- [ ] Nodes wireframe (fill reserved for the 1–2 focal ones), masks only where a node sits on a connector?
- [ ] `role="img"`, first-child `<title>`, prefixed `<slug>-title`/`<slug>-desc` IDs?
- [ ] **Zero color in presentation attributes** — every fill/stroke comes from a `d-*` class (§3)?
- [ ] Checked in BOTH themes and in print preview — no white slab, no invisible stroke?
- [ ] Deterministic checks clean: `python3 <skill-dir>/scripts/verify_geometry.py <file>` and
      `python3 <skill-dir>/scripts/self_check.py <file>` (both adapted from upstream; exit
      non-zero on failure)?

The complete MIT notice is in [diagram-design license](license-diagram-design.md).
