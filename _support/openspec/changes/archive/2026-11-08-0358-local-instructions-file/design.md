# Diseño — local-instructions-file

## Enfoques considerados

| Enfoque | Por qué no |
| --- | --- |
| Script Python con `facts.json` validado en `_support/local/` | Valida y marca lo obsoleto, pero añade código, tests y un formato rígido; el usuario prefiere un documento libre con una sola regla, y `ACCOUNTS.md` muestra que la regla sola funciona. |
| Subcomando `hive facts` en Go | Igual que el anterior, más recompilar el binario en cada máquina y el desfase entre binario y contenido. |
| `CLAUDE.local.md` | Solo Claude Code, y suprime `AGENTS.md` cuando no hay un `CLAUDE.md` que lo importe (HA-HI-03). |
| `AGENTS.override.md` | Codex lo usa en lugar del `AGENTS.md` versionado de ese directorio: se perderían las reglas del repo. |
| Engram con `topic_key` fijos | Es el mecanismo que falló: 21 de 39 sesiones leyeron Engram y re-derivaron el dato. |
| Hook de inicio de sesión | Hive evita hooks sin necesidad medida; la guía global ya se despliega en todos los hosts. |

Elegido: un archivo libre `HIVE.local.md` y una sección de la guía global que lo hace obligatorio. La guía global es la capa que Hive despliega en cada host, así que es la que convierte el archivo en «el `AGENTS.md` de esta máquina». Supuesto sin verificar: que Grok, Pi, OpenCode, Cursor y Antigravity no cargan por su cuenta un archivo local neutral; `harness-audit/references/hierarchy.md` solo lo documenta para Claude Code y Codex.

## Ubicación y qué archivo se escribe

En la raíz del repositorio que contiene el directorio de trabajo, y en la raíz del workspace declarado que lo contiene (cam y carvento tienen su `_support/` a nivel de workspace, fuera de cualquier repo). En un worktree se lee y se escribe el del checkout principal, como `ACCOUNTS.md`. Un dato que concierne a un repositorio va en el de ese repositorio; uno que concierne a varios o a toda la máquina, en el del workspace. Es el único archivo de soporte en la raíz, para que el usuario lo vea en el Explorer; HA-PL-05 y `artifact-layout.md` lo exceptúan.

## Supuestos (decisiones menores, revisión del plan)

- La regla autoriza escribir el archivo y su línea de exclude en cualquier flujo, incluidos los de solo lectura (`flow-status`, `security-advisory`, `harness-audit`): es un archivo local ignorado, como la autorización de «Test accounts» «in any flow or outside one».
- Tope blando de 100 líneas para acotar el coste de leerlo en cada sesión: al pasarlo sin entradas obsoletas se sigue registrando y se informa el tamaño. La revisión de código corrigió la razón anterior: leer un documento entero siempre es lectura grande («Proportionality»).
- Para no pasar el tope solo se quitan entradas marcadas obsoletas, así no se pierden notas del usuario («Preservation»).
- Sin workspace declarado, los datos de toda la máquina van al archivo del repositorio; solo se escribe en el proyecto actual. El exclude sigue el procedimiento de «Secrets» por puntero, sin copiarlo.
- Un dato que vale para todo clon se nombra una sola vez, en el informe final del trabajo que lo verificó.
- `flow-walkthrough` no cambia: también sirve el candidato en local, pero la regla global actúa en todos los flujos y su lectura ocurre al iniciar la sesión.

## Texto de la sección global (borrador para T1)

Va justo después de «Test accounts». El texto final puede ajustar redacción, no reglas.

```markdown
## Local instructions

- At session start, and again after a compaction once the checkpoint «Persistent memory» sets is read, read each `HIVE.local.md` that exists at the root of the repository that holds the working directory and at the root of the declared workspace that holds it; in a worktree, read and write the main checkout's file. It holds facts agents verified on this machine. Each entry is a claim dated when last verified: check it before reporting it as current, and it never outranks the task, the user or the project guidance.
- Record a fact there when it names something of this machine, account or environment, such as a path, a port in use, an installed tool, its version or its absence, a cloud profile, region or host, a local patch or workaround, or the command order that worked here, and this session confirmed it (a command exited 0, its result was observed, or a command was found missing): add or update its entry with that date, in the repository's file when it concerns one repository, and in the declared workspace's file when it concerns several or the whole machine, or in the repository's file when no workspace is declared. Write only in the current project, as «Scope and authorization» sets. When an entry fails, mark it stale with the date and cause before trying another way. A fact that holds for every clone, such as a test command or a CI job, does not go there: name it once, in the final report of the work item that verified it, as a candidate for the repository's tracked guidance.
- Write no secret value: name the file in the secret location that holds it, and point to `ACCOUNTS.md` for accounts. Keep the file at or under 100 lines by removing only entries marked stale; when none remain, keep recording and report its size in the final report so the user can choose what to drop.
- This rule authorizes writing the file in any flow, read-only ones included; before its first write inside a Git repository, keep it out of Git as «Secrets» sets for the secret location. It is the one support file kept at the root rather than in the support home.
- A child reads the file and does not write it: it reports each fact it verified or found stale, and the main thread updates the file.
```

