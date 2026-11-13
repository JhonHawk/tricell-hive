# hive-local-one-line-facts — `HIVE.local.md` como ficha de una línea por hecho, con el detalle en la memoria persistente

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #470, merge 9737eec4 |
| Tracker | ninguno |
| Git | tricell-hive, base `development` en `a8b2632e`, rama `feat/hive-local-one-line-facts` |

## Decisiones

2026-11-13
- **Requisito del usuario** — ¿Qué debe ser `HIVE.local.md`? Una ficha corta de hechos de una línea sobre lo que esta máquina o proyecto tiene o usa («esta máquina usa estos contenedores de Docker», «tiene Figma instalado», «usa acli para conectarse a Jira con la cuenta X», «tiene 3 cuentas de Git, pero la de este proyecto es X»), sin prosa extensa; lo demás puede irse a Engram.
- **Requisito del usuario** — Contemplar en el plan la limpieza de los archivos que ya existen.
- **D11 A** — ¿Los subagentes siguen leyendo los `HIVE.local.md`? Sí; nada cambia para ellos y el costo baja porque los archivos quedan cortos.
- **D12 A** — ¿Dónde vive en Engram el detalle de una entrada del archivo de máquina? En un proyecto fijo, `hive-machine`, que cualquier sesión consulta por nombre.
- **D13 A** — ¿Una limitación de una herramienta con su solución sigue siendo anotable? Sí: la línea dice la limitación y los pasos van a Engram.
- **Aprobación del usuario** — Ampliar la lista de lo que no se registra corrige una regla que no se siguió; el usuario la adopta de inmediato al pedir que el archivo sea una ficha de una línea y que se limpie.
- **Para aprobar con el plan** — Tope de 200 caracteres por entrada, puntero incluido, y de 50 líneas por archivo (antes 100).
- **Para aprobar con el plan** — Sin Engram se escribe la línea sola, sin puntero: la memoria propia de un host no la leen los demás hosts de la máquina.
- **Para aprobar con el plan** — En la limpieza, una decisión por archivo cubre reescribir ese archivo y guardar su detalle en el Engram de su proyecto, aunque sea otro proyecto: es una excepción de esta vez, no un cambio de la skill.
- **Decisión menor** — Sin validador nuevo: un script mediría el largo, pero no distingue un hecho de una bitácora; la skill de curación hace la limpieza con el usuario.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Solicitud

> Lo siguiente a revisar es el tema del HIVE.local.md, hay mmuchas cosas que considero pueden irse a engram, en lugar de ser una entrada de ese archivo, validalo con las sesiones d eestos dias

> es para que ese archivito sea supoer chido
> - Esta maquina usar estos containers de docker
> - Esta maquina tiene figma instalado
> - Esta maquina usa acli para conectarse a jira con la cuenta X
> - Esta maquina tiene 3 cuentas de git, pero la de este proyecto es X
>
> Ese tpo de cosas de una linea, no con una prosa super extensa

> y contemplemos en el plan limpiar los archivos que ya existen

Sin ticket.

## Objetivo

Hoy una entrada puede ser una receta de varios comandos, y los archivos acumulan mediciones, eventos y estado de trabajos: 77 entradas en 8 archivos, 27 KB, que cada sesión y cada subagente cargan al arrancar. La mayoría no se vuelve a usar.

Con el cambio, una entrada es una línea de hasta 200 caracteres que dice qué tiene, usa o aplica aquí; el detalle que haga falta para reutilizarla vive en la memoria persistente y la línea lo señala; y los archivos existentes quedan limpios con la aprobación del usuario.

## Alcance y aceptación

Dentro: la sección «Local instructions» y la lista de momentos de escritura de «Persistent memory» de la guía global, la skill `local-instructions`, la especificación `local-instructions`, las páginas del sitio y la arquitectura que describan qué guarda el archivo, el changelog, y la limpieza de los ocho `HIVE.local.md` de esta máquina.

Fuera: quién lee los archivos (los subagentes siguen leyéndolos), dónde viven y cómo se excluyen de Git, los scripts locales, `ACCOUNTS.md`, y el script `extract_candidates.py`.

