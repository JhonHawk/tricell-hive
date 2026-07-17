
> **Scope & activation.** Distilled, verifiable visual-design criteria (from *Refactoring UI*,
> Wathan & Schoger) to apply WHILE building UI — front-load them, don't bolt them on at the end.
> Cross-stack (React/Angular/Vue/Svelte/HTML+CSS). These globs sometimes match a non-UI or
> non-visual file — apply judgment, exit silently when irrelevant. **Boundaries:** the spacing/
> sizing *scale* itself is owned by `tailwind.md` (use the canonical scale, no arbitrary `[...]`);
> component-level taste is owned by the loaded design-system skill (HeroUI/shadcn); flow-level UX
> (orientation, error recovery, empty-state *invitation*) is owned by the UX rubric
> (`flow-core/references/ux-rubric.md`). The
> numeric values below are starting criteria to map onto the project's design tokens — not literals
> to hardcode.

## Hierarchy
- Encode hierarchy with size **and** weight **and** color together — never size alone. De-emphasize secondary content with a softer color, not a lighter weight.
- Limit UI text to ~2–3 colors (dark = primary, grey = secondary, lighter grey = tertiary) and ~2 weights (400/500 normal, 600/700 emphasis). Never weight < 400 for small text.
- Style actions by prominence, not semantics: **one primary** (solid, high-contrast fill) / secondary (outline or low-contrast fill) / tertiary (link-styled). One primary action per view.
- A destructive action is not automatically big/red — give it secondary/tertiary weight unless deleting *is* the view's primary action (e.g. the confirmation step).
- Decouple semantic level from visual size: pick `h1`–`h6` for semantics, style for hierarchy. Section titles often act as labels — make them small (≈16px), not large.
- Grey text on a colored background washes out; white-at-opacity bleeds. Pick a text color sharing the background's **hue**, then adjust saturation/lightness for lower contrast.
- Balance weight vs contrast: soften a heavy element (solid icon) with a lower-contrast color instead of resizing; widen a too-subtle 1px border before darkening it (darkening reads harsh).
- Labels are a last resort: drop the label when the format implies it (email, price) or fold it into the value (`3 bedrooms`); when kept, make the label the secondary element and let the value dominate — except on scan-for-the-label spec tables.

## Spacing & layout
- **Outer space > inner space, always.** Within a group, elements sit closer to each other than to neighboring groups (a label nearer its own input than the previous field; a heading nearer the text it introduces than the section above). Equal/ambiguous spacing is a bug; aim ~1.5–3× outer:inner.
- Start with too much whitespace and remove, rather than adding until it stops looking cramped. Dense/compact layouts are a deliberate exception, not the default.
- Don't stretch content to fill the viewport — give each element only the width it needs; split over-wide content into columns instead of widening it.
- Use fixed or `max-width` for elements that shouldn't scale (sidebars, cards, avatars); reserve percentage/fluid widths for things you actually want to scale. Adjacent scale steps stay ≥ ~25% apart (the canonical scale already embodies this).
- Don't couple a component's sizes through relative units (`em`): large elements must shrink faster than small ones across breakpoints, and font-size vs padding are tuned independently (padding gets proportionally tighter at small sizes) — not scaled together.

## Typography
- Set the type scale in **px/rem from a small hand-picked set** (e.g. 12/14/16/18/20/24/30/36/48/60/72). Avoid `em` for the scale (nested `em` compounds off-scale) and modular/ratio scales (fractional px).
- Body measure **45–75 characters per line** (~20–35em); cap with `max-width` even inside a wider container — never `max-width: none` on prose.
- Line-height is proportional to measure and inverse to size: body **1.5–2** (taller for wider columns), large headings **≈1**.
- Align mixed font sizes by **baseline** (`align-items: baseline`), not center.
- Right-align numeric table columns; center only headings or blocks ≤ 2–3 lines; if text is justified, set `hyphens: auto`.
- Letter-spacing: leave default; tighten large/headline faces ≈ `-0.05em`, widen all-caps ≈ `+0.05em`.

