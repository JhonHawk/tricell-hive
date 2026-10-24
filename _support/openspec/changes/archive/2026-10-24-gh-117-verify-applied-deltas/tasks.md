# Tareas — gh-117-verify-applied-deltas

Commit base: `dccc1663` (`development`, 2026-10-24). Rama `feat/gh-117-verify-applied-deltas` en el worktree `.claude/worktrees/gh-117`, que también aloja esta carpeta del cambio (sin commit).

## Diseño

- **Comprobación al cerrar.** En `flow-close` «Close the change record», el paso 1 conserva «review the resulting diff» (detecta ediciones fuera de los bloques) y añade después la comprobación: en `<specs>/specs/`, cada requisito ADDED o MODIFIED del delta aparece con el mismo texto, sin contar diferencias de espacios, y no queda ningún encabezado `### Requirement: <name>` de un REMOVED (su cuerpo es un motivo, no texto del spec); en un mapa de producto, el texto nuevo de cada regla ADDED o MODIFIED, sin su marca ni la regla vigente citada, está en su página, y cada REMOVED ya no. Si no coincide, se corrige el spec antes de seguir. Un spec que se crea toma como título y propósito el preámbulo del delta. El paso 2 marca Completado solo tras la comprobación.
- **Abandono.** La línea «Abandon» de `flow-close` añade que, al archivar sin aplicar, la carpeta de deltas del cambio (`specs/` o `product/`) se renombra con el prefijo `unapplied-` (`unapplied-specs/`, `unapplied-product/`). Así ninguna comprobación lee como aplicado un delta que nunca se aplicó, sin depender del idioma del estado. `change-records.md`, hogar canónico del formato de carpetas, lo registra en su inventario como una diferencia con OpenSpec.
- **Prueba del repositorio** (`tests/content/spec_deltas_test.go`). Recorre `_support/openspec/changes/archive/*/specs/*/spec.md` en orden de nombre de carpeta, que empieza por la fecha de cierre, y toma de cada requisito (capacidad + nombre) la operación y el bloque del delta más reciente. Ignora el texto anterior al primer `## ADDED|MODIFIED|REMOVED Requirements` (preámbulo) y los requisitos que toca algún cambio activo en `_support/openspec/changes/<id>/specs/` (fuera de `archive/`), porque se aplican al cerrarlo. Compara el bloque normalizado (espacios colapsados) con el de `_support/openspec/specs/<capacidad>/spec.md`: ADDED o MODIFIED deben estar con el mismo texto; para REMOVED solo cuenta que no quede el encabezado. Cada fallo nombra la capacidad, el requisito, la operación y la carpeta archivada que se usó. Dos carpetas con la misma fecha que tocan el mismo requisito se ordenan por nombre; si eso da un falso fallo, el mensaje nombra la carpeta y un delta posterior lo resuelve (hoy no hay ningún caso que quede así, porque deltas posteriores reemplazan los dos empates conocidos).
- **Base reconciliada con los deltas de este cambio**, que `flow-close` aplica al cerrar:
  - `specs/context-delivery/spec.md` (ADDED, cinco requisitos con los nombres del delta archivado, y un preámbulo `# context-delivery` con su `## Purpose` para crear el spec). Investigación del 2026-10-24 sobre `dccc1663`: «Optional inactive OpenCode context resource» sigue tal cual; «Kernel references validated before installation» y «Portable kernel and situational procedures» siguen, cambiando «kernel» por «global core» (`content/guidance/global.md`); «Fixed-role context delivery without authority expansion» cambia la resolución del rol al identificador nativo (`data.id`, luego `id`, luego el agente del evento, nunca la etiqueta visible; `integrations/opencode/context-loader.js:20`); «Declared source footprint without token claims» cambia la frase de lectura de `reporting` antes de cada respuesta por la de cota superior conservadora del escenario de consulta simple (`content/guidance/global.md:17`, commit `b4417b91`; `tests/content/reading_footprint_test.go:86-89`).
  - `specs/walkthrough-viewer/spec.md` (MODIFIED «Current source classification»): registra el texto vigente del spec, que ya describe el visor actual («project/session list», «archive state»), para que el delta más reciente coincida.
- **Entrega de la prueba con el cierre (D11-A).** El PR principal lleva T1, T4 y T5. La prueba de T2 se escribe y verifica en el árbol de trabajo, pero queda sin commit en la rama principal. Entra con el PR de cierre, junto con los deltas aplicados y el registro archivado, de modo que `development` nunca queda en rojo. Ese PR lleva código y va sin revisión dedicada (D5-A).
- **Arreglo de documentación (D12-A).** `_support/docs/architecture/deployment-manager.md:310` dice que el loader usa `Agent.Info.name`, pero el código resuelve el rol por identificador (`context-loader.js:20`): se corrige esa frase (T5).

## Tareas

### T1 — El cierre comprueba los deltas antes de marcar Completado

- [x] El paso 1 de «Close the change record» conserva la revisión del diff y añade la comprobación por requisito, el caso del mapa de producto y el preámbulo de un spec nuevo; el paso 2 marca Completado solo tras ella; la línea «Abandon» renombra la carpeta de deltas con el prefijo `unapplied-`, y `change-records.md` lo registra.

