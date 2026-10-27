# Deployment manager

The first manager is implemented in Go 1.27. Its core packages (`tooling/management` and `tooling/distribution`) use only the standard library, and a test walks their imports to keep it so. The command-line package depends on `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`, and `github.com/charmbracelet/x/ansi` for the terminal interface only, a deliberate exception to the no-dependency rule because the interface is a user-requested product feature, not configuration; building a release therefore downloads those modules, while installing the offline package does not change. It runs from this checkout with `go run ./tooling/cli`; macOS is the tested platform. The manager never launches models. The separate pilot runner is evaluation tooling.

## Commands

For the one-command user experience, use the offline package's `./install.sh`.
It invokes `hive install` and reuses this manager's planning, transactions and recovery.
See the [installer contract](installer.md) for legacy detection and package verification.

Run from the checkout root. Both hosts and scope must be explicit for planning and status:

```sh
go run ./tooling/cli --help
go run ./tooling/cli setup
go run ./tooling/cli plan install --hosts codex,claude,grok,pi,opencode,cursor --scope user
go run ./tooling/cli plan install --hosts codex --scope project --root /absolute/project --out /absolute/new-plan.json
go run ./tooling/cli apply --plan /absolute/new-plan.json
go run ./tooling/cli status --hosts codex --scope project --root /absolute/project
go run ./tooling/cli plan remove --hosts codex --scope project --root /absolute/project --out /absolute/remove-plan.json
go run ./tooling/cli plan remove --hosts codex --scope user --out /absolute/remove-user-plan.json
go run ./tooling/cli recover --state-dir /absolute/state-directory
```

These are interface examples, not authorization to deploy into the user's real configuration. A plan that installs Grok's [read-whole hook](#grok-read-whole-hook) refuses a `go run` binary; build one with `go build -o <path> ./tooling/cli` and run that instead. `plan` without `--out` is a read-only preview. With `--out`, it writes a new private plan file, never the destination files. `apply` consumes that plan and rejects stale target fingerprints or ownership state. Destination options are not accepted by `apply`: review the destinations bound into the plan instead.

`--home` explicitly selects a synthetic home and ignores host path and compatibility environment overrides. Use it together with `--state-dir` for temporary-home tests. Without `--home`, host configuration environment variables are respected. `--root` is required for project scope. Existing root prefixes are canonicalized; linked target files or ancestors inside those roots are rejected rather than overwritten.

The default source is the current directory; `--source` selects another checkout with the same content layout. `plan install --release <hash>` selects a retained installed snapshot instead, allowing an explicitly planned downgrade. The current binary validates that snapshot's agent sources, so a release whose roles use a retired frontmatter field, such as `claude_effort` before it became `effort`, fails with `unsupported agent field`; plan that downgrade with the manager from the commit that produced the release. Installed releases still update, report status, uninstall, and recover, because those operations do not re-render the old sources. A plan freezes its source bytes; later source edits do not silently alter it.

### Update from a commit and list releases

```sh
go run ./tooling/cli update [--rev HEAD] [--source .] [--home DIR] [--state-dir DIR] [--dry-run] [--out FILE]
go run ./tooling/cli releases [--home DIR] [--state-dir DIR]
```

`hive update` deploys the content of one commit, never uncommitted edits in the checkout. It resolves `--rev` (default `HEAD`) with `git rev-parse` in `--source` (default `.`), with the variables `git rev-parse --local-env-vars` lists removed from Git's environment, archives that commit by its full hash, extracts it into a private temporary directory, and plans an install for every host already registered in user scope. The plan freezes the extracted bytes, so the temporary directory is removed on every return path as soon as the plan is built, before the summary and the confirmation prompt; only a process killed by a signal during extraction or planning may leave a `.hive-extract-*` directory in the system temporary directory. A revision starting with `-` is rejected before Git runs. The command needs Git and a Git checkout of Hive; the offline package does not, and installs through `install.sh`. `export-ignore` or `export-subst` attributes change the archived content, as with any `git archive`: those in the selected commit, and also local ones in `.git/info/attributes` or the file named by `core.attributesFile`, which are not part of the commit.

In a terminal, `update` shows the summary with the commit and asks for confirmation. When nothing changes, `--out FILE` still saves the plan, and applying it records the commit; without `--out`, `update` applies it without asking, which only records the commit, also without a terminal. When the plan changes files and there is no terminal, it fails unless `--dry-run` previews the change or `--out FILE` saves a plan for `hive apply --plan FILE`. In an interactive session, `update` also offers the Codex wait floor (`codex-wait-floor`, selected by default) while Codex's `config.toml` lacks it and Codex is 0.164 or later (a dry run offers nothing, an applied setting is not offered again, and nothing is offered at project scope); a plan that writes that file always asks for confirmation, even when the Hive content is unchanged. `hive plan remove` for the codex host undoes it. See the [installer contract](installer.md). To return to an older commit, run `update --rev <commit>`; the current binary must still validate that commit's sources.

The install and update summary, in the text commands and in the full-screen views that reuse it, adds one line before `Close these CLI sessions before continuing.` when Grok and Cursor are both in the plan or already registered and Grok's `compat.cursor.agents` is on: it names `<Cursor home>/AGENTS.md`, says Grok loads that copy of the guidance a second time (about 10,700 tokens per session), and gives the fix, `[compat.cursor]` `agents = false` in `<Grok home>/config.toml`, with its side effect (the user's own text in `~/.cursor/AGENTS.md` stops loading in Grok). The setting resolves as `GROK_CURSOR_AGENTS_ENABLED` (ignored with `--home`), then the TOML table, then on. Hive only warns and never writes Grok's config; a setting it cannot read or parse omits the line and fails nothing, unlike `compat.claude`, which stays a conflict.

Applying a plan built by `update` records its source commit in `releases/<id>.commits.json`, also when the plan changes no file, because different commits often produce the same release. The release snapshot and its ID do not change. If writing the record fails after the core committed, the result carries a warning and the installation stands. Recovering an already committed journal leaves that release without the commit. Releases installed before this record existed list no commits.

`hive releases` prints, as a JSON array, every retained snapshot from the most recently written to the oldest: `id`, `last_written_at` (the time its snapshot was last written), `commits` (recorded source commits in the order they were applied, possibly empty), and `consumers` (the hosts that currently have it installed, possibly empty). It reads state only.

### Optional voice layer

```sh
go run ./tooling/cli voice list [--source .]
go run ./tooling/cli voice set <id> [--address sir|name|none] [--name NAME] [--intensity subtle|marked] [--source .] [--home DIR] [--state-dir DIR] [--dry-run] [--out FILE]
go run ./tooling/cli voice off [--source .] [--home DIR] [--state-dir DIR] [--dry-run] [--out FILE]
```

A voice is an optional tone for the messages the user reads in the main conversation. It is off by default and never changes the Hive rules. Its text comes from `content/voices/`: `preamble.md`, shared by every voice, followed by `<id>.md`, one line for the form of address (default `none`), and one for the intensity (default `subtle`). The preamble states that the voice never overrides the Hive rules, lists what it never changes, limits tone to transitions, closings, and the form of address, and tells any agent launched by another agent, or writing files, commits, pull requests, tickets, or specifications, to ignore it. That last condition is text, not enforcement: a subagent that reads the global file can still adopt the voice. Voices are not part of the release catalogue, so they never change a release ID.

`voice set` writes one voice span per user-scope instruction file that already holds a Hive block this manager owns, with its own markers, `<!-- === TRICELL HIVE VOICE:BEGIN === -->` and `<!-- === TRICELL HIVE VOICE:END === -->`, placed right after the Hive block's end line. There is one voice per home, applied to every registered user-scope host. `voice off` removes every span and returns each file to its bytes before `voice set`. Both follow the `update` confirmation pattern: a summary, confirmation in a terminal, `--dry-run` to preview, and `--out FILE` to save a plan for `hive apply --plan FILE`.

With a voice active, `plan install`, and therefore `hive install` and `hive update`, keeps the span and regenerates it with the same choice when the voice source text or its rendered form changed; the summary then reports the voice files to regenerate. A host that gains its Hive block in the same plan gets the voice too. A source that cannot render the chosen voice, such as `plan install --release <hash>`, a commit from before `content/voices/` existed, or one where that voice was removed, leaves the spans untouched and the summary says so. `plan remove` removes a file's voice span with its last consumer, and forgets the voice once no span remains. `status` adds a `voice` row per span with the voice, address, and intensity, and reports `drift` when the span was edited by hand. A hand-edited span, or voice markers in a file with no recorded span, is a conflict that every operation on that file preserves.