- AC1 — Una entrada es una línea de hasta 200 caracteres que dice qué tiene, qué le falta, qué usa o qué necesita esta máquina, cuenta o entorno (herramienta, app, contenedor, puerto, alias de host, script, perfil de nube con su región, o la cuenta o identidad que usa una herramienta, y para qué sirve), o una limitación de una herramienta aquí con la solución que funcionó, con la fecha de su última verificación; no lleva pasos, relato ni versión, salvo la que el proyecto fija y la máquina no tiene · falso hoy porque `content/guidance/global.md:54` admite «the command order that worked here», versiones y parches sin tope de largo · Observed in: `content/guidance/global.md` «Local instructions»
- AC2 — Cuando la línea no puede contener el comando con sus opciones en 200 caracteres, o reutilizar el hecho toma dos o más pasos, el detalle se guarda en la memoria persistente bajo un tema `local/<slug>` nombrado por el asunto del hecho, tras buscar uno existente, y la línea termina con su puntero; el detalle de un archivo de proyecto va a ese proyecto y el del archivo de máquina al proyecto `hive-machine`; se lee solo cuando la tarea lo necesita y el hilo principal lo pasa en el brief del hijo que lo necesite; el tema empieza con el nombre de la carpeta del archivo cuando ese archivo está debajo de la carpeta de la configuración de Engram, de modo que dos archivos de un mismo proyecto nunca comparten tema; el proyecto `hive-machine` se crea por la línea de comandos de Engram cuando la herramienta de memoria lo rechaza; un detalle que falla deja su línea obsoleta y se corrige, y antes de quitar una línea obsoleta se anota en su tema que falló; la marca de obsoleta y la fecha y causa de una entrada sin verificar no cuentan para los 200 caracteres; si tampoco la línea de comandos crea `hive-machine`, el detalle va al proyecto actual y el puntero lo nombra; sin Engram se escribe la línea sola; «Persistent memory» permite esa escritura, también en su frase de lo que nunca se guarda, y la regla del nombre de proyecto admite el que nombra un puntero · falso hoy porque «Local instructions» no nombra la memoria persistente `global.md:131` solo permite escribir en una decisión, un plan listo, un cierre, un worktree o un resumen de sesión y prohíbe guardar «per task, finding, fix or discovery», y `global.md:133` solo admite el proyecto del `.engram/config.json` más cercano o el actual · Observed in: `content/guidance/global.md` «Local instructions» y «Persistent memory»
- AC3 — Los archivos no registran un evento, el estado de un trabajo (worktrees, procesos, lo que un cierre conservó), el archivo o la librería de diseño de la app una vez que su `DESIGN.md` los tiene, ni un hecho que los archivos versionados del repositorio ya dicen, además de lo ya excluido; una instrucción del usuario se guarda como decisión en la memoria persistente y el archivo conserva solo una línea con lo que restringe · falso hoy porque `global.md:54` solo excluye «the result of a diagnosis, a count or measurement, or whether a host answers right now» · Observed in: `content/guidance/global.md` «Local instructions»
- AC4 — El tope por archivo es de 50 líneas · falso hoy porque `global.md:55` y la skill dicen 100 · Observed in: `content/guidance/global.md` y `content/skills/local-instructions/SKILL.md`
- AC5 — La skill `local-instructions` escribe entradas de una línea y, al curar, propone para cada entrada existente la línea que queda, la línea más el detalle que pasa a la memoria persistente, su traslado a otro de los archivos, o su retiro, y no mueve ni retira nada sin la respuesta del usuario · falso hoy porque su paso 6 escribe «the fact, its command when it has one, and its date» sin tope y su paso 4 dice que las respuestas viven en el archivo «not in persistent memory» · Observed in: `content/skills/local-instructions/SKILL.md`
- AC6 — En los ocho `HIVE.local.md` de esta máquina, cada línea de entrada (las que empiezan con `- `) tiene 200 caracteres o menos, ninguna está marcada obsoleta ni es de las clases que AC3 excluye, y cada archivo tiene 50 líneas o menos contando títulos; el detalle de cada entrada partida está en Engram, en el proyecto del archivo o en `hive-machine`, leído de vuelta antes de reescribir; ningún detalle guardado contiene un valor secreto; y una copia literal de los ocho archivos anteriores queda en la carpeta ignorada de la limpieza · falso hoy porque suman 77 entradas con una media cercana a 300 bytes, 22 son mediciones, eventos o estado, y 3 están marcadas obsoletas · Observed in: el archivo de máquina `~/.config/hive/HIVE.local.md`, los siete archivos de proyecto que la investigación listó, las entradas `local/…` de Engram de cada proyecto y de `hive-machine`, y `_support/workspace/2026-11-13-hive-local-cleanup/`
- AC7 — La especificación (su propósito y sus requisitos), el sitio, `llms.txt`, la arquitectura y el changelog describen la entrada de una línea y su detalle en la memoria persistente · falso hoy porque el requisito «Verified, dated and secret-free entries» repite el texto de `global.md:54` y su escenario «The local stack starts» guarda los comandos en el archivo · Observed in: el delta de esta carpeta, `CHANGELOG.md` y cada página del sitio o de arquitectura que T3 encuentre afirmando lo anterior

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

