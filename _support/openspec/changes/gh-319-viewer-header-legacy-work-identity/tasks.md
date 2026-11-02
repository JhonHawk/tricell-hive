# Tareas

**Base:** `tricell-hive` en `c72549e4` (`development`), rama `feat/gh-319-viewer-cleanup`.

**Orden y ejecución:**
- T1 va primero y en el hilo principal, porque fija el contrato que consumen T2 y T3, incluido el retiro de `mode:"legacy"` en Go, TypeScript y fixtures.
- T4 (guidance, hilo principal) no depende de nada y avanza mientras corren los hijos.
- Después de T1, T2 (Go, hijo `hive-build-backend`) y T3 (frontend, hijo `hive-build-frontend`) corren en paralelo, cada uno en su propio worktree; el hilo principal integra sus cambios. No comparten archivos: `testdata/`, `timeline-contract.ts` y `timeline_contract*.go` quedan cerrados por T1.
- T5 (spec y documentación) va al final.

### T1 — Identidad de trabajo en `create_session`

- [x] `work_key` opcional en la petición, la derivación del ID, el registro guardado, la salida de sesión, los validadores Go/TS, el chequeo de la CLI y el contrato offline; y el contrato sin `mode:"legacy"` en Go y TypeScript, como fija `design.md`.

**Closes:** AC7, AC5 (parte del índice)

**Depends on:** ninguna

**Locations:**
- `tooling/viewer/timeline_contract.go`, `timeline_store.go`, `timeline_protocol.go` (esquema y ejemplo de `create_session`), `timeline_contract_test.go`;
- `tooling/cli/view_timeline.go` (chequeo de `work_key` en la respuesta);
- `frontend/src/timeline-contract.ts`, `frontend/src/tests/timeline-contract.test.ts`, `frontend/dev/timeline-preview.tsx`;
- `testdata/timeline_contract/`: `cases.json`, `legacy-index-entry.json` y las fixtures de sesión con `work_key` nullable (`session.json`, `v3-session.json`, `stage-session.json`, `catalog-session.json`, `operation-*.json`);
- los tests de store.

**Execution:** hilo principal

**Test approach:** tdd

**Changes:**
- Casos nuevos en `cases.json`:
  - `create_session` con y sin `work_key` válidos;
  - `work_key` con espacio o vacío, inválido;
  - sesión con `work_key: null` y con valor.
- Los tres casos legacy de `cases.json` y `legacy-index-entry.json` esperan `unsupported_mode`; se borran `TimelineLegacyIndexEntry`, `TimelineLegacySession`, la rama reflect `legacy` y los `case "legacy"` en Go y TypeScript.
- Tests de store para:
  - dos claves dan dos IDs;
  - la misma clave da `created:false`;
  - sin clave, con dos `operation_id`, da dos IDs;
  - el replay idéntico da el mismo ID;
  - un registro guardado con la derivación antigua carga.
- No cambia el índice ni ninguna otra operación.

**Verification:** `go test ./tooling/viewer ./tooling/cli` y la cadena de `pnpm` pasan, con los casos nuevos fallando antes del cambio.

### T2 — Runtime Go y CLI sin legacy, guardas ni `--ttl`

- [x] Se borra el código, los assets embebidos y los tests solo legacy de `tooling/viewer` y `tooling/cli`, las tres guardas y `--ttl`.

**Closes:** AC3, AC5, AC6

**Depends on:** T1 (contrato cerrado)

**Locations:** `tooling/viewer/*.go` salvo `timeline_contract*.go` y `testdata/`, `tooling/viewer/static/walkthrough.*`, `tooling/viewer/doc.go`, `tooling/viewer/timeline_protocol.go` (solo `cutoff_required` y su invariante), `tooling/cli/view_command.go`, `tooling/cli/view_source_tracking*.go`, los tests afectados

**Execution:** hijo `hive-build-backend`

**Test approach:** characterization

**Changes:**
- Lo que lista `design.md` en «Retiro del legacy» para Go y la CLI.
- Tests nuevos:
  - `/api/sessions/x` responde 404;
  - un estado con `viewer/index.json`, uno sin marcador y uno con `storage_version` 2 fallan con `invalid_storage` sin escribir;
  - `--ttl` y `status --dir` fallan como flags desconocidos.