The voice is stored in `state.json` as `voice` and `voice_spans` without a schema change. A manager from before this feature keeps those fields while it only reads the state, but drops them the next time it writes it, leaving the spans in the files; its `plan remove` also leaves a file's voice span behind. The current manager then reports those spans as unregistered conflicts, and every operation on such a file, `hive voice off` included, stops until the user deletes the voice span by hand, from its begin marker through its end marker. Run `hive voice off` before returning to an older manager.

### Read-only diagnostics

```sh
go run ./tooling/cli doctor [--home DIR] [--state-dir DIR] [--project DIR]
go run ./tooling/cli models [--home DIR] [--state-dir DIR]
```

`hive doctor` prints five sections with the same text as the Diagnostics, Integrations, and Project views: CLIs, Installation, Sessions, Integrations, and Project. `hive models` prints, per registered host, the effective model and effort of each installed role, with the same words as the Models view. Both write no file or state and exit with status 0 when they report findings; only their own errors, such as a state directory that is a file, give a non-zero status. `hive doctor` has one further failure, for its Project section: it exits non-zero, printing `hive: Project section has blocking findings` on standard error after the report, when the directory is inside a Git repository whose root `AGENTS.md` exists and cannot be read as a regular file within 1 MiB, or holds a `## Hive` section with a blocking finding (see **Project** below). When the repository has no `AGENTS.md` or no section, or the path does not exist, is not a directory, is outside a Git repository, cannot be read, or `git` cannot run, it exits non-zero only when `--project` is given explicitly, so `harness-audit` can tell that a section it wrote did not land. It exits 0 for warnings only, and without `--project` in those cases.

- **CLIs** lists each supported host in installer order: detected on `PATH` or not (Cursor counts as detected when either `cursor-agent` or `cursor` is found, in this report and in `hive install` and the CLIs view alike; both read one list of binary names), `CLI` with the first line of `<binary> --version` (`cursor-agent` for Cursor, so a host with only `cursor` is labeled `editor only` instead of `detected` and shows `CLI unavailable: cursor-agent not found`; the CLIs view labels it `editor only` in its State column too, and it and the installer still offer Cursor then), and `Hive` with the host's release and installation state (`verified`, `partial`, `drift`, `legacy`), as in `claude  detected  Hive 4f8d96b9ce4f (verified)  CLI 2.1.284 (Claude Code)`, so the CLI's version is not mistaken for Hive's. Only a detected host is executed, with fixed arguments, no shell, and a 3-second limit; a failure or timeout shows `CLI unavailable` and its reason.
- **Installation** lists every `status` row other than `installed`, `retained_shared`, and `not_installed`, with a plain phrase and its path, or `No problems found`. Hosts that share one managed file, such as a skill under `~/.agents/skills`, are grouped into one row per status and path that names them all, as in `drift  claude, codex, cursor, grok, opencode, pi  The installed file was changed or cannot be read` followed by the path once, in the order the rows first appear. When a row is `drift`, one line says what to do in either case: a file or block that was edited, or whose permissions changed, makes `install`, `update`, and `plan remove` refuse, so the user undoes the change or fixes its permissions; one that was deleted is put back by `hive install` or `hive update`. Then the user opens Diagnostics again, or runs `hive doctor`, to check. With a pending operation it shows one line naming `hive recover` instead of the `recovery_required` rows, because `status` then marks every row that way; the line adds `--state-dir DIR` unless that directory is the real user home's default, since `hive recover` has no `--home`.
- **Sessions** covers the registered hosts. When Claude Code or Grok is registered, its first line answers whether to restart anything for the hosts Hive checked, such as `2 open Claude Code or Grok sessions should be restarted` or `No open Claude Code or Grok session needs a restart`, and names a host it could not check, as in `No open Claude Code session needs a restart; Grok could not be checked.` When Codex, Pi, or Cursor is registered, the next line says `Hive cannot see Codex, Pi or Cursor sessions; restart them after each update.`, naming only the registered ones. Each host's line then reads, for example, `claude: 3 open sessions, 1 to restart`. It reads Claude Code's `sessions/*.json` under `CLAUDE_CONFIG_DIR` or `~/.claude` and Grok's `active_sessions.json` under `GROK_HOME` or `~/.grok`. It counts only sessions whose process is alive and marks one `started before the installed release; restart it` when it started before the installed release's snapshot was last written. It shows only the process ID, working directory, and start time. A missing session directory or file means no open sessions; a file larger than 1 MiB, unreadable, not a regular file, or in an unknown format, or more than 500 Claude Code session files, makes that host `Session check unavailable` without failing the other sections. OpenCode gets a note that it applies `AGENTS.md` instruction changes on the next message and that role or skill changes need a new session.
- **Integrations** shows Engram, Context7, pi-subagents, `agent-browser`, CodeGraph, only when Pi is a registered host a `Pi codemode` row, and only when Codex is a registered host (or the registration read fails) a `Codex question card` row and a `Codex wait floor` row: the local evidence (the program on `PATH` or the skill file, with its path), the status in the last onboarding record, the official source, and a concrete next step for its case. The Hive-suggested commands are Context7's `npx ctx7@latest setup --cli`, which may ask the user to sign in, and `hive install` for `agent-browser` when it is absent or has no browser, whose record shows the worst status of its onboarding steps; for the others it points to the Source line above, since Hive does not install them. The `agent-browser` check reads `PATH` and files only; the CodeGraph check only looks `codegraph` up on `PATH`, never runs it, and its next step points to CodeGraph's own `codegraph install` and to indexing a repository only by the user's decision. The `Pi codemode` row is built from constants, whether or not codemode is on (`not checked`), and its next step names the `"defaultTools": ["+codemode"]` line for Pi's own `settings.json` and Pi's `/reload`; Pi turns codemode on by itself with MCP servers of `codemode` exposure. Registration is read from Hive's own state (a failed read still shows the row), never from Pi's configuration. Apart from the `Codex wait floor` row, described below, it executes no integration program and reads no host or provider configuration, so pi-subagents shows only its onboarding status and `Pi codemode` none. The `Codex question card` row is built the same way: `not checked`, the catalog source, the status of its `codex-question-card-*` onboarding steps, and a next step naming `codex features enable default_mode_request_user_input`, `codex features disable …` and `codex features list`; it never reads Codex's `config.toml`. The `Codex wait floor` row is the exception: it reads that file and reports whether the two `min_wait_timeout_ms` and `default_wait_timeout_ms` keys under `[features.multi_agent_v2]` are set, whether Hive's state records that it wrote them, and whether `config.toml.hive-backup` exists. It runs `codex --version` only to warn when the table exists with a Codex older than 0.164. The view reserves room for eight rows at 80x24.
- **Project** validates the `## Hive` section of `AGENTS.md` at the root of the Git repository that contains `--project`, or the current directory: missing or duplicated section, missing or empty required values (`Project`, `Base branch`, `Tracker`, `Specs`), a `Specs` path that is not a directory, a `Base branch` that is neither a local branch nor on `origin`, unknown keys, a `Delivery` value other than `direct-base`, or a `Hive guidance` value other than `required`. The retired keys `Review` and `UI verification`, and `Delivery: automatic` or `interactive`, warn without blocking (`<Key>: no longer read; Hive asks it each session — remove it with hive project set --unset <Key> or run harness-audit`, with `Delivery: <value> is no longer read; …` naming the retired value) and are not unknown keys; `Completion` is a known key and warns when it lacks the shape `<state> on <ref> · <state> on <ref>`. The blocking findings, which fail the command, are a missing or empty required value, a `Base branch` name that Git rejects, an invalid `Delivery` or `Hive guidance`, a duplicated section, and an `AGENTS.md` that is not a regular file, cannot be read, or exceeds 1 MiB. A `Base branch` that could not be checked because `git` failed to run is a warning for the doctor, but `hive project set` refuses to write it. The retired keys named in `retiredHiveKeys` must match the "Retired keys:" line of `project-settings.md`, which a test checks. A `--project` directory that does not exist, or is not a directory, is reported as `Directory not found` or `Not a directory` without running `git`. It runs read-only `git` queries with Git's local environment variables removed. A section in a workspace `AGENTS.md` outside a repository is not checked.

With `--home`, as elsewhere, host environment overrides are ignored, no program is looked up on `PATH` or executed (neither the host binaries nor `engram`, `agent-browser`, `codegraph` and `codex`), and sessions are read under that home; the Project check still finds and runs `git` in `--project` or the current directory. Text read from outside the manager (versions, paths, session directories, `AGENTS.md` values) has control characters and escape sequences removed before it is shown.

