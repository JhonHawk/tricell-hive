---
name: flow-report
description: Create a visual HTML report, printable paper, comparison, explainer, or deck when the requested deliverable needs presentation layout. Not for resumable research notes, plans, or routine Markdown findings.
---

# Durable reports

Use this skill when the requested deliverable benefits from an explicit visual or printable presentation. A request to save findings for later does not by itself call for HTML: keep resumable research in its Markdown work record using the research procedure and established workspace location. Preserve an explicitly requested format. Keep ordinary answers, status updates, and handoffs in prose.

Choose the asset that matches what the reader receives:

| Need | Asset |
| --- | --- |
| Operational report or findings | [template](assets/baseline.html) |
| Formal printable document | [template](assets/skeleton-paper.html) |
| Concept or research explanation | [template](assets/skeleton-explainer.html) |
| Code-review verdict | [template](assets/skeleton-review.html) |
| Unresolved option comparison | [template](assets/skeleton-comparison.html) |
| Guided presentation | [template](assets/skeleton-deck.html) |

Copy an asset and replace its placeholders with accurate, audience-appropriate content. Keep its responsive and print behavior unless the report needs a deliberate alternative. Write human-facing content in the session language. Put the report in the destination resolved by the workspace guidance; do not invent a destination or overwrite a retained record without authorization.

Reports are offline by default: keep styles, scripts, images, and diagrams in the file. If an external asset is necessary, add `<meta name="flow-report-assets" content="external">`, name the dependency in the report, and make the readable content work without it. Include a short provenance footer with the date, inputs, and relevant paths.

Run [the self-check](scripts/self_check.py) with `python3 <skill-dir>/scripts/self_check.py <report.html>` after writing and fix failures. For reports containing inline SVG diagrams, also run [the geometry check](scripts/verify_geometry.py) with `python3 <skill-dir>/scripts/verify_geometry.py <report.html>`. Read [the diagram grammar](references/diagram-grammar.md) before drawing or modifying an SVG diagram.
