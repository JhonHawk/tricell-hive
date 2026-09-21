# Shared global instructions

`global.md` will be the canonical source for Hive's first set of global rules. It is intentionally empty: this delivery prepares the structure and proposed integration mechanism, not the rules themselves. An empty source must not be installed or interpreted as an uninstall request.

The content will be inserted as a managed block into the global instruction file actually read by each CLI. It does not replace the entire file or contain the markers; the integration adds them during installation. See the [architecture contract](../../_support/docs/architecture/repository-and-distribution.md).

The root `AGENTS.md` and `CLAUDE.md` govern development of this repository; they are not distributed content. Activity-specific instructions belong in `content/skills/<activity>/SKILL.md`.
