# Tareas

Base del cambio: `0292ff4` (`origin/development`). Worktree: `.claude/worktrees/reinstall`, rama `feat/reinstall-missing-managed-files`.

## T1 — Reponer y quitar lo ausente en el gestor

- [x] (`hive-verify-task`, Opus: AC1 a AC4 cumplidos, con 21 mutaciones. Implementada por `hive-build-backend`, Sonnet configurado, junto con T2.
  - Antes del cambio fallaban con los rechazos viejos todas las pruebas de `reinstall_test.go`: reponer, quitar, recurso compartido, recuperación en cada punto de escritura y `Gone` falsificado. Después pasan.
  - Se ajustaron cuatro pruebas existentes que contradecían el diseño:
    - una firma cambió, porque `transformResource` recibe la acción;
    - `TestManagedDriftAndStalePlans/block` ahora edita dentro de los marcadores, así que sigue ejercitando el rechazo;
    - `TestLegacyReintroductionIsDetected`: un archivo antiguo de Hive reintroducido ya no se rechaza; la búsqueda de instalaciones antiguas lo sigue detectando y lo migra;
    - las dos pruebas de voz esperan «run hive install first».
  - Decisión del orquestador: `voice off` sobre un archivo sin bloques retira el registro sin escribir.) Con `Change.Gone`, instalar y actualizar reponen, y quitar no escribe nada.

**Closes:** AC1, AC2, AC3, AC4.

**Depends on:** ninguna.

**Locations:** primero, un paso mínimo que añade el contrato en `tooling/management/types.go` (`Change.Gone` y `VoiceChange.Gone`), para que T3 pueda empezar (N4). Después `types.go` (`Change`), `plan.go` (`BuildPlan`, `nextRecord`, `validatePlan`), `resources.go` (`transformResource`), `files.go` (el ayudante `isMissing`), `migration.go` (`:136`, `:373`, `:377`), `apply.go` (`:238`, `:299`) y sus pruebas.

**Execution:** delegada a `hive-build-backend`. Es el núcleo del gestor y va acoplado con T2, así que el mismo hijo hace T2 a continuación.

**Test approach:** tdd. Primero las pruebas de los cuatro criterios; fallan en la base, donde todo se rechaza:
- skill, agente y enlace borrados: `install` y `update` los reponen con sus bytes y permisos, `Status` queda limpio y `PlanUnchanged` es falso antes de la reposición;
- bloque de Hive borrado de un archivo que existe: se repone y se conserva el texto del usuario;
- archivo con bloque borrado entero: se crea con el bloque;
- `plan remove` con el archivo ausente: termina y retira el registro; con un recurso compartido, conserva el consumidor que queda y no repone;
- recuperación: se interrumpe una reposición con el mecanismo de fallas existente (el parámetro `fail`), y `Recover` restaura la ausencia cuando la escritura había terminado (`apply.go:680`) (N3);
- un `Gone` que no se corresponde con el disco, como un archivo presente o un plan manipulado, falla cerrado (B2);
- quitar un host de un recurso compartido ausente no escribe, aunque `After` conserve otros consumidores (B1).

**Verification:** `go test -count=1 ./tooling/management` en verde. Las pruebas de lo editado (`TestOwned*`, el enlace reapuntado de `shared_test`) siguen exigiendo el rechazo.

## T2 — Voz

- [x] (`hive-verify-task`: AC5 cumplido; `hive-verify-change`: confirmado en el binario real.) `voice set`, `voice off` y la voz dentro de `install` reponen o dan por quitado el bloque de voz ausente, y omiten las rutas sin bloque de Hive.

**Closes:** AC5.

**Depends on:** T1.

**Locations:** `tooling/management/voice.go` (`checkVoiceConflict`, `composeVoiceStep`, `addVoiceChangesForInstall` `:560`, `BuildVoicePlan` `:452`), `types.go` (`VoiceChange.Gone`), `apply.go:259`, `migration.go:151`, `voice_review_test.go:171` y `:188`.

**Execution:** el mismo hijo de T1, porque es el mismo paquete y el mismo contrato.

**Test approach:** tdd. Las dos pruebas que hoy exigen el rechazo pasan a exigir otra cosa: que se omita la ruta sin bloque, o el mensaje «run hive install first» cuando no queda ninguna. Se añaden pruebas de bloque de voz borrado con `set` y con `off`, de `off` quitando un bloque de voz que quedó sin su bloque de Hive, y de `off` que sigue quitando un bloque presente sin `Gone`.

