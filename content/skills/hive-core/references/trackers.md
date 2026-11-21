# Moving, commenting on and attaching to a ticket

Read after [ticket state](ticket-state.md) decides a move, and before a comment, attachment or worklog.

[Ticket state](ticket-state.md) decides when a ticket moves; this leaf sets how. Act through the first tool that can make the write, in this order: the tracker's CLI, the host's authenticated MCP tool for it, then its REST API with the user's global token through the reference's script, which never shows the token; read each tool's current help for the exact flags: [Jira](trackers/jira.md), [Linear](trackers/linear.md), [GitHub](trackers/github.md). For another tracker, apply the same steps with its own tool.

## Move

1. **Know the states.** Read the persistent-memory topic `tracker/states/<tracker>/<prefix>`, where `<tracker>` is `jira`, `linear` or `github` and `<prefix>` is the project's key prefix, such as `CAM`, followed by the ticket type when the workflow differs by type, such as `CAM-bug`. It lists each state's exact name and category (to do, in progress, done or other), the state the project uses for in progress and the one for integrated work, never a cancelled or rejected state, and the date it was verified. Without it, list the project's states from the tracker.
2. **Read the ticket's transitions live** before every move, when the tracker's tool lists them: what a ticket can move to depends on its current state and type. The cache chooses the target and never replaces this read; a tool that cannot list them, as its reference says, moves by the state's exact name from the cache or the states list.
3. **Move** to the state the project uses for the purpose ticket state names, by its exact name, never a name guessed in another language. When the move fails, take the state the cache or the states list gives for that purpose and retry once; when several states share its category and none is recorded for that purpose, or the retry fails, report a blocker that lists the states found, never pick one.
4. **Keep the cache.** When the topic is missing, a state or a category changed, or a move failed, write the topic with the states, their categories and the two purposes only, never ticket content.
5. **Read the state back** and report it.

## Comment and attach

- Write a comment the work owes, such as its pending work or its evidence, through the tracker's tool, in the ticket's language.
- Evidence of a work item, such as before and after captures, the output of its tests and links to its pull request, reaches the ticket before any step that removes it, as [flow-close](../../flow-close/SKILL.md) sets: attached through the tracker's tool where it uploads files, otherwise linked from the pull request, or listed in the comment for the user to attach. Never commit a capture to link it. A public repository or tracker gets an image only as [PR captures](../../git-workflow/references/pr-captures.md) authorizes it.
- Never attach or paste a secret value, an account identifier or a capture that shows one: crop it, or leave it out and say so in the comment.
- Read the comment or the attachment list back. When a later deletion depends on the evidence, as [support material](../../flow-close/references/support-material.md) sets, read the attachment list: evidence the comment only lists for the user to attach is not a copy.

## Time

When [ticket state](ticket-state.md) says to write or update a ticket's time, measure and write it as [time](trackers/time.md) sets.
