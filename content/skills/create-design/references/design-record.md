# Design record

How to detect a project's design system and keep it in `DESIGN.md`. It applies to any component library; HeroUI and PrimeNG appear as examples, not as the list.

## Detection

Record an item only when you have a file path as evidence.

1. Find the package manifest of the app that owns the UI (in a monorepo, the app the changed screens belong to) and identify the component library from the dependencies and configuration the screens import from (`components.json` marks shadcn and ReUI; the `heroui()` Tailwind plugin marks HeroUI v2; `providePrimeNG` marks PrimeNG). With more than one library, use the one the changed screens import.
2. Take the installed version from the lockfile (`package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `bun.lock`) or the installed package's `package.json`; a manifest range is not a version, and on a difference record the lockfile value. For libraries that copy their code into the project (shadcn, ReUI), also record the `style` and registry from `components.json` and the versions of the primitives the copied components use.
3. Find the customization layer and record the path of each piece: the preset or theme (`definePreset` in PrimeNG v20, the `heroui()` theme in HeroUI v2), CSS variables, `@theme` or the Tailwind theme (HeroUI v3), global overrides (`pt` and `cssLayer` in PrimeNG), and the project's own wrapper components and copied component code.
4. When the library's MCP server or skill is installed, use it before reading files and record which one. Read the documentation of the installed version (Context7 when installed, otherwise the official site) before stating any API or token, since a major version changes the customization mechanism.

When the project has no tokens, theme or component library, record exactly that.

## Format

The file follows the Google Labs DESIGN.md format (alpha): YAML frontmatter (`version: alpha`, `name`, `description`, and only those of `colors`, `typography`, `rounded`, `spacing`, `components` for which you found values), then optional prose sections in this order: Overview, Colors, Typography, Layout, Elevation & Depth, Shapes, Components, Do's and Don'ts. A final `## Hive` section, which the format keeps, holds:

- **Library:** name, installed version, the file the version came from; for a copied-code library also the `style`, the registry and each primitive with its installed version.
- **Customization layer:** kind and path of each piece.
- **Library tools:** the MCP server or skill, if any.
- **Design file:** the design tool and the location of the app's design file, only when the user gives it, written only at the user's instruction, which the plan's task carries out. It says where [create-design](../SKILL.md) draws examples for the app's screens and authorizes no write by itself; it has no source file, carries no fingerprint and never makes the record stale.
- **Design library:** the design tool and the name of its published library that holds the app's components and variables, as the tool lists it, only when the user gives it or confirms the one [create-design](../SKILL.md) found in use, written only at the user's instruction. That stage draws with instances of it. Like the design file, it has no source file, carries no fingerprint, never makes the record stale and authorizes no write; the stage asks for the instruction and keeps a yes pending, writing nothing, as [design library](design-library.md#record-it) sets.
- **Product:** users, main tasks, the mode of each area ([criteria](criteria.md#screen-modes)), constraints, and any accessibility requirement.
- **Decisions:** approved deviations from the library and the app's conventions, each with date and origin. A convention is one the user set while correcting an example of [create-design](../SKILL.md) for screens beyond that one, such as the row actions of every list, written only at the user's instruction that stage keeps in the example's record; its origin is where that record is kept (the plan file that carries it, otherwise the Figma file and the frame's node). This is the list that precedence level 2 reads.
- **Sources:** every source file with its fingerprint, the output of `git hash-object <path>`; a lockfile gets none, since any dependency change alters it.
- **Watch:** globs for the customization layer, such as `src/styles/**/*.css`.

Every path in the record is relative to the directory that holds `DESIGN.md` (a monorepo's root lockfile is `../../pnpm-lock.yaml`), and every `git hash-object` runs from that directory. Every value in the frontmatter and prose names its source file; invent no value. When the project has no design system, write that, and label any starter values as a proposal. Place the file at the root of the app that owns the UI, one per app in a monorepo.

## Staleness

Compare only what the record holds, without repeating the detection. It is stale when a version under **Library** (the library's or a copied-code primitive's) differs from what the lockfile resolves, when the `git hash-object` of a **Sources** file differs from its fingerprint or the file no longer exists, or when a **Watch** glob matches a file not in **Sources**. When stale, name what changed and propose the refresh; do not apply stale tokens silently.

## Writing

This section is the only home of the rule, including the `**Design file:**` and `**Design library:**` fields: a plan task writes them as it writes any other field. [Design library](design-library.md#record-it) is the home of how the drawing stage asks for them and keeps a yes pending.

- Write or refresh the record only through a task in a UI change's plan, which the plan adds when the record is missing or stale and delivers with the change, or when the user asks.
- Never overwrite it without showing the change: present the diff inside the plan task or the reply to the user.
- When `@google/design.md` is already installed in the project, run its `lint` and fix its errors. Never install it, and never run it through a one-off package runner.
