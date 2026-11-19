# Design

## Approaches

Medir el tiempo activo de una sesión (D47):

1. **Comando del gestor `hive time`.** Lee la transcripción nativa de la sesión y suma. Exacto y sin juicio; reusa `resolveViewIdentity` y `locateHostSessionFile`; solo Claude Code, Codex, Grok y Pi.
2. **Marcas del agente.** El agente anota la hora en sus paradas y un script suma. Cualquier host; depende de que no olvide una marca.
3. **Ambos (elegido).** El 1 donde hay transcripción y el 2 en OpenCode, Cursor y Antigravity. Cubre todos los hosts con dos mecanismos.

Escritura en Jira de lo que `acli` no puede (D45): orden por capacidad, `acli` → MCP de Atlassian autenticado → REST con el token del archivo global (D50). Alternativas descartadas: solo token (pierde el MCP cuando lo hay) y solo `acli`/MCP (el worklog podía quedar bloqueado). Origen del token: variable de entorno (D49) descartada porque Codex la filtra y las apps de GUI no leen `.zshrc`.

## Decisions

Ver `proposal.md` «Decisions».

## Proposed shape

### `hive time`

```
hive time [--since <RFC3339>] [--factor <n>] [--session-file <path> --host <host>]
```

- Sin `--session-file`, resuelve host y sesión como `hive view` (`resolveViewIdentity`) y el archivo con `locateHostSessionFile`. Un host sin archivo sale con `{"result":"no_session_file"}` y código 0, para que el agente use las marcas.
- **Lectura completa:** recorre el archivo entero línea por línea, quedándose solo con hora, rol, origen y comandos; no reusa el corte de cola de `readHostSessionFile`. Si no puede leerlo entero (línea ilegible, archivo truncado), sale con `"result":"partial"` y el agente no registra con ese total.
- Ventana: desde `--since` o el primer evento de la sesión hasta el último evento leído.
- **Espera por el usuario:** desde el último evento del agente que precede a un mensaje humano hasta ese mensaje, cuando el turno del agente terminó.
  - Claude: solo entradas `user` con `origin.kind == "human"` (o el formato humano equivalente que el fixture fije) y los resultados de la herramienta de preguntas; las notificaciones de tareas y las continuaciones tras compactar no son humanas, y el hueco que cierran es tiempo de máquina.
  - Codex: `user_message` tras el fin de turno, y el resultado de `request_user_input`.
  - Grok y Pi: su equivalente, fijado leyendo un archivo real de cada host antes de escribir su lector; si un host no lo distingue, su lector sale con `partial`.
- **Espera de CI:** duración de las llamadas de shell cuyo comando espera CI o un despliegue remoto (`gh pr checks … --watch`, `gh run watch`); en Codex, los `wait` sobre la celda que inició uno de esos comandos; en Claude, también el lapso entre lanzar uno de esos comandos en segundo plano y la notificación de su fin.
- **Límite conocido:** la espera por aprobar un permiso (Claude) o una ejecución (Codex) va dentro de la llamada y cuenta como activa.
- Salida JSON: `{"result":"ok|partial|no_session_file","host","session","since","until","active_seconds","user_wait_seconds","ci_wait_seconds","factor","converted_seconds"}` con `converted_seconds = (active − ci) × factor + ci`, redondeado al minuto hacia arriba. `--factor` por omisión 2.0; mayor que 3 o menor que 1 es error de entrada (código 2).

### Marcas (hosts sin transcripción)

`content/skills/hive-rule-trackers/scripts/work_time.py` con `mark <file> start|wait|resume|ci-start|ci-end` y `sum <file> [--factor n]`, que imprime `active_seconds`, `user_wait_seconds`, `ci_wait_seconds`, `factor` y `converted_seconds` con la misma fórmula y redondeo que `hive time`. El archivo de marcas vive junto al checkpoint del trabajo; nunca en el repositorio.

### Escritura en Jira

`content/skills/hive-rule-trackers/scripts/jira_rest.py` con `transitions <KEY>`, `worklog-list <KEY>` (solo los de la cuenta actual: ID y segundos), `worklog-add <KEY> --seconds n`, `worklog-update <KEY> <id> --seconds n` y `attach <KEY> <file>…`.
- Token: lee `~/.config/hive/secrets/jira-api-token` (respeta `XDG_CONFIG_HOME` absoluto); comprueba que la carpeta sea `0700` y el archivo `0600` y del usuario; si falta, sale con código 3 y el mensaje de cómo crearlo, sin crearlo.
- Email y sitio: de `acli jira auth status`; el sitio debe coincidir con el del ticket que da el proyecto.
- Nunca imprime el token ni lo pasa en argumentos; un error HTTP imprime método, ruta sin query y código, nunca el cuerpo de la petición ni cabeceras.
- El worklog no lleva comentario.

### Dónde vive cada regla

- **Cuándo** (entrega, cierre, sesiones previas): `hive-rule-ticket-state`, con su descripción nombrando el worklog.
- **Cómo** (medir, convertir, escribir, qué no escribir): `hive-rule-trackers/references/time.md` (nuevo, porque la hoja tiene tope de 4,000 bytes), enlazado desde la hoja; el orden por capacidad en `references/jira.md`; Linear y GitHub dicen que no registran tiempo.
- **Autorización** del worklog: `global.md` «Scope and authorization»; el token global: `global.md` «Secrets» (un almacén más, que solo el usuario escribe; el agente comprueba su existencia sin leer ni mostrar su contenido; los agentes de otros proveedores nunca lo reciben) y la línea de recursos privados («Read private resources only through their authenticated CLI or MCP tool») amplía a REST con ese token cuando la CLI y el MCP no tienen la capacidad.
- **Sesiones previas:** el estado del plan en memoria persistente (momento ya permitido: «when a plan is ready») guarda host e ID de sesión; `flow-plan` lo nombra. Nada de tiempo va a «Pending work», que se comenta en el ticket.
- **Registro local** del ID del worklog, tiempo real, factor y tipo: el registro de cierre (`flow-close` paso 7).
- **Índice:** las líneas de «Rule leaves» de ticket-state y trackers nombran el worklog; la de interrogate cambia su disparador a «When the user asks for a review by another provider's agent».

## Impact

`global.md` crece por la autorización del worklog, el almacén global de «Secrets», la línea de recursos privados y los disparadores del índice; el tamaño se reporta en el PR. `hive time` es un comando nuevo del gestor: `README.md`, `llms.txt` y `deployment-manager.md`. Pruebas de contenido que cambian: `tests/content/leaf_links_test.go:93` (`delivery-ask` deja de enlazar `interrogate`; `flow-verify` ya lo enlaza) y `single_home_test.go` si fija frases movidas.