Limits: the session files of Claude Code and Grok are internal and undocumented, so any host version may change them and turn the check into `Session check unavailable`. The restart mark is an estimate: a release snapshot is rewritten by every transaction that carries it, such as adding a host, which moves the reference time forward and can mark sessions that already have the current content; `/compact` in Claude Code and voice changes do not move it; the write time has one-second precision; and a reused process ID can make a finished session look alive. Codex, Pi, and Cursor have no session detection.

`hive doctor` and `hive models` without a subcommand never write. Models and `## Hive` values are changed with the commands below or the Models and Project views.

### Model overrides

```sh
go run ./tooling/cli models set --host H (--role R | --group G) [--model M] [--effort E] [--home DIR] [--state-dir DIR] [--dry-run | --out FILE]
go run ./tooling/cli models reset --host H ((--role R | --group G) [--only model|effort] | --all) [--home DIR] [--state-dir DIR] [--dry-run | --out FILE]
```

`integrations/agent-profiles.json` holds the release defaults; `models set` overrides the model, the effort, or both for one role on one registered user-scope host. The override is stored in `state.json` as `model_overrides` and takes precedence over the role's profile and its own effort. It does not change the release ID or its snapshot. The plan reinstalls that host from the release it has installed, rewrites only the role's agent file, and keeps the host's installation receipt, so `hive status` stays `verified`. A part not given keeps the stored override, or the release value when none is stored; the release value is never copied into the override, so an effort-only override follows later model changes. On OpenCode a new effort replaces the profile's `#variant`. A stored override that a later release can no longer render, such as an effort-only OpenCode override for a role whose profile lost its model, stops `update` with an error that names the host and role and the `models reset` command that clears it. `models reset` removes one role's override, one part of it with `--only`, or every override of the host with `--all`; the agent file then matches a fresh installation byte for byte. Both commands follow the `update` confirmation pattern (summary, confirmation in a terminal, `--dry-run`, `--out FILE` for `hive apply --plan FILE`), and a request that changes neither a file nor the stored overrides prints `Nothing to change`.

