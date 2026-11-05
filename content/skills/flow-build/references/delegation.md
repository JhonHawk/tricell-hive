# Delegation

Read before deciding whether a child does a piece of work, in any flow, and before briefing one. What a child may do is the delegation rule of the global guidance's «Scope and authorization».

## When a child does the work

A child does the work for the first of these reasons that applies; when neither applies, the main thread does the work.

1. **Independence: every review.** Every review or verification Hive launches, by a `hive-review-*` or `hive-verify-*` role, always runs in a child, never in the main thread, because the context that produced the work biases its review. The cost of that child does not decide. A code review by the host's native command, when the user chose it in the delivery round, runs as that command runs. The checks the main thread runs for each task are not a review.
2. **Cost: large work the session would carry.** What the main thread reads or writes rides along in each of its later calls, while a fresh child starts at tens of thousands of tokens and returns a short summary. Delegate:
   - in research, each aspect as [flow-research step 4](../../flow-research/SKILL.md) sets;
   - in a build, a task that needs more than 5 large reads or edits more than 3 files, when its interface is settled and its writes do not overlap files another task in progress writes;
   - in any other flow, a check or collection that needs more than 5 large reads.

   A large read is defined in the global guidance's «Proportionality». Work below these limits stays in the main thread, and so does a task whose writes overlap files another task in progress writes.

## When children run at the same time

- Read-only children run side by side, launched in the same step: reviews, research aspects, and the collections or sweeps a flow sends to parallel children, such as flow-status's independent collections and security-advisory's sweep.
- Children that write, such as a backend task and a frontend task each delegated for its cost, run side by side only when all of these hold; otherwise they run one after another, each still in its child:
  - the contract they share is written as a typed artifact, such as types, a schema or an API definition, and does not change while they run;
  - their writes touch disjoint files;
  - the main thread does not need either result to continue;
  - an integration check across the seam, such as a contract test or a walk of the flow that crosses it, runs after both finish.

## Every child

- Brief it with the decision its result informs, what it covers (its aspects, targets or task), what is already settled, and when to stop; a role that names its own inputs, such as `hive-verify-task` (the change folder, the task ID and the diff base), also gets those. A reviewer's brief follows [Reviewer brief](#reviewer-brief) instead, with nothing settled about the work it reviews. Any child that signs in, reviewer or not, also gets the rows of the accounts it uses, as the global guidance's «Test accounts» sets.
- Wait for children with the host's wait tool, which returns when a child finishes or sends a message; never with a loop of sleeps and status checks.
- A child does not commit; the main thread checks the evidence it returns, not its assertions, before using it, and owns the commits.

## Reviewer brief

Give a reviewer what it needs to judge the work and nothing that argues for it:

- the artifact: the diff, the plan files or the captures;
- the requirements it must meet: the request, the ticket's criteria and the acceptance criteria;
- the output of the checks already run, as their commands and results.

Never include the producer's reasoning, its account of why the work is right, a claim such as "verified" or "tests pass" without its output, or a list of facts or conclusions the reviewer should not re-check, such as "already settled" or "facts already gathered", since a framing that the work is correct lowers what a reviewer finds. A UX reviewer covers the regions the change alters and their possible impact on the layouts and other components that contain or reuse them, such as containers, lists, dialogs and shared components, measured and exercised on the captures in its brief. Its brief also gives how to reach each state its role always checks, loading, empty, error and the longest real values, such as the fixture route, the account and how to force the error, or says which of them cannot be reached.

## Build children

- Pick the role by surface: `hive-build-backend`, `hive-build-frontend`, `hive-build-data`, `hive-build-infra`, `hive-build-kmp`, `hive-write-tests` or `hive-write-docs`.
- The brief names the change folder and the task ID and tells the child to read the plan files before editing; it does not paraphrase the task or its criteria.
- The main thread runs the task's checks, commits and marks the task.
