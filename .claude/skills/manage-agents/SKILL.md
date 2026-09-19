---
name: manage-agents
description: >
  Validate Claude Code agent definitions in global/agents/ and rule texts in
  global/rules-situational/ + global/core-sections/ against the hub's design principles —
  the judgment checks no hook or build step covers. Use with:
  /manage-agents validate [--all] [--deep]. Triggers after editing agent or rule files,
  after routing changes, or when reviewing agent quality.
---

Validate agent definitions and rule texts in this workspace. Parse `$ARGUMENTS` to determine the subcommand.

## Subcommands

### `validate` — Validate all agents

Default scope: the agent files changed in the working tree / recent commits, or named by the user; `validate --all` scans the full inventory. Check each in-scope agent against the design principles in CLAUDE.md:

1. **Line weight** — Past 120 lines of the agent's own text, review line by line for filler and report what could go; a long agent whose lines all change its output passes. Count the source file only — rule texts carried through `packs:` are inlined at build and never count. `packs:` names resolve to rule files (the build fails otherwise) and a packed agent carries `agent-core-gates`.
2. **Naming** — kebab-case, 3-50 chars, starts/ends with alphanumeric.
3. **Frontmatter** — Must have: name, description, model, color. `tools` is recommended but optional; if omitted, report that the agent inherits the default toolset and verify that this is intentional for the role. Recognized optional fields: `tools`, `effort`, `permissionMode`, `maxTurns`, `skills`, `disallowedTools`, `mcpServers`, `hooks`, `background`, `isolation`, `initialPrompt`, `memory`, `omitClaudeMd`, and the hive-only `packs` (comma-separated rule basenames, never a YAML list).
4. **Description quality** — Must be 1-3 action-oriented sentences stating WHEN to invoke (triggers). Flag generic descriptions, and flag descriptions that compress the body's procedure into steps — a workflow-summarizing description risks being followed instead of the body.
5. **Tool restriction** — Read-only agents (review/) should not have Write, Edit, or Bash. Quality agents may be remediation-oriented (Write/Edit expected) or audit-oriented (read-only plus Bash for external analyzers); validate the tool surface against the role described. `permissionMode: plan` is a positive signal for read-only review agents.
6. **Color vs directory** — Agent must be in the correct subdirectory for its color (design/=blue, development/=green, review/=cyan, quality/=yellow, ops/=red, docs/=magenta).
7. **No global rule duplication** — Flag agent lines that repeat content the agent already receives: a core section (`global/core-sections/**`, assembled into `global/CLAUDE.md`) or a rule text the agent carries through `packs:`.
8. **Role rules resolution** — every path cited in an agent's `## Role rules` table resolves to a real file: the canonical `~/.claude/...` form exists as deployed state, and the same reference exists in the generated `harness/agents-skills/` tree. `build.py` rebases the table's path root per harness, but only for reference files `SKILL_REFERENCE_INJECTIONS` actually emits — a row pointing at a reference the build never injects is a broken pointer in all four harnesses. Hand-added rows are the typical source of drift.
9. **README inventory sync** (always on `--all`; otherwise when any agent file changed) — `README.md`'s agents table matches disk: heading count == `ls global/agents/*/*.md | wc -l`, one row per agent, tool-surface cell consistent with the agent's `tools`/`disallowedTools` frontmatter.
10. **Cross-harness emission** — the agent is consumed by four harnesses, not one. Run `python3 harness/build.py`: (a) `git status --porcelain harness/` comes back clean — dirty means sources changed without a rebuild; (b) every `WARNING` the build prints for an in-scope agent is a reported finding, never left in scroll-off; (c) per in-scope agent, intent survives translation — a read-only reviewer (`permissionMode: plan`, or no Write/Edit in its surface) emits `sandbox_mode = "read-only"` in its Codex TOML, `permission_mode: plan` in its Grok copy, and `edit: "deny"` in opencode; an `Agent` denial (allowlist omission or `disallowedTools`) emits `task: "deny"` (opencode) and the do-not-spawn clause (Codex/Grok); (d) a `~/.claude/...` path cited in the body must exist as deployed machine state reachable from every harness (the claude scope deploys it), or the citation is Claude-only by accident.
11. **Industry alignment (deep pass — only on `validate --deep` or explicit request)** — check the in-scope agents' Rules sections against current industry consensus (web search, context7); flag outdated or missing practices. The default validate skips this pass.

Then, when the scope includes a rule text (`global/rules-situational/**`) or a core section (`global/core-sections/**`), check the five judgments a build cannot make. `harness/build.py` already refuses a rule no channel delivers, a trigger with no router reference, an `include:` over a triggered rule, a bad `exclusive-with:`, a malformed command prefix and `paths:` anywhere — never re-check those by hand. The contract and the wiring procedure live in `global/rules-situational/README.md`:

R1. **The trigger is an ACT, not an intent** — could a third party reading the transcript say the moment happened? A `globs:` entry names files the model actually writes while the rule applies; a `commands:` prefix names a verb it actually runs. A rule whose real trigger is a decision has no honest trigger: it belongs in the core via `include:`, or router-only. Flag any sentinel glob invented to make a rule look conditional.
R2. **Glob breadth** — flag a pattern that matches far more than the rule governs (`**/*.md` on a rule about specs), one that matches nothing useful (`**/*.xyz`), and any glob expanding past 256 brace alternatives, which the hook drops rather than expand.
R3. **Hold cost** — how many rules does this trigger already hold on the same first write or command, and does the denial reason still name this one whole inside Grok's ~264-character budget? A reference that cannot be named whole is simply not gated there. Report the rules-per-trigger count, not just pass/fail.
R4. **Pack coverage** — each specialized agent carries the packs its stack needs (`agent-core-gates` always, the language/framework rules of the stack it writes, the gates it loses with `omitClaudeMd: true`), and carries no pack for a stack it never touches.
R5. **Enforcement honesty** — a rule phrased as mechanical impossibility ("cannot", "blocked") is backed by a deterministic layer, and the claimed SCOPE matches what that layer actually matches (read the script and its README). A backstop covering one invocation shape while the rule implies all of them keeps the line and gains the qualifier. Taxonomy: `_support/docs/enforcement-layers.md`.

Core sections carry assembly frontmatter (`order`/`targets`/`join`/`include`), not rule frontmatter, and are always-on by definition: R1–R3 do not apply to them, R5 does. `global/CLAUDE.md` is GENERATED — findings land in the section file or the included rule text, never in the output.

Output a summary table, then list specific issues per agent and per rule with suggestions.

### No arguments — Show help

If `$ARGUMENTS` is empty, show `validate [--all] [--deep]` with usage examples.
