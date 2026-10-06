# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
the project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
[VERSION](VERSION) holds the current product version.

## [Unreleased]

### Added

- A content test that keeps `THIRD_PARTY_NOTICES.md` in sync with `go.mod` and
  the local Go toolchain.

### Changed

- Installation now refuses a package that lacks `LICENSE` or
  `THIRD_PARTY_NOTICES.md`.

### Removed

- The `hive bootstrap` subcommand. Plans and onboarding journals written by it
  (they carry an `installer` field) are still read and the field is ignored, so
  `hive recover` keeps working on that state.
- The package outputs `versions/<v>/index.json` and the raw per-platform `hive`
  and `hive.sha256` files next to the archives.

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
  offline with `./install.sh`.

### Removed

- The `bootstrap.sh` online installer. Install from a complete package or a
  source checkout instead.
