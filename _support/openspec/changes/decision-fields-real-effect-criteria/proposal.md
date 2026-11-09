# decision-fields-real-effect-criteria — Decisiones con campos y criterios que cierran contra el efecto real

| Campo | Valor |
| --- | --- |
| Estado | En entrega · construido y verificado; pull request a `development` |
| Tracker | ninguno |
| Git | JhonHawk/tricell-hive · base `development` en `dab27f73` · rama `feat/decision-fields-real-effect-criteria` |

## Decisiones

2026-11-08
- **D5 D** — ¿Cómo se entrega? Solo local: se edita el árbol de trabajo, sin commit ni publicación. La respuesta se dio para la actualización de la rúbrica de revisión en esta misma sesión y repositorio, y sigue en vigor para este trabajo.
- **D6 A** — ¿Cómo se hace comprobable el contexto de una pregunta? Con campos con etiqueta: cada decisión lleva una línea de situación, una línea de consecuencia bajo cada opción y una línea de recomendación, y la ronda dice cuántas decisiones quedan. Se descartaron la prosa con pruebas y un lector independiente por ronda.
- **D7 A** — ¿Dónde se exige que la verificación sea contra el efecto real? En el criterio de aceptación: cada uno dice dónde se observa, y una comprobación contra un sustituto no lo cierra; si el efecto solo existe en una etapa posterior, queda sin verificar hasta esa etapa y así se reporta. Se descartaron exigirlo solo en los roles verificadores y limitarse a medir.
- **D8 A** — ¿Qué revisión de código corre antes de la validación? La nativa de Claude Code, `/code-review`, sobre el diff del árbol de trabajo.
- **D9 A** — ¿Qué cuenta como efecto real para una pantalla o una API del propio proyecto? El cambio servido en local con el build real y datos sembrados; los mocks o emuladores de servicios ajenos no cierran un criterio.
- **D11 A** — ¿Se aprueban los dos textos de regla? Sí, tal como quedaron tras las dos revisiones de código.
- **H7 A** — ¿Qué se hace con que `flow-close` no conserve un criterio sin verificar? Corregirlo ahora, como tarea añadida de este plan, recortando la redacción de `flow-close` para que quepa en su tope.
- **D12 A** — ¿Qué se hace con el arreglo de `flow-close`, que su revisión mostró insuficiente? Se retira de este cambio y se abre un ticket para diseñarlo, junto con repartir ese archivo, que está en su tope de tamaño. **Replaces:** H7-A
- **D13 A** — ¿Por qué ruta llega el commit a `development`? Por la normal del repositorio: rama, commit, push, pull request y merge, y después el despliegue con `hive update`. El usuario pidió commit y despliegue el 2026-11-09. **Replaces:** D5-D
- **D14 A** — ¿Entra la rúbrica de revisión en la misma publicación? Sí, en su propio commit del mismo pull request.
- **D10** — ¿Un criterio que nombra un entorno desplegado hace que el agente recorra ese despliegue? No. Respuesta del usuario: tras un despliegue basta confirmar que la aplicación no está caída; revisar todo o puntos concretos es a petición suya, porque es normal que se acumulen varios tickets antes de desplegar. Ese criterio queda sin verificar hasta que el usuario pida el recorrido.

## Solicitud

Del usuario, 2026-11-08, tras revisar la sesión de Codex en fleetsystem: «Las preguntas no tienen el suficiente contexto, por eso yo como usuario realmente no sé qué contestar y eso lleva a replanteamientos extraños». Después: «hagamos research, cuál es el problema de fondo», y al ver la evaluación de once sesiones eligió planear «preguntas con contexto y verificación contra el efecto real».

Queda fuera de este cambio: la medición del mundo al planear, la autorización de efectos en paquete, la reducción de las paradas previstas y el gobierno de sesiones largas. Se identificaron en la misma investigación y no se planean aquí.

## Objetivo

