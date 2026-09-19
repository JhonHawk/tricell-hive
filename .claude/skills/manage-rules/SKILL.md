---
name: manage-rules
description: >
  Manage global rule files in global/rules/ and global/rules-situational/ — validate frontmatter, content, reachability
  and enforcement honesty (with the always-on load report on --all), or create new rules.
  Use with: /manage-rules validate [--all] [--deep] or /manage-rules create <name>.
  Triggers after editing rule or core-section files, or when checking rule quality.
---

Manage rule files in `global/rules/` and `global/rules-situational/`. Parse `$ARGUMENTS` to determine the subcommand.

## Subcommands

### `validate` — Validate all rules

Default scope: the rule files changed in the working tree / recent commits, or named by the user; `validate --all` scans the full inventory. **The core content is in scope through `global/core-sections/**`** — `global/CLAUDE.md` is GENERATED from those section files by `harness/build.py` (edits to the output get refused or destroyed by the next build), so findings and fixes land in the section files, with the assembled output read for context only. It remains the largest always-on artifact (~19% of everything a session loads); checks 4-10 apply to the section content, checks 1-3 do not (core sections carry assembly frontmatter — `order`/`targets`/`join` — not rule frontmatter, and are always-on by definition). For each file in scope check:

1. **The directory is the mechanism — judge each file class against its own contract.** A **deployed rule** (`global/rules/**`) is always-on by construction: everything there lands in `~/.claude/rules/`, which the docs load unconditionally (*"Rules without a `paths` field are loaded unconditionally"*). `alwaysApply: true` documents that intent; it switches nothing, and no rule here carries `paths:` any more — flag one that does, it is a leftover. A **situational rule** (`global/rules-situational/**`) is never deployed: `globs:` there is read by `harness/build.py` alone, feeding `rule-manifest.json` and the `rule-delivery` hook (`build.py` strips the frontmatter on injection into a router's `references/`). Verify it against that directory's `README.md` contract. A **harness file** (`harness/AGENTS.md`) carries no frontmatter and is out of this check entirely.
2. **Glob validity** — Each pattern in a `globs:` list should match real file types, by path segment. Flag patterns that would never match anything useful (`**/*.xyz`), and any glob expanding past 256 brace alternatives — the hook drops those rather than pay the expansion.
3. **Scope correctness** — Language-specific content (mentions `.ts`, `.java`, JSDoc) belongs in `global/rules-situational/` with `globs:`, never always-on in `global/rules/`; cross-language or action-triggered content is the reverse. Moving a rule out of always-on means it must pass all three reachability tests:
   - **Trigger** — the globs must match files the model will actually touch while the rule applies. A rule that fires on an ACTION rather than a file (deploying, driving a browser, a live incident) has no honest glob: it stays always-on, or lives in `rules-situational/` WITHOUT `globs:` and is reached by its router alone. Never fake conditionality with a sentinel glob — the hook would then never hold a write for it, and the router becomes the only channel by accident rather than by decision.
   - **Nothing safety-bearing is conditional.** A gate that loads only when some glob happens to match is a broken gate; that content stays always-on.
   - **Consumers can reach it.** Grep `global/agents/**` for citations. An agent with a `tools:` allowlist that omits `Skill` cannot invoke a router skill at all, and skills are not inherited from the parent — such a consumer needs the rule always-on, the rule inlined via `packs:`, or the `rule-delivery` hold naming a reference it can `Read` directly. Also fix any agent line that miscalls a rule's scope. Hybrid activation via likely entrypoint files is acceptable when a framework can be config-less (example: Tailwind v4 CSS-first), but the rule body must explicitly require marker verification before applying guidance.
4. **No duplication with core** — Compare rule content against `global/CLAUDE.md`. Flag rules that repeat what the core config already says.
5. **No duplication between rules** — Flag overlapping content across rule files (e.g., same library mentioned in two files).
6. **Size check** — Flag rules under 5 lines (too thin — consider merging). **No line ceiling and no corpus budget:** a long file is not a defect, a dense or incoherent one is. Split by cohesion — a file covering two domains that load on different triggers — never by length. (A line ceiling flagged 6 files permanently and taught everyone to ignore the validator.)
   - **Judge per-bullet density instead of totals:** flag any always-on bullet over ~500 chars or carrying 3+ hard modals — that is `Concise-first`'s "one directive per rule" made checkable.
   - **Deleting an always-on line requires reading its history first** (`git log -L<start>,<end>:<file>`) and quoting the introducing commit's **body**, not its subject — the subject often describes a neighbouring directive. Lines whose walk bottoms out at the initial import have no recoverable rationale: say so rather than treating absence as permission. **Core content splits its history at the core-sections cutover (2026-08):** after it, the line lives in a `global/core-sections/*.md` section file; before it, in `global/CLAUDE.md` (or `harness/AGENTS.md`) — when the section-file walk bottoms out at the cutover commit, continue the walk in the pre-cutover file.
7. **Industry alignment (deep pass — only on `validate --deep` or explicit request)** — verify the in-scope rules still reflect current industry consensus (web search, context7); flag stale rules. The default validate skips this pass.
8. **Enforcement honesty** — A rule phrased as mechanical impossibility ("cannot", "physically blocked") must be backed by a deterministic layer (hook, deny permission, allowlist); otherwise flag it for rewording as confirm-gated or convention. Gates name their enforcement layer. Taxonomy: `_support/docs/enforcement-layers.md`.
   - **Verify the claimed SCOPE, not just the layer's existence.** A rule naming a hook must match what that hook actually matches — read the script and its README. A backstop that covers one invocation shape while the rule implies all of them is an overclaim: the rule keeps the enforcement line and gains the qualifier. (Caught `testing.md`'s post-tool-hub claim, silent for every runner outside the pnpm/turbo shapes.)
