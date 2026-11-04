# Tareas

Base: `development` en `15863834` al empezar (`b35d7544` tras integrar `development` dos veces), rama `feat/gh-311-minimal-research-entry`. Todas en el hilo principal: son reglas que dependen unas de otras y su redacción se revisa junta; un hijo tardaría más en recibir el contexto que en escribirlas.

### T1 — Guía global: entrada, criterio de plan, preguntas abiertas, fin de un análisis y Context7

- [x] «Proportionality» manda todo pedido de trabajo a `flow-research`, define el cambio pequeño como el que no deja decisión del cómo al usuario y se puede retomar con el pedido y el `git diff`, y quita la lista de revisión como filtro de plan; «Questions and routes» permite la pregunta abierta en texto, termina un análisis en los hallazgos, recomienda flujo solo si se pidió un cambio y deja el comando escrito para research→plan y plan→build; dos líneas de Context7.

**Closes:** AC1, AC3, AC4, AC9

**Depends on:** ninguna

**Locations:** `content/guidance/global.md`

**Changes:** Las secciones nombradas, más «Scope changes» para el nuevo criterio. No cambian autorización, secretos ni entrega.

**Verification:** `grep -n "review list" content/guidance/global.md` no aparece como filtro de plan; las frases de AC1, AC3, AC4 y AC9 están en el archivo; tamaño antes y después anotado.

### T2 — `flow-research` mínimo y rol `hive-research`

- [x] La skill queda en los pasos 0–5 de `design.md`, sin visor, sin oferta de reporte ni ruta obligatoria, con la evaluación de ticket corta; el rol pierde la línea de Context7.

**Closes:** AC2, AC3, AC4, AC6, AC9

**Depends on:** T1

**Locations:** `content/skills/flow-research/SKILL.md`, `content/agents/review/hive-research.md`

**Changes:** La skill se reescribe; el rol cambia una línea. No se crea skill ni rol nuevo.

**Verification:** `wc -c` de la skill entre 4.000 y 4.500 bytes; `grep -n -i "walkthrough\|viewer\|flow-report" content/skills/flow-research/SKILL.md` vacío; `go test ./tests/content/...` pasa.

### T3 — Visor desde plan o build

- [x] `flow-walkthrough` da sesión de visor solo desde plan o build, creada con un resumen de un párrafo; `viewer-publication.md` deja de nombrar el snapshot y la pregunta de abrir el visor de research.

**Closes:** AC5

**Depends on:** ninguna

**Locations:** `content/skills/flow-walkthrough/SKILL.md`, `content/skills/flow-walkthrough/references/viewer-publication.md`

**Changes:** Solo las reglas de cuándo se publica. El protocolo y el código del visor no cambian.

**Verification:** `grep -n -i "research" ` en ambos archivos solo deja menciones que no publican; `go test ./tests/content/...` pasa.

### T4 — Plan con enfoques y build desde research

- [x] `flow-plan` presenta 2–3 enfoques en un mensaje propio antes de las tareas cuando el cómo tiene más de una forma razonable y los registra en `design.md`; `flow-build` entra desde research en el mismo turno cuando el cómo está resuelto.

**Closes:** AC7, AC2

**Depends on:** T1

**Locations:** `content/skills/flow-plan/SKILL.md`, `content/skills/flow-plan/assets/change-folder/design.md`, `content/skills/flow-build/SKILL.md`

**Changes:** Un paso nuevo en plan y la línea de entrada de build.

**Verification:** las frases de AC7 y la entrada de AC2 están en los archivos; `go test ./tests/content/...` pasa.

### T5 — Cierre con pie

- [x] Tras integrar, el ticket se mueve, sale el reporte y termina con el pie; lo destructivo corre al elegir cerrar, al empezar otro trabajo o al terminar la sesión; bajo delegación desatendida corre completo.

**Closes:** AC8

**Depends on:** T1

