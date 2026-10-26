# Hive repository structure and distribution

Reviewed: 2026-10-25. The shared guidance, activity skills, six host adapters, Go deployment manager, and packaged installer are implemented in this repository. Synthetic-home lifecycle tests cover the adapters. Historical rollout records describe earlier global installations; verify current host state separately. See the [manager contract](deployment-manager.md) and [installer contract](installer.md) for commands and limits.

## Responsibilities

| Path | Responsibility | Current state |
|---|---|---|
| `content/guidance/global.md` | Canonical core, essential boundaries and linked read conditions | Packaged by the manager; installation is not proof of loading |
| `content/skills/` | Activity procedures and their supporting resources | The authored skills under `content/skills/`; availability and loading depend on each host installation |
| `integrations/{claude,codex,grok,pi,opencode,cursor}/` | Native destination differences | Six user-scope adapters; Codex/Claude project scope; optional inactive OpenCode loader |
| `tooling/cli/` | Command interface | Go CLI, invoked from the checkout or a complete package |
| `tooling/management/` | Shared installation, diagnosis, and removal logic | Managed blocks, snapshots, plans, state, and recovery |
| `tooling/viewer/` | Local walkthrough viewer behind `hive view` | Standard library only; store, ingest, and loopback HTTP server with embedded pages |
| `tooling/package/` | Offline package builder | Builds complete platform packages on demand |
| `tests/{content,fixtures,pilot,skills}/` | Content, fixture, pilot, and skill verification | Integration and management tests live beside their Go packages |
| `site/` | Documentation site for people | Starlight site published from `master`; not packaged or deployed by the manager |
| `_support/docs/` | Durable research and decisions | Existing |
| `_support/sessions/` | Resumable work records | Available |
| `_support/evidence/` | Curated evidence suitable for version control | Available |
| `_support/workspace/` | Local scratch; retains `hive-retirement/` | Only `.gitkeep` markers are versionable |
| `dist/` | Generated distribution artifacts | Ignored by Git; generated when needed |

Only the retained local-workspace directories use `.gitkeep`; populated source directories need no marker. Generated output such as `dist/` is not versioned. Do not create example skills, empty manifests, or executables to imply functionality. Packages may contain derived copies, but the editorial source remains unique.

## First content: global rules

Maintain one source file in this repository and manage a block within the user's existing global instruction file. This preserves user instructions and other integrations.

Managed distribution markers:

```markdown
<!-- === TRICELL HIVE RULES:BEGIN === -->
Content from content/guidance/global.md
<!-- === TRICELL HIVE RULES:END === -->
```

Markers occupy complete lines and have distinct, stable start and end identifiers. HTML comments delimit ownership for the tool; they do not hide instructions from the model or change their priority. The installed version and hash belong in the local installation record, not in marker identifiers.

Each integration must resolve the effective global file from documentation and observed host versions. Do not assume all CLIs read the same file, that `CLAUDE.md` always imports `AGENTS.md`, or write to both when doing so duplicates loading. Host compatibility may give a file multiple consumers. Installation records those consumers, and partial removal preserves the consumers that still use the block.

### Core, references and optional native supply

The core remains in managed instruction channels. Authorization, preservation,
secrets, evidence and basic artifact placement do not depend on selecting a
procedure or activating a plugin. Canonical references carry reporting,
questions, delegation, artifact details and settings with direct read conditions;
reporting is read under the conditions in `content/guidance/global.md`. Reading a reference does not invoke
its owning flow. See [instruction resources](instruction-resources.md) for homes,
runtime resolution, validation and declared reading-footprint scenarios.

