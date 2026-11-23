# Design

## Approaches

Una sola forma sensata: la regla tiene su casa en el Kill de `no-comments`, que `audit-comments` ya declara como su criterio («read it before step 4»), y la sección del proyecto sale del paso 9, que ya escribe en el `AGENTS.md` del repositorio. Una referencia o un paso nuevo agregaría texto sin una condición de lectura distinta.

## Decisions

D3 a D7 de `proposal.md`.

## Texto

**`no-comments.md`, viñeta Kill** (línea 8), queda:

> - **Kill:** it narrates the code (`// loop over users`), restates a name or a type, restates a rule the code, its tests or its configuration in the repository already express, even when it states only part of the rule, marks a section the structure already shows, holds commented-out code, or records a past change (git holds history). A restated rule is a Kill only when the evidence names the line, test or setting that expresses it. A part the code does not express, a detail the code contradicts included, falls outside that Kill: cut the comment down to that part and sort it on its own, as this step sets.

**`no-comments.md`, línea de reporte** (línea 19): tras «Report the number of comments deleted,» agrega «each restated-rule Kill with the line, test or setting it names,».

**`audit-comments/SKILL.md`, Tautology** (línea 22), queda:

> - **Tautology** (Kill): it repeats what the code, a name or a type already says, including a rule the code expresses as the leaf's Kill sets, or marks a section the structure shows.

**`audit-comments/SKILL.md`, paso 9**: tras «registered for every host the project uses» agrega «, and the project's comment rule written to its `AGENTS.md`». **Paso 10**: la viñeta del lint agrega «and the `## Comments` section written, with its path». **Descripción**: tras «generate the project's own skill to repeat it, in .agents/skills with a link in .claude/skills» agrega «, and write the project's comment rule in its AGENTS.md».

**`project-skill.md`**:
- La intro nombra también la regla del proyecto.
- El punto 5 de «What it holds» cambia «and the workaround marker's form» por «and a pointer to the `## Comments` section of `AGENTS.md` for the workaround marker's form», para que la forma viva una vez en el proyecto.
- En «Register it», antes de la viñeta de `## Skills`:

> - In the same repository's `AGENTS.md`, write a `## Comments` section above `## Skills`, outside `## Hive` and any installer-managed block, from this template, with the project's comment lint command, the paths it checks (its `SOURCE_PATTERNS` minus `EXCLUDED`), decision record path and tracker in place of each `<…>`; when the file already has a `## Comments` section, update it instead of adding another. A declared workspace does not get it, since the lint runs in each repository. When that level's `CLAUDE.md` does not import `@AGENTS.md`, report it as audit-harness's [hierarchy](../../audit-harness/references/hierarchy.md) HA-HI-03 sets.
>
> ```markdown
> ## Comments
>
> Do not add comments: `<comment lint command>` rejects a new or returned comment in `<paths the lint checks>`. Put each fact where it belongs instead, and say what the code does through names, types, tests and structure:
>
> - A decision or an outside constraint, such as why a surprising choice was made or a vendor or protocol quirk, goes to an entry in `<decision record>` with the paths it concerns.
> - A shortcut or pending work, such as a TODO, a silent catch or a fallback, goes to a ticket in `<tracker>`, with the workaround marker over the code.
> - Something the next editor must not do becomes a check that enforces it (a type, an assertion, a test or a lint rule), proposed before it is added; until one exists, an entry in `<decision record>`.
>
> Only these comments stay:
>
> - Directives a tool reads, such as a lint suppression with its reason or a bundler's magic comment.
> - License headers.
> - The workaround marker, on one line: the line-comment token, then `workaround: ` and the ticket's key or URL, or the decision entry as `<path>#<anchor>` while no ticket exists; `{/* workaround: <ticket> */}` inside JSX.
>
> Existing comments stay until a cleanup batch removes them.
> ```

**`single_home_test.go`**: fila `{"restates a rule the code", "content/skills/hive-core/review/no-comments.md"}`.

**`_support/docs/architecture/instruction-resources.md:28`**: la descripción de `references/project-skill.md` pasa a «where the generated skill lives, what it holds, its `## Skills` registration and the template of the project's `## Comments` section, an output copy of the leaf's ban paragraph for the audited project».

**`site/src/content/docs/skills.md:46`**: «…leave the project its own skill to repeat the cleanup, readable by every host» pasa a «…leave the project its own skill to repeat the cleanup, readable by every host, and write the project's comment rule in its `AGENTS.md`».

**`CHANGELOG.md`**, `### Changed` de `[Unreleased]`: una entrada con los dos cambios.

## Impact

`no-comments` se lee en cada verify (paquete `stage-entry`) y en cada lote de `audit-comments`: la cláusula borra más comentarios en las revisiones de candidatos, solo cuando la evidencia nombra lo que expresa la regla, y nunca la parte que el código no expresa. La sección `## Comments` llega solo a proyectos que corren la auditoría desde ahora. Una auditoría de harness (HA-IF-05) puede proponer quitar la primera frase de la sección porque el lint la cubre, y el usuario decide entonces; el resto (a dónde va cada hecho, qué se queda) no lo cubre ningún mecanismo.

## Recovery

Revertir el commit; los proyectos que ya recibieron la sección la conservan hasta que alguien la quite.
