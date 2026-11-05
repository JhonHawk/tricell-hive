# Diseño

## Enfoques

1. **Elegido: hook → `hive hook` → API del visor, con cola local y conciliación.** Sigue el patrón que ya funciona con el hook de Grok de Hive (`integrations/grok/hook.go`, `tooling/cli/grok_hook.go`) y con Engram en cinco hosts. La latencia es baja y la pieza nueva es una sola, el subcomando.
2. Los hooks solo escriben en la cola y el visor la procesa. Exige que el visor vigile la cola de forma continua, algo que el plan archivado de publicación bajo demanda descartó.
3. Sin hooks, el visor lee directamente los archivos de sesión. Son seis formatos que no son una interfaz estable y no dan la identidad en el momento.

## Forma propuesta

### Identidad y trabajo actual (binario)

Orden de resolución de la identidad, usando la primera fuente que exista:
1. las banderas explícitas `--host` y `--native-session-id`;
2. la variable documentada del host: `CLAUDE_CODE_SESSION_ID` en Claude Code y `PI_SESSION_ID` en Pi;
3. las variables de los demás hosts (`CODEX_THREAD_ID` u otra en Codex, `OPENCODE_SESSION_ID`), solo después de verificarlas en la documentación oficial y en la versión instalada;
4. el mapa `<state>/hooks/sessions/<pid>.json` que escribe el hook de inicio, buscado subiendo por los procesos padre. Solo vale si nombra una única sesión viva para ese ancestro y es posterior al último cambio de sesión de ese proceso. Así se cubren OpenCode con varias sesiones en un proceso, los hilos de subagente de Codex y `/clear` en Claude Code.

Sin ninguna fuente, o con una ambigua, la orden se rechaza. `host-session-identity.md` se mantiene como la vía explícita por host para ese caso. La identidad resuelta es la real, no un ID inventado. Las operaciones de bajo nivel siguen pidiendo `--host` y `--native-session-id` (D6 A).

El procedimiento deja de pedir que se guarden en el checkpoint los IDs devueltos, porque ya los guarda el binario en el trabajo actual; el checkpoint conserva solo la clave del trabajo.

Trabajo actual: `<state>/publisher/v1/current/<host>/<native-id>.json` guarda el ID del trabajo activo de esa sesión. `work open` y `work publish` en su forma de creación lo fijan, y `work close` lo libera. Las órdenes `work` sin `--work-id` usan ese trabajo; con dos trabajos abiertos en la misma sesión, se exige `--work-id`.

Registro del proyecto: `work open` registra la raíz si falta, en lugar de fallar con `project_identity_unavailable`.

**Formato del mapa pid** (fijado en T2; T6 lo escribe tal cual): `<state>/hooks/sessions/<pid>.json` = `{"host", "pid", "process_started_at", "changed_at" (RFC 3339), "sessions": [{"session_id", "started_at", "ended_at"?}]}`. Las reglas completas de quien lo escribe están en `_support/docs/architecture/host-session-events.md` («Session map»).
- Una sesión está viva si no tiene `ended_at` y su `started_at` es igual o posterior a `changed_at`.
- `changed_at` es el primer inicio y solo se mueve cuando una sesión sustituye a otra (`/clear`, reanudar otra sesión o fin). Una segunda sesión en paralelo no lo mueve, así que dos sesiones vivas siguen detectándose como ambiguas.
- `process_started_at` se compara con la hora de inicio real del proceso; si no coincide, el pid se reutilizó y el mapa no cuenta.
- Si la variable encontrada es de un host distinto al del ancestro más cercano (con mapa, o con el nombre de proceso de una CLI de host), la identidad se rechaza como ambigua.
- Si hay más de una variable de host, la identidad se rechaza como ambigua, porque un host anidado hereda las variables del de fuera.
- Un ID de sesión que no sea seguro como nombre de archivo se guarda en el archivo de trabajo actual con la clave `sha256-<hex>`.

### Órdenes `work` nuevas o ampliadas

