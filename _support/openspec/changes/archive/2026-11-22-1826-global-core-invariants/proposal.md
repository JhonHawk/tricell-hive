# global-core-invariants — El núcleo guarda invariantes; el procedimiento va a su hoja

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #592, 4a2d1c5d |
| Tracker | ninguno |
| Git | tricell-hive · base `development` (`c61c10e3`) · rama propuesta `refactor/global-core-invariants` |
| Playbook | refactor |

## Decisions

2026-11-22
- **D27 A** — La lista de lo que la limpieza de cierre puede hacer (`global.md:32`) pasa a `close-hygiene` «Clean up», que se vuelve excepción de tamaño (hasta 7.000 B); el núcleo conserva el permiso en una frase y los «nunca» que valen siempre.
- **D31 A** **Replaces:** D27 — ¿Dónde va la lista de la limpieza de cierre? A una hoja nueva `hive-core/close/close-cleanup.md`, fuera del paquete de flow-close (que mide 23.189 B de 23.500 permitidos), con su línea en el índice y enlazada desde el paso de limpieza de flow-close; `close-hygiene` no crece. Lo que se queda en el núcleo es lo que fijaba D27, más la regla de detener procesos por ID.
- **D28 A** — Las definiciones de cambio mecánico, cambio pequeño entendido y tarea operativa (`global.md:50-51`) pasan a una hoja nueva `hive-core/plan/proportionality.md` con su línea en el índice; el núcleo conserva el enrutado y los archivos que citan esas definiciones pasan a citar la hoja.
- **D29 A** — La línea del efecto crítico (`global.md:14`) se queda entera.
- **D30 A** — El procedimiento de un proyecto externo (`global.md:16`) pasa a `scope-changes`; el núcleo conserva el límite y una frase que cita la hoja.
- **supuesto** — `global.md:21` (escrituras en el tracker), `:30` (comandos por ruta absoluta), `:39` (dónde viven los secretos) y `:46` (lectura de `HIVE.local.md`, que trata #590) se quedan: son invariantes que deben actuar antes de leer cualquier hoja. Alternativa descartada: moverlas. Se revierte en otro cambio.
- **supuesto** — Las definiciones de lectura grande y cambio riesgoso (`global.md:52-53`) se quedan en el núcleo; las citas a «Proportionality» que las usan no cambian.

## Business rules

None: el trabajo no adopta ninguna regla de negocio.

## Request

«la prosa del global de la línea 11 a la 53 no es muy amplia?» y, ante la propuesta de dejar en el núcleo solo el invariante de cada bullet largo, «A».

## Objective

Hoy las líneas 11–53 de `global.md` ocupan 14.633 B, el 52 % del archivo, y el núcleo está en el techo de 28.000 B. Nueve bullets mezclan su invariante con un procedimiento o una enumeración. Al terminar, tres de ellos (cierre, proporcionalidad, proyecto externo) dejan en el núcleo solo lo que debe actuar sin leer otra hoja, y su procedimiento queda en la hoja que se lee en ese momento.

## Scope and acceptance

Dentro: `global.md`, las hojas nuevas `plan/proportionality.md` y `close/close-cleanup.md`, `close/close-hygiene.md` (su cita), `plan/scope-changes.md`, `flow-close` (su paso de limpieza), los archivos de `content/` que citan las definiciones movidas, `tests/content` (frases registradas, enlaces de hojas y lecturas del escenario plan/build), `instruction-resources.md`, `CHANGELOG.md`. Fuera: el texto de las reglas, que no cambia salvo su ubicación y sus enlaces; los bullets de D29 y del primer supuesto.

- AC1 — `global.md` «Proportionality» conserva el enrutado (toda petición de trabajo sin plan entra a `hive-mode`; quién va a `flow-build` y quién a `flow-plan`) y las definiciones de lectura grande y cambio riesgoso, y no define cambio mecánico, cambio pequeño entendido ni tarea operativa · false today because `global.md:50-51` las define · Observed in: `content/guidance/global.md` del candidato
- AC2 — `hive-core/plan/proportionality.md` existe, empieza con «Read», define esas tres cosas con el texto que hoy tiene `global.md`, el índice la lista bajo «Plan», y `flow-build`, `delivery-ask` y el paso de enrutado de `hive-mode` la enlazan donde aplican esas definiciones · false today because el archivo no existe · Observed in: el archivo y `go test ./tests/content/... -count=1`
- AC3 — Ningún archivo de `content/` cita la guía global por la definición de cambio mecánico, cambio pequeño entendido o tarea operativa; los que la usaban citan `proportionality.md` · false today because 16 archivos citan «Proportionality» · Observed in: `rg -n "«Proportionality»" content` cuyas líneas restantes hablan solo de lectura grande, cambio riesgoso o enrutado
- AC4 — `close/close-cleanup.md` enumera lo que la limpieza de cierre puede hacer y el índice la lista bajo «Close»; `global.md` «Preservation» conserva el permiso en una frase que la cita, la regla de detener solo procesos propios por ID y nunca por nombre o patrón, y los «nunca» (rama sin mergear, worktree con cambios, programas compartidos, contenedores ajenos, evidencia única); `close-hygiene` y `flow-close` la citan · false today because la enumeración está en `global.md:32` · Observed in: las secciones del candidato
- AC5 — `scope-changes.md` fija qué hacer cuando el código está en un proyecto externo, y su primera línea y su línea del índice nombran ese momento; `global.md` conserva el límite del proyecto actual y una frase que cita esa hoja; `defect-fix.md:8` la cita · false today because el procedimiento está en `global.md:16` y la primera línea de `scope-changes` no nombra ese momento · Observed in: las dos líneas del candidato
- AC6 — `global.md` mide 26.000 B o menos · false today because mide 28.000 B · Observed in: `wc -c content/guidance/global.md`
- AC7 — **Retired:** ya es cierto en la base; queda como guarda en la verificación de T3 — El contenido real se instala con todos sus enlaces y anclas resueltos y los paquetes de etapa bajo su tope · Observed in: `go test ./tooling/management/ -run 'TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink|TestRealStagePacketsKeepAMarginUnderTheCap|TestRepositoryCatalogueInstructionReferences' -count=1`

## Prerequisites

None: la construcción no necesita nada de fuera del repositorio.

## Verified context

- `awk 'NR>=11 && NR<=53' content/guidance/global.md | wc -c` · 14.633 B · 2026-11-22
- `wc -c content/guidance/global.md` · 28.000 B · 2026-11-22
- `wc -c` de las hojas destino: `close-hygiene` 3.384 B, `scope-changes` 2.146 B (tope 4.000; excepción 7.000) · 2026-11-22
- Paquete de flow-close generado: 23.189 B; tope con margen 23.500 B · `TestRealStagePacketsKeepAMarginUnderTheCap` (medición de esta sesión) · 2026-11-22
- `rg -l "«Proportionality»" content` · 16 archivos (incluido `global.md`) · 2026-11-22
- La spec «Layered global core with owning skills» exige en el núcleo «the close-cleanup authorization»; no nombra las definiciones de proporcionalidad · `context-delivery/spec.md:94` · 2026-11-22
- `go test ./... -count=1` · ok (tras PR #589) · 2026-11-22
- `python3 -I -m unittest discover -s tests/skills -p '*_test.py'` · 412 OK · 2026-11-22

## Effects

None: las tareas no producen efectos en sistemas externos.

## Delivery

Entrega `hold` y revisión Claude Code `/code-review`, las respuestas vigentes de esta sesión. Sin push, PR ni merge hasta que el usuario lo pida.

## Review status and progress

2026-11-22 — relectura: 0 P1, 0 P2; 4 P3 de redacción aplicados (cláusula de procesos sin duplicar, alcance alineado, test de enlace de flow-close, textos de T2 y AC7). `defect-fix.md` mide 3.844 B: T3 comprueba que sigue bajo 4.000.
2026-11-22 — revisión de `hive-review-plan`: 2 P1 (el paquete de flow-close se pasaba del tope; el núcleo perdía «detener procesos por ID»), 5 P2 y 4 P3. D31 resuelve el primero; el resto se aplica en esta versión.
2026-11-22 — plan escrito.
