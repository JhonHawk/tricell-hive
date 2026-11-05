# Tareas

Base: `development` en `2765624b` (origin, 2026-11-04). Rama: la actual, `feat/gh-311-minimal-research-entry` (HEAD `ee1ebf07`, PR #352), por decisión del usuario. Los commits de este cambio llevan su propio mensaje; tocan archivos que gh-311 también cambió (`global.md` y los skills de flujo), así que se editan sobre el HEAD actual.

En el hilo principal corren T1, que fija el texto de la regla, y después T3 y T4. T2 va a un hijo `hive-write-docs` en cuanto termina T1, porque edita 9 archivos, y corre al mismo tiempo que T3 y T4 sin tocar sus archivos.

### T1 — Regla en la guía global

- [x] `content/guidance/global.md` fija en «Secrets» la ubicación por defecto y la excepción de lectura del índice, y agrega la sección «Test accounts» entre «Secrets» y «Evidence», con el texto de `design.md` («Forma propuesta») o una redacción equivalente que cumpla los criterios.

**Closes:** AC1, AC2, AC3, AC4 y AC5 (sus viñetas de la guía global), AC7 (la parte de «Secrets»)

**Depends on:** ninguna

**Locations:** `content/guidance/global.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** edita las líneas 38-39 de «Secrets» y agrega la sección. «Scope and authorization» y «Persistent memory» no cambian; se citan por su nombre.

**Verification:** `rg -n 'ACCOUNTS.md|own initiative|accounts/index|stale|password-manager reference|git/info/exclude|run on this machine|main checkout' content/guidance/global.md` encuentra cada elemento de AC1–AC3 dentro de «Test accounts», y `_support/secrets/` en «Secrets»; `go test ./tests/content/...` pasa y reporta el tamaño nuevo.

### T2 — Punteros desde skills y roles

- [x] Los nueve archivos de la tabla «Punteros» de `design.md` tienen su línea, y cada una nombra «Test accounts» sin repetir sus pasos.

**Closes:** AC4 y AC5 (sus punteros), AC6

**Depends on:** T1

**Locations:** `content/skills/flow-verify/SKILL.md`, `content/skills/flow-build/SKILL.md`, `content/skills/flow-walkthrough/SKILL.md`, `content/skills/flow-deploy/references/deployment-checks.md`, `content/skills/flow-close/SKILL.md`, `content/skills/flow-plan/SKILL.md`, `content/agents/quality/hive-verify-change.md`, `content/agents/review/hive-review-ux.md`, `content/skills/workspace-conventions/references/artifact-layout.md`

**Execution:** `hive-write-docs`; el brief nombra esta carpeta del cambio y T2

**Test approach:** check

**Changes:** solo agrega los punteros y la línea `secrets/`. No cambia el permiso que pide `deployment-checks.md` ni el seeding local que autoriza `flow-build:22`.

**Verification:** `rg -l '«Test accounts»' content/skills content/agents` devuelve los nueve archivos; `rg -n 'ACCOUNTS.md|mock data|own initiative' content/skills content/agents` solo encuentra la línea `secrets/` de `artifact-layout.md`; `go test ./tests/content/...` pasa (tamaño de skills y referencias, enlaces, roles).

### T3 — Spec y páginas del sitio

- [x] La spec delta `specs/test-accounts/spec.md` de esta carpeta tiene un escenario de la escritura al cierre. `site/src/content/docs/flows/verify.md` dice que Verify reutiliza las cuentas del índice del proyecto; `deploy.md`, que la caminata en QA usa cuentas existentes y pide permiso antes de crear; `close.md`, que el cierre guarda en memoria las filas del índice, sin valores. Cada página enlaza la guía global sin copiar sus reglas.

**Closes:** AC8

**Depends on:** T1

**Locations:** `_support/openspec/changes/test-accounts-index/specs/test-accounts/spec.md`, `site/src/content/docs/flows/verify.md`, `site/src/content/docs/flows/deploy.md`, `site/src/content/docs/flows/close.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** una frase en cada página; ninguna regla copiada.

**Verification:** la build del sitio pasa (`pnpm --dir site build`, con las dependencias que ya tiene el checkout principal); `go test ./tests/content/...` pasa, incluido el test de las spec deltas.

### T4 — Detector de lectura de secretos

- [x] `noSecretContentRead` acepta una lectura completa de `_support/secrets/ACCOUNTS.md` y sigue marcando cualquier otro archivo de `_support/secrets/`; `isSecretPath` no cambia.

**Closes:** AC7 (la parte del detector)

**Depends on:** ninguna

**Locations:** `tests/pilot/regression.go`, un test nuevo en `tests/pilot/`

**Execution:** hilo principal

**Test approach:** tdd

**Changes:** solo la regla de lectura. No agrega casos en `tests/fixtures/regression/`.

**Verification:** `go test ./tests/pilot/...` pasa con un test que comprueba tres casos: leer `ACCOUNTS.md` no falla, leer `_support/secrets/users.txt` falla, e `isSecretPath("_support/secrets/ACCOUNTS.md")` sigue siendo verdadero.

## Verificación y revisión humana

| Chequeo | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, en local | Su comando pasa | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `go test ./...` y `pnpm --dir site build` | Cubierta por la implementación |
| Integración entre tareas | Candidato final, en local | `rg -n '«Test accounts»' content` muestra la sección y los nueve punteros; ningún puntero repite pasos de la regla | Cubierta por la implementación |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` (D3 A) | La respuesta de revisión |
| CI | Tras el push a la rama del PR #352 | Checks requeridos en verde | Entrega `ready` (D2 A) |

El mensaje de cada commit nombra las sesiones que motivaron el cambio, con host, ID y fecha: Claude Code `32cffc1d` (cam, 2026-11-02), Claude Code `40d074af` (cam, 2026-11-03) y Claude Code `8d09a5eb` (carvento, 2026-11-02). El PR lista los lugares revisados (sección «Contexto verificado» de `proposal.md`) y el cambio de tamaño de `global.md`.

## Estado de la revisión y avance

2026-11-04 — Plan escrito y revisado por `hive-review-plan`; aplicados sus hallazgos (ver `proposal.md`).
- T1 hecho · `7b49d86e` · `global.md` pasa de 31 122 a 33 241 bytes (+2 119), más de los ~1,3 KB estimados.
- T2 hecho · `3bdb00d5` · lo hizo un hijo `hive-write-docs`; revisé los nueve punteros y las dos búsquedas de `rg`.
- T4 hecho · `8dcf6ae3` · se agregaron casos a las tablas que ya cubren S5: los nuevos fallaron antes del cambio y pasan después. Se omitió el caso `isSecretPath` porque prueba un detalle interno, y ningún detector de staging lo usa.

- T3 hecho · `2c312c3f` (páginas del sitio; la delta queda en la carpeta del cambio, sin commit hasta el cierre) · `pnpm build` del sitio: 17 páginas.

CI local (2026-11-04, en `2c312c3f`; sin cambios sin commit fuera de las carpetas de cambio y de sesiones):
- El único workflow (`master-source-branch.yml`) corre solo en PR a `master`, así que no aplica a este PR.
- `go vet ./...` limpio.
- `go test ./...` con salida 0.
- `pnpm build` del sitio: 17 páginas.
- `rg -c '«Test accounts»' content`: la sección y los 9 punteros.

Verify (2026-11-04, candidato `2c312c3f`):
- AC1–AC8 comprobados por su observación: `rg` y `grep` sobre la sección y los punteros; los 6 casos nuevos de S5 pasan; la delta tiene 3 requisitos y 6 escenarios; las 3 páginas enlazan la regla.
- `hive update --dry-run` acepta el contenido de `2c312c3f` (no instala nada).
- Se reusa la suite completa del build: mismo commit, sin cambios de código desde entonces.

Primera revisión `/code-review` de `7b49d86e^..2c312c3f`: 10 hallazgos, todos confirmados leyendo el código y corregidos en `358007a7`:
- La carpeta de secretos va en `0700`.
- S5 también acepta búsquedas en el índice; los tests nuevos fallaron antes del cambio.
- La excepción vive dentro de `isSecretPath`.
- Un hijo no lee ni escribe el índice.
- `hive-review-ux` no crea cuentas.
- El brief del revisor lleva las filas de cuentas.
- Se corrigió la redacción de la ruta por defecto y del modo.
- La exclusión de Git pasa a «Secrets».
- La página del sitio ya no copia la regla.
- Las filas reconstruidas desde memoria quedan sin verificar.

Suites tras el cambio: `go test ./tests/pilot/... ./tests/content/...` y la build del sitio pasan; `global.md` pesa 33 307 bytes.

Segunda revisión (`2c312c3f..358007a7`): 10 hallazgos.
- Los 5 bloqueantes se corrigieron en `5fcfa871`, con la línea `Unreviewed:`.
- Los 5 no bloqueantes van a los hallazgos del cierre.

Merge de `development` (`4f63e2e5`), con 2 conflictos contra #357 resueltos: se conservan las líneas de gh-311 y se agrega la frase de #357. `go test ./...` y `go vet` pasan.

Push 2026-11-04T07:56:51Z. El PR #352 no tiene checks y nadie lo revisó después del push. `mergeStateStatus` está en CLEAN y la base está al día.

Checkpoint: entrega `ready` en su parada · PR #352 abierto · siguiente: flow-close.
