---
title: Viewer
description: Follow published work, inspect evidence, and manage the local viewer.
---

The local viewer presents a timeline of work published by an agent. A work can have several native producer sessions contributing to it; the viewer groups their published stages, tasks and evidence. A producer's presence in the timeline does not prove that its host is still running. The viewer is separate from the host's native history and from Engram.

The [flow-walkthrough](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md) skill is what publishes stages, documents and captures to the viewer; the viewer shows what has been published to it. Each work gets its own entry: a work tied to a ticket keeps one entry across the flows of that ticket in the same agent session, and a work without a ticket uses its own stable work key. The agent does not open the viewer by itself: at each point with something to review (a plan or a screen to validate) it offers to open it within the question or reply that point already asks, never as a question of its own, and stops offering once you accept. When a build starts, the agent publishes the plan's tasks so **Tasks** follows their progress.

## Start and stop

Start the viewer explicitly with `hive view start`, then use `hive view open` to open it. Use `hive view status` to check the service and `hive view stop` to stop it. Retention is manual: stopping the service does not mean the stored timeline has been deleted.

Protocol 3 requires a fresh state directory when coming from an earlier storage format. For the timeline revision 2 update within protocol 3, update the CLI, viewer assets and server together, then reload the viewer. Keep a private backup of viewer state before the upgrade and retain it as evidence; the update preserves stored works and the publisher journal. A revision mismatch blocks timeline writes and asks for an update or reload. The [viewer publication procedure](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/references/viewer-publication.md) describes the producer's publication contract.

## Publish Markdown on demand

The agent keeps the plan files in their change folder and calls
`hive view work publish` when a revision is ready to review. The binary scans
that directory, uploads the original Markdown and verifies the stored copy.
There is no watcher: editing a file alone does not publish it. An unchanged
publication returns `unchanged` without adding remote history.

Each work has a stable key and returned ID. Revising ticket 10 reuses its ID;
starting ticket 11 in another folder creates a separate work, even in the same
native session. The explicit directory binding prevents a changed working
directory from selecting the wrong work. `hive view work list --project ROOT`
finds works, and `hive view work show --work-id ID` shows their observed state.

The plan's default tabs are Proposal, optional Design, Tasks and existing spec
deltas. An explicit selection can replace those defaults. Publications have at
most eight tabs, each at most 256 KiB of UTF-8 Markdown; oversized selections
fail instead of being truncated. The renderer stays the same: it does not load
local images, execute HTML or navigate relative file links as a documentation
site. Original sources remain authoritative.

The agent separately registers structured tasks and sends short task/stage
updates. Editing a Markdown checkbox never changes live task progress. Snapshot
status is not approval, and finishing a task does not close the work.

If a response is lost, `work resume` recovers the exact pending request using
its saved bytes; later source edits do not alter that retry. A confirmed conflict
requires explicit reconciliation before a fresh publication. The local journal
is retained outside the viewer's storage, under `publisher/v1/`; deleting it is
not a recovery procedure. No command starts or restarts the viewer.

Run `hive view work contract` for complete commands and required native-source
fields. Advanced membership, evidence and CI operations retain the existing
`hive view contract` interface.

## Interface language

System labels use English by default and Spanish when the browser's primary language is Spanish (es or es-*). Other or unavailable languages fall back to English. The same rule applies to controls, accessible names, status labels and dates; the work does not store a language preference.

Stage names come from their declared stage key and appear once in the browser's language. An agent may add short contextual text after that name. Task descriptions, custom activity titles, captions and published documents stay exactly as authored, so Spanish content can legitimately appear inside an English interface. The offline `hive view contract` catalog exposes the supported languages and reserved label keys.

## Follow a work