A group is the folder of the role's source, `content/agents/<group>/<role>.md`: today `design`, `development`, `docs`, `ops`, `quality`, and `review`. `set --group G` writes an override with only the given parts to every role of the group, replacing each role's own override. The summary of every `set` and `reset`, in the command and in the Models view alike, is a title, a Role/Model/Effort table with `a → b` only for the parts that change, `Replaces the own override of <roles>.` for the roles that lose a part, `Writes N files in <dir>` with the home shown as `~`, and the open-sessions note when a file changes. `reset --group G` removes the overrides of the whole group, or one part with `--only`. `--group` and `--role` cannot be combined, and a group the host's installed release does not have is refused. `hive models` prints one header line per group, alphabetically, with the common model and effort, `mixed` per part when the roles differ, and `*` when a role of the group is overridden. On OpenCode, where the effort travels as the model's `#variant`, a new model without a chosen effort keeps the profile's effort: a separate profile effort is applied as before, and a variant embedded in the profile's model (the `inherit` profile's `#max`) is carried over when the model is replaced, so the override stores only the model and still follows later releases.

Every later installation plan for that host applies the overrides again: `hive update`, `hive install`, and a return to a retained release with `plan install --release`. Removing Hive from a host drops its overrides. An override for a role that the installed release no longer has stays stored and is listed by `hive models` as `not applied: role not in the installed release`; `hive models` marks overridden roles with `*`.

An override is refused, writing nothing, when the host is not registered, the role is not in the host's installed release, the host's agents come from more than one release, the reinstall would change anything other than agent files (a file deleted by hand, a pending legacy migration), the model is empty, longer than 200 characters, or uses a character outside letters, digits, and `._:/@+=[]-`, or the effort is not one of `low`, `medium`, `high`, `xhigh`, `max`, and `ultra`. Grok and Cursor accept no effort, Pi does not accept `ultra` (its thinking levels end at `max`), and an effort on OpenCode needs a model to carry it. On OpenCode, where the release's profiles set no model since 2026-10-26, the release's efforts, including a role's, are mapped to the nearest variant the final model offers (the `variants` table in `agent-profiles.json`; `medium` becomes `high` on the DeepSeek and GLM models), and a model-only override is mapped against the new model, with no mapping when that model has no `variants` entry. An override effort that the final model does not offer is refused, naming the model and the variants it offers, because OpenCode would reject the launch with `Variant unavailable`; a model without a `variants` entry accepts any valid effort. `--model ""` and `--effort ""` are refused; `reset --only` removes one part. The command does not check the host's model list, so a model the host does not list is written as given; the Models view offers the host's list (below). Claude Code's acceptance of `ultra` is not verified against its documentation.

A manager from before this feature drops `model_overrides` the next time it writes the state, leaving the overridden agent files in place; run `hive models reset --host H --all` for each host before returning to an older manager.

### Project settings

```sh
go run ./tooling/cli project set [--project DIR] --set 'Key: value'... [--unset Key...] [--dry-run]
```

`project set` creates or edits the `## Hive` section of `AGENTS.md` at the root of the Git repository that contains `--project`, or the current directory. It replaces only the given `- Key:` lines, appends new keys after the last item in the order of the known keys, and keeps every other line, the file's line endings, and its permissions. Without a section it appends one at the end of the file; without `AGENTS.md` it creates the file, mode `0644`, with only the section. `--unset` accepts optional keys only, and is how a retired key is removed. The summary shows the path, the section's lines before and after, warnings, and `AGENTS.md has uncommitted changes` when Git lists the file as changed; a terminal is required to confirm, and without one only `--dry-run` runs.

It refuses, writing nothing: a duplicated section, an `AGENTS.md` that is a symbolic link, not a regular file, or larger than 1 MiB, a directory outside a Git repository, a missing or empty required value, a value with a line break or with characters that would be removed when shown (also an existing value in a key the request does not touch), an unknown key in `--set`, a retired key (`Review`, `UI verification`) or retired `Delivery` value (`automatic`, `interactive`) in `--set`, named as retired and checked before the unknown ones, a `Delivery` other than `direct-base`, a `Hive guidance` other than `required`, a base branch name Git rejects, and a file that changed after the preview. A base branch or `Specs` path that does not exist yet, a `Completion` without the `<state> on <ref> · <state> on <ref>` shape, an unknown key already in the file, and a retired key already in the file (which is preserved unless unset) are warnings. When the root holds a `CLAUDE.md` that does not import `@AGENTS.md`, the summary warns that Claude Code reads `CLAUDE.md` instead of `AGENTS.md` by default; `CLAUDE.md` is never changed. Only that root `CLAUDE.md` is checked, not `.claude/CLAUDE.md`, `CLAUDE.local.md`, or one in a parent directory.

The write is atomic: a temporary file in the same directory renamed over the old one, or linked into place for a new file so a file created meanwhile is never replaced. It does not enter Hive's transaction journal or `hive recover`, since the file belongs to the project; Git is its recovery. An edit made between the second read and the rename can still be lost, and a hard link to the file is detached.

### Walkthrough viewer

```sh
go run ./tooling/cli view register [--dir DIR] [--home DIR] [--state-dir DIR]
go run ./tooling/cli view start [--port N] [--ttl D] [--home DIR] [--state-dir DIR]
go run ./tooling/cli view status [--dir DIR] [--home DIR] [--state-dir DIR]
go run ./tooling/cli view open|stop [--home DIR] [--state-dir DIR]
go run ./tooling/cli view serve [--port N] [--ttl D] [--home DIR] [--state-dir DIR]
```

`flow-walkthrough` publishes JSON for an installed Hive binary and creates standalone HTML only when that executable is absent; the [manifest publication contract](../../../content/skills/flow-walkthrough/references/walkthrough-manifest.md#publication-target) owns the choice. Ingestion needs only JSON and its referenced captures/recordings, not an HTML page. Server or registration failure does not select HTML.

`view` is a local server for the walkthroughs that `flow-walkthrough` writes (`_support/workspace/*/walkthrough.json` and the files it references). The user starts and stops it; agents register sources through `hive view register`, query ingestion through `hive view status --dir`, and run `hive view open` only when the user chooses that option in a walkthrough question, as their skills set. `register` adds the root of `--dir`, or the current directory, to `viewer/registry.json` and writes nothing when it is already there; it needs no running server. That root is `<root>` when the directory, after resolving links, is inside `<root>/_support/workspace/` at any depth, whether or not `<root>` is a Git repository (a multi-repository workspace keeps its repositories as nested folders); otherwise it is the Git root; outside both it fails, naming both cases. A nested Git repository registered from its own folder keeps its own Git root. Projects group by the normalized `origin` remote only when the root is itself a Git toplevel (`git rev-parse --show-toplevel` equals the root); any other root, such as a non-Git workspace inside a versioned home, is keyed by its path and named from the `Project:` value of its own `AGENTS.md`. `start` launches `serve` detached (its own session, output to `viewer/server.log`), waits up to five seconds for its health check, and reports an already running server instead of starting a second one; `serve` runs in the foreground and holds an exclusive lock on `viewer/server.lock` for its whole life, so one state directory never has two servers. `status` prints JSON with the URL, PID, TTL and counts, never the access token; `open` requests a one-use code and opens `/#code=<code>` in the system's opener (`open` on macOS, `xdg-open` on Linux), never passing the server token or starting the server; `stop` signals the server only after its authenticated health check matches `viewer/server.json`.
Each stored copy retains its resolved project key and name in optional
`project_key` and `project_name` fields in `meta.json`. When a previously resolved
remote becomes unavailable, including after a worktree is removed, its copies
keep that identity across scans and restarts. Projects with the same display name
are not merged by name. Older copies can recover their identity from a matching
cached index entry; an identity already lost before the update needs an explicit,
verified association. Stop the viewer before a manual repair and preserve its
copy contents, archive/pin state and ingestion time. Older binaries reject the
new metadata fields: keep a metadata backup before upgrading, and coordinate any
binary rollback with restoration of compatible metadata.

`status --dir <work folder>` preserves the normal status fields and adds `found`
and `session_status`. It queries the existing authenticated index, without opening
another Store. `found=true` requires the canonical source, session ID, key,
content hash (including referenced files), and explicit status to match the
current source, fingerprinted both before and after the query. A missing or
changed source/copy returns `found=false` and `session_status=unknown`; tokens
and file capabilities never enter the output. `session.status` is optional
`in_progress|closed`; absent status remains valid and projects to `unknown`,
independently of report kind or Git classification. Store load reconstructs the
derived `session_status` from the manifest, without migrating historical
metadata. Status changes use the existing content-replacement behavior, including
pin and archive preservation and refreshed ingestion time. For source closure and retention,
follow [Close the current walkthrough](../../../content/skills/flow-close/SKILL.md#close-the-current-walkthrough). Publication can update JSON without creating HTML; confirmation checks JSON and resources in either output mode.

Access setup uses `POST /api/access-codes` with `X-Hive-Token`, explicit same-origin `Origin` and an empty body. It returns `{code}` without caching: 32 random bytes encoded as 64 hex characters, valid for 60 seconds, held only in memory. At most 64 unconsumed codes may be pending; a full set returns 429 without evicting valid codes. `POST /api/access-codes/exchange` requires the allowed Host and explicit same-origin Origin, `application/json`, and only `{code}` within 1 KiB. It returns `{token}` without caching, consuming the code atomically; expired, missing, invalid or reused codes return 403. Invalid JSON returns 400, unsupported media type 415, and wrong method 405. Codes disappear on server restart. The page removes the fragment before awaiting exchange, stores only the token in sessionStorage for reloads, and ignores responses from superseded openings, including earlier authenticated reads and writes. An explicit failed opening asks for `hive view open` again. An old running server returning 404 to issuance requires `hive view stop`, `hive view start`, then `hive view open`; the CLI never falls back to a token URL or restarts it automatically. A stolen unconsumed code can still be used by a local process; Origin is not authentication for non-browser clients.

The server binds only to `127.0.0.1`, on port 4483 (HIVE on a phone keypad) by default. `--port N` names another port and `--port 0` takes a free one. A port counts as busy when binding it fails or when a connection to it already answers, since on macOS a `127.0.0.1` bind succeeds over a program listening on the wildcard address and would shadow it. When 4483 is busy and no `--port` was given, it takes a free port and `start` says so; an explicit port that is busy fails the start. It checks the `Host` header on every request, and requires the token from `viewer/server.json` (mode 0600) on API requests, reads included, except the access-code exchange below; every write requires an explicit same-origin `Origin`. The page and `/assets/*` carry no data and need no token; copied captures and recordings are served at `/files/<key>/<cap>/<path>`, where `<cap>` is the hex HMAC-SHA256 of the key under the token, handed out only by the authenticated index, because an `<img>` cannot send a header. Every five seconds it scans the registered repositories, validates each manifest with the same rules as the renderer, and copies a valid one with its files into `viewer/sessions/<key>/`, one copy per source folder and session, replaced when the manifest or a referenced file changes. Invalid manifests are listed as ingest errors and never stop the server. Copies older than the TTL (14 days by default, `--ttl`) by ingest time are purged at start and every hour unless pinned; a purged or deleted copy leaves a tombstone so the next scan does not bring it back while its content is unchanged. Deleting from the sidebar removes only the store copy, never a repository file. The package is `tooling/viewer`: no external modules (standard library and in-repository packages), checked by the dependency test. Each workspace folder is opened once as an `os.Root` and that same handle validates, hashes and copies it, and the text of ingest errors is bounded.

Before scanning file content, the viewer checks current Git tracking through a
bounded, NUL-delimited lookup for the owning repository, including worktrees and
ancestor repositories. Confirmed tracked manifests are not ingested; an unavailable
or incomplete check produces an explicit unknown-source warning, not a claim of
agent authorship or confirmed untracked status. Classification changes are applied
even when content hashes do not change and are reconstructed without changing
per-copy `meta.json`. Previously stored copies whose source is now tracked move
to a separate explanatory section. Their normal document and file reads return
409; archive/restore, pin, explicit deletion and normal TTL remain available. Unknown sources remain
accessible with a warning. A local file is not proof of agent authorship.

Archive state belongs to each stored copy's `meta.json` as optional
`archived` metadata (absent means false), independently of `pinned` and the
manifest's session status. The authenticated index exposes `archived` for both
normal sessions and excluded tracked-source entries. Token-authenticated,
same-origin `POST /api/sessions/<key>/archive` archives the copy;
`DELETE` on the same route restores it. Both return 204 on success, preserve
pinning and ingestion time, and never edit source files. Archive state survives
store reconstruction, restarts and replacement from a changed source. Archiving
does not exempt an unpinned copy from normal TTL deletion.

The React shell groups normal visible copies by exact project, declared host and
session ID, with visible-session counts, compact title/ticket/segment/date rows
and no ticket level or connector branches. The selected session stacks its
chronological segments with separators, ticket search jumps and per-copy controls.
This is a derived client view: physical keys, capabilities, store metadata and API
routes remain per copy. Filtering archived copies precedes visual grouping. Projects open
initially only when they contain an explicitly in-progress session; manual
expansion takes precedence. A tooltip-labeled, accessible icon beside Projects
toggles archived copies, hidden by default. The sidebar brand header shows a
route icon. Each segment has a toolbar with informational manifest status,
archive/restore, pin/unpin and delete actions targeting only its own physical key.
The session header aggregates explicit delivery links from its visible accessible
segment manifests, deduplicated by URL; documents and standalone pages retain their
own links. The optional `delivery` contract lives in the manifest reference,
separately from `candidate`; no provider lookup or historical backfill runs.
Both readers validate it and link rendering checks its destination again.
An old strict reader can reject that field, so update binary before content and
follow the maintainer viewer restart rule in the repository `AGENTS.md` when refreshing this checkout; outside that case, obtain explicit permission to restart an already running viewer. Sidebar visibility, project expansion, archive
filter and selection persist in sessionStorage for the same tab and origin.

Manifest and resource ingestion rejects files with more than one hard link using
metadata from each opened descriptor before reading, including hashing and copying.
This does not impose a same-owner rule or authenticate the origin of an inode.
Existing path, symlink, byte-limit and hash checks still apply.

The viewer embeds its assets: replacing the manager binary does not replace a
running viewer process. The updated behavior and assets activate at its next start;
refreshing content does not rewrite previously generated standalone HTML pages.

### Terminal interface

```sh
go run ./tooling/cli
go run ./tooling/cli tui [--home DIR] [--state-dir DIR] [--source DIR]
```

`hive` without arguments in a terminal, or `hive tui`, opens a full-screen application on the terminal's alternate screen: a menu with CLIs, Update, Releases, Voice, Diagnostics, Models, Integrations, Project, and Quit, under a status line with the number of registered hosts, the installed release, and the voice (when Hive's state cannot be read it says `Status unavailable: state could not be read; see Diagnostics.` and never prints the raw error, which Diagnostics shows), and above a help bar that lists the current view's keys. Changing views redraws the whole screen, and leaving the application restores the terminal's previous content. Below 80×24 it shows the minimum size instead of the view, and keeps the view's state until the terminal is large enough again. The theme follows the terminal's light or dark background; `NO_COLOR` removes colors, and checkboxes, the cursor, and the selected button always use symbols.

The application needs a terminal on standard input and output. Without one, `hive` without arguments keeps its usage error, and `hive tui` fails naming the text commands. Those commands (`hive status`, `install`, `update`, `releases`, `voice`, `doctor`, `models`, `project`, `view`, `plan`/`apply` to remove hosts, and `recover`) are the path for scripts and screen readers; the application has no line-based mode. `hive install` remains the plain-text installer that `install.sh` uses.

Keys:

- **Esc** returns to the previous view, and closes the application at the menu. **Backspace** does the same when no text field has the focus; in a text field it deletes a character, and at the menu it does nothing.
- **Ctrl-C** closes the application with exit status 0 from any view. While a change is being applied, every key, Ctrl-C included, and an external interrupt (SIGINT) are ignored until it finishes.
- **↑↓** move, **←→** change a value or choose a button, **Enter** accepts, and **Space** toggles a checkbox.

Each view reuses the plans and summaries of its command, so it writes the same files and state:

- **CLIs** lists each detected, registered, or legacy host with a checkbox (checked means active), its release, the Hive version (the column header shortens to `Hive` when long names leave no room; the CLI's own version is in Diagnostics), and drift. Space changes the desired set and `a` applies it. Removals are previewed, confirmed, and applied first; additions are then planned on the resulting state, with the package verification, required-host notice, optional capabilities, summary, and confirmation of `hive install`. When both happen, the summaries are labeled step 1 and step 2, and stopping at step 2 keeps the removal. When a row reads `editor only`, one line under the table, set apart by a blank line, explains it: `editor only: the Cursor editor is installed, but not its CLI (cursor-agent). Hive can still install Cursor's files.`; without such a row the line is not shown. Checking a host with a legacy installation migrates it. Under a synthetic `--home`, where no executable is ever detected, the view lists every supported CLI and shows the unregistered ones as `not detected`, so a removed host can be installed again from the view. When the scan for legacy installations fails, the view still lists every host with its release and drift, and shows a note under the rows, or under `No CLI hosts were detected or registered` when there are none. For a file that differs from what Hive expects at a path it once used, which may be a file Hive installed or the user's own, the note reads `<path> differs from what Hive expects there. Undo the change, restore it from a backup, or move your own file elsewhere before installing or updating. Removing may also be refused if Hive installed that file.`. The scan reports that case as a `*legacy.ModifiedFileError` carrying the path, which the view finds with `errors.As`; the text installer, `hive plan install` and `hive update` refuse with its message, `<path> differs from what Hive expects there; undo the change, restore it from a backup, or move your own file elsewhere`, and write nothing. It does not point to Diagnostics, which lists only files Hive currently manages. Any other scan failure reads `Hive could not check for a legacy installation, so installing or updating may be refused. The Detail line says why.` For any other scan failure, the raw error goes on a `Detail:` line that gives way to a result or error shown below it; a changed file needs none, since its error says the same as the note. While that lasts, a host with a legacy installation may not be labeled. For a file at a legacy path, installing and updating are refused with the command's own message, writing nothing, until the file is restored (the view starts the refusal it shows with `Nothing was changed. `); removing is refused, with its own message, only when the file is one Hive currently manages. `u` opens Uninstall all, which removes every registered host; its confirmation starts on Cancel and does not accept `y`.
- **Update** edits the source (`.`) and revision (`HEAD`) in place and shows the `hive update` summary. When the revision changes nothing, it records the source commit without asking, as `hive update` does.
- **Releases** lists retained releases, newest first, with the installed one marked. `/` filters by text and Esc clears the filter. Choosing a release returns to it through `plan install --release` for the registered hosts.
- **Voice** shows the active voice as rows (voice, address, name when the address is `name`, and intensity) changed in place with the arrows, then previews `hive voice set` or `off`. With Off, only the voice row shows.

The four diagnostic views show what `hive doctor` and `hive models` print (see [Read-only diagnostics](#read-only-diagnostics)); Models and Project also edit, and change nothing until an edit is confirmed. Each loads in the background with a spinner, `r` reloads it, and a load error shows inside the view with `r to retry`. Long content scrolls with PgUp and PgDn, and with ↑↓ in Diagnostics and Project:

- **Diagnostics** shows the CLIs, Installation, and Sessions sections.
- **Models** shows one host at a time, chosen with ←→ and shown in brackets, as a table of role, profile, model, and effort grouped by the role's group: a header row per group shows the common model and effort, or `mixed`, and `*` when a role in it is overridden. A `>` cursor moved by ↑↓ visits headers and roles; the role name is never cut and an overridden role is marked `*`. Group headers are bold and start with `▾`, their roles are indented under them, and a blank row separates groups. Enter opens a panel above the table for the selected role, or for the whole group on a header (`<Group> group on <host> (N roles)`), with Model and Effort. Enter in the panel previews the matching `hive models set` or `reset` (`--role` or `--group`), sending only what differs from the shown values; a `mixed` part is not sent, `release default` removes that part, and choosing `release default` for both parts of a group previews the group's reset. Pressing `x` on a role with no override answers `Nothing to reset: <role> has no override` in the position row, and on a group header with no overridden role `Nothing to reset: <Group> group has no override`. On Codex, choosing a model that does not offer the chosen effort changes the effort to `release default` and says `Effort <effort> dropped (not offered by <model>); using release default` in the panel's message row. On OpenCode a new model keeps the profile's effort, as with `models set`. `x` previews resetting the selected overridden role or group. Esc closes the list, then the panel, then the view. The footer says `Enter edits the selected role or group; x resets one marked *.` and `Defaults come from integrations/agent-profiles.json in the release.`

  The Model value ends in `▸`, and while it has the focus a muted `→ choose` says how to open the list. → or a letter opens the `Select model` box over the panel and table: a bordered box two columns narrower than the screen, with `Select model · <host> · <role or group>` and `esc` in its top border, the cursor's place among the models (`n of N`) before the key hints on its bottom row while the cursor is on a model, a `Search` line that filters by name, id, or provider, ignoring case, a status line while the list loads or when it failed (`Model list unavailable: <reason>`), and the list. Claude Code, Codex, Grok, and Cursor show one flat list; OpenCode and Pi show one section per provider, from the id's `<provider>/` prefix, with readable names for known providers (`GitHub Copilot`, `OpenAI`, `OpenCode Go`, `xAI`…) and the raw prefix otherwise. A row shows the host's readable name when it gives one (Codex `display_name`, Cursor's `id - Name`) with the id muted, and under a provider the id without its prefix; the stored value is always the full id. The current model is listed even when the host does not list it, marked `●`, and the cursor starts on it. The selected row is highlighted, or marked `>` with `NO_COLOR`. The last line holds `release default` and `Other…`, reached with ↓ past the last model and ←→ between them; `Other…` turns Model into a free-text field validated like `models set --model`. Enter chooses and returns to the panel. On Codex, choosing a model narrows the efforts to that model's `supported_reasoning_levels`; an effort it lacks falls back to `release default`. `Recent`, favorites, and price tags are not offered, since no host reports them to Hive.

  The list comes from the host's own command the first time a panel of that host opens in the session, and is kept for the session; a failure is retried on the next open. The commands are fixed: `codex debug models` (models with `visibility` `list`), `opencode models` (retried once when it prints nothing, which was observed with OpenCode 2.0.21), `pi --list-models`, `grok models` (its session line is dropped), and `cursor-agent models`; Claude Code has no listing command, so its list is the aliases `fable`, `opus`, `sonnet`, and `haiku`. Each runs as the binary found on `PATH`, without a shell, with the user's environment (the host needs it to authenticate), in `os.TempDir()` so it never loads a repository's configuration or plugins, with stdin from `/dev/null`, in its own session so it cannot take the terminal, with an 8-second limit after which its whole process group is killed, and with standard output capped at 8 MiB, beyond which the list is an error rather than a partial list; standard error is discarded. Ids outside the characters `models set` accepts are not offered. With `--home` no host command runs and the list offers only `release default`, the effective model, and `Other…`. `hive models`, `hive doctor`, and browsing the Models view without opening a panel never run them. The hosts' own caches, telemetry, and network calls are outside Hive's control, and the hosts' internal model caches (`~/.codex/models_cache.json`, `~/.cache/opencode/models.json`) are not read. The text command still accepts any valid id, since a host can accept ids it does not list.
- **Integrations** lists five to eight integrations (Pi codemode with Pi registered, the Codex question card and the Codex wait floor with Codex registered) with a `>` cursor moved by ↑↓, and the selected one's detail below.
- **Project** shows the checked `AGENTS.md` path and `Valid` or the findings, with the values read. Inside a repository, `e` opens a form with the eight known keys, required ones first (`Project`, `Base branch`, `Tracker`, `Specs`, `Environments`, `Delivery`, `Completion`, `Hive guidance`): each holds the section's current value or, for a missing required text value, a suggestion marked `suggested` (the repository name and `GitHub Issues · <owner>/<repo>` from a `github.com` `origin`, read without its credentials; the branch `origin/HEAD` points to; `_support/openspec` when it exists). `Delivery` (none or `direct-base`) and `Hive guidance` are chosen with ←→; `Completion` is text, with a placeholder deduced from the `Environments` value. A legacy `Delivery` value shows marked `(retired)` and can be moved to none. Each retired key the section holds adds a row after the eight, choosing between its value marked `(retired)` and `remove`, which unsets it. Enter with an empty required value names it; otherwise it previews `hive project set` with the changed keys. Cancelling keeps what was typed, a write error shows in the form, and a successful write closes the form and checks the section again. While the root `CLAUDE.md` does not import `@AGENTS.md`, the view keeps that warning under the result; `hive doctor` does not report it, since it checks only the section. Esc closes the form only.

In the scrolling views (Diagnostics, Project, the Integrations detail) a line that does not fit continues under its own indent plus two spaces. A path or other token wider than the line breaks after a `/`, keeping the slash at the end of the line, and breaks hard only when a segment has no slash and still does not fit; a token that fits moves whole to the next line, never splitting at a hyphen. The notes and errors of the other views, and the scrolling summaries and error dialogs, wrap the same way.

When Hive's state cannot be read, CLIs and Voice, like Models, lead with `Hive's state in <dir> could not be read. Repair or restore its files; hive doctor shows the same problem.`, then `Press r to retry after fixing it.` (the `r` key reloads the view and is listed in its help bar), and put the raw error on its own `Detail:` line after it; Voice keeps the plain guidance for a source without voices, which is not a state problem.

Apart from that unchanged update, every writing action shows its summary, which scrolls, and asks for confirmation. Declining or going back shows `Cancelled. No changes applied.` inside the originating view; an error shows the command's message there; a successful action refreshes the view and shows its result. When the application opens with a pending operation, it offers to recover it. Recovery messages name `hive recover`, adding `--state-dir DIR` when the application was opened with an explicit state directory. Adding hosts in CLIs, and Voice, read the catalog from `--source`, which defaults to the current directory: outside a Hive checkout or package they report `Run hive from a Hive checkout or package, or pass --source`.

## Optional Context7 setup recommendation

`hive setup` is a read-only onboarding step. It checks known skill locations for nonempty `find-docs/SKILL.md` and `context7-mcp/SKILL.md` files, including the shared `.agents` location and configured host homes. `--home DIR` selects a synthetic home and ignores environment overrides. A user-scope install plan without an existing Hive state file prints a first-setup recommendation; it never blocks installation on Context7 availability.

Context7 is strongly recommended for development workflows that need current library documentation, but is optional. For these terminal-based harnesses, CLI + Skills is the recommended default because it reuses shell execution and native skill discovery without per-host MCP registration. MCP remains valid when native structured tools or an environment without shell access are preferable; existing setups are preserved. A discovered file does not prove host loading, authentication, service availability, or freshness. MCP-only and custom installations may not be detected. Hive never reads credentials, launches authentication, or installs/updates Context7 from this check. Official documentation distinguishes unauthenticated `ctx7 library`/`docs` queries (lower limits) from the hosted setup wizard, which requires authentication. Users without service access should consult current official documentation directly and disclose verification limits.

The suggested voluntary install/refresh command is `npx ctx7@latest setup --cli`. Node.js/npm and network access are needed only if the user chooses that vendor command. The official setup owns agent selection, authentication, its `find-docs` skill, and its instruction rules. Existing MCP users can keep their mode. Do not vendor its content, silently migrate modes, or treat third-party files as Hive-owned removal targets. Selecting `@latest` resolves the current CLI when invoked; rerunning setup downloads the vendor skill and updates its rules. This is not a background-update guarantee or a read-only version check.

Verified against installed `ctx7` 0.5.11 and [official CLI documentation](https://context7.com/docs/clients/cli). Its setup supports Claude, Codex, and OpenCode flags but has no Pi/Grok flags; shared skill-file discovery does not establish runtime support for those hosts. The vendor setup may request authentication and rewrites selected vendor skill/rule files even when they already exist. Hive's setup check does neither.

## Content and native destinations

### Optional OpenCode context loader

Editorial template: `integrations/opencode/context-loader.js`. When present in
the release, the manager generates the installed script and metadata only for
OpenCode user scope at
`<OpenCodeHome>/hive/context-loader/index.js` and
`<OpenCodeHome>/hive/context-loader/context-loader.json`, with no `package.json`, not in the
auto-discovered `plugins/` directory. No other host or project scope gets those
destinations. `integrations/opencode/context_snapshot.go` derives both files from
the validated release payload; the canonical references keep their single
editorial home under their owning skills. No second authored JSON manifest or
live-file fallback is maintained.

The template contains exactly one `__HIVE_CONTEXT_SNAPSHOT_DATA__` marker. The
renderer rejects zero or multiple occurrences before writes and substitutes once
with JSON data serialized by `encoding/json`, not Markdown concatenated as code.
The embedded object has `schema: 2` and `references`, ordered by role:
`hive-review-harness`, then `hive-review-plan`. Each entry contains `role`,
`resource`, `text` and lowercase hexadecimal `sha256`. Text preserves the exact
valid UTF-8 resource bytes in the release, including line endings and BOM; its
digest covers those same bytes. Missing, empty or invalid UTF-8 canonical owners
or references reject generation. The accompanying metadata has `schema: 2` and
the same ordered `role`, `resource`, `sha256` list, without text, a reading root
or `skillsDirectory`. It is derived provenance, not a runtime reading capability.

These files use managed `plugin` ownership, exact bytes/modes, conflict checks
and the existing transaction/recovery machinery. A missing managed file is
reinstalled; a changed or unowned destination is preserved as a conflict.
Removal or a release lacking the loader retires its owned resources without
editing native user configuration. Empty directories created by the manager are
pruned; a preexisting directory or a directory containing unrelated files is
preserved. Unpublished candidate names are not accepted as managed resources.
Saved-plan validation reconstructs both
derived files from the payload and compares their complete bytes before writes,
including when plan hashes have been recalculated. Digests establish exact
canonical-byte provenance, not cryptographic authentication of local executable
code. A party able to replace the plugin is outside the isolation Hive promises;
the host loads local executable code with its own access, not a Hive sandbox.

Normal install/update never activates the plugin or writes `opencode.json(c)`.
After explicitly choosing activation, the user adds the absolute
`<OpenCodeHome>/hive/context-loader` directory path
to the native `plugins` array in their own OpenCode configuration, preserving
existing entries, and starts a new session. The native plugin ID is
`hive.context-loader`. Remove that entry or use OpenCode's native
`-hive.context-loader` directive to deactivate, then reload through the native host. Before uninstalling Hive or
returning to content without the loader, remove the path entry: the manager
preserves it, and OpenCode can otherwise report a missing plugin. Disabling the
plugin does not remove the core's obligation to read required references. Do not
activate the repository's editorial template: it still contains the unsubstituted
snapshot identifier and is not the generated runnable plugin.

Runtime observation on OpenCode 2.0.27 (2026-10-12): a `plugins` entry
pointing to a single `.js` file is rejected with "configured plugin path must be
a directory"; a directory containing `index.js` loads without `package.json`.
Files under the native `plugins/` directory are auto-discovered, so the inactive
copy stays under `hive/context-loader/`. This establishes loading only, not
context injection before dispatch or model effectiveness.

The implementation uses OpenCode V2 `Plugin.define({id, setup})`, a session
`context` hook and `ctx.agent.get({agentID})`. It resolves the role from the
native identifier, `data.id`, then `id`, then the hook event's agent value (also
when the lookup fails), never from the display name. It maps only
`hive-review-plan` to `flow-plan/references/plan-review.md` and
`hive-review-harness` to
`harness-audit/SKILL.md`. Each request receives the corresponding embedded
snapshot again; deduplication touches only parts previously added by this plugin
to that request. The callback performs zero reference-file or metadata reads and
does not resolve reference pathnames. An unavailable snapshot adds the bounded
`missing_snapshot` diagnostic, not an asserted abort or enforcement guarantee.
It changes no history, models, tools, permissions or generation settings. Later
plugins may change context.

Freshness follows delivery, not live filesystem edits: regenerate the plugin with
the manager from the desired validated release payload, then reload the plugin
through the native host before relying on the new snapshot. Editing an installed
Markdown reference alone leaves its embedded snapshot unchanged and can produce
managed drift; the manager preserves detected edits rather than overwriting them.
There is no parallel live-reading loader.

API documentary basis: [OpenCode V2 plugins](https://opencode.ai/v2/docs/build/plugins),
consulted for the change on 2026-10-11 with installed OpenCode 2.0.26 reported in
the plan. Live documentation is not pinned to that binary. Source inspection
and callback tests are not native discovery, final request or behavioral proof.
The bounded model pilot has not yet run and remains blocked on verifying safe,
exact memory cleanup. The snapshot consumer's source/callback checks do not
resolve that prerequisite or prove model benefit; do not infer success from
packaging or private-directory permissions.

New offline packages use `release.json` manifest version `2` and must include
the exact source path `integrations/opencode/context-loader.js`, even when the
user does not select OpenCode. `tooling/package/main.go` copies it into the
package. The current verifier accepts versions `1` and `2`; version `1` keeps
its old required-file set. Old source/snapshot payloads without the loader remain
usable and do not acquire it automatically. This is separate from the state and
product version schemas, and independent of the loader's metadata schema `2`.
Managers whose verifier accepts only manifest `1`
reject new packages: use the manager shipped with the v2 package, or rebuild the
manager before using newer content. Do not assume an old binary handles the new
`plugin` resource kind merely because it can read an old manifest. Returning to
an old *content release with the current manager* is not the same as downgrading
the manager binary; use the current manager to retire the loader resources first.

The catalogue contains `content/guidance/global.md`, each `content/skills/<skill>/SKILL.md`, resources recursively below each skill's `references/`, `scripts/`, and `assets/` directories, and canonical roles at `content/agents/<category>/<name>.md`. Skill resources may be `.md`, `.html`, `.py`, `.sh`, `.json`, `.ts`, `.mjs`, `.astro`, `.mdx`, `.css`, `.yaml`, `.yml`, or `.gitignore`; their source modes are frozen and restored, including executable scripts. Agent sources and `integrations/agent-profiles.json` are validated as one catalogue: role names must be unique across categories. Links, duplicate paths, escaping paths, empty payloads, and reserved block delimiters are rejected. Disposable `__pycache__`, `node_modules`, `dist`, and `.astro` directories beneath skill resources are excluded.

Each release ID is a SHA-256 over its ordered payload manifest, source modes, frozen agent-profile input, and renderer version. Resource identity is its source path (`Target.Source`), so a resource cannot be substituted for another payload. Agent files are rendered in memory for each native host and their exact rendered bytes are frozen into the saved plan and snapshot; the profile file itself is never deployed. Rendering also replaces each inline `skill:owner/path` link outside code fences, in roles and in the global block, with the resource path in the host's skills directory: absolute `<home>/.agents/skills` at user scope, project-root-relative at project scope (`integrations/mdlinks`; `SkillsDir` in each host package). A file shared by hosts must render identically or the plan fails before writing. `agents.Version` is unchanged, so older releases still validate, and a plan saved by an older binary fails the recomputation check in `LoadPlan`. Release sources keep `skill:` locators. No generated distribution directory is required for local use. In the paths below, `<relative-file>` is a skill entrypoint or resource.

| Host | User scope | Project scope |
|---|---|---|
| Codex instructions | Effective `AGENTS.override.md` or `AGENTS.md` under its configured home; an empty override does not win | The same precedence at the selected project root |
| Codex skill | `<user-home>/.agents/skills/<skill>/<relative-file>` | `<root>/.agents/skills/<skill>/<relative-file>` |
| Claude instructions | `<claude-config-dir>/CLAUDE.md` | `<root>/CLAUDE.md` |
| Claude skill | Directory alias `<claude-config-dir>/skills/<skill>` to the shared `.agents` skill | `<root>/.claude/skills/<skill>/<relative-file>` |

The user defaults are documented by [Codex instructions](https://learn.chatgpt.com/docs/agent-configuration/agents-md), [Codex skills](https://learn.chatgpt.com/docs/build-skills), [Claude memory](https://code.claude.com/docs/en/memory), and [Claude skills](https://code.claude.com/docs/en/skills). Observed host versions for this implementation are Codex 0.155.1, Claude Code 2.1.278, Grok Build 1.0.34, Pi 0.86.0, OpenCode 2.0.9, and Cursor CLI 2026.09.18-9a7762b. Documentation does not establish successful loading in every environment. The five-host rollout report (historical evidence omitted from public history) separates installed state, observed loading, and task outcomes.

A physical resource records its registered consumers (`host`, `scope`, `context`). Identical selected destinations share a single resource. Updating shared bytes requires selecting every registered consumer. Removing one consumer retains resources still used by another; `retained_shared` means the native host may still discover that resource. Registration is not a host-level visibility switch.

User scope also supports Grok, Pi, OpenCode, and Cursor. Grok uses the native `~/.claude/CLAUDE.md` compatibility path and the shared `.agents` skill; disabled or ambiguous Claude instruction compatibility is a conflict. Pi uses `PI_CODING_AGENT_DIR` or `~/.pi/agent/AGENTS.md`. OpenCode uses `XDG_CONFIG_HOME/opencode/AGENTS.md` or `~/.config/opencode/AGENTS.md`. Cursor uses `~/.cursor/AGENTS.md` and renders roles to `~/.cursor/agents/`. All four use the shared `.agents` skill and reject project scope.

Cursor CLI does not load `~/.cursor/AGENTS.md` natively: 2026.09.18-9a7762b loaded no global instruction file, and its only documented global channel is User Rules in the app settings. The file reaches Cursor sessions only through the conditional pointer each repository adds to its own `AGENTS.md`; the canonical line and when to require it are in `harness-audit`'s instruction-file rule HA-IF-18. Cursor discovers the shared `.agents` skills natively. Hive only creates and removes `AGENTS.md` and `agents/<role>.md` inside `~/.cursor/`, which belongs to Cursor. Grok's `GROK_HOME` controls its configuration lookup, not the fixed Claude compatibility path.

Each user-scoped Claude skill has one managed directory alias, shared by its entrypoint and references. Only these managed Claude skill-directory aliases are supported. Each alias's exact relative link text is journaled and checked; unknown links, retargeted links, unsafe ancestors, changed override resolution, unowned skill directories, or malformed/unknown Hive blocks conflict. Payload creation precedes alias creation; alias removal precedes final payload removal. Whole skills roots and instruction files are never linked. The installed payload is a versioned copy, not a link to this checkout.

The manager does not follow arbitrary import graphs, adopt unrelated deployments, or detect every unregistered CLI consuming a conventional shared path. A successful `status` reports filesystem state, not runtime discovery or loading.

### Grok read-whole hook

Grok models often read a skill reference with a line `limit` and miss its end:
from 2026-10-01 to 2026-10-13, 29 of 220 per-session reference reads were left
incomplete in 22 of 65 Grok sessions, against 7 of 248 partial reads in Claude
Code. Grok 1.0.55's `read_file` ignores `offset` and `limit` only for `SKILL.md`,
`AGENTS.md` and `CLAUDE.md`, and its own skill guide tells authors to have
references read with them.

For Grok user scope the manager generates
`<GrokHome>/hooks/hive-read-whole-references.json`, a global `PreToolUse` hook
with matcher `read_file` whose command is
`'<hive>' grok-hook read-whole --skills '<SkillsDir>'`: the absolute path of the
`hive` binary that planned the install, and the shared `.agents/skills`
directory, both single-quoted for the `sh -c` Grok runs it with. The hidden
`grok-hook read-whole` subcommand reads the hook event from stdin and, for a
`read_file` with `offset` or `limit` on a Markdown file inside that directory of
at most 1000 lines and 64 KiB, replies with an `updatedInput` that drops both
fields. It prints nothing and exits 0 for every other call and on any error,
including a panic, so Grok continues with the read the model asked for; it never
denies a call. Scripts and assets keep their partial reads.

The record uses kind `hook` with the generated source `integrations/grok/read-whole-hook`,
accepted only for that file name and the `grok` user-scope consumer, like the
OpenCode loader's `plugin` records; a release may not list that source as a
file. Ownership follows agent files: mode 0644, an unowned file at the path is a
conflict, an edited managed file is refused, removal deletes it, and the other
files in `hooks/`, such as the user's own hooks, are untouched. The binary path is
resolved when the plan is built and saved in the plan's `executable` field, so
`apply` regenerates the same bytes; `apply` refuses a plan whose executable is not
an absolute, clean regular file reached without a symbolic link, and a plan that
carries an executable without installing the hook. Only user-scope installs that
include Grok resolve it. A plan that would install the hook from a binary under
the system temporary directory, `GOTMPDIR`, `GOCACHE` or the Go build cache, such
as `go run ./tooling/cli`, is refused: build the binary with `go build` first. The pilot
runner, which uses `go run`, therefore cannot install Grok.

The hook depends on that binary path. Reinstalling or updating from a binary at
another path rewrites the command; deleting the package that installed it leaves
the hook failing open, which Grok shows as a failed hook on each read, until the
next install or update. `hive models` on Grok refuses after the binary moved,
since the reinstall would change the hook, until `hive update` runs. A manager
older than this change rejects state that holds a `hook` record, so rebuild the
binary before deploying. Grok loads hooks when a session starts: restart open
sessions, or reload with `r` in `/hooks`, after installing or removing it.
Subagent sessions run it too. With `toolset.file_toolset = hashline` the read tool
is `hashline_read`, which the matcher does not cover. Sources: Grok 1.0.55
`~/.grok/docs/user-guide/10-hooks.md` and `08-skills.md`, read on 2026-10-13.

## Preservation and state

Only the managed global block is replaced. Existing bytes outside it and the file's permissions survive. The first insertion records any added separator and whether Hive created the file. Removal deletes only that managed span; a newly created file is removed only if no user content remains. Every skill resource and agent file that still exists must match its installed bytes and recorded mode before update or removal; an edited one is refused. Managed content the user deleted by hand (a skill, agent or link, a managed block, or the file that held it) is treated as absent: `install` and `update` write it again as a fresh installation would, and removal completes without writing and drops its record. Removing one host never recreates a deleted shared resource that other hosts still use; their next `install` or `update` does. Plans saved with `--out` and the journal record this with an optional `gone` field, which a manager older than this change rejects. New skills and agents cannot adopt existing files. An already-owned skill can gain absent resources; an unowned file at a planned destination is a conflict. Additional user files outside managed destinations remain untouched.

Installing a release reconciles resources absent from its catalogue, including references removed by a downgrade. It removes only the selected consumers; shared files remain until their last registered consumer is removed. `status` and `plan remove` derive their catalogue from installed state rather than requiring the original source checkout.

Plans contain managed content, fingerprints, and ownership metadata, not the surrounding private instruction text. Preview output shows the old and new managed content rather than a whole-file diff. Backups necessarily contain original target bytes and must stay private.

The macOS default state home is `~/Library/Application Support/tricell-hive`.
Linux uses `$XDG_STATE_HOME/tricell-hive` or `~/.local/state/tricell-hive`, preserving
the historical location when it is the only existing state home. Conflicting locations
require explicit selection. The state home contains:

- `state.json`: schema version, installed resource records, release IDs, owned directories, and scoped legacy-migration receipts.
- `releases/<hash>.json`: immutable installed payload snapshots.
- `releases/<hash>.commits.json`: the source commits recorded by `hive update` for that release, with the time each was applied.
- `voice` and `voice_spans` inside `state.json`: the chosen voice and, per instruction file, the written voice span, its source hash, and its consumers.
- `transactions/<id>.json`: recovery journals with private before/after images and the transaction phase; only the 10 most recent finished ones are kept.
- `pending.json`: an unfinished operation requiring recovery.
- `lock`: process lock, released by the operating system on exit.
- `viewer/`: the walkthrough viewer's registry, server record and log, lock files, index, and copied walkthroughs (`sessions/<key>/`); it is outside `state.json` and its transactions, and `hive view` alone writes and purges it.

New state directories use mode 0700; state files, plans, and backups use 0600. Unknown schemas or malformed state fail rather than trigger automatic rebuilding. Release history is retained and never pruned, and the manager stores no credentials. After each operation commits, the manager keeps the 10 most recent transaction journals in the `committed` or `recovered` phase and deletes older ones; it never deletes the journal `pending.json` names, the core journal a pending onboarding references, a journal in another phase, or one whose leading `Version`, `ID` and `Phase` keys it does not recognize, and it deletes none when it cannot read those pending files. A pruning failure is a warning on the result, never a failed operation. No command restores a finished transaction; `recover` only completes or undoes the pending one. State, operation plans, and new core transaction journals use schema v6. The product-version receipts and version index are separate from historical content release IDs; v6 preserves their hashing. Optional onboarding uses a separate versioned parent journal and does not roll back a committed core on provider failure. It adds verified legacy retirement and scoped migration receipts to v4's source payload modes, renderer/profile identity, and shared consumer ownership. Existing v1-v5 state migrates transactionally when an operation is applied; planning alone does not rewrite it. Retained legacy releases preserve their original payload hashes. Pending v1-v5 transactions remain recoverable with their original checksum semantics. Saved plans before v6 must be regenerated before applying new operations. Plans saved before Cursor support was added also fail to apply (`invalid root`) and must be regenerated.

## Failure handling

The manager validates the whole plan before destination writes, records recovery data before mutation, and rechecks state, resolution, and file fingerprints. Replacements are atomic per file, not across all files. The lock coordinates Hive processes; it cannot lock unrelated editors. The manager checks for competing edits and preserves detected conflicts, but does not claim protection against every malicious filesystem race.

An interrupted operation blocks further plans/applications until `recover`. Recovery preflights all inverse operations. It can preserve newer text outside an installed block while restoring the previous managed span; changes inside the managed span cause a conflict. It never restores a whole old configuration over detected later edits. Recovery of an already committed journal finalizes bookkeeping instead of undoing the installation.

Only directories confirmed created by Hive are cleanup candidates, and only empty directories are removed. A crash between creating a directory and recording ownership may leave an empty directory behind; preserving uncertain ownership is preferable to claiming it retroactively. Directory ownership is recorded by path, not inode: replacing a previously created directory with an empty directory at the same path is not distinguishable during cleanup. No `--force` or blanket purge exists.

## Verification

The repository has no CI: verification runs locally, once on the final candidate of a change.

```sh
go vet ./...
go test ./...
python3 -m unittest discover -s tests/skills -p '*_test.py'
```

Run `go test -race ./...` as well when a change touches concurrency: goroutines, locks, state shared across goroutines, or the manager's concurrent-write tests.

Test binaries of `tooling/management` and `tooling/cli` call `management.DisableDiskSyncForTests` from `TestMain`, because fsync dominates their run time on macOS; the `hive` binary always syncs. The same two `TestMain` functions also clear `test.testlogfile`, which disables Go's test-result cache for these packages on purpose: their tests Lstat path ancestors some 700k times, and the go command re-verifies that log line by line on every cached run (112–300 s and several GB without it, about 20 s with it), so running them each time is cheaper than the cache.

Tests exercise the six user-scope mappings and the two supported project mappings with synthetic homes and projects, preserved LF/CRLF content and permissions, snapshots/downgrades, idempotence, conflicts, stale plans, links, concurrent directory creation, and recovery at each write boundary. Catalogue tests additionally cover multiple skills and nested Markdown references, source identity and forged payload rejection, new references in owned bundles, unowned reference conflicts, partial consumer retirement, old-release rollback, missing-checkout status/removal, v2 migration and legacy journal recovery. The pilot protocol and its limitations are in the [workspace screening fixtures](../../../tests/fixtures/workspace-conventions/README.md).

## Product versions and optional operations

The product version is independent of the state/manifest schemas and Git revision.
`VERSION` is the source-tree label; `hive --version` describes the running manager.
Per-consumer receipts describe the requested artifact and its expected resource
bytes/modes/aliases. Status distinguishes that request from current conformity.
Shared resources may retain their original provenance without implying drift.
A legacy snapshot with no product receipt retains its hash and is not relabeled.

Published version labels cannot be rebound to a different artifact in the local
version index. Development builds use `dev`: between releases, `VERSION` on
`development` reads `dev`, so `hive update` can deploy each merge, and each
release commit sets its number. This local check does not certify a
remote release channel or authorize public distribution. Source-code visibility
and package publication remain independent decisions.

Optional steps are explicit and journaled before execution. A provider execution
with an unknown outcome blocks further mutation until reconciliation. Recovery
inspects that outcome without rerunning an installer or authentication. A committed
core remains installed if an optional provider fails. Status is still readable
while reconciliation is pending. Existing tools, credentials and memory are not
owned merely because the onboarding catalog detected them.
