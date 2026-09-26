# Optional tool dependencies

This inventory describes the optional third-party capabilities the installer
offers. It is not an installation guide, a claim that every host/platform pair
works, or an authorization to change a user environment. The typed catalog is
`tooling/providers`.

## Boundaries

The core Hive package works without all three capabilities. This release has no
provider recipes: the installer never detects, downloads, installs, configures or
authenticates a provider. Selecting a capability records a `manual` step in the
onboarding journal and prints its official source and the reason it is manual;
the core still installs, and the result is reported as partial with a non-zero
exit so the capability is never mistaken for an installed one. Existing binaries,
packages, host configuration, histories, credentials, memory stores and services
are untouched.

| Capability | Purpose | Relevant hosts | Official source | Why manual |
| --- | --- | --- | --- | --- |
| Engram | Persistent memory and optional host integration | Codex, Claude, OpenCode | Engram installation docs (versioned, checksummed release artifacts; `engram setup <agent>` is separate from the binary) | No platform asset, checksum, target location and host integration has passed a native gate |
| Context7 | Current library documentation through its CLI and vendor skills | Codex, Claude, Cursor, OpenCode | Context7 CLI docs (`ctx7 setup --cli`, npm distribution; needs Node.js 18+) | Setup is interactive and may authenticate; Hive installs no runtime |
| pi-subagents | Pi delegated children | Pi | pi-subagents docs (`pi install npm:pi-subagents@<version>`) | Upstream validation covers only Pi 0.86.1 on Linux x64; undoing only an added entry (#29) needs Pi's native gate |

## Evidence and limits

The catalog reflects provider documentation checked on 2026-09-25. It is
documentary evidence, not a native provider-install test.

## Adding a recipe later

A recipe returns only after its host/platform combination passes a native gate,
following the rules kept in the change design (structured process arguments,
exact versions resolved before consent, no replacement of an existing install,
provider output never journaled). The onboarding parent journal already
separates the core commit from optional steps and blocks retries after an
unknown outcome, so a recipe adds a new step kind and status in
`tooling/providers` plus its runner in `tooling/cli`. The code of the earlier
unshipped recipes is archived outside Git under
`_support/workspace/2026-09-26-versioned-installer-onboarding/`.
