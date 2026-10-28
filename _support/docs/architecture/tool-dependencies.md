# Optional tool dependencies

This inventory describes the optional third-party capabilities the installer
offers. It is not an installation guide, a claim that every host/platform pair
works, or an authorization to change a user environment. The typed catalog is
`tooling/providers`.

## Boundaries

The core Hive package works without any of these capabilities. Two have a
provider recipe (below): `agent-browser`, which installs a program, and the Codex
question card, which runs a Codex command. A third, the Codex wait floor, has no
command: Hive writes two keys of Codex's `config.toml` itself (plus the table header when it creates the table). For the others the installer never
detects, downloads, installs, configures or authenticates a provider. Selecting
one of them records a `manual` step in the onboarding journal and prints its
official source and the reason it is manual; the core still installs, and the
result is reported as partial with a non-zero exit so the capability is never
mistaken for an installed one. With the operator's consent, given in the
install summary, Hive runs the host's own command that writes a host
configuration file: `pi install npm:pi-subagents` for Pi's `settings.json` and
`codex features enable default_mode_request_user_input` for Codex's
`config.toml`. The Codex wait floor is the one exception: with the same consent
Hive writes two keys into Codex's `config.toml` itself (below). Existing binaries, packages, histories, credentials,
memory stores and services are untouched.

| Capability | Purpose | Relevant hosts | Official source | Why manual |
| --- | --- | --- | --- | --- |
| Engram | Persistent memory and optional host integration | Claude, Codex, Cursor, OpenCode, Pi; Grok reads Claude's MCP configuration | Engram installation docs (versioned, checksummed release artifacts; `engram setup <agent>` is separate from the binary) | No platform asset, checksum, target location and host integration has passed a native gate |
| Context7 | Current library documentation through its CLI and vendor skills | Codex, Claude, Cursor, OpenCode | Context7 CLI docs (`ctx7 setup --cli`, npm distribution; needs Node.js 18+) | Setup is interactive and may authenticate; Hive installs no runtime |
| pi-subagents | Pi delegated children | Pi | pi-subagents docs (`pi install npm:pi-subagents`) | Upstream validation covers only Pi 0.86.1 on Linux x64; undoing only an added entry (#29) needs Pi's native gate |
| `agent-browser` | Browser automation for UI review and in-vivo checks | All six | agent-browser README (`brew install agent-browser`, `npm install -g agent-browser`, `agent-browser install`) | Installable after consent on macOS arm64 (Homebrew) and Linux amd64 (Node 24+, writable npm prefix); manual elsewhere |
| CodeGraph | A queryable code graph (symbols, callers, impact) through an MCP server and CLI; the `hive-core` evidence-tools reference says when to use it, when to run `codegraph sync -q`, and what to confirm | Claude, Codex, Cursor, OpenCode | CodeGraph README (`github.com/colbymchenry/codegraph`; `codegraph install` writes each host's MCP entry and its own instruction block) | Its installer writes several hosts' configuration, which Hive does not run; manual-only, with no recipe |
| Codex question card (`codex-question-card`) | Codex's synchronous question card in Default mode, so decisions are asked through a card that waits instead of text | Codex | Codex CLI (`codex features enable default_mode_request_user_input`, `codex features disable …`, `codex features list`); feature at stage "under development" in `rust-v0.163.0`, not in the official configuration reference | Executable after consent when `codex` is on `PATH` and `codex features list` reports the feature, selected by default in interactive installs; manual when `codex` is missing, too old, or the home is synthetic |
| Codex wait floor (`codex-wait-floor`) | Keeps a parent agent from waking every 60 seconds while a delegated child works, by raising Codex's minimum and default wait under `[features.multi_agent_v2]` | Codex 0.164.0 or later | Codex source, tag `rust-v0.164.0` (`github.com/openai/codex`); not in the official configuration reference | Not manual: Hive writes the lines after consent, selected by default in interactive `install` and `hive update` while unapplied; not offered with an older Codex, without `codex`, or under a synthetic home |

The Codex question card recipe runs one local command, Codex's own
configuration writer, which adds `default_mode_request_user_input = true` to
the `[features]` table of `config.toml` under the Codex home (`CODEX_HOME`, by
default `~/.codex`) and is idempotent; Hive reads that file to record the step
as verified and never writes it. With the flag, Codex registers the card in
Default mode and the model waits for the answer; the TUI auto-resolves an
untouched card with empty answers after 60 s hidden plus 60 s of visible
countdown, and any key pressed in the card cancels the countdown (observed in
the `rust-v0.163.0` sources). Hive's question rule already asks a timed-out
question again in text. Codex warns that under-development features may change;
`codex features disable default_mode_request_user_input` turns it off.
Installations made before this capability enable it by hand with the same
command, which `hive doctor` names.

The Codex wait floor writes `min_wait_timeout_ms = 300000` and
`default_wait_timeout_ms = 600000` under `[features.multi_agent_v2]` in the same
`config.toml`, inserting only those lines and only the keys that are missing; a
conflicting value or a file Hive cannot edit safely leaves it untouched and
becomes a manual step. The minimum applies to Codex's v2 multi-agent runtime;
v1 is not verified. When `config.toml` exists, Hive saves `config.toml.hive-backup`
(mode 0600) next to it before the first write (a missing file is created with mode 0600), and its journal keeps only fingerprints and the
lines it wrote. `hive plan remove` for the codex host removes only those lines
while they keep their values, then the backup after a clean revert. An older
Hive binary that rewrites Hive's state loses the record, and then `plan remove`
leaves the lines. See the [installer contract](installer.md) for the steps.

`agent-browser` is the program-installing recipe. It installs the latest version, with no
version prompt: Homebrew resolves its formula, which depends on `node` and about
25 other formulas, and npm resolves the latest package, whose postinstall
downloads the native binary from GitHub without a checksum. `agent-browser
install` downloads Chrome for Testing into `$HOME/.agent-browser/browsers`, and
the installer runs it only when no browser is found in the locations
agent-browser checks (on macOS the Google Chrome, Chrome Canary and Chromium
applications; on Linux `google-chrome`, `google-chrome-stable`,
`chromium-browser` and `chromium` on `PATH`), `AGENT_BROWSER_EXECUTABLE_PATH`
names a file, or that folder already holds a browser. Hive never uninstalls
`agent-browser` or Chrome, and never runs `agent-browser install --with-deps`.

## Evidence and limits

The catalog reflects provider documentation checked on 2026-09-25, and
CodeGraph 1.6.2 checked on 2026-10-23. The Codex wait floor source was verified on 2026-10-25 against `rust-v0.164.0`. It is
documentary evidence, not a native provider-install test.

## Adding a recipe later

A recipe returns only after its host/platform combination passes a native gate,
following the rules kept in the change design (structured process arguments,
no replacement of an existing install, provider output never journaled). An
exact version resolved before consent is the default; `agent-browser` installs
the latest version instead, by the user's decision, and says so before consent. The onboarding parent journal already
separates the core commit from optional steps and blocks retries after an
unknown outcome, so a recipe adds a new step kind and status in
`tooling/providers` plus its runner in `tooling/cli`. The code of the earlier
unshipped recipes is archived outside Git under
`_support/workspace/2026-09-26-versioned-installer-onboarding/`.
