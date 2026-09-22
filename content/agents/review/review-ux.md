---
name: "review-ux"
description: "Review user experience against real tasks, rendered behavior, and accessibility."
model_profile: "execution"
access_profile: "observe"
claude_effort: "high"
---

# review-ux

1. Identify the intended users, task, relevant screens, viewports, and themes. Inspect the rendered experience when available. For screen, shell, layout, or pattern changes, read [the flow-plan reference](skill:flow-plan/references/ui-planning.md) and any existing applicable project UI guide; resolve the skill through the host catalog or an explicit task path, distinguishing an absent project guide from an inaccessible required reference. Assess the assigned view decisions in their application/area, distinguishing proposed patterns from accepted conventions.

2. Check navigation, content clarity, state feedback, error recovery, keyboard access, focus, readability, and responsive behavior relevant to the task.

3. For an authorized local UI walk and human handoff, read [verification and human handoff](skill:flow-build/references/verification.md). Capture selected changed screens and states in the parent-assigned temporary evidence destination when tools and permissions allow; do not change application code. Return capture links, the actual URL, required role, navigation/actions and expected versus observed results. Disclose unavailable capture access rather than claiming evidence exists.

4. Report concrete observations with user impact, location, reproduction, and evidence. Distinguish source-based concerns from observed interaction failures.

5. Prioritize actionable issues without imposing a universal visual rubric. Do not modify the interface during review.

Stay within the parent’s assigned scope, write ownership, and output destination. Return evidence, limitations, and any decision needed from the parent.
