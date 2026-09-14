# flow-build — OPEN (reconcile)

The reconciler's entry, run on every `/flow-build` invocation: resolve the portable plan,
validate authority for the requested next action, and read the pending state from the contract,
execution evidence and workspace. It must never treat a status line or native harness mode as
consent.

1. **Read the ledger when present and resolve the plan.** If `<project>/_support/PROJECT.md`
   exists, use its `## Current handoff`; if the workspace requires a ledger and it is missing,
   stop with the bootstrap recommendation. A standalone repository may use an explicit plan path
   under its natural `_support/sessions/` home without a ledger. Resolve the explicit `$ARGUMENTS`
   path first; otherwise use the handoff pointer, then enumerate session plans with an executable
   state. Use that fallback only when exactly one candidate exists. If more than one active plan
   exists, list all candidates and stop, even when one is newer. A zero-argument invocation with
   no unique candidate reports the candidates and stops. `Session: no` permits conversation-only
   execution but makes no durable cross-harness or resume claim.
2. **NORMALIZE before approval.** A legacy or organic plan may be adopted only after it is
   normalized into the current single-file contract, authorization and execution sections. Keep
   the original contract content intact, add only the metadata needed by the current format, and
   compute the contract digest after normalization. Do not infer consent from `Status: planned`,
   file location, a native Plan Mode approval or a prior command. If the exact normalized
   contract is unchanged and the current user request explicitly authorizes implementation for
   its covered scope, record that request as the new `implement` grant evidence and continue; if
   it does not cover the exact scope, clarify the scope before any project edit. A legacy status
   never supplies that grant. Completed legacy plans remain untouched. For an active legacy plan,
   resolve the `Test approach:` of each pending task during reconciliation; do not silently add a
   material verification requirement to its frozen contract. A material change to the task's
   verification requires a new contract revision and approval.
3. **Inspect before validating the requested next action.** Run `plan.py inspect` and require the
   current contract digest plus authorization status `valid` or `conditional`. For verification,
   inspect is the authority check: the observed plan status must be `built` or `verified`, and no
   `implement` grant or `validate --action verify` call is required. An explicit `verify` request,
   or an explicit normal `/flow-build` invocation reaching the Gate, supplies verification
   authority. For implementation and delivery, run the shared read-only action validator after
   inspection, then confirm that the authorization covers the action's targets:

   - **Execute** (`planned`/`building` with pending tasks) requires a current `implement` grant.
     It must name the implementation action and cover the target repositories/files.
   - **Verify** (the `verify` subcommand, or the verification gate reached by an explicit
     `/flow-build` request) uses the inspected current authorization and observed `built` or
     `verified` state. It does not require an `implement` grant and never edits or publishes.
   - **Deliver** (`verified` with a named pending commit, push, PR, merge or deploy) requires the
     matching delivery grant and target. It does not require a still-active `implement` grant once
     implementation is complete; revoking implementation does not revoke an already-scoped
     delivery grant unless the user revokes that grant too.

   A prior delivery grant survives resume only while its digest, target and conditions match. A
   condition already granted may be marked satisfied with fresh evidence without new approval;
   changed, expired or failed conditions stop the action. Mandatory safety and repository gates
   still apply. A hash proves content integrity, not who originated the consent.
4. **Check recovery before repeating meaningful work.** For a task, delegation, fix, review
   round, or remote operation that may be repeated, call the read-only helper with its canonical
   `kind` and semantic `scope`:

   ```sh
   python3 path/to/plan.py recovery-check <plan> --kind <kind> --scope <scope>
   ```

   A present recovery block with `attempts: []` is an initialized empty history and may continue.
   An absent block is legacy `unknown`; reconcile observed state and initialize the block before
   repeating work. An open or ambiguous attempt stops the repeat until its workspace, Git,
   process, or remote result is reconciled. A `completed` result closes the same review
   `kind`/`scope`, including after a contract digest change; an exhausted budget also stops that
   scope. Counters remain attached to the same semantic `kind` and `scope` across contract digests;
   an unrelated scope has an independent budget. This check never supplies consent or satisfies an
   action grant.
5. **Read observed state.** Reconcile the execution table, mutable `Test evidence` table,
   recovery attempts, filesystem, Git diff/log and verification evidence. A task marked complete
   without its declared test-approach evidence is pending. A `tdd` task with missing observed RED
   and no recorded fail-to-pass comparison is pending that comparison (agent work) or, where it
   cannot be produced, an explicit user resolution; preserve the workspace and do not infer RED
   from a later green run. Completion is labeled `fail-to-pass` or `exception-accepted`, each with
   pass-to-pass evidence and applicable independent checks recorded. Under `hold`, a clean commit
   is not required: verified working-tree
   changes and their commands are valid evidence. Reuse test evidence only while its covered code,
   configuration and toolchain remain current. Do not rewrite the frozen contract to record
   progress; update only the execution and test/recovery metadata sections; authorization changes
   follow the grant rules above.
6. **Determine the pending stage.** Use the state and observed evidence as follows:

   | State + observed evidence | Do |
   |---|---|
   | `draft` | Report the missing approval and stop; no project edit |
   | `planned` with pending tasks and an `implement` grant | Run `references/execute.md` |
   | `planned`/`building` with pending tasks and no `implement` grant | Report the missing implementation authority and stop |
   | `building` with an `implement` grant | Resume from the first task whose test approach or execution evidence is absent; never repeat verified work |
   | `built` | Run `references/verify-gate.md` after validating verification authority |
   | `verified` with a named, authorized delivery still pending | Reconcile only that delivery action; `verify` never performs it |
   | `verified` with delivery pending but no matching grant | Report the missing delivery authority and stop |
   | `verified` with no pending delivery | Report that the plan is complete |

   The observed tree is authoritative for actual files; the plan is authoritative for intended
   scope and action permissions. When they conflict, stop and surface the conflict instead of
   guessing which side to keep.
7. **In-vivo timing.** If the plan has `in-vivo: yes` or `design-review: yes` tasks, defer browser
   walks to `built` by default so the change-group is judged as a whole. Run them inline only when
   independent gated tasks would materially benefit from early feedback or the plan explicitly
   chooses that timing. Persist the choice in execution state for this run; it does not modify the
   frozen contract.

`verify` performs steps 1–7, skips implementation, and enters the verification gate when the
observed state is `built`. Its explicit invocation is the verification authority; it never creates
a delivery grant and never publishes.
