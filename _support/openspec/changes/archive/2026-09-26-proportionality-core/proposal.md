# Proporcionalidad base: alcance del código y cambios mecánicos

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `rebuild/harness-engineering` (`8a7e5b9`) |
| Tracker · GitHub Issues | Sin issue; sale de la investigación del 2026-09-26 sobre reglas que empujan a la sobre-ingeniería |
| Git | Commit `8a7e5b9` con push a `rebuild/harness-engineering` (el usuario aprobó la redacción el 2026-09-26) |
| Verificación | Lecturas con `rg` · `go test ./tests/content/... ./integrations/agents/...` · tests de skills · `go vet` y `go test -race` una vez · `/code-review` |
| Siguiente paso | Ninguno en este cambio; el despliegue a los hosts se ofrece aparte |

## Objetivo

La investigación del 2026-09-26 encontró dos causas de sobre-ingeniería en la guía distribuida:

- **Sin límite de alcance para el código:** `global.md` no tiene ninguna regla contra agregar abstracciones, configurabilidad, fallbacks o código defensivo que nadie pidió. Solo lo tocan `flow-build/SKILL.md:28` y `flow-plan/SKILL.md:24`, cuando esos skills se cargan. La [guía de prompting de Anthropic](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices) documenta que sus modelos actuales "overengineer by creating extra files, adding unnecessary abstractions, or building in flexibility that wasn't requested" y recomienda instruirlo explícitamente.
- **Ceremonia sin piso de tamaño:**
  - `global.md:78` dice que las "mechanical edits" no requieren workflow, pero no define qué es mecánico.
  - Varias reglas pesadas no tienen ninguna salida para lo pequeño: la delegación por defecto (`flow-build/SKILL.md:30`), los hijos de UI obligatorios (`:32`, `verification.md:59`), la pregunta de cierre (`global.md:33`), la apertura de issues por cualquier hallazgo incidental (`global.md:41`) y los mandatos de los agentes de desarrollo (`frontend-developer.md:12`, `backend-developer.md:12`).

El Hive legacy tenía una excepción para cambios triviales (`critical-thinking.md:5`) que el rebuild no trajo.

El cambio define "cambio mecánico" en un solo lugar y dice qué ceremonia omite. Agrega una regla de alcance del código con piso de validación en los límites de confianza, y califica en su sitio cada regla pesada, porque una excepción lejana no frena a un modelo débil.

## Alcance y aceptación

**Incluye:**
- `content/guidance/global.md`:
  - `:78`: definición de cambio mecánico y lo que omite.
  - Viñeta nueva con la regla de alcance del código (H1).
  - `:33`: excepción a la pregunta de cierre (D2).
  - `:41`: issues solo para defectos con un escenario de falla (H7).
- `content/skills/flow-build/SKILL.md`: `:30` (delegación), `:32` (hijos de UI) y `:40` (remite a la definición global).
- `content/skills/flow-build/references/verification.md:31` y `:59`, `content/skills/flow-plan/SKILL.md:45`, `content/skills/flow-plan/references/plan-format.md:92` y `content/agents/review/review-ux.md:3`: los cambios de UI solo de texto o estilo los verifica el implementador (D1-A), con la exención en un solo sitio.
- `content/agents/development/frontend-developer.md:12` y `backend-developer.md:12`: los mandatos aplican a lo que el cambio toca (D3-A).
- `tests/content/budget_test.go`: presupuesto al tamaño medido.

**Excluye:**
- **H8** (`plan-review.md:7`, "even when the plan is small"): gobierna el reparto por dominio de una revisión que solo ocurre con plan retenido, no si se revisa.
- **H9** (poda del peso de `global.md`): es la recomendación R3, un cambio aparte.
- **H10** (plugin de Engram): es configuración de terceros y la decide el usuario.
- **Caso de regresión en `tests/pilot`:** el cambio sale de la investigación, no de una falla observada en una sesión. Los pilotos siguen pausados.
- **Despliegue a los hosts:** se ofrece aparte.

**Criterios de aceptación:**
- A1. `global.md` define cambio mecánico con una lista cerrada. Excluye "however small" todo lo que altera lógica, un contrato, permisos, seguridad, dinero, datos almacenados, concurrencia o una dependencia, y la configuración que toca permisos, seguridad o dependencias. Una pregunta simple no necesita workflow ni reporte. Un cambio mecánico no necesita workflow formal, plan retenido ni delegación de la edición, ni revisión dedicada salvo que la guía del proyecto la exija. `:80` usa el mismo término.
- A2. `global.md` tiene una regla de alcance del código:
  - no agregar features, abstracciones, configurabilidad, fallbacks ni capas de compatibilidad que nadie necesite;
  - no validar casos inalcanzables, pero sí validar en los límites de confianza;
  - no refactorizar ni comentar código que el cambio no toca, salvo correcciones bajo la regla de hallazgos incidentales.
- A3. `global.md:33` omite la pregunta de cierre tras un cambio mecánico sin ningún ítem Pendiente. El reporte conserva su etiqueta de limpieza, el estado de entrega y los recordatorios, para que el usuario no tenga que preguntar.
- A4. `global.md:41` abre issue solo para defectos con un escenario de falla concreto; los detalles de estilo o redacción van solo al reporte.
- A5. `flow-build/SKILL.md:30` no delega por defecto un cambio mecánico, y `:40` remite a la definición global en vez de repetir la lista.
- A6. `verification.md:59` exime de `review-ux` y `sdd-verify` a los cambios de UI que solo cambian texto o un valor de estilo sin alterar layout, estado ni flujo. El implementador los revisa en la página renderizada, en los viewports afectados, con el texto realista más largo y en cada tema; si alteran el layout, los hijos aplican. `flow-build:32`, `verification.md:31`, `flow-plan:45`, `plan-format.md:92` y `review-ux.md:3` remiten a esa regla sin redefinirla.
- A7. `frontend-developer.md` y `backend-developer.md` limitan sus mandatos de estados y de chequeos de auth y errores a lo que el cambio introduce o afecta.
- A8. Pasan `go test -count=1 ./tests/content/... ./integrations/agents/...`, los tests de skills, `go vet ./...` y `go test -race -count=1 ./...`.

## Decisiones del usuario (2026-09-26)

- **Ruta:** R1 con `flow-plan`: regla de alcance del código más excepción de cambios mecánicos para la ceremonia.
- **D1-A:** un cambio de UI solo de texto o de un valor de estilo que no altera layout, estado ni flujo lo verifica el implementador en la página renderizada; los hijos quedan para layout, flujos, estados y pantallas nuevas.
- **D2 (respuesta propia del usuario):** tras un cambio mecánico sin nada pendiente, se cierra con el reporte, sin pregunta. El reporte debe indicar la limpieza o higiene hecha, para que el usuario no tenga que preguntar.
- **D3-A:** incluir los calificadores de los agentes de desarrollo.

## Entrega

Respuesta del usuario (2026-09-26):
- **Modo:** `direct-base`. Implementar, verificar y parar para que el usuario lea el diff. Después, commit de solo estas rutas y esta carpeta, con push a `rebuild/harness-engineering`. El archivo del cambio va en un segundo commit con push.
- **Revisión:** `/code-review` de Claude Code antes del commit. Con la regla nueva de revisión, aplica porque el cambio altera contratos de la guía distribuida. Sus hallazgos se tratan por prioridad, con una sola re-revisión.
- **Despliegue:** se ofrece aparte.

El árbol tiene cambios sin commitear de la sesión del instalador. La entrega agrega al commit solo las rutas de este cambio.
