# Pi adapter

This directory contains the Hive integration for [Pi](https://pi.dev/), the
terminal coding harness. It keeps the canonical agent definitions and Bash
hooks shared with Claude Code, Codex, and Grok while adding the Pi-specific
runtime needed for subagents, plan mode, bounded Git reads, and package-backed
tools.

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

## What Pi receives

`python3 harness/build.py` generates the Pi role files alongside the Codex,
opencode, and Grok outputs. The generated files are derived from
`global/agents/**/*.md`; edit the canonical agent, then rebuild. The generator
emits the Pi model, thinking level, context inheritance, tool selection, and
extension fields that Pi can enforce. A `__HIVE_PI_ROOT__` placeholder is
resolved only while deploying, so generated files do not contain a machine
specific home path. MCP-backed Context7 access uses the native `mcp` proxy in
the generated tool allowlists. Hive's closed input guard admits only these two
operations:

```text
{ tool: "context7_resolve-library-id", args: { query, libraryName } }
{ tool: "context7_query-docs", args: { libraryId, query } }
```

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
  output, and never accepts a shell command.
- The plan extension exposes `/hive-plan enter`, `/hive-plan approve [path]`,
  `/hive-plan cancel`, and `/hive-plan status`. While active, the parent can
  use bounded reads and structured Git inspection but has no generic Bash or
  project-write tool. Approval captures the settled plan and starts execution
  in the same explicit action; it does not approve a plan merely because a
  child finished.
- Plan entry snapshots the exact active tool set. Approval and cancellation
  restore that snapshot, and a child launched while planning keeps its
  restrictions for the rest of its lifetime. Entry is rejected while the
  parent or an existing child is active.
- The plan capture bridge calls the canonical
  `global/hooks/flow-plan-capture/flow-plan-capture.sh --from-pi-command` with
  `{ "harness": "pi", "cwd": "...", "tool_response": { "plan": "..." } }`.
  It preserves the plan bytes, returns `{ status: "captured", path, sha256 }`
  for a durable capture, and fails closed for malformed or failed captures. The
  digest is the native `shasum -a 256` hash of the exact plan bytes after the
  artifact's four metadata lines. Before execution, the runtime compares that
  digest with the approved candidate and blocks on a mismatch. A Flow workspace
  uses its ledger, a standalone repository uses its session layer, and a
  directory outside a repository returns `session_only`. `Session: no` returns
  `skipped`.

The general extension is loaded for every Hive role. The reviewer guard is
loaded only for the six canonical roles whose source agent declares the
reviewer hook: `code-reviewer`, `code-scout`, `product-critic`,
`security-reviewer`, `spec-quality-reviewer`, and `workspace-custodian`.
Role identity comes from the generated agent definition; it is never inferred
from a prompt or environment variable. A missing required child extension is
an infrastructure error, not a reason to silently continue without the guard.
The generated `tools` allowlists carry the readiness sentinels:
`hive_hook_readiness` is selected by all 25 roles, and those six reviewer roles
also select `hive_reviewer_readiness`. The corresponding child extensions are
`hive-hooks.ts` for every role and `hive/reviewer-guard.ts` for the six-role
reviewer set. `pi-subagents` 0.67.0 derives the required child-tool checks
from the explicit `tools` selection.

Children run in the background through `pi-subagents`, which keeps the parent
responsive and exposes lifecycle state. The default context is fresh, with
Hive project/global instructions and shared skills explicitly inherited. Child
agents do not recursively delegate. The parent session pointer
`PI_SUBAGENT_PARENT_SESSION` is used for plan inheritance and lifecycle
association; it is not a trusted role identity.

## Pinned packages and external surfaces

| Package or service | Pin / endpoint | Purpose | Loading status |
|---|---|---|---|
| `pi-subagents` | `0.67.0` | Background children, role discovery, lifecycle, and bounded orchestration | Required |
| `gentle-engram` | `0.1.12` | Native Engram HTTP memory integration | Required; no `pi-engram init` |
| `pi-mcp-adapter` | `2.32.1` | MCP transport for Context7 | Required |
| `@juicesharp/rpiv-ask-user-question` | `2.9.0` | Structured user questions in the parent session | Required; children use `contact_supervisor` |
| `pi-web-access` | `0.29.0` | OpenAI-backed web search and source access | Required for research roles |
| Context7 | `https://mcp.context7.com/mcp` | Version-anchored library documentation through the native `mcp` proxy | Closed guard allows only `context7_resolve-library-id` and `context7_query-docs` |
| Web | provider `openai`, search provider `openai-codex` | External research | No workflow provider |

For `pi-mcp-adapter` 2.32.1, the managed Context7 entry sets
`directTools: false`, sets `includeTools` to those two operations, and uses
`lifecycle: "lazy"`. Context7 advertises `ttlMs: 0`, so the adapter discards
its cache and direct or namespace tools do not remain registered; an eager
lifecycle does not change that. The native `mcp` proxy and its closed input
guard are the stable plan-safe surface.

### `pi-subagents` 0.67.0 compatibility patch

Pi 0.67.0 needs the reviewed compatibility patch in
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
| Bounded plan mode | Hive-specific commands and exact tool snapshot/restore | — | Undocumented upstream; enforced by runtime tests |
| Canonical Bash hooks | Hive-specific JSON adapter around existing scripts | [Pi extensions](https://pi.dev/docs/latest/extensions) for lifecycle extension points | Bridge behavior is Hive-specific |
| OS sandbox | Not enabled | [Pi containerization](https://pi.dev/docs/latest/containerization) | Explicitly excluded |

The package reference is version-sensitive. If the installed package changes,
re-check its agent frontmatter and child-extension behavior before changing the
generator or deployment contract.

## Build, test, and deploy

From the repository root:

```sh
python3 harness/build.py
command -v shasum >/dev/null && shasum -a 256 </dev/null
cd harness/pi
pnpm typecheck
pnpm test
```

The `shasum` line is a required PI preflight: the canonical capture hook uses
the native command to produce the returned digest.

The source capture regression suite is an independent check of the canonical
hook contract and should run before the disposable Pi smoke:

```sh
python3 global/hooks/flow-plan-capture/test_pi_capture.py
```

The capture suite does not replace the Pi typecheck, package checks, or
disposable runtime smoke.

The PI-only deployment order is deliberate:

1. Install the exact `pi-subagents@0.67.0` package into the target PI root
   before invoking the helper. For a non-default root, pass the same
   `PI_CODING_AGENT_DIR` to Pi's package installer; begin with:

   ```sh
   PI_CODING_AGENT_DIR="${PI_CODING_AGENT_DIR:-$HOME/.pi/agent}" \
     pi install npm:pi-subagents@0.67.0
   ```

   Stage and audit the remaining pinned packages before activation. The helper
   records all five pins but does not install packages or run OSV auditing.
2. Run the generator, native `shasum` preflight, type checks, package checks,
   and the PI dry run. The dry run is `/deploy-global --only pi`; it refuses a
   missing, wrong-version, or hash-mismatched `pi-subagents` tree.
3. Review the dry-run report, then add `--apply` to copy PI files and hooks,
   merge owned settings, apply the compatibility patch, and create a
   conflict-aware backup under `PI_CODING_AGENT_DIR` (default `~/.pi/agent`).
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
The deploy route remains isolated from Claude, Grok, opencode, and unrelated
Codex configuration.

Before considering the integration usable, run the disposable repository smoke
for a Grok parent, an OpenAI-backed child, the structured question tool,
Context7, the web tools, plan capture, and the reviewer guard. No smoke should
publish a commit, push, or production change.
