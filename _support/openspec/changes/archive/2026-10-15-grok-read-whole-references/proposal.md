# Grok lee completas las referencias de Hive

| Campo | Valor actual |
| --- | --- |
| Estado | Integrado en `development` el 2026-10-15 por el [PR #52](https://github.com/JhonHawk/tricell-hive/pull/52), merge `1307b1fb`; desplegado en local desde `development` y archivado el 2026-10-15. Estado previo (historia): En validación · commits a90e51e9 (T1), 848baf14 (T2), 591ed32f (T3) en `feat/flash-class-guidance`; desplegado localmente en 71ef341d el 2026-10-13 |
| Tracker · GitHub Issues | Sin issue; nace de la auditoría de la sesión carvento Grok `01a1371f` (2026-10-13) |
| Git | Integrado por merge del PR #52; la rama `feat/flash-class-guidance` y su worktree se eliminaron tras el merge |
| Verificación | go vet · go test · `/code-review` · prueba real en Grok (D2-A) |
| Siguiente paso | Ninguno; registro cerrado |

## Request

> ve haciendo research sobre si lo del recorte se puede modificar

Respuesta a la ruta R1 (2026-10-13): "Planificar el hook (Recomendado)" — planificar un hook `PreToolUse` de Grok que quita `offset`/`limit` en las lecturas de referencias de Hive, instalado por el manager.

Decisiones del 2026-10-13: D1-A (el hook lo ejecuta un subcomando oculto del binario `hive`), D2-A (prueba real en Grok tras el build), D3-A (rama de prueba activa, sin commits hasta pedirlo), D4-A (`/code-review` de Claude Code sobre el diff).

## Objective

Los modelos de Grok leen las referencias de Hive con `limit` (80, 120, 40) y se pierden el final: desde el 2026-10-01, 29 de 220 lecturas de referencias por sesión quedaron incompletas, en 22 de 65 sesiones de Grok, con una mediana de 60 líneas perdidas (`verification.md` 10 veces, `plan-format.md` 6, `reporting.md` 5). En Claude Code fueron 7 de 248. La herramienta `read_file` de Grok 1.0.55 lee hasta 1000 líneas y solo ignora `limit` en `SKILL.md`, `AGENTS.md` y `CLAUDE.md`; la guía de skills de Grok (`~/.grok/docs/user-guide/08-skills.md:235`) pide leer las referencias "con offset y limit". Los arreglos solo de redacción no han funcionado antes en Grok.

El cambio instala, con el manager, un hook global de Grok que reescribe esas lecturas para que Grok lea el archivo completo. Las referencias de Hive caben en una lectura (la mayor, `verification.md`, tiene 28 KB). El hook falla abierto: si no puede correr, la lectura sigue como el modelo la pidió.

## Scope and acceptance

Incluye: el subcomando del hook, el tipo de recurso `hook` del manager con propiedad por archivo, su instalación para Grok, la documentación del manager y la prueba real.

Excluye: hooks para otros hosts (Claude Code lee parcial solo en 3 % de las lecturas), cambios a la guía distribuida, lecturas de scripts y assets de las skills (sus lecturas parciales son legítimas), y la sesión de carvento auditada.

Restricciones: `go vet ./...` y `go test ./...` siguen en verde; los archivos ajenos en `~/.grok/hooks/` (engram, herdr) no se tocan; un `hive` anterior que lea un `state.json` con un registro `hook` lo rechaza (`invalid recorded kind`), así que se reconstruye el binario antes de desplegar, como ya pide `AGENTS.md`.

- AC1. `hive grok-hook read-whole --skills <dir>` recibe por stdin un evento `PreToolUse` de `read_file` sobre un `.md` dentro de `<dir>`, de hasta 1000 líneas y 64 KiB, con `offset` o `limit`, e imprime `{"hookSpecificOutput":{"hookEventName":"PreToolUse","updatedInput":…}}` cuyo `updatedInput` es la entrada original sin `offset` ni `limit`, con las demás claves intactas, use la entrada `target_file` o `path`. *False at the base when* el comando no existe.
- AC2. Con una ruta fuera de `<dir>`, un archivo que no es `.md`, uno que supera los límites, uno inexistente, una entrada sin `offset` ni `limit`, un JSON inválido, argumentos inválidos o un pánico interno, el subcomando no imprime nada y sale con 0. *False at the base when* `hive grok-hook` responde `unknown command` con salida distinta de 0.
- AC3. Instalar Grok escribe `<GrokHome>/hooks/hive-read-whole-references.json` con un grupo `PreToolUse` de matcher `read_file` cuyo comando es la ruta absoluta del binario `hive` que planificó, seguida de `grok-hook read-whole --skills <SkillsDir>`, y lo registra en el estado con tipo `hook`; el estado guardado se vuelve a leer sin error (`status`, `update`, quitar). *False at the base when* tras instalar Grok ese archivo no existe.
- AC7. Planificar una instalación de Grok desde un binario bajo el directorio temporal o la caché de compilación de Go falla con un mensaje que pide `go build`, y `apply` de un plan cuyo `executable` no es un archivo regular absoluto falla. *False at the base when* esos planes se aplican sin error.
- AC4. Quitar Grok borra ese archivo y deja intactos los demás de `hooks/`; un archivo ajeno previo en esa ruta hace fallar el plan por colisión; un archivo gestionado editado por el usuario se rechaza como los roles. *False at the base when* el manager no conoce el tipo `hook`.
- AC5. Reinstalar o actualizar con el binario en otra ruta reescribe el comando del hook con la ruta nueva. *False at the base when* no hay hook que reescribir.
- AC6. En una sesión real de Grok 1.0.55, un `read_file` registrado con `limit: 40` sobre una referencia de Hive de más de 40 líneas devuelve el archivo completo, en la sesión principal y en un hijo. *False at the base when* la misma lectura devuelve 40 líneas.

## Delivery

- Repositorio: tricell-hive, base `development`; trabajo en el worktree `.claude/worktrees/feat-flash-class-guidance`, rama de prueba activa `feat/flash-class-guidance` (D3-A).
- Sin commits hasta que el usuario los pida; entonces, uno por tarea verificada. Sin push, PR ni merge hasta que el usuario pruebe y decida.
- Despliegue local (reconstruir el binario y `hive update`) solo a pedido.
- Revisión del código: `/code-review` de Claude Code, una vez, sobre el diff completo, antes de pedir el commit (D4-A).
- Este registro queda sin commit hasta el merge; `flow-close` lo versiona.
