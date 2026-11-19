# ticket-worklog-tracker-writes — Tiempo trabajado en el ticket, escritura en Jira por capacidad y revisión adversarial solo a petición

| Campo | Valor |
| --- | --- |
| Estado | In build · E1 integrado (#523), E2 en PR |
| Tracker | ninguno (lote de ajustes de la revisión de sesiones del 2026-11-19; `engram:tricell-hive/pending/session-review-adjustments`) |
| Git | tricell-hive · `development` en `81d5b4f1` · ramas `feat/hive-time`, `feat/ticket-worklog` |
| Playbook | ninguno (`/flow-plan` tecleado por el usuario) |

## Decisions

2026-11-19
- **Ajuste 1** — ¿Se sigue preguntando la revisión adversarial con otro proveedor? No: se quita la pregunta y el sondeo de hosts; una revisión por el agente de otro proveedor corre solo cuando el usuario la pide explícitamente, y nunca se ofrece ni se sugiere.
- **D38 A** — ¿Se mantiene la pregunta del mecanismo de revisión de código? Sí, sin cambios.
- **D39 A** — ¿Entran al lote los criterios del auditor para la caché de estados no escrita y el «Pruébalo tú» que falta? Sí, como criterios de `--audit-session` con sus fixtures.
- **D40** — ¿Qué tiempo se registra? El tiempo activo de la sesión; solo se restan las esperas por el usuario; CI, builds y todo lo de máquina cuentan.
- **D41** — ¿Cómo se reparte entre varios tickets trabajados a la vez en una sesión? Cada ticket recibe el tiempo completo.
- **D42** — ¿Con qué herramienta? El agente usa la que tenga disponible y pueda registrar tiempo; por ahora se configura el MCP de Atlassian en los hosts (hecho el 2026-11-19 en Claude Code, Codex, OpenCode y Cursor).
- **D43** — ¿Cuándo? Se registra en la entrega y se actualiza al cierre.
- **D44** — ¿Va en el worklog el tiempo convertido a tiempo humano? Sí, el worklog lleva el tiempo convertido.
- **D45 A** — ¿Cómo escribe el agente en Jira lo que `acli` no puede? Por capacidad: `acli`; después el MCP de Atlassian si el host lo tiene autenticado; después REST con token. Se corrige «nunca REST con token» en la hoja de trackers.
- **D49** — ¿De dónde sale el token? De una variable de entorno global `JIRA_API_TOKEN` · **Replaces:** D45 (solo el origen del token)
- **D50 A** — Codex quita de sus comandos las variables con `TOKEN` y las apps abiertas desde la GUI no leen `.zshrc`: ¿dónde vive el token? En un archivo global que el usuario crea una vez, `~/.config/hive/secrets/jira-api-token` (carpeta `0700`, archivo `0600`); el agente le pide crearlo cuando falta y nunca lo escribe · **Replaces:** D49
- **D46 A** — ¿Qué multiplicador? × 2.0 sobre el tiempo activo, × 1 en esperas de CI y despliegue remotos, tope × 3; se recalibra a los 20–30 tickets contra tiempos humanos de tickets parecidos.
- **Corrección del usuario** — ¿Se comenta el tiempo en el ticket? No: el worklog lleva solo el tiempo, sin descripción, y el agente no escribe razonamiento al respecto en ningún lado del tracker; el tiempo real, el factor y el tipo quedan en el registro de cierre local.
- **D47 A** — ¿Cómo se mide el tiempo activo? Con ambos mecanismos: un comando del gestor, `hive time`, lee la transcripción nativa en Claude Code, Codex, Grok y Pi; en OpenCode, Cursor y Antigravity el agente marca la hora en sus paradas y un script suma.
- **D48 A** — ¿Desde cuándo cuenta? Desde el primer mensaje de la sesión que trabaja el ticket: investigación y plan incluidos.
- **D51 A** — Con varios tickets trabajados uno tras otro en una sesión, ¿desde cuándo cuenta cada uno? Desde su propio primer mensaje; los tickets trabajados a la vez siguen recibiendo cada uno el tiempo completo (D41).
- **Revisión de seguridad** — ¿Corre `hive-review-security`? No (respuesta del usuario).
- **Supuesto (no preguntado)** — La compilación y las pruebas locales cuentan como trabajo (× 2.0); solo las esperas de CI y de despliegues remotos van a × 1.
- **Supuesto (no preguntado)** — Una revisión de un plan por otro proveedor no tiene procedimiento y no se agrega: «a petición» aplica a la revisión del candidato que ya define `hive-rule-interrogate`.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «Super importante para el manejo de los tickets, que registren el tiempo trabajado» · «solo se coloca el tiempo, no se razona al respecto» · «quitemos el sondeo, y solo se haga hasta que explícitamente se solicite» · «que se le pida al usuario una variable global» (usuario, 2026-11-19).

Origen: revisión de las sesiones de Codex `01a1f358` y `01a1f37d` (fleetsystem, 2026-11-19) y los watch items `watch/tracker-states-cache-not-written` y `watch/final-report-without-try-it`.

## Objective

Hoy Hive no registra tiempo en los tickets, su hoja de trackers prohíbe el único camino que hoy escribe worklogs y adjuntos en Jira (REST con token), y la ronda de entrega sondea otros hosts para ofrecer una revisión adversarial que el usuario no quiere que se sugiera. Al terminar, cada ticket recibe su tiempo convertido en la entrega y su ajuste al cierre, Jira se escribe por capacidad con un token global, la revisión adversarial corre solo a petición, y el auditor detecta la caché de estados no escrita y el «Pruébalo tú» que falta.

## Scope and acceptance

Dentro: `hive time` en el gestor; el script de marcas y suma; la escritura en Jira por capacidad (transiciones, worklogs, adjuntos); la hoja de trackers y su referencia nueva de tiempo, `ticket-state`, `flow-close`, `flow-plan` (IDs de sesión en el estado del plan) y `global.md`; quitar la pregunta adversarial; dos criterios del auditor; documentación. Fuera: Linear y GitHub no tienen worklog nativo, así que no registran tiempo; la calibración del factor (necesita 20–30 tickets reales); una revisión de planes por otro proveedor; cambiar `acli`. Límite conocido: el tiempo que el usuario tarda en aprobar un permiso de herramienta (Claude) o una ejecución (Codex) queda dentro de la llamada y cuenta como activo.

- AC1 — La ronda de entrega ya no pregunta la revisión adversarial ni sondea otros hosts; `hive-rule-interrogate` se lee solo cuando el usuario pide una revisión por otro proveedor, y su «Availability» corre entonces · false today because `hive-rule-delivery-ask` pregunta «Adversarial review» y la sesión `01a1f37d` corrió `claude auth status` y `cursor-agent status` antes de la ronda · Observed in: guía instalada tras `hive update` y una sesión real posterior sin ese sondeo
- AC2 — `hive time` en una sesión de Claude Code, Codex, Grok o Pi imprime en JSON `active_seconds`, `user_wait_seconds`, `ci_wait_seconds`, `factor` y `converted_seconds` (activo sin CI × factor más CI × 1; factor 2.0 por omisión, `--factor` rechaza más de 3) desde `--since` o el inicio de la sesión; lee la transcripción completa, y cuando no puede reporta `partial` en vez de un total corto; en Claude solo cuentan como espera por el usuario los mensajes con origen humano y las respuestas a tarjetas · false today because el gestor no tiene ese comando · Observed in: `hive time` ejecutado desde esta sesión de Claude Code y sobre los rollouts `01a1f358` y `01a1f37d` con `--session-file`
- AC3 — En un host sin transcripción legible, el script de marcas registra inicio, espera, reanudación e inicio y fin de CI, y su suma da los mismos `active_seconds`, `user_wait_seconds`, `ci_wait_seconds` y `converted_seconds` que `hive time` sobre la misma secuencia · false today because no existe · Observed in: el script sobre un archivo de marcas derivado de un fixture de `hive time`, comparado campo por campo
- AC4 — Al llegar a la entrega (integración, parada `ready` o `hold`) de un ticket de Jira, el agente lee los worklogs de su cuenta en el ticket, agrega uno con solo el tiempo convertido y sin descripción a cada ticket de la sesión, y al cierre actualiza ese mismo worklog con el tiempo sumado después; el ID del worklog, el tiempo real, el factor y el tipo quedan en el registro de cierre en Engram y no en el tracker · false today because ninguna regla registra tiempo · Observed in: una sesión real posterior al despliegue que registre y lea de vuelta el worklog en Jira
- AC5 — El estado del plan en memoria persistente guarda el host y el ID de cada sesión que trabaja el ticket, y la sesión que entrega vuelve a medir esas sesiones con `hive time --session-file` y suma su tiempo, sin escribir nada del tiempo en el ticket · false today because ese tiempo se pierde · Observed in: guía instalada y una sesión real de plan seguida de una de build
- AC6 — La hoja de trackers escribe en Jira por capacidad: `acli`, después el MCP de Atlassian autenticado, después REST con el token de `~/.config/hive/secrets/jira-api-token` y el email y el sitio de `acli jira auth status`; el token nunca va en argumentos, salida, mensajes ni prompts de otros agentes; cuando falta, el agente le pide al usuario crear el archivo y nunca lo escribe; las transiciones se leen en vivo por REST cuando `acli` no las lista y los adjuntos suben por REST · false today because la referencia de Jira dice «Read private resources only through these authenticated tools» y deja los adjuntos para el usuario · Observed in: guía instalada y una sesión real que mueva, adjunte y registre tiempo en Jira
- AC7 — `--audit-session` marca `fail` cuando una sesión busca `tracker/states/<tracker>/<prefijo>` sin resultado (también dentro de una celda `exec` de Codex), mueve un ticket (`acli … transition`, el MCP de Atlassian o de Linear, `gh issue`) y no escribe ese tema por `mem_save`, `mem_update` o `engram save` · false today because no existe el criterio y el rollout `01a1f37d` no da `fail` · Observed in: `go run ./tests/pilot --audit-session` sobre el rollout `01a1f37d` y su fixture
- AC8 — `--audit-session` marca `fail` cuando una sesión edita un archivo de interfaz y el último mensaje del agente tras `flow-verify` no lleva un encabezado de la lista «Pruébalo tú», «Pruébalo», «Try it» · false today because no existe el criterio y el rollout `01a1f37d` no da `fail` · Observed in: `go run ./tests/pilot --audit-session` sobre el rollout `01a1f37d` y su fixture
- AC9 — `README.md`, `llms.txt`, `site/src/content/docs/flows/{plan,verify,close}.md`, `deployment-manager.md` e `instruction-resources.md` describen `hive time`, los scripts, el registro de tiempo y la revisión adversarial a petición · false today because no lo mencionan y `flows/plan.md:28` describe la pregunta adversarial · Observed in: esos archivos en el candidato

## Prerequisites

- [ ] Archivo `~/.config/hive/secrets/jira-api-token` (carpeta `0700`, archivo `0600`) con el token de API de Jira, creado por el usuario · owner: usuario · needed by: T8 (observación tras el despliegue)

## Verified context

- La identidad de host y sesión ya la resuelve el gestor: `tooling/cli/view_identity_resolve.go` (`CLAUDE_CODE_SESSION_ID`, `PI_SESSION_ID`, `OPENCODE_SESSION_ID`, `CODEX_SESSION_ID`, o el proceso padre) · leído 2026-11-19.
- `locateHostSessionFile` en `tooling/cli/view_reconcile.go` localiza la transcripción de `claude-code`, `codex`, `grok` y `pi`; los demás hosts no tienen archivo. `readHostSessionFile` lee solo los últimos 16 MiB (`reconcileReadLimit`), y hay transcripciones de Claude de hasta 59.5 MB en esta máquina · leído 2026-11-19.
- En las transcripciones de Claude, las entradas `user` incluyen notificaciones de tareas (`origin.kind == "task-notification"`) y continuaciones tras compactar, además de las humanas · revisión del plan, 2026-11-19.
- Los lectores de transcripciones de Claude, Codex y Grok con horas de mensajes viven en `tests/pilot/session_audit.go` (no se distribuyen) · leído 2026-11-19.
- `acli` 1.3.40 (la versión actual del cask de Homebrew) no tiene comando de worklog y su JSON de edición no tiene campos de tiempo · `acli jira workitem --help`, `acli jira workitem edit --generate-json` · 2026-11-19.
- La API REST de Jira agrega worklogs y lista transiciones con autenticación básica email:token: la sesión de Codex `01a1f386` (fleetsystem, 2026-11-19) leyó el token de un archivo de una línea, el email y el sitio de `acli jira auth status`, listó `issue/<KEY>/transitions` y registró 25 min en FSR-905 y 15 min en FSR-906, leídos de vuelta · rollout `01a1f386` · 2026-11-19.
- Actualizar un worklog y subir un adjunto por REST · no medido · T8.
- Escribir un worklog o un adjunto por el MCP de Atlassian (`https://mcp.atlassian.com/v2/mcp`; su documentación en Context7 lista búsqueda, creación y comentarios) · no medido · T8.
- Codex quita de los comandos que lanza las variables cuyo nombre contiene `KEY`, `SECRET` o `TOKEN` salvo `ignore_default_excludes` (documentación de Codex, `shell_environment.rs`); `~/.codex/config.toml` no lo cambia · documentación, no ejecutado · 2026-11-19 (motivo de D50).
- Investigación del multiplicador (subagente, 2026-11-19): ningún estudio mide tiempo de reloj del agente contra tiempo humano; rangos de × 1.5 a × 3 para tickets acotados; METR RCT 0.84x, transcript analysis 1.5–13x «soft upper bound».

## Effects

None: the tasks take no effect on an external system. La observación real (T8) ocurre en sesiones de trabajo posteriores, con la autorización que la propia regla da a las escrituras del tracker.

## Delivery

`interactive` (respuesta de esta sesión): parada para validación antes de cada push, luego PR y merge. Revisión: `Native: Claude Code /code-review`. Revisión de seguridad: no (respuesta del usuario, 2026-11-19). Dos unidades (E1 gestor y auditor, E2 guía, scripts y documentación). No autoriza escribir en ningún Jira durante la construcción.

## Review status and progress

2026-11-19 — Plan escrito. Revisión `hive-review-plan`: 21 huecos (5 P1); aplicados todos, con D50 y D51 decididos por el usuario y dos supuestos anotados (plan por otro proveedor fuera de alcance; límite de permisos dentro de la llamada).
