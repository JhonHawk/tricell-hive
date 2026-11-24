# assumed-decisions-ticket-blocks — Decisiones supuestas en lista y evaluación de tickets por bloques

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #528, merge a849a2e3 |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `a0e9b280` · rama propuesta `feat/assumed-decisions-ticket-blocks` |
| Playbook | ninguno (`/flow-plan` tecleado por el usuario) |

## Decisions

2026-11-19
- **Decisión de alcance (A)** — ¿Cómo reducimos rondas sin dejar de informar? Las decisiones con una recomendada clara y reversible se toman y se listan («lo que no estaba definido y decidí»); solo se pregunta lo que cumple un criterio a–f; la evaluación de un ticket se escribe por bloques. Descartado: mover los criterios al plan (rompe «el plan decide el cómo, nunca qué entra») y el modelo de pstack de avanzar sin preguntar.
- **D1 A** — ¿Cómo se produce el veredicto «listo / no listo»? Por juicio sobre los 8 puntos de la lista de `ticket-types.md`, escrito con los números que fallan; el auditor verifica la forma del cierre. `check_issue.py` no se usa sobre tickets ajenos porque solo valida la estructura de nuestros borradores.
- **D2 A** — ¿Medimos cuántas tarjetas terminan en la ★? Sí, como dato informativo del auditor, sin pass/fail.
- **Supuesto** — Los criterios a–f y la lista viven en `hive-rule-questions` «What to ask»; `global.md` guarda el invariante mínimo y `plan-handoff.md` apunta ahí. Descartado: repetirlos en `hive-mode`.
- **Supuesto (reemplazado en la revisión del plan, H2)** — La prueba de `assumption` de las reglas de negocio se alinea con el criterio c: se supone la regla que el ticket, las specs o las palabras del usuario implican. Descartado: la prueba estrecha actual, que contradice c.
- **Supuesto (reemplaza al anterior)** — Una regla de negocio se supone solo cuando no se cumple ninguno de a–f; así permisos, datos, contrato y alcance (a, d, f) siguen yendo a pregunta. Descartado: dejar solo c, que aflojaría la prueba actual (revisión del plan, H2).
- **Supuesto** — La respuesta del usuario a la tarjeta o al bloque de siguiente paso acepta cada supuesto que no corrige, y el mensaje lo dice en una línea. Un supuesto no es una decisión preguntada: no autoriza ningún efecto, y una regla de negocio aceptada así sigue como `assumption` en el registro hasta que el usuario la contesta de forma explícita (revisión, H5).
- **Supuesto** — Como máximo 5 supuestos en la lista; el resto se aprueba en una pregunta de la misma ronda, «¿Aceptas los supuestos 6 a n?», con aceptar y revisar uno por uno como opciones.
- **Supuesto** — a–f gobiernan solo las decisiones abiertas del qué y del cómo (investigación, plan y hechos tardíos del build). Conservan sus preguntas las rondas que otra regla exige: efectos y efectos críticos, preguntas de sesión, parada de validación, rondas de hallazgos, aprobación y reglas propuestas en Figma, rondas de `create-issue` y `create-prd`, la tarjeta sin respuesta, la pregunta por un proyecto externo y la decisión de infraestructura de pruebas (revisión, H6).
- **Supuesto** — Umbrales observables (revisión, H12): una ganadora es «clara» cuando la razón de la recomendada es una consecuencia que las otras opciones no tienen y no se cumple ninguno de a–f; las opciones están «casi parejas» (e) cuando no se puede nombrar una consecuencia que las separe; el ticket, las specs o el usuario «implican» una conducta cuando su texto nombra esa conducta o su caso contrario.
- **Supuesto** — Las condiciones de «no listo» son los 8 puntos de la lista de `ticket-types.md`, no un subconjunto (revisión, H15).
- **Supuesto** — La evaluación de un ticket recupera la forma del 3 de noviembre (`c423e5ef`): bloques fijos, criterio de «no listo», orden explícito y «el plan decide el cómo, nunca qué entra».
- **Supuesto** — Los supuestos se escriben en `## Decisions` del `proposal.md` con la palabra «Supuesto», como en esta carpeta.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «el 80% de preguntas o decisiones termino escogiendo la recomendada […] ese 80% pueden anunciarse talvez en forma de listado "Lo que no está definido y tomé decisión es - 1, 2 - etc, ¿está bien o ajustamos algo?" Y el otro tanto puede ser por varios motivos a consultarse con el usuario, por ejemplo: preguntas que expanden el alcance, A puede ser la opción recomendada pero conviene hacer B a largo plazo, u otros criterios» — usuario, 2026-11-19.
> «algo que sí me gustó es cómo resumen actualmente, eso lo considero un bloque, y abajo poner otros bloques, como veredicto, el tema de los criterios, etc.»