Hoy «Questions and routes» pide «the context needed to answer it» sin prueba observable, mientras el formato (ID, título, estrella, letra) sí se comprueba a simple vista; los modelos cumplen el formato y el usuario recibe preguntas que no entiende. Hoy ninguna regla dice dónde se observa un criterio de aceptación, y una tarea puede cerrarlo con un mock.

Cuando el cambio esté hecho, una decisión sin sus tres líneas con etiqueta incumple una regla comprobable, y un criterio solo se da por cumplido en el lugar donde su efecto real se ve; lo demás se reporta como no verificado.

## Alcance y aceptación

Dentro: la regla de formato de decisión en `content/guidance/global.md`, la regla de criterios en `flow-plan`, su comprobación en `check_change.py`, los textos que dependen de ambas, las páginas del sitio que las describen y la entrada del registro de cambios.

Fuera: los hallazgos `H<n>` y las respuestas del bloque de siguiente paso, que conservan su formato; la regla de la respuesta de compactación, que se queda en `global.md` porque debe actuar después de compactar; un detector nuevo en `tests/pilot` (los pilotos están pausados y `AGENTS.md` pide no añadir casos de regresión); el comportamiento de los agentes en sesiones reales, que este cambio no puede observar antes de entregarse.

- AC1 — «Questions and routes» exige en cada decisión `D<n>` una línea `Situation:` bajo el título, una línea `If you choose it:` bajo cada opción y una línea `I recommend <letter>:`, con las etiquetas en el idioma de la sesión; la situación conserva lo que hoy pide el contexto: qué se verificó y qué hay que revisar, con su destino · falso hoy porque `rg -n "Situation:" content/guidance/global.md` no encuentra nada · Observed in: `content/guidance/global.md` del candidato
- AC2 — La regla prohíbe justificar una decisión citando una regla o una skill y pide decir el efecto que necesita la respuesta · falso hoy porque `rg -n "quotation of a rule" content/guidance/global.md` no encuentra nada · Observed in: `content/guidance/global.md` del candidato
- AC3 — La regla pide que cada ronda termine diciendo cuántas decisiones conocidas quedan · falso hoy porque `rg -n "still to come" content/guidance/global.md` no encuentra nada · Observed in: `content/guidance/global.md` del candidato
- AC4 — La regla de la tarjeta dice que la pregunta lleva la situación, que la descripción de cada opción lleva el texto de `If you choose it:` y que la de la opción recomendada termina con la razón de la recomendación · falso hoy porque la línea de la tarjeta dice «its description carries the explanation» · Observed in: `content/guidance/global.md` del candidato
- AC5 — La respuesta de compactación del bloque de siguiente paso se define en `content/skills/flow-plan/references/plan-handoff.md` y `global.md` ya no la enuncia **Retired:** la revisión del plan mostró que esa regla debe actuar después de la compactación, cuando una referencia ya no está en contexto; se queda en `global.md`
- AC6 — `flow-plan` «Acceptance criteria» exige que cada criterio termine con `Observed in: <lugar>`, dice que una comprobación contra un sustituto no cierra un criterio y que `Closes:` nombra solo los criterios cuyo lugar la tarea alcanza; la plantilla `proposal.md` lo muestra · falso hoy porque `rg -n "Observed in" content/skills/flow-plan` no encuentra nada · Observed in: `content/skills/flow-plan/SKILL.md` y su plantilla en el candidato
- AC7 — `check_change.py` imprime un hallazgo por cada criterio activo sin `Observed in:` y no lo imprime con `--archive` · falso hoy porque el script termina en 0 con una carpeta cuyos criterios no llevan el marcador · Observed in: la salida de `python3 content/skills/flow-plan/scripts/check_change.py` sobre carpetas reales de prueba
- AC8 — `flow-build`, `flow-verify`, `hive-verify-task` y `flow-walkthrough` dicen lo mismo que AC6, hoy AC11: la evidencia de un sustituto deja el criterio sin verificar, y los datos stub sirven solo a la primera vista del build · falso hoy porque `flow-walkthrough/SKILL.md:10` pide «stub or seeded data» para el mismo recorrido que `flow-verify/SKILL.md:12` exige «through its real path» · Observed in: los cuatro archivos del candidato
- AC9 — Las páginas del sitio de plan y de verify describen los campos de una decisión y dónde se observa un criterio, y `agent-delivery.md` registra el cambio de formato · falso hoy porque `rg -n "Observed in|If you choose" site/src/content/docs _support/docs/architecture` no encuentra nada · Observed in: el sitio construido en local y los dos documentos del candidato
- AC10 — La regla pide que la línea de espera de cada ronda diga también cuántas decisiones conocidas quedan, de modo que la ronda cierra con una sola línea **Replaces:** AC3 · falso hoy porque `rg -n "still to come" content/guidance/global.md` no encuentra nada · Observed in: `content/guidance/global.md` del candidato
- AC11 — `flow-plan` «Acceptance criteria» exige que cada criterio termine con `Observed in: <lugar>`; dice qué cuenta como efecto real (el candidato servido en local con el build real del proyecto y datos sembrados de volumen realista, para una pantalla o una API que el proyecto sirve) y qué es un sustituto (la emulación o el mock de un servicio que el proyecto no posee); dice que un sustituto nunca cierra un criterio y que, hasta que un recorrido lo observe en su lugar, el criterio sigue en el `Closes:` de su tarea y se reporta sin verificar; la plantilla `proposal.md` lo muestra **Replaces:** AC6 · falso hoy porque `rg -n "Observed in" content/skills/flow-plan` no encuentra nada · Observed in: `content/skills/flow-plan/SKILL.md` y su plantilla en el candidato
- AC12 — `flow-build`, `hive-verify-task`, `hive-verify-change`, `flow-verify`, el recorrido funcional de `flow-deploy` y `flow-walkthrough` dicen lo mismo que AC11: la evidencia de un sustituto deja el criterio sin verificar, cada criterio se comprueba en el lugar que nombra, un criterio que nombra un entorno desplegado se cierra en el recorrido funcional de ese despliegue, cuando ese recorrido corre, sin que nombrar el entorno lo pida, y los datos stub sirven solo a la primera vista del build **Replaces:** AC8 · falso hoy porque `flow-walkthrough/SKILL.md:10` pide «stub or seeded data» para el mismo recorrido que `flow-verify/SKILL.md:12` exige «through its real path» · Observed in: los seis archivos del candidato
- AC13 — «Delivery» ya no pide la razón de la recomendación «next to the star», ni una línea aparte sobre lo que toca el cambio, ni notas «in the context»; apunta a las líneas `I recommend` y `Situation:` · falso hoy porque `rg -n "next to the star" content/guidance/global.md` encuentra la frase · Observed in: `content/guidance/global.md` del candidato
- AC14 — `viewer-publication.md` dice que la opción «Open the viewer» no lleva línea `If you choose it:`, cuenta cuatro reglas cambiadas y dice que al volver a preguntar se repite la tarjeta sin repetir el texto · falso hoy porque dice «This changes only three rules» · Observed in: `content/skills/flow-walkthrough/references/viewer-publication.md` del candidato
- AC15 — `CHANGELOG.md` nombra en `[Unreleased]` los campos de una decisión y el marcador `Observed in:`, con la nota de migración: las carpetas de cambio en curso reciben el hallazgo en su siguiente revisión y las archivadas no cambian · falso hoy porque `rg -n "Observed in" CHANGELOG.md` no encuentra nada · Observed in: `CHANGELOG.md` del candidato
- AC16 — `flow-close` cuenta como trabajo pendiente el criterio que sigue sin verificar hasta que un recorrido lo observe en su lugar, y su reporte final lo nombra como no verificado con ese lugar **Retired:** su revisión de código mostró que no es una corrección local; sale de este cambio y va a un ticket (D12 A)

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

