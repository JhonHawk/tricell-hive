# Tareas

Base: `b7277d82` (`development`, 2026-10-25). Worktree `.claude/worktrees/gh-167`, rama `feat/gh-167-security-review-on-request`.

**Convención de las comprobaciones.** Como en gh-157 y gh-158: una frase distintiva por cláusula nueva, con conteo 0 en la base antes de escribirla y 1 después, más la lectura del verificador; las frases viejas que deben desaparecer dan 0 después.

## T1 — Lanzamientos automáticos en build, entrega, walkthrough y delegación desatendida

- [x] `flow-build`, `delivery-decisions.md`, `flow-walkthrough` y `unattended-delegation` dejan de lanzar `hive-review-security` sin pedido y lo ofrecen como fija D7-A.
Ronda 1, AC4: la consolidación perdió la condición de que las correcciones toquen un límite en la puerta del candidato aceptado.
Verificado por `hive-verify-task` (~69 s): AC1, AC2, AC3, AC4 y AC8 cumplidos (`flow-build/SKILL.md:58,61`, `verification.md:44`, `delivery-decisions.md:26`, `flow-walkthrough/SKILL.md:62`, `unattended-delegation/SKILL.md:16`). Commit `c43113d3`.

**Closes:** AC1, AC2, AC3, AC4, AC8.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-build/SKILL.md:54-61` (paso 3 y una oración nueva después de la lista); `content/skills/flow-plan/references/delivery-decisions.md:26` (oración del `Review` declarado y pregunta de sí o no); `content/skills/flow-walkthrough/SKILL.md:62`; `content/skills/unattended-delegation/SKILL.md:16`; `content/agents/review/hive-review-code.md:20` («decide on» → «offer»); `content/skills/flow-build/references/verification.md:44` (revisión del 2026-10-25, ver progreso); `content/skills/flow-build/references/security-boundaries.md` (sección «Offer the security review», hogar único desde D11-A).
**Execution:** delegada a `hive-write-docs`: las decisiones están fijadas y las ubicaciones no se pisan con otras tareas.
**Test approach:** check: conteo de frases, lectura y tamaño de los `SKILL.md`.
**Changes:**
- `flow-build` paso 3: correr `hive-review-security` solo cuando el usuario lo pidió; si el cambio cruza un límite de [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md) y nadie lo pidió, el mensaje de validación nombra los límites y ofrece la revisión. Quitar el puntero «within the authorization that the verification reference sets», que no apunta a ninguna regla de seguridad.
- `flow-build`, oración fuera de la lista (la lista solo rige con parada de validación): sin esa parada, o cuando un límite aparece después de la ronda de entrega, por ejemplo en el veredicto de un revisor de código, el siguiente informe nombra los límites y ofrece la revisión sin detener la integración. `SKILL.md` sigue ≤ 16 000 bytes.
- `delivery-decisions.md:26`: quitar que un `Review` declarado, incluido `none`, autoriza `hive-review-security`. Agregar que, cuando el cambio planeado cruza un límite, la misma ronda hace una pregunta aparte de sí o no que ofrece la revisión de seguridad, también cuando `Delivery` y `Review` están declarados; la línea de una frase sigue nombrando los valores declarados; el sí es el pedido.
- `flow-walkthrough/SKILL.md:62`: repetir la revisión de seguridad solo cuando el usuario la pidió y las correcciones tocan un límite; si una corrección cruza un límite y nadie la pidió, el mensaje de validación la ofrece.
- `unattended-delegation/SKILL.md:16`: agregar a la excepción que una revisión de seguridad ofrecida no cuenta como pedida mientras el usuario no la elija; esa pregunta sin contestar cuenta como no y no impide activar la delegación desatendida.
- `hive-review-code.md:20`: «so the parent can decide on a security review» → «so the parent can offer a security review».
- Buscar en `content/`, `site/` y `_support/docs/architecture/` cualquier otra frase que lance o dé por automática la revisión de seguridad y ajustarla o anotarla en el progreso. `CHANGELOG.md:333,362-363` es historia publicada y no se toca.
**Verification:** «also authorizes `hive-review-security`» da 0 en `delivery-decisions.md`; «stays at that reviewer's point» y «within the authorization that the [verification]» dan 0 en `flow-build/SKILL.md`; «authorization a declared `Review` value gives» da 0 y «Offer and run `hive-review-security` only as» da 1 en `verification.md`; «## Offer the security review» da 1 en `security-boundaries.md` y contiene la oferta en una línea, el criterio de recomendación, la repetición solo ante un límite nuevo y la ejecución pedida antes de validar o del push o merge; `flow-build/SKILL.md`, `delivery-decisions.md`, `flow-walkthrough/SKILL.md` y `verification.md` remiten a `#offer-the-security-review` sin repetir la regla; el diff no agrega «yes-or-no question» (la de `questions.md:23` ya está en la base); la frase de la excepción da 1 en `unattended-delegation/SKILL.md`; «decide on a security review» da 0 y «offer a security review» da 1 en `hive-review-code.md`; `rg -n "hive-review-security" content site` no muestra otro lanzamiento sin pedido; `go test ./tests/content/...` pasa.

