# Tareas

### T1 — Reglas de glosa, enlaces y forma en `global.md`

- [x] `global.md` tiene los cambios de [design.md](design.md#guía-globalmd): `:16` con la regla de enlaces, `:17` sin la frase de glosa ni "short labeled paragraphs or a compact list", y los bullets de glosa y de forma después de `:17`.

**Depende de:** nada.
**Ubicación:** `content/guidance/global.md:16-21`, `tests/content/budget_test.go:12`.
**Ejecución:** hilo principal; la redacción está acoplada a las decisiones D1-A, D2-A y D3-A.
**Verificación:** `wc -c content/guidance/global.md` coincide con `globalGuidanceBudget`; `go test ./tests/content/...` pasa; releer los bullets contra A1 y A2 y contra `:19`, `:20`, `:21` y `:25` para confirmar que no hay contradicciones.

### T2 — `flow-report` sin exigir prosa

- [x] `SKILL.md:8` usa el texto de [design.md](design.md#flow-report).

**Depende de:** T1.
**Ubicación:** `content/skills/flow-report/SKILL.md:8`.
**Ejecución:** hilo principal; es una frase.
**Verificación:** `rg -n "in prose" content/skills/flow-report/SKILL.md` no devuelve nada; `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa.

### T3 — Mapa de capacidades de presentación

- [x] `agent-delivery.md` tiene la sección "Presentation capabilities" con la tabla por host, fuentes y etiquetas de evidencia.

**Depende de:** nada.
**Ubicación:** `_support/docs/architecture/agent-delivery.md`, después de "Native user questions".
**Ejecución:** hilo principal; los datos ya están verificados en esta sesión.
**Verificación:** cada fila tiene las cinco capacidades, la versión, la fuente y la etiqueta de evidencia; el dato de enlaces de Claude Code cita la observación del 2026-09-26.

### T4 — Criterios `cited_id_glossed` y `no_bare_url` con fixtures

- [x] Los dos criterios existen en `tests/pilot/regression.go` según [design.md](design.md#criterios-deterministas-testspilotregressiongo), registrados en `regressionCriteria`, con doc comment de límites.
- [x] `tests/fixtures/regression/cited_id_glossed/`: `claude-fail.jsonl`, `claude-pass.jsonl`, `claude-question-after-text-fail.jsonl`, `claude-range-fail.jsonl`, `claude-range-pass.jsonl` y `provenance.md`, con la entrada anidada real de `AskUserQuestion`.
- [x] `tests/fixtures/regression/no_bare_url/`: `claude-fail.jsonl`, `claude-pass.jsonl` (una URL en cada forma excluida) y `provenance.md`.
- [x] Tests unitarios de tokens y glosa del diseño.
- [x] `TestRegressionFixtures`, `TestRegressionCriteriaReturnsAllThree` (renombrado según la nueva cantidad), los doc comments de `regression.go:10` y `flows.go:839` y `tests/fixtures/regression/README.md` actualizados.

**Depende de:** nada; comparte contrato solo con el diseño.
**Ubicación:** `tests/pilot/regression.go`, `tests/pilot/regression_test.go`, `tests/pilot/flows.go` (solo el doc comment), `tests/fixtures/regression/**`.
**Ejecución:** delegado a `test-engineer` (interfaz fija, sin escrituras solapadas con T1–T3). Rutas que puede modificar: las de arriba. Commits locales no permitidos. Lee `tests/fixtures/regression/README.md` antes de derivar fixtures y nunca lee `tool_result` ni líneas completas de sesiones.
**Verificación:**
- `go test -race ./tests/pilot/...` pasa.
- Prueba de reversión por rama: se revierte por separado la rama de superficie de texto, la de pregunta, la de "texto justo antes de la pregunta" y la de rangos; cada una hace fallar su fixture. Se registra el resultado en `provenance.md`.
- El escaneo `rg -l -i -P '/Users/|/home/|/Volumes/|/private/|~/|password=|secret=|(?<!EXAMPLE_API_)token=' -g '!README.md' tests/fixtures/regression` no imprime nada.

### T5 — Verificación conjunta y revisión

- [x] Todo el repositorio pasa y la revisión de código está resuelta.

**Depende de:** T1–T4.
**Ejecución:** hilo principal; la revisión, con el mecanismo que se elija en la entrega.
**Verificación:** `go vet ./...` y `go test -race ./...` pasan; `git status` muestra intactos los cambios ajenos del árbol; los hallazgos de revisión están corregidos o refutados con evidencia.

## Verificación y revisión humana

| Gate | Momento | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Tests y vet | T5 | `go vet ./...`, `go test -race ./...` en verde | Incluido en la implementación |
| Regresión | T4 | Fixtures fail/pass y reversión por rama | Incluido en la implementación |
| Redacción | Antes de la entrega | El usuario lee en el diff los cambios de `global.md:16-17` y los dos bullets nuevos | Parada humana |
| Revisión de código | T5 | Mecanismo elegido en la pregunta de entrega | Pendiente |
| Despliegue a hosts | Después de integrar | Manager de Hive, release nuevo | Fuera de este plan; se ofrece aparte |

## Estado de revisión y progreso

**Revisión del plan (2026-09-26):**
- **Dominios cubiertos:** guía distribuida (T1–T3) y arnés de pruebas (T4), un `review-plan` por dominio en paralelo, sobre la primera versión guardada. Los dos revisores obtuvieron hashes distintos por el orden de concatenación de los archivos; el contenido no cambió entre el guardado y la revisión.
- **Guía, aceptados:** B1 (la pregunta nativa es su propio mensaje), B2 (la frase de `flow-report` mandaba los handoffs al chat), B3 (la lista de tareas nativa ya vive en `flow-build:46`), N1 (la apertura sigue en `:16`), N2 (tablas limitadas al chat; manda el formato del skill), N3 (una sola regla de etiquetas), N4 (verificación de T2 con Python) y N5 (oferta de página con disparador).
- **Guía, aceptado en parte:** B4. Los enlaces se unifican en `:16`. Se descarta mostrar la URL junto a la etiqueta: el usuario pidió explícitamente no pintar URLs, cinco hosts muestran la etiqueta, y OpenCode, según su binario, agrega la URL por su cuenta cuando no hay hipervínculos. El riesgo queda anotado en el diseño.
- **Pruebas, aceptados:** B1 (definición por inicio de línea, normalización de énfasis y código), B2 (entrada anidada de `AskUserQuestion`; `header` excluido), N1 a N6 (límite de `Message`, rangos en código, estados y evidencia de `no_bare_url`, registro y `--assess`, derivación por campos y confirmación de S1–S6, reversión por rama).
- **Sin segunda ronda:** las correcciones aplican lo que propusieron los propios revisores. La excepción es B4 de la guía, descartado con justificación y registrado como riesgo.

**Implementación (2026-09-26):**
- **T1:** hecho. `global.md` pasa de 38015 a 39056 bytes (+1041); `globalGuidanceBudget` sube a 39056. `go test -count=1 ./tests/content/...` pasa.
- **T2:** hecho. `rg "in prose"` sin resultados; `python3 -m unittest discover -s tests/skills -p '*_test.py'`: 23 tests OK.
- **T3:** hecho. Sección "Presentation capabilities" con seis hosts y cinco capacidades.
- **T4:** hecho por `test-engineer`. Siete fixtures derivados de `876d776d`, `9d7ef3ee` y `0f38c529` (dos construidos desde el contrato, declarados en `provenance.md`); reversión por rama registrada. Defecto corregido al derivar: un ID solo se define cuando abre la etiqueta de opción.
- **T5:** `go vet ./...` y `go test -race -count=1 ./...` en verde; escaneo de fixtures vacío. `/code-review` (medium) dio H1–H5; todos corregidos con un test en rojo antes de cada arreglo: clave de mensaje por evento cuando falta `Message` (Codex), preguntas planas de Pi y Grok, definiciones en orden de la traza, rangos expandidos (tope 30) y texto de pregunta que abre con un ID, y la fila de Codex en `agent-delivery.md`. Los fixtures existentes no cambiaron de estado.
- **Entrega:** redacción aprobada por el usuario; commit `66ab311` con push a `rebuild/harness-engineering`; despliegue a los seis hosts desde un worktree limpio en `66ab311`, release `b301a4d80d51`, 425 recursos instalados; texto nuevo verificado en `~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md`, `~/.cursor/AGENTS.md` y `flow-report`.
