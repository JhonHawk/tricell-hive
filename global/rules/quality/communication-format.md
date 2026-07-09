---
alwaysApply: true
---

## Communication Format

> Prefer self-contained HTML over Markdown for substantial multi-modal human-targeted output; Markdown for short, single-purpose, or machine-consumed output.
>
> This rule is the canonical source for the `flow-report` auto-invoke trigger. Other files reference it; the conditions are not restated elsewhere.

### Routing by output shape

| Output shape | Format | Skill |
|---|---|---|
| Static rich deliverable (>300 words, multi-modal, human-targeted) | Self-contained HTML | `flow-report` |
| Interactive UI with live state (sliders, knobs, live re-render) | Self-contained HTML with JS | `playground` |
| UI component, landing page, visual artifact for production | Production HTML/JSX | `frontend-design` / `canvas-design` |
| Short, single-section, conversational, or agent-consumed | Markdown or inline prose | (none) |

### Trigger — deliverable artifacts only

The trigger applies to **deliverable artifacts** — a report, audit, brief, plan, or spec the user will keep, share, or return to. A conversational answer never auto-elevates to HTML on length alone: a long in-turn explanation, review verdict, or status summary stays prose. Discharges: a flow skill that declares its own output format wins over this rule; an explicit inline request ("respond here in text" / "make it a page") wins over any default.

Render the artifact via `flow-report` when **ALL** of these hold:
- **>300 words** AND **>2 distinct sections**
- Mixes **2+ kinds of information** (prose + table, code + diagram, mockup + spec, status + chart)
- Read or shared by a **human**, not consumed by another agent or parsed by a tool
- **No live interactive state** — sliders, knobs, live re-render route to `playground` instead
- No existing UI skill (`frontend-design`, `canvas-design`) already owns the domain

The `flow-report` skill owns the mechanics: template, sections, output location and naming, export pattern.

### Carve-outs (always Markdown or plain text)

Priority over the default — any match means NO HTML even when the trigger threshold is met:

- **Versioned human-edited docs.** `CLAUDE.md`, `AGENTS.md`, `README.md`, `LICENSE`, ADRs, `_support/docs/*`.
- **Agent-to-agent handoffs.** Subagent input, structured task descriptions, tool input → Markdown or JSON.
- **`~/.claude/plans/*.md`.** Plan mode native (Shift+Tab) writes `.md` by Claude Code convention.
- **Short prose.** Under ~300 words OR single-section: inline responses, status checkpoints, confirmation summaries.
- **Harness configs.** `settings.json`, hook scripts, MCP configs, skill frontmatter — these have their own format.

### Cost — and the edge rule

HTML costs **2-4× tokens** vs equivalent Markdown and **2-4× generation time**; the trigger exists so that cost always buys a deliverable. At the edge — conditions arguably met, deliverable status unclear — Markdown. When in doubt, Markdown.
