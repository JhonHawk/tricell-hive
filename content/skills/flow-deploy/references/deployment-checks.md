# Deployment checks

Read by [flow-deploy](../SKILL.md) before it reconciles a promotion, deploys, or verifies what an environment serves. Verification depth for the changed surface stays in [verification by risk and surface](../../flow-build/SKILL.md). These checks grant no deployment, publication, destructive effect or external access.

## Predeploy

Reconcile the complete promoted candidate, affected destinations, migrations, compatibility, recovery and reusable prior evidence. A promotion label does not prove the candidate was reviewed. Before a production deployment of code or infrastructure, record its rollback triggers in the plan, or in the conversation when there is none: each signal to watch, such as error rate, latency, or failing health checks, with its threshold and observation window, and the recovery action it calls for, such as redeploying the previous revision, with the target the action acts on and how to revert it, or a fix-forward, which step 4 of the skill routes as a new run. Step 4 of the skill asks for those actions with the production authorization. Decide them before the deployment, not while watching it. Stored-data changes also follow the abort conditions in [persistent data changes](../../flow-build/SKILL.md).

## Regression and release gates

At integration or promotion, satisfy the project's broader regression and release gates. Confirm what CI actually executes; do not assume it covers a missing check. Local checkpoints, branch publication, merge and release are distinct boundaries, governed by the project's policy and existing authorization.

## Postdeploy

Verify active artifact/revision and routine health per affected surface. The functional walk follows [the functional walk](#functional-walk) below.

### Served identity and health

After authorized deployment, identify the artifact or revision actually served in each affected target and check routine health: health endpoints and the project's smoke checks. Compare the signals against the rollback triggers recorded at predeploy; report a trigger that fires and run its recovery action as step 6 of the skill sets. An HTTP smoke proves liveness, not functional behavior. Prove on the provider itself, not from the pipeline result, that the served artifact is the integrated candidate: its identity, such as the image digest, revision, or build commit, matches what was built from the merged commit, and the rollout finished stable. A green pipeline does not prove this, because a rollout can stall, roll back, or keep old instances serving. Send this check to a `hive-build-infra` child with the candidate identity and, for each target, the concrete provider resource to inspect, such as the ECS cluster and service, the Cloud Run service, the Vercel project, or the VPS host, taken from the project's `AGENTS.md` or deployment configuration; an environment name alone is not enough. That role can write, so the brief forbids any provider change, such as redeploying, scaling, restarting, or rolling back, and passes a read-only identity when the project has one; a stalled or mismatched rollout is reported, not fixed. Evidence by platform includes, for example:

| Platform | Evidence that the change is applied |
| --- | --- |
| Container service, such as AWS ECS | The service runs the new task definition, whose image digest is the candidate's; the deployment is completed; running tasks equal desired tasks; no new task stopped with an error; startup logs are healthy. |
| Kubernetes or serverless containers, such as GKE or Cloud Run | The revision or rollout receiving all traffic references the candidate digest; the rollout is complete and pods or instances are ready. |
| Managed frontend hosting, such as Amplify or Vercel | A deployment of the exact commit succeeded and is the one active on the domain. |
| VPS host, such as Hetzner, with Docker, Compose, or systemd | The running container has the candidate digest and restarted after the deployment; the unit is active; the health endpoint reports the build commit or version. |

A mismatch or an unfinished rollout leaves the deployment not verified, as step 6 of the skill sets, whatever the pipeline reports.

### Functional walk

The walk runs only when the project declares it, in its `AGENTS.md` or in the plan of the work, or when the user asked to validate the deployment; a promotion or deployment request alone does not ask for it, and you do not offer it. Otherwise report it as not run. Before it, ask for the effects it needs in that environment, such as the accounts, records and configuration it creates, as [effect authorization](../../hive-core/SKILL.md) sets; an answer left open leaves them unauthorized and the walk blocked. Run it only after the served identity is verified. Send the functional walk to a `hive-verify-change` child against the environment's URLs, plus `hive-review-ux` when a UI changed, each with its own account or session; brief `hive-review-ux` with the [UI gate's exception](../../flow-verify/SKILL.md) for pre-existing findings and deployed baselines. Launch both before waiting on either, not one after the other: neither needs the other's result. The orchestrator does not script its own walk. A walk that was due and did not run is reported as not verified, and its tickets stay out of their completion state until it runs. Recovery exercises stay within their authorized scope. Use appropriate readbacks, logs, metrics, traces, or journeys; a successful pipeline or an available shell alone does not establish service health. Check recovery against the data/configuration changes made, not only the previous code version.
