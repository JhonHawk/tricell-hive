# rules-situational — reachable only through a router skill

Rules here are **not deployed to `~/.claude/rules/`**, so they are never always-on and never
symlinked into `~/.grok/rules/`. `harness/build.py` injects them into a router skill's
`references/` exactly like the rules under `global/rules/`; the only difference is the
absence of an always-on copy.

## When a rule belongs here instead of in `global/rules/`

A rule leaves `global/rules/` when it is situational, and it lands here rather than gaining
`paths:` when its trigger is an **action**, not a file kind:

| Trigger | Home | Mechanism |
|---|---|---|
| An action with a safety consequence (destructive op, git verb, secret) | `global/rules/` | always-on |
| Touching a kind of file | `global/rules/` + `paths:` | native path-scoping |
| An intent with no file behind it (delegating, planning, saving memory) | **here** | router skill only |

`paths:` cannot express "I am about to delegate". Forcing an artificial glob would make the
rule load on the wrong files and still miss the moment it matters.

## The cost of moving a rule here

It stops being guaranteed. A rule here is present only when the model invokes its router, so
never move one whose absence is unsafe — that is what always-on is for. The compensating
control is a one-line pointer in `global/CLAUDE.md` naming the router and its trigger.
