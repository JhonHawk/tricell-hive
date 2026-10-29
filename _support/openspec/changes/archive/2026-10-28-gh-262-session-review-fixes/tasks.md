# Tareas

Commit base: `668363e2` (`origin/development`, 2026-10-28). Worktree `.claude/worktrees/gh-262-session-review-fixes`, rama `fix/gh-262-session-review-fixes`.

### T1 — H1: el push que despliega desde un repositorio `direct-base` pregunta y lleva la vista previa

- [x] La definición de cambio mecánico, la regla de `direct-base` y los disparadores de la primera vista quedan como dice `design.md` (H1).
  reabierta por la revisión del cambio D7: `8b636a67` (la vista va atada a un push que despliega, la excepción se nombra y aprobar autoriza el push)
  reabierta por la revisión de código: `fd106346` (flow-close y "that question"), re-verificada; reabierta por la re-revisión: `d9ea0d25` (flow-close apunta a la regla)

**Closes:** AC1, AC2, AC3.

**Checkpoint:** C1

**Depends on:** none.

**Locations:**
- `content/guidance/global.md:63`
- `content/skills/flow-plan/references/delivery-decisions.md:32`
- `content/skills/flow-walkthrough/SKILL.md:12,24`

**Execution:** hilo principal. Son tres ediciones de texto acopladas, y el brief sería más largo que el cambio.

**Test approach:** check, con las búsquedas `rg` de Verification.

**Changes:** aplicar los textos de `design.md` › H1.

**Verification:**
- `rg -n "documentation, a text or style change" content/guidance/global.md` no encuentra nada.
- `rg -n "documentation or style that alters no rendered layout" content/guidance/global.md` encuentra 1 línea.
- `rg -n "part of a change to code or configuration stays mechanical" content/guidance/global.md` encuentra 1 línea.
- `rg -n "any deployment that push triggers, including the closing push of a change record" content/skills/flow-plan/references/delivery-decisions.md` encuentra 1 línea.
- `rg -n "and the deployment that push triggers" content/skills/flow-plan/references/delivery-decisions.md` encuentra la línea de la lista de opciones (`:30`).
- `rg -n "other than its change record, are only in repositories that declare" content/skills/flow-plan/references/delivery-decisions.md content/skills/flow-walkthrough/SKILL.md` encuentra 1 línea en `delivery-decisions.md` y 2 en `flow-walkthrough/SKILL.md`.
- En `content/skills/flow-plan/references/delivery-decisions.md`, cada una de estas búsquedas encuentra 1 línea:
  - `rg -n "unless an answer given for this change named that deployment"`
  - `rg -n "same-kind rule of"` (el enlace `effect-authorization.md#what-an-answer-covers` aparece 2 veces: `:28` ya estaba en la base)
  - `rg -n "except a push that triggers a deployment"`
  - `rg -n "that push deploys a new or rearranged rendered layout, show its first rendered view and ask before that push"`
- `rg -c "Delivery: direct-base" content/skills/flow-walkthrough/SKILL.md` da 2.

### T2 — H2: solo bloquea un prerrequisito que únicamente el usuario puede dar

- [x] `flow-build`, los marcadores, `plan-execution` y `plan-format` quedan como dice `design.md` (H2).
  reabierta por la revisión de código: `f8ae1cdc`, re-verificada; reabierta por la re-revisión: `352e4705` (la configuración local fallida es un defecto que bloquea el trabajo; hogar en review-and-gates)

**Closes:** AC4, AC5.

**Checkpoint:** C1

**Depends on:** none.

**Locations:**
- `content/skills/flow-build/SKILL.md:62`
- `content/skills/flow-plan/references/plan-criteria-and-markers.md` (después de la tabla)
- `content/skills/flow-build/references/plan-execution.md:22`
- `content/skills/flow-plan/references/plan-format.md:17`
- `content/skills/flow-verify/references/review-and-gates.md:46` (hogar canónico, agregado por la re-revisión)

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, con las búsquedas `rg` de Verification.

**Changes:** aplicar los textos de `design.md` › H2.

