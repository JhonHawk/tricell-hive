# Jira

Read from [trackers](../SKILL.md) for a ticket whose key the `## Hive` assigns to Jira. Use the Atlassian CLI `acli` 1.3, then the host's Atlassian MCP tool when it is authenticated and has the capability, then [jira_rest.py](../scripts/jira_rest.py) (`python3 -I`), which reads the user's token from `~/.config/hive/secrets/jira-api-token` (under `$XDG_CONFIG_HOME` when it is an absolute path) and the account and site from `acli jira auth status`. Read `acli jira workitem <command> --help` before the first use in a session, since flags change between releases. When the script exits 3, ask the user to create the token file as its message says; never create it, read it or pass its path to another agent.

| Step | Command |
| --- | --- |
| Current state | `acli jira workitem view <KEY> --fields status --json` |
| The project's states | `acli jira workitem search --jql 'project = <PREFIX>' --fields status --paginate --json`, collecting each distinct status with its `statusCategory`; it shows only states some ticket uses, so a state never used reaches the cache when a move or the user names it |
| The ticket's transitions | An Atlassian MCP tool that lists them, otherwise `jira_rest.py transitions <KEY>`; `acli` has none |
| Move | `acli jira workitem transition --key <KEY> --status "<exact name>" --yes --json`; its result `No allowed transitions found for given status` means the name is not a transition of the ticket's current state |
| Comment | `acli jira workitem comment create --key <KEY> --body-file <file>` |
| Attach | `acli` only lists and deletes attachments: an Atlassian MCP tool that attaches files, otherwise `jira_rest.py attach <KEY> <file>…` |
| Worklog | `acli` has none: an Atlassian MCP tool that writes worklogs, otherwise `jira_rest.py worklog-list <KEY>`, then `worklog-add <KEY> --seconds <n>` or `worklog-update <KEY> <id> --seconds <n>`, as [time](time.md) sets |

Categories: `statusCategory.key` is `new` for to do, `indeterminate` for in progress and `done` for done. Status names follow the project's language, such as `En curso` or `Finalizada`; several states can share the `done` category, such as `Finalizada` and `Rechazada`, so record which one means integrated work.
