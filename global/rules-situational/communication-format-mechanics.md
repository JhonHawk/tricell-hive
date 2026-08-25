# Communication Format — rendering mechanics

> Loaded as a `flow-report` skill reference on every harness, alongside `communication-format.md` (the always-on half, canonical for the trigger). This file owns the rendering mechanics: the HTML layout floor, the in-thread prose form, the answer-length bullets, and the diagram norm. Read it whenever rendering a page or shaping a substantial answer.

## Layout floor — any HTML page rendered for the user

Applies to every human-facing HTML deliverable whatever route produced it: `flow-report` output, a published **Artifact**, a one-off page. A design skill's own measure guidance (`artifact-design`'s ~65ch) does NOT override it.

- **The shell is the measure.** One container; headings, prose, lists and tables all fill it. Never cap prose narrower than its container — a text column with a dead band beside it is the defect. Shorter lines wanted → narrow the shell.
- Base `font-size: 18px` / `line-height: 1.6` (16px below 720px); tables never below ~0.95rem.
- Product UI is the opposite case and keeps its 45–75ch cap (`languages/ui-visual-design.md > Typography`).

## In-thread answers — a report's substance, none of its ceremony

An answer the gate keeps in-thread carries everything a page would have carried, minus the packaging: no executive summary, no restatement of the question, no "what follows is…" preamble, no closing recap of what was just said. Headers and lists appear only where the content already has seams — a compact list, or two or three short headers, is the ceiling; an answer whose parts are not genuinely separate takes none. Trim by dropping detail that would not change what the reader does next, never by compressing sentences into fragments, arrow chains, or abbreviations. Explanatory means the reasoning that would change the reader's decision travels with the conclusion: why this over the obvious alternative, what it costs, and what would change the answer.

## Answer length — set by the question, not by the work behind it

A bare-fact or yes/no question closes in 1-3 sentences of plain prose; a substantive one carries the full substance the section above requires. Effort spent is not a reason to write more, exactly as it is not a reason to render a page.

- **Lead with the result** — what happened, or the answer itself, in the first sentence. No "let me…" preamble, no plan narration, no step-by-step account of the tools run: outcomes, decisions, and what the user must act on.
- **Full detail on request.** A request to explain, expand, or justify is answered completely; brevity never withholds what was asked for.
- **Never trade correctness for brevity.** Failing output keeps its actual text, security warnings keep their reasoning, destructive-action confirmations keep what will happen and how to revert, per-criterion state keeps its per-criterion form. Trim detail that would not change what the reader does next — never the evidence a decision rests on.
- **Precedence:** where the harness has a native output style (Claude Code `outputStyle`), it states this same thing and wins on wording. This section is the floor where there is none (Grok, opencode) or where the native switch only tunes the model's own verbosity (Codex `personality`, `model_verbosity`).

## Conversational ASCII diagrams

**Explaining how something works, why it broke, or how parts relate ships WITH a small ASCII diagram by default** — drawing is the norm, not a fresh judgment call each time. The triggers are conversational, not abstract: "how does X work", "why did Y happen", "what's the flow", "walk me through it", "explain this" — and their Spanish equivalents ("cómo funciona", "por qué pasó", "explícame"). Shapes that qualify: **branching** (fallbacks, error paths, mutually exclusive outcomes), **fan-out / fan-in** (one component with N consumers; parallel paths side by side), **cross-layer flow** (data traversing 3+ layers), **staged pipelines with gates** (promotion chains, CI stages — what blocks what), **dependency relationships** (what points at what, what breaks when one moves), **before/after** when a change rearranges the shape. Two distinct topologies in one answer earn two diagrams.

**A diagram is not decoration and never counts against concision.** The brevity directives — in this file, in `concise-first`, in any active output style — do NOT suppress it: a diagram typically replaces more prose than it costs. Skip it only for a genuinely linear sequence (a short ordered list already serializes that) or a bare-fact question. Keep it to a handful of labeled nodes, accompanying the prose rather than replacing it. Prompt-convention: nothing enforces this but the reading.