- Se conservan `registry.go`, `ResolveRoot`, `/files/`, `hardlink.go`, `lock_unix.go`, `bundle.go` y los símbolos que `design.md` lista en «Se conserva o se mueve».

**Verification:** `gofmt -l tooling/` vacío; `go vet ./tooling/...` y `go test ./tooling/...` pasan; el `rg` de AC3 no devuelve nada.

### T3 — Frontend: header compacto, contador de validaciones y sin legacy

- [x] El header sigue el boceto de `design.md`, el canvas pierde el resumen, el contador de validaciones entra en la barra, se borra el código frontend legacy y se regenera el bundle.

**Closes:** AC1, AC2, AC3

**Depends on:** T1

**Locations:**
- `tooling/viewer/frontend/src/{timeline-header,timeline-content,shell,application,api,controller,messages}.tsx?` (no `timeline-contract.ts`, que cierra T1);
- `document.tsx`, `deliveries.tsx`, `sessions.ts` (se borran);
- `vite.config.ts`, `scripts/preview.mjs`, `src/tests/*`, `dev/timeline-preview.tsx`;
- `tooling/viewer/static/app/`.

**Execution:** hijo `hive-build-frontend`

**Test approach:** tdd

**Changes:**
- Tests de render para estos cambios: sin `project_name` ni fecha en el header, el resumen en el header y no en el canvas, el contador con su nombre accesible.
- Se aplican las reglas ReUI del `AGENTS.md`: el plan de UI antes del código y el self-check al terminar.

**Verification:** `pnpm run check && pnpm run lint && pnpm run test && pnpm run build && pnpm run verify:dist` pasan; `go test ./tooling/viewer` pasa con el bundle nuevo; primera vista renderizada mostrada al usuario contra el boceto.

### T4 — Guidance: apertura con pregunta, `work_key` y sin fallback

- [x] `flow-walkthrough` y sus referencias, `flow-close` y `artifact-layout` quedan sin fallback y con las reglas de apertura y de `work_key`; se borran el manifiesto, el renderer, los assets y sus tests Python.

**Closes:** AC4, AC8, AC9

**Depends on:** ninguna

**Locations:**
- `content/skills/flow-walkthrough/` (SKILL.md, `references/viewer-publication.md`; se borran `references/walkthrough-manifest.md`, `scripts/render_walkthrough.py` y `assets/walkthrough.*`);
- `content/skills/flow-close/SKILL.md` (la cláusula completa del walkthrough legacy);
- `viewer-publication.md` «Register once»: cuándo registrar las tareas (AC9), con un puntero de una línea en el párrafo de la lista nativa de tareas de `content/skills/flow-build/SKILL.md`; el commit nombra las sesiones `32cffc1d`, `98476eca` y `8d09a5eb` (Claude Code, 2026-11-02) y la falla observada, como pide `AGENTS.md`;
- punteros de una línea a la regla de apertura en `content/skills/flow-plan/references/plan-handoff.md`, `content/skills/flow-research/SKILL.md` (paso de reporte), `content/skills/flow-build/SKILL.md` (paso 4) y `content/skills/flow-verify/SKILL.md` (paso 2);
- `content/skills/workspace-conventions/references/artifact-layout.md`;
- `tests/skills/walkthrough_*_test.py` y `tests/fixtures/walkthrough/` (se borran).

**Execution:** hilo principal

**Test approach:** check

**Changes:**
- Como fija `design.md` en «Guidance».
- Antes de editar, revisar toda la conducta afectada, como pide `AGENTS.md`: global, skills, roles, `site/`, `_support/docs/architecture/` y los detectores de `tests/pilot`. Listar en el PR los lugares revisados.
- `content/guidance/global.md` no cambia.

**Verification:** `go test ./tests/...` pasa (tamaño de skills y referencias incluido); el `rg` de AC4 sobre `content/` y `tests/` no devuelve nada; la regla nueva aparece completa una sola vez, en `viewer-publication.md`, y los cuatro punteros la nombran sin repetir sus pasos; `flow-walkthrough/SKILL.md` ya no dice que vuelve sin agregar una pregunta.

