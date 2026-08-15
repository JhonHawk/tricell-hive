# Tracker access — shared mechanics

> Single home for the access ladder and batching semantics the tracker touchpoints used to
> restate. Consumers: flow-specs (task sync), flow-build (close), flow-audit /
> audit-playbook (ticket filing), memory-sync (propagation), promotion-playbook (close
> check), and the `state-fetcher` executor. Policy owner stays
> `rules/workflow/memory-routing.md > Tracker sync`; ledger field semantics stay in
> `ledger-template.md` — this file owns only the shared HOW.

## Resolution

1. **The declaration authorizes — never detection.** Flow projects: the ledger's three
   fields (`Tracker`, `Tracker access`, `Deferred marker`) are THE home. Repos without a
   ledger: an AGENTS.md/CLAUDE.md declaration block. Undeclared → the tracker is not
   consulted; propose the declaration instead — its write is confirm-gated (it grants
   standing authorization to an external system).
2. **Access ladder, in order:** `mcp` — load the tracker's tools via ToolSearch (names are
   install-scoped: discover, never assume) · `cli` — `acli` for Jira (often the most
   effective route), `gh issue` for GitHub · `api` — REST/GraphQL with the named env token ·
   `manual` — list the updates for the user. `Tracker: none` → `tasks.md` is the source of
   truth.

## Writes

- Outward writes batch to ONE session-close confirmation — never per-ticket asks mid-run
  (`memory-routing.md > Tracker sync`); unattended runs queue identically.
- Never delete, archive, transfer, or bulk-close; terminal transitions (Done/Closed/
  Cancelled) are propose-only unless explicitly handed as approved.
- An approved batch executes with its payload read from disk (tasks.md rows, audit
  findings): one issue per task, ACs in the description, IDs aligned both ways, per-item
  result reported. A batch or board read big enough to inflate the caller dispatches
  `state-fetcher` (`agent-routing.md`); a single quick state call stays inline.

## Reading

- A read result carries per-ticket `ID | state | title` (+ deferred marker) — IDs always
  survive the synthesis: branch names (`<type>/<TICKET>-<slug>`) are cut from them.
- Live state outranks the tracker: a ticket-vs-git conflict is reported for ticket
  correction, never resolved by editing the ticket to match a memory
  (`debugging.md > Reporting state from ground truth`).
