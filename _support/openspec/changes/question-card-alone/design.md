# Design

## Approaches

Cómo se quita la doble presentación (decidido en la investigación):

- **A, elegida — tarjeta sola; texto solo si vuelve sin respuesta.** Quita la lectura doble en las rondas contestadas por la tarjeta (el 96 % medido). Cuesta un mensaje más en las tarjetas sin contestar.
- **B — texto reducido a una línea por decisión.** Conserva la referencia en el chat, pero es el formato que los agentes incumplieron entre el 2026-10-11 y el 2026-11-04.
- **C — mantener la regla y limitar el tamaño de la ronda.** Ataca la espera y deja la duplicación.

Cómo corre la prueba sin integrar (D1):

- **A, elegida — rama desplegada en todos los hosts** con `hive update --source <checkout> --rev feat/question-card-alone`. Mide el uso real. Obliga a que los refrescos salgan de la rama mientras dure.
- **B — instalación aislada** con `--home DIR` para sesiones elegidas. No toca la instalación normal; pocas sesiones y un inicio de sesión aparte por host.

## Decisions

- **D6 A — línea temporal en `AGENTS.md`.** El `AGENTS.md` raíz manda refrescar desde `development` tras cada merge y pesa más que un `HIVE.local.md`; con 12 merges en tres días la regla de prueba duraría horas instalada. La línea va en «Refreshing a local installation», nombra la rama `feat/question-card-alone` y vale mientras exista sin integrar (`git branch --no-merged development` la lista), de modo que cubre también el tiempo entre el despliegue y la apertura del pull request. El refresco no escribe en la rama ni en su checkout, donde otra sesión puede estar trabajando: une `development` con la rama en un worktree temporal y separado (`git worktree add --detach`), despliega desde ahí y lo retira; el commit de unión no se referencia ni se publica. Si la unión da conflicto: `git merge --abort`, retirar el worktree temporal, no desplegar, avisar y dejar la instalación como está. No es contenido distribuido: no se despliega y no entra en la regla de recurrencia de «Measurement».
- **D7 A — Codex en el criterio.** Dos de las tres tarjetas sin contestar deben ser de Codex, porque su plantilla del modo Default empuja a no volver a preguntar.

- **D2 A — solo el informe antes de la tarjeta.** El texto previo lleva la línea de contexto, el informe o la explicación que la ronda necesita y la línea de espera (qué se espera y cuántas decisiones quedan después, como ya pide «Decisions carry labelled context»); no lleva línea de título, `Situación:`, opciones ni `Recomiendo`. La tarjeta lleva la situación en la pregunta y el efecto de cada opción en su descripción, como ya exige «Write the card to stand alone».
- **D3 A — criterio de la prueba.** Ronda perdida: una tarjeta sin contestar tras la cual el agente no escribe las decisiones abiertas en formato completo, o una ronda en la que el usuario pide el contexto que le faltó para contestar. Codex con dos rondas perdidas vuelve a la regla actual mediante una excepción por host escrita en la misma rama.

## Proposed shape

La regla nueva, en los tres puntos de «Questions and routes» que hoy la sostienen:

1. **Dónde va la ronda.** Condición observable por el agente en su propia sesión: la herramienta de preguntas espera la respuesta y cada opción tiene un campo de descripción.
   - Se cumple: el texto lleva la línea de contexto, el informe y la línea de espera; después, sin otra llamada entre medias, la tarjeta pregunta la ronda (tarjetas consecutivas cuando no cabe en una). El turno termina con la tarjeta.
   - No se cumple (sin herramienta que espere, herramienta sin descripción por opción, ejecución no interactiva, child, delegación desatendida): la ronda completa va en el texto, y la tarjeta sin descripciones la sigue cuando existe, con las mismas reglas de tarjeta (una pregunta por decisión, tarjetas consecutivas, el turno termina con ella); si esa tarjeta vuelve sin respuesta, el agente añade una línea que nombra las decisiones abiertas y no reescribe la ronda.
2. **La tarjeta se sostiene sola.** Sin cambios de formato; pasa de respaldo a único portador, así que la pregunta lleva la situación entera.
3. **Tarjeta sin respuesta.** Vacía, descartada, rechazada (también cuando el host devuelve el descarte como error), contestada solo en parte, con petición de conversar o con texto libre que cambia la ronda: termina las tarjetas de esa ronda y el agente contesta primero lo que el usuario dijo y después escribe en un mensaje cada decisión y cada hallazgo que sigue abierto, incluidos los de tarjetas aún no preguntadas, en formato completo, con la línea de espera que dice que «ok» aplica cada opción con estrella y que una respuesta como `D2 B` cambia una. Si el host cortó el turno al descartarse la tarjeta, con o sin resultado de herramienta, lo hace en respuesta al siguiente mensaje del usuario, tras contestar ese mensaje, salvo que conteste cada decisión abierta por su ID o deje la ronda de lado. Un «ok» solo aplica las opciones con estrella de decisiones que el usuario vio. Las rondas siguientes vuelven a la tarjeta.

Lo que se retira: «A round's first card without the round's text before it is the failure». La línea única tras una tarjeta cerrada («add one line …») se conserva solo donde la ronda ya está en el texto. Lo que se mantiene: una ronda solo en texto donde hay tarjeta que espera sigue siendo el fallo, y también un mensaje después de la última tarjeta que no sea la respuesta a una tarjeta sin contestar.

Punteros que citan el orden y cambian con la regla: `content/skills/flow-plan/SKILL.md` paso 4, `content/skills/flow-verify/SKILL.md` paso 2 y `content/skills/flow-verify/references/viewer-publication.md` «In the card».

## Impact

- Lee la regla toda sesión de los siete hosts; el cambio de conducta se limita a los hosts cuya tarjeta espera y lleva descripciones: Claude Code, Codex (modo Plan, y Default con la función activada), Grok y OpenCode, los cuatro que se ofrecieron al usuario. Pi con su extensión de preguntas cumple la misma condición y queda dentro por capacidad (supuesto dicho al usuario); la prueba no lo mide: no tuvo rondas en la ventana.
- Tres vigilancias de Engram describen como fallo lo que la regla nueva pide o deja de pedir: `watch/card-without-visible-text`, `watch/card-without-message` y `watch/no-card-after-card-timeout`; se revisan al cerrar la prueba.
- La fila de la rúbrica «An empty or timed-out card gets one line, not the round repeated» pasa a comprobar lo contrario.
- No toca producción ni sistemas externos: la instalación local es de esta máquina.

## Recovery

- Volver a la regla actual en la instalación: `hive update --source <checkout> --rev development` y `hive apply`. Las sesiones abiertas cargan la guía al reiniciar; Codex y OpenCode la recargan solos.
- La rama no se integra hasta el cierre de la prueba, así que la regla no necesita reversión en `development`. La línea temporal de `AGENTS.md` sí está en `development`: si la prueba se abandona, un pull request la retira; si la rama se integra, su último commit antes de la integración la borra.
- Lo que no se deshace: las rondas que una sesión ya preguntó con la regla nueva.
