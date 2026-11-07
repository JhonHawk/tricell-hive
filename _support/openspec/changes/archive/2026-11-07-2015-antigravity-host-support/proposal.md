# antigravity-host-support — Antigravity (`agy`) como séptimo host de Hive

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #415, 3933505f |
| Tracker | Sin ticket: el trabajo se hace en esta sesión; ningún issue de GitHub menciona Antigravity (búsqueda 2026-11-06) |
| Git | `tricell-hive`, base `development` en `93ab65ee`; rama propuesta `feat/antigravity-host-support` |

## Decisiones

2026-11-06
- **Modelo de los roles** — ¿Con qué modelo corren los roles de Hive en Antigravity? Con el modelo heredado de la sesión (`model: inherit`); `hive models set` puede cambiarlo por rol más adelante.
- **D1 A** — ¿Se prueba en vivo si `agy` corta también el archivo global? Sí: una llamada a `agy` con marcadores mostró que el global se corta a 24.000 bytes, igual que una regla de proyecto.
- **D2 A** — ¿Cómo llega la guía global a Antigravity? Partida en varias reglas siempre activas (`~/.gemini/config/rules/hive-NN.md`, frontmatter `trigger: always_on`), cortadas por encabezado `##` para que ninguna pase de 24.000 bytes; no por hook ni por plugin.
- **D3 B** — ¿Cómo se limitan los roles de solo lectura? Solo con la instrucción del rol: el rol de Antigravity no lleva lista `tools`, porque `agy` no tiene campo `readonly` y un nombre de herramienta mal escrito cuelga al subagente.
- **D4 A** — ¿Entran los hooks de sesión en esta versión? Sí: Hive registra su hook en `~/.gemini/config/hooks.json` para que las sesiones de `agy` publiquen en el visor como los demás hosts.
- **D5 A** — ¿Se comprueba el candidato en el `agy` real antes de validar? Sí: se instala solo para Antigravity en `~/.gemini/config` y se hacen hasta 3 llamadas de cuota; la instalación queda puesta.
- **D6 A** — ¿Qué modo de entrega? `automatic`: revisión, push, PR y merge sin detenerse.
- **D7 A** — ¿Qué revisión de código? La nativa de Claude Code, `/code-review`, sobre el candidato antes del push.

## Pedido

> "Analiza lo siguiente: Implementar compatibilidad de hive con Antigravity" (2026-11-06), seguido de `/flow-plan Para el modelo que quede heredado`.

Sin ticket ni criterios previos. Fuera de alcance por pedido: nada explícito; el análisis dejó fuera leer el historial nativo de `agy` (protobuf/SQLite).

## Objetivo

Hoy `hive install --hosts antigravity` falla porque el host no existe; una sesión de `agy` (CLI 1.5.2, App 2.29) trabaja sin la guía, las skills ni los roles de Hive, y no puede publicar en el visor. Con el cambio, el gestor instala, actualiza, quita y diagnostica Antigravity en scope usuario: la guía completa llega partida en reglas siempre activas que caben bajo el tope de 24.000 bytes por archivo, las skills compartidas y los roles quedan visibles para `agy`, y el hook de sesión conecta sus sesiones con el visor.

## Alcance y aceptación

Incluye:

- Adaptador `antigravity` del gestor, solo scope usuario, con raíz `~/.gemini/config` (compartida por la CLI, la App y, para las skills, el IDE).
- Guía global partida en reglas `hive-NN.md` con `trigger: always_on`, cada una de 24.000 bytes o menos, con el bloque de voz solo en la primera parte.
- Skills como enlaces simbólicos relativos en `~/.gemini/config/skills/<skill>` hacia el almacén compartido `~/.agents/skills/`, como hace Claude Code.
- Una sonda previa con una regla y un agente de prueba que confirma las rutas `~/.gemini/config/rules/` y `~/.gemini/config/agents/` antes de escribir el código.
- Roles en `~/.gemini/config/agents/<rol>.md` con `name`, `description` y `model: inherit`, sin `tools` ni esfuerzo.
- Hook con nombre `hive` en `~/.gemini/config/hooks.json`, traducción de sus eventos y `antigravity` como host del visor y de la identidad de sesión.
- Contenido distribuido y documentación que listan hosts: la línea de sintaxis de comandos de `global.md`, referencias de `harness-audit` y `flow-walkthrough`, arquitectura, `README.md`, `llms.txt` y el sitio.

Excluye, con su razón:

- Scope proyecto (`.agents/` del repositorio): Cursor, Grok, Pi y OpenCode tampoco lo tienen; se decide aparte si hace falta.
- Leer el historial nativo de `agy` para `doctor` o la revisión de sesiones: está en protobuf y SQLite; queda para un ticket aparte.
- Las extensiones del IDE de Antigravity: no soportan subagentes y su carga de reglas no está verificada.
- `~/.gemini/GEMINI.md`: lo comparten Gemini CLI y el bloque de Engram; Hive no lo toca.

Criterios:

