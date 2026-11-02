# #319 — Visor: header compacto, sin legacy, preguntar antes de abrir y un trabajo por sesión del visor

| Campo | Valor |
| --- | --- |
| Estado | En build |
| Tracker | [JhonHawk/tricell-hive#319](https://github.com/JhonHawk/tricell-hive/issues/319) · abierto · incluye el hallazgo H15 del cierre de los hallazgos de #316 |
| Git | `tricell-hive`, base `development` en `c72549e4`; rama propuesta `feat/gh-319-viewer-cleanup` en `.claude/worktrees/gh-319-viewer-cleanup` |

## Decisiones

2026-11-02
- ¿Los tres cambios van en un solo ticket? Sí, en #319. (D5 B)
- ¿Se quita todo el legacy del walkthrough? Sí, también el código que solo lee o sirve datos viejos: «nos ahorramos mantenimiento (y de todas maneras aun no tenemos instalaciones publicas)». (D10)
- ¿Se conservan las guardas de corte (estado anterior a v3, `/api/sessions/*`, `mode:"legacy"`)? No, se borran las tres. (D11 B)
- ¿Sale el flag obsoleto `--ttl` de `hive view start/serve`? Sí, en este cambio. (D12 A)
- ¿Entra el hallazgo H15, que un segundo trabajo en la misma sesión nativa hereda la sesión del visor del primero? Sí, en este plan. (D13)
- ¿Cómo distingue el visor dos trabajos de una misma sesión nativa? Con una clave de trabajo opcional: con ticket, su clave; sin ticket, el servidor genera una identidad nueva en el momento. (D14, respuesta libre)
- ¿Dónde queda el número de validaciones al quitar la fecha? Como un contador más de la barra de herramientas. (D15 A)
- ¿En qué momentos se pregunta si abrir el visor? En cada parada donde el visor guarda algo que revisar: el plan en su aprobación, el reporte de una investigación y el walkthrough de una UI. (D16 A)
- ¿Cómo se entrega? En modo interactivo: parada para validar el candidato antes del push; luego push, pull request a `development` y merge. (D17 A)
- ¿Qué revisión de código corre? La nativa de Claude Code, `/code-review`, sobre el candidato antes de la validación. (D18 A)
- ¿Entra en #319 que el agente publique las tareas del plan en el visor, cosa que hoy no hace? Sí, dentro de la tarea de guidance. (D19 A)
- El texto bajo la fecha es el resumen de la sesión; pasa a ser la descripción del header sin cambiar lo que es (mismo contenido, mismo corte a dos líneas con «ver más»). (aclaración del usuario)

## Petición

Del ticket (leído el 2026-11-02):

1. «Session header redesign. Remove the date and the project/workspace line shown above the title. New layout: `[Title] [edit]` / `[Description] [Toolbar, right-aligned]`.»
2. «Remove the legacy walkthrough fallback. Drop everything that refers to the legacy `walkthrough.json` manifest and the standalone HTML fallback. Publication works only through real-time interaction with the CLI (structures, states), which keeps it deterministic; there is no HTML plan B.»
3. «Ask before opening the viewer. The viewer no longer opens on its own. The agent asks whether to open it; once the user accepts, it does not ask again in that session; if the user declines, it asks again at each point where there is something to review.»

Y, por D13: «apliquemos con el H15 para incluirlo en el plan».

## Objetivo

Hoy el header de una sesión muestra el proyecto, el título, la fecha y las validaciones, con el resumen suelto al inicio del canvas. El repositorio arrastra un manifiesto `walkthrough.json`, un renderer HTML y unos 10 archivos Go de runtime legacy que ya no corren, con guardas y un flag obsoleto. La guidance abre el visor sola en la primera presentación de UI. Y el visor no puede separar dos trabajos de una misma sesión de Claude Code. Al terminar: el header es `[Título] [editar]` / `[Resumen] [Toolbar →]`, no queda rastro del legacy ni de `--ttl`, el agente pregunta antes de abrir el visor, y cada trabajo tiene su propia sesión del visor.

## Alcance y aceptación

Dentro: el header de sesión y el canvas del frontend; todo el runtime, assets, tests y documentación del walkthrough legacy y del HTML standalone; las tres guardas de corte; el flag `--ttl`; la identidad de trabajo en `create_session`, el registro guardado y los validadores Go/TS; la regla de apertura y de clave de trabajo en `flow-walkthrough`; la spec `walkthrough-viewer`; README, sitio y docs de arquitectura.

Fuera: la lista lateral de la biblioteca (sigue mostrando la fecha); los campos de salida `ttl`/`ttl_seconds` del estado y del índice; cualquier otro cambio de UI; el CHANGELOG histórico y los change records archivados.

- AC1 — El header de una sesión no muestra la línea de proyecto ni la fecha; su primera fila es el título con su botón de editar, y la segunda es el resumen de la sesión (corte a 2 líneas con «ver más») con la barra de herramientas alineada a la derecha desde `lg`; el resumen ya no aparece al inicio del canvas · falso hoy porque `timeline-header.tsx:177-179` pinta `project_name`, `:249-259` pinta la fecha y `timeline-content.tsx:590` pinta el resumen en el canvas.
- AC2 — La barra de herramientas muestra el número de validaciones como un contador más, con el mismo botón `outline` `sm` que tickets, PRs y participantes · falso hoy porque el número va en la línea de la fecha (`timeline-header.tsx:253-258`).
- AC3 — Ningún código de `tooling/` lee, sirve ni embebe el manifiesto legacy o el HTML standalone: `rg -l -i "walkthrough\.json|walkthrough\.(mjs|css)|SourceTracking|session_groups|indexSession"` sobre `tooling/` (sin `static/app`) no devuelve nada, y `hive view status --dir X` falla como flag desconocido · falso hoy: devuelve unos 25 archivos y `--dir` se acepta.
- AC4 — Ningún contenido distribuido, test ni documento vigente menciona el manifiesto ni el fallback: `rg -l -i "walkthrough\.json|walkthrough-manifest|render_walkthrough|standalone HTML|legacy walkthrough"` sobre `content/ site/ tests/ README.md llms.txt _support/docs/ _support/openspec/specs/` no devuelve nada · falso hoy: devuelve `flow-walkthrough`, `flow-close`, `artifact-layout.md`, `viewer-publication.md`, `tests/skills/walkthrough_*`, README, `deployment-manager.md` y la spec.
- AC5 — Sin guardas de corte: `GET /api/sessions/<x>` con token responde 404; una entrada del índice con `mode:"legacy"` falla con `unsupported_mode` en Go y en TypeScript; ni el runtime ni el contrato offline conocen `cutoff_required`, y un `--state-dir` con nombres legacy, sin marcador o con `storage_version` 1 o 2 falla con el `invalid_storage` genérico antes de escribir · falso hoy: 409 `unsupported_protocol` (`server.go:172`), `unsupported_protocol` en el índice, y `cutoff_required` en `state_v3.go:48-71` y `timeline_protocol.go:87,105`.
- AC6 — `hive view start --ttl 1h` y `hive view serve --ttl 1h` fallan como flag desconocido · falso hoy: se aceptan con un aviso (`view_command.go:135,164`).
- AC7 — Desde una misma sesión nativa: dos `create_session` con `work_key` distintas dan dos `viewer_session_id` distintos; repetir una `work_key` devuelve la misma sesión con `created:false`; un `create_session` sin `work_key` con un `operation_id` nuevo crea un trabajo nuevo, y repetir la misma petición devuelve ese mismo trabajo; las sesiones guardadas antes del cambio cargan con su mismo ID · falso hoy: todas las peticiones de una sesión nativa dan el mismo ID (`timeline_store.go:55,616`).
- AC8 — La guidance distribuida no abre el visor por su cuenta: `viewer-publication.md` manda preguntar «¿Abrir el visor?» en cada parada con algo que revisar (plan en su aprobación, reporte de research, walkthrough de UI) hasta que el usuario acepte una vez en la conversación, como una pregunta más de la tarjeta de esa parada o una línea del bloque de siguiente paso; y manda enviar la clave del ticket como `work_key` · falso hoy porque `viewer-publication.md:45` abre la sesión en la primera presentación de UI y no menciona `work_key`.
- AC9 — La guidance distribuida dice cuándo publicar las tareas del plan en el visor: al entrar a build y antes de la primera tarea, registrarlas una vez y cambiar el estado de cada segmento al empezarlo y al terminarlo, en todo host, tenga o no lista nativa de tareas, y en los mismos momentos en que `flow-build` actualiza esa lista donde exista, que no la reemplaza; el párrafo de `flow-build` sobre la lista nativa apunta a ella con una frase aparte, fuera de su condición · falso hoy porque `viewer-publication.md:21` dice cómo pero no cuándo y ninguna skill de flujo apunta a ella, y el agente lleva las tareas solo en la lista nativa (observado por el usuario en COM-907: T1–T7 en la lista de Claude Code y «Sin tareas publicadas» en el visor); las sesiones reales `32cffc1d` (COM-907), `98476eca` (#316) y `8d09a5eb` (COM-908), de Claude Code del 2026-11-02 y posteriores al último cambio de esa referencia (`d7818c88`), llegaron a build con 0 tareas y sin ningún recibo de `task_create`, mientras la API registra una tarea al primer intento en un visor desechable.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- 2026-11-02, `development` en `cf5d77f5`.
- El runtime ya es solo v3: `Store.Sync` devuelve `unsupported_protocol` (`scan.go:343`); `ValidateStateOptions` corta el estado pre-v3 antes de escribir (`state_v3.go:36-71`); `/api/sessions/*` responde 409 (`server.go:172`). La remoción quedó a medias desde la T4 de #306 (`timeline_contract.go:314`; `_support/openspec/changes/archive/2026-11-01-gh-306-viewer-legibility/tasks.md:44`).
- Código solo legacy (mapa de un `hive-research` y de `hive-review-plan`, contrastado en `scan.go`, `state_v3.go`, `server.go`, `view_command.go`): `source_tracking.go`, `source_fingerprint.go` y la parte legacy de `scan.go`, `session_groups.go`, `session_delete.go`, `manifest.go`, `index.go` y `store.go` (lo que siguen usando v3 se conserva o se mueve, ver `design.md`), `static/walkthrough.{mjs,css}`, `assets_test.go`, `timelineLegacyHidden` y `timelineOwns` en `timeline_store.go`, `tooling/cli/view_source_tracking.go`, el flag `status --dir`; en el frontend `document.tsx`, `deliveries.tsx`, `sessions.ts`, los tipos legacy de `api.ts`, `shell.tsx` y `controller.ts`, y `vite.config.ts:36-37` y `scripts/preview.mjs:21`. Se conservan `registry.go`, `ResolveRoot`, `/files/`, `hardlink.go`, `lock_unix.go`, `bundle.go` y los helpers de `manifest.go` que usa `timeline_*` (`contains`, `safeDeliveryURL`, `segmentPattern`, `maxCapture`).
- El instalador no envía ni valida los assets legacy: las skills se copian de forma genérica.
- El estado real del mantenedor es v3 puro (`storage.json`, `timeline-api-v3`), sin nombres legacy.
- La identidad de sesión es `sha256(["timeline-api-v3", proyecto, host, sesión nativa])` (`timeline_store.go:55-63`), y la carga re-deriva y rechaza cualquier registro cuyo ID no coincida (`:218`).
- El header es `TimelineHeader` (`timeline-header.tsx:170-260`); el resumen lo pinta `TimelineSummary` con `lines={2}` en `timeline-content.tsx:590`. El header de `shell.tsx:619-694` solo sirve al modo legacy.
- La regla de apertura vive solo en `viewer-publication.md:45`; el sitio solo documenta `hive view open`.

## Entrega

Modo `interactive`, solo para este trabajo: tras la suite completa, `/code-review` y el recorrido de UI con `hive-review-ux`, parada para que el usuario valide el candidato en un visor de prueba; con su aprobación, push de `feat/gh-319-viewer-cleanup`, pull request a `development` y merge. Durante el build, parada de primera vista del header tras T3. Antes del push se compara la rama con su base y se pregunta por cualquier commit ajeno. No autoriza promover a `master`, publicar una release ni ejecutar `hive update` fuera del refresco local del mantenedor.

## Estado de revisión y avance

2026-11-02 — Plan escrito.
2026-11-02 — Revisión del plan (`hive-review-plan`): 3 bloqueantes, 5 importantes y 4 menores, todos aplicados. La quita de guardas deja fallar el estado viejo con `invalid_storage` en lugar de arrancar, porque tolerarlo debilitaba la detección de estado corrupto; los archivos legacy que aún tienen símbolos de v3 se recortan en lugar de borrarse; el retiro del contrato legacy (Go, TypeScript y fixtures) pasa a T1 para que T2 y T3 no compartan archivos; se modifica también el requisito «Self-describing protocol»; `work_key` en la respuesta es `null` sin clave dada; la regla de apertura gana punteros en las paradas que hacen la pregunta; se suman `index.mdx` y la cláusula completa de `flow-close`.
