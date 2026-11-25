# Publish captures and walk an interface

Read when an interface is captured, or a walkthrough is asked for. It applies when a build or verify captures an interface, when a flow or role publishes captures, and when the user asks for a walkthrough or demo of a built interface, which follows this leaf alone. The close of a work with captures follows [close the viewer](viewer-close.md).

A build or verify flow that was explicitly invoked (its command typed, its route accepted, or research, plan or build continuing into the next flow at the user's request or as a handoff sets), and a walkthrough or demo the user asks for, gets a viewer session at its first UI capture; research and plan publish nothing, and a flow without captures creates no viewer work. The primary producer follows [viewer publication](viewer.md) at that capture; a child never publishes.

For a walkthrough, serve the candidate locally, as [Show it in the host's pane](#show-it-in-the-hosts-pane) sets when the session has that pane, or install and launch a native app on the assigned device or emulator as [Device tool](ui-verification/device-tool.md) sets, with seeded data of realistic volume and text length (stub data only for a build's first view, as [UI decisions](ui-plan.md) defines) and an account that reaches the changed screens, chosen after reading [test accounts](../local/test-accounts.md). Take the capture scope from [UI verification](ui-verification.md) and reuse existing evidence. Publish the current captures, then give the user the target, steps, expected result, observations and evidence paths. The calling flow owns its existing acceptance question and handles corrections; this reference adds only the open-viewer option that [open the viewer](viewer-open.md) sets. On a correction, capture again only the affected screens under their same titles: the viewer numbers the new iteration and keeps the earlier one.

## Show it in the host's pane

This section applies at a stop where the user reviews a served app, such as a build's first view, a walkthrough or a validation request with its Try it block, when this session's tools include a browser pane the user sees inside the host app, such as a desktop app's built-in browser. A terminal session has no such pane, and the stop gives its URLs as links.

- Serve the candidate through the host's own preview mechanism when it has one, such as a launch configuration its preview tool starts, instead of a shell command, keeping the build or preview command the calling flow sets. The Try it block gives the URL and port the served instance reports, not a port `HIVE.local.md` records, since the host may pick a free port each run; a port the host's preview picks is not a fact to record there.
- Open each URL the stop names in that pane, visible to the user when the host keeps it in the background by default. When the host reaches the pane only through its own browser subagent, open it through that subagent: it is the host's access to the pane, not a Hive child, so it does not count toward [delegation](../delegation/delegation.md).
- When opening fails, give the URLs as links and say why.

Automated checks, captures and children keep the tool [Browser tool](ui-verification.md#browser-tool) sets. [Close cleanup](../close/close-cleanup.md) sets what happens to a server the host's preview started.
