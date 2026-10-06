# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
the project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Each release sets [VERSION](VERSION) to its version; between releases it reads `dev`.
[CONTRIBUTING.md](CONTRIBUTING.md#releases-and-version-numbers) says which changes
raise the minor or the patch version.

## [Unreleased]

## [0.2.1] - 2026-10-06

### Added

- Documentation site in `site/`, with one page per flow (research, plan, build,
  close), published at https://hive.tricell.tech/ from this release on.

### Changed

- The `starlight-docs-site` template pins Astro 7.3.5, Starlight 0.42.4, and
  pnpm 12.8.1, denies the `esbuild` install script it does not need, and ships a
  `biome.json` so `pnpm lint` ignores build output.
- `README.md`, `llms.txt`, and pages of the documentation site are updated in
  the same pull request as any change to what they state.
- `CONTRIBUTING.md` defines what the version number tracks and when a release
  raises the minor or the patch version.

## [0.2.0] - 2026-10-05

### Added

- One-line installation: `curl -fsSL https://hive.tricell.tech/install.sh | sh`
  runs `get-hive.sh`, which downloads the latest release package, verifies its
  checksum, keeps it in `~/.local/share/hive/packages/`, and runs its
  `install.sh`. `--dry-run` previews the installation without changing any host
  (the package is still downloaded and kept; add `--hosts` when no terminal is
  available), and `--version` selects a release (#11).

### Changed

- `git-workflow` asks pull-request descriptions to state whether a revert fully
  undoes the merge and what the change would affect if it were wrong.

## [0.1.0] - 2026-10-05

First public release.

### Added

- Shared guidance layer for deliberate work with coding agents: authorization,
  preservation, evidence, and artifact-placement rules.
- Activity skills: `flow-research`, `flow-plan`, `flow-build`, and `flow-close`.
- Specialist role definitions with bounded tasks, owned paths, and evidence to
  return.
- Go manager with the commands `install`, `status`, `doctor`, `update`, `plan`,
  `apply`, and `recover`. It previews destinations, checks ownership and drift,
  preserves user content outside managed resources, and supports recovery.
- Six host adapters: Claude Code, Codex, Grok Build, Pi, OpenCode V2, and
  Cursor CLI.
- Downloadable packages on GitHub Releases for macOS Apple Silicon (arm64) and
  Linux arm64 and amd64. Each package includes the executable, content,
  verification inventory, `LICENSE`, and `THIRD_PARTY_NOTICES.md`, and installs
  offline with `./install.sh`. Installation refuses a package that lacks the
  license files, and a content test keeps the notices in sync with `go.mod` and
  the Go toolchain.

### Removed

- The `bootstrap.sh` online installer and the `hive bootstrap` subcommand.
  Install from a complete package or a source checkout instead. State written by
  the retired subcommand (an `installer` field) is still read and ignored.
