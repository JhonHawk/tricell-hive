---
alwaysApply: true
---

# Code-Search Routing

> Route by operation type, not by tool preference. Deny/warn per repo is enforced deterministically by the `bash-policy` hook (its repo map); everything else is prompt-convention.

## Routing by operation

| Operation | Tool |
|---|---|
| Exhaustive literal search, usage counts, existence checks | `rg` — the only valid evidence for an absence claim |
| Intent discovery, unknown terminology, legacy/untyped code | `rg` with a broad vocabulary sweep (English AND Spanish domain terms, plural/singular, abbreviations), then Read the hits; delegate to `code-scout` when the sweep spans repos or the terminology is unknown |
| Callers / impact / affected tests, structural questions on a known symbol | `codegraph` (`callers`, `impact`, `affected`, `explore`) |
| Ambiguous scope questions (mixed design-vs-code, "where do we handle…") | rg + Read first; add index tools only after the question is concretized |
| Conclusions, flows, "does X exist?" answered for a decision | agent loop (code-scout / Explore) + finding-refuter on negative claims |

## Contraindications

- **CodeGraph in legacy/untyped JS repos**: its anchors mislead there — use `rg` plus Read. The routing hook denies it in declared legacy repos.

## Anti-conclusion discipline

An index result is a pointer, never a verdict:

- **Never conclude absence from an index.** "X does not exist" requires an exhaustive `rg` sweep (broad vocabulary, English AND Spanish domain terms, 0 hits). Index coverage is partial by design.
- **Verify before building on a hit.** A plausible index hit may be dead code or the wrong direction — Read the file and check callers before anchoring a conclusion on it.
- **On conflict between an index and disk (`rg`/Read), disk wins.**

## Model floor for discovery agents

Discovery agents run on **sonnet** (Codex: luna at `max` reasoning via the tier map). The floor is the reasoning depth, not the model size: luna below `max`, and haiku, are NOT approved for discovery — the discipline above is the safety mechanism and degrades first on shallow reasoning. Mechanical harness roles may run at lower efforts.
