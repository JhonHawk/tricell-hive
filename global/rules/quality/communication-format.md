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
| Static rich document (>300 words, multi-modal, human-targeted) | Self-contained HTML | `flow-report` |
| Interactive UI with live state (sliders, knobs, live re-render, export-as-prompt) | Self-contained HTML with JS | `playground` |
| UI component, landing page, visual artifact for production | Production HTML/JSX | `frontend-design` / `canvas-design` |
| Short, single-section, conversational, or agent-consumed | Markdown or inline prose | (none) |

### Default: HTML for substantial human-targeted output

Render output as a self-contained `.html` file via `flow-report` when **ALL** of these conditions hold:

- Output is **>300 words** AND has **>2 distinct sections**
- Output mixes **2+ kinds of information** (prose + table, code + diagram, mockup + spec, status + chart)
- The artifact will be **read or shared by a human**, not consumed by another agent or parsed by a tool
- The output has **no live interactive state** — if the user needs sliders, knobs, or live re-render, route to `playground` instead
- No **existing UI skill** (`frontend-design`, `canvas-design`) already owns the domain

When the rule fires, invoke skill `flow-report` for mechanics (template, sections, export pattern, output dir).

### Carve-outs (always Markdown or plain text)

The carve-outs have **priority over the default**: if any of these match, do NOT apply HTML even if the trigger threshold is met.

- **Versioned human-edited docs.** `CLAUDE.md`, `AGENTS.md`, `README.md`, `LICENSE`, ADRs, `_support/docs/*`.
- **Agent-to-agent handoffs.** When the output is consumed by another agent (subagent input, structured task description, tool input), use Markdown or JSON.
- **`~/.claude/plans/*.md`.** Plan mode native (Shift+Tab) writes `.md` files by Claude Code convention.
- **Short prose.** Outputs under ~300 words OR single-section. Inline conversational responses, status checkpoints, confirmation summaries, one-line answers.
- **Harness configs.** `settings.json`, hook scripts, MCP configs, skill frontmatter — these have their own format.

### Output location

- **Reports, audits, briefs, research, explainers** → `_support/workspace/{kebab-slug}.html`
- **Plans persisted alongside the codebase** → `_support/plan/{kebab-slug}.html` when substantial; `.md` for short plans matching a carve-out
- **Embedded images** → sibling `images/` subfolder, referenced relatively
- After writing, tell the user the path and suggest `open <path>` or upload destination for sharing

### Cost acknowledgment

HTML output costs **2-4× tokens** vs equivalent Markdown and **2-4× generation time**. The triggers above ensure the cost is justified by the artifact's usefulness. When in doubt, default to Markdown.
