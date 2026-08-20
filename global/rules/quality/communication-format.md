---
alwaysApply: true
---

## Communication Format

> **In-thread prose is the default.** HTML is for output with a life outside this conversation — something the user keeps, shares, or returns to. Length, richness, and effort spent never promote an answer into a page.
>
> This rule is the canonical source for the `flow-report` auto-invoke trigger. Other files reference it; the conditions are not restated elsewhere.

### Routing by output shape

| Output shape | Format | Skill |
|---|---|---|
| Artifact with an audience or a second reading beyond this thread | Self-contained HTML | `flow-report` |
| Interactive UI with live state (sliders, knobs, live re-render) | Self-contained HTML with JS | `playground` |
| UI component, landing page, visual artifact for production | Production HTML/JSX | `frontend-design` / `canvas-design` |
| An answer to the user — however long, rich, or hard-won | In-thread prose | (none) |
| Agent-consumed, or a short single-purpose note | Markdown | (none) |

### Trigger — deliverable artifacts only

**The gate is the artifact test, and it decides alone: does this output have a life outside the conversation?** It qualifies only when the user asked for a document, report, or page, OR it has a named audience beyond this thread (a client deliverable, a spec others implement from, a reference to revisit weeks later). Answering the user's question is not that — no matter how long the answer, how many kinds of information it carries, or how much work produced it. Effort spent is never the argument: a costly analysis reported back in-thread is still an answer.

These shapes stay in-thread by default, and the thresholds below never override the line above: an analysis or diagnosis, audit or review findings reported back to you, a status summary, an option comparison for a decision you make now, an explanation of how something works, a post-incident account discussed here, a plan presented for approval in conversation.

Once the artifact test passes, `flow-report` renders it when ALL of these also hold — necessary conditions, never sufficient on their own: **>300 words**; **>2 distinct sections**; **2+ kinds of information** (prose + table, code + diagram, mockup + spec); read by a **human**, not an agent or a parser; **no live interactive state** (that routes to `playground`); no UI skill (`frontend-design`, `canvas-design`) already owns the domain. `flow-report` owns the mechanics — template, sections, output location, export.

### Carve-outs — any match means Markdown even when the trigger fires

- **Versioned human-edited docs:** `CLAUDE.md`, `AGENTS.md`, `README.md`, `LICENSE`, ADRs, `_support/docs/*`.
- **Agent-to-agent handoffs:** subagent input, task descriptions, tool input → Markdown or JSON.
- **`~/.claude/plans/*.md`** — plan mode native (Shift+Tab) writes `.md`.
- **Short prose:** under ~300 words OR single-section.
- **Harness configs:** `settings.json`, hook scripts, MCP configs, skill frontmatter.
- **Discharges:** a flow skill declaring its own output format, or an explicit user request ("respond here in text" / "make it a page"), wins over this rule.

### Layout floor — any HTML page rendered for the user

Applies to every human-facing HTML deliverable whatever route produced it: `flow-report` output, a published **Artifact**, a one-off page. A design skill's own measure guidance (`artifact-design`'s ~65ch) does NOT override it.

- **The shell is the measure.** One container; headings, prose, lists and tables all fill it. Never cap prose narrower than its container — a text column with a dead band beside it is the defect. Shorter lines wanted → narrow the shell.
- Base `font-size: 18px` / `line-height: 1.6` (16px below 720px); tables never below ~0.95rem.
- Product UI is the opposite case and keeps its 45–75ch cap (`languages/ui-visual-design.md > Typography`).

### In-thread answers — a report's substance, none of its ceremony

An answer the gate keeps in-thread carries everything a page would have carried, minus the packaging: no executive summary, no restatement of the question, no "what follows is…" preamble, no closing recap of what was just said. Headers and lists appear only where the content already has seams — a compact list, or two or three short headers, is the ceiling; an answer whose parts are not genuinely separate takes none. Trim by dropping detail that would not change what the reader does next, never by compressing sentences into fragments, arrow chains, or abbreviations. Explanatory means the reasoning that would change the reader's decision travels with the conclusion: why this over the obvious alternative, what it costs, and what would change the answer.

### Conversational ASCII diagrams

**Explaining how something works, why it broke, or how parts relate ships WITH a small ASCII diagram by default** — drawing is the norm, not a fresh judgment call each time. The triggers are conversational, not abstract: "how does X work", "why did Y happen", "what's the flow", "walk me through it", "explain this" — and their Spanish equivalents ("cómo funciona", "por qué pasó", "explícame"). Shapes that qualify: **branching** (fallbacks, error paths, mutually exclusive outcomes), **fan-out / fan-in** (one component with N consumers; parallel paths side by side), **cross-layer flow** (data traversing 3+ layers), **staged pipelines with gates** (promotion chains, CI stages — what blocks what), **dependency relationships** (what points at what, what breaks when one moves), **before/after** when a change rearranges the shape. Two distinct topologies in one answer earn two diagrams.

**A diagram is not decoration and never counts against concision.** The brevity directives — in this file, in `concise-first`, in any active output style — do NOT suppress it: a diagram typically replaces more prose than it costs. Skip it only for a genuinely linear sequence (a short ordered list already serializes that) or a bare-fact question. Keep it to a handful of labeled nodes, accompanying the prose rather than replacing it. Prompt-convention: nothing enforces this but the reading.

HTML costs 2-4× tokens and generation time, so the gate exists to make that cost buy an artifact someone will actually keep. At the edge, answer in-thread.
