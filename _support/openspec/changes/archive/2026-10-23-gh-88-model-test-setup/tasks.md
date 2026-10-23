# Tareas y verificación

Base real de implementación: `7b0775c803a05f1ac97732bec13b246a13e37222`, coincide con planificación y `origin/development` tras fetch. Implementación autorizada el 2026-10-23; entrega automática y sin revisión dedicada.

## T1 — Preparación por host del catálogo grande

- [x] Preparar los seis hosts con planes separados y conservar los 22 roles y ambos tamaños.

**Closes:** AC1 (catálogo completo con planes de un host).

**Dependencias:** ninguna.

**Ubicación y propiedad:** `tooling/cli/tui_models_view_test.go`, exclusivamente `TestModelsViewFitsAndScrollsWithManyRolesOnSixCLIs` y sus assertions locales.

**Ejecución:** hijo de implementación de pruebas; unidad acotada sin archivos compartidos con T2 (hosts mínimos). No modificar helpers generales.

**Enfoque:** `characterization`; conservar la prueba existente y añadir comprobación explícita de host esperado y catálogo completo por host, sin cambiar el resultado esperado.

**Cambio:** D-A (planes reales separados) y restricciones de D-C (preservar escenarios multihost) en `design.md`. Cada llamada del bucle usa el helper existente `installDirect`, con un único host y los mismos home, estado y fuente.

**Verificación:** `go test -race -count=1 -run '^TestModelsViewFitsAndScrollsWithManyRolesOnSixCLIs$' ./tooling/cli`. Debe pasar ambos tamaños, seleccionar exactamente los seis hosts esperados y confirmar los 22 roles por host. Comparar duración con el padre aislado base de 32,21 s; si no mejora, no ampliar el alcance y volver a investigación. Inspeccionar que el plan por llamada contiene un único host.

## T2 — Hosts mínimos en diez escenarios de modelos

- [x] Aplicar exclusivamente las diez reducciones de D-B (preparación de un host por escenario), conservando sus assertions y subcasos.

**Closes:** AC3 (preparaciones mínimas de los diez casos).

**Dependencias:** ninguna; puede ejecutarse en paralelo con T1 (catálogo grande), porque no comparte archivos de escritura.

**Ubicación y propiedad:** los cuatro archivos `tui_models_edit_test.go`, `tui_models_picker_test.go`, `tui_models_review_fixes_test.go` y `tui_models_followups_test.go`; sólo los diez casos de D-B (hosts mínimos). El helper `twin` y los otros casos se preservan.

**Ejecución:** otro hijo de implementación de pruebas, con esta propiedad exclusiva. No editar `tui_models_view_test.go` ni helpers generales.

**Enfoque:** `characterization`; conservar exactamente proveedores, tamaños, entradas, expectativas y aserciones de cada escenario.

**Verificación:** ejecutar el comando siguiente. Los diez padres y sus mismos subcasos deben pasar; inspeccionar cada llamada para confirmar el host de la tabla. No eliminar la navegación/comparación que realmente prueba más de un host.

```sh
go test -race -count=1 -run '^(TestModelsViewMarksAnOverriddenRoleAndKeepsTheMarkNextToTheCursor|TestModelsViewFlatListsHaveNoSectionHeaders|TestModelsViewProviderHeadersAreReadableWithAFallback|TestModelsViewTwoHundredNineModelsReachTheLastOneWithPageDown|TestModelsViewListDoesNotRepeatAnEffectiveModelTheCLIListsAndKeepsCatalogOrder|TestModelsConfirmationFitsEightyColumnsWithLongOpenCodeIDs|TestModelsConfirmationOfASevenRoleGroupKeepsOneRowPerRole|TestModelsViewCodexEffortResetSaysSoInThePanel|TestModelsViewPickerShowsThePositionAmongTheModels|TestModelsViewXOnAGroupWithoutOverrideSaysThereIsNothingToReset)$' ./tooling/cli
```

## T3 — Medición comparable y cierre de cobertura

- [x] Medir el candidato final y decidir con evidencia si cumple el objetivo local de cinco minutos.

**Closes:** AC2 (duración local del CLI inferior a 300 segundos).

