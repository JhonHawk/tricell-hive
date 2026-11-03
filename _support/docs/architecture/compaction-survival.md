# Skill loading and compaction, per host

This page records, for each supported host, how a skill reaches the model, how the global instruction file is framed, and what survives context compaction. It decides where a Hive rule must live to be followed: the instruction file is the only layer every host keeps after compaction, a skill body is context for the turn that loaded it, and a reference file is a tool result that is lost or read in part.

Collected on 2026-10-31 from documentation, published source (Codex, OpenCode, Pi), local transcripts of each host, and practitioner reports; it replaces the 2026-10-27 snapshot, whose Claude, Grok, Codex, OpenCode and Pi cells were wrong. Labels: **doc**, **source**, **obs** (local transcripts or probes), **forum**, **inf** (inferred).

## How a skill reaches the model

| Host (version) | Before invocation | Automatic invocation | Explicit invocation | Reference files |
| --- | --- | --- | --- | --- |
| Claude Code 2.1.316 | Name and description, 1,536 characters per skill, about 1% of the window for the whole list (doc) | `Skill` tool; the body follows in the same user turn as plain text that starts "Base directory for this skill" (obs) | `/name`, the same body (doc) | Read with the file tool; 7.5% of reads partial (obs) |
| Codex CLI 0.165.0 | Name, description and path in a developer message, 2% of the window (source) | The model reads `SKILL.md` through the shell; a tool result (source, obs) | `$name` or `/skills`; the body as a user-role fragment, **cut at 8,000 bytes** (source `MAX_SKILL_PROMPT_BYTES`) | Shell output cut in the middle when long; 100 of about 400 reads truncated, few re-read (obs) |
| Grok Build 1.0.73 | Name, description, "Use when" and path, in a synthetic user message (obs) | `read_file`; a tool result. `SKILL.md`, `AGENTS.md` and `CLAUDE.md` are always read whole (obs, tool definition) | `/name`; the body inlined in the user message after `<user_query>` as `<skill_information>` (obs) | `read_file`; models often pass `limit`, the host usually returns the whole file (obs) |
| Pi 1.4.0 | Name, description and location in the system prompt (source) | `read`; a tool result (doc, obs). The docs admit the model may fail to load a relevant skill | `/skill:name`; the body expanded in the user message (source) | `read` or bash (obs) |
| OpenCode v2.2.4 | Name and description in the system instructions (source) | `skill` tool; a tool result wrapped as `<skill_content>` (source) | `@name`; the same content in the user message (source) | Read tool, up to 2,000 lines or 50 KB (source) |
| Cursor CLI 2026.10.22 | Name, path and description sent as `agent_skills`; descriptions trimmed silently (src, forum) | `Read`; 55% of `SKILL.md` reads partial, usually 60–120 lines (obs, 64 sessions) | `/name`; the full body inlined in the user message as `<manually_attached_skills>` (obs) | `Read`; 59% partial (obs) |

