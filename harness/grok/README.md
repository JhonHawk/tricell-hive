# Grok CLI adapter — what it loads and where that is documented

The thinnest adapter of the three. Grok reads Claude Code's own files through its compat
layer, so most of the pack reaches it with **no** generated tree: `~/.claude/CLAUDE.md`,
`~/.claude/skills/`, and `~/.claude/settings.json` are consumed directly. Only two things
need a Grok-specific shape — the rules (its scan cannot use `paths:`) and the agents (its
frontmatter differs).

Companion files: `global/README.md` (Claude Code), `harness/codex/README.md`,
`harness/opencode/README.md`. All four share the section layout below.

## Pieces

| Piece | Source | Deploy target | Maintained how |
|---|---|---|---|
| Always-on rules | `~/.claude/rules/**` (already deployed) | `~/.grok/rules/<dir>__<file>.md` | `/deploy-global --only grok` — one **flat file symlink** per rule |
| Subagents | `harness/grok/agents/` (generated, versioned) | `~/.grok/agents/` | **Generated** by `harness/build.py` from `global/agents/` — never edit. Spawn takes no `capability_mode` (removed 1.0.6; tools come from the agent type). The `workflow` tool is top-level only (1.0.8+) — generated children never receive it |
| Skills | `~/.claude/skills/` | *(read in place)* | Nothing to deploy — `[compat.claude] skills = true` |
| Global instructions | `~/.claude/CLAUDE.md` | *(read in place)* | Nothing to deploy |
| Hooks | `~/.claude/hooks/` + `~/.claude/settings.json` | *(read in place)* | Nothing to deploy; `bash-policy.sh`, `post-tool-hub.sh` and `flow-plan-capture.sh` are dual-runtime. Two Grok specifics: a matcher must name Grok's own tool name when no alias exists (`exit_plan_mode` has none), and stdout injection is event-specific — see the injection map below |

### Hook injection map — 2026-08-21 probe, qualified against grok 1.0.25

A probe hook registered on five events in `~/.grok/hooks/`, each emitting a distinct token as
`additionalContext`, then asked the model which tokens it could see (2026-08-21, older CLI).
The bundled guide (`~/.grok/docs/user-guide/10-hooks.md`, read at 1.0.25) is the documentary
source for `PreToolUse` and `PostToolUse` — the latter gained model-facing feedback in 1.0.14;
the probe was not re-run after 1.0.13.

