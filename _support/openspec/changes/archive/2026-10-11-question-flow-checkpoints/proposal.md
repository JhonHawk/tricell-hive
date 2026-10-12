# Preguntas en puntos de decisión, con la tarjeta nativa como vía principal

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por PR #46, commit `16028d3575e7286cc028985cd2138c3268257add` · registro archivado |
| Tracker · GitHub Issues | Sin issue; la evidencia vive en este cambio y en la memoria `hive/research/communication-flow-2026-10-round2` |
| Git | `feat/guidance-rework` integrada por [PR #46](https://github.com/JhonHawk/tricell-hive/pull/46) · publicación autorizada D3-A/D10-A el 2026-10-11 · cierre del registro por PR separado |
| Verificación | `go test ./...` (incluye el presupuesto de `global.md`) · comprobaciones `rg` por criterio · medición en sesiones reales después del despliegue |
| Siguiente paso | Sin tareas técnicas pendientes de este cambio; la reorganización nueva continúa en su propio registro |

## Pedido

> Esto es un deep-research, sigue sin encantarme la forma de comonicación cuando deben aparecer las preguntas, quiero una propuesta que se muy organiza para el flujo de trabajo, que no dependa talvez de pasos especificos como el "muestrame las preguntas" o similar y que se incentive el uso de herramientas nativas de los host

> Considera tambien, que durante la etapa la toma de desición o respuesta a una pregunta puede cambiar el flujo de la etapa

> El sigueinte cambio es algo que ya se hacia en hive legacy y es que aunque se tiene el detalle completo del plan en una ruta de archivo, en la sesión se da un resumen junto con el ofrecimiento de flow-build

Respuestas del usuario (2026-10-11), tras la propuesta P1–P6:

- D1-A: planear la propuesta con `flow-plan`.
- D2-A: cuando hay tarjeta, el texto lleva un resumen corto (ID, pregunta y opción recomendada con su razón) y las opciones con su efecto van solo en la tarjeta.

## Objetivo

Después del cambio gh-9 (2026-10-08), cada pregunta se escribe completa en el mensaje y se repite en la tarjeta nativa. Además, la investigación termina con la oferta «¿te muestro las recomendaciones?». En las sesiones posteriores al despliegue, esa oferta apareció 5 veces en Claude Code, Cursor y Codex, y las 5 se contestaron con un simple consentimiento («sí», «a ver, muéstralas», «A1»), con esperas de 7 a 18 minutos. El mensaje repitió las opciones de la tarjeta en el 91 % de los casos. Los fabricantes desaconsejan las dos cosas: Codex pide no escribir preguntas de opción múltiple como texto, y Claude reserva la tarjeta para decisiones reales del usuario, nunca para un «¿sigo?». Hive se aparta de Codex solo en su modo normal, donde la tarjeta no espera la respuesta (ver diseño).

Este cambio ordena las preguntas en puntos de decisión fijos de cada flujo. En cada punto se hace una sola ronda con todas las preguntas que ya se pueden contestar. El mensaje lleva la explicación y un resumen por pregunta, y la tarjeta nativa lleva las opciones. La investigación termina directamente con la pregunta de ruta, sin la oferta. Después de cada respuesta, el agente dice en una línea si la respuesta cierra la pregunta, abre otra ronda o devuelve el trabajo a un flujo anterior, y qué deja de valer. Resultado esperado: menos vueltas sin decisión, sin opciones duplicadas, y más uso de la tarjeta nativa en los hosts que la tienen.

## Alcance y aceptación

Incluido:

- Las reglas de preguntas de `content/guidance/global.md`: líneas 21–23, y las referencias a la oferta en las líneas 31 y 34.
- `content/skills/flow-research/SKILL.md`: oferta y pregunta de ticket.
- La alineación de `content/skills/flow-plan/SKILL.md:16`, `content/skills/flow-plan/references/delivery-decisions.md`, el paso 7 de `content/skills/flow-close/SKILL.md` y la pausa de validación de `content/skills/flow-build/references/verification.md`.
- La sección «Native question cards» de `_support/docs/architecture/agent-delivery.md`.
- Las páginas del sitio `site/src/content/docs/flows/research.md`, `close.md` y `plan.md`.
- La entrada en `CHANGELOG.md`.

Incluido también (P1-A): `tests/content/budget_test.go` y la regla de `AGENTS.md:24` que lo menciona.

Excluido:

- La línea de pausa «Esperando tu decisión» y su etiqueta fija: siguen igual, porque funcionan y la evidencia no pide cambiarlas.
- El formato de reportes, los enlaces, las glosas de IDs y las categorías de cierre.
- La opción experimental de Codex `default_mode_request_user_input` (sigue la decisión D4-A del 2026-10-08).
- Los detectores de `tests/pilot/` y los casos de `tests/fixtures/regression/`. `closeQuestionAfterReport` y `questionAfterDetail` exigen texto antes de cada tarjeta; el resumen de D2-A y la línea antes de cada llamada adicional lo cumplen.
- Crear una release. Este cambio corresponde a una versión de parche (`CONTRIBUTING.md:57`); se propone al cierre.
- Hooks o mecanismos propios de un host.

Restricciones que deben seguir siendo verdad:

- `go test ./...` pasa. El tamaño de `content/guidance/global.md` ya no bloquea (P1-A): se informa.
- La guía distribuida nombra capacidades, no herramientas de un host. Los detalles por host viven en `agent-delivery.md`.
- Cada regla tiene un solo hogar.
- No se pierde ninguna intención de las líneas reemplazadas sin moverla o justificar su salida (lista en `design.md`).

Criterios de aceptación:

- AC1. Ningún archivo distribuido pide ofrecer mostrar las recomendaciones antes de mostrarlas. *Falso en la base cuando* `rg -n -i "offer to show|showing recommendations|show the recommendations|offer below|make the offer" content` encuentra `global.md:21,31,34` y `flow-research/SKILL.md:30,32,36`.
- AC2. La guía global dice que, con una tarjeta nativa que espera la respuesta, el mensaje lleva la explicación y una línea por pregunta (ID, pregunta y opción recomendada con su razón y su ID glosado) sin las demás opciones. Las opciones completas van en el texto siempre que ninguna tarjeta que espera lleve la pregunta. *Falso en la base cuando* `global.md:21` dice «Keep it there even when the options also live in a ticket, a document, or a native question».
- AC3. La guía global limita las preguntas a puntos de decisión. Pide resolver los datos sin preguntar, tomar las decisiones menores (nunca una consecuente ni un efecto que necesite autorización) y declararlas en una línea, y no preguntar solo para pedir permiso de seguir ni para ofrecer mostrar algo. *Falso en la base cuando* `rg -n -i "decision point" content/guidance/global.md`, `rg -n -i "minor call" content/guidance/global.md` y `rg -n -i "leave to proceed" content/guidance/global.md` no encuentran nada.
- AC4. La guía global pide hacer en una sola ronda todas las preguntas abiertas cuya respuesta no dependa de otra pregunta abierta, repartiendo en llamadas seguidas, cada una con una línea de texto antes, las preguntas u opciones que pasen del límite de la tarjeta. *Falso en la base cuando* `rg -n -i "in one round" content/guidance/global.md` no encuentra nada.
- AC5. La guía global pide que, después de cada respuesta, incluida una en texto libre, el agente diga en una línea si cierra la pregunta, abre la siguiente ronda o devuelve el trabajo a un flujo anterior, y qué deja sin validez; vuelve a preguntar algo ya respondido solo si el cambio necesita efectos no autorizados. *Falso en la base cuando* `rg -n -i "next round" content/guidance/global.md` y `rg -n -i "earlier flow" content/guidance/global.md` no encuentran nada.
- AC6. `flow-research` termina una ronda de hallazgos sobre trabajo posible con la pregunta de ruta: las rutas con su flujo (`flow-plan`, `flow-build`), seguir investigando y parar en los hallazgos, la recomendada primero; la pregunta de ticket va en la misma ronda. *Falso en la base cuando* `rg -c -i "offer" content/skills/flow-research/SKILL.md` da 3 o más y `rg -n -i "keep investigating|route question" content/skills/flow-research/SKILL.md` no encuentra nada.
- AC7. `agent-delivery.md` registra las versiones instaladas el 2026-10-11 y cómo se aplica la regla de AC2 en cada host. En Cursor, el efecto va en la etiqueta y se registra el conflicto entre binario y documentación; en Codex en modo normal, las opciones van en texto y se registra la desviación respecto de su fabricante; en Pi, la tarjeta solo existe con la extensión. *Falso en la base cuando* la sección «Native question cards» dice «Checked on 2026-10-08» y `rg -n -i "deviat" _support/docs/architecture/agent-delivery.md` no encuentra la desviación de Codex.
- AC8. Las páginas del sitio de investigación, cierre y plan describen la pregunta de ruta directa, el resumen más la tarjeta y la respuesta que devuelve el trabajo a un flujo anterior, y ninguna dice que descartar la tarjeta no pierde nada. *Falso en la base cuando* `rg -n -i "shortcut|loses nothing|shows its recommendations|offers to show" site/src` encuentra `research.md:37,39,56` y `close.md:58`.
- AC9. La prueba de tamaño de `content/guidance/global.md` informa los bytes y caracteres sin fallar nunca, y `AGENTS.md` pide indicar el cambio de tamaño en el PR en lugar de «fits the size budget». *Falso en la base cuando* `tests/content/budget_test.go:20` llama a `t.Fatalf` por encima de 46.040 bytes y `AGENTS.md:24` dice «fits the size budget».
- AC10. Al entregar un plan listo sin implementar, `flow-plan` resume el plan en la conversación: objetivo, decisiones vigentes con su ID, una línea por tarea, verificación, entrega, lo que queda abierto y la ruta del plan. Termina con la pregunta de ruta para empezar `flow-build` o parar. *Falso en la base cuando* `rg -c -i "summar" content/skills/flow-plan/SKILL.md` da 0 y `flow-plan/SKILL.md:61` solo pide «state the concrete next step».

Sesiones reales que respaldan el cambio, citadas en `design.md`: `b136f225`, `bf305ba1` y `1528596c` (Claude, hive); `2812ca14` (Claude, carvento); `b6185612` (Cursor, carvento); `7d3001ac` y `13035f46` (Cursor, cam); `070aac31` (Codex, dened); `b8c6aeeb` y `5aa1a929` (Claude, 2026-10-03).

## Entrega final

El usuario autorizó integrar la rama previa por PR a `development` y publicar los registros por un PR de cierre (D3-A, integración previa) al iniciar la construcción nueva (D10-A). El PR #46 se fusionó el 2026-10-11 local, commit `16028d3575e7286cc028985cd2138c3268257add`, con HEAD revisado `87290b9e607dc120d2ba050f80c23db9296cce74`. Se preservó la historia de commits y se corrigieron las observaciones de revisión sobre decisiones de revisión ya resueltas, opciones de entrega por tarjetas y alineación del sitio.

Controles: suite Go completa y vet sobre `d5164f71`, 23 tests Python; checks específicos de contenido/referencias tras fixes; sitio check/build y comprobación local de página Build. CodeRabbit aprobó el HEAD final. No hubo workflows de CI ni checks configurados sobre el commit de merge observados; el agregado de statuses vacío no se interpreta como una ejecución exitosa. Esta integración no instaló globalmente ni promovió a `master`; no modifica los despliegues históricos. La eficacia conductual de esta versión no se afirma por estos controles.

## Historial de entrega anterior

Decidido por el usuario el 2026-10-11:

- **Rama de trabajo larga** (respuesta a B1: «No, considera que crearemos una rama nueva y nos quedaremos en ella porque ire pidiendo otros cambios»). Al empezar `flow-build` se crea una rama desde `development`, propuesta: `feat/guidance-rework`, porque reunirá varios cambios. El trabajo sigue en ella para los cambios que el usuario vaya pidiendo. La construcción no empieza hasta que el usuario lo diga.
- **Commits locales** (2026-10-11, respuesta a C1: «commit si, pero nos quedaremos en esta rama hasta uqe te lo indique»). Hay un commit por tarea verificada en `feat/guidance-rework`, sin push ni PR, y el agente sigue en esa rama. Esto reemplaza la instrucción anterior:
- **Sin commits** («de hecho, no hagas commits»), ya reemplazada. Esta instrucción, más estrecha, reemplaza a E1-A (interactivo), que el usuario había elegido minutos antes. Los cambios se implementan y verifican en el árbol de trabajo de esa rama, sin commits, push, PR ni `hive update`, porque `hive update` despliega solo lo que tiene commit. La entrega queda pendiente de que el usuario revise el diff y decida.
- **E2-A, sin revisión dedicada.** Lectura propia del diff, las comprobaciones `rg` del plan y `go test ./...`.

Excluido: crear una release. Se propone en el cierre si se cumple el criterio de `CONTRIBUTING.md`.
