# Tasks

Base: tricell-hive `development` en `c35c3424` (2026-11-16). Ramas: `chore/trial-refresh-line` para T7, que el usuario integra antes de T5, y `feat/question-card-alone` para T2–T4, sin integrar hasta que T6 cumpla AC8. Otras sesiones trabajan en el checkout principal sobre `development`, así que el build pregunta antes de su primera edición si la rama vive en un worktree. Todas las tareas corren en el hilo principal: ninguna edita más de 3 archivos; en T6, la lectura de sesiones de un host que pase de 5 lecturas grandes va a un child `hive-research` por host.

### T1 — Los scripts de medición y la línea base quedan guardados

Primera tarea del build: su origen es una carpeta temporal que un reinicio de la máquina puede borrar.

- [x] `_support/workspace/2026-11-16-question-card-alone/measure/` contiene los scripts de Claude Code, Codex y los otros hosts con sus salidas del 9 al 16 de noviembre, y un `README.md` con el comando de cada uno

**Closes:** AC8

**Depends on:** none

**Locations:** `_support/workspace/2026-11-16-question-card-alone/measure/` (ignorado por Git); origen: la carpeta temporal de la sesión `5aa0de41`, subcarpetas `cc/`, `codex/` (sin el clon del código de Codex) y `others/`

**Execution:** main thread

**Test approach:** check

**Changes:** copia los scripts y sus salidas; añade a cada extractor la fecha de corte como parámetro y la clasificación que T6 necesita: ronda con o sin bloque de decisión antes de la tarjeta, tarjeta sin contestar, y si el mensaje siguiente del agente escribe las decisiones abiertas. No cambia las cifras de la línea base.

**Verification:** `python3 extract.py && python3 aggregate.py` en `measure/cc/` reproduce 306 rondas, 330 tarjetas y 252 rondas con bloque de decisión antes de la tarjeta; `python3 cards.py && python3 agg.py` en `measure/codex/` reproduce 47 tarjetas con 10 vacías. AC8 queda sin verificar hasta T6.

### T2 — La guía global pide la tarjeta sola y el texto tras una tarjeta sin respuesta

- [x] Los tres puntos de «Questions and routes» dicen la regla de `design.md` «Proposed shape», sin restos del orden anterior

**Closes:** AC1

**Depends on:** none

**Locations:** `content/guidance/global.md` (puntos «Write the message as visible text first», «Write the card to stand alone» y «Ask in the text only»)

**Execution:** main thread

**Test approach:** check

**Changes:** reescribe los tres puntos con la condición observable (herramienta que espera y descripción por opción), el contenido del texto previo (línea de contexto e informe), la respuesta a una tarjeta sin contestar y al turno cortado, y los casos que siguen en texto. Antes de editar, revisa cada lugar que depende de la conducta, como pide `AGENTS.md` «Guidance structure», y anota los lugares revisados y el cambio de tamaño de `global.md` para el pull request. No cambia el formato de una decisión, las tarjetas consecutivas, las preguntas abiertas ni el bloque de siguiente paso.

**Verification:** `go test ./tests/...` pasa; `rg -n "visible text first|without repeating the decisions|round's text before it" content/guidance/global.md` no devuelve nada, y `rg -c "through that tool alone when it also gives each option a description field|When a card comes back unanswered and it carried the round alone" content/guidance/global.md` da 2; el tamaño de `global.md` antes y después queda anotado. AC1 queda sin verificar hasta T6.

### T3 — Las skills que citan el orden apuntan a la regla nueva

- [x] `flow-plan` paso 4, `flow-verify` paso 2 y la referencia del visor dejan de pedir el texto antes de la tarjeta

**Closes:** AC1

**Depends on:** T2

