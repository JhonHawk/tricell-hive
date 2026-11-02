---
title: Host limits
description: What Hive cannot hide in each supported coding CLI, and what to do about it.
---

Hive supplies guidance; each coding CLI keeps its own models, tools, permissions
and history. That means some behavior belongs to the CLI and Hive cannot change
it. This page lists the limits that the maintainers have recorded, and what you
do in each case. A claim marked **observed** was seen in a real session or in the
CLI's source; the repository's
[agent delivery](https://github.com/JhonHawk/tricell-hive/blob/master/_support/docs/architecture/agent-delivery.md)
and
[deployment manager](https://github.com/JhonHawk/tricell-hive/blob/master/_support/docs/architecture/deployment-manager.md)
notes hold the evidence and versions. Anything not listed here has not been
verified.

## Limits that apply to every CLI

- **A long session can lose the end of a skill.** A host may keep only the start
  of a skill after it compacts the conversation. That is why each skill's
  essential procedure sits at its beginning. If the agent seems to forget a
  procedure afterwards, name the flow again.
- **Open sessions keep the guidance they started with.** After `hive update`,
  open sessions load the new guidance only after a restart. Two recorded
  exceptions: OpenCode applies edits to its global and upward-discovered
  `AGENTS.md` files before its next model request, and Codex 0.166 reloaded the
  global `AGENTS.md` in an open session within seconds of an update (one session,
  observed). OpenCode still needs a new session for changed roles or skills.
- **Hive can only count sessions on some CLIs.** After an update it reports how
  many open Claude Code and Grok sessions to restart. It cannot see Codex, Pi or
  Cursor sessions, so restart those yourself.

## Claude Code

- Its question card waits for your answer with no time limit (unless you set
  `askUserQuestionTimeout` yourself), so a pending card is the agent waiting for
  you, not a hang.
- While a card is open it covers the screen (an open Claude Code issue), so each
  card carries the context needed to answer on its own.

## Codex

- **Question cards close by themselves.** In Default mode a card that you do not
  touch resolves with no answer after about 120 seconds: 60 seconds hidden, then a
  60-second visible countdown. Pressing a key in the card stops that. When it
  closes, Hive's rule is that the decisions are already in the message above it, so
  answer there (for example `D2 B`).
- Question cards in Default mode need a Codex feature that is off by default,
  `default_mode_request_user_input`. Without it, the agent writes the options in
  the message instead of a card. `hive doctor` shows the state of this and of the
  wait setting below.
- A helper role that declares a read-only sandbox can still run with the session's
  sandbox (observed on Codex 0.163.0). Hive therefore relies on the role's text, not
  the sandbox, to keep such helpers read-only.
- Codex lets an agent give up waiting for helper roles quickly. An optional Hive
  capability writes longer minimum waits (5 and 10 minutes) into your Codex
  `config.toml`; a wait still ends early when a helper finishes or you write.
  `hive doctor` reports whether those keys are set.

## Grok Build

- Its question card closes after 1800 seconds by default, and a close by timeout
  counts as a decline. The `toolset.ask_user_question.timeout_enabled` setting
  controls that.
- Helper roles keep the model written in the installed role file even when the
  session runs another one (observed). The agent can still mistake this for
  inheritance, so if it asks you about which model a verifier used, point it to the
  role file.

## Pi

- Pi has no built-in question tool. Cards exist only if you install the
  third-party `rpiv-ask-user-question` extension; without it the agent writes the
  full options in the message and you answer in text.
- Helper roles need the `pi-subagents` package. Hive declares it with
  `pi install npm:pi-subagents` when you install for Pi, and leaves an existing
  pin alone.
- Pi roles run on your session's model.

## OpenCode

- Dismissing a question card interrupts the whole turn. Answer it, or reply in
  text if you would rather not stop the work.
- Text sent together with a card can be lost, so each card stands on its own
  with its context.
- Roles run on your session's model and variant; per-role effort settings have no
  effect there. A role's choice of a different model needs a subscription you have.
- The optional context loader is off until you add it to your own
  `opencode.json`; see the [loader notes](/#optional-opencode-v2-context-loader).

## Cursor CLI

- **The global rule loads only for projects under your home folder.** Hive writes
  its guidance to `~/.cursor/rules/hive.mdc`, which reached a workspace under the
  home folder and not one outside it (two probes on Cursor Agent 2026.10.22). For a
  repository outside the home folder, the harness audit offers a pointer inside
  the repository.
- Cursor's CLI loads no user-level helper roles (observed on 2026.10.22). The
  agent then gives a generic helper the role's contract, and read-only holds only
  because the brief says so, not because the host enforces it.
- Its question card has no description field and no "Recommended" mark, so
  the effect of each option is written in its label. Whether the agent keeps
  working while a card is open is not settled: the binary waits and the
  documentation says otherwise.
- Helper roles run on your session's model.

## What to do when a limit bites

1. Restart the CLI after `hive update`. OpenCode needs no restart when the change
   was only to the global instructions; Codex 0.166 also reloaded them once,
   but restart it when in doubt.
2. Run `hive doctor` to see what Hive can check on your machine.
3. If a card closes before you answer, reply in the chat with the decision ID
   and letter. The decisions are also written in the message.
