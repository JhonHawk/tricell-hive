## Browser Automation Tooling

> Which tool drives a browser. Default to the `agent-browser` CLI; reach for an MCP browser
> server only for what the CLI cannot do. Distinct from `agent-routing.md` (which *agent*
> verifies) — this picks the *tool* the agent uses.

### Primary: `agent-browser` CLI (via Bash)

For navigating, driving, scraping, and functional/QA verification of a running app, the
default is the `agent-browser` CLI — not an MCP browser server. Why, operationally:

- **Token cost.** Its `snapshot` (a11y tree with refs) is ~1.0× the page; the Playwright
  MCP snapshot of the same page is ~2.15× and chrome-devtools ~1.4×. Its `read` (clean text,
  no refs) is ~15× cheaper than a Playwright snapshot when you only need content. And the CLI
  lets you pipe/`grep`/redirect output — you control what reaches context; an MCP returns its
  full payload whether you need it or not. `--json` gives structured `{success,data}` output
  to parse instead of scrape, and snapshot refs (`@e1`, `@e2`) stay valid across interactions
  until the page changes — no DOM re-query.
- **Turns.** Chain a whole flow in ONE Bash call (`open && fill && click && snapshot`) = one
  model turn. The MCP equivalent is N round-trips (one tool call per action). For multi-step
  flows this dominates both wall-clock and tokens.
- **Latency.** The browser session persists between commands: cold start ~0.7s once, then
  ~30–40ms per warm command.

Coverage is broad enough to be the default: `console` (log/warn/error — the console-error
check), `network requests [--filter]` (verify the call returned what the UI claims),
`network route --abort` (simulate offline/500 for negative testing), `network trace/profiler`,
`auth`, screenshots, PDF, `eval`, CDP `connect`.

- **Profile is always `Tricell`**, enforced by `AGENT_BROWSER_PROFILE=Tricell` in `~/.zshenv` —
  every run inherits its login state (a read-only temp snapshot; original untouched) with no
  flag needed. Override one run with `--profile X` only if the user asks.
- **Batch in one Bash call** to keep it to one model turn; redirect large output to a file and
  read back only the slice you need.
- **Browsing untrusted/external pages → pass `--content-boundaries`** so page content is wrapped
  and distinguishable from tool output — a prompt-injection guard for agentic contexts
  (`quality/security.md`'s injection floor applies the moment you drive a real, public page).

### Reserve the MCP browser servers for their unique strengths

- **chrome-devtools MCP** — Lighthouse audits, performance-insight analysis, heap snapshots.
  These are what `agent-browser` does NOT cover; route performance/diagnostics here.
- **Playwright MCP** — fallback when `agent-browser` is unavailable, or when you specifically
  need a11y refs returned inline for step-by-step reasoning. Otherwise redundant with the CLI.

### Cross-harness note

`agent-browser` is a local CLI reachable from any harness via the shell, so this preference
holds for Claude Code, Codex, and opencode alike. The MCP browser servers may be absent in a
given harness — another reason the CLI is the portable default.
