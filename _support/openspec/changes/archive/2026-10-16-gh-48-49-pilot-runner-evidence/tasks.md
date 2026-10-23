# Tareas — gh-48-49-pilot-runner-evidence

Commit base: `69dce947609f` (`development`).

### T1 — Evidencia privada de operaciones e inventario con tipo (#48)

- [x] El runner guarda texto de `shell`/`execute` y patrones de `glob`/`grep` en un archivo privado `0600` fuera de `run.json`; la inspección decide con él; el inventario registra el tipo de cada entrada o excluye `.git/`.
  Ronda 1 de verificación tras la corrección de la revisión: AC1 not met: `flashStripQuotes` sustituye cada tramo entre comillas por `Q` y el shell concatena, así que `cat ""/etc/passwd`, `cat .''./etc/passwd` y `rg needle ''/etc` salen `pass` fuera de la fixture; AC2 met. Devuelto al implementador para una ronda de corrección; reverificar solo AC1.
  Ronda 2 de verificación (solo AC1): las comillas pegadas quedan cerradas (siete formas rechazadas, dos válidas aceptadas), pero AC1 not met por un hueco nuevo del mismo tipo: `flashStderrRedirect` (`2>/dev/null|2>&1`) no exige borde después, así que `cat docs/a.md 2>&1x` y `cat docs/a.md 2>/dev/nullx` pasan como solo lectura y el shell real crea `1x`. Tope de dos rondas alcanzado: decisión del usuario D1-A (2026-10-16): una ronda más acotada al borde de la redirección (tokens completos), reverificación solo de AC1; si falla, se vuelve a preguntar.
  Ronda 3 de verificación (solo AC1): met. `flashStderrRedirect` exige borde a ambos lados y se aplica hasta estabilizar; las once sondas pedidas y una veintena de formas propias del verificador (redirecciones a archivo, destinos entre comillas o escapados, tokens malformados, encadenados con `tee`/`rm`, rutas fuera de la raíz, flags prohibidos) se rechazan; `2>/dev/null`, `2>&1 | rg x` y `rg -n "git push" README.md` se aceptan. Nit sin caso propio: el bucle de repetición (`2>/dev/null 2>&1`), que si se rompe solo endurece.

**Closes:** AC1, AC2.

**Depends on:** none.

**Locations:** `tests/pilot/flash_class.go:267-285`, `tests/pilot/trace.go`, `tests/pilot/flash_protocol.go`, `tests/pilot/assess.go`, `tests/pilot/main.go:147-171,209,249,878`, pruebas `tests/pilot/flash_command_security_test.go`, `flash_d22_test.go`, `assess_test.go` y las nuevas.

**Execution:** delegada a `hive-build-backend` junto con T2 en un solo encargo, porque ambas tocan `flash_class.go`.

**Test approach:** tdd.

**Changes:** según "Diseño propuesto"; el archivo privado nunca se copia a los artefactos publicados ni se referencia desde `run.json` más que por nombre.

**Verification:** `go test -count=1 ./tests/pilot/` en verde con pruebas nuevas que (a) escriben una operación fuera de la gramática y comprueban el archivo privado con modo `0600`, su contenido y que la inspección la acepta o rechaza por el texto; (b) comprueban que `run.json` no contiene el texto; (c) comprueban el tipo de entrada en el inventario y la exclusión o el tipo de `.git/`.

### T2 — Cobertura por contenido normalizado (#49)

- [x] `inspectReference` y `flashCoverage` comparan contenido normalizado; una lectura con prefijos de línea da `observed`.

**Closes:** AC3.

**Depends on:** T1 (mismo archivo `flash_class.go`; mismo hijo, en orden).

**Locations:** `tests/fixtures/flash-class-guidance/observer.js:46-68`, `tests/pilot/flash_class.go:358-390`, `tests/pilot/flash_class_test.go`, `flash_observer_test.go`.

**Execution:** delegada a `hive-build-backend` (mismo encargo que T1).

**Test approach:** tdd.

**Changes:** confirmar el formato de `read` de OpenCode en su fuente o documentación (Context7 `/sst/opencode`) y anotarlo en el reporte; normalizar antes de comparar; hash del normalizado; conservar la copia byte a byte como caso válido.

**Verification:** prueba nueva con un texto canónico y una lectura con prefijos `N| ` por línea que da `observed`; la prueba existente de copia exacta sigue en verde; `go test -count=1 ./tests/pilot/`.

### T3 — Entrega automática y cierre

- [x] PR fusionada en `development`; #48 y #49 cerrados (se completa en el cierre que versiona este registro en la misma PR). Candidato final: `go vet ./...` limpio, `go test ./...` 18 paquetes en verde.

**Depends on:** T1, T2.

**Verification:** `go vet ./...`, `go test ./...`; merge observado; issues cerrados.

## Verificación y revisión humana

| Compuerta | Momento | Mecanismo | Autoridad |
| --- | --- | --- | --- |
| Verificador por tarea | Tras T1 y tras T2 | `hive-verify-task` sobre AC1–AC2 y AC3 | Implementación |
| Suite completa | Candidato final | `go vet ./...`, `go test ./...` | Implementación |
| Revisión de código | PR abierta | `/code-review` de Claude Code | D3-A |

