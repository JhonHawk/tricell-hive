# local-instructions-file — Instrucciones locales por máquina en `HIVE.local.md`

| Campo | Valor |
| --- | --- |
| Estado | In build (plan revisado: skill añadida) |
| Tracker | ninguno |
| Git | JhonHawk/tricell-hive · base `development` en `e0bb3a95` · rama propuesta `feat/local-instructions-file` |

## Decisiones

2026-11-07
- **D1 A** — ¿Cómo se llama el archivo local que no se commitea? `HIVE.local.md`: dice quién lo lee (la regla de Hive) y que es local, sin sugerir que un host lo carga solo.
- **D2 A** — ¿Cuándo lo lee el agente? Al iniciar la sesión, si existe; no depende de que reconozca el momento.
- **D3 A** — ¿Se copia a Engram al cerrar, como `accounts/index`? No: se reconstruye verificando de nuevo y el cierre no gana una escritura.
- **D4 A** — ¿Se propone añadirlo al `.gitignore` versionado de cada proyecto? No: el exclude que escribe el agente protege cada clon y la regla no añade preguntas.
- **D5 A** — ¿Cómo se entrega? `interactive`: parada antes del push para que el usuario revise la sección y el diff; luego push, PR y merge.
- **D6 A** — ¿Qué revisión de código pasa el candidato? La nativa de Claude Code, `/code-review`.
- **D7 A** — ¿Se aprueba la sección y el diff (`240fc93f`) para push, PR y merge? Sí, sujeto a `/code-review`; como la skill entra en el cambio, el conjunto se valida otra vez antes del push.
- **D8 B** — ¿La skill de arranque entra en este cambio? Sí: se añade a este plan y va en el mismo PR.
- **D9 A** — ¿Skill nueva o modo de `harness-audit`? Skill nueva `local-instructions`, con modos crear y curar.
- **D10 A** — ¿Qué fuentes audita? Todo: transcripciones de los últimos 30 días de este proyecto (leídas por un hijo que devuelve solo comandos que funcionaron, sin valores), memoria persistente, memoria del host y notas locales, `AGENTS.md` o README y nombres de archivos de la carpeta de secretos.
- **D11 A** — ¿Cómo entran los datos conservados? Confirmados con una comprobación de solo lectura cuando existe; el resto con su fecha de origen y marcado «unverified».
- **D13 A** — ¿Se aprueba el conjunto (regla, skill y script) para push, PR y merge? Sí.
- **D14 A** — ¿Quién puede quitar una entrada `unverified` que el usuario conservó? Solo la skill, curando con el usuario; el tope de la regla quita únicamente entradas obsoletas.
- **D12 A** — ¿Cómo se extraen los candidatos de las transcripciones? Con un script de la skill, `extract_candidates.py` (Python de librería estándar, con tests), que ejecuta el hilo principal en lugar del hijo de D10: sesiones principales de los últimos 30 días de Claude Code y Codex, cada comando emparejado con su resultado y cualquier valor enmascarado antes de mostrarlo. Grok, Pi, OpenCode, Cursor y Antigravity quedan sin cubrir.

Decididas en la investigación previa, sin ID: una sola regla y un documento libre en lugar de un script con JSON validado; un archivo neutral en lugar de `CLAUDE.local.md` (solo Claude Code, y suprime `AGENTS.md` sin un `CLAUDE.md` que lo importe) o `AGENTS.override.md` (Codex, reemplaza al `AGENTS.md` versionado); `ACCOUNTS.md` sigue igual y `HIVE.local.md` solo lo apunta. Supuestos menores tras la revisión del plan: ver `design.md` «Supuestos».

## Request

> «Se me ocurre la siguiente premisa a la que llamaré "micro-skills" […] muuuuuy continuamente se olvida de cómo levantar el proyecto localmente, pero no necesariamente se levanta igual en mi compu que en la computadora de mi compañero […] Users de pruebas […] son reutilizables, en lugar de hacer seed cada vez, reutilizamos los que ya sirven.»
>
> «Hagamos research y dame una propuesta, además si podemos irlo escalando a otras cosas que también consumen tiempo en investigar y reinvestigar.»
>
> «Lo otro, podría ser simplemente un .md en raíz (ignorado), para trabajar localmente, y como regla mandatoria hay que poner las instrucciones ahí, eso nos daría una sola regla que mantener, y el dinamismo del documento adaptado a cualquier proyecto.»
>
> «Sería como el AGENTS.md pero que NO se commitea, porque CLAUDE.local.md solo es compatible con claude.»

