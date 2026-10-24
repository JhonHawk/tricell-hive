# Revisión de experiencia de uso en planes y alcance de la regla de admisión (2026-10-24)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · validado por el usuario (D10-A) e integrado en `development` el 2026-10-24 por el [PR #128](https://github.com/JhonHawk/tricell-hive/pull/128), merge `84735609`; desplegado con `hive update` (release `a1521e3d`); registro archivado el mismo día |
| Tracker · GitHub Issues | Sin issue: sale de decisiones del usuario en la revisión de las sesiones del 2026-10-23 (D2, D4, D5, D6) |
| Git | `interactive` (D7) · integrado por merge del PR #128; rama `feat/ux-plan-review-admission-scope` eliminada tras el merge |
| Verificación | `hive-verify-task` por tarea · `go vet ./...`, `go test ./...` y build del sitio sobre el candidato final · sin revisión de código dedicada (D8) |
| Siguiente paso | Ninguno; registro cerrado |

## Solicitud

Respuesta del usuario a D2 (registrar como vigilancia los casos vistos una sola vez), 2026-10-24, tarjeta nativa:

> Si registralo, y consideremos algo de la regla, las features o requisitos nuevos no necesitan necesariamente cumplir el tema de los dos hallazgos, los huecos tampoco, las correcciones si (aunque yo puedo aprobar la corrección inmediata), en este caso creo que el plan necesita un revisor de UX o que se confirme cómo se solventara una UI, para evitar ese tipo de fricciones en el uso

Respuestas siguientes, 2026-10-24, tarjetas nativas:

- D3-A: planear los dos cambios.
- D4-A: confirma la lectura «la exigencia de dos sesiones independientes aplica solo a las correcciones de un comportamiento observado, que el usuario puede aprobar de inmediato; los requisitos y funciones nuevos y los huecos quedan fuera».
- D5-A: registro de cada dato que pide la vista más un área de revisión «Experiencia de uso» aparte.
- D6-A: una función o requisito nuevo, aunque venga de una referencia externa, se adopta con la decisión del usuario y no por repetición; parecer una buena mejora no basta.
- D7-A: entrega interactiva. D8-A: sin revisión de código dedicada.

## Objetivo

En la sesión de carvento del 2026-10-24 (Claude Code, sesión `94440d0a`, COM-645), el plan de la tarjeta de MaxiPublica volvía a pedir por unidad marca, modelo, año y versión, aunque la unidad ya guarda marca y modelo. Ninguno de los tres revisores del plan lo vio: el de Frontend/UI llegó a recomendar reutilizar la cascada marca → modelo, porque su lista de preguntas en `plan-review.md` mira patrones, reutilización, estados y accesibilidad, no qué datos pide la pantalla ni si el sistema ya los tiene. El usuario lo detectó en la primera vista, ya construido.

El cambio hace que un plan con una vista donde el usuario captura o elige datos registre cada dato que pide, de dónde sale y cuántas veces se pide, y que un revisor aparte recorra la tarea real del usuario buscando datos pedidos dos veces. Además ajusta la regla de admisión de `AGENTS.md` a lo que decidió el usuario: la exigencia de dos sesiones reales aplica a las correcciones; las funciones y requisitos nuevos, incluidos los de referencias externas, y los huecos se adoptan con la decisión del usuario.

## Alcance y aceptación

Incluido: `AGENTS.md` (sección Measurement, segundo párrafo); `content/skills/flow-plan/references/ui-planning.md` (condición de lectura y «Record the view decisions»); `content/skills/flow-plan/SKILL.md:28` (condición de lectura de UI planning); `content/skills/flow-plan/references/plan-review.md` («Choose the useful coverage»); `site/src/content/docs/flows/plan.md`; `CHANGELOG.md`.

