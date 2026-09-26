# Tareas

### T1: Niveles de Pi y soporte de `max`

- [x] Tabla de equivalencias de Pi cerrada y límites de `max` anotados en `design.md`.

**Depende de:** nada.

**Ubicaciones:**
- `~/.pi/agent/npm/node_modules/pi-subagents/src/shared/model-info.ts`: `getSupportedThinkingLevels` en las líneas 88-99 y `resolveEffectiveThinking` en las 56-61.
- La documentación de subagentes de Claude Code y la de Codex.
- El modelo de sesión de Codex en `~/.codex/config.toml`.

**Ejecución:** delegada a `sdd-explore`, porque implica leer código instalado. Devuelve solo los niveles aceptados y la fuente de cada uno.

**Cambios:** responder tres preguntas, con la fuente de cada respuesta:
- En Pi, con `model: "inherit"`, ¿`thinking: "max"` se aplica, se baja a otro nivel o se descarta? ¿Qué niveles quedan con un modelo conocido?
- En Codex, cuando `review-security` hereda el modelo de la sesión, ¿el nivel `max` está soportado? ¿La documentación dice qué hace Codex con un nivel no soportado?
- En Claude Code, con un padre Sonnet, ¿el nivel `max` está soportado? ¿La documentación dice qué hace con un nivel no soportado?

Anotar el resultado en `design.md`. Si en Pi `max` se descarta con `inherit`, D5-A exige traducirlo al nivel más alto que se conserve.

**Verificación:** cada respuesta cita su archivo o URL. Si una fuente no está disponible, el límite queda escrito en `design.md`.

### T2: Campo `effort`, perfiles y roles

- [x] El parser acepta `effort` y rechaza `claude_effort`. El renderizador aplica el esfuerzo del rol en Claude, Codex y Pi con la traducción de `design.md`, que según T1 es identidad. El perfil de ejecución de Claude tiene `effort: high`, y las cuatro excepciones están migradas.

**Depende de:** T1, solo por la equivalencia de Pi, y T5.

**Ubicaciones:**
- `integrations/agents/agents.go` (`Role`, `Parse`, `Render`) y `integrations/agents/agents_test.go`.
- `integrations/agent-profiles.json`.
- Declaran `effort`: `content/agents/review/review-security.md`, `content/agents/quality/state-fetcher.md`, `content/agents/review/review-plan.md` y `content/agents/docs/sdd-spec-writer.md`.
- Pierden `claude_effort`: `content/agents/quality/sdd-verify.md`, `content/agents/review/review-ux.md` y `content/agents/review/sdd-explore.md`.

**Ejecución:** delegada a `backend-developer`, que puede modificar solo esos archivos y no hace commits; recibe `design.md` como contrato. Va en una sola tarea porque `TestCatalogueRendersAllRolesForEveryHost` (`agents_test.go:21-43`) renderiza las fuentes reales: rechazar `claude_effort` sin migrar las fuentes rompe la prueba.

**Cambios:** según "Dónde vive el esfuerzo", "Contrato del campo" y "Compatibilidad y versiones" de `design.md`; `agents.Version` no cambia. Pruebas nuevas:
- `effort` inválido falla.
- `claude_effort` falla por campo desconocido.
- Un `effort` de rol reemplaza al del perfil en Claude, Codex y Pi.
- Grok, Cursor y OpenCode no emiten esfuerzo aunque el rol lo declare.
- Una prueba sobre las fuentes del repositorio renderiza los 19 roles en los seis hosts y compara con la tabla "Niveles esperados por rol" (cubre A1 y A2).

**Verificación:**
- `go test ./...` pasa, y las pruebas nuevas fallan antes del cambio y pasan después.
- `rg -n 'claude_effort' content integrations tooling` no devuelve resultados.

### T3: Documentación

- [x] `agent-delivery.md` y `deployment-manager.md` describen el contrato nuevo.

**Depende de:** T2.

**Ubicaciones:** `_support/docs/architecture/agent-delivery.md` (línea 5 del contrato, tabla de perfiles, tabla de traducción y límites de `max`); `_support/docs/architecture/deployment-manager.md` (error `unsupported agent field` con releases anteriores y cómo volver a uno).

**Ejecución:** hilo principal, porque depende del texto final de T2.

