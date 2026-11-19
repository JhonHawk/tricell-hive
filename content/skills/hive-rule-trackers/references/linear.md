# Linear

Read from [trackers](../SKILL.md) for a ticket whose key the `## Hive` assigns to Linear. Use the Linear MCP tools when the host has them, otherwise the Linear CLI or its GraphQL API; read the tool's current description before the first use in a session.

| Step | With the MCP tools |
| --- | --- |
| Current state | `get_issue` with the key |
| The team's states | `list_issue_statuses` for the issue's team, each with its `type` |
| Move | `save_issue` with the key and the state's exact name; a state belongs to its team, so take it from that team's list |
| Comment | `save_comment` on the issue |
| Attach | `prepare_attachment_upload`, then `create_attachment_from_upload`; or embed an uploaded file in the comment as `prepare_file_upload` returns it |

Categories: a state's `type` is `backlog` or `unstarted` for to do, `started` for in progress, `completed` for done, and `canceled` for other. Linear lets an issue move to any state of its team, so the transition read of [trackers](../SKILL.md) step 2 is the team's list of states.
