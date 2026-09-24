---
name: "review-security"
description: "Review security boundaries and detect exposed secrets without leaking or rotating them. Use when a change touches a security boundary: authentication, authorization, untrusted input, command or process execution, secrets, cryptography, dependencies, or data exposure."
model_profile: "inherit"
access_profile: "observe"
---

# review-security

1. Identify the authorized threat surface, assets, trust boundaries, and reachable entry points before scanning or testing. For changes to inputs, permissions, command or process execution, public data exposure, or dependencies, read [the flow-build reference](skill:flow-build/references/security-boundaries.md); resolve the skill through the host catalog or an explicit task path and report missing access.

2. Trace a suspected issue from source to sink with exploitation preconditions and realistic impact. Check current authoritative advisories when version-specific claims matter.

3. Use existing scanning tools when useful. Report secret types and locations without reproducing values; distinguish test fixtures and false positives.

4. Do not exploit live systems, rotate credentials, rewrite history, or remediate source without separate authorization. Return prioritized evidence and explicit coverage limits.

Stay within the parent’s assigned scope, write ownership, and output destination. Return evidence, limitations, and any decision needed from the parent.
