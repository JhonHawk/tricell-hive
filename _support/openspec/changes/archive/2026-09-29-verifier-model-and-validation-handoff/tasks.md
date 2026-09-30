# Tareas

Base del build: `f348588` (`rebuild/harness-engineering`), rama `feat/verifier-model-and-validation-handoff`, 2026-09-29.

Orden: T1 primero (código Go del gestor); T2 y T3 después, en un solo encargo de guía; T4 cierra.

## T1 — Perfil `verifier`

- [x] Los perfiles aceptan `verifier` sin romper releases congeladas, y `hive-verify-task` resuelve al modelo de la tabla por host.

**Closes:** AC1, AC2.

**Depends on:** ninguna.

**Locations:** `integrations/agents/agents.go` (`Parse` línea 101; `ReadProfiles` líneas 145–160; `Resolve` líneas 255–279); `integrations/agent-profiles.json` (agregar `verifier` después de `inherit` en cada host); `content/agents/quality/hive-verify-task.md` frontmatter; `integrations/agents/agents_test.go`, `resolve_test.go`, `render_golden_test.go` y `testdata/render.golden`.

**Execution:** delegada a `hive-build-backend` (perfil `execution`, Sonnet): código Go con pruebas y contrato cerrado.

**Test approach:** tdd. Pruebas que fallan primero:
- `ReadProfiles` acepta un host con cuatro perfiles y valida `verifier` como los demás.
- `ReadProfiles` sigue aceptando perfiles de tres entradas (snapshot anterior), modelada sobre la prueba de cinco hosts (`agents_test.go:347-378`).
- `Parse` acepta `model_profile: "verifier"`.
- `Resolve` de un rol `verifier` contra perfiles sin `verifier` falla con un error que nombra el host y el perfil.
- Los perfiles del repositorio traen `verifier` en los seis hosts.
- Tabla de `Resolve` sobre el archivo real `content/agents/quality/hive-verify-task.md` y el `agent-profiles.json` del repositorio para los seis hosts, con el modelo y el esfuerzo de la tabla del diseño (OpenCode `github-copilot/claude-opus-5.5#high`, Grok `grok-4.6` sin esfuerzo, Cursor `inherit`).

