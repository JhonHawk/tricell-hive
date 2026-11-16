# question-card-alone — Las decisiones van solo en la tarjeta; el texto, cuando la tarjeta vuelve sin respuesta

| Campo | Valor |
| --- | --- |
| Estado | In build |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `c35c3424` · ramas propuestas: `chore/trial-refresh-line` (una línea de `AGENTS.md`, se integra antes de la prueba) y `feat/question-card-alone` (la regla, sin integrar hasta cerrar la prueba) |

## Decisiones

2026-11-16
- **Opción A de la investigación** — ¿Cómo se quita la doble presentación de las preguntas? Las decisiones de una ronda van solo en la tarjeta nativa; el texto completo de la ronda se escribe únicamente cuando la tarjeta vuelve sin respuesta. Se prueba en una rama sin integrar a `development`.
- **D1 A** — ¿Cómo corre la prueba sin integrar? La rama se despliega en la instalación local de los siete hosts; mientras dure, cada refresco de la instalación sale de la rama de prueba puesta al día con `development`, y una línea en el `HIVE.local.md` del repositorio avisa a las otras sesiones.
- **D2 A** — ¿Qué queda en el texto antes de la tarjeta? Solo la línea de contexto y el informe o la explicación; ninguna línea por decisión. Para revisar una decisión ya contestada el usuario pide sus opciones.
- **D3 A** — ¿Qué resultado da por buena la prueba? Al menos 20 rondas en Claude Code y 10 en Codex, que incluyan 3 tarjetas sin contestar, sin ninguna ronda perdida. Si Codex falla dos veces, ese host vuelve a la regla actual.
- **D4 A** — ¿Cómo se entrega la rama de prueba? Modo `ready`: revisión, push y pull request abierto sin integrar; el usuario integra cuando la prueba cumpla D3 A.
- **D5 A** — ¿Qué revisión de código corre antes de desplegar y publicar? La nativa de Claude Code, `/code-review`.
- **D6 A** — ¿Cómo se evita que los refrescos de otras sesiones pisen la regla de prueba? Una línea temporal en `AGENTS.md` «Refreshing a local installation», integrada a `development` antes de la prueba: mientras la rama de prueba exista sin integrar, cada refresco despliega la unión de esa rama con `development`, hecha en un worktree temporal que no toca la rama ni su checkout. El último commit de la rama de prueba antes de integrarla borra la línea; mientras tanto la rama la conserva, porque las sesiones leen `AGENTS.md` del checkout en que trabajan. **Replaces:** la línea de aviso en `HIVE.local.md` que D1 A preveía.
- **D7 A** — ¿La prueba debe ver tarjetas sin contestar en Codex? Sí: al menos 2 de las 3 tarjetas sin contestar son de Codex.

Supuestos tomados sin preguntar, dichos al usuario junto a D1–D3:
- La tarjeta va sola únicamente donde la herramienta espera la respuesta y cada opción admite descripción; donde no la admite (Cursor) o no hay herramienta que espere, la ronda sigue como hoy.
- Cuando el host corta el turno al descartarse la tarjeta, devuelva o no un resultado (Esc en Claude Code, Codex y OpenCode), el agente escribe la ronda en respuesta al siguiente mensaje del usuario, salvo que ese mensaje ya la conteste.
- El mensaje anterior a la tarjeta conserva su línea final de espera, que dice qué se espera y cuántas decisiones quedan después; la indicación de que «ok» aplica las opciones con estrella pasa al texto que se escribe tras una tarjeta sin contestar.
- Los scripts de medición pasan de la carpeta temporal de la sesión a `_support/workspace/2026-11-16-question-card-alone/measure/`.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Petición

Del usuario, 2026-11-16: «Analiza las sesiones de estos dias en los diferentes hosts, un punto relevante que quiero revisar es: el hecho de uqe se presentan las preguntas dos veces; noto que el tiempo de espera para renderizar estas se incrementa; se incrementa el contexto; es más complicado leer todas las preguntas en una lista enorme, prefiero la herramienta, pero talvez para codex despues de no contestarse las preguntas si deban renderizarse, pero debemos validar como funciona el hots para este momento».

