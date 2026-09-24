---
name: starlight-docs-site
description: Scaffold, extend, or audit a Starlight documentation site when documentation is meant to live as a navigable site.
---

# Starlight documentation sites

Use this skill when readers need a maintained, searchable documentation site. A one-file report belongs to `flow-report` instead.

Choose a mode before editing:

- **scaffold**: copy either [user-manual assets](assets/user-manual/) or [spec-site assets](assets/spec-site/) into the explicitly chosen project location. Preserve every established choice of an existing project, including its package manager, locale, theme, content model, and deployment. The supplied assets use exact compatible baseline versions: Astro 7.0.7, Starlight 0.41.3, TypeScript 6.0.3, and pnpm 11.21.0.
- **page**: add or revise a page in an existing site, using its established content structure and navigation. Do not create empty page trees.
- **audit**: check the built site, internal links, page metadata, heading order, keyboard navigation, and meaningful alternative text. Record failures with their page and impact. Report observations separately from changes; do not repair unless authorized.

Use the **user-manual** profile for task-oriented operational documentation: organize pages around actions a reader performs, state prerequisites and verification, and keep critical instructions understandable without screenshots. Use **spec-site** for product rules, acceptance criteria, and decisions: state the scope, sources, status, and open questions accurately. Specs are optional: add a workflow schema only when the project already uses one or explicitly needs one, and do not infer that every documentation site needs it. Follow the project's established voice and locale. The supplied sample content is intentionally small; replace it in the project language.

Install dependencies only in the target project and run its declared checks. The assets do not configure hooks, authentication, deployment, commits, or a fixed language. Measure color contrast only for a concrete foreground/background pair, using [the contrast checker](scripts/contrast-check.py) with `uv run --with coloraide python <skill-dir>/scripts/contrast-check.py <foreground> <background>`.
