# Project skill

Read at [comment audit](../SKILL.md) step 9. It sets where the project's own comment skill lives, what it holds, how every host finds it and the comment rule written to the project's `AGENTS.md`, so the next batch runs from the repository alone.

## Name and place

- Name it `delete-comment` unless the user names it otherwise (matching `^[a-z0-9][a-z0-9-]*$`) and write it to `.agents/skills/<name>/SKILL.md` at the repository root, which Codex, Cursor, OpenCode, Pi and Antigravity read.
- Link the folder per host instead of copying it, since copies drift: `.claude/skills/<name>` → `../../.agents/skills/<name>` for Claude Code, and `.grok/skills/<name>` the same way when the repository has a `.grok/` folder.
- Leave `assets/hive-managed.json` out: that marker tells audit-harness's registry check a skill belongs to Hive, and this one belongs to the project.

## What it holds

Frontmatter with `name` and a `description` that says when to use it, such as "Run the next comment cleanup batch, or fix a failing comment lint. Use before deleting comments or when `lint:comments` fails." Then, in the project's own terms:

1. The inventory command, such as `node scripts/ci/check-comments.mjs --inventory`, and the paths `EXCLUDED` leaves out.
2. The categories and the action for each, as [comment audit](../SKILL.md) step 4 sets them, with the project's examples from the ledger.
3. Each mechanical pattern a batch used, with the exact expression, so the next batch repeats it.
4. Where the ledger of the last audit is, and the comments it left unmarked.
5. The decision record's path, the tracker and parent ticket for excuses, and a pointer to the `## Comments` section of `AGENTS.md` for the workaround marker's form.
6. The checks a batch runs: the comment lint, the lint, the type check and the tests of the touched files, then `--same-code --base <the batch's record commit>` last, since those commands can regenerate files; stage the removal commit by file name.
7. That a bulk edit made by a script does not show in a host's per-turn diff view, and to report `git diff --stat` instead.

## Register it

- In the same repository's `AGENTS.md`, write a `## Comments` section above `## Skills` (which the next bullet creates when it is missing), outside `## Hive` and any installer-managed block, from this template, with the project's comment lint command, the paths it checks (its `SOURCE_PATTERNS` minus `EXCLUDED`), decision record path and tracker in place of each `<…>`; when the file already has a `## Comments` section, update it instead of adding another. A declared workspace does not get it, since the lint runs in each repository. When that level's `CLAUDE.md` does not import `@AGENTS.md`, report it as audit-harness's [hierarchy](../../audit-harness/references/hierarchy.md) HA-HI-03 sets.

  ```markdown
  ## Comments

  Do not add comments: `<comment lint command>` rejects a new or returned comment in `<paths the lint checks>`. Put each fact where it belongs instead, and say what the code does through names, types, tests and structure:

  - A decision or an outside constraint, such as why a surprising choice was made or a vendor or protocol quirk, goes to an entry in `<decision record>` with the paths it concerns.
  - A shortcut or pending work, such as a TODO, a silent catch or a fallback, goes to a ticket in `<tracker>`, with the workaround marker over the code.
  - Something the next editor must not do becomes a check that enforces it (a type, an assertion, a test or a lint rule), proposed before it is added; until one exists, an entry in `<decision record>`.

  Only these comments stay:

  - Directives a tool reads, such as a lint suppression with its reason or a bundler's magic comment.
  - License headers.
  - The workaround marker, on one line: the line-comment token, then `workaround: ` and the ticket's key or URL, or the record entry as `<path>#<anchor>` while no ticket exists; `{/* workaround: <ticket> */}` inside JSX.

  Existing comments stay until a cleanup batch removes them.
  ```

- In the repository's `AGENTS.md`, add a line to its `## Skills` section, creating the section with the opening line [instruction files](../../audit-harness/references/instruction-files.md) HA-IF-19 sets when it is missing:

  ```markdown
  - Before auditing or deleting code comments, or when the comment lint fails: `.agents/skills/<name>/SKILL.md`
  ```

- When the repository is a child of a declared workspace, add the same line to the workspace's `## Skills`, with the child's path as a prefix on the trigger and on the skill's path, such as `app/.agents/skills/<name>/SKILL.md`.
- Check with `python3 -I <audit-harness-dir>/scripts/skill_registry.py check <repository>` ([skill registry](../../audit-harness/scripts/skill_registry.py)), and on the workspace when there is one. Each must exit 0; fix each finding it prints.
