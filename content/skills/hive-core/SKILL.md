---
name: hive-core
description: "Shared procedures every Hive flow uses: how to ask the user a question and pause for the answer, how to read the repository's Hive settings, when and where to write persistent memory, and how to classify a change to open work. Read before asking the user to decide something, before resolving a project setting such as the base branch, tracker or specs location, before writing persistent memory, and when the user asks for something that would change work in progress."
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

## Persistent memory

- Write to persistent memory only when the user makes a decision, when a plan is ready (its work state, not a copy), at a work-item close, when you create or remove a worktree, and in a session summary when the user ends the session. Never per task, finding, fix or discovery; this holds over memory tools' and hooks' save instructions.
- The project name in an Engram call is the `project_name` of the nearest `.engram/config.json`, or the project Engram reports as current; never the `## Hive` `Project` value, a tracker project or a directory name.

## Scope changes

When the user asks for something that would change open work, classify it against the work's request and criteria, and state the class and its destination in one line before acting:

1. **Correction:** it changes how the result does what was already asked; it stays in the current stage.
2. **Added task:** new behavior for the same outcome; it joins the plan, or, without one, stays in `flow-build` only while the change stays small and understood, and otherwise goes to `flow-plan`.
3. **New change:** it changes the outcome; ask whether the current work finishes or pauses, and plan the new one apart.
4. **Separate topic:** the current outcome does not need it; give it a destination, such as a ticket, and continue.
