# Tareas

**Base:** `aa283530` (2026-10-26). Rama `feat/gh-200-unattended-preauthorization`, worktree `.claude/worktrees/gh-200`.

**Convención de las comprobaciones:** `rg -F -c "<frase>" <archivo>` da `1` después del cambio y nada en la base.

## T1 — Vencimiento acotado y anuncio visible (A, C)

- [x] El primer párrafo de la skill y la frase del «me voy» en `global.md` mantienen la delegación ante feedback, recursos, correcciones y preguntas sobre el trabajo, y fijan qué gana cuando un mensaje cae en las dos listas. La activación y su estado se ven en el texto.

**Closes:** AC1, AC4.
**Checkpoint:** C1.
**Depends on:** ninguna.
**Locations:** `content/skills/unattended-delegation/SKILL.md:8` y `:10`, `content/guidance/global.md:35`.
**Execution:** hilo principal; son tres frases y el encargo sería más largo que el cambio.
**Test approach:** check.
- `rg -F "reports something new" content/` y `rg -F "gives a new instruction" content/skills/unattended-delegation/`: sin coincidencias.
- En la skill, una coincidencia de cada frase:
  - «adds information, a resource, feedback or a correction within the objective and scope»
  - «an answer to a question this delegation asked»
  - «When a message fits both lists»
  - «Every other message ends it»
  - «state in visible text, not only in your reasoning»
  - «names the delegation's state» y «or in the first line of a message that takes no context line»
  - «says that it ended and why»
- En `global.md`, una coincidencia de cada frase:
  - «adds information, a resource, feedback or a correction within the objective and scope; a message that fits both lists ends it»
  - «asks how the work is going or about the work»

