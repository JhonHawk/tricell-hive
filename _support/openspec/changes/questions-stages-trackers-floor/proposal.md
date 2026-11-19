# questions-stages-trackers-floor — Preguntas en una fuente, bloque entre etapas, trackers y piso de pruebas

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build · plan revisado el 2026-11-19 |
| Tracker | ninguno |
| Git | `JhonHawk/tricell-hive`, base `development` en `262b64e2`; ramas por unidad: `feat/audit-cam19-criteria`, `feat/questions-and-stages`, `feat/tracker-home`, `feat/verification-floor` |
| Playbook | feature |

## Decisions

2026-11-19
- **D29** — ¿Corregimos ya que Codex pierda el formato de opciones? Sí, con un único origen para todo lo de preguntas: afinar y limpiar lo actual, no volver al texto anterior a skill-leaves.
- **Re-análisis de la sesión Codex del 2026-11-19, punto 1** — ¿Cómo se pasa entre etapas? Cada paso entre etapas muestra el bloque de siguiente paso; contestar una pregunta o aceptar una propuesta solo define o redirige alcance y nunca arranca un flujo. Se conserva la continuación directa de investigación a build solo cuando la propia petición del usuario nombra el cambio y su destino.
- **Re-análisis de la sesión Codex del 2026-11-19, punto 2** — ¿Llevan ID las preguntas obligadas de la sesión? No: entrega, mecanismo de revisión y revisión adversarial llevan un encabezado fijo con el resto del formato; los `D<n>` quedan para las decisiones del trabajo.
- **Re-análisis de la sesión Codex del 2026-11-19, punto 3** — ¿Cómo se mueven los tickets? Un hogar de trackers con reglas cortas y referencias por servicio (Jira, Linear, GitHub); los estados permitidos se descubren y se guardan en Engram bajo `tracker/states/<tracker>/<clave>`.
- **Re-análisis de la sesión Codex del 2026-11-19, punto 4 (piso A)** — ¿Qué pruebas corren siempre? Un piso por tipo de cambio: la revisión de código corre en todo cambio no mecánico; la evidencia (salida de la prueba fallando y pasando, capturas y revisor UX en UI) va en todo modo de entrega, `hold` incluido; un bloque «Pruébalo tú» cierra toda entrega que se pueda ver o ejecutar; la revisión de seguridad se recomienda en cambios de riesgo en vez de quedar solo a petición.
- **D31 A** — ¿Qué forma tiene el hogar único de preguntas? Una hoja `hive-rule-questions` de 7,000 bytes como máximo que absorbe las cuatro hojas actuales, con una línea en el índice.
- **D32 (respuesta libre)** — ¿Cómo pregunta Codex en modo Default? Con la tarjeta que espera respuesta (`request_user_input`) y detiene el flujo hasta que se conteste, aunque las instrucciones de Codex la reservan para preguntas opcionales.
- **D33 A** — ¿Dónde vive el piso de pruebas? En una hoja nueva `hive-rule-verification-floor` que leen plan, build, verify y los playbooks, y que apunta a los hogares existentes sin repetirlos.
- **Revisión adversarial (respuesta libre)** — ¿Corre `interrogate` sobre este cambio? Sí; Hive analiza los agentes de otros proveedores instalados y con sesión, y el usuario elige cuál. Esa elección por host pasa también a la regla.
- **Mensaje del usuario** — Documentar el ticket con sus evidencias va antes de cualquier limpieza, para aprovechar los assets.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «Creo que deberíamos tener un principio sobre lo que tiene que ver con preguntas, un único origen, no? y no es necesario que retomemos las reglas previo a esto, si no que podemos afinar/limpiar e implementar» (D29)

> «cuando pasemos entre las siguientes etapas, debe presentarse el bloque de siguiente paso […] no debe ser el contestar una pregunta. Eso solo sirve para definir alcances o redireccionarlo»

> «algo que NO debería llevar índice, son las preguntas obligadas de la sesión, método de entrega, modo de revisión, etc.»

> «esta skill podría tener reglas sencillas y referencias específicas por servicio, jira, linear, github, etc y qué podemos aprovechar de engram? un tópico específico tal vez tracker-status»

> «ahora y sobre todo basándonos en pstack deben ser más rigurosas las revisiones»

> «documentar los tickets/issues con evidencias va antes de cualquier limpieza (para aprovechar los assets)»

## Objective

Hoy, en la sesión Codex del 2026-11-19 en un proyecto con Jira (rollout `01a1f2f5`), el agente entró a build al responder «1» sin bloque de siguiente paso, preguntó la entrega en viñetas sin formato y sin tarjeta, no movió el ticket al fallar la transición, y verificó un cambio de flujo de UI sin capturas ni revisor UX ni pasos para probarlo. El auditor (`--audit-session`) no detecta ninguna de esas fallas. Al terminar: el auditor las detecta; las reglas de preguntas viven en una hoja; el paso entre etapas siempre muestra el bloque; los trackers tienen hogar y caché de estados; el ticket se documenta con evidencias antes de limpiar; y un piso de pruebas por tipo de cambio fija revisión, evidencia y «Pruébalo tú» en todo modo.

