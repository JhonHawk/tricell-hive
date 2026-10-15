# Instalar pi-subagents sin pin de versión

| Campo | Valor actual |
| --- | --- |
| Estado | Integrado en `development` el 2026-10-15 por el [PR #52](https://github.com/JhonHawk/tricell-hive/pull/52), merge `1307b1fb`; desplegado en local desde `development` y archivado el 2026-10-15. Estado previo (historia): En curso · T1–T3 hechos; publicación no autorizada |
| Tracker · GitHub Issues | Sin ticket vinculado |
| Git | Integrado por merge del PR #52; la rama `feat/flash-class-guidance` y su worktree se eliminaron tras el merge |
| Verificación | `go test ./integrations/pi ./tooling/management` PASS · `hive-review-code` sin bloqueos |
| Siguiente paso | Ninguno; registro cerrado. La publicación que D6 no autorizaba quedó autorizada el 2026-10-15 con la integración de la rama |

## Request

> Planear el pin 0.77.0 (Recommended)

> pregunta, necesitamos pinnear? no es más fácil que se instale el latest?

> y considera hacer lo mismo con las demas herramientas

Decisiones de esta sesión (2026-10-11): D4-B fuente sin versión; D5-A solo pi-subagents; D6 commit a la rama (no el modo automático); D7-A `hive-review-code`.

## Objective

En una instalación o actualización de Hive con el host Pi en alcance de usuario, si `settings.json` no declara `pi-subagents`, el gestor ejecuta `pi install npm:pi-subagents` (sin `@versión`). Pi 1.1.0 guarda esa cadena tal cual; npm resuelve la última versión en ese momento. Un pin que el usuario ya tenga no se reescribe. Así las instalaciones nuevas no se quedan en 0.74.0, que en Pi 1.0+ no arranca hijos en segundo plano.

## Scope and acceptance

Incluido: constante `HiveSubagentsSource`, apply/recover sintético, pruebas, goldens, requisito `versioned-installation`, README, SECURITY.md, `installer.md`, `agent-delivery.md`, `tool-dependencies.md`.

Fuera: instalar gentle-engram, rpiv o pi-web-access; recetas de Engram o Context7; reescribir pins existentes; `pi update --extensions` en máquinas ya instaladas; push, PR o merge; cambiar `feat/guidance-rework`.

Restricciones: no interpolar la fuente en un shell; `go test` no llama a npm ni escribe la settings real; recover solo quita la cadena exacta que esta operación añadió.

- AC1. Cuando un plan de usuario incluye `pi` y la settings no declara `pi-subagents`, plan, apply, recover y el estado usan la fuente `npm:pi-subagents` (sin `@versión`). *Falso en la base cuando* `HiveSubagentsSource` es `npm:pi-subagents@0.74.0`.
- AC2. Los documentos de producto y arquitectura que nombran el comando de instalación dicen `pi install npm:pi-subagents` sin pin. *Falso en la base cuando* README.md:108 y SECURITY.md:29 dicen `@0.74.0`.
- AC3. El requisito «Pi user-scope install declares pi-subagents» y el escenario de paquete de red exigen la fuente sin versión. *Falso en la base cuando* `_support/openspec/specs/versioned-installation/spec.md` nombra `@0.74.0` y «pinned pi-subagents source».

## Delivery

- Repositorio: tricell-hive. Base: `development`.
- Git (D6, texto libre «commit a la rama»): rama de trabajo desde `development`, commits locales. Sin push, PR ni merge hasta una instrucción posterior.
- Revisión de implementación (D7-A): `hive-review-code`; `hive-refute-claim` contrasta hallazgos que bloqueen.
- Cierre del registro: queda local con el código hasta que se autorice publicarlo.
- `hive update` tras un merge futuro sigue pidiendo terminal; no forma parte de AC1–AC3.