**Dependencias:** T1 (catálogo grande) y T2 (hosts mínimos) verificadas e integradas localmente.

**Ejecución:** hilo principal, porque reúne el candidato, las mediciones y la decisión de continuar o retornar a investigación. Verificación independiente del cambio con un hijo distinto de los implementadores, limitado a comparar escenarios, revisar las pruebas conservadas y ejecutar los casos modificados; sin repetir la suite completa.

**Enfoque:** `check`; comparar nombres de escenarios y tiempos con el JSON base, sin sumar subtests y padres. Registrar revisión, Go, plataforma, orden, caché y concurrencia. No afirmar mejora general a partir de una comparación local.

**Verificación:** leer el resultado del gate compartido G1 (corrida completa del CLI con carreras) y comprobar `<300 s` de duración de paquete, ningún fallo y el mismo conjunto de 685 pruebas principales y de subcasos anteriores. Confirmar por diff que se conservan las pruebas de D-C (multihost) y no hay cambios productivos. Si falta un escenario o AC2 (duración inferior a cinco minutos) no se cumple, registrar el incumplimiento y volver a investigación; no declarar resuelto el ticket ni debilitar cobertura.

## Gates compartidos

- **G1 — Medición final del CLI:** una sola corrida del candidato final: `go test -race -json -count=1 -timeout=12m ./tooling/cli`. Comparar con 316,680 s de paquete y 319,812 s totales de la base, sin otra suite simultánea. Es la medición final, no repetirla por tarea.
- **G2 — Regresión Go:** `go vet ./...` y `go test ./...`, como pide `CONTRIBUTING.md`. No hay cambios de producción ni concurrencia; no repetir `go test -race ./...` completo además de G1 (CLI con carreras). Si la implementación sale del alcance de pruebas, detenerse y revisar el plan.
- **G3 — Verificación independiente:** hijo de verificación distinto de los implementadores, candidato final identificado, sólo escenarios modificados y comparación explícita de precondiciones/assertions anteriores y nuevas. Ejecutada y aprobada; evidencia abajo.
- **Revisión dedicada de código:** usuario eligió no ejecutarla. Se conserva G3 (verificación del cambio).

Las pruebas ejercen el Engine real en homes sintéticos y son la verificación local del comportamiento modificado. No cambia la interfaz renderizada del producto; no aplica una nueva revisión visual, navegador, visor cotidiano ni reinicio de hosts. No se incluyen pilotos de modelos. Las suites Python y sitio no se ven afectadas por estos archivos Go de pruebas.

Las verificaciones pasaron sin reintentos de suites completas. Se conservan las mediciones originales como evidencia única.

## Estado de revisión y progreso

