# Project settings

Read before reading or asking for a `## Hive` key. It covers resolving a key too, of a repository's `## Hive` section.

Each repository's `AGENTS.md` holds a `## Hive` section; read it from the repository that holds the changed file, never infer settings from another repository, and ask the user when a key the task needs is missing. Required: `Project`, `Base branch`, `Tracker`, and `Specs`, the path to the `openspec` directory. Optional: `Environments`, the ordered environments from the base branch to production; `Other trackers`, the trackers besides the primary `Tracker`, separated by `;`, each as `<tool> · <location> · keys <PREFIX>[, <PREFIX>]`; `Delivery: direct-base`, the only value that authorizes effects, for a support repository; `Completion`, a later completion point as `<state> on <ref> · <state> on <ref>`, with one value per tracker separated by `;`, where a value starting with `<PREFIX>: ` applies to tickets with that key prefix and a value without a prefix to the primary tracker; and `Hive guidance: required`. Tickets across several trackers follow [ticket state](ticket-state.md).

Retired keys: `Review`, `UI verification`, and `Delivery` with `automatic` or `interactive`. They authorize nothing; propose removing them.

```markdown
## Hive

- Project: my-project
- Base branch: development
- Environments: development → qa → production
- Tracker: GitHub Issues · owner/repo
- Other trackers: Jira · example.atlassian.net · keys XYZ
- Completion: XYZ: QA on qa · Done on production
- Specs: _support/openspec
```
