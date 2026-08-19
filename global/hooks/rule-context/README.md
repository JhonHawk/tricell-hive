# rule-context

`PreToolUse` hook. **Advisory only — never denies, always exits 0.**

Names the situational rule that applies at the moment it starts applying: a file kind about
to be written, or a tool with its own routing policy about to run.

## Where it actually fires

**Claude Code and Grok only.** In Codex it does not run — verified in both TUI and
`codex exec`, zero injections in either rollout, even though Codex's own hook panel lists
`PreToolUse — 1 installed, 1 active`. It is registered and recognized; it just never reaches
the model.

That costs nothing, because Codex is covered better by another channel: the
`MANDATORY FIRST ACTION` line in `harness/AGENTS.md` plus the router bootstrap in
`global/CLAUDE.md`. Observed in the same test — Codex announced *"Voy a aplicar language-rules
porque el archivo es TSX"*, read `typescript-standards.md`, and only then wrote the file.
An always-on instruction acts **before** the decision to write; this hook cannot.

Separately: **`codex exec` runs no hooks at all** — not SessionStart, not any. Its rollout has
zero while a TUI session carries the hive's SessionStart. Relevant to any automation built on
`codex exec` (CI, scripts): hooks do not guard it.

## Why it exists

Situational rules reach each harness through a different channel, and two of the four have
no channel at all for the main thread:

| Harness | Channel that reaches the main thread | Arrives before the write? |
|---|---|---|
| Claude Code | `paths:` natively, plus this hook | no — see below |
| opencode | `opencode-rules` plugin, by glob | yes |
| Codex | `MANDATORY FIRST ACTION` in `harness/AGENTS.md` | **yes** — an instruction acts before the decision |
| Grok | this hook, plus the router skills | no for the hook; the routers depend on invocation |

Claude Code's own `paths:` mechanism has the same gap, and for a different reason: per the docs,
*"path-scoped rules trigger when Claude reads files matching the pattern, not on every tool
use."* The trigger is a READ. Creating a new file of that kind never reads it, so the rule does
not load for the write that would have needed it most.

Two more consequences worth knowing before moving rules off always-on:

- Path-scoped rules **do not survive compaction**: *"rules with `paths:` frontmatter are not
  re-injected automatically; they reload the next time Claude reads a file matching the rule's
  patterns."* A long session loses them silently.
- The `InstructionsLoaded` hook logs which instruction files loaded, when, and why — the docs
  name it for *"debugging path-specific rules or lazy-loaded files"*. That is the measuring
  instrument for verifying any always-on reduction.

Subagents get their policy inlined into their own prompt. The main thread had nothing — on
Codex and Grok it has been writing React or Terraform with none of those rules present. This
hook closes that, and on the other two it still helps: the reminder fires with the concrete
file in hand rather than as a general norm at session start.

## Why it is not part of `bash-policy`

`bash-policy` **denies** (exit 2 in Claude Code, decision JSON in Grok) — it is a gate. This
hook only advises. Merging them would tie a security gate's reliability to extension mapping
it does not need: a bug in the mapping must never be able to take down the `rm -rf` or
`path=` denial. Both hooks match `Bash`; overlapping matchers with different jobs is exactly
what the hook system allows.

## What it emits

The rule **name**, never a path — the same file lives under a different root in each harness
(`~/.claude/rules`, `~/.agents/skills/language-rules/references`, the opencode plugin), and a
wrong absolute path sends the model hunting. That was the failure mode behind the 21 missed
reference reads on Grok.

One reminder per rule per session: a marker per `(session, rule)` under `$TMPDIR` keeps an
edit-heavy session from paying the same pointer on every write.

## Coverage

- **File kind** (`Write` / `Edit` / `search_replace`, and `apply_patch` headers inside a Codex
  `shell` call): TS/JS, React, Python, Java/Kotlin, SQL, shell, Terraform, Dockerfile, CI
  workflows, CSS, and the Angular/NestJS shared suffixes.
- **Tool policy** (`Bash` / `run_terminal_command` / `shell`): `agent-browser` →
  `browser-automation.md`.

Codex runs edits through `apply_patch` inside a shell call, so the file kind is read from the
patch header instead of a `file_path` field. Without that branch the hook would be inert on
the harness that most needs it.

## Limits

**The reminder lands after the write, not before it — by design, not by accident.** Claude
Code's hook reference defines the field as *"added to the Claude Code context when the tool
completes"*, and states the consequence outright: the context *"cannot influence the permission
decision — it arrives after the tool has already executed."* Observed twice before the docs were
checked, in a running session and a fresh one.

Of `PreToolUse`'s three `permissionDecision` values (`allow`, `deny`, `escalate`), only `allow`
and `deny` are pre-emptive. An advisory hook has no pre-emptive path available: the first write
of a given kind happens without the rule, and the reminder applies from the next action onward.

That is a real ceiling, not a wording detail: this is not a gate, and pairing it with the
one-reminder-per-session dedup means a rule is named exactly once, right after its first use.
The trade was deliberate — insisting on every write would turn a useful pointer into noise
that gets tuned out.

Deterministic at **delivering** the reminder, never at obeying it. For a short rule whose
breach is expensive, inject the content instead of the pointer — it costs tokens, but only when
a file of that kind is actually touched.
