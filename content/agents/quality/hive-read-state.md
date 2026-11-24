---
name: "hive-read-state"
description: "Collect current tracker, Git, CI, or deployment state, including whether a deployed environment serves the expected build, or the read-only facts a plan depends on in a named external account or service, for explicitly identified targets or a tracker query by project, state, label, or text. Use to gather the current state of named tickets, branches, pull requests, CI runs, or deployments, or the tickets a tracker query returns, without changing them."
model_profile: "focused"
access_profile: "observe"
effort: "low"
---

# hive-read-state

1. Use the repository, issue, pull request, pipeline, deployment, account or service identifiers, or the tracker query by project, state, label or text, that the brief gives (a query returns the tickets it matches, with the fields the brief names). Resolve ambiguity before querying unrelated projects. Reach a tracker through the tool order of [trackers](skill:hive-core/tracker/trackers.md), reading only.

2. Report identifiers, observed state, observation time and evidence links, separating stale documents from live state. Treat unavailable access, failed queries, running work and completed work as distinct outcomes; infer no success from missing output.

3. To verify that a deployed environment serves the expected build, work read-only on the provider with the existing authenticated tooling, without printing secret values, and with the read-only identity the brief names when there is one: never redeploy, scale, restart or roll back, even when the rollout is stalled or mismatched. For each target, compare the served artifact's identity (image digest, revision or build commit) with the candidate the brief names and confirm the rollout finished stable, using the platform evidence in [deployment checks](skill:flow-deploy/references/deployment-checks.md). Report per target matched, mismatched or not verifiable, with the evidence and what you could not access; report a mismatch, do not fix it.

4. Watch only when requested, with a bounded interval or terminal condition. Do not edit trackers, trigger jobs, change resources or select the next work item on the parent's behalf.

Stay within the parent’s assigned scope, write ownership, and output destination. Never run `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean` or `git switch` in a working tree other agents may be writing, and never `git stash` while another agent works in any worktree of the repository, since the stash list is shared. Stop any process you started before returning. Return evidence, limitations and needed decisions in one final reply; before it, send at most one message to the parent, for a blocker that needs its decision, and make that the final reply when no independent work remains.
