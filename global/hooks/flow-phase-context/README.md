# flow-phase-context

UserPromptSubmit hook (Claude Code only) — the **offer layer** of the flow pack
("organic entry" design, 2026-07-13). In any flow workspace (ledger `_support/PROJECT.md`
at or above cwd), it injects the ledger's `Current phase` and `Next suggested` lines so
the model can OFFER the right `/flow-*` command in plain conversation. Flow skills stay
user-gated (`disable-model-invocation`); this hook makes the *state* ambient, and the
offer-don't-invoke behavior is codified once in `global/CLAUDE.md > Skill Auto-invocation`.

Sibling of `flow-plan-injector` (same ledger walk-up, non-blocking exit 0, stdin cwd),
without the plan-mode gate. Both inherit the lesson of the retired `flow-route-reminder`
(2026-07-10): inject **facts/state**, never per-prompt routing instructions (documented
failure modes: ignored injected instructions, false positives, token pollution —
anthropics/claude-code #19659/#17804).

Design decisions (adversarial investigation, 2026-07-13):
- **Once per session, keyed to ledger mtime** — re-fires after a phase transition
  mid-session instead of leaving stale state injected; otherwise silent.
- **Silent everywhere else**: no ledger → exit 0; ledger without `Current phase` /
  `Next suggested` fields → exit 0 (no noise for pre-template ledgers).
- **Advisory payload** — the injected text itself flags that the ledger can be stale;
  the model verifies before relying on it (`memory-routing.md`: records are claims).
- **Never an execution vector**: the payload never instructs invoking a skill;
  `/flow-deploy prod` and `/deploy-global` are excluded from automatic offers by the
  global rule this hook pairs with.

Codex and opencode have no UserPromptSubmit equivalent — they get the same organic
behavior through the flow-phase-offering block in the workspace `AGENTS.md` written by
`/flow-kickoff` (the cross-harness carrier).

Deployed by `/deploy-global` (script → `~/.claude/hooks/`, block merged into
`settings.json` from `settings-config.json`).
