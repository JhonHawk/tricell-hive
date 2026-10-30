# Delegation and external-work waits

## When to read

Read before launching a child, before a tracker query that lists tickets, or before starting a wait for external work, including reviews, CI, builds and deployments. Children read this resource before dependent work when their assignment consumes this procedure. It does not invoke the research workflow or grant effects beyond the task brief.

## Split work

**Condition:** selecting work to delegate or changing the split.

**Actions:** the main thread coordinates work in proportion to the task. This is the one delegation criterion; skills point here. Keep small, understood tasks direct. Delegate a unit when its interface is settled, its writes do not overlap concurrent work, and its brief is shorter than the change itself; keep it in the main thread when it is coupled to work in progress. For a lookup, which changes nothing, the measure is its output: delegate the rest of a lookup that has not found its answer by the sixth search or read, the count the global guidance uses for a lookup outside a skill, and the probing the next sentence names. Send probing likely to produce long output, such as binaries, installed dependencies, build artifacts, long documentation, or broad searches, to a child that returns only the relevant findings, because early context is re-read for the rest of the session. One exception replaces this size test: the main thread delegates every tracker query that lists tickets, by project, state, label, or text, such as a backlog listing or a duplicate search, to the `hive-read-state` role whatever its expected size, because a listing's size is unknown until it returns; a child that reaches such a query runs it itself, since children launch no children. For a tracker kept in local files, such as a `BACKLOG.md` or a folder of ticket files, first measure what the query would read, such as with `wc -l`: read a short one directly, and send any query over a long one, a text search included, to the role like any other tracker query. The brief gives the query and the fields the parent needs, such as those a report groups by, or each match's title and description for a duplicate search, with the alternative wordings to try for a text search. Reading named tickets stays under the size test. Reassess the split when a direct lookup grows into diagnosis. Delegate independent bounded work; sequence dependencies and avoid overlapping writers. Launch independent children in the same step, not in waves. When several children work on one question, give each a disjoint scope and name in each brief what the others cover, so none repeats another's reads. When your own bounded lookup has not found where the answer lives, launch one child to locate it first and split the rest by what it finds.

## Select the role

**Condition:** the host lists a matching installed role.

**Actions:** select its exact native ID through the actual child-launch tool. Select a child by its deliverable and allowed effects, not just a language label. A task title or role name in the prompt does not select a native profile.

**Condition:** that surface lacks role selection or does not list the role.

**Actions:** give a bounded generic child the resolved canonical contract and disclose the fallback. A role the tool does not list counts as lacking role selection; read its contract from the installed role file. If the contract cannot be read, report the dependency instead of inventing it.

## Supply a bounded brief

**Condition:** launching a child.

**Actions:** give it the objective, repository or declared root, authorized effects, the files or paths it may modify or delete, relevant contracts, expected output, and acceptance evidence. Delegation grants no additional authority. Label delegated premises as inspected facts, source claims, or assumptions, with the relevant evidence or missing check.

Assign the state children would share besides source files; each brief, also for a child that runs alone, says whether its dependency installation is shared with another agent and which instance of a local test service it may use:

