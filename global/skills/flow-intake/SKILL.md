---
name: flow-intake
description: >
  Analyze and improve a client requirements document at project intake (F1 of the flow
  pack): completeness scoring, implicit scope and missing business rules surfaced, and
  client-ready questions split into blocking vs nice-to-know. Use when receiving or
  drafting the initial document for a potential or new project.
argument-hint: "<path-to-requirements-doc>"
disable-model-invocation: true
---

# /flow-intake — requirements intake

The only flow skill that runs BEFORE a workspace exists. If
`<project>/_support/PROJECT.md` exists, follow the flow contract normally
(`~/.claude/skills/flow-core/SKILL.md`); if not, operate standalone — write outputs next
to the source document and note that `/flow-kickoff` will absorb them.

`$1` is the document path (md, pdf, docx). Missing → ask for it; this skill does not
start from a verbal description alone (get it written first, even rough — the document is
what the client implicitly agreed to).

## Phase 1 — Analysis (delegated)

Dispatch **requirement-analyst** per the handoff protocol
(`~/.claude/skills/flow-core/references/handoff-protocol.md`) with:
- The document path, and the rubric path
  (`${CLAUDE_SKILL_DIR}/references/requirements-rubric.md`)
- Whatever context exists about the client (prior projects in the same `<group>/`,
  related conversations the user mentions)
- The intent: "your improved document seeds the specs phase; your questions are what the
  user takes to the next client call — write them client-ready, in the client's language"

Do not analyze the document yourself in the main thread.

## Phase 2 — Outputs

From the analyst's report, write (next to the source doc, or `_support/docs/` if the
workspace exists):
- `requirements-v2.md` — the improved document, additions marked `[PROPUESTO]` so the
  client's words stay distinguishable from ours
- `open-questions.md` — blocking questions first, then nice-to-know each with its
  proposed default assumption

## Phase 3 — Gate

Present to the user: the scorecard, the count of additions, and the **blocking questions
verbatim** — they decide whether to take them to the client now or accept proposed
defaults and proceed to `/flow-kickoff` at risk. Record whichever they choose in
`open-questions.md` (an accepted assumption is a decision: it must survive to the specs
repo later — promotion note at the top of the file).

If the workspace exists, CLOSE per the contract; otherwise end by suggesting
`/flow-kickoff <group> <project>` as the next step.
