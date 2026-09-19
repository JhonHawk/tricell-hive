# flow-build — CLOSE

Runs at the end of every `/flow-build` invocation, after the run's last stage. Close records
observed state and authorized delivery; it does not turn a successful verification into an
implicit publication.

1. **Tracker.** Move task states and attach evidence plus any authorized delivery references via
   the declared tracker access (`flow-core/references/tracker-access.md`). A tracker with no
   declared access remains unchanged; report that fact rather than inventing an integration.
2. **Git and delivery.** Report the current branch, working-tree state, commits and remote
   delivery observed. Under `hold`, a non-clean tree is an expected result and is not a close
   failure. In a normal run, reconcile a delivery action only after the observed plan is
   `verified` and its Authorization grant names that action and target; a grant on a `planned`,
   `building` or `built` plan remains pending. A `verify` invocation never reconciles delivery,
   even when its gate advances the plan to `verified`; leave the authorized action pending for a
   subsequent normal run. A `verified` plan with a `commit` grant lands its seam commits in
   this step, before the report: verified work left only in the working tree is a close
   failure outside `hold`, and whatever stops the commit is asked here — never noted after
   the user has ended the conversation. A Phase B finding or a red check the change caused
   reopens the plan through the declared path (`plan-format.md`, action/state compatibility:
   open `review` attempt → `building` → fix → affected gate → `verified`) before its fix is
   pushed — never an edit on a `verified` plan. If an authorized action remains pending, name the
   exact action and reason. Do not create a commit, push, PR, merge or promotion unless the Authorization section
   and mandatory repository gates permit it.
3. **Ledger, when present.** Update PROJECT.md with the phase, artifacts, decisions, authorization scope and
   delivery state. Finalize the session/initiative only when its lifecycle actually concluded;
   preserve `Session: no` as conversation-only and make no durable resume claim for it. Write
   `## Current handoff` with the paths produced, decisions left, input for the next phase and the
   next suggested action.
4. **Memory.** Upsert every living status record (`topic_key`) this run falsified — a delivery
   observed in step 2 supersedes the "open, not merged" fact written earlier, in the same close;
   it never waits for the user to declare the session over (`memory-routing.md`).
5. **Evidence.** Keep versioned reports written by the verification gate. Purge only raw,
   gitignored evidence whose retention rule allows it; do not delete open-finding evidence or
   anything needed to substantiate a pending delivery.
6. **Docs impact.** State the change-group's specs/product-doc delta: paths updated in the specs
   repo, or `none` with the reason. An operator-facing behavior with no required documentation
   delta is a gap to close in the same change-group.
7. **Report.** Distinguish tasks implemented, tasks verified, delivery actions completed and
   delivery actions still pending. Include paths and exact commands actually run, what was
   unverified or blocked, executor substitutions, servers started/stopped and the next
   suggested scope. `verified` is a quality result, not a delivery claim. Steps 1–6 are done
   before this report is written — a step announced and left open is a close failure. Every
   remaining item is classified by executor (`reporting-integrity.md > Fix at the Root`): work
   the agent can do that waits only on a confirmation (deleting test data, a gated merge) is
   ASKED here with its exact inventory and a recommendation, never listed as pending on the
   user.
