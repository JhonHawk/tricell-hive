---
order: 80
targets: [agents]
---

## Debugging
- SSH blocker: load language-rules `debugging` before escalation.
- Obvious errors get direct fixes; non-obvious → root cause before fix, ground truth from the live source, one variable at a time; three failed fixes on one artifact → stop and question the approach with the user, counted whatever the work is called (debugging, porting, salvaging). The SECOND failed remote apply/deployment/migration attempt within one milestone ends remote execution for the session — remote infrastructure is never the compiler: checkpoint and re-plan from offline evidence (action×resource×context, versions, exact plan) before any further remote attempt; a NEW failure class does not reset the counter, one verified transient transport retry is exempt. A check that can resolve before the state it verifies is a false green: for an async operation (deploy, rollout, migration) observe the TERMINAL state, never an early-returning proxy (a "stable" waiter, a status set at request time, a health probe green before rotation). Salvaging old work (stale diff, abandoned branch) first establishes it ever ran: never-run code is unfinished, not broken, so the question is whether it is needed, not how to repair it — dispensable → discard and say so, keeping its durable half (doctrine, docs, a decision) when the machinery goes. Full discipline (incl. red→green regression proof): the language-rules `debugging` reference.
