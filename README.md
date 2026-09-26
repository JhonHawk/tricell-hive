# Hive

## Install from a complete package

Download the complete package for macOS Apple Silicon or Linux ARM64/AMD64,
extract it, and run `./install.sh`
in a terminal. The installer detects supported CLIs and recognized legacy Hive
resources, previews affected destinations, and asks for confirmation.
No Go installation or terminal authentication is required. Use `--dry-run` for a
read-only preview. Packages are built locally until a release is explicitly published;
GitHub's automatic source archives do not include the executable.
See the [installer contract](_support/docs/architecture/installer.md) for migration,
recovery, supported legacy revision, and maintainer packaging instructions.

Hive is in a rebuild phase for a small, portable guidance layer targeting Claude Code, Codex, Grok Build, Pi, OpenCode V2, and Cursor CLI. Each CLI keeps its native model loop, authentication, tools, permissions, and history. The six-host target is a design scope, not proof that a shared skill set or runtime is installed or working.

Start with [AGENTS.md](AGENTS.md), the common project guidance. Claude Code imports it through [CLAUDE.md](CLAUDE.md). The guidance is designed to work without first selecting an activity skill, especially for authorization, evidence, preservation, and artifact placement.

## Research and measurement

The [harness engineering research index](_support/docs/harness-engineering/README.md) links to the source-backed host comparison and the analysis of the supplied `uber-software-factory.zip` corpus. This branch retains that research; earlier implementation, audit, pilot, and retirement records remain in Git history. The research does not demonstrate a quality improvement. Recheck mutable host documentation against the installed CLI version before adapting its behavior.

## Where work goes

- Source changes stay in the existing source directories.
- Durable repository guidance belongs in `_support/docs/<topic>/`.
- Resumable work records belong in `_support/sessions/YYYY-MM-DD-<slug>/`, reusing the same work folder across conversations.
- Temporary output belongs in git-ignored `_support/workspace/YYYY-MM-DD-<slug>/`.
- Selected evidence belongs in `_support/evidence/YYYY-MM-DD-<slug>/`; only curated material suitable for sharing is eligible for Git.

The [workspace and artifact policy](_support/docs/architecture/workspace-and-artifacts.md) explains repo versus shared workspace scope, established documentation homes, task-local cleanup, and requested archiving.

The [Go deployment manager](_support/docs/architecture/deployment-manager.md) provides explicit plan/apply/status/recovery operations for Codex, Claude, Grok, Pi, OpenCode, and Cursor. Run `go test -race ./...` and `go vet ./...` to verify it. No model runtime or hook framework is provided. Preserve user-owned configuration, credentials, histories, third-party tools, backups, and the Engram workspace identity.

## Repository structure

Shared distributable content lives in `content/`; host differences in `integrations/`; management tooling in `tooling/`; and verification in `tests/`. Only retained local-workspace directories need `.gitkeep` markers. Generated `dist/` and local workspace contents remain ignored.

See the [repository and distribution design](_support/docs/architecture/repository-and-distribution.md). The [shared global rules](content/guidance/global.md) are authored as a body to insert between managed markers in existing global instruction files, preserving user content. The [workspace-conventions skill](content/skills/workspace-conventions/SKILL.md) handles support maintenance; basic placement remains global. [Measurement cases](tests/fixtures/workspace-conventions/README.md) and historical five-host rollout records are retained. Verify current installation separately from behavioral results and filesystem installation tests.

The portable work procedures are [flow-research](content/skills/flow-research/SKILL.md), [flow-plan](content/skills/flow-plan/SKILL.md), and [flow-build](content/skills/flow-build/SKILL.md). Planning includes a conditional [format and template reference](content/skills/flow-plan/references/plan-format.md). They can be used independently. The [TypeScript flow pilot](tests/fixtures/flows/README.md) checks loading and cross-CLI plan handoffs; the current report (historical evidence omitted from public history) distinguishes observations from unverified behavior.

Start with `go run ./tooling/cli setup` for the optional Context7 recommendation and local skill discovery. This check never installs tools or requests credentials; the [manager contract](_support/docs/architecture/deployment-manager.md#optional-context7-setup-recommendation) explains voluntary installation and refresh.

Canonical specialist roles live in `content/agents/`; the [agent delivery contract](_support/docs/architecture/agent-delivery.md) describes profiles, inline native conversion, and verification limits. Additional activities include `flow-report`, `harness-audit`, `engram-init-workspace`, `starlight-docs-site`, `unattended-delegation`, and `workspace-archive`; their resources remain beside each skill.