Origen: sesión Codex en project-g `01a1f5d4` (PRK-20, `gpt-6.2-sol`, 2026-11-19 22:04Z). La evaluación del ticket terminó sin tarjeta, sin bloque de siguiente paso y sin línea de espera; el usuario no se enteró de que los criterios estaban abiertos. Watch: `engram:tricell-hive/watch/ticket-assessment-without-ending`. Decisión: `engram:tricell-hive/decision/assumed-vs-asked-decisions`.

## Objective

Hoy solo se suponen las «decisiones menores» (las que no cambian alcance, contrato, datos, costo ni reversibilidad con ninguna opción), así que casi todo va a una ronda, y la evaluación de un ticket tiene forma libre. Al terminar, el agente decide y lista lo que tiene una ganadora clara y reversible, pregunta solo lo que cumple a–f, y la evaluación de un ticket muestra en bloques el veredicto, los criterios, lo decidido y lo preguntado, en una sola ronda.

## Scope and acceptance

Dentro: `hive-rule-questions`, `global.md` (una línea), `plan-handoff.md`, `hive-rule-business-record` y `hive-rule-business-rules`, la plantilla `proposal.md`, `hive-mode`, las páginas `site/` de hive-mode y plan, la rúbrica, la spec `workflow-lifecycle` y el auditor (`tests/pilot`). Fuera: los efectos críticos y la ronda de efectos, que conservan su propia pregunta; la ronda de entrega y revisión (preguntas de sesión); cambiar `check_issue.py`.

Criterios a–f (van a pregunta si cumple al menos uno): a) alcance: agrega o quita comportamiento, toca otro módulo o parte el trabajo frente a lo que dicen el ticket o el pedido; b) ahora contra después: la recomendada cuesta menos hoy y otra evita rehacer o migrar después; c) regla de negocio o preferencia que ni el ticket, ni las specs, ni las palabras del usuario implican; d) difícil de revertir: contrato público, datos guardados, permisos o seguridad, dinero o producción; e) recomendación débil: depende de un dato sin verificar o las opciones quedan casi parejas; f) contradice al usuario, al ticket o a las specs.

- AC1 — `go run ./tests/pilot --audit-session <rollout 01a1f5d4> --host codex` reporta `ticket_assessment_ending` = `fail` con la línea del mensaje de las 22:04:22Z (evaluación de PRK-20 que termina sin tarjeta, sin bloque ▶ y sin línea de espera) · false today because el criterio no existe y la salida no lo lista · Observed in: la salida del auditor sobre la transcripción real `~/.codex/sessions/2026/11/19/rollout-2026-11-19T16-01-30-01a1f5d4-3ec7-7631-97d8-ba540210ac98.jsonl`
- AC2 — El mismo criterio da `pass` cuando la evaluación de un ticket termina con una tarjeta, con el bloque de siguiente paso o con el bloque de respuestas «▶ ¿Cómo seguimos?»; `review` cuando termina con una pregunta abierta en texto (paso 3 de `hive-mode`); y `not_observed` sin evaluación de ticket · false today because el criterio no existe · Observed in: `go test ./tests/pilot` con fixtures en `tests/fixtures/sessions/ticket_assessment_ending/` y la salida sobre la sesión Claude `bf678880`
- AC3 — `--audit-session` informa, fuera de la lista de criterios y sin estado pass/fail, cuántas tarjetas se contestaron con una opción y cuántas eligieron la recomendada (marcada `(★)`, «(Recommended)» o «(recomendado)»); una tarjeta vacía o una respuesta de texto libre no cuenta como contestada con opción · false today because la salida no trae ese dato · Observed in: la salida del auditor sobre esta sesión Claude `7ee42122` (D1 A, D2 A, Entrega B, Revisión A: 4 contestadas, 3 recomendadas) y sobre `01a1f5d4`
- AC4 — `hive-rule-questions` «What to ask» define los criterios a–f con sus umbrales, a qué rondas no se aplican, la lista de supuestos (como máximo 5 líneas: decisión · lo descartado · cómo se revierte) y que la respuesta acepta cada supuesto que no corrige sin volverlo decisión; `rg -n 'minor call' content site` no devuelve nada · false today because devuelve 4 líneas en `content` y la hoja solo admite «minor calls» · Observed in: la hoja instalada tras `hive update` y la primera sesión real posterior que decide algo (`--audit-session` y lectura del mensaje)
- AC5 — La prueba de `assumption` de `hive-rule-business-record` admite una regla solo cuando no se cumple ninguno de a–f, remitiendo a `hive-rule-questions`, y `flow-figma-design` y su página aplican la misma prueba · false today because la línea 18 repite su propia lista y `site/src/content/docs/flows/figma-design.md:31` la copia · Observed in: las hojas instaladas y la primera sesión real posterior que supone una regla
- AC6 — `hive-mode` «Assess a chosen ticket» fija los bloques en orden (Resumen · Veredicto con los números que fallan · Criterios propuestos · Lo que decidí · Te pregunto · Siguiente paso), las condiciones de «no listo», que la tarjeta lleva solo lo que cumple a–f, que el bloque de siguiente paso va en el mismo mensaje cuando no hay nada que preguntar, y que el plan nunca decide qué entra · false today because la sección solo pide «añadir al reporte» las cuatro dimensiones · Observed in: la primera evaluación real de un ticket tras el despliegue, que termina en como máximo una ronda con los bloques

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `content/skills/hive-rule-questions/SKILL.md:11` define «minor call»; `content/guidance/global.md:131` la restata; `content/skills/flow-plan/references/plan-handoff.md:11` la cita; `content/skills/hive-rule-business-record/SKILL.md:18` aplica la misma prueba estrecha a las reglas de negocio (leído 2026-11-19, `a0e9b280`).
- `content/skills/hive-mode/SKILL.md:36-38` «Assess a chosen ticket»; la forma anterior está en `c423e5ef:content/skills/flow-research/SKILL.md:26-35` (cuatro líneas con etiqueta, «no listo» explícito, orden rondas → respuestas → ruta, «Plan settles how … never what enters it»).
- `content/skills/create-issue/scripts/check_issue.py` valida encabezados de un borrador en Markdown; sobre un ticket ajeno solo daría hallazgos de estructura (D1).
- Auditor: `tests/pilot/session_audit_routes.go` (`flowEnteredWithoutRoute`, `nextStepBlock` en `session_audit_prose.go:30`), fixtures por criterio en `tests/fixtures/sessions/<criterio>/`.
- Comparación de tres vías (pstack, Hive anterior, Hive actual) hecha en la sesión Claude `7ee42122`, regla local `decisions.local.md` del 2026-11-19.

