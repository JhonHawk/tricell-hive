# core-cleanup-batches — Evidencia para relajar reglas, absolutos solo para invariantes, y tres lotes de limpieza de hive-core

| Campo | Valor |
| --- | --- |
| Estado | In build · integrando a `development` por PR (2026-11-22) |
| Tracker | ninguno |
| Git | tricell-hive · base `development` (`8aa361cb`) · `hold`: en el sitio o en un worktree, según pregunte el build; rama `refactor/core-cleanup-batches` solo si luego se entrega |
| Playbook | refactor |

## Decisions

2026-11-22
- **D9 A** — ¿Qué hacer con los +908 B de lectura obligatoria del lote 1? Decidir con la cifra final, al cerrar T6.
- **D10 A** — ¿Se acepta que la lectura obligatoria de plan y build crezca +1.665 B (+0,83 %) contando igual que en la base (195.800 B en la prueba, que ya no cuenta `project-verification` en todo ciclo)? Sí: es el costo de que las reglas del lote 3 lleguen a sus flujos.
- **D3 B** — ¿Cómo seguimos tras la auditoría? Primero decidir los cambios a `AGENTS.md`, y después resolver los tres lotes de una vez.
- **D4 A** — ¿Qué evidencia basta para relajar o quitar una regla? La decisión del usuario más una comparación de sesiones reales con `--audit-session` antes y después, o la razón que la propia regla declara verificada como vencida en documentación o versión actual; quitar una repetición, un puntero colgando o una referencia circular no es relajar.
- **D5 A** — ¿Regla de redacción para «never»/«always»? Sí: solo para invariantes (autorización, secretos, efectos irreversibles, preservación); en lo demás, alcance y razón; aplica al texto que se escribe o edita, sin reescritura masiva de los 556 existentes.
- **D7 A** — ¿Qué lista decide cuándo se recomienda la revisión de seguridad? La amplia: `security-boundaries.md` lista los ocho temas y `hive-review-security` apunta a ella.
- **Entrega** — `hold`: sin commit ni publicación; el árbol de trabajo verificado es el resultado. **Revisión**: nativa, Claude Code `/code-review`.
- **Entrega (cambio)** — El usuario pidió «integremos a development»: PR y merge. Dos commits (guía, pruebas y documentación; carpeta de cambio) en vez de uno por lote, porque los lotes editan los mismos archivos y no se separan sin partir hunks.
- **supuesto** — La casa de «captura reemplazada» es `project-layout/references/layout.md#retain-evidence`; descartada: `support-material.md`.
- **supuesto** — Las condiciones del `--force-with-lease` desatendido viven en `git-commit.md`; descartada: moverlas a `unattended-delegation`.
- **supuesto** — `ticket-state.md` incluye la excepción «salvo que el usuario nombre otro tracker o repositorio» que ya aplica `create-issue`, con delta en la spec; descartada: quitarla.
- **supuesto** — El registro `Playbook` del checkpoint lo leen `flow-build` y `flow-verify` al entrar; descartada: borrarlo de la spec.
- **supuesto** — Los «registros de cierre pendiente» (`close-modes.md:10,17`, `branch-cleanup.md:7,15`, `close-hygiene.md:15`) y el «close record» de `time.md:19` son el registro del resultado que escribe `flow-close` paso 7, que pasa a listar los pasos del cierre que esperan respuesta; descartada: `pending/<change-id>`, que guarda trabajo dejado abierto a propósito, no limpieza pendiente (revisión del plan, H9).
- **supuesto** — Una hoja que pase de 4.000 B (`leafSizeLimit`) al recibir texto lleva la parte nueva a una referencia propia `hive-core/<dominio>/<hoja>/<nombre>.md`; no se pide excepción de tamaño.
- **supuesto** — Si la unión plan/build de `reading_footprint_test.go` crece al cerrar T3 o T6, el build se detiene y lo pregunta; no agrega una allowance por su cuenta.
- **supuesto** — El criterio de paquetes en vigor (AC9) no exige que bajen los 40.403 B de relectura (revisión del plan, H1): T5 solo quita hojas de un solo paquete; la relectura entre skills queda fuera de alcance, como ya dice «Scope and acceptance».

## Business rules

None: el trabajo no adopta ninguna regla de negocio.

## Request

«Considerando que ahora tenemos una organización más concisa sobre nuestras carpetas, veo que aun hay cosas que podrian considerar "repetidas"»; «y aunque nuestro ejemplo va vs git, pasa en otros principios y hojas no?»; «también analicemos qué recomiendan los diferentes proveedores… talvez estamos en primer lugar abusando de la densidad de contenido, y esto claro nos puede llevar a corregir nuestro AGENTS.MD»; «B, y si me interesa resolver los tres lotes de una». Fuera de alcance: una skill `/audit-harness` nueva, una línea base tipo ratchet, el modo incremental de `harness-audit`, la reescritura masiva de los «never» existentes y las pruebas con modelos (en pausa).

