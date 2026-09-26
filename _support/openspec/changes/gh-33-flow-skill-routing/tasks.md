# Tareas

### T1: Regla de ruta

- [x] La viñeta nueva está en `global.md`, junto a la de `flow-research` (`:77`), y el presupuesto de `tests/content/budget_test.go` está ajustado.

**Ejecución:** hilo principal, porque es una línea de guía.

**Verificación:** `wc -c content/guidance/global.md` coincide con `globalGuidanceBudget`, y `go test ./tests/content/...` pasa.

### T2: Criterio `flow_skill_read_before_delivery`

- [x] La función y sus pruebas están en `tests/pilot/regression.go` y `regression_test.go`, con los casos `tests/fixtures/regression/flow_skill_read_before_delivery/{grok-fail,grok-pass,codex-pass}.jsonl` y su `provenance.md`, y la fila correspondiente en el README.

**Ejecución:** delegada a `test-engineer`, que solo toca esas rutas.

**Detalle** (con las correcciones de la revisión):
- Un helper compartido con `observeSkill` (`flows.go:364-432`) devuelve el índice del evento donde se leyó `flow-build/SKILL.md` con contenido: `read`, `cat` en el shell (Codex, `trace.go:500-506`) o la carga nativa con contenido. Un `skill_invocation` sin contenido no cuenta (`trace.go:444-445`).
- Cuenta como acción de Git cualquier intento de `shell`, exitoso o no (el fallo no es confiable en Grok, `trace.go:293-294`), reconocido con `shellSegments` y `gitInvocation`: `git commit`, `git push`, `gh pr create`, `gh pr merge`.
- El criterio falla con la primera acción de Git que no tenga antes esa lectura. La evidencia nunca incluye el texto del comando. No se conecta a `regressionCriteria` ni a `assessFlows`.
- El comentario de documentación de `regression.go:10-16` y el README declaran los límites: comandos escritos (`/flow-build`, `$flow-build`) y despliegues sin Git (G5).
- Casos, reescritos al formato que lee `parseTrace` y con `provenance.md` que mapea las líneas originales:
  - `grok-fail`: globex `01a0daf9`, con la lectura de `git-workflow` en la línea 306, el primer commit en la 334 y `gh pr merge` en la 382.
  - `grok-pass` desde el piloto `grok-b-2`, y `codex-pass` desde `codex-b-1` (lectura con `cat`). Las rutas del home paralelo pasan a ser relativas y el cuerpo de la skill se reemplaza por un texto corto de muestra.
- Rutas permitidas: `tests/pilot/regression.go`, `regression_test.go`, `flows.go` (solo para extraer el helper) y `flows_test.go`, `tests/fixtures/regression/flow_skill_read_before_delivery/**` y el README.

**Verificación:** prueba antes y después del cambio, reversión del criterio para confirmar que el caso que falla sí falla, `go test -race ./tests/pilot/...`, `go vet ./tests/...` y el escaneo del README sin resultados.

### T3: Entrega (D21-A)

- [ ] `/code-review` sin hallazgos abiertos, commit y push a `rebuild/harness-engineering`, despliegue a los seis hosts verificado, issue #33 cerrado con el commit, y la carpeta de cambio archivada.

**Ejecución:** hilo principal.

**Verificación:** `go test ./...`, `-race` y `go vet` pasan. `origin/rebuild/harness-engineering` apunta al commit nuevo, `status` da todo como `installed`, y la viñeta aparece en los `AGENTS.md` de los hosts.

## Estado de la revisión y avance

- Borrador del 2026-09-26 con las decisiones D18-A, D19-A, D20-A y D21-A.
- Ronda 1 de revisión del plan (2026-09-26), 2 dominios, en paralelo:
  - **Guía:** 3 hallazgos (el disparador era demasiado amplio, faltaba el límite para subagentes, el criterio era más débil que la regla).
  - **Criterio:** 8 hallazgos. Entre ellos, que globex sí leyó `git-workflow`, las lecturas de Codex por `cat`, las skills cargadas con un comando escrito, el formato de la traza y contar los intentos de Git.
  - Todas son propuestas de los propios revisores y se aplicaron sin una segunda ronda.
- T1 terminada (2026-09-26): viñeta en `global.md:78`, justo después de la de `flow-research`, con +244 bytes. `globalGuidanceBudget` queda en 36976 y `go test ./tests/content/...` pasa.
- T2 terminada (2026-09-26, `test-engineer`, revisada por el hilo principal):
  - `skillContentReads` y `skillReadIndex` se extrajeron de `observeSkill` sin cambiar su comportamiento; una prueba lo demuestra.
  - `flowSkillReadBeforeDelivery` quedó sin conectar a los criterios de flujo.
  - Hay 3 casos con su `provenance.md`.
  - RED observado antes de GREEN, con la reversión del criterio confirmada.
  - `go test ./...`, `-race` en `tests`, `go vet` y el escaneo pasan.
- `/code-review` alto (2026-09-26), 10 hallazgos, todos con fundamento:
  - **Corregido en el hilo principal:**
    - 3: la viñeta ya no menciona `git-workflow`; queda en 36,936 bytes con el presupuesto ajustado.
    - 5: los límites del modelo de trazas quedaron registrados en [#34](https://github.com/JhonHawk/tricell-hive/issues/34), como pide `AGENTS.md`.
  - **Delegados a `test-engineer`:** 1, 2, 4, 6, 7, 8, 9 y 10 (lectura con contenido, orden por resultado, `git merge` y `gh api`, paridad del helper, comentarios, deduplicación y rendimiento).
- Correcciones de `/code-review` (2026-09-26, `test-engineer`, revisadas):
  - 1, 2, 4, 6, 7, 8, 9 y 10 corregidos.
  - Desvío aceptado: la exigencia de contenido aplica solo al criterio (`skillReadIndex`, en modo estricto) y no a `observeSkill`. La lectura nativa sintética de Claude no tiene un resultado correlacionado con contenido, y una prueba lo documenta.
  - `go test ./...`, `-race`, `go vet`, `gofmt` y el escaneo pasan.
