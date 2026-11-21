# reporting-family-leaves — Reportar tiene dos hogares: el reporte final y cómo termina el mensaje

| Campo | Valor |
| --- | --- |
| Estado | En build |
| Tracker | Sin ticket |
| Git | `JhonHawk/tricell-hive`, base `development` en `7f6277e0` (plan escrito sobre `2376cbbb`); rama `refactor/reporting-family-leaves` en el worktree `.claude/worktrees/reporting-family-leaves` |
| Playbook | ninguno (reestructura que además cambia comportamiento: resuelve contradicciones) |

## Decisiones

2026-11-20
- **Investigación A** — ¿Cómo se organiza una regla con variantes, como reportar un PR, un bug o un despliegue? Una hoja pequeña por familia, con la regla común y una tabla de variantes, leída entera en el momento de reportar; los playbooks solo nombran la variante. Se descartan los punteros a fragmentos (líneas o encabezados) y las variantes dentro de los playbooks.
- **Usuario** — El índice de hojas de `global.md` se queda; no se elimina en ningún adelgazamiento.
- **D1 A** — ¿Qué enfoque? Dos hojas, una por concepto: `hive-rule-close-report` guarda el reporte final de toda variante y `hive-rule-report-endings` guarda cómo termina cada mensaje; `hive-rule-next-step-block` y `hive-rule-close-footer` se funden en ellas y desaparecen; `hive-rule-close-modes` se queda con los pasos.
- **D2 A** — ¿Dónde va el trabajo? Rama propia desde `development`; el PR #554 (prueba de adelgazar `global.md`) se rebasa después, como prevé su protocolo.
- **D3 A** — ¿Cómo se asegura que el índice nombre el momento de su hoja? Prueba literal: el disparador de cada línea del índice es texto literal de la descripción de su hoja; se reescriben las líneas o descripciones que no lo cumplan.
- **D7 A (2026-11-21, plan `hive-core-stage-packets`)** — ¿Quién alinea las líneas del índice con su hoja? **Replaces:** D3 A. La alineación y su prueba pasan al plan `hive-core-stage-packets`, que borra el frontmatter de las hojas y compara cada línea del índice con la primera línea «Read» de su hoja. Este plan conserva solo la corrección de `close-hygiene`.
- **Supuesto** — Las hojas que conservan su nombre (`report-endings`, `close-report`) no se renombran, porque `tests/pilot/session_audit_endings.go:14`, sus fixtures y `single_home_test.go` las citan por nombre. Alternativa descartada: nombres nuevos. Se revierte renombrando y actualizando esos puntos.
- **Supuesto** — `hive-rule-report-endings` entra a las excepciones de 7,000 B de `leaf_skill_test.go`; `hive-rule-close-report` entra solo si, tras quitar duplicados, pasa de 4,000 B. Alternativa descartada: recortar la redacción para caber, que en la prueba #554 perdió matices. Se revierte sacándolas de la lista.
- **Supuesto** — La vista de hallazgos del cierre y su ronda pasan a una sección nueva de `close-report`, «Findings at the close»; `#findings` se queda con la definición y las prioridades que leen los revisores hijos. Alternativa descartada: meterlas en «Findings», que los hijos cargarían. Se revierte moviendo la sección.
- **Supuesto** — El reporte final de un trabajo lo escribe siempre `flow-close` en su paso 6; `git-workflow` paso 9 y `flow-deploy` paso 8 le entregan lo observado. La parada de un despliegue que espera una respuesta de recuperación es un mensaje de bloqueo, no un reporte final. Alternativa descartada: que el flujo escriba el reporte cuando el cierre no corre (ese caso no existe: `ready` y la espera de un revisor son punto de terminación, `hive-rule-ticket-state/SKILL.md:14`). Se revierte devolviendo el reporte a cada flujo.
- **Supuesto** — El paso 9 de `unattended-delegation` es un resumen de la sesión desatendida que apunta al reporte de cada cierre, con los efectos y su reversión y las decisiones tomadas. Alternativa descartada: que sustituya el reporte del cierre. Se revierte devolviéndole las partes del reporte.
- **Supuesto** — Cada playbook nombra su fila de `report-endings` y «la fila de `close-report` que elige el modo de entrega vigente»; solo `pr-review-deploy` nombra además la fila de despliegue. Alternativa descartada: una variante fija por playbook, que choca con la regla de playbooks de `leaf_skill_test.go:23-24`. Se revierte quitando la línea.
- **Supuesto** — La fila de cierre apunta a la línea de espera que ya exige `hive-rule-questions` «Format», sin redefinirla. Alternativa descartada: quitar el criterio. Se revierte quitando el puntero.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Solicitud

