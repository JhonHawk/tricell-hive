# Verificación por tarea: criterios con ID, estados de tarea y un verificador en otro modelo

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `rebuild/harness-engineering` (`88d24ac`) y desplegado como release `ea03e56602d3` |
| Tracker · GitHub Issues | Sin issue: el usuario prefirió no abrirlo (2026-09-26) |
| Git | `direct-base` a `rebuild/harness-engineering` · sin PR · despliegue a los seis hosts tras el push |
| Verificación | `go test ./integrations/...` · `go test -race ./tests/pilot/...` · `go vet ./...` · `/code-review` antes del push · render instalado en los seis hosts |
| Siguiente paso | Ninguno en este cambio. El efecto en sesiones reales se observará en la próxima corrida R1 |

## Objetivo

En el experimento R1 (2026-09-26), Codex `gpt-6-sol` planeó ARK-730 de sample-project y DeepSeek v4.1 Flash lo construyó en OpenCode. DeepSeek resolvió bien lo que el plan dejó abierto, pero omitió dos requisitos que el plan sí traía: un 5xx neutral con test de firma fallida (H2) y abortar la migración ante escrituras antiguas (H3). Ningún revisor de la sesión los detectó, porque en OpenCode todos los perfiles son DeepSeek; una revisión de Opus sí los encontró. Además, el orquestador marcó tres tareas como `[x]` entre 9 y 30 segundos después de lanzar el verificador, sin esperar su resultado.

La causa tiene dos partes. Primero, H2 y H3 estaban en el criterio 4 de `proposal.md` y en `design.md` de ese cambio, pero no en la verificación de ninguna tarea, aunque `flow-plan` ya exige que cada criterio tenga una tarea y una verificación. Segundo, quien marca `[x]` es el propio implementador; la verificación independiente llega por cambio, no por tarea, y en OpenCode ni siquiera corre en otro modelo.

El cambio hace tres cosas:
- **Criterios trazables:** los criterios de aceptación llevan ID y deben poder fallar. Cada tarea declara cuáles cierra.
- **Estados de tarea:** las tareas tienen cinco estados, y solo el orquestador los cambia.
- **Verificador por tarea:** un rol nuevo, `review-task`, en el perfil `reasoning`, verifica cada tarea que cambia comportamiento contra sus criterios, leyéndolos de los archivos canónicos. Devuelve un veredicto por criterio con evidencia `path:line`, y la tarea se marca verificada solo con un veredicto limpio.

El perfil `reasoning` de OpenCode deja de ser DeepSeek.

## Alcance y aceptación

**Incluye:**
- Formato del plan (`plan-format.md`, `change-records.md`): criterios `AC<n>` falsables, la etiqueta `Closes:` y los cinco estados de tarea.
- `flow-plan`: chequeo de preparación y revisión del plan con tres hallazgos bloqueantes nuevos.
- Rol nuevo `review-task` en `content/agents/quality/`.
- `flow-build`: estados, despacho de `review-task`, marcado por el orquestador, rondas, bloqueos y el caso sin verificador disponible.
- `integrations/agent-profiles.json`: modelo `reasoning` de OpenCode, y la documentación de roles y perfiles.
- Criterio de regresión `task_marked_after_verdict`, con la extensión del parser que necesita.

**Excluye:**
- Verificar tareas de `characterization`, mecánicas o de documentación (D3-A): el orquestador las marca con su check.
- Planes sin criterios `AC<n>` (heredados o `<topic>.plan.md` antiguos): quedan fuera de la verificación por tarea, y así se dice de forma explícita.
- Recorridos in vivo y de UI: siguen en los gates de fin de cambio (`sdd-verify`, `review-ux`).
- Cambiar sample-project o su plan de ARK-730: es otro proyecto.
- Medir el efecto en sesiones reales: los pilotos siguen pausados. Se observará en la próxima corrida R1.

**Restricciones que se preservan (no son criterios):**
- `go vet ./...` y `go test -race ./...` siguen en verde.
- La entrega commitea solo las rutas de este cambio.
- `TestRegressionCriteriaReturnsAllSix` no cambia.

**Criterios de aceptación** (la base del cambio es el commit donde empieza el build, que se registra en `tasks.md`):
- AC1. `plan-format.md` y `change-records.md` definen:
  - Criterios `AC<n>` que nombran la observación que los mostraría falsos, falsa en el commit base del cambio.
  - La etiqueta `Closes:` sin traducir en cada tarea que satisface un criterio.
  - Los estados `[ ]`, `[/]`, `[?]`, `[x]` y `[!]` con su significado.

  *Falso en la base si* `rg -n 'AC<n>|Closes:|\[\?\]' content/skills/flow-plan/references/plan-format.md` no devuelve nada.
