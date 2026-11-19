---
name: hive-rule-reuse-before-writing
description: "Hive rule leaf: read before writing new code in a task, or when a plan weighs approaches. Take the first rung that fully does the job: omit it, reuse it, use the platform, an installed dependency."
disable-model-invocation: true
---

# Reuse before writing

Read before writing new code in a task, and when [flow-plan](../flow-plan/SKILL.md) step 2 weighs approaches. A task or an approach that fixes a defect first reads [defect fix](../hive-rule-defect-fix/SKILL.md). Take the first rung that fully does what the task needs:

1. It need not exist: leave out features, options and flexibility the request does not name, and name them in the report. Remove dead code, redundant validators and stub references the change touches before you build on them.
2. The codebase already has it, such as a helper, component or pattern: use it as the surrounding code does, unless it is a workaround, such as one a comment marks or a guard [defect fix](../hive-rule-defect-fix/SKILL.md) does not count as a fix: do not copy it, and name it in the report. Prefer deleting to adding: make the smallest change that solves the problem, with no layer, abstraction or parameter the request does not need.
3. The standard library or a platform feature does it, such as a native input type or a database constraint, unless the project has its own; a design-system component beats a native control.
4. An installed dependency does it.
5. Only then write new code, or add a dependency after running `python3 -I <flow-build-dir>/scripts/check_package.py` ([check package](../flow-build/scripts/check_package.py)) as the global guidance's package check sets, and say in the plan's approach or the report why rungs 2 to 4 do not fit.

No rung drops validation at trust boundaries, error handling that prevents data loss, security or accessibility.
