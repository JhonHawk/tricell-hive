# viewer-capture-only — El visor publica solo capturas con su iteración

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build · plan revisado, pendiente de aprobación |
| Tracker | Ninguno (sin ticket; búsqueda del 2026-11-08: solo #375 y #356 tratan del visor, ninguno de este recorte) |
| Git | tricell-hive · base `development` @ `7ae26b0e` · rama propuesta `feat/viewer-capture-only` en `.claude/worktrees/` |

## Decisions

2026-11-08
- **Research A** — ¿Qué hacemos con el costo del visor? Reducirlo a capturas con su iteración; los agentes dejan de declarar etapas, tareas, documentos y contratos.
- **D1 A** — ¿Con qué enfoque? Un comando nuevo `hive view work capture` en el binario, que publica un lote de capturas en una sola llamada, más el recorte de la guía. Las operaciones de etapas, tareas y documentos siguen en el binario, sin uso en la guía.
- **D2 A** — ¿Quién cierra los trabajos de capturas? Un `work close` en el cierre del trabajo (flow-close), como hoy; se conserva la lectura de vuelta que exige `workflow-lifecycle`.
- **D3 A** — ¿Dónde se ve el número de iteración? En el título de la vista que escribe `capture` («· #N»); la interfaz del visor no cambia.
- **D4 A** — ¿Cómo se entrega? `automatic`: revisión, push, PR y merge sin parar.
- **D5 A** — ¿Qué revisión de código? La nativa, Claude Code `/code-review`.

## Request

> «Tendríamos que analizar cuánto consume de contexto el visor actual, y si vale la pena, porque ahorita lo único que realmente he ocupado es el tema de las imágenes y diferencia de qué iteración son» (usuario, 2026-11-08). Tras la medición eligió «A»: reducir el visor a capturas con su iteración, con plan previo.

## Objective

Hoy el visor se lleva ~9 % de los tokens procesados en las sesiones que lo usan (Hive, 24 sesiones tras #374: 7,2 % en llamadas + 1,9 % arrastrado; cam y carvento, 14 sesiones: 4,9 % + 2,8 %). Solo ~7 % de sus llamadas suben capturas; el resto declara etapas y tareas, lee el contrato (8.995 bytes) y relee la sesión. Publicar una captura exige `work open`, `work sync` y `work evidence --candidate`, y la iteración no lleva número.

Al terminar, un flujo sin capturas no toca el visor; un flujo con capturas publica cada lote con una llamada a `hive view work capture`, que numera la iteración, y el cierre hace un `work close`. La referencia de publicación pasa de 14 KB a ≤ 4 KB.

## Scope and acceptance

Dentro: el comando `work capture` (Go, tests, texto del contrato y su ayuda), la guía distribuida que manda publicar en el visor, el delta del spec `walkthrough-viewer`, las páginas del sitio y los documentos de arquitectura que describen la publicación.

Fuera: la interfaz web del visor (D3 A), los hooks de sesión y la captura pasiva, las operaciones existentes de etapas, tareas, documentos, entregas y CI (se quedan en el binario sin cambios), el arranque lento (#375), y la nueva medición, que necesita sesiones reales después del despliegue.

- AC1 — En un repositorio sin trabajo abierto, con el visor corriendo, `hive view work capture --stage verify --view-title Login --image a.png` deja la vista visible en el visor (`hive view session get`), sin `work open`, `work sync` ni `--candidate` previos · falso hoy porque `work capture` no existe y `work evidence` sin trabajo responde `work-id required: this session has no open work`.
- AC2 — Un segundo `capture` con el mismo título base en el mismo trabajo reemplaza la vista anterior (`supersedes`) y se titula «Login · #2»; la primera conserva «Login · #1», y otro título base empieza en «· #1». Vale también cuando las dos capturas se hicieron con el visor detenido y salen de la cola · falso hoy porque ningún título lleva número y reemplazar exige pasar `--supersedes VIEW` a mano.
- AC3 — Con el visor detenido, `capture` responde `queued` sin arrancarlo; cuando el visor vuelve, el siguiente comando o hook envía en orden la apertura, la etapa y la captura, y la vista aparece una sola vez · falso hoy porque el comando no existe.
- AC4 — Repetir el último `capture` de una pantalla con las mismas imágenes, leyendas y candidato responde `unchanged` y no crea otra iteración; con otro commit crea «· #N+1» · falso hoy porque el comando no existe.
- AC5 — En la guía distribuida, las únicas llamadas al visor son `work capture`, `work close`, un `work sync` en flow-close para `in_review`, `paused` o `awaiting_validation`, `hive view open`, los verbos de gestión local que pide el usuario y la referencia de identidad. `rg -n -o 'hive view [a-z]+( [a-z]+)?|work (open|sync|evidence|publish|contract|resume|reconcile|show|list|bind|tasks|task)\b|session get|--tasks|#tasks|#plan-documents|#one-declaration' content/` solo devuelve esos casos · falso hoy porque `viewer-publication.md` manda `work open`, `sync --tasks`, `tasks register`, `publish`, `contract` y `resume`.
- AC6 — Un flujo sin capturas de interfaz no crea trabajo en el visor: `flow-plan` ya no abre sesión del visor ni publica snapshots del plan, y `flow-build` y `flow-verify` solo la abren con la primera captura · falso hoy porque `flow-plan/SKILL.md:8` abre la sesión al entrar al plan.
- AC7 — `content/skills/flow-walkthrough/references/viewer-publication.md` pesa ≤ 6.000 bytes · falso hoy: 14.077 bytes.
- AC8 — `README.md`, `llms.txt`, `site/src/content/docs/viewer.md` y las páginas de flujos, `skills.md` e `index.mdx` describen la publicación solo de capturas · falso hoy porque `viewer.md` tiene «One declaration per stage change», `README.md:480-486` describe declaraciones de etapa y `llms.txt:57` habla de proyecciones de etapas y tareas.
## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Medición del 2026-11-08 con `_support/sessions/2026-11-02-viewer-cost/measure-viewer-cost.py`; números y método en `viewer-cost.research.md` de esa carpeta.
- Precondiciones de `work evidence`: `tooling/cli/view_work_evidence.go:216` y `:311` (`stage_not_started`), `:346` (`candidate_required`); trabajo requerido en `view_identity.go:266`. Iteración: `--supersedes` crea `kind: iteration`; la interfaz muestra «Iteration» y «Superseded by» sin número (`tooling/viewer/frontend/src/timeline-content.tsx:~388-420`).
- Los hooks nunca cambian el estado del trabajo (`tooling/viewer/host_events.go:220-222`); el autoarchivado exige un cierre explícito (`timeline_contract_maintenance.go`).
- `hive view work contract` imprime 8.995 bytes; un agente que solo publica capturas necesita ~2,8 KB.
- Spec afectado: `walkthrough-viewer/spec.md` (requisitos «Proportional identity-aware publication procedure», l. 235; «Declarative stage publication without a change folder», l. 611; «Queued semantic publication», l. 622). `workflow-lifecycle/spec.md:9-21` (cerrar y releer) no cambia.
- El visor local estaba detenido al planear (`hive view status` → `running: false`); este plan no se publicó en él.

## Delivery

- **D4 A** — `automatic`: revisión, push, PR y merge sin parada de validación; el final es el merge en `development`. Después del merge, la marca de mantenedor (`_support/workspace/maintainer.local`) pide refrescar la instalación local: reconstruir el binario y desplegar el contenido como manda AGENTS.md.
- **D5 A** — Revisión nativa: Claude Code `/code-review` sobre el candidato, antes del push.
- No autoriza: publicar una versión (release), cambiar la configuración global fuera del refresco de la instalación local, ni reiniciar un visor en ejecución salvo el caso que AGENTS.md permite.

## Review status and progress

2026-11-08 — Plan escrito. Revisión `hive-review-plan`: 5 bloqueantes y 6 no bloqueantes, todos aplicados; el detalle está en `tasks.md`.