- 2026-10-23: investigación y perfil aislado completos. Diseño contrastado con código de `installDirect`, `Onboard`, `desiredResources` y preservación de registros de `prepareTransaction`.
- Revisión independiente del plan completada por `hive-review-plan` nativo, dominio pruebas/rendimiento, sin hallazgos materiales. Contrato asignado sólo lectura; aislamiento técnico adicional no verificado. Candidato revisado: hashes SHA256 `f66e1a0e…` de `proposal.md`, `69719ac3…` de `design.md` y `ce6ecc53…` de `tasks.md`. Duración y tokens del revisor no reportados por el host.
- Reconciliación del orquestador: el comando de selección de T2 (diez hosts mínimos) se expandió literalmente y el siguiente paso del encabezado se actualizó; sin cambio de criterio, escenario o contrato. No requiere otra revisión.
- D3-A/D4-A/D5-A (entrega automática, sin revisión dedicada e iniciar) resueltas por el usuario el 2026-10-23. Las decisiones técnicas revisadas no cambiaron. Tracker GitHub sin estado intermedio/Project ni subissues; issue abierto durante implementación y cerrado al integrar.
- No hay delta de requisitos. La rama de implementación sólo cambia cinco archivos de pruebas; `_support/sessions/` previo se conserva en el checkout principal.
- Implementación: T1 (catálogo grande) pasó el comando aislado con carreras, paquete 11,210 s y tiempo total 14,405 s; ambos tamaños y assertions nuevas aprobadas. T2 (diez hosts mínimos) pasó su comando literal con carreras, paquete 5,369 s. Diff cotejado por el orquestador: un test cambia preparación y refuerza catálogo/hosts; diez argumentos locales cambian en los otros cuatro archivos, assertions intactas. Gofmt/diff-check correctos; procesos de implementadores finalizados.
- Los dos implementadores usaron el rol nativo `hive-write-tests`, modelo configurado `gpt-6.1-sol` medium, no observado por metadatos de ejecución.
- Commits por tareas verificadas: `19b74f5b` para T1 (catálogo grande), `67384a0d` para T2 (diez hosts mínimos). Escaneo Gitleaks por stdin del diff de los cinco archivos: exit 0. Hooks de Git no omitidos.
- G3 (verificación independiente) ejecutó selección T1+T2 con carreras y pasó en 14,291 s de paquete; 11 padres, cuatro subtests de listas y dos tamaños del catálogo. G1 (paquete CLI completo) se ejecutó después de terminar sus pruebas, sin otra suite simultánea.
- G3 (verificación independiente) terminó sin defectos; confirmó precondiciones, proveedores, tamaños y assertions conservadas en los diez casos, y los escenarios multihost de D-C sin cambios. Candidato `67384a0d47027a5b713c36711439d1f877425a0d`.
- G1 (medición final) pasó: paquete 274,062 s; tiempo total 274,455 s. Base 316,680 s: diferencia local 42,618 s (13,46 %). El catálogo grande pasó en 8,960 s frente a 31,630 s en la corrida completa base. No demuestra una mejora general ni toda la causa de la duración histórica. Baseline en checkout principal, candidato en worktree aislado; mismo Go/host y sin otra suite simultánea.
- Evidencia G1: `/var/folders/dg/0d963h2158s10jggj2_k2w_00000gn/T/hive-issue88-final-j5y8256u.jsonl` y metadata `.run.json` con comando, revisión, checksum, tiempos y límites. T3 (comparación final) y G2 (vet y regresión Go) aprobadas.
- Verificación independiente T3: `hive-verify-task` nativo, perfil configurado `gpt-6-astra` high, no observado en metadata de ejecución, AC2 (duración final inferior a cinco minutos) `met`. Recalculó checksum y comparó ambos JSON: mismos 685 padres aprobados, 441 subcasos aprobados y tres omisiones; mismo orden de padres y cuerpos multihost protegidos intactos. El orquestador cotejó las cifras y conjuntos en ambos archivos. No se repitió una suite completa para esta verificación. Duración/tokens del verificador no reportados por el host.
- G2 (regresión) pasó: `go vet ./...` exit 0, 0,596 s totales; `go test ./...` exit 0, 20 paquetes, 98,201 s totales. Logs locales `hive-issue88-gates-njcm_gxx.log` y `hive-issue88-gates-5tx9dh0n.log` en el temporal nativo. Los tests ordinarios de `tests/pilot` no ejecutan sesiones de modelos; los pilotos siguen pausados.
- Las tres tareas están verificadas. T3 (medición) no cambia código; su registro se versionará con el archivo del plan, una sola vez al cierre. No se ejecutó revisión dedicada, como pidió el usuario. No quedan procesos propios de pruebas.
- Entrega de código completada: PR #98 (preparación de modelos) integrado en `development` `3e5d789e105316b8ce62194a7082ab843b33f11d`, ticket CLOSED, sin subissues. Se comprobó que el head contenía la base actual antes del merge. No hubo CI; CodeRabbit sólo informó un contexto SUCCESS, no se presenta como revisión dedicada.
- Binario local reconstruido desde ese merge con el comando mantenedor `go build -o "$(command -v hive)" ./tooling/cli`, exit 0. Metadata de Go confirma `vcs.revision=3e5d789e105316b8ce62194a7082ab843b33f11d`; `vcs.modified=true` corresponde al `_support/sessions/` previo sin seguimiento, no a cambios rastreados. Sin reiniciar visor ni desplegar contenido. Registro trasladado a archive y enlaces relativos de código corregidos por su nueva profundidad; no hay cambios en las decisiones técnicas revisadas.