Sin ticket. Fuera de alcance: un script o esquema que guarde o valide los datos del archivo, hooks de inicio de sesión, cambios en `ACCOUNTS.md`, añadir el archivo a `.gitignore` versionados, cambios en `flow-walkthrough`, y escribirlo en proyectos cliente (lo crea el agente allí cuando verifique algo).

## Objective

Hoy ninguna regla dice dónde guardar lo que un agente verificó de esta máquina, así que cada sesión lo vuelve a deducir. Sesiones reales del 2026-10-16 al 11-07:
- carvento, Claude Code `8d09a5eb` (2026-11-02): 53 min re-derivando puertos, el parche de `dist` y CSRF que la sesión `94440d0a` (2026-10-24) ya había resuelto.
- cam: Codex `01a1a844` (2026-11-04), unos 27 min montando un stack aislado; `timeout` inexistente en macOS en 17 sesiones de cam y 22 de Hive; región de AWS barrida en `b7986711` (2026-10-27) y otra vez en `1776e095` (2026-10-28).
- En 21 de 39 sesiones con búsquedas de nube, el agente leyó Engram y aun así re-derivó el dato.

Con el cambio, la guía global manda leer `HIVE.local.md` al iniciar la sesión y añadir ahí, con fecha, cada dato de la máquina que se verificó, sin valores secretos.

## Scope and acceptance

En alcance: una sección de la guía global; la skill `local-instructions` (crear y curar `HIVE.local.md`) con su referencia de transcripciones; punteros en flow-build, flow-verify, delegación, la disposición de artefactos, harness-audit (HA-IF-15 y HA-PL-05) y el rol `hive-review-harness`; las páginas del sitio que describen el comportamiento o listan skills, `llms.txt`, `CHANGELOG.md` y un documento de arquitectura. Fuera: lo listado en Request.

Cada criterio de la guía global se observa con `rg -n -F "<frase>" content/guidance/global.md`, que hoy devuelve 0 coincidencias para cada frase (comprobado en `e0bb3a95`).

