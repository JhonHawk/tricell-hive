# Margen de ahorro de TokenEconomizer en sesiones reales (2026-11-05)

Pregunta: ¿los tres mecanismos de TokenEconomizer (microsoft/TokenEconomizer, arXiv 2610.22732) tienen margen en las sesiones de este equipo? Réplica de las *headroom measurements* del paper (apéndice B) sin llamadas a modelo.

- Script: `measure-headroom.py` (esta carpeta). Resultado crudo: `headroom.json`, sin trackear: se regenera con el script.
- Datos: hilos principales de los últimos 30 días, 194 sesiones de Claude Code (16 repos, 36 480 llamadas) y 35 de Codex (5 repos, 13 807 llamadas; sin `exec` ni subagentes).
- Costo: tokens efectivos relativos con precios de caché; los hosts no se mezclan. Todos los porcentajes son cotas superiores de exposición, no ahorro alcanzable.

## Resultados

| Métrica | Claude Code | Codex |
|---|---|---|
| M1 — llamadas cuya acción ya apareció en una sesión previa del mismo repo (control: otros repos) | 48.7 % (control 36.5 %) | 48.4 % (control 44.7 %) |
| M2 — costo cargado por definiciones de tools que la sesión no usa | 3.2 % agregado, mediana 4.3 % | no medible (sin catálogo en la traza) |
| M3 — costo después de que un error se repite dentro de la sesión | 15 % agregado, mediana 0 %; 27 de 194 sesiones | 60 % agregado, mediana 0 % (baja confianza) |
| M4 — llamadas fallidas con una firma que reaparece en ≥ 3 sesiones del repo | 113 de 474 fallidas (0.3 % de las llamadas) | 52 de 452 (baja confianza) |

### Lectura

- **Micro-skills (M1, M4): margen bajo.** El exceso de repetición sobre el control es de 12 puntos en Claude y 4 en Codex, y repetir no implica desperdicio (`git status`, lecturas de specs). Los errores que reaparecen entre sesiones son el 0.3 % de las llamadas. Las firmas recurrentes con causa clara:
  - Guardia de worktree de Claude Code (no es de Hive): 8 firmas, 56 llamadas, hasta 12 sesiones (`git` compuesto, `cd`/variables calculadas, editar la ruta del checkout compartido, `ExitWorktree` con commits nuevos).
  - Engram: `mem_update` sin `expected_project` (10 sesiones) y `mem_judge` sin `relation` (5 sesiones, 15 llamadas).
  - Visor de Hive en cam (Codex): `invalid timeline input: forbidden_for_kind` en `stage`, `views` y `target` (3 sesiones cada una), desde `tooling/viewer/timeline_contract.go`.
  - `Edit` con texto no encontrado en `design.md` (7 sesiones).
- **Tool pruning (M2): margen pequeño y concentrado.** Claude Code ya difiere los tools MCP (solo un marcador en la forma de la petición; verificado en `api-request-shape`), así que lo que se cobra son los tools nativos cargados. `Artifact` (~9.6 k tokens por petición, sin uso en el 97 % de las sesiones) explica 2.05 % del costo; los siguientes (`Workflow`, `SendFeedback`, `ScheduleWakeup`) suman menos de 1 %.
- **Loop guard (M3): margen incierto.** En Claude, el 15 % del costo cae después de un error repetido, pero lo concentran pocas sesiones largas (mediana 0 %) y el paper muestra que la mayoría de las corridas con repeticiones terminan bien. No hay forma de separar atascos de recuperaciones sin juzgar cada sesión.

## Seguimiento: errores del visor `forbidden_for_kind`

M4 los subcontó: el parser de Codex no emparejaba todas las salidas. Búsqueda directa: unos 37 rechazos `invalid timeline input` en 14 sesiones de Codex (cam, carvento, fleetsystem, tricell-hive, 3–5 nov), unos 22 de ellos `forbidden_for_kind`. También aparecen en 9 sesiones de Claude Code de cam y carvento (detalle no revisado).

- **Causa 1, mensaje engañoso:** `TimelineAppendEventRequest.validate` (`tooling/viewer/timeline_contract.go:1131-1142`) rechaza `stage`, `stage_state` y `expected_session_version` en eventos que no son `stage`, y siempre culpa al campo `stage`. En cam (Codex, 2026-11-04 21:43) un evento `validation` con `expected_session_version: 47` falló cuatro veces seguidas con `forbidden_for_kind on stage`: el modelo quitó `stage`, que ya no estaba, y cambió `kind` en lugar de quitar el campo culpable. Lo mismo pasa en la línea 2092, que agrupa `started_at`, `finished_at` y `document_snapshots` bajo `stage`.
- **Causa 2, contrato sin ramas:** `hive view contract --operation append_event` lista `expected_session_version`, `stage`, `stage_state`, `views` y `target` como opcionales, sin ninguna condición por `kind` (el esquema no trae `if`/`then`). Los invariantes solo dicen que los campos dependen del tipo y que se prohíben las ramas mezcladas, sin decir cuáles. El modelo aprende la regla a prueba y error.

## Límites

- Codex 0.166 ejecuta varios tools por llamada `exec` (modo código); el parser empareja salidas con comandos por orden y toma la primera línea de la salida como firma, lo que mete ruido (banners, `total output lines`). Sus M3 y M4 son de baja confianza.
- Los errores de MCP en modo código de Codex y los fallos sin código de salida no se cuentan.
- No incluye Grok, Cursor, OpenCode ni Pi, ni subagentes.
- Precios relativos aproximados (Claude: entrada 1, escritura de caché 1.25/2, lectura 0.1, salida 5; Codex: 1/0.1/8).