Tras los hallazgos: «A, pero posiblemente lo probemos en una rama sin mergear a development hasta que sepamos que funciona».

## Objetivo

Hoy cada ronda de decisiones se escribe completa en el texto y después se repite en la tarjeta (`content/guidance/global.md:89-91`). Cuando el cambio esté hecho, en los hosts cuya tarjeta espera y lleva descripciones la ronda se lee una sola vez, en la tarjeta; el texto de las decisiones aparece solo si la tarjeta vuelve sin respuesta. La regla corre primero desde una rama desplegada localmente y se integra cuando la prueba cumple D3 A.

## Alcance y aceptación

Dentro: la regla de «Questions and routes» en la guía global (tres puntos), los punteros que la citan en `flow-plan`, `flow-verify` y su referencia del visor, la página de límites por host, la tabla por host de `agent-delivery.md`, la fila de la rúbrica de revisión de sesiones, el delta de la spec `workflow-lifecycle`, la línea temporal de refresco en el `AGENTS.md` raíz, el despliegue local de la rama y la medición de la prueba.

Fuera: el formato de una decisión (título, `Situación:`, opciones, `Recomiendo`), que no cambia y es el que usa el texto cuando se escribe; las tarjetas consecutivas de una ronda; el bloque de siguiente paso entre flujos; las preguntas abiertas de research; los detectores de `tests/pilot` (el criterio S1 sigue pidiendo texto antes de la tarjeta y el informe lo cumple); la integración a `development`, que decide el usuario al cerrar la prueba.

- AC1 — En un host cuya herramienta de preguntas espera la respuesta y da descripción a cada opción, una ronda se pregunta por la tarjeta y el texto anterior lleva la línea de contexto, el informe y la línea de espera, sin ningún bloque de decisión (línea de título con ID, `Situación:`, líneas de opción) · falso hoy porque `content/guidance/global.md:89` exige escribir cada decisión en el texto antes de la tarjeta y 252 de 306 rondas de Claude Code del 9 al 16 de noviembre lo hicieron · Observed in: las sesiones reales de Claude Code y Codex con la rama de prueba desplegada, leídas en sus transcripts
- AC2 — Cuando una tarjeta vuelve sin respuesta (vacía, descartada, rechazada o con petición de conversar), el siguiente mensaje del agente escribe en formato completo cada decisión de la ronda que sigue abierta y no pregunta más tarjetas de esa ronda; cuando el host cortó el turno, lo hace en respuesta al siguiente mensaje del usuario salvo que ese mensaje ya la conteste · falso hoy porque `content/guidance/global.md:91` manda añadir una sola línea sin repetir las decisiones · Observed in: las sesiones reales con la rama de prueba desplegada, en cada tarjeta sin contestar
- AC3 — **Retired:** D7 A exige que 2 de las tarjetas sin contestar sean de Codex; lo sustituye AC6. La prueba reúne al menos 20 rondas en Claude Code y 10 en Codex, que incluyen al menos 3 tarjetas sin contestar, con cero rondas perdidas: ninguna tarjeta sin contestar se queda sin su texto después y en ninguna ronda el usuario pide el contexto que faltaba para contestar · falso hoy porque ninguna sesión ha corrido la regla nueva · Observed in: los transcripts de Claude Code (`~/.claude/projects`) y los rollouts de Codex (`~/.codex/sessions`) posteriores al despliegue de la rama, medidos con los scripts de `_support/workspace/2026-11-16-question-card-alone/measure/`
- AC4 — La instalación local de cada host registrado sirve la regla de la rama de prueba, y la sigue sirviendo después de que otra sesión integre un cambio en `development` y refresque la instalación · falso hoy porque `~/.claude/CLAUDE.md` y `~/.codex/AGENTS.md` contienen «Write the message as visible text first» · Observed in: los archivos de guía instalados de cada host en esta máquina
- AC6 — **Replaces:** AC3. La prueba reúne al menos 20 rondas en Claude Code y 10 en Codex, que incluyen al menos 3 tarjetas sin contestar, 2 de ellas de Codex, con cero rondas perdidas: ninguna tarjeta sin contestar se queda sin su texto después y en ninguna ronda el usuario pide el contexto que faltaba para contestar · falso hoy porque ninguna sesión ha corrido la regla nueva · Observed in: los transcripts de Claude Code (`~/.claude/projects`) y los rollouts de Codex (`~/.codex/sessions`) posteriores al despliegue de la rama, medidos con los scripts de `_support/workspace/2026-11-16-question-card-alone/measure/`
- AC5 — La página de límites por host y la arquitectura describen el orden nuevo: tras una tarjeta que se cierra, el agente escribe las decisiones en el chat · falso hoy porque `site/src/content/docs/host-limits.md:63` y `:156` dicen que las decisiones ya están en el mensaje de arriba, y `agent-delivery.md:253-255` dice «Full options in the text and in the card» · Observed in: la página `host-limits` del sitio construido localmente desde la rama

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

