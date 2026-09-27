# Diseño

## Contexto verificado

Inspeccionado el 2026-09-26 sobre `rebuild/harness-engineering` (`42f0160`).

**Experimento R1, ARK-730 (sample-project):**
- **Sesiones:** el plan lo hizo Codex `gpt-6-sol` (sesión `01a0dfcd`). El build, DeepSeek v4.1 Flash `#max` en OpenCode (sesión `ses_f20143eacffe`), sin cambio de modelo.
- **Revisión de Opus** (`review-code`, sobre `345941a2e`): 0 P0/P1 y 3 P2, verificados en el hilo principal.
  - H1: un script de shell del proyecto expande un arreglo vacío bajo `set -u` y aborta con `/bin/bash` 3.2 (reproducido).
  - H2: tres rutas nuevas propagan el error de firma del proveedor de almacenamiento al cliente, y falta el test de firma fallida que el plan pedía.
  - H3: la migración termina con código 0 aunque queden datos sin migrar, y el plan pedía abortar.
- **Dónde vivían H2 y H3:** en el criterio 4 de `proposal.md` y en `design.md:26` de ese cambio, pero no en la verificación de T2 ni de T4. `content/skills/flow-plan/SKILL.md:42` ya exige «Every acceptance criterion has an implementing task and a concrete verification».
- **Marcas sin veredicto:** el hilo principal lanzó `sdd-verify` y `review-ux` en segundo plano (seq 1848, herramienta `subagent`, `background=true`). El acuse del lanzamiento llegó en 0,5 s. Después marcó tres tareas en `tasks.md` (seq 1861, 1876 y 1889, a los 9, 16 y 30 s). El fin real de `sdd-verify` llegó unos 33 min después, como mensaje `synthetic` con `metadata.source=subagent` (seq 1992). Las marcas no esperaron ningún veredicto. La edición de OpenCode (`edit`, campos `path`, `oldString` y `newString`) no contiene el `T<n>` de la tarea.

**Hive hoy:**
- **Marcado:** las tareas tienen ID `T<n>` (`content/skills/flow-plan/references/plan-format.md:68`), y el implementador marca «only when its expected result is evidenced» (`content/skills/flow-build/SKILL.md:44`).
- **Verificación por cambio, no por tarea:**
  - `sdd-verify` y `review-ux` corren una vez por cambio con UI y «need no separate approval» (`content/skills/flow-build/references/verification.md:59`).
  - `review-code` corre en la publicación (`verification.md:32,41`).
  - La iteración de build dice «not a dedicated independent code-review gate» (`verification.md:30`).
  - `global.md:71` pide confirmar el alcance antes de una revisión de código dedicada.
- **Tareas mecánicas y enfoques de prueba:** `global.md:77` define el cambio mecánico, incluida «a documentation edit». Los enfoques son `tdd`, `characterization` (refactor que conserva comportamiento) y `check` (`flow-plan/SKILL.md:34`, `plan-format.md:78`).
- **Perfiles de los roles:**

  | Perfil | Roles |
  | --- | --- |
  | `reasoning` | `database-specialist`, `performance-engineer`, `review-code`, `review-harness`, `review-plan` |
  | `inherit` | `solution-architect`, `visual-designer`, `review-refuter`, `review-security` |
  | `execution` | los otros diez |

  En OpenCode los tres perfiles son `opencode-go/deepseek-v4.1-flash#max` (`integrations/agent-profiles.json`). `AGENTS.md:19` dice «every role on OpenCode», y `agent-delivery.md:28` dice «OpenCode runs every role on its fixed `#max` variant».
- **Renderer:** acepta cualquier modelo sin `\r`, `\n` ni `\x00` (`integrations/agents/agents.go:151`) y lo escribe entre comillas para OpenCode (`agents.go:264-265`), por ejemplo `model: "opencode-go/deepseek-v4.1-flash#max"`.
- **Tests de roles:** `integrations/agents/agents_test.go:28-29` fija 19 roles, y el mapa `expected` de `:253-273` fija el esfuerzo por rol y, a través de `len(expected)`, también el recuento.
- **Parser de trazas:**
  - `tests/pilot/trace.go` lee el stream de `opencode run --format json` (`part.tool`, `part.state.input`, `callID`). El export de OpenCode solo se usa para comprobar el cierre (`trace.go:131-196`). La rama de OpenCode (`trace.go:571-588`) procesa `tool_use`, `text` y `step_finish`.
  - Un mensaje con contenido de texto plano no produce eventos (`trace.go:372`). Por eso hoy el parser no ve el aviso de fin de un hijo en segundo plano: ni el mensaje `synthetic` de OpenCode ni el `<task-notification>` de Claude.
  - `file_change` de Codex no conserva `Input` (`trace.go:515`).
  - `go test ./tests/pilot/...` pasa.