## Scope and acceptance

Dentro: guía distribuida en `content/`, criterios y fixtures de `tests/pilot`, pruebas de `tests/content`, `README.md`, `llms.txt`, `site/`, `_support/docs/` y el delta de `workflow-lifecycle`. Fuera: cambios del manager (`tooling/`, `integrations/`), pilotos de modelo, y reescribir las hojas que el piso solo enlaza.

- AC1 — `--audit-session` marca `fail` en un criterio nuevo cuando el agente lee la skill de un flujo (`flow-build`, `flow-plan`, `flow-figma-design`) en el turno que sigue a un mensaje del usuario, si el mensaje anterior del agente terminó en una pregunta o una propuesta sin bloque `▶` ni comando de flujo; la continuación directa sobre la petición original queda fuera porque no la precede un mensaje del agente que pregunte · false today because el rollout `01a1f2f5` no da ningún `fail` · Observed in: `go run ./tests/pilot --audit-session` sobre el rollout `01a1f2f5` y su fixture
- AC2 — `--audit-session` marca `fail` cuando una ronda de decisiones, con título `D<n>` o con encabezado fijo de pregunta de sesión, va en texto sin línea de situación ni `(★)` en un turno sin tarjeta que espere · false today because `text_round_without_card` da `not_observed` en el rollout `01a1f2f5` · Observed in: `go run ./tests/pilot --audit-session` sobre el rollout `01a1f2f5` y su fixture
- AC3 — `--audit-session` marca `fail` cuando una sesión edita un archivo de interfaz (`.tsx`, `.jsx`, `.vue`, `.svelte`, `.css`, `.html`) y termina con un reporte final sin haber lanzado `hive-review-ux` · false today because no existe ese criterio · Observed in: `go run ./tests/pilot --audit-session` sobre el rollout `01a1f2f5` y su fixture
- AC4 — La guía instalada tiene una sola hoja de preguntas, `hive-rule-questions`, de 7,000 bytes como máximo, y ya no existen `hive-rule-decision-format`, `hive-rule-question-channel`, `hive-rule-question-text` ni `hive-rule-card-unanswered`; el índice de `global.md` tiene una línea de preguntas · false today because existen las cuatro hojas y cuatro líneas · Observed in: `~/.agents/skills/` tras `hive update`
- AC5 — En una sesión real de Codex en modo Default posterior al despliegue, una ronda de decisiones va por `request_user_input` y el turno espera la respuesta · false today because la sesión del 2026-11-19 preguntó en prosa sin tarjeta · Observed in: auditoría de una sesión real de Codex posterior al despliegue
- AC6 — Responder una pregunta o aceptar una propuesta nunca arranca un flujo: `hive-mode`, el bloque de siguiente paso y los finales de reporte lo dicen, y la continuación directa a build exige que la propia petición nombre el cambio y su destino · false today because `hive-mode` paso 5 no excluye una respuesta a una pregunta · Observed in: guía instalada y el criterio de AC1 en `pass` sobre una sesión real posterior
- AC7 — Las preguntas de entrega, de mecanismo de revisión y de revisión adversarial llevan encabezado fijo sin `D<n>`, y la pregunta de revisión ya no recomienda «sin revisión» salvo en un cambio mecánico · false today because `hive-rule-delivery-ask` titula `D<n> — Which code review…` y recomienda `No dedicated review` · Observed in: hoja instalada y fixtures de `review_question_native_option`
- AC8 — Antes de una revisión adversarial, Hive lista los agentes de otros proveedores instalados y con sesión, y el usuario elige uno · false today because `hive-rule-interrogate` fija «Codex from Claude Code, Claude from Codex» · Observed in: hoja instalada y la ronda de este plan
- AC9 — Antes de mover un ticket se leen en vivo sus transiciones disponibles; si una falla se toma la equivalente; hay una hoja `hive-rule-trackers` con referencias de Jira, Linear y GitHub para mover, comentar y adjuntar, y `create-issue` conserva su procedimiento de creación enlazando esas referencias · false today because ninguna regla dice cómo mover y la sesión del 2026-11-19 se rindió ante «No allowed transitions found» · Observed in: guía instalada y una sesión real que mueva un ticket de Jira con estados en español
- AC10 — `global.md` permite escribir en Engram `tracker/states/<tracker>/<prefijo>` (el prefijo de clave del proyecto, más el tipo de ticket cuando su flujo difiere) con los nombres y categorías de los estados, al descubrirlos o tras una transición fallida; la hoja lo usa para elegir el estado y nunca en lugar de la lectura en vivo · false today because «Persistent memory» no permite esa escritura · Observed in: `global.md` instalado y la entrada de Engram creada en una sesión real
- AC11 — Al cerrar, el ticket se documenta con sus evidencias (capturas, salida de pruebas) antes de cualquier limpieza, y `global.md` autoriza ese comentario · false today because `flow-close` limpia en su paso 5 sin paso de evidencias y la autorización no lo nombra · Observed in: `flow-close` instalado y un cierre real
- AC12 — La hoja `hive-rule-verification-floor` fija por tipo de cambio (mecánico, bugfix, feature, UI, de riesgo, rendimiento) lo que siempre corre, la enlazan `flow-plan`, `flow-build`, `flow-verify` y los playbooks, y la revisión de seguridad se recomienda en cambios de riesgo · false today because no existe y la seguridad corre «solo a petición» en 10 lugares · Observed in: guía instalada y `tests/content/leaf_links_test.go`
- AC13 — El reporte final lleva la salida de la prueba de regresión fallando y pasando de un defecto, y un bloque «Pruébalo tú» (URL o comando, fila de la cuenta, pasos, resultado esperado) cuando el cambio se puede ver o ejecutar, en todo modo de entrega, `hold` incluido · false today because `hive-rule-close-report` no pide ninguno de los dos · Observed in: hoja instalada y la auditoría de una entrega `hold` real
- AC14 — `README.md`, `llms.txt` y las páginas del sitio describen la revisión en todo cambio no mecánico y la seguridad recomendada en cambios de riesgo · false today because `README.md:441` dice «a security review only when you ask for it» · Observed in: `README.md` y `site/src/content/docs/flows/`

