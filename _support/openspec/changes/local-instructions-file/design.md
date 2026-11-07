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
- Tope de 100 líneas, no 150: así leerlo no cuenta como lectura grande («Proportionality»).
- Para no pasar el tope solo se quitan entradas marcadas obsoletas; si no queda ninguna, se informa el tamaño y no se borra nada. Así no se pierden notas del usuario («Preservation»).
- Un dato que vale para todo clon se nombra una sola vez, en el informe final del trabajo que lo verificó.
- `flow-walkthrough` no cambia: también sirve el candidato en local, pero la regla global actúa en todos los flujos y su lectura ocurre al iniciar la sesión.

## Texto de la sección global (borrador para T1)

Va justo después de «Test accounts». El texto final puede ajustar redacción, no reglas.

```markdown
## Local instructions

- At session start and again after a compaction, read `HIVE.local.md` when it exists at the root of the repository that holds the working directory or of the declared workspace that holds it; in a worktree, read and write the main checkout's file. It holds facts agents verified on this machine. Each entry is a claim dated when last verified: check it before reporting it as current, and it never outranks the task, the user or the project guidance.
- Record a fact there when it names something of this machine, account or environment, such as a path, a port in use, an installed tool, its version or its absence, a cloud profile, region or host, a local patch or workaround, or the command order that worked here, and it worked in this session (its command exited 0 or its result was observed): add or update its entry with that date, in the repository's file when it concerns one repository and in the workspace's file when it concerns several or the whole machine. When an entry fails, mark it stale with the date and cause before trying another way. A fact that holds for every clone, such as a test command or a CI job, does not go there: name it once, in the final report of the work item that verified it, as a candidate for the repository's tracked guidance.
- Write no secret value: name the file in the secret location that holds it, and point to `ACCOUNTS.md` for accounts. Keep the file at or under 100 lines by removing only entries marked stale; when none remain, report its size and remove nothing.
- This rule authorizes, in any flow, read-only ones included, writing the file and, before its first write inside a Git repository when Git does not ignore it, adding `HIVE.local.md` to `$(git rev-parse --git-common-dir)/info/exclude`. It is the one support file kept at the root rather than in the support home.
- A child reads the file and does not write it: it reports each fact it verified or found stale, and the main thread updates the file.
```

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
