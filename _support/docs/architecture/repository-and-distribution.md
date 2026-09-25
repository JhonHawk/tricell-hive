# Hive repository structure and distribution

Reviewed: 2026-09-21. Structure agreed and directories created. The global rules, workspace skill, and first Go deployment manager are implemented. The adapters have synthetic-home lifecycle tests; the five-host user-global deployment is installed for an experimental observational rollout. See the [manager contract](deployment-manager.md) for commands and current limits.

## Responsibilities

| Path | Responsibility | Current state |
|---|---|---|
| `content/guidance/global.md` | Single source for distributed global rules | Installed globally; behavioral evidence is reported separately |
| `content/skills/` | Activity procedures and their supporting resources | `workspace-conventions` and three `flow-*` skills installed; flow planning includes a Markdown reference |
| `integrations/{claude,codex,grok,pi,opencode,cursor}/` | Native destination differences | Six user-scope adapters; project scope for Codex and Claude |
| `tooling/cli/` | Command interface | Go CLI, invoked from the checkout |
| `tooling/management/` | Shared installation, diagnosis, and removal logic | Managed blocks, snapshots, plans, state, and recovery |
| `tooling/tui/` | Future interface over the same operations | Outside initial scope |
| `tests/{content,integrations,management,fixtures}/` | Content, integration, and installation lifecycle verification | Lifecycle tests, workspace fixtures, TypeScript flow fixtures, frozen baseline, and native CLI pilot runner |
| `_support/docs/` | Durable research and decisions | Existing |
| `_support/sessions/` | Resumable work records | Available |
| `_support/evidence/` | Curated evidence suitable for version control | Available |
| `_support/workspace/` | Local scratch; retains `hive-retirement/` | Only `.gitkeep` markers are versionable |
| `dist/` | Generated distribution artifacts | Neither created nor versioned; generated when needed |

Empty source directories use `.gitkeep`; generated output such as `dist/` is not retained. Do not create example skills, empty manifests, or executables to imply functionality. Packages may contain derived copies, but the editorial source remains unique. This structure does not select a programming language, dependencies, or per-host package formats.

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

### Why an instruction-file block and not a session-start hook

Researched on 2026-09-24 against the installed hosts: Claude Code 2.1.281, Codex 0.156.1, Grok Build 1.0.38, Pi 0.86.1 and OpenCode 2.0.15. A hook injects the same text, so it does not reduce density. It also cannot deliver the block with the same effect:

| Host | Session-start injection | Limitation |
|---|---|---|
| Claude Code | `SessionStart` `additionalContext` | Silently truncated at 10,000 characters ([anthropics/claude-code#94358](https://github.com/anthropics/claude-code/issues/94358), open). Does not fire for subagents, which do inherit `CLAUDE.md`. `disableAllHooks` in any settings level silences it. Observed: the host frames `CLAUDE.md` as overriding default behavior; hook output gets no such framing. |
| Codex | `SessionStart` `additionalContext` | About 2,500 tokens by default. Skipped in `codex exec` without a bypass flag ([openai/codex#46210](https://github.com/openai/codex/issues/46210), open). A project `.codex/config.toml` can set `features.hooks = false`, while `AGENTS.md` has no such switch. Its output fills the TUI; Engram moved off this hook for that reason. |
| Grok Build | None | Bundled hook documentation: `SessionStart` stdout is ignored. |
| OpenCode | None | Requested and closed three times ([anomalyco/opencode#5409](https://github.com/anomalyco/opencode/issues/5409), not planned). |
| Pi | Extension `before_agent_start` | Can inject; no documented size limit. |

The reference projects agree. `optional reference project` injects only a ~3 KB bootstrap by hook, removed its Codex hook as worse UX, and never edits the user's global file. `optional reference project` uses a marker block and thinned it (`75c6b2c7`). `optional reference project` injects full rules only because they are 2.6–6.6 KB. The legacy Hive never injected rule text by hook: forcing a reference into the user turn, a stronger channel than any hook, did not improve compliance (legacy `_support/docs/methodology-bibliography.md:1323`).

A hook fits a small payload, conditional injection by session source or project, or re-injection after compaction. The lever for density is a thinner always-on block that keeps gates in place and moves procedure to skills. Rules that must always hold stay in the block, because router and reference delivery reached the model in only about 45–70% of measured sessions.

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