Los criterios que se observan en sesiones reales (AC5, la mitad real de AC6, AC9, AC10, AC11 y AC13) cierran cada unidad como no verificados; T9 los observa tras el despliegue. Las unidades siguientes no esperan esas observaciones.

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Rollout Codex `01a1f2f5-7f54-7791-bef9-b333778d6a43` (Codex 0.170.0, `gpt-6.2-sol`, modo Default): lectura de hojas a las 08:40, mensaje D1/D2 a las 08:43, `request_user_input` ofrecida y no usada, transición `In Progress` fallida, sin `spawn_agent`.
- `go run ./tests/pilot --audit-session` sobre ese rollout en `262b64e2`: todos los criterios `pass` o `not_observed`.
- Tamaños: las cuatro hojas de preguntas suman 12,873 bytes; `hive-rule-delivery-ask` 3,993 (7 libres); `hive-rule-tests` 3,997; `global.md` 30,800; el tope de hoja es 4,000 y 7,000 con excepción nombrada (`tests/content/leaf_skill_test.go`).
- Sesiones Codex anteriores: `request_user_input` usada en modo Default (dos proyectos, 2026-11-17 y 2026-11-18); formato con `Situación:` y `(★)` en texto (un proyecto, 2026-11-17, Codex 0.169.0). Transiciones fallidas recuperadas con el nombre en español o leyendo `issue/<key>/transitions` (2026-11-07, 2026-11-11, 2026-11-18).
- Nombre de herramienta de host en contenido distribuido: La sesión del 2026-11-19 muestra que la regla por capacidad («la herramienta nativa que espera, cuando el host la permite») no llevó a Codex a la tarjeta, porque su modo Default la limita a preguntas opcionales; por eso la hoja nombra `request_user_input`. T2 verifica antes la documentación oficial vigente de Codex sobre ese modo.
- Documentación oficial de Codex (Context7 `/openai/codex`, 2026-11-19): en modo Default, `request_user_input` existe solo con la función `default_mode_request_user_input` (en desarrollo, apagada por defecto; activa en esta máquina), y su manejador la marca bloqueante solo en modo Plan; en Default la tarjeta puede cerrarse sola, lo que cubre la regla de tarjeta sin respuesta.
- Agentes de otros proveedores en esta máquina: `codex` 0.170.0 con sesión (ChatGPT), `cursor-agent` con sesión, `grok` 1.0.89 y `agy` 1.5.11 sin estado de sesión verificado; `pi` y `opencode` instalados.

## Effects

None: the tasks take no effect on an external system.

## Delivery

Respuestas de la sesión en este repositorio: entrega `interactive` (parada de validación antes de cada push) y revisión con Claude Code `/code-review`. Revisión adversarial sí, con Codex (OpenAI, `codex` 0.170.0 con sesión de ChatGPT), elegido por el usuario el 2026-11-19; autoriza el gasto de su cuota para este trabajo, una corrida por unidad y otra tras las correcciones. Cuatro PRs en orden E1 → E4. No autoriza releases ni pilotos.

## Review status and progress

2026-11-19 — Plan escrito; revisión `hive-review-plan` lanzada.
2026-11-19 — Revisión del plan: 2 P1 y 7 P2 aplicados (auditor y fixtures que nombran las hojas borradas; encabezados fijos visibles al auditor; criterios AC1 y AC3 observables; deltas que contradecían; presupuesto de bytes; frases de seguridad; observación posterior al despliegue; clave de la caché; `create-issue` conserva su procedimiento) y los P3 (ruta de `rule-enforcers.md`, orden entre `ticket-state` y `trackers`, nombre de herramienta de Codex, sin secretos en adjuntos).