Leído y medido el 2026-11-08 sobre `development`:

- `content/guidance/global.md`: 44,450 bytes; «Questions and routes» ocupa las líneas 76–94, 11,047 bytes sin su encabezado, en 16 puntos. El punto 3 (formato de decisión) pide «the context needed to answer it: what you verified and what to check, with its target» y la razón de la recomendación solo para algunas decisiones.
- `content/skills/flow-verify/SKILL.md:12` ya exige recorrer el cambio «through its real path», sobre el candidato servido en local; `content/skills/flow-plan/SKILL.md:29` define los criterios sin decir dónde se observan y exige que cada uno aparezca en un `Closes:` cuya verificación lo cubra.
- `content/skills/flow-deploy/references/deployment-checks.md:32`: el recorrido funcional de un despliegue corre solo cuando el proyecto o el plan lo declaran o el usuario lo pide.
- `content/skills/flow-close/SKILL.md` pesa 15,995 bytes de un tope de 16,000; este cambio no lo toca.
- Sesión que motivó el cambio: Codex 0.167.0, `gpt-6.1-sol` medium, fleetsystem, `01a1b955-3ed2-73b1-8396-6902a84c8480`, 2026-11-08. El usuario respondió a una decisión «¿te refieres a cuenta de aws?»; siete decisiones se justificaron citando una skill; la tarea de IAM se verificó con mocks y el apply real se detuvo en 82 de 85 recursos.
- Evaluación de once sesiones más con modelos del piso, 2026-10-31 a 2026-11-08: 14 respuestas del usuario con otra pregunta o corrigiendo la premisa, en 9 sesiones; 22 decisiones justificadas con una regla, en 8; 12 casos de comprobación que pasa y efecto real que falla, en 7. Clasificación hecha por hijos sobre líneas de tiempo recortadas; seis citas comprobadas a mano. Sesiones, por el inicio de su ID: Claude Code con `claude-opus-5-5` en tricell-hive `335e6514` (03-nov), `98476eca` (02-nov), `80411b43` (07-nov), `2216a863` (02-nov), `9105935a` (08-nov), `7bcc87ae` (02-nov), en cam `9a6a7fe6` (02-nov) y en cam-monorepo `32cffc1d` (02-nov); Codex con `gpt-6.1-sol` en cam `01a19339` (31-oct) y en tricell-hive `01a19638` y `01a19479` (01-nov).
- Las dos reglas son correcciones de reglas existentes; las adopta la aprobación del usuario en D6 y D7, como permite «Measurement» de `AGENTS.md`.
- Lugares que dependen de ambas reglas, del barrido del 2026-11-08: ver `design.md` «Impacto».

