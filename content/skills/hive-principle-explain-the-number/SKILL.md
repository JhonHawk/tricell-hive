---
name: hive-principle-explain-the-number
description: "Hive principle leaf: read before you measure, and before you trust, report or act on a measured number such as a speedup, latency, throughput or eval score. Find its limiter and what else it measured."
disable-model-invocation: true
---

# Explain the number

Read before you measure, and before you report or act on a number you measured.

1. Write the decision before the run: what result would change it, and how large the noise is. When the spread between runs is larger than the effect, the number decides nothing.
2. Ask why the result is not twice as good: name the resource or code path that bounds it, from a profile or counters taken during a run. A guess from reading the code is not a limiter.
3. List what else the number could be measuring, such as errors, skipped or cached work, code that never ran, an unequal setup, or a piece too small to matter, and rule each out with evidence.
4. Report the number with its run count, its spread and its limiter.

For an evaluation score, also ask whether every run did the task and whether the gap holds across trials. A number you cannot explain is reported as unexplained, not as a result.
