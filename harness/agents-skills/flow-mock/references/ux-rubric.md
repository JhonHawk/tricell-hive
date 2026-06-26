# UX flow rubric

Checked per flow by `ux-flow-reviewer` while navigating the live mock or QA deployment.
Plain markdown so any harness can consume it. Pass/fail per dimension, per flow.

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