- Regla leída en `a8b2632e` (2026-11-13): `content/guidance/global.md` líneas 51 a 58 («Local instructions») y 131 («Persistent memory»), 48,751 bytes; `content/skills/local-instructions/SKILL.md`, 8,060 bytes; especificación `local-instructions` con seis requisitos.
- Lugares que nombran `HIVE.local.md` o «Local instructions»: en `content/`, las skills `flow-build`, `flow-verify`, `flow-research`, `flow-status`, `harness-audit` (tres referencias), `security-advisory`, `workspace-conventions` (`artifact-layout.md`) y el rol `hive-review-harness`; en el sitio, `flows/build.md`, `flows/verify.md`, `getting-started.md`, `host-limits.md`, `index.mdx` y `skills.md`; `_support/docs/architecture/workspace-and-artifacts.md`; la especificación `test-accounts`; `tests/pilot/flows.go` y `record_placement_test.go`, que solo comprueban la ruta del archivo; `llms.txt` (línea 46) y `README.md`; `CHANGELOG.md`. `tooling/management/antigravity_rules.go` limita los bytes de cada sección `##` de la guía para Antigravity.
- Medición sobre las sesiones reales del 7 al 13 de noviembre de 2026 (dos hijos de investigación, búsqueda por términos literales, que subcuenta):
  - Claude Code, 53 sesiones: 49 de 77 entradas sin uso en una sesión posterior; 6 usadas en dos o más; 58 lecturas completas del archivo de máquina, la mitad de subagentes.
  - Codex, 276 rollouts: 63 de 77 sin uso; 11 usadas en dos o más; 201 sesiones leyeron el archivo de máquina, 185 de ellas subagentes.
  - 22 entradas son mediciones, eventos, diagnósticos o estado, que la regla vigente ya excluye en parte; 3 marcadas obsoletas; 4 grupos de duplicados; 7 entradas del archivo de máquina son de un solo proyecto.
  - Lo que sí se reutilizó son hechos operativos cortos: qué perfil de nube sirve a un proyecto, adjuntos de un ticket por REST, copiar la configuración local a un worktree nuevo, el navegador instalado para capturas.
  - El inventario de versiones y rutas no se usó y las sesiones lo vuelven a comprobar por comando.
- Engram 3.7.1: `engram save` acepta `--project`, `--type` y `--topic`; el puntero `engram:<project>/<topic_key>` es el que «Persistent memory» ya define.
- Naturaleza de la regla, para «Measurement» de `AGENTS.md`: el tope de línea, el detalle en la memoria persistente y el tope de 50 son requisitos nuevos que la decisión del usuario adopta; ampliar la lista de lo que no se registra corrige una regla que no se siguió en muchas sesiones de Claude Code y Codex, con la aprobación del usuario.

## Effects

None: the tasks take no effect on an external system.

## Entrega

Entrega automática y revisión nativa de Claude Code `/code-review`, las respuestas vigentes de la sesión para este repositorio: revisión, push, PR hacia `development` y merge sin detenerse, y después el refresco de la instalación local. La limpieza de los archivos (T4) no forma parte del PR: son archivos sin versionar de esta máquina, y se detiene una vez para que el usuario apruebe la tabla de entradas. No autoriza un release ni la promoción a `master`.

## Estado de revisión y progreso

2026-11-13 — Plan escrito. Revisión del plan (`hive-review-plan`): sin P0; aplicados tres P1 (el detalle del archivo de máquina no tenía proyecto de Engram: D12 A; el delta se contradecía sobre versiones, cuentas, comandos y el tope; «Persistent memory» seguía prohibiendo la escritura nueva), doce P2 (limitaciones con su solución: D13 A; clase de traslado entre archivos; la entrada de diseño se conserva hasta que `DESIGN.md` la tenga; la aprobación por archivo nombra las escrituras en otros proyectos; respaldo literal antes de reescribir; ciclo de vida del puntero; el detalle llega a los hijos por el brief; búsqueda de T3 ampliada; observaciones en el commit) y cinco P3.

2026-11-13 — El usuario aprobó la tabla de limpieza de los ocho archivos (D14 a D21, todos A). Revisión de código (`/code-review` sobre `8c6dfeaa`): diez hallazgos, corregidos en una ronda: el proyecto `hive-machine` no existía y la herramienta de memoria rechaza nombres desconocidos (se crea por la línea de comandos); dos archivos que comparten proyecto podían pisarse un tema; la pregunta de curación no tenía opción para partir o reubicar; la skill contradecía guardar una instrucción como decisión; la marca de obsoleta no cabía en el tope; quitar una línea ya no borra memoria; una herramienta que falta sigue siendo anotable; la skill ya no pide versiones que no se guardan; y el sitio describía un tope rígido y Engram como único destino.

2026-11-13 — Segunda revisión de código (`8c6dfeaa..cea40ee3`) y revisión del plan revisado: sin P0 ni P1. Corregido en un último commit, que ya no pasa por otra revisión: el tema de un detalle lleva el nombre de la carpeta de su archivo cuando varios archivos comparten proyecto, en vez de una regla de pertenencia que no se podía comprobar; antes de quitar una línea obsoleta se anota el fallo en su tema; una instrucción del usuario se guarda como decisión y deja una línea con lo que restringe; la pregunta de curación por archivo usa la opción D como los grupos y es la única que decide las entradas existentes; hay salida cuando tampoco la línea de comandos crea `hive-machine`; la fecha y causa de una entrada sin verificar no cuentan para el tope; y el delta de curación, la página de build, el changelog y el checkpoint dicen lo mismo.
