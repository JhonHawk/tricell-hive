# Tareas

Base: `development` en `7ae26b0e`; rama `feat/model-qualification-tool` en un worktree. Las tareas de Go escriben en el mismo paquete `tests/pilot` y van en serie: T4 → T1 → T6 → T2 → T3 → T7. Cada tarea de Go la ejecuta un hijo `hive-build-backend` (cada una lee más de 5 archivos grandes de `tests/pilot`); T7 el hilo principal. Todo test nuevo que lance el runner usa un `HOME` temporal y binarios falsos en `PATH`.

### T4 — Suite, registro de detectores y set de tareas

- [x] `qualify_suite.go` con el tipo, el cargador, el validador (forma y cantidades por separado), el registro de detectores y los chequeos nuevos; `suite.json` con al menos 15 tareas que pasan la validación; README de la suite.

**Closes:** AC39

**Depends on:** ninguno

**Locations:** `tests/pilot/qualify_suite.go` (nuevo), `tests/pilot/qualify_suite_test.go` (nuevo), `tests/fixtures/qualification/suite.json` (nuevo), `tests/fixtures/qualification/README.md` (nuevo)

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** `design.md` «Suite y registro» y «Set de tareas». Los adaptadores devuelven `criterionAssessment`.

**Verification:** `go test ./tests/pilot/ -run QualifySuite` pasa sobre el `suite.json` real y rechaza cada copia modificada que lista AC39; una suite de 2 tareas pasa la validación de forma y falla la de cantidades; tests de `git_unchanged`, `no_new_commit` (un commit sobre `main` existente falla) y `commands_must_not` (`git -C x push` se detecta, `echo "git push"` no).

### T1 — Brazos `deployed`, A y `none` en la suite `qualification`

- [ ] El runner corre una tarea de `suite.json` con `--suite qualification` y los tres brazos según AC40, con fixtures en `os.TempDir()`, y registra en `run.json` lo que pide AC56.

**Closes:** AC40, AC56

**Depends on:** T4

**Locations:** `tests/pilot/main.go`, `tests/pilot/guidance_variant.go`, `tests/pilot/guidance_variant_test.go`, `tests/pilot/qualify_fixture.go` (nuevo), `tests/pilot/qualify_fixture_test.go` (nuevo)

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** `design.md` «Brazos del runner».

**Verification:** `go test ./tests/pilot/ -run 'GuidanceVariant|QualificationArm|QualificationFixture|GuidanceHash'` pasa con hosts simulados (binarios falsos en `PATH` y `HOME` temporal, como `flash_runner_test.go`): `deployed` sin `--guidance-source` se acepta en los cinco hosts; `none` en `claude`, `grok` y `pi` sale con el error sin lanzar; el home `none` de Codex y OpenCode no tiene ningún archivo con el marcador del bloque gestionado ni skills o agentes de Hive, y sí tiene la configuración de credenciales y memoria; un `none` de OpenCode con directorio de trabajo bajo el home real se rechaza; la raíz del fixture no está bajo `--source` ni bajo `--out`; `run.json` trae tarea, `split`, brazo, esfuerzo y hash; `guidanceHash` da el mismo valor con el mismo contenido y entre dos instalaciones del mismo checkout en dos homes temporales distintos, otro tras editar un archivo gestionado o cambiar el destino de un registro `symlink`, y no cambia al editar texto fuera de los marcadores ni al cambiar un registro de otro host; una tarea con `case: research` recibe el prompt de `flowTaskPrompt`; los tests existentes de `guidance_variant_test.go`, `flows` y `flash` pasan sin cambios.

### T6 — Modelo servido de OpenCode desde el export

- [ ] Una corrida `qualification` de OpenCode fija `ModelObserved` desde `opencode session export` antes de limpiar el home aislado; Codex queda `not_observed`.

**Closes:** AC49

**Depends on:** T1

**Locations:** `tests/pilot/main.go`, `tests/pilot/trace.go`, `tests/pilot/trace_test.go`, `tests/pilot/qualify_fixture_test.go`

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** `design.md` «Modelo servido en OpenCode».

