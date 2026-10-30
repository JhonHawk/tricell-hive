---
name: hive-core
description: "Shared procedures every Hive flow uses: how to ask the user a question and pause for the answer, and how to read the repository's Hive settings. Read before asking the user to decide something, and before resolving a project setting such as the base branch, tracker or specs location."
---

# Shared procedures

## Ask the user

- Ask only for a choice that is the user's and changes what happens next; reuse decisions already given.
- Use the host's native question tool when it has one; otherwise list the options in the message, each with a short label.
- Put the recommended option first, with its reason. Ask related choices in one round.
- End the message with one line that says what you are waiting for, and wait for the answer.

## Declare settings

Each repository's `AGENTS.md` holds a `## Hive` section; read it from the repository that holds the changed file, and ask the user when a key the task needs is missing. Required: `Project`, `Base branch`, `Tracker`, and `Specs`, the path to the `openspec` directory. Optional: `Environments`, the ordered environments from the base branch to production; `Delivery: direct-base`, the only value that authorizes effects, for a support repository; `Completion`, a later completion point as `<state> on <ref> · <state> on <ref>`; and `Hive guidance: required`.

Retired keys: `Review`, `UI verification`, and `Delivery` with `automatic` or `interactive`. They authorize nothing; propose removing them.

```markdown
## Hive

- Project: my-project
- Base branch: development
- Environments: development → qa → production
- Tracker: GitHub Issues · owner/repo
- Specs: _support/openspec
```