**Locations:** `content/skills/flow-close/SKILL.md`, `content/skills/flow-verify/SKILL.md`, `content/skills/flow-build/SKILL.md`, `content/skills/flow-deploy/SKILL.md`, `content/skills/git-workflow/SKILL.md`, `content/guidance/global.md` «Preservation», `tests/pilot/regression.go` si su detector exige la tarjeta

**Changes:** El orden de los pasos de `flow-close` y la pregunta de cierre como pie; los llamadores dicen «termina con el pie» en vez de «run flow-close». Seguimiento de PR abierto y `hold` conservan su lógica.

**Verification:** `grep -rn "run \[flow-close\]\|Then run \`flow-close\`" content/skills` no deja una corrida automática fuera de la desatendida; `go test ./...` pasa.

### T6 — Documentación

- [x] El README, `llms.txt`, las páginas de flujos del sitio y la arquitectura describen la entrada por research, el fin de un análisis en los hallazgos y el cierre con pie; entrada en `CHANGELOG.md`.

**Closes:** AC10

**Depends on:** T1, T2, T3, T4, T5

**Locations:** `README.md`, `llms.txt`, `site/src/content/docs/flows/{overview,research,plan,build,close}.md`, `_support/docs/architecture/instruction-resources.md`, `_support/docs/architecture/agent-delivery.md`, `CHANGELOG.md`

**Changes:** Las páginas describen el comportamiento y enlazan la skill; no copian reglas.

**Verification:** el diagrama del README muestra research como entrada; el sitio compila con su comando de build.

### T7 — Barrido de consistencia

- [x] Cada lugar que dice o depende de la entrada, el fin de research o el cierre queda alineado; se anotan los lugares revisados y el cambio de tamaño de `global.md` para el PR.

**Closes:** AC1, AC4, AC8

**Depends on:** T1–T6

**Locations:** `content/`, `site/`, `_support/docs/architecture/`, `tests/pilot/`

**Changes:** Ninguno nuevo; solo correcciones que el barrido encuentre.

**Verification:** `grep -rn -i "small understood\|Next step\|run .flow-close\|review list" content site/src _support/docs/architecture` sin contradicciones; `go test ./...` pasa.

### T8 — La carpeta de cambio se versiona con el código

- [x] `git-workflow` paso 4 tiene la regla canónica: con `Specs` en el mismo repositorio, la carpeta (con su `tasks.md` como esté) se commitea en la rama del trabajo antes del push y, con worktree, se mueve antes a la misma ruta del worktree y se edita solo ahí; con `Specs` en otro repositorio, se commitea allí en el punto de completitud, incluida la parada de `ready`, entregada según su `Delivery` o el modo del trabajo; bajo `hold`, nada. `flow-close` corre ese commit en la primera parte del cierre, como cláusula propia ligada al punto de completitud y no a la integración, solo enlazando la regla sin repetir modos; el paso 4 archiva como hoy. `plan-handoff.md:30` apunta a esa regla.

**Closes:** AC11

**Depends on:** ninguna

**Locations:** `content/skills/git-workflow/SKILL.md`, `content/skills/flow-close/SKILL.md`, `content/skills/flow-plan/references/plan-handoff.md`

**Changes:** Una viñeta en el paso 4 de `git-workflow`; una cláusula en la primera parte de `flow-close`; la línea 30 de `plan-handoff.md`. Sin renumerar pasos: «steps 1, 2, 4 and 5 are pending» sigue valiendo.

**Verification:** `grep -rn -i "uncommitted" content/skills/flow-plan content/skills/flow-close content/skills/git-workflow` no deja la carpeta sin commit como regla; la regla de `git-workflow` nombra el caso del mismo repositorio, el de worktree, el aparte con `ready` y `hold`; `flow-close` la enlaza sin repetir modos; `go test ./tests/content/...` pasa (tamaños).

### T9 — Pendientes en el ticket y en memoria, y su borrado

