# global-guidance-slimming — Recortar la guía global a ~20 KB y frenar su crecimiento

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | none · relacionado: #553 (router de reglas por tecnología, espera a este cambio) |
| Git | tricell-hive · base `development` @ 27875867 · rama propuesta `refactor/global-guidance-slimming` |
| Playbook | refactor |

## Decisions

2026-11-20
- **D1 A** — ¿Qué hacer con las 47 reglas y el índice de `content/guidance/global.md`? Adoptar la clasificación completa: eliminar las copias de reglas que ya viven en su hoja, convertir la evidencia en un principio, mover el procedimiento a su hoja dejando el invariante en una línea, compactar el índice y conservar los invariantes. Se construye en una rama sin fusionar y se prueba antes de decidir la fusión.
- **D2 A** — ¿Qué tope impone el test de presupuesto? 22.000 bytes; el test falla por encima y su mensaje nombra la regla de destino.
- **D3 A** — ¿Cómo se prueba antes de fusionar? Instalar la rama en local, que el usuario trabaje sesiones reales en Claude, Codex y Grok, y comparar con `--audit-session` contra sesiones previas; el usuario decide con ese informe.
- **D4 A** — ¿Ticket para el router de reglas por tecnología? Creado: #553.
- **D5 A** — ¿Qué pasa con el maintainer refresh durante la prueba? Tras cada refresh desde `development`, la rama se rebasa sobre lo nuevo y se reinstala; cada sesión registra el commit instalado.
- **Supuesto** — La definición de efecto crítico se queda en el global (descartado: moverla a `hive-rule-external-effects`), porque otras reglas la usan sin leer una hoja; se revierte moviéndola.
- **Supuesto** — El índice no tiene tope propio (descartado: 5.200 bytes); sus líneas pueden agrupar hojas del mismo momento y AC1 acota el tamaño; se revierte reponiendo el criterio de tope del índice.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> "32kb son demasiadas reglas que aun se conservan ahi, analizalo" · "muchas de esas reglas podrian vivir en algun principio, ser un principio o simplemente eliminarse" · "Vamos a adoptar completo, pero desde una rama sin mergear, y haremos algunas pruebas antes de hacerlo, esto es por si hay regresiones o reglas ignoradas, lo podamos evaluar y si el impacto es grande, replantearlo".

