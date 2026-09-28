# Deployment manager

The first manager is implemented in Go 1.27 using only the standard library. It runs from this checkout with `go run ./tooling/cli`; macOS is the tested platform. The manager never launches models. The separate pilot runner is evaluation tooling.

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

With a voice active, `plan install`, and therefore `hive update`, keeps the span and regenerates it with the same choice when the voice source text or its rendered form changed; the summary then reports the voice files to regenerate. A host that gains its Hive block in the same plan gets the voice too. `plan install --release <hash>` carries no voice text and leaves spans untouched. `plan remove` removes a file's voice span with its last consumer, and forgets the voice once no span remains. `status` adds a `voice` row per span with the voice, address, and intensity, and reports `drift` when the span was edited by hand. A hand-edited span, or voice markers in a file with no recorded span, is a conflict that every operation on that file preserves.

The voice is stored in `state.json` as `voice` and `voice_spans` without a schema change. A manager from before this feature reads that state but drops those fields, leaving the spans in the files; the current manager then reports them as unregistered conflicts. Run `hive voice off` before returning to an older manager.

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
go test -race ./...
go vet ./...
```

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
