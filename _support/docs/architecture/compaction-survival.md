# What survives context compaction, per host

This page records, for each supported host, whether the global instruction file and the body of an invoked skill survive context compaction, and which surface can inject context again after it. It informs the re-injection layer of the layered global guidance (change `gh-240`): a host gets a re-injection mechanism only where a loss is verified.

Collected on 2026-10-27 from documentation and published source, without running a compaction. Labels: **doc** (official or installed documentation), **source** (published source code, not pinned to the installed version), **inferred** (reasoned from how the prompt is built), **undocumented** (no source settles it).

| Host | Installed | Global instruction file after compaction | Invoked skill body after compaction | Re-injection surface after compaction |
| --- | --- | --- | --- | --- |
| Claude Code | 2.1.312 | Project-root `CLAUDE.md` is re-injected from disk (doc). The user-level `~/.claude/CLAUDE.md` is undocumented; inferred to reload through `InstructionsLoaded` with `load_reason "compact"`. | Re-injected, first 5,000 tokens per skill, 25,000 in total, oldest dropped first (doc). | `SessionStart` with matcher `compact`; output over 10,000 characters is saved to a file and replaced by its path and a 2,000-character preview (doc). `PostCompact` does not inject. |
| Codex CLI | 0.165.0 | Re-injected: compaction rebuilds the initial context, including user instructions (source, `compact.rs`, `agents_md.rs` on `main`). Docs only say the chain is rebuilt on every run. | Undocumented. | `SessionStart` with `source: "compact"` runs before the next model request, for root sessions only, with `additionalContext` (about 2,500 tokens by default); unmanaged hooks need trust before they run (doc). `PreCompact` and `PostCompact` do not inject. |
| Grok Build | 1.0.72 | Undocumented. Rules reach the model as user rules; with `--rules`, the text goes to the system prompt (doc); its survival is inferred. | Undocumented. | None. `SessionStart` never fires on compaction; `PreCompact` and `PostCompact` drop their `systemMessage` (doc). |
| Pi | 1.4.0 | Survives by construction (inferred): the system prompt is built from base instructions and discovered context files, and compaction replaces only history messages (doc). | Undocumented. The body enters when the model reads `SKILL.md`; it stays in the kept recent tokens or is summarized (inferred). | Extension events `before_agent_start`, `context` and `context_with_system` run per turn; `session_before_compact` can replace the summary (doc). No post-compaction event found. |
| OpenCode | v2.2.0 | Survives: "After compaction, the current versions of your instruction files become the session's baseline"; the system prompt is not summarized (doc). | Not re-injected (source, dev branch): a `skill` tool result is treated like any tool result. | A plugin `context` hook runs on every request and can edit the system parts (doc); the `compaction` hook only summarizes. |
| Cursor CLI | 2026.10.18 | Undocumented. Rules go at the start of the model context; a user-level global file is not documented for the CLI. | Undocumented. | None after compaction: `sessionStart` fires when a conversation is created and `preCompact` is observational only (doc). |

## Sources

- Claude Code: https://code.claude.com/docs/en/context-window, https://code.claude.com/docs/en/memory, https://code.claude.com/docs/en/hooks, https://code.claude.com/docs/en/skills (read 2026-10-27).
- Codex: https://developers.openai.com/codex/hooks, https://developers.openai.com/codex/guides/agents-md; source through Context7 `/openai/codex` (`codex-rs/core/src/compact.rs`, `agents_md.rs`, branch `main`).
- Grok Build: installed guide `~/.grok/docs/user-guide/10-hooks.md` and `12-project-rules.md` (version 1.0.72).
- Pi: installed docs `compaction.md`, `how-pi-works.md`, `skills.md`, `extensions.md`, `configuration.md` (version 1.4.0).
- OpenCode: https://opencode.ai/docs/plugins; Context7 `/anomalyco/opencode` (`compaction.mdx`, `instructions.mdx`, `compaction.ts`, branch `dev`; installed v2.2.0 not confirmed against it).
- Cursor CLI: https://cursor.com/docs/hooks, https://cursor.com/docs/context/rules; Context7 `/websites/cursor_cli`.

## Open cells

No behavior was observed. These cells rest on no documentation and stay open until a compaction probe is authorized and run:

- Claude Code: the user-level `~/.claude/CLAUDE.md`.
- Codex: the invoked skill body; the global file rests on unpinned source.
- Grok Build: all three columns.
- Pi: explicit survival of the global file and of a skill body.
- OpenCode: the skill body rests on source only.
- Cursor CLI: all three columns.

## What it means for re-injection

- A host whose global file survives needs no re-injection of the always-loaded layer; the loss to cover is the `hive-core` skill body.
- Claude Code truncates a re-injected skill body to its first 5,000 tokens: what must survive goes at the start of `SKILL.md`, and a `compact` hook is the only documented way to restore more.
- Codex and Claude Code offer a post-compaction event; OpenCode and Pi only per-request or per-turn surfaces; Grok and Cursor none.
