# local-instructions-scope — Solo datos útiles en HIVE.local.md y un archivo de máquina

| Campo | Valor |
| --- | --- |
| Estado | Draft · plan en revisión |
| Tracker | ninguno |
| Git | JhonHawk/tricell-hive · base `development` en `831e6032` · rama propuesta `fix/local-instructions-scope` |

## Decisiones

2026-11-07
- **Revisión de la sesión `de72f693`, opción A** — ¿Se ajusta la regla ya, sin esperar una segunda sesión? Sí: el usuario aprobó la corrección de inmediato, como permite «Measurement» del `AGENTS.md` del repositorio (registro `watch/local-instructions-scope`).
- **D16 A** — ¿Dónde vive el archivo de máquina? En `~/.config/hive/HIVE.local.md`: la convención de las CLIs del usuario, y una carpeta que Hive no gestiona. Supuesto menor tras la revisión del plan: con `XDG_CONFIG_HOME` definida, en `$XDG_CONFIG_HOME/hive/`.
- **D17 A** — ¿Dónde va qué perfil de nube usa cada entorno? Partido: los perfiles existentes y su región al archivo de máquina; qué perfil sirve a cada entorno, al archivo del repositorio.

Rigen las respuestas de entrega de esta sesión para este repositorio: `interactive` (D5 A) con `/code-review` de Claude Code (D6 A).

## Request

> «revisa la sesión en paralelo de Hive con claude code... esta usando HIVE.local.md como otra memoria... no sé si sea lo correcto»
>
> Respuesta A al bloque de la revisión: «Ajustar la regla ya, con tu aprobación: dos criterios observables […] y un archivo de máquina leído en cada sesión para los datos de toda la máquina.»

Sin ticket. Fuera de alcance: cambios en `extract_candidates.py` y en los detectores de pilotos (pausados; un piloto que use el HOME real podría escribir el archivo de máquina real, sin verificar) (sus listas ya separan lo de la máquina, `missing` y `cloud`, de lo del proyecto, `commands` y `ports` con `where`), y cualquier archivo de otros proyectos.

## Objective

La sesión de Claude Code `de72f693` (2026-11-07, tarea de mantenimiento del Mac lanzada desde el checkout de tricell-hive) escribió en `tricell-hive/HIVE.local.md`:
- tres datos de la máquina que no son del proyecto (T3 Code, cloudflared, las dos apps de Antigravity);
- un diagnóstico de macOS con un conteo (SetStoreUpdateService);
- un estado pasajero (`ssh bb1` sin ruta).

La regla lo permitía: no pide relevancia ni durabilidad, y sin workspace declarado manda los datos de toda la máquina al repo donde arrancó la sesión, donde otros proyectos nunca los ven.

Con el cambio, solo se guarda lo que una sesión futura volvería a ejecutar o comprobar, nunca un diagnóstico, un conteo ni la disponibilidad momentánea de un host. Los datos que valen en cualquier proyecto van a `~/.config/hive/HIVE.local.md`, que toda sesión lee.

## Scope and acceptance

En alcance: «Local instructions» de la guía global, la skill `local-instructions`, Getting started, `workspace-and-artifacts.md`, `CHANGELOG.md`, la spec y la migración de los dos archivos locales de esta máquina.

Cada criterio de la guía se observa con `rg -c -F "<frase>" content/guidance/global.md`, que hoy da 0 (comprobado en `831e6032`).

- AC34 — La regla manda leer en cada sesión el archivo de máquina `~/.config/hive/HIVE.local.md` además de los del repositorio y el workspace · falso hoy: `~/.config/hive/HIVE.local.md` no aparece.
- AC35 — **Retired:** la prueba era una predicción; la sustituye AC44, que nombra el contenido. La regla solo deja registrar un dato que una sesión futura volvería a ejecutar o comprobar.
- AC36 — La regla prohíbe registrar el resultado de un diagnóstico, un conteo o medición, o si un host responde ahora · falso hoy: `Never record the result of a diagnosis` no aparece.
- AC37 — La regla manda al archivo de máquina lo que vale en cualquier proyecto (herramienta instalada, su versión o ausencia, perfil o región de nube, alias de host) y al del repositorio o workspace lo de ese proyecto · falso hoy: `in the machine file when it holds whatever project you work in` no aparece.
- AC38 — **Retired:** la redacción pasa a nombrar la carpeta del archivo de máquina, que puede venir de `XDG_CONFIG_HOME`; la sustituye AC45. La regla autoriza escribir el archivo de máquina y crear `~/.config/hive/` en cualquier flujo.
- AC39 — La skill manda los candidatos `missing` y `cloud` al archivo de máquina y los `commands` y `ports` al archivo que indica su `where` · falso hoy: `rg -F "go to the machine file" content/skills/local-instructions/SKILL.md` no devuelve nada.
- AC40 — La clase de descarte de la skill incluye diagnósticos, conteos y la disponibilidad momentánea de un host · falso hoy: `rg -F "whether a host answered" content/skills/local-instructions/SKILL.md` no devuelve nada.
- AC41 — Getting started y `workspace-and-artifacts.md` nombran `~/.config/hive/HIVE.local.md` · falso hoy: `rg -l "config/hive" site/src _support/docs/architecture` no devuelve nada.
- AC42 — `CHANGELOG.md` («Unreleased») describe el cambio · falso hoy: `rg -c "config/hive" CHANGELOG.md` da 0.
- AC43 — En esta máquina, `tricell-hive/HIVE.local.md` ya no tiene las entradas de SetStoreUpdateService ni de `bb1`, y `~/.config/hive/HIVE.local.md` tiene las de T3 Code, cloudflared y Antigravity · falso hoy: `rg -c "SetStoreUpdateService|bb1" HIVE.local.md` da 2 y `~/.config/hive/` no existe.

