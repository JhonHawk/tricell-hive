# Tasks

Base: `development` en `94ed2264`, rama `fix/archive-order-within-day`. Todas las tareas corren en el hilo principal: son un archivo de test, dos frases de guía y un renombre que se verifican juntos.

### T1 — Nombre del archivo en `flow-close` y `artifact-layout`, y CHANGELOG

- [x] `flow-close` paso 4 mueve la carpeta a `archive/YYYY-MM-DD-HHMM-<change-id>/`, con la fecha y la hora en UTC del archivo (`date -u +%Y-%m-%d-%H%M`, D8 A). La frase debe ser corta, porque la skill tiene 15 881 bytes y el límite es de 16 000. `artifact-layout.md:34` deja de repetir la regla y apunta a `flow-close` paso 4. La entrada del `CHANGELOG` sobre el test de deltas (`CHANGELOG.md:63–66`) se reescribe en su sitio, sin añadir otra que la contradiga.

**Closes:** AC1

**Depends on:** ninguna

**Locations:** `content/skills/flow-close/SKILL.md`, `content/skills/workspace-conventions/references/artifact-layout.md`, `CHANGELOG.md`

**Changes:** Solo el nombre de destino y su puntero. No cambian los demás pasos del cierre.

**Verification:**
- `rg -n 'YYYY-MM-DD-HHMM-<change-id>'` está ausente en `git show 94ed2264:content/skills/flow-close/SKILL.md` y presente después.
- `artifact-layout.md` ya no dice «which use the closing date».
- `go test ./tests/content/...` pasa, incluido el límite de tamaño.

### T2 — Orden por hora y duplicados en `spec_deltas_test.go`

- [x] **Orden.** Dos carpetas del mismo día con horas distintas se ordenan de forma estricta. Si a alguna le falta la hora, o las dos tienen la misma, se conserva la tolerancia (`SameDay`).
- [x] **Duplicados.** Se detectan en una pasada aparte y sus bloques se excluyen de las comparaciones y de `SameDay`. Se reportan solo cuando su carpeta tiene el delta más reciente del requisito. Cualquier carpeta que ordene después los supera, también una del mismo día sin hora.
- [x] Se resuelven H29–H36.

**Closes:** AC2, AC4, AC5

**Depends on:** ninguna

**Locations:** `tests/content/spec_deltas_test.go`

**Test approach:** tdd. Los casos de AC2, AC4 y AC5 deben fallar en `94ed2264`:

- **AC2:** `2026-01-01-1403-a` y `2026-01-01-2003-b` modifican Alpha y la especificación coincide con `a`. Esperado: un fallo, `DIFFERS`, que nombra `b`.
- **AC4:** `2026-01-01-a` lista Alpha dos veces, `2026-02-01-c` lo modifica y la especificación coincide con `c`. Esperado: ningún fallo. En la base da uno, el duplicado.
- **AC5:** `2026-01-01-a` lista Alpha dos veces (el primer bloque es «X»), `2026-01-01-b`, sin hora, lo modifica a «Z» y la especificación dice «X». Esperado: un fallo, `DIFFERS`, que nombra `b`. En la base el bloque «X» lo oculta.
- **Cobertura (H30, no tiene que fallar en la base):** el caso de AC4 con la especificación contraria a `c`. Esperado: un solo fallo, `DIFFERS` de `c`, sin el duplicado.
- **Se conservan:**
  - la regresión del mismo día sin hora (`TestSpecDeltaSameDateAcceptsEitherDelta`);
  - un duplicado solo en su carpeta, que se reporta con sus dos operaciones en cualquier orden de archivo;
  - un duplicado en la carpeta más reciente con un cambio activo sobre el requisito, que se sigue reportando.

**Changes:**
- Se reescriben las aserciones de `TestSpecDeltaSameFolderDuplicateFails` sobre carpetas posteriores (hoy esperan el duplicado; con AC4 ya no) y su comentario.
- Se actualizan los comentarios de `latestDelta`, `collectLatestDeltas`, `checkSpecDeltas` y `TestSpecDeltaSameDateAcceptsEitherDelta`.
- No cambian `parseDeltaBlocks`, `parseSpecBlocks` ni `activeChangeRequirements`.

**Verification:** Los casos de AC2, AC4 y AC5 fallan en `94ed2264` y pasan después. `go test -count=1 ./tests/content/...` pasa.

### T3 — Renombrar el par del 2026-11-04

