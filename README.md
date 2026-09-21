# Hive

Hive is in a rebuild phase for a small, portable guidance layer targeting Claude Code, Codex, Grok Build, Pi, and OpenCode V2. Each CLI keeps its native model loop, authentication, tools, permissions, and history. The five-host target is a design scope, not proof that a shared skill set or runtime is installed or working.

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

The [Go deployment manager](_support/docs/architecture/deployment-manager.md) provides explicit plan/apply/status/recovery operations for Codex, Claude, Grok, Pi, and OpenCode. Run `go test -race ./...` and `go vet ./...` to verify it. No model runtime or hook framework is provided. Preserve user-owned configuration, credentials, histories, third-party tools, backups, and the Engram workspace identity.

## Repository structure

Shared distributable content lives in `content/`; host differences in `integrations/`; management tooling in `tooling/`; and verification in `tests/`. Empty directories are reserved with `.gitkeep`, not implemented features. Generated `dist/` and local workspace contents remain ignored.

See the [repository and distribution design](_support/docs/architecture/repository-and-distribution.md). The first [shared global rules](content/guidance/global.md) are authored as a body to insert between managed markers in existing global instruction files, preserving user content. The [workspace-conventions skill](content/skills/workspace-conventions/SKILL.md) handles support maintenance; basic placement remains global. [Four measurement cases](tests/fixtures/workspace-conventions/README.md) are prepared. The manager is implemented; the five-host user-global deployment is installed for an experimental observational rollout. Behavioral results and limitations are recorded separately from filesystem installation tests.
