---
name: hive-principle-index
description: "Hive principle index: read before writing or reviewing code, or when a line names your moment. One line per principle with the moment it applies; read only the leaves whose moment you are in."
disable-model-invocation: true
---

# Code principles

Read before writing new code in a task, before reviewing a diff, and at any other moment a line below names, such as presenting memory as current in an answer. Each line names the observable moment a principle applies; read the leaf only when you are at that moment, as a file by its path, never through a skill-invocation tool. A principle never fixes a delivery mode and never outranks the project's own rules. Name in your report each principle you read that changed a decision, and none otherwise.

- **Designing a type or a function signature, or writing statically typed code:** read [type system discipline](../hive-principle-type-system-discipline/SKILL.md).
- **Adding validation, error handling or a framework or transport adapter:** read [boundary discipline](../hive-principle-boundary-discipline/SKILL.md).
- **Writing stateful logic, or adding another branch to a chain that repeats a shape assumption:** read [model the domain](../hive-principle-model-the-domain/SKILL.md).
- **Designing a command, job, migration step or loop that can crash, restart or retry:** read [make operations idempotent](../hive-principle-make-operations-idempotent/SKILL.md).
- **Adding an internal API while callers of the old one remain:** read [migrate callers, then delete the legacy API](../hive-principle-migrate-callers-then-delete-legacy-apis/SKILL.md).
- **Shaping or reviewing code that is hard to trace, with layers that only forward or state that is hard to follow:** read [minimize reader load](../hive-principle-minimize-reader-load/SKILL.md).
- **Two or more actors could write the same file, branch, key or object:** read [separate before serializing shared state](../hive-principle-separate-before-serializing-shared-state/SKILL.md).
- **Before measuring, or before reporting or acting on a measured number:** read [explain the number](../hive-principle-explain-the-number/SKILL.md).
- **Before presenting memory, a summary or a tool's behavior as current, or answering about a library, framework, SDK, API, CLI tool or cloud service:** read [verify current state](../hive-principle-verify-current-state/SKILL.md).
- **When editing `.ts` or `.tsx`:** read [TypeScript](../hive-typescript/SKILL.md).

Three principles live inside rule leaves that already apply at their moment, so this index only points to them:

- **Fixing a defect, including the always-on signal and the shared premise of repeated failed fixes:** [defect fix](../hive-core/references/defect-fix.md).
- **Writing or keeping a test, including whether it asserts behavior:** [Tests](../hive-core/references/tests.md).
- **Deciding what to add, with removal and the smallest change first:** [Reuse before writing](../hive-core/references/reuse-before-writing.md).