**Changes:** según [diseño, perfil `verifier`](design.md#perfil-verifier). No cambiar `execution`, `reasoning` ni `inherit`. Regenerar el golden con `-update-golden` y revisar que su diff se limite a los seis bloques de `hive-verify-task`.

**Verification:**

```bash
go test ./integrations/... ./tooling/management/... ./tooling/cli/...
git diff --stat -- integrations/agents/testdata/render.golden
rg -n 'model_profile: "verifier"' content/agents/quality/hive-verify-task.md
```

Resultado esperado: las pruebas nuevas se vieron fallar antes del cambio y pasan después; el diff del golden solo toca `hive-verify-task`; el `rg` encuentra el frontmatter.

## T2 — Comparación de modelos antes de lanzar el verificador

- [x] `flow-build` compara los modelos antes de lanzar `hive-verify-task`, conserva el caso de verificador que no puede correr y anota los modelos.

**Closes:** AC3.

**Depends on:** T1.

**Locations:** `content/skills/flow-build/SKILL.md` ciclo de `hive-verify-task` (pasos 1, 5 y 6); `_support/docs/architecture/agent-delivery.md` líneas 11–19, 33 y tabla de perfiles.

**Execution:** delegada a `hive-write-spec` junto con T3 (un solo escritor de guía).

**Test approach:** check.

**Changes:** según [diseño, comparación](design.md#comparación-antes-de-lanzar).

**Verification:**

```bash
rg -n -i "configured but not observed|configured, not observed" content/skills/flow-build/SKILL.md
rg -n -i "treat .*unknown|inherits the session" content/skills/flow-build/SKILL.md
rg -n -i "out of quota|cannot run" content/skills/flow-build/SKILL.md
rg -n -i "rest of the build" content/skills/flow-build/SKILL.md
rg -n "verifier" _support/docs/architecture/agent-delivery.md
```

Resultado esperado: cada `rg` de `SKILL.md` devuelve la línea de la regla nueva (ninguna frase existe en la base salvo "cannot run", que debe seguir presente); el último muestra la tabla con `verifier`, la nota que lo distingue del perfil de acceso `verify` y la evidencia de catálogo de Pi y OpenCode.

## T3 — Validación manual en texto

- [x] La petición de validación manual termina el turno con la pregunta en texto, sin la herramienta de preguntas.

**Closes:** AC4.

**Depends on:** T1 (solo por orden de commits).

**Locations:** `content/guidance/global.md` regla de esperas del usuario (línea 22); `content/skills/flow-build/references/verification.md` "Human UI handoff" (líneas 63–71); `tests/content/budget_test.go` solo si el tope no alcanza.

**Execution:** delegada a `hive-write-spec` junto con T2.

**Test approach:** check.

**Changes:** según [diseño, validación en texto](design.md#validación-manual-en-texto).

**Verification:**

```bash
rg -n -i "manual validation" content/guidance/global.md content/skills/flow-build/references/verification.md
rg -n -i "question in text|text question" content/guidance/global.md content/skills/flow-build/references/verification.md
go test ./tests/content/
```

Resultado esperado: los dos `rg` encuentran la regla en `global.md` y en "Human UI handoff" (ninguna de esas frases existe en la base), con el alcance del diseño y la nota de que las preguntas de entrega, ruta y cierre no cambian; `tests/content` pasa, con el tope sin cambios o subido en este cambio con su razón en el commit.

## T4 — Validación final

- [x] El candidato final pasa las comprobaciones del repo.

T4 sobre `fb8f61b`: `go vet ./...` sin hallazgos; `go test -count=1 ./...` con código de salida 0 y ninguna línea `FAIL` en la salida completa (16 paquetes `ok`); `update --dry-run --source .` con código 0 y los seis `hive-verify-task`. Commits: `a9cdaab` (T1) y `fb8f61b` (T2 y T3).

Hallazgo corregido en este cambio: `TestRenderGolden` fallaba en la base `f348588`, comprobado en un worktree temporal. El PR #67 editó el cuerpo de `hive-verify-task.md` sin regenerar el golden, y su T5 leyó `go test` a través de `tail -15`, que ocultó el fallo y devolvió el código de salida de `tail`. T1 regeneró el golden.

**Depends on:** T1–T3.

**Execution:** hilo principal.

**Verification:**

```bash
go vet ./...
go test ./...
go run ./tooling/cli update --dry-run --source .
```

Resultado esperado: sin fallos; el dry-run sobre el commit final valida el catálogo.

## Verificación y revisión humana

Sin superficie de UI. `hive-verify-task` verifica T1–T3 con los comandos de cada tarea, sin correr la suite completa (la corre T4 una vez). Los implementadores corren en Sonnet (`hive-build-backend`, `hive-write-spec`) y el verificador en Opus; se anotan ambos modelos por tarea, como pide la regla de T2.

| Gate | Momento | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | Tras cada tarea | `hive-verify-task` con los comandos de la tarea | Parte de implementar |
| Suite y catálogo | Candidato final | T4 | Parte de implementar |
| Revisión de código | Antes del merge | Ninguna dedicada (D5-C reutilizada) | Decisión del usuario en esta sesión |

## Revisión del plan y avance

Revisión del plan, ronda 1 (candidato proposal `bb035d15`, design `8b5bad19`, tasks `9a8356c7`), dos `hive-review-plan` en paralelo:

- Gestor Go (backend/integración): B1, exigir `verify` rompía releases congeladas que el gestor relee (`models.go:24-26`, `plan.go:300`, `apply.go:78`): se aplicó su opción (a), `verifier` opcional al leer y obligatorio en los perfiles del repo. N1, Pi y OpenCode sin evidencia de catálogo: comprobados (`pi --list-models` lista `xai grok-4.7`; `opencode models` lista `github-copilot/claude-opus-5.5`); T3 original (modelo de Pi) se resolvió en la planificación y se quitó con su AC. N3, golden y rol real: agregados a T1. N4: condicionado a un cambio de Pi que no ocurre.
- Guía distribuida: B1, se perdía el caso de verificador que no puede correr: vuelve dentro de la regla. B2, modelo desconocido: se trata como igual, cada modelo se anota como observado o configurado, y una respuesta vale para el build. N1–N3, alcance de la validación en texto: acotado a recorrido, rondas y parada de `interactive`, con las demás preguntas del turno en texto y las de entrega, ruta y cierre sin cambio. N4, `rg` que pasaban con el texto de la base: reemplazados por frases nuevas. N5, nombre `verify` repetido: el perfil se llama `verifier`.

Todas las correcciones aplican la propuesta de cada revisor tal como la hizo; no requieren otra ronda.

Build (2026-09-29): T2 y T3 los implementó `hive-write-spec` (Sonnet, configurado, no observado) en paralelo con T1, sin archivos en común. Verificadores `hive-verify-task`: Opus 5.5, observado (`claude-opus-5-5[1m]` según el host); AC3 y AC4 `met` en 41–47 s cada uno. `global.md` quedó en 43 312 bytes y el tope de `tests/content` subió de 43 000 a 43 320: la cláusula de validación agrega unos 430 bytes y comprimir la misma regla recuperó unos 110. Después de los veredictos, el orquestador restauró "or wait" en la frase de la herramienta de preguntas sin respuesta (el implementador la había recortado sin que el diseño lo pidiera) y agregó dónde se anotan los modelos ("in the plan's progress section"), observación O1 del verificador de T2; ninguno toca lo que comprueban los `rg` de AC3 y AC4.

Entrega: PR #68 mergeado en `rebuild/harness-engineering` como `e9c4c01` (2026-09-29), sin CI en el repo y sin revisión dedicada (D5-C). Cierre: esta carpeta archivada por push directo a la base (D6-A); luego binario recompilado y `hive update`.
