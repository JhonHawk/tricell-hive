# Design

## Approaches

Para los eventos de subagente de Grok se presentaron tres enfoques; los tres sacan al hijo del mapa y dejan al padre como único productor.

- **Patrón Codex (elegido, D1 A).** `SubagentStart` (que trae el `sessionId` del padre) se convierte en `subagent_started` con el `subagentId`; todo otro evento con `subagentType` se descarta. Reutiliza el patrón de `integrations/codex/hook_event.go` y no agrega estado. Costo: el visor no ve el fin del subagente en Grok, lo que solo afecta la última actividad.
- **Inicio y fin por el mapa.** Igual, más `SubagentStop` atribuido a la única sesión viva del mapa del proceso. Simétrico, pero el traductor tendría que leer el mapa.
- **Enlace por subagente.** Guardar `subagentId → padre` al iniciar y leerlo al terminar. Preciso, pero agrega estado nuevo con limpieza propia.

Para Codex se presentaron la variable `CODEX_SESSION_ID` con prueba en vivo (elegida, D2 A), la misma sin prueba en vivo, y fallar cerrado sin variable (el daemon dejaría de publicar).

## Decisions

- **D1 A** — patrón Codex en Grok. Elegido por ser el arreglo mínimo con un precedente en el código.
- **D2 A** — `CODEX_SESSION_ID` con una ejecución real de `codex exec`. Es la única fuente que separa los hilos del daemon; la prueba cubre que no esté documentada.
- **Decisión menor (sin preguntar):** con varias variables de identidad, el resolver elige la del host del proceso más cercano. Sin esto, agregar `CODEX_SESSION_ID` volvería ambiguo un `claude` lanzado desde una terminal de Codex.
- **Decisión menor (sin preguntar, hallazgo de la revisión del plan):** cuando el host más cercano no tiene variable entre las presentes, se ignoran las heredadas y se usa su mapa. La alternativa, rechazar, haría que un Grok lanzado desde una terminal de Codex dejara de publicar (hoy resuelve por su mapa). El mapa no es una fuente nueva y sigue sujeto a su regla de una sola sesión viva.

## Proposed shape

**Grok (`integrations/grok/hook_event.go`).** Después de construir `base` y normalizar el evento:

- con `subagentType` presente: si el evento es `subagentstart`, devolver `base.Subagent(SubagentStarted, ID(subagentId))` (sesión padre); en cualquier otro caso, `nil`;
- sin `subagentType`, `subagentstart` y `subagentstop` devuelven `nil` (un subagente sin tipo no se atribuye);
- el resto sigue por `hostevent.Generic(base, norm, true)`.

El comentario del archivo deja de decir que los subagentes comparten la sesión del padre. `integrations/grok/hook_register.go` agrega `SubagentStart` a `sessionEventNames`.

**Resolver (`tooling/cli/view_identity_resolve.go`).**

- `viewIdentityVariables` agrega `{"CODEX_SESSION_ID", "codex"}`; `CODEX_THREAD_ID` sigue sin ser fuente (es el id propio del subagente).
- Se recorre la ascendencia y se toma el host del primer proceso con mapa válido o que sea un host CLI, reconocido con `processHost` (nombre del comando, y argumentos para `node`, `bun` y `deno`).
- Si ese host tiene variable entre las presentes, se usa esa. Si no la tiene, se ignoran las heredadas y se resuelve por su mapa con la regla actual. Sin host reconocido y con más de una variable, el resultado sigue siendo ambiguo.
- `confirmVariableHost` también reconoce el host con `processHost`.

**Spec.** Delta MODIFIED de «Resolved native identity and current work» en `walkthrough-viewer`:
- la variable se acepta por documentación oficial, o por código publicado más una ejecución observada;
- con varias variables manda el host más cercano, y un host sin variable resuelve por su mapa;
- los subagentes de Claude Code, Codex y Grok nunca entran al mapa (Cursor queda fuera hasta verificarlo).

## Impact

- Los mapas de Grok dejan de registrar sesiones de subagentes; el timeline de una work de Grok gana eventos `subagent_started` (antes, actividad genérica).
- Cualquier shell de Codex 0.167 resuelve por variable: el app-server, `codex exec` y la TUI. Los subagentes de Codex comparten la sesión raíz, así que publicarían como el padre; la guía ya reserva la publicación al padre.
- Los hosts anidados con dos variables pasan de rechazo a resolución por el host más cercano.
- Nada cambia en OpenCode, Pi y Claude Code salvo la selección entre varias variables; Cursor sigue igual.
- Ninguna escritura de datos persistidos: los mapas viejos con entradas de subagentes ya no se leen (proceso muerto) o se reemplazan al iniciar otra sesión.

## Recovery

Revertir el commit restaura el comportamiento anterior; no hay migración ni datos que reparar. Una versión de Codex que quite `CODEX_SESSION_ID` vuelve a resolver por el mapa, como hoy.
