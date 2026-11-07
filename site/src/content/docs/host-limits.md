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
  many open Claude Code and Grok sessions to restart. It cannot see Codex, Pi,
  Cursor or Antigravity sessions, so restart those yourself.
- **Host hooks load when a session starts.** The hooks that report a session to
  the [viewer](/viewer/#host-observed-activity) are read at session start, so
  restart open sessions after installing them. Codex also skips them until you
  approve them in `/hooks`, and Cursor CLI's use of its hooks file is verified
  only in the installed code. Antigravity's turn-end hook reply is not yet
  verified in a live session.

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
  pin alone. It also sets `"toolActivation": "eager"` in
  `extensions/subagent/config.json` when you have not set it, so helper roles
  can start in the same turn instead of after your next message; restart Pi
  after the install.
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

## Antigravity

- **Covered:** the `agy` CLI (1.5.2 observed) and the Antigravity App (2.29
  observed), which share `~/.gemini/config`. Hive installs at user scope only;
  the IDE extensions are not covered.
- **The guidance arrives in several rule files.** `agy` cuts every rule file at
  24,000 bytes, global ones included (observed 2026-11-06 on `agy` 1.5.2), and
  always-on and global rules share a 20,000-token budget (documented). Hive
  writes `~/.gemini/config/rules/hive-NN.md` parts that stay under that cut;
  how many there are depends on the guidance size and the length of your home
  path. Do not put your own files at those names.
- Helper roles run on your session's model. `agy` has no per-role effort and no
  read-only setting, so read-only holds only because the role's instructions say
  so. `hive models set` offers `inherit`, `flash_lite`, `flash` and `pro`.
- Hive cannot read `agy`'s native history, which is not a plain-text format, so
  session review and `hive doctor` cannot inspect Antigravity sessions. `agy`
  keeps its own memory ("Knowledge"), which Hive does not manage.
- Hive does not write `~/.gemini/GEMINI.md`, which Gemini CLI also uses.
- Not yet verified for Antigravity: how its context behaves after compaction,
  its question card, and how it launches helper roles.

## What to do when a limit bites

1. Restart the CLI after `hive update`. OpenCode needs no restart when the change
   was only to the global instructions; Codex 0.166 also reloaded them once,
   but restart it when in doubt.
2. Run `hive doctor` to see what Hive can check on your machine.
3. If a card closes before you answer, reply in the chat with the decision ID
   and letter. The decisions are also written in the message, and the next
   decision still comes in a card.