- AC44 — La regla solo deja registrar un dato que nombra un comando, ruta, puerto, herramienta, versión, perfil, región, alias de host o parche local que una sesión futura pueda ejecutar o comprobar · falso hoy: `a later session can run or check it` no aparece. **Replaces:** AC35
- AC45 — La regla autoriza, en cualquier flujo, escribir los archivos y crear la carpeta del archivo de máquina · falso hoy: `creating the machine file's folder` no aparece. **Replaces:** AC38
- AC46 — La regla pide excluir de Git cualquiera de los archivos, el de máquina incluido, cuando está dentro de un repositorio · falso hoy: `before the first write to any of them inside a Git repository` no aparece.
- AC47 — **Retired:** «solo la entrada propia» impedía marcar obsoleta la entrada de otra sesión (revisión de código); lo sustituye AC53. La regla pide releer el archivo de máquina justo antes de escribirlo y cambiar solo la entrada propia.
- AC48 — La regla manda al archivo del repositorio qué perfil sirve a un entorno de ese repositorio · falso hoy: `such as which profile serves one of its environments` no aparece.
- AC49 — La skill permite escribir en el proyecto y en el archivo de máquina, y su descripción ya no habla de un único archivo del proyecto · falso hoy: `rg -F "Work only in the current project and the machine file" content/skills/local-instructions/SKILL.md` no devuelve nada.
- AC50 — `security-advisory/SKILL.md`, `harness-audit/SKILL.md`, `flow-build/references/delegation.md` y `llms.txt` nombran el archivo de máquina · falso hoy: `rg -l "machine file" <los cuatro>` no devuelve nada.
- AC51 — La regla manda nombrar, para un dato de máquina, el archivo de credenciales propio de la herramienta, sin su valor · falso hoy: `the tool's own credential file` no aparece.
- AC52 — La regla nombra `$XDG_CONFIG_HOME/hive/` cuando esa variable está definida · falso hoy: `` `$XDG_CONFIG_HOME/hive/` `` no aparece.

- AC53 — La regla pide releer el archivo de máquina justo antes de escribirlo y cambiar solo las entradas que se añaden, actualizan o marcan obsoletas, conservando las demás · falso hoy: `change only the entries you add, update or mark stale` no aparece. **Replaces:** AC47
- AC54 — La regla dice que el archivo de máquina es la única escritura fuera del proyecto que «Scope and authorization» permite y que no es configuración global · falso hoy: `is not global configuration` no aparece en `global.md`.
- AC55 — La regla manda ejecutar Git desde la carpeta del propio archivo para excluirlo · falso hoy: `running Git from the file's own folder` no aparece.
- AC56 — Si el archivo de máquina y el de un proyecto se contradicen, manda el del proyecto · falso hoy: `the project's file holds for that project` no aparece.

## Prerequisites

Ninguno: el build no necesita nada fuera del repositorio. La migración de AC43 toca dos archivos locales sin seguimiento de esta máquina y quita dos entradas: la autoriza la aprobación del plan por el usuario, porque la regla nueva no estará desplegada hasta el merge y el refresco.

## Verified context

- `content/guidance/global.md:52-58` («Local instructions», 40 430 bytes en `831e6032`).
- `_support/openspec/specs/local-instructions/spec.md` (5 requisitos, archivado en #420).
- `content/skills/local-instructions/SKILL.md`: pasos 3 (clasificación) y 6 (escritura); el script devuelve `missing` y `cloud` sin `where` de proyecto, y `commands` y `ports` con `where`.
- Transcripción de la sesión Claude Code `de72f693`: escrituras en `HIVE.local.md` a las 03:48, 03:53, 04:41 y 04:44 UTC del 2026-11-08; dos `mem_save` de la decisión de T3 Code.
- `~/.config` ya aloja `gh`, `acli`, `gcloud` y otras CLIs; no está dentro de un repositorio Git; `~/.config/hive` no existe. De `~/.config`, Hive solo gestiona `~/.config/opencode` y respeta `XDG_CONFIG_HOME` (`tooling/cli/setup.go:90`).

## Delivery

`interactive` (D5 A): parada antes del push para leer la regla y el diff; luego push, PR a `development`, merge y refresco del contenido instalado. Revisión: `/code-review` de Claude Code (D6 A).

## Review status and progress

2026-11-07 — Plan escrito. `hive-review-plan`: 1 P1, 6 P2, 7 P3. Aplicados: la skill y otros cuatro archivos admiten el archivo de máquina (P1, nuevo T11); prueba de relevancia por contenido (AC35 → AC44); D17 A para el mapeo de perfiles; exclusión de Git y relectura antes de escribir en el archivo de máquina; credenciales de un dato de máquina; XDG; comprobaciones negativas en T7; línea de cloudflared reescrita en T10 y su autorización; fechas locales; entrada existente de `CHANGELOG.md` actualizada en vez de añadir otra. Pendiente para el cierre: el `## Purpose` de la spec, que un delta MODIFIED no cambia.
2026-11-07 — Revisión del plan durante el build: `/code-review` (10 hallazgos) llevó a retirar AC47 y añadir AC53-AC56; también: detector de pilotos para el archivo de máquina con test, `flow-research` y `flow-status`, cuatro páginas del sitio, descarte de binarios del proyecto en la skill, `XDG_CONFIG_HOME` solo absoluto, herramientas instaladas en la spec. T7, T8, T9 y T11 se reabren y cierran en el commit de esta ronda.
