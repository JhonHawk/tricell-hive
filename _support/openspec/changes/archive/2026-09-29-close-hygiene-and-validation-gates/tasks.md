# Tareas

Base del build: `deb2958` (`rebuild/harness-engineering`), rama `feat/close-hygiene-and-validation-gates`, 2026-09-29.

Orden: T1 primero (reubica texto en `global.md` y `flow-build`); T2, T3 y T4 editan los mismos archivos y van en serie después; T5 cierra.

Cambio de ejecución al empezar el build: T2, T3 y T4 van a `hive-write-spec` (Sonnet) en un solo encargo después de T1, no al hilo principal. El hilo principal corre en Opus, igual que `hive-verify-task`, y la regla de `flow-build` prohíbe verificar con el mismo modelo que implementó.

## T1 — Skill `flow-close` y reubicación de la higiene

- [x] `flow-close` existe, concentra la higiene y el registro del cambio se versiona una sola vez al cierre.

**Closes:** AC1, AC2, AC3, AC4.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-close/SKILL.md` (nuevo); `content/guidance/global.md` (líneas 123, 124 y 154; la 144 se conserva); `content/skills/flow-build/SKILL.md` "Verify and close" (línea 75) y el párrafo del estado del plan (línea 44); `content/skills/git-workflow/SKILL.md` línea 8 y "Clean up after a merge"; `content/skills/flow-plan/SKILL.md` último párrafo; `content/skills/flow-plan/references/change-records.md` "Close a change"; `content/skills/flow-plan/references/delivery-decisions.md` líneas 14 y 18 y el registro de la línea 32; `content/skills/workspace-conventions/SKILL.md` línea 22.

**Execution:** delegada a `hive-write-spec`: el diseño está cerrado y es un solo escritor sobre archivos que ninguna otra tarea toca mientras corre. Debe leer `_support/docs/architecture/instruction-resources.md` y AGENTS.md (reglas de redacción de guía distribuida: un hogar por regla, criterio con su alcance, capacidad en vez de nombre de herramienta, redactar para el modelo menos capaz).

**Test approach:** check, con los `rg` de abajo.

**Changes:** según [diseño, `flow-close`](design.md#flow-close). Mover el texto, no duplicarlo: cada procedimiento queda solo en `flow-close` y el origen conserva un enlace con la condición de lectura. `global.md` conserva la autorización de borrado (incluida la de ramas mergeadas, que hoy concede `git-workflow` línea 36) y el disparador. El procedimiento de limpieza de `git-workflow` pasa sin cambios de contenido salvo esa frase de concesión. Reescribir las referencias en prosa que el diseño lista.

**Verification:** desde la raíz del repo,

```bash
test -f content/skills/flow-close/SKILL.md && head -5 content/skills/flow-close/SKILL.md
rg -n "when the plan is ready|commit and push it now|at each delivery milestone|when the plan is approved" content/
rg -n "git fetch --prune" content/
rg -n "^## Close a change" content/
rg -n "versioning points|points the delivery decision names|change-folder paragraph|post-merge cleanup in \[git-workflow|closing procedure in \[change records|flow-plan. skill carries" content/
rg -n "authorizes deleting merged work branches" content/
rg -n -i "closing route" content/skills/flow-plan/references/delivery-decisions.md content/skills/flow-close/SKILL.md
wc -c content/guidance/global.md
rg -n "flow-close" content/guidance/global.md content/skills/flow-build/SKILL.md content/skills/git-workflow/SKILL.md content/skills/flow-plan/ content/skills/workspace-conventions/SKILL.md
```

Resultado esperado:
- el `head` muestra `name: flow-close` y una descripción que menciona el cierre, el fin de sesión y la petición de higiene;
- el segundo, el cuarto y el quinto `rg` no devuelven nada;
- el tercero solo encuentra `content/skills/flow-close/SKILL.md`;
- el sexto solo encuentra `content/guidance/global.md`, con el texto fijo del diseño;
- el séptimo encuentra la vía de cierre en la pregunta de entrega y en `flow-close`;
- `wc` da 43 180 bytes o menos, dejando 100 para T4;
- el octavo encuentra el enlace en cada archivo de origen.

Leer `flow-close` y confirmar los casos de versionado de AC3 (`direct-base`, specs en el repo de código después del merge, `mv` sin historia, `hold`, pausa y abandono) y el orden del cierre con specs en el repo de código: base avanzada, commit del registro, publicación, borrado de ramas incluida la de cierre.

## T2 — Correcciones sin gate intermedio

- [x] Las correcciones durante la validación no relanzan los hijos y el gate final se acota a la diferencia.

**Closes:** AC5.

**Depends on:** T1 (ambas tocan `flow-build/SKILL.md`).

**Locations:** `content/skills/flow-build/references/verification.md` "Human UI handoff" (línea 67); `content/skills/flow-build/SKILL.md` "Work outside the plan's tasks".

**Execution:** hilo principal: dos párrafos acoplados a T3 y T4 en los mismos archivos; el brief sería más largo que el cambio.

**Test approach:** check.

**Changes:** según [diseño, correcciones](design.md#correcciones-durante-la-validación).

**Verification:**

```bash
rg -n "layout adjustments" content/
rg -n -i "correction" content/skills/flow-build/references/verification.md content/skills/flow-build/SKILL.md
```

Resultado esperado: el primero no devuelve nada; el segundo muestra el párrafo generalizado, que nombra los cuatro mecanismos omitidos entre rondas (`hive-review-ux`, `hive-verify-change`, `hive-verify-task`, revisión de código), el gate único acotado a la diferencia entre el último candidato revisado y el aceptado, la tarea reabierta a `[?]` con "reopened by correction round" y que TDD sigue aplicando en cada ronda.

## T3 — Tests acotados y re-verificación por criterio

- [x] La iteración usa la selección de tests afectados con disparadores de ampliación, y la re-verificación revisa solo lo que falló.

**Closes:** AC6, AC7.

**Depends on:** T2.

**Locations:** `content/skills/flow-build/references/verification.md` "Frequency and coverage" primera viñeta; `content/skills/flow-build/SKILL.md` paso 3 del ciclo de `hive-verify-task` y "While iterating, run the targeted tests"; `content/agents/quality/hive-verify-task.md` líneas 10 y 16 (obligatorio: hoy califica todos los criterios que la tarea cierra).

**Execution:** hilo principal, por el mismo motivo que T2.

**Test approach:** check.

**Changes:** según [diseño, tests acotados](design.md#tests-acotados).

**Verification:**

```bash
rg -n -i "selects no test|unchanged candidate" content/skills/flow-build/
rg -n -i "lockfile|migration" content/skills/flow-build/references/verification.md
rg -n -i "not met" content/skills/flow-build/SKILL.md
rg -n -i "names criteria|named criteria|only those" content/agents/quality/hive-verify-task.md
```

Resultado esperado: el primero encuentra la selección vacía como no verificada y la prohibición de repetir sobre un candidato sin cambios; el segundo, los disparadores de ampliación; el tercero, que la re-verificación se limita a los criterios `not met` y el brief los nombra; el cuarto, que el rol califica solo los criterios nombrados cuando el brief los nombra.

## T4 — Atascos como hallazgos y respuesta "no volver a mencionar"

- [x] Los atascos con umbral se reportan como hallazgos y cada hallazgo admite tres respuestas.

**Closes:** AC8, AC9.

**Depends on:** T3.

**Locations:** `content/skills/flow-build/references/verification.md` "Frequency and coverage"; `content/guidance/global.md` línea 40; `content/skills/flow-close/SKILL.md` paso de hallazgos.

**Execution:** hilo principal, por el mismo motivo que T2.

**Test approach:** check.

**Changes:** según [diseño, atascos](design.md#atascos-como-hallazgos).

**Verification:**

```bash
rg -n -i "deadlock|lock timeout" content/skills/flow-build/references/verification.md
rg -n -i "mention again" content/guidance/global.md content/skills/flow-close/SKILL.md
wc -c content/guidance/global.md
```

Resultado esperado: el primero muestra la lista de seis umbrales absolutos y la separación proyecto/Hive; el segundo, la tercera respuesta con memoria persistente por proyecto y tema y la búsqueda previa que descarta silenciados; `wc` da 43 280 bytes o menos.

## T5 — Documentación, lista protegida y validación

- [x] La documentación y el piloto conocen `flow-close`, y el catálogo valida.

Ediciones hechas en paralelo a T1 (`README.md`, `workspace-and-artifacts.md`, `protected.go`); la validación corre sobre el commit final.

**Depends on:** T4.

**Locations:** `tests/pilot/protected.go` línea 56 (lista de skills protegidas); `README.md` línea 46 (procedimientos portables); `_support/docs/architecture/workspace-and-artifacts.md` línea 86 (cierre).

**Execution:** hilo principal: tres líneas.

**Changes:** agregar `flow-close` a la lista protegida y a la mención de procedimientos; alinear la descripción del cierre en la documentación de arquitectura con un enlace a la skill.

**Verification:**

```bash
go vet ./...
go test ./...
go run ./tooling/cli update --dry-run --source .
```

Resultado esperado: `vet` y `test` sin fallos; `update --dry-run` termina sin errores de catálogo ni de enlaces y lista `flow-close` entre las skills. Como `update` despliega el `HEAD` confirmado, esta comprobación se corre sobre el commit final de la rama.

## Verificación y revisión humana

Sin superficie de UI ni runtime: la verificación es la lectura de la guía y los comandos de cada tarea. No aplican `hive-review-ux`, `hive-verify-change` ni paseo in-vivo. `hive-verify-task` verifica T1 a T4 (test approach `check`) contra sus criterios.

| Gate | Momento | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | Tras cada tarea | `hive-verify-task` con los `rg` de la tarea | Parte de implementar |
| Catálogo y tests | Candidato final | T5 | Parte de implementar |
| Revisión de código | Antes del merge | Ninguna dedicada (D5-C) | Decisión del usuario |

## Revisión del plan y avance

Revisión del plan, ronda 1 (`hive-review-plan`, Claude Code nativo, dominio guía distribuida; candidato proposal `45442e39`, design `65a11899`, tasks `cc6dee51`): cinco huecos bloqueantes y tres sugerencias.

- B1, archivar antes del merge chocaba con `verification.md` 41 y 81 y con `change-records.md` 59: decisión del usuario D6-A (archivar después del merge).
- B2, pausa y abandono: pausa no archiva; abandono archiva con razón por la vía del cierre.
- B3, `hive-verify-task.md` calificaba todos los criterios: su edición pasa a obligatoria en T3.
- B4, referencias en prosa sin chequeo: `rg` agregado en T1.
- B5, dueño de la concesión de borrado de ramas: `global.md`, con `rg` en T1.
- N1, presupuesto de `global.md`: tope +300 bytes.
- N2, umbrales: absolutos de 5 minutos y 3 minutos sin salida; memoria nombrada como capacidad.
- N3, marcadores en rondas de corrección: tarea reabierta a `[?]`.

Re-revisión (mismo revisor reanudado; candidato proposal `976247a7`, design `32d59337`, tasks `b5b1c892`): confirmó B2, B3, N2, N3 y el tope de bytes, y encontró cinco puntos nuevos. Se aplicaron tal como los propuso, sin otra ronda:

- R1, la vía de cierre necesita autorización del usuario: la pregunta de entrega la nombra y `flow-close` actúa solo sobre la respuesta.
- R2, borrar worktrees podía destruir el plan sin seguimiento: solo worktrees limpios y sin forzar.
- R3, orden imposible del cierre con specs en el repo de código: base primero, commit del registro, ramas al final.
- R4, `rg` de T1 que rechazaban texto correcto: anclados a los enlaces viejos y frase de concesión fija.
- R5, tope de bytes comprobado tarde: `wc` también en T1 (43 180).

Build: T1 verificada (AC1–AC4 `met`, `hive-verify-task` Opus sobre implementación Sonnet) y commiteada en `b160923`. El verificador dejó dos observaciones que entran con T2–T4: O2, `flow-build` línea 71 une con "and" el cierre y la finalización, y no nombra la pausa; O3, un repo de specs aparte sin `direct-base` no tenía caso de versionado, y se trata como el de specs en un repo de código (vía de cierre en la pregunta de entrega).

T2, T3 y T4 los implementó `hive-write-spec` (Sonnet) con O2 y O3, y los verificaron tres `hive-verify-task` (Opus) en paralelo, de 34 a 37 s cada uno, solo con `rg` y lectura: AC5, AC6, AC7, AC8 y AC9 `met`. Después del veredicto de T3, el orquestador agregó una cláusula que su verificador marcó como hueco (O4): sin selección de tests en el proyecto, se corren los tests del paquete cambiado. No toca lo que comprueban los `rg` de AC6. Quedan sin cambiar dos observaciones de redacción del verificador de T4: los hallazgos de flujo nombran "declinar" solo en `flow-close`, y lo declinado se guarda por tema sin proyecto, como antes.

T5: `go vet ./...` sin hallazgos; `go test ./...` falló en un solo test, `tests/content` (`TestGlobalGuidanceStaysWithinRatchetedBudget`), porque `global.md` quedó en 43 254 bytes contra un tope ya fijado en el repo de 43 000 bytes, que el plan no había visto (el plan permitía 43 280). Se condensaron las líneas de `global.md` que ya remiten a `flow-close` hasta 42 988 bytes, sin tocar la frase de concesión ni la tercera respuesta, y `tests/content` pasa. El resto de paquetes pasó en la misma corrida (`tooling/cli` 205 s). `update --dry-run --source .` sobre `861a1ef` valida el catálogo y lista `flow-close`. Commits: `b160923` (T1), `25b671e` (T2–T4), `861a1ef` (T5 y condensación).

Entrega: push de `feat/close-hygiene-and-validation-gates`, PR #67 y merge en `rebuild/harness-engineering` como `343d694` (2026-09-29). Sin CI en el repo y sin revisión dedicada (D5-C). Cierre por la vía elegida (D6-A): esta carpeta se archiva y se versiona con un push directo a la base.

Límites: la re-revisión no cubre la redacción final, que verifica `hive-verify-task`. `rebuild/harness-engineering` no tiene protección de rama ni rulesets (`gh api`, 2026-09-29), así que el push directo del cierre de este cambio es posible. Plan listo.