- AC1 — `hive plan install --hosts antigravity --scope user --home <tmp>` propone las partes `<tmp>/.gemini/config/rules/hive-NN.md`, cada una empezando por el frontmatter `trigger: always_on` y de 24.000 bytes o menos, cuyos textos gestionados unidos en orden reproducen la guía reescrita; un enlace relativo `<tmp>/.gemini/config/skills/<skill>` → `../../../.agents/skills/<skill>` (que resuelve a `<tmp>/.agents/skills/<skill>`) por skill; y un `<tmp>/.gemini/config/agents/<rol>.md` por rol con `model: "inherit"` y sin `tools` ni esfuerzo. Con `--scope project` falla con el error de scope no soportado. · falso hoy porque `--hosts antigravity` se rechaza como host desconocido.
- AC2 — En un home sintético, install, update, remove, status y recover de Antigravity dejan byte a byte iguales las reglas ajenas de `rules/`, los enlaces ajenos de `skills/` y los hooks ajenos de `hooks.json` (como `herdr`); remove no deja archivos de Hive; y cuando la guía cambia de tamaño y cruza el límite de una parte, update agrega o quita las partes sobrantes. · falso hoy porque no existe adaptador.
- AC3 — La instalación agrega a `~/.gemini/config/hooks.json` un hook con nombre `hive` con `SessionStart`, `PreInvocation` y `Stop` hacia `hive hook`; con una carga de `agy` (claves camelCase, `conversationId`, `workspacePaths`), `SessionStart` produce el inicio de sesión canónico con host `antigravity` y el directorio de `workspacePaths[0]`, `Stop` produce un fin de turno y nunca un fin de sesión; cada ejecución, también con error, argumentos desconocidos o tiempo agotado, imprime la salida mínima aceptada para su evento; y el visor acepta `antigravity` en su línea de tiempo. · falso hoy porque `hive hook` y el visor no conocen el host.
- AC4 — `content/guidance/global.md` nombra Antigravity en la línea de sintaxis de comandos con `/flow-x`, y `README.md`, `llms.txt` y la página de límites de hosts del sitio listan Antigravity con su scope y sus límites (guía partida por el tope de 24 KB, roles de solo lectura por instrucción). · falso hoy porque ninguno lo menciona.
- AC5 — Con el candidato instalado para Antigravity en el `~/.gemini/config` real, una llamada `agy -p` cita literalmente el primer y el último encabezado `##` de la guía de Hive sin leer archivos, `agy agents` lista cada rol de Hive una sola vez, y una sesión de `agy` deja su inicio registrado en el visor. · falso hoy porque Hive no está instalado para `agy`.

## Prerrequisitos

- [x] Autorización para instalar el candidato en el `~/.gemini/config` real y gastar hasta 3 llamadas de cuota de Antigravity · responsable: el usuario · lo necesitan: T0 (una llamada) y T6 (hasta dos) (dada el 2026-11-06, D5 A)
- [x] `agy` instalado y con sesión iniciada · responsable: el usuario · lo necesita: T6 (verificado 2026-11-06: `agy` 1.5.2 en `~/.local/bin/agy`, una llamada `-p` respondió)

## Contexto verificado

- `agy` 1.5.2, App 2.29.0 y IDE 2.5.5 instalados; la configuración compartida vive en `~/.gemini/config/` (doc oficial `customization-discovery` y la skill integrada `agy-customizations` de `agy`).
- Tope por archivo de regla de 24.000 bytes, aplicado también al global: llamada `agy -p` con `gemini-3.8-flash-low` el 2026-11-06 sobre una copia de `global.md` con marcadores en los bytes 8, 20.342, 26.150 y 37.852; vio los dos primeros y no los dos últimos (conversación `a720fff1-7e01-4b24-83cd-2d490787e333`).
- Presupuesto de 20.000 tokens compartido por las reglas globales y `always_on` (doc oficial `rules`); la guía estimada en unos 9.400 tokens más el bloque de Engram de `~/.gemini/GEMINI.md` (unos 1.800).
- `hooks.json`: objeto de hooks con nombre en la raíz, cada uno con sus eventos (`SessionStart`, `PreToolUse`, `PostToolUse`, `PreInvocation`, `PostInvocation`, `Stop`); el comando lee JSON camelCase por stdin y debe escribir JSON por stdout. Hoy contiene solo el hook `herdr`.
- `~/.gemini/config/skills/` ya contiene enlaces ajenos (`firecrawl`, `shadcn`) hacia `~/.agents/skills/`.
- Precedente externo: una referencia en producción entrega su guía a `agy` partida en reglas `always_on` por encabezado, desde el 2026-11-03.
- Código de Hive: el host está fijado en al menos 12 listas independientes; el gestor no sabe escribir una fuente en varios archivos (`target/catalog.go:16-18`, `plan.go:406`, `ownership.go:271`); el adaptador de Cursor (`integrations/cursor/cursor.go`, commit `58ac1d9b`) es el modelo.

## Entrega

- Modo `automatic` (D6 A): tras T6 y la revisión, push de `feat/antigravity-host-support`, PR hacia `development`, espera de CI y merge, sin parada de validación. El punto final es el merge en `development`; con la marca de mantenedor de este checkout, el refresco local (binario y contenido) sigue al merge.
- Revisión de código: `/code-review` de Claude Code sobre el candidato antes del push (D7 A).
- Comprobación en vivo antes del push (D5 A), dentro de las 3 llamadas autorizadas.
- No autoriza: un release, ni tocar `~/.gemini/GEMINI.md`, ni pilotos de conducta más allá de T6.

## Estado de la revisión y progreso

2026-11-06 — Plan escrito.
2026-11-07 — Revisión del plan (`hive-review-plan`): 10 hallazgos, todos aplicados. Cambiaron los eventos del hook (sin `PreToolUse`, `Stop` como fin de turno, directorio de `workspacePaths`, salida JSON en todo camino), el corte de la guía (encabezados fuera de bloques de código, sin corte por `###`), la voz solo en la primera parte con sus pruebas, una sonda de formato previa (T0), los deltas (registro de hooks, modelos por rol, adaptador de scope usuario), los puntos de código y documentación faltantes, el procedimiento de pilotos y la reversión en T6, el destino relativo de los enlaces y la forma de `model`.
