---
alwaysApply: true
---

## Cross-Service Coordination

When a feature spans multiple services, repos, or requires frontend-backend agreement:

### Architect-First Workflow
1. **Check for existing spec** — before implementing a cross-cutting feature, look for a spec file in `<project>/_support/spec/` (workspace-level, cross-repo) or `<repo>/_support/spec/` (repo-level) that covers the required contracts.
2. **If no spec exists** — design the contract first using the system-designer agent. Do not begin implementation until the spec is written and the user has confirmed it.
3. **If a spec exists** — implementation agents must read and follow it. Do not deviate from the spec without updating it first.

### Mandatory Triggers
You MUST check for a spec and invoke system-designer BEFORE implementation when:
- The feature requires a new API endpoint that a frontend will consume
- The change adds or modifies request/response shapes between services
- A new event, message, or webhook contract is introduced
- An existing shared contract (API, event, schema) is modified across services

Touching multiple repos is NOT itself a trigger — the change must affect a shared contract. Renaming an env var across two services, updating CI workflows in both repos, or applying the same bug fix in parallel are mechanical changes that don't need a system-designer pass.

### Spec Consumption Rules
- **All implementation agents**: implement against the spec's OpenAPI schemas exactly (paths, methods, status codes, request/response shapes). Do not invent your own shapes.
- **Event/message/webhook contracts** use AsyncAPI or CloudEvents (CNCF), not OpenAPI (HTTP/REST-only); "implement against the spec exactly" applies equally to those for the async triggers above.
- **When language-specific type examples exist** in the spec, use them as starting points. When they don't, derive types from the OpenAPI schemas.
- **When the spec is insufficient** (missing endpoint, unclear field): flag the gap and update the spec before implementing a workaround.

### When This Applies
- New API endpoint that a frontend will consume
- New event/message/webhook contract between services
- Migration or refactor that changes existing shared contracts (API, event, schema)

### Multi-Agent Chaining
For cross-service features requiring multiple agents, follow the chaining protocol in the agent-routing rule: identify agents → present plan → execute sequentially with context passing.

### When This Does NOT Apply
- Single-repo changes that don't affect external contracts
- Internal refactors that preserve the public API
- Bug fixes that don't change the interface
