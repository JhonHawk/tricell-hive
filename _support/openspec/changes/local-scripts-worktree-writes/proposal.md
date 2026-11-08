# local-scripts-worktree-writes — Scripts locales junto a HIVE.local.md y escritura diferida desde worktrees

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build · pendiente de tu aprobación |
| Tracker | ninguno |
| Git | JhonHawk/tricell-hive · base `development` en `b14ae746` · rama propuesta `feat/local-scripts-worktree-writes` |

## Decisiones

2026-11-08
- **D1 A** — ¿Cómo escribe el agente `HIVE.local.md` y `ACCOUNTS.md` del checkout principal cuando el host lo bloquea desde un worktree? Se difiere: la actualización queda pendiente y se escribe desde el checkout principal en cuanto la sesión sale del worktree, a más tardar en el cierre. Nunca se escribe una copia dentro del worktree.
- **D2 A** — ¿Dónde viven los scripts personales que nombra `HIVE.local.md`? En `_support/scripts/` del repositorio, o del workspace declarado cuando abarca varios, siempre fuera de Git como `secrets/`.
- **D3 B** — ¿Quién autoriza crear o modificar un script local? El agente los crea sin preguntar mientras construye, los nombra en el informe final y el usuario los revisa después. Tras la revisión del plan (H8) se concreta en una lista: `flow-build`, incluidos los arreglos de `flow-verify`, y una tarea operativa; nunca research, status, plan, una revisión ni un hijo.
- **D6 A** — ¿Dónde quedan las escrituras diferidas mientras la sesión está en el worktree? En un archivo ignorado del worktree, `_support/workspace/YYYY-MM-DD-<trabajo>/pending-local.md`; el checkpoint solo lo señala, porque `tasks.md` se versiona. **Replaces:** D1 A, solo en dónde se guarda lo pendiente (antes: el checkpoint).
- **D7 A** — Con lo que mostró la segunda revisión, ¿cómo se escribe desde un worktree aislado en `HIVE.local.md`, `ACCOUNTS.md` o un script local del checkout principal? Por shell, bajo los permisos del usuario: ante el rechazo, la misma escritura se hace con un comando de shell (`>>` para una entrada nueva, edición de esa sola línea para una existente, o la escritura del script). Nunca se escribe una copia en el worktree. **Replaces:** D1 A y D6 A. La razón: diferir dependía de una salida del worktree que en Claude Code solo ocurre a pedido del usuario y que con `remove` borra lo pendiente (H11). Claude Code documenta que una escritura por shell al checkout principal «isn't refused» por el aislamiento y queda bajo «your permissions and sandboxing settings».
- **D8 A** — Sabiendo que la ruta prevista por Claude Code para cambiar el checkout principal es «edit it yourself or start an interactive session there», y que la escritura por shell usa un hueco que el aislamiento no vigila, ¿se mantiene D7 A? Sí, con la edición de una línea comprobada con `diff`, la comprobación del exclude desde el worktree y los scripts ejecutables. `host-limits.md` lo describe como un hueco conocido que usa Hive. Corrige la presentación de D7, que lo había llamado «vía documentada».
- **D4 B** y **D5 A**, en «Delivery».

## Request

> «ese tipo de cosas son lo que debería generarse para HIVE.local.md no? creo que esto va pudiendo ser una carpeta .hive ignorada, el HIVE.local.md tendría las instrucciones y se pueden tener subcarpetas con scripts o cosas similares»
>
> «entonces, podríamos pedir que se escriban los scripts en _support/scripts/ y referenciarlo en el .md no? Aunque tengo duda, qué pasa con las rutas y archivos ignorados cuando se crean los worktrees»
>
> Respuesta A al bloque del hallazgo H1: «Llevar a plan un solo cambio en Hive: scripts personales en `_support/scripts/` referenciados desde `HIVE.local.md`, más la corrección de H1».

Sin ticket. Viene de evaluar pgx/pgrun: la alternativa nativa (una base de Postgres clonada por worktree con `CREATE DATABASE … TEMPLATE`) necesita un script personal que no debe entrar al repo del cliente.

## Objective

Hoy:
- La guía global y las specs piden, en un worktree, leer y escribir el `HIVE.local.md` y el `ACCOUNTS.md` del checkout principal. Pero `git-workflow` pide usar la función nativa de worktrees del host, y en Claude Code 2.1.319 una sesión aislada en un worktree rechaza `Edit`/`Write` sobre el checkout principal. El error sugiere editar una copia en el worktree, que está ignorada y se pierde al borrarlo (H1, Engram `hosts/claude-code-worktree-isolation-main-checkout`).
- No hay lugar para scripts personales del agente: «Local instructions» solo guarda hechos, el layout de `_support/` no tiene `scripts/`, «Support folder» dice que los scripts se quedan en su ubicación de origen y «Proportionality» pide preguntar antes de agregar un script.

