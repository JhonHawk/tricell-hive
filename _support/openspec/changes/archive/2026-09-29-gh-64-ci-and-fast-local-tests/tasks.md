# Tareas

Base del cambio: `a5f3d1f` (`origin/development`, creada el 2026-09-29 desde `origin/rebuild/harness-engineering`). Worktree: `.claude/worktrees/gh-64`, rama `feat/gh-64-ci-and-fast-local-tests`.

## T0 — Rama `development` publicada

- [x] `origin/development` existe y apunta a la punta de `origin/rebuild/harness-engineering`. Evidencia: push `origin/rebuild/harness-engineering -> development` en `a5f3d1f`; `git merge-base --is-ancestor` salió con 0. La comprobación antes del merge sigue pendiente.

**Closes:** AC6 (la parte de la rama).

**Depends on:** ninguna.

**Locations:** remoto `origin`.

**Execution:** hilo principal, porque es una escritura remota que autorizan D4-A y D5-A.

**Test approach:** check: `git merge-base --is-ancestor origin/rebuild/harness-engineering origin/development`.

**Changes:**
- `git fetch origin` justo antes del push, y después `git push origin origin/rebuild/harness-engineering:refs/heads/development`, sin tocar `rebuild/harness-engineering`.
- Antes del merge (paso 4 de la entrega), `git fetch origin` y `git rev-list origin/development..origin/rebuild/harness-engineering`. Si devuelve commits, otra sesión entregó en `rebuild` después de T0: se integran en `development` antes del merge y se informa cuáles fueron.

**Verification:** el comando de check sale con 0 tras el push y otra vez antes del merge, y el `rev-list` de ese momento está vacío.

## T1 — Pruebas sin fsync en `management` y `cli`

- [x] Los binarios de prueba de `tooling/management` y `tooling/cli` corren sin fsync, y el binario real sigue sincronizando. `hive-verify-task` (Opus) dio por cumplidos AC1 y AC2:
  - tiempos: `management` 9,43 s y `cli` 141,5 s, cada paquete solo;
  - la prueba `TestDiskSyncSwitch` falla con la condición invertida;
  - `.Sync()` solo aparece dentro de `syncFile` (`files.go:136`).

  Evidencia del implementador: antes del cambio la compilación fallaba con `undefined: diskSync` y después pasa; con la condición invertida la prueba falla en `sync_test.go:34`.

**Closes:** AC1, AC2.

**Depends on:** T0.

**Locations:** `tooling/management/files.go` (`write`), `tooling/management/resources.go` (reemplazo de enlace, alrededor de la línea 162), un `TestMain` nuevo en `tooling/management`, `tooling/cli/tui_hosts_test.go:80` (`TestMain`).

