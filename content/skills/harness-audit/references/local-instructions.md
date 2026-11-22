# Create or curate HIVE.local.md

Read from [harness-audit](../SKILL.md) when the user asks to create, bootstrap, review or prune `HIVE.local.md`, or to record how the project runs on this machine; not for recording one fact verified during other work, which [local record](../../hive-core/local/local-record.md) covers. This mode replaces the audit procedure.

[local record](../../hive-core/local/local-record.md), [local write](../../hive-core/local/local-write.md) and [local script](../../hive-core/local/local-script.md) are the canonical home of `HIVE.local.md` (the project's file and the machine file): where each lives, what an entry is, secrets, size, what `unverified` means and how Git ignores it. This mode adds the first fill and the pruning, and changes one element of those rules, named in step 5.

*Create* when no `HIVE.local.md` exists where [local write](../../hive-core/local/local-write.md) places it. *Curate* when it exists and the user asks to rebuild, review or prune it, or after [local write](../../hive-core/local/local-write.md) recommended this mode past 50 lines; curate lists the existing entries beside the new candidates, so no entry, the user's hand-written ones included, leaves the file without the user's answer. Work only in the current project and the machine file.

## 1. Read what exists

In the main thread, read each `HIVE.local.md` that [local write](../../hive-core/local/local-write.md) places for this session; the project's tracked guidance on running it locally (the `AGENTS.md` files from the working directory up to the repository or workspace root, and the README's setup section), whose facts are not recorded, as [local record](../../hive-core/local/local-record.md) sets; `ACCOUNTS.md`, only to point to it; the names, never the contents, of the files in the secret location; the host's own memory and local notes for this project (such as Claude Code's project memory folder and `CLAUDE.local.md`); and persistent memory, with queries of one or two distinctive terms such as the project name with `local`, `port`, `start`, `region` or `profile`.

## 2. Extract candidates from transcripts

```sh
python3 -I <harness-audit-dir>/scripts/extract_candidates.py --root <repository root> [--root <declared workspace root>] --days 30
```

It reads the top-level Claude Code and Codex sessions of the last 30 days started in those roots, below or above them, keeps each command that ran in a root, leaves out subagent and background sessions, and prints JSON: `commands` that exited 0 (with the environment variables they set), `missing` commands, `ports`, and `cloud` profiles and regions, each with its session count, last-seen date and `where` (the folder below the root it ran in, `.` for the root itself, which picks the repository's or the workspace's file). It prints no command output and masks every value that could be a secret; never open a transcript to check a candidate. `hosts` gives per host the transcripts `read` and the `sessions` with a command in a root, and `not_covered` the hosts it does not read; a host you used that shows zero `read` has transcripts the script does not locate (it reads Claude Code and Codex only): say so in the report.

## 3. Classify

Sort every candidate into one of:

- **record:** it passes the test [local record](../../hive-core/local/local-record.md) sets. The `missing` candidates, existing cloud profiles with their regions and installed tools go to the machine file; the `commands`, `ports` and which profile serves an environment of the project go to the file their `where` names. Each becomes one entry as [local record](../../hive-core/local/local-record.md) defines it (a `commands` candidate is a line that says what it starts or does, with the command in its detail when it does not fit);
- **every clone:** a fact [local record](../../hive-core/local/local-record.md) keeps out of the files because it holds in any checkout; it goes to step 7's report;
- **discard:** a `missing` program the project's own dependencies provide, a stack, branch, ticket or path that no longer exists, a duplicate, a fact a later session contradicts, a command a single session used for one task, or anything [local record](../../hive-core/local/local-record.md) excludes.

In curate mode give each existing entry one proposal instead: **keep** it; **split** it, with the line that stays and the detail that moves to persistent memory, when it is longer than its line or holds steps; **relocate** it to another file; or **retire** it when [local record](../../hive-core/local/local-record.md) excludes it or it is a discard. An instruction of the user found among the entries is split too: one line saying what it constrains, and the instruction saved as a decision in the file's project, or in `hive-machine` for the machine file.

Group the record candidates by file, then by subject (starting the stack, isolated stacks, tools and missing commands, cloud and remote access, other). Recommend keeping a candidate seen in two or more sessions or confirmed by a step 5 check.

## 4. Ask what to keep

Before writing anything, list in the message text every group with its candidates numbered, each with its session count and last date, and the **discard** class as one more group (in curate mode, each existing entry with its proposal and the line it would become). Then ask through the host's native question tool, as [questions](../../hive-core/messages/questions.md) sets, one question per group, with these options: `A - (★) Keep the recommended` (the numbers you recommend, named in the option); `B - Keep all`; `C - Discard all`; `D - Choose by number`. A `D` answer needs the user's numbers in the chat: finish the round's cards, then ask in one message for the numbers of each group or file answered `D`; that message after the last card is this reference's exception to the rule that the turn ends with the round's last card. In curate mode, also ask one question per file about its existing entries (`A - (★) Apply the proposals`; `B - Keep every entry as it is`; `D - Choose by number`); an existing entry is split, relocated or retired only when the answer applies its proposal. A candidate the user does not keep is not written. The answers live in `HIVE.local.md`, not in persistent memory, because the file is the user's record; only the detail of an entry and an instruction of the user whose proposal the answer applied (saved as a decision) go to persistent memory, as [local record](../../hive-core/local/local-record.md) sets. Discarding a candidate deletes nothing from its source: cleaning host or persistent memory is `harness-audit`'s job.

## 5. Confirm what the user kept

Confirm each kept fact with one of these read-only checks when one applies, and run nothing else:

- `command -v <tool>`; then `<tool> --version` only when the entry records that the project pins a version this machine lacks, only for a tool resolved outside a version-manager shim folder (`shims`, `.asdf`, `mise`), never for `pnpm`, `yarn` or `corepack`, which may download a version, and never through `npx`, a package runner or a project script;
- `docker ps --format '{{.Names}} {{.Ports}}'` and `docker compose ls`;
- `lsof -nP -iTCP:<port> -sTCP:LISTEN`;
- `aws configure list-profiles` and `aws configure get region --profile <profile>`;
- `test -e <path>`;
- `git config --get <key>` only for `user.name`, `user.email`, `init.defaultBranch` and `core.*` keys.

Never run a project command that starts, stops, seeds, migrates, installs, signs in or writes, even one the user kept: it stays as recorded. A fact a check confirms takes today's date; a fact no check covers, or whose check fails while the user keeps it, takes the date it was last seen and the mark `unverified`. This is the one element of [local record](../../hive-core/local/local-record.md) this mode changes: there, an entry needs this session's confirmation.

## 6. Write the files

Write each **record** candidate to the file step 3 names, through the operations of [local write](../../hive-core/local/local-write.md) and their result checks, with `unverified` where step 5 requires it. Save and read back each detail before removing it from a file. In curate mode, keep the user's wording for entries they kept.

## 7. Report

Report in the session language: each file written with its line count, the entries written per group and how many are `unverified`, what the user discarded, each **every clone** fact as a candidate for the repository's tracked guidance, and the hosts the script did not cover.