- Regla vigente: `content/guidance/global.md:89-91`, leída el 2026-11-16; su tamaño se mide en T2.
- Historia: del 2026-10-11 al 2026-11-04 las opciones iban solo en la tarjeta y el texto llevaba una línea resumen por pregunta; el commit `4a6b0ea6` lo reemplazó porque Claude Code `3cac4995`, Grok `01a1a4f3` y Codex `01a1a844` escribieron solo prosa y la ronda quedó invisible donde el host ocultó la tarjeta (`_support/docs/architecture/agent-delivery.md:228`).
- Medición del 2026-11-09 al 2026-11-16 (scripts en la carpeta temporal de la sesión `5aa0de41`):
  - Claude Code 2.1.325: 306 rondas en 49 sesiones; la tarjeta mide el 88 % del texto de decisiones (1.186 contra 1.346 caracteres de mediana); 325 de 330 tarjetas contestadas por la tarjeta, 5 rechazadas con Esc, ningún cierre por tiempo; 54 rondas ya salieron solo en tarjeta y 53 se contestaron por ella.
  - Codex 0.167–0.169, modo Default: 46 rondas en 8 sesiones; 10 de 47 tarjetas se cerraron a los 120,0 s con `{"answers":{}}`.
  - OpenCode 2.3.0: 16 rondas, 2 descartadas con error «The user dismissed this question». Grok 1.0.87: 10 rondas. Cursor 2026.11.13: 2 rondas. Pi y Antigravity: ninguna.
- Codex `rust-v0.169.0`, leído en su código: 60 s ocultos más 60 s de cuenta atrás, sin clave de configuración (`tui/src/bottom_pane/request_user_input/mod.rs:73-74`); solo el modo Plan bloquea (`handlers/request_user_input.rs:83`); tras el cierre queda «Questions 0/N answered» con los enunciados y sin opciones (`tui/src/history_cell/request_user_input.rs`); la plantilla del modo Default dice «If `request_user_input` returns no answers, continue with best judgment instead of asking again» y «Never write a multiple choice question as a textual assistant message» (`collaboration-mode-templates/templates/default.md:15,19`).
- `hive update` acepta `--source DIR --rev REV` y despliega a todos los hosts registrados; no tiene opción por host (`hive update --help`, 2026-11-16).
- Ningún detector de `tests/pilot` comprueba líneas de decisión en el texto; S1 (`question_after_detail`) pide 40 caracteres de texto del asistente antes de la tarjeta (`tests/pilot/regression.go:65-127`).
- Ningún requisito de `_support/openspec/specs` fija hoy el orden texto y tarjeta; «Decisions carry labelled context» (`workflow-lifecycle/spec.md:238`) fija solo el formato.
- No verificado: qué deja Claude Code 2.1.325 en pantalla tras un Esc o tras contestar; si la herramienta `ask_question` de Antigravity 1.5.9 espera la respuesta.