Sin efecto de UI.

## Estado de revisión y progreso

- 2026-10-16: rama creada desde `69dce947`; T1 y T2 delegadas a `hive-build-backend` (Sonnet, configurado, no observado) en un encargo, TDD con RED observado en las seis pruebas nuevas.
- Decisiones de diseño registradas (implementador, aceptadas por el orquestador): archivo `operations-private.jsonl` junto a `run.json`, `O_EXCL` y `0600`, una línea por operación (sesión, agente, herramienta, clase `outside_grammar`/`pattern`, ruta, texto); `run.json` solo lleva el nombre (`OperationsFile`) y `traceEvent.PrivateText` va con `json:"-"`. `execute` se guarda pero nunca se acepta automáticamente (ejecuta JS arbitrario): queda para la revisión humana tras el lote, como pide #48. `.git/` se inventaría con tipo en vez de excluirse, porque los casos de regresión de entrega Git afirman sobre `.git/HEAD` y `.git/refs/*`; la contención exige que toda entrada del inventario sea `file`. Cobertura: se normaliza solo el candidato (prefijos `^[ \t]*\d+[|:][ \t]?` por línea), el canónico solo se recorta al final; normalizar ambos lados rompería un canónico con líneas como `10: x`. Formato de `read` de OpenCode: 2.x emite `Read file <path>, lines a-b` y `N: línea` (fuente `packages/core/src/tool/plugin/read.ts`); el formato 1.x `NNNNN| línea` se cubre como histórico; no hay evidencia retenida del lote `fcg-20261012c` para confirmar lo que emitió 2.0.27.
- Verificación: `hive-verify-task` (Opus, configurado, no observado) sobre T1: AC1 met (con la salvedad de `execute`), AC2 met; sobre T2: AC3 met. `go vet ./...` limpio, `go test ./...` 18 paquetes en verde, `go test ./tests/pilot/` ~22 s.
- 2026-10-16: PR [#63](https://github.com/JhonHawk/tricell-hive/pull/63) abierta (b2e6a470 T1+T2, changelog). `/code-review` ronda 1: 10 hallazgos, 6 de corrección o seguridad en el clasificador nuevo (falso `unauthorized_write` por `git push` entre comillas; flag entre comillas elude la lista de argumentos; búsquedas de shell sin contención de ruta; `r.After` nulo en la corrida en vivo, la contención solo se evalúa en `--assess`; `grep` sin patrón pasa en silencio; `scan.Err()` sin comprobar), 4 de reutilización o claridad (error sobrescrito en `main.go:807`, regex duplicada en `operation_safety.go`, contención duplicada frente a `eventPath`/`under`, clase `execute` propia). Ronda de corrección devuelta al mismo implementador; T1 vuelve a `[/]` hasta reverificar AC1 y AC2.
- 2026-10-16: correcciones de revisión y verificación commiteadas en ef043c49. `/code-review` ronda 2: 9 hallazgos, 3 de corrección confirmados con sonda (texto no reconocido pasado a `flashGitWrite` → falso `unauthorized_write` en `git log --grep commit`; patrón `{/etc,x}/**` escapa la contención; `cat` de un enlace simbólico que apunta fuera no pasa por la prueba de contención, que solo corre tras búsquedas), 1 de sobrerrechazo (patrón regex de `grep` juzgado como ruta), 5 de robustez o claridad (texto sin tope al escribir frente a tope de lectura, error sobrescrito, duplicación del marcado opaco, tres pasadas de la compuerta por corrida, comilla tipográfica en un comentario). Devueltos al implementador; AC1 se reverifica una vez más; sin tercera ronda de revisión (correcciones pequeñas dentro del cambio, cubiertas por pruebas).
- 2026-10-16: nueve hallazgos de la ronda 2 corregidos (TDD, 9 RED). Verificación final de AC1: met; 13 sondas pedidas y 24 formas propias del verificador con el resultado esperado. Decisión de diseño registrada (orquestador): el texto guardado se corta a 64 KiB y se marca `truncated`; un texto truncado nunca se acepta solo, así que "texto completo" de AC1 se lee como "completo hasta 64 KiB, con marca de corte" para que una entrada enorme de `execute` no deje ilegibles las siguientes. Nits sin corregir: la aserción "texto truncado no se acepta" de `TestFlashPrivateOperationTextIsCappedAtWrite` usa un `cat` que la gramática ya rechaza por tamaño (debería usar un glob); `git --config-env x=Y push` sale opaco en vez de `fail` (no pasa en silencio).
- Hallazgos incidentales (no corregidos aquí): (a) el texto completo de los comandos `shell` del agente primario llega a `run.json` y `events.json` por `addTool` (`trace.go:344`), sin pasar por la gramática D22; preexistente en la base, gravedad media si hay comandos con forma de credencial; acción sugerida: pasar los eventos `shell` de stdout por `flashOperationCommand`. (b) `gofmt -l` señala `tests/pilot/regression.go` desde antes de este cambio. (c) El clasificador reutiliza la gramática y rechaza `head -5`, `rg -e` y `--glob`; ampliar la gramática queda fuera de alcance.
