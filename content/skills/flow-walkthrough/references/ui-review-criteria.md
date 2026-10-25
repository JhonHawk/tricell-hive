# UI review criteria

Read when reviewing a rendered UI change, or when briefing the reviewer. [The UI gate](ui-gate.md) decides when the review is required; [browser automation](../../flow-build/references/browser-automation.md) governs how the browser is driven. The [design baseline](#design-baseline) sets which project and library sources take precedence over these criteria.

## Stance

- Review independently of the implementer. The verdict starts as not passed and becomes a pass only when every applicable check has a recorded result.
- An observed defect keeps its severity unless a written rule below lowers it; cite that rule when lowering. Do not talk a finding down because it looks minor, matches a sibling screen, or was not in the plan.
- The parent supplies scope, data, and viewports but does not pre-rate findings or say what to ignore. Two exceptions apply. The first, named in [the UI gate](ui-gate.md#the-independent-children), limits pre-existing findings to the touched views and may supply a baseline report; it does not limit introduced or worsened findings. The second, the [design baseline](#design-baseline), lets the brief carry the library configuration and the project UI guide's recorded decisions and global findings, which this review applies as that section sets: a library default or a recorded decision is not applicable whether or not the change adds it, and skipping a recorded global finding never covers one the change introduces or worsens.
- Judge from measurement and interaction. A capture locates a problem; computed geometry, styles, and behavior decide it. Vision judgments of wrapping, overlap, and alignment are unreliable on their own.

## Design baseline

The UI library the application chose, with its theme and its installed components as published, is the baseline: customizing it is an explicit decision, the user's or one the project already records.

1. **Read the baseline before the checks.** Read the UI library the application uses and its configuration, such as `components.json`; its theme tokens and installed components; the project UI guide, with its theme deviations, rules not adopted and global findings; and any accessibility requirement the project declares in that guide or in its `AGENTS.md`, or that the ticket requests. An absent guide or requirement is a fact to report, not a failed check.
2. **Precedence, highest first:** a declared accessibility requirement; the project's recorded decisions and UI guide; the UI library's defaults; the screen's accepted pattern, which decides composition only and cannot turn a failed integrity, state or accessibility check into a pass; the checks below.
3. **A library default** is a theme-token value, or a style the installed component sets inside its own file, that the project has not changed. Establish it with all of these checks:
   - (a) `git diff <base> <candidate>` on the library configuration, the theme file and the installed-component folder: a value the change edits in a file that existed at the base is introduced by the change, not a default; a file the change adds, such as a newly installed component, goes to check (b) like any other;
   - (b) the value matches a value the library published for the declared configuration, such as `style` and `baseColor` in `components.json`, or the installed package when the library copies no files, either its current value or the one of the version the project installed, read from the library's documentation, registry or release history with no private identifier in the request; and `git log -p --follow` on the file shows no later edit to that value;
   - (c) a class or style set where the component is used, rather than inside it, is a project value;
   - (d) position, layering and composition, such as a sticky header covering focus, are always project values.

   When the published value cannot be read, or the repository has no history, as in a shallow clone, the value is a project value.
4. **Not applicable.** A library default, or a value that a recorded project decision sets, is correct for this review in the checks that measure that value itself, such as its contrast, color, size, spacing or radius: those checks record `not applicable: library default` or `not applicable: project decision <reference>`, never a finding. Without a declared accessibility requirement this holds even when the value misses a WCAG threshold. With one, a failure the requirement covers is a finding, and its fix is a theme deviation. Neither a default nor a decision makes the layout and content integrity checks, the state checks, keyboard order, a focus indicator that is missing, or accessible names not applicable.
5. **Theme deviation.** A fix is a `theme deviation` when it changes a theme token, edits an installed library component, or uses raw palette values where the project has a token for that purpose. Raw palette values that the project UI guide allows, such as categorical identity colors, are not one. Other files point to this definition rather than restating it.
6. **Global findings.** A global finding is one whose cause lives in a shared source: a theme token, an installed or shared component, or the shell or layout, located by the checks in 3. A class set on one view over a shared token is that view's finding. Before reporting a finding or recommending a change, locate its cause, and report a global finding once, at its source, with its measured value per theme and the affected views as examples, not once per view. Skip a global finding the project UI guide already records only when the change does not touch its source and the measurement matches the value recorded for the same theme; a theme with no recorded value is measured and reported once to complete the record. Otherwise report it, attributed as below.

## Prepare the comparison

1. Inventory each affected screen as a whole: regions, columns, rows, and components, including those the change only displaces; for pre-existing findings, report only on the touched views that the UI gate's exception sets.
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
- The screen conforms to its accepted pattern. A sibling screen is a valid reference only for checks it passes itself; a shared defect whose cause is not a library default or a recorded decision is a global finding, as the [design baseline](#design-baseline) sets, not a defense.

**States**

- Loading, empty, error, disabled, hover, focus, selected, and success states exist where relevant and are captured. Loading does not shift the final layout.
- Errors state what went wrong and how to recover; empty states lead to the correct first action.

**Flow and feedback**

- The user can tell where they are and what to do next; every action gives visible feedback in the same view.
- Destructive or sensitive actions state consequences and require deliberate confirmation.
- Labels and messages use the user's vocabulary; each role in scope sees the right data and affordances.
- The console shows no new errors during the walk.

**Accessibility (WCAG 2.2 AA)**

Compute contrast from rendered colors, after the [design baseline](#design-baseline) has set which values are not applicable.

- Text contrast is at least 4.5:1, or 3:1 for large text (SC 1.4.3), including placeholder text and text on tinted surfaces.
- Non-text contrast is at least 3:1 (SC 1.4.11) for the visual information required to identify a component or its state: the indicator that shows where a text input or select is, such as its border, bottom border or fill, even when a label is present; the boundary or fill of any other control without visible text or a contrasting icon of its own; checkbox and radio outlines, state indicators and meaningful icons; and a focus indicator the author styled. A focused text input's visible caret counts as its focus cue; a custom low-contrast caret does not.
- A button or link identified by its own visible text or a contrasting icon needs no boundary contrast. Non-text contrast does not apply to decorative borders and separators, to disabled components, or to a user-agent style the author did not modify.
- Each contrast finding names the success criterion and its level, the element, the cue that identifies it, and the measured ratio.
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

Breakage introduced or worsened by the change is at least High. Classify each finding as `introduced`, `worsened` (present at the baseline and made worse), or `pre-existing` (baseline evidence shows it unchanged). Without baseline evidence, treat it as introduced; a baseline report the parent supplies under the UI gate's exception only means you do not report that report's findings again: they keep the classification the local review gave them, and the report is not evidence that turns an introduced finding into a pre-existing one. Report pre-existing findings separately with a suggested disposition; the parent applies the shared incidental-findings rule, listing them as [incidental findings](../../flow-close/references/incidental-findings.md#which-findings-are-listed-and-how) sets.

A suggested fix uses the project's existing tokens, variants and components. Mark a fix as a `theme deviation` when the [design baseline](#design-baseline) defines it as one; the parent applies a theme deviation only with an explicit decision, as [the UI gate](ui-gate.md#the-independent-children) sets.

## Verdict

Pass only when no Blocker or High finding is introduced or worsened, and every applicable check has a pass, fail, or not-applicable result with evidence. A finding whose only fix is a theme deviation that a recorded decision declined becomes `not applicable: project decision <reference>`. A missing prerequisite leaves its check not verified, never passed. Report per screen: the viewports and themes covered, the result of each check, then each finding with severity, attribution, location and state, measurement, capture path, and a suggested fix. List global findings apart, once each, with their source, measured value and example views. End with what was not verified and why.
