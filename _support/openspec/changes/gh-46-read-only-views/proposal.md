# Vistas de solo lectura del gestor: diagnóstico, modelos, integraciones y proyecto

| Campo | Valor actual |
| --- | --- |
| Estado | En curso · implementación autorizada por el usuario el 2026-09-29 |
| Tracker · GitHub Issues | • [#46 — Instalador TUI: CLIs, voz y tono, estado, diagnóstico, modelos por rol, proyecto e integraciones](https://github.com/JhonHawk/tricell-hive/issues/46) (pantallas 4 a 7, solo lectura) |
| Git | `direct-base` a `rebuild/harness-engineering` desde el worktree (D5-B) · recorrido del usuario y `/code-review` antes del push (D6-A) |
| Verificación | `go vet ./...` · `go test -race -timeout 20m ./...` · `hive-review-ux` y `hive-verify-change` en `tmux` · recorrido del usuario |
| Siguiente paso | T1 y T4 en paralelo |

## Objetivo

La interfaz de terminal ya instala, quita, actualiza, vuelve a una release y configura la voz. Pero no responde tres preguntas: si la instalación de Hive está sana, qué modelo usa cada subagente en cada CLI y si un proyecto tiene bien declarada su sección `## Hive`. Hoy la respuesta está repartida entre `hive status`, `hive setup`, archivos internos de cada CLI y la lectura a mano de `AGENTS.md`.

Este cambio agrega cuatro vistas que solo leen: Diagnostics, Models, Integrations y Project. También agrega dos comandos de texto equivalentes, `hive doctor` y `hive models`. Ninguna vista escribe archivos ni estado. Cambiar modelos y crear o editar `## Hive` queda para un segundo cambio (D1-A).

## Decisiones del usuario (2026-09-29)

- **D1-A:** primero solo lectura. Las escrituras de las pantallas 5 y 6 irán en otro cambio.
- **D2-A:** la pantalla Models solo muestra el modelo y el esfuerzo efectivos. No hay ajustes personales en este cambio.
- **D3-A:** sesiones abiertas.
  - Claude Code y Grok: se compara el inicio de cada sesión viva con la hora en que se escribió la release instalada.
  - Codex, Pi y Cursor: un aviso de reiniciar.
  - OpenCode: se dice que recarga las instrucciones.
- **D4-A:** la pantalla Project solo valida `## Hive` y muestra qué falta.

## Alcance y aceptación

**Incluye:**
- Las cuatro vistas y sus entradas en el menú.
- `hive doctor` y `hive models`, con el mismo contenido que las vistas.
- Un lector del último registro de integraciones (`onboarding/<id>.json`).
- La resolución del modelo efectivo por rol, sin duplicar la lógica de `agents.Render`.
- La especificación `versioned-installation` y `deployment-manager.md`.

**Excluye:**
- Escribir modelos, `agent-profiles.json` o `AGENTS.md` de proyectos (D1-A, D2-A, D4-A).
- Leer la configuración de terceros: `config.toml`, registros MCP y ajustes de hosts. #37 lo excluyó.
- Ejecutar programas de integraciones (Engram, `agent-browser`, `npx`). Solo se ejecuta `<cli> --version` de los seis CLIs.
- Actualizar `agent-delivery.md` (hallazgo aparte de la investigación, pendiente de ticket).
- Validar la sección `## Hive` del `AGENTS.md` de un workspace. La guía la permite, pero Project solo valida la raíz del repositorio Git (D4-A). Fuera de un repositorio, la vista lo dice.
- Windows.

**Restricciones que deben seguir siendo ciertas:**
- `go vet ./...` y `go test -race -timeout 20m ./...` pasan.
- La salida de `hive status`, `install`, `update`, `releases`, `voice` y `setup` no cambia.
- `tooling/management` y `tooling/distribution` siguen sin módulos externos (`deps_test.go`).
- Abrir cualquiera de las vistas nuevas, o correr `hive doctor` o `hive models`, deja `state.json`, el home de prueba y el proyecto idénticos.
- El tema sigue cumpliendo la prueba de contraste WCAG AA.

**Criterios de aceptación.** Base: `82f2dcc`.

- AC1. El menú muestra, en este orden, CLIs, Update, Releases, Voice, Diagnostics, Models, Integrations, Project y Quit. Enter abre cada vista nueva, y Esc o Backspace vuelven al menú. *Falso en la base cuando* el menú tiene cinco entradas.
- AC2. La sección de CLIs de Diagnostics muestra una fila por CLI de `installerHosts` con estos datos:
  - detectado o no;
  - la primera línea de `<binario> --version`, con `cursor-agent` para Cursor;
  - la release corta y el estado de la instalación (`verified`, `partial`, `drift`, `legacy`).

  Un CLI no detectado no se ejecuta. Un `--version` que falla o tarda más de 3 s muestra «CLI version unavailable» y el motivo. La versión del CLI y la release de Hive llevan etiquetas distintas («CLI version» y «Hive release») (D1-B, M6). *Falso en la base cuando* ninguna salida del gestor muestra la versión de un CLI.
- AC3. La sección Installation muestra cada fila de `management.Status` que no está en `installed`, `retained_shared` ni `not_installed`, con su ruta y una frase clara. Cuando hay una operación pendiente, las filas `recovery_required` se resumen en una sola línea con `hive recover`. Por ejemplo, `unowned_or_conflicting` dice «Hive markers are missing, duplicated or broken». También muestra una operación pendiente con el comando `hive recover`. Con una fila `drift`, una línea explica que Hive no repara un archivo cambiado y qué puede hacer el usuario (D1-B, M4). Sin problemas, dice «No problems found». *Falso en la base cuando* no existe una vista ni un comando que traduzca los estados de `Status`.
- AC4. La sección Sessions:
  - **Claude Code:** lee `sessions/*.json` de su directorio de configuración.
  - **Grok:** lee `active_sessions.json` de su home.
  - **Solo sesiones vivas:** en ambos casos considera solo las sesiones cuyo `pid` sigue vivo.
  - **Resumen:** con Claude Code o Grok registrado, la primera línea dice cuántas sesiones hay que reiniciar, o que ninguna, y qué CLI no se pudo comprobar (D1-B, M4).
  - **Marca de desactualizada:** marca «started before the installed release; restart it» cuando la sesión empezó antes del `LastWrittenAt` de la release instalada para ese CLI.
  - **Formato no reconocido:** un archivo ilegible, que pasa de 1 MiB o tiene un formato desconocido da «Session check unavailable for <host>: <motivo>», sin fallar la vista.
  - **Codex, Pi y Cursor registrados:** un aviso de reiniciar tras cada actualización.
  - **OpenCode:** una nota de que recarga en el siguiente mensaje.

  *Falso en la base cuando* nada en `tooling/` lee esos archivos.
- AC5. Models muestra, para cada CLI registrado, una fila por rol de la release instalada en ese CLI: rol, perfil, modelo y esfuerzo.
  - Los valores coinciden con los campos que `agents.Render` escribe en el archivo del agente, incluido el esfuerzo propio del rol en Claude, Codex y Pi.
  - Sin modelo, la fila dice «host default». Un `inherit` literal dice «inherit (parent session)».
  - En OpenCode, el sufijo `#variant` del modelo se muestra en la columna de esfuerzo.
  - La vista dice cómo cambiarlos: editar `integrations/agent-profiles.json` y correr `hive update`.

  *Falso en la base cuando* ninguna salida muestra el modelo efectivo por rol.
- AC6. Integrations muestra filas para Engram, Context7, pi-subagents y `agent-browser` con cuatro datos:
  - la evidencia local: el ejecutable en `PATH` o el archivo de skill con su ruta;
  - el estado del último registro de integraciones, cuando lo hay;
  - la fuente oficial;
  - un siguiente paso concreto para su caso; el único comando que sugiere es el de Context7 (D1-B, M1).

  Ningún programa de integraciones se ejecuta y no se lee configuración de hosts. *Falso en la base cuando* solo existe `hive setup`, que únicamente revisa Context7.
- AC7. Project valida la sección `## Hive` del `AGENTS.md` en la raíz del repositorio Git que contiene el directorio actual, o `--project DIR` en `hive doctor`. Reporta cada uno de estos casos con una frase:
  - fuera de un repositorio, con una segunda línea que dice que las secciones de workspace no se validan;
  - sin `AGENTS.md`;
  - sin sección o con la sección duplicada;
  - falta un valor obligatorio (`Project`, `Base branch`, `Tracker`, `Specs`) o está vacío;
  - la ruta de `Specs` no existe como directorio;
  - `Base branch` no existe como rama local ni en `origin`;
  - una clave desconocida;
  - `Delivery` distinto de `direct-base`;
  - `Hive guidance` distinto de `required`.

  Una sección válida dice «Valid» y lista los valores. *Falso en la base cuando* `rg -n '## Hive' tooling integrations` no encuentra código que la lea.
- AC8. `hive doctor [--home DIR] [--state-dir DIR] [--project DIR]` imprime las secciones CLIs, Installation, Sessions, Integrations y Project con el mismo contenido que las vistas. `hive models [--home DIR] [--state-dir DIR]` imprime las filas de Models. Ambos terminan con 0 aunque haya hallazgos, y con distinto de 0 solo ante un error propio. Sin terminal, el mensaje de `hive tui` nombra los dos comandos. *Falso en la base cuando* `hive doctor` falla con «unknown command».
- AC9. A 80×24 y a 120×40, ninguna línea de las vistas nuevas supera el ancho, y el contenido más largo que la pantalla se desplaza con PgUp y PgDn, y también con ↑↓ salvo en Integrations, donde ↑↓ mueve el cursor. Con 6 CLIs y 20 roles, se llega a la última fila. *Falso en la base cuando* las vistas no existen.
- AC10. Una prueba lee la sección «Project settings» de `content/guidance/global.md` y falla en dos casos:
  - una clave de `hiveSettingKeys` no es el comienzo de ningún texto entre comillas invertidas de esa sección (por ejemplo, `Delivery` aparece como `` `Delivery: direct-base` ``);
  - una línea `- Clave: valor` del bloque de código de ejemplo de esa sección usa una clave que `hiveSettingKeys` no conoce. Solo se leen las líneas dentro del bloque de código, no la lista de la sección. *Falso en la base cuando* no hay prueba que las compare.
- AC11. La documentación y la especificación reflejan el cambio:
  - `deployment-manager.md` documenta las cuatro vistas, `hive doctor` y `hive models`, con sus límites: formatos internos no documentados, y Codex, Pi y Cursor sin detección.
  - La especificación `versioned-installation` tiene el requisito nuevo.

  *Falso en la base cuando* `rg -n "hive doctor" _support/docs/architecture/deployment-manager.md` no devuelve líneas.

## Entrega (D5-B y D6-A, decididas por el usuario el 2026-09-29)

- **Modo:** `direct-base` a `rebuild/harness-engineering`, sin PR.
  - Se trabaja en el worktree `.claude/worktrees/gh-46-read-only-views`, con la rama local `feat/gh-46-read-only-views`, cortada de la punta de la base.
  - Los worktrees de las tareas en paralelo se integran en esa rama.
  - Antes del push, la rama se rebasa sobre la punta de la base.
- **Orden:**
  1. Verificación por tarea.
  2. `hive-review-ux` y `hive-verify-change` en paralelo.
  3. El recorrido del usuario.
  4. `/code-review` de Claude Code sobre el diff completo contra la base.
  5. El push a la base.
- **Carpeta del cambio:** viaja en los commits de la rama. Se versiona con el primer commit y se archiva en el último, con la especificación integrada en `specs/versioned-installation`.
- **Sin CI:** el repositorio no tiene workflows, así que la suite local es la única verificación automática.
- **Despliegue:** por la regla temporal de `AGENTS.md`, tras el push se reconstruye el binario con `go build -o "$(command -v hive)" ./tooling/cli`. Como el cambio no toca `content/`, `hive update` no despliega nada nuevo, y se dice así en el reporte.
- **Sin autorización todavía:** la implementación espera a que el usuario diga que se construya.