- [x] `flow-close` paso 7 escribe el registro `pending/<change-id>` (título «Pending: <change-id>», qué falta, por qué, quién y los tickets o condiciones que lo destraban) y el comentario en el ticket cuando algo queda abierto a propósito, también en el cierre de pausa; excluye la promoción por la ruta declarada, los pasos diferidos del cierre en dos partes y la tarea operativa; sin ticket, no hay comentario; sin carpeta, la clave es el `change-id` de `flow-plan`. Cada cierre busca los registros «Pending:» que nombran su ticket o `change-id` (consulta con `project` resuelto como «Persistent memory» y `limit` explícito), los actualiza y borra con borrado suave (`engram delete <id>`) el que queda vacío; el de abandono borra el suyo; sin Engram, borra la entrada del host o dice que no pudo. La guía global autoriza el comentario («Scope and authorization») y el borrado de esos registros («Preservation»).

**Closes:** AC12, AC14

**Depends on:** T8 (mismo archivo, pasos vecinos)

**Locations:** `content/skills/flow-close/SKILL.md`, `content/guidance/global.md`

**Changes:** El paso 7 de `flow-close`, el párrafo de cierre de pausa y el de abandono; una cláusula en cada sección de la guía global. «Persistent memory» no cambia: el registro se escribe y se borra en un cierre, que ya es un momento permitido.

**Verification:** las frases de AC12 y AC14 están en ambos archivos, incluidos el cierre de pausa, el de abandono, la exclusión de promoción, pasos diferidos y tarea operativa, y el fallback sin Engram; tamaño de `global.md` antes y después anotado; `go test ./tests/content/...` pasa.

### T10 — `flow-status` lista los pendientes

- [x] El paso 1 recoge también los registros «Pending:» del proyecto (consulta con `project` y `limit` explícitos) y las carpetas sin versionar bajo `Specs` (`git status --short -- <Specs>/changes` en su repositorio); un registro sale como desactualizado cuando su ticket está en un estado terminal y sus condiciones se cumplen, o su carpeta ya está en `archive/`. El reporte los muestra en el trabajo local con su acción.

**Closes:** AC13

**Depends on:** T9 (formato del registro)

**Locations:** `content/skills/flow-status/SKILL.md`

**Changes:** El paso 1, la línea de qué se lee y el punto 3 del reporte.

**Verification:** las frases de AC13, incluida la condición de desactualizado, están en el archivo; `go test ./tests/content/...` pasa.

### T11 — Documentación y barrido

- [x] Las páginas `plan`, `close` y `status` del sitio y `CHANGELOG.md` describen el comportamiento y enlazan la skill; cada lugar de `content/`, `site/`, `_support/docs/architecture/` y `tests/pilot/` que dice o depende de la carpeta sin commit o de la memoria del cierre queda alineado.

**Closes:** AC15, AC11

**Depends on:** T8, T9, T10

**Locations:** `site/src/content/docs/flows/{plan,close,status}.md` (`plan.md:45`, `status.md:26`), `CHANGELOG.md`, y lo que encuentre el barrido

**Changes:** Las páginas describen; no copian reglas.

**Verification:** `grep -rn -i "uncommitted\|versions it once\|pending/" content site/src _support/docs/architecture tests/pilot` sin contradicciones; `go test ./...` y el build del sitio pasan.

## Verificación y revisión humana