## Effects

None: the tasks take no effect on an external system.

## Entrega

Modo `ready` (D4 A): tras la revisión de código, push de `feat/question-card-alone` y pull request no borrador contra `development`; se espera CI y se responden sus revisiones; el pull request queda abierto durante la prueba y solo el usuario lo integra. Revisión de código: nativa, `/code-review` de Claude Code (D5 A), sobre el candidato antes del despliegue local y del push.

Esta elección no autoriza integrar, activar auto-merge ni desplegar a otra máquina. El despliegue a la instalación local de esta máquina lo autoriza D1 A.

## Estado de la revisión y progreso

2026-11-16 — Plan escrito.
2026-11-16 — Revisión del plan (`hive-review-plan`): 7 hallazgos. Corregidos en la carpeta: la línea de espera se conserva antes de la tarjeta, las búsquedas de T4, el comando de T5, la línea base de T1 y las vigilancias en T6. Dos fueron al usuario: D6 A (línea temporal en `AGENTS.md`, nueva tarea T7, sustituye el aviso en `HIVE.local.md`) y D7 A (AC6 sustituye a AC3). Invalidó: el texto de T5 y el `Closes:` de T1 y T6.
2026-11-16 — Segunda revisión del plan: 5 hallazgos, todos corregidos en la carpeta. El refresco de T7 une la rama con `development` en un worktree temporal en vez de escribir en la rama, y vale mientras la rama exista sin integrar; T3 cubre también la lista de reglas de la referencia del visor; T2 y T5 comprueban que la regla nueva está; la recuperación ya no nombra `HIVE.local.md`; Pi queda dicho como supuesto.
2026-11-16 — Build: T1, T7, T2, T3 y T4 hechas. Revisión de código (`/code-review`, nativa) sobre `6f9acaf5`: 10 hallazgos; 9 confirmados y corregidos en una ronda (Esc en Claude Code corta el turno con resultado; los hosts con la ronda ya en el texto reciben una línea, no la ronda otra vez; los hallazgos `H<n>` vuelven a las reglas del texto; la tarjeta lleva el ID y el título; un descarte devuelto como error no cuenta como herramienta que falla; tras una petición de conversar se contesta antes de escribir la ronda; la referencia de `harness-audit`; la rúbrica; este registro). Uno refutado: «10 de 47» en Codex es correcto, y la nota de T1 que lo contradecía era el error. La línea temporal de `AGENTS.md` se borra en el último commit antes de integrar, no al desplegar.
2026-11-16 — Segunda revisión de código, limitada a las correcciones (una pasada a esfuerzo bajo no leyó el texto y no cuenta; la válida fue a esfuerzo medio): 10 hallazgos, todos corregidos sin nueva revisión. Un «ok» solo aplica decisiones que el usuario vio; una tarjeta contestada en parte cuenta como sin contestar; las reglas de tarjeta valen también donde la ronda va antes en el texto; tras un turno cortado el agente contesta el mensaje del usuario antes de escribir la ronda; la frase inicial de la regla ya no coincide con el marcador de la regla anterior que usan T2, T5 y AC4. `global.md`: 52386 -> 54293 bytes.
2026-11-16 — Entrega en modo `ready`: pull request #491 (`feat/question-card-alone`, la regla, sin integrar hasta cerrar la prueba) y #492 (`chore/trial-refresh-line`, la línea temporal, que integra el usuario antes de T5). Revisión de código de la línea temporal: 9 hallazgos en cada una de sus dos pasadas; el procedimiento se rehízo sin worktree (`git merge-tree`, `git commit-tree`, `refs/hive/trial-join`) y las correcciones de la segunda pasada no se revisaron de nuevo. Queda abierto para el cierre: la unión se despliega sin correr las pruebas de contenido sobre el árbol combinado.
