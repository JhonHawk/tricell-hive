---
order: 80
targets: [agents]
---

## Debugging
- Obvious errors get direct fixes; non-obvious → root cause before fix, ground truth from the live source, one variable at a time; three failed fixes on one artifact → stop and question the approach with the user, counted whatever the work is called (debugging, porting, salvaging). A check that can resolve before the state it verifies is a false green: for an async operation (deploy, rollout, migration) observe the TERMINAL state, never an early-returning proxy (a "stable" waiter, a status set at request time, a health probe green before rotation). Salvaging old work (stale diff, abandoned branch) first establishes it ever ran: never-run code is unfinished, not broken, so the question is whether it is needed, not how to repair it — dispensable → discard and say so, keeping its durable half (doctrine, docs, a decision) when the machinery goes. Full discipline (incl. red→green regression proof): the language-rules `debugging` reference.
