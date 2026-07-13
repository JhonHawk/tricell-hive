# Agent-Verifiable Quality Gates — Decision Record

> **For agentic workers:** This is a decision record (ADR), not an implementation plan. It captures WHAT changed in the global rules and WHY. The mechanical edits live in the implementation plan that accompanies this record.

**Status:** accepted (2026-06-22)

**Scope:** `global/rules/quality/{testing,security,critical-thinking,development-principles}.md`, `global/CLAUDE.md`, `global/rules/workflow/agent-routing.md`, the flow skills that referenced maturity tiers, and the condensed cross-harness mirror `harness/AGENTS.md`.

---

## Context

Two problems converged.

1. **`project-maturity` was load-bearing on paper, unused in practice.** The tier ladder (`spike`/`poc`/`mvp`/`production`) relaxed test breadth, docs, and scope by declared stage. In practice no client workspace declared a tier other than the `production` default, so the machinery — a dedicated rule file, a `## Project Maturity` block in every kickoff template, tier carve-outs threaded through `testing.md`, `critical-thinking.md`, and `security.md`, a ledger row, a handoff field — bought nothing but cross-reference surface and the standing risk that a successful POC gets silently promoted with zero tests.

2. **The rules are read by AI agents, not humans, and were written as if read by humans.** `testing.md` *preached the TDD ritual* ("tests are part of the implementation", "write the failing test first") without imposing anything an agent's output could be *checked against*. For a human, the ritual and the result are coupled by professional discipline. For an agent under task pressure they are not — the agent can perform the ritual's appearance while producing tests that verify nothing, or skip it entirely.

The trigger was a question about whether to mandate TDD. Research reframed it.

## Finding — instructing the ritual backfires; verifiable context is what helps

