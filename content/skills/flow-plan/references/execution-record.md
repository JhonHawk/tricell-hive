# Local execution record

Read when the coordinator uses [E2E coordination](../../flow-build/references/e2e-coordination.md). The existing [tasks helper](../scripts/tasks.py) owns the optional `execution` commands below; ordinary `status`, `done` and `checkpoint` retain their interfaces. [Change validation](../scripts/check_change.py) also checks a checkpoint's execution pointer unless `--archive` is given.

Only the coordinator writes the record or invokes its mutations. `--coordinator` must match the recorded native identity; it is a consistency guard, not authentication, a capability token or permission to act. Native role boundaries and the approved grant govern who can invoke the helper. Record files are mode 0600. Owner and verifier returns are inputs the coordinator inspects, never direct writes to the record.

## Location and initialization

Create the work's ignored workspace directory using the project's existing support layout. The helper requires `<repo>/_support/workspace/<run>/execution.json`, outside the change folder, already Git-ignored and untracked, with no symlink escape. It creates no ignore rules or worktrees. Evidence paths are relative files inside the record's directory, including subdirectories; missing files, `..`, absolute paths and symlink escapes fail validation. Keep native identities and raw outputs there, out of versioned plans.

Run with absolute change-folder, record and input paths:

```text
python3 -I <tasks.py> <change-folder> execution init --record <execution.json> --input <manifest.json> --coordinator <native-root-id>
python3 -I <tasks.py> <change-folder> execution check --record <execution.json>
python3 -I <tasks.py> <change-folder> execution update --record <execution.json> --input <candidate.json> --coordinator <native-root-id> --expected-generation 0
python3 -I <tasks.py> <change-folder> execution validate-handoff --record <execution.json> --unit U1
```

These commands return 0 on success and 2 on an invalid input, unavailable evidence/Git fact, stale generation or write failure. No command executes the checks named in metadata, accesses the network or publishes Git changes. Git reads reuse the existing [Git state](../../git-workflow/scripts/git_state.py) process boundary and timeout.

`init` reads a JSON manifest, refuses overwrite, adds the server-owned fields below and links the record from `## Checkpoint` in `tasks.md` with a relative `- Execution record:` pointer. Later checkpoint writes preserve that pointer. Supply these manifest fields:

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

Replace the example paths, IDs and SHA with actual values. An allowed path without a trailing `/` is an exact file; one ending in `/` is a literal subtree prefix. Globs, `.` and non-normalized paths are rejected. Resources are exclusive string identities, not commands.

The stored root additionally has `schema_version: 1`, `generation: 0`, `coordinator_id`, absolute `plan_ref` and `plan_revision` (SHA-256 over proposal, design, task definitions and spec deltas). Checkboxes, checkpoint and progress sections are excluded from that fingerprint. Unknown fields/versions, duplicate JSON keys/unit or assignment IDs, unknown task/criterion/dependency IDs and dependency cycles are rejected. Every assigned, running, returned, verified or blocked unit protects its branch, worktree, paths and resources: overlapping worktrees, same-repository branches or allowed paths, and shared exclusive resources fail validation. `integrated` units retain history but are no longer active assignments; this alone never authorizes cleanup.

## Updates and exact-head verification

Read the current JSON, edit a complete copy, then pass that copy as `--input` with its current generation as `--expected-generation`. The helper requires both generations to match, increments once and replaces the record atomically under a short exclusive lock. It binds one plan fingerprint for the operation and checks that fingerprint again before replacement; a plan edited during validation is rejected without adopting its new revision. Concurrent losers and invalid updates preserve the prior record. A lock left after interruption is retained: reconcile the native writer before removing its exact lock directory. It is never broken by age. Init links the checkpoint first; an interrupted init can leave a dangling pointer, which change validation reports, without overwriting an earlier record.

