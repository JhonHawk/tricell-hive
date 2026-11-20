# Diseño — create-design

## Enfoques para los criterios de UI (D3)

- **A, elegido — mover tal cual.** `hive-ui-criteria/SKILL.md` pasa a `create-design/references/ui-criteria.md` con su precedencia, sus pasos y su formato de salida; `criteria.md`, `design-record.md`, `refinement.md` y `notice.md` pasan a `create-design/references/` sin cambios de contenido. Los consumidores cambian solo la ruta. Es el primer escalón de reutilizar (mover sin reescribir) y los roles leen lo mismo, con los mismos saltos. Costo: la revisión de UX y el build leen un archivo que vive dentro de la skill de diseño.
- **B — hoja de regla aparte** (`hive-rule-ui-criteria`). Revertía D5 en parte y dejaba una carpeta más.
- **C — disolver** la precedencia en `criteria.md` y la detección en `design-record.md`. Reescribía reglas y enlaces internos sin ahorrar ninguna lectura: en la medición del 2026-11-16, los roles abrieron primero el `SKILL.md` (57 lecturas contra 39 de `criteria.md`).

## Forma de la skill

`create-design/SKILL.md` abre con sus entradas y a cuál ir:

| Entrada | Cuándo | Qué hace |
| --- | --- | --- |
| Dibujar | Una pantalla nueva o recompuesta antes del plan, o un pedido cuyo entregable es el diseño | Los pasos 1–8 actuales |
| Corregir | Un pedido que nombra frames existentes como destino | Los mismos pasos, con `references/special-cases.md` «A request for the design itself» |
| Revisar un diseño existente | Un link de Figma o una imagen de un diseño que dibujó otra persona | `references/review.md`: un hijo `hive-review-ux` en un hilo aparte; no dibuja |
| Crear o actualizar `DESIGN.md` | El usuario lo pide, o un plan de UI lo necesita | `references/ui-criteria.md` y `references/design-record.md` «Writing» |

La precedencia sigue con un solo hogar, ahora `references/ui-criteria.md#precedence`; el `SKILL.md` de dibujo enlaza ahí donde hoy enlaza `../hive-ui-criteria/SKILL.md#precedence`.

## Ancho de los frames (D1)

El paso 5 distingue la plataforma de la app: una app web se dibuja al ancho de escritorio de sus frames existentes, si no 1440 px, más el frame de 440 px cuando se sirve en móvil; una app que corre solo en móvil se dibuja una vez, a 440 px, sin segundo frame. La señal es la que el paso 3 ya lee: una app nativa, o un shell de teléfono sin layout de escritorio en el código y en los frames hermanos, es solo móvil; un layout adaptable (un breakpoint o una etiqueta meta viewport, la misma señal de `hive-rule-ui-verification`) es una app web servida en móvil. Sin código ni frames, decide la plataforma que nombran el pedido o el ticket; sin ninguna, se dibuja como web con su frame de 440 px y se dice en el mensaje de aprobación. Con un solo ancho, `sections.md` no coloca un segundo frame debajo de cada pantalla. Los pares de kin vinieron de la sesión del 2026-11-16, previa a la regla de secciones, que agrupó cada vista con sus dos anchos a pedido del usuario: no rompe ninguna regla vigente.

Una consecuencia de D4: hoy un sí a registrar una convención ya envía la obra al plan (`approval.md` «Close the stage»); con D4, un sí al archivo o la biblioteca también, porque siempre queda `pending`.

## Entrega de `DESIGN.md` (D4)

`design-library.md` «Record it» deja de escribir y entregar: un sí siempre queda como `record: <campos> · pending` en `examples.md`, la vía que hoy ya sigue cuando no hay sección `## Hive`. `hive-rule-ui-example` y `hive-rule-ui-plan` ya agregan la tarea que escribe un `pending`; se revisa que su texto cubra también el caso con sección `## Hive`. Desaparecen la pregunta de entrega a mitad del diseño y el commit aparte.

## Entrada de revisión (D7)

`references/review.md` fija: qué acepta (link de Figma legible por la conexión, o imagen, que no necesita conexión), el brief del hijo `hive-review-ux` según `hive-rule-reviewer-brief`, que gana el caso de un diseño de otra persona en su único hogar (sin el razonamiento de quien dibujó), contra qué revisa (el pedido si lo hay, las pantallas existentes, las decisiones registradas y `ui-criteria.md`), qué devuelve (hallazgos por prioridad, estados faltantes, reglas sin decidir como preguntas con respuesta recomendada, veredicto) y cómo termina (bloque de respuestas sin comando, según `hive-rule-report-endings`). No escribe en Figma ni en el repositorio. El rol `hive-review-ux` ya cubre «frames dibujados en un archivo de diseño»; se agrega el caso de una imagen sin nodos.