**Closes:** AC1.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-close/SKILL.md` («Close the change record», pasos 1 y 2, y la línea «Abandon»); `content/skills/flow-plan/references/change-records.md` (inventario de la línea 13).

**Execution:** hilo principal; son pocas frases y el brief sería más largo.

**Test approach:** check — `rg` sobre frases que solo añade el cambio y la prueba de tamaño de skills.

**Changes:** según el Diseño. El archivo tiene 15 083 bytes de 16 000: redactar breve (borrador del revisor: unos 580 bytes). El paso 2 usa el orden «… to Completed only after …» para que coincida el patrón.

**Verification:** `rg -n "same text" content/skills/flow-close/SKILL.md` muestra la comprobación en el paso 1, `rg -n "unapplied-" content/skills/flow-close/SKILL.md content/skills/flow-plan/references/change-records.md` muestra la línea «Abandon» y el inventario, y `rg -n "Completed.*(after|only)" content/skills/flow-close/SKILL.md` muestra la condición del paso 2; los tres no devuelven nada en `dccc1663`. `go test ./tests/content/ -run 'Size'` pasa.

### T2 — Prueba de deltas aplicados

- [x] `tests/content/spec_deltas_test.go` falla con un delta archivado sin aplicar, distinto o con un REMOVED presente, ignora los requisitos de un cambio activo y el preámbulo, y sus casos sintéticos lo demuestran.

**Closes:** AC2.

**Depends on:** ninguna.

**Locations:** `tests/content/spec_deltas_test.go` (nuevo); convenciones de `tests/content/skill_size_test.go` (rutas relativas `../../`, mensajes con la ruta del repositorio).

**Execution:** delegada a `hive-write-tests`: unidad con interfaz fijada en el Diseño, sin escrituras compartidas con T1, T3 ni T4.

**Test approach:** tdd — primero esqueletos de las funciones que compilan y devuelven un resultado vacío, para que los casos sintéticos fallen por sus aserciones (no por compilación); luego la lógica.

**Changes:** funciones de lectura de bloques y de comparación; casos sintéticos: requisito ausente, texto distinto, diferencia solo de espacios (pasa), REMOVED con encabezado presente y REMOVED ausente (pasa), MODIFIED posterior que reemplaza un ADDED, requisito ignorado por un cambio activo, carpeta `archive/` que no cuenta como cambio activo, preámbulo antes de la primera sección y último requisito al final del archivo; prueba sobre `_support/openspec`.

**Verification:** `go test ./tests/content/ -run 'SpecDelta' -v`: los casos sintéticos pasan, y antes de la lógica fallan por sus aserciones. La prueba sobre el repositorio, en una copia temporal sin la carpeta de este cambio, falla nombrando los 6 requisitos del escaneo (5 de `context-delivery`, 1 de `walkthrough-viewer`).

### T3 — Deltas que reconcilian la base

- [x] La carpeta de este cambio contiene `specs/context-delivery/spec.md` (preámbulo y ADDED de cinco requisitos reconciliados) y `specs/walkthrough-viewer/spec.md` (MODIFIED «Current source classification» con el texto vigente).

**Closes:** AC3.

**Depends on:** T2 (la prueba es la verificación).

**Locations:** `_support/openspec/changes/gh-117-verify-applied-deltas/specs/`; textos de origen en `_support/openspec/changes/archive/2026-10-15-flash-class-guidance/specs/context-delivery/spec.md` y `_support/openspec/specs/walkthrough-viewer/spec.md`.

**Execution:** hilo principal; los textos reconciliados vienen de la investigación y deben pasar por la lectura del orquestador.

**Test approach:** check — la prueba de T2.

**Changes:** escribir los dos archivos de delta según el Diseño.

**Verification:** con la carpeta presente, `go test ./tests/content/ -run 'SpecDelta'` pasa en el árbol de trabajo (los 6 requisitos quedan ignorados). En una copia temporal del repositorio, aplicar los deltas a `specs/` como dice `flow-close`, mover la carpeta a `changes/archive/2026-10-24-gh-117-verify-applied-deltas/` y correr la misma prueba: pasa sin ignorar nada. La comprobación definitiva corre en el PR de cierre.

### T4 — El sitio y el changelog describen la comprobación

- [x] `flows/close.md` dice que el cierre comprueba que cada cambio de requisito quedó aplicado antes de marcarlo completado; `CHANGELOG.md` registra el cambio en `[Unreleased]` → `Changed` con referencia a #117.

**Closes:** AC4.

**Depends on:** T1.

**Locations:** `site/src/content/docs/flows/close.md:44`; `CHANGELOG.md`.

**Execution:** hilo principal; depende del texto exacto de T1.

**Test approach:** check — `rg` y el build del sitio.

**Changes:** una frase en el paso 2 de la página; una entrada de changelog.

**Verification:** `rg -n "matches|confirms" site/src/content/docs/flows/close.md` muestra la frase nueva en el paso 2 y no devuelve nada en `dccc1663`; el build del sitio termina sin errores.

### T5 — La documentación del loader resuelve el rol por identificador

- [x] `_support/docs/architecture/deployment-manager.md` describe que el loader resuelve el rol por el identificador nativo (`data.id`, luego `id`, luego el agente del evento), no por `Agent.Info.name`.

**Closes:** AC5.

**Depends on:** ninguna.

**Locations:** `_support/docs/architecture/deployment-manager.md:310`; `integrations/opencode/context-loader.js:20`.

**Execution:** hilo principal; una frase.

**Test approach:** check — `rg`.

**Changes:** reemplazar la mención de `Agent.Info.name` por la resolución por identificador.

**Verification:** `rg -n "Info.name" _support/docs/architecture/deployment-manager.md` no devuelve nada (en `dccc1663` devuelve la línea 310), y `rg -n "data.id" _support/docs/architecture/deployment-manager.md` muestra la frase nueva.

## Verificación y revisión humana

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | Tras cada tarea, en la rama de trabajo | `hive-verify-task` con contexto limpio: `AC<n>: met` | Parte de la implementación |
| Gates compartidos | Candidato final, local | `go vet ./...` sin salida; `go test ./...` con todos los paquetes `ok`; build del sitio completo | Parte de la implementación |
| Validación humana | Antes del push (`interactive`) | El usuario lee el diff y los deltas reconciliados | D4-A |
| Prueba de deltas en el cierre | PR de cierre, que añade la prueba (D11-A), aplica los deltas y archiva | `go test ./tests/content/ -run 'SpecDelta'` y `go test ./tests/content/` en verde sobre el candidato del PR de cierre, antes de fusionar | Ruta de cierre registrada y D11-A |

Sin revisión de código dedicada (D5-A), sin revisión de seguridad y sin recorrido de UI.

## Estado de la revisión y progreso

- Plan escrito el 2026-10-24. Revisión: dos `hive-review-plan` en paralelo sobre la primera versión: guía y sitio (69 s, unos 91 mil tokens) y prueba y specs (115 s, unos 100 mil tokens).
- Aceptado tal como lo propusieron los revisores, sin re-revisión: conservar la revisión del diff; patrones `rg` que no coinciden en la base (T1, T4); redacción del mapa de producto y de REMOVED por encabezado; esqueletos para un RED genuino (N2) y la verificación dependiente de T3 movida a T3; casos sintéticos añadidos (N3); preámbulo con `Purpose` en el delta (N4); desempate por nombre de carpeta con el nombre en el mensaje (N1).
- Corrección propia de B2: el abandono renombra la carpeta de deltas con el prefijo `unapplied-`. Re-revisada por el mismo revisor de guía (30 s, unos 97 mil tokens): sin bloqueantes. Se aplicaron tal como los propuso el renombre de `product/` y el registro en `change-records.md`; el borrador del revisor estima que cabe en el límite de tamaño.
- Decisiones del usuario: D11-A (la prueba se entrega con el cierre; B1 resuelto) y D12-A (arreglo de `deployment-manager.md:310`, T5).
- Ruta D13-A (2026-10-24): `flow-build` iniciado, AC1–AC5 aprobados. Durante el build, el usuario preguntó si AC1 debía ir en `flow-build`; tras la explicación (D14) y una propuesta de aplicar los deltas en el build (D15), decidió dejar el plan original.
- T1 `[x]` (`7f2fec47`; tras su veredicto se añadió a `change-records.md` el preámbulo opcional de un delta que crea un spec, nota del verificador), T5 `[x]` (`6faf6a6f`, con una línea reajustada), T4 `[x]` (`a7b911cc`), T2 `[x]` (sin commit por D11-A; RED con esqueletos: 10 casos fallaron por aserción; 17 casos verdes; sobre el repositorio sin deltas activos falló nombrando los 6 requisitos), T3 `[x]` (deltas en la carpeta del cambio; prueba verde en el árbol y en la simulación del cierre con `sim117.sh` en el scratch). Cada tarea con veredicto `met` de un `hive-verify-task` con contexto limpio. Modelo del implementador: Opus 5.5 (observado); verificadores y `hive-write-tests`: configurado, no observado.
- Gates compartidos: `go vet ./...` sin salida; `go test ./...` con 20 paquetes `ok` (árbol con la prueba de T2 presente); build del sitio completo (10 páginas) sobre T4. Sin revisión de código dedicada (D5-A).
- Validación del usuario D17-A (2026-10-24), tras preguntar por el costo de la comprobación de T1. Push, [PR #137](https://github.com/JhonHawk/tricell-hive/pull/137) (T1, T4, T5), merge `39cfe446` con la base al día.
- Cierre: rama `chore/close-gh-117-record` desde `39cfe446`; `specs/context-delivery/spec.md` creado desde el delta (preámbulo + requisitos); `walkthrough-viewer` sin edición porque su bloque ya era idéntico al MODIFIED; comprobación del paso 1 por la prueba de T2 sobre el repositorio real tras archivar.
- Descartado: «falta la sección `### Changed` en el changelog»; existe en `CHANGELOG.md:35`.
