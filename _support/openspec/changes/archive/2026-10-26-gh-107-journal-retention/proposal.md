# Retención de los journals de transacciones

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el [PR #222](https://github.com/JhonHawk/tricell-hive/pull/222) (`b9cb57cb`) |
| Tracker · GitHub Issues | • [#107 — Operación: añadir CI y robustecer tests y journals](https://github.com/JhonHawk/tricell-hive/issues/107) |
| Git | `automatic` · PR a `development` y merge · sin parada humana |
| Verificación | tests de `tooling/management` y `tooling/cli` · suite completa · in-vivo sintético · `/code-review` · medición tras el refresco local |
| Siguiente paso | Ninguno. La poda real del estado del mantenedor ocurre en el próximo `hive update` que cambie contenido |

## Request

> vamos con #107 evalua la premisa, pero si eso hace nuestro flujo más lento, no aplicara (en caso de que si aplique, despues del research haz el plan

Acción sugerida de [#107](https://github.com/JhonHawk/tricell-hive/issues/107), leída el 2026-10-26 (el ticket no trae criterios de aceptación):

> 1. Añadir CI mínima para `go vet ./...`, `go test ./...` y las pruebas Python; incluir `verify:dist` del frontend cuando esté disponible el toolchain declarado.
> 2. Separar los tests que legitimen un repositorio Git de los que deberían funcionar desde un paquete fuente o una ubicación temporal.
> 3. Definir retención y restauración de journals: qué se conserva, durante cuánto tiempo y cómo se usa sin debilitar la recuperación.

## Objective

Cada operación confirmada del gestor deja en `<StateDir>/transactions/` un journal de unos 12 MB, que nunca se borra. En la máquina del mantenedor eran 262 journals y 2,9 GB tras cinco semanas, a un ritmo de 100–230 MB por día. Solo `hive recover` necesita el journal de la operación pendiente, y la recuperación de onboarding el del core que referencia. Los demás no los lee nadie, aunque el instalador presenta la carpeta como "Private backups".

El cambio poda automáticamente los journals ya terminados y conserva los 10 más recientes. Nunca toca un journal que una recuperación pueda necesitar, y el texto y la documentación pasan a describir lo que la carpeta es de verdad. El disco queda acotado (unos 120 MB) sin un comando nuevo y sin coste apreciable por actualización.

## Scope and acceptance

**Incluido:** poda en `tooling/management`, el texto de la ruta en la salida del instalador y de la vista de hosts, `deployment-manager.md`, `README.md` y un delta de `versioned-installation`.

**Excluido, con su destino:**
- Punto 1 de #107 (CI): descartado. Haría el flujo más lento; la CI se retiró el 2026-09-30 (unos 27→8 min frente a unos 50 s de la suite local). Decisión del usuario: descartar lo que frene el flujo.
- Punto 2 de #107 (tests fuera de un clon Git o bajo `os.TempDir()`): descartado (D2-A). Ningún camino soportado llega a esos casos: `CONTRIBUTING.md` pide un clon y el paquete offline no incluye `tests/`. La protección de `temporaryLocation` (`tooling/management/plan.go:103`) es deliberada.
- Journals padre de onboarding en `<StateDir>/onboarding/`: fuera de alcance. Son pequeños (recibos de pasos) y en la máquina medida había 0.
- Un comando de restauración de transacciones ya confirmadas: no existe y no se crea. `hive recover` sigue siendo la única restauración.

**Restricciones que deben seguir ciertas:**
- Nunca se borra el journal que nombra `pending.json`, ni el `CoreID` de `onboarding-pending.json`, ni un journal en una fase distinta de `committed` o `recovered`.
- Si la poda falla, la operación sigue siendo un éxito y el resultado lleva un aviso. Nunca falla una instalación ya confirmada.
- `Recover` y `RecoverOnboarding` no cambian su comportamiento.
- La poda no decodifica journals completos: lee solo los campos iniciales, para no sumar segundos a cada `hive update`.
- La suite existente sigue pasando (`go vet ./...`, `go test ./...`, tests de skills).

**Búsqueda de especificaciones:** términos `journal`, `transaction`, `backup`, `prune`, `retain` en `_support/openspec/specs/` y en `deployment-manager.md`. Solo coincide `versioned-installation/spec.md`. Se leyeron "External operation recovery" (línea 88), "Retained release listing" (línea 129) y todas sus prohibiciones `SHALL NOT`; ninguna regula la retención de journals ni la prohíbe. `deployment-manager.md:456` afirma "Backups and release history are retained; the manager does not prune them", que este cambio modifica. Se añade el requisito "Transaction journal retention" en el delta `specs/versioned-installation/spec.md`.

- AC1. Tras una operación confirmada con `hive update` o `hive install` sin pasos opcionales, `transactions/` conserva como máximo los 10 journals `committed` o `recovered` más recientes, y el de esa operación es uno de ellos. *Falso en la base cuando* 12 aplicaciones seguidas dejan 12 journals.
- AC2. Una instalación con pasos opcionales (`Onboard` con pasos) poda igual al terminar. *Falso en la base cuando* 12 onboardings dejan 12 journals de core.
- AC3. La salida del instalador y de la vista de hosts ya no llama "Private backups" a `transactions/`, sino journals de recuperación, e indica que se conservan los 10 más recientes. *Falso en la base cuando* `tooling/cli/install.go:825` imprime `Private backups:`.
- AC4. `deployment-manager.md` y la especificación `versioned-installation` describen la retención: los 10 más recientes y nunca uno que una recuperación necesite. *Falso en la base cuando* `deployment-manager.md:456` dice que el gestor no poda.

## Delivery

- **Modo:** `automatic` (D3-A, respuesta del 2026-10-26). Rama `fix/gh-107-journal-retention` desde `development`, commits, `/code-review` sobre el candidato local, push, PR a `development` y merge cuando pasen las comprobaciones locales. Sin parada para validación humana; el repositorio no tiene CI.
- **Revisión:** Claude Code `/code-review` (D4-A) sobre el candidato local, después de la suite completa y antes del push.
- **Registro del cambio:** se cierra al final con un PR pequeño, sin revisión dedicada, que hace merge el agente.
- **Tras el merge:** el marcador `_support/workspace/maintainer.local` existe, así que AGENTS.md pide recompilar el binario y ejecutar `hive update`. Ese refresco poda los journals reales (unos 250 archivos, unos 2,8 GB); el usuario aceptó ese efecto al elegir D1-A.
- **Ticket:** al integrar, #107 se cierra con un comentario que explica que los puntos 1 y 2 se descartaron y por qué.
- No hay despliegue ni promoción a `master` en este alcance.