## Color
- Author colors in **HSL** (or OKLCH), not hex/RGB — so visually related colors stay related in code.
- Define the full palette up front, not ~5 ad-hoc values: **8–10 grey shades**, **1–2 primaries with 5–10 shades each**, plus accent hues for semantic states (red/yellow/green) with several shades. Use a fixed scale named **100 (lightest) → 500 (base) → 900 (darkest)**.
- Never generate shades at runtime (`lighten()`/`darken()`/opacity) — pick fixed shades; don't invent new ad-hoc ones.
- Compensate saturation as lightness leaves 50%: raise saturation toward the light and dark ends so extreme shades don't wash out.
- To shift brightness without washing out, rotate hue toward the nearest **bright hue (60/180/300°)** to lighten or **dark hue (0/120/240°)** to darken; cap rotation ≤ ~20–30°.
- Greys carry temperature: cool ≈ hue 200–210, warm ≈ hue 40, at low saturation (~12–21%); keep one temperature and bump saturation at the extreme shades. Start the darkest grey from very-dark-grey, not true black.
- Meet **WCAG AA contrast: ≥ 4.5:1 normal text, ≥ 3:1 large text** (≥ ~18px or bold). For colored badges/labels, prefer dark text on a light tint of the same hue over white-on-saturated.
- Never encode meaning by color alone — pair it with an icon/label/shape. For multi-series charts, prefer one hue light→dark over many distinct hues.

## Depth & shadows
- Model one light source from **directly above**: raised elements get a subtle light top edge + a dark drop shadow below; inset elements get a dark top inner shadow + a light bottom edge. Pick the lighter highlight color by hand — not transparent white (it desaturates).
- Treat elevation as a **fixed shadow scale (~5 steps)**: larger/softer shadow = higher/closer (button < dropdown < modal). Raise the shadow on drag, shrink it on press.
- Build a shadow from **two layers** — a larger soft cast plus a tighter darker ambient shadow; let the tight one fade as elevation rises.
- Flat depth without blur: lighter = closer, darker = further; use solid shadows (small vertical offset, **zero blur**).
- When elements overlap across a background boundary, give an overlapping image an "invisible border" matching the background so edges don't clash.

## Images
- Text over a photo needs consistent contrast — reduce the image's dynamic range first (a semi-transparent overlay, lowered image contrast, a colorize/`multiply` tint, or a large-blur zero-offset text-shadow). Don't just pick a text color.
- Respect each asset's intended size: don't scale a 16–24px icon up 3–4× (chunky, detail-poor) — use an icon drawn for the size, or wrap the small icon in a larger filled shape. Don't shrink a full screenshot to illegibility — recapture at a smaller layout, crop, or redraw simplified.
- Constrain user-uploaded images to a fixed container with `background-size: cover` (center + crop), never intrinsic aspect ratio in a grid; prevent edge bleed with an inner shadow, not a border.

## Borders & finishing touches
- Reach for a border **last**. To separate elements prefer, in order: a box-shadow, a second background color, or extra spacing. If you already use background colors plus a border, drop the border.
- Add an accent border (a colored rectangle — card top, active-nav underline, alert left edge, headline underline, layout top) to inject polish cheaply.
- Treat empty backgrounds: a background-color change, a slight gradient (**two hues ≤ ~30° apart**), or a low-contrast pattern/shape.
- Empty states — visual-craft layer (the flow-level *invitation* is the UX rubric's): hide tabs/filters/search that do nothing until content exists, and emphasize the single first-action CTA.
- Supercharge defaults instead of adding chrome: bullets → contextual icons; default checkboxes/radios → brand-colored selected states.
- Components aren't locked to their stereotype: merge related non-sortable table columns into one primary/secondary cell (avatar + name + meta + status pill); turn an important radio group into selectable cards.

## Component simplicity
- Prefer the project's design-system component (HeroUI/shadcn) over hand-rolling one it already provides.
- Fewer variants and fewer "just in case" config props — add a variant when a real second use exists, not preemptively.
