# Menos preguntas a mitad de build: fallo de CI, efectos previstos y decisiones que pueden esperar al final

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el [PR #251](https://github.com/JhonHawk/tricell-hive/pull/251) (`f9e2a323`) |
| Tracker · GitHub Issues | • [#249 — Preguntas evitables a mitad de build: fallo de CI, efectos de entorno no previstos y decisiones que podían esperar al final](https://github.com/JhonHawk/tricell-hive/issues/249) |
| Git | automatic · PR a `development` y merge · registro en un PR de cierre pequeño que integra el agente · sin parada humana |
| Verificación | comprobaciones `rg` por criterio · `go test ./...` · `/code-review` de Claude Code |
| Siguiente paso | ninguno; registro archivado |

## Solicitud

> La premisa es la sigueinte: 
>
> - Y lo podemos ver en multiples sesiones: Mucho tiempo de espera por preguntas que surgen a mitad de un build
> - En build solo deberian salir preguntas de super alto impacto, pero ha habido preguntas como "Hubo errores en el CI, me autorizas otro PR para solucionarlo?" o "No lo medi en QA, me autorizas crear data dummie?", la primera es innecesaria y la segunda aunque genuina, debio ser parte del diseño del plan 
>
> Analizalo y expande el research si es necesario

> Busca en todos los hosts, ademas de que solo te dije dos ejemplos, hay más similares

> Aunque eso si, fue antes de los ajustes de hoy, sin embargo no estoy seguro de que sea un tema resuelto

> Aqui importante, no es solo si la pregunta es genuina o no, si no si la lanza en un momento indicado, en research o plan, se tiene la atención del usuario, en build no, normalmente el usuario espera revisar hasta que ya este acabado 
> - Sobre el tema de --admin, es relativo al usuario, en mi caso yo tengo admin de todas las organizaciones en las que estoy, pero y en las que no? sin embargo mientras si lo sea, es ruido y no creo que deba ser una regla, si no algo que recupera de memoria al trabajar segun el usuario (no se puede poner como regla en el repo)

> esto ultimo, anotemoslo como en observación porsi llegamos a subir el nivel de pausar a D7-A

Respuestas en las tarjetas (2026-10-28):
- **Ruta:** D1-C (seguir investigando) y D2-A (planear G1, G2 y G4; aprueba adoptar las correcciones sin esperar la recurrencia de dos sesiones).
- **G4:** D3-A (solo promociones) quedó sin efecto por el mensaje sobre `--admin`, y G4 salió del alcance.
- **Diseño:** D4-A (G2 también para trabajo sin plan) y D7-B (asumir lo reversible y confirmar al final).
- **Entrega y revisión:** D5-A (automatic) y D6-A (`/code-review` de Claude Code).
- **Tras la revisión del plan:** D8-A (una decisión de contrato interna y no publicada se asume, y `global.md:29` gana un puntero) y D9-A (bajo `automatic` y `direct-base`, los supuestos se confirman en el reporte final, después del merge).

## Objetivo

Hoy los builds se detienen a preguntar cosas que ya estaban cubiertas o que se podían prever, en un momento en que el usuario no está atento: deja el build corriendo y revisa al final. Una revisión de sesiones reales encontró tres causas que siguen en la guía después de #152 y #224:

- el texto que el agente lee mientras espera el CI manda reportar como pendiente el fallo causado por el cambio;
- el plan solo nombra los entornos y los efectos de los gates sustanciales, así que los datos de prueba, los accesos y los efectos de verificación en QA salen a mitad del build;
- una decisión genuina que aparece a mitad del build detiene el trabajo aunque sea reversible o no bloquee el resto.

El cambio hace que el fallo de CI causado por el cambio se arregle sin preguntar, y que el plan, o el build sin plan antes de su primera edición, nombre lo que la verificación necesita fuera de esta máquina para pedirlo en la ronda de entrega. También hace que el build siga ante decisiones reversibles o que no bloquean, y las confirme todas juntas al final.

## Evidencia

Las sesiones las revisaron tres hijos `hive-research` el 2026-10-28; el hilo principal comprobó las citas de las reglas.

- **Sesiones revisadas:** 2026-09-20 a 2026-10-28, en Claude Code, Codex, Grok, Pi, OpenCode y Cursor.
  - En Claude, Codex y Grok hubo 227 pausas a mitad de build en 65 sesiones: 18 % evitables o previsibles, 25 % genuinas y 59 % paradas prescritas.
  - Después de #224 siguieron siendo evitables 6 de 25.
- **Fallo de CI:**
  - Codex cam `01a15822`, 2026-10-23 06:32Z: «una prueba falla en CI… ¿Cómo continuamos?».
  - Claude carvento `576e4f6d`, 2026-10-17 08:15Z: «CI fallida… ¿Qué hacemos?».
  - Claude cam `df8c65a1`, 2026-10-27 04:00Z: D10.
- **Efectos de entorno:**
  - Codex cam `01a1679c`, COM-867: nueve pausas para el primer corte a QA.
  - Pi carvento `01a17176`: cinco autorizaciones escalonadas de datos de prueba en QA.
  - Pi carvento `01a175b4`, 2026-10-26, D7: la allowlist de Atlas, tres minutos después de iniciar el build.
  - Claude cam `df8c65a1`, D11: un grant en QA.
- **Decisiones de diseño a mitad de build:** 42 pausas, entre ellas Claude carvento `94440d`, D39 y D40, con 538 min de espera.
- **Reglas que lo provocan** (leídas en `development` `3bb1e623`):
  - `content/skills/flow-build/references/verification.md:56`: «the failure comes from the change's code… report the failed jobs… as a pending item».
  - `content/skills/flow-build/references/foreseen-authorizations.md:9`: «what the plan, or for work without a plan the request, already names… an effect neither covers is asked when it arises».
  - `content/skills/flow-plan/SKILL.md:40`: entornos y efectos solo «for substantial gates».
  - `content/skills/flow-plan/SKILL.md:44-54` y `content/skills/flow-plan/references/plan-review.md:22`: no hay ninguna comprobación de entornos.
  - `content/skills/flow-build/SKILL.md:20`: «ask for the missing decision before dependent edits».
  - `content/skills/hive-core/references/questions.md:11`: «ask only at that decision point».

## Alcance y aceptación

**Incluido:**
- la viñeta del CI en `verification.md`;
- la lista de readiness de `flow-plan/SKILL.md`, las comprobaciones bloqueantes de `plan-review.md`, `foreseen-authorizations.md` y `delivery-decisions.md:7`;
- una condición nueva en `questions.md` para las decisiones a mitad de build, con su excepción en `questions.md:11`, el puntero de `flow-build/SKILL.md:20`, el de `global.md:29` (D8-A) y el de `reporting.md:84`;
- las páginas Build y Plan de `site/` y el `CHANGELOG.md`.

**Excluido, con su razón:**
- El bypass `--admin` (G4): depende de cada usuario. `git-workflow/SKILL.md:42` ya permite el admin merge cuando el repositorio no aplica sus reglas a los admins, y el dato va en Engram (`user/admin-in-own-orgs`).
- La ronda de entrega tardía en builds sin `flow-build` (G3): está en observación (`watch/late-delivery-round-without-flow-build`).
- El paso a D7-A: está en observación (`watch/build-assumptions-d7b`).
- Las paradas de validación, los efectos críticos y las decisiones que bloquean todo y no son reversibles: siguen preguntándose cuando aparecen.
- Las reglas de cambio de alcance (`scope-changes.md`): un pedido nuevo del usuario se clasifica como hasta ahora.

**Restricciones:**
- `go test ./...` pasa, incluido el límite de 16 000 bytes por skill y referencia. Hoy quedan `verification.md` en 15 380 B, `flow-plan/SKILL.md` en 14 987 B, `flow-build/SKILL.md` en 14 915 B y `questions.md` en 14 383 B.
- Cada regla conserva un único hogar y los demás sitios apuntan a ella.
- `content/guidance/global.md` solo gana el puntero de la línea 29 (D8-A). El PR reporta su cambio de tamaño, que hoy es 17 539 B.
- La delegación desatendida mantiene su propio régimen de decisiones.

**Búsqueda de specs que decidió los deltas:** busqué «question», «authoriz», «CI», «assumption», «readiness», «plan review» y «SHALL NOT» en `_support/openspec/specs/*/spec.md`. Leí `context-delivery` (requisitos «Portable kernel…» y «Layered global core…»; las preguntas solo aparecen como parte de la unión de fuentes, en la línea 84) y los SHALL NOT de `documentation-site` y `versioned-installation`. Ningún requisito describe estos comportamientos, así que el cambio no lleva deltas.

- AC1. `verification.md` indica que un fallo de CI causado por el cambio se arregla sin preguntar, bajo la regla global de defectos que bloquean: un commit nuevo en el mismo PR antes del merge, o un PR de seguimiento bajo la misma entrega después. Solo queda pendiente cuando el arreglo corre en producción, necesita un efecto no cubierto o agota las rondas. *Falso en la base cuando* `verification.md:56` dice «or the failure comes from the change's code, the integration is not complete… as a pending item».
- AC2. La lista de readiness de `flow-plan/SKILL.md` exige que cada criterio que se verifica fuera de esta máquina nombre su entorno y los efectos, datos y accesos que su verificación necesita allí. `plan-review.md` reporta como hallazgo bloqueante el criterio que no lo hace. *Falso en la base cuando* `rg -n "outside this machine" content/skills/flow-plan/SKILL.md content/skills/flow-plan/references/plan-review.md` no encuentra nada.
- AC3. `foreseen-authorizations.md` define que, para trabajo sin plan, lo previsto es lo que la petición y su verificación necesitan, listado antes de preguntar, y no solo lo que la petición nombra. *Falso en la base cuando* la línea 9 dice «what the plan, or for work without a plan the request, already names».
- AC4. `questions.md` define cómo se trata, durante un build y fuera de una parada de validación, una decisión del usuario que no autoriza efectos ni cambia el alcance:
  - si es reversible, se aplica la opción recomendada como supuesto registrado y se sigue;
  - si no lo es pero no bloquea el resto, se sigue con el trabajo independiente;
  - se pregunta en el momento solo cuando bloquea todo lo restante y no es reversible.

  Reversible quiere decir que deshacerla solo cambia trabajo de esta tarea en esta máquina, aún no publicado. Eso incluye las decisiones de contrato internas del cambio (D8-A).

  Las decisiones pendientes se preguntan juntas cuando ya no queda trabajo independiente. Los supuestos se confirman según el modo de entrega:
  - en la parada de validación bajo `interactive`;
  - después del merge bajo `automatic` y `direct-base` (D9-A): se listan entre las decisiones tomadas sin preguntar y se confirman en la ronda de la pregunta de cierre; una corrección va como PR de seguimiento bajo la misma entrega;
  - al entregar el resultado bajo `hold`, también en la ronda de cierre.

  Un supuesto no es un ítem pendiente, y `reporting.md:84` lo apunta. Un rechazo que llega después del cierre es trabajo nuevo.

  La regla excluye:
  - las tareas añadidas y los cambios nuevos, que siguen `scope-changes.md`;
  - las decisiones que otra regla reserva a la respuesta explícita del usuario;
  - la delegación desatendida.

  `questions.md:11`, `flow-build/SKILL.md:20` y `global.md:29` apuntan a la regla. *Falso en la base cuando* `rg -n "recorded assumption" content/skills/hive-core/references/questions.md` no encuentra nada y `flow-build/SKILL.md:20` dice «or ask for the missing decision before dependent edits».
- AC5. Las páginas Build y Plan de `site/` describen los tres comportamientos sin copiar las reglas. *Falso en la base cuando* `site/src/content/docs/flows/build.md:84` dice «asks you for the missing decision before it keeps editing» y ninguna de las dos páginas menciona el entorno de verificación ni el fallo de CI.

## Entrega

- **Repositorio:** `tricell-hive`, base `development`, en la rama propuesta `fix/gh-249-mid-build-questions`.
- **Modo automatic (D5-A):**
  - commits por tarea;
  - `/code-review` de Claude Code sobre el candidato local (D6-A);
  - push, PR a `development` y merge con las comprobaciones locales en verde. El repo no tiene CI ni protección de rama.
  - Después, `hive update` por la regla del mantenedor.
- **Cierre del registro:** un PR pequeño a `development`, sin revisión dedicada, que integra el agente.
- **Tickets:** #249 se cierra al integrar el cambio.
- **Implementación:** pendiente de la pregunta de build.