### T5 — Spec y documentación

- [x] El delta de `walkthrough-viewer` se aplica a la spec en el cierre; README, `llms.txt` si cambia algo que declara, `site/src/content/docs/viewer.md` y `_support/docs/architecture/deployment-manager.md` quedan sin legacy ni `--ttl`, con `work_key` y la regla de apertura descritas.

**Closes:** AC4

**Depends on:** T2, T3, T4

**Locations:** `README.md`, `llms.txt`, `site/src/content/docs/viewer.md`, `site/src/content/docs/index.mdx` (`:98`), `_support/docs/architecture/deployment-manager.md`, `tooling/viewer/frontend/README.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** solo texto; las páginas describen la conducta y enlazan la skill, sin copiar sus reglas.

**Verification:** el `rg` de AC4 completo no devuelve nada; el build del sitio pasa si el repositorio tiene su comando.

## Verificación y revisión humana

| Comprobación | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, en local | El comando de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `gofmt -l tooling/ tests/ && go vet ./tooling/... ./tests/... && go test ./tooling/... ./tests/...`, la cadena de `pnpm` y `python3 -m unittest` de `tests/skills` si queda alguno | Cubierta por la implementación |
| Conservación | Candidato final, en local | Un visor de prueba con el binario del candidato sobre una copia del estado real, sin abrir el navegador, carga las 5+ sesiones con sus IDs, 0 errores, y `create_session` desde una sesión nativa existente con `work_key` nuevo crea una sesión distinta | Cubierta por la implementación |
| UI | Primera vista tras T3 y candidato final, en local | Header en escritorio (≥1280) y móvil (390), tema claro y oscuro, con resumen largo y título largo; un hijo `hive-review-ux` con esas capturas | Cubierta por la implementación |
| Revisión de código | Candidato final, antes del push o de la parada de validación | Según la respuesta de revisión | La respuesta de revisión |

**Recorrido humano:**
- Servir el candidato con `hive view start` en un puerto libre, sobre una copia del estado.
- Abrir la sesión de una obra con resumen largo.
- Comprobar que el header muestra título y editar en la primera fila, y resumen con «ver más» más la barra a la derecha en la segunda, sin fecha ni proyecto.

## Estado de revisión y avance

2026-11-02 — Plan escrito; ninguna tarea empezada.
2026-11-02 — T1 hecha · `49a0d10d` · Go y pnpm en verde; los registros guardados sin `work_key` se leen con un decodificador tolerante (`decodeTimelineStoredJSON`) y la carga comprueba que la clave efectiva coincida; `hive-verify-task` en curso.
2026-11-02 — T4 hecha · `d54fb51d` · `go test ./tests/...` en verde; el renderer, los assets y los fixtures legacy se borran al integrar T2 y T3, que quitan sus últimos consumidores.
2026-11-02 — T5 (documentación) · `cfb46d8a` · README, sitio, `deployment-manager.md`; el delta de spec se aplica en el cierre.
2026-11-02 — Checkpoint: build, paso 2. T2 (hijo backend) y T3 (hijo frontend) corren en worktrees propios desde `49a0d10d`. Siguiente: integrar T2 y T3, borrar `content/skills/flow-walkthrough/assets/`, primera vista del header al usuario, suite completa.
2026-11-02 — T3 hecha · `611a6b10` · 700 tests de pnpm; además se corrigió el corte a 2 líneas del resumen (raíz flex del markdown) y el `h1` de los estados sin sesión; primera vista aceptada por el usuario en un visor de prueba con el binario del candidato (D20 A).
2026-11-02 — T2 hecha · `37df77ed` · suite Go completa en verde; también se borró `PartialDeleteError`, que quedó sin uso, y los assets del skill.
2026-11-02 — `d97349b3`: la CLI muestra el código de error del servidor; `viewer-publication` dice cómo asociar segmentos (hallado al publicar las tareas). `8ce07f90`: H16, el test de la derivación vieja ahora fija su valor.
2026-11-02 — Checkpoint: build terminado (T1–T5); sin workflows de CI. Siguiente: `flow-verify` (conservación con el binario final, recorrido de UI con `hive-review-ux`, `/code-review`), parada de validación.
