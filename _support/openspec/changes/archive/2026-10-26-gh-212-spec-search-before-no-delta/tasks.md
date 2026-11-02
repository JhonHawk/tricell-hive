# Tareas

Commit base: `f14eb791` (`origin/development`, 2026-10-26, con #204 ya integrado). Worktree `.claude/worktrees/gh-212`, rama `feat/gh-212-spec-search-before-no-delta`.

## T1 — Regla de búsqueda antes de declarar «sin delta»

- [x] `change-records.md` pide buscar en el hogar de requisitos vigentes, leer los requisitos que nombran lo que el cambio toca (incluidas listas cerradas y `SHALL NOT`) y anotar la búsqueda en `proposal.md`; el sitio y el changelog lo describen.

  reabierta por la corrección de la revisión de código (`c54f5f77`): la búsqueda incluye la lista o el catálogo al que se suma un elemento, sin archivos de código, y la tabla «Files» nombra el registro de la búsqueda. AC6 met en `c54f5f77`.
  reabierta por las correcciones tras la re-revisión (`0bb52a38`): busca también los archivos que lee un usuario o host, lee todas las prohibiciones de cada spec tocada, enlaza la sección de documentos antiguos; `plan-format.md` apunta al registro; sitio y changelog describen sin copiar la regla. AC6 met en `0bb52a38`.

**Closes:** AC6

**Checkpoint:** C1

**Depends on:** none

**Locations:** `content/skills/flow-plan/references/change-records.md:15,24`; `content/skills/flow-plan/references/plan-format.md:61` (puntero, añadido tras la revisión de código); `site/src/content/docs/flows/plan.md:45`; `CHANGELOG.md` (`[Unreleased]` → `Changed`).

**Execution:** hilo principal; son tres ediciones cortas de texto ya redactado en `design.md`.

**Test approach:** check — `rg` sobre la regla y sus punteros, y las pruebas de contenido.

**Changes:** insertar la oración de [Regla de búsqueda](design.md#regla-de-búsqueda-ac6) después de «…needs no delta.»; añadir la frase del sitio y la entrada del changelog descritas en [Sitio y changelog](design.md#sitio-y-changelog).

**Verification:**

- `rg -n "terms searched and the requirements read" content/` → una sola coincidencia, en `change-records.md`.
- `rg -n -i "closed set|SHALL NOT" content/skills/flow-plan/references/change-records.md` → la oración nueva.
- `go test ./tests/content/...` → pasa (enlaces, tamaño de skills y huella de lectura).

## T2 — Deltas contrastados con el código

- [x] Los cuatro MODIFIED de `specs/` describen el código de la base del build, y aplicados en una copia temporal cambian solo esos cuatro requisitos y pasan la comprobación de deltas.
  ronda 1 · AC1 · `hive plan remove` borra tabla y archivo solo si no queda otra cosa en ellos, y el respaldo solo si quitó todas las líneas registradas en esa pasada (`integrations/codex/wait_floor.go:433-461`, `tooling/management/codex_wait_floor.go:338-346`); el delta prometía borrarlos siempre que no quedaran líneas registradas. Corregido; re-verificación de AC1.
  ronda 2 · AC1 · la retirada borra la clave cuyo valor sigue igual aunque su línea cambie (`integrations/codex/wait_floor.go:423-431`); el escenario no excluía un respaldo previo ni un archivo sin salto final. Corregido (claves con su valor, escenario acotado, «other lines»); ronda 3 de AC1.
  ronda 3 · AC1 · el texto del requisito coincide con el código; el escenario «Codex wait floor applied and removed» admite estados en que su THEN falla (una cadena `"""` añadida fuera de la tabla hace que la retirada no borre nada, `integrations/codex/wait_floor.go:221-223,415-417`; otro home de Codex al retirar, `tooling/management/codex_wait_floor.go:153-159`; oferta solo manual seleccionada). Decisión del usuario D4-A (2026-10-26): se quita el escenario; el texto del requisito, verificado en la ronda 3, cumple AC1. AC2–AC5 met en la ronda 1.

**Closes:** AC1, AC2, AC3, AC4, AC5

**Checkpoint:** C1

**Depends on:** none

**Locations:** `_support/openspec/changes/gh-212-spec-search-before-no-delta/specs/{versioned-installation,documentation-site}/spec.md`; `tooling/providers/catalog.go`, `tooling/providers/types.go`; `tooling/cli/doctor_integrations.go` (`codexWaitFloorRow` y su llamada); `integrations/codex/wait_floor.go`; `integrations/agents/agents.go` (rechazo de `ultra` en Pi); `content/skills/`, `site/src/content/docs/flows/`.

**Execution:** hilo principal; los deltas ya están escritos en el plan y la tarea solo los contrasta y corrige.

**Test approach:** check — contraste por afirmación y aplicación en copia temporal.

**Changes:** en la base del build, comprobar cada afirmación nueva de los deltas contra el código citado (IDs del catálogo, cuándo se ofrece, qué inserta, qué rechaza, qué guarda el journal, qué deshace `hive plan remove`, estados y evidencia de la fila del doctor, rechazo de `ultra`, skills sin página) y corregir el delta donde no coincida. Luego copiar `_support/openspec` a una carpeta del scratchpad, aplicar los deltas como fija `flow-close/SKILL.md:69`, mover el change a `changes/archive/<fecha>-gh-212-…` en la copia y comprobar.

**Verification:**

- `rg -n "ID:" tooling/providers/catalog.go` y `rg -n ' ID = "' tooling/providers/types.go` → los siete IDs del catálogo, con sus cadenas, coinciden con la lista de AC1.
- En la copia: `diff` entre la spec original y la aplicada → solo cambian los bloques de los cuatro requisitos. Es la comprobación principal de AC5.
- En la copia: copiar solo `tests/content/spec_deltas_test.go`, que importa solo la biblioteca estándar, a un módulo del scratchpad con su propio `go.mod` y la constante `openspecRoot` apuntando a la copia, y correr `go test -run TestSpecDeltasApplied` → pasa. El archivo rastreado no se edita.

## Verificación y revisión humana

- **C1** — T1, T2 — ninguna tarea cambia un contrato público ni tiene dependientes; las dos se verifican juntas al final.

| Puerta | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Pruebas de contenido | por tarea, en la rama de trabajo | `go test ./tests/content/...` pasa | implementación |
| Suite completa | candidato final | `go vet ./...` y `go test ./...` pasan | implementación |
| Sitio | candidato final | `npm run build` en `site/` termina sin errores | implementación |
| Revisión dedicada | candidato local, antes de publicar | según la respuesta de revisión | pendiente (ronda de entrega) |
| Cierre | después de integrar | `flow-close` aplica los deltas, archiva y `TestSpecDeltasApplied` pasa sobre el repositorio | ruta de cierre del registro, pendiente (ronda de entrega) |

Sin verificación de UI: el sitio solo cambia una frase de texto.

## Estado de revisión y progreso

- **Revisión del plan** sobre la revisión `39a854e4d86e` (2026-10-26), con dos `hive-review-plan` en paralelo:
  - **Coherencia de la guía** (~109 mil tokens, 85 s): sin bloqueantes; el criterio del delta solo vive en `change-records.md:15`. Aceptadas: la regla cubre también los planes que escriben delta, va antes de la exención, alcanza mapas de producto y documentos antiguos, se quita la afirmación de que la revisión del plan comprueba el registro (queda en vigilancia), los requisitos del PR y del commit pasan a «Entrega», y el commit dice que no se recuperó el ID de sesión de gh-142.
  - **Deltas contra el código** (~140 mil tokens, 144 s): sin bloqueantes; encabezados y texto sin cambios copiados exactos; doctor solo lee `config.toml` de Codex. Aceptadas: el tiempo mínimo crea tabla o archivo si faltan, respalda solo un archivo existente, se ofrece en alcance de usuario cuando el plan apunta a Codex con `codex` en `PATH`, queda como paso manual si el registro es de otro home, `hive plan remove` borra tabla, archivo y respaldo creados por Hive; la fila del doctor no tiene estado de onboarding; la verificación de T2 usa `types.go` y un módulo aparte para la prueba.
  - Sin nueva ronda: las correcciones aplican las propuestas de los revisores.
- **Decisión tomada sin preguntar:** AC6 se amplió respecto del AC6 de #212 (también cuando el plan escribe delta, y antes de la exención), por la propuesta del revisor de guía; se revierte volviendo al texto del ticket.
- **Build:** T1 y T2 `[x]`. Implementador: Claude Opus 5.5 (sesión principal, observado). Verificadores `hive-verify-task`: `opus` configurado, no observado.
- **Revisión dedicada:** `/code-review` (medium) completo sobre `a8db0496` (8 hallazgos, corregidos en `c54f5f77`) y una re-revisión de `a8db0496..c54f5f77` (7 hallazgos, corregidos en `0bb52a38` sin otra revisión). Quedan sin corregir: nadie revisa en la revisión del plan la búsqueda anotada (en vigilancia) y la lista de siete páginas de `documentation-site`.
- **Puertas:** `go vet ./...` y `go test ./...` pasan (20 paquetes); `go test ./tests/content/...` pasa sobre `0bb52a38`; `npm run build` en `site/` termina.
- **Integración:** [PR #215](https://github.com/JhonHawk/tricell-hive/pull/215), merge `678010d2` (2026-10-26). Deltas aplicados y registro archivado en el PR de cierre.
- **Siguiente paso:** ninguno.
