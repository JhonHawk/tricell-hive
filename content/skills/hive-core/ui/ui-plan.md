# UI decisions

Read when planning a screen change. Read it also when building or reviewing a screen from the `UI` section it writes.

Settle the UI before the plan is approved:

1. Load the project's design system first, as [UI criteria](../../create-design/references/ui-criteria.md) sets: the project's `DESIGN.md`, the UI skills and component library the project guidance names, and the real component catalog, through the library's MCP server or its installed components folder. When those criteria find the `DESIGN.md` missing or stale, add the task its [writing rule](../../create-design/references/design-record.md#writing) sets; [approved example](ui-example.md) sets the same task for the work's `examples.md`.
2. Fix the data contract each changed screen consumes: the response shape, its states and its errors.
3. For each region that changes, decide the existing screen it follows (with its path), each component by its library name with its variant and size, where each action goes with one primary action per region, the loading, empty, error and permission states, and the copy.
4. Draw each region before and after as an ASCII sketch, at most 60 characters wide, with the real labels, item counts and component names, and write the decisions and sketches in the `UI` section of `design.md`. When the user chooses between compositions, show each option as such a sketch, in its preview as [asking a round](../messages/questions.md) sets. When the work's folder under `_support/workspace/` holds the `examples.md` that [create-design](../../create-design/SKILL.md) writes, read [approved example](ui-example.md) before drawing.
5. When no existing pattern fits, settle the composition with one `hive-design-ui` child before the plan is approved, unless an approved example already settles it.

For a new or rearranged screen, the contract in code (types, schema or client) comes first because both sides consume it; the UI against stub data typed by that contract, with realistic volume and text length, and the backend that fulfils it then need only the contract, so they run side by side. A task's `Depends on` names only what it needs to start (a file, data or an answer): an ordering preference or a validation stop the task does not use, such as the first-view acceptance for a backend task, is not a dependency, since declaring it makes independent work wait.
