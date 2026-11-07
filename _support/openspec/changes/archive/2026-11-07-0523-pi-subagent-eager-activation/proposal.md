# pi-subagent-eager-activation — Pi starts with the subagent tool active

| Field | Value |
| --- | --- |
| Estado | Completado · PR #413, merge d12d72bc |
| Tracker | Ninguno |
| Git | tricell-hive, base `development` en `8956c9a3`, rama propuesta `feat/pi-subagent-eager-activation` |

## Decisiones

2026-11-06
- **H18 A** — ¿Se corrige que en Pi la herramienta `subagent` llegue tarde? Sí: lo hace la integración de Hive para Pi. Además, H14 a H17 quedan como observaciones.
- **D5 A** — ¿Cómo queda `subagent` disponible desde el inicio? El manager fija una sola clave, `toolActivation: "eager"`, en `extensions/subagent/config.json` cuando falta. La registra en su estado, no toca un valor que puso el usuario y, al desinstalar, la quita solo si sigue siendo la que puso Hive.
- **D6 A** — ¿Se mantiene ese enfoque aunque la revisión del plan subió el alcance a unos 15 archivos? Sí, con el alcance completo y tres supuestos:
  - la recuperación se hace por clave, como el piso de espera de Codex;
  - la instalación desde paquete también incluye el paso;
  - las actualizaciones de release no escriben la clave, y `hive doctor` indica `hive install` cuando falta.

## Pedido

El usuario aprobó corregir H18 («A», 2026-11-06), observado en la sesión Pi `01a1b3d3` (Pi 1.5.0, Grok 4.7, Carvento CAR-5):

- pi-subagents 0.83.0 abre cada sesión solo con el cargador `subagents_enable`;
- el agente lo llamó a las 20:33 y a las 20:59, y se detuvo con «responde lo que sea para seguir»;
- quedó parado 11 minutos, hasta que el usuario respondió a las 21:10;
- no lanzó revisiones mientras el usuario validaba.

La documentación de la extensión (`docs/configuration.md:83-99`, 0.83.0 instalada) explica la causa y el arreglo. Con algunos proveedores, `subagent` aparece recién en el siguiente prompt del usuario. Con `{"toolActivation": "eager"}` en `<Pi home>/extensions/subagent/config.json`, el cargador no se registra. La extensión encuentra el archivo con `getAgentDir()`, que respeta `PI_CODING_AGENT_DIR`, y acepta `auto`, `dynamic` o `eager` (`src/extension/config.js:223,242`).

## Objetivo

Hoy Hive instala `pi-subagents` con `pi install`, pero nunca toca su archivo de configuración. Por eso cada sesión de Pi arranca en modo diferido.

Con este cambio, cuando la clave falta, la instalación de alcance usuario, la instalación desde paquete y la actualización desde el código fuente que incluyen Pi dejan `toolActivation` en `"eager"`. El valor que haya puesto el usuario se respeta, y al desinstalar se quita solo lo que Hive escribió. `hive doctor` muestra el estado de la clave.

## Alcance y aceptación

**Entra:**
- el paso de plan y apply del manager, con registro, journal, recuperación por clave y eliminación;
- su convivencia con la migración del instalador anterior, que ya administraba `forceTopLevelAsync` en ese archivo;
- los resúmenes de instalación y desinstalación;
- la comparación de planes sin cambios;
- las validaciones de actualización de release;
- la fila de `hive doctor` y de la vista de integraciones de la TUI;
- la spec y la documentación que el cambio deja falsa.

**No entra:**
- escribir la clave en una actualización de release;
- cualquier otra clave del archivo;
- cambiar la versión de `pi-subagents`;
- la guía de las skills.

- AC1 — Un `hive install` de alcance usuario, una instalación desde paquete o una actualización desde el código fuente que incluye Pi, con `<Pi home>/extensions/subagent/config.json` ausente: el resumen muestra la ruta, la clave, el valor y que Pi necesita reiniciarse. Apply crea el archivo con `{"toolActivation": "eager"}` (modo 0600, carpetas en 0700), y el estado registra la ruta, el valor, que Hive creó el archivo y las carpetas que creó. · Hoy es falso: el plan no tiene ese paso y el archivo no se crea.
- AC2 — Si el archivo es un objeto JSON sin `toolActivation`, apply agrega solo esa clave y deja las demás claves, sus valores y su orden. Conserva el modo del archivo y crea `config.json.hive-backup` cuando no existe una copia. · Hoy es falso: el archivo no cambia.
- AC3 — Una actualización posterior, con la clave en `"eager"` registrada por Hive para esa ruta, informa el paso como ya aplicado, no escribe y conserva el registro. · Hoy es falso: no hay registro que conservar.
- AC4 — Si el archivo tiene `toolActivation` con cualquier valor que Hive no registró, el plan no escribe, no registra e informa el valor como puesto por el usuario. El plan tampoco escribe e informa un paso manual con la clave y el valor cuando el archivo:
  - no es un objeto JSON con claves únicas;
  - es un enlace simbólico o no es un archivo regular;
  - supera 1 MiB;
  - o está en una carpeta que no se puede escribir de forma segura.

  Si el registro de Hive es de otro Pi home, no escribe en el nuevo y avisa del registro que conserva. · Hoy es falso: el plan no menciona la clave.