Sin ticket. Fuera de alcance: el router de reglas por tecnología (#553).

## Objective

Hoy `content/guidance/global.md` mide 32.148 bytes (153 líneas, 47 reglas más un índice de 61 hojas de 7.483 bytes) y nada limita su tamaño: `tests/content/budget_test.go` solo lo informa. Tras cada recorte vuelve a crecer en dos o tres semanas (45 → 25 → 42 → 12 → 57 → 31 → 32 KB entre el 8 de octubre y el 20 de noviembre). Al terminar, mide 22.000 bytes o menos, conserva solo los invariantes que deben actuar sin leer otra hoja más un puntero a cada hoja, un test impide que vuelva a pasar de 22.000 bytes y la guía del repositorio dice dónde va una regla nueva.

## Scope and acceptance

Dentro: `content/guidance/global.md`, las hojas que reciben el procedimiento movido, un principio nuevo, el índice de principios, `tests/content/`, `AGENTS.md` «Guidance structure», la especificación `context-delivery`, y la documentación que describe el global (`site/`, `README.md`, `llms.txt`) cuando cambie algo que afirma. Fuera: borrar o relajar el comportamiento de una regla (cada regla sigue existiendo en un solo lugar), el router por tecnología y cualquier cambio del manager.

- AC1 — El archivo fuente `content/guidance/global.md` mide 22.000 bytes o menos (el texto instalado crece ~28 bytes por enlace al reescribir rutas, así que no se mide instalado) · false today because mide 32.148 · Observed in: `wc -c content/guidance/global.md` en el candidato
- AC2 — `go test ./tests/content/...` falla cuando el global supera 22.000 bytes, y el mensaje nombra la regla de destino de `AGENTS.md` · false today because `TestGlobalGuidanceSize` solo registra el tamaño con `t.Logf` · Observed in: el caso de prueba del test con un global de 22.001 bytes
- AC3 — Cada procedimiento que sale del global aparece una sola vez bajo `content/`, en la hoja o la referencia que `design.md` «Mapa de reglas» le asigna con su condición de lectura, y el global conserva para él una línea con su invariante y el enlace; las autorizaciones de escritura en tickets, de limpieza al cierre y la escalada de un P0, la definición de efecto crítico, de lectura grande, de cambio riesgoso y de tarea operativa, y la pregunta por un proyecto externo siguen en el global · false today because, por ejemplo, la estimación de duración de un check solo está en el global y no en `hive-rule-checks` · Observed in: el candidato, comparando el mapa con `rg` de la frase clave de cada regla
- AC4 — Las reglas que el global repetía de su hoja (canal de preguntas, tarjeta sin responder, bloque de siguiente paso, pregunta de entrega, commits ajenos) quedan como una línea mínima o solo como entrada del índice; E2E e identidad de Engram quedan en una línea cada una que conserva su autorización o prohibición; el test de hogar único registra la frase clave de cada una · false today because el global contiene el texto completo de cada una (1.117 bytes solo las tres de preguntas) · Observed in: `go test ./tests/content/...` y `rg` de cada frase clave
- AC5 — Un principio oculto `hive-principle-verify-current-state` contiene la regla de evidencia (verificar memoria y resúmenes, leer la documentación vigente), `hive-principle-index` lo enlaza con su momento, y el global conserva una línea de invariante que enlaza el principio · false today because la regla vive completa en la sección «Evidence» del global y el principio no existe · Observed in: el candidato y `go test ./tests/content/...`
- AC6 — **Retired:** sin una cifra propia, cada hoja una vez con su momento ya es cierto hoy y AC1 acota el tamaño; las líneas del índice pueden agrupar hojas que comparten momento
- AC7 — Un informe de prueba compara las sesiones reales hechas con la rama instalada contra sesiones previas del mismo host y modelo del piso: (a) por cada criterio de `go run ./tests/pilot --audit-session`, fallos sobre las sesiones donde el criterio se observó (pass + fail; `review` y `not_observed` se cuentan aparte), y es regresión si los fallos suben de 0 a 2 o más, o la tasa sube 20 puntos o más con al menos 5 sesiones observadas; (b) por cada regla movida del «Mapa de reglas», su detector o «revisión manual» con lo que se busca en cada sesión, y su resultado; cada sesión registra host, modelo y commit instalado · false today because no existe sesión con la rama instalada ni informe · Observed in: sesiones reales de Claude, Codex y Grok en esta máquina tras instalar la rama

## Prerequisites

- [ ] Sesiones reales de trabajo con la rama instalada en Claude, Codex y Grok · owner: el usuario · needed by: T7

## Verified context

- Tamaño y secciones medidos el 2026-11-20 sobre 27875867: Scope and authorization 7.544 B, Rule leaves 7.483 B, Proportionality 3.293 B, Preservation 2.851 B, Persistent memory 2.305 B; el resto ≤ 2 KB por sección.
- `_support/openspec/specs/context-delivery/spec.md` «Layered global core with owning skills» exige que el global contenga las autorizaciones (escrituras en tickets, limpieza al cierre, escalada de P0) y el índice; «Rule leaves hidden from model invocation» exige cada hoja una vez en el índice; `tests/content/leaf_skill_test.go:156` lo comprueba.
- 60 de las 61 hojas del índice ya las enlaza al menos un flujo, rol u otra hoja (conteo del 2026-11-20); `hive-rule-project-verification-format` no tiene enlaces fuera del índice.
- `tests/pilot/session_audit.go` audita transcripciones de Claude, Codex y Grok sin correr modelos; criterios actuales: `leaf_read_before_action`, `decision_round_format`, `text_round_without_card`, `prose_choice_without_card`, `analysis_reply_block`, `ticket_assessment_ending`, `close_before_branch_cleanup`, `flow_entered_without_route`, `skill_named_as_reason`, `review_question_native_option`, `self_feeding_log`, `tracker_states_cache_written`, `try_it_block_in_report`, `ui_change_without_ux_review`.
- La guía global llega a los subagentes de roles Hive en 6 de 7 hosts (investigación del 2026-11-20, `engram:tricell-hive/decision/language-rules-router`).

## Effects

- Instalación local de la rama en los hosts registrados (`hive update --source <worktree> --out FILE` y `hive apply --plan FILE`) · target: esta máquina · ceiling: 0 · tasks: T7 · answer: approved, D3 A
- Restaurar la instalación desde `development` si el usuario descarta la rama · target: esta máquina · ceiling: 0 · tasks: T7 · answer: approved, D3 A
- Rebasar la rama sobre `development` y reinstalarla tras cada maintainer refresh durante la prueba · target: rama remota y esta máquina · ceiling: 0 · tasks: T7 · answer: approved, D5 A

## Delivery

Modo `ready` (2026-11-20): revisión, push de `refactor/global-guidance-slimming` y PR listo hacia `development`, que se queda abierto; nunca se fusiona, ni se activa auto-merge. La rama se instala en local para la prueba (Effects, D3 A). Revisión de código: nativa, `/code-review` de Claude Code, sobre el candidato antes del push. No autoriza: fusionar, desplegar a `master` ni publicar una versión.

## Review status and progress

2026-11-20 — plan escrito.
2026-11-20 — revisión del plan (hive-review-plan): 17 hallazgos; aplicados los de criterio, destino, privacidad y documentación (detalle en tasks.md). AC6 retirado. D5 respondida.
