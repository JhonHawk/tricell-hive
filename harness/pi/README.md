# Pi adapter

This directory contains the Hive integration for [Pi](https://pi.dev/), the
terminal coding harness. It keeps the canonical agent definitions and Bash
hooks shared with Claude Code, Codex, and Grok while adding the Pi-specific
runtime needed for subagents, portable Flow planning, bounded Git reads, and
package-backed tools.

The source and runtime have one boundary:

```text
global/agents + global/hooks
          │ generated roles / canonical hook scripts
          ▼
  harness/pi runtime ──► ~/.pi/agent
          │                 ├─ extensions: Hive bridge + reviewer guard
          │                 ├─ agents: 25 generated role definitions
          │                 ├─ settings: managed package/resource entries
          │                 └─ packages: pinned third-party capabilities
          ▼
      Pi parent ──► pi-subagents background children
                         ├─ fresh context by default
                         ├─ inherited project/global instructions and skills
                         └─ no specialist-to-specialist delegation
```

The default installation root is `~/.pi/agent`. Set `PI_CODING_AGENT_DIR` when
the managed Pi state must live elsewhere. Deployment is additive and owns only
the Hive files and package entries it declares; credentials, sessions, theme,
model defaults, trust decisions, and unrelated packages remain operator-owned.

`/deploy-global` includes Pi in its default, `all`, and `harness` selections, and
comma-separated harness selectors may be mixed. `--only pi` is the isolated write
boundary: it preflights Pi and the neutral shared-skills surface, writes only those
targets, and leaves Claude, Codex, opencode, and Grok installations untouched.

## Deployment surface

The deployment contract separates bytes copied into the Pi root from the registrations
that make them active, and from the policy used when a prerequisite is missing:

| Surface | Copied or staged | Wired or registered | Failure policy |
|---|---|---|---|
| Core and roles | `harness/AGENTS.md` and the 25 generated files under `harness/pi/agents/` | Pi agent discovery and `pi-subagents` role selection | Missing, stale, or hand-edited generated output fails the selected preflight |
| Runtime and extensions | `harness/pi/src/`, `extensions/`, and the canonical hook scripts | Managed extension/package entries under the PI root | Missing or non-executable required files block; advisory extension failures warn and continue |
| Managed configuration | Owned fields in `settings.json`, `extensions/subagent/config.json`, `mcp.json`, and `web-search.json` | Five exact package entries, `forceTopLevelAsync`, the managed Context7, Linear, and HeroUI Pro `mcp` proxies, and OpenAI web search | Invalid config, ownership conflict, or missing required pin blocks before a selected write; unrelated user fields remain intact |
| Packages and patch | Preinstalled packages under `<PI root>/npm` plus the reviewed patch metadata in this checkout | Exact package identities, versions, and patch target hashes | No auto-install; a missing package, third hash, symlink, or patch mismatch blocks before writes |
| Shared skills | Generated universal skills through the neutral shared-skills target | Shared manifest and neutral backup/rollback engine | Legacy conflicts remain untouched and are reported with hash/source-commit provenance |

The neutral deploy engine records the Pi and shared-skills manifests and rollback backups. A
legacy path-only manifest is read-only migration evidence: matching current bytes, or matching
blobs at its recorded source commit, may be adopted; unknown differences are preserved as
conflicts. Other selected harnesses retain their legacy writers and backup paths.

## What Pi receives

`python3 harness/build.py` generates the Pi role files alongside the Codex,
opencode, and Grok outputs. The generated files are derived from
`global/agents/**/*.md`; edit the canonical agent, then rebuild. The generator
emits the Pi model, thinking level, context inheritance, tool selection, and
extension fields that Pi can enforce. A `__HIVE_PI_ROOT__` placeholder is
resolved only while deploying, so generated files do not contain a machine
specific home path. MCP access uses the native `mcp` proxy in the generated
tool allowlists.

[`src/mcp-allowlist.json`](src/mcp-allowlist.json) is the single source of the
cut, applied in two layers: the runtime guard in `src/hooks.ts` validates every
`mcp` call, and `deploy.py` writes the same tool names as `includeTools` on each
managed `mcp.json` server entry. Proxy tool names are `<server>_<tool>`:

```text
{ tool: "context7_query-docs", args: { libraryId, query } }
{ tool: "linear_list_issues", args: { team: "FAC", limit: 20 } }
{ tool: "heroui_get_component_docs", args: { component: "Button" } }
```

Context7 tools keep their exact required string arguments. Linear and HeroUI
tools accept JSON-serializable plain arguments up to depth 4 and 16 KiB.
Everything outside the allowlist — other servers included — is blocked before the
adapter runs, each rejection naming its own cause. Linear writes are limited to
`save_issue` and `save_comment`, and the guard refuses them in child sessions
(`PI_SUBAGENT_CHILD=1`): they run from the parent session only, with no approval
dialog. An unreadable allowlist fails closed — every `mcp` call is refused and the
parent is warned once per session. The guidance above reaches only roles whose
active tools include `mcp`.

The source WebFetch grant expands to Pi's
fetch_content/get_search_content retrieval pair. WebSearch expands to that
pair plus web_search/source_check for the full web profile. Source
disallowedTools entries expand to the same Pi names and remain authoritative
through the generated excludeTools field.

The 25 roles keep the Hive roster and prompts. Their OpenAI-backed model policy
is:

| Canonical source policy | Generated Pi `model` | Thinking level | Roles |
|---|---|---|---|
| `opus` + declared `high` | `openai-codex/gpt-6-astra` | `medium` | Five approved Astra roles; the source `high` declaration remains visible for auditability |
| `sonnet` + declared `max` | `openai-codex/gpt-5.6-luna` | `max` | 11 executor and discovery roles |
| `sonnet` + declared `high` | `openai-codex/gpt-5.6-luna` | `high` | Three validator roles |
| `sonnet` + declared `medium` | `openai-codex/gpt-5.6-luna` | `medium` | `state-fetcher` |
| `inherit` + declared `high` | `inherit` | `high` | Five judgment roles; the generated definition keeps `model: inherit` |

The source frontmatter remains unchanged for Claude Code and Grok. The mapping
above is a Codex/Pi generator policy, so changing it does not alter the other
installed harnesses. The Astra mapping applies to exactly five roles:
`code-reviewer`, `product-critic`, `database-specialist`,
`performance-engineer`, and `prompt-engineer`.

## Runtime responsibilities

The local TypeScript runtime under `src/` supplies the pieces that Pi does not
provide as a single native Hive surface:

- `hook-runner.ts` invokes canonical Bash hooks with `spawn` and an argument
  array, passes JSON on stdin, enforces the existing 10-second timeout, and
  classifies blocking failures as fail-closed and advisory failures as
  warnings.
- `git-read.ts` exposes `hive_git_read`, a bounded read-only tool for status,
  diff, log, show, blame, file listing, branches, and remotes. It rejects
  unsafe arguments, disables external diff/pager/fsmonitor behavior, caps
  output, and never accepts a shell command. In a multi-repo workspace the
  optional `repo` parameter selects the repository directory (resolved against
  the session cwd and refused outside it); `path` is always a pathspec inside
  that repository. The general extension registers it for every Hive role,
  including reviewers.
- `research.ts` exposes `hive_research_readiness`, which inspects the active
  tools in the current Pi agent. The `documentation` profile requires
  `fetch_content` and `get_search_content`; the `web` profile additionally
  requires `web_search` and `source_check`. It reports `available`,
  `missing`, and `ready` without blocking local work when a research tool
  is unavailable.
- `hive-status` reports the current hook readiness, per-hook status, and active
  tools, including both research profiles. `hive_hook_readiness` remains the
  structured hook check used by generated roles.
- Flow planning and execution use the shared `/flow-plan` and `/flow-build`
  skills. Pi has no Hive-specific plan mode, plan approval state, or plan
  capture adapter; the workflow artifacts and their approvals remain
  harness-independent.

The general extension is loaded for every Hive role. The reviewer guard is
loaded only for the six canonical roles whose source agent declares the
reviewer hook: `code-reviewer`, `code-scout`, `product-critic`,
`security-reviewer`, `spec-quality-reviewer`, and `workspace-custodian`.
Role identity comes from the generated agent definition; it is never inferred
from a prompt or environment variable. A missing required child extension is
an infrastructure error, not a reason to silently continue without the guard.
The generated `tools` allowlists carry the readiness sentinels:
`hive_hook_readiness` is selected by all 25 roles, and research-capable roles
also select `hive_research_readiness`; the six reviewer roles additionally
select `hive_reviewer_readiness`. The corresponding child extensions are
`hive-hooks.ts` for every role and `hive/reviewer-guard.ts` for the six-role
reviewer set. `pi-subagents` 0.67.0 derives the required child-tool checks
from the explicit `tools` selection.

### Parent advisory hooks

The parent extension delivers the three canonical advisory surfaces without changing
the existing child guards:

| Advisory | Parent trigger | Durable behavior | Failure policy |
|---|---|---|---|
| `flow-session-context` | Fresh context and successful compaction | Fresh flow context is restored once; compaction receives the condensed Flow recovery; pending initial context survives reload | Missing or malformed advisory output warns and continues |
| `rule-context` | Before a tool invocation | Translates Pi `Write`/`Edit`/`Bash` inputs to the canonical JSON payload and keeps canonical per-session rule markers | Advisory only; it never replaces the blocking Bash or reviewer guards |
| `session-hygiene-report` | Fresh parent context | Reuses the canonical report-only scan and its six-hour fingerprint cooldown | Missing report warns and continues; it never kills a process or blocks the parent |

Freshness comes from active context entries, not from treating every startup as a new
session. A resumed CLI context may therefore emit the fresh advisory when its active
context is absent. The Pi adapter queues the initial advisory as a native hidden
`custom_message` with `triggerTurn: false` and no `deliverAs`; native session entries
carry it through an extension reload and deliver it once before model work. Pi 0.85.1's
`SessionManager` does not create the session file until the first assistant response, so
disk durability cannot be promised before that response. Children receive no new advisory
injection. Advisory invocations use the canonical JSON stdin contract, serialized calls,
and the bounded 10-second runner.

Children run in the background through `pi-subagents`, which keeps the parent
responsive and exposes lifecycle state. The default context is fresh, with
Hive project/global instructions and shared skills explicitly inherited. Child
agents do not recursively delegate. The parent session pointer
`PI_SUBAGENT_PARENT_SESSION` is used for child lifecycle association; it is not
a trusted role identity.

## Pinned packages and external surfaces

| Package or service | Pin / endpoint | Purpose | Loading status |
|---|---|---|---|
| `pi-subagents` | `0.67.0` | Background children, role discovery, lifecycle, and bounded orchestration | Required |
| `gentle-engram` | `0.1.12` | Native Engram HTTP memory integration | Required; no `pi-engram init` |
| `pi-mcp-adapter` | `2.33.0` | MCP transport for Context7, Linear, and HeroUI Pro | Required |
| `@juicesharp/rpiv-ask-user-question` | `2.9.0` | Structured user questions in the parent session | Required; children use `contact_supervisor` |
| `pi-web-access` | `0.29.0` | OpenAI-backed web search and source access | Required for research roles |
| Context7 | `https://mcp.context7.com/mcp` | Version-anchored library documentation through the native `mcp` proxy | Allowlisted `context7_resolve-library-id` and `context7_query-docs` |
| Linear | `https://mcp.linear.app/mcp` | Tracker reads for ledger-declared `Tracker access: mcp`, plus approved ticket and comment writes | Allowlisted read tools + `save_issue`/`save_comment` (parent session only); `auth: "oauth"` |
| HeroUI Pro | `https://mcp.heroui.pro/mcp` | Component, CSS, theme, and design-system documentation for roles whose tool allowlist includes `mcp` | Allowlisted read tools; token from `HEROUI_PERSONAL_TOKEN` in the environment that launches `pi` |

`HEROUI_PERSONAL_TOKEN` has no fallback: the adapter interpolates `${VAR}` in
`headers` and substitutes an empty string when the variable is unset — only `url`
raises on a missing variable (`utils.ts`: `interpolateEnvRecord` versus
`resolveServerUrl`), so an unset token yields an empty header and authentication
fails on first use. The deploy plan warns for any unset `${VAR}` in an allowlist
header; it never blocks.
| Web | provider `openai`, search provider `openai-codex` | External research | No workflow provider |

For `pi-mcp-adapter` 2.33.0, every managed entry sets `directTools: false`,
sets `includeTools` to its allowlisted tools, and uses `lifecycle: "lazy"`.
Context7 advertises `ttlMs: 0`, so the adapter discards its cache and direct or
namespace tools do not remain registered; an eager lifecycle does not change
that. The HeroUI entry stores the literal `${HEROUI_PERSONAL_TOKEN}` string in
`headers`, which the adapter interpolates from the environment at connect time —
no token is ever written to a file. Linear authenticates through the adapter's
own OAuth flow: a browser authorization on first use, independent of Claude
Code's Linear token. The managed block also sets `settings.scriptMode: false`,
which hides the `mcpScript` tool that no Hive guard can validate in the parent
session. The native `mcp` proxy and its input guard are the stable bounded
surface.

### `pi-subagents` 0.67.0 compatibility patch

`pi-subagents` 0.67.0 needs the reviewed compatibility patch in
[`patches/pi-subagents-0.67.0.patch`](patches/pi-subagents-0.67.0.patch), with
its exact metadata in
[`patches/pi-subagents-0.67.0.json`](patches/pi-subagents-0.67.0.json). It
corrects two upstream boundaries used by Hive:

- Custom extension tools were classified against the built-in-only inventory
  and silently pruned. The patch preserves explicitly requested custom tools
  and diagnoses only unavailable built-ins.
- Required child-tool validation thrown from `agent_start` could be swallowed
  or arrive too late. The patch handles the check on `input`, before the
  provider request, and stops the input when a required tool is missing.

The bundle hash is
`fb534c3f1c9acd8d2c8c7e83459bf2d0115747264381b76e577966967312f3d1`. The
target hashes below are SHA-256 values for files under the installed package
root (`<PI root>/npm/node_modules/pi-subagents/`):

| Package source path | Before patch | After patch |
|---|---|---|
| `src/runs/shared/child-tool-plan.ts` | `90a8135c4afa87e56ff739acea8fa83b151e56323ea8eea167d832dfd70c1ddb` | `39c081aab64fb2617703043bfdb9184f2a38c7a48f7293e34649754c6c8c432b` |
| `src/runs/shared/subagent-prompt-runtime.ts` | `ab6b6176dc55fa330bd8d0174bce143fe68203b05d50a10809102c629e58a641` | `d46f26e9cc1b01a58c64ab6b32e7cc720877f50e30902bf92dbad5e7bc8330d1` |

The PI deploy helper validates the exact package name and version, patch
bytes, and every before/after hash before it plans a write. Each target must
be pristine or already patched; a third hash, a symlink, or a package outside
the PI staging root stops the deploy. The helper applies the patch inside the
PI root and records it in the PI manifest; it does not modify a global npm
installation.

Pi packages can execute code with the privileges of the local process. Review
the pinned package sources before changing a pin, then run the repository's
dependency and type checks. The integration deliberately does not enable an
OS sandbox or a peer-to-peer team runtime; its safety boundary is Pi's tool
surface, the role restrictions, canonical hooks, and the review gate.

## Loading matrix

The following links are the upstream sources consulted for the installed Pi
0.85.1 integration on 2026-09-10. Package-specific controls are marked when
the Pi core documentation does not define them.

| Surface | Hive behavior | Upstream source | Status |
|---|---|---|---|
| Global settings | Managed entries merge into `~/.pi/agent/settings.json`; project settings override global settings | [Pi settings](https://pi.dev/docs/latest/settings) | Official |
| Packages | Pinned npm packages are installed and referenced from settings | [Pi packages](https://pi.dev/docs/latest/packages) | Official |
| Extensions | Hive extensions load from the Pi agent directory and package resources | [Pi extensions](https://pi.dev/docs/latest/extensions) | Official |
| Shared skills | Pi discovers `~/.agents/skills/` and project `.agents/skills/` | [Pi skills](https://pi.dev/docs/latest/skills) | Official |
| Custom agents | Generated files load from `~/.pi/agent/agents/**/*.md` | [pi-subagents agents reference](https://github.com/nicobailon/pi-subagents/blob/main/docs/agents.md) | Package reference |
| Background children | Hive uses `pi-subagents` async children with explicit child-only extensions | [pi-subagents package](https://pi.dev/packages/pi-subagents?type=extension) · [tool and extension selection](https://github.com/nicobailon/pi-subagents/blob/main/docs/agents.md) | Package reference |
| Engram memory bridge | `gentle-engram@0.1.12` provides Pi-native memory tools and the native HTTP path used here | [gentle-engram package](https://pi.dev/packages/gentle-engram) | Package reference; this integration does not run `pi-engram init` |
| Required child extensions | `hive-hooks.ts` on all 25 roles; `hive/reviewer-guard.ts` on six roles; explicit `tools` sentinels `hive_hook_readiness` and `hive_reviewer_readiness` | — | Undocumented upstream; enforced by generated tool selection and runtime checks |
| Portable Flow planning | Shared `/flow-plan` and `/flow-build` skills own phases, artifacts, and approvals; Pi contributes no native plan mode | — | Hive workflow |
| Canonical Bash hooks | Hive-specific JSON adapter around existing scripts | [Pi extensions](https://pi.dev/docs/latest/extensions) for lifecycle extension points | Bridge behavior is Hive-specific |
| OS sandbox | Not enabled | [Pi containerization](https://pi.dev/docs/latest/containerization) | Explicitly excluded |

The package reference is version-sensitive. If the installed package changes,
re-check its agent frontmatter and child-extension behavior before changing the
generator or deployment contract.

## Build, test, and deploy

From the repository root:

```sh
python3 harness/build.py
cd harness/pi
pnpm typecheck
pnpm test
```

The selected deployment order is deliberate:

1. Install all five exact packages into the target PI root before invoking the helper:
   `pi-subagents@0.67.0`, `gentle-engram@0.1.12`, `pi-mcp-adapter@2.33.0`,
   `@juicesharp/rpiv-ask-user-question@2.9.0`, and `pi-web-access@0.29.0`.
   For a non-default root, pass the same `PI_CODING_AGENT_DIR` to Pi's package
   manager; for example:

   ```sh
   PI_CODING_AGENT_DIR="${PI_CODING_AGENT_DIR:-$HOME/.pi/agent}" \
     pi install npm:pi-subagents@0.67.0
   ```

   Repeat the native package-manager operation for the other four pins and stage
   the reviewed patch before activation. The helper never installs packages or
   runs OSV auditing; it fails closed when any exact pin, package identity, or
   patch hash is missing.
2. Run the generator, native `shasum` preflight, type checks, package checks,
   and a read-only dry run. The dry run is `/deploy-global` with the desired
   selectors (for example `--only pi`); it preflights every selected root and runs the
   read-only generated-tree parity check without writing tracked files.
3. Review the dry-run report, then add `--apply`. Apply runs the generated build
   once before the diff, preflights all selected roots, and only then creates backups and
   copies files, merges owned settings, and applies the compatibility patch. Apply does not
   run a second generated-tree check.
   `--only pi` writes PI plus shared skills; mixed selections write only their
   named harnesses and shared metadata.
4. Run the disposable parent/child/tool smoke after apply. A second dry run
   should report no file/config changes when the managed state is current.

The apply journal prints the backup directory. To restore it, first run the
rollback dry run, then repeat with `--apply`:

```sh
python3 harness/pi/deploy.py rollback \
  --pi-dir "${PI_CODING_AGENT_DIR:-$HOME/.pi/agent}" \
  --backup-dir "${PI_CODING_AGENT_DIR:-$HOME/.pi/agent}/.hive-deploy-backups/<timestamp>" \
  --dry-run
python3 harness/pi/deploy.py rollback \
  --pi-dir "${PI_CODING_AGENT_DIR:-$HOME/.pi/agent}" \
  --backup-dir "${PI_CODING_AGENT_DIR:-$HOME/.pi/agent}/.hive-deploy-backups/<timestamp>" \
  --apply
```

Rollback preserves targets changed after the backup and reports the conflict;
it never overwrites a later user edit. It restores the state immediately before
helper apply, including pristine package files; it does not undo earlier native
`pi install` calls. Before first installation, retain the original settings and
package inventory. A full removal also removes only the newly installed packages
with the native package manager, preserving unrelated packages and preferences.
The PI-only route remains isolated from Claude, Grok, opencode, and unrelated
Codex configuration. The neutral engine covers the Pi and shared-skills roots; the default,
`all`, `harness`, and mixed routes retain the legacy writer and backup path for other
explicitly selected harness roots.

Before considering the integration usable, run the disposable repository smoke
for a Grok parent, an OpenAI-backed child, the structured question tool,
Context7, Linear, HeroUI Pro, the web tools, `hive-status` and hook readiness, bounded Git reads,
and the reviewer guard. No smoke should publish a commit, push, or production
change.
