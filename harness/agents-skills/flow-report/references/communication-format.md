
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

Applies to **deliverable artifacts** — a report, audit, brief, plan, or spec the user will keep, share, or return to. A conversational answer never auto-elevates on length alone: a long explanation, review verdict, or status summary stays prose.

Render via `flow-report` when **ALL** hold: **>300 words** AND **>2 distinct sections**; mixes **2+ kinds of information** (prose + table, code + diagram, mockup + spec); read by a **human**, not an agent or a parser; **no live interactive state** (that routes to `playground`); no UI skill (`frontend-design`, `canvas-design`) already owns the domain. `flow-report` owns the mechanics — template, sections, output location, export.

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

### Conversational ASCII diagrams

Within a prose answer, include a small ASCII diagram only when the explanation's topology is non-linear — the shapes prose serializes badly: **branching** (fallbacks, error paths, mutually exclusive outcomes), **fan-out / fan-in** (one component with N consumers; parallel paths compared side by side), **cross-layer flow** (data traversing 3+ layers). The diagram accompanies the prose, never replaces it.

HTML costs 2-4× tokens and generation time, so the trigger exists to make that cost buy a deliverable. At the edge, Markdown.
