---
name: local-instructions
description: Create or curate a project's HIVE.local.md from what already worked on this machine, by gathering candidates from recent session transcripts, persistent and host memory, local notes and the project's guidance, asking the user which to keep, and confirming them with read-only checks. Use when the user asks to create, bootstrap, review or prune HIVE.local.md, or to record how the project runs on this machine; not for recording one fact verified during other work, which the global guidance's «Local instructions» covers.
---

# Create or curate HIVE.local.md

The global guidance's «Local instructions» is the canonical home of `HIVE.local.md`: where it lives, what goes in it, its secrets, its size, what `unverified` means and how Git ignores it. This skill adds only the first fill and the pruning, and changes one element of that section, named in step 5.

**Modes.** *Create* when no `HIVE.local.md` exists where «Local instructions» places it. *Curate* when it exists and the user asks to rebuild, review or prune it, or after «Local instructions» recommended this skill past 100 lines; curate lists the existing entries beside the new candidates, so no entry, the user's hand-written ones included, leaves the file without the user's answer.

Work only in the current project, as the global guidance's «Scope and authorization» sets.

## 1. Read what exists

In the main thread, read:

- each `HIVE.local.md` that «Local instructions» places for this session;
- the project's tracked guidance on running it locally: the `AGENTS.md` files from the working directory up to the repository or workspace root, and the README's setup section. A fact found there holds for every clone and is never written to the file;
- `ACCOUNTS.md`, only to point to it: accounts stay there;
- the names of the files in the secret location, never their contents, as the global guidance's «Secrets» sets;
- the host's own memory and local notes for this project, such as Claude Code's project memory folder and `CLAUDE.local.md`;
- persistent memory, with short queries of one or two distinctive terms, such as the project name with `local`, `port`, `start`, `region` or `profile`.

## 2. Extract candidates from transcripts

Run, from the main thread:

```sh
python3 -I <this skill>/scripts/extract_candidates.py --root <repository root> [--root <declared workspace root>] --days 30
```

It reads the top-level Claude Code and Codex sessions of the last 30 days started in those roots, below or above them, keeps each command that ran in a root (its own directory, moved by any `cd` before it on the line), leaves out subagent sessions and commands started in the background, and prints JSON: `commands` that exited 0 (with the environment variables they set), `missing` commands, `ports`, and `cloud` profiles and regions, each with its session count, the date it was last seen, and `where`: the folder below the root it ran in, `.` for the root itself. Use `where` to choose the repository's or the workspace's file. It prints no command output. It masks every value that could be a secret before printing; never open a transcript to check a candidate, since the transcript holds the values the script masked. `hosts` gives per host the transcripts `read` and the `sessions` with a command in a root, and `not_covered` the hosts it does not read; when a host you used shows zero sessions, read [transcript sources](references/transcript-sources.md).

## 3. Classify

Sort every candidate, and in curate mode every existing entry, into one of:

- **machine:** it names something of this machine, account or environment, the test «Local instructions» sets;
- **every clone:** it holds for any checkout, such as a test command already in the tracked guidance; it goes to step 7's report, never into the file;
- **discard:** a stack, branch, ticket or path that no longer exists, a duplicate, a fact a later session contradicts, or a command a single session used for one task.

Group the machine candidates by subject: starting the stack, isolated stacks, tools and missing commands, cloud and remote access, and other. Recommend keeping a candidate seen in two or more sessions or confirmed by a step 5 check.

## 4. Ask what to keep

Before writing anything, list in the message text every group with its candidates numbered, each with its session count and last date, and the **discard** class as one more group. Then ask through the host's native question tool, as the global guidance's «Questions and routes» sets, one question per group, with these options:

- `A - (★) Keep the recommended`: the numbers you recommend, named in the option;
- `B - Keep all`;
- `C - Discard all`;
- `D - Choose by number`: the user replies in the chat with the numbers to keep.

Ask at most as many questions per call as the tool allows and the rest in a later call. A candidate the user does not keep is not written; in curate mode, an existing entry the user does not keep is removed. The answers live in `HIVE.local.md`, not in persistent memory: this departs from the global guidance's «Persistent memory», which saves a user's decision, because the file is their record. Discarding a candidate deletes nothing from its source: cleaning host or persistent memory is `harness-audit`'s job.

## 5. Confirm what the user kept

Confirm each kept fact with one of these read-only checks when one applies, and run nothing else:

- `command -v <tool>`; then `<tool> --version` only for a tool `command -v` resolved to a path outside a version-manager shim folder (such as `shims`, `.asdf` or `mise`), never for `pnpm`, `yarn` or `corepack`, which may download a version, and never through `npx`, a package runner or a project script;
- `docker ps --format '{{.Names}} {{.Ports}}'` and `docker compose ls`;
- `lsof -nP -iTCP:<port> -sTCP:LISTEN`;
- `aws configure list-profiles` and `aws configure get region --profile <profile>`;
- `test -e <path>`;
- `git config --get <key>` only for `user.name`, `user.email`, `init.defaultBranch` and `core.*` keys.

Never run a project command that starts, stops, seeds, migrates, installs, signs in or writes, even one the user kept: it stays as recorded. A fact a check confirms takes today's date. A fact no check covers, or whose check fails while the user keeps it, takes the date it was last seen and the mark `unverified`. This is the one element of «Local instructions» this skill changes: there, an entry needs this session's confirmation.

## 6. Write the file

Write as «Local instructions» sets. Start the file with one line saying it is per-machine, untracked and kept by agents under «Local instructions». Give each group a heading and each entry one line: the fact, its command when it has one, and its date, followed by `unverified` when marked. In curate mode, keep the user's own wording of an entry they kept.

## 7. Report

Report in the session language: each file written with its line count, the entries written per group and how many are `unverified`, what the user discarded, each **every clone** fact as a candidate for the repository's tracked guidance, and the hosts the script did not cover.
