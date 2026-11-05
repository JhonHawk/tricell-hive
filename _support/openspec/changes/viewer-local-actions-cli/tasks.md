# Tareas

Base: `development` 192697cb, rama `feat/viewer-local-actions-cli`. T1 fija el contrato en el servidor. Después T2 (CLI, `hive-build-backend`) y T3 (frontend, `hive-build-frontend`) consumen ese contrato congelado, tocan archivos disjuntos y corren en paralelo; la prueba de integración de T6 cruza los dos. T4 y T5 son redacción en el hilo principal.

### T1 — Origen del actor y capacidad en el servidor

- [x] El servidor acepta `origin` en el actor `local_viewer`, lo guarda donde dice `design.md`, anuncia `local_actions_cli` en `/api/health` y publica el esquema.

**Closes:** AC4 (servidor), AC7 (capacidad)

**Depends on:** ninguna

**Locations:** `tooling/viewer` (contrato v3, `timeline_protocol.go`, dominio, almacén, mantenimiento, borrado, `server.go` y sus tests), `tooling/viewer/testdata/timeline_contract/`

**Execution:** hijo `hive-build-backend`, con `design.md` como contrato congelado

**Test approach:** tdd (contrato persistido y permisos)

**Changes:** sin `origin` vale `browser`; un valor distinto se rechaza; las reglas de quién hace cada acción no cambian; la revisión del contrato no sube.

**Verification:** `go test ./tooling/viewer/...` con casos que fallan antes: `origin: cli` aceptado y guardado en evento, recibo y auditoría de borrado; sin `origin` guardado como `browser`; valor desconocido rechazado sin cambios; `/api/health` con la capacidad; esquema con `origin`.

### T2 — Comandos de CLI

- [x] `hive view work rename|archive|restore|pin|unpin|recover|delete` funcionan como en `design.md`.

**Closes:** AC1, AC2, AC3, AC7, AC6 (contrato embebido y errores de `session update`/`delete`)

**Depends on:** T1

**Locations:** `tooling/cli/view_work*.go`, `tooling/cli/view_work_contract.go`, `tooling/cli/view_timeline.go` (solo los errores que apuntan a los comandos nuevos) y sus tests

**Execution:** hijo `hive-build-backend`, después de T1

**Test approach:** tdd

**Verification:** tests contra un visor de prueba: cada acción cambia el estado y deja el origen `cli`; `recover` vuelve conductora la solicitud indicada y esa sesión cierra el trabajo; `delete` sin `--confirm` igual al ID no envía nada; un visor sin la capacidad no recibe nada; una respuesta incierta relee y reporta. Prueba real con el binario candidato sobre un visor en un estado temporal, nunca sobre el visor del usuario.

### T3 — Frontend: actor con origen

- [x] El validador y el tipo del actor aceptan `origin` y la UI muestra el origen donde ya muestra el actor.

**Closes:** AC4 (UI)

**Depends on:** T1

**Locations:** `tooling/viewer/frontend/src/timeline-contract-v3.ts` y sus tests, el componente que muestra el actor, el bundle versionado en `tooling/viewer/static/app/`

**Execution:** hijo `hive-build-frontend`, en paralelo con T2

**Test approach:** tdd

**Verification:** `pnpm test`, `pnpm check`, `pnpm lint` y `pnpm build` pasan; el bundle versionado coincide con el build; un caso compartido con `origin: cli` se lee sin error.

### T4 — Regla de la guía

- [x] `viewer-publication.md` da la regla de D13 A y D14 A y corrige las frases que la contradicen.

**Closes:** AC5

**Depends on:** ninguna

**Locations:** `content/skills/flow-walkthrough/references/viewer-publication.md` (incluido el puntero para cerrar un trabajo cuya sesión conductora terminó; `flow-close` no tiene bytes libres)

**Execution:** hilo principal

**Test approach:** check

**Verification:** `go test ./tests/...` (tamaños); búsqueda de cada lugar que nombra `local_viewer` o la recuperación, listada en el PR.

### T5 — Spec, documentación, sitio y CHANGELOG

- [x] El delta de spec, `README.md`, `llms.txt`, `deployment-manager.md` (comandos y vuelta atrás), `site/…/viewer.md` y el CHANGELOG describen los comandos.

**Closes:** AC6

**Depends on:** T2 (forma final de los comandos)

**Execution:** hilo principal

**Test approach:** check

**Verification:** build del sitio pasa.

### T6 — Verificación del conjunto

- [ ] Suite completa, integración CLI–UI, verificación independiente y revisión.

**Closes:** AC1–AC7

**Depends on:** T1–T5

**Execution:** hilo principal; un hijo `hive-verify-task` verifica T1 y T2 por tocar permisos y un contrato persistido

**Verification:** `go test ./...`, `go vet ./...` y los checks del frontend pasan; en un visor de prueba, una acción hecha con la CLI se abre en la UI con su origen.

## Verificación y revisión humana

| Chequeo | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar la tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go test ./...`, `go vet ./...`, checks del frontend, build del sitio | Cubierto por la implementación |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` (D16 A) | La respuesta de revisión |

Entrega automática (D15 B): push, PR, merge y despliegue. Al integrar, el binario se recompila y el visor local se reinicia conservando puerto, estado y TTL, como `AGENTS.md` autoriza al mantenedor; el token cambia y el reporte ofrece «abre el visor».

## Estado de la revisión y avance

2026-11-04 — plan escrito; revisión independiente: 14 hallazgos aplicados; D17 A, D18 A y D19 A resueltos por el usuario.
- Checkpoint: plan listo; visor work `8e54541d9af61f82d08d5ad73e90cdb09dc928d5e89f66a75c262b61c9bf9ce4`, snapshot `revision-c9afb937c9b71a6dbb28365b16208bbd`. Siguiente: `/flow-build` (T1, luego T2 y T3 en paralelo, T4 y T5 en el hilo principal).
- 2026-11-04 · T1 7fd390da, T4 127ef419, T3 a3029783, T2 4d1b42bc. Prueba real con el binario candidato sobre un visor aislado (puerto 4597, estado temporal, detenido y borrado): las cinco acciones, recover con cierre posterior, delete sin y con --confirm, auditoría con origin cli.