**Verification:**
- `rg -n "unavailable prerequisites leave" content/skills/flow-build/SKILL.md` no encuentra nada.
- `rg -n "review-and-gates.md#runtime-and-ui-gates" content/skills/flow-build/SKILL.md content/skills/flow-plan/references/plan-criteria-and-markers.md content/skills/flow-build/references/plan-execution.md` encuentra una línea en cada archivo.
- `rg -n "and re-verify those criteria" content/skills/flow-build/references/plan-execution.md` encuentra 1 línea.
- `rg -n "how the build creates them locally" content/skills/flow-plan/references/plan-format.md` encuentra 1 línea.

### T3 — H4: la excepción de la línea de contexto cubre solo una oración

- [x] `global.md:13` y `reporting.md:13,15` quedan como dice `design.md` (H4), con `reporting.md` en 16 000 B o menos.
  reabierta por la revisión de código: `7585191e` y por la re-revisión: `910ab7e8` ("progress note" en questions.md, flow-walkthrough y site)

**Closes:** AC6.

**Checkpoint:** C1

**Depends on:** none.

**Locations:**
- `content/guidance/global.md:13`
- `content/skills/flow-close/references/reporting.md:13,15`

**Execution:** hilo principal, porque hay que ajustar los bytes en el mismo archivo.

**Test approach:** check, con las búsquedas `rg` y la prueba de tamaño de Verification.

**Changes:** aplicar los textos de `design.md` › H4.

**Verification:** `rg -n "one-line progress note" content/guidance/global.md content/skills/flow-close/references/reporting.md` no encuentra nada. `rg -n "progress note of a single sentence" content/guidance/global.md` encuentra 1 línea. `rg -n "two or more sentences is not one" content/skills/flow-close/references/reporting.md` encuentra 1 línea. `go test ./tests/content/ -run Size` pasa.

### T4 — H8: la CLI de Context7 queda exenta de la revisión de dependencias

- [x] `security-boundaries.md` y `global.md:70` quedan como dice `design.md` (H8).

**Closes:** AC7.

**Checkpoint:** C1

**Depends on:** none.

**Locations:**
- `content/skills/flow-build/references/security-boundaries.md:22`
- `content/guidance/global.md:70`

**Execution:** hilo principal, porque son dos líneas.

**Test approach:** check, con las búsquedas `rg` de Verification.

**Changes:** aplicar los textos de `design.md` › H8.

**Verification:**
- `rg -n "with its \`library\` or \`docs\` subcommand" content/skills/flow-build/references/security-boundaries.md` encuentra 1 línea.
- `rg -n "\`login\` included" content/skills/flow-build/references/security-boundaries.md` encuentra 1 línea.
- `rg -n "third-party requests" content/skills/flow-build/references/security-boundaries.md` encuentra la cláusula nueva.
- `rg -n "except running \`ctx7 library\` or \`ctx7 docs\`" content/guidance/global.md` encuentra 1 línea.
- `rg -n "installing it is not excepted" content/guidance/global.md` encuentra 1 línea.

### T5 — Sitio y registro de cambios

- [x] `site/src/content/docs/flows/build.md` y `CHANGELOG.md` describen el cambio, como dice `design.md` (Sitio y registro de cambios).
  ronda 1 (tras D7) · AC8 · el CHANGELOG no reflejaba D7-A ni la configuración local fallida; corregido en el commit siguiente
  ronda 1 · AC8 · (corregido en `9f597c03`, verificado en la ronda 2) · `build.md:90` limitaba la pregunta a documentación y omitía la excepción de la respuesta para el mismo cambio

**Closes:** AC8.

**Checkpoint:** C2

**Depends on:** T1, T2, T3, T4.

**Locations:**
- `site/src/content/docs/flows/build.md:42,90`
- `CHANGELOG.md` (Unreleased)

**Execution:** hilo principal.

**Test approach:** check, con las búsquedas `rg` de Verification.

**Changes:** una oración en la página y una entrada en el registro de cambios.

**Verification:**
- `rg -n -i "support repositor" site/src/content/docs/flows/build.md` encuentra la oración nueva.
- `rg -n "\`direct-base\` and unattended delegation there is no such stop" site/src/content/docs/flows/build.md` no encuentra nada.
- `rg -n "in a plain line" site/src/content/docs/flows/build.md` no encuentra nada, y `rg -n "single-sentence" site/src/content/docs/flows/build.md` encuentra la oración de `:42`.
- `rg -n "#262" CHANGELOG.md` encuentra 1 línea.

