# viewer-passive-capture — Visor más estable y casi sin costo de contexto

| Campo | Valor |
| --- | --- |
| Estado | Draft · plan escrito, pendiente de revisión y aprobación |
| Tracker | Sin ticket propio. Resuelve #358 (transiciones de tareas fallan en builds de tres hosts) en la vía semántica; T3 comprueba si el bajo nivel también lo sufre. Relacionado: #356 (reinicio sin perder pestañas, fuera de alcance) |
| Git | JhonHawk/tricell-hive · base `development` en `d23f9a88` · rama propuesta `feat/viewer-passive-capture` |

## Decisiones

2026-11-04
- **D4 A** — ¿Cómo llegan los eventos del host al visor? Por hook: cada host llama a `hive hook <evento>`, que los envía a la API del visor sin imprimir nada, con cola local si el visor está caído y conciliación desde los archivos de sesión.
- **D5 A** — ¿Qué puede deducir el visor sin que el agente lo publique? El ciclo de vida de la sesión (inicio, actividad, fin) y las tareas de la lista nativa del host. Etapas, motivos, documentos, capturas y validaciones los sigue publicando el agente, una vez por etapa.
- **D6 A** — ¿Qué pasa con las operaciones de bajo nivel de `hive view`? Siguen funcionando igual; la guía pasa a los comandos nuevos y el contrato las marca como avanzadas.
- **D7 B** — ¿Cómo se entrega? En un solo pull request.
- **D8 A** — ¿Cómo se registra el hook en Claude Code? Con entradas propias en `~/.claude/settings.json`, mediante un editor que solo añade y quita las que llaman a `hive hook`.
- **D9 A** — ¿Quién aprueba el hook en Codex? El usuario, una vez, desde `/hooks`; Hive escribe `~/.codex/hooks.json` y no toca `trusted_hash`.
- **D10 A** — ¿Se activa un plugin en OpenCode? Sí, solo el de eventos, en `~/.config/opencode/plugins/`; el cargador de contexto sigue sin activarse.
- **D11 B** — ¿Entra Cursor? Sí, con un `hooks.json`, aunque los hooks de Cursor CLI no estén verificados.
- **D12 A** — ¿Qué se pone en cola con el visor caído? La declaración semántica (la entrada de `open`, `sync` o `evidence`). Al enviarla, las peticiones se construyen sobre el estado actual, y lo que ya no aplique se informa como rechazado sin sobrescribir.
- **D13 A** — ¿Cómo conviven las tareas del host con las del plan? Si el trabajo tiene tareas registradas, las del host solo suman actividad. Si no tiene ninguna, las tareas nativas pasan a ser las del trabajo, hasta 50.
- **D14** — ¿Cómo se comprueba en real? Primero se eligió `bb1` solo para la instalación. El usuario lo ajustó después: `bb1` queda para temas técnicos, y la comprobación se hace directamente en esta Mac antes del merge. Aquí las seis CLIs ya están instaladas y con sesión iniciada.

## Solicitud

«creo que debemos hacer research profundo de cómo podemos tener el visor mas estable y con mucho menox costo de contexto» y, tras el research, la opción «B — Planificar las opciones 1 y 3 juntas: identidad y ciclo de vida por hooks, con el agente publicando solo lo que nadie escribe a disco».

## Objetivo