**Changes:** la redacción de [design.md](design.md#redacción-propuesta): párrafos 1 y 2 de la skill y la frase de `global.md:35`.
**Verification:**
- Los comandos del enfoque de prueba dan lo esperado.
- Al leer la skill y el bullet global, coinciden en los casos que mantienen la delegación, en los que la terminan y en la regla de precedencia.
- Ninguna frase toca «Effects» ni la lista de fuentes.

## T2 — Ronda de preautorización al activarse (B)

- [x] La skill tiene la sección `## Pre-authorization round` y su excepción en «Decisions». `questions.md` excluye la ronda de la tarjeta y la incluye entre los casos en texto. `verification.md`, `delivery-decisions.md` y `flow-build/SKILL.md` remiten a ella.
Ronda 2: AC2 cumplido; AC3 no cumplido: `global.md:20` aún mandaba la ronda a la tarjeta en lectura literal. Arreglo de ronda 3 en `7be672a6`. Ronda 3: AC3 cumplido.

**Closes:** AC2, AC3.
**Checkpoint:** C2.
**Depends on:** T1 (mismo archivo; la sección se escribe sobre el texto de T1).
**Locations:**
- `content/skills/unattended-delegation/SKILL.md` (sección nueva antes de `## Decisions`, y la oración de `:16`);
- `content/skills/flow-plan/references/questions.md:47` y `:60`;
- `content/skills/flow-build/references/verification.md:27`;
- `content/skills/flow-plan/references/delivery-decisions.md:5`;
- `content/skills/flow-build/SKILL.md:24`.

**Execution:** hilo principal; un párrafo y cinco frases.
**Test approach:** check.
- `rg -n "^## Pre-authorization round" content/skills/unattended-delegation/SKILL.md` da una coincidencia, y `rg -n "^## Gate selection and timing" content/skills/flow-build/references/verification.md` también. Es la comprobación manual de las anclas: ninguna prueba las valida (`tooling/management/references.go:12-13`).
- En la skill, una coincidencia de cada frase:
  - «and the questions of the pre-authorization round above, which only the user answers» (dice «above» porque la sección quedó antes de «Decisions»; el diseño decía «below»)
  - «Ask the missing ones in one round in the activation message»
  - «never through a tool that blocks your turn until the answer»
  - «then keep working»
  - «A delivery point and the effects it authorizes take effect only once every question that point depends on is answered» y «a reply that changes one of them answers this round»
  - «each critical effect with its exact target and ceiling»
  - «run no dedicated review without a review answer»
  - «with the full suite when it runs, without commit or publication» y «stay pending gates, so the work is not reported complete»
  - «names the dedicated review and each independent UI child that did not run as not run» y «lists the unanswered questions with the same IDs»
- En `questions.md`, una coincidencia de cada frase: «and the question is not the round that» y «whose pause line adds that the work continues meanwhile without commit or publication».
- Una coincidencia de cada puntero:
  - «When an explicit delegation or the user's leaving message activates unattended delegation before that round» en `verification.md`;
  - «When an explicit delegation or the user's leaving message activates unattended delegation before this round» en `flow-build/SKILL.md` y en `delivery-decisions.md`;
- `wc -c content/skills/unattended-delegation/SKILL.md content/skills/flow-build/SKILL.md`: cada uno da 16 000 o menos.

**Changes:** la redacción de [design.md](design.md#redacción-propuesta): sección nueva, excepción en «Decisions», condiciones de `questions.md` y punteros.
**Verification:**
- Los comandos del enfoque de prueba dan lo esperado.
- `go test ./tooling/management/... ./tests/content/...` en verde: enlaces relativos y tamaño de las skills.
- La lectura de la skill confirma tres cosas:
  - «Decisions» ya no manda decidir las preguntas de la ronda;
  - la sección nombra las dos reglas de `questions.md` que cambia;
  - un modelo literal no puede elegir la tarjeta para esta ronda.

## T3 — Sitio y changelog

- [x] `index.mdx`, `skills.md`, `flows/build.md` y `CHANGELOG.md` describen el vencimiento acotado y la ronda de preautorización.

**Closes:** AC5.
**Checkpoint:** C2.
**Depends on:** T1, T2.
**Locations:** `site/src/content/docs/index.mdx:97`, `site/src/content/docs/skills.md:30`, `site/src/content/docs/flows/build.md:48`, `CHANGELOG.md` (`## [Unreleased]`, `### Changed`).
**Execution:** hilo principal; depende del texto final de T1 y T2.
**Test approach:** check, con las frases que el implementador escribe y anota aquí antes de verificar.
- `rg -F "anything else hands control back to you" site/`: sin coincidencias.
- Cada frase anotada da una coincidencia en su archivo y ninguna en la base.
- Frases anotadas: «or adds a resource, feedback or a correction within the goal does not end that» y «it asks them at once without stopping» (`index.mdx`), «feedback within the goal keeps it going» (`skills.md`), «the agent asks it in text right away and keeps working» (`flows/build.md`), «Unattended delegation no longer ends when you add a resource» (`CHANGELOG.md`).

**Changes:** ver [design.md](design.md#redacción-propuesta), «Sitio» y el puntero de `flows/build.md`.
**Verification:**
- Los comandos del enfoque de prueba dan lo esperado.
- `npm run build` en `site/` termina sin errores.

## Verificación y revisión

**Checkpoints:**
- C1 — T1. Lo cierra T1 porque T2 y T3 dependen de él.
- C2 — T2 y T3. Lo cierra T3, la última tarea.

**Comprobación de tamaño:** `wc -c content/guidance/global.md` antes y después, para el PR.

| Puerta | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| `hive-verify-task` | Al cerrar C1 y C2, en el worktree | Veredicto por cada AC que cierra el checkpoint | Rutina del build |
| Suite local | Candidato final | `go vet ./...` y `go test ./...` en verde; `npm run build` en `site/` | Rutina del build |
| Revisión dedicada | Candidato local, antes del push | `/code-review` de Claude Code; hallazgos corregidos o refutados con evidencia | D5-A |
| Lugares revisados | Antes del PR | El PR lista los lugares que encuentra `rg -i "unattended\|leaving\|first edit\|pause line\|context line\|second report"` en `content/`, `site/src` y `_support/docs/architecture`, además de `tests/pilot` | Regla de `AGENTS.md` |

No hay verificación in vivo: el cambio es solo de guía y los pilotos de comportamiento siguen en pausa (`AGENTS.md`, Measurement).

## Revisión del plan y avance

**Revisión del plan.** La revisión `e0daf9a80d24` la hizo un `hive-review-plan` del dominio de consistencia de la guía.

- **Seis hallazgos bloqueantes**, todos aplicados:
  - la excepción en «Decisions»;
  - los punteros de primera edición en `flow-build`, `delivery-decisions` y el sitio;
  - la excepción al segundo reporte y al «ask once»;
  - la exclusión de la tarjeta en `questions.md:47`;
  - los checkpoints C1/C2;
  - la comprobación manual de anclas.
- **Tres no bloqueantes**, aplicados: la regla de precedencia, la posición del estado en la línea de contexto, y los detalles sin respuesta junto con la copia completa en `global.md` y la búsqueda más amplia.
- **Cobertura no resuelta:** el supuesto de que una tarjeta nativa bloquea el turno en cada host. Por eso la regla se escribe por capacidad.

**Avance (build).**

- **T1** (`ca78245a`): AC1 y AC4 se cumplen en la ronda 1. Vuelve a `[?]` porque el arreglo de T2 (`5b46a772`) toca `global.md`, que T1 comparte.
- **T2** (`7e534b3c`): AC2 y AC3 se cumplen en la ronda 1. El verificador encontró dos huecos incidentales que afectan esta obra: la copia de la regla de la tarjeta en `global.md:20` y una remisión a una regla de la puerta de UI que no existe. Se corrigieron en `5b46a772` y T2 se está reverificando.
- **T3** (`1115a772`, `beed3ca2`): AC5 se cumple en las rondas 1 y 2. La ronda 2 corrige «if you are already gone» por «without your answers». Se reverifica por el arreglo de T2.
- **Suite completa** en `1115a772`: `go vet ./...` y `go test ./...` en verde.
- **Modelos:**
  - implementador: hilo principal, `claude-opus-5-5` (observado);
  - verificadores: perfil `verifier` de `hive-verify-task`, configurado y no observado.
- **`global.md`:** 39 046 → 39 380 bytes.

- **`/code-review`** (Claude Code, medium) sobre `7be672a6`: ocho hallazgos, todos aplicados en `b6fe5e94` (T1, T2) y `ece81278` (T3).
  - P1: una sola respuesta de entrega podía publicar sin revisión.
  - P2: el respaldo `hold` sin suite completa; punteros sin condición de fuente; la ronda ignoraba las distinciones de `flow-build`; la fuente explícita omitida; un cambio a una respuesta reutilizada terminaba la delegación.
  - P3: la excepción global no nombraba la skill; el lugar del estado sin línea de contexto.
  - Las frases de comprobación de T1 y T2 se actualizaron a la redacción nueva (revisión del plan durante el build).
- **Suite completa** en `7be672a6`: verde.

- **Reverificación tras `b6fe5e94`/`ece81278`:** T2 (AC2, AC3) y T3 (AC5) se cumplen.
- **`/code-review` acotado** a `7be672a6..ece81278`: nueve hallazgos.
  - Ocho se corrigieron sin otra ronda de revisión, como permite `review-and-gates.md`, en `a297efaa` (T1, T2) y `91e6c7a8` (T3): la excepción de la respuesta reutilizada; las autorizaciones previstas; las verificaciones pendientes sin bajar el nivel de `flow-verify`; los punteros para la ronda sin contestar; la regla global de la tarjeta; la línea de pausa; la cuarta parte de la línea de contexto; y el ancho del CHANGELOG. También se corrigió `skills.md:30`, que vio el verificador de T3.
  - Uno no se corrigió: un trabajo programado declarado sin respuestas de entrega no tiene regla para la ronda. El hueco ya existía en la base, este cambio no lo introduce, y queda en el PR como hallazgo sin confirmar.

- **Cierre de la ambigüedad:** la excepción de la respuesta reutilizada nombra también la regla de precedencia (`2ecc2f82`).
- **Reverificación final:** T1 (AC1, AC4), T2 (AC2, AC3) y T3 (AC5) se cumplen sobre `2ecc2f82` y `91e6c7a8`.
- **Suite completa** en `91e6c7a8`: `go vet ./...` y `go test ./...` en verde, y `npm run build` completo.
- **Entrega:** [PR #201](https://github.com/JhonHawk/tricell-hive/pull/201), merge `5cb8ff6b` en `development`; #200 cerrado. `hive update` aplicado desde `5cb8ff6b`.

**Siguiente paso:** ninguno; registro cerrado.
