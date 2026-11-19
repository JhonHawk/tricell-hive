# GitHub Issues

Read from [trackers](../SKILL.md) for a ticket whose key the `## Hive` assigns to GitHub Issues. Use the `gh` CLI; read `gh <command> --help` before the first use in a session.

| Step | Command |
| --- | --- |
| Current state | `gh issue view <number> -R <owner>/<repo> --json state,projectItems` |
| States | An issue is open or closed. A project board adds a status field: list it with `gh project field-list <project> --owner <owner> --format json`, each option of the single-select `Status` field being a state |
| Move | Done: `gh issue close <number> -R <owner>/<repo>`; reopen with `gh issue reopen`. On a board: `gh project item-edit --id <item> --project-id <id> --field-id <field> --single-select-option-id <option>` |
| Comment | `gh issue comment <number> -R <owner>/<repo> --body-file <file>` |
| Attach | `gh issue artifact create <number> -R <owner>/<repo> --attach <file>#<alt text>` for a capture, `--body-file <file>` for text such as test output; where the command or issue artifacts are not available, link the evidence from the pull request body, as [PR captures](../../git-workflow/references/pr-captures.md) sets, or list it in the comment for the user to attach |

Categories: a closed issue is done and an open one to do, unless the board's status option says in progress; record the board's option names when the project uses one.
