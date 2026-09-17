# Tracker access — shared mechanics

> Single home for the access ladder and batching semantics the tracker touchpoints used to
> restate. Consumers: the spec-writing playbook (task sync), flow-build (close),
> audit-playbook (ticket filing), memory-sync (propagation), promotion-playbook (close
> check), and the `state-fetcher` executor. Policy owner stays
> `rules-situational/memory-routing.md > Tracker sync`; ledger field semantics stay in
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
   truth. Linear names, so the `mcp` step costs one lookup instead of a search loop — Grok
   (observed): `linear__get_issue`, `linear__list_issues`, `linear__save_issue`,
   `linear__save_comment`; the operations are Linear's own, so other harnesses expose them
   under their MCP prefix (`mcp__<server>__save_issue`) — confirm the prefix once per session.
   A name not listed here is still discovered, never assumed.

## Writes

- Outward writes batch to ONE close confirmation at the finished plan — never per-ticket asks mid-run
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
  (`quality/reporting-integrity.md > Reporting state from ground truth`).