> Supongamos que tenemos 1 sola fuente de verdad para el tema de como reportar algo, la entrega de 1 pr, la entrega de un 1 bug, un despliegue etc, en teoria es lo mismo con variariones. […] Haz research sobre como recomiendan en documentación hacerlo, recuerda que debe ser compatible con todos los hosts.

Respuesta elegida (A): planear la reorganización de la familia de reportar en una hoja con tabla de variantes, con los playbooks nombrando la variante; incluye resolver las 2 contradicciones y alinear los disparadores del índice.

## Objetivo

Hoy reportar vive en 5 hojas (`report-endings` 2,602 B, `close-report` 3,776 B, `close-footer` 3,154 B, `close-modes` 3,197 B, `next-step-block` 3,815 B) y en pasos de cuatro flujos (`git-workflow` paso 9, `flow-deploy` paso 8, `flow-close` pasos 6 y 8, `unattended-delegation` paso 9). Consecuencias observadas:

- Una entrega con PR, un despliegue y el trabajo desatendido escriben dos reportes finales: el del flujo y el de `flow-close`.
- `flow-close` paso 8 y su descripción dicen que el cierre termina con el footer sin condición, y `close-modes` dice que seis situaciones terminan sin footer.
- Tres formas distintas de «reporte final» apuntan a la misma hoja.
- La continuación en el mismo turno de una investigación a `flow-build` está escrita en 6 lugares: `hive-mode/SKILL.md:14,25-26`, `report-endings`, `next-step-block`, `flow-build/SKILL.md:14,20`, `flow-plan/references/plan-handoff.md:21,28` y `global.md:137`.
- Las líneas del índice y las descripciones de sus hojas nombran el momento con palabras distintas, y nada lo comprueba.

Al terminar: el contenido del reporte final tiene un hogar con una fila por variante, cómo termina cada mensaje tiene otro hogar con una fila por forma de terminar, cada flujo apunta a su fila, cada trabajo produce un solo reporte final, y `close-hygiene` ya no reclama el paso 6. La prueba literal del índice la hace el plan `hive-core-stage-packets` (D7 A).

## Alcance y aceptación

Dentro: las 5 hojas de la familia; los pasos de los flujos que reportan o terminan; la descripción de `flow-close`; las líneas de `global.md` que las nombran («Rule leaves» y «Questions and routes»); las líneas del índice de las hojas de la familia y la de `close-hygiene` (la alineación de las 62 líneas pasa a `hive-core-stage-packets`, D7 A); los 13 playbooks (una línea cada uno); las pruebas de contenido; `site/` (`flows/close.md`, `flows/deploy.md`), `llms.txt`, `CHANGELOG.md` y la documentación de arquitectura que las citan; el delta de `workflow-lifecycle`.

Fuera: el PR #554 y su presupuesto de bytes; el formato de las preguntas (`hive-rule-questions`); que cada hijo reporte las hojas que leyó y las rutas en el brief (opciones aprendidas de gentle-ai, otro trabajo); cualquier prueba piloto con modelos.

