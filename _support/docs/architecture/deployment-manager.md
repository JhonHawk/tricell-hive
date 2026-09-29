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
go run ./tooling/cli recover --state-dir /absolute/state-directory
```

These are interface examples, not authorization to deploy into the user's real configuration. `plan` without `--out` is a read-only preview. With `--out`, it writes a new private plan file, never the destination files. `apply` consumes that plan and rejects stale target fingerprints or ownership state. Destination options are not accepted by `apply`: review the destinations bound into the plan instead.

`--home` explicitly selects a synthetic home and ignores host path and compatibility environment overrides. Use it together with `--state-dir` for temporary-home tests. Without `--home`, host configuration environment variables are respected. `--root` is required for project scope. Existing root prefixes are canonicalized; linked target files or ancestors inside those roots are rejected rather than overwritten.

The default source is the current directory; `--source` selects another checkout with the same content layout. `plan install --release <hash>` selects a retained installed snapshot instead, allowing an explicitly planned downgrade. The current binary validates that snapshot's agent sources, so a release whose roles use a retired frontmatter field, such as `claude_effort` before it became `effort`, fails with `unsupported agent field`; plan that downgrade with the manager from the commit that produced the release. Installed releases still update, report status, uninstall, and recover, because those operations do not re-render the old sources. A plan freezes its source bytes; later source edits do not silently alter it.

### Update from a commit and list releases

```sh
go run ./tooling/cli update [--rev HEAD] [--source .] [--home DIR] [--state-dir DIR] [--dry-run] [--out FILE]
go run ./tooling/cli releases [--home DIR] [--state-dir DIR]
```

`hive update` deploys the content of one commit, never uncommitted edits in the checkout. It resolves `--rev` (default `HEAD`) with `git rev-parse` in `--source` (default `.`), with the variables `git rev-parse --local-env-vars` lists removed from Git's environment, archives that commit by its full hash, extracts it into a private temporary directory, and plans an install for every host already registered in user scope. The plan freezes the extracted bytes, so the temporary directory is removed on every return path as soon as the plan is built, before the summary and the confirmation prompt; only a process killed by a signal during extraction or planning may leave a `.hive-extract-*` directory in the system temporary directory. A revision starting with `-` is rejected before Git runs. The command needs Git and a Git checkout of Hive; the offline package does not, and installs through `install.sh`. `export-ignore` or `export-subst` attributes change the archived content, as with any `git archive`: those in the selected commit, and also local ones in `.git/info/attributes` or the file named by `core.attributesFile`, which are not part of the commit.

In a terminal, `update` shows the summary with the commit and asks for confirmation. When nothing changes, `--out FILE` still saves the plan, and applying it records the commit; without `--out`, `update` applies it without asking, which only records the commit, also without a terminal. When the plan changes files and there is no terminal, it fails unless `--dry-run` previews the change or `--out FILE` saves a plan for `hive apply --plan FILE`. To return to an older commit, run `update --rev <commit>`; the current binary must still validate that commit's sources.

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

`hive doctor` prints five sections with the same text as the Diagnostics, Integrations, and Project views: CLIs, Installation, Sessions, Integrations, and Project. `hive models` prints, per registered host, the effective model and effort of each installed role, with the same words as the Models view. Both write no file or state and exit with status 0 when they report findings; only their own errors, such as a state directory that is a file, give a non-zero status.

- **CLIs** lists each supported host in installer order: detected on `PATH` or not (Cursor counts as detected when either `cursor-agent` or `cursor` is found, in this report and in `hive install` and the CLIs view alike; both read one list of binary names), `CLI` with the first line of `<binary> --version` (`cursor-agent` for Cursor, so a host with only `cursor` shows `CLI unavailable: cursor-agent not found`), and `Hive` with the host's release and installation state (`verified`, `partial`, `drift`, `legacy`), as in `claude  detected  Hive 4f8d96b9ce4f (verified)  CLI 2.1.284 (Claude Code)`, so the CLI's version is not mistaken for Hive's. Only a detected host is executed, with fixed arguments, no shell, and a 3-second limit; a failure or timeout shows `CLI unavailable` and its reason.
- **Installation** lists every `status` row other than `installed`, `retained_shared`, and `not_installed`, with a plain phrase and its path, or `No problems found`. Hosts that share one managed file, such as a skill under `~/.agents/skills`, are grouped into one row per status and path that names them all, as in `drift  claude, codex, cursor, grok, opencode, pi  The installed file was changed or cannot be read` followed by the path once, in the order the rows first appear. When a row is `drift`, one line explains that Hive cannot repair a changed file by itself: `install`, `update`, and `plan remove` refuse to run while a managed file differs from what Hive wrote, so the user undoes the change, fixes its permissions, or restores the file from a backup, then opens Diagnostics again, or runs `hive doctor`, to check. With a pending operation it shows one line naming `hive recover` instead of the `recovery_required` rows, because `status` then marks every row that way; the line adds `--state-dir DIR` unless that directory is the real user home's default, since `hive recover` has no `--home`.
- **Sessions** covers the registered hosts. When Claude Code or Grok is registered, its first line answers whether to restart anything for the hosts Hive checked, such as `2 open Claude Code or Grok sessions should be restarted` or `No open Claude Code or Grok session needs a restart`, and names a host it could not check, as in `No open Claude Code session needs a restart; Grok could not be checked.` When Codex, Pi, or Cursor is registered, the next line says `Hive cannot see Codex, Pi or Cursor sessions; restart them after each update.`, naming only the registered ones. Each host's line then reads, for example, `claude: 3 open sessions, 1 to restart`. It reads Claude Code's `sessions/*.json` under `CLAUDE_CONFIG_DIR` or `~/.claude` and Grok's `active_sessions.json` under `GROK_HOME` or `~/.grok`. It counts only sessions whose process is alive and marks one `started before the installed release; restart it` when it started before the installed release's snapshot was last written. It shows only the process ID, working directory, and start time. A missing session directory or file means no open sessions; a file larger than 1 MiB, unreadable, not a regular file, or in an unknown format, or more than 500 Claude Code session files, makes that host `Session check unavailable` without failing the other sections. OpenCode gets a note that it reloads its instructions on the next message.
- **Integrations** shows Engram, Context7, pi-subagents, and `agent-browser`: the local evidence (the program on `PATH` or the skill file, with its path), the status in the last onboarding record, the official source, and a concrete next step for its case. The only command it suggests is Context7's `npx ctx7@latest setup --cli`, which may ask the user to sign in; for the others it points to the Source line above, since Hive does not install them. It executes no integration program and reads no host or provider configuration, so pi-subagents shows only its onboarding status. `agent-browser` is detected but is not part of the optional capability catalog.
- **Project** validates the `## Hive` section of `AGENTS.md` at the root of the Git repository that contains `--project`, or the current directory: missing or duplicated section, missing or empty required values (`Project`, `Base branch`, `Tracker`, `Specs`), a `Specs` path that is not a directory, a `Base branch` that is neither a local branch nor on `origin`, unknown keys, and `Delivery` or `Hive guidance` values other than `direct-base` and `required`. A `--project` directory that does not exist, or is not a directory, is reported as `Directory not found` or `Not a directory` without running `git`. It runs read-only `git` queries with Git's local environment variables removed. A section in a workspace `AGENTS.md` outside a repository is not checked.

With `--home`, as elsewhere, host environment overrides are ignored, no program is looked up on `PATH` or executed (neither the host binaries nor `engram` and `agent-browser`), and sessions are read under that home; the Project check still finds and runs `git` in `--project` or the current directory. Text read from outside the manager (versions, paths, session directories, `AGENTS.md` values) has control characters and escape sequences removed before it is shown.

Limits: the session files of Claude Code and Grok are internal and undocumented, so any host version may change them and turn the check into `Session check unavailable`. The restart mark is an estimate: a release snapshot is rewritten by every transaction that carries it, such as adding a host, which moves the reference time forward and can mark sessions that already have the current content; `/compact` in Claude Code and voice changes do not move it; the write time has one-second precision; and a reused process ID can make a finished session look alive. Codex, Pi, and Cursor have no session detection.

To change a model or effort, edit `integrations/agent-profiles.json` in the Hive checkout and run `hive update`; the diagnostics never write models or `## Hive` values.

### Terminal interface

```sh
go run ./tooling/cli
go run ./tooling/cli tui [--home DIR] [--state-dir DIR] [--source DIR]
```

`hive` without arguments in a terminal, or `hive tui`, opens a full-screen application on the terminal's alternate screen: a menu with CLIs, Update, Releases, Voice, Diagnostics, Models, Integrations, Project, and Quit, under a status line with the number of registered hosts, the installed release, and the voice (when Hive's state cannot be read it says `Status unavailable: state could not be read; see Diagnostics.` and never prints the raw error, which Diagnostics shows), and above a help bar that lists the current view's keys. Changing views redraws the whole screen, and leaving the application restores the terminal's previous content. Below 80×24 it shows the minimum size instead of the view, and keeps the view's state until the terminal is large enough again. The theme follows the terminal's light or dark background; `NO_COLOR` removes colors, and checkboxes, the cursor, and the selected button always use symbols.

The application needs a terminal on standard input and output. Without one, `hive` without arguments keeps its usage error, and `hive tui` fails naming the text commands. Those commands (`hive status`, `install`, `update`, `releases`, `voice`, `doctor`, `models`, `plan`/`apply` to remove hosts, and `recover`) are the path for scripts and screen readers; the application has no line-based mode. `hive install` remains the plain-text installer that `install.sh` and the online bootstrap use.

Keys:

- **Esc** returns to the previous view, and closes the application at the menu. **Backspace** does the same when no text field has the focus; in a text field it deletes a character, and at the menu it does nothing.
- **Ctrl-C** closes the application with exit status 0 from any view. While a change is being applied, every key, Ctrl-C included, and an external interrupt (SIGINT) are ignored until it finishes.
- **↑↓** move, **←→** change a value or choose a button, **Enter** accepts, and **Space** toggles a checkbox.

Each view reuses the plans and summaries of its command, so it writes the same files and state:

- **CLIs** lists each detected, registered, or legacy host with a checkbox (checked means active), its release, the Hive version (the column header shortens to `Hive` when long names leave no room; the CLI's own version is in Diagnostics), and drift. Space changes the desired set and `a` applies it. Removals are previewed, confirmed, and applied first; additions are then planned on the resulting state, with the package verification, required-host notice, optional capabilities, summary, and confirmation of `hive install`. When both happen, the summaries are labeled step 1 and step 2, and stopping at step 2 keeps the removal. Checking a host with a legacy installation migrates it. When a file Hive installed was changed by hand, the scan for legacy installations fails on it (it reads the file as an old Hive file it cannot remove); the view still lists every host with its release and drift and adds `A file Hive installed was changed (<path>). Open Diagnostics to see how to restore it before installing or removing.`, with the raw error on a `Detail:` line that gives way to a result or error shown below it. While that lasts, a host with a legacy installation may not be labeled, and removing or installing is refused with the command's own message, writing nothing, until the file is restored. `u` opens Uninstall all, which removes every registered host; its confirmation starts on Cancel and does not accept `y`.
- **Update** edits the source (`.`) and revision (`HEAD`) in place and shows the `hive update` summary. When the revision changes nothing, it records the source commit without asking, as `hive update` does.
- **Releases** lists retained releases, newest first, with the installed one marked. `/` filters by text and Esc clears the filter. Choosing a release returns to it through `plan install --release` for the registered hosts.
- **Voice** shows the active voice as rows (voice, address, name when the address is `name`, and intensity) changed in place with the arrows, then previews `hive voice set` or `off`. With Off, only the voice row shows.

The four read-only views show what `hive doctor` and `hive models` print (see [Read-only diagnostics](#read-only-diagnostics)) and change nothing. Each loads in the background with a spinner, `r` reloads it, and a load error shows inside the view with `r to retry`. Long content scrolls with PgUp and PgDn, and with ↑↓ except in Integrations:

- **Diagnostics** shows the CLIs, Installation, and Sessions sections.
- **Models** shows one host at a time, chosen with ←→ and shown in brackets, as a table of role, profile, model, and effort; the role name is never cut.
- **Integrations** lists the four integrations with a `>` cursor moved by ↑↓, and the selected one's detail below.
- **Project** shows the checked `AGENTS.md` path and `Valid` or the findings, with the values read.

Apart from that unchanged update, every writing action shows its summary, which scrolls, and asks for confirmation. Declining or going back shows `Cancelled. No changes applied.` inside the originating view; an error shows the command's message there; a successful action refreshes the view and shows its result. When the application opens with a pending operation, it offers to recover it. Recovery messages name `hive recover`, adding `--state-dir DIR` when the application was opened with an explicit state directory. Adding hosts in CLIs, and Voice, read the catalog from `--source`, which defaults to the current directory: outside a Hive checkout or package they report `Run hive from a Hive checkout or package, or pass --source`.

## Optional Context7 setup recommendation

`hive setup` is a read-only onboarding step. It checks known skill locations for nonempty `find-docs/SKILL.md` and `context7-mcp/SKILL.md` files, including the shared `.agents` location and configured host homes. `--home DIR` selects a synthetic home and ignores environment overrides. A user-scope install plan without an existing Hive state file prints a first-setup recommendation; it never blocks installation on Context7 availability.

Context7 is strongly recommended for development workflows that need current library documentation, but is optional. For these terminal-based harnesses, CLI + Skills is the recommended default because it reuses shell execution and native skill discovery without per-host MCP registration. MCP remains valid when native structured tools or an environment without shell access are preferable; existing setups are preserved. A discovered file does not prove host loading, authentication, service availability, or freshness. MCP-only and custom installations may not be detected. Hive never reads credentials, launches authentication, or installs/updates Context7 from this check. Official documentation distinguishes unauthenticated `ctx7 library`/`docs` queries (lower limits) from the hosted setup wizard, which requires authentication. Users without service access should consult current official documentation directly and disclose verification limits.

The suggested voluntary install/refresh command is `npx ctx7@latest setup --cli`. Node.js/npm and network access are needed only if the user chooses that vendor command. The official setup owns agent selection, authentication, its `find-docs` skill, and its instruction rules. Existing MCP users can keep their mode. Do not vendor its content, silently migrate modes, or treat third-party files as Hive-owned removal targets. Selecting `@latest` resolves the current CLI when invoked; rerunning setup downloads the vendor skill and updates its rules. This is not a background-update guarantee or a read-only version check.

Verified against installed `ctx7` 0.5.11 and [official CLI documentation](https://context7.com/docs/clients/cli). Its setup supports Claude, Codex, and OpenCode flags but has no Pi/Grok flags; shared skill-file discovery does not establish runtime support for those hosts. The vendor setup may request authentication and rewrites selected vendor skill/rule files even when they already exist. Hive's setup check does neither.

## Content and native destinations

The catalogue contains `content/guidance/global.md`, each `content/skills/<skill>/SKILL.md`, resources recursively below each skill's `references/`, `scripts/`, and `assets/` directories, and canonical roles at `content/agents/<category>/<name>.md`. Skill resources may be `.md`, `.html`, `.py`, `.sh`, `.json`, `.ts`, `.mjs`, `.astro`, `.mdx`, `.css`, `.yaml`, `.yml`, or `.gitignore`; their source modes are frozen and restored, including executable scripts. Agent sources and `integrations/agent-profiles.json` are validated as one catalogue: role names must be unique across categories. Links, duplicate paths, escaping paths, empty payloads, and reserved block delimiters are rejected. Disposable `__pycache__`, `node_modules`, `dist`, and `.astro` directories beneath skill resources are excluded.

Each release ID is a SHA-256 over its ordered payload manifest, source modes, frozen agent-profile input, and renderer version. Resource identity is its source path (`Target.Source`), so a resource cannot be substituted for another payload. Agent files are rendered in memory for each native host and their exact rendered bytes are frozen into the saved plan and snapshot; the profile file itself is never deployed. No generated distribution directory is required for local use. In the paths below, `<relative-file>` is a skill entrypoint or resource.

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

## Preservation and state

Only the managed global block is replaced. Existing bytes outside it and the file's permissions survive. The first insertion records any added separator and whether Hive created the file. Removal deletes only that managed span; a newly created file is removed only if no user content remains. Every skill resource and agent file must still match its installed bytes and recorded mode before update or removal. New skills and agents cannot adopt existing files. An already-owned skill can gain absent resources; an unowned file at a planned destination is a conflict. Additional user files outside managed destinations remain untouched.

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
- `transactions/<id>.json`: private before/after images and transaction phase.
- `pending.json`: an unfinished operation requiring recovery.
- `lock`: process lock, released by the operating system on exit.

New state directories use mode 0700; state files, plans, and backups use 0600. Unknown schemas or malformed state fail rather than trigger automatic rebuilding. Backups and release history are retained; the manager does not prune them or store credentials. State, operation plans, and new core transaction journals use schema v6. The product-version receipts and version index are separate from historical content release IDs; v6 preserves their hashing. Optional onboarding uses a separate versioned parent journal and does not roll back a committed core on provider failure. It adds verified legacy retirement and scoped migration receipts to v4's source payload modes, renderer/profile identity, and shared consumer ownership. Existing v1-v5 state migrates transactionally when an operation is applied; planning alone does not rewrite it. Retained legacy releases preserve their original payload hashes. Pending v1-v5 transactions remain recoverable with their original checksum semantics. Saved plans before v6 must be regenerated before applying new operations. Plans saved before Cursor support was added also fail to apply (`invalid root`) and must be regenerated.

## Failure handling

The manager validates the whole plan before destination writes, records recovery data before mutation, and rechecks state, resolution, and file fingerprints. Replacements are atomic per file, not across all files. The lock coordinates Hive processes; it cannot lock unrelated editors. The manager checks for competing edits and preserves detected conflicts, but does not claim protection against every malicious filesystem race.

An interrupted operation blocks further plans/applications until `recover`. Recovery preflights all inverse operations. It can preserve newer text outside an installed block while restoring the previous managed span; changes inside the managed span cause a conflict. It never restores a whole old configuration over detected later edits. Recovery of an already committed journal finalizes bookkeeping instead of undoing the installation.

Only directories confirmed created by Hive are cleanup candidates, and only empty directories are removed. A crash between creating a directory and recording ownership may leave an empty directory behind; preserving uncertain ownership is preferable to claiming it retroactively. Directory ownership is recorded by path, not inode: replacing a previously created directory with an empty directory at the same path is not distinguishable during cleanup. No `--force` or blanket purge exists.

## Verification

```sh
go test ./...
go test -race -timeout 20m ./...
go vet ./...
```

The race run of `tooling/cli` takes about thirteen minutes, beyond `go test`'s default ten-minute limit, so pass `-timeout` explicitly.

Tests exercise the six user-scope mappings and the two supported project mappings with synthetic homes and projects, preserved LF/CRLF content and permissions, snapshots/downgrades, idempotence, conflicts, stale plans, links, concurrent directory creation, and recovery at each write boundary. Catalogue tests additionally cover multiple skills and nested Markdown references, source identity and forged payload rejection, new references in owned bundles, unowned reference conflicts, partial consumer retirement, old-release rollback, missing-checkout status/removal, v2 migration and legacy journal recovery. The pilot protocol and its limitations are in the [workspace screening fixtures](../../../tests/fixtures/workspace-conventions/README.md).

## Product versions and optional operations

The product version is independent of the state/manifest schemas and Git revision.
`VERSION` is the source-tree label; `hive --version` describes the running manager.
Per-consumer receipts describe the requested artifact and its expected resource
bytes/modes/aliases. Status distinguishes that request from current conformity.
Shared resources may retain their original provenance without implying drift.
A legacy snapshot with no product receipt retains its hash and is not relabeled.

Published version labels cannot be rebound to a different artifact in the local
version index. Development builds use `dev`. This local check does not certify a
remote release channel or authorize public distribution. Source-code visibility
and package publication remain independent decisions.

Optional steps are explicit and journaled before execution. A provider execution
with an unknown outcome blocks further mutation until reconciliation. Recovery
inspects that outcome without rerunning an installer or authentication. A committed
core remains installed if an optional provider fails. Status is still readable
while reconciliation is pending. Existing tools, credentials and memory are not
owned merely because the onboarding catalog detected them.
