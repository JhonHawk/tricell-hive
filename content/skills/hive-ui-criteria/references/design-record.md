# Design record

How to detect a project's design system and keep it in `DESIGN.md`. It applies to any component library; HeroUI and PrimeNG appear as examples, not as the list.

## Detection

Record an item only when you have a file path as evidence.

1. Find the package manifest of the app that owns the UI. In a monorepo, choose the app that the changed screens belong to.
2. Identify the component library from the dependencies and configuration files the screens import from. Examples: `components.json` marks shadcn and ReUI; the `heroui()` plugin in the Tailwind configuration marks HeroUI v2; `providePrimeNG` in the application configuration marks PrimeNG. With more than one library, use the one the changed screens import.
3. Take the installed version from the lockfile (`package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `bun.lock`) or from the `package.json` inside the installed package. The manifest range (`^2.4.0`) is not a version: when they differ, record the lockfile value. For libraries that copy their code into the project (shadcn, ReUI), also record the `style` and registry from `components.json` and the versions of the primitives the copied components use.
4. Find the customization layer and record the path of each piece:
   - the preset or theme: `definePreset` from `@primeuix/themes` in PrimeNG v20; the `heroui()` plugin theme in HeroUI v2;
   - CSS variables, `@theme` or the Tailwind theme: CSS variables over `@heroui/styles` in HeroUI v3;
   - global overrides: `pt` and `cssLayer` in PrimeNG;
   - the project's own wrapper components and copied component code.
5. When the library's MCP server or skill is installed, use it before reading files. Record which one.
6. Read the documentation of the installed version (Context7 when installed, otherwise the official site) before you state any API or token. A major version changes the customization mechanism, as HeroUI v2 and v3 show.

When the project has no tokens, theme or component library, record exactly that.

## Format

The file follows the Google Labs DESIGN.md format (alpha): YAML frontmatter, then prose sections.

- Frontmatter: `version: alpha`, `name`, `description`, and only those of `colors`, `typography`, `rounded`, `spacing`, `components` for which you found values.
- Prose sections, each optional, in this order: Overview, Colors, Typography, Layout, Elevation & Depth, Shapes, Components, Do's and Don'ts.
- A final `## Hive` section; the format keeps sections it does not know. It holds:
  - **Library:** name, installed version, the file the version came from; for a library that copies its code, also the `style`, the registry and each primitive with its installed version.
  - **Customization layer:** kind and path of each piece.
  - **Library tools:** the MCP server or skill, if any.
  - **Product:** users, main tasks, the mode of each area ([criteria](criteria.md#screen-modes)), constraints, and any accessibility requirement.
  - **Decisions:** approved deviations from the library, each with date and origin. This is the list that precedence level 2 reads.
  - **Sources:** every source file with its fingerprint, the output of `git hash-object <path>`. A lockfile gets no fingerprint, because any dependency change alters it; reading it covers the version only.
  - **Watch:** globs for the customization layer, such as `src/styles/**/*.css`.
- Every path in the record, in **Library**, **Sources**, **Watch** and the value comments, is relative to the directory that holds `DESIGN.md` (a monorepo's root lockfile is written as `../../pnpm-lock.yaml`, for example), and every `git hash-object` runs from that directory.
- Every value in the frontmatter and prose names its source file. Invent no value. When the project has no design system, write that, and label any starter values as a proposal, not as project tokens.
- Place it at the root of the app that owns the UI; one per app in a monorepo.

Example:

```markdown
---
version: alpha
name: Example app
description: Back-office UI on a component library.
colors:
  primary: "#2563eb"   # src/styles/theme.css
rounded:
  md: 6px              # src/styles/theme.css
---

## Overview

Dense back-office screens for operators.

## Hive

- **Library:** ExampleUI 3.1.0 (pnpm-lock.yaml)
- **Customization layer:** CSS variables, src/styles/theme.css; wrapper, src/components/Button.tsx
- **Library tools:** none found
- **Product:** operators; review and approve orders; area Orders = operate
- **Decisions:** 2026-01-10, primary button uses the brand color (ticket EX-12)
- **Sources:** src/styles/theme.css 1f3a9c0e...; src/components/Button.tsx 7be21d44...
- **Watch:** src/styles/**/*.css
```

## Staleness

Do not repeat the detection; compare only what the record holds. The record is stale when any of these is true:

- a version recorded under **Library**, the library's own or one of the primitives a copied-code library uses, differs from the version the lockfile resolves;
- the `git hash-object` of a file in **Sources** differs from its recorded fingerprint, or the file no longer exists;
- a **Watch** glob matches a file that is not in **Sources**.

When stale, name what changed and propose the refresh. Do not apply stale tokens silently.

## Writing

This section is the only home of the rule.

- Write or refresh the record only through a task in a UI change's plan, which the plan adds when the record is missing or stale and delivers with the change, or when the user asks.
- Never overwrite it without showing the change: present the diff inside the plan task or the reply to the user.
- When `@google/design.md` is already installed in the project, run its `lint` and fix its errors. Never install it, and never run it through a one-off package runner.
