# Actualizar desde un commit y listar releases: `hive update` y `hive releases`

| Campo | Valor actual |
| --- | --- |
| Estado | En validación · T1–T5 verificadas · falta `/code-review` antes del push |
| Tracker · GitHub Issues | • [#46 — Instalador TUI: CLIs, voz y tono, estado, diagnóstico, modelos por rol, proyecto e integraciones](https://github.com/JhonHawk/tricell-hive/issues/46) (primera entrega, sin la TUI)<br>• [#42 — Diferido: extraer el despacho de subcomandos de los run del CLI y del empaquetador](https://github.com/JhonHawk/tricell-hive/issues/42) (solo el `run` del CLI) |
| Git | `direct-base` a `rebuild/harness-engineering` · sin PR · `/code-review` antes del push · sin despliegue |
| Verificación | `go vet ./...` · `go test -race ./...` · prueba local de `update --dry-run` y `releases` sobre el estado real, solo lectura |
| Siguiente paso | `/code-review` del diff; después rebase y push directo a `rebuild/harness-engineering` |

## Objetivo

Hoy, para desplegar Hive desde este repositorio sin arrastrar los cambios sin commit de otra sesión, hay que hacer a mano tres pasos: un `git archive` de `HEAD` en una carpeta temporal, un `plan install --source <carpeta> --hosts <los seis> --scope user --out <plan>` y un `apply`. El 2026-09-27 se hizo unas 12 veces. Además, el gestor guarda cada release instalada (116 en el equipo del usuario), pero ningún comando las lista, y ninguna guarda de qué commit salió. Para volver atrás hay que conocer el hash.

Este cambio agrega dos comandos al gestor. `hive update` saca los archivos de un commit (por defecto `HEAD`), arma el plan para los CLIs que ya tienen Hive instalado en alcance de usuario, muestra el resumen y aplica tras confirmación. Sin terminal, deja el plan en un archivo para `hive apply`. `hive releases` lista las releases guardadas con su fecha, su commit cuando se conoce y los CLIs donde está instalada cada una. Es la primera entrega de #46: la TUI llegará después como capa sobre estos comandos, con `charmbracelet/huh`, cuando exista la voz (decisión D4-A del 2026-09-27).

## Alcance y aceptación

**Incluye:**
- `hive update` en `tooling/cli`, con las opciones `--rev`, `--source`, `--home`, `--state-dir`, `--dry-run` y `--out`.
- El commit de origen en el plan y un registro aparte, junto a cada release, de los commits de los que salió.
- `hive releases` en `tooling/cli` y la función de listado en `tooling/management`.
- La extracción del despacho de subcomandos de `run` en `tooling/cli/main.go` (#42), porque este cambio lo toca.
- La documentación en `_support/docs/architecture/deployment-manager.md` y el cambio en la especificación `versioned-installation`.

**Excluye:**
- La TUI y cualquier dependencia nueva. `huh` llegará con la TUI (D4-A).
- Las pantallas 2, 4, 5 y 6 de #46 (voz, diagnóstico, modelos por rol y sección `## Hive`) y la pantalla 7 (integraciones). Siguen dentro de #46 (D3-B).
- El `run` de `tooling/package/main.go`, `Engine.Recover` y `BuildPlan`: la condición de #42 solo se activa para el código que se toca.
- Instalar `pi-subagents` (#29).
- Cambiar el paquete offline: `install.sh` sigue sin necesitar Git. `hive update` solo sirve desde un checkout de Git.
- Desplegar en la configuración global real del usuario. Es un efecto aparte que necesita su propia autorización.

**Restricciones que deben seguir siendo ciertas:**
- `go vet ./...`, `go test ./...` y `go test -race ./...` pasan.
- Los planes guardados sin commit conservan su ID: el campo nuevo se omite cuando está vacío.
- Los snapshots de release no cambian de bytes, así que el ID de las releases existentes no cambia.
- `go.mod` sigue sin dependencias externas.

**Criterios de aceptación:**
- AC1. En una terminal interactiva, sobre un home sintético con Hive instalado para dos CLIs y un repositorio Git sintético con un cambio en `content/` con commit y otro sin commit, `hive update` muestra el resumen con el commit y esos dos CLIs, aplica tras `y`, y los archivos instalados tienen el contenido del commit y no el cambio sin commit. *Falso en la base cuando* `go run ./tooling/cli update` devuelve `unknown command "update"`.
- AC2. `hive update --dry-run` muestra el mismo resumen y no cambia el estado ni los destinos, también sin terminal. *Falso en la base cuando* el comando no existe (mismo error que AC1).
- AC3. Cuando el plan cambia archivos, sin terminal y sin `--dry-run`, `hive update` falla sin cambiar nada y el mensaje nombra `--out`. Cuando el plan no cambia nada, el diseño lo aplica sin preguntar, y su única escritura es el registro del commit. Con `--out FILE` escribe un plan que `hive apply --plan FILE` aplica, y ese plan lleva el commit de origen completo. *Falso en la base cuando* el comando no existe.
- AC4. `--rev` elige el commit. Un valor que empieza por `-`, un commit que no existe, un `--source` que no es un checkout de Git o la falta de `git` en el `PATH` fallan con un error claro antes de escribir nada. *Falso en la base cuando* el comando no existe.
- AC5. Después de un `hive update`, la carpeta temporal donde se extrajo el commit (`.hive-extract-*`) ya no existe, tanto si el comando termina bien como si falla después de extraer. *Falso en la base cuando* el comando no existe.
- AC6. Aplicar un plan con commit de origen registra ese commit junto a la release (`releases/<id>.commits.json`), también cuando el plan no cambia ningún archivo. Aplicar un plan sin commit no crea ese registro. Dos commits con el mismo contenido quedan listados una vez cada uno. *Falso en la base cuando* `rg -n "commits.json" tooling/` no devuelve nada.
- AC7. `hive releases` lista cada snapshot de `releases/`, del más reciente al más antiguo, con su ID, la fecha de su última escritura, sus commits (vacío si no hay registro) y los CLIs donde está instalada, también después de volver atrás con `--release`. Sobre el estado real del usuario muestra tantas releases como snapshots hay y no modifica nada. *Falso en la base cuando* `go run ./tooling/cli releases` devuelve `unknown command "releases"`.
- AC8. `run` en `tooling/cli/main.go` solo resuelve el subcomando y delega en una función por subcomando, y no pasa de 40 líneas. *Falso en la base cuando* `run` mide 149 líneas.
- AC9. `deployment-manager.md` documenta `hive update` y `hive releases`, incluido que `update` necesita Git y un checkout, y que el paquete offline no. *Falso en la base cuando* `rg -n "hive update|hive releases" _support/docs/architecture/deployment-manager.md` no devuelve nada.

## Entrega

Decisiones del usuario del 2026-09-27:

- **Modo (D7-A):** `direct-base` sobre `rebuild/harness-engineering`, sin PR, como en #43.
  - La carpeta del cambio se versiona al aprobar el plan, en un commit propio en esa misma rama.
  - El código va en commits locales verificados. Después vienen la revisión de código y el push directo a la rama.
  - El cierre archiva la carpeta del cambio en otro commit y mueve #46 y #42 según corresponda: #42 se cierra; #46 sigue abierto, porque esta es solo su primera entrega.
- **Revisión de código (D8-A):** `/code-review` de Claude Code sobre el diff local, antes del push.
  - Los hallazgos que bloquean se corrigen o se refutan con evidencia antes de publicar.
  - La elección no autoriza por sí sola a ejecutarla: se lanza al terminar T5.
- **Sin despliegue:** el despliegue en los seis CLIs del usuario no está incluido. Se ofrece aparte al terminar.