**Verification:** `go test ./tests/pilot/ -run 'QualificationServedModel|Trace'` pasa: con un `opencode` falso cuyo export trae `{"id":"deepseek-v4.1-flash","providerID":"opencode-go","variant":"high"}`, `ModelObserved` es `opencode-go/deepseek-v4.1-flash#high`; con `Info.Model` y el último mensaje del asistente distintos gana el mensaje; el export se lanza una sola vez por corrida; en el brazo `none` el export corre antes de que se borre el directorio de datos del home aislado; con la corrida cerca de su plazo, el export usa su propio timeout; sin modelo en el export, `not_observed`; en Codex, `not_observed`; `TestNativeTraces` sigue igual.

### T2 — Puntuador y veredicto

- [ ] `qualify score` agrega intentos utilizables (D21), calcula la estadística de D18, MISS de D19 y la lectura de la skill, y emite el veredicto o se niega con código 1.

**Closes:** AC43, AC44, AC57

**Depends on:** T4, T1, T6

**Locations:** `tests/pilot/qualify_score.go` (nuevo), `tests/pilot/qualify_score_test.go` (nuevo)

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** `design.md` «Puntuador». Una función del paquete que recibe el directorio, la suite cargada, el documento del piso y la referencia, y devuelve el informe y el veredicto; T3 la expone.

**Verification:** `go test ./tests/pilot/ -run QualifyScore` pasa con evidencia sintética y suites chicas:
- tasa, error entre tareas y diferencia pareada iguales a valores calculados a mano; varianza cero con `d_t = 0` pasa y con `d_t` iguales y negativas falla;
- `candidate`; `candidate-partial` en Claude; `not-candidate` por tasa 0.89, por MISS (referencia 9/9, candidato 3/9) y por diferencia menor que −2 errores;
- candidato 4/9 contra referencia 9/9 no es MISS; candidato 9/9 contra referencia 0/9 no es MISS; una tarea `tuning` no cambia el veredicto;
- `timeout` y `not_verified` no cuentan y el intento utilizable anterior sí; `blocked_permission` y `background_unverified` cuentan como fallo y `blocked_permission` aparece en el conteo del informe; `OutsideChanges` hace fallar; un criterio `not_observed` no listado como aceptado hace fallar;
- intentos `completed` no utilizables (D21): hash distinto del de `batch.json`, esfuerzo distinto, variante de OpenCode distinta, modelo servido distinto o ausente en Claude; cada uno no cuenta, aparece por causa en el informe y deja la celda sin repetición utilizable si era la única;
- un intento de OpenCode con `Effort` vacío en `run.json`, pedido `opencode-go/qwen3.8-flash#medium` y `batch.json` con `medium` es utilizable; con `#high` pedido no lo es;
- cada negativa de AC57 sale con código 1 y sin veredicto: sin `batch.json`, `--reference` distinta de la de `batch.json`, repeticiones utilizables desiguales entre brazos comparados, celda programada sin repetición utilizable;
- `opencode-go/claude-haiku-5-5#high` pedido contra servido `claude-haiku-5-5` de otro proveedor se acepta; `opus` pedido contra `claude-opus-5-5[1m]` servido se acepta vía «Aliases»;
- dos modelos de Codex con el mismo `--configured-model` se informan cada uno con su `ModelRequested` como «configured».

### T3 — Subcomandos `qualify run` y `qualify score`

- [ ] `qualify run` escribe o comprueba `batch.json`, calcula el calendario, exige `--confirm-sessions` igual a lo que va a lanzar, resuelve el esfuerzo, lanza el runner por intento y relanza solo repeticiones sin intento utilizable; `qualify score` imprime el informe de T2.

**Closes:** AC58

**Depends on:** T1, T6, T2

**Locations:** `tests/pilot/main.go` (despacho de `qualify`), `tests/pilot/qualify_run.go` (nuevo), `tests/pilot/qualify_run_test.go` (nuevo)

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** `design.md` «Driver».

