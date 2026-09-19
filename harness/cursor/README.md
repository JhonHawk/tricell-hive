# Cursor adapter — what it loads and where that is documented

The only harness with **no generated tree and no deploy step**. Cursor reads three of the
four Claude Code targets directly through its third-party compatibility layer, and reads
`~/.agents/skills` as a first-class native path — so the whole pack arrives without an
adapter. This folder holds documentation only.

"Cursor Desktop" is not a separate product: it is the desktop app. The surfaces are
Desktop, CLI (`cursor-agent`) and Web/Cloud Agents, all over one `~/.cursor/`.

Companion files: `global/README.md` (Claude Code), `harness/codex/README.md`,
`harness/opencode/README.md`, `harness/grok/README.md`.

## Pieces

| Piece | Source | How it arrives | Maintained how |
|---|---|---|---|
| Skills | `global/skills/` → `~/.agents/skills/` | Native path, scanned directly | `/deploy-global` (shared with Codex/opencode) |
| Agents | `global/agents/` → `~/.claude/agents/` | Third-party compat, Desktop only — the CLI loads project-level agents, never user-level ones | `/deploy-global` (claude scope) |
| Hooks | `global/hooks/` → `~/.claude/settings.json` | Third-party compat, read and merged | `/deploy-global` (claude scope) |
| Rules | — | **Nothing reaches it** | — |

## What this harness loads

CLI column for Skills, Agents and Rules re-verified 2026-09-18 by driving
`cursor-agent 2026.09.15` headless against an isolated lab (canary subagents that return a
token on dispatch). The Desktop column and the hook payload were last measured 2026-08-21
with `cursor-agent 2026.08.11`. Measured, not inferred from docs.

