# Deployment manager

The first manager is implemented in Go 1.27 using only the standard library. It runs from this checkout with `go run ./tooling/cli`; macOS is the tested platform. The manager never launches models. The separate pilot runner is evaluation tooling.

## Commands

Run from the checkout root. Both hosts and scope must be explicit for planning and status:

```sh
go run ./tooling/cli --help
go run ./tooling/cli plan install --hosts codex,claude,grok,pi,opencode --scope user
go run ./tooling/cli plan install --hosts codex --scope project --root /absolute/project --out /absolute/new-plan.json
go run ./tooling/cli apply --plan /absolute/new-plan.json
go run ./tooling/cli status --hosts codex --scope project --root /absolute/project
go run ./tooling/cli plan remove --hosts codex --scope project --root /absolute/project --out /absolute/remove-plan.json
go run ./tooling/cli recover --state-dir /absolute/state-directory
```

These are interface examples, not authorization to deploy into the user's real configuration. `plan` without `--out` is a read-only preview. With `--out`, it writes a new private plan file, never the destination files. `apply` consumes that plan and rejects stale target fingerprints or ownership state. Destination options are not accepted by `apply`: review the destinations bound into the plan instead.

`--home` explicitly selects a synthetic home and ignores host path and compatibility environment overrides. Use it together with `--state-dir` for temporary-home tests. Without `--home`, host configuration environment variables are respected. `--root` is required for project scope. Existing root prefixes are canonicalized; linked target files or ancestors inside those roots are rejected rather than overwritten.

The default source is the current directory; `--source` selects another checkout with the same content layout. `plan install --release <hash>` selects a retained installed snapshot instead, allowing an explicitly planned downgrade. A plan freezes its source bytes; later source edits do not silently alter it.

## Content and native destinations

The allowlist contains exactly the global guidance body and the workspace-conventions `SKILL.md`. Research, fixture inputs, reports, caches, README files, and `.gitkeep` markers are not payload. Each release ID is a SHA-256 over its ordered payload manifest. No generated distribution directory is required for local use.

| Host | User scope | Project scope |
|---|---|---|
| Codex instructions | Effective `AGENTS.override.md` or `AGENTS.md` under its configured home; an empty override does not win | The same precedence at the selected project root |
| Codex skill | `<user-home>/.agents/skills/workspace-conventions/SKILL.md` | `<root>/.agents/skills/workspace-conventions/SKILL.md` |
| Claude instructions | `<claude-config-dir>/CLAUDE.md` | `<root>/CLAUDE.md` |
| Claude skill | Directory alias `<claude-config-dir>/skills/workspace-conventions` to the shared `.agents` skill | `<root>/.claude/skills/workspace-conventions/SKILL.md` |

The user defaults are documented by [Codex instructions](https://learn.chatgpt.com/docs/agent-configuration/agents-md), [Codex skills](https://learn.chatgpt.com/docs/build-skills), [Claude memory](https://code.claude.com/docs/en/memory), and [Claude skills](https://code.claude.com/docs/en/skills). Observed host versions for this implementation are Codex 0.155.1, Claude Code 2.1.278, Grok Build 1.0.34, Pi 0.86.0, and OpenCode 2.0.9. Documentation does not establish successful loading in every environment. The five-host rollout report (historical evidence omitted from public history) separates installed state, observed loading, and task outcomes.

A physical resource records its registered consumers (`host`, `scope`, `context`). Identical selected destinations share a single resource. Updating shared bytes requires selecting every registered consumer. Removing one consumer retains resources still used by another; `retained_shared` means the native host may still discover that resource. Registration is not a host-level visibility switch.

User scope also supports Grok, Pi, and OpenCode. Grok uses the native `~/.claude/CLAUDE.md` compatibility path and the shared `.agents` skill; disabled or ambiguous Claude instruction compatibility is a conflict. Pi uses `PI_CODING_AGENT_DIR` or `~/.pi/agent/AGENTS.md`. OpenCode uses `XDG_CONFIG_HOME/opencode/AGENTS.md` or `~/.config/opencode/AGENTS.md`. All three use the shared `.agents` skill and reject project scope. Grok's `GROK_HOME` controls its configuration lookup, not the fixed Claude compatibility path.

Only the managed Claude skill-directory alias is supported. Its exact relative link text is journaled and checked; unknown links, retargeted links, unsafe ancestors, changed override resolution, unowned skill directories, or malformed/unknown Hive blocks conflict. Payload creation precedes alias creation; alias removal precedes final payload removal. Whole skills roots and instruction files are never linked. The installed payload is a versioned copy, not a link to this checkout.

The manager does not follow arbitrary import graphs, adopt unrelated deployments, or detect every unregistered CLI consuming a conventional shared path. A successful `status` reports filesystem state, not runtime discovery or loading.

## Preservation and state

Only the managed global block is replaced. Existing bytes outside it and the file's permissions survive. The first insertion records any added separator and whether Hive created the file. Removal deletes only that managed span; a newly created file is removed only if no user content remains. Skill content must still match the installed bytes before update or removal. Additional user files in a skill directory remain untouched.

Plans contain managed content, fingerprints, and ownership metadata, not the surrounding private instruction text. Preview output shows the old and new managed content rather than a whole-file diff. Backups necessarily contain original target bytes and must stay private.

The default state home is `~/Library/Application Support/tricell-hive`:

- `state.json`: schema version, installed resource records, release IDs, and owned directories.
- `releases/<hash>.json`: immutable installed payload snapshots.
- `transactions/<id>.json`: private before/after images and transaction phase.
- `pending.json`: an unfinished operation requiring recovery.
- `lock`: process lock, released by the operating system on exit.

New state directories use mode 0700; state files, plans, and backups use 0600. Unknown schemas or malformed state fail rather than trigger automatic rebuilding. Backups and release history are retained; the manager does not prune them or store credentials. State schema v2 supports shared consumers and managed aliases. Existing v1 state migrates transactionally; pending v1 recovery keeps its original checksum semantics. Regenerate saved v1 plans before applying new operations.

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

Tests exercise the five user-scope mappings and the two supported project mappings with synthetic homes and projects, preserved LF/CRLF content and permissions, snapshots/downgrades, idempotence, conflicts, stale plans, links, concurrent directory creation, and recovery at each write boundary. The pilot protocol and its limitations are in the [workspace screening fixtures](../../../tests/fixtures/workspace-conventions/README.md).
