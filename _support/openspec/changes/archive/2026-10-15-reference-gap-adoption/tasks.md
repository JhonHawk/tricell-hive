# Tareas

Base del build: `f8116af5` (`feat/flash-class-guidance`, worktree `.claude/worktrees/feat-flash-class-guidance`), 2026-10-14.

Todas las tareas se ejecutan en el hilo principal: cada una son una a cinco frases y un brief sería más largo que el cambio. Enfoque de prueba `check` en todas, con las observaciones `rg` de cada AC y `go test ./tests/content/...`.

### T1 — Herencia de contexto del hijo (H1)

- [x] `global.md` y `delegation.md` dejan de afirmar que el hijo no hereda; dicen que se lanza sin la historia del padre cuando la herramienta lo permite y conservan la razón de pasar rutas explícitas. `agent-delivery.md` registra el hecho de Codex.

**Closes:** AC1.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md:80`; `content/skills/flow-research/references/delegation.md:35` (sección "Deliver and resolve instructions"); `_support/docs/architecture/agent-delivery.md`, sección "Native selection and evidence", fila Codex de la tabla por host o una fila de evidencia nueva con versión, fuente y etiqueta.

**Execution:** hilo principal.

**Test approach:** check: observación `rg` de AC1 y `go test ./tests/content/...`.

**Changes:** aplicar DD2 en los dos archivos, redacción en inglés y por capacidad. En `agent-delivery.md` registrar el hecho tal como lo enuncia el contexto verificado de `design.md`: en Codex 0.162.1 un rol por `agent_type` exige `fork_turns: "none"` y no hereda; un hijo genérico sin `agent_type` hereda historia, modelo y esfuerzo por defecto; fuentes `multi_agents_spec.rs` (Context7, 2026-10-13; obs (src)) y `strings` del binario (obs (bin)); no observado en ejecución.

**Verification:** `rg -n "not inherit the parent" content/guidance/global.md content/skills/flow-research/references/delegation.md` sin resultados; `rg -n fork_turns _support/docs/architecture/agent-delivery.md` con la fila nueva; `go test ./tests/content/...` en verde.

### T2 — Un hijo no lanza hijos (H5)

- [x] La guía global y el contrato del hijo en `delegation.md` incluyen "launch children of their own" entre los efectos vedados.

**Closes:** AC2.

**Depends on:** T1 (misma sección de `delegation.md`; evitar ediciones cruzadas).

**Locations:** `content/guidance/global.md:79`; `content/skills/flow-research/references/delegation.md:29`.

**Execution:** hilo principal.

**Test approach:** check: observación `rg` de AC2 y `go test ./tests/content/...`.

**Changes:** aplicar DD1. Añadir a ambas listas "…or launch children of their own; a child that cannot finish alone reports it" (o equivalente), sin tocar los roles.

**Verification:** `rg -n -i "children of (its|their) own" content/guidance/global.md content/skills/flow-research/references/delegation.md` devuelve una línea por archivo; `go test ./tests/content/...` en verde.

### T3 — Comprobar código compartido antes de "mecánico" (H9)

- [x] La definición de cambio mecánico en `global.md` incluye el chequeo de código compartido.

**Closes:** AC4.

**Depends on:** T2 (mismo archivo; secuencial para no cruzar ediciones).

**Locations:** `content/guidance/global.md:72`.

**Execution:** hilo principal.

**Test approach:** check: observación `rg` de AC4 y `go test ./tests/content/...`.

**Changes:** aplicar DD4: tras "…is not mechanical, however small." añadir una frase acotada: before calling mechanical a change that alters what a piece of code does, as opposed to its name or location, check whether that code also serves behavior the request did not name, such as a shared helper, parser, or option; if it does, the change alters a contract. Un rename o un move no se ven afectados.

**Verification:** `rg -n "did not name" content/guidance/global.md` devuelve la línea 72; `go test ./tests/content/...` en verde y el log de `TestGlobalGuidanceSize` anotado para el PR.

### T4 — Comentarios del revisor humano del PR (H2)

- [x] `git-workflow/SKILL.md` tiene la regla acotada de DD3 y `site/.../build.md` la describe en una frase.

**Closes:** AC3, AC6 (parte de `build.md`).

**Depends on:** ninguna.

**Locations:** `content/skills/git-workflow/SKILL.md:30` (nuevo párrafo después); `site/src/content/docs/flows/build.md:41-43`.

**Execution:** hilo principal.

**Test approach:** check: observaciones `rg` de AC3 y AC6 y `go test ./tests/content/...`.

**Changes:** aplicar DD3 y DD7 en cuatro o cinco frases en inglés, con la condición explícita "when a human reviewer is assigned or requested on the pull request", la autoridad (replies and pushes run under the delivery authorization already given for that pull request) y el disparador (on the user's request or when checking the pull request's state). En `build.md` una frase en el párrafo de revisión que remita a la skill.

**Verification:** `rg -n -i "human reviewer" content/skills/git-workflow/SKILL.md site/src/content/docs/flows/build.md` devuelve una línea por archivo; `go test ./tests/content/...` en verde.

### T5 — Decisiones tomadas sin preguntar en el reporte de cierre (H11)

- [x] `reporting.md` exige la lista de decisiones (con umbral) en el reporte de cierre, `unattended-delegation` remite a ella y `site/.../close.md` lo describe en una frase.

**Closes:** AC5, AC6 (parte de `close.md`).

**Depends on:** ninguna.

**Locations:** `content/skills/flow-close/references/reporting.md:51` (condición "a work item completes"); `content/skills/unattended-delegation/SKILL.md:16`; `site/src/content/docs/flows/close.md:34`.

**Execution:** hilo principal.

**Test approach:** check: observaciones `rg` de AC5 y AC6 y `go test ./tests/content/...`.

**Changes:** aplicar DD5: en el reporte de cierre, list each implementation decision taken without asking that chose between behaviors or contracts the user could reasonably have chosen differently, with its reason and how to revert it; omit the list when none was taken. En `unattended-delegation/SKILL.md:16` conservar "decide every question yourself… record the decision and its reason in the task record" y sustituir "list each one in the final report with how to revert it" por una remisión a la regla de `reporting.md`. En `close.md` ampliar el punto 6 con esa frase.

**Verification:** `rg -n "without asking" content/skills/flow-close/references/reporting.md site/src/content/docs/flows/close.md` devuelve una línea por archivo; `sed -n 16p content/skills/unattended-delegation/SKILL.md | rg "how to revert"` sin resultados (las líneas 20 y 29 del archivo usan la frase en otro contexto y no cambian); `go test ./tests/content/...` en verde (incluye `skill_links_test.go` para el enlace nuevo).

## Verificación y revisión humana

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad / condición pendiente |
| --- | --- | --- | --- |
| Pruebas de contenido | Tras cada tarea, worktree de la rama de prueba | `go test ./tests/content/...` en verde | Cubierta por la implementación |
| Suite completa | Candidato final | `go test ./...` en verde; anotar el tamaño de `global.md` que registra `TestGlobalGuidanceSize` | Cubierta por la implementación |
| Revisión del plan | Antes de declarar el plan listo | `hive-review-plan`, dominio guía y skills (un dominio afectado) | Parte de la planificación |
| Verificación independiente del cambio | Al terminar T1–T5 | `hive-verify-change`: relee cada AC contra el diff y corre las observaciones | Cubierta por la implementación |
| Revisión de código dedicada | — | Pendiente de la pregunta de revisión | Pendiente de elección |

Sin efecto de UI renderizada: no aplica `hive-review-ux`.

## Entrega

Pendiente de la respuesta del usuario (ver `proposal.md`).

## Estado de revisión y progreso

Plan escrito el 2026-10-13 sobre la base `f8116af5`. Revisión del plan: una ronda con `hive-review-plan` (rol nativo de Claude Code, dominio guía y skills, candidato = los tres archivos tal como se escribieron). Hallazgos reconciliados por el orquestador, todos aplicados como los propuso el revisor, sin segunda ronda:

- B1 (la guía global repite la afirmación falsa en `global.md:80`): aceptado; T1 y AC1 cubren ahora los dos archivos.
- B2 (el hecho de Codex estaba mal: un rol por `agent_type` no puede heredar historia; solo el hijo genérico hereda por defecto; la etiqueta "doc (src)" no existe en la leyenda): verificado en el binario y corregido en `design.md` y T1 con etiquetas `obs (src)` y `obs (bin)`.
- B3 (el chequeo de código compartido haría no mecánico cualquier rename): aceptado; AC4, DD4 y T3 acotan la regla a cambios que alteran lo que hace el código.
- N1 (H2 sin autoridad ni disparador): aceptado como DD7 en T4.
- N2 (H11 sin umbral y duplicado con el modo desatendido): aceptado; DD5 fija el umbral y hace de `reporting.md` el único hogar; T5 edita también `unattended-delegation/SKILL.md:16`.
- N3 (conservar la razón de pasar rutas explícitas): aceptado en DD2.

Límite declarado por el revisor: no observó en ejecución el comportamiento de Codex ni leyó `multi_agents_spec.rs`; el orquestador sí lo leyó vía Context7. Progreso (2026-10-14): T1–T5 implementadas en el hilo principal sobre `f8116af5` (diff en el worktree, sin commit por `hold`). Observaciones de AC1–AC6 confirmadas por el orquestador; `go test ./tests/content/...` en verde; `global.md` pasa de 26.440 a 26.965 bytes (+525). El único resultado restante de `rg "not inherit the parent" content/` es `plan-review.md:26`, una frase distinta ("not inherit the parent's interpretation") que no forma parte del cambio. Modelos para `hive-verify-task`: implementador Mythos 5.1 (observado, hilo principal); verificador perfil `verifier` = opus (configurado, no observado); difieren, se lanza como está configurado. Verificaciones: T1, T4 y T5 lanzadas juntas (sin archivos compartidos); T2 y T3 tras el veredicto de T1 por compartir `global.md` y `delegation.md`.

Veredictos (2026-10-14, `hive-verify-task`, perfil verifier configurado): T1 AC1 met; T4 AC3 met y AC6 (`build.md`) met; T5 AC5 met y AC6 (`close.md`) met; el orquestador abrió las líneas citadas. Ajustes aplicados tras los veredictos, propuestos por los verificadores y sin cambio de contrato: `build.md:43` dice ahora "assigned or requested" como la regla; `unattended-delegation/SKILL.md:29` deja de listar "every decision taken" y remite a la regla de `reporting.md`, conservando los efectos (completa DD5); la observación de T1 se acota a sus dos archivos. Límite registrado por T5: `tests/content/skill_links_test.go` solo valida los enlaces `skill:` de `global.md`; el enlace relativo nuevo de `unattended-delegation` lo valida `validateInstructionReferences` (`tooling/management/references.go`) al construir el plan de instalación, sin anclas; el ancla se comprobó a mano. T2 AC2 met (roles sin cambios, sin contradicción). T3 AC4 met (rename y move siguen mecánicos; sin conflicto con flow-build:73; TestGlobalGuidanceSize registra 26.965 bytes). Suite completa repetida sin caché tras los ajustes: en verde.

`hive-verify-change` (2026-10-14): AC1–AC6 met contra base y árbol; sin afirmaciones viejas en sitio, docs ni detectores; README.md y llms.txt no describen el comportamiento cambiado; diff limitado a los 8 archivos del plan. Hallazgo aceptado bajo la regla de hallazgos incidentales (pequeño, local, cubierto por las pruebas de contenido): diez roles (`hive-build-*`, `hive-design-*`, `hive-tune-performance`, `hive-write-tests`, `hive-write-spec`) repetían la lista de efectos vedados sin "launch children"; se alineó su frase con la guía global ("write to trackers, or launch children of your own"). Nota aplicada: `unattended-delegation/SKILL.md:29` reformulada para que la lista de decisiones al expirar siga el formato del reporte de cierre. Límite: la afirmación sobre el tipo `fork` de Claude Code en `agent-delivery.md:136` la observó el orquestador en la descripción de su propia herramienta de lanzamiento (obs (local)); el verificador no la vio en la suya. Suite completa `go test ./...` en verde (18 paquetes) sobre el candidato antes de estos dos ajustes de texto; se repite al final.