- **Dependencies.** No agent, the main thread included, runs a command that installs, prunes, or relinks a dependency installation while another agent builds or tests on it, in the same working tree or through a link from another tree, such as `npm ci`, `pnpm install`, or adding a package; the main thread runs them once no child uses that installation. Never link one tree's dependency directory into another tree where a child builds or tests.
- **Local test services.** Give each child whose test suite sets up, resets, or binds a local service, such as a test database, a cache, or a port, its own instance when the project can provision one, such as a separate or disposable database, a cache database index, or a free port. Otherwise launch the children that need the shared instance one after another, naming it in each brief, or have the main thread run those suites after the children return; a suite stopped because its service was busy follows the shared-service rule of [verification](../../flow-build/references/verification.md#frequency-and-coverage).
- **Served candidates.** A served candidate that children only walk or inspect stays shared, with the isolated data, accounts and browser sessions that [the UI gate](../../flow-verify/references/ui-gate.md) sets.

A child works only in its assigned working tree and named paths, lists planned deletions before applying them, and stops and reports when it needs a change outside them. A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, changes paths outside the child's own: the child never runs one in a working tree shared with other writers, and neither does the main thread while a child works in that tree; both compare with the base through read-only Git commands or a separate worktree. `git stash` also writes the stash list that every worktree of the repository shares, so neither the child nor the main thread runs it while another writer works in any worktree of the same repository. Unless its brief says its dependency installation is not shared, the child runs no command that installs, prunes, or relinks dependencies, such as `npm ci`, `pnpm install`, or adding a package, and it runs a test suite that sets up, resets, or binds a local service, such as a test database, a cache, or a port, only on the instance its brief assigns; when its brief assigns none, it does not run the suite and reports it as not run. The child may commit locally in its own working tree and branch only when the brief allows, and never pushes, opens or merges pull requests, writes to trackers, or launches children of its own, even when the delivery mode authorizes those effects; when it cannot finish alone, it reports what remains. It returns only its final result; before it, it sends the parent at most one message, for a blocker that needs the parent's decision, and returns that blocker as its final reply when no independent work remains. The main thread reviews its diff before integrating it and owns those effects.

## Deliver and resolve instructions

**Condition:** preparing a child brief or reading a required skill resource.

**Actions:** pass accessible project-guidance paths, task decisions, existing project conventions, and the exact skill or resource identities or paths the child needs, including the skills and project documents that the project's rules direct for this kind of work, and those a conditional rule activates when this host lists the skill or tool it names, by identity or path. When the launch tool offers to pass the parent's history, launch the child without it, since the brief carries what it needs; whether or not it inherits history, it does not receive the skills the parent read, so pass their paths. After the parent's context is compacted, a child still reads what its brief names. Availability, loading, compliance and effective permission or model controls are distinct facts: one does not prove another. Pass the smallest relevant set, not whole manuals or unrelated stacks.

Deployed roles and global guidance link installed paths, root-relative at project scope. Resolve any other skill through the host's catalog or loader or an explicit task path and read its resources, including a remaining `skill:owner/path` link, relative to that skill's directory, not the current directory or agent file. A `skill:` link is not a host command or auto-expanded URI, and reading a resource does not invoke the skill's whole workflow.

The child reads these sources before dependent work, and the skills, references and role files among these sources whole, as the global guidance sets. If resolution is ambiguous, inaccessible or unavailable, it pauses only the work requiring that source, continues independent work, and reports the unavailable source in its final reply with what it could not do; only when the source blocks the whole assignment is it the one blocker message the global delegation rule allows. An installed file or listed skill is not proof of reading.

## Launch and wait

This section applies to the main thread; a child runs its commands as its role sets.

**Condition:** children or other external processes are running.

**Actions:** continue useful independent work. Give each external process you start a completion signal, such as the host's monitor or watch tool, before your turn ends.

**Condition:** giving an external process a completion signal.

**Actions:** make the signal end on every outcome the process can reach: success, failure, a target that does not exist yet, and a process that hangs. When the host signals only when the process exits, also set a bounded watch, such as a monitor or timer that ends after one wait stretch, so a hung process still wakes you; on that wake, check whether it is making progress and apply the rule in [verification](../../flow-build/references/verification.md) for a run that makes no progress. Prefer the process's exit status, or a command that reports a final state, over matching its output; when you must match text, first check the pattern against the tool's actual success and failure output, including indentation. When the target may not exist yet, such as checks not yet registered on a pull request you just pushed, use one command that waits for the target to appear and then watches it to a final state, under the wait-stretch rule below. A watch that exits reporting no target, such as `no checks reported`, has not finished: rerun that same command, never a new short sleep-and-check cycle.

**Condition:** a child is expected to outlast one wait stretch and the host supports background launch with a completion signal.

**Actions:** launch it in the background.

**Condition:** no independent work remains.

**Actions:** wait on the completion signal instead of polling. When the host delivers that signal to a later turn, such as a child's result or a background command's completion notification, the main thread handles that turn as [reporting](../../flow-close/references/reporting.md#report-a-running-or-completed-item) sets for a turn a notification starts. When the host drops results that arrive after you return, collect them before returning.

**Condition:** the host needs an explicit blocking wait rather than delivering the signal.

**Actions:** keep one blocking wait, such as a subagent wait call or `gh run watch`, with its timeout set to a whole stretch of about five to ten minutes, or the host's maximum when lower, never a 30- or 60-second slice. Re-await the same process or rerun the same watch, never starting new short sleep-and-check cycles. A delegated child returns only its final result, as the global delegation rule sets.

## Check returned work

**Condition:** a child returns.

**Actions:** children return findings or changes, evidence, unresolved limits, and any required instruction-access gaps. The main thread checks the result against the brief and source evidence, resolves disagreements, and owns integration and closure. An agent's assertion is not independent verification. Relay every defect a child reports, at any severity scale it uses, or propose it under the global incidental-findings rule; never drop it from your summary. Before relaying a child's result that touched credentials, cookies, or tokens, check whether it put a secret in a durable or shared place, and if so tell the user under the global Secrets rule.

**Condition:** a child is silent longer than its expected duration, or a child returns.

**Actions:** keep the processes it starts identifiable. Before telling the user that a child is still in progress after a silence longer than its expected duration, check its liveness through its tool activity or processes; when a child returns, confirm it left no running processes: record each one it left, on purpose or not, such as a verifier's server, with its PID or port in the plan's progress section, or in the report without a plan; stop at once those no longer needed, and leave the rest for [flow-close](../../flow-close/SKILL.md) to stop at close.

**Condition:** a child returns, or the host reports that it stopped, while it waits on work of its own, such as a watcher, a wait or a background command.

**Actions:** treat that child as stopped, because its result may be provisional. Check at once whether the child and that work are still alive. When that work has finished, resume the same child so it returns its final result; when it is still running, give it a completion signal as [Launch and wait](#launch-and-wait) sets and resume the same child when it ends; when the work is gone or the host cannot resume the child, launch the same role again with the same brief, once. Never report it as in progress, and never wait on it in short check cycles.

## Record consequential delegation

**Condition:** the delegated work is consequential.

**Actions:** retain the requested role/path, native or generic dispatch, and observed model/access controls in the existing work record or handoff. Mark unavailable evidence as unverified. Reading a role's frontmatter does not apply its permissions or model; a configured profile, tool exclusion, or domain label is not proof of effective isolation, a different model, or skill loading. Report the actual control and account for parent overrides.

## Worked wait example

A background review has a host completion signal and an independent file can be inspected. Inspect that file rather than polling. Once no independent work remains, wait for the signal. If an explicit wait is required and its five-minute stretch expires, re-await that review, not a new reviewer. When it returns, check the cited evidence and include every reported defect before integration.