| Layer | Cursor CLI | Cursor Desktop | Mechanism | Official doc |
|---|---|---|---|---|
| Skills | **yes** | **yes** | `~/.agents/skills` is a documented native path; `.claude/skills` is also scanned as legacy-compat | [cursor.com/docs/skills](https://cursor.com/docs/skills) |
| Agents — project level | **yes** | **yes** | `.cursor/agents/` and `.claude/agents/`, subdirectories included; dispatched through the CLI's Task tool | [cursor.com/docs/subagents](https://cursor.com/docs/subagents) |
| Agents — user level | **no** | **yes** | `~/.claude/agents/` read via compat; `.cursor/` wins name conflicts. The CLI loads neither it nor the native `~/.cursor/agents/` — see below | [cursor.com/docs/subagents](https://cursor.com/docs/subagents) |
| Hooks | **yes** | **yes** | `~/.claude/settings.json` hooks are read and merged; 8 events map 1:1 | [reference/third-party-hooks](https://cursor.com/docs/reference/third-party-hooks) |
| Rules | **no** | **no** | `~/.claude/rules/` is in no scan list — and nothing is deployed there any more — see below | [cursor.com/docs/rules](https://cursor.com/docs/rules) |
| `AGENTS.md` / `CLAUDE.md` | **yes** | **yes** | Project root only, nested subdirectories supported | [cursor.com/docs/rules](https://cursor.com/docs/rules) |

**Better skill parity than Codex or opencode:** Cursor honors `disable-model-invocation`
natively, so the `flow-*` gate travels with no translation — no generated `openai.yaml`
(Codex) and no command wrapper (opencode); the 2026-09-18 run lists the hive skills and
none of the gated ones. It also honors `paths:` in `SKILL.md` frontmatter.
`user-invocable` is **undocumented** as of 2026-09-18: the skills page no longer mentions it.

**The CLI diverges from the subagents doc at user level.** The page says *"You can use
subagents in the editor, CLI, and Cloud Agents"* and lists `~/.cursor/agents/` and
`~/.claude/agents/` as user-level paths, yet `cursor-agent 2026.09.15` lists no hive agent
and no canary placed in `~/.cursor/agents/` (flat or nested), while project-level canaries
in both directories are listed and dispatch. Real hive agents copied into a project's
`.claude/agents/` load with their frontmatter untouched — so the roster reaches the CLI
per project, never through the global deploy. Untested: a flat file in `~/.claude/agents/`.

**The compat layer is OFF by default** — Cursor Settings → Agents → Third-Party Imports →
*"Include Third-Party Plugins, Skills, and Other Configs"* (menu path per the docs,
2026-09-18). Without it, nothing under `~/.claude/` is read. Skills arrive regardless,
because `~/.agents/skills` is native rather than compat. A CLI equivalent of the toggle is
**undocumented**.

## What does NOT reach it

- **Every rule.** `~/.claude/rules/` appears in no scan list — and this repo deploys
  nothing there any more — so none of the 41 texts under `global/rules-situational/` reaches
  it as a rule file, the 7 the Claude core inlines included (Cursor does not read
  `~/.claude/CLAUDE.md` either). Cursor's own rules are `.cursor/rules/*.mdc` per project
  (`alwaysApply:` ≡ always-on, `globs:` ≡ conditional), project `AGENTS.md`, and User Rules
  as a Settings text field. A **global rules directory** is not documented;
  `~/.cursor/rules/` exists on this machine and appears to work, but that is unverified and
  there are open upstream feature requests asking for exactly it.
- **`global/CLAUDE.md`.** Cursor reads `CLAUDE.md` at the PROJECT root, never `~/.claude/CLAUDE.md`.
- **The per-agent tool allowlist.** Cursor's subagent frontmatter is `name`, `description`,
  `model`, `readonly`, `is_background` — no `tools`/`disallowedTools`. The seven cyan
  reviewers, whose read-only is *enforced* by the `tools:` allowlist, arrive unconstrained;
  `readonly: true` is the only lever and it is wrong for agents that must run commands
  (`sdd-verify` drives `agent-browser` over Bash). Their never-mutate prompt clause is
  what holds, behaviorally.
- **Every hive hook maps to a Cursor event** since `instructions-audit` (`InstructionsLoaded`,
  which has no Cursor equivalent) was retired on 2026-09-13.

## Payload shape — why the shell gate needed a fix

Cursor is a THIRD runtime shape, and `bash-policy` matched only two:

```
Cursor   tool_name: "Shell"      hook_event_name: "preToolUse"   cwd: ""   workspace_roots: [...]
Claude   tool_name: "Bash"       hook_event_name: "PreToolUse"   cwd: "/real/path"
Grok     toolName: "run_terminal_command"
```

Three differences beyond the tool name: camelCase event, **`cwd` as an empty string** (jq's
`//` only falls through on `null`, so a naive fallback does not fire — the real path is in
`workspace_roots[]`), and no `permission_mode` / `effort`. Fixed in `bash-policy.sh` and
`post-tool-hub.sh`; verified red→green against the captured payload.

**Hooks do not double-fire:** one shell call produced exactly one `preToolUse` and one
`postToolUse`.

## How to re-verify

```bash
cursor-agent status                        # auth
cursor-agent -p --trust --force "..."      # headless; without --trust it aborts on workspace trust
cursor-agent about                         # version, tier, model
```

`cursor-agent` has no equivalent of `opencode debug skill` or `grok inspect` — there is no
deterministic listing of what it resolved. The only route is asking a headless run what it
sees, which is a model claim, not a resolved manifest. Weigh findings accordingly.

**Name collision:** the installer creates BOTH `cursor-agent` and `agent` in `~/.local/bin`,
and `agent` collides with Grok's own (`~/.grok/bin/agent`), which wins on PATH order here.
Always invoke `cursor-agent`.

## Hive planning contract

`flow-plan` and `flow-build` use the shared plan artifact and read-only validator. Approval,
action/target scope and execution evidence travel with the plan; entering or leaving native
plan mode grants no Hive authority. During drafting, the parent's no-implementation boundary
is a prompt convention, not a universal write sandbox. `Session: no` retains conversation-only
operation without durable validation or cross-harness resume guarantees.

Source: `global/skills/flow-core/references/plan-format.md` (Hive convention, 2026-09-10).
