# Claude Code layer — what it loads and where that is documented

`global/` is the deployable source of truth for `~/.claude/`. Claude Code is the only
harness that reads these files **natively** — no adapter, no generated tree, no
translation. Every other harness gets a lossy projection of the same sources
(`harness/README.md`).

Companion files: `harness/codex/README.md`, `harness/opencode/README.md`,
`harness/grok/README.md`. All four share the section layout below so they can be read
side by side.

## What this harness loads

Every URL below was fetched and its quote extracted from the page body on the date in the
last column. When a mechanism changes upstream, this is where you go to check.

| Layer | What | Where it lands | Mechanism | Official doc | Verified |
|---|---|---|---|---|---|
| Always-on core | `global/CLAUDE.md`, with the 7 always-on rule texts inlined into it by their `rule-*` core sections | `~/.claude/CLAUDE.md` | Read at launch, walking up from cwd; `@file` imports expanded inline | [memory#how-claude-md-files-load](https://code.claude.com/docs/en/memory#how-claude-md-files-load) · [memory#import-additional-files](https://code.claude.com/docs/en/memory#import-additional-files) | 2026-08-20 |
| Delivered rules | the other 35 texts under `global/rules-situational/`, 26 of them carrying `globs:`/`commands:` | `~/.claude/skills/<router>/references/`; the pack-only text (`agent-core-gates`) reaches an agent only inlined through its `packs:` | **Not** deployed as rule files — the conditional-rule channel is retired (below). A triggered rule is held by the `rule-delivery` hook on a matching write or command until its reference is read; otherwise reachable by invoking a router skill, or carried inside a packed agent | (no upstream mechanism; a repo convention — see below) | — |
| Skills | 18 skills, including automatic `flow-research` with conditional references | `~/.claude/skills/**` | `SKILL.md` frontmatter drives invocation gating (`disable-model-invocation`, `user-invocable`, `allowed-tools`) | [skills#frontmatter-reference](https://code.claude.com/docs/en/skills#frontmatter-reference) — *"`user-invocable` … Set to `false` when only Claude should invoke the skill"* | 2026-08-20 |
| Agents | 26 subagents, shipped from the generated `harness/claude/agents/` (the `global/agents/` source plus any `packs:` inlined) | `~/.claude/agents/**` | Discovered recursively; `tools:` is the enforcing allowlist, `model:` defaults to `inherit` | [sub-agents#supported-frontmatter-fields](https://code.claude.com/docs/en/sub-agents#supported-frontmatter-fields) | 2026-08-20 |
| Hooks | 8 hook dirs (a script + a `settings-config.json` block each) | scripts → `~/.claude/hooks/`, registration → `~/.claude/settings.json` | Hooks are registered **in settings**, never auto-scanned from a directory | [hooks#hook-locations](https://code.claude.com/docs/en/hooks#hook-locations) — *"Hooks are defined in JSON settings files."* | 2026-08-20 |
| Settings precedence | merged blocks only | `~/.claude/settings.json` | Managed › CLI args › Local › Project › User | [settings#settings-precedence](https://code.claude.com/docs/en/settings#settings-precedence) | 2026-08-20 |

**There is no `/docs/en/rules` page.** The entire `.claude/rules/` mechanism — the `paths:`
key included — is documented inside the **memory** page. A link to `code.claude.com/docs/en/rules`
is a 404; do not write one. The mechanism itself is retired here: `paths:` is a build error and
`~/.claude/rules/` receives nothing, so the *"loaded unconditionally"* clause now describes
only the core file above.

## What does NOT reach it

Any rule text as a loadable file. Everything else still arrives: the always-on core with its
7 inlined rules, all 18 skills with their native gates, all 26 agents with enforced tool
allowlists, and all 8 hooks — every constraint documented in the other three READMEs is a
subtraction from this baseline.

The triggered coverage includes modular TS/JS suffixes, modern and legacy Compose names,
Next server filename conventions, and package/workspace manifests. The last request only
development and security guidance; generated lockfiles have no blanket filename trigger.
Framework applicability stays in the rule text, not in hook-side framework detection.

The other 35 texts reach Claude Code exactly as they reach Grok and Codex: the 26 carrying
`globs:`/`commands:` by the `rule-delivery` hook's hold on a matching write or command, all
but the pack-only one through a router skill's `references/` (injected by `harness/build.py`,
`SKILL_REFERENCE_INJECTIONS`), or inlined into a packed agent. `paths:` still works upstream;
this repo stopped using it. It fires on a READ, so it misses the creation of the first file of
a kind, it charges every read-only agent that opens a `.ts` for a rule it will never apply,
and it does not survive compaction — a deny that names the file to read is the primitive four
of the five harnesses enforce (opencode has no hook channel —
`global/hooks/rule-delivery/README.md`).

## Unverified / undocumented dependencies

None. Every loading mechanism this layer depends on is covered by the official docs above.

Two upstream statements independently **corroborate** claims this repo already made:

- Rules carrying `paths:` are not re-injected after `/compact` — one of the reasons this
  repo replaced that channel. `rule-delivery` drops its state on `SessionStart`
  `compact|clear`, so the hold re-arms instead of the rule silently vanishing.
- Claude Code does **not** fall back to `AGENTS.md` when `CLAUDE.md` is absent, which is
  why every repo here needs a `CLAUDE.md` containing `@AGENTS.md`
  (`_support/docs/methodology-bibliography.md > Harness context-loading mechanics`).

## How to re-verify

```bash
# What is actually deployed right now (dry-run; never pass --apply to inspect)
.claude/skills/deploy-global/scripts/deploy-global.sh --only claude --verbose

# Always-on vs delivered, from the generated manifest (never by subtraction)
python3 - <<'PY'
import json
rules = json.load(open('harness/rule-manifest.json'))['rules']
print(len(rules), 'texts')
print(sum(r['always_on'] for r in rules), 'always-on (core include)')
print(sum(bool(r['globs'] or r['commands']) for r in rules), 'hook-triggered')
PY
```

To observe which instruction files actually load in a session and why, register an
`InstructionsLoaded` hook: the hive's `instructions-audit` instrument was retired on
2026-09-13 once the always-on reduction was verified (last use:
`_support/archive/audits/2026-09-13-hive-control-remeasurement.md`).

## Hive planning contract

`flow-plan` and `flow-build` use the shared plan artifact and read-only validator. Approval,
action/target scope and execution evidence travel with the plan; entering or leaving native
plan mode grants no Hive authority. During drafting, the parent's no-implementation boundary
is a prompt convention, not a universal write sandbox. `Session: no` retains conversation-only
operation without durable validation or cross-harness resume guarantees.

Source: `global/skills/flow-core/references/plan-format.md` (Hive convention, 2026-09-10).
