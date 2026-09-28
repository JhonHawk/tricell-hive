# Interfaz de terminal de Hive con `huh`

| Campo | Valor actual |
| --- | --- |
| Estado | En curso · plan aprobado el 2026-09-28 · T1 en marcha |
| Tracker · GitHub Issues | • [#46 — Instalador TUI: CLIs, voz y tono, estado, diagnóstico, modelos por rol, proyecto e integraciones](https://github.com/JhonHawk/tricell-hive/issues/46) (pantallas 1, 2 y 3) |
| Git | `direct-base` a `rebuild/harness-engineering` · `/code-review` antes del push · recorrido del usuario antes de publicar (D7-A, D8-A) |
| Verificación | `go vet ./...` · `go test -race ./...` · `review-ux` y `sdd-verify` sobre el binario en una terminal real · recorrido del usuario |
| Siguiente paso | `flow-build`, empezando por T1 |

## Objetivo

Hive ya tiene en el gestor todo lo que #46 quería mostrar en una interfaz: instalar y quitar por CLI, estado, actualizar desde un commit, releases y la voz. Pero cada cosa es un comando distinto con sus opciones. Este cambio agrega una interfaz de terminal con `charmbracelet/huh` v2: `hive` sin argumentos abre un menú desde el que se hace todo eso con vista previa y confirmación. Los comandos actuales, `install.sh` y la instalación en línea no cambian (D14-A).

La interfaz no duplica la lógica. Los comandos ya aceptan `(args, in, out, interactive)` y reutilizan el resumen y la confirmación. El cambio separa esas preguntas en una interfaz `prompter`, con la implementación de texto actual y otra con `huh`, y la interfaz nueva recorre los mismos flujos con la segunda.

## Alcance y aceptación

**Incluye (D15-B):**
- **Menú:** Status, Install CLIs, Remove CLIs, Update, Releases, Voice y Quit.
- **Pantallas:** una por entrada, con vista previa y confirmación antes de cualquier escritura.
- **Recuperación al abrir:** si hay una operación pendiente, la interfaz lo muestra al empezar y ofrece recuperarla.
- **`hive tui [--home] [--state-dir] [--source]`:** la forma explícita, que permite abrir la interfaz sobre un home de prueba.
- **Modo accesible:** preguntas en texto plano con `HIVE_ACCESSIBLE=1`, para lectores de pantalla y para las pruebas.
- **La dependencia:** `charm.land/huh/v2` v2.0.3 (D16-A), registrada como excepción explícita a la regla de `AGENTS.md` sobre dependencias.
- **Documentación:** en `deployment-manager.md` y el cambio en la especificación `versioned-installation`.

**Excluye:**
- Reemplazar las preguntas de texto de `hive install` (D14-A).
- Las pantallas 4, 5 y 6 de #46: diagnóstico, modelos por rol y sección `## Hive`.
- Una vista de diff con desplazamiento: la vista previa usa el mismo resumen de texto de los comandos.
- Windows, que el gestor no soporta hoy.
- Temas visuales configurables.

**Restricciones que deben seguir siendo ciertas:**
- `go vet ./...`, `go test ./...` y `go test -race ./...` pasan.
- Las pruebas de texto de `install`, `update` y `voice` pasan sin cambios.
- `install.sh`, `bootstrap.sh` y el paquete offline funcionan igual.
- `tooling/management` y `tooling/distribution` no importan módulos externos.

**Criterios de aceptación:**
- AC1. `hive` sin argumentos, o `hive tui [--home] [--state-dir] [--source]`, en una terminal, abre un menú con Status, Install CLIs, Remove CLIs, Update, Releases, Voice y Quit, en ese orden y siempre visibles, con una línea de estado encima.
  - Sin terminal, `hive` sin argumentos imprime el mismo error de uso de hoy y sale con 1.
  - Con `HIVE_ACCESSIBLE=1`, el mismo menú corre como preguntas de texto que leen la entrada estándar, y el fin de la entrada sale con 0.

  *Falso en la base cuando* `hive` sin argumentos siempre imprime el error de uso y `hive tui` es un comando desconocido.
- AC2. Install CLIs: elegir varios CLIs detectados (ninguno marcado de entrada, cada uno con su estado), confirmar los CLIs adicionales que exige un recurso compartido, elegir capacidades opcionales, ver el resumen y confirmar. El resultado en archivos y estado es el mismo que el de `hive install --hosts` con esas elecciones. *Falso en la base cuando* no hay interfaz.
- AC3. Remove CLIs: elegir CLIs instalados, ver qué archivos se quitan (incluidos los bloques de voz) y confirmar. El resultado es el mismo que el de `plan remove` seguido de `apply`. *Falso en la base cuando* no hay interfaz.
- AC4. Status muestra, por CLI registrado, la release instalada (ID corto), la versión del producto, cuántos recursos están en `drift` y la voz activa. Solo lee: `state.json` y los archivos del home quedan idénticos. *Falso en la base cuando* no hay interfaz.
- AC5. En Update se escriben la fuente (por defecto `.`) y la revisión (por defecto `HEAD`), se ve el mismo resumen que en `hive update` y se confirma. En Releases se ven las releases de la más reciente a la más antigua; elegir una ofrece volver a ella con `plan install --release` para los CLIs registrados, con vista previa y confirmación. *Falso en la base cuando* no hay interfaz.
- AC6. Voice muestra la voz activa y permite elegir voz, tratamiento, nombre e intensidad, o apagarla, con vista previa y confirmación. El resultado es el mismo que con `hive voice set` u `off`. *Falso en la base cuando* no hay interfaz.
- AC7. `go.mod` requiere `charm.land/huh/v2` v2.0.3, y `go mod verify` pasa. Una prueba recorre las importaciones de `tooling/management` y `tooling/distribution` y comprueba que no dependen de ningún módulo externo. *Falso en la base cuando* `go.mod` no tiene `require`.
- AC8. `install`, `update` y `voice` obtienen las elecciones del usuario a través de una interfaz `prompter` con una implementación de texto, y sus pruebas de texto pasan sin modificarse. *Falso en la base cuando* `rg -n "type prompter" tooling/cli` no devuelve nada.
- AC9. `deployment-manager.md` documenta la interfaz, `hive tui`, su modo accesible y la dependencia, y `hive --help` menciona `hive tui` y `HIVE_ACCESSIBLE`. *Falso en la base cuando* `rg -n "HIVE_ACCESSIBLE" _support/docs/architecture/deployment-manager.md` no devuelve nada.
- AC10. Dentro de cualquier pantalla, pasa lo siguiente y la interfaz sigue abierta:
  - rechazar la confirmación, cancelar (Ctrl-C o Esc) o llegar al fin de la entrada imprime «Cancelled. No changes applied.», no cambia nada y vuelve al menú;
  - un error del flujo imprime el mensaje del comando y vuelve al menú.

  *Falso en la base cuando* no hay interfaz.
- AC11. Cada pantalla tiene su mensaje para el caso vacío o límite, según la tabla de [la interfaz](design.md#la-interfaz): sin CLIs registrados, sin releases, la release ya instalada y una fuente sin catálogo o sin voces. *Falso en la base cuando* no hay interfaz.

## Entrega

Se reutilizan las decisiones de esta sesión para el mismo repositorio:
- **Modo (D7-A):** `direct-base` a `rebuild/harness-engineering`, desde un worktree aislado.
- **Revisión de código (D8-A):** `/code-review` antes del push.
- **Validación humana:** la interfaz es algo que el usuario tiene que ver, así que el usuario la recorre sobre un home de prueba antes de la revisión de código y del push.
- **Sin despliegue:** usar la interfaz sobre la configuración real es decisión aparte del usuario.
