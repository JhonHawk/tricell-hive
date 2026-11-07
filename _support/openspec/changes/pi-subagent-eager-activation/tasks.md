# Tasks

Base `development` en `8956c9a3`, rama `feat/pi-subagent-eager-activation`.

- T1 corre en el hilo principal: son funciones puras de la integración.
- T2 corre en un hijo `hive-build-backend`: edita más de tres archivos del manager. Su brief lleva esta carpeta, el resultado de T1 y el patrón de `codex_wait_floor.go`.
- T3 y T4 corren en el hilo principal, después de T2.
- T5, que es documentación, corre en el hilo principal al final.

### T1 — Clasificar y editar la clave en la integración de Pi

- [x] `integrations/pi` expone la ruta relativa del archivo, la clave y el valor, más estas funciones:
  - clasificar los bytes de `config.json` en: ausente, objeto sin la clave, clave presente con su valor, o no editable (no es un objeto, tiene claves duplicadas o es JSON inválido);
  - insertar `toolActivation` conservando las demás claves, sus valores y su orden;
  - quitarla sin tocar el resto;
  - comprobar la seguridad del archivo y su carpeta igual que `integrations/codex/wait_floor_target.go`: enlace simbólico, archivo no regular, más de 1 MiB, carpeta no escribible.

**Closes:** AC2, AC4

**Depends on:** ninguna

**Locations:** `integrations/pi/subagent_config.go`, `integrations/pi/subagent_config_test.go`

**Execution:** hilo principal

**Test approach:** tdd

**Changes:** Funciones sobre bytes y sobre metadatos de archivos. No escribe nada.

**Verification:** `go test ./integrations/pi/...` pasa. La tabla de casos cubre:
- archivo ausente, `{}`, un objeto con otras claves (con su orden verificado después de insertar y de quitar);
- la clave en `auto`, en `eager` y en `null`;
- claves duplicadas, JSON inválido y un arreglo;
- un enlace simbólico, un archivo de más de 1 MiB y una carpeta de solo lectura.

### T2 — Paso del manager: plan, apply, recuperación, eliminación, estado y validaciones

- [x] Paso nuevo en el plan y registro nuevo en el estado (el esquema no cambia, con `omitempty`), con su accesor `StoredPiSubagentConfig`. Se planifica igual que `planPiPackage`: install, instalación desde paquete y actualización desde el código fuente. Cubre todos los estados de AC1, AC3 y AC4, con el caso del otro Pi home.
- [x] Apply:
  - revalida el fingerprint;
  - toma como estado previo lo que deja la migración del instalador anterior cuando ella toca el mismo archivo;
  - crea el backup;
  - escribe conservando el modo, o con 0600 y carpetas en 0700 si el archivo es nuevo;
  - registra en el estado y en el journal;
  - se recupera por clave.
- [x] `hive plan remove` sobre Pi aplica AC5. Una eliminación interrumpida se recupera según AC6.
- [x] `PlanUnchanged` conoce el paso (AC7). `validatePlan`, `update_coordinator.go` y `update_providers.go` lo rechazan en una actualización de release (AC9).
- [x] Los tests existentes con un Pi home sintético se ajustan al paso nuevo sin perder lo que verificaban.

**Closes:** AC1, AC2, AC3, AC4, AC5, AC6, AC7, AC8, AC9

**Depends on:** T1

**Locations:**
- `tooling/management/pi_subagent_config.go` (nuevo) y `types.go`;
- `plan.go`, `apply.go`, `migration.go`;
- `update_providers.go` y `update_coordinator.go`;
- los tests nuevos `pi_subagent_config_{plan,apply,remove}_test.go` y los existentes que use un Pi home sintético (`install_test.go`, `migration_test.go`, `release_update_test.go`, `shared_test.go` y los que correspondan).

**Execution:** hijo `hive-build-backend`

**Test approach:** tdd

**Changes:** Sigue el patrón de `codex_wait_floor.go` sin pedir consentimiento aparte, igual que la instalación de pi-subagents. No toca `settings.json` ni la instalación del paquete. Nunca usa el `~/.pi` real.

**Verification:** `go test ./tooling/management/...` pasa, con un caso por AC (AC1 a AC9). Incluye:
- un fallo inyectado después de la escritura, con su recuperación;
- una eliminación interrumpida;
- una migración anterior con `forceTopLevelAsync` y sin la clave;
- `PlanUnchanged` en falso cuando solo falta la clave;
- el rechazo en una actualización de release.

### T3 — Resúmenes de instalación y desinstalación

- [x] `showInstallSummary` muestra el paso: ruta, clave, valor, «ya aplicado», «puesto por el usuario» o «paso manual», y el aviso de reiniciar Pi.
- [x] El resumen de desinstalación y las notas de apply informan lo que se quitó y lo que se dejó.

**Closes:** AC1, AC4, AC5

**Depends on:** T2

