# test-accounts-index — Índice de cuentas de prueba en `_support/secrets/` y reutilizar antes de crear

| Campo | Valor |
| --- | --- |
| Estado | Borrador · plan escrito, falta la revisión del plan y la aprobación |
| Tracker | Sin ticket. Retoma la intención de #284 (cerrado; su regla se eliminó en `7a37381b`) |
| Git | `JhonHawk/tricell-hive`, base `development` en `2765624b` (origin, 2026-11-04). Por decisión del usuario (2026-11-04), el cambio va en la rama actual `feat/gh-311-minimal-research-entry` (HEAD `ee1ebf07`), que tiene el PR #352 abierto y 48 commits de gh-311 |

## Decisiones

2026-11-04
- ¿Dónde se guarda el índice de cuentas? En `_support/secrets/` de cada proyecto, con las cuentas locales y también las existentes de QA y de los entornos desplegados.
- ¿Qué se genera en QA y en los entornos desplegados? Nada por iniciativa propia: se usan las cuentas y los datos que ya existen, y se sigue pidiendo permiso antes de crear algo, como ya fija `deployment-checks.md`.
- ¿Qué guarda Engram? La existencia de cada cuenta y dónde está su secreto (el archivo en `_support/secrets/` o la referencia en un gestor como 1Password), nunca el valor.
- ¿Basta con poner la regla en `flow-walkthrough`? No: a las pruebas de una API por su camino real (`flow-verify` paso 1, `flow-build:52`) no las cubre `flow-walkthrough`.
- ¿Dónde vive la regla? En la guía global, como sección «Test accounts», con punteros desde los flujos que usan cuentas. (D1 A)
- ¿En qué rama va? En la rama actual de gh-311, con su PR #352, no en una rama nueva.
- ¿Qué revisión de código corre? La nativa de Claude Code, `/code-review`. (D3 A)

## Pedido

> Podemos empezar a usar la carpeta de _support/secrets para generar y guardar un indice de cuentas, ojo, esto es para entornos locales, ni en QA, ni en entornos despegados deberian generarse información mock de manera proactiva, ahi no queda mas que usar las cuentas e info que ya existe

> Si, que siga pidiendo permiso, eso es lo correcto
> Build, no prueba? por ejemplo si esta haciendo un api no prueba que funcione? (este tipo de peticiones no llega a flow-walkthrough)
> Si, incluye las cuentas existentes, en engram se puede guardar su existencia y donde localizar el archivo con los secretos o la referencia donde se hayan guardado, un gestor o similar

## Objetivo

Hoy ninguna regla de Hive dice dónde buscar una cuenta de prueba antes de crearla. En las sesiones del 11-02 al 11-04, las cuentas de cam estaban repartidas en 7 archivos, uno por ticket y sin índice; Engram no devolvió nada en 3 búsquedas; una org se recreó; y en QA se perdieron ~39 min por usar una cuenta que no veía el dato a probar. En carvento, la cuenta que documenta `AGENTS.md:151` da 401 en local.

Con el cambio, toda prueba que necesite iniciar sesión, en cualquier flujo o fuera de uno, lee primero el índice del proyecto y reutiliza la cuenta que encaje. En local se crea solo lo que falta y se anota en el índice. En QA y en desplegados no se crea nada sin permiso. Al cierre del trabajo, las filas del índice, sin valores, quedan en Engram.

## Alcance y aceptación

Dentro: «Secrets» y la sección nueva «Test accounts» de `content/guidance/global.md`; los punteros en `flow-verify`, `flow-build`, `flow-walkthrough`, `flow-deploy/references/deployment-checks.md`, `flow-close`, `flow-plan`, los roles `hive-verify-change` y `hive-review-ux`, y `workspace-conventions/references/artifact-layout.md`; el detector de lectura de secretos de `tests/pilot/regression.go`; una spec delta `test-accounts`; las páginas `verify.md`, `deploy.md` y `close.md` del sitio.

Fuera: llenar los índices de cam, carvento, fleetsystem o facti (va en sesiones de cada proyecto); scripts de seed; una clave nueva en `## Hive`; cambiar el permiso que pide `deployment-checks.md`; los casos de `tests/fixtures/regression/`.

