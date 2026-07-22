# code-search-routing hook

Deterministic backstop for `rules/tools/code-search.md`, born from the v3/v4 benchmarks
(`_support/archive/audits/2026-07-21-benchmark-*.html`).

- **PreToolUse** (`Bash` + codegraph/jbcontext MCP): blocks (exit 2) an index tool in repos
  where it is contraindicated, per the versioned map `code-search-routing.json`
  (`repos` substrings matched against the command/projectPath/pathFilter/cwd). The stderr
  message redirects the agent to the routed alternative. Extend the map with evidence only.
- **PostToolUse** (same matchers): injects the anti-conclusion discipline reminder in the
  turn right after an index tool ran (the two failure modes it prevents were measured in v4:
  false-absence anchoring on CodeGraph, dead-code anchoring on jbcontext).

Deploy: `/deploy-global` copies `code-search-routing.sh` and `code-search-routing.json` to
`~/.claude/hooks/` and merges `settings-config.json` into `~/.claude/settings.json`.
Enforcement layers: deny = deterministic (hook exit 2); discipline reminder = prompt-convention
with deterministic delivery.
