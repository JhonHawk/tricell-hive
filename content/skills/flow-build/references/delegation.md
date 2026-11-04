# Delegation

Read before deciding whether a child does a piece of work, in any flow, and before briefing one. What a child may do is the delegation rule of the global guidance's «Scope and authorization».

## When a child does the work

Delegate for the first of these reasons that applies; when none applies, the main thread does the work.

1. **Independence: every review.** A review Hive launches itself, by `hive-review-code`, `hive-review-ux`, `hive-review-plan` or `hive-review-security`, and a `hive-verify-task` check always run in a child, never in the main thread, because the context that produced the work biases its review. The cost of that child does not decide. A code review by the host's native command, when the user chose it in the delivery round, runs as that command runs. The checks the main thread runs for each task are not a review.
2. **Parallelism: tasks that write at the same time.** Give tasks that write and can run side by side, such as a backend task and a frontend task, each to its own child, launched in the same step, only when all of these hold:
   - each task takes the main thread more than 5 tool calls;
   - the contract they share is written as a typed artifact, such as types, a schema or an API definition, and does not change while they run;
   - their writes touch disjoint files;
   - the main thread does not need either result to continue;
   - an integration check across the seam, such as a contract test or a walk of the flow that crosses it, runs after both finish.

   Otherwise run them one after another. Read-only work runs in parallel children only where its flow says so, such as flow-status's independent collections; elsewhere its cost decides.
3. **Cost: large work the session would carry.** What the main thread reads or writes rides along in each of its later calls, while a fresh child starts at tens of thousands of tokens and returns a short summary. Delegate:
   - in research, each aspect as [flow-research step 4](../../flow-research/SKILL.md) sets;
   - in a build, a task that needs more than 5 large reads or edits more than 3 files, when its interface is settled and its writes do not overlap files another task in progress writes;
   - in any other flow, a check or collection that needs more than 5 large reads.

   A large read is defined in the global guidance's «Proportionality». Work below these limits stays in the main thread, and so does a task whose writes overlap files another task in progress writes.

## Every child

- Brief it with the decision its result informs, what it covers (its aspects, targets or task), what is already settled, and when to stop; a role that names its own inputs, such as `hive-verify-task` (the change folder, the task ID and the diff base), also gets those.
- Launch independent children together, not in waves.
- Wait for children with the host's wait tool, which returns when a child finishes or sends a message; never with a loop of sleeps and status checks.
- Check the evidence a child returns, not its assertions, before using it.

## Reviewer brief

Give a reviewer what it needs to judge the work and nothing that argues for it:

- the artifact: the diff, the plan files or the captures;
- the requirements it must meet: the request, the ticket's criteria and the acceptance criteria;
- the output of the checks already run, as their commands and results.

Never include the producer's reasoning, its account of why the work is right, or a claim such as "verified" or "tests pass" without its output, since a framing that the work is correct lowers what a reviewer finds. A UX reviewer covers the regions the change alters and their possible impact on the layouts and other components that contain or reuse them, such as containers, lists, dialogs and shared components, measured and exercised on the captures in its brief.

## Build children

- Pick the role by surface: `hive-build-backend`, `hive-build-frontend`, `hive-build-data`, `hive-build-infra`, `hive-build-kmp`, `hive-write-tests` or `hive-write-docs`.
- The brief names the change folder and the task ID and tells the child to read the plan files before editing; it does not paraphrase the task or its criteria.
- A build child does not commit. The main thread runs the task's checks, commits and marks the task.