- AC1 — `hive-rule-close-report` tiene una tabla con una fila por variante del reporte final (cierre en el punto de terminación, PR dejado abierto en `ready`, despliegue, `hold`, pausa) que dice qué partes lleva y qué añade cada una, con la fila de despliegue completa (la ejecución, por ambiente la identidad, la salud y la aceptación funcional como cumplida, fallida, bloqueada o «no corrió», los disparadores, las fallas de CI previas que dejó pasar y los tickets que esperan un chequeo), y `git-workflow` paso 9, `flow-deploy` paso 8 y `unattended-delegation` paso 9 apuntan a su fila sin listar partes propias · false today because esos pasos listan sus propias partes y la hoja no tiene tabla · Observed in: los archivos de `content/skills/` del candidato
- AC2 — Cada trabajo produce un solo reporte final, escrito por `flow-close` en su paso 6: `git-workflow` paso 9 y `flow-deploy` paso 8 le entregan lo observado sin escribir reporte, y el paso 9 de `unattended-delegation` es un resumen de la sesión que apunta a cada reporte de cierre. El mensaje con el resultado del CI de la base que llega después del cierre (`hive-rule-merge-gate`) no es un reporte final · false today because `git-workflow/SKILL.md:22` reporta «as the final report sets» antes del paso 10, `flow-deploy` paso 8 reporta antes de su cierre de limpieza, y `close-modes:15` corre el paso 6 también bajo trabajo desatendido · Observed in: los archivos de `content/skills/` del candidato
- AC3 — `hive-rule-report-endings` tiene una tabla con una fila por forma de terminar (análisis, evaluación de un ticket, paso a otro flujo con su bloque y la sintaxis por host, cierre con footer y sus respuestas, cierre sin footer, acción única) y la sección de continuaciones en el mismo turno, y `hive-rule-next-step-block` y `hive-rule-close-footer` ya no existen ni se enlazan · false today because los dos directorios existen y 11 archivos los enlazan · Observed in: los archivos de `content/` del candidato
- AC4 — `flow-close` paso 8, su descripción, `llms.txt` y `site/src/content/docs/flows/close.md` dicen que el cierre termina como fija la fila de `close-modes`, no siempre con el footer · false today because `flow-close/SKILL.md:21` dice «End with the close footer», su descripción (línea 3) «reports the outcome with a close footer» y `llms.txt:41` «end[s] with a close footer» · Observed in: los archivos del candidato
- AC5 — La continuación en el mismo turno de una investigación a `flow-build` está escrita entera solo en `report-endings`; `hive-mode`, `flow-build`, `plan-handoff` y `global.md` la nombran con un enlace y conservan solo su propio paso; `single_home_test.go` fija su frase en `report-endings` · false today because está escrita en los 6 lugares de «Objetivo» · Observed in: `go test ./tests/content/` en el candidato
- AC6 — Una prueba de contenido exige que el disparador de cada una de las 62 líneas de `## Rule leaves` sea texto literal de la descripción de su hoja, y pasa; la línea de `close-hygiene` y su hoja ya no incluyen el paso 6 de `flow-close` · false today because la prueba no existe, la mayoría de las líneas no cita su descripción literalmente, y `global.md:69` y `close-hygiene` dicen «steps 4 to 7» aunque el paso 6 es de `close-report` · Observed in: `go test ./tests/content/` en el candidato, vista fallar en la base **Retired:** D7 A pasa la prueba literal al plan `hive-core-stage-packets`; lo reemplaza AC10.
- AC7 — Cada uno de los 13 playbooks tiene una línea que enlaza su fila de `report-endings` y «la fila de `close-report` que elige el modo de entrega vigente», sin nombrar un modo · false today because 10 de 13 no enlazan ninguna hoja de reporte · Observed in: `go test ./tests/content/` en el candidato
- AC8 — La fila de cierre de `report-endings` enlaza la línea de espera de `hive-rule-questions` «Format» · false today because `close-footer` y `close-report` no la mencionan (watch `#17750`) · Observed in: los archivos de `content/skills/` del candidato
- AC9 — Ninguna regla se pierde: cada oración de las fuentes listadas en T1 tiene su destino en `inventory.md` (archivo y sección, o «se borra: duplicado de <destino>»), y `/code-review` lo confirma con el inventario · false today because `inventory.md` no existe · Observed in: `inventory.md` y el informe de `/code-review`
- AC10 — **Replaces:** AC6. La línea de `close-hygiene` en `## Rule leaves` y la hoja `hive-rule-close-hygiene` ya no incluyen el paso 6 de `flow-close`, que pertenece a `close-report` · false today because `global.md:69` y `close-hygiene` dicen «steps 4 to 7» · Observed in: los archivos de `content/` del candidato