Within each project group, works are ordered by their last substantive work activity, with its local date and time shown alongside the state: In progress, In review, Awaiting validation, Paused, Finished or Failed. Renaming, pinning, archiving and restoring do not change that activity. Select a work to see its timeline in chronological order. Stages group their activity; **Go to current** takes you to the current stage. Open **Sessions** to see the participating producers and the lead. A producer can contribute without taking the lead; the browser can reassign the lead through the explicit **Reassign lead** dialog when recovery is needed, or `hive view work recover` can do it from the terminal.

Open **Tasks** in the horizontal toolbar for the global task list; each stage shows only its own task segments. The list shows published task states and their segments in each stage. A task can span stages: for example, seven tasks can have six segments completed in Build while the seventh has a completed Verification segment and a pending Delivery segment. Its overall state is not complete until its work is complete across its segments. Tasks with review content open a detail panel with **Summary**, **Evidence** and **History**. A task containing only a short description expands inline instead of opening an empty panel. Image, document, block and CI indicators are visible even while the task is collapsed. For a blocked task, **See how to unblock** opens its instructions, including the responsible person and the condition for resuming when published.

## Project groups and source

The sidebar groups registered roots that explicitly declare the same `project_name` in their root `.engram/config.json`. For example, a workspace and its API repository can both declare `example-project`: their works appear in one flat group, with the original workspace or repository shown beside each work's ticket and date. This reads the local configuration file; it does not require Engram to be installed or running. Names are compared exactly after trimming outer whitespace, including case.

Grouping is navigation only. Each work keeps its ID, links, original source, documents and collaboration permissions. The viewer does not infer membership from parent folders, matching `AGENTS.md` names or an Engram alias. A root without this configuration keeps its original project group. Changing or removing the name updates the grouping on the next refresh, including existing works; an invalid or unreadable configuration leaves works accessible in their original groups and shows a retryable notice.

Register a configured non-Git workspace directly with `hive view register --dir /path/to/workspace`. The directory itself must contain a valid `.engram/config.json` with a nonempty `project_name`; registration does not search its parents or register its children. Existing Git registration and registration from `_support/workspace` keep working. Repeating registration does not duplicate the root.

## Inspect evidence and documents

Open a capture to view its gallery and move between images. Open a document to inspect its revisions; when an event has multiple document chains, select the chain before choosing a revision. The detail you opened is a snapshot. If a newer version is published, the viewer signals an update without replacing what you are reading. Use **Show latest** where offered, or close and reopen a document to see the latest revision.

References resolve to the current work's canonical records. **Published** preserves the context recorded at publication; **Current** shows the current record when it differs. The published context remains historical evidence rather than silently changing with the record.

## Read delivery checks

Open the CI detail for **Summary**, **Checks** and **Attempts**. The viewer displays published observations and the applicable attempt for each check; an unconfirmed required-check policy or observation remains unconfirmed. A later observation from an older attempt does not make the current attempt pass. Viewing this information does not rerun checks, and passing checks do not confirm integration or merge.

## Manage a work

Use **Edit title** to rename a work and **Session actions** to archive or restore, pin or unpin, or delete it. A finished, unpinned work is automatically archived two hours after its recorded closure; if that time cannot be established, it gets a full two-hour window from the first evaluation. Other states are not automatically archived. Archiving changes visibility only: use **Include archived** to find a work and **Restore** to bring it back. Pinning prevents automatic archival; unpinning uses the original deadline. A manual restore exempts that closure until the work is reopened and closed again. Reopening restores an automatically archived work, while a manually archived work stays archived. Manually archiving shows a temporary notification naming the archived work, with **Undo** to restore it. The notification hides after eight seconds or can be dismissed; navigating to another work does not make that work archived. Content retention is manual, with no age-based deletion or TTL setting. Deleting a session removes its stored timeline events and captures from the viewer; original project files are preserved. Each management action also has a command, `hive view work rename|archive|restore|pin|unpin|delete`; `delete` needs `--confirm` with the work ID. They act as the local viewer, distinct from the native producer's publication of work and evidence, and an agent runs them only when you ask for that action on a named work. For the agent's procedure, see [flow-walkthrough](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md).
