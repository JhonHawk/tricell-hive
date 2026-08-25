---
order: 140
targets: [claude]
---

### Task Execution
- **Multi-step tasks: keep a task list.** Break the work down before starting — in the harness's task-list tool where one exists, in the reply otherwise — update state as you go, and mark an item complete only after verification.
- **Persistence is bounded to the CURRENT MILESTONE** — one rollback boundary (IaC state, DB schema, app release, external integration), never the whole ticket or epic; when duties collide: safety > scope > milestone breaker > verification > publication.