- AC1 — Antes de una prueba por el camino real que inicia sesión o usa registros que el entorno ya tiene, la guía global indica leer `ACCOUNTS.md` en la ubicación de secretos y reutilizar la cuenta que encaje por entorno, rol y alcance; desde un worktree se lee y escribe el archivo del checkout principal, y si falta se reconstruye desde la entrada de memoria `accounts/index` · falso hoy porque `rg -n 'ACCOUNTS.md' content` no devuelve nada.
- AC2 — La guía global define «local» como base de datos y servicios en esta máquina, donde se crea solo lo que falta y se anota; en QA, una base de desarrollo compartida o en la nube y los entornos desplegados no se crean cuentas, registros ni datos mock por iniciativa propia, se usa lo existente y se pide permiso antes de crear · falso hoy porque `rg -n -i 'mock data|own initiative|run on this machine' content/guidance/global.md` no devuelve nada.
- AC3 — La guía global fija el formato de cada fila (entorno, rol, identificador, fuente del secreto como archivo o referencia en un gestor y nunca un valor, qué alcanza, estado con fecha), la marca de obsoleta con fecha y causa, el modo `0600` y el agregado de la ubicación a `.git/info/exclude` cuando Git no la ignora · falso hoy porque no existe ningún formato de índice.
- AC4 — Un hijo recibe filas y fuentes de secreto, nunca valores, y reporta las cuentas en lugar de escribir el índice; `hive-verify-change` y `hive-review-ux` apuntan a la regla · falso hoy porque `rg -n -i 'account' content/agents` no encuentra ninguna regla de cuentas.
- AC5 — Al cierre de un trabajo, las filas del índice, sin valores, se copian a la memoria persistente en `accounts/index`, y `flow-close` paso 7 apunta a la regla · falso hoy porque `flow-close` paso 7 dice solo «Write the outcome to persistent memory».
- AC6 — `flow-verify` paso 1, `flow-build:22`, `flow-walkthrough:10`, `deployment-checks.md:32` y `flow-plan` («Prerequisites») apuntan a «Test accounts» con una condición de lectura y sin repetir sus pasos, y `artifact-layout.md` lista `secrets/` · falso hoy porque ninguno nombra un índice de cuentas ni la carpeta `secrets/`.
- AC7 — «Secrets» nombra `_support/secrets/` como ubicación por defecto y exceptúa `ACCOUNTS.md` de la prohibición de leer archivos de secretos, y el detector S5 acepta leer ese archivo y sigue marcando los demás · falso hoy porque `global.md:38` no nombra ninguna ruta y `noSecretContentRead` marca cualquier lectura bajo `_support/secrets/`.
- AC8 — La carpeta del cambio tiene la delta `test-accounts` con un escenario por requisito (incluida la escritura al cierre), y las páginas `verify.md`, `deploy.md` y `close.md` describen el comportamiento en una frase cada una, enlazando la guía · falso hoy porque ninguna de las tres páginas menciona cuentas de prueba y la delta no tiene escenario de cierre.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- 2026-11-04: análisis de 117 transcripts del 11-02 al 11-04 (cam, carvento, fleetsystem, facti y el visor de Hive), con 4 hijos `hive-research`; los hallazgos H1 a H5 se dieron en la conversación.
- `62c3afaa` (#284, 2026-10-29) agregó la regla de reutilizar stack y cuentas en `flow-verify/references/review-and-gates.md`; `7a37381b` (2026-10-30, recorte «minimal core») la eliminó, junto con su texto en `site/.../flows/build.md`.
- `content/guidance/global.md:38`: el único almacén de secretos es «the project's declared untracked secret location», en modo `0600`.
- `_support/openspec/specs/context-delivery/spec.md:87`: la guía global contiene toda regla que aplica en cualquier flujo o fuera de uno.
- Puntos que tocan cuentas o semillas hoy: `flow-build/SKILL.md:22` y `:52`, `flow-verify/SKILL.md:12`, `flow-walkthrough/SKILL.md:10`, `flow-deploy/references/deployment-checks.md:32`, `flow-close/SKILL.md:18` y `:46`.
- `tests/pilot/regression.go:742-945`: `isSecretPath` trata todo `_support/secrets/` como secreto y `noSecretContentRead` marca cualquier lectura ahí (por eso existe T4).
- Roles que inician sesión: `content/agents/quality/hive-verify-change.md` y `content/agents/review/hive-review-ux.md`; ninguno menciona cuentas.
- Revisados sin cambio necesario: `_support/docs/architecture/` (no nombran la ubicación de secretos), `README.md` y `llms.txt` (no cambian comandos, hosts ni carpetas).
- `content/guidance/global.md` pesa 31 122 bytes en el árbol de trabajo (`ee1ebf07`) y 27 416 en `development`; `tests/content/budget_test.go` lo reporta sin cuota.

## Entrega

`ready` (D2 A): revisión nativa `/code-review` (D3 A), push a la rama actual, espera del CI del PR #352 y respuesta a sus revisiones; termina con el PR abierto para que el usuario haga el merge. No autoriza el merge, el auto-merge ni la cola de merge, ni cerrar gh-311.

## Estado de la revisión y avance

2026-11-04 — Revisión del plan (`hive-review-plan`). Cambios aplicados:
- La guía exceptúa `ACCOUNTS.md` de la prohibición de leer secretos, y el detector S5 se ajusta (T4).
- Se fija dónde se escribe desde un worktree o un hijo.
- Se agrega `.git/info/exclude` para cuando Git no ignora la carpeta.
- Se define «local» por dónde vive la base de datos.
- La ubicación por defecto se fija en «Secrets».
- Se agregan punteros en los roles `hive-verify-change` y `hive-review-ux` y en `flow-plan`.
- Los punteros pasan a ser solo condiciones de lectura.
- Se ajusta AC8 a cómo se fusionan las deltas.
- T3 vuelve al hilo principal.
- Se agregan las páginas `deploy.md` y `close.md`, la carpeta `secrets/` en `artifact-layout.md`, las sesiones en el mensaje de commit, el respaldo en la memoria del host y el disparador acotado.
- No aplican, porque el trabajo va en la rama actual y no en un worktree nuevo: el conflicto con #352, la cifra base del tamaño y las dependencias del sitio.
