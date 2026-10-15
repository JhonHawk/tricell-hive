# Poda de la guía por procedencia: lote 1 (2026-10-15)

| Campo | Valor actual |
| --- | --- |
| Estado | Integrado en `development` el 2026-10-15 por el [PR #57](https://github.com/JhonHawk/tricell-hive/pull/57), merge `d68d49b1` (commit `27d32fca`); desplegado en local desde `development` y archivado el 2026-10-15. Estado previo (historia): completado en local, T1–T5 `[x]`, AC1–AC8 cumplidos, `global.md` 27 318 → 26 955 bytes |
| Tracker · GitHub Issues | Sin issue: el usuario lleva este trabajo desde la memoria y los registros de cambio |
| Git | Integrado por merge del PR #57 sobre la base `0b2e0bac`; la rama `feat/guidance-pruning` y su worktree se eliminaron tras el merge |
| Verificación | `go test ./tests/content/...` por tarea · `go test ./...` y tamaños sobre el candidato final · `hive-verify-change` al cerrar |
| Siguiente paso | Ninguno; registro cerrado. Cambio aparte decidido: clave `Done:` por proyecto (C2b-A), cuando cam, carvento y dened fijen su punto de cierre. Lotes 2 y 3: no ahora (DD7) |

## Solicitud

> cómo podriamos retomar la investigación anterior en la practica?

> avanza en un analisis adversarial sobre la propuesta, el inventario y la revisión en general, no hagamos commit de estos archivos

> tenemos que usar la maxima exisgencia sobre la revisión de esta propuesta por favor, es un deep-research profundo sobre el tema

> Lo que me preocupa es que en todas las propuestas que me das, global.md crece

Respuestas del usuario (2026-10-15, tarjetas nativas y texto):

- D1 (ruta inicial): "Inventario primero". D1 (tras el inventario): "plan de los tres lotes, aunque lo vayamos aplicando por etapas"; sustituye la decisión D1-A "Conservar" del cambio `flash-class-guidance` (2026-10-12).
- D2 (retención): "Sí, en el registro de cambio": los dos inventarios viven en `inventory/`.
- D3 (tras el análisis adversarial): "Reducir a lote 1 corregido": duplicados literales reales menos los que lee un rol sin guía global, los tres conflictos resueltos por tarjeta, punteros intactos, sin tocar `AGENTS.md`; los lotes 2 y 3 quedan como "no ahora".
- C1 (hallazgos de revisión), texto: "Movamos el tema de la severidad solo cuando es ejecutado desde flow-build, o acotemoslo al flujo completo (cuando hay research o plan previo), la intención es que las ejecuciones directas para tareas menores no tengan esta carga de razonamiento, solo que ataquen o resuelvan los que se les indico".
- C2 (ticket hecho frente a CI posterior al merge), texto: "esto talvez vaya a ser dinamico, en el caso de de cam / carvento, no hay qa explicito que detenga los tickets de linear (pero si los de jira), por eso el ciclo funciona diferente, en muchos equipos tendran su propia definición y ciclo de cierre, talvez sea algo que se tenga que definir si no lo esta y colocar en las reglas del AGENTS.md de cada proyecto (research sobre esto)".
- C3 (alcance del borrado de ramas): "Ampliar global a las ramas fusionadas del repo".
- Restricción del usuario (texto): `global.md` no debe crecer con este cambio.
- C1b (tras la segunda ronda de revisión): "No, C1 solo cubre la revisión": la regla de hallazgos incidentales de `global.md:31` no cambia.
- D4 (ruta): "Arrancar flow-build ahora": T2–T4 en el hilo principal, T5 delegada; diff sin commit.
- C2b (tras el research de T5): "Clave `Done:` en otro cambio": opción A (`Done: base | <entorno> | manual`) como registro aparte, cuando cam, carvento y dened fijen su punto; este cambio no toca el punto de "ticket hecho".

## Objetivo

El inventario de `inventory/` (559 reglas entre `global.md` y las cuatro skills de flujo) y el análisis adversarial del 2026-10-15 (`design.md`, "Contexto verificado") dejaron tres resultados: los lotes 2 y 3 de la idea original no tienen método de detección que funcione sin evaluaciones propias y ahorrarían poco; los duplicados "literales" son en su mayoría copias deliberadas que leen roles sin guía global, o punteros de lectura que Cursor necesita; y las tres contradicciones entre `global.md` y `flow-close` o `verification.md` son defectos reales con origen rastreable. El lote 1 corrige esas contradicciones según las decisiones C1–C3, retira de `global.md` el texto que ya tiene casa en una referencia leída siempre en el mismo momento, y deja `global.md` más corto, sin cambiar ninguna regla de clase `observed` o `decision` más allá de lo que esas decisiones fijan. El punto de "ticket hecho" se investiga como ajuste por proyecto (C2) en vez de resolverse en la guía.

## Alcance y aceptación

Incluido: `content/guidance/global.md` líneas 23, 31, 93, 108; `content/skills/flow-close/SKILL.md:8,74` y `content/skills/workspace-conventions/references/artifact-layout.md:41` (alcance de la concesión de C3); `content/skills/flow-build/SKILL.md:18` ("code review" en la lista de lecturas de `verification.md`); `content/skills/flow-build/references/verification.md:41,43,61,71` solo para comprobar que siguen siendo la única casa de la severidad; `content/skills/flow-research/SKILL.md:40` (una oración duplicada con `reporting.md:77`); `site/src/content/docs/flows/build.md:43` y `close.md:31,54` para describir C1 y C3; `CHANGELOG.md` (`[Unreleased]`); un research acotado para C2 (T5) que termina en una tarjeta, no en una edición.

Excluido, con razón:

- Lotes 2 y 3 (reglas de comportamiento por defecto y reglas sin origen): sin método de detección (ventana de diez sesiones detecta solo fallos con tasa ≥ 20 %; las reglas de secretos, efectos irreversibles, autorización e inyección tienen tasa base de 0–2 eventos al mes) y con ahorro de ~1,5 KB (lote 2) y 3–5 KB (lote 3) frente a veinte tarjetas y dos ventanas. Quedan como "no ahora" con el inventario como referencia; reabrirlos exige evaluaciones por sonda como las que la referencia que midió un recorte usó (control 8/10 → tratamiento 5/10, revertido).
- Duplicados entre `global.md` y `change-records.md:7`, `artifact-layout.md:41` y `reporting.md:19` como copias: los leen roles (`hive-write-spec` y otros) que en cinco hosts no cargan la guía global según lo que el repositorio establece; la referencia conserva el texto completo. En `global.md` solo se reduce lo que el hilo principal relee siempre en el mismo momento por un puntero (T3).
- Los 34 punteros "read X" de las skills y los siete de `global.md`: Cursor CLI no carga la guía global (`README.md:185`) y ninguna prueba detecta un puntero perdido.
- `AGENTS.md` (Measurement) y una lista de vigilancia: sin lote 2 no hace falta una vía de retiro nueva; el texto retirado por T2–T3 queda en este registro.
- Roles de `content/agents/`: copias deliberadas para hijos sin guía global.

Restricciones que deben mantenerse: `go test ./...` en verde en cada candidato; `global.md` termina con menos bytes que en la base (`TestGlobalGuidanceSize`); ninguna regla de clase `observed` o `decision` cambia de significado salvo lo que C1 y C3 fijan; cada oración retirada de `global.md` queda citada en `tasks.md` con su casa restante; el PR declara los lugares revisados y los tamaños antes y después.

Criterios de aceptación (falsos en la base `0b2e0bac`):

- AC1. `global.md:31` ya no fija cómo se tratan los hallazgos de una revisión-gate: conserva "A review requested on its own does not authorize fixes" y deja la severidad a `verification.md:43`, que sigue siendo la única casa de ese criterio. *Falso en la base cuando* `rg -n "are fixed within it" content/guidance/global.md` devuelve la línea 31.
- AC2. `global.md:108` concede borrar las ramas de trabajo fusionadas, de este trabajo o de trabajos anteriores, con sus worktrees limpios (las ramas de entorno no son ramas de trabajo; las preview del trabajo sin PR sí); `flow-close/SKILL.md:8,74` y `artifact-layout.md:41` citan la concesión con el mismo alcance. *Falso en la base cuando* `rg -c "from this work or earlier" content/guidance/global.md content/skills/flow-close/SKILL.md content/skills/workspace-conventions/references/artifact-layout.md` no devuelve nada.
- AC3. `global.md:93` conserva solo la ubicación de `openspec` en una oración y el puntero a `change-records.md`; las otras dos oraciones viven solo en `change-records.md:7`. *Falso en la base cuando* `rg -c "never place it in a workspace outside Git" content/guidance/global.md` devuelve 1.
- AC4. `global.md:108` ya no contiene "Repeating a test may not reproduce the same evidence", que vive en `artifact-layout.md:41`; `global.md:23` ya no contiene la oración descriptiva que empieza "Reporting details and closing categories live in"; `flow-research/SKILL.md:40` ya no repite "Start the chosen flow without a new command", que vive en `reporting.md:77`. *Falso en la base cuando* `rg -F "Repeating a test may not reproduce" content/guidance/global.md`, `rg -F "Reporting details and closing categories live in" content/guidance/global.md` y `rg -F "Start the chosen flow without a new command" content/skills/flow-research/SKILL.md` devuelven una línea cada uno.
- AC5. `global.md` pesa menos que en la base. *Falso en la base cuando* `wc -c content/guidance/global.md` da 27318 y el candidato no da menos.
- AC6. `site/.../build.md:43` describe C1 (hallazgos corregidos por severidad bajo Build) en una frase. *Falso en la base cuando* `rg -n "severity" site/src/content/docs/flows/build.md` no devuelve nada. (`close.md:31,54` ya describe C3 en la base y no cambia.)
- AC7. `CHANGELOG.md` tiene la entrada de este cambio bajo `[Unreleased]`. *Falso en la base cuando* `rg -n "from this work or" CHANGELOG.md` no devuelve nada.
- AC8. Existe una tarjeta de decisión sobre el punto de "ticket hecho" por proyecto (C2) cuyas opciones fundamentadas están en `design.md` y cuya respuesta está registrada en este `proposal.md`. *Falso en la base cuando* `rg -n "C2b" _support/openspec/changes/guidance-pruning/design.md _support/openspec/changes/guidance-pruning/proposal.md` no devuelve nada.

## Entrega

Decisiones vigentes para este repositorio en esta sesión (flujo de rama de prueba, 2026-10-13 y 2026-10-15): modo `hold` en `feat/guidance-pruning`; sin commit, push ni PR hasta que el usuario lo pida (el usuario lo pidió explícitamente el 2026-10-15: "no hagamos commit de estos archivos"); despliegue local desde la rama a pedido; sin revisión de código dedicada (inspección propia del diff, `go test`, `hive-verify-change` al cerrar); integración en `development` por PR tras las pruebas del usuario, declarando lugares revisados y tamaños.
