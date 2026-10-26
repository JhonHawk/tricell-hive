# Delegación desatendida que sobrevive a un «continúa», preguntas previsibles y tareas tardías verificadas (2026-10-24)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` el 2026-10-24 por el [PR #145](https://github.com/JhonHawk/tricell-hive/pull/145), merge `2a100cd5`; #143 y #144 cerrados; registro archivado el mismo día |
| Tracker · GitHub Issues | • [#143 — Delegación desatendida: un «continúa» la termina aunque no cambie nada](https://github.com/JhonHawk/tricell-hive/issues/143)<br>• [#144 — Preguntas previsibles: definir prueba costosa y pedir las autorizaciones del plan mientras el usuario está presente](https://github.com/JhonHawk/tricell-hive/issues/144) |
| Git | `automatic` (E3) · rama desde `development` · sin parada antes del push |
| Verificación | `hive-verify-task` por tarea · `go vet ./...`, `go test ./...` y `npm run build` en `site/` sobre el candidato final · `/code-review` de Claude Code (E4) |
| Siguiente paso | Ninguno; registro cerrado |

## Solicitud

Mensajes del usuario, 2026-10-24:

> Ahora, considerando esto, necesito que revisemos la sesión de cam en codex, se esta haciendo eterna, y pareciera haber perdido ciertas reglas, asi como la de carvento, que inicialmente se inicio con claude y ahorita se retomo con PI, muchas cosas quedaron inconclusas hasta el momento de que ya era un problema [...] ve si puedes retroceder al 19 de octubre para analizar las sesiones

> no solo de esos harness, de todos los que se hana visto involucrados en ese rango de fechas, porfa

> por cierto, creo que el problema se esta presentando con tareas largas o complejas

Respuestas siguientes, 2026-10-24, en tarjetas nativas:

- D6-A: investigar las 48 decisiones de cam.
- P1-A: «Plan de los 3 cambios»: la duración y reactivación del modo desatendido; las pruebas costosas y las preguntas previsibles mientras el usuario está presente, incluida la autorización anticipada de efectos encadenados; el tercer punto, que E1 reemplazó.
- P2-A: abrir los issues [#143](https://github.com/JhonHawk/tricell-hive/issues/143) y [#144](https://github.com/JhonHawk/tricell-hive/issues/144).
- E1-B: «Endurecer ya las tareas tardías». Una tarea agregada a mitad del build o tras la validación cumple los mismos criterios y pasa por `hive-verify-task`, y los cambios posteriores a la revisión se preguntan. El usuario adopta esta corrección sin esperar una segunda sesión (AGENTS.md, Measurement).
- E2-A: una prueba es costosa si gasta dinero o cuota externa, usa o cambia recursos compartidos o remotos, o se espera que tarde más de 30 minutos. Un ensayo local de minutos sin costo externo no se pregunta.
- E3: entrega `automatic`. E4: `/code-review` de Claude Code.

Criterios de aceptación de #143, citados textualmente (leído el 2026-10-24):

> - La skill distingue entre un mensaje que solo pide continuar y uno que cambia el objetivo, el alcance o las respuestas, o pide detenerse. Para cada caso dice si la delegación sigue, se renueva o termina.
> - Un caso de comprobación con «continúa» después de una delegación por ausencia da como resultado que el agente sigue decidiendo y registrando, sin preguntar.
> - No se amplían los efectos que la delegación autoriza: siguen excluidos los irreversibles y los destructivos.

Criterios de aceptación de #144, citados textualmente (leído el 2026-10-24):

> - La guía define "costoso" con un criterio observable, y un ensayo local de minutos sin costo externo queda fuera.
> - La guía pide preguntar en una sola ronda, mientras el usuario está presente, las confirmaciones y los efectos que el plan ya prevé para pasos posteriores, cada uno nombrado con su alcance.
> - Los efectos sobre producción conservan su propia pregunta, salvo que el usuario ya los haya nombrado.

## Objetivo

La revisión de las sesiones de cam en Codex, del 20 al 24 de octubre, mostró que unas 88 h de la obra fueron espera de respuestas del usuario.

- **Delegación desatendida.** Unas 20 h se perdieron porque un «continúa» terminó la delegación: la skill dice que expira cuando «the user writes again», y el agente volvió a preguntar con el usuario ausente.
- **Preguntas tardías.** Otras ~25 h vinieron de confirmaciones que el plan ya preveía (un presupuesto para un ensayo local de 128 s, publicar imágenes, preparar el VPS de QA). Se preguntaron al llegar a cada paso, no en la ronda en la que el usuario estaba presente.
- **Tarea tardía sin verificar.** En carvento, una tarea agregada al plan después de la validación (T19) se entregó sin `hive-verify-task` y después de la revisión de código, y su defecto apareció en el primer uso real.

Tras el cambio:

- un mensaje que solo pide continuar mantiene la delegación;
- «costoso» tiene un criterio observable;
- el agente pide en una sola ronda, mientras el usuario está, las autorizaciones que el plan ya prevé;
- una tarea agregada tarde se planea y verifica como cualquier otra, y si llega después de la revisión se pregunta si revisarla.

## Alcance y aceptación

Incluido:

- `content/skills/unattended-delegation/SKILL.md` (párrafo de activación y expiración);
- `content/guidance/global.md` (viñeta de trabajo desatendido y viñeta de gates);
- `content/skills/flow-build/references/verification.md` (gates y su párrafo de suites largas);
- `content/skills/flow-plan/SKILL.md` (puntero en la frase de gates sustanciales);
- `content/skills/flow-plan/references/plan-revision.md` (tareas agregadas);
- `site/src/content/docs/index.mdx`, `site/src/content/docs/flows/build.md` y `site/src/content/docs/flows/plan.md`;
- `CHANGELOG.md`.

Excluido:

- El texto de #45 (frase D1-B): queda como vigilancia con 1 caso; E1-B lo reemplaza para este plan.
- Editar `content/skills/flow-close/SKILL.md` («Close question») y `content/skills/flow-close/references/reporting.md`. Ya dicen que una respuesta autoriza el conjunto y que producción tiene su propia pregunta (`reporting.md:106`), así que la regla nueva remite a `reporting.md` en vez de repetirlo. Además, `flow-close/SKILL.md` está a 333 bytes del límite.
- La espera en tramos de 60 s de Codex: la sigue #142.
- Las tarjetas sin mensaje y las etiquetas en inglés: las siguen #70 y #71.
- Deltas de requisitos: ningún spec de `_support/openspec/specs/` describe los flujos de Hive.

Restricciones:

- Cada `content/skills/*/SKILL.md` tocado se mantiene en 16 000 bytes o menos (`tests/content/skill_size_test.go`).
- `go vet ./...`, `go test ./...` y `npm run build` en `site/` pasan sobre el candidato final.
- El PR declara el cambio de tamaño de `content/guidance/global.md` (35 036 bytes en la base) y lista los lugares revisados (AGENTS.md).
- No se amplían los efectos que autoriza la delegación desatendida.

Criterios de aceptación:

- AC1. El párrafo de expiración de `unattended-delegation/SKILL.md` dice que la delegación se mantiene, con el mismo objetivo y expiración, ante tres tipos de mensaje en cualquier idioma: el que solo pide continuar, el que acusa recibo y el que pregunta por el estado. También dice que la terminan un mensaje que cambia el objetivo, el alcance o una respuesta vigente, hace otra pregunta, da una instrucción nueva o pide detenerse. *Falso en la base cuando* `content/skills/unattended-delegation/SKILL.md:8` dice «expires when that objective completes or the user writes again».
- AC2. La viñeta de trabajo desatendido de `global.md` expresa la misma expiración que la skill: no la terminan un mensaje que solo pide continuar, uno que acusa recibo ni uno que pregunta por el estado. *Falso en la base cuando* `content/guidance/global.md:34` dice «expiring when it completes or the user returns».
- AC3. La viñeta de gates de `global.md` define como costosa la prueba que gasta dinero o cuota de un servicio externo, cambia recursos compartidos o remotos, o se espera que tarde más de 30 minutos. Excluye de esa definición el uso de modelo de la propia sesión y sus hijos, la CI del proyecto y las lecturas normales como bajar dependencias, y dice que una prueba local de minutos sin costo externo ni efecto irreversible es rutinaria. *Falso en la base cuando* `content/guidance/global.md:68` dice «costly load/hardware exercises» sin criterio.
- AC4. `verification.md` exige pedir, mientras el usuario está presente y en la ronda en que se resuelven las preguntas de entrega y revisión, las autorizaciones que el plan prevé para pasos posteriores: pruebas costosas o efectos como publicar un artefacto o preparar un entorno compartido. Cada una va nombrada con su alcance, destino y límites, agrupadas según la regla «Ask once for a set of actions» de `reporting.md`. Esa ronda es la previa a declarar listo un plan o a la primera edición de un trabajo sin plan; cuando no hay ronda porque la configuración, una respuesta anterior o la solicitud ya resuelven esas preguntas, las autorizaciones se piden solas en ese mismo punto, y `delivery-decisions.md` remite a esta regla. Una autorización sin respuesta sigue sin autorizar, también bajo delegación desatendida; la skill `unattended-delegation` lo repite en su sección «Decisions». La regla declara que es una excepción a «ask only at that decision point» de `questions.md`, limitada a lo que el plan o, sin plan, la solicitud ya nombran. Remite a `reporting.md` para agrupar y para el trato de producción, y dice que bajo delegación desatendida esas autorizaciones se deciden según su sección «Effects», sin preguntar. *Falso en la base cuando* `content/skills/flow-build/references/verification.md:26` solo dice «A plan may carry explicit advance authorization; its mere inclusion of a check is insufficient».
- AC5. El párrafo de suites largas de `verification.md` exige evaluar alcance y efectos solo cuando la prueba es costosa según la guía global o puede causar un efecto irreversible (como flashear un dispositivo), aunque arranque localmente. *Falso en la base cuando* `content/skills/flow-build/references/verification.md:39` exige la evaluación «regardless of whether their command starts locally» sin condicionarla al costo.
- AC6. La frase de gates sustanciales de `flow-plan/SKILL.md` remite a la regla de AC4 para pedir las autorizaciones pendientes en la ronda de entrega y revisión. *Falso en la base cuando* `content/skills/flow-plan/SKILL.md:36` solo dice «record whether execution authorization is pending or already sufficient under the shared rule».
- AC7. `plan-revision.md` dice que una tarea agregada durante el build o tras la validación tiene criterios falsables en la base, `Closes:` y enfoque de prueba. También dice que, si su enfoque es `tdd` o `check`, la verifica `hive-verify-task` antes de su commit, como fija `plan-execution.md`, y que el informe final no atribuye verificación independiente a una tarea que no la tuvo. Su «When to read» incluye agregar una tarea, y distingue una corrección del resultado validado (rondas de corrección de `flow-walkthrough`) de un comportamiento nuevo pedido entonces (tarea agregada). *Falso en la base cuando* `content/skills/flow-plan/references/plan-revision.md:10` solo habla de «a changed task».
- AC8. `plan-revision.md` dice que una tarea agregada después de que corrió la revisión de código dedicada se informa y se pregunta al usuario si revisarla, con remisión a `verification.md`. *Falso en la base cuando* `plan-revision.md` no menciona la revisión dedicada.
- AC9. El sitio describe los cambios en tres páginas. *Falso en la base cuando* ninguna de las tres menciona esos puntos.
  - `site/src/content/docs/index.mdx` dice que un mensaje que solo pide continuar no termina el trabajo desatendido.
  - `site/src/content/docs/flows/build.md` dice qué cuenta como prueba costosa y que una tarea agregada tarde también pasa por el verificador.
  - `site/src/content/docs/flows/plan.md` dice que las autorizaciones que el plan prevé se piden junto con las preguntas de entrega y revisión.

Sesiones reales que motivan el cambio (para los mensajes de commit, AGENTS.md, Measurement):

- Codex CLI 0.163.0 y 0.164.0, proyecto cam, hilos `01a15822-b9bc-7923-8fc2-ba2b68161fa2` (2026-10-20 a 10-23) y `01a1679c-7ed3-7030-b5f5-345d62c1dcba` (2026-10-23 a 10-24):
  - la delegación termina con un «continúa» (D22 y D23, unas 20 h de espera);
  - autorizaciones previsibles preguntadas tarde (D29, D30, D42, D43).
- Claude Code, proyecto carvento, sesión `94440d0a-42ac-491f-ba69-d327f8cdae5b` (2026-10-24): T19, una tarea agregada tras la validación, entregada sin `hive-verify-task` y después de la revisión.

## Entrega

- **Repositorio y modo:** `tricell-hive`, base `development`, modo `automatic` (E3, 2026-10-24). Se trabaja en una rama de trabajo desde `development` en un worktree propio, porque otra sesión usa el checkout principal. Commits por tarea verificada, gates compartidos, push, PR, `/code-review` de Claude Code (E4) sobre el candidato, y merge a `development` cuando pasan las pruebas locales y la revisión. Sin parada para validación humana: el cambio no tiene superficie visual. Después del merge, `hive update` local por el marcador de maintainer.
- **Registro del cambio:** se versiona una vez al cierre con `flow-close`, en un PR pequeño de cierre sin revisión dedicada que fusiona el agente.
- **Promoción a `master`:** no incluida.
