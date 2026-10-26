# Diseño

## Contexto verificado (2026-10-26, `development` en `2e45e292`)

- **Specs contra código.**
  - `_support/openspec/specs/versioned-installation/spec.md:62`: «The optional capability catalog SHALL be limited to Engram, Context7, pi-subagents, agent-browser, the Codex question card (`codex-question-card`), and CodeGraph.» El catálogo añade `CodexWaitFloor` (`tooling/providers/catalog.go:49`; ID `codex-wait-floor`, `tooling/providers/types.go:32`).
  - `versioned-installation/spec.md:179`: «… SHALL NOT read host or provider configuration.» `codexWaitFloorRow` (`tooling/cli/doctor_integrations.go:498`) lee `config.toml` de Codex y nunca repite su texto. La fila aparece solo con Codex registrado o con el registro ilegible (`:176-184`). Sus estados son `applied`, `partial`, `not applied`, `cannot edit` y `not checked`. Informa si Hive escribió las líneas y si existe `config.toml.hive-backup`.
  - `versioned-installation/spec.md:245` acepta `ultra` en cualquier host que admita esfuerzo. `integrations/agents/agents.go:390` lo rechaza en Pi («its thinking levels end at max»).
  - `_support/openspec/specs/documentation-site/spec.md:10`: «one page per flow skill», con siete páginas. Hay nueve skills `flow-*` en `content/skills/`, y `site/src/content/docs/flows/` tiene siete páginas.
- **Comportamiento del tiempo mínimo de espera** (diseño archivado de gh-142, `_support/openspec/changes/archive/2026-10-25-gh-142-codex-wait-floor/design.md`, puntos 2–4 de «Enfoque» y la tabla del editor):
  - no hay paso externo; solo se ofrece en interactivo, con Codex que carga la tabla, y si el mínimo no está ya puesto;
  - inserta solo las claves que faltan, copia antes el archivo a `config.toml.hive-backup` si no existe la copia, y se niega con un paso manual ante valores en conflicto, formas que no edita, enlaces simbólicos o falta de permiso;
  - el journal guarda huellas y las líneas propias, nunca el resto del archivo;
  - `hive plan remove` con Codex quita solo las líneas registradas que siguen iguales.
- **Causa.** `_support/openspec/changes/archive/2026-10-25-gh-142-codex-wait-floor/proposal.md:45`: «Delta de specs: no hace falta. Ninguna capability de `_support/openspec/specs/` cubre el instalador de capacidades opcionales…». Registro cerrado en `b1d3171e` (2026-10-25).
- **Regla actual.** `content/skills/flow-plan/references/change-records.md:15` da el criterio del delta y la exención de «implementation, tooling, wording, or the plan's own steps». No pide buscar ni registrar.
- **Prueba existente.** `tests/content/spec_deltas_test.go:410` (`TestSpecDeltasApplied`) recorre `changes/archive/*/specs/*/spec.md` y exige que el último delta de cada requisito coincida con la spec. No lee código.
- **Hogares y punteros del criterio del delta** (búsqueda de «delta», «needs no delta», «requirement delta» y «change records» en `content/`, `site/`, `_support/docs/architecture/`, `tests/pilot/`, `AGENTS.md` y `CONTRIBUTING.md`):
  - el criterio solo está en `change-records.md:15`;
  - `content/guidance/global.md:104` solo apunta a `change-records.md`;
  - `content/skills/flow-close/SKILL.md:69`, `content/skills/flow-plan/references/plan-revision.md:9` y `content/agents/quality/hive-verify-change.md:10` usan los deltas sin enunciar cuándo se escriben;
  - `site/src/content/docs/flows/plan.md:45` describe el change folder sin el criterio;
  - `content/skills/harness-audit/references/project-layout.md:10` (HA-PL-04) revisa la ubicación de las specs, no el criterio;
  - `_support/docs/architecture/workspace-and-artifacts.md:25` solo da la ubicación;
  - `tests/pilot/` no tiene detectores del criterio (sus coincidencias son de huellas, `protected_test.go`).
- **Trabajo en paralelo.** La rama de #204 (`feat/gh-204-whole-reads-reference-split`) modifica `global.md`, `plan-format.md`, `site/src/content/docs/flows/plan.md` y `CHANGELOG.md`, pero no `change-records.md` ni las specs (`git diff --name-only 2e45e292 feat/gh-204-whole-reads-reference-split`). Lo que comparten las dos ramas en `site/.../plan.md` y `CHANGELOG.md` son líneas distintas; quien integre segundo resuelve el conflicto de texto.

## Enfoque

### Regla de búsqueda (AC6)

Va en `change-records.md:15`, justo después de la frase de exención, porque es el único sitio que decide si hay delta; ningún otro archivo la repite. Texto propuesto:

