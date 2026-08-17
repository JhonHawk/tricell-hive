---
alwaysApply: true
---

## Cross-Service Coordination

> **Always-on in Claude Code**, and reached via the `workspace-conventions` skill on Codex/opencode (`references/cross-service-workflow.md`). It is not conditional here because `system-designer` — the agent that enforces spec-first — has a `tools:` allowlist with no `Skill`, so it could never load this from a router.

> Spec-first for shared contracts: the contract is designed before code implements it, and nobody deviates from a spec without updating it first. Confirmation is by signal (below), not a per-contract stop. Contract specs live in `<project>-specs/contracts/`; pre-specs-repo fallback: `<project>/_support/spec/` (repo-local contracts: `<repo>/_support/spec/`).

### Trigger — a shared contract, not repo count
Spec check + contract design happen BEFORE implementation when the change touches a shared contract: a new API endpoint a frontend will consume, an added/modified request-response shape between services, a new event/message/webhook contract, or a modified existing cross-service contract. Touching multiple repos is NOT itself a trigger — mechanical parallel changes (env var rename, CI updates, the same bug fix in both repos) need no designer pass.

### Workflow
1. **Check for an existing spec** covering the contract. Exists → implement against it.
2. **No spec → write it first.** Dispatch scales with blast radius: a new contract or multi-consumer/multi-shape change → system-designer; a 1-field, 1-consumer change → inline spec edit by the implementing agent.
3. **Confirmation is by signal.** An approved plan that includes the contract IS the confirmation. Outside a plan gate, confirm up-front only for externally-visible or hard-to-change contracts — public/partner API, event schema, a shape that persists into a DB, a webhook a third party consumes, or invented domain semantics. An internal contract with both sides in the same change-group proceeds with the spec written, flagged in the close summary.
4. **Never deviate from the spec without updating it first.** A missing endpoint or unclear field is a gap: update the spec, then implement — never a workaround. When the fix is the implementer's to make, the update is a non-blocking note in the close report; contested cases follow `gap-resolution.md > Divergence Between Sources`.

Multi-agent chains follow `agent-routing.md`: declare the chain and execute, passing each agent's key outputs (spec paths, schemas) forward.

### Contract distribution
How the contract reaches consumers is part of the design, not the implementer's leftover.
- The handshake is a **published, versioned artifact** pinned by exact version — never a local pointer (`file:`, `link:`, `workspace:*`) in a consumer repo, and never "wait until the producer's feature merges".
- Publish a **prerelease** before implementing: both sides pin it and build in parallel. A wire change is a new prerelease, not an ad-hoc adjustment at integration time.
- A prerelease pin never reaches a production-deploying branch — gate it where the project gates promotion.
- New projects adopt this by default. Existing projects keep their mechanism; propose a change only when theirs causes drift or blocks parallel work.

### Spec consumption
- Implement against the spec's schemas exactly (paths, methods, status codes, request/response shapes) — never invent shapes. HTTP/REST → OpenAPI; event/message/webhook contracts → AsyncAPI or CloudEvents, not OpenAPI.
- Language-specific type examples in the spec are starting points; absent those, derive types from the schemas.

**Does NOT apply:** single-repo changes without external contracts, internal refactors that preserve the public API, bug fixes that don't change the interface.
