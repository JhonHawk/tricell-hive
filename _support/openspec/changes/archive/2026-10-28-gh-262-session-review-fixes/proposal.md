# Push que despliega sin vista previa, prerrequisitos generables, línea de contexto y excepción de Context7

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el [PR #263](https://github.com/JhonHawk/tricell-hive/pull/263) (`8ed2e6a4`) · desplegado localmente con `hive update` |
| Tracker · GitHub Issues | • [#262 — Push que despliega sin vista previa, prerrequisitos generables tratados como bloqueo, línea de contexto omitida y contradicción sobre npx ctx7](https://github.com/JhonHawk/tricell-hive/issues/262) |
| Git | automatic · PR a `development` y merge · registro en un PR de cierre pequeño que integra el agente · sin parada humana |
| Verificación | comprobaciones `rg` por criterio · `go vet ./...` · `go test ./...` · pruebas de skills en Python · `/code-review` de Claude Code |
| Siguiente paso | ninguno; registro archivado |

## Solicitud

> Creo que ya podemos ir revisando sesiones de los diferentes proyectos y hosts, entre los temas más relevantes:
>
> - En el cambio de cam, del manual hizo todos los cambios pero no me presento como iba a quedar y subio directamente
> - En carvento, una funcionalidad no la revisó porque no habia información, cuando localmente se puede generar sin problema alguno, lo dejo como bloqueante para esa parte y aun asi cerro
>
> Esas son mis anotaciones, pero busca más temas

> lo de cam, fue con grok

> Planear H1,H2,H4,H6,H8 (Recomendada)

> no tengo tema con  Etiqueta de pausa en inglés

> en AC1, pongamos que esto solo aplica cuando la sesión solo esta trabajando con documentación, no cuando el trabajo base es sobre el proyecto real y se incluye un paso de actualización de doc, lo digo para que no nos vaya a meter fricción en el dia a dia, creo que tambien afecta como se ve AC2 Y AC3

Respuestas de la ronda de plan (2026-10-28): D2 "Excepción para Context7", D3 "automatic (Recomendada)", D4 "/code-review (Recomendada)". Durante el build: D7 "Mostrar siempre y preguntar (Recomendada)".

El issue #262 se abrió el 2026-10-28 desde esta sesión; sus criterios de aceptación son la base de los AC de abajo.

## Objetivo

Una revisión de 24 sesiones reales iniciadas después de hive-core (2026-10-28T03:18Z) en Claude Code, Codex y Grok confirmó las dos notas del usuario y encontró dos fallas más que vale la pena corregir.

- **H1.** En una sesión Grok de cam (`01a183d9`), el agente rehízo la tabla del manual y en un solo comando hizo commit y push a `master` de un repositorio `Delivery: direct-base`. Ese push desplegó en Vercel sin build, sin captura previa y sin pregunta. La regla de `direct-base` pide preguntar antes de un push que despliega "when no question is otherwise asked". El agente aplicó al rediseño el "actualicemos" que el usuario había dado para el cambio anterior. Además, la guía clasifica "documentación" como cambio mecánico aunque altere el layout, así que el agente no cargó la primera vista renderizada.
- **H2.** En una sesión Grok de carvento (`01a18466`, COM-900), el plan condicionó la caminata funcional a que "exista la cuenta" y el build terminó sin sembrar datos locales. El agente los generó en 2 minutos cuando el usuario se lo pidió. `review-and-gates.md` ya pide preparar cuentas, roles y datos semilla locales, pero `flow-build` y los marcadores del plan dicen "prerequisites unavailable" sin separar lo que el agente puede crear de lo que solo el usuario puede dar.
- **H4.** Codex omitió la línea de contexto en 181 de 202 mensajes. Leyó la excepción "a one-line progress note" como "cualquier párrafo escrito en una sola línea".
- **H8.** La guía manda consultar Context7 con la skill instalada, que corre `npx ctx7@latest`, y la revisión de dependencias exige revisar todo `npx` puntual y prohíbe `@latest`. El usuario eligió eximir esa CLI (D2-B).

El resultado esperado: esos cuatro textos dejan de admitir la lectura que causó cada falla, sin agregar preguntas ni ceremonia.

## Alcance y aceptación

**Incluye:** `content/guidance/global.md`, `flow-plan/references/delivery-decisions.md` (con la excepción declarada a `hive-core/references/effect-authorization.md`, que no se edita), `flow-walkthrough/SKILL.md`, `flow-build/SKILL.md`, `flow-plan/references/plan-criteria-and-markers.md`, `flow-build/references/plan-execution.md`, `flow-plan/references/plan-format.md`, `flow-close/references/reporting.md`, `flow-build/references/security-boundaries.md`, `site/src/content/docs/flows/build.md` y `CHANGELOG.md`.

**Excluye:**
- H6, la etiqueta de pausa en inglés: el usuario no la considera una falla.
- H3, H5 y H7 (memoria fuera de los momentos permitidos, tarjeta sin mensaje, esperas con sondeos cortos): su causa está en la configuración de hosts y plugins, y quedan como observación en Engram (`watch/memory-without-memory-md`, `watch/card-without-message`, `watch/external-wait-without-delegation`).
- Los casos sueltos H9 a H17: quedan como observación.
- `evidence-tools.md` y la regla personal `~/.claude/rules/context7.md` del usuario no cambian: con la excepción ya no contradicen la revisión de dependencias.

**Restricciones:**
- Cada `SKILL.md` y cada referencia sigue en 16 000 bytes o menos (`tests/content/skill_size_test.go`). `reporting.md` está en 15 893 B, así que su edición no debe crecer más de 107 B.
- La suite local completa sigue pasando.
- El PR declara el tamaño de `global.md` antes y después, y lista los lugares revisados (`AGENTS.md`, Guidance structure).
- Admisión (`AGENTS.md`, Measurement): H1 y H2 tienen un caso cada una y las adoptó el usuario; H4 tiene cinco sesiones independientes; H8 es una contradicción de texto. Cada commit nombra la sesión real que motiva su corrección (host, ID y fecha).

**Búsqueda de especificaciones.** Busqué en `_support/openspec/specs/` (context-delivery, documentation-site, versioned-installation, walkthrough-viewer) los términos `mechanical`, `context line`, `direct-base`, `Context7`, `ctx7`, `seed data`, `prerequisite`, `first rendered`, `one-line progress` y `dependency before install`. Leí también las prohibiciones (`SHALL NOT`, "limited to") de context-delivery y documentation-site. Las únicas coincidencias son la lista de capacidades opcionales (Context7 en versioned-installation:62), la entrega del núcleo global (context-delivery:10, 87) y el aislamiento de `site/` (documentation-site:35, 42). El cambio no altera ninguna: solo reescribe el texto de reglas existentes. No lleva delta.

- AC1. La definición de cambio mecánico de `global.md` exige que la documentación, como el texto y el estilo, no altere el layout renderizado, salvo una actualización de documentación que forma parte de un cambio de código o configuración, que sigue siendo mecánica. *Falsa en la base cuando* `global.md` lista "documentation" sin condición.
- AC2. La regla de `Delivery: direct-base` de `delivery-decisions.md` exige nombrar siempre el despliegue y preguntar antes de cada push que despliega. Solo lo cubre una respuesta dada para ese cambio que nombró el despliegue; una respuesta de otro cambio o un modo reutilizado no lo cubre, y el texto lo declara como excepción a la regla de "same kind" de `effect-authorization.md`. "Publish without asking" excluye ese push. Cada opción de entrega de un trabajo mixto nombra el despliegue que dispara el push de un repositorio de soporte, incluido el push del registro al cierre, para que esa respuesta lo cubra sin otra pregunta. Solo en un trabajo cuyos cambios, aparte de su registro, están únicamente en repositorios `Delivery: direct-base`, cuando el cambio agrega o reacomoda un layout, el agente muestra la primera vista renderizada y pregunta antes de ese push aunque una respuesta ya haya nombrado el despliegue (D7-A). *Falsa en la base cuando* el texto dice "ask before that push when no question is otherwise asked", publica sin preguntar cuando solo hay repositorios de soporte, las opciones no nombran ese despliegue y no menciona la vista.
- AC3. `flow-walkthrough/SKILL.md` dispara la captura de la primera vista antes del push que despliega un layout nuevo o reacomodado desde un repositorio `Delivery: direct-base`, solo en un trabajo cuyos cambios, aparte de su registro, están únicamente en esos repositorios, y pregunta antes de ese push si publicarlo; aprobarla autoriza el push. *Falsa en la base cuando* sus disparadores no mencionan ese push.
- AC4. `flow-build/SKILL.md`, los marcadores de `plan-criteria-and-markers.md` y el paso 4 de `plan-execution.md` solo dejan un criterio sin verificar o una tarea `blocked:` por un prerrequisito que la regla de Runtime and UI gates (`review-and-gates.md`) deja al usuario, y apuntan a ella. El paso 4 crea los prerrequisitos locales (datos semilla, servicio local) y vuelve a verificar. *Falsa en la base cuando* `flow-build/SKILL.md` dice "unavailable prerequisites leave the affected criterion unverified" sin esa distinción y el paso 4 solo marca `blocked:`.
- AC5. `plan-format.md` pide registrar, junto a los prerrequisitos de cuentas y datos, cómo los crea el build en local. *Falsa en la base cuando* solo dice "accounts/data prerequisites".
- AC6. `global.md` y `reporting.md` limitan la excepción de la línea de contexto a una nota de una sola oración. *Falsa en la base cuando* ambos dicen "one-line progress note".
- AC7. "Check a dependency before installing" exime solo el paquete npm `ctx7` con los subcomandos `library` o `docs`, corrido como indica la skill de documentación. Cualquier otro nombre o subcomando (incluido `login`) y su instalación no quedan exentos, y sus consultas siguen la regla de solicitudes a terceros. `global.md:70` declara la misma excepción con el mismo límite de instalación. *Falsa en la base cuando* ningún texto la exime.
- AC8. `site/src/content/docs/flows/build.md` describe la vista previa y la pregunta antes del push que despliega desde un repositorio de soporte, y limita el "sin parada" de `direct-base` al elegido para un repositorio de código, y su `:42` dice que una nota de avance de una sola oración va sin línea de contexto. `CHANGELOG.md` registra el cambio en Unreleased. *Falsa en la base cuando* la página dice "Under `automatic`, `direct-base` and unattended delegation there is no such stop" y ninguno menciona el cambio.

## Entrega

| Repositorio | Base | Modo | Efectos autorizados | Parada humana |
| --- | --- | --- | --- | --- |
| `tricell-hive` | `development` | automatic (D3-A, 2026-10-28) | rama `fix/gh-262-session-review-fixes` en el worktree `.claude/worktrees/gh-262-session-review-fixes`, commits, `/code-review` (D4-A) sobre el candidato local, push, PR a `development` y merge cuando pase la suite local | ninguna |

- **Registro del cambio:** se versiona una vez al cierre en un PR pequeño sin revisión dedicada que integra el agente, como en #259.
- **Instalación local:** al aterrizar en `development`, la marca `_support/workspace/maintainer.local` cuenta como pedido de refrescarla, así que corre `hive update` desde esta copia (`AGENTS.md`, Refreshing a local installation).
- **Ticket:** #262 se cierra al integrarse (los issues de GitHub no tienen estado intermedio en este repositorio).
- **Seguridad:** el cambio toca la regla de dependencias (H8), un límite de seguridad. La revisión de seguridad queda ofrecida en una línea, no como pregunta.

## Estado

Completado: integrado por el PR #263 (`8ed2e6a4`), issue #262 cerrada. Sin deltas de requisitos. Revisión del plan, verificación por tarea y revisión de código: ver `tasks.md`.