| Chequeo | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar la tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go test ./...` y el build del sitio | Cubierto por la implementación |
| Revisión de código | Candidato final, antes de la validación o el push | Según la respuesta de revisión | La respuesta de revisión |

## Estado de la revisión y avance

2026-11-03 — Plan escrito.
- T1 hecho · c423e5ef · `global.md` 27.172 → 28.152 bytes.
- T2 hecho · 0fb4c949 · `flow-research/SKILL.md` 5.075 → 4.574 bytes (estimado 4.000–4.500; 74 bytes arriba), sin lecturas obligatorias del visor.
- T3 hecho · 20293948.
- T4 hecho · 1f3e3197.
- T5 hecho · 73046b86 · los llamadores (`flow-verify`, `flow-build`, `git-workflow`) no cambian: siguen corriendo `flow-close`, que ahora espera la respuesta del pie. El detector `close_question_after_report` de `tests/pilot/` sigue esperando la tarjeta en Grok y OpenCode; no se cambia mientras los pilotos estén en pausa.
- T7 parcial · 14320b3e · referencias a pasos de `flow-research` en `security-advisory` y `test-audit`.
- T6 hecho · 4d7a8ae4 · por `hive-write-docs`, revisado; build del sitio pasa (17 páginas). `instruction-resources.md:80` es una medición del 2026-10-23 y queda igual.
- T7 hecho · 14320b3e, 4d7a8ae4 · `go test ./...` pasa. Los casos de `tests/fixtures/` no cambian (pilotos en pausa).
- Verify · `/code-review` (high) sobre la rama hasta 4d7a8ae4: 10 hallazgos + 3 menores. Corregidos en 38c2f6d7 y en el delta `specs/walkthrough-viewer`: paso 0 del research saltaba el criterio de plan; entrada de build muy estrecha; «¿se puede…?» contra «Scope and authorization»; el cierre dividido sin reporte de lo diferido ni registro de lo pendiente; frase de excepciones ambigua y `hold`; búsqueda del siguiente ticket en cada cierre; spec del visor; `viewer.md` y `workspace-and-artifacts.md`. No se corrige: los cambios de contrato, datos o seguridad ya no van a plan por sí solos (es D2 A; build conserva `hive-verify-task`, y la revisión se recomienda para esa lista); el detector `close_question_after_report` (pilotos en pausa, decisión del usuario); la línea duplicada de `global.md` (no lo es: una es para preguntas y la otra para trabajo).
- Segunda revisión de 38c2f6d7: 10 hallazgos, corregidos en 78df78d5 (`Unreviewed:`). Memoria escrita al integrar con los pasos pendientes, para otra sesión.
- 2026-11-03 · PR #352 abierto (ready), sin checks de CI ni revisiones, base al día. `hive update` desde 78df78d5 a los 6 hosts; el binario no se reconstruyó (71b9c591 del visor, de #351).
- 2026-11-03 · 175a0263 · D12 (pedido amable con acción y destino cuenta como pedido) y D13 (viñeta propia para la pregunta de intención; la regla de rutas partida en cuatro). Lo integrado en development durante la sesión (#340–#351) ya estaba en la base; binario al día (effe57fe). `hive update` desde 175a0263 a los 6 hosts.
- 2026-11-03 · correcciones tras probar en sesiones reales, todas en el PR #352 y desplegadas: 5a2a8b25 (bloque de análisis con pregunta; documentación en línea cuenta para delegar), 012bc02e (respuestas con letra), 905eb5df (★ en el bloque de análisis, nunca un flujo no pedido), 14f0dd4c (flow-research se lee una vez por research), 56704a0a (ticket solo si la opción elegida lo nombra).
- 2026-11-03 · bd4e3f09 · merge de `origin/development` (#353, publicación bajo demanda) en la rama; conflictos en `viewer-publication.md` y `site/.../viewer.md` resueltos con los comandos `hive view work` de #353 sin la publicación del research. `go test ./...` y el sitio pasan; push y `hive update` desde bd4e3f09 a los 6 hosts; binario en c22ad325 (lo construyó la sesión de Codex). PR #352 MERGEABLE/CLEAN.
- 2026-11-03 · 19805fc3 · un build sin plan publica en el visor con `session create` y `task create` de bajo nivel, porque `work open` pide una carpeta y `work tasks register` un plan publicado. Push y `hive update` desde 19805fc3 a los 6 hosts.
- 2026-11-03 · 0d79e82a (nunca cerrar, reiniciar ni matar un programa compartido; contenedores ajenos con confirmación; D15 A) y 54eeabf5 (research lista aspectos y fuentes antes de leer y delega en la sexta lectura). Push y `hive update` desde 54eeabf5. `global.md` 29.566 bytes. Revisión con rúbrica de 5 sesiones; watch items en Engram.
- 2026-11-03 · 453efaac (D16 A: rondas de entrega y de verify como texto antes de la tarjeta) y 3b6b04c4 (D17 A: procesos de la tarea por PID). Push y `hive update` desde 3b6b04c4. `global.md` 29.661 bytes.
- 2026-11-03 · 90250180 (sesión Codex 01a1a2b3, premisa de @poteto, D1 A de esa sesión): verify reutiliza checks vigentes de build («Check reuse») y los hallazgos de revisión se confirman con reproducción o fuente concluyente («Review evidence»), también al juzgar la salida de un /code-review nativo. Validado aquí: solo esos dos ajustes, conserva D16 y la regla de segunda revisión, `global.md` sin cambio, tests de contenido y sitio pasan. Push y `hive update` desde 90250180.
- 2026-11-03 · descripción del PR #352 actualizada (MERGEABLE/CLEAN); `hive apply` desde 90250180 sin cambios pendientes, binario c22ad325. Siguiente: seguir la sesión Codex cam 01a1a2e6 (COM-916 + COM-921, primera con toda la guía) tras la respuesta del usuario a su D1; pendiente decidir el recordatorio de texto antes de la tarjeta en flow-research (watch #17098, conteo 2).
- 2026-11-03 19:43Z · sesión Codex cam 01a1a2e6 en build (D1 A, D2 interactiva, D3 agente): ronda D2/D3 solo en tarjeta (repite #17098 en la misma sesión; el recordatorio de D16 no cubre la entrega de flow-build sin plan); recomendó el agente y no `/review` nativo sin preferencia del proyecto (watch #17124). El reporte `_support/sessions/2026-10-30-skills-rebuild/skills-minimal.report.html` está desactualizado (estado al 2026-11-01); pendiente la decisión del usuario sobre actualizarlo.
- 2026-11-03 · 7f84af01 (D18 A: la oferta de abrir el visor nunca va como pregunta propia; Codex cam 01a1a2e6 mandó D6 solo en tarjeta). Push y `hive update` desde 7f84af01; 524 archivos instalados y verificados. Siguiente: seguir 01a1a2e6 hasta su validación; pendiente decidir el reporte de skills mínimas y el recordatorio de #17098 en flow-research y la entrega de flow-build.
- 2026-11-03 · 47d8ad84 (research escribe la lista de aspectos, fuentes y delegación en el mensaje antes de la primera lectura; recurrencia de #17110 en Codex 01a1a2b3 y 01a1a30f, aprobada por el usuario). `flow-research/SKILL.md` 5.197 bytes. Push y `hive update` desde 47d8ad84. Siguiente: seguir Codex cam 01a1a2e6 y Codex hive 01a1a30f (toasts del visor, worktree `fix/viewer-archive-toasts`).
- 2026-11-03 22:53Z · Codex cam 01a1a2e6 cerrada (PR #340 integrado, CI posterior verde, tickets en To Deploy, visor cerrado, rama borrada, procesos por PID, Engram al cierre). Casos menores: pie de cierre antes de leer flow-close; cierre final sin línea de espera. Codex hive 01a1a30f: D3 validada (oferta del visor dentro de la ronda, conforme a D18 A), PR #355 integrado, binario reconstruido y visor reiniciado conforme; espera su pie de cierre. Costo del visor (variante Codex): hive 12,3 %/5,2 %, cam 15,5 %/15,0 %; sin señal de baja.
- 2026-11-03 · af227749 (D19 A: tras un reinicio de mantenedor, avisar que las pestañas del visor perdieron acceso y ofrecer *open the viewer*). La rama ya traía `development` (merge 2ff2d43a con #354 y #355, hecho fuera de esta sesión); rebase del commit local encima. Sin cambios en `content/` desde 47d8ad84: no hace falta `hive update`. Pendiente en el issue #356: `hive view restart` que conserve el token (engram decisions/viewer-restart-token).
- 2026-11-03 · Codex hive 01a1a30f cerrada (integró `development` en esta rama a pedido del usuario, 2ff2d43a; limpieza D4). b30c7cf9: una tarjeta cerrada por tiempo sigue siendo herramienta que espera (watch/no-card-after-card-timeout, cam 01a1a2e6 D7 y hive 01a1a30f D4; adoptada por el usuario). `global.md` 29.751 bytes. `hive update` desde b30c7cf9: 524 archivos verificados. Cuerpo del PR #352 actualizado (tamaños y lugares revisados). Issue #356 creado para `hive view restart`.
- 2026-11-03 · d9b51eb9 (research delegar vs inline, opción A del usuario: el límite de `flow-research` paso 4 cuenta solo lecturas grandes; `SKILL.md` 5.578 bytes). Push, `hive update` desde d9b51eb9 (524 verificados), cuerpo del PR actualizado. El research quedó en la conversación (no se guardó).
- 2026-11-03 · Criterio de delegación en todos los flujos (research de costo y sesgo de revisores; plan en la conversación con D20 A referencia propia, D21 A revisor de plan en todo plan con carpeta, D22 A más de 5 lecturas grandes o más de 3 archivos; entrega D23 A `ready`, D24 A `/code-review`, D25 A dentro de #352): aed20048 referencia `flow-build/references/delegation.md`, 8812fe1c «lectura grande» y lista de riesgo en `global.md`, baac0f3b flow-build, e631a085 flow-plan, 98b51e59 revisores, cc083aae flow-deploy y flow-status, 0e4e58a3 punteros, 8937d245 sitio y CHANGELOG. `/code-review` high: 10 hallazgos corregidos en fa2c0fe9; segunda revisión limitada: 9 corregidos en 02adf4f9 (línea `Unreviewed:`), 1 descartado al medirlo (el fork de `/code-review` arranca con ~58k tokens, no hereda la conversación). Push, `hive update` desde 02adf4f9 (530 verificados), cuerpo del PR actualizado. Tamaños: `global.md` 30.227 bytes, `flow-research/SKILL.md` 5.252, `delegation.md` 4.359.
- 2026-11-03 · 5cc5d4aa: `_support/sessions/2026-11-03-delegation-cost/measure-delegation.py` (Codex y Claude Code) y filas de delegación en la rúbrica de sesiones. Validación de documentos sin commitear: esta carpeta reconciliada (base, alcance, AC6, design, delta del visor), cabecera de `viewer-cost.research.md` corregida; `subagent-effort-per-call` y `viewer-cost` son de otro trabajo y quedan locales.
- 2026-11-04 · Tarea agregada (sesión Claude Code `f9f55a91`): T8–T11 y AC11–AC15, D26 A y D27 A, por las cinco carpetas sin versionar de cam-specs (COM-911, 913, 915, 917, 920).
- 2026-11-04 · `hive-review-plan` sobre T8–T11: 10 hallazgos, aplicados: se quita «archivar solo sin pendientes» (1), commit aparte ligado al punto de completitud incluida la parada de `ready` (2), mover la carpeta al worktree (3), cada cierre busca y borra los registros que nombran su ticket (4), consulta y fallback sin Engram (5), la promoción no es pendiente (6), separación de los pasos diferidos y clave sin carpeta (7), condición de desactualizado y borrado al abandonar (8), verificaciones ampliadas (9), `flow-close` solo enlaza (10). Entrega D28 A (`ready` en #352) y revisión D29 A (`/code-review`). Visor sin publicar: esta sesión no es productor aceptado del trabajo #311.
- 2026-11-04 · T8 hecho · 3a95ec70. T9 hecho · 03a8107a · `global.md` 33.680 → 33.953 bytes, `flow-close/SKILL.md` 15.544 bytes. T10 hecho · d22c5d1a. T11 hecho · 2cfe7146 · `go test ./...` pasa salvo `TestExecApplyPiTimesOutAndKillsDescendants` (falló una vez en la corrida completa, pasa 5 de 5 sola; no toca este cambio); sitio: 17 páginas. Visor sin publicar (esta sesión no es productor aceptado).
- Checkpoint: build de T8–T11 hecho; sigue verify con `/code-review` (D29 A) y push al PR #352 (D28 A); al merge: mover #311, archivar esta carpeta (con el delta de `walkthrough-viewer`), limpiar la rama.
