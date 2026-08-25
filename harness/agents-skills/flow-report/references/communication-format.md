
## Communication Format

> **In-thread prose is the default.** HTML is for output with a life outside this conversation — something the user keeps, shares, or returns to. Length, richness, and effort spent never promote an answer into a page.
>
> This rule is the canonical source for the `flow-report` auto-invoke trigger. Other files reference it; the conditions are not restated elsewhere. Rendering mechanics (HTML layout floor, in-thread prose form, extended answer-length bullets, ASCII-diagram norm) live in `rules-situational/communication-format-mechanics.md`, a `flow-report` reference.

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

### Answer length — the in-thread floor

A bare-fact or yes/no question closes in 1-3 sentences of plain prose; a substantive one carries its full substance with none of a report's ceremony (no executive summary, no restated question, no closing recap). Lead with the result; full detail on request; never trade correctness for brevity — the evidence a decision rests on is never trimmed. Extended form (per-bullet floor, layout floor for HTML pages, diagram norm, precedence vs native output styles): `rules-situational/communication-format-mechanics.md`.

HTML costs 2-4× tokens and generation time, so the gate exists to make that cost buy an artifact someone will actually keep. At the edge, answer in-thread.
