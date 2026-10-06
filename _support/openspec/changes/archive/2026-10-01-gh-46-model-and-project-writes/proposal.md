# Escrituras de las pantallas Models y Project

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado el 2026-10-01 en `development` por el [PR #89](https://github.com/JhonHawk/tricell-hive-private/pull/89) (`c0cbedb`) |
| Tracker · GitHub Issues | • [#46 — Instalador TUI: CLIs, voz y tono, estado, diagnóstico, modelos por rol, proyecto e integraciones](https://github.com/JhonHawk/tricell-hive-private/issues/46) (pantallas 5 y 6, escrituras) |
| Git | `interactive` a `development` (D4-A) · parada antes del push para el recorrido del usuario · `/code-review` antes del merge (D5-A) |
| Verificación | `go vet ./...` · `go test ./...` · `hive-review-ux` y `hive-verify-change` en `tmux` · recorrido del usuario |
| Siguiente paso | Ninguno en este cambio |

## Objetivo

La interfaz de terminal ya muestra qué modelo y esfuerzo recibe cada subagente en cada CLI, y si un repositorio tiene bien declarada su sección `## Hive`. Pero no deja cambiar ninguna de las dos cosas.
- **Modelos:** hoy cambiar el modelo de un rol exige editar `integrations/agent-profiles.json` en el checkout, hacer un commit y correr `hive update`. Eso no sirve en una instalación desde el paquete offline, que no tiene checkout. Además, el cambio vale para todos los roles del perfil.
- **`## Hive`:** hay que escribirla a mano, copiando el formato de la guía.

Este cambio agrega las dos escrituras:
- **Modelos:** elegir modelo y esfuerzo por rol y por CLI, guardados en el estado de Hive como la voz y vueltos a aplicar en cada `update` o vuelta a una release anterior.
- **Proyecto:** crear o editar la sección `## Hive` del `AGENTS.md` de un repositorio, con valores sugeridos desde git que el usuario confirma.

Cada escritura tiene su comando de texto equivalente, como exige la especificación de la interfaz.

## Decisiones del usuario (2026-10-01)

- **D1-A:** el ajuste de modelo vive en el estado de Hive, no en `agent-profiles.json`. No cambia el ID de la release y funciona sin checkout.
- **D2-A:** el ajuste es por rol y por CLI, con modelo y esfuerzo. El modelo es texto libre, porque no hay un catálogo fiable de modelos. El esfuerzo se elige de una lista fija.
- **D3-A:** la pantalla Project crea o edita `## Hive` y crea `AGENTS.md` si falta.
  - Sugiere valores desde git, muestra el cambio antes de escribir y conserva lo que no conoce.
  - Avisa, sin tocarlo, cuando un `CLAUDE.md` no importa `@AGENTS.md`.

## Ampliación durante la validación (2026-10-01)

Al probar la vista Models, el usuario pidió elegir el modelo de una lista en lugar de escribirlo, y agrupar los roles para ajustar todo un grupo a la vez. Decidió lo siguiente después de ver cómo lo hace optional reference project (`b388eb31`) y qué lista expone cada CLI:

- **D6-A:** el selector ofrece la lista del propio CLI.
  - Se consulta al abrir el panel y se guarda mientras dura la sesión.
  - Claude usa alias fijos, porque no tiene comando de listado.
  - Siempre queda «Other…» para escribir un id.
- **D7-A:** los roles se agrupan por la carpeta de función que ya existe en `content/agents/<grupo>/` (`design`, `development`, `docs`, `ops`, `quality`, `review`).
  - La cabecera del grupo ajusta todos sus roles a la vez.
  - Ajustar el grupo reemplaza los ajustes propios de sus roles; un rol se vuelve a personalizar después.
- **D8-A:** este trabajo entra en este mismo cambio, como T7 a T10, antes del recorrido del usuario.

## Entrega (D4-A y D5-A, decididas por el usuario el 2026-10-01)

- **Modo:** `interactive` a `development`.
  - Una rama de trabajo cortada de `development` (propuesta: `feat/gh-46-model-and-project-writes`), con worktrees por línea que se integran en ella.
  - Commits locales y verificación completa.
  - Parada antes del push, con el binario construido, para el recorrido del usuario (`tasks.md`, «Recorrido del usuario»).
- **Después de que el usuario valide:**
  1. Push de la rama.
  2. PR a `development`.
  3. `/code-review` de Claude Code sobre el diff completo contra la punta de la base, con sus hallazgos atendidos o refutados con evidencia.
  4. Merge por el agente con la suite local en verde, porque no hay CI.
- **Despliegue local:** tras el merge, por la regla temporal de `AGENTS.md`, se reconstruye el binario con `go build -o "$(command -v hive)" ./tooling/cli`. El cambio no toca `content/`, así que `hive update` no despliega nada nuevo, y el reporte lo dice.
- **Carpeta del cambio:** queda sin versionar mientras dura el trabajo. `flow-close` la archiva, con el delta integrado en `specs/versioned-installation/spec.md`, en un PR pequeño de cierre sin revisión dedicada, que fusiona el agente.
- **Tracker:** #46 se cierra al integrarse, porque estas son sus últimas pantallas pendientes.
- **Sin autorización todavía:** la implementación espera a que el usuario la pida.

## Alcance y aceptación

**Incluye:**
- El ajuste por rol en `tooling/management`: estado, plan, aplicación, recibos y modelo efectivo, y un parámetro nuevo en `agents.Resolve` y `agents.Render`.
- `hive models set` y `hive models reset`, y la edición en la vista Models.
- `hive project set` y la edición en la vista Project.
- La especificación `versioned-installation` y `deployment-manager.md`.

**Excluye:**
- **Cambiar los perfiles de la release:** `agent-profiles.json` sigue siendo el valor por defecto que se distribuye.
- **Configuración y cachés de los CLIs:** no se leen, porque #37 y el cambio anterior lo excluyeron. La lista de modelos de T7 sale de los comandos de listado de cada CLI (D6-A), no de sus cachés internos.
- **`## Hive` de un workspace:** solo se escribe la sección de la raíz del repositorio Git, como la valida hoy la vista.
- **`CLAUDE.md`:** no se crea ni se modifica; solo se avisa.
- **Journal y recuperación:** la escritura de `AGENTS.md` no entra en el journal de Hive ni en `hive recover`, porque el archivo es del proyecto y no gestionado. Se recupera con Git.
- **Alcance de proyecto y Windows:** los ajustes de modelo son de alcance `user`, y Windows queda fuera.

**Restricciones que deben seguir siendo ciertas:**
- `go vet ./...` y `go test ./...` pasan. `go test -race ./...` pasa también, porque el cambio toca el estado que comparten las pruebas de escritura concurrente (`deployment-manager.md`, «Verification»).
- **Sin ajustes, nada cambia:**
  - la salida de `agents.Render` es idéntica byte a byte a la de la base (golden de 20 roles × 6 CLIs);
  - `state.json` se codifica igual que antes;
  - `hive models` sin subcomando imprime las mismas filas de roles. T8 solo agrega las cabeceras de grupo (AC14).
- `hive doctor` y `hive models` sin subcomando siguen sin escribir nada.
- `tooling/management` y `tooling/distribution` siguen sin módulos externos (`deps_test.go`).
- El tema sigue cumpliendo la prueba de contraste WCAG AA.

**Criterios de aceptación.** Base: `d4c014d`.

- AC1. `hive models set --host <h> --role <r> [--model M] [--effort E]` muestra un resumen y pide confirmación, o acepta `--dry-run` y `--out` como `hive voice set`. Al aplicar:
  - reescribe solo el archivo del rol en ese CLI con el modelo y el esfuerzo nuevos;
  - deja idénticos byte a byte los demás archivos de agente y los demás CLIs;
  - guarda el ajuste en el estado;
  - `hive models` muestra los valores nuevos con la marca `*` y la nota «* set with hive models set».

  *Falso en la base cuando* `hive models set` termina en error por subcomando o argumento desconocido.
- AC2. Un ajuste sobrevive a `hive update` (release nueva) y a `hive plan install --release <id anterior>` + `hive apply`: el archivo del rol sigue con el valor ajustado.
  - Después de `models set` y de `update`, `hive status` da la instalación de ese CLI como `verified`, no `partial` ni `drift`.
  - Después de la vuelta atrás, da el mismo estado de versión que la misma vuelta atrás sin ajustes. Hoy es `legacy`, porque esa reinstalación no lleva versión de producto (`versioning.go:46-48`).

  *Falso en la base cuando* no existe ajuste que sobreviva.
- AC3. `hive models reset --host <h> --role <r> --only model|effort` quita solo esa parte del ajuste. Después de `hive models reset --host <h> --role <r>` o `--all`:
  - el archivo del rol vuelve a ser idéntico byte a byte al que instala una release sin ajustes;
  - el ajuste sale del estado.

  Quitar Hive de un CLI con `plan remove` borra sus ajustes del estado. *Falso en la base cuando* `hive models reset` no existe.
- AC4. Estos ajustes se rechazan sin escribir nada y con un mensaje que nombra el problema:
  - un esfuerzo en Grok o Cursor;
  - un esfuerzo fuera de `low`, `medium`, `high`, `xhigh`, `max` y `ultra`, la lista que el código ya acepta en los perfiles (`agents.go:161`);
  - un modelo vacío, de más de 200 caracteres, o con un carácter fuera de `[A-Za-z0-9._:/@+=\[\]-]`, lo que también excluye `#`, espacios, comillas, caracteres de control e invisibles;
  - un rol que no está en la release instalada en ese CLI;
  - un CLI no registrado;
  - un esfuerzo en OpenCode sin un modelo que lo lleve;
  - un CLI cuya reinstalación cambiaría algo más que archivos de agente ajustados: recursos borrados a mano o una migración heredada pendiente. El mensaje remite a `hive install` o `hive doctor`.

  `hive apply` rechaza un plan guardado con `--out` y editado a mano, con su ID recalculado, que meta un ajuste inválido, un ajuste distinto del estado para un CLI fuera del plan o un CLI desconocido. *Falso en la base cuando* no existe validación de ajustes.
- AC5. En la vista Models:
  - ↑↓ mueve un cursor `>` por los roles.
  - Enter abre la edición del rol: un campo de texto con el modelo efectivo y el esfuerzo, que se cambia con ←→. En Grok y Cursor el esfuerzo se muestra como «not supported by <host>» y no se edita.
  - Enter en la edición muestra una confirmación con el CLI, el rol, el modelo y el esfuerzo antes → después, y la carpeta de los archivos que escribe.
  - Apply escribe lo mismo que AC1, y Cancel o Esc no cambian nada.
  - `x` sobre un rol ajustado pide confirmación y hace lo mismo que `reset`.
  - Los roles ajustados llevan la marca `*`.
  - Las dos líneas del pie dicen «Enter edits the selected role; x resets one marked *.» y «Defaults come from integrations/agent-profiles.json in the release.»
  - Cambiar solo el esfuerzo en el panel deja el mismo estado que `hive models set --effort` sin `--model`.

  *Falso en la base cuando* la vista no tiene cursor ni edición.
- AC6. `hive project set [--project DIR] --set 'Key: value'… [--unset Key…] [--dry-run]` muestra el resumen y pide confirmación; sin terminal solo acepta `--dry-run`. Al aplicarse sobre un `AGENTS.md` con la sección:
  - reemplaza solo las líneas `- Key:` de las claves dadas;
  - agrega las claves nuevas después del último elemento;
  - conserva sin cambios las claves desconocidas, los comentarios y las líneas en blanco o de prosa de la sección, y todo el texto fuera de ella;
  - conserva los finales de línea CRLF y los permisos del archivo.

  `--unset` solo acepta claves opcionales. *Falso en la base cuando* `hive project` es un comando desconocido.
- AC7. Si no hay sección, `hive project set` la agrega al final del archivo, separada por una línea en blanco. Si no hay `AGENTS.md`, lo crea con solo la sección y permisos `0644`. Si en la raíz hay un `CLAUDE.md` sin una línea `@AGENTS.md`:
  - la salida avisa que Claude Code no leerá la sección;
  - `CLAUDE.md` queda idéntico.

  *Falso en la base cuando* ningún comando escribe `AGENTS.md`.
- AC8. Estos casos se rechazan sin escribir nada:
  - una sección duplicada;
  - un `AGENTS.md` que es un enlace simbólico (symlink) o no es un archivo regular;
  - un archivo de más de 1 MiB;
  - fuera de un repositorio Git;
  - un valor obligatorio que falta o queda vacío después del cambio;
  - un valor con saltos de línea, caracteres de control o caracteres que `sanitizeLine` quitaría al mostrarlo;
  - `Delivery` distinto de `direct-base` o `Hive guidance` distinto de `required`;
  - un `Base branch` que no pasa `git check-ref-format --branch`;
  - una clave que `hiveSettingKeys` no conoce en `--set`;
  - un `AGENTS.md` que cambió entre la vista previa y la escritura.

  Una rama o una ruta de `Specs` que todavía no existen no bloquean: se avisan en la vista previa. *Falso en la base cuando* no hay escritor.
- AC9. En la vista Project, `e` abre un formulario con las ocho claves de `hiveSettingKeys`, las obligatorias primero:
  - **Valores iniciales:** cada campo trae el valor actual. Si no hay valor, trae una sugerencia marcada «suggested»:
    - `Project`: el nombre del repositorio en el remoto `origin`, o el del directorio;
    - `Base branch`: la rama de `origin/HEAD`;
    - `Tracker`: `GitHub Issues · <owner>/<repo>` para un remoto de `github.com`;
    - `Specs`: `_support/openspec` cuando existe.
  - **`Delivery` y `Hive guidance`:** se eligen con ←→ entre vacío y su único valor.
  - **Guardar:** con un obligatorio vacío, Enter no avanza y dice cuál falta. Si no, Enter muestra una confirmación con las líneas de la sección antes y después, los avisos de AC7 y AC8, y «AGENTS.md has uncommitted changes» cuando `git status --porcelain` lo lista.
  - **Al aplicar:** escribe lo mismo que `hive project set`, y la vista vuelve a validar y muestra «Valid».

  *Falso en la base cuando* la vista no tiene edición.
- AC10. A 80×24 y a 120×40, ninguna línea supera el ancho en estas pantallas: la edición de Models, la lista de modelos abierta, una cabecera «mixed», las confirmaciones de rol y de grupo, el formulario de Project y su confirmación. Con 20 roles se llega al último con ↑↓. *Falso en la base cuando* estas pantallas no existen.
- AC11. La documentación y la especificación reflejan el cambio:
  - `deployment-manager.md` documenta `hive models set|reset` y `hive project set`, quita la frase «the diagnostics never write models or `## Hive` values» y explica sus límites: el comando acepta ids que el CLI no lista, y `AGENTS.md` va sin journal.
  - La especificación `versioned-installation` tiene los requisitos de este cambio.

  *Falso en la base cuando* `rg -n "models set" _support/docs/architecture/deployment-manager.md` no devuelve líneas.
- AC12. En el panel de Models, el campo Model es una lista que se puede filtrar escribiendo. Contiene:
  - **El valor efectivo actual,** siempre, aunque el CLI no lo liste.
  - **Los modelos del propio CLI,** según cada uno:
    - Codex: los `slug` con `visibility` igual a `list` de `codex debug models`;
    - OpenCode: `opencode models`, con un reintento si la salida sale vacía;
    - Pi: `pi --list-models`, como `provider/model`;
    - Grok: `grok models`;
    - Cursor: `cursor-agent models`;
    - Claude: los alias fijos `fable`, `opus`, `sonnet` y `haiku`.
  - **«Other…»,** que abre el campo de texto.

  En Codex, la lista de esfuerzos del modelo elegido es la de sus `supported_reasoning_levels`, recortada a los valores que acepta `ValidateOverride`. Si el CLI falla, tarda más de 8 s o devuelve algo ilegible, la lista muestra solo el valor actual y «Other…», más una línea «Model list unavailable: <motivo>». Con `--home` no se ejecuta ningún CLI. Los ids que no cumplen la lista de caracteres de AC4 no se ofrecen. *Falso en la base cuando* el campo Model es texto libre.
- AC13. La consulta de la lista cumple estas condiciones:
  - ejecuta solo el binario del CLI detectado en `PATH`, con argumentos fijos, sin shell, con límite de 8 s, `WaitDelay` de 1 s y una salida acotada a 8 MiB;
  - se ejecuta una vez por CLI y por sesión de la interfaz, al abrir el panel de ese CLI;
  - `hive models`, `hive doctor` y las vistas que no editan no la ejecutan nunca.

  *Falso en la base cuando* ningún código ejecuta un comando de listado.
- AC14. La vista Models agrupa los roles de cada CLI por su carpeta (`content/agents/<grupo>/`), con una cabecera por grupo en orden alfabético. La cabecera muestra el modelo y el esfuerzo comunes del grupo, o «mixed» si difieren. El cursor puede caer en las cabeceras. `hive models` imprime los mismos grupos. *Falso en la base cuando* la tabla es una lista plana de roles.
- AC15. Para editar un grupo hay dos caminos equivalentes:
  - en la interfaz, Enter sobre la cabecera abre el mismo panel para el grupo;
  - en texto, `hive models set --host H --group G [--model M] [--effort E]`.

  Aplicar el cambio deja en cada rol del grupo un ajuste con solo las partes dadas. Reemplaza el ajuste propio que tuviera, y la confirmación nombra cada rol cuyo ajuste propio se pierde. `x` sobre la cabecera, o `hive models reset --host H --group G [--only model|effort]`, quita los ajustes de todos los roles del grupo. `--group` y `--role` son excluyentes, y un grupo desconocido se rechaza sin escribir. *Falso en la base cuando* `--group` es una opción desconocida.
- AC16. `deployment-manager.md` y la especificación documentan la lista de modelos, sus fuentes y límites por CLI, la ejecución acotada de esos comandos y la edición por grupo. *Falso en la base cuando* `rg -n "debug models|--group" _support/docs/architecture/deployment-manager.md` no devuelve líneas.
