# opencode adapter — flow pack

Thin layer that lets opencode run the flow pack. The content lives once (skills +
rubrics deploy to `~/.agents/skills/`); this folder holds only what opencode needs
natively.

## Pieces

| Piece | Source | Deploy target | Maintained how |
|---|---|---|---|
| Skills + rubrics | 18 skills, including `flow-research` | `~/.agents/skills/` | `/deploy-global` (copy); `flow-research` is eligible for automatic invocation |
| Subagents | `harness/opencode/agents/` (generated, versioned) | `~/.config/opencode/agents/` | **Generated** by `harness/build.py` from `global/agents/` — never edit |
| Command wrappers (every user-invoked skill, gated or not; model-invoked routers get none) | `harness/opencode/commands/` | `~/.config/opencode/commands/` | `/deploy-global` (copy) |
| SessionStart plugin (V2 API) | `global/hooks/flow-session-context/flow-session-context.ts` | `~/.config/opencode/plugins/` | `/deploy-global` writes it unless it positively detects a non-V2 opencode (V1, or a version it cannot read); **an absent opencode still gets the file**, inert until one exists. The report names the detected version |
| Config additions | `opencode.jsonc.snippet` | merge into `~/.config/opencode/opencode.json` | Manual, once |
| Permission keys + `shell` | `permission-config.json` | `permission.*` and a discovered, version-checked Bash 5+ `shell` when no user value exists | `/deploy-global` (surgical jq merge, idempotent; user values win; missing Bash only skips the optional shell key) |

## One-time setup

1. Merge `opencode.jsonc.snippet` into `~/.config/opencode/opencode.json`
   (skill gating + Linear MCP). The `permission` keys are NOT part of this step —
   `permission-config.json` is merged by the deploy on every run.
2. Run `/deploy-global` from the hub; start a fresh opencode session.

## What this harness loads

Every URL below was fetched and its quote extracted from the page body on the date in the
last column. Companion files: `global/README.md` (Claude Code),
`harness/codex/README.md`, `harness/grok/README.md` — all four share this layout.

> **The upstream repo moved orgs.** `github.com/sst/opencode` now redirects to
> `github.com/anomalyco/opencode`. Write `anomalyco`.

