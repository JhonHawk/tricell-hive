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

Scan every `.md` file in `global/rules/` and check:

1. **Frontmatter** — Must have exactly one of:
   - `alwaysApply: true` (no `paths` field) — for cross-language rules
   - `paths: [...]` array with valid glob patterns — for language/framework-scoped rules
   - Flag files with both, neither, or invalid frontmatter.
2. **Glob validity** — Each pattern in `paths` should match real file types. Flag patterns that would never match anything useful (e.g., `**/*.xyz`).
3. **Scope correctness** — Rules with language-specific content (mentions `.ts`, `.java`, JSDoc, etc.) should be path-scoped, not `alwaysApply`. Rules with cross-language content should be `alwaysApply`, not path-scoped. Hybrid activation via likely entrypoint files is acceptable when a framework can be config-less (example: Tailwind v4 CSS-first), but the rule body must explicitly require marker verification before applying guidance.
4. **No duplication with core** — Compare rule content against `global/CLAUDE.md`. Flag rules that repeat what the core config already says.
5. **No duplication between rules** — Flag overlapping content across rule files (e.g., same library mentioned in two files).
6. **Size check** — Flag rules under 5 lines (too thin — consider merging) or over 50 lines (consider splitting).
7. **Industry alignment** — For each rule, verify its recommendations still reflect current industry consensus. Use web search and context7 to check if any rule has become outdated or if a better practice has emerged. Flag stale rules.
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
