---
name: "hive-review-security"
description: "Review security boundaries and detect exposed secrets without leaking or rotating them. Use when the user asks for a security review or accepts one recommended for the change; it reviews changes that touch a security boundary: authentication, authorization, untrusted input, command or process execution, secrets, cryptography, dependencies, or data exposure."
model_profile: "inherit"
access_profile: "observe"
---

# hive-review-security

1. Identify the authorized threat surface, assets, trust boundaries and reachable entry points before scanning or testing. Read [Security boundaries](skill:hive-core/build/security-boundaries.md) before judging the change, since its list of boundaries sets what this review covers.

2. Trace a suspected issue from source to sink with exploitation preconditions and realistic impact; check current authoritative advisories when version-specific claims matter.

3. Run the project's existing scanning tools that cover the changed paths, and name any you did not run. Report secret types and locations without reproducing values, distinguishing test fixtures and false positives.

4. Do not exploit live systems, rotate credentials, rewrite history or remediate source without separate authorization. Return prioritized evidence and explicit coverage limits.

Stay within the parent’s assigned scope, write ownership, and output destination. Never run `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean` or `git switch` in a working tree other agents may be writing, and never `git stash` while another agent works in any worktree of the repository, since the stash list is shared. Stop any process you started before returning. Return evidence, limitations and needed decisions in one final reply; before it, send at most one message to the parent, for a blocker that needs its decision, and make that the final reply when no independent work remains.
