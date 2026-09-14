# Harness mechanics — translation key for non-Claude harnesses

The flow skills are written against Claude Code's mechanics by name. They are also
published to the universal skill location (`~/.agents/skills/`), where Codex, opencode and Pi
read the SAME files; Grok reads them from `~/.claude/skills/` through its Claude-compat
layer, which is its lowest-precedence source. When you are NOT Claude Code, translate
mechanic names per this table instead of treating them as unknown tools. The skill's
*process* (phases, gates, contracts) is harness-neutral; only the machinery differs.

| When a flow skill says | Claude Code | Codex CLI | opencode | Grok CLI | Pi CLI |
|---|---|---|---|---|---|
| "dispatch agent X" / Agent tool | `Agent` tool with `subagent_type` | Spawn the custom agent `X` (defined in `~/.codex/agents/X.toml`) — spawning on a skill's instruction counts as the required explicit request | `task` tool with the subagent `X` (defined in `~/.config/opencode/agents/X.md`) | `spawn_subagent` with the agent name (defined in `~/.grok/agents/X.md`). No `capability_mode` (1.0.6+); tools come from the agent type. The `workflow` tool is parent-only — do not tell a child to launch one | `pi-subagents` or the configured child mechanism; use the role's generated Pi definition |
| "dispatch Explore" (read-only research) | Built-in `Explore` agent | General agent with `sandbox_mode = "read-only"` | General subagent with read-only `permission` denies | General subagent with the write tool (`search_replace`) withheld | A child with only the role's read-only tools |
| "the question tool" / decision gate | `AskUserQuestion` tool | Ask the options inline in chat and wait for the reply | Ask the options inline in chat and wait for the reply | Ask the options inline in chat and wait for the reply | Use the configured question skill/tool; ask inline when it is unavailable |
| "load X via ToolSearch" (MCP) | `ToolSearch` then call | MCP servers from `config.toml` are already available — call directly | MCP tools from `opencode.json` are already available — call directly | MCP servers are read from the Claude Code config via compat — call directly | Use the configured MCP extension/tool surface |
| "the task-list tool" / session task list | Task tools | Maintain the checklist in your plan/working notes | Todo tooling if available; otherwise plan notes | Todo tooling if available; otherwise plan notes | Maintain the checklist in the session or plan notes |
| "render via flow-report skill" | `flow-report` skill | Write the self-contained HTML directly (same single-file rules) | Write the self-contained HTML directly | `flow-report` skill, read from `~/.claude/skills/` via compat | Write the self-contained HTML directly (same single-file rules) |
| `$ARGUMENTS` / `$1` | Native substitution | No substitution in skills — interpret the free text after `$skill-name` as the arguments | Native in commands (`/flow-*` wrappers); skills invoked via command receive them in the prompt | Substitution unverified — interpret the free text after the invocation as the arguments | Text after `/skill:<name>` is passed as skill arguments |
| Portable Hive plan gate | `/flow-plan` writes the contract and records explicit approval; native Plan Mode (Shift+Tab) is optional drafting | `$flow-plan` writes the same contract and records explicit approval; `/plan` is optional drafting | Same `/flow-plan` contract and approval | Same `/flow-plan` contract and approval | `/skill:flow-plan` writes the same contract and records explicit approval; Pi has no Hive-specific plan mode |
| Engram memory saves | Engram MCP | Skip unless the Engram MCP is configured in this harness | Skip unless configured | Skip unless the Engram MCP reaches Grok through the Claude-compat config | Skip unless Engram is configured in the Pi surface |

The portable plan is the authority shared across harnesses. A native plan-mode transition or a
harness-specific approval is input to normalization at most; neither changes Hive's contract digest
nor grants implementation and delivery actions. In Codex, invoke the
shared skills as `$flow-plan` and `$flow-build`; in Pi, invoke them as `/skill:flow-plan` and
`/skill:flow-build`. Do not add a `hive-plan` command for this workflow.

## Enforcement differences to respect

- **Read-only reviewers** (sdd-spec-reviewer, sdd-product-critic, review-ux,
  workspace-custodian, review-code, review-security, sdd-explore): Claude
  Code enforces via tool allowlists; Codex via `sandbox_mode = "read-only"` in the agent
  TOML; opencode via `permission` denies; Grok by withholding `search_replace`, its only
  write tool. If your harness lost the enforcement in translation, honor it behaviorally —
  these agents propose, never mutate.
- **Phase skills are user-invoked gates.** Claude Code enforces with
  `disable-model-invocation`; Codex with `agents/openai.yaml` →
  `policy.allow_implicit_invocation: false`; opencode via `permission.skill` config
  (`"flow-*": "ask"`). Grok honors `disable-model-invocation` natively — the gate travels
  with the skill, no translation needed. Never self-trigger a flow skill because the
  conversation resembles its description.
- **Cross-references resolve on disk.** Paths like
  `~/.claude/skills/flow-core/references/...` exist on the same machine in every
  harness — read them directly; do not report them as Claude-only.
- **Grok's skill source is lowest-precedence.** It reads `~/.claude/skills/` below its own
  and repo-local skills, so a same-named Grok-native or project skill silently wins. Check
  `grok inspect` before concluding a flow skill is missing.
- **MCP schema cost is harness-level.** Codex and opencode load every configured MCP
  server's tool schemas into every context upfront (no deferred ToolSearch). Delegating
  research keeps the *result dumps* out of the orchestrator, but cannot reclaim the
  schema overhead — that is fixed by trimming the harness's MCP config for the project,
  not by the flow skill.