**Verification:** `go test -count=1 ./tooling/management` en verde.

## T3 — CLI: resúmenes y texto de doctor

- [x] Ronda de corrección aplicada:
  - `TestInstallSummaryCountsReinstalledDeletedFile` (M9) y `TestShowVoiceSummaryListsGoneEntryWithEqualSpans` (M10) fallan con su mutación.
  - En `management` fallan con la suya `TestReinstallRecreatesAWholeDeletedSkillDirectory` (M8) y `TestRemoveOfOneHostWithADeletedVoiceBlockInASharedFileWritesNothing` (M18, con claude y grok). La segunda muestra que al quitar se descarta el registro compartido del bloque de voz sin escribir, y el siguiente `install` repone el bloque; se acepta por D1-A.
  - `TestReinstallRestoresADeletedAgentAndStillRefusesAnEditedOne` cubre el agente.

  Re-verificación (ronda 2, `hive-verify-task`, Opus): AC1 cumplido. M9, M10, M8, M18 y el agente fallan con su mutación, y `go test ./...` pasa.
- (Historial) Ronda 1, AC1 (la parte del resumen): no se cumple. El código es correcto, pero ninguna prueba falla si se quita `ch.Gone` de `changedChangeCount` (M9) o `vc.Gone` del resumen de voz (M10). En T1 faltan pruebas para un directorio de skill borrado entero (M8), para quitar a claude con el bloque de voz compartido con grok (M18) y para la reposición de un agente. Se abre una ronda de corrección que añade esas pruebas.
  (Implementada por `hive-build-backend`, Sonnet configurado. `changedChangeCount` y el resumen de voz cuentan `Gone`. `driftRepairText` tiene dos casos. `TestEditedManagedFilesAreRefusedAndDeletedOnesReinstalled` reemplaza a `TestNoCommandRepairsADriftedManagedFile`: lo editado se rechaza, y lo borrado se repone con `update` e `install`, mientras que `plan remove` no lo repone.) Los resúmenes cuentan `Gone`, y el texto de ayuda de `drift` distingue editado de borrado.

**Closes:** AC6, y la parte de CLI de AC1 (el resumen cuenta la reposición).

**Depends on:** el paso de contrato de T1. Arranca cuando T1 lo termina, porque `types.go` es de T1 y no debe haber dos escritores en el mismo archivo.

**Locations:**
- `tooling/cli/install.go:791` (`changedChangeCount`);
- `tooling/cli/voice.go:213`;
- `tooling/cli/doctor.go:431-437` (`driftRepairText`) y `doctor_test.go:630`;
- `tooling/cli/doctor_drift_test.go:35-130` (`TestNoCommandRepairsADriftedManagedFile`): se renombra, porque ya no describe la regla. El caso de archivo borrado pasa a exigir la reposición; los casos editados siguen rechazando; se añade «skill borrada».

**Execution:** delegada a un segundo `hive-build-backend`. Solo toca `tooling/cli`, sin cruzarse con T1.

**Test approach:** tdd, sobre la prueba de `doctor_drift_test.go` y el texto de doctor.

**Verification:** `go test -count=1 ./tooling/cli` en verde cuando T1 y T2 estén dentro.

## T4 — Documentación y especificación

- [x] (La búsqueda con `rg` de la regla vieja en la documentación y en `doctor.go` no devuelve nada.) `deployment-manager.md` y el delta de la especificación describen la regla nueva. Documentación hecha en el worktree:
  - `:71`: la ayuda de `drift` distingue editado de borrado;
  - la sección de preservación: lo borrado se repone y lo editado se rechaza; quitar un host nunca repone un recurso compartido; los planes y el diario llevan `gone`.

  El delta de la especificación está escrito en la carpeta del cambio. Falta que T3 cambie el texto de `doctor.go` para cerrar la búsqueda con `rg`.

**Closes:** AC6 (parte de documentación).

**Depends on:** ninguna.

**Locations:**
- `_support/docs/architecture/deployment-manager.md` (`:71` y `:156`);
- `_support/openspec/changes/reinstall-missing-managed-files/specs/versioned-installation/spec.md`, un delta nuevo con el escenario «Deleted managed file». Al cerrar, `flow-close` lo aplica a la especificación vigente.