- **Modelo de OpenCode:**
  - `openai/gpt-6-sol` corrió en OpenCode el 2026-09-25 (sesión `ses_f242d1b48ffe`) y se cortó por cuota de OpenAI. El usuario prefiere reservar esa cuota para Codex nativo.
  - El catálogo local `~/.cache/opencode/models.json` lista `github-copilot/claude-opus-5.5` con `reasoning: true` y sin variantes. Por eso el modelo va sin sufijo `#`.
  - `auth.json` de OpenCode tiene la clave `github-copilot`; solo leí el nombre.
  - El usuario mostró su plan Copilot Pro Plus activo: 0 de 7 000 créditos, con renovación el 2026-09-30.
  - El proveedor `github-copilot` se usó en OpenCode en 29 sesiones, hasta el 2026-08-16.
  - Según los [planes de Copilot](https://docs.github.com/en/copilot/get-started/plans) (consultados el 2026-09-26), Opus 5.5 requiere Pro+ ($39, 7 000 créditos al mes). La página no dice nada sobre clientes de terceros ni sobre qué pasa al agotar los créditos.
  - La sintaxis `provider/model#variant` coincide con la documentación de OpenCode V2 (Context7 `/websites/opencode_ai_v2`); versión instalada: 2.0.18.
  - No se verificó qué hace OpenCode con un modelo que no resuelve o sin cuota.

**Referencias externas** (investigación del 2026-09-26):
- **optional reference project** (`subagent-driven-development`): un revisor nuevo por tarea, con evidencia `archivo:línea`. El controlador marca solo con un veredicto limpio. Admite como problema sin resolver los requisitos que se pierden al extraer el texto de la tarea.
- **optional reference project** (`a9e36e9b`):
  - El verificador independiente es obligatorio cuando el implementador corrió en un perfil chico (`internal/assets/skills/_shared/odd-orchestrator-sections.md:19,29-30`).
  - «A task checkbox is not review authority» (`persistence-contract.md:11`).
  - En todos sus presets, el verificador tiene un nivel igual o superior al del implementador.
- **optional reference project** (`c55ee46`):
  - Los criterios son checkboxes sin ID que nadie marca.
  - La revisión en la misma sesión es «confirmation bias with a slash command» (`docs/engineering/code-review.md:60`).
  - «Sub-agent output is a hypothesis, not evidence» (`:66-68`).
  - «The acceptance criteria graded nothing: some passed before any work was done» (`docs/engineering/to-tickets.md:76-77`), que allí quedó como consejo manual.
- **Anthropic y la literatura:** el harness de tareas largas de Anthropic deja que el mismo agente marque `passes` y reconoce que marca antes de tiempo. La investigación sobre autoverificación y autopreferencia favorece un verificador nuevo y de otra familia.

## Enfoque

### Criterios, `Closes:` y estados de tarea

- **En `proposal.md`, los criterios `AC<n>`:** se numeran desde 1, con un comportamiento observable cada uno. Cada uno nombra la observación que lo mostraría falso, y esa observación es falsa en el commit base del cambio. El build registra ese commit en `tasks.md` al empezar.
  - **No son criterios:** las restricciones de preservación («X sigue pasando»); van en una lista aparte de restricciones.
  - **No califican nada:** un criterio que ya se cumple en la base, que solo cumple otro cambio o que repite el pedido.
- **En `tasks.md`, la línea `Closes:`:** toda tarea que satisface un criterio lleva `**Closes:** AC<n>, …`, incluidas las de documentación o entrega. Su **Verification** cubre cada criterio que cierra.
  - **Sin traducir:** `Closes:`, `AC<n>` y `T<n>` quedan sin traducir en cualquier idioma del plan. Así el verificador y el criterio de pilot los encuentran.
- **Estados de tarea** (D7-A):

  | Marca | Estado | Transición |
  | --- | --- | --- |
  | `[ ]` | Pendiente | Al crear el plan |
  | `[/]` | En construcción | Al asignar o empezar la tarea |
  | `[?]` | Hecha, sin verificar | Cuando el implementador reporta una tarea que dispara `review-task`. Si está bloqueada, lleva debajo la línea `blocked: <prerequisite>` |
  | `[x]` | Verificada | Con veredicto limpio, o, en tareas que no disparan `review-task`, con la evidencia de su check. Si el usuario la aceptó sin verificar (D9-A), lleva debajo la línea `accepted unverified by the user: <reason>` |
  | `[!]` | Rechazada | Con algún `not met`. Debajo va una línea con la ronda, el `AC<n>` y el motivo |

  Qué pasa después de una decisión del usuario (D9-A) sobre un `[!]` definitivo o un `[?]` sin verificador:
  - **Si la acepta:** pasa a `[x]` con la línea de aceptación.
  - **Si pide rehacerla:** vuelve a `[/]`.
  - **Si la retira:** sale de la lista de tareas y queda una línea con su motivo en la sección de progreso.

  Se evita `[-]`, porque en la convención más difundida de tareas en Markdown significa cancelada. GitHub y VS Code solo muestran como checkbox `[ ]` y `[x]`; las demás marcas se leen como texto.
- **Preparación del plan:** `flow-plan/SKILL.md:42` pasa a exigir dos cosas. Todo `AC<n>` está en el `Closes:` de al menos una tarea, cuya verificación lo cubre. Y toda tarea `tdd` o `check` lleva `Closes:` (D10-A): si no cierra ningún criterio, o falta un criterio o la tarea sobra.
- **Revisión del plan:** en `plan-review.md`, cada revisor señala como bloqueantes tres casos: criterios huérfanos, tareas cuya verificación no cubre lo que cierran, y criterios que ya se cumplen en la base.
- **Planes heredados:** un plan sin `AC<n>` no entra en la verificación por tarea. Usa solo `[ ]` y `[x]`, y el orquestador marca con la evidencia de cada check, como hasta ahora. `plan-format.md` lo dice.

### Rol `review-task`

`content/agents/quality/review-task.md`, con `model_profile: "reasoning"` y `access_profile: "verify"`. `verify` conserva los permisos del host porque re-ejecutar tests escribe artefactos, igual que en `sdd-verify`. El cuerpo va en inglés, escrito para el modelo menos capaz, y solo enlaza recursos con localizadores `skill:owner/path` (`tooling/management/references.go:112`):

1. **Entrada:** la ruta de la carpeta del cambio, el `T<n>` y la base del diff. Lee él mismo `tasks.md`, los `AC<n>` del `Closes:` en `proposal.md` y las secciones de `design.md` que la tarea enlaza. Si el padre incluye texto parafraseado, prevalecen los archivos.
2. **Qué revisa:** el diff de las rutas en **Locations** de la tarea, contra la base. Si hay trabajo sin commitear de otras tareas en esas rutas, lo declara como límite.
3. **Verificación:** re-ejecuta la **Verification** de la tarea, dentro de los efectos que asigna el padre. No hace recorridos in vivo ni de UI; esos son de `sdd-verify` y `review-ux`.
4. **Veredicto:** uno por criterio: `met`, `not met` o `cannot verify`. Cada uno lleva evidencia `path:line` o la salida del comando.
   - **Falla del plan:** si la verificación de la tarea no cubre un criterio que la tarea cierra, es `not met` por el plan.
   - **Prerrequisito:** `cannot verify` nombra el prerrequisito que falta (servicio, credencial, entorno).
5. **Límites:** no edita fuentes, tests ni la carpeta del cambio, y no delega. Devuelve el veredicto y sus límites al padre.

Esfuerzo esperado en `agents_test.go`: `{"high", "medium", "medium"}` (Claude, Codex, Pi), igual que `review-code`.

### `flow-build`: estados, verificación y marcado

La frase de marcado de `content/skills/flow-build/SKILL.md:44` se reemplaza por este procedimiento. Aplica en un plan con criterios `AC<n>`:

- **Quién escribe el estado:** durante el build, solo el orquestador cambia las marcas y el estado de las tareas. Los briefs de los hijos no les asignan esas líneas.
- **Disparador:** `review-task` se lanza para cada tarea con enfoque de prueba `tdd` o `check`. `characterization`, las tareas sin enfoque y las de entrega pasan de `[/]` a `[x]` con la evidencia de su check.
- **Ciclo:**
  1. Cuando el implementador reporta, la tarea pasa a `[?]`.
  2. El orquestador lanza `review-task` con la carpeta del cambio, el `T<n>` y la base, y espera su señal de fin, según la regla de espera de `global.md:82`, sin repetirla. El acuse de un lanzamiento en segundo plano no es el veredicto.
  3. Con todo `met`, comprueba puntualmente la evidencia citada (abre el `path:line` o lee la salida) y marca `[x]`.
  4. Con algún `not met`, aunque otros criterios sean `cannot verify`, marca `[!]` con la línea de motivo y devuelve los hallazgos al implementador para una ronda de corrección. Después la tarea vuelve a `[?]` para una sola re-verificación. Si vuelve a haber un `not met`, queda en `[!]` como decisión del usuario.
  5. Un `cannot verify` por un prerrequisito faltante, sin ningún `not met`, en la primera verificación o en la re-verificación, deja la tarea en `[?]` con la línea `blocked: <prerequisite>` y la reporta como bloqueada, sin ronda de corrección.
  6. Si `review-task` no puede correr (cuota, modelo no disponible), la tarea queda en `[?]` y el orquestador pregunta al usuario (D8-A). Las opciones son esperar o reintentar, verificar con otro modelo que elija el usuario, o aceptarla sin verificar. Nunca la sustituye por un hijo del mismo modelo que el implementador.
  7. La decisión del usuario sigue D9-A: aceptada pasa a `[x]` con la línea de aceptación, rehacer vuelve a `[/]`, y retirada sale de la lista con su motivo.
- **Lista de tareas nativa:** `flow-build/SKILL.md:46` completa la tarea nativa cuando la tarea del plan llega a `[x]`, no a `[?]`.
- **Autorización y alcance:** `review-task` es parte de la implementación autorizada y no requiere una aprobación aparte. Verifica cumplimiento del plan, no es code review, y no reemplaza a `review-code` en la publicación ni a `sdd-verify` y `review-ux` en el fin del cambio.

Estas rondas son del verificador de tareas y no cambian las de `verification.md:41`, que rigen la re-revisión de hallazgos de code review. `verification.md:30`, la fila de iteración del build, agrega solo un puntero a este procedimiento.

Límite conocido: en Grok y Cursor el perfil `reasoning` es `inherit`, así que ahí verifica el modelo de la sesión. Se documenta en `agent-delivery.md`.

### Perfil `reasoning` de OpenCode

`hosts.opencode.models.reasoning.model` pasa a `github-copilot/claude-opus-5.5` (D6-E); `execution` e `inherit` siguen en DeepSeek.
- **Comprobación antes del apply:** el catálogo local sigue listando el modelo, y `auth.json` sigue teniendo la clave `github-copilot` (solo el nombre).
- **Sin verificar:** qué hace OpenCode al agotar los créditos o si el modelo no resuelve. Esos casos quedan cubiertos por D8-A.
- **Qué cambia en OpenCode:** `review-task`, `review-code`, `review-plan`, `review-harness`, `database-specialist` y `performance-engineer` corren en Opus 5.5. El hilo principal y los roles `execution` e `inherit`, incluido `solution-architect`, siguen en DeepSeek.
- **`AGENTS.md:19`:** pasa a decir que los modelos `execution` corren los roles delegados en cada host que fija uno, y el hilo principal y todo rol no `reasoning` en OpenCode.
- **`agent-delivery.md`:**
  - `:9` sube a veinte roles con `review-task`.
  - `:13-15` actualiza la fila `reasoning` de OpenCode.
  - `:28` corrige «every role on its fixed `#max` variant».
  - Una nota del riesgo de cuota enlaza con D8-A.

### Criterio `task_marked_after_verdict`

Va en `tests/pilot/regression.go`, declarado y sin conectar a `regressionCriteria`, como `flowSkillReadBeforeDelivery` y `noPollWaitChain`, porque solo aplica a builds con un plan que tiene criterios. El comentario de `regressionCriteria` lo menciona.

**Extensión del parser** (`tests/pilot/trace.go`):
- **OpenCode:** el lanzamiento es la herramienta `subagent` con `agent` y `background` en `Input`.
  - **Primer plano** (`background` no verdadero): el fin es el propio `tool_use` con estado completo.
  - **Segundo plano:** el fin es el mensaje `synthetic` con `metadata.source=subagent`, `agent` y `state=completed`. Se agrega un evento de fin con el rol.
  - **Supuesto de formato, más fuerte de lo que parece:** en la base de datos V2 el mensaje `synthetic` no es un `part`, y el parser de stream decide por `type` y `part`. El fixture tiene que inventar el tipo de ese evento, no solo suponer que `opencode run --format json` lo emite. Se declara en `provenance.md`, igual que `no_poll_wait_chain/provenance.md`.
- **Claude:** el lanzamiento es `Agent` o `Task` con `subagent_type` y `run_in_background`. En primer plano, el fin es el `tool_result` del lanzamiento. En segundo plano, es el mensaje de usuario con `origin.kind=task-notification`, y para ese caso se agrega un evento de fin.
  - **Supuesto de formato:** `origin.kind` solo se observó en transcripciones interactivas de Claude Code 2.1.283, no en `claude -p --output-format stream-json`. Se declara.

**Regla** (no puede atribuir la marca a una tarea: la edición no trae su `T<n>`):
- **Qué es una marca:** una edición en un `Path` que termina en `tasks.md`, en la que sube el número de líneas que empiezan (tras la sangría) por `- [x]` en minúscula, entre el texto viejo y el nuevo. Cada unidad de subida es una marca. Una subida de líneas `- [!]` es un rechazo.
  - Campos: en OpenCode, `oldString` y `newString`; en Claude `Edit`, `old_string` y `new_string`.
- **Consumo:** cada marca y cada rechazo consumen un fin de `review-task`. Si una edición marca varias tareas, consume un fin por cada una.
- **Excepción:** una marca cuya línea nueva va seguida de `accepted unverified by the user` (D9-A) no requiere ni consume un fin.
- **Falla** si alguna marca llega sin un fin de `review-task` sin consumir.
- **Pasa** si toda marca tiene su fin sin consumir.
- **No observado** si no hay marcas.
- **Límites que el doc comment declara:**
  - No atribuye marcas a tareas.
  - Dan falsos positivos, así que el criterio se lee junto con el plan:
    - las tareas que no disparan `review-task`;
    - agregar una tarea que ya viene en `[x]`;
    - reescribir el plan.
  - Un `cannot verify` deja la tarea en `[?]`, así que no consume su fin, y ese fin puede validar una marca posterior.
  - `[X]` y `* [x]` no cuentan.
  - `file_change` de Codex, y `Write` y `MultiEdit` de Claude, dan `not_observed`.

**Fixtures** (`tests/fixtures/regression/task_marked_after_verdict/`):
- **`opencode-fail`:** se lanza `review-task` en segundo plano, llega el acuse y la marca va antes del fin. El rol se sustituyó de `sdd-verify` a `review-task`, y así se declara.
- **`opencode-no-launch-fail`:** marca sin ningún lanzamiento.
- **`opencode-double-mark-fail`:** dos marcas en ediciones separadas, con un solo fin.
- **`opencode-batch-mark-fail`:** una edición que marca dos tareas, con un solo fin.
- **`opencode-reject-then-mark-fail`:** un fin consumido por un rechazo `[!]`, y después una marca sin otro fin.
- **`opencode-pass`:** el fin y después la marca.
- **`opencode-accepted-pass`:** una marca con la línea de aceptación del usuario y sin fin.
- **`claude-background-ack-fail`:** la marca después del acuse y antes de la notificación de fin.
- **`claude-background-pass`:** la notificación de fin y después la marca.
- **`provenance.md`:** la procedencia de cada uno y los supuestos de formato.

Se derivan de la sesión con `sqlite3 -readonly` y `json_extract`, campo por campo, y se traducen a la forma del stream. No se usa `opencode session export`, que vuelca la conversación completa. Van sanitizados según `tests/fixtures/regression/README.md:16-24`: la ruta genérica `openspec/changes/example-change/tasks.md`, líneas de checkbox genéricas, sin texto del cliente ni secretos.

**Tests:**
- `TestTaskMarkedAfterVerdictNotWiredIntoRegressionCriteria`, como `regression_test.go:795,885`.
- Tests del parser para los dos eventos de fin nuevos.
- Reversión por rama, cada una con su fixture:

  | Rama revertida | Fixture que deja de fallar |
  | --- | --- |
  | contar el acuse de un lanzamiento en segundo plano como fin | `opencode-fail` y `claude-background-ack-fail` |
  | quitar la exigencia de un fin previo | `opencode-no-launch-fail` |
  | quitar el consumo | `opencode-double-mark-fail` |
  | consumir un fin por edición, no por marca | `opencode-batch-mark-fail` |
  | no consumir en el rechazo | `opencode-reject-then-mark-fail` |

## Alternativas descartadas

- **Ampliar `sdd-verify`** (D2-B): pasaría todos los recorridos in vivo y de UI a `reasoning`.
- **Reusar `review-code`** (D2-C): en Claude usa el modo plan, así que no re-ejecuta comandos, y su foco son defectos, no el cumplimiento del plan.
- **Verificar todas las tareas** (D3-B): costo sin beneficio en tareas mecánicas.
- **Que el verificador marque:** habría dos escritores del plan. optional reference project y optional reference project dejan el marcado al padre.
- **Atribuir la marca por tarea en el criterio:** la edición no trae el `T<n>`, y la traza no conserva el contenido del archivo.
- **Un sustituto del mismo modelo cuando `review-task` no corre** (D8-C): perdería la independencia que motiva el cambio.
