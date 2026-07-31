# UX rubric

Checked per flow by `ux-flow-reviewer` while navigating the live mock or QA deployment.
Plain markdown so any harness can consume it. **Pass/fail per dimension, per flow** — both tables.
Two axes: **Flow** (does the journey work?) and **Visual craft** (is it built to the design
criteria?). Craft derives from `languages/ui-visual-design.md` (the implementer's front-loaded
criteria); this rubric is the reviewer's verification of the same principles. A failed dimension on
either axis is a finding.

## Flow

| # | Dimension | Pass means |
|---|---|---|
| 1 | Orientation | At every step the user can tell where they are (titles, breadcrumbs, active nav) |
| 2 | Next step | The primary action on each screen is visually obvious and singular |
| 3 | Feedback | Every action produces visible confirmation or progress within the same view |
| 4 | Error recovery | Error states say what went wrong AND how to fix it — never a dead end |
| 5 | Empty states | Zero-data views invite the correct first action instead of showing a void |
| 6 | Destructive friction | Destructive/sensitive actions state consequences and require deliberate confirmation |
| 7 | Hierarchy | The layout prioritizes the decision the user came to make; secondary info recedes |
| 8 | Copy | Labels and messages use the user's vocabulary and state consequences, not internals |
| 9 | Role coherence | The flow works for each role/tenant named in the spec — no leakage, no missing affordances |
| 10 | Responsive & a11y floor | Usable at the spec's target widths; focus order, labels, and contrast not obviously broken |

## Visual craft

| # | Dimension | Pass means |
|---|---|---|
| 11 | Type scale | Font sizes come from a small consistent scale (px/rem, not `em`); body measure is 45–75 chars; line-height ~1.5–2 for body, ~1 for large headings |
| 12 | Spacing system | Spacing comes from the scale; **space around a group exceeds space within it** (no ambiguous equal spacing); whitespace is generous, not cramped |
| 13 | Color & contrast | Palette is systematic (greys + 1–2 primaries + semantic accents, fixed shades); text meets **WCAG AA — ≥4.5:1 normal, ≥3:1 large**; meaning is never carried by color alone |
| 14 | Action hierarchy | Per view exactly one primary action (solid); secondary = outline/low-contrast; tertiary = link-styled; a destructive action isn't auto-styled big/red unless it's the view's primary |
| 15 | Elevation/shadows | Shadows form a consistent elevation scale (button < dropdown < modal), one light source from above — not arbitrary per element |
| 16 | Borders restraint | Separation prefers spacing / background / shadow over borders; no redundant border on top of a background-color change; consistent border-radius personality |
| 17 | Component simplicity | Uses the design system's components over hand-rolled ones; no gratuitous variants or "just in case" configurability |
| 18 | Net improvement *(changes to an existing screen only)* | Judged against the pre-change capture at the same viewport and theme (the desktop baseline in `tools/browser-automation.md`, set explicitly — never inherited): nothing that read as resolved before now reads as unresolved — no region emptied without a replacement, no element left without a surface, no heading demoted out of the hierarchy. A screen that only lost content did not improve. No pre-change capture → the dimension is **not-verified**, never a pass |

Dimensions 1–17 are conformance checks: an empty screen passes most of them, and some (overflow, contrast) pass *better* the emptier it gets. 18 is the only comparative one — it is what a redesign is actually judged on.
