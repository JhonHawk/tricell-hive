# Deployment checks

Read by [flow-deploy](../SKILL.md) before it reconciles a promotion, deploys, or verifies what an environment serves. Verification depth for the changed surface stays in [flow-verify](../../flow-verify/SKILL.md). These checks grant no deployment, publication, destructive effect or external access.

## Predeploy

Reconcile the complete promoted candidate, affected destinations, migrations, compatibility, recovery and reusable prior evidence. A promotion label does not prove the candidate was reviewed. Before a production deployment of code or infrastructure, record its rollback triggers in the plan, or in the conversation when there is none: each signal to watch, such as error rate, latency or failing health checks, with its threshold and observation window, and the recovery action it calls for, such as redeploying the previous revision, with the target the action acts on and how to revert it, or a fix-forward, which step 4 of the skill routes as a new run. Step 4 asks for those actions with the production authorization: decide them before the deployment, not while watching it. Stored-data changes also follow the abort conditions in [stored data](../../hive-core/build/stored-data.md).

## Postdeploy

### Served identity and health

After authorized deployment, identify the artifact or revision actually served in each affected target and check routine health: health endpoints and the project's smoke checks, compared against the rollback triggers recorded at predeploy; report a trigger that fires and run its recovery action as step 6 of the skill sets. An HTTP smoke proves liveness, not functional behavior.

Prove on the provider itself, not from the pipeline result, that the served artifact is the integrated candidate: its identity, such as the image digest, revision or build commit, matches what was built from the merged commit, and the rollout finished stable. A green pipeline does not prove this, because a rollout can stall, roll back or keep old instances serving.

Run the check:
- read-only, with the target, the expected build identity and, for each target, the concrete provider resource to inspect, such as the ECS cluster and service, the Cloud Run service, the Vercel project or the VPS host, taken from the project's `AGENTS.md` or deployment configuration; an environment name alone is not enough;
- with the project's read-only identity when it has one;
- in the main thread, or in a `hive-read-state` child given those inputs when [delegation](../../hive-core/delegation/delegation.md) sends it to one.

The check reports and never fixes: a stalled or mismatched rollout is reported, not repaired.

A mismatch or an unfinished rollout (a service not on the new task definition, a revision not receiving all traffic, a deployment of another commit active on the domain) leaves the deployment not verified, as step 6 of the skill sets, whatever the pipeline reports.

### Functional walk

The walk runs only when the project declares it, in its `AGENTS.md` or in the plan of the work, or when the user asked to validate the deployment; a promotion or deployment request alone does not ask for it, and you do not offer it. Otherwise report it as not run. A criterion whose `Observed in:` names this environment does not declare the walk. When the walk runs, it closes those criteria, as [flow-plan](../../flow-plan/SKILL.md#acceptance-criteria) sets; when it does not, report them as not verified.

Before the walk:
- read [test accounts](../../hive-core/local/test-accounts.md) before choosing its accounts and records, and drive a browser as [Browser tool](../../hive-core/ui/ui-verification.md#browser-tool) sets;
- ask for the effects it needs in that environment, such as the accounts, records and configuration it creates, as the global guidance's «Scope and authorization» section sets; an answer left open leaves them unauthorized and the walk blocked;
- run it only after the served identity is verified.

Run the walk:
- send it to a `hive-verify-change` child against the environment's URLs, plus `hive-review-ux` when a UI changed. Give each its own account. When they share one, at most one of them loads its saved state; the other signs in on a session of its own and saves nothing, as [Browser tool](../../hive-core/ui/ui-verification.md#browser-tool) sets;
- brief `hive-review-ux` as [reviewer brief](../../hive-core/delegation/reviewer-brief.md) sets, with the environment's URLs in place of the captures, and report the defects it lists as already there as findings, as [close report](../../hive-core/close/close-report.md#findings) sets, and each state it lists as not covered as not verified;
- run both side by side, launched as [delegation](../../hive-core/delegation/delegation.md#when-children-run-at-the-same-time) sets; the orchestrator does not script its own walk.

After it:
- a walk that was due and did not run is reported as not verified, and its tickets move as the `Completion` rule of [ticket state](../../hive-core/tracker/ticket-state.md) sets;
- recovery exercises stay within their authorized scope and are checked against the data and configuration changes made, not only the previous code version.