**Execution:** delegada a `hive-build-backend`: la interfaz está decidida en [design.md](design.md#opción-de-prueba-para-fsync-d1-a) y no se cruza con otras escrituras.

**Test approach:** tdd. Primero la prueba de AC2, que falla porque no hay opción; después la opción y los `TestMain`.

**Changes:** los que describe el diseño. Ningún cambio de comportamiento fuera de las pruebas.

**Verification:**
- `go test -count=1 -run '<prueba de AC2>' -v ./tooling/management` pasa. Hay que comprobar que falla si `syncFile` ignora la variable: invertir la condición a propósito, confirmar el fallo y deshacer el cambio.
- `rg -n 'DisableDiskSyncForTests' --glob '!*_test.go' tooling` devuelve solo la definición y su comentario.
- `go test -count=1 ./tooling/management` tarda 25 s o menos y `go test -count=1 ./tooling/cli` 170 s o menos, corriendo cada paquete solo, en esta máquina (AC1).
- `go vet ./tooling/...` sin hallazgos.

## T2 — La prueba de flujos se salta sin Node

- [x] `TestNewFlowCasesHaveDistinctFixturesAndContracts` se salta cuando falta `node`. `hive-verify-task` (Sonnet) dio AC3 por cumplido: reprodujo el fallo en una copia limpia de `a5f3d1f` y el salto en `flows_test.go:168-170`. Evidencia de la implementación:
  - antes del cambio: sin Node, la prueba falla con `FAIL tricell-hive/tests/pilot`;
  - después: sin Node el paquete pasa (`ok … 6.174s`) y la prueba aparece como `SKIP` en `flows_test.go:169`;
  - con Node pasa completa.

**Closes:** AC3.

**Depends on:** T0.

**Locations:** `tests/pilot/flows_test.go:147`, con el salto justo antes de `dir := t.TempDir()` (línea 164).

**Execution:** hilo principal, porque son dos líneas y el brief sería más largo que el cambio.

**Test approach:** tdd. La corrida sin Node falla en la base y pasa después.

**Changes:** `exec.LookPath("node")` y `t.Skip("Node unavailable")` antes de la parte que usa Node, como en `flows_test.go:108`. La comprobación de `cases.json` sigue corriendo sin Node.

**Verification:** con `P="$(dirname "$(command -v go)"):/usr/bin:/bin"`, tras confirmar que esas rutas no traen `node`:
- `env PATH="$P" go test -count=1 ./tests/pilot/` pasa con el paquete entero (AC3).
- `env PATH="$P" go test -count=1 -run 'TestNewFlowCasesHaveDistinctFixturesAndContracts' -v ./tests/pilot/` muestra `SKIP`.
- Con Node en el `PATH`, la prueba corre completa y pasa.

## T3 — Workflow de CI

- [x] `.github/workflows/ci.yml` según [design.md](design.md#workflow-de-ci-d2). `hive-verify-task` (Sonnet) dio por cumplida la parte de AC4 anterior al merge. Después del merge en `31cdd43`, `gh run list --workflow CI --event push` devolvió 0 corridas. El tiempo de 27 min supera los 20 min previstos, así que queda como hallazgo para #60 y #61. La [corrida 36673081482](https://github.com/JhonHawk/tricell-hive/actions/runs/36673081482) del PR #69 terminó en verde:
  - duración total: 27 min 13 s;
  - `go vet`: 32 s;
  - `go test -race`: 26 min 14 s;
  - `unittest`: 1 s.

  El archivo ya está escrito:
  - usa `checkout@v7`, `setup-go@v7` (`go-version-file`) y `setup-node@v7` (Node 24, `package-manager-cache: false`); las etiquetas `v7` existen en los tres repositorios y las entradas coinciden con sus README;
  - el YAML se lee bien.

  La verificación espera al PR.

**Closes:** AC4.

**Depends on:** T0; su verificación final depende del PR (paso 3 de la entrega).

**Locations:** `.github/workflows/ci.yml` (nuevo).

**Execution:** hilo principal, porque su verificación es el propio PR que abre el hilo principal.

**Test approach:** check: el PR de este cambio dispara el workflow.

**Changes:** los que describe el diseño. Antes de escribir, confirmar con la documentación oficial las entradas de `actions/setup-go@v7` (`go-version-file`) y la sintaxis de `paths-ignore` y `concurrency`.

**Verification:**
- En el PR, `gh pr checks <PR>` muestra `CI` en verde. Los logs muestran los pasos `go vet`, `go test -race` y `unittest` completos, con el tiempo total anotado en este archivo.
- `gh run list --workflow CI --event push` no devuelve corridas después del merge.

## T4 — Documentación de verificación y rama base

- [x] La verificación documentada y la rama base declarada coinciden con [design.md](design.md#verificación-local-y-rama-base-d4-a). `hive-verify-task` (Sonnet) dio por cumplidos AC5 y la parte de AC6 que toca `AGENTS.md`.
  - Matiz del verificador: solo `deployment-manager.md` nombra a la vez la excepción de concurrencia y la CI. `README.md` enlaza a esa sección e `installer.md` nombra la CI.
  - Lo acepté porque la sección enlazada es el único lugar donde vive esa regla.

  Evidencia de la implementación:
  - `go test -race` solo aparece acotado a la CI o a cambios de concurrencia, en `deployment-manager.md:194` e `installer.md:130`;
  - `go vet ./...` y `go test ./...` aparecen en los tres archivos;
  - `AGENTS.md:8` dice `Base branch: development`;
  - `tests/content` y las pruebas de Python pasan.

**Closes:** AC5, AC6 (la parte de `AGENTS.md`).

**Depends on:** T0.

**Locations:** `README.md:38`, `_support/docs/architecture/deployment-manager.md` (sección «Verification»), `_support/docs/architecture/installer.md:130`, `AGENTS.md:8`, `AGENTS.md:54`.

**Execution:** hilo principal, porque es un cambio mecánico de documentación.

**Test approach:** check con `rg`.

**Verification:**
- `rg -n 'go test -race' README.md _support/docs/architecture/deployment-manager.md _support/docs/architecture/installer.md` solo devuelve líneas que acotan `-race` a concurrencia o a la CI, incluida la de `deployment-manager.md:189` con `-timeout`.
- `rg -c 'go vet \./\.\.\.' <los tres archivos>` y `rg -c 'go test \./\.\.\.' <los tres archivos>` devuelven al menos 1 en cada archivo, y en ellos esa es la verificación local por defecto.
- `rg -n 'rebuild/harness-engineering' AGENTS.md` solo devuelve la mención de la rama congelada, si se deja alguna.
- `go test -count=1 ./tests/content/...` y las pruebas de Python pasan.

## Verificación y revisión

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Pruebas por tarea | Al cerrar cada tarea, en local | Los comandos de cada tarea | Parte de la implementación |
| Suite local sin `-race` | Una vez sobre el candidato final, antes del push | `go vet ./...` y `go test ./...` en verde, con el tiempo total anotado | Parte de la implementación |
| Revisión de código | Antes del push | `/code-review` de Claude Code sobre el diff (D6-A); cada hallazgo resuelto o refutado con evidencia | Autorizada (D6-A) |
| CI del PR | Tras el push | `CI` en verde en el PR, con `-race` en Linux | Autorizada (D5-A) |
| Verificación independiente por tarea | T1 y T3 | `hive-verify-task` contra sus `AC<n>` | Parte de `flow-build` |

Este cambio no tiene efecto visible en la interfaz ni un recorrido local para validar a mano. Su resultado se observa en los tiempos, en las pruebas y en la CI.

## Estado de la revisión y avance

Plan escrito el 2026-09-29 sobre `a5f3d1f`. Se revisó una ronda con dos `hive-review-plan` nativos de Claude Code, en paralelo y de solo lectura, sobre la versión con prefijos de hash `a7591ace11a0` (proposal), `d629b4f0e6bc` (design) y `522640f71ddd` (tasks):

- **Backend (T1 y T2, AC1 a AC3):** sin hallazgos bloqueantes. Aceptados:
  - N1: el helper `syncFile` y la aserción con archivo cerrado, para que AC2 falle si la opción está mal conectada;
  - N2: el salto por Node va antes de la parte que lo usa, no al inicio de la prueba;
  - N3: la verificación de T2 corre el paquete entero sin `-run`.
- **CI y entrega (T0, T3 y T4, AC4 a AC6):** aceptados:
  - B1: `fetch` antes del push, comprobación de commits nuevos en `rebuild` antes del merge, y AC6 verificado por ancestría y no por SHA idéntico;
  - B2: el patrón de T4 cubre `-race -timeout` y comprueba que `go vet ./...` y `go test ./...` aparecen;
  - N1: `setup-node` con Node 24;
  - N2: el PR de cierre salta la CI a propósito;
  - N3: `--base development` explícito;
  - N4: actualizar las memorias locales que nombran la base.

Todas las correcciones aplican la propuesta del propio revisor, así que no hubo segunda ronda.

Suite local sin `-race` sobre el candidato: `go vet ./...` y `go test -count=1 ./...` en verde en 2 min 23 s. `tooling/cli` tardó 142,6 s y `tooling/management` 11,7 s.

`/code-review` de Claude Code (D6-A) devolvió ocho hallazgos.

Corregidos:
- **F3:** el trabajo de CI tiene 55 min, para que venza antes el `-timeout 40m` de Go.
- **F5:** la variable pasó a ser `skipDiskSync atomic.Bool`.
- **F6:** la prueba `TestDisableDiskSyncOnlyCalledFromTests` reemplaza la comprobación manual con `rg`. Falla si un archivo que no es de prueba llama a la función; lo comprobé con un archivo de sondeo temporal.
- **F7:** `tests/fixtures/regression/README.md:35` ya no pide `-race`. No se tocó ningún caso.
- **F8:** la documentación nombra el comando exacto de la CI.

Después de los arreglos: `go test -race ./tooling/management` pasa en 50 s, y `tests/pilot` y `tests/content` pasan.

No aplicados, porque contradicen decisiones del usuario o el diseño:
- **F1:** la CI no corre en push a `development`, así que el resultado integrado no se prueba. Es la decisión D2; se le informa al usuario.
- **F2:** riesgo de agotar el cupo de minutos. Se aceptó con la estimación de D2; la primera corrida da el dato real.
- **F4:** `paths-ignore` no reporta ningún check. Es intencional: no hay checks requeridos y el cierre lo tiene previsto.

Modelos de la construcción. El hilo principal corre en Claude Opus 5.5, observado.

| Tarea | Implementa | Verifica |
| --- | --- | --- |
| T1 | `hive-build-backend`, Sonnet (configurado, no observado) | `hive-verify-task`, Opus (configurado, no observado) |
| T2 y T4 | hilo principal, Opus | `hive-verify-task`, Sonnet (forzado con la opción de modelo del lanzamiento) | Límites que siguen abiertos:

- no se midió el tiempo de `-race` en Linux con 2 CPU; lo mide la primera corrida del PR;
- las mediciones de AC1 son de esta máquina.

Cierre (2026-09-29): el PR #69 se integró en `development` con el merge `31cdd43`, sin commits nuevos en `rebuild` antes del merge. El push del merge no disparó la CI. `hive` se recompiló desde `31cdd43`: `hive doctor` muestra los seis CLIs verificados, y no hizo falta `hive update` porque `content/` no cambió. #62 y #64 se cerraron con un comentario. La memoria local sobre actualizar la base ya nombra `development`.