- `work open --key K --title T --summary S` sin `--change` crea un trabajo sin carpeta de cambio; `--change` sigue siendo válido.
- `work sync --stage S --state X --reason R [--tasks FILE|-] [--close OUTCOME]` es la orden declarativa por etapa:
  - inicia o actualiza la etapa;
  - registra las tareas que falten, sin exigir revisión de plan cuando el trabajo no tiene carpeta;
  - vincula sus segmentos a la etapa y aplica los estados de tarea;
  - con `--close`, completa la etapa activa y cierra el trabajo.

  Es idempotente: repetir el mismo estado devuelve `unchanged`. La respuesta trae los IDs de tarea y de segmento y cabe en 4.096 bytes.
- `work evidence --stage S --view-title T --image FILE[:caption] … [--candidate COMMIT] [--supersedes VIEW]` sube las capturas, registra el candidato si hace falta y añade el evento con su vista.
- `work close` y `work sync --close` completan la etapa activa antes de cerrar y leen el trabajo de vuelta dentro de la misma orden. Esa es la lectura de cierre que exige la spec.

Las operaciones de bajo nivel siguen funcionando sin cambios (D6 A), y el contrato las presenta como avanzadas.

### Cola y lock

- **Lock por trabajo:** `publisher/v1/locks/<work-id>.lock` reemplaza al lock global en las órdenes que afectan a un solo trabajo. La creación y la migración del journal siguen usando el lock global.
- **Cola semántica (D12 A):** si el visor no está en marcha o no responde, `work open`, `work sync` y `work evidence` guardan su entrada semántica y su ID de operación en el journal con el estado `queued`, y responden `queued`. Una imagen se copia a la carpeta del publicador para que sus bytes no cambien.
  - La siguiente orden `work`, o el siguiente `hive hook`, con el visor en marcha procesa la cola en orden: construye las peticiones sobre el estado actual y las registra con sus bytes exactos antes de transmitirlas, como cualquier intención.
  - Un `open` en cola devuelve un identificador local: un hash del proyecto, el host, la sesión nativa y la clave. El ID real del visor no se puede derivar sin el servidor, porque también usa la identidad del proyecto en el servidor. Al enviarse, el journal asocia el identificador local con el ID real y reescribe el archivo de trabajo actual.
  - Los locks son por trabajo, incluida la creación (con el identificador local). Las imágenes de `evidence` se congelan con un máximo de 8 MiB por imagen, por el límite del journal.
  - Lo que ya no aplique (una etapa completada por otro, un trabajo cerrado) se informa como rechazado con su motivo y no sobrescribe nada.

  Las entradas en cola sobreviven al proceso y no caducan.

### Eventos del host

- **Servidor:** una ruta nueva, `POST /api/host-events`, con la capacidad `host_events` anunciada en `/api/health`. Usa el token `X-Hive-Token` y no exige `Origin`, porque la llama un proceso local y no el navegador.
- **Cuerpo:** host, ID nativo, proyecto (cwd), tipo y datos. Los tipos son `session_started`, `activity`, `session_ended`, `task_created`, `task_updated`, `task_completed`, `subagent_started` y `subagent_stopped`.
- **Actor:** el evento actúa como el productor nativo aceptado de esa sesión, que ya es miembro del trabajo. Se marca con la procedencia `host_observed` y lleva un ID de operación derivado del evento, para que un reenvío se aplique una sola vez. Un evento de una sesión que no es productora del trabajo se rechaza. Así se cumplen el asignado aceptado y la autoridad de transición sin crear un actor nuevo.
- **Trabajo y etapa:** `hive hook` los resuelve desde el archivo de trabajo actual de esa sesión y los envía. Sin trabajo actual o sin etapa activa, el evento solo mueve la última actividad.
- **Efectos:**
  - **Actividad:** actualiza `last_activity_at` de los trabajos de esa sesión como metadato de mantenimiento, no como evento del historial. No consume el tope de 508 eventos ni invalida las versiones que espera el agente.
  - **Tareas nativas (D13 A):** solo se reflejan cuando el trabajo no tiene tareas registradas por el agente. Se guardan con la clave `host:<id-nativo-de-tarea>`, asignadas al productor de la sesión, vinculadas a la etapa activa y con un tope de 50. Si el trabajo tiene tareas del plan, solo suman actividad.
  - **Antigüedad:** un evento de tarea con fecha anterior a la creación del trabajo se ignora. Esto incluye los de un trabajo creado por un `open` en cola, que nace al enviarse la cola.
  - **Límites:** ningún evento del host inicia, completa o cierra una etapa, cambia la autoridad, ni cierra o pausa un trabajo. `session_ended` solo marca la última actividad.
