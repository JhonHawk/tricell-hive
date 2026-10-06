# Instalar pi-subagents al desplegar Pi

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `81682f9` (PR 94) |
| Tracker · GitHub Issues | • [29 — Instalar pi-subagents al desplegar el host Pi](https://github.com/JhonHawk/tricell-hive-private/issues/29) |
| Git | automático · tricell-hive `development` · sin parada humana antes del push |
| Verificación | pruebas Go con `pi` falso y home sintético · sin piloto de modelo |
| Siguiente paso | commit, PR y `hive-review-code` |

## Objetivo

Al instalar o actualizar Hive con el host Pi en alcance de usuario, el gestor deja `pi-subagents` declarado en la settings de usuario de Pi con el instalador nativo (`pi install`). Hoy escribe el bloque de guía y los roles, pero sin esa extensión `subagent` no selecciona esos roles. El resultado es una instalación de Pi que Hive da por correcta y que no puede delegar.

## Alcance y aceptación

Incluido: plan, apply y recover del paquete `pi-subagents` cuando el plan de usuario incluye `pi`; detección por nombre en `settings.json`; pin Hive solo si falta la entrada; pruebas con instalador falso.

Fuera: parchear `pi-subagents`; instalar Engram u otras extensiones; alcance de proyecto (`pi install -l`); reescribir `settings.json` entero; lanzar un modelo; cambiar el formato de los roles o `excludeTools`; el hallazgo de hijos asíncronos en Pi 1.0 (H1, declinado 2026-10-01).

Restricciones: no interpolar la fuente en un shell; no tocar la settings real desde `go test`; no reinstalar ni cambiar un pin que el usuario ya tenga; `recover` no quita una entrada preexistente.

- AC1. Un plan de usuario que incluye `pi` muestra la instalación de `npm:pi-subagents@0.74.0` si esa settings no declara el paquete por nombre, u omisión idempotente si ya está declarado. *Falso en la base cuando* `integrations/pi/pi.go` solo expone `AGENTS.md`, la skill y `agents/`, y el plan no tiene un paso de paquete.
- AC2. `apply` de ese plan deja la fuente Hive en la settings de usuario del home planificado y no altera las demás entradas de `packages`. *Falso en la base cuando* el gestor no ejecuta `pi install`.
- AC3. Tras apply, la settings del home planificado declara la fuente Hive y existen los archivos convencionales del paquete en ese home (sin lanzar Pi ni un modelo). *Falso en la base cuando* Hive no instala el paquete.
- AC4. `recover` quita solo la entrada que esta operación añadió, y solo mientras siga siendo esa cadena; si el usuario ya tenía el paquete, recover no lo quita. *Falso en la base cuando* no hay paso journalizado de paquete.
- AC5. Las pruebas cubren ausencia, presencia previa y conflicto (`extensions: []`), con home sintético, sin llamar a npm ni escribir la settings real. *Falso en la base cuando* no existen esas pruebas.

## Entrega

- Repositorio: tricell-hive. Base: `development`.
- Modo Git: D1-A automático (sesión 2026-10-01, mismo repositorio): rama, commit, push, PR, `hive-review-code` (D2-A) y merge cuando CI y revisión pasen.
- Pin de instalaciones nuevas: D3-B `npm:pi-subagents@0.74.0` (2026-10-01). Un pin distinto ya declarado no se reescribe.
- Cierre del registro de cambio: con el código, en el PR a `development`.
- `hive update` tras el merge sigue pidiendo terminal interactiva; no forma parte de AC1–AC5.