**Locations:** `content/skills/flow-plan/SKILL.md`, `content/skills/flow-verify/SKILL.md`, `content/skills/flow-verify/references/viewer-publication.md` (el punto «In the card» y la lista de reglas que la referencia dice cambiar, que hoy incluye «a round's text is written in full before its card»)

**Execution:** main thread

**Test approach:** check

**Changes:** cada mención pasa a ser un puntero a «Questions and routes» sin repetir sus pasos; la opción «Open the viewer» se describe como opción de la tarjeta y como última línea de opción cuando la ronda se escribe en texto. No cambia qué se pregunta ni cuándo.

**Verification:** `rg -n -i "visible text before the question tool's card|written in the visible text|written in full before its card" content` no devuelve nada; `go test ./tests/content/...` pasa (tamaño de skills y referencias). AC1 queda sin verificar hasta T6.

### T4 — El sitio, la arquitectura y la rúbrica describen el orden nuevo

- [x] La página de límites por host, `agent-delivery.md` y la rúbrica de revisión de sesiones coinciden con la regla de T2

**Closes:** AC5

**Depends on:** T2

**Locations:** `site/src/content/docs/host-limits.md`, `_support/docs/architecture/agent-delivery.md`, `_support/docs/harness-engineering/session-review-rubric.md`

**Execution:** main thread

**Test approach:** check

**Changes:** en la página, Codex y el paso 3 de la lista final dicen que tras un cierre el agente escribe las decisiones en el chat; en la arquitectura, la regla del 2026-11-04 pasa a historia con las sesiones y cifras que la reemplazan, y la tabla por host cambia «Full options in the text and in the card» por el reparto nuevo, con Codex 0.169.0 revisado; en la rúbrica, la fila de la tarjeta vacía comprueba el texto completo después y se añade la fila de la ronda sin bloque de decisión antes de la tarjeta. `README.md` y `llms.txt` no cambian: no afirman este orden.

**Verification:** `rg -n -i "already in the message above it|also written in the message|full options in the text and in the card|writes each decision's title line and each option line in the visible text|gets one line, not the round repeated" site _support/docs` no devuelve nada fuera de los párrafos marcados como historia; `npm run build` en `site/` termina sin errores y la página `host-limits` construida muestra el texto nuevo.

### T5 — La rama de prueba queda desplegada en la instalación local

- [x] Los archivos de guía instalados de cada host registrado contienen la regla de T2

**Closes:** AC4

**Depends on:** T2, T3, T4, y T7 integrada por el usuario en `development`

**Locations:** la instalación local de Hive (los siete hosts registrados)

**Execution:** main thread

**Test approach:** check

**Changes:** con los commits de la rama hechos y la revisión de código aplicada, nombra los commits aún no desplegados, ejecuta `hive update --source <checkout de la rama> --rev feat/question-card-alone --out FILE` y `hive apply --plan FILE`, y dice al usuario que las sesiones abiertas necesitan reinicio. Antes de desplegar incorpora `development` a la rama, de modo que lleve la línea de T7; el commit que la borra se añade al cerrar la prueba, justo antes de integrar. El binario no se recompila: el cambio es solo de contenido. Los refrescos posteriores de otras sesiones siguen la línea de T7; una sesión que aun así despliegue desde `development` pisa la regla, y T6 lo detecta por las ventanas de despliegue y no cuenta esas rondas.

**Verification:** en el archivo de guía instalado de cada host que `hive status --hosts codex,claude,grok,pi,opencode,cursor,antigravity --scope user` da como instalado (sus rutas las lista `hive doctor`), `rg -q "Write the message as visible text first" <archivo>` termina con código 1 y `rg -q "through that tool alone when it also gives each option a description field" <archivo>` termina con código 0; `hive releases` muestra la instantánea desplegada desde la rama.

### T6 — La prueba se mide y el usuario decide la integración

- [ ] Un informe con las rondas posteriores al despliegue de T8, las tarjetas sin contestar y las rondas perdidas por host, contra AC8

**Closes:** AC1, AC7, AC8, AC4

**Depends on:** T1, T5, T8, y que las sesiones reales reúnan el volumen de AC8

**Locations:** `_support/workspace/2026-11-16-question-card-alone/measure/`; `~/.claude/projects`, `~/.codex/sessions` y los registros de los otros hosts, solo lectura

**Execution:** main thread; un child `hive-research` por host cuando leer sus sesiones pase de 5 lecturas grandes

**Test approach:** check

**Changes:** ejecuta los scripts con el despliegue de T8 como corte, cuenta solo sesiones iniciadas después de él (en Codex y OpenCode, que recargan la guía solas, las rondas posteriores a él), descarta las rondas de sesiones que corrieron con la guía de `development` (ventanas de `hive releases`), lee a ojo cada tarjeta sin contestar y cada ronda marcada como perdida, e informa según la rúbrica de revisión de sesiones; el informe revisa las vigilancias de Engram que el cambio toca (`watch/card-without-visible-text`, `watch/card-without-message`, `watch/no-card-after-card-timeout`) y propone cerrar o redefinir cada una. Si Codex suma dos rondas perdidas, el plan se revisa para dejarlo en la regla anterior mediante una excepción por host. No integra la rama: la integración es la decisión del usuario tras el informe.

**Verification:** el informe muestra al menos 20 rondas en Claude Code y 10 en Codex con al menos 3 tarjetas sin contestar, 2 de ellas de Codex; en cada ronda contada el texto previo no tiene bloque de decisión (AC1); cada tarjeta cerrada sola va seguida del texto completo de sus decisiones abiertas y de ninguna otra tarjeta de esa ronda, y cada tarjeta cancelada por el usuario va seguida de lo que corresponde a lo que el usuario dijo: respuestas por ID aplicadas y el resto preguntado otra vez por la tarjeta, la misma tarjeta, el replanteo, el descarte o el recordatorio (AC7); cero rondas perdidas (AC8); los archivos instalados siguen sirviendo la regla de la rama al medir, después de al menos un refresco hecho por otra sesión tras un merge a `development` (AC4).

### T7 — La línea temporal de refresco está en `development`

- [x] El `AGENTS.md` raíz dice que, mientras la rama `feat/question-card-alone` exista sin integrar, el refresco de la instalación despliega su unión con `development`

**Closes:** AC4

**Depends on:** none

**Locations:** `AGENTS.md` raíz, sección «Refreshing a local installation»

**Execution:** main thread

**Test approach:** check

**Changes:** añade una frase al párrafo del checkout del mantenedor: mientras la rama de prueba `feat/question-card-alone` exista sin integrar, cada refresco construye la unión de esa rama con `development` sin worktree (`git merge-tree --write-tree` y `git commit-tree`), despliega con `--rev` en ese commit y, tras un despliegue correcto, guarda el commit en `refs/hive/trial-join`; no escribe en la rama ni en el remoto. Si la unión falla, no despliega contenido e informa la salida del comando. (Revisado tras la revisión de código, que descartó el worktree temporal.) Rama `chore/trial-refresh-line`, pull request propio en modo `ready`; el usuario lo integra. No toca `content/`, así que no se despliega nada.

**Verification:** `rg -n "feat/question-card-alone" AGENTS.md` muestra la frase en «Refreshing a local installation»; un ensayo de sus comandos (unión con `git merge-tree`, `hive update --rev <unión> --dry-run`) termina con código 0 sin cambios en la rama, y una referencia inexistente hace salir a `git merge-tree` con código 1; `go test ./tests/...` pasa; el pull request tiene CI en verde. AC4 queda sin verificar hasta T6.

### T8 — La regla distingue la tarjeta que se cierra sola de la que cancela el usuario

- [x] El punto «When a card comes back unanswered» de la guía global, sus dependientes y la instalación local siguen D9

**Closes:** AC7, AC8

**Depends on:** T5

**Locations:** `content/guidance/global.md`; `content/skills/harness-audit/references/local-instructions.md`; `site/src/content/docs/host-limits.md`; `_support/docs/architecture/agent-delivery.md`; `_support/docs/harness-engineering/session-review-rubric.md`; la instalación local; `_support/workspace/2026-11-16-question-card-alone/measure/cc/extract.py` y `measure/codex/cards.py`

**Execution:** main thread

**Test approach:** check

**Changes:** reescribe ese punto en dos ramas, «The card closed by itself» y «The user left it unanswered», como fija `design.md` «D9»; alinea la referencia de `harness-audit`, la página de límites por host (sección nueva «If you cancel a card»), la arquitectura y la fila de la rúbrica; añade a los extractores de Claude Code y de Codex, para cada tarjeta sin contestar (en Codex también la que no tiene resultado, que es la cancelada con Esc), si el agente volvió a preguntar por la tarjeta, y el siguiente mensaje del usuario para la lectura a ojo; anota para el pull request el cambio de tamaño de `global.md` y los lugares revisados; pasa `/code-review`, hace commit, push y redespliega con los pasos del párrafo temporal de `AGENTS.md`. Más de 3 archivos, pero en el hilo principal: es una corrección de una frase por archivo sobre texto que esta sesión acaba de escribir.

**Verification:** `go test ./...` y `npm --prefix site run build` terminan con código 0; `rg -c "The card closed by itself:|The user left it unanswered:" content/guidance/global.md` da 2 y la misma búsqueda encuentra ambas frases en la guía instalada de cada host tras el despliegue. AC7 y AC8 quedan sin verificar hasta T6.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` and `npm run build` in `site/` | Covered by the implementation |
| Code review | Final candidate, before the push or the validation stop | Nativa: Claude Code `/code-review` (D5 A) | The review answer |
| Trial | After T8's deploy, in real sessions on this machine | The report of T6 against AC8 | D1 A |

Sin cambio de interfaz: no hay recorrido de UI. La validación humana es la prueba misma: el usuario trabaja en sus sesiones con la regla nueva y decide la integración con el informe de T6.

## Checkpoint

- Updated: 2026-11-16
- Stage: build · T6 (espera el volumen de la prueba)
- Last commit: 6d1c40ed
- Open findings: none
- Next: T6: seguir _support/workspace/2026-11-16-question-card-alone/handoff.md (en el checkout principal, ignorado por Git); corte de la prueba 2026-11-17T00:29:27Z

### Evidence

- go -C <repo> test ./... · exit 0 · 27e752a0 (árbol sin cambios desde la ejecución)
- npm --prefix site run build · exit 0, 19 páginas · 27e752a0
- go test ./tests/... en chore/trial-refresh-line · exit 0 · 27da2b19
- ensayo del refresco: worktree temporal + hive update --dry-run · exit 0, sin worktrees restantes · 27da2b19
- go -C <repo> test ./... · exit 0 · candidato con las correcciones de ambas revisiones
- npm --prefix site run build · exit 0 · mismo candidato
- hive update --source <checkout> --rev aceff0f0 --out FILE + hive apply · exit 0 · unión de 15f74421 y 582997ab, guardada en refs/hive/trial-join
- rg de la regla nueva y de la anterior en los archivos de guía instalados · nueva en 6 archivos (7 hosts), anterior en ninguno
- go -C <repo> test ./... · exit 0 · candidato a1ee85f7
- npm --prefix site run build · exit 0 · candidato a1ee85f7
- hive update --rev 4c5d9bc3 --out FILE + hive apply · exit 0 · unión de ab4db72f y 582997ab, en refs/hive/trial-join
- rg de «The card closed by itself:» y «The user left it unanswered:» en la guía instalada · ambas en los 6 archivos de los 7 hosts; regla anterior en ninguno
- git merge origin/development (007d3999) en la rama + go test ./... + npm --prefix site run build · exit 0 · 6d1c40ed
- hive update --rev fe3c4438 + hive apply · unchanged: la instalación ya tenía ese contenido · refs/hive/trial-join = fe3c4438

## Review status and progress

2026-11-16 — Plan escrito.
2026-11-16 — Revisión tras la revisión del plan: D6 A añade T7 y cambia T5; D7 A sustituye AC3 por AC6 en T1 y T6.
2026-11-16 — T1 · uncommitted · unverified: AC6 hasta T6; la línea base de Codex se reproduce: 63 filas en los archivos, de ellas 47 tarjetas que esperan desde el 2026-11-09, con 10 vacías
2026-11-16 — T7 · 27da2b19 · unverified: AC4 hasta T6; el ensayo de los comandos de refresco con --dry-run terminó sin dejar worktrees; push y pull request pendientes de la revisión de código
2026-11-16 — T2 · 6545eca3 · unverified: AC1 y AC2 hasta T6 (sesiones reales)
2026-11-16 — T3 · 7fab11bf · unverified: AC1 hasta T6
2026-11-16 — T4 · 27e752a0 · unverified: none
2026-11-16 — T5 · 15f74421 · unverified: AC4: la regla nueva está en la guía instalada de los siete hosts (Grok la lee de ~/.claude/CLAUDE.md) y la anterior en ninguno; falta observar que sobrevive a un refresco de otra sesión, en T6
2026-11-16 — Revisión por D9: T8 añadida; T2 deja de cerrar AC2 (retirado); T6 cierra AC7 y AC8.
2026-11-16 — T8 · a1ee85f7 · unverified: AC7 y AC8 hasta T6 (sesiones reales); las correcciones de la segunda pasada de revisión no se revisaron de nuevo
