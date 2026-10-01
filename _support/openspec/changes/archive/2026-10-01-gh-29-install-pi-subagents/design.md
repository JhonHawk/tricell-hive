# Diseño: pi-subagents en el plan de Pi

## Contexto verificado

- `integrations/pi/pi.go` (HEAD `e4fae19`) resuelve bloque, skill compartida y `<Pi home>/agents/`. Alcance de proyecto: error `Pi project scope is unsupported`.
- El gestor no ejecuta procesos de proveedor: `tooling/providers/run.go` solo admite pasos `manual`. No hay `exec` en `tooling/management`.
- Comentario del [29](https://github.com/JhonHawk/tricell-hive/issues/29) (2026-09-25, Pi 0.87.1): la identidad npm es el nombre (`pi-subagents`), no el pin; `pi install` con otra cadena sustituye la entrada; `pi remove` no vale de recover si la entrada ya era del usuario; `extensions: []` está declarado y no carga la extensión (conflicto); `pi list` no es JSON, hay que leer `settings.json`.
- Pi 1.0.0 instalado: `pi install <source>` sigue siendo el mecanismo (`docs/packages.md`). Esta máquina declara `pi-subagents` 0.74.0. D3-B usa ese pin solo si falta la entrada.
- Criterio de “rol seleccionable”: los pilotos de modelo están pausados. El 29 lo baja a descubrimiento de archivos.
- Spec actual (`Optional operation capabilities`): no ejecutar un instalador de proveedor hasta que la receta pase su puerta nativa; hasta entonces, manual.

Hechos vs supuestos: el comportamiento de `addSourceToSettings` en Pi 1.0 no se releyó en fuente; se asume el hallazgo de 0.87.1 salvo que una prueba con `pi` falso demuestre lo contrario. `pi install` de usuario no pide `--approve` (eso es proyecto).

## Enfoque

`pi-subagents` es **prerrequisito del host Pi**, no una capacidad opcional de onboarding. Si el plan de usuario incluye `pi`, el núcleo planifica un paso de paquete. Engram y Context7 siguen manuales.

No escribir `settings.json` como archivo gestionado de Hive: Pi lo reescribe. Invocar el binario `pi` resuelto (ruta absoluta, argumentos estructurados) con `PI_CODING_AGENT_DIR` igual al `PiHome` del plan.

### Detección

Leer `packages` en `<PiHome>/settings.json`. Una entrada cuenta como el paquete si su `source` (cadena u objeto con `source`) nombra `npm:pi-subagents` con o sin `@pin`.

| Estado | Plan | Apply | Recover (deshacer Hive) |
| --- | --- | --- | --- |
| Ausente | paso `install` con `npm:pi-subagents@0.74.0` | `pi install` esa fuente, tras revalidar que sigue ausente | `pi remove` esa fuente si el journal dice que esta operación la añadió, la cadena sigue igual, y el apply terminó | 
| Presente, filtro omitido o que no es `[]` | omisión idempotente | no llama a install | no quita |
| Presente con `extensions: []` | conflicto, no éxito | no aplica | — |
| Declarado pero faltan archivos o el pin del usuario no coincide con lo instalado | conflicto / informar; no reinstalar | no llama a install | no quita |
| Declaración de proyecto del mismo paquete | fuera de alcance | — | — |

Apply relee `settings.json` inmediatamente antes del efecto. Si la identidad `npm:pi-subagents` apareció desde el plan, se niega sin ejecutar `pi` (el plan está obsoleto). No se reinstala un paquete preexistente: el 29 lo exige decisión explícita; D3-B no la da.

Un `extensions` distinto de omitido y de `[]` no se trata como éxito de carga: o se puede ver que el filtro admite la extensión, o el plan informa disponibilidad no verificada / conflicto y no cambia settings.

Interrupción: si `pi install` arranca y el gestor no confirma que la settings quedó con la fuente Hive, el resultado es desconocido: no `pi remove`, bloqueo hasta reconciliar (spec de recuperación de operaciones externas). Recover tras un apply **terminado** que Hive añadió sí quita esa entrada: es el deshacer del 29, no un rollback de fallo.

Home sintético: `ExpandHostHomes` ya ignora el entorno. Las pruebas inyectan un ejecutable falso; nunca `npm`.

### Contratos

- Nuevo paso en el plan (campo propio, no un `Change` de archivo): fuente, acción (`install` / `omit` / `reinstall-user`), si Hive es dueña para recover.
- Journal: solo lo que esta operación añadió (fuente exacta). Recover compara esa cadena con la settings actual.
- Binario: `LookPath("pi")` o error nombrando que Pi debe estar en PATH. Argumentos: `install`, `npm:pi-subagents@0.74.0`. Entorno: copia acotada más `PI_CODING_AGENT_DIR=<PiHome>`.
- Confirmación de Hive: el paso de paquete cuenta como cambio de archivos a efectos de “plan que cambia algo”; un plan solo-omisión no pide confirmación extra por el paquete.

### Recuperación

Si el journal no marca dueño Hive, recover no llama a `pi remove`. Si el usuario cambió el pin después, la cadena ya no coincide: no quitar.

## Spec

Requisito nuevo en `versioned-installation`: el host Pi en alcance de usuario instala o omite `pi-subagents` como arriba. El catálogo opcional sigue limitado a Engram, Context7 y pi-subagents para el resumen de doctor; la receta ejecutada es solo Pi+usuario.
