# Jira

Read from [trackers](../SKILL.md) for a ticket whose key the `## Hive` assigns to Jira. Use the Atlassian CLI `acli` 1.3, or the host's Atlassian MCP tool when it has one, and read `acli jira workitem <command> --help` before the first use in a session, since flags change between releases. Read private resources only through these authenticated tools.

| Step | Command |
| --- | --- |
| Current state | `acli jira workitem view <KEY> --fields status --json` |
| The project's states | `acli jira workitem search --jql 'project = <PREFIX>' --fields status --paginate --json`, collecting each distinct status with its `statusCategory`; it shows only states some ticket uses, so a state never used reaches the cache when a move or the user names it |
| The ticket's transitions | An Atlassian MCP tool that lists an issue's transitions; `acli` has none, so move by the exact name from the cache or the states |
| Move | `acli jira workitem transition --key <KEY> --status "<exact name>" --yes --json`; its result `No allowed transitions found for given status` means the name is not a transition of the ticket's current state |
| Comment | `acli jira workitem comment create --key <KEY> --body-file <file>` |
| Attach | `acli` only lists and deletes attachments: upload through an Atlassian MCP tool that attaches files, otherwise list the files in the comment for the user to attach |

Categories: `statusCategory.key` is `new` for to do, `indeterminate` for in progress and `done` for done. Status names follow the project's language, such as `En curso` or `Finalizada`; several states can share the `done` category, such as `Finalizada` and `Rechazada`, so record which one means integrated work.
