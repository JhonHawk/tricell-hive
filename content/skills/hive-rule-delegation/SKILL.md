---
name: hive-rule-delegation
description: "Hive rule leaf: read before deciding whether a child does a piece of work. The first reason that applies decides; children that run side by side."
disable-model-invocation: true
---

# Delegation

Read before deciding whether a child does a piece of work, in any flow. Before briefing a child or waiting for one, read [child brief](../hive-rule-child-brief/SKILL.md); for a reviewer's brief, [reviewer brief](../hive-rule-reviewer-brief/SKILL.md). What a child may do is the delegation rule of the global guidance's «Scope and authorization».

For explicitly activated E2E work, read [E2E coordination](../flow-build/references/e2e-coordination.md) before assigning a local owner; it owns the named exception to ordinary child commits and checkout defaults. Ordinary tasks keep the rules below.

## When a child does the work

A child does the work for the first of these reasons that applies; when none applies, the main thread does the work.

1. **Independence: every review.** Every review or verification Hive launches, by a `hive-review-*` or `hive-verify-*` role, always runs in a child, never in the main thread, because the context that produced the work biases its review. The cost of that child does not decide. A code review by the host's native command, when the user chose it in the delivery round, runs as that command runs. The checks the main thread runs for each task are not a review.
2. **Cost: large work the session would carry.** What the main thread reads or writes rides along in each of its later calls, while a fresh child starts at tens of thousands of tokens and returns a short summary. Delegate:
   - in research, each aspect as [hive-mode step 4](../hive-mode/SKILL.md) sets;
   - in a build, a task that needs more than 5 large reads or edits more than 3 files, when its interface is settled and its writes do not overlap files another task in progress writes;
   - in any other flow, a check or collection that needs more than 5 large reads.

   A large read is defined in the global guidance's «Proportionality». Work below these limits stays in the main thread unless reason 3 applies, and so does a task whose writes overlap files another task in progress writes. Drawing the frames of [flow-figma-design](../flow-figma-design/SKILL.md) stays in the main thread whatever its size, as its step 5 sets; their review goes to a child under reason 1.
3. **Cross-check: the external facts of a plan.** When a plan's approach depends on two or more external systems, one read-only child per external system measures that system's facts, whatever their size, as [outside the repository](../hive-rule-external-facts/SKILL.md) sets, so the main thread checks each returned fact instead of producing it.

## When children run at the same time

- Children that run side by side are all launched before the main thread waits on any of them. When the host's launch call returns only once its child finishes, put all the launches in one response with no other tool call between them: launching one per response makes them run one after another.
- Read-only children run side by side: reviews, research aspects, the external facts of a plan, and the collections or sweeps a flow sends to parallel children, such as the independent collections of research's status report and its advisory sweep.
- Children that write, such as a backend task and a frontend task each delegated for its cost, run side by side only when all of these hold; otherwise they run one after another, each still in its child:
  - the contract they share is written as a typed artifact, such as types, a schema or an API definition, and does not change while they run;
  - their writes touch disjoint files;
  - the main thread does not need either result to continue;
  - an integration check across the seam, such as a contract test or a walk of the flow that crosses it, runs after both finish.