- AC2. La preparación de `flow-plan` y `plan-review.md` tratan como bloqueantes tres casos: criterios huérfanos, tareas cuya verificación no cubre lo que cierran, y criterios que ya se cumplen en la base. *Falso en la base si* `rg -n 'orphan' content/skills/flow-plan/` no devuelve nada.
- AC3. Existe `content/agents/quality/review-task.md` con perfil `reasoning` y acceso `verify`, y los tests de roles cuentan 20 roles. *Falso en la base si* `go test ./integrations/agents/...` con el test ya actualizado falla con «catalogue has 19 roles, want 20».
- AC4. `flow-build` define:
  - Los estados y sus transiciones.
  - El disparador de `review-task`: tareas con enfoque `tdd` o `check`.
  - Que el orquestador espera el veredicto y comprueba la evidencia citada antes de `[x]`.
  - Las rondas: `[!]`, una corrección y una re-verificación, y luego el usuario.
  - Que un `cannot verify` por un prerrequisito faltante deja la tarea bloqueada, sin ronda.
  - Qué pasa si `review-task` no puede correr (D8-A).
  - Que durante el build el orquestador es el único que cambia el estado de las tareas.
  - Que `review-task` es parte de la implementación autorizada y no reemplaza a `review-code`, `sdd-verify` ni `review-ux`.

  *Falso en la base si* `rg -n 'review-task' content/skills/flow-build/SKILL.md` no devuelve nada.
- AC5. En OpenCode, los seis roles `reasoning` instalados llevan `model: "github-copilot/claude-opus-5.5"`, y los roles `execution` e `inherit` siguen en DeepSeek. `agent-delivery.md` (`:9`, `:13-15`, `:28`) y `AGENTS.md:19` describen eso. *Falso en la base si* `grep -l 'claude-opus-5.5' ~/.config/opencode/agents/*.md` no devuelve nada.
- AC6. `task_marked_after_verdict` falla con las trazas fallidas derivadas de ARK-730 y pasa con sus pares. La reversión de cada rama hace pasar su fixture fallido. *Falso en la base si* el criterio no existe.
- AC7. La release desplegada instala `review-task` en los seis hosts. *Falso en la base si* no existe ningún `review-task*` en los directorios de agentes de los hosts; hoy no hay ninguno.

## Decisiones del usuario (2026-09-26)

- **D1-A:** alcance completo: trazabilidad de criterios, verificador por tarea, el orquestador marca y OpenCode `reasoning` fuera de DeepSeek.
- **D2-A:** rol nuevo `review-task` (perfil `reasoning`, acceso `verify`); `sdd-verify` sigue en `execution` para los recorridos in vivo.
- **D3-A:** solo las tareas que cambian comportamiento pasan por el verificador. Se concreta como enfoque de prueba `tdd` o `check`.
- **D6-E** (reemplaza a D6-C): el perfil `reasoning` de OpenCode usa `github-copilot/claude-opus-5.5`, con el plan Copilot Pro+ del usuario, ya activo el 2026-09-26 (7 000 créditos al mes, renovación el 30 de septiembre). La cuota de OpenAI queda para Codex nativo.
- **D7-A:** cinco estados de tarea: `[ ]` pendiente, `[/]` en construcción, `[?]` hecha sin verificar, `[x]` verificada, `[!]` rechazada con una línea de motivo. Se evita `[-]` porque se lee como cancelada.
- **D8-A:** si `review-task` no puede correr, la tarea queda en `[?]` y el orquestador pregunta al usuario si esperar o reintentar, verificar con otro modelo que el usuario elija, o aceptar la tarea sin verificar. Nunca la sustituye por un hijo del mismo modelo que el implementador.
- **D9-A:** cuando el usuario decide sobre un `[!]` definitivo o un `[?]` sin verificador:
  - si la acepta, pasa a `[x]` con la línea `accepted unverified by the user: <reason>`;
  - si pide rehacerla, vuelve a `[/]`;
  - si la retira, sale de la lista con su motivo.
- **D10-A:** toda tarea `tdd` o `check` lleva `Closes:`. Si una no cierra ningún criterio, o falta un criterio o la tarea sobra.
- **D11-A:** antes del push, Codex verifica T1–T5 contra sus AC con el contrato de `review-task`, porque `/code-review` corre en la misma familia de modelos que el implementador (Claude). Después sigue `/code-review` (D5-A).
- **Sin issue:** el usuario prefirió no abrir uno por ahora.

## Entrega

Respuesta del usuario (2026-09-26):
- **D4-A, `direct-base`:** un commit con las rutas de este cambio y su carpeta `openspec`, push a `rebuild/harness-engineering`, sin PR. Después, release a los seis hosts desde un worktree limpio del commit, verificación del rol y del modelo en cada host, y borrado del worktree y del archivo de plan del manager. Al final, el commit de cierre con el archivo del cambio, también con push.
- **D5-A, revisión:** `/code-review` de Claude Code sobre el diff antes del push.

**Concurrencia:** otra sesión commitea en paralelo en este repositorio (`42f0160`), y tiene sin aplicar una propuesta para renombrar roles a `hive-<role>` (`_support/sessions/2026-09-26-hive-agent-names/`). Para convivir con eso:
- La entrega hace `git pull --ff-only` antes del commit.
- Commitea con `git commit --only -- <rutas>`.
- Si el remoto avanzó, rebasa antes del push.
- Si el renombre llega antes, este cambio adopta el nombre nuevo del rol al rebasar.