The earlier September host comparison is historical, not a current prohibition
on hooks. The October change consulted current official sources, including
[Claude hooks](https://code.claude.com/docs/en/hooks) and
[OpenCode V2 plugins](https://opencode.ai/v2/docs/build/plugins). Claude now
documents a file/preview for long context rather than the previous silent
truncation claim; pre-tool context arriving with the tool result is not a
preventive control over that operation. The old 45–70% delivery estimate is not
reproducible from the available evidence and is not an acceptance threshold.

Grok gets one managed `PreToolUse` hook that drops `offset` and `limit` when a
session reads a skill reference, because Grok models read references partially
in about one session in three; for skill references read with Grok's read tool,
it enforces part of the read-whole rule that the global guidance states for
every host. Its contract is in [the deployment manager](deployment-manager.md#grok-read-whole-hook).

An optional OpenCode user-scope plugin supplies two fixed-role reference snapshots
in the session context hook when explicitly activated. The manager generates
them from the validated release payload; canonical references retain one editorial
home. Requests read no reference files or metadata. Refresh requires manager
regeneration and plugin reload, not a live installed-Markdown edit. It is delivered inactive
at `<OpenCodeHome>/hive/context-loader/{index.js,context-loader.json}`
with no `package.json`, outside auto-discovery, not as a replacement for the core or a six-host loader.
Its editorial JS template is included in manifest v2 offline packages; activation
targets only the installed `hive/context-loader` directory. Metadata schema 2 records fixed
role/resource/digest provenance and is independent of package manifest v2; the current manager
also verifies v1 packages. Older v1-only managers cannot verify v2 packages.
See [activation and lifecycle](deployment-manager.md#optional-opencode-context-loader).

Packaging, native discovery, reference supply and behavior remain separate
claims. Current source/callback checks do not establish native final-request
delivery or better outcomes for Haiku/Flash. The bounded A/B/C model experiment
is authorized for this change but has not yet run and remains blocked on verifying
safe, exact memory cleanup. That limit does not invalidate the loader's separately
checked generation and lifecycle contracts. A smaller global file is not
lower total context: the simple-query union counts reporting, and
the declared retained plan/build cycle grows in source bytes. No arbitrary quota
or byte-to-token estimate establishes reliability.

Coexistence with user rules: the block is appended after the user's content, and the markers do not set priority. Precedence between the block and user-authored text in the same file is undeclared. Grok Build gives later text precedence, so there the block effectively wins over the user's earlier rules. Hook delivery would leave the user's file untouched, but each host would order the text differently and could silently disable it.

## Installation and update contract

1. Resolve explicit destination and scope; detect existing files, symbolic links, and shared resources before writing. Do not inadvertently replace a symbolic link.
2. Read and show the proposed change. Preserve bytes outside the managed block, including line endings. If appending the block requires a separator, record the added bytes as part of the managed resource.
3. With no markers, append one block without replacing existing content. With one valid pair and recorded ownership, update only that block. Do not automatically adopt blocks of unknown origin.
4. Incomplete, duplicate, nested, or reversed markers are a conflict: preserve the file and explain the problem. Source content must not contain reserved delimiters.
5. If the block changed since installation, present the conflict and preserve the edit rather than silently overwriting it.
6. Check that the destination has not changed between reading and writing; use safe replacement, preserve permissions, retain a backup, and record the result. Repeating an operation with the same input must produce no changes.
7. Verify the resulting file and distinguish installation on disk from loading in a session. Report any host-required reload or restart.

The manager can target user or project scope, but implementation verification uses temporary destinations. Real global deployment is a separate explicit action. An empty source is not distributed; instruction removal is a separate operation.

## Deactivation and removal

Remove only the identified block and separators added by the tool, preserving everything else. Detect subsequent modifications, shared resources, and sessions that may still use an earlier version. Do not restore an entire configuration file over later user changes.

Record creation of a previously absent global file: remove that entire file only if it remains exclusively the resource Hive created. If the user added content, preserve the file. Uninstalling Hive does not delete authentication, preferences, history, Engram, or backups; purging these is separate.

Installation state and backups live outside the checkout in the local directory described by the [manager contract](deployment-manager.md). Tests use an explicit temporary state directory. Global state is not created merely by reading or planning a deployment.

## Agreed future capability

The manager may incorporate persistent configuration and reusable decisions with scope, conditions, provenance, revision, and revocation. CLI, TUI, and a possible skill use the same validated operations rather than independent direct database writes.

This capability is outside the first version. The database, schema, and commands remain undecided. Repeated preferences or past approvals do not become standing authorization by repetition.

## Design sources

- [Workflow map](../harness-engineering/2026-09-20-workflow-map.md).
- [Harness engineering research](../harness-engineering/2026-09-20-portable-harness-research.md).
- [Workspace and artifact organization](workspace-and-artifacts.md).
- Session agreements: shared content, minimal per-CLI integration, reversible distribution, CLI before TUI, and future decision storage.
