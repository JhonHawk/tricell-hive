# leaf-restatements — Citar las reglas fuera de su casa, sin resúmenes parciales

| Campo | Valor |
| --- | --- |
| Estado | Completed |
| Tracker | ninguno |
| Git | tricell-hive · base `development` (`d4f1eedb`) · rama propuesta `refactor/leaf-restatements` |
| Playbook | refactor |

## Decisions

2026-11-22
- **D13** — ¿Se planea quitar los resúmenes parciales de las hojas en `global.md`, las skills y las hojas? Sí, sin el test de solapamiento: el escaneo de esta sesión sirve una vez para la limpieza y no queda como guardia; #588 revisa los tests de texto.
- **D15 A** — ¿Qué levanta un permiso de admin-merge? Lo que dice `merge-gate.md:12` (y ya dice la spec «An admin-merge grant lasts the session»); `global.md:116` pasa a puntero.
- **D16 A** — Un commit ajeno que aparece después del push, como el autofix de un bot, también se pregunta antes del merge o de la parada `ready`; la casa es `foreign-commits.md` y `pr-reviews.md:9` pasa a puntero.
- **D17 A** — La evaluación de un ticket cuyos hallazgos no confirman trabajo termina como la fila de análisis de `report-endings` (como ya dice la spec «A chosen ticket's assessment is written in blocks»); `hive-mode` deja de devolverla.
- **D18 A** — Un ticket pasa a su estado de `Completion` cuando el recorrido funcional del despliegue observó sus criterios en ese entorno; la casa es `ticket-state.md` y `flow-deploy:34` apunta a ella; desaparece «due check».
- **D19 A** — Fuera de su casa, una regla se cita solo con su momento y un enlace a la sección exacta (`hoja.md#ancla`), sin texto propio sobre su contenido; las secciones citadas mantienen su título como identificador. Descartada: numeración tipo constitución, porque insertar una cláusula rompe en silencio las citas siguientes.
- **D20 A** — El índice «Rule leaves» se queda en `global.md` (solo 18–67 % de las sesiones medidas cumplen una lectura obligatoria en sus primeras 5 llamadas, según el host, `engram:tricell-hive/research/mandatory-read-compliance`) y se agrupa por dominio, un subtítulo `###` por dominio de `domains.json`, una hoja por línea.
- **D21 A** — ¿Qué queda de «Questions and routes» en `global.md`? Una línea invariante de ~300 B: preguntar solo lo que es del usuario, y solo su respuesta o un comando arranca otro flujo, salvo las continuaciones que lista `report-endings`, con enlaces a `questions`, `reasons-and-blockers` y `report-endings`, sin ningún formato; la spec «Layered global core with owning skills» se cumple.
- **D22 A** **Replaces:** D18 — ¿Cuándo pasa un ticket a `Completion` si el recorrido funcional no corre? Cuando corre, solo si observó sus criterios; cuando no corre, con el despliegue verificado, y el reporte lo nombra como no recorrido.
- **D23 A** — Una tarjeta que se cerró sola (vacía o por timeout, sin rechazo) se escribe completa en texto, como dice `messages/questions/unanswered.md`; `global.md:110` lo prohibía en general.
- **D24 A** — Una instrucción del propio proyecto se puede nombrar como razón con su efecto, como dice `reasons-and-blockers.md:5`; las reglas de Hive nunca son la razón.
- **supuesto** — Sin carpeta de cambio, las reglas de un ejemplo aprobado van al checkpoint, como dice `business-record.md:23-24`, no a `proposal.md` como dice `ui-example.md:5`.
- **supuesto** — Una clase del effects round no cubre los efectos que `external-effects.md:10` deja con su propia pregunta; `flow-plan:14`, `flow-deploy:24` y `data-change:10` decían lo contrario.
- **supuesto** — Un plan en la conversación lleva cada hecho con su medición; `plan-handoff` enlaza en vez de listar.
- **supuesto** — `python3 -I` en `change-folder.md:24` y `batch-archive.md`.
- **supuesto** — Entran los ~10 casos cercanos que encontró la clasificación de skills.

## Business rules

None: el trabajo no adopta ninguna regla de negocio.

## Request

«Las reglas se repiten a medias, por eso algunas veces la misma situación se comporta diferente»; «si pasa en global.md puede pasar en las skills referenciando esas hojas o de hojas referenciando a otras hojas»; «hagamos el plan»; «que el plan incluya mejorar el índice, además del ajuste que hablamos». Fuera de alcance: la carpeta `questions/` (el usuario la sacó de la mesa) y los tests de texto, que son #588.

## Objective

Hoy 77 lugares de `global.md`, skills, roles y hojas repiten parte de una hoja que además enlazan; varias copias contradicen a su casa (lo que deciden D15–D17, D22 y D21–D24), y en este mismo trabajo el agente siguió la versión corta (`global.md:109`) al escribir las rondas D3 y D5 sin `Situation:` ni `If you choose it:`. Al terminar, cada referencia fuera de la casa de una regla es su momento y un enlace anclado, o un invariante completo y corto que debe actuar antes de leer la hoja; cada contradicción queda resuelta en su casa, y el índice queda agrupado por dominio.

## Scope and acceptance

Dentro: los casos D y X de `design.md` en `global.md`, skills, playbooks, roles y hojas; los comportamientos D15–D17, D22 en su casa; el índice por dominio; `AGENTS.md` «Guidance structure» con la regla de citar con ancla; `instruction-resources.md`; `CHANGELOG.md`; los tests de contenido que dependen de las frases movidas. Fuera: `questions/`, la reforma de tests de #588, un test nuevo de solapamiento.

- AC1 — En `global.md`, «Questions and routes» es una sola línea invariante sin formato (D21) y los textos que `design.md` marca D o X en «Global» son punteros con ancla o invariantes completos · false today because `global.md:109-111` resume el formato de las decisiones sin `If you choose it:` y G3, G10, G16, G17 y G18 repiten parte de su hoja · Observed in: `content/guidance/global.md` del candidato
- AC2 — El índice «Rule leaves» tiene un subtítulo por dominio de `hive-core/assets/domains.json`, en ese orden, una hoja por línea, e incluye `principles/verify-current-state.md` · false today because el índice va en orden de flujo, junta hasta 3 hojas por línea y no lista ese principio · Observed in: lectura de la sección «Rule leaves» de `content/guidance/global.md` del candidato
- AC3 — Cada caso D o X de «Skills», «Roles» y «Leaves» de `design.md` queda como puntero con ancla o con su excepción local, sin pasos de su hoja · false today because el escaneo de 6 palabras encuentra 58 pares, 39 de ellos D o X · Observed in: el escaneo de 6 palabras sobre el candidato, que solo devuelve pares clasificados P, E o I en `design.md`
- AC4 — `foreign-commits.md` pide preguntar por un commit ajeno aparecido después del push antes del merge o de la parada `ready`, y `pr-review-deploy.md` y la rúbrica de sesiones no repiten el momento viejo · false today because `foreign-commits.md:3,5` solo dice «before validation or push» y `hive-mode/playbooks/pr-review-deploy.md:12` lo repite · Observed in: las hojas y el playbook del candidato
- AC5 — `ticket-state.md` mueve un ticket a su estado de `Completion` como fija D22 (con el recorrido, si corre, solo cuando observó sus criterios; sin él, con el despliegue verificado y nombrado como no recorrido), `flow-deploy` apunta a esa línea y el término «due check» no aparece en `content/` · false today because `flow-deploy/SKILL.md:34` dice «a ticket whose due check did not run» y remite a `ticket-state.md:9`, que remite de vuelta · Observed in: `rg -n 'due check' content` sin resultados y las dos líneas del candidato
- AC6 — `hive-mode` termina una evaluación sin trabajo confirmado con la fila de análisis de `report-endings`, sin devolverla, y `report-endings` no devuelve esa evaluación a `hive-mode` · false today because `hive-mode/SKILL.md:27` y la fila «Assessment of a ticket» se remiten mutuamente · Observed in: las dos líneas del candidato
- AC7 — `global.md` mide 28.000 B o menos (estimado tras D21: ~27.850 B) · false today because mide 29.210 B · Observed in: `wc -c content/guidance/global.md`
- AC8 — Ningún documento enlaza la sección «Evidence» borrada, describe la condición vieja de `Completion` ni da a «Questions and routes» el manejo de la tarjeta sin respuesta · false today because `README.md:476` dice «and no functional walk is pending» y `_support/docs/architecture/agent-delivery.md:258` dice que «Questions and routes» «owns … unanswered-card handling» · Observed in: `rg -n '#evidence|functional walk is pending' README.md llms.txt AGENTS.md site _support/docs` sin resultados, y lectura de `agent-delivery.md:258` y `tool-dependencies.md:52`, que apuntan el manejo de la tarjeta a `unanswered.md`

## Prerequisites

None: la construcción no necesita nada de fuera del repositorio.

## Verified context

- Clasificación de 77 casos por tres agentes de investigación sobre `d4f1eedb`: «Global» 18 bullets + 14 de roles, «Skills» 32 pares + ~10 cercanos, «Leaves» 17 pares · 2026-11-22
- La spec «An admin-merge grant lasts the session» ya coincide con `merge-gate.md:12`, y «A chosen ticket's assessment is written in blocks» ya termina con la fila de respuestas · lectura de `workflow-lifecycle/spec.md:1374,1422` · 2026-11-22
- Una lectura obligatoria se cumple en 18–67 % de las sesiones en sus primeras 5 llamadas, según el host · `engram:tricell-hive/research/mandatory-read-compliance` · 2026-11-22
- El manager valida cada ancla de un enlace de instrucciones al instalar · `tooling/management/references.go` · 2026-11-22
- Fuentes externas (investigación del 2026-11-22): Anthropic y Cursor recomiendan una copia canónica más punteros («Reference files with @ instead of copying», cursor.com/docs/context/rules; «Prefer pointers to copies», HumanLayer 2025-11-25); ninguna fuente recomienda IDs de cláusula para instrucciones, y la «constitution» de Anthropic es un documento de entrenamiento, no un formato; Vercel midió (2026-01-27) que un índice de 8 KB dentro de AGENTS.md pasó el 100 % de sus evals contra 53–79 % de skills que dependen de invocarse, y que «You MUST invoke» dio peores resultados; la documentación de Claude Code re-inyecta el archivo raíz tras compactar y no las rutas cargadas bajo demanda; ninguna fuente defiende tests de frases obligatorias, y Anthropic recomienda evals de comportamiento · 2026-11-22
- `wc -c content/guidance/global.md` · 29.210 B · 2026-11-22
- `go test ./... -count=1` · ok · 2026-11-22 (tras PR #587)
- Escaneo de 6 palabras, como lo define `design.md` «Overlap scan» (copia de esta sesión: `/private/tmp/claude-501/-Users-jmartinez-Development-projects-tricell-tricell-hive/2e5de8b8-cadc-4f7b-af48-d72138a2a0ab/scratchpad/overlap_detail.py <repo> <salida.json>`) · 58 pares · 2026-11-22
- `python3 -I -m unittest discover -s tests/skills -p '*_test.py'` · 412 OK · 2026-11-22

## Effects

None: las tareas no producen efectos en sistemas externos.

## Delivery

Entrega `hold` y revisión Claude Code `/code-review` sobre lo no mecánico, las respuestas vigentes de esta sesión. Sin push, PR, merge ni refresco de la instalación.

## Review status and progress

2026-11-22 — revisión de `hive-review-plan`: 3 P1 (la spec pide la línea mínima de preguntas; D22 bloqueaba tickets sin recorrido; G1 perdía calificativos), 5 P2 y 4 P3; D21–D24 los resuelve el usuario y el resto se aplica en esta versión.
2026-11-22 — plan escrito; la investigación externa sobre documentación y opiniones de usuarios confirma D19 y D20 y no cambia ninguna decisión.
