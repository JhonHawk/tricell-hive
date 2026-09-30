# Diseño

## Contexto verificado

Base de las observaciones: `a5f3d1f`, el 2026-09-29.

- **Dónde se sincroniza.** `tooling/management` llama a fsync en tres puntos. En `write` sincroniza el archivo temporal (`files.go:100`) y el directorio después del rename (`files.go:120`). En el reemplazo de un enlace gestionado sincroniza el directorio (`resources.go:162`). `tooling/cli` y `tests/pilot` escriben a través de `management`.
- **Mediciones sin `-race`, en la misma máquina (macOS, Apple Silicon).** En una copia aparte con los tres fsync convertidos en no-ops, `tooling/management` bajó de 52,7 s a 10,6 s y `tooling/cli` de 204,0 s a 140,9 s, con todas las pruebas en verde. `tests/pilot` tarda 5,9 s y `tests/content` 0,25 s. Observación: `tooling/cli` sin fsync corrió junto a `tooling/management`, así que 140,9 s es una cota alta.
- **Por qué solo en macOS.** En `darwin`, `File.Sync` de Go usa `F_FULLFSYNC`. En Linux (Docker, `golang:1.27`), `tooling/management` tarda 10,7 s con fsync, así que ese costo es de macOS. Esto es inferencia a partir de las mediciones; no se consultó la documentación de Go.
- **Suite en Linux.** Pasa entera salvo `tests/pilot` `TestNewFlowCasesHaveDistinctFixturesAndContracts`. Esa prueba (`flows_test.go:147`, falla en la línea 173) necesita `node` y no se salta sin él, mientras que `TestTypeScriptFixturesAndIndependentContract` (`flows_test.go:107`) sí se salta. Las máquinas `ubuntu-latest` de GitHub traen Node, pero la prueba no debe depender de eso.
- **`TestMain` existentes.** `tooling/cli` ya tiene uno en `tui_hosts_test.go:80`, que borra la fuente mínima de catálogo al terminar. `tooling/management` no tiene ninguno.
- **Pruebas de Python.** `python3 -m unittest discover -s tests/skills -p '*_test.py'` corre 23 pruebas en menos de 1 s. `harness_audit_rules_test.py` lee `_support/docs/harness-engineering/harness-audit-rules.md`.
- **GitHub.** El repositorio es privado y la cuenta es Pro: 3000 minutos de Actions al mes en repositorios privados. `ubuntu-latest` tiene 2 CPU y 8 GB en repositorios privados ([facturación](https://docs.github.com/en/billing/concepts/product-billing/github-actions), [máquinas](https://docs.github.com/en/actions/reference/runners/github-hosted-runners), consultado el 2026-09-29). Las Actions están habilitadas y `.github/` no existe. Últimas versiones: `actions/checkout` v7.0.1 y `actions/setup-go` v7.0.0 (releases de GitHub, consultado el 2026-09-29).
- **Ritmo de entrega.** `rebuild/harness-engineering` recibe de 22 a 61 commits diarios en primer nivel (del 23 al 29 de septiembre).
- **Referencias a la rama base.** Fuera de los registros archivados y las sesiones, `rebuild/harness-engineering` aparece solo en `AGENTS.md:8` y `AGENTS.md:54`.

## Enfoque

### Opción de prueba para fsync (D1-A)

En `tooling/management`, una variable de paquete que vale `true` por defecto decide si se sincroniza. Los tres puntos de fsync pasan por un único helper, `syncFile(f *os.File) error`, que consulta esa variable y llama a `f.Sync()` solo si vale `true`. Sirve tanto para archivos como para directorios. Una función exportada, de nombre explícito como `DisableDiskSyncForTests()`, la pone en `false`. La llaman los `TestMain` de `tooling/management` (nuevo) y `tooling/cli` (el existente, antes de `m.Run()`). Como se asigna antes de que empiece cualquier prueba, no hay carrera, incluso si #60 introduce `t.Parallel` más adelante.

Las alternativas descartadas son dos. Una variable de entorno, porque el binario real la leería y un usuario podría desactivar la durabilidad sin querer. Una etiqueta de compilación, porque obligaría a pasar `-tags` en cada invocación y en la CI.

La prueba de AC2 guarda el valor por defecto en una variable de paquete del archivo `_test`, que Go inicializa antes de `TestMain`, y comprueba que era `true`. Después pasa a `syncFile` un archivo ya cerrado:

- con la sincronización activa, debe devolver `os.ErrClosed`;
- con la sincronización desactivada, debe devolver `nil`.

Así la prueba falla si la variable se ignora o se lee al revés. Además, con la sincronización activa solo durante esa prueba (que no usa `t.Parallel`), ejecuta una escritura y un reemplazo de enlace con `writeResource` y un destino `symlink`, partiendo de `target.Canonical(t.TempDir())`, como hace `management_test.go:33`. Al terminar vuelve a desactivarla. La CI con `-race` tampoco ejercita el fsync real fuera de esta prueba, y eso es aceptable. Que ningún código de producción llame a la función lo comprueba una prueba, `TestDisableDiskSyncOnlyCalledFromTests`: recorre los archivos `.go` que no son de prueba en `integrations/`, `tests/` y `tooling/`. Tras `/code-review`, la variable pasó a ser `skipDiskSync atomic.Bool`, cuyo valor cero mantiene la sincronización. Así, cambiarla dentro de `TestDiskSyncSwitch` no provoca una carrera de datos si #60 introduce pruebas en paralelo.

Pregunta abierta de #62: ¿hay alguna prueba de recuperación que dependa de que fsync ocurra antes de un punto de falla simulado? No. fsync no cambia lo que otro proceso ve en el sistema de archivos, solo lo que sobrevive a una caída real, que ninguna prueba provoca. Además, la copia aparte sin fsync pasó todas las pruebas de `management` y `cli`.

### Prueba de flujos sin Node

`TestNewFlowCasesHaveDistinctFixturesAndContracts` (`flows_test.go:147`) se salta con `t.Skip` cuando `exec.LookPath("node")` falla, igual que su prueba hermana. La comprobación va justo antes de `dir := t.TempDir()` (línea 164) y no al inicio de la prueba: la parte que lee `cases.json` y exige los nueve casos de flujo no necesita Node y debe seguir corriendo. Esto no modifica ningún caso de `tests/fixtures/regression/`.

### Workflow de CI (D2)

Archivo `.github/workflows/ci.yml`, con nombre `CI`:

- Disparo: solo `pull_request` hacia `development`, con `paths-ignore` de `_support/openspec/**` y `_support/sessions/**`, que son registros de cambio y de sesión que ninguna prueba lee. Sin `push`.
- `concurrency`: un grupo por número de pull request con `cancel-in-progress: true`, para que un push nuevo al mismo PR cancele la corrida anterior.
- `permissions: contents: read`.
- Un solo trabajo en `ubuntu-latest` con `timeout-minutes: 55`, para dejar margen a la preparación y la compilación con `-race` antes de que venza el `-timeout 40m` de Go. Pasos:
  1. `actions/checkout@v7`.
  2. `actions/setup-go@v7` con `go-version-file: go.mod`.
  3. `actions/setup-node@v7` con `node-version: '24'`. Las pruebas de flujos ejecutan `.ts` con `node`, y la corrida en Linux del diseño se hizo sin Node; en local pasan con Node v24.19.0. Fijarla evita depender de la versión que traiga la imagen.
  4. `go vet ./...`.
  5. `go test -race -timeout 40m ./...`.
  6. `python3 -m unittest discover -s tests/skills -p '*_test.py'`.
- No hace falta configurar identidad de git, historial completo, `script` ni cgo. Las pruebas definen su identidad en repositorios temporales (`update_test.go:47`), `script` viene con `util-linux` y `-race` usa el cgo de la imagen.
- Si el PR tiene conflictos, GitHub no crea el commit de merge y el workflow no corre. Hay que resolver los conflictos antes de esperar la CI.

El límite de 40 min es una estimación hasta tener la primera corrida: con `-race` y 2 CPU, `tooling/cli` no está medido. La primera corrida del PR da el tiempo real, que se anota en `tasks.md`. Si pasa de 20 min, se registra como hallazgo para #60 y #61.

### Verificación local y rama base (D4-A)

- `README.md:38` y la sección de verificación de `deployment-manager.md` e `installer.md` dejan `go vet ./...` y `go test ./...` como verificación local. `-race` queda en local solo cuando el cambio toca concurrencia (goroutines, locks, estado compartido o las pruebas de escritura concurrente del gestor), y en la CI de cada PR. Se conserva la nota del `-timeout`, que solo aplica a `-race`.
- `AGENTS.md`: `Base branch: development`. La excepción temporal de refresco pasa a decir «each commit or merged pull request that lands on `development`», y conserva su condición de retiro (la integración en `master`).

Con la CI, `flow-build` cierra la cobertura de regresión en la CI del PR y no en una corrida local completa (`content/skills/flow-build/SKILL.md:65`). El agente corre en local las pruebas de los paquetes afectados.