## Verificación y revisión humana

- C1 — T1, T2, T3 y T4 — T4 toca un límite de seguridad (la revisión de dependencias), así que cierra el grupo; las otras tres son ediciones de texto independientes.
- C2 — T5 — documenta el comportamiento ya integrado de C1.

| Gate | Cuándo y dónde | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local completa | candidato final, en el worktree | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'`, todo en verde | implementación |
| Revisión dedicada | candidato local, antes del push | `/code-review` de Claude Code sobre el diff contra `development`; hallazgos corregidos o refutados con evidencia | D4-A |
| Revisión de conducta completa | antes del PR | lista de lugares revisados (`global.md`, skills, referencias, roles, `site/`, `_support/docs/architecture/`, `tests/pilot/`) en el cuerpo del PR, con el tamaño de `global.md` antes y después | `AGENTS.md` |
| Verificación local en vivo y de UI | — | no aplica: el cambio es texto de guía, sin interfaz renderizada ni comportamiento ejecutable | — |

## Estado de revisión y avance

### Revisión del plan

Revisión `060fec9f512f` (2026-10-28), con dos hijos `hive-review-plan` nativos, de solo lectura.

**Coherencia de la guía** (unos 153 s, 130 mil tokens). Cinco bloqueantes, todos aceptados:
- "publish without asking" contradecía H1: acotado.
- `effect-authorization.md:29` permite reusar una respuesta del mismo tipo: H1 declara la excepción en su punto.
- `site/.../build.md:90` decía que `direct-base` no tiene parada: acotado a "chosen for a code repository".
- El paso 4 no trataba un `cannot verify` por prerrequisito local: ahora lo crea y vuelve a verificar.
- Faltaban búsquedas para la segunda mitad de AC2 y para AC6: agregadas.

Sus dos no bloqueantes:
- "progress line" en `reporting.md:61,110`: aceptado si los bytes alcanzan.
- H2 repetida en tres lugares: aceptado; ahora son punteros a Runtime and UI gates.

**Límite de seguridad** (unos 62 s, 96 mil tokens). Tres bloqueantes, todos aceptados:
- B1: el "unless" podía omitir nombrar el despliegue y cubrir el push con un modo reutilizado o un pedido sin aviso. Ahora el despliegue se nombra siempre y solo lo cubre una respuesta para ese cambio que lo nombró.
- B2: `global.md:70` perdía el límite de instalación. Lo conserva.
- B3: la exención no nombraba el paquete ni los subcomandos. Ahora nombra `ctx7` con `library` y `docs`, y excluye `login` y la instalación.

Su no bloqueante, la regla de solicitudes a terceros: aceptado.

**Sin nueva ronda.** Las correcciones aplican las propuestas de los propios revisores, así que se reconcilian sin otra revisión (`plan-review.md`).

**Decisiones tomadas sin preguntar:**
- `login` no queda exento. Es la lectura conservadora de D2-B, porque `login` configura credenciales en lugar de consultar documentación.
- No se edita `hive-verify-task.md`: su `cannot verify` sigue siendo correcto y quien decide crear el prerrequisito es el padre.

### Revisión 2 (2026-10-28, pedida por el usuario)

El usuario acotó AC1 a un trabajo que es únicamente documentación, y con ello AC2 y AC3, para no meter fricción cuando la documentación acompaña un trabajo sobre el proyecto real.

**Qué cambia:**
- AC1, AC2 y AC3 (texto en `proposal.md`).
- `design.md` › H1:
  - excepción para la documentación dentro de otro trabajo;
  - la vista previa y la captura solo en trabajo que toca únicamente repositorios de soporte;
  - las opciones de entrega nombran el despliegue del push del repositorio de soporte.
- La verificación de T1 y la oración de T5.

**Qué invalida:** ninguna tarea estaba hecha.

