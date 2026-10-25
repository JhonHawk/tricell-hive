# Menos preguntas en los flujos: autorizar efectos por entorno y criticidad (2026-10-25)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` el 2026-10-25 por el [PR #152](https://github.com/JhonHawk/tricell-hive/pull/152), merge `4f06cc28` · #151 cerrado · desplegado en local con `hive update` · registro archivado el mismo día |
| Tracker · GitHub Issues | • [#151 — Preguntas de más en los flujos: autorizar efectos por entorno y costo, y no preguntar decisiones menores](https://github.com/JhonHawk/tricell-hive/issues/151) |
| Git | `automatic` (D7-A) · rama desde `development` · PR, `/code-review` y merge con la suite local en verde (no hay CI) · sin parada humana |
| Verificación | `hive-verify-task` por tarea · `go vet ./...`, `go test ./...`, `unittest`, y `pnpm check` y `pnpm build` en `site/` sobre el candidato final · `/code-review` de Claude Code (D8-A) |
| Siguiente paso | Ninguno; registro cerrado |

## Solicitud

Mensajes del usuario, 2026-10-25:

> Analiza las sesiones de los ultimos 5 dias en los diferentes proyectos, y sobre todo la sesiones de cam y carvento.
>
> La premisa es que ahora siento que hace muchas más preguntas innecesarias, por ejemplo que si a alguna le dices que no, simplemente ya no se puede cumplir la tarea/ticket, o que se podrian haber agrupado en una sola pregunta:
>
> "Para hacer la prueba e2e necesito crear un socio, lo creo? si/no", "Ahora necesito crear una sucursal, la creo? si/no?", esto es un cuello de botella en contestar mini preguntas y no creo que deba ser el objetivo

> Ojo que esto no es algo que deba arreglarse en un solo punto, esto va sobre todo nuestros flujo y skills de flow

> tambien tenemos una sesión activa en cam de claude y para un ticket pequeño senti muchas preguntas

> Para ciertas cosas se debe tener cierto criterio:
> - No es lo mismo crear insumos para un entorno de desarrollo o qa, que para producción. Ademas, de que el usuario puede optar por usar una cuenta o insumo ya existente, talvez una de las opciones que proponga esta pregunta es investigar si se pueden usar ya existentes (no se hace por defecto porque esta investigación toma tiempo),
> - Algo parecido para en crear infra o cosas que consuman dinero real, este criterio es infinitamente critico como para asegurar con el usuario si se hace o no

Respuesta a D4, en texto libre:

> No solo para costos, si no para cosas con efectos irreversibles, gasto de dinero, destrucción de infra, update de data en bs de datos (y otros que no menciono pero tienen un tono similar)

Corrección a la pregunta de ruta D11, 2026-10-25:

> ojo, en caso de tricell y sus clientes es local -> dev -> qa -> prod, pero para otros puede ser diferente, de hecho nosotros tenemos facti, que es local -> gamma -> prod, hay que tomar en cuenta que se pregunte desde la etapa de planeación los permisos para lo que desbloquee la validación de la implementación en caso de carvento, pues eran los insumos de QA (consideremos que talvez el ticket solo consideraba hasta development, por eso no entraba este cuestionamiento y lo siguiente si entraba a una etapa posterior)

Criterios de aceptación de [#151](https://github.com/JhonHawk/tricell-hive/issues/151), leídos el 2026-10-25:

> - La guía define los cuatro niveles de efecto (local, desarrollo/QA, producción, dinero o infraestructura) y qué pide cada uno. La pregunta de desarrollo/QA ofrece reutilizar insumos existentes como opción.
> - El plan o la pregunta de entrega carga la lista de efectos previstos, incluidos los datos de verificación en entornos no locales, y una sola respuesta los autoriza.
> - Un efecto de la misma clase y del mismo entorno ya autorizado no genera una pregunta nueva.
> - Los huecos de la revisión del plan con opción convencional no llegan al usuario como pregunta y quedan listados como decisiones tomadas.
> - Los PRs de corrección del mismo ticket no piden merge por separado cuando ya hay permiso y no hay condición contraria.
> - Las páginas de `site/` que describen esos flujos reflejan el cambio.

## Decisiones

- D1-A (2026-10-25): planear R1–R4 con `flow-plan`. D2-A: abrir [#151](https://github.com/JhonHawk/tricell-hive/issues/151).
- D3-A: en producción, una pregunta por tipo de efecto y tarea, con cada registro o destino, el cambio, la reversión y la duración. El mismo tipo de efecto sobre el mismo destino no se repregunta dentro de la tarea.
- D9-A, que reemplaza a D4: los **efectos críticos** llevan siempre su propia pregunta y nunca se agrupan. Son lo irreversible o difícil de revertir (borrar sin respaldo, enviar mensajes a personas reales, publicar a terceros, rotar credenciales), gastar dinero real o cuota de pago, destruir o reemplazar infraestructura, y modificar o borrar datos existentes en cualquier base de datos no local, migraciones y backfills incluidos. Se pueden preguntar al planear si ya se prevén, con el destino exacto, el costo o tope y la reversión cuando la hay. Un destino distinto o pasar el tope se vuelve a preguntar.
- D5-A: R3 se reduce a dos aclaraciones. Una condición del usuario dura lo que él dijo. Un PR de corrección del mismo ticket es parte de la misma entrega.
- D6-A: cada opción de ruta de la investigación dice si abre un ticket, y elegirla autoriza esa escritura en el tracker. Ya no hay una pregunta de ticket aparte.
- D7-A: entrega `automatic`. D8-A: `/code-review` de Claude Code.
- D13-A (2026-10-25, durante la revisión de código): en modo desatendido, con autorización previa para ese destino y tope, se puede gastar dentro del tope y hacer lo crítico reversible. El resto de lo irreversible sigue prohibido, aunque esté autorizado.
- D10-A (2026-10-25): en modo desatendido, un efecto crítico solo se ejecuta si se autorizó de antemano para ese destino y ese tope exactos. Si no, se salta, se sigue con lo que no depende de él y se reporta con destino, costo y reversión. Lo reversible y sin costo sigue el modo desatendido como hoy.
- Decisiones del agente, sin preguntar:
  - La regla vive en una referencia nueva, `content/skills/flow-plan/references/effect-authorization.md`, junto a `questions.md`, y en la guía global queda solo lo mínimo indispensable. Se eligió así porque estas autorizaciones también aparecen en la investigación (los accesos en facti), no solo en build.
  - Sin delta de specs: el cambio altera solo la guía distribuida, como en los cambios #143/#144 y #147.

## Objetivo

En cinco días de sesiones reales los agentes hicieron unas 245 preguntas. Muchas no aportaban una decisión: se respondían con la opción recomendada en segundos, mientras las esperas en ausencia del usuario llegaban a 3–12 h. El exceso no viene de una regla, sino de varias que se suman:

- Confirmar cada escritura en datos "compartidos", sin definir qué es compartido.
- Adelantar autorizaciones solo en tres casos, dejando que lo no previsto "se pregunte cuando surja".
- Prohibir que una opción de ruta incluya efectos.
- Mandar al usuario los huecos menores que encuentra la revisión del plan.
- No tener ninguna regla para agrupar los pasos de una misma cadena.

El cambio gradúa las autorizaciones por entorno y criticidad, con tus criterios. Hace que el plan o la primera pregunta de build lleve todos los efectos previstos. Deja que el agente decida y anote las decisiones menores, agrupa las cadenas, aclara la vigencia de las condiciones de merge y mete el ticket dentro de la ruta. El resultado esperado es menos rondas por ticket sin perder el control de lo que no se puede deshacer.

## Alcance y aceptación

Incluye la guía global, `questions.md`, la referencia nueva, `verification.md`, `delivery-decisions.md`, `plan-review.md`, `flow-plan/SKILL.md`, `flow-build/SKILL.md`, `flow-research/SKILL.md`, `reporting.md`, las páginas `site/` de research, plan, build y close, y `CHANGELOG.md`.

Exclusiones:
- Declarar `Delivery`, `Review` o `UI verification` en los proyectos de cliente: pertenece a la sesión de cada proyecto.
- Los gates obligatorios de validación de UI y la pregunta de cierre no cambian.
- Quedan como observación, con un caso cada uno:
  - La delegación desatendida que terminó con "continúa" (cam Codex).
  - El hallazgo incidental propuesto a mitad del build (cam Codex D36).
  - Los detalles de estilo preguntados al cierre (carvento F5–F8).
- `README.md` y `llms.txt` no describen estas reglas y no cambian.

Restricciones que deben seguir cumpliéndose:
- `go test ./...` en verde.
- Cada `SKILL.md` en 16 000 bytes o menos.
- Ninguna regla con dos hogares (AGENTS.md, «Guidance structure»).
- El PR declara el cambio de tamaño de `global.md` (35 750 bytes en `2bdf9357`) y la lista de lugares revisados.

Las frases de comprobación de cada criterio están en [tasks.md](tasks.md) y se cuentan con `rg -F -c`.

- AC1. Existe `content/skills/flow-plan/references/effect-authorization.md`. Define cuatro clases (local, no productivo, producción, crítica) y lo que pide cada una. Los entornos salen de `Environments`, que `project-settings.md` define así: el último es producción, los anteriores son no productivos (por ejemplo dev y QA, o gamma) y `local` nunca cuenta. La clase crítica enumera los efectos de D9-A. *Falso en la base cuando* el archivo no existe en `2bdf9357`.
- AC2. La pregunta para un entorno no productivo lista todo el conjunto con cómo se elimina al final. Ofrece como opción buscar primero cuentas o registros existentes, diciendo que esa búsqueda toma tiempo. *Falso en la base cuando* ningún archivo de `content/` ofrece reutilizar insumos existentes.
- AC3. Una respuesta autoriza su clase para la tarea. Un efecto posterior de la misma clase y entorno dentro del conjunto, o de la misma clase de producción sobre el mismo destino, no genera una pregunta nueva. Un entorno o destino distinto, un efecto crítico o pasar un tope sí la generan. Un gate que necesita esos efectos los pregunta en la misma pregunta que el gate. *Falso en la base cuando* `content/` no contiene esa cobertura.
- AC4. `global.md` ya no dice «Before a write to shared or production data, confirm the exact records». En su lugar remite a la referencia nueva y enuncia completo su invariante mínimo: las cuatro clases y que un efecto crítico siempre lleva su propia pregunta, aunque el usuario lo nombre. *Falso en la base cuando* la frase vieja está en `global.md:28`.
- AC5. `verification.md` pide, en la ronda de entrega y revisión, cada efecto que la tarea y su verificación necesitarán en cada entorno que alcanza su alcance, clasificado según la referencia nueva. Cuando el trabajo pasa a un entorno que no cubría, incluso por una promoción, que solo cubre la cadena de entrega, pide todo lo de ese entorno al empezar, antes de su primer efecto. Ya no dice «decisions they do not foresee are still asked when they arise». Lo no previsto se pregunta al surgir, junto con todos los efectos que necesite ese mismo paso. *Falso en la base cuando* la frase vieja está en `verification.md:26`.
- AC6. En `flow-build/SKILL.md`, el trabajo que no termina en una entrega Git también pide las autorizaciones previstas antes de la primera edición. *Falso en la base cuando* `flow-build/SKILL.md:20` omite las autorizaciones previstas.
- AC7. `questions.md` dice tres cosas que hoy no dice:
  - Las autorizaciones o elecciones de varios pasos de un mismo objetivo se preguntan como una sola pregunta con todo el conjunto, aunque los pasos dependan entre sí.
  - Antes de preguntar se comprueba si ya cubre la respuesta una decisión vigente: una respuesta anterior, los efectos ya autorizados, o una decisión registrada en el plan o en la memoria.
  - Si ya está cubierta, se actúa y se dice qué respuesta la cubre.

  *Falso en la base cuando* `questions.md` solo agrupa preguntas independientes y no menciona plan ni memoria.
- AC8. `plan-review.md` aplica sin preguntar los hallazgos que tienen una opción convencional y no cambian el alcance, un contrato ni un criterio. Los anota en el progreso como decisiones tomadas sin preguntar, y el resumen del plan en `flow-plan/SKILL.md` los muestra. *Falso en la base cuando* `plan-review.md:37` solo dice «Ask the user only for consequential choices».
- AC9. `delivery-decisions.md` dice dos cosas:
  - Un PR de corrección del cambio del mismo ticket pertenece a la misma entrega y no pide un merge aparte.
  - Una condición que el usuario pone a un permiso dura solo el alcance que él declaró, y un traspaso de sesión la lleva con ese alcance.

  *Falso en la base cuando* `delivery-decisions.md:24` no lo menciona.
- AC10. En la investigación, cada opción de ruta que actúa sobre el trabajo dice si abre un ticket, y elegirla lo autoriza. Hay además una opción de registrar el ticket y parar. `flow-research/SKILL.md`, `global.md:76` y `reporting.md:98` dicen lo mismo, y ya no existe la pregunta de ticket aparte. *Falso en la base cuando* `flow-research/SKILL.md:38` pide el ticket «as its own question».
- AC11. Las páginas `site/` de research, plan, build y close y el `CHANGELOG.md` describen el comportamiento nuevo sin copiar las reglas, y el sitio compila. *Falso en la base cuando* `site/src/content/docs/flows/research.md:47` dice «A ticket offer can join the same round as its own numbered question».

- AC12. En modo desatendido, un efecto crítico no autorizado de antemano para ese destino y ese tope se salta y se reporta con destino, costo y reversión (D10-A). La sección Effects de `unattended-delegation/SKILL.md` lo dice completo, y `effect-authorization.md` y `global.md:28` remiten a ella. *Falso en la base cuando* `unattended-delegation/SKILL.md` permite crear recursos en la nube y migraciones sin distinguir gasto ni autorización previa.

**Not yet specified:** nada.