## T2 — Descripción del rol y barrido de avisos

- [x] La descripción de `hive-review-security` dice que se usa solo a pedido del usuario, y el paso 5 de `security-advisory` pasa a ofrecer la evaluación de alcance en el informe.
Verificado por `hive-verify-task` (~35 s): AC5 y AC6 cumplidos (`hive-review-security.md:3`, `security-advisory/SKILL.md:27,36`). Commit `6320de65`.

**Closes:** AC5, AC6.
**Depends on:** ninguna.
**Locations:** `content/agents/review/hive-review-security.md:3`; `content/skills/security-advisory/SKILL.md:27` (paso 5) y su sección «Report» (`reachability, when step 5 assessed it`).
**Execution:** delegada a `hive-write-docs`, junto con T1 en el mismo hijo (textos cortos con la misma decisión).
**Test approach:** check.
**Changes:** descripción: «Use only when the user asks for a security review …», conservando la lista de límites como lo que revisa; paso 5: cuando la exposición depende de cómo el código llega a la ruta vulnerable, el informe lo marca y ofrece la evaluación con `hive-review-security`, que corre solo si el usuario la pide; la línea del informe sobre el alcance queda condicionada a que se haya pedido.
**Verification:** «Use when a change touches a security boundary» da 0 y la frase nueva da 1 en el rol; «launch a `hive-review-security` child» da 0 en `security-advisory/SKILL.md` y la oferta nueva da 1; lectura.

## T3 — Sitio y changelog

- [x] `site/src/content/docs/flows/build.md` explica que la revisión de seguridad corre solo a pedido y cuándo se ofrece, con la oferta de una línea aunque `Delivery` y `Review` estén declarados (D11-A); `CHANGELOG.md` lo registra en `[Unreleased]` con `(#167)`.
Verificado por `hive-verify-task` (~55 s): AC7 cumplido (`CHANGELOG.md:83-90`, `site/.../flows/build.md:48,80`). El orquestador aplicó sus tres observaciones de redacción (orden de la frase en `:48`, gramática en `:80`, «role description» en el changelog) antes del commit `6c8698bb`.

**Closes:** AC7.
**Depends on:** T1 (el sitio describe el texto ya escrito).
**Locations:** `site/src/content/docs/flows/build.md:48` (párrafo de la ronda de entrega y revisión, «states them in one line before the first edit instead of asking»); `site/src/content/docs/flows/build.md:80` («the security review when the change crosses a security boundary», que describe el lanzamiento automático; lo señaló el verificador de T1); `CHANGELOG.md` `[Unreleased]` → `### Changed`.
**Execution:** delegada a `hive-write-docs`.
**Test approach:** check.
**Verification:** `#167` da 1 en `CHANGELOG.md`; la frase nueva da 1 en el párrafo de la línea 48; lectura contra el texto de T1; la página no copia reglas de la skill, enlaza a ella.

## Verificación y revisión

**Suite completa:** `go vet ./...` y `go test ./...` pasan sobre `6c8698bb` y de nuevo sobre el candidato final `560c2424` (2026-10-25).


| Puerta | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Por tarea | tras cada tarea, en el worktree | `hive-verify-task` por criterio | incluida en la implementación |
| Suite completa | candidato final | `go vet ./...` y `go test ./...` pasan | incluida |
| Revisión de código | antes del merge | `/code-review` (declarado), con un tope de una re-revisión | `Review` declarado |
| CI | PR a `development` | checks del repositorio en verde sobre la base actual | `Delivery: automatic` |