- **Contrato:** pasa a la revisión 3 (`TimelineContractRevision`), con el origen `host_observed` y su reflejo en `timeline-contract*.ts`.
- **Frontend:** las tareas y eventos `host_observed` muestran su procedencia («observado del host»), y el trabajo muestra la última actividad.

### `hive hook`

`hive hook <host> <evento>` lee el JSON por stdin con un límite de 1 MiB. Hace lo siguiente:
- convierte los campos de cada host, en snake_case o en camelCase;
- en el inicio, escribe el mapa pid → sesión;
- envía el evento con un tiempo límite de 2 s;
- si el visor está caído, deja el evento en la cola de eventos de host.

No imprime nada y siempre sale con 0. Sigue el patrón de `grok_hook.go`. Cada traductor de carga vive en su `integrations/<host>/`.

**Correcciones de T0** (fuentes en `_support/docs/architecture/host-session-events.md`):
- **El host real se detecta siempre por la carga y el entorno.** Grok y Cursor CLI también ejecutan los hooks de `~/.claude/settings.json`, así que el argumento `<host>` es solo una pista.
- **Tareas nativas por host:**
  - Claude Code: `TaskCreated` y `TaskCompleted`.
  - OpenCode: el hook del plugin sobre la herramienta `todowrite`, con un ID derivado del contenido. En 2.2.5 no existe `todo.updated`.
  - Codex, Grok, Pi y Cursor: no tienen.
- **OpenCode y Pi** pasan el ID de sesión de forma explícita desde el plugin o la extensión. En OpenCode el mapa pid no sirve porque todas las sesiones comparten el servicio.
- **Cursor** registra el mapa también en `beforeSubmitPrompt`, porque `sessionStart` no se dispara al reanudar.
- **Codex:**
  - no se registran `Stop` ni `SubagentStop`, que esperan JSON en la salida;
  - un evento con `agent_id` es de un subagente y no cambia la sesión principal;
  - `CODEX_THREAD_ID` queda sin usar mientras no se verifique.
- **Claude Code:** `SessionEnd` lleva `timeout` explícito, y los subagentes con `agent_type` vacío se ignoran.
- **Fuentes de identidad en el shell** (solo el shell del agente, no los `!` del usuario): `CLAUDE_CODE_SESSION_ID`, `PI_SESSION_ID` y `OPENCODE_SESSION_ID`. `CURSOR_CONVERSATION_ID` no se usa, porque no está verificado que coincida con el `conversation_id` del hook. Cursor y Grok resuelven por el mapa pid.

Las cargas que entiende por host son:
- **Claude Code:** `session_id`, `transcript_path`, `cwd`, `CLAUDE_PROJECT_DIR`, y para tareas `task_id` y `task_subject` de TaskCreated y TaskCompleted.
- **Codex:** `session_id` y `cwd` de los hooks SessionStart, UserPromptSubmit, SubagentStart, SubagentStop y SessionEnd.
- **Grok:** `sessionId`, `cwd` y `workspaceRoot`.
- **OpenCode:** el plugin envía `session.created`, `session.deleted` y `session.execution.*`, y los cambios de `todowrite`, con el ID de sesión explícito.
- **Pi:** la extensión envía `session_start`, `agent_settled` y `session_shutdown`.
- **Cursor:** `conversation_id` y `workspace_roots`.

**Conciliación:** `hive view work reconcile`, y cada `hive hook session-start`, leen el archivo de sesión del host (Claude Code y Codex JSONL, Grok `updates.jsonl`, Pi JSONL) para recuperar la última actividad y las tareas nativas que falten. Nunca deducen la muerte del proceso.

### Registro por host (gestor)

