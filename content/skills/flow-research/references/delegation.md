# Delegation and external-work waits

## When to read

Read before launching a child or starting a wait for external work, including reviews, CI, builds and deployments. Children read this resource before dependent work when their assignment consumes this procedure. It does not invoke the research workflow or grant effects beyond the task brief.

## Split and announce work

**Condition:** selecting work to delegate or changing the split.

**Actions:** the main thread coordinates work in proportion to the task. Tell the user in one sentence which work goes to which subagents and why; do not announce that work stays in the main thread. Keep small, understood tasks direct. Send probing likely to produce long output, such as binaries, installed dependencies, build artifacts, long documentation, or broad searches, to a child that returns only the relevant findings, because early context is re-read for the rest of the session. Reassess the split when a direct lookup grows into diagnosis. Delegate independent bounded work; sequence dependencies and avoid overlapping writers.

## Select the role

**Condition:** the host lists a matching installed role.

**Actions:** select its exact native ID through the actual child-launch tool. Select a child by its deliverable and allowed effects, not just a language label. A task title or role name in the prompt does not select a native profile.

**Condition:** that surface lacks role selection or does not list the role.

**Actions:** give a bounded generic child the resolved canonical contract and disclose the fallback. A role the tool does not list counts as lacking role selection; read its contract from the installed role file. If the contract cannot be read, report the dependency instead of inventing it.

## Supply a bounded brief

**Condition:** launching a child.

**Actions:** give it the objective, repository or declared root, authorized effects, the files or paths it may modify or delete, relevant contracts, expected output, and acceptance evidence. Delegation grants no additional authority. Label delegated premises as verified facts, source claims, or assumptions, with the relevant evidence or missing check.

A child works only in its assigned working tree and named paths, lists planned deletions before applying them, and stops and reports when it needs a change outside them. It may commit locally in its own working tree and branch only when the brief allows, and never pushes, opens or merges pull requests, or writes to trackers, even when the delivery mode authorizes those effects. The main thread reviews its diff before integrating it and owns those effects.

## Deliver and resolve instructions

**Condition:** preparing a child brief or reading a required skill resource.

**Actions:** pass accessible project-guidance paths, task decisions, existing project conventions, and the exact skill or resource identities or paths the child needs. It does not inherit the parent's conversation or skills already read. Pass the smallest relevant set, not whole manuals or unrelated stacks.

Deployed roles and global guidance link installed paths, root-relative at project scope. Resolve any other skill through the host's catalog or loader or an explicit task path and read its resources, including a remaining `skill:owner/path` link, relative to that skill's directory, not the current directory or agent file. A `skill:` link is not a host command or auto-expanded URI, and reading a resource does not invoke the skill's whole workflow.

The child reads these sources before dependent work and reports any unavailable one before proceeding with what depends on it. If resolution is ambiguous, inaccessible or unavailable, report the dependency and pause only the work requiring it; continue independent work. An installed file or listed skill is not proof of reading.

## Launch and wait

**Condition:** children or other external processes are running.

**Actions:** continue useful independent work. Give each external process you start a completion signal, such as the host's monitor or watch tool, before your turn ends.

**Condition:** a child is expected to outlast one wait stretch and the host supports background launch with a completion signal.

**Actions:** launch it in the background.

**Condition:** the host offers only a launch that blocks your turn.

**Actions:** first tell the user in one line what the child does, roughly how long it may take, and what its result decides.

**Condition:** no independent work remains.

**Actions:** wait on the completion signal instead of polling. When the host delivers that signal to a later turn, such as a child's result or a background command's completion notification, the main thread ends its turn with the in-progress item described in [reporting](../../flow-close/references/reporting.md#report-a-running-or-completed-item). When the host drops results that arrive after you return, collect them before returning.

**Condition:** the host needs an explicit blocking wait rather than delivering the signal.

**Actions:** keep one blocking wait, such as a subagent wait call or `gh run watch`, with its timeout set to a whole stretch of about five to ten minutes, or the host's maximum when lower, never a 30- or 60-second slice. Re-await the same process or rerun the same watch, never starting new short sleep-and-check cycles. After each stretch that ends without the result, the main thread gives the user one line naming what is still running and what its result decides. A delegated child still returns its final result, not interim status.

## Check returned work

**Condition:** a child returns.

**Actions:** children return findings or changes, evidence, unresolved limits, and any required instruction-access gaps. The main thread checks the result against the brief and source evidence, resolves disagreements, and owns integration and closure. An agent's assertion is not independent verification. Relay every defect a child reports, at any severity scale it uses, or propose it under the global incidental-findings rule; never drop it from your summary. Before relaying a child's result that touched credentials, cookies, or tokens, check whether it put a secret in a durable or shared place, and if so tell the user under the global Secrets rule.

## Record consequential delegation

**Condition:** the delegated work is consequential.

**Actions:** retain the requested role/path, native or generic dispatch, and observed model/access controls in the existing work record or handoff. Mark unavailable evidence as unverified. Reading a role's frontmatter does not apply its permissions or model; a configured profile, tool exclusion, or domain label is not proof of effective isolation, a different model, or skill loading. Report the actual control and account for parent overrides.

## Worked wait example

A background review has a host completion signal and an independent file can be inspected. Inspect that file rather than polling. Once no independent work remains, wait for the signal. If an explicit wait is required and its five-minute stretch expires, report: "The review is still running; its verdict determines whether this candidate may be published. I will continue waiting for the same review." Re-await that review, not a new reviewer. When it returns, check the cited evidence and include every reported defect before integration.
