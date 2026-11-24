# Child brief

Read before briefing or waiting for a child. It also covers building a child's role, in any flow.

## Every child

- Brief it with the decision its result informs, what it covers (its aspects, targets or task), what is already settled, and when to stop; a role that names its own inputs, such as `hive-verify-task` (the plan file's path, the task ID and the diff base), also gets those. A reviewer's brief follows [Reviewer brief](reviewer-brief.md) instead, with nothing settled about the work it reviews. A child gets these further items when it needs them:
  - the rows of the accounts it uses, when it signs in, as [test accounts](../local/test-accounts.md) sets;
  - the tool or device serial that [Browser tool](../ui/ui-verification.md#browser-tool) or [Device tool](../ui/ui-verification/device-tool.md) sets, when it drives a browser or a device;
  - the paths of the machine file and of the project's `HIVE.local.md` (the main checkout's in a worktree), when it starts, serves or tests the project or uses its environments, with the detail of each entry with a pointer that its task needs, read by the main thread from persistent memory; it reports the facts it verified or found stale, as [local record](../local/local-record.md) sets.
- When the repository's or workspace's `AGENTS.md` has a `## Skills` section, copy into the brief of every child, builder or reviewer, each of its lines whose trigger applies to its task, verbatim: its host may not load `AGENTS.md`.
- Wait for children with the host's wait tool, which returns when a child finishes or sends a message; never with a loop of sleeps and status checks.
- Except for an explicitly assigned `hive-build-owner` under the [local owner boundary](../../flow-build/references/e2e-coordination.md#local-owner-boundary), a child does not commit; the main thread owns ordinary child commits and checks returned evidence, not assertions, before using it.
- An ordinary child works in the main thread's checkout and never creates a worktree. E2E owners and their independent verifiers use the assigned checkout under [E2E coordination](../../flow-build/references/e2e-coordination.md). Launching a child isolated in its own worktree, such as through a host's worktree isolation option, is the main thread creating a worktree, as [git-workflow](../../git-workflow/SKILL.md) step 3 sets.

## Build children

- Pick the role by surface: the `hive-build-*` roles, `hive-write-tests` or `hive-write-docs`.
- The brief gives the plan file's absolute path and the task ID and tells the child to read the plan, down to its `## Checkpoint`, before editing; for an sbx sandbox it gives the path of the copy in the worktree, as [plan record](../plan/plan-record.md#plan-file) sets; it does not paraphrase the task or its criteria. For a task that fixes a defect, it also tells the child to read [defect fix](../build/defect-fix.md) before its first edit, and it asks the child to report each business rule it had to assume, as [business rules](../business/business-rules.md#when-a-rule-enters-the-record) sets.
- For ordinary build children, the main thread runs the task's checks, commits and marks the task. For an E2E local owner, follow [check, correct and integrate](../../flow-build/references/e2e-coordination.md#check-correct-and-integrate); the coordinator retains the record, independent verification and integration.