Excluido:
- `hive-review-plan` y `hive-review-ux`: el rol de revisión de planes no tiene áreas propias (las áreas viven en `plan-review.md`) y `hive-review-ux` revisa la interfaz renderizada, no planes.
- `content/guidance/global.md`: la regla de admisión solo vive en `AGENTS.md`, que gobierna este repositorio; `global.md` no cambia de tamaño.
- Las vigilancias ya registradas de referencias externas (HF1, Q1, WA1–WA5, PR1/PR2, PM1, SEC1 y otras): la regla nueva no las reabre; cada una se puede reconsiderar cuando el usuario lo pida.
- Los hallazgos de otros proyectos (carvento: cilindrada, color y carrocería todavía duplicados; facti sin `Delivery` ni `Review`; cam): se reportan para sus propias sesiones.

Restricciones: `go vet ./...`, `go test ./...` y el build del sitio siguen pasando; la validación de enlaces de instrucciones pasa; ninguna regla queda con dos hogares (`AGENTS.md`, «Guidance structure»).

- AC1. El segundo párrafo de Measurement en `AGENTS.md` exige dos sesiones reales independientes, o la aprobación inmediata del usuario, solo a las correcciones (una regla nueva o más estricta para una situación que una regla existente ya cubre), y dice que una función o requisito nuevo, aunque venga de una referencia externa, y un hueco (una situación que ninguna regla cubre, aunque la haya revelado una sesión real) se adoptan con la decisión del usuario sin recurrencia. *Falso en la base cuando* `rg -c "gap, a situation no rule addresses" AGENTS.md` no encuentra nada.
- AC2. «Record the view decisions» de `ui-planning.md` pide, para una vista donde el usuario captura o elige datos, cada dato con el lugar donde ya vive o del que se deriva, o la razón por la que lo da el usuario, y cuántas veces lo da por registro. *Falso en la base cuando* `rg -c -i "already lives" content/skills/flow-plan/references/ui-planning.md` no encuentra nada.
- AC3. `plan-review.md` asigna un área «User experience», como encargo propio de `hive-review-plan` aparte de Frontend / UI, a todo plan que crea o cambia una pantalla, diálogo o formulario donde el usuario final captura o elige datos, o que cambia los pasos de una tarea que el usuario final hace en la interfaz; sus preguntas cubren el registro de datos (su ausencia bloquea), datos que el sistema ya tiene o puede derivar, capturas repetidas por registro, y una premisa o decisión previa de quien pidió el plan que obliga a capturar dos veces, elevada como decisión de esa persona. *Falso en la base cuando* `rg -c "User experience" content/skills/flow-plan/references/plan-review.md` no encuentra nada.
- AC4. La página del flujo Plan del sitio dice que un plan con una pantalla donde el usuario captura datos recibe además un revisor que recorre su tarea buscando datos que tendría que capturar dos veces. *Falso en la base cuando* `rg -c -i "twice" site/src/content/docs/flows/plan.md` no encuentra nada.
- AC5. `ui-planning.md` y `flow-plan/SKILL.md` piden leer UI planning también al crear o cambiar un diálogo o formulario donde el usuario final captura o elige datos. *Falso en la base cuando* `rg -c -i "dialog or form" content/skills/flow-plan/SKILL.md content/skills/flow-plan/references/ui-planning.md` no encuentra nada.

## Entrega

- Modo `interactive` (D7-A, 2026-10-24) en `tricell-hive`, base `development`: rama de trabajo, commits locales, verificación, parada antes del push para que el usuario valide el texto; tras su validación, push, PR, sin revisión dedicada (D8-A) e integración en `development` cuando pasen las comprobaciones.
- Este registro se versiona al cierre con un PR pequeño sin revisión dedicada, que integra el agente.
- Tras la integración, en este checkout de mantenedor, `hive update` despliega el contenido (`AGENTS.md`, «Refreshing a local installation»); en OpenCode el snapshot de `plan-review.md` se regenera en ese despliegue.
- `UI verification`: no aplica; el cambio del sitio es solo texto y lo cubre el build del sitio.
