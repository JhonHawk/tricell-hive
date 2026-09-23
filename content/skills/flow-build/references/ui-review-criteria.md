# UI review criteria

Read when reviewing a rendered UI change, or when briefing the reviewer. [Verification](verification.md) decides when the review is required; [browser automation](browser-automation.md) governs how the browser is driven. Project guides and accepted patterns refine these criteria; they never excuse a failed check.

## Stance

- Review independently of the implementer. The verdict starts as not passed and becomes a pass only when every applicable check has a recorded result.
- An observed defect keeps its severity unless a written rule below lowers it; cite that rule when lowering. Do not talk a finding down because it looks minor, matches a sibling screen, or was not in the plan.
- The parent supplies scope, data, and viewports but does not pre-rate findings or say what to ignore.
- Judge from measurement and interaction. A capture locates a problem; computed geometry, styles, and behavior decide it. Vision judgments of wrapping, overlap, and alignment are unreliable on their own.

## Prepare the comparison

1. Inventory each affected screen as a whole: regions, columns, rows, and components, including those the change only displaces.
2. For a change to an existing screen, capture and measure the base revision in the same environment, data, viewport, and theme. Without that baseline, the comparison is not verified, and a defect cannot be classified as pre-existing.
3. Use realistic and extreme content: the longest real values, text 30–50% longer, empty and missing values, one row, and many rows.
4. Set viewports explicitly: 1440 and 1280 desktop widths, a mobile width such as 390 when the application supports mobile, and further widths where the layout changes. For pages other than data tables, also check 320 CSS px for reflow. When the change contains dialogs or fixed or sticky regions, also check a short desktop height such as 1280×720. Check each supported theme.

## Checks

Apply the checks relevant to the change at every assigned viewport and theme. Each check names its measurement.

**Layout and content integrity**

- No word broken mid-word in names, identifiers, or headers: a word's text range returns client rects on more than one line.
- Short values stay on one line: dates, statuses, badges, amounts, identifiers, and button labels. The rendered line count is 1.
- No silent clipping: when `scrollWidth > clientWidth` under hidden overflow, an ellipsis is shown and the full value is available, such as in a tooltip or accessible name. Names and primary identifiers are not truncated.
- No page-level horizontal scroll (`documentElement.scrollWidth <= innerWidth`) unless a contained, reachable scroll region is the intended pattern.
- No overlap between elements: bounding boxes of distinct content do not intersect.
- Fixed and sticky regions, such as dialog headers and footers or pinned bars, do not cover content: scrolled to each end, the last and first content boxes stay clear of them, and content taller than the space scrolls inside its container.
- No regression in neighbors: compared with the baseline, existing columns do not narrow into wrapping, rows do not grow from new line breaks, and regions do not empty without a replacement.

**Hierarchy and consistency**

- Text is left-aligned, and quantities are right-aligned with tabular figures; each header aligns with its column.
- Type sizes, spacing, radii, and colors come from the project's scale or tokens; space around a group exceeds space within it.
- Each view has one primary action; secondary and destructive actions follow the project's hierarchy.
- Components in scope follow their library's documented anatomy, checked against its current documentation when available; for example, a checkbox or radio control sits on its label's row with the description beneath.
- The screen conforms to its accepted pattern. A sibling screen is a valid reference only for checks it passes itself; a shared defect is a systemic finding, not a defense.

**States**

- Loading, empty, error, disabled, hover, focus, selected, and success states exist where relevant and are captured. Loading does not shift the final layout.
- Errors state what went wrong and how to recover; empty states lead to the correct first action.

**Flow and feedback**

- The user can tell where they are and what to do next; every action gives visible feedback in the same view.
- Destructive or sensitive actions state consequences and require deliberate confirmation.
- Labels and messages use the user's vocabulary; each role in scope sees the right data and affordances.
- The console shows no new errors during the walk.

**Accessibility (WCAG 2.2 AA)**

- Text contrast is at least 4.5:1, or 3:1 for large text, and meaningful non-text contrast is at least 3:1, computed from rendered colors.
- Keyboard order is logical; focus is visible and not hidden by sticky content.
- Interactive targets are at least 24×24 CSS px or adequately spaced; controls have accessible names.
- With WCAG text-spacing overrides applied, the integrity checks above still pass.

## Severity and attribution

| Severity | Meaning |
| --- | --- |
| Blocker | A flow cannot be completed, data is lost or exposed, or primary content becomes unreadable |
| High | Breakage or a regression across repeated content or a primary identifier, or a WCAG AA failure |
| Medium | Noticeable defect or inconsistency limited in reach that does not impede the task |
| Nit | Polish that does not affect comprehension or use |

Breakage introduced or worsened by the change is at least High. Classify each finding as `introduced`, `worsened` (present at the baseline and made worse), or `pre-existing` (baseline evidence shows it unchanged). Without baseline evidence, treat it as introduced. Report pre-existing findings separately with a suggested disposition; the parent applies the shared incidental-findings rule.

## Verdict

Pass only when no Blocker or High finding is introduced or worsened, and every applicable check has a pass, fail, or not-applicable result with evidence. A missing prerequisite leaves its check not verified, never passed. Report per screen: the viewports and themes covered, the result of each check, then each finding with severity, attribution, location and state, measurement, capture path, and a suggested fix. End with what was not verified and why.
