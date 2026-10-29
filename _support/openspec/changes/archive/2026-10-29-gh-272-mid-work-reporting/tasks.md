# Tareas

Base del cambio: `development` fa3e3948, confirmada al iniciar el build. La rama es `fix/gh-272-mid-work-reporting`, en `.claude/worktrees/gh-272-mid-work-reporting`. El change folder vive en ese worktree desde el inicio del build; la copia del checkout principal quedó obsoleta.

## T1 — Reglas canónicas: global, reporting, work-item boundary, questions y delegation

- [x] Las reglas canónicas dejan de exigir reportar a mitad del trabajo. Al esperar, se pide una oración. La lista de lo entregado tiene un solo hogar. El lenguaje se reduce al idioma de la sesión. El acuse de una respuesta sigue D2-A.

**Closes:** AC2, AC5, AC6, AC7

**Checkpoint:** C1

**Depends on:** none

**Locations:**

- `content/guidance/global.md:13,20,29,63`
- `content/skills/flow-close/references/reporting.md`
- `content/skills/flow-close/references/work-item-boundary.md`
- `content/skills/hive-core/references/questions.md:88,90,118`
- `content/skills/flow-research/references/delegation.md:57-59,67,89-91`; el título de la sección de reparto se cambia en T2
- `content/voices/preamble.md`
- `tests/content/voices_test.go`

**Execution:** main thread. Es la definición canónica de la que dependen T2 y T3, y su redacción está acoplada entre archivos.

**Test approach:** check: `go test ./tests/content/...` y las búsquedas de AC2, AC5, AC6 y AC7.

**Changes:** aplicar los apartados de [design.md](design.md) "La línea de contexto", "Espera", "Turno de notificación", "Información que se mueve al cierre", "Reporte de cierre" (la lista única de `reporting.md`), "Lenguaje" y "Acuses".

- Reescribir el ejemplo de `reporting.md`, que hoy tiene un bloque "In progress", como un reporte de cierre con un pendiente.
- Reescribir también el ejemplo de espera de `delegation.md`.
- No tocar el encabezado `## Report a running or completed item`.

**Verification:**

- `rg -n -F -e "labeled in-progress item" -e "one progress sentence" -e "everyday words" -e "not the user's vocabulary" -e "plain everyday words" -e "also sets the line that tells the user which stages run" content tests` y `rg -n -F "report it in one line" content/guidance/global.md` no devuelven nada.
- `rg -n "picks an offered option needs no line" content/skills/hive-core/references/questions.md` muestra la condición de D2-A.
- `rg -n "## Report a running or completed item" content/skills/flow-close/references/reporting.md` devuelve una línea.
- `wc -c content/guidance/global.md` da 19,218 o menos.
- `wc -c content/skills/hive-core/references/questions.md` da 15,984 o menos.
- `go test ./tests/content/...` pasa.

## T2 — Skills: entradas de etapa, punteros, anclas, lista única y reporte de cierre

- [x] Las skills ya no piden líneas de entrada, anuncios del reparto, líneas por ronda ni notas de progreso. Apuntan a la lista única del reporte y enlazan `#split-work`. El paso 7 de flow-close sigue D5-A.

**Closes:** AC1, AC3, AC4, AC7

**Checkpoint:** C2

**Depends on:** T1

**Locations:**

- `content/skills/flow-research/references/delegation.md`: el título "Split and announce work" pasa a "Split work".
- `content/skills/flow-build/SKILL.md:12,38,42,66`
- `content/skills/flow-build/references/iterate-and-integrate.md:9,15`
- `content/skills/flow-verify/SKILL.md:14,18,33,65`
- `content/skills/flow-research/SKILL.md:26,48`
- `content/skills/flow-deploy/SKILL.md:14,18,31`
- `content/skills/flow-walkthrough/SKILL.md:24`
- `content/skills/flow-plan/SKILL.md:57,67`
- `content/skills/flow-security/SKILL.md:14`
- `content/skills/security-advisory/SKILL.md:19`
- `content/skills/flow-close/SKILL.md`: el paso 7 y las excepciones de la pregunta de cierre.
- `content/skills/unattended-delegation/SKILL.md:10,20`

**Execution:** main thread. Las ediciones son cortas y dependen de la redacción de T1.

**Test approach:** check: las búsquedas de AC1, AC3 y AC4, y `go test ./tests/content/...`.

**Changes:**

- Aplicar los apartados de [design.md](design.md) "Entradas de etapa", "Rondas iterativas", "Reparto", "Punteros" y "Reporte de cierre".
- En `flow-verify:65` e `iterate-and-integrate.md:15`, la lista de campos pasa a ser un puntero; la regla de colocación de `flow-verify:65` se queda.
- `unattended-delegation:20` apunta a la condición "a work item completes" de `reporting.md`.

**Verification:**

- `rg -n -F -e "announce the split" -e "announced before its first edit" -e "At entry, say in one line" -e "one by one at entry" -e "Report each round in one line" -e "Say in one line the candidate" -e "in the progress note" -e "progress update" -e "opening with the in-progress item" -e "split-and-announce-work" -e "Split and announce" -e "Before progress reports" -e "Before progress or completion reports" -e "a cleanup line stating" -e "still carries its cleanup label" content` no devuelve nada.
- `rg -n "not in a message of its own" content/skills/flow-verify/SKILL.md` devuelve una línea.
- `rg -n "what was verified" content/skills/flow-verify/SKILL.md content/skills/flow-build/references/iterate-and-integrate.md` no devuelve la lista; solo queda el puntero a `reporting.md`.
- `go test ./tests/content/...` pasa: `skill_size_test`, `skill_links_test`, `voices_test` y `reading_footprint_test`.

## T3 — Sitio y documentación de arquitectura

