# Esperas de CI y subagentes: señal de fin en tramos, con aviso al usuario

| Campo | Valor actual |
| --- | --- |
| Estado | En validación · T1 y T2 verificados; commit, push y release en curso |
| Tracker · GitHub Issues | • [#36 — Esperas de CI y subagentes: una espera bloqueante y aviso al usuario](https://github.com/JhonHawk/tricell-hive/issues/36) |
| Git | `direct-base` a `rebuild/harness-engineering` · sin PR · despliegue a los seis hosts tras el push |
| Verificación | `go test ./tests/content/...` · `go test -race ./tests/pilot/...` · `go vet ./...` · reversión por rama de cada fixture que falla |
| Siguiente paso | `flow-build` sobre esta carpeta: T1 y T2 en paralelo, después T3 |

## Objetivo

Codex espera los procesos externos con bucles cortos. Para CI encadena `sleep` + `gh run view` (87 y 145 consultas en dos sesiones, contra 4 y 1 `gh run watch`). Para subagentes encadena `wait_agent` de 60 s o menos, y entre el 46 % y el 66 % se agotan sin resultado. Mientras tanto no le dice nada al usuario durante decenas de minutos. Cada consulta es una vuelta completa del modelo sobre el contexto, y desde fuera la sesión parece atorada.

La herramienta no es el límite. `wait_agent` acepta hasta una hora y el prompt del propio Codex ya pide esperas de minutos, pero el modelo lo ignora (ver [design.md](design.md#contexto-verificado)). El cambio agrega a la guía global una regla escrita por capacidad: esperar la señal de fin del proceso, en tramos de varios minutos, con una línea al usuario tras cada tramo. Agrega también un criterio determinista con fixtures derivados de las sesiones reales, y extiende el parser de trazas de Codex para que vea las llamadas de subagentes.

## Alcance y aceptación

**Incluye:**
- Regla de espera en `content/guidance/global.md` (regla de delegación, `:82`), con el presupuesto de bytes subido de forma deliberada.
- El parser de Codex en `tests/pilot/trace.go` registra los items `collab_tool_call`.
- Criterio `no_poll_wait_chain` en `tests/pilot/regression.go`, declarado pero sin conectar a `regressionCriteria` ([design.md](design.md#criterio-no_poll_wait_chain)), con fixtures de CI y de `wait_agent`.
- Un comentario en #36 que explica el cambio de alcance (D2-A).

**Excluye:**
- La regla 3 del issue (reintentar o reportar un spawn fallido por capacidad), por D2-A. Hay un solo caso de tarea descartada en silencio (`01a0c7a7`); en `01a0c10d` el modelo reaccionó a las 11 veces que vio el error. Una regla necesita recurrencia.
- Subir el límite de hilos de Codex (D2-A del issue original).
- Medir el efecto en sesiones reales: los pilotos siguen pausados.

**Criterios de aceptación:**
- A1. `global.md` dice lo siguiente, sin nombrar herramientas de un host:
  - Mientras corre un proceso externo y no queda trabajo independiente, se espera la señal de fin.
  - Si el host la entrega en un turno posterior, el hilo principal termina el turno con el aviso de trabajo en curso.
  - Si no, usa una espera bloqueante reutilizada en tramos de varios minutos, nunca ciclos nuevos de sleep y consulta.
  - Tras cada tramo sin resultado, el hilo principal da al usuario una línea con qué sigue corriendo y qué decide su resultado, en la misma respuesta que la siguiente espera.
  - Conserva lo que ya decía `:82` sobre resultados descartados y sobre el hijo que devuelve su resultado final.
- A2. `globalGuidanceBudget` sube exactamente al tamaño nuevo y `go test ./tests/content/...` pasa.
- A3. Un item `collab_tool_call` de `codex exec --json` produce un evento de herramienta con nombre `collab_<tool>`, y los fixtures existentes conservan su estado.
- A4. `no_poll_wait_chain` falla con los fixtures derivados de las sesiones (CI y subagentes) y pasa con sus pares corregidos. Revertir cada rama del criterio hace fallar su fixture. `regressionCriteria` sigue devolviendo seis criterios.
- A5. `go vet ./...` y `go test -race ./tests/pilot/...` pasan, y el escaneo de fixtures del README no imprime nada.

## Decisiones del usuario (2026-09-26)

- **D1-A:** esperar la señal de fin en tramos de varios minutos, con una línea al usuario tras cada tramo. Une las reglas 1 y 2 del issue.
- **D2-A:** quitar la regla 3 de este cambio y anotar el motivo en el issue.
- **D3-A:** fixtures de CI y extensión del parser de Codex para `collab_tool_call`, con fixtures de `wait_agent`.
- **Ruta:** `flow-plan` en esta sesión.

## Entrega

Respuesta del usuario (2026-09-26):
- **E1-A, `direct-base`:** un commit con las rutas de este cambio y su carpeta `openspec`, con push a `rebuild/harness-engineering`. Después, el commit de cierre con el archivo del cambio, también con push. Sin PR.
- **E2-A, revisión:** sin revisión dedicada del código. Los fixtures y la reversión de cada rama demuestran el criterio, y la redacción de la regla la cubrió la revisión del plan.
- **E3-A, despliegue:** release nueva a los seis hosts tras el push, desde un worktree limpio del commit. Se verifica el texto en los archivos de cada host y se borra el worktree después.
- **D2-A:** autoriza el comentario en #36.

Otra sesión trabaja en paralelo en este repositorio: commiteó `a64fd38` en `flow-build`/`flow-plan` sin tocar `global.md` ni los tests. La entrega commitea solo las rutas de este cambio y hace `git pull --ff-only` antes del push.