- [x] Dos `git mv`: `2026-11-04-viewer-session-lifecycle-and-autoarchive` y `2026-11-04-viewer-local-actions-cli` pasan a `2026-11-04-2003-viewer-session-lifecycle-and-autoarchive` y `2026-11-05-0203-viewer-local-actions-cli`, con la hora UTC de sus commits de archivo, `bdb2e8f4` (20:03Z) y `bb61f71f` (02:03Z del 11-05). Ninguna referencia del repositorio apunta a esas carpetas, así que no hay que corregir nada más.

**Closes:** AC3

**Depends on:** T2 (su verificación usa el orden estricto)

**Locations:** `_support/openspec/changes/archive/`

**Changes:** Solo los nombres de las dos carpetas. Su contenido no cambia.

**Verification:**
- `ls _support/openspec/changes/archive` muestra los dos nombres nuevos.
- `go test -count=1 ./tests/content/...` pasa sobre el repositorio real.
- Comprobación sin tocar el repositorio: en un worktree temporal con las carpetas renombradas, se cambia el bloque «Local viewer management actor» de `specs/walkthrough-viewer/spec.md` por el delta de las 14:03. `go test ./tests/content/ -run TestSpecDeltasApplied` da exactamente un `DIFFERS` que nombra `2026-11-05-0203-viewer-local-actions-cli`. Después se borra ese worktree.

## Verification and human review

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de contenido | Al cerrar cada tarea, local | `go test -count=1 ./tests/content/...` | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go test ./...` | Cubierto por la implementación |
| Revisión del comportamiento entero | Antes del push | Lugares revisados (incluido `artifact-layout.md:34`) y tamaño de `content/guidance/global.md` (sin cambio) en el PR | Cubierto por la implementación |
| Code review | Candidato final, antes del push | Claude Code `/code-review` (D5 A) | D5 A |

La entrega es automática (D4 A), sin parada de validación.

## Review status and progress

2026-11-04 — Plan escrito y revisado. La revisión del plan encontró 9 huecos:
- T3 no corrige ninguna ruta;
- el caso de H30 queda escrito y se exime de fallar en la base;
- se define qué supera a un duplicado y qué aserciones cambian;
- `artifact-layout.md:34`;
- la entrada del `CHANGELOG` se reescribe;
- una comprobación de AC3 que demuestra la exigencia;
- dos carpetas con la misma hora se toleran;
- margen de tamaño y comentarios obsoletos.

La zona horaria se resolvió con D8 A (UTC).
- T2 hecho en `998316d2`: los casos de AC2, AC4 y AC5 fallaron en la base (`94ed2264`) y pasan; `go test -count=1 ./tests/content/...` pasa.
- T3 hecho en `b2356aac`: dos `git mv`. En un worktree temporal, la especificación con el delta antiguo dio un único `DIFFERS` que nombra `2026-11-05-0203-viewer-local-actions-cli`.
- T1 hecho en el commit de guía: `flow-close` ocupa 15 933 bytes, por debajo del límite; `artifact-layout.md` apunta a `flow-close`; el `CHANGELOG` está reescrito en su sitio.

## Checkpoint

- Etapa: verify.
- Rama: `fix/archive-order-within-day` en el worktree `.claude/worktrees/archive-order-within-day`.
- Visor: work `05a290ef…`.
- Primera revisión `/code-review` (high) sobre `1f128ed3`: 9 hallazgos.
  - Corregidos en `43e9866a`: transición de fecha local a UTC, alternativas con todo el historial, texto del CHANGELOG, `artifact-layout.md:35` y el bucle.
  - Van al cierre: `doctor` no exige la hora; un duplicado superado por una carpeta sin orden; un ID de 4 dígitos que se lee como hora; dos archivos en el mismo minuto.
- Segunda revisión, limitada a `1f128ed3..43e9866a`: 9 hallazgos.
  - Bloquea uno, arreglado en local con `Unreviewed:`: se aceptaba un delta superado por una carpeta posterior conocida. Ahora se descartan al recolectar y hay un caso negativo, que falla en `43e9866a`. Con él se corrigió el texto del diseño.
  - Van al cierre: la ventana de un día ignora hora y dirección; el duplicado de una carpeta anterior sin orden no se reporta; el escenario del delta y el texto del proposal para la ventana entre días; los nombres de los tests.
- Integrado con el PR #371 (`8b4d1eee`) y desplegado.

2026-11-04 — Cierre: el delta se fusionó en `specs/workflow-lifecycle/spec.md` y la carpeta se movió al archivo con su nombre UTC, el primero con el formato nuevo. Los hallazgos H49–H57 quedan listados en el reporte de cierre.
