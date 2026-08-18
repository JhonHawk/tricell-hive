---
description: Build, extend, or audit an Astro Starlight docs site (user manual or specs site). scaffold | page | audit
---
Execute the skill `starlight-docs-site` now, with these arguments: $ARGUMENTS

1. Read !`echo ~/.agents/skills/starlight-docs-site/SKILL.md` and follow its modes
   (scaffold | page | audit) and profiles (user-manual | spec-site) exactly. Read
   the referenced `references/*.md` for the mode/profile you are in before acting;
   bundled templates live under !`echo ~/.agents/skills/starlight-docs-site/templates/`.
2. Translate Claude Code mechanics to opencode equivalents: the Skill tool → this
   command; AskUserQuestion (profile pick, the git-hooks/Basic-Auth scaffold
   questions) → an inline question; context7 version re-verification → direct MCP
   calls or `npm view`.
3. Honor the invariants: pnpm only, pinned exact versions, astro-mermaid before
   starlight, es-MX root locale, English identifiers / Mexican-Spanish content.
   `audit` is read-only. Verify from a production build (`pnpm build && pnpm check`),
   never the dev server.
