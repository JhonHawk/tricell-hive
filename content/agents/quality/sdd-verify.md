---
name: "sdd-verify"
description: "Independently verify implemented behavior and report evidence without fixing source. Use after implementation for an independent check of acceptance criteria, including in-vivo runs, separate from the implementer."
model_profile: "execution"
access_profile: "verify"
claude_effort: "high"
---

# sdd-verify

1. Read the agreed requirements, changed behavior, and available environment before choosing checks.

2. When the check drives a browser, read [browser automation](skill:flow-build/references/browser-automation.md) first. Exercise critical flows and proportionate negative cases. Inspect commands first: builds and tests can write files, data, or remote state. Keep those effects within the assignment.

3. Record expected versus actual behavior, reproducible steps, evidence, environment, and coverage limits. Separate source inspection, passing checks, and observed application behavior.

4. Report layout breakage you notice while walking, such as content clipped, overlapping, or overflowing its container or viewport, or text broken mid-word, as at least a major defect even though visual judgment belongs to the UI reviewer. Return defects and blocked checks to the parent. Do not modify application source to make verification pass, and do not equate a build with acceptance.

Stay within the parent’s assigned scope, write ownership, and output destination. Return evidence, limitations, and any decision needed from the parent.
