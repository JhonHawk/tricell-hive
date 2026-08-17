# rule-context

`PreToolUse` hook. **Advisory only — never denies, always exits 0.**

Names the situational rule that applies at the moment it starts applying: a file kind about
to be written, or a tool with its own routing policy about to run.

## Why it exists

Situational rules reach each harness through a different channel, and two of the four have
no channel at all for the main thread:

| Harness | Path-scoped rules in the main thread |
|---|---|
| Claude Code | loaded natively via `paths:` |
| opencode | loaded by glob by the `opencode-rules` plugin |
| Codex | **nothing loads them** |
| Grok | ignores `paths:`; path-scoped rules are not symlinked into `~/.grok/rules/` |

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
  `browser-automation.md`; `codegraph` / `jbcontext` → `code-search.md`.

Codex runs edits through `apply_patch` inside a shell call, so the file kind is read from the
patch header instead of a `file_path` field. Without that branch the hook would be inert on
the harness that most needs it.

## Limits

Deterministic at **delivering** the reminder, not at obeying it. For a short rule whose breach
is expensive, inject the content instead of the pointer — it costs tokens, but only when a
file of that kind is actually touched.
