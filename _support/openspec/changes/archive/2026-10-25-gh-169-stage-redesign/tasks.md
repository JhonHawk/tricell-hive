# Tareas

Base: `f4946d63` (`development`, 2026-10-25). Worktree `.claude/worktrees/gh-169`, rama `feat/gh-169-stage-redesign`.

**Convención de las comprobaciones.** Una frase distintiva por cláusula nueva, con conteo 0 en la base y 1 después; las frases movidas dan 0 en su origen y 1 en su destino; lectura del verificador contra [diseño](design.md#enfoque). Cada tarea que mueve texto o enlaces corre `go test ./tests/content/...` (tamaños) y `go test ./tooling/management -run TestRepositoryCatalogueInstructionReferences` (archivos de enlace). Las anclas de encabezado no las comprueba ninguna prueba: cada ancla que se mueve se cuenta con `rg` en `content`, `site` y `_support/docs`.

## T1 — La puerta de UI pasa a `flow-verify`

- [x] `ui-gate.md` y `ui-review-criteria.md` se mueven con `git mv` a `content/skills/flow-verify/references/`; sus enlaces relativos internos y todos los enlaces entrantes, absolutos o relativos, apuntan a la ruta correcta.
Verificado por `hive-verify-task` (~63 s): movimiento y enlaces cumplidos (`git diff -M` 87 % y 100 %, rg sin rutas viejas, unos 30 enlaces resuelven, anclas existen). El verificador marcó AC3 `not met` solo porque el AC incluía la remisión de `flow-walkthrough`, asignada a T2; AC3 se dividió (revisión del plan): AC3 queda con el movimiento y los enlaces, cumplido por ese mismo veredicto, y AC10 con la remisión, que cierra T2. Commit `afebe7d1`.

**Closes:** AC3.
**Checkpoint:** C1.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-walkthrough/references/ui-gate.md`, `ui-review-criteria.md` → `content/skills/flow-verify/references/`; enlaces internos `ui-gate.md:3,19,28,36,38,46`; enlaces entrantes en `content/skills/flow-walkthrough/SKILL.md:8,12-14,38,62,64`, `content/agents/` (`hive-review-ux.md`, `hive-design-ui.md`, `hive-build-frontend.md:11,17`), `content/skills/flow-build/SKILL.md`, `content/skills/flow-build/references/verification.md`, `content/skills/flow-build/references/browser-automation.md:3`, `content/skills/flow-plan/SKILL.md:49`, `content/skills/flow-plan/references/ui-planning.md`, `delivery-decisions.md:36`, `plan-format.md:136`, `content/skills/flow-close/SKILL.md`, `content/skills/flow-close/references/incidental-findings.md:21`, `site/src/content/docs/`, `_support/docs/architecture/instruction-resources.md:18`.
**Execution:** delegada a `hive-write-docs`: movimiento y enlaces, sin cambiar otro texto.
**Test approach:** check.
**Verification:** `rg -n "flow-walkthrough/references/(ui-gate|ui-review-criteria)|\]\(references/(ui-gate|ui-review-criteria)" content site _support/docs` no deja enlaces a las rutas viejas desde fuera de `flow-verify`; `git diff -M --stat` muestra los dos archivos como renombrados (los únicos cambios internos son los enlaces); `go test ./tests/content/...` pasa. `TestRepositoryCatalogueInstructionReferences` no puede pasar hasta que exista `flow-verify/SKILL.md` («missing owning skill»): esa prueba de los enlaces de T1 pasa a la Verification de T2 (revisión del 2026-10-25, sin cambiar criterios).

## T2 — `flow-verify` y el final de `flow-build`

- [x] Existe `flow-verify/SKILL.md` con contrato, decisión de nivel (incluida la omisión como aceptación, D21-A), secuencia de verificación e integración (D24-A); `review-and-gates.md` recibe los párrafos de revisión, prueba en vivo y entrega humana; `flow-build` termina en el candidato e integra solo cuando verify se omite; `flow-walkthrough` remite a `flow-verify` para la puerta y las rondas de corrección; las frases de propiedad que el diseño lista quedan reescritas.
Verificado por `hive-verify-task` (~123 s): AC1, AC2, AC4 y AC10 cumplidos (`flow-verify/SKILL.md:10,18-33,39-46,62`, `review-and-gates.md`, `flow-build/SKILL.md:12,62,64`, `flow-walkthrough/SKILL.md:3,8,12-13`). Arreglo posterior al commit `3dbea246` (frase ambigua de `flow-build` sobre revisión de UI y prueba en vivo), re-verificado en el momento (~71 s): AC1 cumplido. Commits `97344024`, `3dbea246`.

**Closes:** AC1, AC2, AC4, AC10.
**Checkpoint:** C2.
**Depends on:** T1.
**Locations:** `content/skills/flow-verify/SKILL.md` (nuevo); `content/skills/flow-verify/references/review-and-gates.md` (nuevo); `content/skills/flow-build/SKILL.md:3,34,36,52-69`; `content/skills/flow-build/references/verification.md:34-35,42-46,50-66`; `content/skills/flow-walkthrough/SKILL.md:3,8,12,60-66`; `content/skills/flow-verify/references/ui-gate.md:3,19,36`; `content/skills/flow-plan/references/delivery-decisions.md:9`; `content/skills/flow-plan/references/plan-revision.md:11,13`; `content/skills/flow-plan/references/plan-format.md:47`; `content/skills/flow-build/references/plan-execution.md:40`; `content/agents/review/hive-review-ux.md:14`; `content/skills/flow-close/SKILL.md` (enlace de la puerta); `content/guidance/global.md:71,78`; `content/skills/flow-plan/references/delivery-decisions.md:24`; `content/skills/flow-walkthrough/SKILL.md:13-14`.
**Execution:** delegada a `hive-write-docs`, con `design.md` («Etapas y contratos», «Nivel», «Integración», «Movimientos») como fuente.
**Test approach:** check.
**Verification:** la secuencia numerada de validación da 0 en `flow-build/SKILL.md` y 1 en `flow-verify/SKILL.md`; las frases de refutación de P0/P1, mapa de severidades y tope de re-revisión dan 0 en `verification.md` y 1 en `review-and-gates.md`; «Correction rounds and markers» da 0 en `flow-walkthrough/SKILL.md`; dan 1 en `flow-verify/SKILL.md` la frase de omisión (cambio mecánico o pedido), la de aceptación sin verificar de lo omitido, la tabla de nivel y la de integración con `git-workflow`; `flow-build/SKILL.md` tiene la frase de integración cuando verify se omite; «verify results, and close» da 0 en la descripción de `flow-build`; `TestRepositoryCatalogueInstructionReferences` pasa también para los enlaces que movió T1; las frases de propiedad de la lista del diseño no nombran a `flow-walkthrough` como dueña de la puerta ni «the order flow-build sets»; `global.md:71` y `delivery-decisions.md:24` distinguen el chequeo aceptado sin verificar de una puerta pendiente y dicen que D21-A no exime CI, protección de ramas ni revisión exigida; `global.md:78` nombra la etapa que integra; las dos pruebas de la convención pasan.

## T3 — `flow-deploy`

- [x] Existe `flow-deploy/SKILL.md` con contrato (etapa aparte, después de integrar) y procedimiento, y sus referencias reciben los chequeos de despliegue y los arreglos urgentes; los demás archivos remiten a ella.
Verificado por `hive-verify-task` (~96 s): AC5 cumplido (`flow-deploy/SKILL.md:7-24`, `deployment-checks.md`, `urgent-fixes.md`, sin `#urgent-production-fixes`). Commit `513a9bc1`. Arreglo posterior al commit por D27-A (`2edda714`, la descripción dice que solo abre y mergea el PR de promoción), re-verificado (~53 s): AC5 cumplido. Observación P3 sin cambio: `urgent-fixes.md:7` no dice si el PR del hotfix lo abre deploy o la etapa que integra.

**Closes:** AC5.
**Checkpoint:** C3.
**Depends on:** T2 (comparte `flow-build/SKILL.md` y `verification.md`).
**Locations:** `content/skills/flow-deploy/SKILL.md`, `references/deployment-checks.md`, `references/urgent-fixes.md` (nuevos); `content/skills/flow-build/references/verification.md:36-37,60,74,77-86`; `content/skills/flow-plan/references/delivery-decisions.md:40-46`; `content/skills/git-workflow/SKILL.md:44` (ancla `#urgent-production-fixes`); `content/skills/flow-build/SKILL.md:3,18`; `content/skills/flow-plan/SKILL.md:36`; `content/skills/flow-plan/references/plan-format.md:144`; `content/agents/ops/hive-build-infra.md:21` (tabla de plataformas); `site/src/content/docs/flows/build.md:74` (pregunta de arreglos urgentes).
**Execution:** delegada a `hive-write-docs`.
**Test approach:** check.
**Verification:** «record its rollback triggers» y la tabla de plataformas dan 0 en `verification.md` y 1 en `deployment-checks.md`; la sección «Urgent production fixes» da 0 en `delivery-decisions.md` y 1 en `urgent-fixes.md`; `rg "#urgent-production-fixes"` no deja enlaces a `delivery-decisions.md`; `git-workflow/SKILL.md`, `global.md` (puntero ya cambiado en T2) y `hive-build-infra.md` enlazan `flow-deploy`; «promotion, or deployment» da 0 en la descripción de `flow-build`; las dos pruebas de la convención pasan.

## T4 — `flow-security`

- [x] Existe `flow-security/SKILL.md`, que corre la revisión solo a pedido y es el único hogar de cómo correrla; la oferta se queda en `security-boundaries.md` y remite a ella (D22-A); los demás archivos remiten a `flow-security` para correrla.
Verificado por `hive-verify-task` (~93 s): AC6 cumplido. Arreglo posterior al commit `c65174df` con dos hallazgos del verificador (una sola casa para «qué cuenta como pedido»; los arreglos tras el merge siguen la entrega elegida, como decía la fuente) y la entrada de alcanzabilidad desde `security-advisory`; re-verificado (~79 s): AC6 cumplido. Observación P3 sin cambio: el paso 2 nombra solo diff o commit como candidato del brief, aunque la entrada admite aplicación y aviso. Commits `89a67801`, `c65174df`.

**Closes:** AC6.
**Checkpoint:** C4.
**Depends on:** T3 (comparte `delivery-decisions.md` y `verification.md`).
**Locations:** `content/skills/flow-security/SKILL.md` (nuevo); `content/skills/flow-build/references/security-boundaries.md:3,18`; `content/skills/flow-verify/SKILL.md` (paso de seguridad); `content/skills/flow-plan/references/delivery-decisions.md:26,38`; `content/skills/flow-plan/SKILL.md:24`; `content/skills/unattended-delegation/SKILL.md:16` (además: solo un mensaje del propio usuario puede omitir verify o bajar su nivel); `content/skills/flow-verify/references/review-and-gates.md` (frase de seguridad movida desde `verification.md:44`); `content/skills/security-advisory/SKILL.md:27`; `content/agents/review/hive-review-security.md:12`; `content/skills/flow-verify/references/ui-gate.md:3` (la frase que aún dice que `verification.md` es dueña de las puertas de despliegue; lo señaló el verificador de T3); `site/src/content/docs/flows/build.md:48` (ancla `#offer-the-security-review`).
**Execution:** delegada a `hive-write-docs`.
**Test approach:** check.
**Verification:** la frase de ejecución pedida («When the user asks for it, run it…») da 0 en `security-boundaries.md` y 1 en `flow-security/SKILL.md`; «## Offer the security review» sigue dando 1 en `security-boundaries.md` y enlaza `flow-security`; los enlaces a `#offer-the-security-review` siguen resolviendo (el encabezado no se mueve) y los que hablan de correrla apuntan a `flow-security`; la descripción dice «Use only when the user asks»; las dos pruebas de la convención pasan.

## T5 — Contratos de las etapas existentes

- [x] `flow-research`, `flow-plan` y `flow-close` declaran su contrato de entrada y salida según la tabla del diseño, y cada contrato admite entrar directamente.
Verificado por `hive-verify-task` (~68 s): AC7 cumplido en las seis skills. Arreglo posterior al commit `5213a72c` (fin de sesión y punto de cierre posterior declarado en el contrato de `flow-close`), re-verificado (~57 s): AC7 cumplido. Commits `60958a10`, `5213a72c`. Pendiente para T7: `site/.../flows/deploy.md:6` dice sin condición que deploy no es parte del cierre.

**Closes:** AC7.
**Checkpoint:** C5.
**Depends on:** T2, T3, T4 (comparte `flow-plan/SKILL.md` con T3 y T4).
**Locations:** `content/skills/flow-research/SKILL.md`, `content/skills/flow-plan/SKILL.md`, `content/skills/flow-close/SKILL.md` (un párrafo «Entry and exit» al inicio, de dos o tres frases).
**Execution:** delegada a `hive-write-docs`.
**Test approach:** check.
**Changes:** `flow-close` (525 bytes libres) y `flow-plan` (685 libres antes de T3 y T4) recortan frases repetidas por puntero si el párrafo no cabe; la tarea dice cuáles en el progreso.
**Verification:** «Entry and exit» da 1 en cada uno de los seis `SKILL.md` de etapa; los tres archivos siguen ≤ 16 000 bytes; las dos pruebas de la convención pasan.

## T6 — Pruebas de estructura

- [x] Las pruebas que fijan la estructura incluyen las etapas nuevas con estos cambios concretos y pasan.
Verificado por `hive-verify-task` (~45 s): AC8 cumplido (`protected.go:56`, `flows_test.go:221,224`, `reading_footprint_test.go:57,80-81`); los fixtures de pilotos pausados no cambian, por ser entradas de pilotos que nadie corre. Commit `0c8fa0c8`.

**Closes:** AC8.
**Checkpoint:** C5.
**Depends on:** T2, T3, T4.
**Locations:** `tests/content/reading_footprint_test.go:52-81` (agregar `flow-verify/SKILL.md` y `review-and-gates.md` a `cycle` y a `moved`, para que la línea base no busque archivos inexistentes); `tests/pilot/protected.go:56` (agregar las tres skills); `tests/pilot/flows_test.go:221-226` (`TestAllManagedSkillsProtected`: su lista y su conteo); `_support/docs/architecture/instruction-resources.md:73+` (descripción del escenario). Los fixtures de pilotos (`tests/pilot/flows.go:548-550`, `flows_test.go:756`, `tests/fixtures/flows/cases.json`) quedan como están: son entradas de pilotos pausados y cambiarlas no prueba nada; la decisión queda escrita en el progreso.
**Execution:** delegada a `hive-write-tests`.
**Test approach:** check: `go test ./tests/...`.
**Verification:** `protected.go` y `TestAllManagedSkillsProtected` nombran `flow-verify`, `flow-deploy` y `flow-security`; `reading_footprint_test.go` tiene los dos archivos nuevos en `cycle` y en `moved`; `go test ./tests/...` pasa.

## T7 — Documentación

- [!] README, `llms.txt`, `CONTRIBUTING.md`, el sitio y el changelog describen las etapas nuevas, la seguridad a pedido y `flow-deploy` como etapa aparte.
Ronda 1, AC9: `site/.../flows/build.md:50` atribuía a Build el manejo de hallazgos de la revisión y `:83` enlazaba la oferta de seguridad a `flow-build`; arreglado en `cb0fa7b2`.
Ronda 2, AC9: `site/.../flows/build.md:54` todavía decía que Build despliega a pedido, y `deploy.md:6` negaba sin condición que deploy sea parte del cierre (lo vio el verificador de T5); arreglado en `1c11f890`.
Ronda 3, AC9 (~134 s): quedan `README.md:357` y `deploy.md:39` (deploy fuera del cierre sin la salvedad), `verify.md:42` (sugiere que Verify despliega), `close.md:61` (Build antes de Close; `close.md` no estaba en las Locations) y una tensión entre `flow-deploy` («not for … pull requests or merges») y el PR de promoción que autoriza `delivery-decisions.md:5`. Tercera ronda sin cumplir: va al usuario. D27-A (2026-10-25): se corrigen las cuatro frases y las dos imprecisiones, `close.md` entra en las Locations de T7, y la descripción de `flow-deploy` pasa a decir que solo abre y mergea el PR de promoción (T3 se re-verifica por esa frase); una re-verificación de T7.
Re-verificación tras D27-A (~176 s): los seis puntos cumplidos; dos frases nuevas (`verify.md:27`, validación humana sin condición; `deploy.md:32`, commits atribuidos a Verify). D28-B (2026-10-25): se arreglan (`66ad97d0`, con el matiz del README) y se lanza otra verificación de T7, que cumple AC9 (~181 s). Commits `d0f8f1c6`, `cb0fa7b2`, `1c11f890`, `ddf87ab1`, `66ad97d0`. Tres imprecisiones menores anotadas sin cambio (Verify edita en las correcciones; resúmenes de una línea sin la condición de `hold`).

**Closes:** AC9.
**Checkpoint:** C5.
**Depends on:** T2, T3, T4, T5.
**Locations:** `README.md:309-342` (Mermaid y viñetas), `:415-417`; `llms.txt:29-32`; `CONTRIBUTING.md:76-78`; `site/astro.config.mjs:75-78`; `site/src/content/docs/flows/verify.md`, `deploy.md`, `security.md` (nuevos); `site/src/content/docs/skills.md`, `index.mdx`, `flows/build.md`; `CHANGELOG.md` (`[Unreleased]` → `### Added` y `### Changed`).
**Execution:** delegada a `hive-write-docs`.
**Test approach:** check: los tres scripts del sitio.
**Verification:** el Mermaid y las viñetas del README y `llms.txt` nombran `flow-verify`, `flow-deploy` y `flow-security`; `CONTRIBUTING.md` deja de decir «four main flow guides»; la barra lateral tiene las tres páginas; `pnpm --dir site build`, `pnpm --dir site check` y `pnpm --dir site lint` terminan sin error; `#169` da 1 en `CHANGELOG.md`; que las páginas describan y enlacen sin copiar reglas lo juzga el verificador leyendo (ninguna prueba lo comprueba); los enlaces del sitio apuntan a `blob/master`, como la convención actual, y resolverán tras la promoción.

## Verificación y revisión

Puntos de control (criterio de `plan-format.md`):

- C1 — T1 — cierra porque T2 y T7 dependen de las rutas nuevas.
- C2 — T2 — cierra porque T3, T4, T5, T6 y T7 dependen de `flow-verify`.
- C3 — T3 — cierra porque T4, T5, T6 y T7 dependen de `flow-deploy`.
- C4 — T4 — cierra porque T5, T6 y T7 dependen de `flow-security`.
- C5 — T5, T6, T7 — último punto; T5 y T7 no comparten archivos con T6; T7 depende de T5 y corre después.

| Puerta | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Por tarea | en cada punto de control | `hive-verify-task` por criterio | incluida en la implementación |
| Suite completa | candidato final | `go vet ./...` y `go test ./...` pasan | incluida |
| Sitio | candidato final | `pnpm --dir site build`, `check` y `lint` sin error | incluida |
| Copias instaladas | candidato final, ya commiteado en el worktree | `hive update --dry-run --source .claude/worktrees/gh-169`: las rutas instaladas viejas `flow-walkthrough/references/ui-gate.md` y `ui-review-criteria.md` aparecen en la lista y no están en la nueva versión (la vista previa no marca el retiro; `ownership.go:203-230` lo hace); se anota la versión instalada contra la que corrió | incluida (solo vista previa) |
| UI del sitio | candidato final | según el valor de `UI verification` que se resuelva en la ronda de entrega | pendiente |
| Revisión de código | antes del merge | `/code-review` (declarado), una ronda y una re-revisión | `Review` declarado |

## Estado de la revisión y progreso

- **Revisión del plan, ronda 1** (sobre `f4946d63`), dos `hive-review-plan`:
  - Consistencia de la guía (~167 s, ~137 k tokens): cuatro bloqueantes (integración sin dueño, omitir verify frente a puertas pendientes, enlaces relativos de `ui-gate.md`, frases de propiedad) y cuatro no bloqueantes (ubicaciones, anclas, orden de etapas, dependencias y bytes).
  - Pruebas, distribución y sitio (~154 s, ~120 k tokens): dos bloqueantes (la prueba de enlaces correcta, anclas sin comprobar) y tres no bloqueantes (AC8 concreto, `CONTRIBUTING.md` y scripts del sitio, Mermaid y `llms.txt`); copias instaladas viejas sin verificar.
  - Llevados al usuario: D20/D24 (integra la última etapa que corre), D21-A, D22-A, D23-A. Aplicado el resto por el orquestador.
- **Re-revisión** (los mismos revisores): pruebas y sitio (~37 s) sin bloqueantes, con un ajuste de redacción en la puerta de copias instaladas; consistencia (~43 s) con dos bloqueantes y tres no bloqueantes, todos con una corrección propuesta y aplicada tal como la propuso, sin otra ronda: las frases de `global.md:71`, `delivery-decisions.md:24` y `verification.md:50` distinguen la aceptación sin verificar de una puerta pendiente (T2); se quita «tarea chica» y el salto mecánico informa lo que lo cubre; `global.md:78` nombra la etapa que integra; la delegación desatendida no puede omitir verify; `flow-walkthrough/SKILL.md:13-14` y `project-settings.md:9` quedan anotados.
- Plan listo el 2026-10-25.
- **Pregunta del usuario tras el plan** («qué pasa con el code-review? se queda en build o pasa a verify?»): pasa a `flow-verify`; aclarado en «Integración» del diseño que un `Review` declarado es el mecanismo por defecto, no una exigencia. D26-A: `UI verification` `functional` para el sitio, vigente en la sesión.
- **Segunda pregunta del usuario** («las preguntas de qué se ejecuta … se siguen lanzando desde el diseño del plan verdad?»): sí; aclarado en «Integración» del diseño (plan, o entrada de build sin plan; `flow-verify` solo pregunta lo que falte cuando se entra directo).
- **Build.** T1 escrita por un `hive-write-docs` (~35 s, ~83 k tokens) y commiteada al reportarla (`afebe7d1`); además corrigió el enlace de `plan-execution.md`, fuera de las Locations. Revisión: la prueba de referencias del catálogo falla con «missing owning skill» hasta que T2 cree `flow-verify/SKILL.md`, así que pasa a la Verification de T2.
- **T2** escrita por un `hive-write-docs` (~255 s, ~166 k tokens), commiteada al reportarla (`97344024`). El orquestador agregó «Entry and exit» a `flow-build` (AC7 lo nombra y T5 no lo cubría). `global.md` crece +308 bytes (líneas 71 y 78), a declarar en el PR. Pendientes para T3/T4: el puntero de despliegue de `global.md:78` y `verification.md:60`; la fila de seguridad de la tabla de nivel apunta a «Offer the security review» hasta T4.
- **T3** escrita por un `hive-write-docs` (~145 s, ~117 k tokens), commiteada al reportarla; `global.md` +1 byte.
- **T4** escrita por un `hive-write-docs` (~58 s, ~94 k tokens), commiteada al reportarla.
- **T6** escrita por un `hive-write-tests` (~41 s, ~81 k tokens), commiteada al reportarla; espera el cierre de C5. La tabla de bytes fechada de `instruction-resources.md` (medida 2026-10-23) queda desactualizada al sumar dos archivos al ciclo: se re-mide en la suite final.
- **T5** escrita por un `hive-write-docs` (~25 s, ~82 k tokens), sin recortes; commiteada al reportarla. `flow-close` 15 890 y `flow-plan` 15 728 bytes.
- **T7** escrita por un `hive-write-docs` (~88 s, ~120 k tokens); `pnpm --dir site build`, `check` y `lint` pasan; commiteada al reportarla. C5 alcanzado: verificadores de T5, T6 y T7 en paralelo (no comparten archivos).
- **Puertas finales sobre `1c11f890`:** `go vet ./...` y `go test ./...` pasan; `pnpm --dir site build`, `check` y `lint` pasan; `hive update --dry-run --source .claude/worktrees/gh-169` lista `~/.agents/skills/flow-walkthrough/references/ui-gate.md` y `ui-review-criteria.md`, que no están en la versión nueva (se retiran), junto a los archivos nuevos de `flow-verify`, `flow-deploy` y `flow-security`; corrió contra la instalación refrescada desde `f4946d63`. Pendientes: puerta de UI `functional` del sitio y `/code-review`.
- Observación P3 sin cambio: `flow-build/SKILL.md:26` todavía lista «deployment of existing code» entre los cambios que leen referencias situacionales; es lectura de más, no un error de comportamiento.
- **PR #177, `/code-review` ronda 1** (~104 s): 9 hallazgos P2, todos aplicados en `d479004b`: `global.md:78` distingue los commits de la obra (`flow-build`) de PR y merge (etapa que integra); `global.md:71` y `flow-verify` distinguen omitir antes de correr de rechazar una puerta; `delivery-decisions.md` remite a `flow-verify` en vez de repetir la regla; la espera de CI de una promoción sigue la regla de `verification.md`; la caminata funcional desplegada vive entera en su subsección; el vocabulario de estados pasa a `verification.md#report-check-status`; `flow-verify` y `flow-deploy` remiten a `flow-close#when-to-run-it`; enlace del P3; `flow-walkthrough:51`. `global.md` queda en 38 338 bytes (+472 sobre la base). Re-revisión (~117 s, la única): 10 hallazgos P2, todos arreglos chicos dentro del cambio y sin cambio de contrato, aplicados sin otra ronda en `5294c970` (entre ellos, el despliegue que dispara el merge de integración vuelve a chequear identidad servida y salud). Suite completa y build del sitio pasan sobre `5294c970`. Integrado por PR #177 (`9b8ce042`).
