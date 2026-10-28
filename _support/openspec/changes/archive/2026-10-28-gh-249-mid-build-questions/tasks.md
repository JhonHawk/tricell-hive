# Tareas — gh-249-mid-build-questions

Base: `3bb1e623` (development), confirmada al iniciar el build el 2026-10-28. Rama `fix/gh-249-mid-build-questions` en `.claude/worktrees/gh-249-mid-build-questions`.

## T1 — El fallo de CI causado por el cambio se arregla sin preguntar

- [x] `verification.md` separa el fallo de infraestructura que vuelve a fallar, que queda pendiente, del fallo causado por el cambio, que se arregla bajo la regla global: con un commit nuevo en el mismo PR o con un PR de seguimiento bajo la misma entrega.

**Closes:** AC1, AC5.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:**
- `content/skills/flow-build/references/verification.md:56`, con el disparador ampliado al run del PR antes del merge;
- `site/src/content/docs/flows/build.md:39`.

**Execution:** hilo principal. Son pocas líneas, y el sitio lo comparten T2 y T3.

**Test approach:** check: los `rg` de abajo y `go test ./tests/content/...`.

**Changes:** según el enfoque 1 de `design.md`.

**Verification:**
- `rg -n "or the failure comes from the change's code, the integration is not complete" content/skills/flow-build/references/verification.md` no encuentra nada.
- `rg -n "follow-up pull request|same pull request" content/skills/flow-build/references/verification.md` encuentra la regla nueva, con su enlace a `delivery-decisions.md`.
- `site/.../build.md:39` menciona el fallo de CI.
- `go test ./tests/content/...` pasa, con el archivo por debajo de 16 000 B.

## T2 — El plan, y el build sin plan, nombran lo que la verificación necesita fuera de esta máquina

- [x] La lista de readiness de `flow-plan/SKILL.md` y las comprobaciones bloqueantes de `plan-review.md` exigen el entorno y los efectos, datos y accesos de cada criterio que se verifica fuera de esta máquina. `foreseen-authorizations.md` hace que el trabajo sin plan los liste desde la petición y su verificación.

**Closes:** AC2, AC3, AC5.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:**
- `content/skills/flow-plan/SKILL.md:44-54`;
- `content/skills/flow-plan/references/plan-review.md:22`;
- `content/skills/flow-build/references/foreseen-authorizations.md:9`;
- `content/skills/flow-plan/references/delivery-decisions.md:7`;
- `site/src/content/docs/flows/plan.md:70`;
- `site/src/content/docs/flows/build.md:50`.

**Execution:** hilo principal. Comparte las páginas del sitio con T1 y T3.

**Test approach:** check: los `rg` de abajo y `go test ./tests/content/...`.

**Changes:** según el enfoque 2 de `design.md`.

**Verification:**
- `rg -n "outside this machine" content/skills/flow-plan/SKILL.md` y `rg -n "verification environment" content/skills/flow-plan/references/plan-review.md` encuentran una línea cada uno. La búsqueda de `plan-review.md` cambió tras la revisión de código, para que la regla tenga una sola casa en `flow-plan`.
- `rg -n "for work without a plan the request, already names" content/skills/flow-build/references/foreseen-authorizations.md` no encuentra nada.
- `rg -n "the plan or request foresees" content/skills/flow-plan/references/delivery-decisions.md` no encuentra nada.
- La lectura de la línea nueva de `foreseen-authorizations.md` muestra que el agente lista lo que la petición y su verificación necesitan antes de preguntar.
- Las páginas Plan y Build mencionan el entorno de verificación.
- `go test ./tests/content/...` pasa.

## T3 — El build asume lo reversible y confirma todo al final (D7-B)

- [x] `questions.md` tiene la condición para decisiones a mitad de build, con las exclusiones, la definición de reversible y la confirmación por modo (D9-A).
  - Si es reversible, la asume como supuesto registrado.
  - Si no lo es pero no bloquea, sigue con el resto.
  - Solo pregunta en el momento si bloquea todo y no es reversible.
  - `questions.md:11`, `flow-build/SKILL.md:20` y `global.md:29` apuntan a la regla.

**Closes:** AC4, AC5.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:**
- `content/skills/hive-core/references/questions.md`, en «Decide whether a question is needed» y en `:11`;
- `content/skills/flow-build/SKILL.md:20`;
- `content/guidance/global.md:29`;
- `content/skills/flow-close/references/reporting.md:84`, con un puntero que dice que un supuesto no es un ítem pendiente;
- `site/src/content/docs/flows/build.md:78`, `:84` y la lista «What it does».

**Execution:** hilo principal.

**Test approach:** check: los `rg` de abajo, la lectura de la condición nueva frente a `unattended-delegation/SKILL.md` y `scope-changes.md` para descartar choques, y `go test ./...`.

**Changes:** según el enfoque 3 de `design.md`.

**Verification:**
- `rg -n "recorded assumption" content/skills/hive-core/references/questions.md` encuentra la condición nueva, con sus tres casos, las exclusiones y la confirmación por modo.
- `rg -n "except during a build" content/skills/hive-core/references/questions.md` encuentra la excepción en `:11`.
- `global.md:29` enlaza la regla, y el PR reporta el tamaño de `global.md`.
- La lectura de la condición muestra:
  - las reglas excluidas, nombradas una por una;
  - que los supuestos van en la lista de decisiones tomadas sin preguntar y se confirman en la ronda de la pregunta de cierre bajo `automatic`, `direct-base` y `hold`;
  - que un supuesto no es un ítem pendiente, y que `reporting.md:84` lo apunta;
  - que un rechazo posterior al cierre es trabajo nuevo.
