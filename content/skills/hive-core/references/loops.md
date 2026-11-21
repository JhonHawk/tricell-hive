# Loops

Read before the first cycle of work that repeats changing and measuring until a condition holds. Such work is cycles of changing something and measuring it, such as a [perf](../../hive-mode/playbooks/perf.md) goal or a loop under [unattended delegation](../../unattended-delegation/SKILL.md).

1. **Before the first cycle, state:**
   - the measure: an observable check run every cycle, the direction that counts as better and the threshold that counts as done;
   - the cap: cycles, time or tokens;
   - the guards: the checks that must pass after every cycle, such as the existing tests of the code the loop changes and each invariant the user named.

   State the measure, the cap and the guards in the activation message when unattended, otherwise in the plan. Without a measure and a cap, the loop does not start. Run the measure and the guards once on the unchanged state, and log the measure's value there as the baseline, the loop's first row. When a guard fails there, the loop does not start until it passes: that failure is a defect that blocks the work, handled as [reasons and blockers](reasons-and-blockers.md) sets.
2. **Each cycle** makes one change, then runs the measure and the guards. Keep the change only when the measure improves on the last kept value, or on the baseline before any change is kept, and every guard passes; otherwise revert it whole. A change, or a run of the measure or a guard, the baseline run included, that loads or changes a system outside the working tree is an effect: take it as the [effects round](external-effects.md) sets, or [unattended effects](unattended-effects.md) under unattended delegation. Log the cycle, with its change, measure, guard results and whether it was kept, in the [decision log](../../unattended-delegation/SKILL.md#decision-log) when unattended, otherwise in the [checkpoint](checkpoint.md).
3. **Never relax what judges the loop.** Do not lower the threshold, change how the measure is taken, or weaken, skip or remove a guard to reach done; only the user changes them.
4. **On a plateau, change approach.** After two cycles in a row with no measurable improvement, do not stop: try a different kind of change, such as another category of hypothesis, a combination of near-misses already measured one by one, whose log row names each part it joins, or rereading the code where the cost is, and name the new approach in the log.
5. **Stop** when the measure reaches its threshold with every guard passing, when the cap is reached, or when the log shows three approaches, the first and two changes of approach, each ending in two cycles without measurable improvement. In the last two cases, report what was tried, what failed, the current measure and what is needed to continue. Record the final status.
