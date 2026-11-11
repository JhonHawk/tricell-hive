# figma-design-main-thread-review — Figma: el hilo principal dibuja por script y un revisor evalúa

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #453 (f6ee0520), PR #454 (20956270) |
| Tracker | ninguno |
| Git | tricell-hive · base `development` @ `ea4195ab` · rama propuesta `feat/figma-design-main-thread-review` |

## Decisiones

2026-11-10
- **Diagnóstico** — ¿Por qué las vistas de FSR-857 y FSR-858 salieron chuecas y sin homologar? Por el proceso más que por el modelo: un subagente sin el contexto de la sesión dibujó alternativas A/B y corrigió con decenas de parches sueltos de tamaño y posición; las instrucciones del subagente no llevaban las pantallas hermanas ni las decisiones de UI que el usuario ya había tomado. El usuario aceptó los hallazgos y aprobó adoptar la corrección sin esperar una segunda ocurrencia.
- **Hilo principal** — ¿Quién dibuja? El hilo principal, no un subagente, porque conserva las convenciones de la sesión.
- **Script** — ¿Cómo se dibuja? Por script, como lo hizo Fable 5.2 en FSR-856: un kit de piezas reutilizables sobre autolayout, el marco de la app clonado de una pantalla existente y variables del archivo; una corrección cambia una pieza o los datos y vuelve a dibujar, no mueve ni redimensiona nodos sueltos.
- **Un ejemplo** — ¿Cuántos ejemplos? Uno por pantalla, siguiendo las pantallas hermanas; dos o tres solo cuando el usuario los pide explícitamente para una pantalla que nombra, para no dibujar y descartar flujos completos.
- **Evaluador** — ¿Qué hace el subagente? Pasa de ejecutor a evaluador: revisa los frames contra lo pedido antes de que el usuario los vea.
- **D1 A** — ¿Quién revisa los frames? El revisor de UX (`hive-review-ux`), ampliado con un modo para frames de un archivo de diseño; no se crea un rol nuevo.
- **D2 A** — ¿Research recomienda la etapa para toda pantalla nueva aunque haya pantallas hermanas? Sí, cuando el host tiene la conexión que escribe en Figma y no hay delegación desatendida; los cambios que conservan el layout siguen sin recomendación.
- **D3 A** — ¿Una convención corregida que aplica más allá de la pantalla se registra en `DESIGN.md`? En la ronda de aprobación se pregunta una vez; con un sí, el plan agrega la tarea que la escribe en las decisiones de `DESIGN.md`.

## Solicitud

> Valida las sesiones de codex / claude en fleetsystem […] quiero saber si es problema del modelo, de las skills o qué pasa.

> No me importa que lo escriba el hilo principal si eso aumenta la calidad, además el proceso que usó fable fue diferente, por script y creo que eso lo dejó mucho más exacto. Estoy de acuerdo en los hallazgos, pero en la parte de propuestas que solo se haga por petición explícita del usuario y sobre una pantalla específica (para no desechar flujos completos).

> Lo que puede hacer el subagente ahora es evaluar esas pantallas para ver que cumplan con lo que se solicita (pasa de ejecutor a evaluador).

Sin ticket.

## Objetivo

Hoy `flow-figma-design` dibuja dos o tres ejemplos cuando cree que la composición está abierta, delega el dibujo a un subagente `hive-design-ui` cuyas instrucciones no incluyen las pantallas hermanas, y nadie revisa los frames antes de que el usuario los vea. Al terminar, el hilo principal dibuja un solo ejemplo por pantalla, por script y siguiendo las pantallas hermanas y las decisiones ya registradas; un subagente `hive-review-ux` lo evalúa y el hilo principal corrige lo que se sostiene antes de preguntar; las alternativas solo aparecen cuando el usuario las pide para una pantalla, y una convención que el usuario corrige puede quedar en `DESIGN.md`.

## Alcance y aceptación

Dentro: la skill `flow-figma-design`, la regla de delegación, los roles `hive-review-ux` y `hive-design-ui`, la recomendación de `flow-research`, el paso de UI de `flow-plan`, el registro de diseño de `hive-ui-criteria`, los requisitos de las especificaciones `workflow-lifecycle` y `ui-criteria`, y las páginas del sitio que describen la etapa. Fuera: rehacer los frames de FSR-857 y FSR-858 (es trabajo de fleetsystem), roles nuevos y un modo de navegador para frames. La guía global cambia solo en dos frases de rutas (+88 bytes), que la revisión de código mostró contradictorias.