Al terminar:
- Una escritura que el host rechaza desde un worktree se hace por shell sobre el archivo del checkout principal, bajo los permisos del usuario.
- Una entrada de `HIVE.local.md` puede nombrar un script local en `_support/scripts/`, fuera de Git, que se ejecuta desde cualquier worktree por su ruta bajo el checkout principal.

## Scope and acceptance

Dentro:
- la guía global (`content/guidance/global.md`: «Local instructions», «Test accounts», «Support folder» y la frase de scripts de «Proportionality»);
- `content/skills/workspace-conventions/references/artifact-layout.md`;
- las specs `local-instructions` y `test-accounts`;
- `site/src/content/docs/getting-started.md`, `site/src/content/docs/flows/build.md`, `site/src/content/docs/host-limits.md` y `_support/docs/architecture/workspace-and-artifacts.md`.

Fuera:
- cambiar cómo `git-workflow` crea los worktrees (D1 C descartada);
- copiar ignorados con `.worktreeinclude`;
- resolver la raíz principal en las copias independientes de Grok (issue #97: aquí solo se pregunta la ruta);
- que el skill `local-instructions` proponga scripts como candidatos. Límite conocido (H10, P3): al curar `HIVE.local.md`, quitar una entrada deja su script en disco sin referencia;
- los detectores de `tests/pilot` (pilotos pausados; los scripts nunca se escriben en flujos de solo lectura);
- cualquier archivo de otros proyectos.

- AC1 — **Retired:** reemplazado por AC7 tras la revisión del plan (H3, H5): el checkpoint se versiona y el cierre lo borraba antes de aplicar lo pendiente.
- AC2 — **Retired:** reemplazado por AC13 (D7 A). Texto anterior: «Test accounts» aplica a `ACCOUNTS.md` la escritura diferida con un puntero a «Local instructions», sin repetir sus pasos · falso hoy porque «Test accounts» no menciona qué hacer cuando la escritura se rechaza.
- AC3 — **Retired:** reemplazado por AC8 (H2, H7: un script no se podía escribir desde el worktree y la ruta era ambigua).
- AC4 — **Retired:** reemplazado por AC9 (H4, H8: «Support folder» contradecía la regla y «un flujo que cambia archivos» no era comprobable).
- AC5 — `artifact-layout.md` lista `scripts/` como destino del hogar de apoyo, nunca versionado, que remite a «Local instructions» · falso hoy porque el layout no tiene `scripts/`.
- AC6 — **Retired:** reemplazado por AC14 (D7 A y H15). Texto anterior: `getting-started.md` y `workspace-and-artifacts.md` describen los scripts locales y la escritura diferida desde un worktree sin copiar la regla, y la frase de la arquitectura sobre «permanent maintenance scripts» admite la excepción · falso hoy porque ninguno lo menciona.
- AC7 — **Retired:** reemplazado por AC11 (D7 A). Texto anterior: «Local instructions» dice que, cuando el host rechaza desde un worktree la escritura del archivo del checkout principal, la actualización se agrega a `_support/workspace/YYYY-MM-DD-<trabajo>/pending-local.md` dentro del worktree y el checkpoint lo señala. Cuando la sesión sale del worktree sin borrarlo, cada pendiente se aplica al archivo del checkout principal, a más tardar antes de que el cierre borre el worktree. Nunca se escribe una copia del archivo dentro del worktree · falso hoy porque el texto solo dice «in a worktree, read and write the main checkout's file».
- AC8 — **Retired:** reemplazado por AC12 (D7 A: el script ya no se mueve). Texto anterior: «Local instructions» permite que el `HIVE.local.md` de un proyecto (nunca el archivo de máquina) nombre un script local en `_support/scripts/<nombre>` bajo la raíz del archivo que lo nombra, fuera de Git como fija «Secrets». Desde un worktree, el script de un repositorio se ejecuta por la misma ruta bajo el checkout principal (la carpeta que contiene `git rev-parse --path-format=absolute --git-common-dir`), con el worktree como directorio de trabajo; con una copia independiente se pregunta la ruta. Cuando el host rechaza escribir el script en el checkout principal, se escribe en la misma ruta del worktree y se mueve con los pendientes · falso hoy porque la guía no menciona scripts.
- AC9 — **Replaces:** AC4. La regla fija quién escribe un script local: el hilo principal, sin preguntar, solo en `flow-build` (incluidos los arreglos de `flow-verify`) o en una tarea operativa; nunca en research, status, plan, una revisión ni un hijo, que reporta lo que necesita. El script no guarda valores secretos, el informe final nombra cada script creado o cambiado, y un script que necesita todo clon queda como candidato al código versionado. «Support folder» y «Proportionality» lo exceptúan con un puntero · falso hoy porque «Support folder» manda los scripts a su ubicación de origen y «Proportionality» pide preguntar.
- AC10 — **Retired:** D7 A eliminó la escritura diferida, así que la limpieza de ramas no aplica nada pendiente. Texto anterior: `branch-cleanup.md`, después de salir del worktree propio del trabajo, aplica sus pendientes, borra `pending-local.md` y mueve sus scripts locales al checkout principal como fija «Local instructions», antes de las comprobaciones de borrado · falso hoy porque la limpieza no menciona escrituras pendientes.

- AC11 — **Retired:** reemplazado por AC15 (tercera revisión: H17, H18, H21). Texto anterior: «Local instructions» dice que, cuando el host rechaza desde un worktree la edición del archivo del checkout principal, la misma escritura se hace con un comando de shell, bajo los permisos del host: una entrada nueva se agrega con `>>` y una existente se cambia editando esa sola línea. Nunca se escribe una copia del archivo dentro del worktree · falso hoy porque el texto solo dice «in a worktree, read and write the main checkout's file».
- AC12 — **Replaces:** AC8. «Local instructions» permite que el `HIVE.local.md` de un proyecto (nunca el archivo de máquina) nombre un script local en `_support/scripts/<nombre>` bajo la raíz del archivo que lo nombra, fuera de Git como fija «Secrets». Desde un worktree, el script de un repositorio se ejecuta por la misma ruta bajo el checkout principal (la carpeta que contiene `git rev-parse --path-format=absolute --git-common-dir`), con el worktree como directorio de trabajo; con una copia independiente se pregunta la ruta. Cuando el host rechaza escribir el script en el checkout principal, se escribe ahí por shell · falso hoy porque la guía no menciona scripts.
- AC13 — **Replaces:** AC2. «Test accounts» remite a «Local instructions» para la escritura de `ACCOUNTS.md` que el host rechaza desde un worktree, sin repetir sus pasos · falso hoy porque «Test accounts» no lo menciona.
- AC14 — **Replaces:** AC6. Describen los scripts locales y la escritura por shell desde un worktree aislado, sin copiar la regla: `getting-started.md`; `flows/build.md` (lo que build registra incluye los scripts locales); `host-limits.md` (sección Claude Code: el aislamiento rechaza `Edit`/`Write` al checkout principal); y `workspace-and-artifacts.md`, cuya frase sobre «permanent maintenance scripts» admite la excepción · falso hoy porque ninguno lo menciona.

- AC15 — **Retired:** reemplazado por AC16 (cuarta revisión: H22, H24, H25, H28). Texto anterior: «Local instructions» dice que, cuando el host rechaza desde un worktree la edición del archivo del checkout principal, la misma escritura se hace con un comando de shell, bajo los permisos del host:
  - una entrada nueva se agrega con `>>`;
  - una línea existente se cambia con `sed -i.bak` y luego se borra el `.bak`, conservando el modo del archivo;
  - que Git lo ignore se comprueba desde el worktree con `git check-ignore` sobre la misma ruta relativa, y una entrada que falta se agrega al exclude de `--git-common-dir`;
  - si la escritura por shell también se rechaza, no se escribe nada y el informe final nombra la actualización;
  - nunca se escribe una copia del archivo dentro del worktree.

  Falso hoy porque el texto solo dice «in a worktree, read and write the main checkout's file».
- AC16 — **Replaces:** AC15. «Local instructions» dice que, cuando el host rechaza desde un worktree la edición del archivo del checkout principal, la misma escritura se hace con un comando de shell, bajo los permisos del host:
  - una entrada nueva se agrega con `>>`;
  - una línea existente se reescribe por su número con `sed -i.bak`, un `diff` contra el `.bak` confirma que solo cambió esa línea antes de borrar el `.bak`, y si falla se restaura;
  - si la escritura por shell también se rechaza, no se escribe nada y el informe final lo nombra;
  - nunca se escribe una copia dentro del worktree.

  Además, la viñeta de autorización admite que la comprobación del exclude corra desde el worktree sobre la misma ruta relativa; la viñeta 2 incluye los scripts locales entre lo que se escribe; y todo script nuevo se hace ejecutable. Falso hoy porque el texto solo dice «in a worktree, read and write the main checkout's file».

## Prerequisites

Ninguno: el build no necesita nada de fuera del repositorio.

## Verified context

- 2026-11-08, Claude Code 2.1.319, con un worktree creado con `EnterWorktree`:
  - Desde el worktree: `Read` del checkout principal funciona; `Edit` y `Write` se rechazan («Edit the worktree copy of this file instead of the shared-checkout path»); `>>` desde Bash funciona.
  - Un script del checkout principal se ejecuta con el worktree como directorio de trabajo: `--show-toplevel` devuelve el worktree y `dirname` de `--git-common-dir` devuelve el checkout principal.
  - Los ignorados (`HIVE.local.md`, `_support/workspace/`) no existen en el worktree.
  - Después de `ExitWorktree`, la misma sesión escribió con `Write` los archivos de esta carpeta de cambio en el checkout principal (H6).
- Documentación de Claude Code, «How Claude Code enforces isolation». La ruta prevista para cambiar un archivo del checkout principal es: «To change a file in the main checkout, edit it yourself or start an interactive session there.» Y sobre lo que el aislamiento no revisa: «a command that writes into the main checkout without running git there, such as `cp` or a shell redirect, isn't refused by them. Claude Code treats that command like any other shell command under your permissions and sandboxing settings.»
- La documentación de Claude Code (worktrees) confirma el bloqueo de `Edit`/`Write` y que un worktree solo trae lo versionado; `.worktreeinclude` copia ignorados, pero solo en los worktrees que crea Claude Code.
- `content/guidance/global.md` mide 41 843 bytes en `b14ae746`; ninguna prueba lo limita. `content/skills/flow-close/SKILL.md` mide 15 998 bytes (límite de 16 000, `tests/content/skill_size_test.go`).
- `git-workflow` paso 3: «Use the host's own worktree feature when it has one». Engram `hosts/grok-native-worktrees-specs` (issue #97): los worktrees de Grok son copias independientes.
- Casos que motivan el cambio: cam clona una base por worker de tests con `CREATE DATABASE … TEMPLATE` (`comsys/cam/cam-monorepo/apps/backend/test/global-setup-integration.ts`), y el workspace cam ya tiene `_support/scripts/`.

## Delivery

- **D4 B** — ¿Qué modo de entrega usa este cambio? `automatic`: tras la revisión, push, PR y merge en `development` sin parar a validar.
- **D5 A** — ¿Qué revisión de código corre sobre el candidato antes del push? Claude Code `/code-review`.

El refresco de la instalación local después del merge sigue la regla del `AGENTS.md` del repositorio. Esta respuesta no autoriza un release.

## Review status and progress

- 2026-11-08 — Plan escrito.
- 2026-11-08 — `hive-review-plan` encontró H2–H10. Se aplicaron H2, H4–H9; H3 se resolvió con D6 A; H10 queda listado. Se retiraron AC1, AC3 y AC4 y se agregaron AC7–AC10, con la nueva T4. 
- 2026-11-08 — Segunda revisión: H11–H16. H11–H14 dependían de diferir y desaparecen con D7 A. Se retiran AC2, AC6, AC7, AC8 y AC10, se agregan AC11–AC14 y se retira T4. Se aplicaron H15 (más páginas del sitio) y H16 (verificaciones concretas).
- 2026-11-08 — Tercera revisión, sobre D7: H17–H21, sin P0 ni P1. Se aplicaron:
  - H17: comprobación del exclude desde el worktree;
  - H19: la viñeta de scripts va después de la de autorización;
  - H20: verificaciones que hoy fallan;
  - H21: forma portable de editar una línea, conservar el modo y hacer ejecutable el script.

  H18 se resolvió como decisión menor: si la escritura por shell también se rechaza, no se escribe nada y el informe final lo nombra. Se retira AC11 y se agrega AC15.
- 2026-11-08 — Cuarta revisión: H22–H28, sin P0 ni P1. H23 corrigió la cita de la documentación y se resolvió con D8 A. Se aplicaron H22 (el exclude se comprueba desde el worktree dentro de la viñeta de autorización, con un delta de «Ignored before the first write»), H24 (`sed` por número de línea más `diff`), H25 (todo script nuevo es ejecutable) y H28 (la viñeta 2 incluye los scripts). Quedan listados: H26 (`check-ignore` usa el `.gitignore` de la rama del worktree) y H27 (`flow-close` y `branch-cleanup` mueven cosas al checkout principal desde un worktree, fuera de alcance). Se retira AC15 y se agrega AC16.
- 2026-11-08 — Quinta revisión: H29–H33, sin P0 ni P1. Se aplicaron H29–H32 en `design.md` y `tasks.md`; H33 queda listado. No cambian criterios ni deltas, así que no hace falta otra ronda.
