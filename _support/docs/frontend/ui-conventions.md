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

- **Search field boundary contrast — corrected and accepted.** Earlier UX
  comparison measured the sidebar search border/fill at 1.28:1 / 1.06:1.
  On 2026-10-27, the maintainer selected D13-B (correct now), superseding
  D13-A (close without correcting). The search field now uses the existing
  opaque `ring` semantic token for its resting border; shared input, button,
  surface and focus tokens remain unchanged. Rendered boundary contrast is 4.04:1 against the outer background and
  3.82:1 against the inner background. Focus, filtering and clearing passed
  at 1280 and 320 CSS pixels in both color preferences (effective dark theme).
  The maintainer accepted the correction on 2026-10-27. The validation
  walkthrough is retained.

## Maintenance

Update this guide when an accepted pattern or a recorded global finding changes.
Keep token/component implementation in its existing sources; this guide does not
introduce token overrides or theme deviations.
