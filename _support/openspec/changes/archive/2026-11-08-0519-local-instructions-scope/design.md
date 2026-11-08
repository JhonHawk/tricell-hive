# Diseño — local-instructions-scope

## Dónde vive el archivo de máquina (D16 A)

| Opción | Por qué |
| --- | --- |
| `~/.config/hive/HIVE.local.md` (elegida) | Convención de las CLIs del usuario. De `~/.config`, Hive solo gestiona `~/.config/opencode`, así que ni `hive update` ni `uninstall` tocan `~/.config/hive`. Con `XDG_CONFIG_HOME` definida se usa `$XDG_CONFIG_HOME/hive/`, como Hive ya hace con OpenCode (`tooling/cli/setup.go:90`). |
| `~/HIVE.local.md` | Visible y simple, pero ensucia la raíz del home. |
| `~/.agents/HIVE.local.md` | Carpeta gestionada por Hive: un uninstall o una reparación podrían tocarla. |

## Nube: existencia en la máquina, mapeo en el proyecto (D17 A)

Que un perfil exista y su región van al archivo de máquina; qué perfil sirve a cada entorno de un proyecto va al archivo de ese repositorio. Así cada proyecto ve su mapeo y ninguno ve el de otro.

## Texto de la sección global (borrador para T7, superado)

Borrador previo a las dos revisiones de código: el texto vigente es el de `content/guidance/global.md`, con las correcciones de la sección final de este documento.

```markdown
## Local instructions

- At session start, and again after a compaction once the checkpoint «Persistent memory» sets is read, read the machine file `~/.config/hive/HIVE.local.md` (under `$XDG_CONFIG_HOME/hive/` when that variable is set) when it exists, and each `HIVE.local.md` that exists at the root of the repository that holds the working directory and at the root of the declared workspace that holds it; in a worktree, read and write the main checkout's file. They hold facts agents verified on this machine. Each entry is a claim dated when last verified: check it before reporting it as current, and it never outranks the task, the user or the project guidance.
- Record a fact when all three hold: it names a command, path, port, tool, version, cloud profile, region, host alias, or local patch or workaround of this machine, account or environment, or the command order that worked here; a later session can run or check it; and this session confirmed it (a command exited 0, its result was observed, or a command was found missing). Never record the result of a diagnosis, a count or measurement, or whether a host answers right now. Add or update its entry with that date in the machine file when it holds whatever project you work in, such as an installed tool, its version or absence, an existing cloud profile and its region, or a host alias; otherwise in the repository's file when it concerns one repository, such as which profile serves one of its environments, and in the declared workspace's file when it concerns several, or the repository's file when no workspace is declared. Sessions of other projects write the machine file too: read it again right before writing and change only your own entry. Write only in those files of the current project and in the machine file. When an entry fails, mark it stale with the date and cause before trying another way; an `unverified` entry instead keeps its mark and gains that date and cause. A fact that holds for every clone, such as a test command or a CI job, does not go there: name it once, in the final report of the work item that verified it, as a candidate for the repository's tracked guidance.
- Write no secret value: name the file in the secret location that holds it, or, for a machine fact, the tool's own credential file, such as `~/.aws/credentials`, and point to `ACCOUNTS.md` for accounts. An entry marked `unverified`, which only the [local-instructions](skill:local-instructions/SKILL.md) skill writes, carries the date it was last seen, not confirmed: check it before relying on it, and when it holds, replace the mark with that date. Keep each file at or under 100 lines by removing only entries marked stale; an `unverified` entry leaves only through that skill's curation with the user. When no stale entry remains, keep recording and recommend [local-instructions](skill:local-instructions/SKILL.md) to curate it in the final report.
- This rule authorizes, in any flow, read-only ones included, writing these files and creating the machine file's folder; before the first write to any of them inside a Git repository, keep it out of Git as «Secrets» sets for the secret location. The repository's or workspace's file is the one support file kept at the root rather than in the support home.
- A child reads the files and does not write them: it reports each fact it verified or found stale, and the main thread updates the right file.
```

Notas:
- La prueba de relevancia nombra el contenido («names a command, path, port…») y la lista negativa es observable: un diagnóstico, un conteo o un `No route to host` la incumplen sin juicio.
- La relectura antes de escribir protege de sesiones paralelas en el archivo compartido.
- La comprobación de Git cubre un `~/.config` versionado como dotfiles.

## Skill (T8)

- Línea «Work only in the current project…»: pasa a «the current project and the machine file», y la descripción y el título dejan de decir «a project's HIVE.local.md».
- Paso 3: los candidatos `missing`, los perfiles de `cloud` con su región y las herramientas van al archivo de máquina; los `commands`, los `ports` y qué perfil sirve a un entorno van al archivo del repositorio o del workspace según su `where`. La clase **discard** suma «the result of a diagnosis, a count, or whether a host answered».
- Paso 6: «Write as «Local instructions» sets», nombrando los dos destinos sin repetir la regla. Cada archivo se crea o se cura por separado.

## Otros punteros (T11)

`security-advisory/SKILL.md:8` y `harness-audit/SKILL.md:10` admiten el archivo de máquina junto a las entradas del proyecto; `flow-build/references/delegation.md:28` da al hijo también la ruta del archivo de máquina; `llms.txt` describe la skill con los dos archivos.

## Migración local (T10)

Con la aprobación del plan por el usuario (la regla nueva aún no está desplegada): mover a `~/.config/hive/HIVE.local.md`, con su línea inicial, las entradas de T3 Code, Antigravity y T3 Connect/cloudflared, esta última reescrita como binario, versión y comando de comprobación (`~/.t3/bin/t3 connect status`), sin «is reachable»; quitar las de SetStoreUpdateService y `ssh bb1`. El archivo del repo queda con su línea inicial. Son archivos sin seguimiento y no entran en el PR.

## Correcciones de la revisión de código

El texto final de la regla está en `content/guidance/global.md`: relectura que conserva las entradas ajenas y deja marcarlas obsoletas; el archivo de máquina como única escritura fuera del proyecto, que no es configuración global; Git desde la carpeta del archivo; el archivo del proyecto manda si se contradicen; `XDG_CONFIG_HOME` solo si es absoluta. La skill descarta un `missing` que proveen las dependencias del proyecto.