**Cambios:**
- Reemplazar `claude_effort` por `effort`, y agregar la tabla de traducción y la regla del nivel efectivo.
- Agregar una frase propia para los hosts donde las excepciones no tienen efecto: en Grok y Cursor todos los roles heredan el esfuerzo de la sesión, y en OpenCode todos corren en `#max`.
- Registrar que la traducción de D5-A usa la escala del host, no la del modelo: en un rol `inherit`, `max` depende del modelo de la sesión.
- Registrar el límite para volver a releases anteriores.
- Hallazgo incidental del conteo D1-A: la base de OpenCode registra llamadas `task` con `subagent_type`, mientras la tabla de dialectos dice `subagent` con `agent` en V2. Queda abierto en [#32](https://github.com/JhonHawk/tricell-hive/issues/32), sin tocar la fila.

**Verificación:** `rg -n 'claude_effort' _support/docs` no devuelve resultados. Una lectura final confirma que la tabla de perfiles coincide con `agent-profiles.json`.

### T4: Actualización en home sintético

- [x] Una instalación hecha desde la base actual se actualiza con el working tree; se comprueban la recuperación y el rechazo del release anterior (A5).

**Depende de:** T2 y T5.

**Ubicaciones:** una copia de trabajo de `HEAD` en `_support/workspace/2026-09-25-subagent-effort-profiles/base/`, y el home y el directorio de estado sintéticos en la misma carpeta.

**Ejecución:** delegada a `sdd-verify`, para que un agente distinto al que implementó compruebe el comportamiento. Escribe solo dentro de esa carpeta de `_support/workspace/`.

**Cambios:** ninguno en las fuentes. Rutas de ejemplo: `<ws>` es `_support/workspace/2026-09-25-subagent-effort-profiles`, con `<ws>/home` y `<ws>/state`. Cada plan va a un archivo nuevo, porque `--out` no puede apuntar a un archivo existente.
1. Desde `<ws>/base`, con el binario de la base: `go run ./tooling/cli plan install --hosts codex,claude,grok,pi,opencode,cursor --scope user --home <ws>/home --state-dir <ws>/state --out <ws>/p1.json`, y después `go run ./tooling/cli apply --plan <ws>/p1.json`.
2. Desde el working tree: el mismo `plan install` con `--out <ws>/p2.json`, y `apply --plan <ws>/p2.json`.
3. `go run ./tooling/cli status --hosts codex,claude,grok,pi,opencode,cursor --scope user --home <ws>/home --state-dir <ws>/state`.
4. Desde el working tree: `plan install --release <hash de p1> --hosts claude --scope user --home <ws>/home --state-dir <ws>/state --out <ws>/p3.json`.

**Verificación:**
- El paso 2 no reporta conflictos y quita `cloud-architect` de los seis hosts (T5).
- `status` no reporta diferencias.
- En `<ws>/home`, `.claude/agents/backend-developer.md` tiene `effort: "high"` y `review-security.md` tiene `effort: "max"`; `.codex/agents/state-fetcher.toml` tiene `model_reasoning_effort = "low"`.
- El paso 4 falla con `unsupported agent field "claude_effort"`.

### T5: Fusionar `cloud-architect` en `solution-architect` (D8-B)

- [x] `solution-architect` cubre también el diseño de infraestructura cloud, `cloud-architect` deja de existir y el catálogo queda en 19 roles.

**Depende de:** nada. Se hace antes de T2, mientras corre T1, para que la prueba de T2 ya use el catálogo de 19 roles.

**Ubicaciones:** `content/agents/design/solution-architect.md` y `content/agents/design/cloud-architect.md` (se elimina). En `_support/docs/architecture/agent-delivery.md`, la frase "Twenty roles" y el párrafo de consolidaciones.

**Ejecución:** hilo principal, por ser una redacción corta sobre el contrato de un rol.

**Cambios:**
- Ampliar la descripción de `solution-architect` para que también se elija ante decisiones de infraestructura cloud (compute, red, almacenamiento y servicios gestionados, con su costo y sus modos de falla).
- Agregar al cuerpo lo que solo aportaba `cloud-architect`: los objetivos de disponibilidad y recuperación medibles, separar costos estimados de costos medidos, y consultar la referencia de nombres de infraestructura de `flow-plan`.
- Eliminar `cloud-architect.md` y anotar la consolidación en `agent-delivery.md`, como las consolidaciones anteriores.

**Verificación:**
- `go test ./...` pasa con 19 roles.
- `rg -n 'cloud-architect' content integrations tooling _support/docs` solo devuelve la nota de consolidación.
- En T4, después del paso 2, `<ws>/home/.claude/agents/cloud-architect.md` y sus equivalentes en los otros cinco hosts ya no existen.

## Verificación y revisión humana

| Comprobación | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Pruebas | Después de T2; checkout local | `go test ./...`, `go test -race ./...`, `go vet ./...` pasan | Parte de la implementación |
| Home sintético | T4; `_support/workspace/` | Pasos y resultados de T4 | Parte de la implementación, sin efectos globales |
| Revisión del código | Al terminar T3, sobre el diff del working tree | `/code-review` de Claude Code (D7-A); los hallazgos se corrigen o se refutan con evidencia | Elegida el 2026-09-25; se lanza en `flow-build` |
| Tu revisión | Al final (`hold`) | Diff del working tree en VS Code; comparar con la tabla de niveles de `design.md` | Tuya: decides commit o despliegue |

No hay pantallas, así que no aplica verificación de interfaz.

## Estado de la revisión y avance

- Revisión del plan del 2026-09-25 sobre la versión e310048660ff / d67fbbdf33cb / ee72d28fefaf, con `review-plan` en paralelo para dos dominios. En Claude Code se seleccionó el rol nativo.
  - Renderizador y gestor en Go: dos bloqueantes, T2/T3 indivisibles y un error esperado equivocado en A5. Se corrigieron uniendo T2 con T3 y esperando `unsupported agent field`. También se corrigieron los comandos de T4 y `Version` se queda en `"1"`.
  - Contenido y portabilidad: un bloqueante, T1 verificaba `max` contra el modelo equivocado. Se corrigió cambiando las preguntas de T1 y agregando el límite a `design.md`. También se precisó por qué Cursor queda fuera y se agregó la frase de límites a T3.
  - Todas las correcciones son las que propusieron los propios revisores, así que no hubo segunda ronda.
  - Queda abierto: el soporte de `max` según el modelo, que depende de T1.
- D1-A resuelto con el conteo en seis hosts; D8-B (2026-09-25): se fusiona solo `cloud-architect` y se mantiene `kotlin-multiplatform-developer`.
- Implementación autorizada con `flow-build` (2026-09-25), en `hold`. Orden: T1 (delegada) en paralelo con T5 (hilo principal), después T2, T3 y T4.
- T1, T2, T3 y T5 terminadas (2026-09-25).
  - T2: de las cinco pruebas nuevas, cuatro fallaron antes del cambio por la razón esperada. `TestParseRejectsInvalidEffortValue` ya pasaba antes, porque el campo desconocido también la rechazaba.
  - Después del cambio pasan `go test ./...`, `-race` en `integrations` y `tooling`, y `go vet ./...` (según el hijo). El hilo principal repitió `go test -count=1 ./integrations/... ./tests/...` y `go vet ./integrations/...`, y pasaron.
  - Desvío aceptado: `rg 'claude_effort'` encuentra el literal en la prueba que exige rechazarlo (`agents_test.go:205-208`) y en la nota de `deployment-manager.md`; ningún rol ni el código de producción lo usan.
  - Hallazgo: otra sesión tiene cambios sin commit en `tooling/` e `install.sh` de este mismo checkout, y seguía editando a las 16:19. T4 y la revisión se hacen en worktrees aislados con solo los archivos de este cambio.
- T4 terminada (2026-09-25), verificada por `sdd-verify` en worktrees aislados; el hilo principal repitió una muestra.
  - La base instaló `cloud-architect` en los seis hosts, y la actualización lo quitó sin conflictos.
  - `status` mostró las 425 entradas en `installed`.
  - Los niveles renderizados coinciden con la tabla, y Grok, Cursor y OpenCode no tienen línea de esfuerzo.
  - `plan install --release 6465f0e9…` falló con `unsupported agent field "claude_effort"`.
  - En el candidato aislado (HEAD más este cambio) pasan `go test ./...`, `-race` y `go vet`.
- `/code-review` alto (D7-A) sobre el diff aislado del candidato, 2026-09-25. Nueve hallazgos:
  - Corregidos (6):
    - La lista de hosts que aceptan esfuerzo estaba duplicada; ahora usa `acceptsEffort`.
    - La fusión había perdido "assumptions" en `solution-architect`.
    - El comentario de la prueba de cinco hosts había quedado desplazado.
    - La búsqueda de `effort` en la salida era frágil; ahora usa `emittedEffort`, anclada a la clave.
    - El comentario de `Parse` estaba desactualizado.
    - Un comentario de prueba tenía un encabezado en español y apuntaba a `design.md`.
  - Refutados o aceptados sin cambio (3):
    - `max` en Codex: es el límite documentado en la sección "Riesgos acotados" de `design.md` y en `agent-delivery.md`. D5-A traduce la escala del host, que incluye `max`.
    - `ultra` solo existe en perfiles de Codex, y el contrato portable de roles se limita a `low`…`max`.
    - El conteo de roles en dos pruebas es anterior a este cambio y queda fuera de alcance.
  - Después de las correcciones, en el candidato pasan `go test ./...`, `-race` y `go vet`.
- Cerrado en `hold` (2026-09-25): implementación verificada en el working tree, sin commits. La carpeta de cambio queda sin versionar.
- Entrega autorizada por el usuario (2026-09-26): "commit + push y deploy cuando acabe". Commit selectivo por rutas, push a `rebuild/harness-engineering` y despliegue a los seis hosts cuando termine `work-close-sequence`. Al integrar, la carpeta de cambio se cierra y se archiva.
