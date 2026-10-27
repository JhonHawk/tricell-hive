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

- **Shared control contrast — P2, pending disposition.** Independent UX comparison
  of `051d45d9` and base `82ccf111` on 2026-10-27 measured button focus rings at
  1.88:1 against their background, and the sidebar search border/fill at
  1.28:1 / 1.06:1. Both color preferences produced the same effective dark
  theme and the same values. The reviewer identified these as SC 1.4.11 AA
  failures, unchanged by the session-toolbar candidate. Source: shared
  `components/ui/button.tsx` and `src/styles.css`; captures
  `ux-matrix-focus-320.png` and `ux-matrix-search-1280.png` in the local change
  walkthrough. Token/component changes are a proposed theme deviation, not
  adopted in this work. No duplicate was found by the bounded tracker search.
  The close question must settle the grouped finding before a final disposition
  is recorded; its presence does not authorize changing shared theme tokens.

## Maintenance

Update this guide when an accepted pattern or a recorded global finding changes.
Keep token/component implementation in its existing sources; this guide does not
introduce token overrides or theme deviations.