Normal transitions are `assigned → running → returned → verified → integrated`. Running, returned or verified units can become blocked; assigned units can also be blocked before dispatch. Corrections return to running and clear the verifier. Blocked units resume running, or receive a replacement. A returned head change clears its previous verdict and cannot simultaneously become verified. Unit IDs remain fixed for an execution; changing assignment boundaries requires replacement. Plan changes require explicit reconciliation, not silently accepting another fingerprint.

For `returned`, record a full head and at least one evidence file in `evidence`. `validate-handoff` reads only local Git: repository common-directory identity, exact worktree root and branch, full current head, commit ancestry from the assigned base, clean tracked/untracked status and every added/deleted/renamed path in the complete base-to-head diff. A failed handoff stays unverified. Before `verified`, supply:

```json
{
  "id": "independent-native-verifier",
  "head": "full-40-character-returned-head",
  "criteria": ["AC1"],
  "checks": [{"command": "actual command", "result": "pass", "path": "verification/output.txt"}]
}
```

The verifier identity must differ from the current owner and every historical owner of the unit, so replacement cannot make an earlier author's preserved work eligible for self-verification. This v1 check conservatively excludes all historical owners. Its head must match exactly, its criteria must equal the unit's criteria, and each check needs a command, passing result and available confined evidence file. The helper repeats the Git handoff check before verified and when first integrating. A metadata check cannot establish that a command ran, that an identity is independent or that an output satisfies a criterion. The coordinator reads the actual evidence and the verifier's reasoning before accepting the transition.

## Replacement and plan reconciliation

To replace an owner, preserve the unit's returned/partial head, evidence and limitations. Set a fresh `assignment_id`, the new owner and assignment boundaries, `state: assigned`, and this `closure` object for the former writer:

```json
{
  "assignment_id": "assignment-1",
  "owner_id": "native-owner-1",
  "kind": "completed",
  "path": "native/completion.txt"
}
```

`kind` is `completed` or `stop_ack`. The coordinator first observes that native completion or stop acknowledgment and reconciles Git/dirty work; timeout or an absent ID is insufficient. The helper checks closure identity and available evidence, archives the previous assignment with that closure into append-only `history`, clears the current closure/verifier and accepts the new assignment. Its next action carries the recovery checkpoint. History cannot be edited or an assignment ID reused; the helper never resets or deletes working-tree files.

When proposal/design/task definitions or specs change, `check` and ordinary updates report plan revision drift. Reconcile every unit against the revised plan, keep the current record's fingerprint in the input and use `update ... --reconcile-plan`. Each candidate unit must be blocked; all current verdicts clear and the helper stamps the new fingerprint. This permits updating only task, criterion and dependency references for the revised plan. Owner, role, repository, worktree, branch, base SHA, allowed paths and exclusive resources remain unchanged: blocked status does not prove the former writer stopped or release its boundaries. Resume through running only after the coordinator confirms the assignment is still valid. If an owner or an assignment boundary must change, first reconcile with the former boundaries protected, then use the confirmed replacement procedure above in a separate update. Adding/removing unit IDs or changing the integration target requires a separately initialized execution after closing the old assignments; v1 does not silently rewrite those boundaries.

## Integration and publication

Only independently verified units can become integrated. Before that transition, merge the exact verified heads in the declared target and run the joint check. Record root fields `integration_scope` (`local` or `remote`), `integration_head` (full target HEAD) and `joint_check`:

```json
{"head": "full-integration-head", "command": "actual joint command", "result": "pass", "path": "joint/output.txt"}
```

The helper reads back the declared repository/worktree/branch, requires a clean target at that head, checks that every integrated unit's head is its ancestor and checks available joint evidence bound to it. Integrated history can remain after an owner worktree is cleaned up; the integration repository and target must remain available for live checks. `remote` is a declaration whose remote effect the delivery flow verifies, never an inferred push.

`publication` remains null after local integration. If delivery occurs, the coordinator separately records `{"head": "full-published-head", "result": "pass", "path": "publication/evidence.txt"}`; `fail` and `blocked` are also valid publication results. This validates available metadata only. Publication authorization and actual remote readback remain with the existing delivery flow.
