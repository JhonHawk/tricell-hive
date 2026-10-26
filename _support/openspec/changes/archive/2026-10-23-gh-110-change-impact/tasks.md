# Tareas · gh-110-change-impact

Commit base: `ad6099c1` (`development`, 2026-10-23); rama `feat/gh-110-change-impact`.

### T1 — Referencia canónica de impacto

- [x] `content/skills/flow-research/references/change-impact.md` existe con el texto de [design.md](design.md#texto-propuesto-de-la-referencia-en-inglés-como-el-contenido-distribuido).

**Closes:** AC1.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-research/references/change-impact.md` (nuevo).

**Execution:** hilo principal; es texto de guía acoplado a T2 y su brief sería más largo que el cambio.

**Test approach:** check, con un `rg` por cláusula de AC1 sobre el archivo nuevo (`F`):

- disparador: `rg -n "^## When to read" F`, `rg -n "moves it to another host or process" F`, `rg -n "no test or CI run covers" F`;
- exclusión: `rg -n "Renaming a symbol that only this code uses is neither" F`;
- paso 1: `rg -n "Always run step 1" F`, `rg -n "search for every direct caller of the code that changes" F`, `rg -n "Search for the value, not only the symbol" F`;
- señales: `rg -c "Signal:" F`;
- paso 2: `rg -n "which host or process runs what" F`;
- paso 3: `rg -n "code paths the change makes reachable" F`;
- paso 4: `rg -n "Mark that consumer unverified" F`, `rg -n "window long enough" F`, `rg -n "how to revert" F`;
- proporcionalidad: `rg -n "needs no further step" F`.

**Changes:** escribir la referencia según el diseño; sin nombres de herramientas de un host.

**Verification:** cada `rg -n` devuelve una línea y `rg -c "Signal:"` devuelve 3.

### T2 — Punteros y reconciliación de la regla de interfaces

- [x] `flow-research`, `flow-plan`, `flow-build`, `plan-review.md` y `data-changes.md` enlazan la referencia en sus puntos de uso; la regla de interfaces y su copia quedan reconciliadas.

**Closes:** AC2, AC3, AC4, AC5, AC8.

**Depends on:** T1.

**Locations:** `content/skills/flow-research/SKILL.md:16`; `content/skills/flow-plan/SKILL.md:22,24`; `content/skills/flow-build/SKILL.md:16,42`; `content/skills/flow-plan/references/plan-review.md:15,17`; `content/skills/flow-build/references/data-changes.md:7`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, con un `rg` por cláusula:

- AC2: `rg -n "change-impact" content/skills/flow-research/SKILL.md` en la sección «Frame the investigation».
- AC3: `rg -n "Before planning a change that removes" content/skills/flow-plan/SKILL.md` (puntero junto a seguridad y datos); `rg -n "search for consumers as it sets" content/skills/flow-plan/SKILL.md` y `rg -n "never becomes the condition that retires a transition" content/skills/flow-plan/SKILL.md` (regla de interfaces).
- AC4: `rg -n "When no plan exists, or the plan records no change-impact result" content/skills/flow-build/SKILL.md`; `rg -n "does not account for" content/skills/flow-build/SKILL.md`; `rg -n "observation that only chose the approach" content/skills/flow-build/SKILL.md`.
- AC5: `rg -n "change-impact" content/skills/flow-plan/references/plan-review.md` en las filas «Backend / data / integration» e «Infrastructure / delivery».
- AC8: `rg -n "change-impact" content/skills/flow-build/references/data-changes.md`.

**Changes:** aplicar las secciones «Reconciliación de la regla de interfaces» y «Punteros» de [design.md](design.md).

**Verification:** cada `rg` devuelve su línea en la ubicación indicada; el de AC5 devuelve exactamente las dos filas. La validación de enlaces de `go test ./...` (puerta compartida) resuelve todos los enlaces nuevos.

### T3 — Code-review cuenta caminos que el cambio vuelve alcanzables

- [x] `hive-review-code.md` paso 3 trata como introducido un defecto previo que el cambio vuelve alcanzable o amplía, y `verification.md` no lo devuelve a la regla de incidentales.

**Closes:** AC6.

**Depends on:** ninguna.

**Locations:** `content/agents/review/hive-review-code.md:14`; `content/skills/flow-build/references/verification.md:43`.

**Execution:** hilo principal; dos frases.

**Test approach:** check: `rg -n "makes reachable or widens access" content/agents/review/hive-review-code.md`; `rg -c "not code that was already there" content/agents/review/hive-review-code.md`; `rg -n "path the change makes reachable" content/skills/flow-build/references/verification.md`.

**Changes:** aplicar la sección «Code-review (P4)» de [design.md](design.md).

**Verification:** el primer y el tercer `rg` devuelven su línea; el segundo devuelve 0.

### T4 — Sitio y changelog

- [x] Las páginas de Research y Plan describen el chequeo y el changelog lo registra.

**Closes:** AC7.

**Depends on:** T1, T2.

**Locations:** `site/src/content/docs/flows/research.md` («What it produces»), `site/src/content/docs/flows/plan.md`, `CHANGELOG.md` (`[Unreleased]`).

**Execution:** hilo principal; dos párrafos y una entrada.

**Test approach:** check: `rg -n "change-impact" site/src/content/docs/flows/research.md site/src/content/docs/flows/plan.md` y `rg -n "#110" CHANGELOG.md`.

**Changes:** aplicar la sección «Sitio y changelog» de [design.md](design.md); las páginas describen el comportamiento y enlazan la referencia sin copiar sus pasos.

**Verification:** el primer `rg` devuelve una línea en cada página; el segundo, la entrada bajo `[Unreleased]`.

### T5 — Renombres y movimientos fuera del código no son mecánicos

- [x] La definición de cambio mecánico de `global.md` excluye renombrar o mover algo que lee o fija algo fuera del código.

**Closes:** AC9.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md:73`.

**Execution:** hilo principal; una cláusula.

**Test approach:** check: `rg -n "renames or moves something a host, process, or configuration outside the code reads or sets" content/guidance/global.md` y `wc -c content/guidance/global.md`.

**Changes:** añadir la cláusula a la lista de lo que «is not mechanical, however small».

**Verification:** el `rg` devuelve la línea 73; `wc -c` da el tamaño nuevo, que el PR declara.

## Verificación compartida

| Puerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | Candidato final, checkout local | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`: todo pasa | Parte de la implementación |
| Tamaño de la guía global | Candidato final | `wc -c content/guidance/global.md`: 33 608 → 33 710 bytes (+102, solo T5) | Parte de la implementación |
| Revisión del cambio | Tras T1–T4 | Mecanismo pendiente de la pregunta de revisión | Pendiente |

No aplica verificación de UI ni en vivo: el cambio es texto de guía. No hay prueba de comportamiento del modelo (pilotos en pausa por `AGENTS.md`); la validación real será observar las próximas sesiones sobre legacy.

## Revisión y progreso

- 2026-10-23: plan escrito sobre `ad6099c1`, revisión `ac5ae5fbc8da`.
- 2026-10-23: revisión del plan (`hive-review-plan`, dominio guía distribuida; 88 s, 113 724 tokens): 3 bloqueantes y 3 no bloqueantes, todos aceptados.
  - B1: la frase P2 reabría retiros condicionados al uso → la observación solo elige el enfoque y nunca retira una transición; `flow-build:42` alineado.
  - B2: disparador demasiado amplio → excluye cambios mecánicos, acota «redeploys» a mover de host o proceso y el código sin pruebas a cambios de comportamiento en producción; paso 1 limitado al repositorio salvo valor compartido.
  - B3: verificaciones incompletas → un `rg` por cláusula; Locations corregidas (`flow-build:16,42`); la fila de Infraestructura enlaza la referencia.
  - N1: el paso 3 ya no redefine `global.md:31`; `verification.md:43` añadido a T3 y AC6.
  - N2: `data-changes.md:7` con puntero; nuevo AC8.
  - N3: puntero breve en `flow-plan:24`.
- 2026-10-23: re-revisión de `738d9d952b7a` (mismo revisor; 49 s, 124 701 tokens): B1 y B3 resueltos, casos de origen cubiertos, AC1–AC8 falsos en la base y cubiertos. Dos ajustes propuestos por el revisor, aplicados tal cual sin otra ronda: el paso 1 busca en todo el workspace cuando el valor es un recurso de red; el puntero de `flow-research:16` usa la forma breve que remite a «When to read».
- 2026-10-23: build en `feat/gh-110-change-impact` sobre `ad6099c1`. Implementador: hilo principal (Opus 5.5, observado). Verificadores `hive-verify-task`: modelo configurado, no observado.
  - T1 `ddcba98a`: AC1 met (38 s).
  - T3 `7676123a`: AC6 met (40 s); ajustes de redacción del verificador aplicados («counts as introduced», sin punto y coma).
  - T2 `146ce34d`: AC2–AC5 y AC8 met (88 s); validación de enlaces `ok`; frase de `flow-build:16` reordenada según el verificador.
  - T4 `64f51c82`: AC7 met (70 s); dos imprecisiones de las páginas corregidas (exclusión de cambios mecánicos, señales del paso 3, disparador completo en Plan).
- Suite completa sobre el candidato: `go vet ./...`, `go test ./... -count=1` y unittest de skills (55 pruebas) pasan; `global.md` sigue en 33608 bytes. Ninguna prueba cubre las páginas del sitio, así que no se repitió tras los ajustes de T4.
- Sin revisión de seguridad: el cambio no cruza ningún límite de `security-boundaries.md`.
- 2026-10-23: validado por el usuario (D6-A); push y PR #111. `/code-review` (115 s, 111 280 tokens) dio 10 hallazgos:
  - P1 corregido: F1, la exclusión de cambios mecánicos dejaba fuera renombres y movimientos; ahora solo quedan fuera los que nada fuera del código lee.
  - P2 corregidos: F2 (build con plan sin lista de impacto), F3 (`verification.md` sin depender de una etiqueta del revisor), F4 (punteros circulares: ahora nombran el disparador), F5 (la observación vive en el paso 4 de la referencia; `flow-plan` solo remite), F6 (la búsqueda de la regla de interfaces solo cuando la referencia aplica), F8 (el paso 1 corre cuando hay un valor con nombre).
  - P3: F7 corregido (párrafo del sitio acortado); F10 solo reportado (`plan-format` sin hueco para la lista de impacto).
  - Descartado: F9; el sitio se publica desde `master` con la release (CONTRIBUTING.md, «Documentation site»), así que la página y la referencia llegan juntas.
  - Revisión del plan: AC1 cambia de redacción (exclusión y paso 1); T1 vuelve a `[?]` para reverificar. Las búsquedas de T2 y T3 se actualizan sin cambiar sus criterios.
- 2026-10-23: re-revisión de `/code-review` sobre `64f51c82..fc13f88e` (108 s, 107 230 tokens), 10 hallazgos:
  - Corregidos sin otra revisión (locales, dentro del alcance): R2 (exclusión circular → solo símbolos que usa este código; ante la duda, paso 1), R4 (paso 1 siempre; añade ruta de archivo, paquete y llamadores de un cambio de comportamiento), R5 y R6 (sitio y changelog con el disparador completo), R7 (la observación informa cuando puede correr; sin ella se decide con la evidencia y el consumidor queda no verificado), R8 (punteros alineados con «moves it to another host or process»), R9 (el orquestador comprueba contra el diff cada hallazgo marcado como previo), R10 (build relee solo si el plan no registra el resultado; la línea de señales va al contexto verificado).
  - Descartado: R3; la regla de interfaces ya exige «identify its actual producers and consumers» en todo cambio de interfaz.
  - A decisión del usuario: R1, `global.md:73` llama mecánicos a renombres y movimientos, así que una tarea así no carga ninguna skill ni la referencia.
  - AC1 cambia de redacción otra vez (exclusión y paso 1); T1 vuelve a `[?]`.
- 2026-10-23: T1 reverificada (AC1 met, 47 s); ajuste «every direct caller» del verificador aplicado. D7-A: nueva T5 con AC9 en `global.md:73`; sin nueva revisión, como decía la opción.
- 2026-10-23: T5 verificada (AC9 met, 51 s); suite completa pasa. PR #111 fusionado con merge commit `1000ee23`; #110 cerrada por GitHub; `hive update` desplegó `1000ee23` en local (la referencia desplegada coincide con el checkout). Sugerencia del verificador de T5 (pista «ante la duda» en `global.md`, ~60 bytes) no aplicada: excede lo decidido en D7-A.
