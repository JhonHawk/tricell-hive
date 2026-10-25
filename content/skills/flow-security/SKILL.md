---
name: flow-security
description: Review a change's security boundaries with an independent reviewer and report prioritized findings without remediating. Use only when the user asks for a security review; not for advisory sweeps across repositories or for checking a package before installing it.
---

# Run the requested security review

## Entry and exit

This stage runs only on an explicit user request; what counts as one, and the rule for offering the review, are in [Offer the security review](../flow-build/references/security-boundaries.md#offer-the-security-review). Enter with the request and a candidate: a diff, a branch or a merged commit, or, for a reachability assessment that [security-advisory](../security-advisory/SKILL.md) offers, the application and the advisory. When both exist, enter directly, with no plan or build before it. Exit with prioritized findings and the coverage limits; this stage does not remediate.

## Prepare

Before resolving needed project settings, read [project settings](../flow-plan/references/project-settings.md). Before delegating or starting an external-work wait, read [delegation and waits](../flow-research/references/delegation.md). Before progress reports, completion reports or messages containing citations, links or externally defined IDs, read [reporting](../flow-close/references/reporting.md). To brief the reviewer and check its result, read [security boundaries](../flow-build/references/security-boundaries.md), which names the boundaries and the dependency check.

## Procedure

1. **Place the run.** When the user asks for it, run it on the current candidate before asking the user to validate or, when the delivery has no validation stop, before the push or merge. Once asked for, it also covers boundaries that appear later in the same change, at the same points. At the accepted-candidate gate of a walkthrough, rerun it only when the corrections touch a boundary. When the user asks after the change was merged, run it on the merged commit.
2. **Brief `hive-review-security`.** Give it the candidate (the diff and its base, or the merged commit), the boundaries the change crosses, the plan or the request that states the intent, and the path of [security boundaries](../flow-build/references/security-boundaries.md). For a reachability assessment, give it the application, the advisory, and the dependency path instead of a diff, and ask which entry points reach the affected code. Tell it to stay inside that scope, to read and test without exploiting anything, and to report a secret by its type and location only, never its value.
3. **Check the result.** Treat the findings as claims: each needs preconditions, impact and evidence from source to sink. Prioritize them P0–P3 by observable impact, as [review and gates](../flow-verify/references/review-and-gates.md) maps severity. A finding about a dependency's reachability follows [security-advisory](../security-advisory/SKILL.md) step 5, not a copy of it here.
4. **Report** the findings by priority, the boundaries covered and not covered, what was not run, and the coverage limits. Do not fix anything in this stage. Fixes for a merged commit go in a follow-up pull request of the same ticket, under the delivery already chosen for it. When a fix touches UI files, run the UI gate again on that diff before asking the user to validate, as [the UI gate](../flow-verify/references/ui-gate.md) sets.
