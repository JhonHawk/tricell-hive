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
| Always-on core | `global/CLAUDE.md` | `~/.claude/CLAUDE.md` | Read at launch, walking up from cwd; `@file` imports expanded inline | [memory#how-claude-md-files-load](https://code.claude.com/docs/en/memory#how-claude-md-files-load) · [memory#import-additional-files](https://code.claude.com/docs/en/memory#import-additional-files) | 2026-08-20 |
| Always-on rules | 12 rule files under `global/rules/`, none with `paths:` | `~/.claude/rules/**` | Loaded unconditionally, every session, every project | [memory#path-specific-rules](https://code.claude.com/docs/en/memory#path-specific-rules) — *"Rules without a `paths` field are loaded unconditionally and apply to all files."* | 2026-08-20 |
| Delivered rules | 28 rule files under `global/rules-situational/`, 19 of them with `globs:` | `~/.claude/skills/<router>/references/`; the two pack-only texts (`agent-core-gates`, `test-gate`) reach an agent only inlined through its `packs:` | **Not** deployed to `rules/`. A `globs:` rule is held by the `rule-delivery` hook on a matching write until its reference is read; otherwise reachable by invoking a router skill, or carried inside a packed agent | (no upstream mechanism; a repo convention — see below) | — |
| Skills | 17 skills | `~/.claude/skills/**` | `SKILL.md` frontmatter drives invocation gating (`disable-model-invocation`, `user-invocable`, `allowed-tools`) | [skills#frontmatter-reference](https://code.claude.com/docs/en/skills#frontmatter-reference) — *"`user-invocable` … Set to `false` when only Claude should invoke the skill"* | 2026-08-20 |
| Agents | 26 subagents, shipped from the generated `harness/claude/agents/` (the `global/agents/` source plus any `packs:` inlined) | `~/.claude/agents/**` | Discovered recursively; `tools:` is the enforcing allowlist, `model:` defaults to `inherit` | [sub-agents#supported-frontmatter-fields](https://code.claude.com/docs/en/sub-agents#supported-frontmatter-fields) | 2026-08-20 |
| Hooks | 9 hook dirs (a script + a `settings-config.json` block each) | scripts → `~/.claude/hooks/`, registration → `~/.claude/settings.json` | Hooks are registered **in settings**, never auto-scanned from a directory | [hooks#hook-locations](https://code.claude.com/docs/en/hooks#hook-locations) — *"Hooks are defined in JSON settings files."* | 2026-08-20 |
| Settings precedence | merged blocks only | `~/.claude/settings.json` | Managed › CLI args › Local › Project › User | [settings#settings-precedence](https://code.claude.com/docs/en/settings#settings-precedence) | 2026-08-20 |

**There is no `/docs/en/rules` page.** The entire `.claude/rules/` mechanism — the `paths:`
key included — is documented inside the **memory** page. A link to `code.claude.com/docs/en/rules`
is a 404; do not write one.

## What does NOT reach it

The 28 rule texts under `global/rules-situational/`, as loadable files. Everything else
still arrives: all 12 always-on rules, all 17 skills with their native gates, all 26 agents
with enforced tool allowlists, and all 9 hooks — every constraint documented in the other
three READMEs is a subtraction from this baseline.

Those 28 reach Claude Code exactly as they reach Grok and Codex: the 19 carrying `globs:`
by the `rule-delivery` hook's hold on a matching write, all of them through a router skill's
`references/` (injected by `harness/build.py`, `SKILL_REFERENCE_INJECTIONS`) or a pack
inlined into the agent — the other 9 have no hook channel on any harness. `paths:` still
works upstream; this repo stopped using it. It fires on a READ, so it misses the creation of
the first file of a kind, it charges every read-only agent that opens a `.ts` for a rule it
will never apply, and it does not survive compaction — a deny that names the file to read is
the primitive four of the five harnesses enforce (opencode has no hook channel —
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

# Always-on vs delivered, from the source tree
find global/rules -name '*.md' | wc -l                    # always-on
find global/rules-situational -name '*.md' ! -name README.md | wc -l   # delivered
rg -l '^globs:' global/rules-situational | wc -l          # of those, hook-deliverable
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