**Verification:** `go test ./tests/pilot/ -run QualifyRun` pasa con el binario de test como runner, hosts falsos y suites chicas:
- sin confirmación o con un número distinto sale con código 2 y cero lanzamientos; con el correcto lanza exactamente ese número;
- el calendario de un candidato de Codex con 2 tareas `heldback` (1 `compare_none`) y 3 repeticiones es 2·3 + 1·3·2 + 2·3 = 18; tras completar 5, la segunda llamada exige 13 y no relanza repeticiones con intento utilizable; una repetición cuyo único intento es `completed` con otro modelo servido entra en el conteo y se relanza en `attempt-2`;
- la primera llamada escribe `batch.json` con hashes y esfuerzos; si antes de la segunda cambia un archivo gestionado del home `deployed` simulado, sale con código 2 nombrando ambos hashes y no lanza nada; lo mismo si cambia `--models`, `--reference` o un `--effort`;
- un intento A con la instalación real del checkout en el home aislado de la corrida es utilizable (su hash coincide con el de `batch.json`);
- solo A recibe `--guidance-source`; `--provider` solo en Pi;
- esfuerzo: `sonnet` en Claude recibe `high`; `gpt-6.1-sol` en Codex sin `--effort` sale con código 2; `grok-4.6` no recibe `--effort`; la referencia fuera de la tabla sin `--effort` sale con código 2; OpenCode recibe `#high` en el modelo y nunca `--effort`;
- `<out>` fuera de `_support/workspace/` se rechaza; un intento con solo el `run.json` previo se conserva y el siguiente usa `attempt-2`; un modelo con `/` y `#` produce una ruta válida;
- `qualify score` sobre esa salida imprime el informe con veredicto.

### T7 — Documentación del piso y de contribución

- [ ] `model-floor.md` y `CONTRIBUTING.md` según AC59.

**Closes:** AC59

**Depends on:** T3

**Locations:** `_support/docs/harness-engineering/model-floor.md`, `CONTRIBUTING.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** texto de AC59; las tablas «Floor» y «Aliases» y las reglas 1 y 2 de «Membership and normalization» no cambian; de la regla 3 solo cambia su paréntesis.

**Verification:** `go test ./tests/content/` pasa (incluido `TestModelFloor`); `rg -n "candidate none|not built yet|open for revision|of its runs|record the served model yet|record it yet" _support/docs/harness-engineering/model-floor.md` no devuelve nada; `git diff` de «Membership and normalization» solo toca el paréntesis de la regla 3; cada flag que nombra la sección existe en `qualify run --help` o `qualify score --help`.

## Verificación y revisión humana

| Chequeo | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar la tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go vet ./... && go test ./...` | Cubierto por la implementación |
| Sin modelos reales | Candidato final, local | Cada `exec.Command` y cada `PATH` de los tests nuevos o cambiados apunta a binarios falsos escritos por el test o al binario de test, con `HOME` temporal | Cubierto por la implementación |
| Revisión de código | Candidato final, antes del push | `/code-review` de Claude Code (D6 A) | La respuesta de revisión |

Sin UI. Correr una calificación con modelos reales queda fuera de este cambio.

## Checkpoint

- Updated: 2026-11-08
- Stage: build · 2
- Last commit: 1265c17a
- Open findings: none
- Next: T1 via hive-build-backend child

### Evidence

- go test ./tests/pilot/ -run QualifySuite: ok @ 1265c17a

## Estado de revisión y progreso

- 2026-11-08 — tareas movidas desde `model-support-floor` por D13 A, sin revisar.
- 2026-11-08 — revisión 1: T4, T1, T2, T3 y T6 reescritas tras D18–D20 y la lectura del código; T7 nueva; ninguna tarea hecha.
- 2026-11-08 — revisión 2: hallazgos H19–H30 de `hive-review-plan`; AC41, AC42, AC45, AC46 y AC47 retirados y reemplazados por AC56–AC59; T7 pasa a depender de T3; ninguna tarea hecha.
- 2026-11-08 — revisión 3: D21 A (intentos no utilizables se relanzan) y H31–H39; T2, T3, T6 y T7 ajustadas; ninguna tarea hecha.
- 2026-11-08 — revisión 4: D22 A, hallazgos H40–H45 de la tercera revisión aplicados sin cuarta revisión; AC48, AC53, AC54 y AC55 retirados y reemplazados por AC56–AC59; ninguna tarea hecha.
2026-11-08 — T4 · 1265c17a · unverified: none
