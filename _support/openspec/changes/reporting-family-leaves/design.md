# Diseño

## Enfoques

Reportar son dos conceptos con variantes: qué dice el reporte final (cierre, PR abierto en `ready`, despliegue, `hold`, pausa) y cómo termina el mensaje (análisis, evaluación de ticket, paso a otro flujo, cierre con o sin footer, acción única, continuaciones).

1. **Dos hojas, una por concepto, con tabla de variantes (elegido, D1 A).** `close-report` guarda el reporte final; `report-endings` absorbe `next-step-block` y el encabezado y las respuestas de `close-footer`; `close-modes` se queda con los pasos. Reutiliza hojas existentes (escalón 2 de reutilizar antes de escribir). Costo: `report-endings` necesita la excepción de 7,000 B.
2. **Una sola hoja de reportar.** Junta 13.3 KB en ≤7,000 B; exige recortar ~45 % de la redacción, que en la prueba #554 perdió matices (`decision/global-guidance-slimming`, «Learned»).
3. **Mínimo.** Deja las 5 hojas, corrige las 2 contradicciones, quita copias y pone la tabla solo en `close-report`. El concepto sigue en 5 archivos.

Descartados en la investigación previa: punteros a fragmentos de una hoja (ningún proveedor los documenta; Grok ignora `offset`/`limit` en skills; los números de línea se desfasan) y variantes escritas dentro de los playbooks (se leen una vez al rutear y no sobreviven a la compactación, `#17836`).

### Comparación en tres versiones (`decisions.local.md`, regla de evaluación durante la reforja)

| | pstack | Hive antes | Hive ahora |
| --- | --- | --- | --- |
| Regla común | Hojas de principios leídas «en completo» desde un ruteador | Un `reporting.md` grande en `hive-core`; se leyó parcial y se perdió la regla de la línea 108 (Cursor `afaeb558`, `#16727`) | Repartida en 12 hojas y 5 flujos |
| Variantes | El playbook copia sus pasos a una lista de tareas | Dentro de la misma referencia grande | Dispersas, con 2 contradicciones |
| Lectura | Siempre completa | Parcial en Grok (29 de 220) | Completa por regla; hook en Grok |

Recomendación dada en la sesión: dos hojas por concepto, leídas enteras, con tabla de variantes. El usuario eligió D1 A.

## Decisiones

- **D1 A** — dos hojas por concepto.
- **D2 A** — rama propia; #554 se rebasa después.
- **D3 A** — prueba literal del índice contra la descripción de cada hoja. Reemplazada por **D7 A** (2026-11-21): la prueba pasa al plan `hive-core-stage-packets`.
- Supuestos (nombres conservados, excepciones de tamaño, «Findings at the close», un solo escritor del reporte, resumen desatendido, playbooks neutrales al modo, línea de espera por enlace): ver `proposal.md` «Decisiones».

## Forma propuesta

`hive-rule-close-report` («Final report, findings and candidates»):

- Lectura: «Read when you write a work item's final report, in any variant below, or classify a finding».
- «Final report»: las partes de hoy (Integrated, Resolved, Try it, Omitted, Repaired, Support material, Findings, Candidates, Next).
- Tabla de variantes. El reporte lo escribe siempre `flow-close` en su paso 6; la fila la elige el modo de entrega vigente o el flujo que llamó al cierre:

| Variante | Partes | Añade |
| --- | --- | --- |
| Cierre en el punto de terminación (merge hecho) | todas las que tengan contenido | — |
| PR dejado abierto en `ready` o esperando a un revisor | todas | head, base del chequeo y CI en Integrated; el merge pendiente y de quién en Next |
| Despliegue (cierre de limpieza que llama `flow-deploy`) | todas | la ejecución; por ambiente la identidad servida, la salud y la aceptación funcional como cumplida, fallida, bloqueada o «no corrió»; los disparadores decididos; las fallas de CI previas que dejó pasar el paso 5; los tickets que esperan un chequeo |
| `hold` | lo pendiente, la prueba del defecto y Try it | — |
| Pausa fuera de `hold` | como `hold` | el commit de la carpeta de cambio |

- «Findings»: definición y prioridades, sin cambios (la leen los revisores hijos).
- «Findings at the close»: la tabla de hallazgos y su ronda que hoy están en `close-footer`, más «Find the next ticket».
- «Candidates», sin cambios.

`hive-rule-report-endings` («How a message ends»):

- Lectura: «Read before you end a message that reports, hands over or closes».
- Tabla, una fila por forma de terminar: análisis, explicación, revisión o diagnóstico → bloque de respuestas con letras y `(★)`, sin comando; evaluación de un ticket elegido → `hive-mode` paso 5; paso a otro flujo → bloque de siguiente paso (`---`, encabezado, razón, comando con argumento real, respuestas con letra) y compactar como respuesta cuando aplica; cierre con footer → su encabezado y sus respuestas (Close, Resolve findings now, See the findings, Explain, Turn a candidate into a mechanism, Find the next ticket, Hygiene, Continue), según la fila de `close-modes`; cierre sin footer → el reporte; acción única → una respuesta más del bloque o una decisión. Todo mensaje que cierra trabajo termina con la línea de espera de `hive-rule-questions` «Format».
- Sección «Same-turn continuations»: las cinco continuaciones (hogar único) y la sintaxis del comando por host.

`hive-rule-close-modes`: sin cambios en los pasos; su columna «Ends with» enlaza la fila de `report-endings`.

Flujos: `git-workflow` paso 9 y `flow-deploy` paso 8 verifican y entregan lo observado a `flow-close`; `flow-close` paso 6 escribe el único reporte con la fila que aplica y el paso 8 termina como dice la fila de `close-modes`; `unattended-delegation` paso 9 resume la sesión y enlaza el reporte de cada cierre. La parada de `flow-deploy` paso 6 que espera una respuesta de recuperación es un mensaje de bloqueo (`hive-rule-reasons-and-blockers`).

Índice: este plan solo corrige la línea de `close-hygiene` y su hoja, que no deben incluir el paso 6 de `flow-close`. La prueba literal del índice (antes D3 A) la construye el plan `hive-core-stage-packets` (D7 A), contra la primera línea «Read» de cada hoja.

## Impacto

40 archivos fuera del archivo histórico enlazan alguna de las 5 hojas; 11 enlazan las dos que se borran y pasan a `report-endings` o `close-report` con ancla de sección. `tests/pilot/session_audit_endings.go` y sus fixtures citan `report-endings`, que conserva su nombre. La regla de la spec «every other file SHALL point to its row» admite una excepción: el invariante mínimo que `global.md` «Questions and routes» debe conservar para actuar sin leer la hoja (`AGENTS.md` «Guidance structure»). Nada toca producción ni datos.

## Recuperación

Descartar el árbol de trabajo de la rama local (`hold`) restaura todo; si el cambio ya llegó a `development`, revertir su commit y `hive update` desde la base revertida reinstala la guía anterior.
