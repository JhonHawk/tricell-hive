# Deployment manager

The first manager is implemented in Go 1.27 using only the standard library. It runs from this checkout with `go run ./tooling/cli`; macOS is the tested platform. The manager never launches models. The separate pilot runner is evaluation tooling.

## Commands

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

The default source is the current directory; `--source` selects another checkout with the same content layout. `plan install --release <hash>` selects a retained installed snapshot instead, allowing an explicitly planned downgrade. A plan freezes its source bytes; later source edits do not silently alter it.

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

The default state home is `~/Library/Application Support/tricell-hive`:

- `state.json`: schema version, installed resource records, release IDs, and owned directories.
- `releases/<hash>.json`: immutable installed payload snapshots.
- `transactions/<id>.json`: private before/after images and transaction phase.
- `pending.json`: an unfinished operation requiring recovery.
- `lock`: process lock, released by the operating system on exit.

New state directories use mode 0700; state files, plans, and backups use 0600. Unknown schemas or malformed state fail rather than trigger automatic rebuilding. Backups and release history are retained; the manager does not prune them or store credentials. State, operation plans, and new transaction journals use schema v4. It adds source payload modes plus the renderer/profile identity used for agent output, while retaining explicit source identity and shared consumer/alias ownership. Existing v1-v3 state migrates transactionally when an operation is applied; planning alone does not rewrite it. Retained legacy releases preserve their original payload hashes. Pending v1-v3 transactions remain recoverable with their original checksum semantics. Saved plans before v4 must be regenerated before applying new operations. Plans saved before Cursor support was added also fail to apply (`invalid root`) and must be regenerated.

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
