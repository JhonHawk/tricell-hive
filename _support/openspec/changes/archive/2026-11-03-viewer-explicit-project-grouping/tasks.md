# Tasks

Base: development `2653b0b185572f9207d0b4196969891b21008f7c`. Rama propuesta `feat/viewer-explicit-project-grouping`; implementación en un worktree aislado, plan sin commit. Ejecución paralela por archivos disjuntos: backend (Go T1/T2) y frontend (TS T1/T3) con contrato aprobado; hilo principal integra, documenta, verifica y commitea; revisión independiente del plan y UX por roles requeridos. Los briefs citan esta carpeta. Implementación autorizada por `$flow-build`.

## Preparación del build

Antes de T1, inspeccionar Git y leer git-workflow; crear un worktree limpio desde development actualizado en el lugar nativo que corresponda a este host y a AGENTS (sin carpeta hermana inventada). Mantener el checkout actual y sus archivos ajenos intactos. Llevar sólo esta carpeta de plan al worktree, verificando los hashes y conservando su copia original hasta confirmar el traslado; no hacer commit del plan en el checkout de origen. Registrar ruta/branch/base en este checkpoint. La creación queda autorizada por la instrucción de worktree; se ejecuta cuando empiece flow-build. Al cerrar, eliminar únicamente rama mergeada y worktree limpio según las reglas de preservación.

### T1 — Contrato de proyección de navegación y compatibilidad

- [x] Tipos Go/TS y validación de la proyección opcional y compatibilidad cliente/servidor. Fixtures por lenguaje; integración HTTP real comprueba el contrato compartido.

**Closes:** AC5
**Depends on:** decisiones de contrato del plan aprobadas
**Locations:** tooling/viewer/server.go; tipos de navegación separados del contrato persistido; frontend/src/api.ts y tests
**Execution:** hilo principal
**Test approach:** tdd
**Changes:** petición opt-in, tipos/fixtures, fallback legacy y rechazo aislado de proyección inválida. No agregar campos persistidos a TimelineSession ni relajar validadores existentes.
**Verification:** tests old-client/new-server (sin campo), new-client/old-server (legacy), metadata válida/inválida, IDs extra/duplicados/versiones; `go test ./tooling/viewer` y `pnpm --dir tooling/viewer/frontend test`.

### T2 — Resolver configuración explícita y permitir registro de raíz no Git

- [x] Un único lector estricto y registro directo idempotente; proyección dinámica por raíz original.

**Closes:** AC1, AC3, AC4
**Depends on:** T1 (tipos y fixture)
**Locations:** tooling/viewer/index.go, registry.go, server.go; helper nuevo y tests en tooling/viewer; tooling/cli/view_command_test.go
**Execution:** hilo principal
**Test approach:** tdd
**Changes:** igualdad exacta trim, límites, ausencia/invalidación; caché sólo dentro de petición por raíz; namespaces de grupos y origen. Resolver raíz no Git sólo ante configuración explícita válida en entrada exacta; preservar Git y camino legacy.
**Verification:** fixtures con raíz no Git y dos repos, worktree, grupos con mismo nombre visual sin declaración, nombres case-sensitive, Unicode, enlace/archivo grande/JSON inválido/tipo/control/trailing data; dos registros con added_at estable; refrescar config sin reinicio y comparar hashes de archivos históricos; guardar snapshot de IDs/keys/receipts/blobs antes/después; join entre proyectos sigue devolviendo foreign_project. Tests de registro, store/server/CLI pasan; usar state temporal, nunca el workspace original ni visor habitual.

### T3 — Lista directa agrupada con origen y estados

- [x] Sidebar consume proyección, conserva selección/filtros y muestra origen, fallback y avisos.

**Closes:** AC1, AC2, AC4
**Depends on:** T1 (contrato y fixture); no depende de T2 para trabajar con fixtures
**Locations:** frontend/src/api.ts, controller.ts, shell.tsx, messages/types/tests según ubicación vigente
**Execution:** hilo principal; revisión independiente hive-review-ux en verificación
**Test approach:** tdd
**Changes:** usar componentes instalados sin retocar geometría; preferencias con fallback legacy; conteos de filas visibles; textos ES/EN; error de agrupación aislado. Ningún nuevo menú ni subgrupo.
**Verification:** pruebas de agrupación, filtro por título/ticket, archive/pin, expansión persistida, selección al reagrupar, fallback y retry. `pnpm --dir tooling/viewer/frontend check`, `lint`, `test`. Browser real a 1440×900 y 440×956 light/dark con 2 y 30 recorridos, títulos largos y raíces de mismo basename; teclado y no horizontal overflow; self-check ReUI y hive-review-ux.

