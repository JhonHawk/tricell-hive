# Local execution record

Read when the coordinator uses [E2E coordination](../../flow-build/references/e2e-coordination.md). The [tasks helper](../scripts/tasks.py) owns the `execution` commands (`init`, `check`, `update`, `validate-handoff`; `--help` lists their flags); [change validation](../scripts/check_change.py) also checks a checkpoint's execution pointer unless `--archive` is given.

Only the coordinator writes the record or invokes its mutations. `--coordinator` must match the recorded native identity; it is a consistency guard, not authentication or permission to act. Owner and verifier returns are inputs the coordinator inspects, never direct writes to the record. No command runs the checks it records, reaches the network or publishes Git changes; each exits 2 on invalid input, unavailable evidence or Git fact, a stale generation or a write failure.

## Location and initialization

The record lives at `<repo>/_support/workspace/<run>/execution.json`: outside the change folder, already Git-ignored and untracked, with no symlink escape, mode 0600; the helper creates no ignore rules or worktrees. Evidence paths are relative files inside the record's directory; missing files, `..`, absolute paths and symlink escapes fail validation. Keep native identities and raw outputs there, out of versioned plans.

`init` reads a JSON manifest, refuses overwrite, adds the server-owned fields (`schema_version`, `generation`, `coordinator_id`, `plan_ref`, `plan_revision`) and links the record from `## Checkpoint` in `tasks.md` with a relative `- Execution record:` pointer that later checkpoint writes keep. Run it with absolute paths: `python3 -I <tasks.py> <change-folder> execution init --record <execution.json> --input <manifest.json> --coordinator <native-root-id>`. The manifest fields:

```json
{
  "grant_ref": "proposal.md#decisions",
  "integration_target": {
    "repo": "/absolute/repository",
    "worktree": "/absolute/integration-worktree",
    "branch": "integration-branch"
  },
  "units": [{
    "id": "U1",
    "tasks": ["T1"],
    "criteria": ["AC1"],
    "depends_on": [],
    "assignment_id": "assignment-1",
    "owner_id": "native-owner-1",
    "role": "hive-build-owner",
    "repo": "/absolute/repository",
    "worktree": "/absolute/assigned-worktree",
    "branch": "owner-branch",
    "base_sha": "0000000000000000000000000000000000000000",
    "allowed_paths": ["src/unit-a/", "tests/unit_a_test.py"],
    "resources": ["local-database:unit-a"],
    "state": "assigned",
    "returned_head": null,
    "evidence": [],
    "verifier": null,
    "limitations": [],
    "next_action": "Implement the assigned task",
    "cleanup": null
  }],
  "publication": null
}
```

An allowed path without a trailing `/` is an exact file; one ending in `/` is a literal subtree prefix; globs, `.` and non-normalized paths are rejected. Resources are exclusive string identities, not commands. Every assigned, running, returned, verified or blocked unit protects its branch, worktree, paths and resources: overlaps fail validation. An `integrated` unit retains history but is no longer an active assignment, and that alone never authorizes cleanup. `plan_revision` is a SHA-256 over proposal, design, task definitions and spec deltas, so checkboxes, checkpoint and progress sections do not change it.

## Updates and exact-head verification

Read the current JSON, edit a complete copy, and pass it as `--input` with its generation as `--expected-generation`; the helper replaces the record atomically under a short lock, and a stale generation or a plan edited during validation is rejected with the prior record preserved. A lock left after an interruption is retained: reconcile the native writer before removing its exact lock directory.

Transitions are `assigned → running → returned → verified → integrated`; a unit can become blocked before `verified`, and a blocked one resumes running or receives a replacement. A correction returns to running and clears the verifier, and a changed returned head clears its verdict. Unit IDs stay fixed for an execution; changing assignment boundaries requires replacement.

For `returned`, record a full head and at least one evidence file; `validate-handoff` reads local Git only (repository identity, worktree, branch, ancestry from the assigned base, clean status, and every added, deleted or renamed path against `allowed_paths`), and a failed handoff stays unverified. For `verified`, supply a `verifier` object with `id`, the exact returned `head`, `criteria` equal to the unit's, and `checks` of `{command, result, path}`. The verifier identity must differ from the current owner and every historical owner of the unit, so a replacement cannot make an earlier author's preserved work eligible for self-verification. Metadata cannot establish that a command ran, that an identity is independent or that an output satisfies a criterion: the coordinator reads the actual evidence and the verifier's reasoning before accepting the transition.

## Replacement and plan reconciliation

To replace an owner, preserve the unit's returned or partial head, evidence and limitations, set a fresh `assignment_id`, the new owner and boundaries and `state: assigned`, and add a `closure` of kind `completed` or `stop_ack` for the former writer, with its `assignment_id`, `owner_id` and evidence `path`. The coordinator first observes that native completion or stop acknowledgment and reconciles Git and dirty work; a timeout or an absent ID is not enough. The helper archives the previous assignment into append-only `history` and never resets or deletes working-tree files.

When proposal, design, task definitions or specs change, `check` and ordinary updates report plan drift. Block every candidate unit, keep the current fingerprint in the input and use `update ... --reconcile-plan`: all verdicts clear and only task, criterion and dependency references may change, while owner, role, worktree, branch, base SHA, allowed paths and resources stay, since blocked status does not prove the former writer stopped. Resume through running only after the coordinator confirms the assignment is still valid. A change of owner or boundary goes first through that reconciliation, then through the replacement above in a separate update; adding or removing unit IDs or changing the integration target requires a separately initialized execution after the old assignments close.

## Integration and publication

Only independently verified units can become integrated. Before that, merge the exact verified heads in the declared target, run the joint check, and record the root fields `integration_scope` (`local` or `remote`; `remote` is a declaration whose remote effect the delivery flow verifies, never an inferred push), `integration_head` and `joint_check` (`head`, `command`, `result`, `path`). The helper reads back the target, requires it clean at that head with every integrated unit's head as an ancestor, and checks the joint evidence. `publication` stays null after local integration; if delivery occurs, the coordinator records its `head`, `result` (`pass`, `fail` or `blocked`) and evidence `path`. Publication authorization and the remote readback stay with the delivery flow.
