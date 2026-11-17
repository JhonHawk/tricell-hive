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

- **D6 A — línea temporal en `AGENTS.md`.** El `AGENTS.md` raíz manda refrescar desde `development` tras cada merge y pesa más que un `HIVE.local.md`; con 12 merges en tres días la regla de prueba duraría horas instalada. La línea va en «Refreshing a local installation», nombra la rama `feat/question-card-alone` y vale mientras el párrafo esté presente en `AGENTS.md`, bajo el marcador del mantenedor, de modo que cubre también el tiempo entre el despliegue y la apertura del pull request. El refresco no escribe en la rama ni en ningún checkout: construye la unión con `git merge-tree --write-tree` y `git commit-tree`, despliega con `--rev` en ese commit y, solo si el despliegue termina bien, lo deja localizable en la referencia local `refs/hive/trial-join`; la unión no se publica. Si la unión falla, no se despliega contenido y se informa la salida del comando; el paso del binario no cambia. (La primera versión usaba un worktree temporal; la revisión de código la descartó por sus rutas de limpieza.) No es contenido distribuido: no se despliega y no entra en la regla de recurrencia de «Measurement».
- **D7 A — Codex en el criterio.** Dos de las tres tarjetas sin contestar deben ser de Codex, porque su plantilla del modo Default empuja a no volver a preguntar.

- **D2 A — solo el informe antes de la tarjeta.** El texto previo lleva la línea de contexto, el informe o la explicación que la ronda necesita y la línea de espera (qué se espera y cuántas decisiones quedan después, como ya pide «Decisions carry labelled context»); no lleva línea de título, `Situación:`, opciones ni `Recomiendo`. La tarjeta lleva la situación en la pregunta y el efecto de cada opción en su descripción, como ya exige «Write the card to stand alone».
- **D3 A — criterio de la prueba.** Ronda perdida: una tarjeta sin contestar tras la cual el agente no escribe las decisiones abiertas en formato completo, o una ronda en la que el usuario pide el contexto que le faltó para contestar. Codex con dos rondas perdidas vuelve a la regla actual mediante una excepción por host escrita en la misma rama. (La definición de ronda perdida vigente es la de AC8, tras D9; el corte de la prueba es el despliegue de T8.)

- **D9 — según quién dejó la tarjeta sin contestar.** Se cerró sola (la herramienta devuelve vacío o un cierre por tiempo con el turno en curso, sin rechazo, descarte ni declinación en el resultado): el agente escribe la ronda abierta en texto. El usuario la dejó sin contestar (rechazo, descarte o declinación, también como error y también donde el host devuelve una declinación tras su propio cierre; turno cortado con la tarjeta abierta; petición de conversar; respuesta a medias; texto libre que cambia la ronda): el agente no vuelca la ronda y toma el primer caso que encaje con lo que el usuario dijo, en la tarjeta o en su siguiente mensaje: (a) responde por ID o con «ok»: se aplica, y lo abierto vuelve a la tarjeta; (b) fue un error o pide las preguntas: la misma ronda otra vez por la tarjeta, con los mismos IDs; (c) dirección equivocada, cambia en qué se apoyan las decisiones, o texto libre que añade, quita o reemplaza una decisión u opción: se descartan las que toca, se replantean y se pregunta una ronda nueva con IDs nuevos; (d) pide olvidar las preguntas: se descartan y solo vuelve la que un paso posterior no pueda saltarse; (e) cambia de tema o pregunta algo que deja las decisiones como estaban: respuesta y una línea que nombra las decisiones sin contestar; se vuelven a preguntar cuando el usuario lo dice o pide continuar trabajo que no puede seguir sin ellas. Sin nada dicho aún sobre las decisiones abiertas (declinación con el turno en curso, tarjeta contestada a medias, petición de conversar sin tema): la línea de recordatorio sola, que dice cómo recuperar la tarjeta. Una decisión preguntada otra vez conserva su ID, y una línea que dice qué decisiones vuelven precede a la tarjeta que regresa. Lo que la tarjeta sí devolvió contestado vale en todos los casos. La separación entre «nadie actuó» y «el usuario canceló» es observable en el resultado de la herramienta: Codex devuelve vacío en el cierre por tiempo y corta el turno con Esc; Claude Code devuelve un rechazo y corta el turno; OpenCode devuelve el error de descarte; Grok devuelve «declined» tanto por descarte como por su cierre a los 1800 s, y ahí el agente trata el resultado como cancelación.

## Proposed shape

La regla nueva, en los tres puntos de «Questions and routes» que hoy la sostienen:

1. **Dónde va la ronda.** Condición observable por el agente en su propia sesión: la herramienta de preguntas espera la respuesta y cada opción tiene un campo de descripción.
   - Se cumple: el texto lleva la línea de contexto, el informe y la línea de espera; después, sin otra llamada entre medias, la tarjeta pregunta la ronda (tarjetas consecutivas cuando no cabe en una). El turno termina con la tarjeta.
   - No se cumple (sin herramienta que espere, herramienta sin descripción por opción, ejecución no interactiva, child, delegación desatendida): la ronda completa va en el texto, y la tarjeta sin descripciones la sigue cuando existe, con las mismas reglas de tarjeta (una pregunta por decisión, tarjetas consecutivas, el turno termina con ella); si esa tarjeta vuelve sin respuesta, el agente añade una línea que nombra las decisiones abiertas y no reescribe la ronda.
2. **La tarjeta se sostiene sola.** Sin cambios de formato; pasa de respaldo a único portador, así que la pregunta lleva la situación entera.
3. **Tarjeta sin respuesta.** Según quién la dejó sin contestar, como fija «D9» en «Decisions», más arriba; lo que sigue en este punto es la versión anterior a D9, que ya no rige. Vacía, descartada, rechazada (también cuando el host devuelve el descarte como error), contestada solo en parte, con petición de conversar o con texto libre que cambia la ronda: termina las tarjetas de esa ronda y el agente contesta primero lo que el usuario dijo y después escribe en un mensaje cada decisión y cada hallazgo que sigue abierto, incluidos los de tarjetas aún no preguntadas, en formato completo, con la línea de espera que dice que «ok» aplica cada opción con estrella y que una respuesta como `D2 B` cambia una. Si el host cortó el turno al descartarse la tarjeta, con o sin resultado de herramienta, lo hace en respuesta al siguiente mensaje del usuario, tras contestar ese mensaje, salvo que conteste cada decisión abierta por su ID o deje la ronda de lado. Un «ok» solo aplica las opciones con estrella de decisiones que el usuario vio. Las rondas siguientes vuelven a la tarjeta.

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
