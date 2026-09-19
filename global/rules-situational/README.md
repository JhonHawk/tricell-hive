# rules-situational — reachable through a router skill, or carried by an agent

Rules here are **not deployed to `~/.claude/rules/`**, so they are never always-on and never
symlinked into `~/.grok/rules/`. `harness/build.py` injects them into a router skill's
`references/` exactly like the rules under `global/rules/`; the only difference is the
absence of an always-on copy. The pack-only texts below are the exception: no router serves
them.

## When a rule belongs here instead of in `global/rules/`

A rule leaves `global/rules/` when it is situational, and it lands here rather than gaining
`paths:` when its trigger is an **action**, not a file kind:

| Trigger | Home | Mechanism |
|---|---|---|
| An action with a safety consequence (destructive op, git verb, secret) | `global/rules/` | always-on |
| Touching a kind of file | `global/rules/` + `paths:` | native path-scoping |
| An intent with no file behind it (delegating, planning, saving memory) | **here** | router skill only |
| Being a specialized executor agent (`packs:` in its frontmatter) | **here**, no `globs:` | inlined into the agent at build |

`paths:` cannot express "I am about to delegate". Forcing an artificial glob would make the
rule load on the wrong files and still miss the moment it matters.

A second shape lives here since F2: the **detail half of a demoted always-on rule**. The
gate stays always-on as a stub in `global/rules/` (so Grok's flat symlink and Claude Code's
unconditional load keep it); the mechanics land here and reach every harness through the
owning skill's `references/` (`browser-automation-reference.md` → `language-rules`,
`communication-format-mechanics.md` → `flow-report`, `unattended-autonomy-mode.md` →
`unattended-delegation`).

A third shape: **pack-only texts** (`agent-core-gates.md`, `test-gate.md`). No router, no
globs, never hook-delivered — `harness/build.py` inlines them into every agent that declares
them in `packs:`, so their presence is a build-time guarantee, not a routing probability.
`agent-core-gates` is required in every packed agent: on Claude Code a packed agent runs with
`omitClaudeMd: true` and this text is what puts the gates back. Accepted cost: Codex, opencode,
Grok and Pi have no per-agent corpus switch, so a packed agent there receives these gates on
top of the core it already loads.

## The cost of moving a rule here

It stops being guaranteed. A rule here is present only when the model invokes its router, so
never move one whose absence is unsafe — that is what always-on is for. The compensating
control is a one-line pointer in `global/CLAUDE.md` naming the router and its trigger.
