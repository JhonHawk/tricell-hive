# La revisión de UI respeta el sistema de diseño de la librería de UI

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el PR [#174](https://github.com/JhonHawk/tricell-hive/pull/174) (merge `00fde904`) |
| Tracker · GitHub Issues | • [#173 — La revisión de UI empuja cambios al tema y a los componentes de la librería de UI](https://github.com/JhonHawk/tricell-hive/issues/173) |
| Git | `automatic` (declarado) · rama `feat/gh-173-ui-review-design-baseline` propuesta desde `development` · sin parada humana |
| Verificación | revisión del texto por criterio · `go test ./...` · `/code-review` |
| Siguiente paso | Ninguno: cerrado y archivado |

## Pedido

> Sobre el verificador de UX/UI:
> - Valida la sesión de pi en carvento, y aunque en contraste tiene algo de razón esto obliga a personalizar el tema por defecto de shadcn, tema que ya tiene numerosos estudios al respecto sobre su funcionamiento y valides visual.
> - Veamos qué temas deben respetarse por: tema, framework de ui, reglas previas del proyecto (y otros puntos que no estoy mencionando)
> - Esto implicara ajustar lo que valida esa etapa y subagente, dame una propuesta al respecto
> - Puedes analizar otros tickets para ver si esto nos ha hecho cambios que eran innecesarios, tiempo consumido o que vaya en contra de lo que menciono arriba

> Tal vez shadcn no respalde la desición con una investigación o articulo, sin embargo dejar el sistema de diseño del framework o libreria de UI nos ayudaria a alivianar el soporte personalizado y sobre todo a dedicar menos tiempo a ello, ya que normalmente se escoge una libreria UI por como ya funciona en ese momento y las personalizaciones son explicitas.

> explicita mio o ya del proyecto, y tienes razon, una libreria no siempre cumple AA, esa es la magia de la humanidad en contraste de las reglas y el caso para generar belleza (relativa), y si, si se solicita explicitamente es otro tema a atenderse, analiza esto ultimo y R1-A

> D1-A, D2-A, pero aqui ojo, si lo pone como una recomendación antes de aplicar un cambios debe validarse que no sea un problema global (para que el mismo hallazgo o recomendación no siga apareciendo por cada vista que se vaya tocando)

Criterios de [#173](https://github.com/JhonHawk/tricell-hive/issues/173), leídos el 2026-10-25:

> - **AC1:** la guía de la revisión de UI pide leer la base de diseño (configuración de la librería, tokens del tema, componentes instalados, guía UI del proyecto con sus desviaciones y reglas no adoptadas) antes de revisar, y fija el orden de precedencia.
> - **AC2:** un valor por defecto de la librería o una decisión registrada del proyecto no se reporta como hallazgo; el check queda como «no aplica» con su motivo.
> - **AC3:** el criterio de contraste no textual nombra cuándo hay falla AA, cuándo depende de interpretación y cuándo no aplica (decorativo, deshabilitado, valor por defecto del navegador, control identificado por otra señal).
> - **AC4:** el arreglo sugerido usa solo tokens, variantes y componentes existentes; cambiar el tema o un componente instalado, o usar colores sueltos, se reporta como desviación que requiere una decisión explícita.
> - **AC5:** la guía dice dónde se declara un requisito de accesibilidad y que, declarado o pedido en el ticket, prevalece sobre la librería.
> - **AC6:** el diseño de UI (`hive-design-ui`) y la planificación de UI aplican la misma base.
> - **AC7:** la plantilla de guía UI del proyecto tiene lugar para las desviaciones, las reglas no adoptadas y el requisito de accesibilidad.
> - **AC8:** las páginas de `site/` afectadas reflejan el cambio.

## Objetivo

La revisión de UI de Hive (`hive-review-ux`, guiada por `ui-review-criteria.md` y `ui-gate.md`) trata los criterios WCAG y los genéricos de Hive como superiores a todo lo demás. Por eso empuja cambios al tema y a los componentes instalados de la librería de UI. En un proyecto de cliente con shadcn/ui esto produjo cambios que el usuario rechazó, desviaciones sin registrar y regresiones que la misma revisión volvió a reportar.

El cambio fija como base el sistema de diseño de la librería elegida: su tema y sus componentes sin modificar son correctos, y personalizarlos es una decisión explícita del usuario o del proyecto. La revisión lee esa base antes de revisar y aplica un orden de precedencia. Mide el contraste con la precisión de WCAG 2.2, sugiere arreglos solo dentro del sistema y reporta una sola vez por proyecto un hallazgo cuya causa es global. El resultado esperado son menos cambios al tema, menos rondas de corrección y ninguna recomendación repetida vista por vista.

## Alcance y aceptación

**Incluye:**
- La guía distribuida de la revisión de UI: `ui-review-criteria.md`, `ui-gate.md`, `hive-review-ux.md`.
- Los puntos de uso de otras reglas que la nueva regla cambia: `verification.md:44` (P3 sin registrar) e `incidental-findings.md` (destino de un hallazgo declinado).
- Los roles y referencias que diseñan o implementan UI: `hive-design-ui.md`, `hive-build-frontend.md`, `ui-planning.md`.
- La plantilla `ui-conventions-template.md`.
- La página `site/src/content/docs/flows/build.md` y el `CHANGELOG.md`.

**Excluye:**
- Los hallazgos P3 como tickets y el alcance de la revisión a las vistas tocadas, ya resueltos por #159.
- La regla global de hallazgos incidentales.
- Una clave nueva en `## Hive`, descartada por D1-A.
- El código del gestor (`tooling/`).
- Los cambios en el proyecto del cliente, que le tocan a su propia sesión.

**Restricciones:**
- `go test ./...` sigue pasando, incluida la prueba de tamaño de los `SKILL.md`. `flow-walkthrough/SKILL.md` mide 15 540 bytes y no se toca.
- Cada regla tiene un solo hogar, `ui-review-criteria.md`; los demás archivos solo apuntan a ella, como pide `AGENTS.md`.

**Criterios:**

- AC1. `ui-review-criteria.md` pide leer, antes de los checks, la base de diseño: la librería de UI y su configuración, los tokens del tema, los componentes instalados, la guía UI del proyecto con sus desviaciones, reglas no adoptadas y hallazgos globales, y el requisito de accesibilidad si existe. También fija el orden de precedencia: requisito de accesibilidad declarado › decisiones y guía UI del proyecto › valores por defecto de la librería › patrón aceptado de la pantalla › criterios de Hive. El patrón aceptado solo decide la composición. La base queda nombrada como segunda excepción a «no decir qué ignorar» en `ui-review-criteria.md:9` y `ui-gate.md:28`. *Falso en la base cuando* la línea 3 dice «they never excuse a failed check» y no existen esa lectura, ese orden ni esa excepción.
- AC2. Un valor por defecto de la librería o uno que fija una decisión registrada del proyecto no se reporta como hallazgo. El check registra «no aplica» con su motivo. Sin requisito de accesibilidad declarado, esto vale aunque el valor no alcance un umbral WCAG (D2-A). La guía define «valor por defecto de la librería» con comprobaciones que el revisor puede ejecutar:
  - el diff contra la base en configuración, tema y componentes instalados;
  - la coincidencia con el valor que publica la librería para la configuración declarada (D3-A), más el historial del archivo;
  - una clase en el sitio de uso, que es valor del proyecto;
  - la composición, que siempre es del proyecto.

  Un origen que no se puede establecer cuenta como valor del proyecto. *Falso en la base cuando* no hay regla que convierta esos valores en «no aplica» ni que defina cómo se reconocen.
- AC3. El check de contraste define cuándo hay falla AA: el indicador de dónde está un input de texto o un select aunque tenga etiqueta, el límite de otro control sin texto ni icono propio, el contorno de checkbox y radio, los indicadores de estado y un indicador de foco con estilo del autor. También define cuándo no aplica: botón o enlace identificado por su propio texto o icono, borde decorativo, componente deshabilitado, estilo del navegador sin modificar. Cuenta el caret visible como señal de foco de un input. Cada hallazgo nombra el criterio y su nivel, el elemento, la señal que lo identifica y la razón medida. *Falso en la base cuando* la línea 55 pide 3:1 para todo contraste no textual «meaningful» sin definirlo.
- AC4. El arreglo sugerido usa tokens, variantes y componentes existentes. Cambiar un token del tema, editar un componente instalado de la librería o usar valores sueltos de la paleta se reporta como desviación del tema. `ui-gate.md`, incluida su línea 34, la aplica solo con una decisión explícita: un requisito de accesibilidad declarado, una desviación ya registrada o la respuesta del usuario, preguntada en cualquier modo de entrega. Una desviación aplicada se registra en la guía UI. Una rechazada se registra como regla no adoptada y su hallazgo pasa a «no aplica: decisión del proyecto». *Falso en la base cuando* `ui-gate.md:26` y `:34` mandan arreglar todo hallazgo Blocker o High introducido sin distinguir un arreglo que personaliza el tema.
- AC5. `ui-review-criteria.md` dice dónde se declara un requisito de accesibilidad: en la guía UI o en el `AGENTS.md` del proyecto, o pedido por el ticket (D1-A). Declarado, prevalece sobre la librería y cuenta como la decisión de las desviaciones que sus fallas exigen. Su arreglo se registra y pasa por la validación visual cuando el modo de entrega la tiene; bajo `automatic`, el informe de cierre lo lista con cómo revertirlo. *Falso en la base cuando* la guía no menciona un requisito de accesibilidad declarado por el proyecto.
- AC6. `hive-design-ui.md`, `hive-build-frontend.md` y `ui-planning.md` apuntan a la misma base: no personalizan el tema ni los componentes instalados sin una decisión explícita. *Falso en la base cuando* `hive-design-ui.md:15` verifica contraste sin base y ninguno de los tres menciona la base de diseño.
- AC7. `ui-conventions-template.md` tiene lugar para:
  - la librería y su base;
  - el requisito de accesibilidad, registrado como `undecided` hasta que alguien lo decida;
  - las desviaciones del tema;
  - las reglas no adoptadas;
  - los hallazgos globales con su fuente, valor medido y destino.

  *Falso en la base cuando* la plantilla solo tiene «Responsive/accessibility expectations» y «Decisions and exceptions» genéricas.
- AC8. `site/src/content/docs/flows/build.md` explica que la revisión de UI mantiene el sistema de diseño de la librería como base, que personalizarlo requiere tu decisión o la del proyecto y que un hallazgo global se reporta una vez. Su línea 39 deja de decir que todo defecto introducido se arregla sin excepción. *Falso en la base cuando* la página no lo menciona y la línea 39 no tiene excepción.
- AC9. Antes de reportar un hallazgo o una recomendación, el revisor ubica su causa. Si está en una fuente compartida (token del tema, componente instalado o compartido, shell o layout), lo reporta una sola vez como hallazgo global en su fuente, con su valor medido y las vistas afectadas como ejemplos. Una clase puesta en una sola vista sobre un token compartido es hallazgo de esa vista. Lo omite solo cuando la guía UI ya lo registra, el cambio no toca su fuente y la medición coincide. `ui-gate.md` le da un solo destino:
  - una desviación decidida;
  - la pregunta de cierre de hallazgos incidentales, si es P2 o más;
  - un registro sin pregunta, si es P3, incluidos los Nit (D4-A).

  Un hallazgo global introducido o empeorado sigue la regla de arreglo y no estos destinos.

  Ese destino se registra en la guía UI, creada como fija `ui-planning.md` cuando no existe. `verification.md:44` y `incidental-findings.md` nombran esa excepción. *Falso en la base cuando* `ui-review-criteria.md:39` solo dice «a shared defect is a systemic finding, not a defense» y ningún archivo dice cómo evitar que se repita por vista.

AC1–AC8 vienen de #173. AC9 viene de la decisión del usuario del 2026-10-25 y no está en el ticket.

## Decisiones vigentes

- **Base de diseño:** el sistema de diseño de la librería de UI es la base. Personalizar es explícito: lo decide el usuario o ya está registrado en el proyecto (2026-10-25).
- **D1-A:** el requisito de accesibilidad se declara como texto en la guía UI o en el `AGENTS.md` del proyecto, o lo pide el ticket. No hay clave nueva en `## Hive`.
- **D2-A:** sin requisito declarado, las brechas AA de la propia librería no son hallazgos. El check queda como «no aplica: valor de la librería».
- **Hallazgos globales:** se validan como globales antes de recomendar un cambio, se reportan una vez y se registran (AC9).
- **D3-A:** un valor es de la librería solo si coincide con el valor que ella publica para la configuración declarada, leído de su documentación o registro sin identificadores privados. Si no se puede comparar, es valor del proyecto.
- **D4-A:** los Nit globales también se registran en la guía UI, una vez, como excepción nombrada en `ui-gate.md` a la regla global de «nits solo en el reporte» (`content/guidance/global.md:33`), sin editar `global.md`.
- **Fuera de este cambio:** P5 (preexistentes menores sin ticket) ya lo resolvió #159.

## Entrega

Ver [tasks.md](tasks.md#entrega).