| Layer | What | Where it lands | Mechanism | Official doc | Verified |
|---|---|---|---|---|---|
| Always-on core | `harness/AGENTS.md` (condensed cross-harness core) | `~/.config/opencode/AGENTS.md` | Global instructions file, loaded before the per-directory chain | [v2 instructions](https://opencode.ai/v2/docs/instructions/) — *"OpenCode loads the global file followed by every `AGENTS.md` from the current workspace directory toward the home directory."* | 2026-09-20 |
| Skills | 18 skills with injected `references/`, including automatically eligible `flow-research` | `~/.agents/skills/` | Global compatibility skill source | [v2 skills](https://opencode.ai/v2/docs/skills/) — scope table: *"Global compatibility"* → *"`~/.claude/skills`, `~/.agents/skills`"* | 2026-09-20 |
| Commands | 10 wrappers, one per user-invoked skill | `~/.config/opencode/commands/` | Global command folder | [v2 commands](https://opencode.ai/v2/docs/commands/) — *"`~/.config/opencode/commands/` and project commands in `.opencode/commands/`."* | 2026-09-20 |
| Agents | 26 generated `.md` | `~/.config/opencode/agents/` | Global agent folder | [v2 agents](https://opencode.ai/v2/docs/agents/) — *"`~/.config/opencode/agents/<name>.md`"* | 2026-09-20 |
| Plugin | `flow-session-context.ts` (V2 API) | `~/.config/opencode/plugins/` | Global plugin folder; deployed unless a non-V2 opencode is positively detected | [v2 plugins](https://opencode.ai/v2/docs/plugins/) — *"Global plugins use the same discovery layout under the OpenCode config directory."* The page contrasts this with a `plugins/` folder beside a root `opencode.json`, which is NOT discovered; it does not use the word "automatic" for the global folder, so that half is **measured** — `opencode run --print-logs` logs `loading plugin id=~/.config/opencode/plugins/…` for files no config key references (2.0.9, 2026-09-20) | 2026-09-20 |
| Permissions | `permission-config.json` + the manual snippet | V1-shaped `permission.*` map in `~/.config/opencode/opencode.json` | **Deliberately V1-shaped**: V2 reads and normalizes it to a `permissions` array with the order preserved — measured with `opencode debug config` (2.0.9, 2026-09-20), `bash`→`shell` and `task`→`subagent`. An additive jq merge into an ordered, last-match-wins array is the harder thing to get right; what must not happen is BOTH shapes in one file | [v2 permissions](https://opencode.ai/v2/docs/permissions/) — *"The last matching rule wins, so the specific exceptions follow the broad rule."*; [v2 migration](https://opencode.ai/v2/docs/migrate-v1/) | 2026-09-20 |

The glob channel was retired in M3: it never loaded a rule in the deployed state (below),
and `harness/opencode/rules/` no longer exists. **Cleanup is manual** — the deploy manifest
never tracked `~/.config/opencode/rules/`, so those files are not orphan-cleaned: delete the
directory yourself. The `opencode-rules@0.6.4` plugin key is already gone from the live
config (`opencode debug config`, 2026-09-20, shows no plugin key at all); leftovers of that
era that remain on disk are `~/.config/opencode/rules/` and `~/.opencode/state/opencode-rules/`.

## What does NOT reach it

- **The 35 rule texts under `global/rules-situational/` that the core does not inline, as
  files.** There is no glob or command channel on this harness: every one of them arrives
  injected inside a router skill (`language-rules`, `workspace-conventions`, …), which the
  model must invoke, or inlined into a packed agent. opencode is the harness where not
  invoking the router is literally not having the rule — the `rule-delivery` hook has no
  channel here.
- **The 7 core-included rules, as files.** They are not deployed here at all; their
  condensed form is already in `harness/AGENTS.md`, which loads every session.
- **Claude Code's `disable-model-invocation`.** opencode does not honor it. What V2 reads for
  the same purpose is `metadata.opencode/autoinvoke: false` in SKILL.md frontmatter
  ([v2 skills](https://opencode.ai/v2/docs/skills/), 2026-09-20) — and the pack's skills do
  not carry it. The gate is therefore entirely `permission.skill."flow-*": "ask"` in
  `opencode.json` plus the command wrappers, which are the only way a gated skill is exposed.
  Confirmed present after normalization: `{"action":"skill","resource":"flow-*","effect":"ask"}`.

- **`~/.claude/CLAUDE.md`, on V2.** V1 read it; V2 does not. [v2 instructions](https://opencode.ai/v2/docs/instructions/)
  (2026-09-20): *"OpenCode V2 recognizes `AGENTS.md` only. It does not use `CLAUDE.md` as a
  fallback."* Nothing is lost — the deploy writes the condensed core to
  `~/.config/opencode/AGENTS.md`, which is the channel that was already doing the work — but
  any statement elsewhere that opencode picks up the Claude core for free is now wrong.

`~/.claude/skills/` IS still scanned on V2 (the scope table above). The V1-era measurement
this file used to cite — 2026-08-21, `opencode debug skill`: 33 pack skills resolving from
`~/.agents/skills`, zero duplicate names, `~/.claude/skills` absent from every location — is
**not reproducible on V2**, which removed that subcommand (`opencode debug` now offers only
`agents`, `config`, `paths`). Treat skill resolution here as documented-but-unverified.
**Caveat recorded in that V1 run and not re-checked since:** the scan is RECURSIVE, so a skill
directory that vendors its own `.codex/skills/` or nested `.agents/skills/` publishes those
too — `herdr` contributed three skills nobody chose.

## The plugin layer — fixed for the flow plugin, still broken for the rest

**History.** On opencode `1.18.18` (measured 2026-08-21) the *whole* plugin layer failed:
`~/.local/share/opencode/log/opencode.log` showed **160 load attempts, 160 failures**, every
plugin, with `SchemaError: Missing key at ["default"]`. A headless run in a ledger workspace,
asked whether its context carried the `flow-process-protocol` marker, answered **AUSENTE**.
That was the V1 plugin contract changing under V1.

**Now.** The same error is what OpenCode 2 raises for any V1-shaped plugin, and it is no
longer a bug to wait out — it is the contract. Measured on `2.0.9`, 2026-09-20, with
`opencode run --print-logs`:

```
WARN message="failed to load plugin" target=~/.config/opencode/plugins/engram.ts
  cause="PluginModule.LoadError: Plugin must export a default definition with an id
  and an effect or setup function. (cause: SchemaError(Missing key at [\"default\"]))"
```

| Plugin | Owner | State on V2 | What is lost |
|---|---|---|---|
| `flow-session-context.ts` | this repo | **ported** — V2 `export default {id, setup}` | — |
| `engram.ts` | Engram | fails to load | the plugin's automatic session hooks; Engram itself still works, it is also wired as an MCP server (`mcp.engram` in `opencode.json`) |
| `gk-hooks.js`, `herdr-agent-state.js` | third-party | fail to load | their own features — not ours to port |

A leftover V1 plugin is not inert: the file is discovered, refused, and logged as a WARN on
every boot. Delete the ones you no longer want.

### What the port had to work around

- **`experimental.chat.system.transform` is gone.** It was the V1 surface used to inject
  SessionStart context (opencode has no real SessionStart event), and it was never documented
  — read from source. The V2 equivalent is `ctx.session.hook("context", event => …)`, which
  IS documented ([v2 plugin API](https://opencode.ai/v2/docs/build/plugins/), 2026-09-20).
  `event.system` is an array of `{type:"text", text}` parts rather than V1's array of strings.
- **The documented import does not work for a local plugin file.** The docs show
  `import { Plugin } from "@opencode/plugin"` + `Plugin.define({…})`. A local file resolves
  imports against its own directory, and `~/.config/opencode/` has no node_modules for the
  `@opencode` scope — measured: `Cannot find package '@opencode/plugin'`, plugin refused. The
  port uses `import type` (erased at runtime) plus a plain `export default`, which loads
  cleanly. This is sound because `Plugin.define` is the identity function
  (`@opencode/plugin@2.0.11`, `dist/promise/plugin.js`: `return plugin`). Revisit if the
  deploy ever installs the package next to the file.
- **The per-session gate had to go.** V1 injected once per session id. In V2 `system` is
  rebuilt for every model invocation — the `context` hook runs once per invocation and
  opencode's own internal plugins push into `system` unconditionally each time — so a
  once-per-session gate does not prevent duplication, it prevents delivery past the first
  model call. The port injects on every invocation and caches only the TEXT (three
  synchronous git probes whose only input, the directory, is fixed per plugin instance).
- **Post-compaction recovery is moot, not impossible.** V2 does expose a dedicated
  `compaction` hook ([v2 plugin API](https://opencode.ai/v2/docs/build/plugins/), 2026-09-20,
  alongside `prompt`, `context`, `generate`, `title`, …). It is simply not needed once the
  block is re-injected every invocation: the first invocation after a compaction carries it
  again. The earlier wording here — that V2 offers no compaction signal — was wrong.

**Verification status.** Split deliberately, because the two halves have different evidence:

| Claim | How it is verified | Reproducible |
|---|---|---|
| The export shape loads on 2.0.9 (`import type` + `export default`; the runtime import does not resolve) | in vivo, `opencode run --print-logs` against a probe plugin | yes, see the commands below |
| The plugin's own behavior — both sections, push-not-append, every-invocation delivery, per-session text caching, marker guard, empty array, absent directory, never-throw | `global/hooks/flow-session-context/flow-session-context.test.ts`, 12 checks driving `setup()` with a fake ctx: `node --experimental-strip-types --test global/hooks/flow-session-context/flow-session-context.test.ts` | yes |
| `ctx.location.directory` is the SESSION's project directory | **not established.** The probe showed it populated with the directory opencode was started in; the docs describe `location` as where the plugin instance is loaded. The whole plugin (ledger walk + git probes) assumes they are the same. Falsify it by opening a session elsewhere and logging the value | no |
| A real session shows the injected text end to end | **not done.** `opencode run` hung on every attempt in this environment after the first probe; the hang was not traced | no |

## How to re-verify

```bash
opencode --version                        # the version gate keys on this
ls ~/.config/opencode/agents/  | wc -l    # expect 26
ls ~/.config/opencode/commands/| wc -l    # expect 10
opencode debug agents | jq 'length'       # agents actually resolved (26 + 7 built-ins)
opencode debug config  | jq '.[0].info.permissions'   # the normalized V2 array
opencode run --print-logs -m x/y x 2>&1 | grep -E 'loading plugin|failed to load'
```

Two ways to see the plugin state, both measured on 2.0.9, 2026-09-20:

- **`opencode plugin list`** lists local file plugins by path, and its **ID column is the
  load signal** — populated for a plugin that loaded, `-` for one that was refused. The four
  V1 plugins show `-`; the ported one shows `flow-session-context`. Caveat: the listing's
  scope varies by working directory in a way this file does not explain — from `/tmp` it
  showed the four global plugins, from the hub repo root only a package plugin.
- **`opencode run --print-logs`** (the last line above) shows the load attempt and the
  refusal reason. `opencode debug` has no plugin subcommand (only `agents`, `config`, `paths`).

## How it works

`/flow-build <args>` (opencode command) → instructs the agent to read
`harness-mechanics.md` (mechanic translation) + the skill body from
`~/.agents/skills/` → dispatches the generated subagents via the task tool.
Updating a skill or agent in the hub + `/deploy-global` updates this harness with no
adapter edits; only a renamed agent or a new skill requires touching this folder
(new command wrapper).

**Wrapper paths — no literal `~`.** Every `~/.agents/…` path in a wrapper is written
as a `` !`echo ~/…` `` shell injection, so opencode expands it to an absolute path at
prompt-build time. The Read tool does not expand tilde, and a model that passes the
literal `~` gets a failed read plus a recovery detour.

## Hive planning contract

`flow-plan` and `flow-build` use the shared plan artifact and read-only validator. Approval,
action/target scope and execution evidence travel with the plan; entering or leaving native
plan mode grants no Hive authority. During drafting, the parent's no-implementation boundary
is a prompt convention, not a universal write sandbox. `Session: no` retains conversation-only
operation without durable validation or cross-harness resume guarantees.

Source: `global/skills/flow-core/references/plan-format.md` (Hive convention, 2026-09-10).
