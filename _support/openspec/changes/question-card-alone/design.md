# Design

Historical choices below remain recorded; the 2026-11-17 correction governs the current candidate wherever they conflict.

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
- **D3 A — criterio de la prueba.** Ronda perdida: una tarjeta sin contestar tras la cual el agente no escribe las decisiones abiertas en formato completo, o una ronda en la que el usuario pide el contexto que le faltó para contestar. Codex con dos rondas perdidas vuelve a la regla actual mediante una excepción por host escrita en la misma rama. (Historical definition and cutoff: the D9 version deployed by T8. The current definition is AC11, with a new cutoff pending an authorized T9 deployment.)

- **D9 — según quién dejó la tarjeta sin contestar.** Se cerró sola (la herramienta devuelve vacío o un cierre por tiempo con el turno en curso, sin rechazo, descarte ni declinación en el resultado): el agente escribe la ronda abierta en texto. El usuario la dejó sin contestar (rechazo, descarte o declinación, también como error y también donde el host devuelve una declinación tras su propio cierre; turno cortado con la tarjeta abierta; petición de conversar; respuesta a medias; texto libre que cambia la ronda): el agente no vuelca la ronda y toma el primer caso que encaje con lo que el usuario dijo, en la tarjeta o en su siguiente mensaje: (a) responde por ID o con «ok»: se aplica, y lo abierto vuelve a la tarjeta; (b) fue un error o pide las preguntas: la misma ronda otra vez por la tarjeta, con los mismos IDs; (c) dirección equivocada, cambia en qué se apoyan las decisiones, o texto libre que añade, quita o reemplaza una decisión u opción: se descartan las que toca, se replantean y se pregunta una ronda nueva con IDs nuevos; (d) pide olvidar las preguntas: se descartan y solo vuelve la que un paso posterior no pueda saltarse; (e) cambia de tema o pregunta algo que deja las decisiones como estaban: respuesta y una línea que nombra las decisiones sin contestar; se vuelven a preguntar cuando el usuario lo dice o pide continuar trabajo que no puede seguir sin ellas. Sin nada dicho aún sobre las decisiones abiertas (declinación con el turno en curso, tarjeta contestada a medias, petición de conversar sin tema): la línea de recordatorio sola, que dice cómo recuperar la tarjeta. Una decisión preguntada otra vez conserva su ID, y una línea que dice qué decisiones vuelven precede a la tarjeta que regresa. Lo que la tarjeta sí devolvió contestado vale en todos los casos. La separación entre «nadie actuó» y «el usuario canceló» es observable en el resultado de la herramienta: Codex devuelve vacío en el cierre por tiempo y corta el turno con Esc; Claude Code devuelve un rechazo y corta el turno; OpenCode devuelve el error de descarte; Grok devuelve «declined» tanto por descarte como por su cierre a los 1800 s, y ahí el agente trata el resultado como cancelación.

## Proposed shape

The global guidance's «Questions and routes» is the canonical procedure. T9, approved on 2026-11-17, adds channel choice before the existing presentation and unanswered-card rules: one channel per pending question, within the host's current tool and mode restrictions. The waiting card remains preferred where permitted; an async receipt is not an answer and cannot justify a duplicate waiting card. The async path requires independent work and the conditions in that canonical rule. Chat answers settle only the answered decisions, without asserting that a UI queue cleared.

D9's cancellation and timeout semantics remain applicable where the host leaves that response to Hive; they do not override a native required response or format. Codex Default mode's optional-question restriction, concise text for required input and permissions, and best-judgment response to an unanswered optional question are documented with source evidence in `agent-delivery.md`.

The earlier shape (2026-11-16) selected by waiting capability alone and required text after every empty timeout. That behavior is retired where it conflicts with native instructions; its observations remain historical. Skill references continue to point to the global rule instead of copying its procedure.

## Impact

- Lee la regla toda sesión de los siete hosts; la presentación solo en tarjeta aplica cuando el host permite esa pregunta y la tarjeta espera y lleva descripciones. La investigación inicial cubrió: Claude Code, Codex (modo Plan, y Default con la función activada), Grok y OpenCode, los cuatro que se ofrecieron al usuario. Pi con su extensión de preguntas cumple la misma condición y queda dentro por capacidad (supuesto dicho al usuario); la prueba no lo mide: no tuvo rondas en la ventana.
- Tres vigilancias de Engram describen como fallo lo que la regla nueva pide o deja de pedir: `watch/card-without-visible-text`, `watch/card-without-message` y `watch/no-card-after-card-timeout`; se revisan al cerrar la prueba.
- La fila de la rúbrica «An empty or timed-out card gets one line, not the round repeated» pasa a comprobar lo contrario.
- After approving the T9 local correction, the user explicitly authorized native Codex review, commit/push to the trial branch and local deployment, without merge. Real-session counting restarts after that deployment, with the existing thresholds and no model pilots.

## Recovery

Historical recovery routes below require authorization in the current session and must not restore a known conflict with native host instructions.

- Volver a la regla actual en la instalación: `hive update --source <checkout> --rev development` y `hive apply`. Las sesiones abiertas cargan la guía al reiniciar; Codex y OpenCode la recargan solos.
- La rama no se integra hasta el cierre de la prueba, así que la regla no necesita reversión en `development`. La línea temporal de `AGENTS.md` sí está en `development`: si la prueba se abandona, un pull request la retira; si la rama se integra, su último commit antes de la integración la borra.
- Lo que no se deshace: las rondas que una sesión ya preguntó con la regla nueva.
