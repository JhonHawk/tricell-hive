---
name: "hive-write-spec"
description: "Write source-backed specifications and documentation for an agreed scope. Use to draft or update requirement specs, change records, or technical documentation from verified sources."
model_profile: "execution"
access_profile: "implement"
effort: "medium"
---

# hive-write-spec

1. Locate the project's declared specs directory, existing documentation conventions, and authoritative code, decisions, and contracts.

2. Distinguish observed behavior, accepted requirements, proposed changes, and unresolved questions. Cite evidence precisely enough to resume verification.

3. Describe boundaries, examples, and acceptance criteria in proportion to the change, in the project's existing formats.

4. Place artifacts in the destination supplied by the parent or repository guidance, preserve useful legacy plan names, and keep implementation status truthful.

Before resolving needed project settings, read [project settings](skill:flow-plan/references/project-settings.md). Before creating or continuing a change record, read [change records](skill:flow-plan/references/change-records.md). Before writing a support artifact, read [artifact layout and retention](skill:workspace-conventions/references/artifact-layout.md). Before reporting findings with citations, links or externally defined IDs, read [reporting](skill:flow-close/references/reporting.md). Reading a reference does not invoke its owner's whole workflow.

Stay within the parent’s assigned scope, write ownership, and output destination. Commit only in your assigned working tree and branch, and only when the brief allows it; never push, open or merge pull requests, write to trackers, or launch children of your own. Modify or delete only the paths the brief names, list planned deletions before applying them, and stop and report when the work needs a change outside them. Run installs, builds, and other long commands in the foreground with a bounded timeout rather than behind a background monitor, and stop any process you started before returning. Return evidence, limitations, and any decision needed from the parent.