- **TDAD** (arXiv [2603.17973](https://arxiv.org/abs/2603.17973)): on smaller models, instructing the TDD ritual *by itself increased* regressions (+9.94%). What cut regressions ~70% was giving the agent **verifiable context about which tests to check** (impact analysis over a dependency graph), not the procedural instruction to write tests first.
- **Reward hacking is measured, not hypothetical.** METR reports o3 and Claude 3.7 reward-hack in 30%+ of runs on some tasks; observed tactics include a `conftest.py` that rewrites outcomes to "passed" and reading `git log` to copy the expected answer. An agent under pressure modifies or weakens tests to make them pass rather than fixing the code.
- **The industry antidote is a dual metric, not a ritual.** SWE-bench Verified scores a patch on **fail-to-pass** (the bug's tests now pass) AND **pass-to-pass** (previously-passing tests still pass) — precisely to make "I made the tests green" insufficient on its own. Kent Beck (*Augmented Coding*) calls TDD a "superpower" with agents while explicitly warning about *cheating*.

**Design conclusion:** the rules should not *preach* the ritual — they should impose a **verifiable gate on the agent's output**. That gate is what fills the hole left by removing maturity-scaled testing.

## Decisions

1. **Gates, not tiers.** Remove `project-maturity` entirely — no stage-based escape valve. Tests are required whenever a change alters behavior, full stop. The one genuinely load-bearing piece of the tier model — the **exposure-gated security floor** (auth, CVE gates, injection/SSRF apply whenever the system touches real user data, real production systems, or the public network) — survives intact, relocated into `security.md` and `critical-thinking.md`, decoupled from any mention of tiers.

2. **Verifiable gates over the output, not the ritual.** Require the *verifiable result* (fail-to-pass + pass-to-pass; the verifier runs the tests and watches for test-gaming). The test-first ritual is recommended, not mandated — TDAD shows the ritual alone is neutral-to-harmful; the checkable outcome is what carries the guarantee.

3. **Decision recorded before the edit.** This ADR lands in `_support/spec/` before the rule files change — the rules are high-leverage (they deploy to Claude Code, Codex, and opencode), so the "why" is written down first.

## The unifying concept — "verifiable test gate"

Lives **canonically in `testing.md`**; the other rules reference it without restating it (repo style). Operational definition:

> A behavior change ships with tests that **would fail without it** (fail-to-pass) and **do not weaken the existing ones** (pass-to-pass). Verification **runs** the tests; it does not trust the report. Disabling, weakening, or bypassing tests (skip/xit, `--no-verify`, manipulating the runner config) to make a suite "pass" is test-gaming, not a fix.

It replaces the deleted notion of "tests scaled by tier" and ties together three rules that were previously loose:

- `testing.md` — *what* the gate demands (the dual metric, anti-gaming).
- `agent-routing.md` — *who* verifies it (a fresh-context verifier that runs the tests, not the implementing thread; the subagent's report is a claim to check).
- `development-principles.md` — *what violates it* (test-gaming is the agent's version of silencing a guardrail).

Anchored to **behavior change** (non-trivial), using the canonical definition of *trivial* in `critical-thinking.md`, so it does not over-apply to typos, renames, or formatting.

## What changed, by file

| File | Change |
|---|---|
| `global/rules/quality/project-maturity.md` | deleted |
| `global/CLAUDE.md` | removed `## Project Maturity Tiers` |
| `global/rules/quality/security.md` | exposure-gated floor promoted to its own `### Exposure-gated security floor` subsection; tier references dropped |
| `global/rules/quality/critical-thinking.md` | pre-ship ownership test reframed as "scaled by exposure" (Q1-2 gated on real exposure, Q3 absolute); maturity cite dropped |
| `global/rules/quality/testing.md` | tier carve-outs removed; Execution Scope reframed defensive→enabling; **verifiable test gate added** as canonical subsection |
| `global/rules/workflow/agent-routing.md` | fresh-context verification strengthened: verifier runs the tests (fail-to-pass + pass-to-pass) and checks for test-gaming; implementer handoff carries the tests added + run evidence |
| `global/rules/quality/development-principles.md` | TDD-refactor carve-out on "Only change what was asked"; test-gaming added to "Fix the cause, not the check"; agent over-abstraction angle on Rule of Three |
| flow skills + `harness/AGENTS.md` | maturity references removed; gate + Execution Scope reflected in the condensed mirror |

## Backing research

Recorded in `_support/docs/methodology-bibliography.md` (section "Agentes IA + evals"):

- **K. Beck, *Augmented Coding* (Tidy First? substack, 2025)** — TDD as agent "superpower"; the cheating warning. Backs the verifiable test gate's recommend-the-ritual / mandate-the-result split.
- **SWE-bench / SWE-bench Verified (Jimenez et al., 2023; OpenAI Verified subset, 2024)** — fail-to-pass + pass-to-pass dual metric. Backs the gate's definition.
- **TDAD (arXiv 2603.17973)** — ritual-alone raises regressions; verifiable impact context lowers them. Backs *why* the gate targets the result, not the ritual; and the Execution Scope "let the tool compute the impact graph" angle.
- **METR (reward hacking, 2025); DeepMind specification gaming; L. Weng, "Reward Hacking in RL" (2024)** — the failure mode the anti-gaming clauses defend against.
- **Rothermel & Harrold, "A safe, efficient regression test selection technique", *ACM TOSEM* 6(2), 1997** — safe regression test selection by import/dependency graph. Backs Execution Scope's "prefer the tool-computed subset to the agent's guess".
- **M. Fowler, "Is High Quality Software Worth the Cost?" (martinfowler.com, 2019)** — high internal quality is not a stage-gated luxury; it pays back almost immediately. Backs "gates, not tiers".

## Consequences

- **Simpler rule graph.** One canonical gate, three referencing rules; the maturity cross-reference web is gone.
- **No early-stage escape valve.** A genuine throwaway spike no longer has a declared way to skip tests. Accepted: the cost of an unused relaxation lane outweighed its benefit, and the exposure-gated floor (the part that mattered) is preserved.
- **Inert legacy declarations.** A `## Project Maturity` block left in some project's `CLAUDE.md`/`AGENTS.md` is simply ignored by the new rules — it does not break anything.
- **High blast radius.** Deploys to three harnesses on the user's next `/deploy-global`. Mitigated by keeping the change isolated and the `harness/AGENTS.md` byte budget (`~/.codex/AGENTS.md` < 49152) measured.
