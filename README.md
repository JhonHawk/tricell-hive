# Hive

Hive is in a rebuild phase for a small, portable guidance layer targeting Claude Code, Codex, Grok Build, Pi, and OpenCode V2. Each CLI keeps its native model loop, authentication, tools, permissions, and history. The five-host target is a design scope, not proof that a shared skill set or runtime is installed or working.

Start with [AGENTS.md](AGENTS.md), the common project guidance. Claude Code imports it through [CLAUDE.md](CLAUDE.md). The guidance is designed to work without first selecting an activity skill, especially for authorization, evidence, preservation, and artifact placement.

## Research and measurement

The [harness engineering research index](_support/docs/harness-engineering/README.md) links to the source-backed host comparison and the analysis of the supplied `uber-software-factory.zip` corpus. This branch retains that research; earlier implementation, audit, pilot, and retirement records remain in Git history. The research does not demonstrate a quality improvement. Recheck mutable host documentation against the installed CLI version before adapting its behavior.

## Where work goes

- Source changes stay in the existing source directories.
- Durable repository guidance belongs in `_support/docs/`.
- Resumable work records belong in `_support/sessions/YYYY-MM-DD-<slug>/`.
- Reproducible temporary output belongs in git-ignored `_support/workspace/`.
- Curated non-reproducible evidence belongs in `_support/evidence/YYYY-MM-DD-<slug>/`.

This repository does not promise a runtime, deploy mechanism, hook framework, or build/test command. Inspect what exists before relying on it. Preserve user-owned configuration, credentials, histories, third-party tools, backups, and the Engram workspace identity.