## Effects

None: the tasks take no effect on an external system.

## Delivery

`automatic` (2026-11-19): revisión, push, PR y merge a `development` sin detenerse; la aprobación pendiente de un revisor humano detiene el merge. Revisión de código: nativa, `/code-review` de Claude Code. Sin revisión de seguridad (no toca un límite de seguridad). No autoriza promover a `master`, publicar una versión ni actualizar la instalación local fuera de lo que `AGENTS.md` «Refreshing a local installation» ya autoriza en esta copia.

## Review status and progress

2026-11-19 — Plan escrito. Revisión `hive-review-plan` (2 P1, 10 P2, 6 P3), todo aplicado: cuarta línea de «minor call» en `hive-rule-ui-example` (T4); la prueba de `assumption` exige que no se cumpla ninguno de a–f, no solo c; delta MODIFIED de «Business rules that appear while drawing» y la página de Figma (T5); invariante de `global.md` sin lista parcial; un supuesto no es una decisión ni una autorización; a–f limitados al qué y al cómo, con las rondas exentas nombradas; finales permitidos (pregunta abierta → `review`, bloque de respuestas → `pass`); medición con «(Recommended)», sin tarjetas vacías; más fixtures; `rule-enforcers.md`, `single_home_test.go` y `CHANGELOG.md` (T7, T8); umbrales observables; «no listo» con los 8 puntos de la lista; detección de claves y de `hive-mode`. `check_change.py` sale con 0. Estado: listo para `flow-build`.
2026-11-19 — Build T1–T8 (`f74ff1c`…`6cb6ada`). `/code-review` (4 importantes, 5 menores): corregidos los 4 importantes y 3 menores en `134f423` y `0584c28` (criterio c reformulado como «ninguna conducta nombrada la requiere», para que una regla implícita se pueda suponer; el enfoque del plan y las reglas `proposed` de Figma siguen preguntándose; `approval.md` remite a «Decide and list»; la tarjeta solo aprueba si cierra el turno; `/hive-mode` tecleado cuenta como carga). Quedan como hallazgos de cierre: ▶ o comando en cualquier parte del mensaje de cierre y `#N` sin la palabra «issue».
2026-11-19 — Segunda revisión limitada a las correcciones (`6cb6ada..0584c28`): 1 importante corregido en `6f39a06` con su prueba (`/hive-mode` cuenta como carga solo al inicio del prompt; commit con `Unreviewed:`). Hallazgos de cierre no bloqueantes: `example-record.md:7` aún dice «that leaf's test» (debería apuntar al registro de reglas); comentario de `ticketAssessmentEnding` y filas de la rúbrica sin el caso del comando y de la tarjeta que no cierra; redacción de c con doble negación; «what a–f send» en `hive-mode` perdió su referente; el disparador del requisito de la spec dice «picks a ticket for research». `TestE2ESetupDeadlinesOwnDescendantsAndPipes` es intermitente también en la base (1 de 5 en `a0e9b280`); la suite completa pasó en `6f39a06`.
