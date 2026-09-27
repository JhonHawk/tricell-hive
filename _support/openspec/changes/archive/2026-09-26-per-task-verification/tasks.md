# Tareas

**Base del cambio:** `42f0160` (`rebuild/harness-engineering`, 2026-09-26).

Este cambio todavía no puede usar `review-task`, porque el rol existe recién después del despliegue. En su lugar, antes del push Codex verifica cada tarea T1–T5 contra sus AC con el contrato de `review-task` (D11-A, decisión del usuario del 2026-09-26): otra familia de modelos, veredicto por AC con evidencia `path:line`, re-ejecución de la verificación y solo lectura. Las tareas pasan a `[x]` solo con veredicto limpio y evidencia comprobada. Después corre `/code-review` sobre el diff (D5-A). Las marcas y la etiqueta `Closes:` ya siguen el formato nuevo.

Estados: `[ ]` pendiente · `[/]` en construcción · `[?]` hecha sin verificar · `[x]` verificada · `[!]` rechazada.

### T1 — Formato del plan, estados y revisión de huérfanos

- [x] La plantilla, los registros de cambio y la revisión del plan siguen [design.md](design.md#criterios-closes-y-estados-de-tarea).

**Closes:** AC1, AC2.
**Depende de:** nada.
**Ubicaciones:**
- `content/skills/flow-plan/references/plan-format.md`: lista «What the document must carry», sección `Scope and acceptance` de la plantilla, bloque de tarea `:68-82`, leyenda de estados y nota de planes heredados.
- `content/skills/flow-plan/references/change-records.md`: fila de `proposal.md` y fila de `tasks.md`.
- `content/skills/flow-plan/SKILL.md:42`.
- `content/skills/flow-plan/references/plan-review.md`: párrafo «Reconcile and close» o la tabla de dominios.

**Ejecución:** hilo principal. La redacción es la decisión, y el brief sería más largo que la edición.
**Verificación:**
- AC1: `rg -n 'AC<n>|Closes:|\[\?\]|\[!\]' content/skills/flow-plan/references/plan-format.md content/skills/flow-plan/references/change-records.md` muestra la definición falsable de los criterios, la etiqueta sin traducir y la leyenda de los cinco estados.
- AC2: `rg -n 'orphan' content/skills/flow-plan/SKILL.md content/skills/flow-plan/references/plan-review.md` muestra el chequeo de preparación y los tres hallazgos bloqueantes.

### T2 — Rol `review-task`

- [x] `content/agents/quality/review-task.md` sigue [design.md](design.md#rol-review-task), y los tests de roles lo cubren.

**Closes:** AC3.
**Depende de:** nada.
**Ubicaciones:** `content/agents/quality/review-task.md` (nuevo); en `integrations/agents/agents_test.go`, las líneas `:28-29` (19 → 20) y el mapa `expected` de `:253-273` (`"review-task": {"high", "medium", "medium"}`).
**Ejecución:** hilo principal. Es un archivo corto, acoplado a la redacción de T3.
**Enfoque de prueba:** `tdd`. Primero los tests, que fallan con «catalogue has 19 roles, want 20»; después el rol.
**Verificación:**
- AC3: `go test -count=1 ./integrations/agents/...` en verde después de haber fallado.
- El render para Claude tiene `model: "opus"` y no lleva `permissionMode: "plan"`, porque el acceso es `verify`.
- Se relee el cuerpo contra los cinco puntos de [design.md](design.md#rol-review-task).

### T3 — `flow-build`: estados, verificación y marcado

- [x] `flow-build` sigue [design.md](design.md#flow-build-estados-verificación-y-marcado).

**Closes:** AC4.
**Depende de:** T1 (estados y etiqueta) y T2 (nombre del rol).
**Ubicaciones:** `content/skills/flow-build/SKILL.md:44` (procedimiento de marcado) y `content/skills/flow-build/references/verification.md:30` (solo el puntero).
**Ejecución:** hilo principal.
**Verificación:**
- AC4: `rg -n 'review-task' content/skills/flow-build/SKILL.md content/skills/flow-build/references/verification.md` muestra el procedimiento en `SKILL.md` y solo el puntero en `verification.md`.
- Cada uno de estos patrones devuelve al menos una línea en `content/skills/flow-build/SKILL.md`: `\[\?\]`, `\[!\]`, `cannot verify`, `blocked:`, `accepted unverified by the user`, `sdd-verify`, `review-code`.
- Se relee el procedimiento contra los puntos de AC4: disparador `tdd`/`check`, esperar la señal de fin, comprobar la evidencia, rondas, bloqueo por prerrequisito, D8-A, D9-A, único escritor del estado, la tarea nativa completa en `[x]`, y que no reemplaza otros gates.
- **Límite:** T1 y T3 son ediciones de guía sin enfoque de prueba, así que no pasan por `review-task` (D3-A). La regla central del cambio no tiene más verificación independiente que `/code-review` en T6.

### T4 — Perfil `reasoning` de OpenCode y documentación de perfiles

- [x] OpenCode `reasoning` usa `github-copilot/claude-opus-5.5`, y la documentación sigue [design.md](design.md#perfil-reasoning-de-opencode).

**Closes:** AC5.
**Depende de:** T2 (recuento de roles).
**Ubicaciones:**
- `integrations/agent-profiles.json`: `hosts.opencode.models.reasoning`.
- `_support/docs/architecture/agent-delivery.md`: `:9` (recuento y lista de roles), `:13-15` (tabla de perfiles), `:28` (fila de esfuerzo), y la nota de cuota y el límite de Grok/Cursor.
- `AGENTS.md:19`.

**Ejecución:** hilo principal.
**Enfoque de prueba:** `check`. `go test ./integrations/...` y la lectura del render.
**Verificación:**
- `jq -r '.hosts.opencode.models | .reasoning.model, .execution.model' integrations/agent-profiles.json` imprime `github-copilot/claude-opus-5.5` y el DeepSeek actual.
- `go test -count=1 ./integrations/...` en verde.
- El render de OpenCode de `review-task` lleva `model: "github-copilot/claude-opus-5.5"`, y el de `sdd-verify` y el de `solution-architect` siguen en DeepSeek.
- `rg -n 'every role on OpenCode|every role on its fixed' AGENTS.md _support/docs/architecture/agent-delivery.md` no devuelve nada.
- AC5, instalado: se comprueba en T6 después del apply.

### T5 — Criterio `task_marked_after_verdict`, parser y fixtures

- [x] El parser, el criterio, los fixtures y los tests siguen [design.md](design.md#criterio-task_marked_after_verdict).
  Ronda 1, AC6 `not met` (Codex), corregido y re-verificado `met`: un `tool_result` fallido de un lanzamiento en primer plano cuenta como fin (`tests/pilot/regression.go:2379` no revisa `Success`), y falta un test explícito de `not_observed` sin marcas.

**Closes:** AC6.
**Depende de:** T2 (nombre del rol). Es independiente de T1, T3 y T4.
**Ubicaciones:**
- `tests/pilot/trace.go` y `tests/pilot/trace_test.go`: eventos de fin de un hijo en segundo plano, para OpenCode y Claude.
- `tests/pilot/regression.go`: la función y el comentario de `regressionCriteria` en `:26-35`.
- `tests/pilot/regression_test.go`: `TestRegressionFixtures` y el test de «sin conectar».
- `tests/fixtures/regression/task_marked_after_verdict/**`.
- `tests/fixtures/regression/README.md`: una fila en la tabla y una nota en «Limits».
- Fuente: sesión OpenCode `ses_f20143eacffe`, seq 1848, 1861, 1876, 1889 y 1992, leída campo por campo con `sqlite3 -readonly` y `json_extract`.

**Ejecución:** delegada a `test-engineer`. La interfaz está fijada aquí y no comparte archivos con T1–T4; así la lectura de `trace.go`, de `regression.go` y de la sesión queda fuera del hilo principal. El brief incluye `tests/fixtures/regression/README.md` y las reglas de saneamiento, y no autoriza commits. El hilo principal revisa el diff y confirma que ningún texto del cliente llegó a los fixtures.
**Enfoque de prueba:** `tdd`. Primero los tests del parser y el registro de fixtures, que fallan porque falta la función; después la implementación.
**Verificación:**
- AC6: `go test -race -count=1 ./tests/pilot/...` en verde:
  - los seis fixtures `*-fail` fallan;
  - `opencode-pass`, `opencode-accepted-pass` y `claude-background-pass` pasan;
  - una traza sin marcas da `not_observed`.
- Reversión por rama, restaurando después, según la tabla de [design.md](design.md#criterio-task_marked_after_verdict): cada una de las cinco ramas revertidas hace dejar de fallar su fixture.

- El escaneo de fixtures del README no imprime nada.
- `TestRegressionCriteriaReturnsAllSix` no cambia.

### T6 — Verificación conjunta, revisión y entrega

- [x] Gates en verde, `/code-review` atendido, entrega y despliegue según [proposal.md](proposal.md#entrega).

**Closes:** AC7.
**Depende de:** T1–T5.
**Ejecución:** hilo principal, dueño de la integración, Git y el despliegue.
**Verificación:**
- **Checks:** `go vet ./...` y `go test -race -count=1 ./...` en verde.
- **Revisión:** `/code-review` sobre el diff, con los hallazgos resueltos según `verification.md`.
- **Commit y push:**
  1. `git pull --ff-only`.
  2. Escáner de secretos del proyecto sin hallazgos.
  3. `git commit --only -- <rutas del cambio>`.
  4. Push. Si el remoto avanzó, rebase antes del push.
- **Release** desde un worktree limpio del commit, según `_support/docs/architecture/deployment-manager.md:16-18`:
  - `go run ./tooling/cli plan install --hosts codex,claude,grok,pi,opencode,cursor --scope user --out <plan>` y después `go run ./tooling/cli apply --plan <plan>`.
  - Antes del apply: `jq -r '."github-copilot".models | has("claude-opus-5.5")' ~/.cache/opencode/models.json` imprime `true`, y `jq -r 'has("github-copilot")' ~/.local/share/opencode/auth.json` imprime `true`.
  - Se registran el ID de la release, el ID del plan y el recuento de recursos.
- **Verificación por host:**
  - AC7: `review-task` instalado en `~/.claude/agents/`, `~/.codex/agents/` (`.toml`), `~/.grok/agents/`, `~/.pi/agent/agents/`, `~/.config/opencode/agents/` y `~/.cursor/agents/`.
  - AC5: `grep -l 'github-copilot/claude-opus-5.5' ~/.config/opencode/agents/*.md` lista los seis roles `reasoning`, y `sdd-verify.md` sigue en DeepSeek.
- **Limpieza:** se borran el worktree y el archivo de plan del manager, no la carpeta del cambio.
- **Cierre:** no hay deltas de specs. La carpeta se archiva con `git mv` y el estado cerrado, en un commit con push.

## Verificación compartida

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Tests de roles, integraciones y pilot | Tras T1–T5, local | `go vet ./...`, `go test -race ./...` en verde | Incluido en la implementación |
| Reversión de ramas del criterio | T5, local | Cada fixture fallido deja de fallar al quitar su rama | Incluido en la implementación |
| Revisión del código | T6, antes del push | `/code-review` de Claude Code sobre el diff | D5-A |
| Despliegue | T6, tras el push | Release a los seis hosts y lectura del rol y del modelo en cada uno | D4-A |
| Verificación in vivo | — | No aplica: la regla actúa en sesiones de modelo y los pilotos están pausados. El efecto se observará en la próxima corrida R1 | — |

## Revisión y progreso

**Revisión del plan, ronda 1** (revisión `11f1806a8292`, HEAD `42f0160`): tres `review-plan` nativos en paralelo, de solo lectura.

- **Guía:** los bloqueantes B1–B4 se aceptaron:
  - B1: `Closes:` queda separado del disparador de `review-task`.
  - B2: la falsabilidad se mide desde la base del cambio, y la preservación pasa a restricciones.
  - B3: `review-task` es parte de la implementación autorizada, no code review.
  - B4: `cannot verify` por prerrequisito deja la tarea bloqueada, y el caso sin verificador lo resuelve D8-A, que decidió el usuario.

  Las sugerencias S1–S6 se aplicaron: rondas propias, sin evidencia duplicada, escritor del estado acotado al build, alcance por **Locations**, etiqueta sin traducir y planes heredados explícitos.
- **Trazas:** los bloqueantes B1–B3 se aceptaron:
  - B1: el parser gana el evento de fin de un hijo en segundo plano.
  - B2: la regla es más débil, sin atribución por tarea.
  - B3: los fixtures prueban el orden, cada uno con su rama.

  Las sugerencias N1–N5 se aplicaron: forma del stream y campos de OpenCode, límites de Codex y de `Write`/`MultiEdit`, hechos corregidos (tres marcas), test de «sin conectar» y extracción por campos.
- **Integración:** los bloqueantes B1–B2 se aceptaron:
  - B1: la lista exacta de roles `reasoning` y la redacción de `AGENTS.md:19`.
  - B2: AC5 falsable desde lo instalado, con la conducta sin créditos no verificada.

  Las sugerencias N1–N8 se aplicaron: `agent-delivery.md:28`, render entre comillas, comandos de release, orden de Git y riesgo del renombre, recuentos, líneas corregidas.

**Cambios posteriores a la ronda, decididos por el usuario:**
- D7-A: cinco estados de tarea.
- D8-A: qué hacer cuando `review-task` no corre.
- D6-E: `github-copilot/claude-opus-5.5` en lugar de `openai/gpt-6-sol#medium`. El plan Pro+ ya está activo.

**Re-revisión** (revisión `75621c007ea6`; los mismos revisores reanudados; única ronda):
- **Guía:**
  - B5–B7 entraron así:
    - B5, salida después de la decisión del usuario: resuelta con D9-A, decisión del usuario.
    - B6, tarea `tdd`/`check` sin `Closes:`: resuelta con D10-A, decisión del usuario.
    - B7, precedencia de `not met` y la línea `blocked:`: la propuesta del revisor, tal cual.
  - N1–N4 aplicadas tal cual:
    - la tarea nativa se completa en `[x]`;
    - se cita `global.md:82`;
    - los planes heredados usan solo `[ ]`/`[x]`;
    - se agregan `rg` por término clave.
  - N5 queda como límite declarado en T3.
- **Trazas:**
  - B4, reversión alineada con cada rama, aplicada tal cual.
  - N6–N12 aplicadas tal cual:
    - consumo por marca;
    - el rechazo consume un fin;
    - falsos positivos declarados;
    - el fin en primer plano de OpenCode;
    - supuestos de formato declarados;
    - el fixture `claude-background-ack-fail`.
  - N9 confirmó que `[?]`→`[x]` y `[!]`→`[x]` funcionan.
  - La excepción de aceptación de D9-A se agregó al criterio con su fixture. Es consecuencia directa de una decisión del usuario, no un cambio de contrato nuevo.
- **Integración:** las correcciones son propuestas del propio revisor. El cambio de modelo (D6-E) se comprueba en T4 y T6.

**Límites que quedan:**
- No se ejecutó ningún test durante la revisión.
- No se verificó que `opencode run --format json` ni `claude -p --output-format stream-json` emitan los eventos de fin; queda declarado como supuesto de formato.
- No se verificó qué hace OpenCode al agotar los créditos de Copilot; lo cubre D8-A.
- El revisor de guía no leyó T4–T6.

**Verificación por tarea (D11-A)**, Codex `gpt-6-sol`, esfuerzo `medium`, sesión `01a0e074`, solo lectura, sobre la base `42f0160` y el working tree:
- **T1 AC1 y AC2: `met`.** Evidencia: `plan-format.md:23`, `change-records.md:9`, `flow-plan/SKILL.md:42`, `plan-review.md:18` y los `rg` de la tarea.
- **T2 AC3: `met`.** Evidencia: `review-task.md` con `reasoning` y `verify`; `agents_test.go:22` exige 20 roles; `go test ./integrations/agents/...` en verde después del RED observado («catalogue has 19 roles, want 20»).
- **T3 AC4: `met`.** Evidencia: `flow-build/SKILL.md:46` y siguientes; `verification.md:30` tiene solo el puntero; los `rg` pasan.
- **T4 AC5: `cannot verify`** en la parte instalada, porque todavía no está desplegada. La parte del repositorio es `met`: `agent-profiles.json:20`, `agent-delivery.md:9`, `AGENTS.md:19`, `jq`, el render comprobado y `go test ./integrations/...`.
- **Evidencia comprobada por el orquestador:** las líneas citadas existen y dicen lo que el veredicto afirma.
- **Hallazgo descartado:** `agents_test.go` no fija el modelo exacto del render. Un test con el literal repetiría la configuración. La regresión que importa, el modelo instalado, la comprueba T6 en cada host.

- **T5 AC6, ronda 1: `not met`.** Codex encontró dos cosas:
  - Un `tool_result` fallido de un lanzamiento en primer plano contaba como fin.
  - Faltaba un test de `not_observed`.

  El orquestador lo confirmó en `regression.go:2379`, y `test-engineer` corrigió con TDD:
  - RED: `claude-foreground-failed-fail.jsonl: got pass want fail`.
  - La condición `Success` quedó en `regression.go:2384`.
  - Se agregaron los fixtures `claude-foreground-failed-fail` y `claude-foreground-pass`, y el test `TestTaskMarkedAfterVerdictEndWithoutMarksIsNotObserved`.
  - La reversión de la condición afecta solo al fixture nuevo.
- **T5 AC6, re-verificación: `met`.** Codex citó el fixture con `is_error:true`, `trace.go:381` y el test nuevo, y `go test -race ./tests/pilot/...` pasó. El orquestador comprobó esa evidencia.
- **Saneamiento:** se quitaron el slug del cambio de sample-project de `provenance.md` y de la fila del README, y los identificadores de código del cliente de `design.md`, porque el repositorio es público.

**`/code-review`** (Claude Code, nivel medio, sobre el working tree): dos hallazgos, ambos corregidos en el hilo principal.
- **H1, P2, `regression.go`:** una edición fallida de `tasks.md` contaba como marca y consumía el veredicto, así que un reintento correcto daba `fail`.
  - Se agregó el fixture `claude-failed-mark-retry-pass`, con RED `got fail want pass`.
  - `isTasksMarkdownEdit` ahora excluye las ediciones con `Success=false`.
  - Revertir esa exclusión hace fallar el fixture; se restauró.
- **H2, P3, `flow-build/SKILL.md`, paso 3:** decía «leave `[!]`» cuando la tarea ya estaba en `[?]`. Ahora dice «set `[!]` again with the round-2 line».
- **Sin re-revisión:** son arreglos P2/P3 que no cambian un contrato, según `verification.md`.

**Entrega (T6):**
- **Checks:** `go vet ./...` y `go test -race -count=1 ./...` en verde; `gitleaks` sin hallazgos en las 32 rutas.
- **Commit y push:** `88d24ac`, con `git commit --only` de las 32 rutas del cambio, y push a `rebuild/harness-engineering` (cabeza remota `88d24ac`). La carpeta `_support/sessions/2026-09-26-hive-agent-names/` de la otra sesión quedó fuera.
- **Release** `ea03e56602d3`:
  - Plan `300ca78fbc83`, desplegado desde un worktree limpio de `88d24ac`: 437 recursos `installed`.
  - Dos `apply` repetidos por error con el mismo plan fueron rechazados por el manager («stale ownership»), sin efecto.
- **AC7 `met`:** `review-task` instalado en `~/.claude/agents/`, `~/.codex/agents/` (`.toml`), `~/.grok/agents/`, `~/.pi/agent/agents/`, `~/.config/opencode/agents/` y `~/.cursor/agents/`.
- **AC5 `met`:**
  - `grep -l 'github-copilot/claude-opus-5.5' ~/.config/opencode/agents/*.md` lista `database-specialist`, `performance-engineer`, `review-code`, `review-harness`, `review-plan` y `review-task`.
  - `sdd-verify` sigue en `opencode-go/deepseek-v4.1-flash#max`.
  - T4 sale del bloqueo y pasa a `[x]`.
- **Limpieza:** se borraron el worktree y el archivo de plan del manager.

**Estado:** cerrado. Integrado en `rebuild/harness-engineering` (`88d24ac`) y desplegado como release `ea03e56602d3`.
