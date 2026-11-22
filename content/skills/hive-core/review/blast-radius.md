# Blast radius

Read before a review, for impact beyond the diff. The review is that of a candidate. Also read it when the user asks what a change could break. A list of callers is not the job, since a search returns it in seconds; the job is the breakage a search does not show.

Prose that sounds right proves nothing: find the one or two facts the safety rests on and run code to test them.

1. Read the change: the symbols it adds, changes and deletes, and what behaves differently now, including what the diff does not spell out.
2. Name the one fact the change is safe because of, such as "this call only drops entries that are already dead". If it holds, most risky cases clear at once; spend the effort here.
3. Look where a search stops: the source of each library the change calls, at the pinned version and with any local patch; timing, such as teardown order or async scheduling; data another reader consumes, such as a payload, a column, a wire format, a file another language reads or a feature flag; code several calls downstream.
4. Rate each risk by likelihood and cost, with a real `file:line`.
5. Prove the one fact with a script or test that runs the real code and fails loudly when the fact is false. Run it and keep its output.

Hand back what the change does (including the non-obvious part); the fact it is safe because of, with the run's output, or `unproven` when no run could test it; each risk with how it breaks, `file:line`, likelihood, cost and how to check it; what was cleared and why; and the cheapest check before merge, with its script written. A risk that needs a change joins the review's findings; a fix inside the candidate's scope follows [defect fix](../build/defect-fix.md). Write the hand-back as [prose](../build/prose.md) sets.