## Objective

Hoy, quitar una regla exige evidencia de sesiones reales sin esa regla, que no puede existir mientras la regla está y las pruebas con modelos están en pausa, así que la guía solo crece; y nada orienta el uso de «never», que los modelos actuales leen como absoluto. Además, el inventario de la sesión encontró unos 88 hallazgos en `hive-core` y sus consumidores: citas a pasos o secciones que no existen, definiciones circulares, reglas repetidas a medias, paquetes que precargan reglas condicionales y reglas que un flujo necesita pero no enlaza. Al terminar, `AGENTS.md` y su spec fijan la evidencia de D4 y la redacción de D5, una prueba detecta las citas a secciones o pasos inexistentes, y cada hallazgo de `design.md` queda resuelto en su casa con punteros en los demás lugares.

## Scope and acceptance

Dentro: `AGENTS.md` «Measurement» y «Guidance structure»; las specs `model-qualification` y `workflow-lifecycle` (deltas); una prueba nueva en `tests/content`; los hallazgos de los lotes 1, 2a, 2b y 3 de `design.md` en `content/`; los tests de contenido y de `tooling/management` que dependan de frases, paquetes o rutas movidas; `_support/docs/architecture/instruction-resources.md`, `_support/docs/harness-engineering/rule-enforcers.md`, las páginas de `site/` que describan un comportamiento que cambia, `README.md` y `llms.txt` si alguno lo menciona, y `CHANGELOG.md`. Fuera: lo que lista «Request», los hallazgos descartados de `design.md` y los ~40 KB de relectura de paquetes entre skills que se cargan en la misma sesión cuando cada skill necesita el suyo para funcionar sola.