9. **Harness reachability** — Claude Code is not the only consumer; per in-scope rule verify how it reaches the other three:
   - **Always-on** → lands in Grok's flat symlink set (no `paths:`, filename without `__` — `deploy-global.sh > grok_always_on_rules`). When it belongs to the cross-harness core, it lands in a `global/core-sections/` section targeting `agents` (or both) and `build.py` assembles `harness/AGENTS.md` from it — there is no manual mirror any more; what gets recorded is the deliberate decision to keep a rule Claude-only (a `claude`-only section, or no core section at all).
   - **Glob-scoped (`rules-situational/` with `globs:`)** → present in `harness/rule-manifest.json` (so the `rule-delivery` hook can hold a matching write on Claude Code, Grok and Codex) AND in a router skill's injected references (`SKILL_REFERENCE_INJECTIONS`) — opencode and PI have only the second.
   - **Router-only (`rules-situational/` without `globs:`)** → reachable through some router skill's injections, or inlined into the agents that declare it in `packs:` (say which, in its header).
   - **Numeric parity of the delegation gates is already deterministic** — `harness/build.py` compares the structured thresholds in `agent-routing.md > Delegation Gates` against the core bullet and exits non-zero on drift. Do not re-check those numbers by hand; do flag a threshold restated in a THIRD place, which the check does not see.
   - **Transition hazard, flag it every time it appears in a diff:** moving a rule out of `global/rules/` silently removes it from Grok's flat symlink set; moving one in silently adds it to every Grok session.
   - **`SKILL_REFERENCE_INJECTIONS` has a second writer:** `/agents-md-primary apply` edits that map via its `inject-to-router` outcome. This check is what verifies those edits — run it after one, and treat an injection added there as in-scope here even when no rule file changed.

10. **Load report (`--all` only)** — what a session actually pays, in BYTES (the corpus grows by mass, not by rule count): always-on total (`global/CLAUDE.md` + every rule under `global/rules/`) with the tokenizer proxy (`bytes ÷ 3.7`) and each file's share; per-stack totals (always-on + that stack's glob-scoped rules); and the trend since the last prune (bytes added vs deleted), not just the level. `build.py` reports the two cores' size — this report adds the rule corpus around them.

Output a summary table, then specific issues per rule with suggestions.

### `create <name>` — Create a new rule

`<name>` is the rule filename without `.md`, e.g., `python-standards`.

1. Ask the user what technology/topic this rule covers.
2. Determine the home: language- or file-specific → `global/rules-situational/<name>.md` with `globs:`; cross-language and action-triggered → `global/rules/<dir>/<name>.md` with `alwaysApply: true`.
3. Read `global/CLAUDE.md` and existing rules to avoid duplication.
4. Draft the rule following this structure:
   ```yaml
   ---
   globs:            # rules-situational/ — or alwaysApply: true under rules/
     - "**/*.ext"
   ---

   ## Title

   - Concrete rule 1
   - Concrete rule 2
   ```
5. Present the draft for review before saving.
6. Save to the home chosen in step 2.

Rules for creating:
- Each rule must change Claude's behavior vs default. No generic advice.
- Gates name their enforcement layer (deterministic / confirm-gated / prompt-convention) — see `_support/docs/enforcement-layers.md`; never phrase a convention as mechanical impossibility.
- Prefer few strong rules over many weak ones.
- If the content fits naturally in an existing rule file, suggest merging instead of creating a new file.
- **Research before drafting**: use web search, context7, and authoritative sources (official docs, recognized books, RFC/specs) to ground the rule in current industry best practices. Don't write rules based solely on internal conventions — validate against the broader ecosystem.

### No arguments — Show help

If `$ARGUMENTS` is empty, show the available subcommands with usage examples.
