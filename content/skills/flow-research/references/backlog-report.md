# Backlog report

Read before answering a question about the project's status or a request to list its backlog. The global guidance already requires the delivery state, a query across all non-terminal states, grouping five or more tickets by product module in the first answer, and no flat list, per-state dump, or grouping by tracker state, application, repository, or layer. A different cut of the tickets, such as those that fit one session, structures its own answer instead of this report.

- Take modules from tracker labels or projects, otherwise from titles; a tracker state, application, repository, or layer may appear only as a suffix.
- After the modules, group cross-cutting work such as infrastructure, CI, tests, tooling, and technical debt by type.
- Keep an initiative's umbrella and its phases together and in order, list deferred tickets apart, and name those blocking the next promotion.
- Give every ticket its ID.
- Open with one line per non-terminal state and its count, in-progress first; after the backlog, list relevant workspace state outside the tracker.
- In a status report, not a plain backlog listing, give the rework of recent integrations right after the state counts, or within the delivery state when the project declares no tracker. Its window starts at the last promotion to the last environment that the repository's `Environments` setting declares after the base branch, as Git or the project's promotion mechanism records it, such as a merge into that environment's branch, a release, or a tag; with no such environment or no recorded promotion, the window is the last 30 days. Open with one line that names the window's start and gives three counts, then put each counted item on its own list line with its ID or link:
  - reverts made in the window of commits integrated into the base branch in it, found by the revert message, such as `This reverts commit` or `Revert "`;
  - tickets closed by work integrated in the window and reopened since;
  - bug tickets, meaning those the tracker labels or types as a bug, opened in the window whose body or links cite a ticket or pull request integrated in it.

  Give counts, never rates or targets, collected read-only through Git and the tracker's existing CLI or tool; delegate a collection that would produce long output as [flow-research](../SKILL.md) sets for tracker state. Without a tracker, give only the reverts and say that the tracker counts are unavailable; when nothing was integrated in the window, say so in the opening line instead of giving zeros.
- Close by recommending which group or ticket to take first and why: production user-visible bugs, then security or data integrity, then promotion blockers, then debt and tooling; among equals, the smallest and most visible first. A status report ends with these recommendations; ask the route question only when the user asked what to take next.
