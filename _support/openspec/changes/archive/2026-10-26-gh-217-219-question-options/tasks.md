# Tareas — gh-217-219-question-options

Base: `26012ce0` (development), confirmada al iniciar el build. Rama `fix/gh-217-219-question-options` en `.claude/worktrees/gh-217-219`.

## T1 — La pregunta de entrega nombra la ruta de cierre del registro en sus opciones (#217)

- [x] El párrafo de la tarjeta y su variante de texto en `delivery-decisions.md` nombran la ruta de cierre del registro del cambio; en texto, la ruta y la publicación de los repositorios de soporte van dentro de cada opción (E2-A).

**Closes:** AC1.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `content/skills/flow-plan/references/delivery-decisions.md:5` y `:16`; `site/src/content/docs/flows/build.md:48`.

**Execution:** hilo principal; es una edición de una frase acoplada a la regla de la línea 22.

**Test approach:** check — `rg -n "closing route" content/skills/flow-plan/references/delivery-decisions.md` y `go test ./tests/content/...`.

**Changes:** según el diseño de #217 en `proposal.md`.

**Verification:** `rg` muestra «closing route» en la línea 16 (frase de la tarjeta; la de texto dice «how it versions the change record»), además de las líneas 5, 22 y 28; `go test ./tests/content/...` pasa (incluye el límite de 16 000 bytes).

## T2 — La pregunta de ruta nunca reduce las opciones a abrir un ticket (#219)

- [x] `flow-research/SKILL.md` y la página de investigación del sitio dicen que cada pregunta de ruta, incluida la repetida, ofrece al menos una alternativa que resuelve el trabajo, además de abrir el ticket y parar.

**Closes:** AC2, AC3, AC4.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `content/skills/flow-research/SKILL.md:46`; `content/guidance/global.md:24`; `site/src/content/docs/flows/research.md:47`; `CHANGELOG.md`.

**Execution:** hilo principal; son dos frases y una entrada del changelog.

**Test approach:** check — `rg -n "never replaces" content/skills/flow-research/SKILL.md`, lectura de la línea del sitio y `go test ./tests/content/...`.

**Changes:** según el diseño de #219 en `proposal.md`; una frase equivalente, en lenguaje llano, en el sitio; una entrada en `CHANGELOG.md` para los dos issues.

**Verification:** `rg` encuentra la invariante en la línea 46 de `SKILL.md`; las líneas 24 de `global.md` y 47 del sitio mencionan la pregunta repetida; `go test ./tests/content/...` pasa e informa el tamaño nuevo de `global.md`.

## Verificación y revisión

- C1 — T1, T2 — ninguna de las dos tiene dependientes ni cambia un contrato persistido; se verifican juntas al final.

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Pruebas de contenido | en C1, worktree | `go test ./tests/content/...` en verde | implementación |
| Suite completa | candidato final, local | `go test ./...` en verde (~50 s) | implementación |
| Búsqueda de consistencia | antes del commit final | `rg` de «route question», «closing route» y «ask it again» en `content/`, `site/`, `_support/docs/architecture/` y `tests/pilot/`; ninguna frase contradice el cambio | implementación |
| Revisión dedicada | candidato local, antes del push | Claude Code `/code-review`, hallazgos resueltos o refutados | D3-A |
| UI | no aplica | no hay interfaz renderizada | — |

Pruebas en vivo: no aplican; el cambio es texto de guía y los pilotos están en pausa.

## Estado y revisión del plan

- Revisión del plan: `hive-review-plan`, dominio de consistencia de guía, sobre la primera versión (78 s, unos 104 000 tokens).
  - Bloqueante, aplicado sin preguntar porque es la propuesta del revisor: en la variante de texto, la ruta de cierre va dentro de cada opción, no en la línea compartida (contradecía `delivery-decisions.md:22`).
  - No bloqueante, aplicado: el falsador de AC3 ahora es «no menciona la pregunta repetida».
  - Decisiones: E1-A (cláusula en `global.md:24`, tras comprobar que `8100a51b` no leyó `flow-research`) y E2-A (corregir `:16` contra `:18`). Las dos aplican la propuesta del propio revisor, así que no hace falta otra ronda.
- Para AC1, `rg` no separa las dos variantes de la línea 16; hay que leer la línea.
- Build: T1 (`7697e84f`) y T2 (`38ca7779`) verificados por `hive-verify-task` (AC1–AC4 met; T1 49 s y unos 88 000 tokens, T2 63 s y unos 91 000 tokens; implementador claude-opus-5-5 observado, verificador configurado, no observado). Suite completa `go test ./...` en verde. `global.md` crece 156 bytes tras la revisión 1 (40024 → 40180).
- Revisión 1 (hallazgos de `/code-review` medium sobre `38ca7779`, 67 s y unos 95 000 tokens): 9 hallazgos. Se aplicaron 8 como correcciones locales en los commits `(T1)` y `(T2)` siguientes:
  - F1, F5, F6 y F7: la invariante de #219 ya no depende de que se ofrezca el ticket, y `global.md` usa la misma forma.
  - F2 y F3: una respuesta reutilizada sin ruta de cierre la pregunta antes del primer edit.
  - F4: la variante de texto lleva una condición.
  - F9: `build.md` describe el cambio.
  - F8 se descarta: `8100a51b` y `1624fa4f` son sesiones independientes, y R4-A es la aprobación del usuario.
  - Los AC no cambian. T1 amplía sus ubicaciones y vuelve a `[?]` (revisión 1); T2 vuelve a `[?]` (revisión 1). Sin otra ronda de revisión del plan ni de código: son correcciones P2, sin cambio de lógica ni de contrato, que aplican la propuesta del revisor.
- Reverificación tras la revisión 1: T1 (AC1) y T2 (AC2–AC4) met (63 s y 52 s, unos 98 000 y 93 000 tokens). Suite completa en verde otra vez.
- Integrado: [PR #220](https://github.com/JhonHawk/tricell-hive/pull/220), merge `b5a4358e`; #217 y #219 cerrados por el PR.
