# Tareas

**Commit base:** `18847d6`. El 2026-09-27 otra sesión trabajaba en `rebuild/harness-engineering`, así que la construcción va en el worktree `.claude/worktrees/gh-46`, rama local `feat/gh-46-update-releases`. Al final se rebasa sobre `origin/rebuild/harness-engineering` y se publica directo a esa rama (D7-A).

Los criterios de AC1 a AC7 no dependen del commit base: `update`, `releases` y `commits.json` tampoco existen en `18847d6`.

## Orden y ejecución

T1 y T2 pueden correr en paralelo, porque escriben en paquetes distintos (`tooling/cli` y `tooling/management`). T3 y T4 dependen de las dos y los hace un solo hijo en secuencia, porque ambas tocan `tooling/cli/main.go`. T5 va al final en el hilo principal.

Cada hijo recibe `AGENTS.md`, este cambio (`proposal.md`, `design.md` y `tasks.md`) y [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md) (solo T3). Puede hacer commits locales en su rama de trabajo solo si la entrega elegida lo permite.

### T1 — Despacho de subcomandos por función (#42)

- [ ] `run` delega cada subcomando en su propia función, con el comportamiento visible sin cambios.

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

- [ ] `Plan` lleva el commit de origen, `Apply` lo registra junto a la release y `Releases` lista los snapshots.

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

- [ ] `hive update` actualiza desde un commit, según [el diseño](design.md#hive-update).

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

- [ ] `hive releases` imprime en JSON el resultado de `management.Releases`.

**Closes:** AC7.

**Depends on:** T1, T2.

**Locations:** `tooling/cli/main.go` (función del subcomando y `--help`), `tooling/cli/main_test.go`.

**Execution:** delegada, el mismo hijo que T3, después de T3.

**Test approach:** tdd.

**Changes:** una prueba con un estado sintético de dos releases, una instalada y otra con registro de commits, que comprueba la salida JSON y el orden.

**Verification:** `go test ./tooling/cli` pasa.

### T5 — Documentación y comprobación local sobre el estado real

- [ ] `deployment-manager.md` documenta los dos comandos, y las comprobaciones de solo lectura sobre el estado real pasan.

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

**Avance:** plan listo para implementar; ninguna tarea empezada.

**Siguiente paso:** la pregunta de entrega y, después, `flow-build`.
