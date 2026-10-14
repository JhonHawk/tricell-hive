# Optional tool dependencies

This inventory describes the optional third-party capabilities the installer
offers. It is not an installation guide, a claim that every host/platform pair
works, or an authorization to change a user environment. The typed catalog is
`tooling/providers`.

## Boundaries

The core Hive package works without all four capabilities. Only `agent-browser`
has a provider recipe (below); for the others the installer never detects,
downloads, installs, configures or authenticates a provider. Selecting one of
them records a `manual` step in the onboarding journal and prints its official
source and the reason it is manual;
the core still installs, and the result is reported as partial with a non-zero
exit so the capability is never mistaken for an installed one. Existing binaries,
packages, host configuration, histories, credentials, memory stores and services
are untouched.

| Capability | Purpose | Relevant hosts | Official source | Why manual |
| --- | --- | --- | --- | --- |
| Engram | Persistent memory and optional host integration | Codex, Claude, OpenCode | Engram installation docs (versioned, checksummed release artifacts; `engram setup <agent>` is separate from the binary) | No platform asset, checksum, target location and host integration has passed a native gate |
| Context7 | Current library documentation through its CLI and vendor skills | Codex, Claude, Cursor, OpenCode | Context7 CLI docs (`ctx7 setup --cli`, npm distribution; needs Node.js 18+) | Setup is interactive and may authenticate; Hive installs no runtime |
| pi-subagents | Pi delegated children | Pi | pi-subagents docs (`pi install npm:pi-subagents`) | Upstream validation covers only Pi 0.86.1 on Linux x64; undoing only an added entry (#29) needs Pi's native gate |
| `agent-browser` | Browser automation for UI review and in-vivo checks | All six | agent-browser README (`brew install agent-browser`, `npm install -g agent-browser`, `agent-browser install`) | Installable after consent on macOS arm64 (Homebrew) and Linux amd64 (Node 24+, writable npm prefix); manual elsewhere |

`agent-browser` is the one recipe. It installs the latest version, with no
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

The catalog reflects provider documentation checked on 2026-09-25. It is
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