- AC1 — `AGENTS.md` «Measurement» y la spec «Recurrence evidence from floor candidates» aceptan, junto con la decisión del usuario, una comparación de sesiones reales con `--audit-session` antes y después o la razón declarada de la regla verificada como vencida, y dicen que quitar una repetición, un puntero colgando o una referencia circular no es relajar · false today because `AGENTS.md:86` y la spec solo aceptan «real-session evidence that the floor's candidates behave correctly without it» · Observed in: `AGENTS.md` y `_support/openspec/changes/core-cleanup-batches/specs/model-qualification/spec.md` del candidato
- AC2 — `AGENTS.md` «Guidance structure» dice que «never» y «always» se usan solo para invariantes (autorización, secretos, efectos irreversibles, preservación) y que las demás reglas dan alcance y razón, en el texto que se escribe o edita · false today because `rg -n -i 'only for invariants' AGENTS.md` no devuelve nada · Observed in: `AGENTS.md` del candidato
- AC3 — `go test ./tests/content/...` falla, nombrando archivo, línea y destino, cuando una «sección» citada junto a un enlace a un `.md`, o un «step N» citado junto a un enlace a un `SKILL.md`, no existe en el archivo enlazado (para la sección, un título `#` o una etiqueta en negrita al inicio de un ítem o párrafo, `**Nombre**` seguido de `.`, `:` o `,`; para el paso, un ítem `N.`), y pasa en el candidato · false today because la suite pasa (2026-11-22) aunque `content/skills/hive-core/plan/scope-changes.md:12` cita flow-plan «Revise a plan in progress», que `flow-plan/SKILL.md` no tiene · Observed in: la salida de la prueba con la base más la prueba (falla nombrando `scope-changes.md:12`) y con el candidato (pasa)
- AC4 — Cada fila del lote 1 de `design.md` queda resuelta: la cita apunta a un destino que dice lo que se le atribuye, cada par circular tiene un solo lugar que define, y cada conflicto queda en la redacción de su casa · false today because `rg -n 'Tests section' content` devuelve 3 líneas, `deployment-checks.md:11` cita «git-workflow step 5» (hooks) para leer CI, y `close-report.md:49` y `review-findings.md:5` se remiten P1 mutuamente · Observed in: los archivos del candidato, fila por fila, con `rg` de cada frase citada
- AC5 — Cada fila del lote 2a queda como enlace con ancla a su casa, sin pasos, condiciones ni excepciones de ella fuera de la casa · false today because, por ejemplo, `git-workflow/SKILL.md:24` repite la espera `gh run watch --exit-status` de `merge-gate.md:5` y `hotfix.md:8` repite `defect-fix.md` · Observed in: los archivos del candidato, fila por fila
- AC6 — `flow-verify` no precarga `project-verification.md`, `flow-build` no precarga `stored-data.md` ni `security-boundaries.md`, el paquete de `flow-close` trae `business-ticket.md`, `reviewer-brief.md` no tiene las reglas de marcos de create-design, y los bytes de reglas que se releen en un ciclo hive-mode → flow-plan → flow-build → flow-verify → git-workflow bajan de 40.403 B · false today because esos `assets/stage.json` las listan y el recuento da 40.403 B (2026-11-22) · Observed in: los `stage.json` del candidato y el recuento **Retired:** los cambios de T5 tocan hojas de un solo paquete y no mueven ese recuento (revisión del plan, H1)
- AC8 — `flow-verify` no precarga `project-verification.md`, `flow-build` no precarga `stored-data.md` ni `security-boundaries.md`, el paquete de `flow-close` trae `business-ticket.md`, `reviewer-brief.md` no tiene el brief de marcos dibujados de create-design, y la línea «it holds the leaves» de cada `SKILL.md` cuyo paquete cambia lista exactamente las hojas de su `stage.json` · false today because `content/skills/flow-verify/assets/stage.json` lista `project-verification.md`, `flow-build/assets/stage.json` lista `stored-data.md` y `security-boundaries.md`, y `reviewer-brief.md:11` tiene el brief de marcos · Observed in: los `stage.json`, las líneas 8 de `flow-build/SKILL.md`, `flow-close/SKILL.md` y la de `flow-verify/SKILL.md`, y `reviewer-brief.md` del candidato **Replaces:** AC6 **Retired:** el paquete de `flow-close` pasaría de su tope de 24.000 B con `business-ticket.md` (27.089 B); `business-rules.md`, que el paquete ya trae, enlaza `business-ticket.md` en el momento de la escritura 3
- AC9 — `flow-verify` no precarga `project-verification.md` y solo lo lee para trabajo E2E o un cambio de comando de verificación, `flow-build` no precarga `stored-data.md` ni `security-boundaries.md`, `reviewer-brief.md` no tiene el brief de marcos dibujados de create-design, y la línea «it holds the leaves» de cada `SKILL.md` con paquete lista exactamente su `stage.json`, comprobado por una prueba · false today because `flow-verify/assets/stage.json` lista `project-verification.md`, `flow-build/assets/stage.json` lista `stored-data.md` y `security-boundaries.md`, y `reviewer-brief.md:11` tiene el brief de marcos · Observed in: los `stage.json`, las líneas «it holds» y `reviewer-brief.md` del candidato, y `go test ./tests/content/ -run TestPacketHoldsLinesMatchManifests` **Replaces:** AC8
- AC7 — Cada fila del lote 3 queda enlazada en el paso del flujo donde aplica, y `security-boundaries.md` lista los ocho temas de D7 con `hive-review-security` apuntando a esa lista · false today because `rg -c 'git-workflow' content/skills/flow-verify/SKILL.md content/skills/flow-deploy/SKILL.md` no cuenta nada y `security-boundaries.md:3` lista cuatro temas · Observed in: los archivos del candidato, fila por fila

## Prerequisites

None: la construcción no necesita nada de fuera del repositorio.

## Verified context

