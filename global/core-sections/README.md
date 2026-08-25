# core-sections — canonical source of the always-on cores

`global/CLAUDE.md` (Claude Code, deployed to `~/.claude/CLAUDE.md`) and
`harness/AGENTS.md` (Codex + opencode + Grok always-on core) are **generated** from this
directory by `harness/build.py`. Edit the section files here, run
`python3 harness/build.py`, and commit the regenerated outputs. The guard works at
assembly time: a stale output (matches HEAD, differs from the new assembly) regenerates
silently; a hand-edited output (differs from both) aborts the build with the edit
preserved on disk. `python3 harness/build.py --check` runs every parity check without
writing anything — the deploy-preflight / pre-commit form, where a stale or hand-edited
output fails instead of being regenerated.

## File format

Each `<name>.md` is one section, written in English, with frontmatter:

```yaml
---
order: 20                 # assembly position (gaps of 10)
targets: [claude, agents] # claude -> global/CLAUDE.md; agents -> harness/AGENTS.md
join: tight               # optional: attach to the previous section with a single
                          # newline (a bullet continuing a list); default is a blank line
---
```

- `order_claude:` / `order_agents:` override `order` per target — required when a shared
  section sits at different positions in the two outputs. Every target needs an order
  from one of the two keys.
- A sibling `<name>.claude.md` / `<name>.agents.md` **replaces the body** for that target
  only. Use sparingly — only where a shared section needs a small per-harness wording
  variation. An overlay that replaces the WHOLE body is apparent sharing, worse than two
  files: split it into two per-target sections instead (that is why the routing policy
  lives as `skill-routing.md` (claude) + `on-demand-rules.md` (agents), not as an
  overlay). Overlays carry no frontmatter.
- Shared sections (`targets: [claude, agents]`) are the point of this directory: one
  canonical text, one edit, both outputs. Prefer promoting a section to shared over
  keeping per-target twins of the same policy.
- Heading text is load-bearing: rule files, hooks, and `build.py` parity checks reference
  sections as `CLAUDE.md > <heading>` — renaming a heading requires sweeping those
  references first.
- `build.py` fails closed on an empty directory, malformed frontmatter, an unknown
  target, or an orphan overlay. This `README.md` is excluded from assembly.
