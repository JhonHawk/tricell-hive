# figma-design-library — Figma: dibujar con la biblioteca publicada de la app

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #461, merge 5be1ee32 |
| Tracker | ninguno |
| Git | tricell-hive · base `development` @ `6b81dbf0` · rama `feat/figma-design-library` |

## Decisiones

2026-11-11
- **Origen** — Hueco: ninguna regla dice qué hacer cuando la app tiene una biblioteca de Figma publicada. El paso 5 de `flow-figma-design` pide definir piezas propias en cada script, así que una biblioteca existente se ignoraría y las páginas de cada ticket seguirían divergiendo.
- **A** — El usuario eligió ajustar Hive antes de crear la biblioteca de fleetsystem: un campo `**Design library:**` en `DESIGN.md`, escrito solo por instrucción del usuario, y dibujar con instancias de esa biblioteca; las piezas que le falten se definen localmente y se anotan como candidatas.
- **Supuesto: sin pregunta nueva** — La etapa no pregunta por la biblioteca; el usuario la declara cuando exista. Sin campo, todo sigue igual.

## Solicitud

> Lo que se me ocurre es validar si existen un sistema de diseño, por ejemplo en cam si estan las librerias de componente, y las vistas ya son super parecidas, talvez en fleetsystem lo que valga la pena hacer es crear el system desing … y se usa como libreria

> ahi, pregunta importante eso que te digo de la biblioteca ajusta algo de la skill?

**Evidencia:** sesiones Codex `01a1c93d-7172-74f0-a2cc-581c66a4aeb8` y `01a1c9ca-91f9-7833-9390-9c5e7ac60efc` (fleetsystem, gpt-6.1-sol, 2026-11-11): el archivo de Figma solo tiene bibliotecas genéricas de comunidad; cada sesión dibuja con piezas propias; se encontraron barra superior de 64 px frente a 76 px, márgenes de 0 px y tres variantes de columna de acciones.

## Alcance y aceptación

**Dentro:** `flow-figma-design` pasos 4 y 5, `hive-review-ux` (modo frames), `hive-ui-criteria` registro de diseño, specs `ui-criteria` y `workflow-lifecycle`, página del sitio, `instruction-resources.md`, CHANGELOG.

**Fuera:** crear la biblioteca de fleetsystem (trabajo de ese proyecto).

**Tras la revisión de código:** el revisor de UX señala una pieza que la biblioteca tiene y no se dibujó como instancia; las variables siguen la misma regla; añadir la biblioteca al archivo es un paso manual del usuario; sin acceso a ella se dibuja con piezas propias y se avisa; las candidatas tienen línea `candidate:` y se nombran en el mensaje de aprobación.

- **AC1** — Con `**Design library:**` declarado, el paso 5 dibuja con instancias de sus componentes y variables y solo define piezas para lo que falta, anotadas como candidatas. Observed in: `content/skills/flow-figma-design/SKILL.md`.
- **AC2** — El registro de diseño define el campo, escrito solo por instrucción del usuario, sin huella ni efecto en la caducidad. Observed in: `content/skills/hive-ui-criteria/references/design-record.md`.