Sin recorrido en vivo ni de UI: el cambio es texto de guía; el sitio no cambia de maquetación.

## Estado de la revisión y progreso

- **Revisión del plan, ronda 1** (revisión `41f585bad461` / `97c16749d36c`), dominio consistencia de la guía, un `hive-review-plan` (~120 s, ~120 k tokens). Tres bloqueantes y uno parcial, todos aceptados y aplicados por el orquestador como decisiones menores que se siguen de D7-A y de la supernota:
  - B1: la oferta para `automatic` quedaba dentro de una lista que solo rige con parada de validación; pasa a una oración propia (AC3).
  - B2: la oferta no cabía como opción de la pregunta de revisión de código; pasa a pregunta propia de sí o no en la misma ronda (AC2).
  - B3: la delegación desatendida podía contestarse la oferta; nuevo AC8.
  - Parcial: la corrección del walkthrough que cruza un límite ofrece la revisión (AC4).
  - No bloqueantes aplicados: quitar el puntero sobrante del paso 3, «offer» en `hive-review-code.md:20`, la excepción en `site/.../build.md:48`.
- **Re-revisión** (revisión `b5afc652e3f7` / `f7d7373a753e`, el mismo revisor, ~33 s): B1, B2, B3 y el parcial resueltos, sin bloqueantes. Aplicados dos no bloqueantes: la viñeta y el conteo de `hive-review-code.md:20`, y que la pregunta de seguridad sin contestar cuenta como no y no frena la delegación desatendida. Anotado sin cambio: sin plan, el límite se detecta antes de la primera edición solo como respaldo (`flow-build/SKILL.md:22`); la oración de AC3 lo cubre.
- Plan listo el 2026-10-25.
- **Build.** T1 y T2 escritos por un `hive-write-docs` (~33 s, ~89 k tokens). Revisión del plan durante el build: el hijo encontró que `verification.md:44` también lanzaba `hive-review-security` (por cruzar un límite, por el veredicto del revisor de código o con el `Review` declarado); ninguna revisión del plan lo había visto. Se agregó a las Locations y a la Verification de T1, sin cambiar criterios (AC1 y AC3 ya lo cubren), y el orquestador escribió la frase nueva.
- **`/code-review` del PR #170, ronda 1** (~96 s): 9 hallazgos P2 sin verificación propia; todos aplicados en `286cd040` (pregunta de seguridad en el camino sin Git, activación desatendida, registro de la respuesta, oferta tras aceptar el walkthrough, revisión pedida ante un límite posterior, criterio de recomendación, trabajo que sigue, referente de «These steps», reanudación del barrido de avisos).
- **Re-revisión** (~116 s, la única permitida): 10 hallazgos. El central, una contradicción con `questions.md` (la pregunta de sí o no no podía esperar sin pausar) y la oferta repetida en cinco lugares, se llevó al usuario: D11-A, la oferta es una línea sin pregunta, con un solo hogar en `security-boundaries.md` «Offer the security review». Aplicados como arreglos chicos: revisión pedida antes del push o merge sin parada de validación, pedido después del merge, límites nuevos dentro del mismo cambio, `global.md` vuelve a su texto original (sin cambio de tamaño), cuerpo del PR corregido. Reportado sin cambio: sin respuesta, un cambio de alto riesgo se integra sin revisión de seguridad; es la consecuencia de la decisión «solo a pedido» y de D7-A.
- **Revisión del plan durante el build:** AC2 y AC3 reescritos por D11-A; T1 y T3 vuelven a `[?]` para re-verificar AC1–AC4, AC8 y AC7 sobre el texto consolidado.
- **Re-verificación tras D11-A:** T3 cumple AC7 (~55 s). T1 cumple AC1–AC3 y AC8, pero AC4 no se cumple (~114 s): al consolidar se perdió la condición de que las correcciones toquen un límite. Ronda de arreglo 1: la condición vuelve a `security-boundaries.md`; además, el paso 3 de `flow-build` y `site/.../build.md:80` ofrecen solo ante un límite que ninguna oferta anterior nombró, y AC4, AC7 y la línea de T3 se ajustan a D11-A.
- **Ronda 1 verificada:** AC4 cumplido (~31 s, `security-boundaries.md:18`); AC7 cumplido con la redacción de D11-A (~49 s). Commits `9bfb95aa` (T1) y `560c2424` (T3).