- [x] Las páginas del sitio y `instruction-resources.md` describen el comportamiento nuevo.
  Ronda 1 rechazada, AC8: `build.md:34` decía que el agente "tells you how it split the work". Se corrigió en el commit e1b187e5, y la nueva verificación dio AC8 cumplido.

**Closes:** AC8

**Checkpoint:** C2

**Depends on:** T1

**Locations:**

- `site/src/content/docs/flows/build.md:6,20,39,42,43,75,90`
- `site/src/content/docs/flows/verify.md:22,31,41`
- `site/src/content/docs/flows/research.md:52`
- `site/src/content/docs/flows/plan.md:75`
- `site/src/content/docs/flows/close.md:9,49,80`
- `_support/docs/architecture/instruction-resources.md:17`

**Execution:** main thread. Son pocas páginas y describen la redacción de T1 y T2.

**Test approach:** check: la búsqueda de AC8 y el build del sitio.

**Changes:** describir el comportamiento, sin copiar las reglas:

- el agente no manda avisos de progreso;
- al esperar, dice en una oración qué sigue corriendo;
- la línea de contexto abre las decisiones, las validaciones, las entregas, los cierres y los bloqueos;
- el reporte de cierre tiene la forma de D5-A;
- los chequeos saltados se nombran en el reporte;
- la ruta del primer render va en el reporte final.

**Verification:**

- `rg -n -i -e "every progress or decision message" -e "answers in one line too" -e "says in one line whether verify runs" -e "progress updates" -e "single-sentence note" -e "progress note" -e "in one line with its measurement" -e "names each one at the start" -e "what was cleaned" -e "how it split the work" site/src/content/docs _support/docs/architecture/instruction-resources.md` no devuelve nada.
- `npm --prefix site run build` (`astro build`) termina sin errores.

## Verificación y revisión

**Puntos de control:**

- C1 — T1 — dos tareas posteriores dependen de su redacción canónica.
- C2 — T2 y T3 — el resto de los hogares y la documentación.

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite completa | Candidato final, local | `go test ./...` pasa | Cubierto por la implementación |
| Revisión dedicada | Candidato local, antes del push | `/code-review` de Claude Code; hallazgos resueltos o refutados con evidencia | D7-A (2026-10-29) |
| Prueba local en vivo | — | No aplica: es contenido de guía, sin runtime. Su comportamiento real solo se observa en sesiones futuras, que quedan en observación (los pilotos están pausados por `AGENTS.md`). | — |

## Estado de la revisión y progreso

**Revisión del plan.** Un revisor `hive-review-plan` de consistencia de la guía (el único dominio afectado) revisó la revisión `d020c1be`/`7590f390`/`8e8e9e61`. Tardó unos 195 s y usó unos 147 k tokens. Dio 7 hallazgos bloqueantes y 5 no bloqueantes. Se aplicaron todos sin preguntar, porque ninguno cambia el alcance, un criterio ni una decisión del usuario:

- **B1:** `global.md:63` va en T1 y `build.md:75` en T3.
- **B2:** `flow-verify:33` nombra los saltos solo en el reporte.
- **B3:** T3 agrega `close.md:49`, `build.md:90` y `verify.md:31,41`, y amplía su `rg`.
- **B4:** AC4 pasa a T2, con su búsqueda.
- **B5:** el cambio de título y los cuatro enlaces quedan juntos en T2.
- **B6:** solo la lista de campos se vuelve puntero; la regla de colocación se queda.
- **B7:** los pasos conservan su número y la oración de D3-A aplica al turno de notificación que termina esperando.
- **N1–N5:** el tamaño de `questions.md` y el ejemplo de `:118`; los dos disparadores del puntero a `work-item-boundary`; la identidad del deploy en el reporte; "reports" en `reporting.md:36,78`; el encabezado sin cambios.

No hubo segunda ronda de revisión. Cada corrección aplica la propuesta del revisor tal como la planteó; B7 elige entre dos lecturas, y la elegida es la que sigue D3-A.

**Progreso.**

- T1 se implementó en el commit 7e75b8fa y quedó `[x]`. El verificador `hive-verify-task` (configurado, no observado; unos 104 s y 109 k tokens) dio AC2, AC5, AC6 y AC7 como cumplidos, y `go test ./tests/content/...` pasa.
- Corregí dos búsquedas del plan, sin cambiar ningún criterio. La de frases retiradas encontraba `flow-walkthrough/SKILL.md:48`, que ya estaba en la base y es un aviso de error dentro de una pregunta; ahora "report it in one line" se busca solo en `global.md`. La de AC6 buscaba un texto que el diseño reemplaza.
- Quedan dos observaciones no bloqueantes del verificador para la ronda de `/code-review`: "asks a decision" en `global.md:13` y "only picks" en `questions.md:88`.
- El implementador es el hilo principal (Opus 5.5, observado). El modelo del verificador no está observado.
- T2 se implementó en el commit 92d4bfaa y T3 en el 67b57138; los dos están en `[?]`. En `unattended-delegation:10` y `:20` no hizo falta cambio: `:20` ya apunta a la sección de `reporting.md` que contiene la condición "a work item completes".
- Para el build del sitio enlacé `site/node_modules` al del checkout principal, sin instalar nada; ese enlace se quita al cerrar.
- `go test ./...` pasa sobre el candidato final 39c7bfcb (20 paquetes). `/code-review` (high): la ronda 1 dio 10 hallazgos, corregidos en a6bd6b31; la ronda 2, limitada a los arreglos, dio otros 10, corregidos en 39c7bfcb sin otra revisión. Integrado por el PR #273 (merge 9f25ff36).
- Además corregí `build.md:90` ("in one line") y `research.md:56` (la línea tras cada respuesta, según D2-A), que la búsqueda de T3 no cubría.