- `site/.../build.md:78` dice que bajo `automatic` los supuestos se confirman después del merge (N6).
- `rg -n "or ask for the missing decision before dependent edits" content/skills/flow-build/SKILL.md` no encuentra nada.
- `rg -n "asks you for the missing decision before it keeps editing" site/src/content/docs/flows/build.md` no encuentra nada.
- La lectura muestra que la delegación desatendida y los pedidos del usuario quedan fuera de la regla.
- `go test ./tests/content/...` pasa.

## T4 — Changelog

- [x] `CHANGELOG.md` registra el cambio en `[Unreleased] / Changed`, con la indicación de correr `hive update`.

**Depends on:** T1, T2, T3.

**Locations:** `CHANGELOG.md`.

**Execution:** hilo principal.

**Verification:** `rg -n "#249" CHANGELOG.md` encuentra la entrada.

## Verificación y revisión

- C1: T1, T2 y T3. Ninguna tarea depende de otra, y ninguna cambia un contrato público ni datos guardados, así que basta un solo checkpoint al final.

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite completa | candidato final, local | `go test ./...` pasa (unos 50 s) | implementación |
| Revisión dedicada | candidato local, antes del push | `/code-review` de Claude Code sin hallazgos abiertos | D6-A |
| `hive update` | tras el merge a `development` | deploy del commit integrado a los hosts registrados | regla del mantenedor (AGENTS.md) |

No hay UI ni verificación en un entorno desplegado.

## Revisión del plan y progreso

- Revisión del plan, ronda 1: un `hive-review-plan` de consistencia de guía revisó la revisión `711694cd82ae` y tardó unos 190 s, con unos 152 000 tokens. Encontró 8 hallazgos bloqueantes y 3 menores.
  - Aplicados sin preguntar, por ser correcciones convencionales: B1 (exclusiones), B4 (manda `scope-changes.md`), B5, B6, B7, B8, N1, N2 y N3.
  - Llevados al usuario: B2, resuelto con D8-A, y B3, resuelto con D9-A.
  - Re-revisión, la única ronda, con el mismo revisor sobre `0da78aba0f0c`: unos 63 s y 168 000 tokens.
    - Comprobó que B1, B4–B8 y N1–N3 están corregidos y que B2 y B3 quedan resueltos con D8-A y D9-A.
    - Encontró R1, bloqueante: un supuesto sin confirmar impedía cerrar.
    - Encontró N4, N5 y N6.
    - Los cuatro se aplicaron tal como los propuso el revisor, así que no hubo otra ronda.
- Plan listo en la revisión final. No queda ninguna decisión abierta.

## Progreso del build

- **Commits:**
  - T1: `1463a595`;
  - T2: `447a15fa`;
  - T3: `bd796772` y `f609ae7e`;
  - T4: `6a69a72d`;
  - correcciones tras la revisión: `0d86efa7` y `a14add1f`.
- **`hive-verify-task`:** T1, T2 y T3 quedaron en `met`. Se lanzó en su perfil configurado; el modelo no se observó en el lanzamiento. El implementador es claude-opus-5-5, en el hilo principal.
  - T3 tuvo una corrección de redacción en `reporting.md:84`, con una re-verificación que dio `met`.
- **Suite completa:** `go test ./...` pasa; código de salida 0, unos 107 s.
- **`/code-review` medio sobre `6a69a72d`:** 10 hallazgos.
  - Pasaron por `hive-refute-claim` los cuatro que bloquearían el merge. F1 y F4 quedaron refutados. F2 (el arreglo directo bajo `direct-base`) y F3 (la ronda de corrección ante un supuesto rechazado) eran plausibles y se corrigieron en `a14add1f`.
  - Corregidos en `0d86efa7`: el puntero de `global.md`, la ambigüedad de `reporting.md`, la casa única de la regla de entornos (con el CI excluido) y el texto del sitio.
  - Descartado: presentar los supuestos en `flow-verify` y `flow-close`, porque `questions.md` se lee antes de toda validación o pregunta de cierre.
  - Solo reportado: el margen de 22 B de `questions.md`.
  - Hallazgo incidental, fuera de #249: en `verification.md:55`, una falla previa de la base que no bloquea igual retiene el ticket.
- **Re-revisión única** sobre `6a69a72d..a14add1f`: 9 hallazgos, sin uno bloqueante. Los locales se corrigieron sin otra revisión en `ec58364a`:
  - el anclaje de «later rejection»;
  - la regla de entornos legible para el revisor del plan, con su alcance por criterio;
  - una excepción de CI más estrecha;
  - la ruta sin PR;
  - `reporting.md`, sin ambigüedad;
  - el motivo de la regla, restituido.
- **Descartado:** que «such a choice» en `global.md:29` difiera las decisiones no reversibles. Es el diseño D7-B.
- **Re-verificación acotada tras las correcciones:** T1 (AC1), T2 (AC2) y T3 (AC4 y AC5) dieron `met`.
- **Integración:** merge de `development` (#250) en la rama; suite completa en verde; PR #251 integrado con merge commit `f9e2a323`; #249 cerrado. `hive update` aplicó `f9e2a323` a los 7 destinos registrados, y se verificó en `~/.agents/skills`.
- Progreso: nada implementado.
