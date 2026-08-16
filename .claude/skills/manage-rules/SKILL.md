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

Default scope: the rule files changed in the working tree / recent commits, or named by the user; `validate --all` scans the full inventory. **`global/CLAUDE.md` is in scope too** — it is the largest always-on artifact (~19% of everything a session loads) and checks 4-10 apply to it; checks 1-3 do not (it carries no frontmatter and is never path-scoped). For each file in scope check:

1. **Frontmatter — `paths:` is the only key Claude Code actually reads.** Per the official docs: *"Rules without a `paths` field are loaded unconditionally and apply to all files."* So there are exactly two real states:
   - **`paths: [...]`** with valid globs — loads only when a matching file is touched. This is the ONLY mechanism that keeps a rule out of the always-on set.
   - **No `paths:`** — always-on, whatever else the frontmatter says. We write `alwaysApply: true` to state the intent, but it is **documentation, not mechanism**: the file would load identically with empty frontmatter. Never treat it as a switch, and never invent a third scope key — an unrecognized key does not suppress loading.
   - Flag: a file with both `paths:` and `alwaysApply: true` (contradictory intent), and any rule whose prose claims it is not always-on while carrying no `paths:`.
2. **Glob validity** — Each pattern in `paths` should match real file types. Flag patterns that would never match anything useful (e.g., `**/*.xyz`).
3. **Scope correctness** — Rules with language-specific content (mentions `.ts`, `.java`, JSDoc, etc.) should be path-scoped, not `alwaysApply`. Rules with cross-language content should be `alwaysApply`, not path-scoped. Making a rule conditional means giving it `paths:`, and it must then pass all three reachability tests:
   - **Trigger** — the globs must match files the model will actually touch while the rule applies. A rule that fires on an ACTION rather than a file (deploying, driving a browser, a live incident) has no honest glob and stays always-on. Do not reach for a never-matching sentinel glob to fake conditionality: it works mechanically but leaves the rule reachable only through a skill, which is the failure mode below.
   - **Nothing safety-bearing is conditional.** A gate that loads only when some glob happens to match is a broken gate; that content stays always-on.
   - **Consumers can reach it.** Grep `global/agents/**` for citations. An agent with a `tools:` allowlist that omits `Skill` cannot invoke a router skill at all, and skills are not inherited from the parent — such a consumer needs the rule always-on, or must cite the deployed path (`~/.claude/rules/...`) and read it directly. Also fix any agent line that miscalls a rule's scope. Hybrid activation via likely entrypoint files is acceptable when a framework can be config-less (example: Tailwind v4 CSS-first), but the rule body must explicitly require marker verification before applying guidance.
4. **No duplication with core** — Compare rule content against `global/CLAUDE.md`. Flag rules that repeat what the core config already says.
5. **No duplication between rules** — Flag overlapping content across rule files (e.g., same library mentioned in two files).
6. **Size check** — Flag rules under 5 lines (too thin — consider merging). **No line ceiling and no corpus budget:** a long file is not a defect, a dense or incoherent one is. Split by cohesion — a file covering two domains that load on different triggers — never by length. (A line ceiling flagged 6 files permanently and taught everyone to ignore the validator.)
   - **Judge per-bullet density instead of totals:** flag any always-on bullet over ~500 chars or carrying 3+ hard modals — that is `Concise-first`'s "one directive per rule" made checkable.
   - **Deleting an always-on line requires reading its history first** (`git log -L<start>,<end>:<file>`) and quoting the introducing commit's **body**, not its subject — the subject often describes a neighbouring directive. Lines whose walk bottoms out at the initial import have no recoverable rationale: say so rather than treating absence as permission.
7. **Industry alignment (deep pass — only on `validate --deep` or explicit request)** — verify the in-scope rules still reflect current industry consensus (web search, context7); flag stale rules. The default validate skips this pass.
8. **Enforcement honesty** — A rule phrased as mechanical impossibility ("cannot", "physically blocked") must be backed by a deterministic layer (hook, deny permission, allowlist); otherwise flag it for rewording as confirm-gated or convention. Gates name their enforcement layer. Taxonomy: `_support/docs/enforcement-layers.md`.
   - **Verify the claimed SCOPE, not just the layer's existence.** A rule naming a hook must match what that hook actually matches — read the script and its README. A backstop that covers one invocation shape while the rule implies all of them is an overclaim: the rule keeps the enforcement line and gains the qualifier. (Caught `testing.md`'s post-tool-hub claim, silent for every runner outside the pnpm/turbo shapes.)
9. **Harness reachability** — Claude Code is not the only consumer; per in-scope rule verify how it reaches the other three:
   - **Always-on** → lands in Grok's flat symlink set (no `paths:`, filename without `__` — `deploy-global.sh > grok_always_on_rules`), and when it belongs to the cross-harness core its manual mirror in `harness/AGENTS.md` is a recorded decision (mirrored, or consciously Claude-only).
   - **Path-scoped under `languages/`** → present BOTH in the opencode rules pipeline (`build.py` conversion) AND in `language-rules`' injected references (`SKILL_REFERENCE_INJECTIONS`).
   - **Path-scoped elsewhere** → reachable through some router skill's injections, or explicitly Claude-only by design (say so in its header).
   - **Transition hazard, flag it every time it appears in a diff:** adding `paths:` to a previously always-on rule silently removes it from Grok; removing `paths:` silently adds it to every Grok session.
   - **`SKILL_REFERENCE_INJECTIONS` has a second writer:** `/agents-md-primary apply` edits that map via its `inject-to-router` outcome. This check is what verifies those edits — run it after one, and treat an injection added there as in-scope here even when no rule file changed.

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
4. **Load analysis** — What a session actually pays, in BYTES (the corpus grows by mass, not by
   rule count — line or rule counts report it stable while cost compounds):
   - Always-on total: `global/CLAUDE.md` + every rule with no `paths:`. Report bytes and the
     tokenizer proxy (`bytes ÷ 3.7`), plus each file's share so the biggest are visible.
   - Per-stack: always-on + that stack's path-scoped rules.
   - The trend, not just the level: bytes added vs deleted across the window since the last prune.

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
