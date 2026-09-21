# Shared global instructions

[`global.md`](global.md) is the canonical source for Hive's first set of shared global rules: scope and authorization, preservation, evidence, proportionality, continuity, and artifact placement. The source is authored but has not been installed or behaviorally validated across the target CLIs.

The content will be inserted as a managed block into the global instruction file actually read by each CLI. It does not replace the entire file or contain the markers; the integration adds them during installation. See the [architecture contract](../../_support/docs/architecture/repository-and-distribution.md).

The root `AGENTS.md` and `CLAUDE.md` govern development of this repository; they are not distributed content. Activity-specific instructions belong in `content/skills/<activity>/SKILL.md`.

The essential workspace rules are included in `global.md` and do not depend on a skill or an external reference being loaded. The [workspace-conventions skill](../skills/workspace-conventions/SKILL.md) carries the procedure for requested organization, evidence curation, promotion, and archiving. Its availability and loading must be verified in each host; authoring it here does not install it.

The [workspace and artifact explanation](../../_support/docs/architecture/workspace-and-artifacts.md) supports maintainers with examples; it is not a runtime dependency. An empty source must not be installed or interpreted as an uninstall request. The [frozen measurement baseline](../../tests/fixtures/workspace-conventions/baseline/global.md) is historical test input, not another maintained rule source.