**Locations:** `tooling/cli/install.go`, `tooling/cli/tui_hosts.go`, `tooling/management/apply.go` (notas) y sus tests

**Execution:** hilo principal

**Test approach:** check, extendiendo los tests que ya cubren esos resúmenes

**Changes:** Solo texto de los resúmenes. Nunca muestra otro contenido del archivo.

**Verification:** `go test ./tooling/cli/...` pasa, con un caso por cada estado del resumen.

### T4 — Fila de `hive doctor` y de la TUI

- [x] Con Pi registrado, `hive doctor` muestra la fila «Pi subagent activation» con sus cuatro estados (AC10), y en «no fijado» indica `hive install`.
- [x] Los textos de `doctor_integrations.go` que dicen que Hive no lee la configuración de Pi pasan a nombrar esta única excepción.
- [x] La vista de integraciones de la TUI reserva las filas necesarias para mostrar todas cuando Pi y Codex están registrados.

**Closes:** AC10

**Depends on:** T2

**Locations:** `tooling/cli/doctor_integrations.go`, `tooling/cli/tui_integrations_view.go` y sus tests

**Execution:** hilo principal

**Test approach:** check, extendiendo los tests que ya cubren las filas de integraciones y la vista

**Changes:** Solo lee ese archivo con las comprobaciones de seguridad de T1, y el estado mediante `StoredPiSubagentConfig`.

**Verification:** `go test ./tooling/cli/...` pasa, con los cuatro estados y con la vista de Pi y Codex registrados.

### T5 — Spec y documentación

- [x] El delta de `versioned-installation` refleja lo construido: el requisito agregado y el de diagnósticos modificado.
- [x] `README.md:191`, `llms.txt:62`, `_support/docs/architecture/tool-dependencies.md:20` y `deployment-manager.md:111` dejan de afirmar que el piso de Codex es lo único que Hive escribe en la configuración de un host.
- [x] La sección de Pi de `agent-delivery.md` describe el paso y su causa: pi-subagents 0.83 en modo diferido.
- [x] `site/src/content/docs/host-limits.md` cambia solo si afirma algo que el cambio contradice.

**Closes:** ninguno; documenta AC1 a AC10

**Depends on:** T2, T3, T4

**Locations:** el delta de esta carpeta, `README.md`, `llms.txt`, `_support/docs/architecture/{tool-dependencies,deployment-manager,agent-delivery}.md` y, si aplica, `site/src/content/docs/host-limits.md`

**Execution:** hilo principal

**Changes:** Documentación del comportamiento. No cambia la guía distribuida en `content/`.

**Verification:** `go test ./tests/...` pasa, incluida la verificación de deltas; `check_change.py` sobre esta carpeta sale con 0.

## Verificación y revisión humana

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar cada tarea, en local | El comando de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `go test ./...` y `go vet ./...` | Cubierta por la implementación |
| Plan y apply en un home sintético | Candidato final, en local | `hive install`, la actualización desde el código fuente y `hive plan remove` con `--home` temporal: el archivo, el estado, el resumen y la desinstalación | Cubierta por la implementación |
| Revisión de código | Candidato final, antes del push | `Native: Claude Code /code-review` | La respuesta de revisión |

Sin validación humana: la entrega es `automatic` y no hay pantalla.

## Estado de la revisión y progreso

2026-11-06 — Plan escrito y revisado (R1–R19 aplicados); alcance confirmado en D6 A.
2026-11-06 — T1 hecha: `go test ./integrations/pi/` y `go vet` pasan. Siguiente: T2 en un hijo `hive-build-backend`.
2026-11-06 — T2 hecha por `hive-build-backend`: `go test ./tooling/management/... ./integrations/... ./tooling/cli/...` y `go vet` pasan. El hijo escribió el código antes que las pruebas y lo compensó con 6 mutaciones; 5 las detectó la prueba prevista, y la sexta (`update_providers.go`) es una comprobación redundante. Brecha en AC4: con la instalación, un `config.json` inválido, con claves duplicadas, que no es un objeto, que es un enlace simbólico o que no es un archivo regular hace fallar todo `BuildPlan` en el escaneo legacy, previo, en vez de dar un paso manual; no se escribe nada. El caso de más de 1 MiB y el de carpeta no escribible sí llegan al paso. La eliminación interrumpida se revierte hacia atrás (AC6), salvo un backup ya borrado.
2026-11-06 — T3 `079e5e74`, T4 `2fe6fd1e` (columna de nombre a 24 y vista con 9 filas), T5 `f32de7ca`. Las notas de eliminación ya salen en el resultado de Apply; el helper de pruebas se ajustó en `16cbc85d`. Suite local en `f32de7ca`+`16cbc85d`: `go vet ./...` pasa y `go test ./...` pasa. El único workflow (`master-source-branch.yml`) no corre pruebas. Siguiente: verify con `/code-review`.
