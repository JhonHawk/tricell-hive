---
title: Getting started
description: Install Hive and start your first deliberate session with a coding agent.
---

Install your coding CLI separately first. Hive supplies guidance; your CLI keeps
its own models, sign-in, permissions, tools, and history.

## Install Hive

Close affected CLI sessions. On macOS Apple Silicon or Linux (arm64 or amd64), run:

```sh
curl -fsSL https://hive.tricell.tech/install.sh | sh
```

The installer asks which hosts to install and confirms before writing. To preview
the installation first:

```sh
curl -fsSL https://hive.tricell.tech/install.sh | sh -s -- --dry-run
```

The preview downloads the package but makes no host changes. Start a new CLI
session after installation. Users of 0.8.0 and earlier must run the installer
once to migrate to 0.9.0 or newer; the old updater cannot replace its binary.
Use `~/.local/bin/hive` if PATH or an alias still selects an old binary. Once
migrated, use `hive update` to preview and confirm a stable release; use
`hive update --dry-run` for a preview. Existing public packages can be updated
with the installer. For requirements, other installation routes, and recovery, see the
[repository's installation guide](https://github.com/JhonHawk/tricell-hive#get-started).

## Give the session context

Add a `## Hive` section to your repository's `AGENTS.md` with the project name,
base branch, tracker, and specs folder. The agent asks for missing settings when
the work first needs them.

The section holds facts about the repository, not decisions. Optional keys in it
are `Environments`, `Other trackers` (trackers besides the main one, such as a
client's Jira, each with its ticket key prefixes), `Delivery: direct-base` (marks
a support repository), `Completion` (a later completion point for tickets, one
per tracker), and `Hive guidance`. With `Other trackers`, the agent reads and
searches every declared tracker, moves and comments each ticket in the tracker its
key belongs to, and creates new tickets only in the main one.
The delivery mode, the review mechanism and the `UI verification` level are not
keys: the agent asks them each session and reuses your answers for that session.
A section that still holds `Review`, `UI verification` or `Delivery: automatic`
or `interactive` is retired: it answers no question, the agent asks it each session,
and the old value only informs the recommended option. `hive doctor` warns
about those lines.
Each key's meaning and accepted values are in
[project settings](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).
`hive project set` edits the section for you after a confirmation, and the
[harness audit](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/harness-audit/SKILL.md)
can create or update it for you: ask it to set up your repository's Hive section,
and it proposes the facts with their evidence, asks the decisions it cannot
deduce, and writes only what you accept.

Facts that hold only on your machine, such as the ports your local stack ends
up on, a command your system lacks, or the cloud profile and region of each
environment, go in `HIVE.local.md` at the repository root, or at the workspace
root when your work spans several repositories. You neither create nor commit
it: the agent reads it at the start of each session, adds there each fact it
verified with its date, marks the ones that stop working as stale, names the
file that holds a password instead of writing it, and adds the file to
`.git/info/exclude` so Git ignores it. You can edit it by hand. A fact that
holds for every clone belongs in `AGENTS.md`, so the agent names it in its
report instead. The [local instructions rule](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md)
holds the details.

Then ask for an outcome and state its limits:

```text wrap
Investigate why this endpoint times out. Read only; report evidence and options.
```

For an understood change, you can request implementation directly:

```text wrap
Implement the agreed change locally and verify it. Do not commit or publish.
```

## Choose the right depth

If a CLI behaves differently from what these pages describe, check [Host limits](/host-limits/).

Start with [Research](/flows/research/) when facts are missing, [Plan](/flows/plan/)
for consequential decisions, or [Build](/flows/build/) for an authorized change.
[Close](/flows/close/) records the result and cleans up the session. The
[Skills catalog](/skills/) links to each procedure and supporting capability.
