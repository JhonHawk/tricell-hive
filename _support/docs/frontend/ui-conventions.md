# Viewer UI conventions

## Scope and sources

This guide covers the local walkthrough viewer in `tooling/viewer/frontend`.
The library baseline is ReUI on shadcn/ui, configured by the frontend
`components.json`; installed components and semantic tokens remain authoritative.
The repository ReUI guidance applies to every UI edit. A formal accessibility
conformance level is undecided; keyboard and focus behavior are verified for
changed interactions.

## Viewer shell

The session detail pattern uses the existing project sidebar and global header,
a session toolbar outside the scrolling canvas, and stacked walkthroughs inside
that canvas. Session actions operate on the native conversation identity across
its tickets. Work status remains on each walkthrough. The session toolbar
pattern was accepted by the maintainer and integrated in [PR #227](https://github.com/JhonHawk/tricell-hive/pull/227).
Source examples: `tooling/viewer/frontend/src/shell.tsx` and `application.tsx`.

## Global findings

- **Repeated landmark names — P3, reported with no action.** Independent UX
  comparison on 2026-10-27 found axe `landmark-unique` groups containing six
  nodes in candidate `39187977` and seven in base `82ccf111`, in the existing
  dark theme. Session/walkthrough regions named `Entregas`, navigation regions
  named `Índice`, and repeated flow/capture regions need surrounding context
  to identify their walkthrough. This affects orientation through landmark
  navigation; no task was observed to be impossible and this axe rule alone
  does not establish WCAG nonconformance. This is a shared stacked-walkthrough
  pattern, not an introduced session-toolbar defect. No separate light-theme
  measurement is available (the viewer retains its existing dark appearance).
  Source: independent `hive-review-ux` comparison; exact counts have not been
  remeasured on later candidates. Disposition: recorded once, no implementation
  or ticket requested. Revisit if impact worsens or a distinct theme is added.

- **Button focus ring — published library baseline, not applicable.** The
  1.88:1 ring measurement remains valid in the effective dark theme, with both
  color preferences. An independent source/capture check on 2026-10-27 found
  the same focus classes in the [official base-nova button registry](https://ui.shadcn.com/r/styles/base-nova/button.json)
  and the same ring/background tokens in the [published neutral theme](https://github.com/shadcn-ui/ui/blob/main/apps/v4/content/docs/%28root%29/theming.mdx).
  The outline variant's dark border wins over the opaque focus border; the
  measured ring is still visible. Under the project's review baseline, this
  published default is not a reportable theme finding while formal AA
  conformance remains undecided. This is not a claim of general WCAG
  conformance. Publication was checked current, not pinned to a historical
  registry version. Revisit if an AA requirement is adopted or the component
  is customized. Source capture: `ux-matrix-focus-320.png` in the local change
  walkthrough.

- **Search field boundary contrast — P2, declined for this delivery.** Independent UX
  comparison of `051d45d9` and base `82ccf111` measured the sidebar search
  border/fill at 1.28:1 / 1.06:1, unchanged by the toolbar. Both color preferences
  produced the same effective dark theme and values. The project maps
  `--color-input` to the 10%-white border token; the published neutral theme
  uses a distinct 15%-white input token. This local alias prevents treating
  the field values as an exact published-input baseline. Source:
  `src/styles.css`; capture: `ux-matrix-search-1280.png` in the local change
  walkthrough. Changing the shared input appearance requires a theme decision
  and is not part of the session-toolbar or screenshot-only guidance change.
  No duplicate was found by the bounded tracker search. On 2026-10-27, the
  maintainer chose D13-A (close without correcting), preserving the accepted
  theme and retaining the validation walkthrough. No issue was opened and no
  theme token was changed. This is a decline for the current delivery, not a
  suppression; recurrence count: 1.

## Rules not adopted

- Increasing the search field boundary contrast in this delivery was declined
  by the maintainer under D13-A (close without correcting). Revisit if a later
  change worsens the boundary, introduces another effective theme, or adopts
  a different accessibility requirement. The current theme remains accepted.

## Maintenance

Update this guide when an accepted pattern or a recorded global finding changes.
Keep token/component implementation in its existing sources; this guide does not
introduce token overrides or theme deviations.