| Host | Destino | Notas |
|---|---|---|
| Claude Code | `~/.claude/settings.json`, entradas `hooks.<evento>[]` cuyo comando es `'<hive>' hook claude <evento>` | Editor JSON que conserva el resto del archivo byte a byte, guarda una copia y solo reconoce como propias las entradas con ese comando (D8 A) |
| Codex | `~/.codex/hooks.json`, mismas reglas | No escribe `trusted_hash`; el informe de `hive update` pide aprobar el hook en `/hooks` (D9 A) |
| Grok | `~/.grok/hooks/hive-session-events.json`, un segundo archivo nativo | Comando que llama al programa directamente, sin ruta entre comillas |
| OpenCode | `~/.config/opencode/plugins/hive-events.js` | Activo; el cargador de contexto sigue sin activarse (D10 A) |
| Pi | `~/.pi/agent/extensions/hive-events.ts` | Archivo propio |
| Cursor | `~/.cursor/hooks.json`, entradas propias | Sin verificar en CLI (D11 B); el informe lo dice |

La propiedad (`ownership.go`) admite los nuevos tipos y hosts. La comprobación del ejecutable (`checkHookExecutable`) se aplica a todos. Desinstalar quita solo lo propio.

## UI

Patrones que se siguen:
- **Fila de tarea:** `tooling/viewer/frontend/src/timeline-tasks.tsx`, con `TaskIndicator`, `TimelineHeader` y `TimelineTitle`.
- **Cabecera del trabajo:** `timeline-header.tsx:280-300`, con `Badge` de estado dentro de `ButtonGroup`.

No hay acciones nuevas.

**Fila de tarea observada del host.** Lleva `<Badge variant="outline">`, tamaño por defecto, después del título, como el estado de entrega en `timeline-tasks.tsx:387`. Su tooltip y su texto accesible dicen de dónde viene. El indicador de estado no cambia.

```
Antes                              Después
○ Ajustar placeholders  ⌄          ○ Ajustar placeholders [Del host] ⌄
✓ Correr typecheck      ⌄          ✓ Correr typecheck     [Del host] ⌄
```

**Cabecera.** La última actividad va en `text-xs text-muted-foreground`, antes del `Badge` de estado y en el mismo grupo `flex gap-2`. Solo aparece cuando hay actividad del host. En móvil (390 px) baja con el resto del grupo, que ya usa `flex-wrap`.

```
Antes                                   Después
Resumen…        [● Activo] [⋯]          Resumen…  Activo hace 3 min [● Activo] [⋯]
```

**Estados:**
- **Sin eventos del host:** no se muestra nada nuevo.
- **Carga y error:** los que ya tiene la vista del trabajo.
- **Más de 50 tareas del host:** no se añaden más; la vista no cambia.

## Textos

| Clave | es | en |
| --- | --- | --- |
| `hostObserved` | Del host | From host |
| `hostObservedHint` | Tarea tomada de la lista de tareas del host | Task taken from the host's task list |
| `lastActivity` | Activo hace {tiempo} | Active {time} ago |

## Impacto

- Productores que no actualicen la guía siguen con las operaciones de bajo nivel (D6 A).
- La revisión 3 del contrato obliga a recargar el visor y a actualizar el CLI a la vez (`reload_required`), como en la revisión 2.
- Las configuraciones globales de seis hosts reciben hooks: cada hook añade un proceso por evento, sin tokens.
- Un binario movido hace que los hooks fallen en silencio hasta el siguiente `hive update`, el mismo riesgo que ya tiene el hook de Grok.

## Recuperación

Antes de revertir, quitar los hooks con el binario de este cambio: `hive plan remove`, o `hive update` desde `development` con el binario nuevo. Un binario anterior no entiende el estado nuevo: el gestor sube la versión del esquema de estado, así que el binario viejo se niega con su error de «actualiza Hive» en vez de perder las entradas. Después ya se puede revertir el PR, reconstruir el binario anterior y recargar el visor en el contrato anterior.

Hay dos cosas que no se deshacen:
- los eventos del host ya guardados se quedan en el historial de los trabajos;
- las copias `.hive-backup` de `settings.json` y `hooks.json` se conservan con modo 0600 para restaurarlos a mano, y la siguiente instalación las sobrescribe.