**Execution:** hilo principal, porque es texto corto.

**Test approach:** check con `rg`.

**Verification:** `rg -n 'cannot repair|must still match its installed bytes' _support/docs/architecture/deployment-manager.md tooling/cli/doctor.go` no devuelve la regla vieja.

## T5 — Verificación de punta a punta

- [x] (`hive-verify-change`: AC1 a AC5 cumplidos en el binario real frente a la base; lo editado sigue rechazado y un `gone` falso falla cerrado. Salidas `e2e-*` en `_support/workspace/2026-09-30-reinstall-missing-managed-files/`.) `hive-verify-change` confirma AC1 a AC5 con el binario real.

**Closes:** AC1, AC2, AC3, AC4, AC5 (confirmación en el binario real).

**Depends on:** T1 a T4.

**Execution:** `hive-verify-change`.

**Test approach:** check.

**Verification.** Binario del candidato en un `HOME` temporal, sin `--home`, con `--state-dir` explícito, `PATH` limitado y `CODEX_HOME`, `CLAUDE_CONFIG_DIR` y `XDG_STATE_HOME` sin definir. Casos:
- borrar una skill y correr `update`: se repone y `status` queda limpio;
- borrar el bloque de `CLAUDE.md`: `install` lo repone y conserva el texto del usuario;
- borrar `CLAUDE.md` entero: `install` lo crea;
- `plan remove` y `apply` de un host con su archivo borrado: terminan; con un recurso compartido, el archivo sigue ausente y el siguiente `install` lo repone (segunda mitad de AC4);
- `voice set` y `voice off` con el bloque de voz borrado;
- editar la skill: sigue rechazándose.

Las salidas se guardan en `_support/workspace/2026-09-30-reinstall-missing-managed-files/`.

## Verificación y revisión

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | T1 a T3 | `hive-verify-task` con otro modelo que el implementador | Parte de `flow-build` |
| De punta a punta | T5 | `hive-verify-change` con el binario real | Parte de la implementación |
| Suite local | Una vez sobre el candidato final | `go vet ./...`, `go test ./...` y las pruebas de Python. Sin `-race`: el paquete no usa goroutines ni `sync` fuera de las pruebas, y la concurrencia del diario es entre procesos, protegida con un lock (N5) | Parte de la implementación |
| Revisión de código | Antes del push | `/code-review` de Claude Code | Autorizada (decisión de la sesión) |

El cambio visible en la interfaz es solo texto: la ayuda de `drift` en Diagnostics. Según la regla global es mecánico, así que no lleva revisión de experiencia de uso.

## Estado de la revisión y avance

Cierre (2026-09-30):
- commit `12fcf71`, integrado con el PR #77 (merge `f95ea11`);
- `/code-review` devolvió F1 a F10, todos aplicados; F2 queda como límite documentado;
- una re-verificación final dio AC1 a AC6 cumplidos, con 22 mutaciones y F1 y F5 comprobados en el binario real;
- suite local en verde; `hive` recompilado desde `f95ea11`, y `hive doctor` muestra los seis CLIs verificados;
- el delta de la especificación se aplicó a `_support/openspec/specs/versioned-installation/spec.md`.

Hueco menor que queda: F3 en los marcadores de voz no tiene prueba propia; el comportamiento se comprobó con una prueba temporal.

Plan escrito el 2026-09-30 sobre `0292ff4`. Lo revisó un `hive-review-plan` nativo, de solo lectura, en el dominio del gestor. Con una prueba desechable comprobó que la búsqueda de instalaciones antiguas no bloquea antes de la regla nueva, y que AC1 a AC4 fallan en la base.

Aceptados:
- **B1:** el ayudante recibe la acción, y al quitar nunca repone.
- **B2:** `Gone` se comprueba contra la foto actual.
- **N1:** el formato del plan y del diario, y el límite al volver a una versión anterior.
- **N2:** voz: la condición `vc.Gone`, y `off` quita un bloque de voz huérfano.
- **N3:** la prueba de recuperación afirma el resultado.
- **N4:** primero se fija el contrato, y T3 arranca después.
- **N5:** sin `-race`.
- **Menor:** se acepta que el resumen de quitar liste lo ya borrado.
- **AC4:** se completa en T5.

Todas aplican la propuesta del revisor, así que no hubo segunda ronda.
