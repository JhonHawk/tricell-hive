---
name: adversarial-research
description: >
  Adversarial multi-proposal research for a consequential question, claim, or design
  decision: N independent generators (one may be Codex via codex-rescue when the codex
  plugin is present) produce proposals in parallel without seeing each other; a
  finding-refuter cross-examines every claim against declared ground truth (repo files,
  docs, live state), attempting to refute each one; the main thread synthesizes a canon —
  refuted / weakened(+fix) / surviving / net-new — with a verdict on the most faithful
  proposal. Use when being wrong is expensive and a single investigation would anchor on
  one framing: architecture mappings, root-cause disputes, "which of these designs matches
  reality", migration/compat claims. Do NOT use for single-fact lookups, questions one
  grep or file read answers, or verifying one already-stated claim (dispatch
  finding-refuter directly per agent-routing).
---

# /adversarial-research — independent proposals, adversarial cross-exam, synthesized canon

One investigator anchors on one framing; this skill buys N framings and pays for them with
one adversarial pass. Explicit invocation IS the multi-agent opt-in
(`agent-routing.md > Workflow Tool vs Subagents vs Agent Teams`) — never auto-invoke this.
All agent prompts are written in English; user-facing output follows the conversation
language.

**Routing.** `$ARGUMENTS` is the question, claim, or decision verbatim. Recognize modifiers
by intent, not flag syntax: a generator count ("3 generators", "n=4") sets N (default 3,
min 2, max 5); "no codex" / "sin codex" skips the Codex slot. Empty `$ARGUMENTS` → ask what
the question is and stop.

## 1 — Frame

- Restate the question as ONE falsifiable sentence. If the user's phrasing bundles several
  questions, pick the consequential one and say so.
- Declare the **ground-truth corpus**: explicit paths/globs, docs, and read-only commands
  (git log, live-state reads) the refuter will treat as canonical. Declared once, pasted
  identically into every generator prompt and the refuter prompt. Genuinely ambiguous
  corpus → one `AskUserQuestion`; otherwise infer and state it.
- Assess the risk tier and set the refuter count M per the refuter budget in
  `agent-routing.md > Verification runs in fresh context` (do not restate its numbers —
  that rule is the single owner). State the plan: "N generators, M refuter(s)".

## 2 — Generate (parallel, independent)

- Dispatch ALL generators in ONE message as parallel `Agent` calls. Independence is
  mechanical: never show one generator another's output; never re-dispatch a generator
  after any other has returned.
- Slot assignment: prefer domain specialists per the `agent-routing.md` disambiguation
  table. One slot is `codex:codex-rescue` when the codex plugin is available and not opted
  out; absent → fill it with a Claude generator seeded with a deliberately distinct
  perspective (code-outward vs docs-outward vs constraints-first) and note the substitution
  in the synthesis header.
- Every generator prompt MUST contain: (a) the framed question; (b) the identical
  ground-truth corpus manifest; (c) the punishment clause, verbatim in intent: "An
  adversarial cross-examiner will attempt to refute every claim you make; a claim that
  implies wrong behavior counts against you. Cite file:line or command output for every
  claim. Say 'unknown' rather than guess."; (d) the required output shape: numbered
  claims plus a stated overall position.
- Garbage gate: a proposal that is empty, off-topic, or citation-free is dropped. ≥2 usable
  proposals → proceed; <2 → redispatch the failed slots ONCE; still <2 → abort and explain —
  never degrade silently into single-thread research.

## 3 — Refute

- Dispatch `finding-refuter` (M in parallel when M>1) with: all surviving proposals labeled
  A/B/C…, the same corpus manifest, and this instruction: for EVERY numbered claim, attempt
  refutation against ground truth and return a per-claim verdict — REFUTED (with the
  counterexample citation), WEAKENED (what is off + the minimal fix), or CONFIRMED
  (refutation attempted and failed; cite what was checked) — plus NET-NEW: true facts in
  the corpus that no proposal surfaced.
- Multiple refuters: apply the budget's majority rule per claim; a split with no majority →
  mark the claim CONTESTED for resolution in step 4.

## 4 — Synthesize (main thread — it holds the conversation context)

- Build the **canon table**: claim | source proposal(s) | verdict
  (refuted / weakened / surviving / net-new / contested) | ground-truth citation | fix
  applied (weakened only). Every row cites; a verdict without a citation is not canon.
- Resolve CONTESTED rows by reading the cited evidence directly; never leave a row
  contested in the final output.
- Produce a **per-proposal scorecard** (refuted/weakened/surviving counts) and a one-line
  **most-faithful verdict** with the reason.
- State the **recommendation**: the answer to the framed question supported ONLY by
  surviving + fixed + net-new claims, with a stated confidence and what would change it.
- Everything-refuted is a valid outcome, not a failure — the canon of refutations IS the
  deliverable ("the premises fail because…"). Offer ONE regeneration round with corrected
  premises, user-gated; never loop automatically.

## 5 — Deliver

- Route the synthesis format through `communication-format.md` (flow-report HTML when its
  trigger conditions hold; inline Markdown otherwise) — do not hardcode a format here.
- Always close inline, even when an HTML report is produced, with: the verdict, the
  refuted/weakened/net-new counts, and the recommendation in ≤3 sentences.