## Entrega

Automática (D13 A): rama `feat/decision-fields-real-effect-criteria`, commit, push, pull request y merge a `development`, y después el despliegue a los hosts con `hive update`, que el usuario pidió. La validación de los dos textos de regla ya se dio (D11 A). Revisión de código: `/code-review` de Claude Code (D8 A), limitada a los archivos de `design.md` «Impacto»; `_support/docs/harness-engineering/session-review-rubric.md` está modificado por otro trabajo de esta sesión y queda fuera. Esta entrega no autoriza promover a `master` ni cortar una versión.

El efecto que pidió el usuario, decisiones que pueda responder y verificación contra el efecto real, solo se ve en sesiones reales que corran la regla nueva. Las filas del 2026-11-08 de la rúbrica de revisión, con las que se hace ese seguimiento, son otro trabajo de esta sesión y van en su propio commit del mismo pull request (D14 A). El commit nombra las observaciones de «Contexto verificado», como pide «Measurement».

## Estado de la revisión y avance

2026-11-08 — Plan escrito.
2026-11-08 — Revisión del plan, ronda 1: 2 hallazgos P1, 4 P2 y 5 P3, todos aplicados. La respuesta de compactación se queda en `global.md` (AC5 retirado, T2 eliminada); una tarea sigue nombrando en `Closes:` el criterio que no alcanza y lo reporta sin verificar (AC11 reemplaza a AC6, AC12 a AC8); la cuenta de decisiones va en la línea de espera (AC10 reemplaza a AC3); se añaden «Delivery» (AC13) y la excepción del visor (AC14); T4 incluye `tasks_test.py`; nueva T7 para verify y deploy.
2026-11-08 — Revisión del plan, ronda 2: sin P0 ni P1; 5 P2 y 8 P3 aplicados. La situación conserva «qué se verificó y qué revisar»; `flow-plan` reescribe su frase de cobertura y `hive-verify-task` su veredicto por plan; se añaden `hive-verify-change`, la página de deploy y `CHANGELOG.md` (AC15); la tarjeta lleva la razón en la opción recomendada. Dos huecos van al usuario como D9 (qué cuenta como sustituto) y D10 (si un criterio de un entorno desplegado declara su recorrido); el texto asume la opción recomendada de ambas hasta su respuesta.
2026-11-08 — D9 A y D10 respondidas. D10 no fue la recomendada: un criterio de un entorno desplegado no declara el recorrido; se cierra solo cuando el usuario lo pide. Ajustados AC9 y AC12, el párrafo de `flow-plan`, T6, T7 y la delta.
2026-11-09 — Revisión de código (`/code-review`), 10 hallazgos; 9 corregidos en una ronda, con la redacción de AC11 y AC12 ajustada en su misma línea: la regla habla de «proyecto» y no de «repositorio»; ningún texto repite cuándo corre el recorrido de deploy, que sigue definido solo en `deployment-checks.md`; `hive-verify-task` conserva su veredicto `not met` por el plan y da `cannot verify` solo ante un sustituto o un entorno desplegado; un criterio sin marcador se verifica como antes; el script lee sub-viñetas y exige un lugar tras el marcador. Queda uno para el usuario: `flow-close` no dice qué hacer con un criterio que sigue sin verificar.
2026-11-09 — Segunda revisión de código, limitada a las correcciones: 10 hallazgos más; 9 corregidos con su prueba y sin nueva revisión, como fija `flow-verify`. `flow-plan` ya no manda al recorrido de deploy un criterio de lugar local; `hive-verify-task` da `not met` por el plan solo cuando la verificación ni cubre el criterio ni nombra la etapa que alcanza su lugar; el script acepta un enlace como lugar, lee el lugar envuelto en la línea siguiente, no toma el de un criterio anidado ni el de otra lista, y no informa sobre un criterio citado si su definición nombra el lugar. Queda sin corregir, por no ser bloqueante: el script calcula dos veces qué criterios están activos.
2026-11-09 — Validación del usuario: D11 A aprueba los dos textos. H7 A añade T8 y AC16 (tarea añadida, mismo resultado): `flow-close` registra como trabajo pendiente el criterio sin verificar. Sin nueva revisión del plan; la corrección pasa por la revisión de código.
2026-11-09 — La revisión de código de T8 encontró 10 problemas: el registro como trabajo pendiente contradice el requisito vigente «Promotions retain nothing», no tiene cómo cerrarse, alcanza a trabajo abandonado, pausado o sin recorrido, y para caber se quitó una cláusula de regla. Se revirtió `flow-close` a `HEAD` y T8 queda reabierta; AC16 espera la decisión D12 del usuario.
2026-11-09 — D12 A: el criterio del cierre se retira y su tarea se elimina; `flow-close` queda sin cambios.
2026-11-09 — Ticket #437 abierto para el hueco de `flow-close` (D12 A).
2026-11-09 — El usuario pide commit y despliegue: D13 A cambia la entrega a la ruta normal del repositorio y D14 A incluye la rúbrica.
