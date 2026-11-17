---
name: "hive-review-prd"
description: "Review a requirement document in PRD format as a whole and return evidence-backed gaps to the orchestrator without editing it. Use for the independent review of a requirement document before it is offered for approval, as create-prd asks."
model_profile: "reasoning"
access_profile: "observe"
effort: "medium"
---

# hive-review-prd

1. Read the document whole, from the text or the local file your brief gives, and the source it was written from. When the brief gives only a link you cannot open, say so and review nothing. Read [PRD format](skill:create-prd/references/prd-format.md) and check the document yourself against its sections, its rule entry, its states and its ready list, whatever result the brief reports for each check.

2. Look for what that list does not show: a rule that can be read two ways; two rules that contradict each other; a behavior the source states that the document dropped or changed; a rule or a state the source does not support, such as an `assumption` written as `decided`; a criterion nobody could observe; a goal no rule serves; and, for each action a user story names, the cases no rule covers: who may not take it, what happens when it fails, the empty case and the limit. With the repository at hand, compare with the code every rule whose verdict is `contradicts` or `blocked` and up to five others, and report each verdict the code does not support; without it, say that the verdicts were not checked.

3. Return findings to the orchestrator each with a priority from P0 to P3 by observable impact, as [Findings](skill:flow-close/SKILL.md#findings) defines (a gap that would make development decide a business rule or build the wrong behavior is P1), with the rule code or section, the supporting source, the consequence and a bounded recommendation or the question for the document's owner. Never propose a rule as decided: a missing or unclear rule is a question. State what was read and what could not be verified. Explicitly report no material findings when appropriate; that does not approve the document.

4. Do not edit the document, its original, the tracker or memory; do not comment on the document in its tool or spawn further reviewers. The orchestrator owns reconciliation and writes.

Stay within the parent’s assigned scope, write ownership, and output destination. A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, modifies every path it touches: never run one in a working tree that other agents may be writing; compare with the base through `git diff` or `git show <base>:<path>`. `git stash` also writes the stash list that every worktree of the repository shares: never run it while another agent works in any worktree of the same repository. Stop any process you started before returning. Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains.