**Re-revisión** (el mismo revisor de coherencia, una ronda, unos 64 s, 142 mil tokens). Un bloqueante y tres no bloqueantes, todos aceptados con la propuesta del propio revisor, así que no hay otra ronda:
- **Bloqueante:** criterios de alcance distintos en AC1 y en AC2/AC3, y el registro de cambios en un repositorio de código sacaba el trabajo del alcance. Ahora hay un criterio único que excluye el registro, y AC1 dice "part of a change to code or configuration".
- **No bloqueante:** la excepción de AC1 casi no quita fricción, porque la clasificación es por cambio entero. Queda documentado en `design.md` que el disparador general de la primera vista no cambia.
- **No bloqueante:** `delivery-decisions.md:30` no listaba el despliegue en cada opción. Ahora lo lista.
- **No bloqueante:** el push del registro al cierre no quedaba cubierto. Las opciones ahora lo nombran.

**Durante el build:** el verificador de T3 notó que `site/.../build.md:42` todavía describe la nota de avance como "a plain line". Entra en T5 y AC8 como decisión tomada sin preguntar: es la página que `AGENTS.md` pide actualizar en el mismo PR. El ejemplo de `delivery-decisions.md:32` sin despliegue, que el verificador de T1 marcó como no bloqueante, queda igual: la regla nombra el despliegue de forma explícita y el ejemplo describe repositorios que no despliegan.

### Revisión de código (`/code-review` medium, candidato `9f597c03`)

10 hallazgos. Corregidos 7 en tres commits adicionales:
- `fd106346` (T1): el push del registro al cierre en `flow-close` ahora pregunta si despliega, y "that question" ya no es ambiguo.
- `f8ae1cdc` (T2): un prerrequisito local cuya creación falla bloquea, nombrando la falla, con un solo intento de creación.
- `7585191e` (T3): la nota de avance se llama "sentence" o "note" en `delegation.md` y `reporting.md`.

Descartados 3:
- ~~Pedir "reacomoda y publica" nombra la publicación, así que no mostrar vista previa es lo pedido.~~ Revertido por D7-A: en trabajo solo de documentación con layout, siempre se muestra la vista y se pregunta.
- La excepción `@latest` de ctx7 es la decisión D2-B del usuario.
- El conflicto con la delegación desatendida ya existía: el texto anterior también pedía preguntar.

Re-revisión: una sola, limitada al diff `9f597c03..7585191e`, porque el arreglo de `flow-close` cambia una regla.

### Re-revisión (`/code-review` medium sobre `9f597c03..7585191e`)

Encontró 9 puntos. Corregidos sin otra revisión, porque el arreglo es local y está dentro del alcance:
- `352e4705` (T2): `review-and-gates.md:46` trata la configuración local que falla como un defecto que bloquea el trabajo, dentro de los límites de rondas de corrección. Los demás archivos apuntan ahí. Esto quita el "una sola vez" y la deriva entre tres textos.
- `d9ea0d25` (T1): `flow-close` apunta a la regla en lugar de restatearla en parte.
- `910ab7e8` (T3): las menciones restantes de "progress line" pasan a "progress note".

Reportado sin corregir:
- P3: a `reporting.md` le quedan 36 B de margen.

Al usuario como decisión:
- D7: la vista previa se salta cuando una respuesta anterior de este mismo cambio ya nombró el despliegue.

### D7 (2026-10-28)

El usuario eligió D7-A: en trabajo solo de documentación con un layout nuevo o reacomodado, siempre se muestra la primera vista y se pregunta antes del push que despliega, aunque una respuesta ya haya nombrado el despliegue.

Commits:
- `ed52b514` (T1): `delivery-decisions.md:32` y `flow-walkthrough/SKILL.md:24`.
- `e3ab8aa2` (T5): `site/.../build.md:90`.

Se actualizaron AC2 y la verificación de T1.

**Cierre:** las cinco tareas quedaron `[x]` y la suite local completa pasó sobre `8b636a67`; `9b09ba6a` solo cambia una palabra y pasa las pruebas de contenido. La revisión del cambio D7 (`910ab7e8...e3ab8aa2`) dio 6 hallazgos, corregidos en `8b636a67`, `9b09ba6a` y el registro. PR #263 integrado (`8ed2e6a4`).
