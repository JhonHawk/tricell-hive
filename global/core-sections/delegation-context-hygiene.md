---
order: 190
targets: [claude]
---

### Delegation & Context Hygiene

Keep the main thread focused: delegate executable work, reason in the main thread. Governing question: does this inflate my context without need? Yes → delegate; no → inline. **Invoke `task-routing` at the first of the acts its row lists above** — it settles the requested deliverable, the authorized effects, the completion evidence and the stopping condition before work starts, and carries the routing table, delegation gates, gap analysis and git mechanics. Nothing loads it for you: not invoking it is identical to not having those rules. A short self-contained question and trivial mechanical work skip it; the project-edit and git gates still apply.

- **Parallelize only independent work, and name how the agent asks back.** Concurrent subagents have no dependency between them — an output another consumes, a co-written file, a pending decision; a split that would degrade a result serializes; every dispatch prompt tells the agent to STOP at a real fork it does not settle and ask you — by message, or by ending its turn with the question — never to pick an option and flag it; you answer and resume the same agent (`agent-routing.md > Delegation Gates`). A decision that is the user's, or irreversible (a migration, a published contract), is asked BEFORE dispatching — never settled inside the dispatch prompt as an assumption to confirm later.
- **Delegate with the intent, not only the task.** Subagent prompts state the why — the larger goal, who or what consumes the output, and what it enables — so the agent connects the task to relevant context instead of inferring it.