### T4 — Documentación y verificación integrada

- [x] Manual, arquitectura y delta reflejan el comportamiento; candidato integrado validado con API real y assets generados.

**Closes:** AC1, AC2, AC3, AC4, AC5
**Depends on:** T2, T3
**Locations:** site/src/content/docs/viewer.md; _support/docs/architecture/deployment-manager.md; delta de esta carpeta; frontend/static generados por build del repo
**Execution:** hilo principal
**Test approach:** check
**Changes:** documentar agrupación versus identidad, lectura explícita opcional sin dependencia Engram, registro directo, cambios de configuración y fallback; revisar README/llms juntos según AGENTS. Mantener sin tocar instrucciones distribuidas.
**Verification:** escenario end-to-end en estado temporal: registrar raíces, crear dos recorridos con claves distintas, publicar contenido, abrir grupo único, mutar sólo configs de fixture, invalidar/restaurar, reiniciar servidor temporal, comprobar enlaces/IDs/bytes y foreign_project. `go test ./...`; frontend check/lint/test/build/verify:dist; checks existentes del sitio/specs y `git diff --check`. Documentar comandos reales descubiertos antes de ejecutar, sin instalar dependencias. No pilot de modelo.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests de tarea | estado temporal, al terminar tarea | tests y assertions de contrato | flow-build |
| Suite completa | candidato final | Go y frontend y checks docs existentes | flow-build |
| UI check | servidor temporal del candidato, desktop/mobile, ambos temas | capturas afectadas, browser walk y hive-review-ux | flow-build/verify |
| Code review | antes de validación o push | nativa Codex `/review` | decisión de revisión |
| Validación humana | entrega interactiva | agrupación real, origen legible, navegar dos recorridos | decisión de entrega |

No se necesita cuenta externa: visor temporal autenticado por su mecanismo local. No imprimir tokens. El usuario valida sólo las capturas/vista del candidato; no editar CAM para el ejercicio. Sólo detener y limpiar procesos/state reproducible creados por esta tarea al cierre; preservar evidencia única y estado habitual.

## Completion evidence

All tasks completed in `71b9c591`; changelog in `52fc0c4f`. Integrated by PR #349 at `effe57fe`. User approved the rendered desktop/mobile view before publication.

- T1 / AC5: independent contract verification, old/new server-client compatibility and strict metadata fallback passed. Go/TS use separate fixtures; authenticated real runtime verifies their shared contract.
- T2 / AC1, AC3, AC4: direct registration, refresh and unchanged persisted bytes passed. Native Codex review and independent verification found one inconsistent historical fallback-name defect; a failing-first HTTP regression and request-local name selection fixed it. Focused second review passed.
- T3 / AC1, AC2, AC4: 706 frontend tests plus 3 script tests, TS/Biome, desktop/mobile browser walks, keyboard, selection and filters passed. Captures cover 2 and 30 works, degraded config and simulated CSS light mode (the application natively uses dark). User accepted.
- T4 / all criteria: Go vet/full Go suite, 23 Python checks, deterministic bundle rebuild/verification, Astro check/build passed. Restart of isolated server preserved all stored bytes. Maintainer viewer refreshed at the same port/state/TTL and its two affected real works form one group. No model pilots or race suite (no concurrency changes).

Review: native Codex on implementation `71b9c591`, then focused fix review. Independent UX found no functional blocker or measured navigation WCAG violation; preexisting three-line titles and tooltip landmark best-practice advisory remain P3 observations. Inconclusive contrast nodes were not claimed as compliance.

Local setup: dependency symlinks were removed after build because the offline-package test correctly rejects linked sources. No packages installed. Seven screenshots were copied to the viewer and authenticated bytes/hash verified before cleanup. The change preserves identities, history and collaboration authority.
