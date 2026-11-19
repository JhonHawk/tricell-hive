# Jira

Read from [trackers](../SKILL.md) for a ticket whose key the `## Hive` assigns to Jira. The commands are those of the Atlassian CLI `acli` 1.3; read `acli jira workitem <command> --help` before the first use in a session, since flags change between releases.

| Step | Command |
| --- | --- |
| Current state | `acli jira workitem view <KEY> --fields status --json` |
| The project's states | `acli jira workitem search --jql 'project = <PREFIX>' --fields status --json`, collecting each distinct status with its `statusCategory` |
| The ticket's transitions | `acli` has no command for them: `GET /rest/api/3/issue/<KEY>/transitions` of the Jira REST API, each with `to.name` and `to.statusCategory.key` |
| Move | `acli jira workitem transition --key <KEY> --status "<exact name>" --yes --json`; its result `No allowed transitions found for given status` means the name is not a transition of the ticket's current state |
| Comment | `acli jira workitem comment create --key <KEY> --body-file <file>` |
| Attach | `acli` only lists and deletes attachments: upload with `POST /rest/api/3/issue/<KEY>/attachments` and the header `X-Atlassian-Token: no-check` |

Categories: `statusCategory.key` is `new` for to do, `indeterminate` for in progress and `done` for done. Status names follow the project's language, such as `En curso` or `Finalizada`; take them from the transitions, never from the English defaults.

For a REST call, use the site and the API token the project's secret store holds, as [test accounts](../../hive-rule-test-accounts/SKILL.md) sets, passed through a file or a variable scoped to the command, never in its arguments or output.
