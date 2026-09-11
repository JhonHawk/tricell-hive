
# Code-Search Routing

> Route by operation type, not by tool preference. Prompt-convention.

## Routing by operation

| Operation | Tool |
|---|---|
| Exhaustive literal search, usage counts, existence checks | `rg` — the only valid evidence for an absence claim |
| Intent discovery, unknown terminology, legacy/untyped code | `rg` with a broad vocabulary sweep (English AND Spanish domain terms, plural/singular, abbreviations), then Read the hits; delegate to `sdd-explore` when the sweep spans repos or the terminology is unknown |
| Callers / impact / affected tests, structural questions on a known symbol | `rg` for the call sites, then Read; the test selection comes from the runner (`vitest related`, `jest --findRelatedTests`, `turbo --affected`), never from a code index |
| Ambiguous scope questions (mixed design-vs-code, "where do we handle…") | rg + Read first; concretize the question before widening the sweep |
| Conclusions, flows, "does X exist?" answered for a decision | agent loop (sdd-explore / Explore) + review-refuter on negative claims |
| Sweep spanning 2+ repos, or 3+ concurrent searches (subagent fan-out) over the workspace | `tgw` when installed — same flags and output as `rg`, served from the trigram index of `~/Development/projects`; falls back to `rg` by itself when its server is down. Single-repo searches stay on `rg`. Under evaluation: `_support/docs/tgrep-workspace-server.md` in the hive |

## Anti-conclusion discipline

A search hit is a pointer, never a verdict:

- **An absence claim needs an exhaustive `rg` sweep** — broad vocabulary, English AND Spanish domain terms, 0 hits. Nothing weaker supports "X does not exist". **An assumed absence counts as a claimed one:** fixing the occurrence that failed and moving on asserts there are no others without ever saying so, so the sweep is owed before acting, not before writing the sentence (`quality/critical-thinking.md > Count the instances`).
- **Verify before building on a hit.** It may be dead code or the wrong direction — Read the file and check its call sites before anchoring a conclusion on it.
- **A tool that infers relations instead of resolving them is not evidence.** Any index or assistant that answers "who calls this" without the compiler's own resolution guesses on ambiguity, and a wrong guess is indistinguishable from a right one in its output. Confirm the target file before acting on a relation it reports.

## Model floor for discovery agents

Discovery agents run on **sonnet** (Codex: luna at `max` reasoning via the tier map). The floor is the reasoning depth, not the model size: luna below `max`, and haiku, are NOT approved for discovery — the discipline above is the safety mechanism and degrades first on shallow reasoning. Mechanical harness roles may run at lower efforts.