Hoy el agente publica en el visor a mano, comando a comando. Cada comando es una llamada al modelo que relee 200–500k de contexto, por lo que el visor consume el 7–11 % de los tokens del hilo principal en Claude Code, Codex y Grok. Además falla entre el 10 y el 24 % de las mutaciones, sobre todo al registrar tareas (#358) y al añadir capturas, y cuando el visor está caído el agente deja de publicar.

Cuando el cambio esté hecho:
- el binario resolverá la identidad de la sesión y el trabajo actual;
- una sola orden por etapa publicará la etapa, sus tareas y el cierre;
- si el visor está caído, la publicación quedará en cola;
- los hooks de cada host informarán del ciclo de vida y de las tareas nativas sin llamadas al modelo.

## Alcance y aceptación

**Dentro:**
- binario `hive`: identidad, trabajo actual, `work` sin carpeta de cambio, orden declarativa por etapa, evidencia, cola local y lock por trabajo;
- servidor del visor y frontend: eventos del host con su procedencia;
- `hive hook` y la conciliación;
- el registro por host en el gestor para Claude Code, Codex, Grok, OpenCode, Pi y Cursor;
- la guía de `flow-walkthrough` y de los flujos que la llaman, los specs y la documentación.

**Fuera:**
- reiniciar sin cambiar el token (#356) y el arranque automático del servicio;
- deducir etapas, motivos o validaciones;
- dar por muerta una sesión sin su evento de fin;
- escribir la confianza de Codex.

**Criterios:**
- AC1 — Un build sin plan publica su etapa, tres tareas con sus estados y el cierre, con la lectura final incluida, con como mucho 3 comandos de `hive view` y sin parche de vinculación · falso hoy porque necesita unas 4N+5 mutaciones y falla con `stage_binding_mismatch` (#358).
- AC2 — `hive view work` sin `--host` ni `--native-session-id` resuelve la identidad real en Claude Code (variable de entorno) y, en un host sin variable, por el mapa pid → sesión que escribe el hook de inicio. Sin fuente, o con una ambigua (dos sesiones bajo el mismo proceso), se rechaza · falso hoy porque esas banderas son obligatorias (`tooling/cli/view_work.go:271`).
- AC3 — `hive hook session-start` registra el mapa pid → sesión e informa al visor de la sesión. No imprime nada y sale con 0 aunque el visor esté caído o la entrada sea inválida · falso hoy porque no existe el subcomando `hook`.
- AC4 — En un build sin tareas registradas, una tarea creada y completada en la lista nativa del host (Claude Code `TaskCreated`/`TaskCompleted`, OpenCode `todowrite`) aparece en el trabajo actual con procedencia `host_observed` sin ninguna llamada del agente. En un trabajo con tareas del plan, no se añade ninguna · falso hoy porque el visor solo acepta orígenes `agent_reported` o `server` (`tooling/viewer/timeline_contract.go:1380`).
- AC5 — Con el visor parado al inicio, `work open` y `work sync` responden `queued`. La siguiente orden con el visor activo crea el trabajo y la etapa una sola vez y en orden, y una declaración que ya no aplica se informa como rechazada · falso hoy porque `openViewWorkEngine` falla antes de registrar nada (`tooling/cli/view_work_engine.go:44`).
- AC6 — Dos trabajos distintos publican a la vez sin `busy` · falso hoy por el lock global `publisher.lock` (`tooling/cli/view_work_journal.go:81`).
- AC7 — Una orden `work` publica capturas y su vista de validación en una sola invocación · falso hoy porque solo existe con operaciones de bajo nivel (`artifact add` y `event add` con `views`).
- AC8 — `hive update` instala, y la desinstalación quita, solo las entradas de Hive en `~/.claude/settings.json`, `~/.codex/hooks.json`, el hook nativo de Grok, el plugin de eventos de OpenCode, la extensión de Pi y `~/.cursor/hooks.json`, dejando intactos los hooks ajenos (herdr, rex, codegraph, Engram) · falso hoy porque el gestor solo registra el hook de Grok (`integrations/grok/hook.go`).
- AC9 — El visor muestra la última actividad del host por trabajo y distingue lo observado del host de lo publicado por el agente, sin cerrar ni pausar trabajos por inactividad · falso hoy porque no existe la procedencia del host.
- AC10 — El procedimiento de `flow-walkthrough` publica una vez por cambio de etapa con las órdenes nuevas, sin operaciones de bajo nivel ni banderas de identidad, y cada archivo queda bajo el límite de 16.000 bytes · falso hoy porque `viewer-publication.md:7-9` exige contrato, identidad y operaciones de bajo nivel para un build sin plan.

## Prerrequisitos

- [x] Archivar `gh-311-minimal-research-entry`: archivada en el PR #370 (`d23f9a88`), con su delta incorporado a la spec base · lo necesitaba: T1 y T8.
- [ ] Aprobar el hook de Hive una vez en Codex (`/hooks`) cuando T10 lo instale en esta Mac · responsable: usuario · lo necesita: T10.

## Contexto verificado

- Research de la sesión `5ff8bac1` del 2026-11-04 (H11–H19):
  - el visor consume el 7–11 % del hilo principal;
  - entre el 10 y el 24 % de las mutaciones fallan;
  - los hooks de Engram funcionan en Claude Code, Codex, OpenCode y Pi, y el hook de Hive en Grok tiene 3.182 ejecuciones correctas;
  - Engram ya escribe `~/.engram/host-sessions/<pid>.json`;
  - `CLAUDE_CODE_SESSION_ID` coincide con el ID de la sesión.
- Mapa de código con base en `1dd67c3b`:
  - `tooling/cli/view_work*.go`, `tooling/viewer/timeline_*.go`, `tooling/viewer/server.go`;
  - `integrations/grok/hook.go`, `tooling/management/{plan,ownership,files}.go`, `tooling/management/codex_wait_floor.go`;
  - `_support/openspec/specs/walkthrough-viewer/spec.md`.
- `gh-311-minimal-research-entry` estaba fusionada (PR #352, `12dbca13`) y quedó archivada. Su texto de «Viewer-first walkthrough publication» ya está en la spec base, y este delta parte de él.
- CLIs instaladas en esta Mac: `claude`, `codex`, `grok`, `opencode`, `pi` y `cursor-agent`.

## Entrega

Se reutilizan las respuestas que siguen vigentes en esta sesión para este repositorio:
- **Entrega `interactive` (D1 B):** se detiene antes del push para que el usuario valide; después hace push, abre el pull request y lo fusiona.
- **Revisión `Native: Claude Code /code-review` (D2 A):** corre sobre el candidato antes de la validación.

La interfaz (T5) también la recorre `flow-verify` y la revisa `hive-review-ux`. Estas respuestas no autorizan desplegar en los directorios globales de los hosts ni aprobar el hook en Codex, que siguen siendo prerrequisitos.

## Estado de revisión y progreso

2026-11-04 — Plan escrito.

2026-11-04 — Revisión `hive-review-plan`: 4 P1, 9 P2 y 6 P3. Cambios aplicados al plan:
- la cola pasa a ser semántica (D12 A);
- la convivencia de las tareas del host con las del plan (D13 A);
- la comprobación real en `bb1` (D14 B);
- el actor de los eventos del host es el productor de su sesión;
- se rechaza la identidad ambigua y se conserva la vía explícita;
- bloques MODIFIED para la revisión 3, las tareas, los eventos, la recuperación y «Viewer-first»;
- T0 para verificar cada host;
- dependencias de T8 y pruebas añadidas.