- AC1 — La etapa dibuja un solo ejemplo por pantalla, que sigue las pantallas hermanas del mismo módulo; dibuja dos o tres solo para una pantalla que el usuario nombra al pedir alternativas · false today because el paso 5 de `content/skills/flow-figma-design/SKILL.md` dice "two or three examples when the composition is open" · Observed in: `content/skills/flow-figma-design/SKILL.md` del candidato
- AC2 — El hilo principal dibuja los frames por script (piezas reutilizables sobre autolayout, marco clonado de una pantalla existente, variables del archivo) y corrige cambiando una pieza o los datos y volviendo a dibujar; ninguna regla manda el dibujo a un subagente · false today because el paso 5 y `content/skills/flow-build/references/delegation.md` delegan el dibujo a un `hive-design-ui` · Observed in: esos dos archivos del candidato
- AC3 — El registro se escribe en cuanto se dibuja, con estado en revisión; antes de preguntar al usuario, un subagente `hive-review-ux` revisa los frames contra lo pedido, las pantallas hermanas, las decisiones registradas y la alineación, y el hilo principal corrige lo que se sostiene en una ronda; el rol describe ese modo y su veredicto sin chequeos de navegador, y sin conexión al archivo juzga por las imágenes y reporta como no corridas las medidas de alineación · false today because ni la skill ni `content/agents/review/hive-review-ux.md` revisan frames de un archivo de diseño · Observed in: la skill, el rol y `content/skills/flow-build/references/delegation.md` del candidato
- AC4 — Research termina con el bloque que recomienda la etapa para toda pantalla nueva cuando el host tiene la conexión que escribe en Figma, aunque haya pantallas hermanas y aunque el cambio fuera directo a build, y no la recomienda para un cambio que conserva el layout ni bajo delegación desatendida · false today because el paso 5 de `content/skills/flow-research/SKILL.md` exige que nada fije la composición y solo aplica cuando research iba a terminar hacia el plan · Observed in: `content/skills/flow-research/SKILL.md` del candidato
- AC5 — Cuando el usuario corrige una convención que aplica más allá de la pantalla, la ronda de aprobación pregunta una vez si se registra en las decisiones de `DESIGN.md`; un sí queda en el registro del ejemplo, el plan agrega la tarea que la escribe y, si el entregable era el diseño, el cierre la nombra como pendiente · false today because la etapa solo pregunta por el archivo de diseño · Observed in: la skill, `content/skills/flow-plan/SKILL.md` y `content/skills/hive-ui-criteria/references/design-record.md` del candidato
- AC6 — Los deltas de especificación, las páginas del sitio, `README.md` y `llms.txt` describen el comportamiento nuevo y ninguno dice que la etapa dibuja dos o tres propuestas por defecto · false today because `site/src/content/docs/flows/figma-design.md` y `README.md` dicen "two or three proposals" · Observed in: esos archivos del candidato y el sitio construido con `pnpm --dir site build`
- AC7 — Las pruebas de contenido pasan · **Retired:** no es falso en la base; pasa a la tabla de verificación como chequeo de la suite.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- Sesión de Claude Code `c71dd314` (fleetsystem, FSR-856): Fable 5.2 en el hilo principal, 4 escrituras con `use_figma` entre 05:02 y 05:05 (hora local, 2026-11-10), un kit de piezas (`B` sobre `createAutoLayout`, `btn`, `badge`, `table`, `mkRow`, `select`, `pager`, `shell`) y el marco clonado del frame `2:3737`; `flow-figma-design` todavía no existía (PR #450 integrado a las 11:02).
- Codex `01a1c819` (subagente `design_figma_857`, `gpt-6-astra` high): ~70 operaciones entre 18:54 y 19:48, 112 `resize()`, 21 posiciones `x` a mano, 11 nodos `ABSOLUTE`, alternativas A/B.
- Codex `01a1c8e0` (subagente `draw_858`, `gpt-6.1-sol` high): ~50 operaciones, alternativas A/B, midió los patrones de las pantallas de 856 a las 22:50, después de la queja del usuario a las 22:49.
- Frames actuales del archivo Figma de fleetsystem leídos el 2026-11-10: barra superior de 76 px en 857/858 frente a 64 px en 856, alta de 857 en un diálogo frente a una página en 856, y acciones de texto en 858 frente a iconos con tooltip decididos en 857.
- Las instrucciones de los subagentes de Codex están cifradas en el registro: no se pudo leer qué recibieron.

## Effects

None: the tasks take no effect on an external system.

## Entrega

- **D4 B** — ¿Cómo se entrega? Automático: revisión, push, PR hacia `development` e integración sin parada de validación. Al tener la marca de mantenedor, el refresco de la instalación local sigue a la integración, como fija `AGENTS.md`.
- **D5 A** — ¿Qué revisión corre? `/code-review` nativo de Claude Code sobre el candidato antes del push.

No autoriza: un release, cambios en fleetsystem ni rehacer los frames de FSR-857 y FSR-858.

## Estado de revisión y avance

2026-11-10 — Plan escrito.
2026-11-10 — Revisión del plan (`hive-review-plan`): 1 P1, 7 P2, 4 P3, todas aplicadas. Research recomienda la etapa también cuando iría directo a build; el brief del revisor vive en «Reviewer brief»; el rol revisor recibe un modo de frames con su veredicto y su respaldo sin conexión; el registro se escribe al dibujar con estado en revisión; README, llms.txt, la página del plan y el CHANGELOG entran en T4; las convenciones en una petición de solo diseño quedan pendientes en el cierre; las decisiones previas se leen de `DESIGN.md` y de las secciones `UI` de carpetas de cambio; AC7 retirado a la tabla de verificación. Supuesto: el primer dibujo y el frame de 440 px reciben una revisión cada uno; una corrección que pidió el usuario no, porque el usuario la ve en ese mismo momento.
2026-11-10 — Revisión de código (`/code-review` high): 10 hallazgos confirmados y corregidos en una ronda. Aprobación por pantalla y borrado solo de alternativas; ruta de la etapa a build para cambios pequeños; excepción de pantalla nueva en la guía global, el sitio y el README; informe del revisor de frames; decisiones previas leídas de `DESIGN.md` y del registro de cada hermana, sin recorrer todas las carpetas archivadas; regla de dibujo con un solo hogar; origen de una convención sin carpeta de cambio; escenario del registro aclarado; flecha del diagrama del README.
2026-11-11 — H6 y H7 (rama `fix/figma-design-sibling-records`): la búsqueda del registro de cada hermana vuelve a incluir carpetas archivadas y suma los `examples.md` del workspace, por archivo y nodo; reemplaza el «sin recorrer todas las carpetas archivadas» de la ronda anterior porque esa omisión perdía decisiones. Las instrucciones del revisor llevan lo que fijan las hermanas y sus correcciones, y el rol mide las hermanas cuando faltan.
