# Diagram SVG construction — skin, primitives, connector rules, accessibility

## When to read

Read before writing or changing SVG markup in a report: the CSS-class skin (§3), the embedding, marker, node, label, legend and zone primitives (§4), the mandatory connector rules (§5) and the accessible SVG contract (§8). When to draw, type selection, the layout budget, anti-patterns and the pre-output checklist are in [the diagram grammar](diagram-grammar.md) (§1, §2, §6, §7, §9), which shares this numbering.

> Adapted from [cathrynlavery/diagram-design](https://github.com/cathrynlavery/diagram-design)
> (v2.3.2, MIT) and re-skinned onto the create-report baseline tokens (`baseline.html`). It covers
> static inline SVG only: no Mermaid/draw.io import, animation, or PNG/SVG export.

## 3. Skin — color lives in CSS classes, never in attributes

A color in an SVG presentation attribute cannot follow the report's theme toggle or invert for print: the diagram renders as a white slab in a dark page. Put every color on a class defined in the report's `<style>`, where `var()` resolves; geometry (`x`, `width`, `rx`, a `stroke-dasharray` used for meaning) stays on the element.

```svg
<rect class="d-node" x="20" y="62" width="170" height="66" rx="8"/>   <!-- yes -->
<rect fill="#fffdf8" stroke="#1b1916" .../>                          <!-- never -->
```

| Class | Role |
|---|---|
| `d-bg` | the SVG's background rect, matching the page |
| `d-node` | default node: no fill, ink stroke |
| `d-focal` | the 1–2 focal nodes: accent tint fill, accent stroke |
| `d-store` | store / state: muted stroke |
| `d-ext` | external / optional / async: faint stroke, `stroke-dasharray:4 3` (also on the connector) |
| `d-arrow` / `d-arrow-accent` | connectors, default and headline |
| `d-label` | node name (sans, 12px / 600) |
| `d-sub` | technical sublabel: ports, URLs, types (mono, 9px) |
| `d-alabel` | arrow label, legend text (mono, 8px, uppercase, ≤14 chars) |
| `d-mask` | opaque page-colored rect behind a label or under a node |
| `d-tag` / `d-tagtext` | the rectangular type tag on a node (7px mono uppercase, `letter-spacing .08em`) |
| `d-rule` / `d-legend-rule` | hairlines and the legend separator |

The class definitions ship in every archetype skeleton: copy them with it. Nodes are wireframe: stroke carries the meaning and fill is reserved for the 1–2 focal nodes, since a grid of filled boxes flattens the hierarchy. A node that sits on a connector gets a `d-mask` rect underneath, because an unfilled node would show the connector through it. Sizes stay on the class, never on the element; the diagram's title is not inside the SVG, the report's heading owns it.

### Motion

No scroll-in reveal ships with this system: a stroke trace or `clip-path` wipe can render a blank diagram. Motion added later must be opt-in from JS over a figure that is already visible, and honor `prefers-reduced-motion` and `beforeprint`.

## 4. Core primitives

### Embedding in the report

The scroll container must be an inner element: a `.framed` that also carries `overflow` clips its own `[ TITLE ]` label.

```html
<figure class="diagram framed">
  <span class="frame-t">Flow</span>
  <div class="scroll-x">
    <svg viewBox="0 0 960 520" role="img" aria-labelledby="slug-title slug-desc">…</svg>
  </div>
  <figcaption>One-line reading of the diagram.</figcaption>
</figure>
```

`.diagram svg` keeps `min-width:640px` and scrolls inside `.scroll-x`, so the page never scrolls horizontally. Background: a single `<rect class="d-bg" width="100%" height="100%"/>`, no dot patterns.

### Arrow markers — one marker, `context-stroke`

One marker serves every color, so arrowheads follow the theme:

```svg
<marker id="ah" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7"
        orient="auto-start-reverse">
  <path d="M0,0 L10,5 L0,10 z" fill="context-stroke"/>
</marker>
```

Default arrows `d-arrow`; `d-arrow-accent` for the 1–2 headline flows; `stroke-dasharray="5,4"` and a lighter weight for optional, passive, return or async. **Draw arrows before nodes**: bg → zones → arrows → nodes.

### Node box — full pattern

```svg
<!-- 1. Mask, ONLY where the node sits on top of a connector -->
<rect class="d-mask" x="X" y="Y" width="W" height="H" rx="8"/>
<!-- 2. The node itself (d-focal for the 1-2 focal ones) -->
<rect class="d-node" x="X" y="Y" width="W" height="H" rx="8"/>
<!-- 3. Type tag STRADDLING the top edge, on its own mask (rx=2, NOT a pill) -->
<rect class="d-mask" x="X+5" y="Y-6" width="34" height="12"/>
<rect class="d-tag ink" x="X+8" y="Y-6" width="28" height="12" rx="2"/>
<text class="d-tagtext ink" x="X+22" y="Y+3" text-anchor="middle">API</text>
<!-- 4-5. Name + sublabel, optically centred on the WHOLE box -->
<text class="d-label" x="CX" y="CY-2" text-anchor="middle">Node Name</text>
<text class="d-sub"   x="CX" y="CY+13" text-anchor="middle">tech:port</text>
```

The tag rides the top edge: inside the box it steals the top third and the text block reads top-heavy. Drop the `ink` modifier when the node is `d-focal`.

### Arrow labels, legend, zones

Every arrow label gets an opaque `d-mask` rect behind it (about 14px above the arrow: 8px text plus a 6px gap) and a visible gap off the line, never on top of it, centered on the segment midpoint; for vertical segments the label sits beside the line with the same 6–10px gap. The legend is a hairline separator (`d-legend-rule`) and a horizontal row of `d-alabel` text at the bottom, after all nodes, covering every treatment used and nothing extra; expand the `viewBox` height ~60px. Group 2+ nodes sharing a tier or trust boundary with a `d-ext` zone rect (`rx=8`) drawn before arrows and nodes, with a mono eyebrow label on a `d-mask` at the zone's top edge, ≥16px clear of the first enclosed node; max 3 zones, since more reads as a swimlane.

## 5. Mandatory connector rules

Verify with the §9 checklist in the diagram grammar before shipping.

1. **Orthogonal connectors only.** No diagonal `<line>` between nodes that do not share an x or y axis. Every bend is a quarter-arc, `r=8` (`r=6` minimum in tight layouts). Two-bend elbow, `mid = (x1+x2)/2`:

   ```svg
   <!-- right+down: (x1,y1) → (x2,y2) -->
   <path class="d-arrow" d="M x1,y1 H mid-8 Q mid,y1 mid,y1+8 V y2-8 Q mid,y2 mid+8,y2 H x2"
         marker-end="url(#ah)"/>
   ```

   Flip the vertical signs for right+up; plain `<line>` only when endpoints share an axis. A mainly-vertical path exits and enters the top and bottom edges with a single-bend L-path (entering a side face looks like a puncture); reserve left and right ports for mainly-horizontal travel.

2. **Label-to-connector margin: 6–10px visible gap.** The mask stops bleed-through and the gap keeps the connector traceable; a label that hides its own arrow is a hard fail.

3. **No overlapping connectors.** Two connectors never share a stroke path. Crossings get a hop on the LESS important arrow, never both; parallel runs stay ≥12px apart. If connectors keep wanting to stack, the layout is over budget.

   ```svg
   <!-- horizontal hop over a vertical crossing at x=cx, line at y -->
   <path class="d-arrow" d="M x1,y H cx-8 a 8,8 0 0,1 16,0 H x2" marker-end="url(#ah)"/>
   ```

   (Vertical hop over a horizontal: `a 8,8 0 0,0 0,16`.)

4. **Shared edge → fan the attach points.** N connectors on an edge of length L attach at `L·k/(N+1)`, k=1..N, ≥12px apart (8px floor on very small boxes).

5. **No transit behind non-endpoint boxes.** Reroute around them. Sole exception: a cross-cutting bar on the only direct orthogonal path, where the stroke is dashed (`4,3`), the label sits at the visible end and the arrowhead resolves only at the true destination.

6. **A label mask must not overlap a node drawn after it:** the node fill clips the mask and the text renders as a fragment on the border. Place labels on segments crossing open canvas; a mask fully inside a node is a badge chip, and one over a zone is fine.

Dashed paths follow the same routing, port and bridge rules; when a dashed and a solid path cross, bridge the dashed one.

## 8. Accessible SVG contract

1. `<svg>` carries `role="img"` and `aria-labelledby` naming its `<title>` and `<desc>`.
2. `<title>` is the FIRST child of `<svg>`, before `<defs>`; both `<title>` and `<desc>` are filled.
3. IDs are prefixed per diagram (`<slug>-title` / `<slug>-desc`), since bare IDs collide when a report embeds two diagrams.
4. `<title>` is the figure's short name (≤60 chars); `<desc>` is one sentence describing the content, not the geometry.
5. Purely decorative SVG gets `aria-hidden="true"` instead.

The complete MIT notice is in [diagram-design license](license-diagram-design.md).