Cláusula de T6 (final, tras D14 A): «An entry marked `unverified`, which only the [local-instructions](skill:local-instructions/SKILL.md) skill writes, carries the date it was last seen, not confirmed: check it before relying on it, and when it holds, replace the mark with that date. Keep the file at or under 100 lines by removing only entries marked stale; an `unverified` entry leaves only through that skill's curation with the user. When no stale entry remains, keep recording and recommend [local-instructions](skill:local-instructions/SKILL.md) to curate it in the final report.» Y en el punto de escritura: un fallo marca la entrada obsoleta, salvo una `unverified`, que conserva su marca con la fecha y la causa.

No toca «Persistent memory» (D3 A) ni añade preguntas (D4 A).

## Punteros y documentos (T2, T3)

Cada uno nombra la regla y cuándo leerla, sin repetir sus pasos:
- `flow-build/SKILL.md`, párrafo de reparación del entorno local: tras repararlo, registrar lo verificado como fija «Local instructions».
- `flow-verify/SKILL.md` paso 1, junto a «Test accounts»: servir el candidato con lo que registra `HIVE.local.md`, actualizándolo como fija «Local instructions».
- `flow-build/references/delegation.md` «Every child»: un hijo que levanta, sirve o prueba el proyecto, o usa sus entornos, recibe la ruta de `HIVE.local.md` (la del checkout principal en un worktree) y reporta los datos que verificó o encontró obsoletos. Supuesto sin verificar: que algún host no entrega la guía global a los hijos; el puntero cubre ese caso sin coste si no ocurre.
- `workspace-conventions/references/artifact-layout.md`, junto a la línea «Even a small task or single file belongs there rather than the repository root»: la excepción de `HIVE.local.md`.
- `harness-audit/references/project-layout.md` HA-PL-05: `HIVE.local.md` en la raíz no es material fuera de sitio.
- `harness-audit/references/instruction-files.md` HA-IF-15: leer y verificar `HIVE.local.md` como `CLAUDE.local.md` e informar como `info`; no lo propone como hallazgo a escribir, y eso no impide que la regla global lo escriba.
- `content/agents/review/hive-review-harness.md` paso 4: misma distinción que HA-IF-15.
- Sitio: `getting-started.md` («Give the session context»), `flows/build.md` y `flows/verify.md` («What it does»).
- `_support/docs/architecture/workspace-and-artifacts.md`: la excepción de la raíz.

## Skill `local-instructions` (D8 B, D9-D12 A)

Crea o cura `HIVE.local.md` a partir de lo que ya funcionó en la máquina. La regla global sigue siendo el hogar canónico del archivo; la skill añade el arranque y la curación, y nombra la única excepción que introduce.

