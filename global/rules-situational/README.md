# rules-situational — the one rule-text store

Every rule text lives here, in one flat directory. Nothing here is deployed to
`~/.claude/rules/` and nothing is symlinked into `~/.grok/rules/`: no harness loads a file
here by file kind. A rule reaches a model through the channels below, and `harness/build.py`
wires all of them. **A rule with no channel is a build error** — "it is in
`global/rules-situational/`" reaches nobody.

| Channel | How it arrives | Declared by |
|---|---|---|
| Core include | a `global/core-sections/rule-<name>.md` section carries `include: rules-situational/<name>.md`; the build inlines the body into `global/CLAUDE.md`, always-on in every session | the core section (the rule file itself says nothing) |
| `rule-delivery` hook | denies the first matching write — or the first matching command — naming the reference to read; the re-issued call passes (Claude Code, Grok, Codex, PI) | `globs:` and/or `commands:` **plus** a router reference |
| Router skill | injected into a skill's `references/`; present only when the model invokes the router | `SKILL_REFERENCE_INJECTIONS` in `harness/build.py` |
| `packs:` | inlined into an agent's prompt at build — a guarantee, not a routing probability | the `packs:` line on the agent |

**A trigger counts only together with a router reference.** The hook can gate only what it can
tell the agent to READ, so `globs:`/`commands:` with no `SKILL_REFERENCE_INJECTIONS` entry is a
build error, and a pack carrying the text does not excuse it. **A rule the core includes may
not also declare a trigger** — the hook would push what the core already carries; the build
refuses that too.

Today: 42 rule texts — 7 always-on, 26 triggered, 8 router-only, 1 pack-only
(`agent-core-gates.md`). Recount from `harness/rule-manifest.json`, never by subtraction.

## Adding a rule — decide the shape first

| The rule governs… | Shape | Wire |
|---|---|---|
| how the agent thinks or answers in EVERY session, with no observable act to hang it on (a safety gate, a reporting duty, a decision discipline) | **core include** | the text here + `global/core-sections/rule-<name>.md` carrying only `order:`, `targets:`, `include:`. No trigger, no body in the section |
| an observable act — writing a file of a kind, running a command | **triggered** | `globs:` and/or `commands:` + a `SKILL_REFERENCE_INJECTIONS` entry + a row in that router's routing table + `RULE_READER_NAMES` if read-only agents need it + `packs:` on the specialized agents that should carry it + `exclusive-with:` where another rule claims the same file conventions |
| an intent with no act (delegating, planning, saving memory) | **router-only** | a `SKILL_REFERENCE_INJECTIONS` entry and its routing-table row. When absence is unsafe, a one-line gate in the core naming the router and its trigger |
| only what a specialized executor does, and the main thread never needs it | **pack-only** | the `packs:` line on each agent. No trigger, no injection |

A trigger is written as an ACT a third party could point to in the transcript — the write, the
command — never as an intent. If you cannot name the act, the shape is core include or
router-only, not a sentinel glob.

## Frontmatter

Only these keys, and only on a rule that needs them:

```yaml
---
globs:                       # hook trigger: paths this rule scopes, matched by segment
  - "**/*.{ts,tsx}"          #   against the target's absolute realpath
commands:                    # hook trigger: command PREFIXES, matched on a parsed stage
  - "git commit"             #   token for token after argv[0]'s basename
  - "npm install +"          #   trailing ` +` = needs one more non-option argument
exclusive-with:              # rule stems whose pack excludes this one (framework overlap)
  - "nestjs-patterns"
---
```

`paths:` is a build error — it was Claude Code's native key and that mechanism is retired.
`readers` and `always_on` are **derived**, never authored: `readers` from `RULE_READER_NAMES`
in `harness/build.py` (a rule a reviewer needs, not only a writer), `always_on` from the core
section that includes the file.

## Before you finish

- **Run the build.** `python3 harness/build.py` regenerates the manifest, the cores and every
  generated tree; `--check` runs the parity checks without writing. It refuses: a rule no
  channel delivers, a trigger with no reference, an `include:` over a triggered rule, two
  sections including the same text, an `exclusive-with:` naming no rule, a malformed command
  prefix, `paths:` anywhere.
- **Run the hook's tests.** `python3 global/hooks/rule-delivery/test_rule_delivery.py` — its
  matrix runs against the real manifest, so a new trigger is exercised there, not only parsed.
- **Cost the rule before keeping it.** A core include is paid by every session in every
  project — state its KB. A trigger is paid in denial ROUNDS on a first write or command:
  check how many rules the new one joins on that trigger, and that the reason still names it
  whole inside Grok's ~264-character budget (a path that cannot be named whole is simply not
  gated there).
- **Exercise it in vivo, in a sandbox outside the temp roots.** `$TMPDIR`, `/tmp`,
  `/private/tmp` and `/var/folders` are ungated by design, so a probe run there proves nothing
  about a hold.
- **Update the inventories** the change touches: `README.md`'s rules section, `global/README.md`
  and the per-harness loading tables, `AGENTS.md`'s File Structure.

## The cost of living here

Outside the core-include shape, a rule is present only when its channel fires — a router the
model must invoke, or a hook the model must obey after a hold. Never leave one whose absence is
unsafe to those channels; that is what the core include is for. The compensating control for
the rest is a one-line pointer in `global/CLAUDE.md` naming the router and its trigger.

Packed agents run with `omitClaudeMd: true` on Claude Code, so `agent-core-gates.md` is
required in every `packs:` line — it is what puts the core gates back. Codex, opencode, Grok
and PI have no per-agent corpus switch, so a packed agent there receives those gates on top of
the core it already loads.
