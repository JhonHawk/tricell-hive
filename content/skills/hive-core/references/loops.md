# Loops

## When to read

Read before starting or resuming work that repeats cycles of gathering context, acting and verifying until a condition is met: the iterative route of [flow-research](../../flow-research/SKILL.md), unattended work that repeats cycles, such as making a failing suite pass or migrating a set of files, or a loop the user asks the host to run. This reference sets the loop's contract; it grants no effect, and authorization, critical effects and promotion keep their own rules.

## Start the loop

At the start, state in visible text:

1. **Objective and measure.** The observable end state and the check that shows it, such as a command and its expected result. Prefer a quantitative check you can run every cycle. When the host offers a goal evaluator that judges the transcript, write the check so its output appears in the transcript.
2. **Cap.** A limit of cycles, time or tokens, set before the first cycle. When the host accepts a cap inside the goal, such as a turn limit, include it. Reaching the cap stops the loop.
3. **State.** Where the state lives, as the State section below sets.

A loop without a measure or a cap does not start: ask for the missing part when the user is present; otherwise report it and stop.

## Run each cycle

Each cycle gathers context, acts and runs the measure, then records the result in the state before the next cycle. When a result decides something consequential, such as marking the objective met, prefer a check by a child with fresh context, such as `hive-verify-task` or `hive-verify-change`, over your own reading.

**No progress.** When two cycles in a row bring no measurable improvement, change the approach, recording what changes and why; this happens once per loop. After that change, stop when the next cycle brings no improvement, or later when two cycles in a row bring none. A count of attempts alone does not show progress; only the measure does.

Stop when the objective is met, the cap is reached, the no-progress rule says so, a step needs an effect no answer or authorization covers, or a decision is genuinely the user's and the user is present. The rules that set the loop's authority, such as [unattended delegation](../../unattended-delegation/SKILL.md), say which decisions you take yourself.

## State

Keep the loop's state on disk and update it after every cycle, so a compaction, a stopped host or a new session can resume it. With a retained plan, it is the progress section of its `tasks.md`; otherwise, the saved-state destination that [artifact layout](../../workspace-conventions/references/artifact-layout.md) sets. Record:

- the objective, its measure and the cap;
- each cycle with its measure, and when the last cycle started;
- the decisions taken and each effect applied with how to revert it;
- the run that wrote it, such as its session ID, the ticket, the next step, and the status: `running`, `done` when the objective is met, or `waiting for the user: <what>` when the loop stops for any other reason, the cap included.

Commits serve as checkpoints only when the delivery answer authorizes commits. The state is a working record: it is not a ledger, not a retained research document, and not persistent memory, which keeps its own write cadence. It authorizes nothing: answers and authorizations come only from where their own rules set, such as a plan or the user's answers.

**Resume.** Before any cycle, read the state. A `running` status your own run wrote, such as after a compaction, continues, and so does one the user present asks you to resume, whatever the next sentence says. A `running` status another run wrote stops this one, without writing to the state, while its last cycle started within the loop's time cap, or within a day when the cap is not a time; past that, record `waiting for the user: interrupted run` and report it. A `waiting` status is terminal: run no cycle, report it again, and clear it only by the user's edit of the state or, with the user present, their message choosing to resume. A `done` status ends the loop; a later loop on the same work starts a new record and keeps the earlier one.

## Escalate and end

A run that stops because another run is active only reports that in one line and does nothing else, `flow-close` included. Whenever the loop stops for any other reason, record its final status. When it stops without meeting the objective, report what was tried, what failed, the current measure and what is needed to continue.

Then run [flow-close](../../flow-close/SKILL.md): as the completion point when the objective is met, as a pause otherwise.

## Host capabilities

Use the capability the host offers, named by what it does: a goal evaluator with a turn limit, a scheduler that runs a prompt at an interval or on a schedule, a stop hook that continues the session, or a wait with a completion signal. A scheduled or background turn activates no unattended delegation, as [unattended delegation](../../unattended-delegation/SKILL.md) sets. A loop such a turn runs with nobody present only measures, investigates and edits locally as `hold` sets; a decision that is the user's, a missing measure or cap, or an effect beyond this machine records `waiting for the user: <what>` and stops it. When the host offers none of these, run the cycles in the session, wait as [delegation and waits](../../flow-research/references/delegation.md) sets, and keep the state on disk. The host-specific names are recorded for maintainers, not here.
