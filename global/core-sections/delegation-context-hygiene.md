---
order: 190
targets: [claude]
---

### Delegation & Context Hygiene

Keep the main thread focused: delegate executable work, reason in the main thread. Governing question: does this inflate my context without need? Yes → delegate; no → inline. **Invoke the `task-routing` skill at FIRST EDIT-INTENT on a software project — or at the first git verb of a session that never reaches one — and before the first delegation or a plan's tasks.** A read-only investigation, a short question, and work outside a software project never reach either trigger; the trivial carve-out is out too. — it carries the routing table, the delegation gates, inline-vs-delegate, fresh-context verification, the gap analysis, and the git-mechanics reference. Neither rule is always-on: their trigger is an intent, not a file, so nothing loads them for you and not invoking the skill is the same as not having them.

- **Parallelize only independent work, and name how the agent asks back.** Concurrent subagents have no dependency between them — an output another consumes, a co-written file, a pending decision; a split that would degrade a result serializes; every dispatch prompt tells the agent to STOP at a real fork it does not settle and ask you — by message, or by ending its turn with the question — never to pick an option and flag it; you answer and resume the same agent (`agent-routing.md > Delegation Gates`). A decision that is the user's, or irreversible (a migration, a published contract), is asked BEFORE dispatching — never settled inside the dispatch prompt as an assumption to confirm later.
- **Delegate with the intent, not only the task.** Subagent prompts state the why — the larger goal, who or what consumes the output, and what it enables — so the agent connects the task to relevant context instead of inferring it.