- Inventario de la sesión del 2026-11-22: siete agentes de investigación (git; build y review; close, tracker y business; messages, plan y delegation; ui y local; comparación pstack y Hive anterior; rúbrica contra mecanismos existentes) y dos de proveedores (Anthropic y OpenAI; Google, Cursor, xAI, OpenCode, Pi y estudios); verificados en el hilo principal G-C1, G-C7, G-C8, B-1, B-2, B-6, B-8, LO-1, PL-1, PK y los descartes CL-6 y UI-5. Decisión en `engram:tricell-hive/decision/relax-evidence-and-absolutes`.
- Prototipo de la prueba de citas (script desechable en el temporal de la sesión), base `8aa361cb`: 1 cita colgante (`scope-changes.md:12`) aceptando etiquetas en negrita; 7 sin aceptarlas. 2026-11-22.
- `go test ./tests/content/...` · ok, 2,3 s · 2026-11-22
- `go test ./tooling/management/ -run 'TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink|TestRealStagePacketsKeepAMarginUnderTheCap' -count=1` · ok, 25,3 s · 2026-11-22
- `go vet ./... && go test ./...` · ok, 23 paquetes, 2 min 57 s · 2026-11-22
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` · ok, 412 tests, 1 min 40 s · 2026-11-22
- `go test ./tests/content/ -run TestCitedSectionsAndStepsExist -count=1` · not measured: T2 crea la prueba; su forma más cercana, `go test ./tests/content/...`, pasa (arriba)
- `python3 -I content/skills/flow-plan/scripts/check_change.py _support/openspec/changes/core-cleanup-batches` · exit 0 · 2026-11-22
- `rg -n 'Tests section' content` · 3 líneas (`test-audit/SKILL.md:3,47`, `flow-plan/assets/change-folder/tasks.md:19`) · 2026-11-22
- `rg -n -i 'audit-session.{0,80}before and after|before and after.{0,80}audit-session' AGENTS.md` · sin resultados (las 2 menciones actuales de `audit-session` son de fixtures) · 2026-11-22
- `rg -n -i 'only for invariants' AGENTS.md` · sin resultados · 2026-11-22
- `rg -n 'git-workflow step 5' content` · 1 línea, `flow-deploy/references/deployment-checks.md:11` · 2026-11-22
- `rg -c 'git-workflow' content/skills/flow-verify/SKILL.md content/skills/flow-deploy/SKILL.md` · 0 en ambos · 2026-11-22
- `rg -n -i 'authentication' content/skills/hive-core/build/security-boundaries.md` · sin resultados · 2026-11-22
- `go test ./tests/content/ -run TestRequiredProcedureReadingFootprint -count=1 -v` · pasa; unión plan/build 200.373 B (+8,61 % sobre `2bb96478`), la referencia «igual o menor» de T3 y T6 · 2026-11-22
- `rg -n 'flow-verify/references/host-session-identity' content site README.md llms.txt _support/docs` · 1 línea, `site/src/content/docs/viewer.md:68` · 2026-11-22
- `rg -n -i -l 'process execution' content site README.md llms.txt _support/docs` · 7 archivos: `global.md`, `security-boundaries.md`, `flow-build/SKILL.md`, `hive-review-security.md`, `hive-mode/playbooks/infra-ci.md`, `site/src/content/docs/flows/security.md`, `_support/docs/architecture/instruction-resources.md` · 2026-11-22
- Tamaños: `build/tests.md` 3.705 B, `plan/checkpoint.md` 3.230 B, `git/merge-gate.md` 4.489 B, `content/guidance/global.md` 25.856 B · 2026-11-22

## Effects

None: las tareas no toman ningún efecto fuera del repositorio; el push, la PR y el merge los rige el modo de entrega.

## Delivery

`hold` (respuesta del usuario, 2026-11-22) hasta la verificación; después el usuario pidió «integremos a development»: rama `refactor/core-cleanup-batches`, PR a `development`, espera de CI y merge. Revisión de código: nativa, Claude Code `/code-review`, sobre el diff del candidato antes de entregarlo. Este modo no autoriza commit, push, PR, merge ni `hive update`.

## Review status and progress

2026-11-22 — Hallazgos de cierre H15–H20 (P3) corregidos a pedido del usuario: `flow-deploy` «Completion» nombra el ref desplegado; `flow-close` paso 3 apunta a `ticket-state` para el tiempo; `hive-review-security` lee la hoja siempre; `flow-build` escribe el `Playbook` sin carpeta de cambio; `reviewer-brief` limita los estados a pantallas capturadas; `batch-archive` nombra su propio borrado suave. Revisión limitada (`hive-review-code`): H17–H20 resueltos; la regresión P2 de H16 (worklog de tarea operativa) y el P3 de H15 corregidos después. `tests/content` y las pruebas de instalación pasan; huella 196.065 B.
2026-11-22 — Revisión en el build (T5): AC8 retirado y reemplazado por AC9 sin el paquete de `flow-close` (tope de 24.000 B); `reading_footprint_test.go` pasa `project-verification` a la ruta E2E, porque flow-verify ya no lo lee en todo ciclo. No cambia lo que el usuario aprobó salvo esa cláusula.
2026-11-22 — Plan escrito. Revisión `hive-review-plan`: 4 P1, 5 P2 y 5 P3; aplicadas todas. AC6 sin el recuento de bytes (H1); listas «it holds the leaves» en T5 (H2); `host-session-identity.md` a `hive-core/ui/viewer/` como referencia de `viewer.md` (H3); `reviewer-brief.md` se parte en tres destinos (H4); D7 sin condición en T6 y búsqueda de T7 sobre `content/` (H5); PL-1 se arregla en T2 (H6); topes de hoja y huella con supuestos (H7, H8); CL-1 decidido (H9); líneas de verificación agregadas (H10); `proposal.md:39` de la plantilla al lote 1 y regla de negrita fijada (H11); control de reescrituras de D5 (H12); medición de `global.md` para una PR futura (H13); T7 cierra solo AC7 y B-2 con líneas 3 y 47 (H14). Relectura del mismo revisor: H1–H14 resueltos; tres P3 nuevos (ruta de T6, recuperación de archivos nuevos y movidos, dónde se trabaja bajo `hold`) corregidos. `check_change.py` sale 0.