- AC1 — La guía global manda leer `HIVE.local.md` al iniciar la sesión, en la raíz del repositorio y en la del workspace declarado, y en un worktree el del checkout principal. **Retired:** agrupaba tres conductas con una sola búsqueda; lo sustituyen AC6 y AC7.
- AC2 — La guía global manda escribir solo datos verificados, con fecha, sin secretos, con tope de líneas y sin datos de todo clon. **Retired:** agrupaba cinco conductas; lo sustituyen AC8 a AC12.
- AC3 — La guía global manda el exclude antes de la primera escritura y que un hijo reporte en vez de escribir. **Retired:** agrupaba dos conductas; lo sustituyen AC13 y AC14.
- AC4 — Punteros en cinco archivos. **Retired:** faltaban HA-PL-05 y el rol de auditoría; lo sustituye AC15.
- AC5 — Dos páginas del sitio. **Retired:** faltaban `flows/verify.md` y el documento de arquitectura; lo sustituye AC16.
- AC6 — La guía global manda leer `HIVE.local.md` al iniciar la sesión y otra vez tras una compactación, en la raíz del repositorio y del workspace declarado, y en un worktree el del checkout principal · falso hoy: `again after a compaction` no aparece. **Replaces:** AC1
- AC7 — Trata cada entrada como afirmación con fecha que se comprueba antes de darla por vigente y que no se impone a la tarea, al usuario ni a la guía del proyecto · falso hoy: `check it before reporting it as current` no aparece. **Replaces:** AC1
- AC8 — Manda registrar un dato que nombra algo de esta máquina, cuenta o entorno y que funcionó en la sesión, con fecha, en el archivo del repositorio o del workspace según a cuántos repositorios concierne · falso hoy: `names something of this machine` no aparece. **Replaces:** AC2
- AC9 — Manda marcar obsoleta, con fecha y causa, una entrada que falla antes de probar otra vía · falso hoy: `before trying another way` no aparece. **Replaces:** AC2
- AC10 — Prohíbe valores secretos y manda nombrar el archivo de la carpeta de secretos y `ACCOUNTS.md` · falso hoy: `Write no secret value` no aparece. **Replaces:** AC2
- AC11 — Fija el tope de 100 líneas quitando solo entradas obsoletas, y si no quedan, informar sin borrar · falso hoy: `removing only entries marked stale` no aparece. **Replaces:** AC2
- AC12 — Un dato que vale para todo clon no se guarda y se nombra una vez en el informe final como candidato a la guía versionada · falso hoy: `candidate for the repository's tracked guidance` no aparece. **Replaces:** AC2
- AC13 — Autoriza escribirlo en cualquier flujo, de solo lectura incluidos, y añadirlo al exclude antes de la primera escritura si Git no lo ignora · falso hoy: `read-only ones included` no aparece. **Replaces:** AC3
- AC14 — Un hijo lo lee, no lo escribe, y reporta lo verificado u obsoleto · falso hoy: `reads the file and does not write it` no aparece. **Replaces:** AC3
- AC15 — `flow-build/SKILL.md`, `flow-verify/SKILL.md`, `flow-build/references/delegation.md`, `workspace-conventions/references/artifact-layout.md`, `harness-audit/references/project-layout.md` (HA-PL-05), `harness-audit/references/instruction-files.md` (HA-IF-15) y `content/agents/review/hive-review-harness.md` nombran «Local instructions» o `HIVE.local.md` · falso hoy: `rg -c "Local instructions|HIVE.local.md" content` no devuelve nada. **Replaces:** AC4
- AC16 — `site/src/content/docs/getting-started.md`, `flows/build.md`, `flows/verify.md` y `_support/docs/architecture/workspace-and-artifacts.md` describen `HIVE.local.md` · falso hoy: `rg -c "HIVE.local.md" site/src _support/docs/architecture` no devuelve nada. **Replaces:** AC5
- AC17 — Existe `content/skills/local-instructions/SKILL.md` con `name` y `description`, a lo sumo 16 000 bytes, con dos modos: crear el archivo y curarlo · falso hoy: el archivo no existe.
- AC18 — **Retired:** el hijo lector de transcripciones lo sustituye el script (D12 A); lo sustituyen AC24 y AC25. La skill reúne candidatos de memoria persistente, memoria del host y notas locales, `AGENTS.md` o README y nombres de archivos de la carpeta de secretos, y delega en un hijo las transcripciones de los últimos 30 días de este proyecto, que devuelve solo comandos que funcionaron sin ningún valor · falso hoy: no existe la skill.
- AC19 — **Retired:** faltaba cómo cabe la entrevista en la herramienta de preguntas y qué pasa con la memoria; lo sustituye AC26. Antes de escribir nada, la skill entrevista al usuario con la herramienta de preguntas para conservar o descartar cada grupo de candidatos, y en modo curar también las entradas existentes · falso hoy: no existe la skill.
- AC20 — **Retired:** faltaban las restricciones de las comprobaciones y qué hace un fallo; lo sustituye AC27. La skill confirma cada dato conservado con una comprobación de solo lectura de una lista fija cuando existe una, escribe el resto con su fecha de origen marcado `unverified`, y nunca ejecuta comandos que arrancan, siembran o migran · falso hoy: no existe la skill.
- AC21 — **Retired:** repetía pasos de la regla; lo sustituye AC28. La skill escribe como fija «Local instructions» (sin valores secretos, tope, exclude) y nombra en su informe, sin guardarlos, los datos que valen para todo clon · falso hoy: no existe la skill.
- AC22 — **Retired:** con el script cambia qué hosts se cubren; lo sustituye AC29. `content/skills/local-instructions/references/transcript-sources.md` da la ruta de las transcripciones de Claude Code, Codex, Grok y Pi, y dice que las de OpenCode, Cursor y Antigravity no están localizadas · falso hoy: el archivo no existe.
- AC23 — **Retired:** faltaba `getting-started.md`; lo sustituye AC30. La línea del tope en «Local instructions» recomienda la skill, y `site/src/content/docs/skills.md`, `index.mdx`, `llms.txt` y `CHANGELOG.md` la nombran · falso hoy: `rg -l "local-instructions" content/guidance site/src llms.txt CHANGELOG.md` no devuelve nada.
- AC24 — La skill reúne candidatos de memoria persistente, memoria del host y notas locales, la guía versionada y los nombres (nunca el contenido) de los archivos de la carpeta de secretos, y ejecuta `scripts/extract_candidates.py` sobre las sesiones principales de los últimos 30 días de Claude Code y Codex de este proyecto, incluidos sus worktrees y las sesiones abiertas en la raíz del workspace · falso hoy: no existe la skill. **Replaces:** AC18
- AC25 — `extract_candidates.py` no imprime ningún valor: con transcripciones de prueba que llevan una contraseña en una asignación de entorno, en la parte de usuario de una URL, en una cabecera `Authorization` y tras `--password`, ninguno de esos textos aparece en su salida; devuelve los comandos que terminaron con 0 y los comandos que faltaron, con número de sesiones y fecha más reciente, y deja fuera otro proyecto con el mismo prefijo de ruta, las transcripciones de subagentes y las sesiones fuera de la ventana · falso hoy: el script y su test no existen. **Replaces:** AC18
- AC26 — Antes de escribir nada, la skill lista en el texto cada grupo con sus candidatos numerados y pregunta por grupo con la herramienta de preguntas (conservar los recomendados, conservar todos, descartar todos o elegir por número en el chat); las respuestas quedan en `HIVE.local.md`, no en memoria persistente, y descartar un candidato no borra su fuente · falso hoy: no existe la skill. **Replaces:** AC19
- AC27 — La skill confirma lo conservado solo con su lista de comprobaciones de solo lectura (`git config --get` solo con claves que no guardan credenciales; `--version` solo tras `command -v` y nunca con `npx` ni scripts del proyecto), escribe como `unverified` con su fecha de origen lo que no se confirma o falla y el usuario conserva, y nunca ejecuta un comando del proyecto que arranca, para, siembra, migra, instala, inicia sesión o escribe · falso hoy: no existe la skill. **Replaces:** AC20
- AC28 — La skill escribe el archivo con un puntero a «Local instructions», sin repetir sus pasos, y nombra en su informe los datos que valen para todo clon · falso hoy: no existe la skill. **Replaces:** AC21
- AC29 — `references/transcript-sources.md` da las rutas y codificaciones de carpeta de Claude Code (con `/` y `.` como `-`, y las carpetas de worktrees) y de Codex, y dice que Grok, Pi, OpenCode, Cursor y Antigravity no los cubre el script · falso hoy: el archivo no existe. **Replaces:** AC22
- AC30 — La línea del tope en «Local instructions» recomienda la skill, y `site/src/content/docs/skills.md`, `index.mdx`, `getting-started.md`, `llms.txt` y `CHANGELOG.md` la nombran · falso hoy: `rg -l "local-instructions" content/guidance site/src llms.txt CHANGELOG.md` no devuelve nada. **Replaces:** AC23
- AC31 — **Retired:** el tope la dejaba quitar a cualquier agente (D14 A); lo sustituye AC32. «Local instructions» dice que una entrada marcada `unverified` lleva la fecha en que se vio, se comprueba antes de usarla y el tope puede quitarla como una obsoleta · falso hoy: `rg -F "unverified" content/guidance/global.md` no devuelve nada.

