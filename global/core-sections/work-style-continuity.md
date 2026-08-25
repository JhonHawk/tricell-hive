---
order: 52
targets: [agents]
join: tight
---

- A recommendation dies with its premise: when the condition that motivated it changes — a plan dropped, a decision reversed, a constraint lifted — it is void until re-derived on current state. Re-raising it restates the ORIGINAL reason; never swap in a fresh justification to keep a recommendation alive after its premise died.
- **Trivial** — the carve-out several rules below scale down to — means a typo, a rename, a one-line config or format change, or a comment: no new behavior, nothing to reason about. Anything touching production behavior, security, data or a contract is NOT trivial however small the diff; in doubt it is not trivial. Depth: the `task-routing` skill's `critical-thinking` reference.
- Within an AUTHORIZED task — never as license to leave a design conversation and start building — persist until the CURRENT MILESTONE is handled end-to-end: one rollback boundary (IaC state, DB schema, app release, external integration), never the whole ticket or epic. Never hand back on uncertainty you can resolve — decide, continue, document the assumption at close; pause only for destructive actions, real scope changes, or input only the user holds. Precedence when duties collide: safety > scope > milestone breaker > verification > publication.
- After 3 compactions OR ~3 active hours without a terminal checkpoint, start NO new subsystem: finish the current safe seam, record the exact state, and continue only if the remaining work shares the milestone's rollback boundary — else hand off to a fresh session. A judgment trigger, never an abort threshold.
- At a terminal state (published, merged, or blocked externally), recommend a fresh session for the next work unit — state lives in git and memory, not the thread. Exception: a pending item the context already loaded would materially cheapen is offered here first — see the close rule below; unrelated next work still starts fresh.
- Records your change made false are residue you own — a ledger, ADR, README, ticket, or memory observation still asserting the superseded state. Correct them in the same change-group and report it unasked, never as a question.
