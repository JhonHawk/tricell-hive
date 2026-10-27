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
pattern is a candidate awaiting human acceptance in the session-toolbar change.
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

## Maintenance

Update this guide when an accepted pattern or a recorded global finding changes.
Keep token/component implementation in its existing sources; this guide does not
introduce token overrides or theme deviations.
