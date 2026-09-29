---
name: "hive-verify-change"
description: "Independently verify a completed change's behavior end to end, including in-vivo and browser runs, and report evidence without fixing source. Use after a change is implemented, whatever its number of tasks and with or without a plan, separate from the implementer; not for verifying one task of a retained plan during the build."
model_profile: "execution"
access_profile: "verify"
---

# hive-verify-change

1. Read the agreed requirements, changed behavior, and available environment before choosing checks.

2. When the check drives a browser, read [browser automation](skill:flow-build/references/browser-automation.md) first. Exercise critical flows and proportionate negative cases. Inspect commands first: builds and tests can write files, data, or remote state. Keep those effects within the assignment.

3. Check authentication by behavior and by cookie or token names and attributes, never by printing their values; read [browser automation](skill:flow-build/references/browser-automation.md) for the allowed checks. If a secret value reaches your output anyway, report which one and where at the top of your result.

4. Record expected versus actual behavior, reproducible steps, evidence, environment, and coverage limits. Separate source inspection, passing checks, and observed application behavior.

5. Report layout breakage you notice while walking, such as content clipped, overlapping, or overflowing its container or viewport, or text broken mid-word, as at least a major defect even though visual judgment belongs to the UI reviewer. Return defects and blocked checks to the parent. Do not modify application source to make verification pass, and do not equate a build with acceptance.

Stay within the parent’s assigned scope, write ownership, and output destination. Return evidence, limitations, and any decision needed from the parent.
