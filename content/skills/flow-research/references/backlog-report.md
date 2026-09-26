# Backlog report

Read before answering a question about the project's status or a request to list its backlog. The global guidance already requires the delivery state, a query across all non-terminal states, grouping five or more tickets by product module in the first answer, and no flat list, per-state dump, or grouping by tracker state, application, repository, or layer. A different cut of the tickets, such as those that fit one session, structures its own answer instead of this report.

- Take modules from tracker labels or projects, otherwise from titles; a tracker state, application, repository, or layer may appear only as a suffix.
- After the modules, group cross-cutting work such as infrastructure, CI, tests, tooling, and technical debt by type.
- Keep an initiative's umbrella and its phases together and in order, list deferred tickets apart, and name those blocking the next promotion.
- Give every ticket its ID.
- Open with one line per non-terminal state and its count, in-progress first; after the backlog, list relevant workspace state outside the tracker.
- Close by recommending which group or ticket to take first and why: production user-visible bugs, then security or data integrity, then promotion blockers, then debt and tooling; among equals, the smallest and most visible first.
