# Tareas

Base: `2bdf9357` (2026-10-25). Rama `feat/gh-151-question-economy`, worktree `.claude/worktrees/gh-151` (`automatic`, D7-A).

**Convención de las comprobaciones.** Cada frase se cuenta con `rg -F -c "<frase>" <archivo>`, que debe dar `1` después del cambio. Las frases marcadas «ausente» deben dar 0 en todo `content/`. El 2026-10-25 se verificó que ninguna de las frases nuevas existe en `content/`, `site/src/`, `_support/docs/` ni `tests/`. La redacción completa está en [design.md](design.md#redacción-propuesta).

## T1 — Autorización de efectos por clase (R1, D3-A, D9-A)

- [x] Existe la referencia nueva y la guía global, `verification.md` y `flow-build` remiten a ella.
Verificado por `hive-verify-task`: AC1–AC6 y AC12 cumplidos. Antes del commit se aplicó una observación del verificador: en `global.md:28`, crítico pasa a ser «irreversible or hard-to-reverse», igual que en la referencia. No cambia ninguna frase de comprobación. Commit `3b3349da`. `global.md` pasa de 35 750 a 36 469 bytes.
Desvío del diseño: `tests/content/reading_footprint_test.go` queda sin cambios. Esa prueba lee cada fuente en commits históricos (`d5164f71`), donde la referencia nueva no existe, y falla con «exit status 128». El diseño suponía que la prueba solo medía.

**Closes:** AC1, AC2, AC3, AC4, AC5, AC6, AC12.
**Depends on:** ninguna.
**Locations:**
- `content/skills/flow-plan/references/effect-authorization.md` (nuevo).
- `content/skills/unattended-delegation/SKILL.md`, sección «Effects» (D10).
- `content/guidance/global.md:28`.
- `content/skills/flow-build/references/verification.md:26`.
- `content/skills/flow-build/SKILL.md:20`.
- `content/skills/flow-plan/references/delivery-decisions.md:5`.
- `content/skills/flow-plan/references/project-settings.md:9`.
- `_support/docs/architecture/instruction-resources.md` (lista de hogares canónicos).
- `tests/content/reading_footprint_test.go` (escenario `cycle`).

**Execution:** hilo principal. La redacción ya está escrita en el diseño y un encargo sería más largo que el cambio. Antes de crear el archivo, leer `_support/docs/architecture/instruction-resources.md` (AGENTS.md, «Guidance structure»).
**Test approach:** check.
- `test -f content/skills/flow-plan/references/effect-authorization.md`, y en ese archivo una coincidencia de cada frase:
  - «| Critical |»
  - «| Non-production |»
  - «as [project settings](project-settings.md) defines the `Environments` setting»
  - «first look for existing accounts or records»
  - «the search takes time»
  - «Do not split the set into stages»
  - «changing DNS»
  - «data the task did not create in any non-local database»
  - «needs no new question»
  - «of that kind, on those targets»
  - «does not cover a new session»
  - «sets how critical effects and the other classes are handled instead of asking»
- `content/guidance/global.md`:
  - «Before a write to shared or production data, confirm the exact records» ausente.
  - Una coincidencia de «Before asking to authorize an effect beyond this machine, read [effect authorization]».
  - Una coincidencia de «local effects take no question».
  - Una coincidencia de «one before the last of the repository's `Environments`».
  - Una coincidencia de «each critical effect takes its own question, even when the user's request names it».
  - Una coincidencia de «changing or deleting data the task did not create».
- `content/skills/flow-build/references/verification.md`:
  - «decisions they do not foresee are still asked when they arise» ausente.
  - Una coincidencia de «every effect beyond this machine that the work and its verification will need in each environment its scope reaches».
  - Una coincidencia de «ask when the work in that environment starts, before its first effect».
  - Una coincidencia de «which covers only its delivery chain».
- `content/skills/flow-plan/references/delivery-decisions.md`: una coincidencia de «which that reference asks when the work there starts».
- `content/skills/flow-plan/references/project-settings.md`: una coincidencia de «the last one is production and every earlier one is non-production» y otra de «`local` is never one of them».
  - Una coincidencia de «together with every other effect that same step needs».
  - Una coincidencia de «ask for them in that gate's question».
- `content/skills/unattended-delegation/SKILL.md` (D10-A): una coincidencia de cada frase: «is taken only when it was authorized in advance for that exact target and ceiling», «report it with its target, cost and revert» y «that spends no money or paid quota».
- `content/skills/flow-build/SKILL.md`: una coincidencia de «review question, the foreseen authorizations».
- `_support/docs/architecture/instruction-resources.md`: una coincidencia de «flow-plan/references/effect-authorization.md».
- `wc -c content/skills/flow-build/SKILL.md` da 16 000 o menos.

**Changes:** redacción de [design.md](design.md#redacción-propuesta).
**Verification:**
- Los comandos del enfoque de prueba dan lo esperado.
- `go test ./tests/content/... ./tooling/management/...` en verde: enlaces `skill:` y relativos, tamaño de las skills y la prueba de huella.
- Leer la referencia confirma seis cosas:
  - Las cuatro clases cubren el caso del socio y la sucursal en QA, el acceso temporal en producción y la migración de datos existentes.
  - Con `local → gamma → prod` (facti), gamma es no productivo y prod es producción. Sin `Environments`, un entorno de nombre ambiguo se trata como producción.
  - Reutilizar una cuenta de QA existente como insumo cae en desarrollo o QA, y modificar sus campos cae en crítica.
  - La clase crítica coincide punto por punto con D9-A.
  - Ninguna frase de la referencia fija el momento de la pregunta, que vive en `verification.md`, ni repite la cobertura que `verification.md` enlaza.
  - En modo desatendido, `unattended-delegation` enuncia la regla completa y la referencia y `global.md:28` solo remiten a ella.
  - `global.md:28` enuncia el invariante completo y no copia pasos.

## T2 — Agrupar cadenas y reutilizar decisiones vigentes (R4)

- [x] `questions.md` agrupa las autorizaciones de un mismo objetivo y comprueba las decisiones vigentes antes de preguntar.
Verificado por `hive-verify-task` (AC7 cumplido). El verificador señaló que `reporting.md:106` («Ask once for a set of actions») repite la regla de agrupar. Decisión tomada sin preguntar, por la regla de un solo hogar: en T5, `reporting.md:106` remite a `questions.md`, y `verification.md:26` enlaza `questions.md` en vez de `reporting.md`.

**Closes:** AC7.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-plan/references/questions.md`, sección «Decide whether a question is needed» (`:11`, `:19`, `:23`).
**Execution:** hilo principal; son dos condiciones y una frase.
**Test approach:** check.
- «Reuse explicit decisions already in force» y «such as the route, close or ticket question» ausentes.
- Una coincidencia de cada frase:
  - «persistent memory from an earlier session»
  - «carries over from an earlier session only as»
  - «ask them as one question that lists the whole set»
  - «and a production deployment keep their own question»
  - «Silence, preselection, and a dismissed or timed-out question grant nothing»

**Changes:** redacción de [design.md](design.md#redacción-propuesta).
**Verification:** los comandos dan lo esperado. Leer la sección confirma dos cosas: hay un solo orden (comprobar lo vigente, decidir si preguntar, agrupar) y no queda ninguna regla repetida con la condición de preguntas independientes.

## T3 — Decisiones menores de la revisión del plan (R2)

- [x] `plan-review.md` aplica y anota los hallazgos menores, y el resumen del plan los muestra.
Verificado por `hive-verify-task` (AC8 cumplido). Commit `868118b0`.

**Closes:** AC8.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-plan/references/plan-review.md:37`, `content/skills/flow-plan/SKILL.md` (párrafo del resumen del plan).
**Execution:** hilo principal; son dos frases.
**Test approach:** check.
- `plan-review.md`: una coincidencia de «Apply without asking a finding whose fix has a conventional default».
- `flow-plan/SKILL.md`: una coincidencia de «the minor calls taken without asking with their reasons».
- `wc -c content/skills/flow-plan/SKILL.md` da 16 000 o menos.

**Changes:** redacción de [design.md](design.md#redacción-propuesta).
**Verification:** los comandos dan lo esperado. Leer `plan-review.md:37–39` confirma que la segunda ronda sigue mandando al usuario solo los huecos materiales.

## T4 — Vigencia de las condiciones y PRs de corrección (D5-A)

- [x] `delivery-decisions.md` cubre el PR de corrección del mismo ticket y la vigencia de una condición.
Verificado por `hive-verify-task` (AC9 cumplido). Se aplicó su observación: «the chosen delivery contract, with its exceptions above, covers its merge» en lugar de «the mode's merge rule covers it», para que también valga la ruta de revisión humana.

**Closes:** AC9.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-plan/references/delivery-decisions.md:24`.
**Execution:** hilo principal; es un párrafo.
**Test approach:** check. Una coincidencia de cada frase: «belongs to the same delivery» y «lasts only for the scope it states».
**Changes:** redacción de [design.md](design.md#redacción-propuesta).
**Verification:** los comandos dan lo esperado. Leer el párrafo confirma que siguen vigentes la excepción de revisión humana y la restricción explícita de no hacer merge.

## T5 — Ticket dentro de la pregunta de ruta (D6-A)

- [x] La investigación ofrece el ticket dentro de las opciones de ruta, y la guía global y `reporting.md` coinciden.
Verificado por `hive-verify-task` (AC10 cumplido). El verificador hizo tres observaciones:
- `global.md:23` resumía las opciones de ruta sin la de abrir el ticket y parar. Se añadió: «plus opening a ticket and stopping when no ticket covers the work».
- La salvedad de `reporting.md:106` («an action whose risk or target needs its own decision») se deja así, porque es más amplia que la de `questions.md` y no la contradice.
- `go test ./tooling/management/...` no comprueba anclas: el ancla `#deliver-the-finding` se confirmó leyendo el archivo.

**Closes:** AC10.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-research/SKILL.md:38` y `:42`, `content/guidance/global.md:76`, `content/skills/flow-close/references/reporting.md:98` y `:106`, `content/skills/flow-build/references/verification.md:26`. Las dos últimas se añadieron tras verificar T2, para que la regla de agrupar tenga un solo hogar.
**Execution:** hilo principal; son cuatro frases que deben coincidir.
**Test approach:** check.
- Ausentes: «as its own question with an ID» (en `flow-research/SKILL.md`) y «end with the ticket question and the route question».
- Una coincidencia de cada frase:
  - «Ask no separate ticket question»
  - «one more option opens the ticket and stops»
  - «whose options say whether they open a ticket» (en `global.md`)
  - «except opening the ticket that a research route option names» (en `reporting.md`)
  - «as [questions](../../flow-plan/references/questions.md) sets for the steps of one objective», una vez en `reporting.md` y otra en `verification.md`
- Ausente en `verification.md`: «"Ask once for a set of actions" rule in [reporting]».

**Changes:** redacción de [design.md](design.md#redacción-propuesta).
**Verification:**
- Los comandos dan lo esperado.
- `rg -n -i "ticket question" content/` no deja ninguna instrucción que pida la pregunta de ticket aparte en la investigación.
- `go test ./tooling/management/...` en verde, porque el ancla `#deliver-the-finding` debe resolver.

## T6 — Sitio y changelog

- [x] Las páginas de research, plan, build y close y el changelog describen el comportamiento nuevo.
Verificado por `hive-verify-task` (AC11 cumplido). Antes del commit se aplicaron dos de sus observaciones en `build.md:45`: se corrigió la errata «…set» con un enlace repetido, y se mencionó la opción de buscar insumos existentes. `pnpm check` y `pnpm build` volvieron a pasar. Commit `e8b6160f`.
Su tercera observación afecta a T1: `flow-close/SKILL.md:48` todavía metía «cloud resources or DNS» en una sola respuesta. Se corrige en un commit adicional `(T1)` y pasa por una re-verificación de T1.
Implementado por `hive-write-docs`. Frases anotadas; cada una falta en `2bdf9357` y aparece una vez:
- `research.md`: «so there is no separate ticket question».
- `plan.md`:
  - «is applied without asking and noted in the plan's progress section» (:34)
  - «and the minor calls it took without asking with their reasons» (:64)
  - «Effects beyond your machine are listed by class» (:67)
- `build.md`:
  - «The effects beyond your machine that the work and its verification need are classified»
  - «is covered by the delivery you already chose, without a new question»
- `close.md`: «grouped by effect class».
- `CHANGELOG.md`: «Fewer questions in the flows».

**Closes:** AC11.
**Depends on:** T1, T2, T3, T4, T5.
**Locations:**
- `site/src/content/docs/flows/research.md:47`: ticket dentro de la ruta.
- `site/src/content/docs/flows/plan.md:34`: los hallazgos menores de los revisores se aplican y se anotan.
- `site/src/content/docs/flows/plan.md:64`: el resumen muestra las decisiones menores.
- `site/src/content/docs/flows/plan.md:67`: autorizaciones previstas, con las clases.
- `site/src/content/docs/flows/build.md:45`: clases de efecto.
- `site/src/content/docs/flows/build.md:67-68`: PR de corrección y duración de una condición.
- `site/src/content/docs/flows/close.md:49`: los permisos que faltan siguen las clases.
- `CHANGELOG.md`, entrada nueva en `## [Unreleased]`, sin editar la de #144.

**Execution:** delegado a `hive-write-docs`. Las fuentes están fijadas en T1–T5 y en el diseño, y el hilo principal sigue con la conciliación. El encargo incluye las rutas de AGENTS.md («README and llms.txt»: describir y enlazar, no copiar reglas) y de este registro.
**Test approach:** check.
- `rg -F "A ticket offer can join the same round as its own numbered question" site/` da 0.
- Una coincidencia de cada frase, que el implementador anota aquí antes de la verificación, ausente en la base:
  - `research.md`: la ruta dice si abre un ticket.
  - `plan.md`: una frase por cada una de sus tres líneas.
  - `build.md`: una frase sobre las clases y otra sobre el PR de corrección.
  - `close.md`: una frase.
  - `CHANGELOG.md`: una frase.
- En `site/`, `pnpm install --frozen-lockfile` si el worktree es nuevo, y luego `pnpm check` y `pnpm build` en verde.

**Changes:**
- `build.md:45`: conservar la frase de prueba costosa, que sigue en `global.md:68`, y añadir al lado las clases de efecto con su enlace.
- `close.md:49`: la opción de continuar sigue pidiendo una vez los permisos que faltan. Se añade que se piden por clase de efecto y que los críticos van aparte.
- Los enlaces del sitio apuntan a `blob/master`, donde la referencia nueva no existe hasta la próxima release. Hasta entonces se enlaza `verification.md#gate-selection-and-timing`, que ya existe en `master` y remite a la referencia.

**Verification:** los comandos dan lo esperado. Leer las cuatro páginas confirma que describen y enlazan, y que no copian.

## Verificación y revisión

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | Después de cada tarea, en el worktree | `hive-verify-task` con los criterios de la tarea; veredicto por criterio | Parte de la implementación |
| Suite completa | Candidato final, local, sobre la base actual; el repositorio no tiene CI (`CONTRIBUTING.md:14`) | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` en verde; en `site/`, `pnpm check` y `pnpm build` en verde | Parte de la implementación |
| Tamaño de `global.md` | Candidato final | `wc -c content/guidance/global.md` frente a 35 750 en la base, declarado en el PR | Parte de la implementación |
| Revisión de código | Sobre el PR a `development` | `/code-review` de Claude Code (D8-A) a esfuerzo medio o más; se corrigen los hallazgos según `verification.md` | D8-A |
| Validación local | No aplica | El cambio no tiene efecto renderizado ni runtime; la evidencia es el diff y las comprobaciones | — |

No hay autorizaciones previsibles fuera de las de la entrega: el cambio no toca entornos remotos, datos ni gasto.

## Entrega

`automatic` (D7-A), 2026-10-25:
1. Rama de trabajo desde `development` en un worktree.
2. Commits por tarea, push y PR a `development`.
3. `/code-review` sobre el PR, y merge cuando la suite local está en verde sobre la base actual y sus hallazgos quedaron resueltos o refutados. No hay CI que esperar.
4. Mover #151 a cerrado al integrarse.
5. Con la marca de mantenedor presente, `hive update` desde el checkout principal y reporte del commit desplegado (AGENTS.md, «Refreshing a local installation»).

El registro se versiona una vez al cierre con un PR pequeño sin revisión dedicada, que hago yo. No dispara ningún despliegue.

## Estado de la revisión y progreso

**Revisión de código (2026-10-25).** Corrió el `/code-review` de Claude Code (medio) sobre `942d6331` y devolvió 10 hallazgos sin verificar.
- `hive-refute-claim` revisó los tres más graves:
  - El 1 (en modo desatendido, lo crítico autorizado de antemano chocaba con «nunca lo irreversible») es plausible. Decisión tomada sin preguntar: lo irreversible sigue prohibido en modo desatendido, como dice `global.md:34`, y la autorización previa solo desbloquea lo crítico reversible. Se revierte con una línea si el usuario quiere lo contrario.
  - El 2 (la memoria podía saltarse preguntas) es menor. Se limitó a decisiones del usuario sobre este trabajo.
  - El 5 (una condición sin alcance declarado) queda refutado, aunque igual se añadió que una condición sin fin dura hasta que el usuario la levanta.
- Corregidos: el 3 (en una cadena, lo que se pregunta aparte va en la misma ronda), el 4 (cómo revertir la configuración en no productivo), el 6 (el PR de corrección hereda la revisión), el 8 (un registro que se queda en el entorno) y el 9 (el invariante global sin `Environments`).
- Descartados:
  - El 7: el procedimiento de dividir tarjetas de `questions.md` ya cubre el límite de opciones.
  - El 10: `build.md` resume las clases en una frase y enlaza, sin copiar la tabla.
- Commits `84d1b067` (T1), `9093936a` (T2) y `b0e3994d` (T4). `tests/content`, `tooling/management` y `unittest` en verde.
- Re-revisión única del diff de esas correcciones: 8 hallazgos.
  - Cinco eran locales y sin cambio de contrato, y se corrigieron sin otra ronda:
    - cambiar registros existentes sigue siendo crítico;
    - el informe final lista los registros que se quedan;
    - una condición sin fin sobrevive al traspaso de sesión;
    - un PR de corrección conserva la ruta de revisión humana;
    - un despliegue a producción ya nombrado no lleva pregunta aparte.
  - Van en commits adicionales `(T1)`, `(T2)` y `(T4)`. La frase de T2 «and a production deployment keep their own question» pasa a «and a production deployment the user has not already named keep their own question».
  - Los otros tres (1, 3 y 7) son uno solo: gastar dinero es irreversible y choca con el «never permits irreversible» de `global.md:34`. El usuario lo resolvió con D13-A, que dejaron alineado los commits `(T1)` y `(T6)`.
- Suite final sobre el candidato, local: `go vet ./...`, `go test ./...` y `unittest` (64 pruebas) en verde; `pnpm check` (0 errores) y `pnpm build` (10 páginas) en verde. La base estaba al día, y el merge de [PR #152](https://github.com/JhonHawk/tricell-hive/pull/152) dio `4f06cc28`. `hive update` aplicó el commit `4f06cc28` en los siete hosts registrados. `global.md`: 35 750 → 36 769 bytes.
- Modelos: implementación en el hilo principal (Opus 5.5, observado). `hive-verify-task`, `hive-write-docs` y `hive-refute-claim` según su perfil configurado, no observado.

**Revisión del plan, ronda 1 (2026-10-25, sobre el borrador sin commit)**

Hubo dos revisores `hive-review-plan`: guía y reglas de T1–T5 (unos 113 k tokens, 108 s) y sitio, changelog y gates de T6 (unos 103 k tokens, 70 s).

Hallazgos aceptados y aplicados:
- La cobertura de un crítico se repreguntaba siempre. Ahora solo se repregunta si no fue autorizado para ese destino exacto.
- «Clase» tenía dos sentidos. Ahora es «kind of effect» dentro de producción, y la respuesta cubre lo listado, de ese tipo y en esos destinos.
- La agrupación de `questions.md` no excluía los críticos ni un despliegue a producción. Ahora sí.
- La memoria podía pasar autorizaciones entre sesiones. Ahora de la memoria solo vienen elecciones de ruta, alcance o diseño.
- La reutilización de QA chocaba con «already existed». Ahora dice «data the task did not create», y usar insumos existentes cae en desarrollo o QA.
- El momento de la pregunta tenía dos hogares. La pregunta del gate pasa a `verification.md` y la referencia solo remite.
- `questions.md:15` todavía nombraba la pregunta de ticket.
- AC4: el invariante ahora nombra la clase local y dice «changing or deleting».
- El repositorio no tiene CI, así que el merge depende de la suite local.
- Faltaban páginas: `plan.md:34` y `:64` y `build.md:67-68`. Se quitó `plan.md:73`.
- `build.md:45` conserva la definición de prueba costosa.
- El gate suma `unittest`, `pnpm check` y la instalación con lockfile.
- Las comprobaciones de T6 ahora usan frases fijas.
- El sitio enlaza a `master`, donde la referencia todavía no existe.

Desestimado: «shared» sin definir en `global.md:68`. La prueba costosa sigue siendo un gate y sus efectos van en la misma pregunta, como dice el diseño. No hace falta definirlo para las clases.

El usuario resolvió D10-A, sobre los efectos críticos en modo desatendido.

**Re-revisión (única ronda; mismo revisor de guía, unos 123 k tokens, 37 s).** Las siete correcciones quedaron bien. Encontró un hallazgo nuevo: D10-A tenía tres hogares. Se aplicó tal como lo propuso el revisor: el hogar completo es la sección Effects de `unattended-delegation`, y la referencia y `global.md:28` remiten a ella. Se ajustaron AC12 y las frases de T1. También se corrigió una cuenta menor («cuatro» por «seis»). Como son correcciones que propuso el propio revisor, se concilian sin otra ronda.

Cobertura: la guía y las reglas (T1–T5), y el sitio, el changelog y los gates (T6). Ninguna corrección queda sin revisar.

**Revisión del usuario (2026-10-25, respuesta a D11).**
- La clase «Development or QA» pasa a «Non-production», definida por `Environments`: el último entorno es producción y los anteriores no lo son. Sin ese ajuste, un nombre ambiguo cuenta como producción.
- `verification.md:26` pide los efectos de cada entorno que alcanza el alcance. Cuando el trabajo llega a un entorno no cubierto, se piden todos al empezar esa etapa.
- Afecta a AC1, AC2 y AC5, a las frases de T1 y a la frase de clases de T6 en `build.md`. No hay tareas hechas que reabrir.
- Re-revisión propia: la hizo el mismo revisor (unos 129 k tokens, 31 s) y encontró dos hallazgos más uno menor, que se aplicaron tal como los propuso:
  - Una promoción a un entorno nuevo cubre solo su cadena de entrega. Los efectos de verificación de ese entorno se piden en la ronda de inicio, y `delivery-decisions.md:5` remite a ella.
  - El significado de `Environments` pasa a `project-settings.md`, donde `local` nunca cuenta, y la referencia solo remite.
  - «stage» se cambió por «the work in that environment».
- Como son sus propias correcciones, se concilian sin otra ronda. Plan listo.
