# rules-situational — the rule-text store: delivered by hook, router, or pack

Rules here are **not deployed to `~/.claude/rules/`**, so they are never always-on and never
symlinked into `~/.grok/rules/`. No harness loads them by file kind. They reach an agent
three ways, and `harness/build.py` is what wires all three:

| Channel | How it arrives | Applies to |
|---|---|---|
| `rule-delivery` hook | holds a matching write until the deployed reference is read, then the re-issued call passes (Claude Code, Grok, Codex, PI; a read, inside a read-only agent scoped by `readers`) | a file here **with `globs:`** |
| Router skill | injected into a `references/` directory (`SKILL_REFERENCE_INJECTIONS`); present only when the model invokes the router | every file here except the pack-only texts |
| `packs:` | inlined into the agent's prompt at build — a build-time guarantee, not a routing probability | the agents that declare it |

**The contract: `globs:` is what makes a file hook-deliverable.** Without it the hook never
fires for that rule, whatever its content — the router (or a pack) is then the only channel.

## What belongs here instead of in `global/rules/`

`global/rules/` holds only the always-on rules: safety gates and policy whose trigger is an
action rather than a file, paid every session in every project. Everything else lives here,
in one of four shapes:

- **Intent-triggered** (delegating, planning, saving memory) — no `globs:`, router only.
  `globs:` cannot express "I am about to delegate"; an artificial glob would fire on the
  wrong files and still miss the moment that matters.
- **Glob-scoped** (`globs:`) — the language, workspace and devops rules: hook, router and
  packs all deliver them.
- **The detail half of a demoted always-on rule.** The gate stays always-on as a stub in
  `global/rules/`; the mechanics land here and reach every harness through the owning
  skill's `references/` (`browser-automation-reference.md` → `language-rules`,
  `communication-format-mechanics.md` → `flow-report`, `unattended-autonomy-mode.md` →
  `unattended-delegation`).
- **Pack-only texts** (`agent-core-gates.md`, `test-gate.md`). No router, no globs, never
  hook-delivered. `agent-core-gates` is required in every packed agent: on Claude Code a
  packed agent runs with `omitClaudeMd: true` and this text is what puts the gates back.
  Accepted cost: Codex, opencode, Grok and Pi have no per-agent corpus switch, so a packed
  agent there receives these gates on top of the core it already loads.

## The cost of living here

A rule here is present only when its channel fires — a router the model must invoke, or a
hook the model must obey after a hold. Never move one whose absence is unsafe; that is what
always-on is for. The compensating control is a one-line pointer in `global/CLAUDE.md`
naming the router and its trigger.
