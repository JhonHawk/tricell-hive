# Criteria

Each check gives a threshold or an observation. A check marked **[measured]** is judged as a defect (P0 to P3) by the UX review when the change introduced it, subject to the precedence in [SKILL.md](../SKILL.md#precedence). Unmarked items are design guidance and are never reported as defects.

## Screen modes

Pick one per area of the app. When the record exists, take the mode from its «Product» entry. A new or changed mode reaches the record only as [design record](design-record.md#writing) «Writing» sets; never edit the record for it on your own.

- **Operate** (default in an app with a component library): the user scans, compares and acts repeatedly. Use the library's standard controls and layout. Project style contributes only typography, palette, density and one distinctive trait, never layout or controls.
- **Read**: documentation or long text. One calm column; the line length is whatever the project or library sets.
- **Persuade**: public pages only. The offer and one working action are visible in the first viewport.

## Accessibility floors

All **[measured]**. Each names its standard and level.

- **[measured]** Status is never shown by color alone: a second cue (text, icon or shape) exists. WCAG 2.2 1.4.1, level A.
- **[measured]** Each pointer target is at least 24x24 CSS px, or meets an exception of 2.5.8 (spacing, equivalent control, inline in text, user-agent control, essential). WCAG 2.2 2.5.8, level AA. Guide value, not a floor: 44 px when the library's size allows it; its absence is not a defect.
- **[measured]** With text enlarged to 200 %, nothing is lost (1.4.4, AA).
- **[measured]** When [UI verification](../../flow-verify/references/ui-verification.md) checks the mobile viewport: at its 440 CSS px width, content and functions remain without two-dimensional scrolling. This Hive floor is based on WCAG 1.4.10 (AA) but is looser, since that criterion tests at 320 CSS px, so passing it never establishes 1.4.10 conformance; a declared accessibility requirement does not add the 320 px test unless the project or the ticket names that width. In a desktop-only app this is guidance, not a defect.
- **[measured]** A focused element is not entirely hidden by sticky or overlay content. WCAG 2.2 2.4.11, level AA.
- **[measured]** The focus indicator, and any state conveyed only by contrast, has at least 3:1 contrast against adjacent colors. WCAG 2.2 1.4.11, level AA.
- **[measured]** Motion triggered by an interaction and not essential is disabled or reduced under `prefers-reduced-motion`. WCAG 2.2 2.3.3, which is level **AAA**: judge it only in what the change builds, and cite that level.
- **[measured]** Closing a modal overlay returns focus to the element that opened it, unless that element no longer exists. Source: the WAI-ARIA Authoring Practices Guide modal dialog pattern, with WCAG 2.4.3 (A). Judge it only in what the change builds.

Text contrast, keyboard order and accessible names are already measured by the UX review; do not repeat them.

## Collections

Applies to lists, tables and card grids.

- **[measured]** The scope of a search is named (for example, "Search orders").
- **[measured]** A result count is shown.
- **[measured]** Active filters are visible and removable one by one.
- **[measured]** A control resets all filters.
- **[measured]** "No results" keeps the query and offers a way to widen it.
- **[measured]** An empty catalog is a different state from "No results" and offers an action.
- **[measured]** When the app routes, search, filters, sort and page are in the URL.
- **[measured]** Returning from a detail view restores the page, scroll position and focus.
- Guide: choose a table to compare, cards or a list to scan.

## Copy

- **[measured]** A button label names the action or its result.
- **[measured]** A destructive confirmation names the action; "OK" or "Yes" alone fails.
- **[measured]** An error says what failed and how to recover.
- **[measured]** A placeholder is never the only label.
- **[measured]** In code that ships to production, no demo data is presented as real. Test or seeded data of a walkthrough does not count.
- Guide: the label predicts the result.

## Extremes

- **[measured]** With more than 1000 items, the list paginates or virtualizes: the DOM never holds all items at once.
- **[measured]** While a submit is pending, its control is disabled and the network log shows one request.
- Guide: design for interrupted gestures (a drag released outside the target).

Long values and text 30 to 50 % longer are already covered by the UX review.
