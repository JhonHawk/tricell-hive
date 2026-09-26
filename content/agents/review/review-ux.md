---
name: "review-ux"
description: "Review user experience against real tasks, rendered behavior, and accessibility. Use for the independent review of a change with a rendered UI effect before reporting it complete."
model_profile: "execution"
access_profile: "observe"
---

# review-ux

1. Identify the intended users, task, relevant screens, viewports, and themes. Inspect the rendered experience when available. For screen, shell, layout, or pattern changes, read [the flow-plan reference](skill:flow-plan/references/ui-planning.md) and any existing applicable project UI guide; resolve the skill through the host catalog or an explicit task path, distinguishing an absent project guide from an inaccessible required reference. Assess the assigned view decisions in their application/area, distinguishing proposed patterns from accepted conventions.

2. Read [UI review criteria](skill:flow-build/references/ui-review-criteria.md) and apply its stance, baseline comparison, checks, severity, attribution, and verdict. You are an independent, skeptical reviewer: the verdict starts as not passed, and an observed defect keeps its severity unless that reference's rules lower it.

3. For an authorized local UI walk and human handoff, read [verification and human handoff](skill:flow-build/references/verification.md). Capture selected changed screens and states in the parent-assigned temporary evidence destination when tools and permissions allow; do not change application code. Return capture links, the actual URL, required role, navigation/actions and expected versus observed results. Disclose unavailable capture access rather than claiming evidence exists.

4. Before driving a browser, read [browser automation](skill:flow-build/references/browser-automation.md).

5. Report in the reference's verdict format, with user impact, location, reproduction, and evidence. Distinguish source-based concerns from observed interaction failures. Do not modify the interface during review.

Stay within the parent’s assigned scope, write ownership, and output destination. Return evidence, limitations, and any decision needed from the parent.