Automatic selection often fails: in Vercel's Next.js evals the skill was not invoked in 56% of cases, and an explicit "use the skill" reached 79% (forum, 2026-01). Cursor staff call `/skill-name` "the only guaranteed trigger today" (forum 163761); Pi's docs say to use `/skill:name` to force it. An explicit invocation guarantees the load, not adherence (Codex issue #48877) or persistence.

## How the instruction file is framed

| Host | Global instruction file |
| --- | --- |
| Claude Code | A user message after the system prompt, as a reminder that these instructions override defaults (doc, obs) |
| Codex CLI | A user message, `# AGENTS.md instructions for <dir>`, below base instructions that say the user's instruction takes precedence over skills and to proceed unless a skill explicitly requires approval (source, obs) |
| Grok Build | `~/.claude/CLAUDE.md` arrives as `<user_rules>` "rules … you should follow **if appropriate**"; the project `AGENTS.md` as `<always_applied_workspace_rules>` "the agent **must always follow**" (obs, 85 sessions) |
| Pi | Global and project `AGENTS.md` in the same system-prompt section (source) |
| OpenCode | System instructions, combined with the agent prompt (doc) |
| Cursor CLI | Rules at the start of the context. `~/.cursor/AGENTS.md` is never loaded; `<dir>/.cursor/rules/*.mdc` loads for each folder from the workspace up to `/`, so `~/.cursor/rules/hive.mdc` with `alwaysApply` reaches workspaces under the home folder only (src, two probes 2026-10-31) |

No host documents that skill or reference text is treated as untrusted data; Meta's SkillInject study found the opposite, skills loaded as a trusted channel.

## What survives compaction

| Host | Instruction file | Invoked skill body | References |
| --- | --- | --- | --- |
| Claude Code | Re-read from disk (doc, obs) | Re-injected as `invoked_skills`, first 5,000 tokens per skill, 25,000 in total (doc, obs within one process) | Summarized; up to five recent files re-read (doc) |
| Codex CLI | Initial context rebuilt (source) | An explicit fragment is dropped from kept messages; a tool-result body is summarized (source) | Summarized |
| Grok Build | Rules message and skill list re-injected (obs, 71 compacted sessions) | **Never re-injected**; it survives only if the summary copies it (obs) | Summarized |
| Pi | System prompt rebuilt (doc) | Tool results cut to 2,000 characters for the summarizer, so a read-loaded body is mostly lost; a `/skill:name` body reaches the summarizer whole (doc, inf) | Same cut |
| OpenCode | Baseline rebuilt from current files (doc) | Skills stripped from kept messages; survives only in the ~15k-token verbatim tail (source) | Summarized |
| Cursor CLI | Rules sent with every request (inf) | Probably summarized (inf) | Probably summarized (inf) |

Corrections to the 2026-10-27 probe: on Claude the probe's skill survived only inside the summary, because `/compact` ran in a separate `--resume` process; inside one process the body is re-injected. On Grok the probe's tiny skill was copied by the summarizer, not re-injected by the host.

### Native task list after compaction

Collected on 2026-11-02 from documentation, source and local transcripts of the installed versions.

| Host · version | Tool and default | Kept after an automatic compaction |
| --- | --- | --- |
| Claude Code 2.1.316 | `TaskCreate`/`TaskUpdate`, one JSON file per task under `~/.claude/tasks/<session>/`; off by default on Opus 4.8, Sonnet 5 and newer models unless `CLAUDE_CODE_ENABLE_TODO_TOOLS=1` (doc, changelog 2.1.233) | Yes: a `task_reminder` attachment re-injects the open list after the `compact_boundary` (obs, five sessions; doc "Tasks persist across context compactions") |
| Grok Build 1.0.73 | `todo_write`; `plan.json` stays empty locally (obs) | Yes: a "## TODO List … still active" user message follows the compaction (obs) |
| Codex CLI 0.166.0 | `update_plan`, off by default since 0.152 (`tools.update_plan.enabled`); the handler stores nothing (source `plan.rs`) | No: the compacted history keeps no plan, and the model rebuilt it itself (obs, 0.145.0) |
| OpenCode v2.2.5 | `todowrite`, only with `experimental.todo: true`; the list is the latest call in the message log (source) | No re-injection: issue #5934 is still reported after its stale close, and PR #25982 was closed unmerged (source, inf) |
| Pi 1.4.0 | None built in; an example extension keeps state in tool-result details (doc) | Not applicable |
| Cursor CLI 2026.10.22 | `TodoWrite` through `CallDynamicTool` (obs) | Unknown: no documentation or transcript shows it either way |

## What it means for Hive

- A rule that must hold every turn and after compaction lives in the global guidance. On Grok, the guidance states its own authority because the host frames it as optional.
- Each flow skill keeps its binding steps in the first ~60 lines and stays under 8 KB, for Cursor's partial reads and Codex's explicit cut.
- References are read at a fixed moment, such as `plan-handoff.md` before the build route; a rule that only lives in a reference holds only while that read is in context.
- A native task list is a live view, not the record: four of six hosts do not re-inject it after a compaction. A plan kept in the conversation must pass the resume test of `flow-plan`, and its checkpoint line in the message carries the pending tasks, as `flow-build` sets.
- A flow boundary names the command for the next flow in the host's syntax (`/flow-x`, `$flow-x`, `/skill:flow-x`, `@flow-x`), so each flow starts with its skill loaded whole.

## Sources

- Claude Code: code.claude.com/docs/en/skills, /memory, /context-window; issues #82017, #85207, #83306, #95745.
- Codex: developers.openai.com/codex/skills; source at tag `rust-v0.165.0` (`ext/skills`, `user_instructions.rs`, `compact.rs`, `contextual_user_message.rs`); issues #56807, #48877, #34712, #34885.
- Grok Build: installed user guide 1.0.73 (`08-skills.md`, `12-project-rules.md`, `10-hooks.md`); docs.x.ai/build/features/skills-plugins-marketplaces; 188 local session folders.
- Pi: installed docs and `dist` 1.4.0 (`skills.md`, `compaction.md`, `core/skills.js`, `agent-session.js`); issues #11019, #9874.
- OpenCode: source at tag `v2.2.4` (`skill.ts`, `session/prompt.ts`, `compaction.ts`); docs `skills.mdx`, `instructions.mdx`; issues #10744, #46815.
- Cursor CLI: cursor.com/docs/context/skills, /context/rules, /cli/using; forum threads 163761, 161142, 163768; local transcripts and the installed bundle.
- Native task lists (2026-11-02): code.claude.com/docs/en/interactive-mode ("Task list") and /tools-reference, Claude Code CHANGELOG 2.1.233 and 2.1.311; openai/codex issue #42365 and `codex-rs/core/src/tools/handlers/plan.rs`; Grok installed user guide `17-sessions.md`; OpenCode `packages/core/src/tool/plugin/todo.ts`, issue #5934, PR #25982, opencode.ai/docs/tools; Pi `docs/extensions.md`; cursor.com/docs/cli/changelog.
- Evals and reports: vercel.com/blog/agents-md-outperforms-skills-in-our-agent-evals; Meta SkillInject summary; X posts by Codex, Claude Code and OpenCode staff (2026).