> Before deciding which requirements a change alters, or that it needs no delta, and before applying the exemption above, search the current-requirements home, and any older document that [Areas still described by older documents](#areas-still-described-by-older-documents) keeps authoritative for the area, for each capability, behavior, command, view, setting, and integration the change adds or alters, each file a user, host, or other process reads or writes that it adds or alters, and the list, catalog, or report that an added item joins, since a requirement that enumerates a closed set never names the item being added. Read every requirement or product-map rule those searches find, including one that enumerates a closed set, such as a "limited to" list or the list of what a command reports, and read every prohibition, such as a SHALL NOT, in each capability spec or product-map page the searches touch, since a prohibition in general terms names neither the item nor its list: adding to that set or doing what it forbids alters it. Record in `proposal.md` the terms searched and the requirements read, or that nothing matched.

Por qué así:

- La falla de gh-142 fue no buscar. Nombrar las listas cerradas y las prohibiciones apunta a las dos formas que rompió, y son las que una búsqueda por la palabra de la capacidad nueva no encuentra, porque la capacidad nueva aún no aparece en ellas. Por eso la búsqueda incluye lo que el cambio altera, no solo su nombre nuevo: `doctor`, `config.toml`, el catálogo.
- La búsqueda vale también cuando el plan sí escribe un delta: un delta parcial que toca una capability puede dejar rota una lista cerrada de otra, como pasó en gh-142.
- Va antes de la exención de «implementation, tooling…», porque en Hive `tooling/` es producto (el instalador de gh-142 lo era) y un modelo literal aplicaría la exención sin buscar.
- El registro en `proposal.md` sigue la práctica existente: los planes ya escriben «Deltas de requisitos: ninguna spec…» en Excluido (por ejemplo `2026-10-26-gh-200-unattended-expiry-preauthorization/proposal.md`). La regla pide que esa línea diga qué se buscó.
- «current-requirements home», «product-map rule» y «older document» cubren los proyectos con mapa de producto y las áreas que aún describen documentos antiguos (`change-records.md:59-61`); «such as with SHALL NOT» cubre las reglas de negocio que prohíben sin esa palabra.

Alternativas descartadas:

- **Una prueba que compare specs con código:** el paper muestra que esos chequeos son caros y poco fiables, y no sabría qué lado está mal. Queda en vigilancia.
- **Ponerla en `flow-close`:** al cerrar, el código ya está integrado y el plan ya decidió; la búsqueda sirve antes de escribir las tareas.
- **Repetirla en `plan-format.md` o en `hive-review-plan`:** sería una segunda casa de la misma regla (AGENTS.md, «one canonical home»), y #204 está reescribiendo `plan-format.md`.

Regla de admisión (AGENTS.md, Measurement): es una corrección con un solo caso observado, gh-142 (cerrado en `b1d3171e`, 2026-10-25, sin ID de sesión registrado en su change record). El usuario aprobó adoptarla ya (respuesta «R2 specs + regla», 2026-10-26). El commit nombra la falla (gh-142, 2026-10-25), dice que el ID de esa sesión de planificación no se pudo recuperar porque no consta en su change record ni en `_support/sessions/`, y nombra esta sesión (Claude Code, `abdd6ea1-e8bb-4ebc-8d12-6284d022d322`, 2026-10-26), que la observó.

### Deltas (AC1–AC5)

- Cada MODIFIED es el requisito completo con sus escenarios, copiado de la spec en `2e45e292` y cambiado solo donde hace falta:
  - **Optional operation capabilities:** `codex-wait-floor` en la lista y una oración sobre qué escribe, qué rechaza y cómo se deshace. El escenario «Codex wait floor applied and removed» se quitó por decisión del usuario (D4-A, 2026-10-26), tras tres verificaciones que encontraron estados en que su resultado no se cumple. No fija valores en milisegundos ni versiones de Codex, para no meter detalle que envejece.
  - **Read-only diagnostics:** la fila del tiempo mínimo de espera entre lo que se informa, y la excepción acotada a `config.toml` de Codex en la prohibición, más el escenario «Codex wait floor read without repeating the file».
  - **Per-role model overrides:** `ultra` en Pi en la condición de rechazo y en el escenario «Unsupported effort».
  - **Published documentation site:** «one page for each of these flow skills» y una oración que nombra `flow-report` y `flow-walkthrough` como skills sin página.
- Se aplican en el cierre, como fija `flow-close/SKILL.md:69`. En el build solo se comprueban contra el código y en una copia temporal.

## Sitio y changelog

- `site/src/content/docs/flows/plan.md:45`: una frase después de la que describe el change folder: el plan dice en qué requisitos buscó antes de concluir que el cambio no toca ninguno.
- `CHANGELOG.md`, `[Unreleased]` → `Changed`: una entrada con ese mismo efecto y la corrección de las cuatro specs.
- README y `llms.txt` no describen este comportamiento; no cambian.