1. **Lectura:** los `HIVE.local.md` existentes, `AGENTS.md` y README (para separar lo que vale para todo clon), `ACCOUNTS.md` solo como puntero, los nombres (nunca el contenido) de los archivos de la carpeta de secretos, la memoria del host y `CLAUDE.local.md`, y la memoria persistente con búsquedas cortas.
2. **Transcripciones (D12 A):** contrato del script (tercera revisión): ventana por fecha de modificación del archivo; Claude Code cuenta como éxito `is_error` falso sin `run_in_background`; Codex, `exit_code` 0; «faltó» solo con un mensaje de error de línea completa (`zsh: command not found: x`, `bash: line N: x: command not found`, su forma en español o `sh: N: x: not found`) en un resultado con error o con 127; se excluyen subagentes de ambos hosts; Codex desenvuelve `[shell, -lc, script]`; se toma la primera línea del comando, partida en `&&`, `||`, `;`, `|`; un segmento es candidato si su programa está en la lista `RELEVANT` del script, empieza por `./`, o es `curl` a localhost; se omite el código en línea (`python3 -`, `node -e`); clave de agrupación: el segmento enmascarado; hasta 60 filas por lista, por número de sesiones y luego fecha; campos `command`, `env`, `sessions`, `last_seen`, `last_session`, `where`. Enmascarado por lista de permitidos: se conservan programa, palabras, nombres de banderas, números y rutas; se enmascaran el valor tras banderas de credencial o carga útil (`-H`, `-u`, `-a`, `-p`, `--password`, `--data`…), el valor de una bandera larga que no esté en `KEEP_NEXT`, los grupos de banderas cortas que llevan valor (`-pS3cret`), la consulta y la parte de usuario de las URL, los textos con comillas o espacios, los largos y los de alta entropía, y los valores de asignaciones salvo números, booleanos, rutas y variables públicas o de entorno (`AWS_PROFILE`, `NEXT_PUBLIC_*`); un nombre de variable con `pass`, `secret`, `token` o `key` siempre se enmascara. Nunca imprime la salida de un comando. El hilo principal ejecuta `scripts/extract_candidates.py --root <raíz> [--root <raíz del workspace>] --days 30`. El script lee solo sesiones principales de Claude Code y Codex cuyo `cwd` es la raíz o está debajo (worktrees incluidos), empareja cada comando con su resultado (Claude: `tool_use` Bash con su `tool_result` e `is_error`; Codex: `CommandExecution` con `exit_code`), toma como «faltó» el código 127 o el texto «command not found» de una herramienta concreta, enmascara antes de imprimir los valores de asignaciones de entorno (salvo los solo numéricos), la parte de usuario de las URL, lo que sigue a cabeceras y a banderas de credenciales, y los textos entre comillas o largos, y devuelve JSON agregado: comandos de arranque, servicio, prueba, nube y acceso remoto que terminaron con 0, comandos que faltaron, puertos, perfiles y regiones, cada uno con número de sesiones y fecha más reciente, hasta un límite. No hay hijo: la salida ya está acotada y enmascarada.
3. **Clasificación:** dato de la máquina, dato de todo clon (se informa, no se guarda) o descarte (stack, ticket o ruta que ya no existe, duplicado o contradicho).
4. **Entrevista:** el texto lista cada grupo con sus candidatos numerados; la herramienta de preguntas hace una pregunta por grupo con cuatro opciones: conservar los recomendados (★), conservar todos, descartar todos, elegir por número respondiendo en el chat. Se recomienda conservar lo confirmado o visto en más de una sesión. Las respuestas quedan en `HIVE.local.md`, no en memoria persistente; descartar no borra la fuente (para limpiar memoria está `harness-audit`).
5. **Confirmación (D11 A):** lista fija de solo lectura: `command -v`; `<tool> --version` solo si `command -v` lo resolvió y nunca con `npx` ni scripts del proyecto; `docker ps --format` y `docker compose ls`; `lsof -nP -iTCP:<port> -sTCP:LISTEN`; `aws configure list-profiles` y `aws configure get region --profile <p>`; `test -e <path>`; `git config --get` solo para `user.name`, `user.email`, `core.*` e `init.defaultBranch`. Si no hay comprobación, o falla y el usuario conserva el dato, entra con su fecha de origen y `unverified`. Nunca un comando del proyecto que arranca, para, siembra, migra, instala, inicia sesión o escribe.
6. **Escritura:** «write as «Local instructions» sets», más el formato propio: una línea inicial que lo identifica, un encabezado por grupo, una línea por entrada con su fecha y `unverified` cuando aplica.
7. **Informe:** archivos y líneas, entradas por grupo y cuántas `unverified`, lo descartado, los datos de todo clon como candidatos para `AGENTS.md`, y los hosts no cubiertos.

La excepción a «this session confirmed it» la nombra la skill; lo que significa `unverified` para cualquier lector lo dice la regla global (AC32), porque las sesiones posteriores solo leen la regla.

Referencia `references/transcript-sources.md`: Claude Code `<Claude home>/projects/<cwd con / y . como ->/*.jsonl` (los worktrees bajo `.claude/worktrees/` tienen su propia carpeta; los subagentes en `<sesión>/subagents/` se excluyen); Codex `<Codex home>/sessions/YYYY/MM/DD/rollout-*.jsonl` filtrado por el `cwd` de `session_meta`. Grok (`<Grok home>/sessions/<cwd URL-encoded>/<id>/updates.jsonl`) y Pi (`<Pi home>/sessions/--<cwd con / como ->--/*_<id>.jsonl`) se localizaron pero el script no los cubre; OpenCode, Cursor y Antigravity no se localizaron.
