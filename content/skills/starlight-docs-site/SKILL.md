---
name: starlight-docs-site
description: Scaffold, extend, or audit a Starlight documentation site when documentation is meant to live as a navigable site.
---

# Starlight documentation sites

Use this skill when readers need a maintained, searchable documentation site. A one-file report belongs to `flow-report` instead.

Choose a mode before editing:

- **scaffold**: copy either `assets/user-manual/` or `assets/spec-site/` into the explicitly chosen project location. Preserve an existing project's package manager, locale, theme, content model, and deployment choices. The supplied assets use exact compatible baseline versions: Astro 7.0.7, Starlight 0.41.3, TypeScript 6.0.3, and pnpm 11.21.0.
- **page**: add or revise a page in an existing site, using its established content structure and navigation. Do not create empty page trees.
- **audit**: inspect structure, links, metadata, accessibility, and build results. Report observations separately from changes; do not repair unless authorized.

Use the **user-manual** profile for task-oriented operational documentation. Use **spec-site** for product rules and approval evidence. Specs are optional: do not infer that every documentation site needs their schema or workflow fields. The supplied sample content is intentionally small; replace it in the project language.

Install dependencies only in the target project and run its declared checks. The assets do not configure hooks, authentication, deployment, commits, or a fixed language. Read [the user-manual profile](references/profile-user-manual.md) or [the spec-site profile](references/profile-spec-site.md) when selecting that profile. Read [the audit checklist](references/audit-checklist.md) only for an audit. Use `uv run --with coloraide python scripts/contrast-check.py <foreground> <background>` when a color-pair contrast measurement is needed.
