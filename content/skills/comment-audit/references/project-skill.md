# Project skill

Read at [comment audit](../SKILL.md) step 9. It sets where the project's own comment skill lives, what it holds and how every host finds it, so the next batch runs from the repository alone.

## Name and place

- Name it `delete-comment` unless the user names it otherwise (matching `^[a-z0-9][a-z0-9-]*$`) and write it to `.agents/skills/<name>/SKILL.md` at the repository root, which Codex, Cursor, OpenCode, Pi and Antigravity read.
- Link the folder per host instead of copying it, since copies drift: `.claude/skills/<name>` → `../../.agents/skills/<name>` for Claude Code, and `.grok/skills/<name>` the same way when the repository has a `.grok/` folder.
- Leave `assets/hive-managed.json` out: that marker tells harness-audit's registry check a skill belongs to Hive, and this one belongs to the project.

## What it holds

Frontmatter with `name` and a `description` that says when to use it, such as "Run the next comment cleanup batch, or fix a failing comment lint. Use before deleting comments or when `lint:comments` fails." Then, in the project's own terms:

1. The inventory command, such as `node scripts/ci/check-comments.mjs --inventory`, and the paths `EXCLUDED` leaves out.
2. The categories and the action for each, as [comment audit](../SKILL.md) step 4 sets them, with the project's examples from the ledger.
3. Each mechanical pattern a batch used, with the exact expression, so the next batch repeats it.
4. Where the ledger of the last audit is, and the comments it left unmarked.
5. The decision record's path, the tracker and parent ticket for excuses, and the workaround marker's form.
6. The checks a batch runs: `--same-code` against the commit before the batch, the comment lint, the lint, the type check and the tests of the touched files.
7. That a bulk edit made by a script does not show in a host's per-turn diff view, and to report `git diff --stat` instead.

## Register it

- In the repository's `AGENTS.md`, add a line to its `## Skills` section, creating the section with the opening line [instruction files](../../harness-audit/references/instruction-files.md) HA-IF-19 sets when it is missing:

  ```markdown
  - Before auditing or deleting code comments, or when the comment lint fails: `.agents/skills/<name>/SKILL.md`
  ```

- When the repository is a child of a declared workspace, add the same line to the workspace's `## Skills`, with the child's path as a prefix on the trigger and on the skill's path, such as `app/.agents/skills/<name>/SKILL.md`.
- Check with `python3 -I <harness-audit-dir>/scripts/skill_registry.py check <repository>` ([skill registry](../../harness-audit/scripts/skill_registry.py)), and on the workspace when there is one. Each must exit 0; fix each finding it prints.
