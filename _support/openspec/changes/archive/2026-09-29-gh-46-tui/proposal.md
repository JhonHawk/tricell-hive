# Interfaz de terminal de Hive a pantalla completa

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado el 2026-09-29 · integrado en `rebuild/harness-engineering` con `2c6e6d3` (push directo, D4-A) · especificación integrada en `specs/versioned-installation` |
| Tracker · GitHub Issues | • [#46 — Instalador TUI: CLIs, voz y tono, estado, diagnóstico, modelos por rol, proyecto e integraciones](https://github.com/JhonHawk/tricell-hive-private/issues/46) (pantallas 1, 2 y 3) |
| Git | `direct-base` a `rebuild/harness-engineering` desde el worktree · recorrido del usuario y `/code-review` antes del push (D4-A) |
| Verificación | `go vet ./...` · `go test -race -timeout 20m ./...` · `review-ux` y `sdd-verify` en `tmux` · recorrido del usuario |
| Siguiente paso | Ninguno en este cambio. Las pantallas 4 a 6 de #46 quedan fuera de alcance |

## Objetivo

El gestor ya hace con comandos todo lo que #46 quería mostrar: instalar y quitar CLIs, estado, actualizar desde un commit, volver a una release y la voz. La primera versión de la interfaz (T1–T6) encadenaba formularios de `huh`. En su recorrido del 2026-09-28, el usuario vio que cada vista quedaba impresa debajo de la anterior y pidió cuatro cambios:
- instalar y quitar en una sola vista, con casillas;
- volver atrás con una tecla desde cualquier vista;
- cambiar los ajustes en el lugar con las flechas;
- redibujar la pantalla completa.

Este replanteo reemplaza la interfaz por una aplicación de pantalla completa con Bubble Tea v2 y `bubbles`, sin `huh`. Hay una pila de vistas y una sola vista de CLIs que también muestra el estado.

`hive` sin argumentos y `hive tui` siguen abriéndola. Los comandos de texto (`hive install`, `update`, `voice`, `status`, `releases`, `plan`/`apply`) no cambian y son el camino para lectores de pantalla y guiones, porque se retira el modo `HIVE_ACCESSIBLE` (D1-A).

## Alcance y aceptación

**Incluye:**
- **Aplicación:** pantalla alterna, pila de vistas, barra de ayuda con las teclas, línea de estado, tema según el fondo y `NO_COLOR`.
- **Menú (D2-A):** CLIs, Update, Releases, Voice y Quit.
- **Vista CLIs:**
  - una fila por CLI con casilla (marcada = activo), release, versión y `drift`;
  - aplicar los cambios marcados, en dos pasos cuando se agregan y se quitan CLIs a la vez (D3-A);
  - la acción «Uninstall all».
- **Vistas Update, Releases y Voice:** ajustes en el lugar, resumen y confirmación.
- **Recuperación al abrir:** una operación pendiente se ofrece en una vista propia.
- **Retiro:** el modo `HIVE_ACCESSIBLE`, el `prompter` de `huh`, las pantallas secuenciales y la dependencia `charm.land/huh/v2`.
- **Documentación:** `deployment-manager.md` y el cambio en la especificación `versioned-installation`.

**Excluye:**
- Cambiar `hive install` (D14-A), `install.sh`, `bootstrap.sh` o el paquete offline.
- Las pantallas 4, 5 y 6 de #46: diagnóstico, modelos por rol y sección `## Hive`.
- Windows, que el gestor no soporta hoy.
- Temas configurables.
- Soporte de ratón.
- Terminales de menos de 80×24: se muestra un aviso en lugar de la vista.

**Restricciones que deben seguir siendo ciertas:**
- `go vet ./...` y `go test -race -timeout 20m ./...` pasan.
- Las pruebas de texto de `install`, `update`, `voice` y `bootstrap` pasan sin cambios.
- `tooling/management` y `tooling/distribution` no importan módulos externos (`deps_test.go`).
- La recuperación de una operación pendiente al abrir sigue funcionando.
- El tema cumple el contraste WCAG AA de su prueba.

**Criterios de aceptación.** Base: `dd146c0`, la interfaz secuencial con `huh`.

- AC1. `hive` sin argumentos o `hive tui [--home] [--state-dir] [--source]`, en una terminal, abren la aplicación en la pantalla alterna.
  - Muestra un menú con CLIs, Update, Releases, Voice y Quit, y una línea de estado.
  - Cambiar de vista redibuja la pantalla entera.
  - Al salir, la terminal vuelve a su contenido anterior sin nada impreso por la interfaz.
  - Sin terminal, `hive` imprime el error de uso de hoy, y `hive tui` falla con un mensaje que nombra los comandos de texto, incluidos `plan`/`apply` para quitar y `recover`.

  *Falso en la base cuando* el menú tiene siete entradas y cada vista deja impresa la anterior (captura del recorrido).
- AC2. Teclas para volver y salir:
  - Esc vuelve a la vista anterior desde cualquier vista.
  - Backspace también vuelve cuando ningún campo de texto tiene el foco; dentro de uno, borra un carácter. En el menú no hace nada.
  - Ctrl-C cierra la aplicación con 0 desde cualquier vista. Mientras se aplica un cambio, Ctrl-C y una interrupción externa (SIGINT) se ignoran hasta que termina.

  *Falso en la base cuando* Backspace no vuelve al menú desde Releases.
- AC3. La vista CLIs muestra una fila por CLI detectado o registrado: casilla marcada si está registrado, release (ID corto), versión y cantidad de recursos en `drift`. Abrirla y salir deja `state.json` y el home idénticos. *Falso en la base cuando* instalar, quitar y el estado son tres pantallas separadas.
- AC4. Aplicar en CLIs:
  - **Solo altas:** confirmar los CLIs adicionales que exige un recurso compartido, elegir capacidades opcionales (con su versión cuando la piden), ver el resumen y confirmar. Los archivos y el estado quedan iguales que con `hive install --hosts` y esas elecciones. Marcar un CLI con una instalación antigua significa migrarla e instalar, y el resumen lo dice. Un paquete que no pasa su verificación falla igual que `hive install`.
  - **Solo bajas:** el resumen (archivos que se quitan, recursos compartidos que se conservan, bloques de voz que se quitan) y la confirmación. El resultado es igual que el de `plan remove` seguido de `apply`.
  - **Altas y bajas juntas (D3-A):** dos pasos, primero quitar y después instalar. Cada paso tiene su resumen y su confirmación, y el segundo se calcula después de aplicar el primero. Rechazar el segundo deja aplicado el primero, y la vista lo dice.

  *Falso en la base cuando* no hay una vista con casillas que haga el diff.
- AC5. «Uninstall all» en la vista CLIs muestra el resumen de quitar todos los CLIs registrados, con Cancel seleccionado de entrada y sin atajo de una tecla para aplicar. Al confirmar, deja el mismo resultado que `plan remove` con todos ellos. Después no queda ningún CLI registrado ni voz activa. *Falso en la base cuando* no existe esa acción.
- AC6. La vista Voice muestra las filas Voice (Off y las voces de la fuente), Address, Name (solo con `name`) e Intensity, con la voz activa ya cargada.
  - ↑↓ cambia de fila y ←→ cambia el valor.
  - Enter muestra el resumen y la confirmación, y el resultado es igual que el de `hive voice set` u `off`.
  - Sin cambios respecto de la voz activa, dice que no hay nada que aplicar.

  *Falso en la base cuando* la voz se configura con un asistente de preguntas sucesivas.
- AC7. Update y Releases:
  - **Update:** muestra en una sola vista los campos de fuente (`.`) y revisión (`HEAD`), y Enter lleva al mismo resumen de `hive update` y a la confirmación.
  - **Releases:** lista las releases de la más reciente a la más antigua, con la instalada marcada y un filtro. Elegir una lleva al resumen de `plan install --release` para los CLIs registrados y a la confirmación.
  - **Resultados:** los dos dejan los mismos archivos y estado que sus comandos, y muestran sus estados vacíos según [la tabla de vistas](design.md#vistas).

  *Falso en la base cuando* fuente y revisión son un formulario secuencial que queda impreso.
- AC8. Resúmenes y confirmaciones:
  - El resumen se desplaza dentro de la pantalla.
  - Rechazar, pulsar Esc o un error del flujo vuelven a la vista de origen con el mensaje dentro de ella, no impreso debajo, y sin cambios.
  - Mientras se aplica, la vista muestra un indicador de progreso e ignora las teclas.
  - Al terminar con éxito, la vista de origen se refresca desde el estado y muestra el resultado dentro de ella.

  *Falso en la base cuando* «Cancelled. No changes applied.» queda impreso en la terminal.
- AC9. A 80×24 y a 120×40, ninguna fila se parte ni se corta sin «…», y las listas largas se desplazan dentro de la pantalla. Por debajo de 80×24, la aplicación muestra un aviso con el tamaño mínimo. *Falso en la base cuando* la fila de una release se corta en el borde derecho (captura del recorrido).
- AC10. `go.mod` no requiere `charm.land/huh/v2` y requiere `charm.land/bubbletea/v2` como dependencia directa. `go list -m all` no muestra ningún módulo que no estuviera en `dd146c0`. `HIVE_ACCESSIBLE` no aparece en `tooling/` ni en la ayuda. `go mod verify` y `deps_test.go` pasan. *Falso en la base cuando* `rg -n "charm.land/huh" go.mod` devuelve una línea.
- AC11. `deployment-manager.md` documenta:
  - la aplicación, sus vistas y teclas;
  - que los comandos de texto son el camino sin terminal;
  - la excepción de dependencias actualizada (Bubble Tea, `bubbles` y `lipgloss`).

  *Falso en la base cuando* `rg -n "HIVE_ACCESSIBLE" _support/docs/architecture/deployment-manager.md` devuelve líneas.

## Entrega (D4-A)

- **Modo:** `direct-base` a `rebuild/harness-engineering`, desde el worktree `.claude/worktrees/gh-46-tui` y la rama local `feat/gh-46-tui`. Antes del push, la rama se rebasa sobre la punta de la base, que avanzó con `4b82fd2` y `f18a7cb`, dos commits de guía que no tocan `tooling/`.
- **Orden:** verificación por tarea, luego `review-ux` y `sdd-verify`, luego el recorrido del usuario sobre un home de prueba, luego `/code-review`, y por último el push.
- **Sin despliegue:** usar la interfaz sobre la configuración real es decisión aparte del usuario.
- **Historial:** los commits de T1–T6 se conservan en la rama; el replanteo agrega commits nuevos, sin reescribirlos.
