# /flow-start — Stage A: Intake (produce the requirements doc)

The idea/brief arrives in **any** form: raw notes, a conversation, a client email, a rough
document. This stage *produces* the requirements doc from that input — it never demands a
written one as a prerequisite (that demand is exactly what got the old step bypassed).

1. Dispatch **requirement-analyst** per the handoff protocol
   (`~/.claude/skills/flow-core/references/handoff-protocol.md`) with: the raw input,
   the rubric path (`${CLAUDE_SKILL_DIR}/references/requirements-rubric.md` — after deploy,
   `~/.claude/skills/flow-start/references/requirements-rubric.md`), whatever
   client context exists (sibling projects under the same `<group>/`), and the intent:
   "your improved document seeds the specs phase; your questions are what the user takes to
   the next client call — write them client-ready, in the client's language". Do not
   analyze the input yourself in the main thread.
2. From the analyst's report, write into `_support/docs/` (created in Stage B if not yet
   present — stage the files beside the source until then):
   - `requirements-v2.md` — the improved document, additions marked `[PROPUESTO]` so the
     client's words stay distinguishable from ours.
   - `open-questions.md` — blocking questions first, then nice-to-know, each with its
     proposed default assumption.
3. **Gate:** present the scorecard, the count of additions, and the **blocking questions
   verbatim**. The user takes them to the client now, or accepts the proposed defaults and
   proceeds at risk. **An accepted assumption is a decision** — it must survive to the specs
   repo later; record it at the top of `open-questions.md` as a promotion note.

Already have a satisfactory requirements doc → skip to Stage B, noting what you're reusing.
