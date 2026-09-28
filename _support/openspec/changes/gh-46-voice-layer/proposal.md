# Capa opcional de voz y tono: `hive voice`

| Campo | Valor actual |
| --- | --- |
| Estado | En curso · plan aprobado el 2026-09-27 · T1 y T3 en marcha |
| Tracker · GitHub Issues | • [#46 — Instalador TUI: CLIs, voz y tono, estado, diagnóstico, modelos por rol, proyecto e integraciones](https://github.com/JhonHawk/tricell-hive/issues/46) (pantalla 2, sin TUI) |
| Git | `direct-base` a `rebuild/harness-engineering` · `/code-review` antes del push · sin despliegue (decisiones de la sesión D7-A y D8-A) |
| Verificación | `go vet ./...` · `go test -race ./...` · prueba local sobre un home de prueba · validación humana del texto de las tres voces |
| Siguiente paso | `flow-build` sobre esta carpeta, empezando por T1 y T3 en paralelo |

## Objetivo

#46 propone una capa de voz y tono opcional, apagada por defecto, encima de las reglas de comunicación de Hive y sin cambiarlas. La TUI con `charmbracelet/huh` se construirá cuando esta capa exista (decisión D4-A del 2026-09-27). Este cambio construye la capa en el gestor, con comandos normales, para que la TUI después solo la llame.

La voz vive en un bloque gestionado propio, separado del bloque de Hive, en el mismo archivo de instrucciones de cada CLI (D9-A). El gestor la activa, la cambia y la quita con su mecanismo de plan, aplicación y recuperación. El ID de las releases no cambia, porque la voz es una preferencia del usuario y no parte de la guía del producto. La primera entrega trae tres voces: Jarvis, Senior directo y Mentor (D10-A).

## Alcance y aceptación

**Incluye:**
- Los textos de tres voces en `content/voices/`, en inglés, que describen un estilo sin copiar diálogos de películas.
- El bloque de voz en el gestor (`tooling/management`): su modelo, su plan, su aplicación con journal y recuperación, su estado, `status` y `plan remove`.
- `hive voice list`, `hive voice set` y `hive voice off` en `tooling/cli`.
- Que `hive update` y `plan install` conserven la voz y la regeneren cuando cambie su texto (D12-A).
- La documentación en `deployment-manager.md` y el cambio en la especificación `versioned-installation`.

**Excluye:**
- La TUI.
- Las otras cinco voces: Alfred, TARS, Data, C-3PO y Marvin. Se suman después como texto.
- Una voz distinta por CLI. Hay una sola voz por home, aplicada a todos los CLIs registrados en alcance de usuario.
- El alcance de proyecto.
- Los estilos de salida de Claude Code, que solo existen en ese host y reemplazarían el estilo «Concise» que el usuario ya usa.
- Desplegar la voz en la configuración real del usuario, que necesita autorización aparte.

**Restricciones que deben seguir siendo ciertas:**
- `go vet ./...`, `go test ./...` y `go test -race ./...` pasan.
- El ID de las releases no depende de la voz.
- Los bytes fuera de los dos bloques gestionados se conservan.
- `go.mod` sigue sin dependencias externas.
- Las reglas de comunicación de `content/guidance/global.md` no cambian.

**Criterios de aceptación:**
- AC1. `hive voice list` imprime las tres voces con su ID y una línea de descripción. *Falso en la base cuando* `go run ./tooling/cli voice list` devuelve `unknown command "voice"`.
- AC2. En un home de prueba con Hive instalado para Codex y Claude, `hive voice set jarvis --address sir --intensity subtle` pide confirmación en una terminal. Después:
  - cada archivo de instrucciones de esos CLIs tiene un solo bloque de voz, justo después del bloque de Hive, con el texto de la voz y los parámetros elegidos;
  - los bytes fuera de los dos bloques no cambian;
  - el ID de la release instalada no cambia.

  *Falso en la base cuando* el comando no existe.
- AC3. `hive voice off` deja cada archivo byte a byte como estaba antes de `voice set`. *Falso en la base cuando* el comando no existe.
- AC4. Con una voz activa, `plan install` de una release nueva conserva el bloque de voz. Si el texto de esa voz cambió en la fuente, el plan lo regenera con la misma elección y el resumen de `hive update` lo muestra. Si no cambió, no lo toca. *Falso en la base cuando* `rg -n "Voice" tooling/management/` no devuelve nada.
- AC5. `plan remove` de un CLI quita también su bloque de voz. En un archivo compartido por varios CLIs, como Grok y Claude en `~/.claude/CLAUDE.md`, se quita solo con el último consumidor, igual que el bloque de Hive. *Falso en la base cuando* no existe bloque de voz.
- AC6. `hive status` agrega una fila por bloque de voz, con `Kind` `voice` y un campo `Voice` con el ID, el tratamiento y la intensidad, y marca como `drift` un bloque de voz editado a mano. *Falso en la base cuando* `status` no menciona la voz.
- AC7. Si el usuario editó a mano el bloque de voz, `voice set`, `voice off` y `update` presentan un conflicto y conservan la edición. *Falso en la base cuando* no existe bloque de voz.
- AC8. Una operación de voz interrumpida en cualquier punto de escritura se recupera con `hive recover`: el archivo vuelve a un estado coherente con `state.json` y conserva el texto fuera de los bloques. *Falso en la base cuando* no existe bloque de voz.
- AC9. `voice set` falla antes de escribir con una voz desconocida, con `--address name` sin `--name`, o sin Hive instalado. El catálogo rechaza un texto de voz que contenga cualquiera de los marcadores. *Falso en la base cuando* el comando no existe.
- AC10. El bloque generado de cada voz empieza con el preámbulo común. Ese preámbulo dice:
  - que nunca anula las reglas de Hive aunque vaya después;
  - la lista de lo que la voz no cambia: el resultado primero, las secciones con etiqueta, el vocabulario llano, la glosa de los IDs, el desacuerdo sin halagos y el idioma;
  - que el tono entra solo en las transiciones, los cierres y el tratamiento;
  - que debe ignorarla quien fue lanzado por otro agente o escribe en archivos, commits, PRs, tickets, especificaciones o para otro agente. *Falso en la base cuando* `content/voices/` no existe.
- AC11. `deployment-manager.md` documenta `hive voice`, el bloque de voz y su relación con `update` y `plan remove`. *Falso en la base cuando* `rg -n "hive voice" _support/docs/architecture/deployment-manager.md` no devuelve nada.

## Entrega

Se reutilizan las decisiones de esta sesión para el mismo repositorio (D7-A y D8-A del 2026-09-27):

- **Modo:** `direct-base` a `rebuild/harness-engineering`, sin PR, construyendo en un worktree aislado.
  - La carpeta del cambio se versiona al aprobar el plan.
  - El código va en commits locales verificados, y después vienen la revisión de código y el push directo.
  - El cierre archiva la carpeta en otro commit.
- **Revisión de código:** `/code-review` de Claude Code antes del push.
- **Validación humana:** el usuario lee las tres voces, tanto el texto como un ejemplo de mensaje en cada una, antes de la revisión de código. Es la compuerta subjetiva del cambio.
- **Sin despliegue:** activar la voz en los CLIs reales del usuario se ofrece aparte.
