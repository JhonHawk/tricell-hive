---
name: hive-rule-blast-radius
description: "Hive rule leaf: read before the review of a candidate. Finds what the change can break beyond the diff and proves, with real code, the one fact that makes it safe."
disable-model-invocation: true
---

# Blast radius

Read before the review of a candidate, and when the user asks what a change could break. A list of callers is not the job: a search returns it in seconds. The job is the breakage a search does not show.

A write-up that sounds right proves nothing. Do not hand back prose alone: find the one or two facts the safety rests on and run code to test them.

## Steps

1. Read the change: the symbols it adds, changes and deletes, and what behaves differently now, including what the diff does not spell out.
2. Name the one fact the change is safe because of, such as "this call only drops entries that are already dead". If it holds, most risky cases clear at once. Spend the effort here, not on a long list of maybes.
3. Look where a search stops: the source of each library the change calls, at the pinned version and with any local patch; timing, such as teardown order or async scheduling; data another reader consumes, such as a returned payload, a column, a wire format, a file another language reads, or a feature flag; and code several calls downstream.
4. Rate each risk by how likely it is and what it costs when it happens. Cite a real `file:line` for each. A search that finds nothing is an answer; never invent a caller or an API.
5. Prove the one fact with a script or test that runs the real code and fails loudly when the fact is false. Run it and keep its output.

## Evidence ladder

Take each fact as far down this list as is cheap, and say where it stopped:

1. Asserted only. Worth nothing alone.
2. Pointed at a line, with its `file:line` or the library's own source.
3. Shown that the failure path cannot be reached, step by step.
4. Ran it: a script or test calls the real code.
5. Reproduced it in the running application.

## Hand back

- **What it does,** including the non-obvious part.
- **The fact it is safe because of,** the rung it reached, and the output of the run; write `unproven` when no run could test it.
- **Risks,** each with how it breaks, `file:line`, likelihood and cost, and how to check it.
- **Cleared,** what was checked and why it is fine.
- **Before merge,** the cheapest check that would catch the real bug, with the script written.

A risk that needs a change joins the review's findings; a fix inside the candidate's scope follows [defect fix](../hive-rule-defect-fix/SKILL.md). Write the hand-back as [prose](../hive-rule-prose/SKILL.md) sets.
