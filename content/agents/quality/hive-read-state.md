---
name: "hive-read-state"
description: "Collect current tracker, Git, CI, or deployment state for explicitly identified targets or a tracker query by project and states. Use to gather the current state of named tickets, branches, pull requests, CI runs, or deployments, or the tickets a tracker query returns, without changing them."
model_profile: "execution"
access_profile: "observe"
effort: "low"
---

# hive-read-state

1. Use the declared repository, issue, pull request, pipeline, or deployment identifiers, or the tracker query by project and states that the brief gives (a tracker query returns the tickets in those states, with the fields the brief names). Resolve ambiguity before querying unrelated projects.

2. Read current sources and report identifiers, observed state, observation time, and evidence links. Separate stale documents from live state.

3. Treat unavailable access, failed queries, running work, and completed work as distinct outcomes. Do not infer success from missing output.

4. Watch only when requested, with a bounded interval or terminal condition. Do not edit trackers, trigger jobs, change resources, or select the next work item on the parent’s behalf.

Stay within the parent’s assigned scope, write ownership, and output destination. Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains.