| Event | Hook runs? | Model receives its `additionalContext`? |
|---|---|---|
| `SessionStart` | yes | **no** (guide: stdout ignored) |
| `UserPromptSubmit` | yes | **no** (guide: allowing stdout discarded) |
| `PreToolUse` | yes | **after the call** (1.0.13 guide); also `ask` / `deny` / `allow` / `defer` / `updatedInput`. 2026-08-21 probe on older CLI saw none |
| `PostToolUse` | yes | **YES, since 1.0.14** — `additionalContext`, `decision: "block"` + `reason`, and `updatedToolOutput` (replaces the model's copy of the result); all land after the tool result, in the same turn. Exit 2 now feeds stderr to the model — a hook must end with an explicit `exit 0` to stay silent |
| `Stop` / `SubagentStop` | yes | **YES** — and it keeps the turn working (capped at 8 continuations) |

A hook on a non-injecting event can still ACT (write a file, deny a call). Consequences for
this repo's hooks in Grok: `bash-policy` works (its product is a deny, not context);
`post-tool-hub` and `flow-plan-capture` inject since 1.0.14 (`post-tool-hub` reads Claude's
`.tool_response`, which grok emits as an alias of `toolResult`); `rule-context` lands as a
post-call note; `flow-session-context`, `session-hygiene-report` (`SessionStart`) and
`flow-context` (`UserPromptSubmit`) run but their injection never lands. **Policy still
belongs in the rules and skills layer** — a hook that speaks only after the call cannot
carry a rule the model needed before it; use hooks in Grok for effects, for post-call
feedback, and for PreToolUse `ask`/`deny`/`updatedInput` when a call must be gated.
`instructions-audit` is a separate case: its `InstructionsLoaded` event does not exist in
Grok at all.

There is deliberately **no** `harness/grok/skills/` tree: it would duplicate what
`~/.claude/skills/` already provides. Contrast with Codex, which needs a generated
`openai.yaml` to carry the same gate.

## What this harness loads

Every URL below was fetched and its quote extracted from the page body on the date in the
last column.

| Layer | What | Where it lands | Mechanism | Official doc | Verified |
|---|---|---|---|---|---|
| Always-on core | `global/CLAUDE.md` | `~/.claude/CLAUDE.md` (read in place) | Claude-compat instruction-file discovery | [docs.x.ai/build/features/skills-plugins-marketplaces](https://docs.x.ai/build/features/skills-plugins-marketplaces) — *"Grok automatically reads Claude Code marketplaces, plugins, skills, MCPs, agents, hooks, and instruction files"* | 2026-08-20 |
| Always-on rules | 12 flat symlinks `<dir>__<file>.md` | `~/.grok/rules/` | Every `*.md` in a rules dir is loaded regardless of name | [docs.x.ai/build/features/project-rules](https://docs.x.ai/build/features/project-rules) — *"every `*.md` file in a `.grok/rules/` directory"*; `.claude/rules/` also read for compat | 2026-08-20 |
| Path-scoped rules | **none** | — | No file-pattern scoping key exists — see *What does NOT reach it* | [docs.x.ai/build/features/project-rules](https://docs.x.ai/build/features/project-rules) (absence) | 2026-08-20 |
| Skills | 16 skills, incl. the router `references/` | `~/.claude/skills/` (read in place) | Claude-compat scan at the **lowest** precedence; `disable-model-invocation` honored natively | [docs.x.ai/build/features/skills-plugins-marketplaces](https://docs.x.ai/build/features/skills-plugins-marketplaces) — *"`disable-model-invocation`: Slash command only; no automatic invoke. Default `false`."* Bundled 1.0.25 `08-skills.md` still matches | 2026-09-09 |
| Agents | 25 Grok-shaped `.md` (real files, not symlinks) | `~/.grok/agents/` | `.md` agent definitions in `~/.grok/agents/` | [docs.x.ai/build/features/subagents](https://docs.x.ai/build/features/subagents); spawn args in bundled `16-subagents.md` (grok 1.0.25) — no `capability_mode` | 2026-09-09 |
| Hooks | shared with Claude Code | `~/.claude/settings.json` (merged by compat) | Grok hooks + Claude settings compat; matcher aliases cover only common Claude tool names; stdout injection is event-specific (map above) | [docs.x.ai/build/features/hooks](https://docs.x.ai/build/features/hooks); alias list, `ask`/`defer`, and per-event stdout rules (PostToolUse feedback since 1.0.14) from bundled `~/.grok/docs/user-guide/10-hooks.md` (grok 1.0.25) | 2026-09-09 |
| Scope & precedence table | — | — | Which dirs are scanned, in what order | bundled doc `~/.grok/docs/user-guide/12-project-rules.md` + `08-skills.md` (grok 1.0.25) — local file, not a URL | 2026-09-09 |

**Dead URLs — never write these** (all verified 404 on 2026-08-20):
`github.com/xai-org/grok-cli` · `docs.x.ai/docs/grok-cli` ·
`docs.x.ai/build/project-rules` (the real path carries `/features/`) ·
`github.com/xai-org/grok-build/blob/main/docs/user-guide/...` — that repo is public but is
the Rust harness and has **no `docs/` directory**; the user guide exists only on disk.

## What does NOT reach it

- **The 19 path-scoped rules.** Grok has no `paths:`-equivalent, so shipping them into
  `~/.grok/rules/` would load all of them, always. They are excluded on purpose and reach
  Grok the same way they reach Codex: injected into a router skill's `references/`
  (`language-rules`, `workspace-conventions`, `task-routing`, …) plus the advisory
  `rule-context` hook. **Consequence:** giving an existing always-on rule a `paths:` key
  silently removes it from Grok — route that content through a router skill in the same
  change.
- **Directory symlinks.** `~/.grok/rules/` is populated with one symlink **per file**, not
  a single link to `~/.claude/rules/`. See the recursion caveat below.
- **Nothing yet deployed to `~/.claude/rules/`.** The grok scope skips a missing rule with
  a WARNING rather than leaving a dangling link — run `--only claude` at least once first.

## Unverified / undocumented dependencies

Two claims this adapter is built on that **no official source states positively**. Both are
recorded here rather than dressed up as citations:

- **"The rules scan is not recursive."** This is the entire reason for the flat
  `<dir>__<file>.md` naming. Neither the public docs nor the bundled user guide say
  whether the scan descends into subdirectories. Compare Claude Code, whose docs are
  explicit that `.md` files are discovered recursively. **Status: assumed, unconfirmed.**
  Settle it empirically — drop a `.md` into a subdirectory of `~/.grok/rules/`, run
  `grok inspect`, and see whether it is listed.
- **"Grok ignores `paths:`."** Supported as *documentary absence*: no source mentions any
  file-pattern scoping key at all (`paths:`, `globs:`, or otherwise); the only documented
  scoping is per-directory. Reasonable, but weaker than the flat assertion sounds.

## How to re-verify

```bash
grok inspect          # lists every config source, rules file, skill, plugin, hook and MCP
                      # server Grok discovered here — the deterministic auditor,
                      # analogous to `codex debug prompt-input`

ls -la ~/.grok/rules/ # should be 12 symlinks named <dir>__<file>.md → ~/.claude/rules/...
ls    ~/.grok/agents/ # should be 25 real .md files
```

The bundled user guide under `~/.grok/docs/user-guide/` ships with the CLI and is version-
stamped: check `grok --version` before trusting a quote from it.
