# Tareas

**Commit base:** `18847d6`. El 2026-09-27 otra sesión trabajaba en `rebuild/harness-engineering`, así que la construcción va en el worktree `.claude/worktrees/gh-46`, rama local `feat/gh-46-update-releases`. Al final se rebasa sobre `origin/rebuild/harness-engineering` y se publica directo a esa rama (D7-A).

Los criterios de AC1 a AC7 no dependen del commit base: `update`, `releases` y `commits.json` tampoco existen en `18847d6`.

## Orden y ejecución

T1 y T2 pueden correr en paralelo, porque escriben en paquetes distintos (`tooling/cli` y `tooling/management`). T3 y T4 dependen de las dos y los hace un solo hijo en secuencia, porque ambas tocan `tooling/cli/main.go`. T5 va al final en el hilo principal.

Cada hijo recibe `AGENTS.md`, este cambio (`proposal.md`, `design.md` y `tasks.md`) y [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md) (solo T3). Puede hacer commits locales en su rama de trabajo solo si la entrega elegida lo permite.

### T1 — Despacho de subcomandos por función (#42)

- [x] `run` delega cada subcomando en su propia función, con el comportamiento visible sin cambios. Evidencia: `d6b1c41`. Las pruebas de caracterización pasan antes y después sin cambios (`main_test.go` con el mismo sha256), `run` mide 32 líneas, y `go vet` y `go test ./tooling/cli` pasan. Cambio de conducta aceptado: `hive foo bar` responde ahora `unknown command "foo"` en lugar de `unexpected positional arguments`.

**Closes:** AC8.

**Depends on:** ninguna.

**Locations:** `tooling/cli/main.go` (`run`), `tooling/cli/main_test.go`.

**Execution:** delegada a `backend-developer`. Tiene una interfaz ya fijada y no se solapa con T2.

**Test approach:** characterization.

**Changes:**
1. Antes de mover código, agregar a `main_test.go` un helper que capture la salida estándar con `os.Pipe`.
2. Agregar pruebas de caracterización para:
   - `plan install` de solo vista previa;
   - `plan` con `--out` seguido de `apply`;
   - `status`;
   - los errores de uso: sin subcomando, `plan` sin acción, argumentos posicionales y comando desconocido.

   Usar home y estado sintéticos, como las pruebas existentes. Deben pasar sobre la base.