Restricción de todas las tareas (no es un criterio, porque hoy ya se cumple): `global.md` no crece (≤ 28,593 B, su tamaño en `7f6277e0`, medido con `wc -c`; el tope de la prueba es 30,000 B) y cada hoja cumple su límite (4,000 B, o 7,000 B para las excepciones listadas).

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

- Mapa de la familia, tamaños y contradicciones: investigación de esta sesión (2026-11-20) sobre `development` en `2376cbbb`; comprobado a mano `flow-close/SKILL.md:21` frente a `hive-rule-close-modes/SKILL.md:15`, `git-workflow/SKILL.md:22-23`, `flow-deploy/SKILL.md:31,33`, `unattended-delegation/SKILL.md:20` y `hive-rule-ticket-state/SKILL.md:14`.
- Enlaces a las dos hojas que se borran: 11 archivos fuera del archivo histórico, entre ellos `tests/content/leaf_links_test.go:30`. Enlaces a las 5: 40 archivos, entre ellos `content/agents/review/*` (que leen `close-report#findings`), `tests/pilot/session_audit_endings.go:14`, `tests/content/single_home_test.go:48-50,82` y `tests/content/reading_footprint_test.go:137,148`.
- Límites y regla de playbooks: `tests/content/leaf_skill_test.go:19-24` (4,000 B; excepción 7,000 B para `hive-rule-questions` y `hive-rule-merge-gate`; playbooks 7,000 B y solo la frase del modo vigente).
- El índice cubre las 62 hojas (62 directorios, 62 entradas, ninguna sobrante).
- Evidencia de diseño: documentación oficial (Anthropic, agentskills.io, Cursor) pide archivos pequeños leídos enteros a un nivel del índice; Grok ignora `offset`/`limit` en markdown de skills; el hook `hive-read-whole-references` existe por 29 de 220 lecturas incompletas (`_support/docs/architecture/deployment-manager.md:542`); gentle-ai quitó los resúmenes de skills por distorsionarlas (`faa4f7fe2`). Engram: `decision/reporting-family-leaf`.
- Spec afectada: `_support/openspec/specs/workflow-lifecycle/spec.md` «Moving between stages shows the next-step block» y «The final report carries evidence and a try-it block»; `project-layout/spec.md:54` menciona el footer y se conserva. Carpeta abierta en la base: `skills-rebuild` (2026-11-01, sin relación con estas hojas).

## Effects

None: the tasks take no effect on an external system.

## Entrega

2026-11-21, respuestas en vigor de la sesión del build. **Replaces:** la entrega `hold` de 2026-11-20, que respondió otra sesión.
- **Entrega: `automatic`.** Reviso, hago push, abro el PR y lo mergeo sin parar.
  - Se trabaja en el worktree `.claude/worktrees/reporting-family-leaves`, rama `refactor/reporting-family-leaves`.
  - Tras el merge, el refresco local sigue la regla del mantenedor de `AGENTS.md`. El reporte avisa que las sesiones abiertas deben reiniciarse, porque desaparecen dos hojas instaladas.
- **Revisión: nativa, `/code-review` de Claude Code**, sobre el candidato antes del push, con `inventory.md` como insumo para AC9.

## Estado de revisión y progreso

2026-11-20 — Plan escrito tras D1 A y D2 A.
2026-11-20 — Revisión del plan (`hive-review-plan`): 4 P1 y 10 P2 aplicados antes de aprobar, con los criterios reescritos en su sitio (el plan aún no estaba aprobado): un solo escritor del reporte (supuesto), resumen desatendido, «Findings at the close», playbooks neutrales al modo, fila de despliegue completa, inventario por oración, criterios observados en el candidato, `site/flows/deploy.md`, `llms.txt`, descripción de `flow-close`, `leaf_links_test` y `CHANGELOG.md` añadidos; D3 A decide la prueba literal del índice.
2026-11-21 — Revisión por D7 A del plan `hive-core-stage-packets`, que se construye después de este: la prueba literal del índice y la alineación de sus líneas pasan a ese plan. AC6 queda retirado y lo reemplaza AC10, que conserva solo la corrección de `close-hygiene`. T6 se reescribe y no invalida ninguna tarea marcada, porque ninguna lo está. La entrega `hold` de 2026-11-20 era una respuesta de aquella sesión; al construir este plan se pregunta de nuevo.
