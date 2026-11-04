# Delegation

Read before deciding whether a child does a piece of work, in any flow, and before briefing one. The global guidance's «Scope and authorization» keeps the boundaries: a child starts with a fresh context, not a fork of the conversation, stays in its working tree, paths and effects, and never pushes, opens or merges pull requests, writes trackers or launches children; a child also does not commit. The main thread coordinates and checks what children return.

## When a child does the work

Delegate for the first of these reasons that applies; when none applies, the main thread does the work.

1. **Independence: every review.** A code, UX, plan or security review and a task verification always run in a child, never in the main thread, because the context that produced the work biases its review. The cost of that child does not decide. Brief it as [Reviewer brief](#reviewer-brief) sets.
2. **Parallelism: independent work at the same time.** Give pieces of work that can run side by side, such as independent read-only collections, each to its own child, launched in the same step. When they write, such as a backend task and a frontend task, do so only when all four hold:
   - the contract they share is written as a typed artifact, such as types, a schema or an API definition, and does not change while they run;
   - their writes touch disjoint files;
   - the main thread does not need either result to continue;
   - an integration check across the seam, such as a contract test or a walk of the flow that crosses it, runs after both finish.

   Otherwise run them one after another. A task the main thread can finish in 5 tool calls or fewer stays in the main thread.
3. **Cost: large work the session would carry.** What the main thread reads or writes rides along in each of its later calls, while a fresh child starts at tens of thousands of tokens and returns a short summary. Delegate:
   - in research, each aspect as [flow-research step 4](../../flow-research/SKILL.md) sets;
   - in a build, a task that needs more than 5 large reads or edits more than 3 files, when its interface is settled and its writes do not overlap files another task in progress writes;
   - in any other flow, a check or collection that needs more than 5 large reads.

   A large read is defined in the global guidance's «Proportionality». Work below these limits stays in the main thread, and so does a task whose writes overlap files another task in progress writes.

## Reviewer brief

Give a reviewer what it needs to judge the work and nothing that argues for it:

- the artifact: the diff, the plan files or the captures;
- the requirements it must meet: the request, the ticket's criteria and the acceptance criteria;
- the output of the checks already run, as their commands and results.

Never include the producer's reasoning, its account of why the work is right, or a claim such as "verified" or "tests pass" without its output, since a framing that the work is correct lowers what a reviewer finds. A UX reviewer covers the regions the change alters and their possible impact on the layouts and other components that contain or reuse them, such as containers, lists, dialogs and shared components, measured and exercised on the captures in its brief.

## Build children

- Pick the role by surface: `hive-build-backend`, `hive-build-frontend`, `hive-build-data`, `hive-build-infra`, `hive-build-kmp`, `hive-write-tests` or `hive-write-docs`.
- Launch independent children in the same step, not in waves; tasks that share a file go one after another.
- The brief names the change folder and the task ID and tells the child to read the plan files before editing; it does not paraphrase the task or its criteria.
- Wait for children with the host's wait tool, which returns when a child finishes or sends a message; never with a loop of sleeps and status checks.
- The main thread checks the evidence a child returns, runs the task's checks, commits and marks the task.