3. Extraer una función por subcomando y otra para la ayuda, según [el diseño](design.md#despacho-de-subcomandos-42), sin cambiar mensajes, opciones ni salidas.

**Verification:** las pruebas nuevas pasan antes y después del refactor, sin modificarlas entre los dos momentos. `go test ./tooling/cli` pasa. Contar las líneas de `run` con `awk '/^func run\(/,/^}/' tooling/cli/main.go | wc -l`: debe dar 40 o menos.

### T2 — Commit de origen y listado de releases en el gestor

- [x] `Plan` lleva el commit de origen, `Apply` lo registra junto a la release y `Releases` lista los snapshots. Evidencia: `review-task`, AC6 cumplido y AC7 cumplido en la parte del gestor. Rompió el código a propósito de 14 maneras y cada prueba falló cuando faltaba lo que cubre. `go test -race ./tooling/management` pasa. Commit local con la versión de T2. Límites menores: el orden de los commits dentro de una release no se prueba, y la lista vacía en JSON (`[]` y no `null`) pasa a la prueba de T4.

**Closes:** AC6, AC7.

**Depends on:** ninguna.

**Locations:**
- `tooling/management/types.go`: `Plan` y el nuevo `ReleaseEntry`.
- `tooling/management/plan.go`: `validatePlan` y la nueva `Releases`.
- `tooling/management/apply.go`: `Engine.Apply`, antes de devolver `"unchanged"` y después de `commitTransaction`, con el bloqueo tomado.
- `tooling/management/management_test.go`.

**Execution:** delegada a `backend-developer`. Es el único que escribe en `tooling/management`.

**Test approach:** tdd.

**Changes:** según [commit de origen](design.md#commit-de-origen) y [`hive releases`](design.md#hive-releases). Pruebas nuevas, escritas antes del código:
- Un plan con `SourceCommit` aplicado crea `releases/<id>.commits.json` con ese commit.
- Un plan con `SourceCommit` cuyo `Apply` devuelve `"unchanged"` también lo registra.
- Un plan sin commit no crea el registro.
- Dos commits con el mismo contenido quedan listados una vez cada uno.
- Una transacción que falla antes de confirmarse no deja commit registrado.
- Un plan sin `SourceCommit` conserva el ID que tenía en la base: fijar un plan de ejemplo con su ID esperado.
- `validatePlan` rechaza un `SourceCommit` que no sea hexadecimal de 40 o 64 caracteres.
- `Releases` devuelve las entradas ordenadas, con sus consumidores tomados de `Records` y la lista de commits vacía cuando falta el registro.
- Tras un `plan install --release` de una release anterior, `Releases` sigue mostrando los CLIs de esa release.
- `Releases` falla nombrando el archivo cuando un snapshot tiene JSON inválido.

**Verification:** `go test -race ./tooling/management` pasa, incluidas las pruebas nuevas, que fallan sobre la base.

### T3 — Comando `hive update`

- [x] `hive update` actualiza desde un commit, según [el diseño](design.md#hive-update).
  - **Evidencia:** `review-task` da AC1 a AC5 cumplidos. Rompió el código a propósito de 15 maneras para comprobar que las pruebas fallan cuando falta lo que cubren, y ejecutó el CLI sobre un home de prueba.
  - **Desviaciones aceptadas:**
    - La extracción usa `target.Canonical(os.TempDir())`, porque en macOS `/var` y `/tmp` son enlaces simbólicos que el gestor rechaza.
    - El commit se adjunta al plan con `management.BindSourceCommit`, que sigue el mismo patrón que `BindInstaller`. Reemplaza a una copia del cálculo del ID del plan que el implementador había puesto en `tooling/cli`.
  - **Arreglado después de la verificación:** con un archivo por encima del límite de tamaño, el comando se colgaba en vez de fallar. Ahora detiene a Git antes de esperarlo. La prueba `TestArchiveGitCommitFailsPastSizeLimit` falló antes del arreglo, por tiempo agotado a los 10 s, y pasa después.
  - **Prueba añadida:** `TestFilteredGitEnvDropsRepositoryOverrides`, que comprueba el comportamiento existente; no se escribió antes de un cambio de código.
  - **AC3:** se aclaró para que coincida con el diseño en el caso sin cambios (decisión P1-A).

**Closes:** AC1, AC2, AC3, AC4, AC5.

**Depends on:** T1, T2.

**Locations:**
- Nuevos: `tooling/cli/update.go`, con el comando y el adaptador del tar de Git, y `tooling/cli/update_test.go`.
- La entrada `update` en el despacho de `tooling/cli/main.go` y en la ayuda.
- `tooling/distribution` no cambia.

**Execution:** delegada a `backend-developer`, el mismo hijo que T4, en secuencia.

**Test approach:** tdd.

**Changes:**

Montaje de las pruebas:
- Crean en `t.TempDir()` un repositorio Git con el catálogo mínimo: `content/guidance/global.md`, una skill, un rol en `content/agents/` e `integrations/agent-profiles.json`.
- Montan su propio catálogo dentro de `tooling/cli`, porque los helpers de `tooling/management` no se exportan. Toman como molde `management_test.go:31-43`.
- Instalan un primer commit para dos CLIs en un home sintético. Luego hacen un commit que cambia un archivo de `content/` y dejan otro cambio sin commit.
- Usan `t.Setenv("TMPDIR", dir)` para controlar dónde se extrae; por eso esas pruebas no usan `t.Parallel`.
- Necesitan `git` en el `PATH`, salvo el caso de Git ausente.

Casos:
- **AC1:** modo interactivo con la entrada `y`.
- **AC2:** `--dry-run`, con y sin modo interactivo.
- **AC3:** sin modo interactivo, primero sin opciones y después con `--out`, seguido de `apply`.
- **AC4:** `--rev -x`, un commit inexistente, un `--source` que no es un checkout de Git y un `PATH` vacío.
- **AC5:** `dir` queda vacío al terminar, en éxito y en un fallo después de extraer. El fallo se provoca sin gancho, con un commit cuyo catálogo es inválido, por ejemplo un rol duplicado.
- **Adaptador:** una prueba que pasa por `distribution.Extract` la salida real de `git archive --format=tar --prefix=…/`, con su cabecera pax global y sus directorios con barra final.
- **Mismo contenido:** un segundo commit que solo cambia un archivo fuera de `content/` termina con «sin cambios» y registra el commit.

**Verification:** `go test -race ./tooling/cli` pasa, con las pruebas nuevas en rojo sobre la base.

### T4 — Comando `hive releases`

- [x] `hive releases` imprime en JSON el resultado de `management.Releases`. Evidencia: `review-task` confirma la parte de la CLI de AC7. Las pruebas de `tooling/cli/releases_test.go` comprueban el orden, los commits, los consumidores y que las listas vacías salgan como `[]`; siete roturas deliberadas del código las hicieron fallar. `run` mide 36 líneas. El comando no modificó nada al ejecutarse sobre un estado de prueba.

**Closes:** AC7.

**Depends on:** T1, T2.

**Locations:** `tooling/cli/main.go` (función del subcomando y `--help`), `tooling/cli/main_test.go`.

**Execution:** delegada, el mismo hijo que T3, después de T3.

**Test approach:** tdd.

**Changes:** una prueba con un estado sintético de dos releases, una instalada y otra con registro de commits, que comprueba la salida JSON y el orden.

**Verification:** `go test ./tooling/cli` pasa.

### T5 — Documentación y comprobación local sobre el estado real

- [x] `deployment-manager.md` documenta los dos comandos, y las comprobaciones de solo lectura sobre el estado real pasan. Evidencia del implementador, del 2026-09-27 sobre el commit `25d124e`: `state.json` tiene el mismo hash antes y después (`cf0d0b5a839d11e4`), y hay 117 snapshots. `releases` lista 117, y la más reciente, `622087a518ff`, figura en los seis CLIs. `update --dry-run` nombra los seis CLIs y el commit `25d124eb73bb`, e informa «already up to date». No se creó ningún `*.commits.json`.
  - **Verificación:** `review-task` da AC9 y las partes sobre el estado real de AC7 y AC2 como cumplidos, repetidos sobre `24759bb` con el mismo hash de `state.json`.
  - **Documentación corregida tras la verificación:** los ejemplos de uso no mostraban `--home` y `--state-dir`, y faltaba decir qué pasa con `--out` cuando no hay cambios.
  - **Código corregido tras la verificación:** la carpeta temporal se borraba al final del comando y no antes de la confirmación, como pide el diseño. Ahora `planFromCommit` la borra en cuanto existe el plan (`f6da52d`). La prueba `TestUpdateRemovesExtractionBeforeConfirmation` falló antes del arreglo y pasa después.
  - **Mensaje de uso:** el que sale al ejecutar el CLI sin argumentos ahora incluye `update` y `releases`.

**Closes:** AC9, AC7 (la parte del estado real), AC2 (la parte del estado real).

**Depends on:** T3, T4.

**Locations:** `_support/docs/architecture/deployment-manager.md` (secciones «Commands» y «Preservation and state»).

**Execution:** hilo principal. El texto depende de las salidas finales de T3 y T4, y el brief sería más largo que el cambio.

**Test approach:** check. La comprobación son los comandos de solo lectura de la verificación.

**Changes:** documentar:
- `update`: sus opciones, qué CLIs elige, qué hace sin terminal, que necesita Git y un checkout (el paquete offline no), y que los atributos `export-ignore` o `export-subst` de un commit alterarían lo extraído;
- el registro `releases/<id>.commits.json`, incluido que un journal ya confirmado y recuperado deja la release sin commit;
- `releases`.

**Verification:**
- `rg -n "hive update|hive releases|commits.json" _support/docs/architecture/deployment-manager.md` devuelve las secciones nuevas.
- Sobre el estado real del usuario, que es solo lectura, anotar el hash de `state.json` con `shasum -a 256` antes y después.
- `go run ./tooling/cli releases | jq length` devuelve el número de snapshots de `releases/` sin contar los `*.commits.json` (116 al planear).
- `go run ./tooling/cli update --dry-run` muestra los seis CLIs instalados y el commit de `HEAD`.
- El hash de `state.json` no cambia.

## Verificación y revisión humana

| Comprobación | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Pruebas del gestor | Al cerrar cada tarea y al final, en local | `go vet ./...`, `go test ./...` y `go test -race ./...` en verde | Parte de la implementación |
| Verificación por tarea | Tras T1 a T4 | `review-task` por tarea contra sus `AC<n>` | Parte de la implementación |
| Prueba local sobre el estado real | En T5 | Comandos de solo lectura; hash de `state.json` sin cambios | Parte de la implementación: no escribe en la configuración global |
| Revisión de código | Antes de integrar | Pendiente de la pregunta de entrega | Pendiente |
| Despliegue real con `hive update` | Después de integrar, si el usuario lo pide | `go run ./tooling/cli update` en una terminal, que escribe en la configuración global de los seis CLIs | Necesita autorización explícita (`AGENTS.md`, «Scope and preservation») |

No hay superficie de interfaz gráfica, así que no aplican `review-ux` ni un recorrido en navegador. La salida de terminal de `update` se revisa en la prueba local de T5.

## Estado de la revisión y avance

**Revisión de código** (2026-09-27): `/code-review` de Claude Code con esfuerzo alto sobre `18847d6..1a2c850`. Diez hallazgos, ninguno de seguridad crítica.

- **Arreglados:**
  - **F1:** con `--out` y sin cambios, el plan no se guardaba. Ahora se guarda; la especificación, AC3 y el diseño se ajustaron.
  - **F2:** un aviso se perdía si además fallaba `finishApply`.
  - **F4:** el entorno de Git solo quitaba tres variables. Ahora quita todas las que lista `git rev-parse --local-env-vars`.
  - **F5:** faltaba documentar que también cuentan los atributos locales de Git.
  - **F6:** el error de `rev-parse` escondía la causa real, por ejemplo `safe.directory`.
  - **F8:** gzip con `BestSpeed`.
  - **F9:** se quitaron una validación redundante y un reordenamiento que no hacía falta.
  - **F10:** el comentario del paquete apuntaba a una ruta que se archiva al cerrar el cambio.
  - Pruebas nuevas, las dos en rojo antes de su arreglo: `TestUpdateOutSavesPlanWhenContentUnchanged` (falló de forma observable) y `TestFilteredGitEnvDropsRepositoryOverrides`, ampliada (su rojo fue de compilación por el cambio de firma).
- **No aplicados, como sugerencias:**
  - **F3:** que el aviso de `Apply` viaje como un valor aparte. Cambiaría la firma de `Apply` en todos sus llamadores, y el aviso solo aparece si falla la escritura del registro.
  - **F7:** que `Releases` no decodifique cada snapshot completo. Hoy son 62 MB en 117 releases, y el comando respondió bien sobre el estado real.

**Revisión del plan** (2026-09-27): versión revisada `184e0ee0ebf4`. Dos revisores `review-plan` en paralelo, despachados como rol nativo de Claude Code, de solo lectura.

- **Contratos del gestor y pruebas.** Dos bloqueantes, los dos corregidos con la propuesta del revisor:
  - Un commit con el mismo contenido nunca quedaba registrado, porque `Apply` devuelve «unchanged» antes de escribir. Ahora se registra también en ese caso.
  - Los CLIs y las releases se tomaban de `installations`, que pierde los CLIs tras `--release`. Ahora se usan `RegisteredHosts` y `Records`.

  Se incorporaron las sugerencias sobre el nombre `LastWrittenAt`, el punto exacto de escritura en `Apply`, el catálogo propio de las pruebas de `tooling/cli`, el helper de salida de T1, la función de ayuda y la redacción de AC7.
- **Git, extracción y seguridad.** Un bloqueante, corregido con la opción recomendada por el revisor: `Extract` rechaza la cabecera pax global y los directorios con barra final de `git archive`. Ahora un adaptador en `tooling/cli` normaliza el tar. Se incorporaron también:
  - borrar la carpeta `.hive-extract-*` entera;
  - `--source .` por defecto y el caso de Git ausente;
  - pedir `tar` sin comprimir;
  - limpiar las variables `GIT_*` heredadas;
  - leer con un límite de tamaño;
  - elegir los CLIs antes de ejecutar Git;
  - «every return path» en la especificación.
- **Descartado:** `--end-of-options` en `rev-parse`. El revisor confirmó que el rechazo del `-` inicial basta con el sufijo `^{commit}`.
- **Sin nueva ronda:** todas las correcciones aplican propuestas de los propios revisores. Límite pendiente: no se comprobó el comportamiento con Git anterior a 2.38.

**Avance:** T1 a T5 verificadas. `go vet ./...` y `go test -race ./...` pasan sobre `24759bb`, y `go test -race` de las pruebas afectadas pasa sobre `f6da52d`. Falta la revisión de código con `/code-review`.

**Siguiente paso:** `/code-review` sobre el diff contra `18847d6`, rebase sobre `origin/rebuild/harness-engineering` y push directo (D7-A).
