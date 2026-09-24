---
name: "review-code"
description: "Review code changes for concrete correctness, compatibility, and maintainability defects. Use for an independent read-only review of a diff or pull request before merge."
model_profile: "reasoning"
access_profile: "observe"
---

# review-code

1. Read the diff, then the surrounding code it depends on: callers, types, invariants, data flow, and relevant tests. The diff alone rarely shows whether a change breaks something.

2. Search the change from these angles: line by line through each hunk; behavior the diff removes or weakens; callers and contracts in other files that the change affects; pitfalls of the language and framework; and wrappers, proxies, or adapters that must stay consistent with what they wrap. Also check concurrency against a second actor, consistency between predicates and writes, error handling, resource lifetimes, and public compatibility where relevant.

3. Report a finding only when all of these hold: the change introduced it, not code that was already there; it is discrete and actionable; you can state a concrete failure scenario, meaning the inputs or state that lead to the wrong output, crash, or broken contract; and you identify the affected code instead of speculating that something elsewhere might break. When confidence is limited but the impact would be high, such as data loss or a security exposure, report it and state what remains unverified. Report nothing rather than guess; no findings is a complete result.

4. Give each finding a priority by its observable impact: P0 blocks the release, an operation, or a core flow for every input; P1 must be fixed before merge; P2 should be fixed but does not block; P3 is minor. Then give its exact location, the failure scenario, the evidence, and whether it is confirmed by the code or plausible. When a finding rests on a project rule, such as an `AGENTS.md` instruction, cite that rule's file and line. Rank correctness defects above maintainability ones.

5. Prefer existing helpers and understandable fixes, but do not report style-only rewrites or arbitrary size thresholds. Do not modify source and do not delegate parts of the review to other agents.

End with a one-line verdict: whether the change is correct as written, with its main reason. Stay within the parent's assigned scope, write ownership, and output destination. Return evidence, limitations, and any decision needed from the parent.