- AC5 — `hive plan remove` sobre Pi borra `toolActivation` solo si el estado registra que Hive la escribió y el valor sigue en `"eager"`. Borra el archivo y las carpetas solo si los creó Hive y quedaron vacíos, y el backup solo si lo creó Hive y la clave se quitó. El resumen informa cada cosa que deja. · Hoy es falso: la desinstalación no toca ese archivo.
- AC6 — Un apply que falla después de insertar la clave, o una eliminación interrumpida, se recupera por clave desde el journal: la recuperación quita la clave insertada o restituye la borrada, sin reescribir el resto del archivo. · Hoy es falso: no existe ese paso en el journal.
- AC7 — Una actualización desde el código fuente cuya única diferencia es este paso no se trata como «sin cambios»: pide confirmación como cualquier plan que escribe. · Hoy es falso: `PlanUnchanged` (`migration.go:103-171`) no conoce el paso.
- AC8 — En una migración del instalador anterior que edita el mismo archivo, apply no falla por «cambio concurrente»: el paso parte del estado que deja la migración. · Hoy es falso: un paso que revalida el fingerprint del plan fallaría después de la escritura de la migración (`apply.go:396-446`).
- AC9 — Una actualización de release no planifica el paso y rechaza un plan que lo lleve. · Hoy es falso: el campo no existe, así que no hay rechazo.
- AC10 — Con Pi registrado, `hive doctor` y la vista de integraciones de la TUI muestran la fila «Pi subagent activation» con uno de cuatro estados:
  - eager, escrito por Hive;
  - puesto por el usuario, con su valor;
  - no fijado, indicando `hive install`;
  - no leído.

  Solo lee ese archivo y el estado de Hive, y no muestra otro texto del archivo. Con Pi y Codex registrados, la vista de la TUI muestra todas sus filas. · Hoy es falso: no existe la fila.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

Leído el 2026-11-06 sobre `development` `8956c9a3` y revisado por `hive-review-plan`:

- `tooling/management/pi_package.go` y `plan.go:506-507,640-680`, donde `planPiPackage` corre en install, en la instalación desde paquete y en la actualización desde el código fuente.
- `codex_wait_floor.go` (paso, registro y recuperación por clave), `integrations/codex/wait_floor_target.go:19-49` (seguridad de archivos) y `codex_wait_floor.go:397` (`StoredCodexWaitFloor`): es el patrón a seguir.
- `apply.go:219-223` (arrastre de registros), `396-446` (orden: migración y después pasos), `762-765` (notas) y `858-866,970-980` (recuperación de una eliminación).
- `migration.go:103-171`, con `PlanUnchanged`, y `497-510`, con el journal de la migración.
- `update_providers.go:33-49`, `update_coordinator.go:94` y `plan.go:871` (`validatePlan`).
- `install.go:883-960` (`showInstallSummary`), `tui_hosts.go:49` (resumen de desinstalación) y `tui_integrations_view.go:18-22` (límite de filas).
- `doctor_integrations.go:1-8,142-168`: hoy dice que Hive no lee la configuración de Pi.
- `tooling/legacy/transforms.go:313-363`: el instalador anterior es dueño de `forceTopLevelAsync` en el mismo archivo.
- `README.md:191`, `llms.txt:62`, `_support/docs/architecture/tool-dependencies.md:20` y `deployment-manager.md:111`: dicen que el piso de Codex es lo único que Hive escribe en la configuración de un host.
- `_support/openspec/specs/versioned-installation/spec.md`: los requisitos «Pi user-scope install declares pi-subagents» y «Read-only diagnostics».

## Entrega

Entrega `automatic` y revisión `Native: Claude Code /code-review`, ya respondidas en esta sesión para tricell-hive: revisión, push, PR y merge sin parar. Después del merge se reconstruye el binario y se hace `hive update`, como pide `AGENTS.md` del repositorio; ese `hive update` es el que aplica la clave en el `~/.pi` real.

## Estado de la revisión y progreso

2026-11-06 — `hive-review-plan` encontró 4 P1, 8 P2 y 7 P3 (R1–R19). Se aplicaron al plan:
- R1, la clave ya escrita por Hive (AC3);
- R2, el requisito de diagnósticos modificado;
- R3, la migración anterior (AC8);
- R4, `PlanUnchanged` (AC7);
- R5, los resúmenes;
- R6, R7 y R11, como supuestos de D6;
- R8, la seguridad de archivos;
- R9, otro Pi home;
- R10, la vista de la TUI;
- R12, el modo;
- R13, los documentos;
- R14, la trazabilidad;
- R15, `--home`;
- R16, el aviso de reinicio;
- R17, el orden y las claves únicas;
- R18, las carpetas;
- R19, los tests existentes.

El usuario confirmó el alcance completo (D6 A).
