---
name: manage-rules
description: >
  Manage global rule files in global/rules/ — validate frontmatter and content,
  audit coverage across technology stacks, or create new rules. Use with:
  /manage-rules validate, /manage-rules audit, or /manage-rules create <name>.
  Triggers when working with rule files, reviewing global config coverage,
  adding new technology standards, or checking rule quality.
---

Manage rule files in `global/rules/`. Parse `$ARGUMENTS` to determine the subcommand.

## Subcommands

### `validate` — Validate all rules

Default scope: the rule files changed in the working tree / recent commits, or named by the user; `validate --all` scans the full inventory. For each file in scope check:

1. **Frontmatter** — Must have exactly one of:
   - `alwaysApply: true` (no `paths` field) — for cross-language rules
   - `paths: [...]` array with valid glob patterns — for language/framework-scoped rules
   - `loadedBy: <skill-name>` — progressive disclosure: the rule loads when that router skill is invoked, not always. Verify the named skill exists and actually routes to the rule, and that the rule is listed in `SKILL_REFERENCE_INJECTIONS` in `harness/build.py` so the other harnesses receive it.
   - Flag files with more than one, none, or invalid frontmatter.
2. **Glob validity** — Each pattern in `paths` should match real file types. Flag patterns that would never match anything useful (e.g., `**/*.xyz`).
3. **Scope correctness** — Rules with language-specific content (mentions `.ts`, `.java`, JSDoc, etc.) should be path-scoped, not `alwaysApply`. Rules with cross-language content should be `alwaysApply`, not path-scoped. A `loadedBy:` rule must pass three reachability tests, all of them:
   - **Trigger** — a file glob the model will actually touch, or an unmistakable conversational signal. "The model will remember" is not a trigger. A rule that fires on an ACTION rather than a file (deploying, driving a browser, a live incident) has no glob, so it may only be conditional when an **always-on rule instructs the load at the moment of the action** — a pointer that merely names the file is not an instruction. Absent that, it stays always-on. Worked example: `browser-automation.md` is `loadedBy:` and legitimate, because `agent-routing.md > Skill & Browser Disambiguation` says to load it before driving a browser AND its security half was promoted to always-on; the deployment readiness gate is always-on precisely because nothing could reliably instruct its load.
   - **Nothing safety-bearing is conditional.** A gate reachable only when a skill happens to load is a broken gate; that content stays in an `alwaysApply` owner.
   - **Consumers can reach it.** Grep `global/agents/**` for citations of the rule. An agent with a `tools:` allowlist that omits `Skill` **cannot invoke a router skill at all**, and skills are not inherited from the parent — for those consumers the rule must stay always-on, or the agent must cite the deployed path (`~/.claude/rules/...`) and read it directly. Also fix any agent line that still calls the rule "always on". Hybrid activation via likely entrypoint files is acceptable when a framework can be config-less (example: Tailwind v4 CSS-first), but the rule body must explicitly require marker verification before applying guidance.
4. **No duplication with core** — Compare rule content against `global/CLAUDE.md`. Flag rules that repeat what the core config already says.
5. **No duplication between rules** — Flag overlapping content across rule files (e.g., same library mentioned in two files).
6. **Size check** — Flag rules under 5 lines (too thin — consider merging) or over 50 lines (consider splitting).
7. **Industry alignment (deep pass — only on `validate --deep` or explicit request)** — verify the in-scope rules still reflect current industry consensus (web search, context7); flag stale rules. The default validate skips this pass.
8. **Enforcement honesty** — A rule phrased as mechanical impossibility ("cannot", "physically blocked") must be backed by a deterministic layer (hook, deny permission, allowlist); otherwise flag it for rewording as confirm-gated or convention. Gates name their enforcement layer. Taxonomy: `_support/docs/enforcement-layers.md`.

Output a summary table, then specific issues per rule with suggestions.

### `audit` — Coverage report

Analyze what technology stacks are covered by rules and which have gaps:

1. **Inventory** — List all rules with: name, scope type (always/path-scoped), line count, and glob patterns.
2. **Stack coverage** — Map rules to technology stacks based on their paths and content:
   - TypeScript/JavaScript: which rules apply?
   - Angular: which rules apply?
   - Java/Kotlin: which rules apply?
   - CSS/Tailwind: which rules apply?
   - Python: any rules? (user has Python projects)
3. **Gap analysis** — Identify stacks in the user's ecosystem (from CLAUDE.md context: Next.js, Angular, NestJS, Express, Spring, Kotlin, Python, DevOps) that have no dedicated path-scoped rule.
4. **Load analysis** — Calculate how many rule lines load per stack:
   - Always-loaded lines (sum of all `alwaysApply` rules)
   - Per-stack lines (always-loaded + stack-specific path-scoped rules)
   - Compare to the old monolithic CLAUDE.GLOBAL.md (181 lines) to show savings.

### `create <name>` — Create a new rule

`<name>` is the rule filename without `.md`, e.g., `python-standards`.

1. Ask the user what technology/topic this rule covers.
2. Determine scope: if the rule is language-specific, use `paths` with appropriate globs. If cross-language, use `alwaysApply: true`.
3. Read `global/CLAUDE.md` and existing rules to avoid duplication.
4. Draft the rule following this structure:
   ```yaml
   ---
   paths:            # or alwaysApply: true
     - "**/*.ext"
   ---

   ## Title

   - Concrete rule 1
   - Concrete rule 2
   ```
5. Present the draft for review before saving.
6. Save to `global/rules/<name>.md`.

Rules for creating:
- Each rule must change Claude's behavior vs default. No generic advice.
- Gates name their enforcement layer (deterministic / confirm-gated / prompt-convention) — see `_support/docs/enforcement-layers.md`; never phrase a convention as mechanical impossibility.
- Prefer few strong rules over many weak ones.
- If the content fits naturally in an existing rule file, suggest merging instead of creating a new file.
- **Research before drafting**: use web search, context7, and authoritative sources (official docs, recognized books, RFC/specs) to ground the rule in current industry best practices. Don't write rules based solely on internal conventions — validate against the broader ecosystem.

### No arguments — Show help

If `$ARGUMENTS` is empty, show the available subcommands with usage examples.
