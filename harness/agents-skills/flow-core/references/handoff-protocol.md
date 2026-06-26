# Handoff protocol — dispatching agents from flow skills

Agents start cold: they see none of the conversation, none of the plan, none of the other
agents' work. Everything they need travels in the prompt. A handoff that skips one of the
four elements below produces plausible-looking but misaligned work — and the cost of
re-running an agent is always higher than the cost of writing one good prompt.

## The four elements — every dispatch includes all of them

1. **Intent** — the why and the consumer. Not "implement the messages endpoint" but
   "implement the messages endpoint; its response shape feeds PROJ-414's UI bubble, and the
   contract is already frozen in the OpenAPI spec — deviating from it breaks the frontend
   task running in parallel". Agents with intent make aligned micro-decisions; agents
   without it guess.
2. **Bounded context** — concrete paths and excerpts, never "read everything". Give the
   spec section (quote it or give the exact path + heading), the contract file, the 2-3
   existing files whose conventions to mirror. When the epic has a reviewed mock, the
   **mock repo is part of the bounded context** for an implementation dispatch — name it
   (the ledger's `## Current handoff` carries the path), so the agent builds against the
   prototype the UX was validated on instead of re-inventing the screen. If you cannot name
   the files, you are not ready to dispatch — explore first.
3. **Applicable conventions** — only the ones the agent cannot infer: the project naming
   table (`<project>-specs/conventions/naming.md`), the branch to work on. Do not paste
   global rules — agents inherit them.
4. **Expected output shape** — what comes back and in what form: "return: files
   created/modified with paths, migration ID, what you verified (command + result), open
   issues". The agent's final message is the only thing the orchestrator sees; an
   unspecified return forces a follow-up round-trip. When the dispatch (or a plan step)
   includes a command to run, pair it with its expected result — `Run: pnpm test
   messages.spec — Expected: 12 passing, 0 failing` — a step whose expected output you
   cannot state is not a verification step yet, and silent drift ("it ran" but not as
   expected) becomes detectable.

## The return contract

Tell every agent its final message must report, raw and compact:
- What changed (paths) and what was created
- Verification: the exact command run and its actual result — a claim of "build passes"
  without the command is not verification; the orchestrator re-checks via Bash
- Anything it could NOT do, with the reason — silent scope-trimming is the worst outcome
- Open questions that affect downstream tasks

**Every return opens with one of four statuses**, and the orchestrator's response to
each is fixed — never a silent retry of the same agent unchanged:

| Status | Meaning | Orchestrator response |
|---|---|---|
| `DONE` | Complete, verified as dispatched | Verify against the diff, then proceed |
| `DONE_WITH_CONCERNS` | Complete, but the agent flags risks or doubts | Read the concerns BEFORE consuming the output; real ones route to review or the user |
| `NEEDS_CONTEXT` | Missing information to proceed correctly | The gap was in the handoff — re-dispatch the same task with the missing context added |
| `BLOCKED` | Cannot proceed: failed verification, contradiction, missing resource | Diagnose before re-dispatching: task too large → split it; reasoning beyond the agent → escalate model or specialist; plan defect → surface to the user |

Instruct agents that escalating is always legitimate — "this is too hard for me" is an
acceptable return; bad work costs more than no work.

## Reviewer and verifier dispatches

When the dispatched agent's job is to review or verify ANOTHER agent's work, two rules
stack on top of the four elements:

- **Prime the reviewer against the report.** The implementer's report travels as claims
  to check, never as context to trust — state in the dispatch that it may be incomplete,
  optimistic, or wrong, and that no claim is accepted without reading the code it points
  at. The reviewer's primary input is the diff (or SHA range), not the narrative.
- **The orchestrator holds itself to the same rule.** Subagent work is verified against
  the VCS diff and by re-running reported commands — "the agent said it's done" is never
  evidence; the diff is.

## Example dispatch (flow-build, backend task)

```
Implement PROJ-413 (send message endpoint) in chat-hub-backend, branch feat/proj-413-send.

Intent: this endpoint feeds the customer portal's message composer (PROJ-414, runs next).
The response shape is frozen in the OpenAPI contract — the frontend will be built against
it without seeing your code.

Context:
- Spec: chat-hub-specs/epics/E07-mensajeria/PRODUCT.md § "Send flow" (Gherkin ACs there)
- Contract: chat-hub-specs/contracts/messages.openapi.yml — implement exactly
- Mirror conventions of: src/modules/templates/ (module layout, DTO validation, error shape)
- Naming: chat-hub-specs/conventions/naming.md applies to any queue/bucket you touch

Return: modified paths, migration IDs if any, test files added, the verify command you ran
with its output, and anything you couldn't complete with the reason.
```

## Anti-patterns

- **"Read the repo and figure it out"** — burns the agent's context on discovery you
  already did, and it may discover a different convention than the one you wanted.
- **Dispatching dependent tasks in parallel** — if B consumes A's output, B waits;
  parallel-dispatching them means B invents A's interface.
- **Re-dispatching a failed task to a different agent silently** — surface the failure to
  the user; a second agent inherits the same blocker plus less context.
- **Omitting the consumer** — "write tests for X" yields generic tests; "write tests for X;
  the deploy gate runs them against QA after every deploy" yields the right tests.

## Phase handoff — the research → write → build → verify chain

Everything above governs **agent dispatch** (skill → agent). The dev trinity
(`flow-plan research` → `flow-plan write` → `flow-build` → its `verify` gate) also hands off
**phase → phase**, and its carrier is an **artifact**, not the ledger's `## Current handoff`.
This matters because a phase may run on a *different harness* than the one before it: the
carrier has to be a durable file any harness reads cold, not session state.

| Transition | Carrier (what the next phase reads) | What the consumer does |
|---|---|---|
| research → write | `<slug>-findings.md` | `write` turns the chosen approach + gaps into the executable plan |
| write → build | `<slug>-plan.md` (`plan-format.md`) | `build` reconciles `Status` + git and executes the pending tasks |
| build → verify | `Status: built` + git (the landed task commits) | the in-vivo/review gate walks the `in-vivo: yes` tasks |
| verify → close | `Status: verified` + the versioned in-vivo report | the ledger CLOSE records the outcome and seals back-references |

- **The plan is the cross-harness carrier.** Unlike the ledger handoff (consumed by the same
  orchestrator at the next phase), `plan.md` is consumed by whoever runs `flow-build` — which
  may be a cheaper model on another harness. So it obeys `plan-format.md`: self-contained,
  harness-neutral, state-bearing.
- **Status advances forward only, and is verified against git.** No phase trusts the header
  alone — `built` is believed because the task commits are in the log, `verified` because the
  report exists. A header that contradicts git is stale; git wins (`memory-routing.md`).
- **The ledger still closes each phase.** The artifact chain carries the *work*; the ledger's
  `## Current handoff` still carries the *coordination* note (paths produced, next phase
  suggested) per the flow contract. The two are complementary — artifact for the executor,
  ledger for the orchestrator.