- AC32 — «Local instructions» dice que una entrada `unverified` la escribe solo la skill, lleva la fecha en que se vio, se comprueba antes de usarla, pierde la marca al confirmarse y solo sale del archivo curando con el usuario · falso hoy: `rg -F "leaves only through that skill's curation" content/guidance/global.md` no devuelve nada. **Replaces:** AC31

## Prerequisites

Ninguno: el build no necesita nada fuera del repositorio.

## Verified context

- `content/guidance/global.md` (37 751 bytes en `e0bb3a95`): «Secrets» (líneas 34-41) ya fija el exclude en `$(git rev-parse --git-common-dir)/info/exclude`; «Test accounts» (43-50) es el patrón de índice por máquina con hijo que reporta y autoriza «in any flow or outside one»; «Support folder» (64-66) envía los artefactos a `_support/`; «Proportionality» (62) define la lectura grande en más de 100 líneas.
- `_support/openspec/specs/test-accounts/spec.md`: el precedente que funciona; 8 sesiones de Codex leyeron `ACCOUNTS.md` sin redescubrir cuentas.
- Diseño archivado `2026-11-05-0920-test-accounts-index/design.md`: una regla en una referencia se perdió (#284), por eso la regla va en la guía global.
- `content/skills/harness-audit/references/hierarchy.md:31-35`: en Claude Code y Codex ningún archivo local neutral se carga solo; `CLAUDE.local.md` suprime `AGENTS.md` (HA-HI-03) y `AGENTS.override.md` reemplaza. Los demás hosts no están documentados ahí.
- `harness-audit/references/project-layout.md:11` (HA-PL-05) informa «support material at a repository root»; `artifact-layout.md:18` lo prohíbe; `hive-review-harness.md:16` prohíbe proponer escribir `CLAUDE.local.md`.
- `flow-status/SKILL.md:8` es de solo lectura: la autorización de AC13 lo cubre.
- `tests/pilot/regression.go:910` exenta solo `ACCOUNTS.md` del detector de lectura de secretos; `HIVE.local.md` no está en la carpeta de secretos, así que no cambia.
- `git check-ignore HIVE.local.md` no devuelve nada en este repositorio.

## Delivery

`interactive` (D5 A): tras el verify, parada antes del push para que el usuario lea la sección nueva de `global.md` y el diff; con su visto bueno, push de `feat/local-instructions-file`, PR a `development` y merge. Revisión: `/code-review` de Claude Code (D6 A) sobre el candidato antes de la parada. Tras añadir la skill (D8 B) hay una segunda parada de validación con el conjunto completo antes del push. Con el marcador de mantenedor presente, el merge en `development` es la petición de refrescar la instalación local (`hive update --source`). Esto no autoriza un release ni cambios en proyectos cliente.

## Review status and progress

2026-11-07 — Plan escrito; `hive-review-plan` devolvió 7 hallazgos P2 y 9 P3, sin P0 ni P1. Aplicados: criterios divididos con observación propia (AC1-AC5 retirados, AC6-AC16), HA-PL-05, `hive-review-harness`, `flows/verify.md` y `workspace-and-artifacts.md` en alcance, autorización en flujos de solo lectura, archivo de destino según el número de repositorios, umbral observable de dato de la máquina, poda que no borra entradas vigentes, `## Purpose` en el delta, delta alineado con el borrador, comprobación antes de dar una entrada por vigente, relectura tras compactación, tope de 100 líneas, informe del hijo en el puntero de delegación, sesiones en el mensaje de commit, supuestos marcados y comando del build del sitio.
2026-11-07 — Revisión del plan: `/code-review` dio 10 hallazgos, aplicados en `87fef783`; el usuario añadió la skill de arranque (D8 B, D9-D11 A). Nuevos AC17-AC23 y tareas T4-T5; no invalida T1-T3.
2026-11-07 — Tercera revisión del plan (revisión con D12): 1 P1, 7 P2, 6 P3. Aplicados: subagentes de Codex excluidos (`exec` y `vscode` cuentan como sesiones principales, supuesto menor), enmascarado por lista de permitidos con los casos de la revisión en el test, éxito de Claude sin comandos en segundo plano, «faltó» anclado a mensajes de error completos, `where` por candidato, `--version` sin shims, excepción de memoria nombrada, alcance reescrito.
2026-11-07 — Segunda revisión del plan: 2 P1, 7 P2, 5 P3. El usuario eligió el script (D12 A). Retirados AC18-AC23; nuevos AC24-AC31; T4 pasa a ser el script, T5 la skill y T6 los enlaces.
2026-11-07 — Segunda validación: el usuario aprobó (D13 A) y eligió que solo la skill quite entradas `unverified` (D14 A). AC31 retirado, AC32 nuevo; cambia una frase de la regla y la spec.
