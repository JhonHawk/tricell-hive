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
session after installation. Run the installation command again when you want to
update. For requirements, other installation routes, and recovery, see the
[repository's installation guide](https://github.com/JhonHawk/tricell-hive#get-started).

## Give the session context

Add a `## Hive` section to your repository's `AGENTS.md` with the project name,
base branch, tracker, and specs folder. The agent asks for missing settings when
the work first needs them.

Optional keys in the same section let you answer the recurring questions once:
`Environments`, `Review`, `Delivery`, `UI verification`, and `Hive guidance`.
Each key's meaning and accepted values are in
[project settings](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/references/project-settings.md).
`hive project set` edits the section for you after a confirmation.

Then ask for an outcome and state its limits:

```text wrap
Investigate why this endpoint times out. Read only; report evidence and options.
```

For an understood change, you can request implementation directly:

```text wrap
Implement the agreed change locally and verify it. Do not commit or publish.
```

## Choose the right depth

Start with [Research](/flows/research/) when facts are missing, [Plan](/flows/plan/)
for consequential decisions, or [Build](/flows/build/) for an authorized change.
[Close](/flows/close/) records the result and cleans up the session. The
[Skills catalog](/skills/) links to each procedure and supporting capability.
