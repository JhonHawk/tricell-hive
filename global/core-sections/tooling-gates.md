---
order: 270
targets: [claude]
---

## Tooling Gates

> The gate half of four rules whose detail is situational. Each names where its full text comes from; the gate holds whether or not that text was ever loaded — including where nothing would hold it: a decision taken before any command runs, or a tool that fires no trigger.

- **The query that counts is the write-time one.** The obligation attaches to the exact signature, options/props, and config shape at the moment you write the version-sensitive code. A prior query — planning-phase or earlier in the session — discharges it only when it demonstrably returned that exact detail (same surface, same version); a pattern-level answer ("use App Router") does not cover the call detail you then write. Re-query on signal: the detail isn't in what you already fetched, or the build/typecheck failed on that surface.
- **A hard stop remains only for framework-critical patterns** (auth flows, middleware, transaction management) with NO docs AND NO version anchor (greenfield without a manifest) — and that single question folds into the plan gate, never a mid-run stop. Signals, version anchoring by phase, query mechanics and the fallback chain: `context7.md`, held on a dependency install and carried by the `language-rules` skill.
- **An absence claim needs an exhaustive `rg` sweep** — broad vocabulary, English AND Spanish domain terms, 0 hits. Nothing weaker supports "X does not exist". **An assumed absence counts as a claimed one:** fixing the occurrence that failed and moving on asserts there are no others without ever saying so, so the sweep is owed before acting, not before writing the sentence (`CLAUDE.md > Critical Thinking` > Count the instances). Routing by operation type and the rest of the anti-conclusion discipline: `code-search.md` via the `language-rules` skill.
- **Delegate the flow** to the subagent whose intent matches (`review-ux`, `sdd-verify`, `visual-designer` — `agent-routing.md`) whenever it captures a screenshot for visual judgment, takes a `snapshot`, or chains 3+ interactions. What returns is findings in text; the captures die with the agent's context. **Keep the glance inline:** `read`, `console`, filtered `network requests`, a one-off `eval`.
- **That decision is made before any command runs** — and an MCP browser tool fires no hold at all, which is why it lives here. The rest of the gate block (profile, viewport, untrusted-content boundaries) is `browser-automation.md`, held on the first `agent-browser` command; the CLI mechanics are `browser-automation-reference.md` via the `language-rules` skill.
- **A change that alters behavior follows the verifiable test gate** — observed RED-to-GREEN, pass-to-pass, and verification that runs the checks rather than claiming they would pass. `test-gate.md` is held on the first code write; the full policy (test approach, coverage by change type, execution scope) is `testing.md` via the `language-rules` skill.
