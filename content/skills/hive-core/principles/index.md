# Code principles

Read before writing new code in a task, before reviewing a diff, and at any other moment a line of the code principles index names, such as presenting memory as current in an answer. Each line names the observable moment a principle applies; read the leaf only at that moment, as a file by its path, not through a skill-invocation tool. A principle never fixes a delivery mode and never outranks the project's own rules. Name in your report each principle you read that changed a decision, and none otherwise.

- **Designing a type or a function signature, or writing statically typed code:** read [type system discipline](type-system-discipline.md).
- **Adding validation, error handling or a framework or transport adapter:** read [boundary discipline](boundary-discipline.md).
- **Writing stateful logic, or adding another branch to a chain that repeats a shape assumption:** read [model the domain](model-the-domain.md).
- **Designing a command, job, migration step or loop that can crash, restart or retry:** read [make operations idempotent](make-operations-idempotent.md).
- **Adding an internal API while callers of the old one remain:** read [migrate callers, then delete the legacy API](migrate-callers-then-delete-legacy-apis.md).
- **Shaping or reviewing code that is hard to trace, with layers that only forward or state that is hard to follow:** read [minimize reader load](minimize-reader-load.md).
- **Two or more actors could write the same file, branch, key or object:** read [separate before serializing shared state](separate-before-serializing-shared-state.md).
- **Before measuring, or before reporting or acting on a measured number:** read [explain the number](explain-the-number.md).
- **Before presenting memory, a summary or a tool's behavior as current, or answering about a library, framework, SDK, API, CLI tool or cloud service:** read [verify current state](verify-current-state.md).
- **When editing `.ts` or `.tsx`:** read [TypeScript](../../hive-typescript/SKILL.md).

Three principles live inside rule leaves that already apply at their moment:

- **Fixing a defect, including the always-on signal and the shared premise of repeated failed fixes:** [defect fix](../build/defect-fix.md).
- **Writing or keeping a test, including whether it asserts behavior:** [Tests](../build/tests.md).
- **Deciding what to add, with removal and the smallest change first:** [Reuse before writing](../plan/reuse-before-writing.md).
