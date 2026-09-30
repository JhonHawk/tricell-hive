# Diseño

## Contexto verificado

Base de las observaciones: `3530d13` (`development`), del 2026-09-30.

- **Qué hace lenta la CI.** Corrida 36673081482 (`ubuntu-latest`, 2 CPU): `go test -race` 26 min 14 s, de los que `tooling/cli` ocupa 1464 s.
- **Reparto en local con `-race`** (M4 Max, `go test -race -json`): 839 s en total.
  - 19 pruebas que instalan el árbol real: unos 676 s, entre 32 y 68 s cada una frente a ~1,5 s sin `-race`.
  - La subprueba `no_CLI_hosts`: 63 s.
  - Unas 360 pruebas más: unos 100 s.
- **Qué afirman esas pruebas** (investigación de un subagente, con las afirmaciones clave comprobadas): el asistente de instalación, los planes que caducan, la recuperación, el bootstrap y los textos, sin datos del contenido real. La única que lee contenido real es `TestModelsViewFitsAndScrollsWithTheFullCatalogueOnSixCLIs`, que copia el front matter de los agentes reales y exige al menos 20 roles (`tui_models_view_test.go:388-420`).
- **Por qué pesan.** Planificar el árbol real con `-race` tarda 0,74 s (`TestRepositoryCatalogueInstructionReferences`). Lo caro es escribir y calcular el hash de todos los archivos al aplicar, multiplicado por el número de CLIs. Con una fuente mínima (`global.md` y una voz) esa escritura es de unos pocos archivos.
- **Causa de la espera de 60 s**, verificada con un parche temporal que se deshizo después. `no_CLI_hosts` escribe unos 110 caracteres antes de que la vista se dibuje por primera vez. Hasta ese primer dibujo, el cursor intermitente del campo de texto genera comandos que el controlador ejecuta de forma síncrona, unos 530 ms cada uno. Con un `mustShow` después de `openMenuEntry`, la subprueba bajó de 61,07 s a 0,53 s.
- **Usos del árbol real en `tooling/cli`** (`rg`):
  - `filepath.Abs("../..")` en 20 lugares: `install_test.go` (8, incluido `installArgs:110`, con 12 llamadas), `bootstrap_test.go` (4: `bootstrapFixtureOrigin:39`, `newBootstrapFixture:338`, `:597`, `:638`), `provider_adapter_test.go` (4), `tui_test.go:35` (`interfaceTestOptions`, en la que se apoya `testAppConfig`, con 47 llamadas) y `:1012`, `main_test.go:50` y `voice_test.go:67`.
  - Los fixtures de bootstrap copian `content/` completo y `integrations/agent-profiles.json` en un tarball (`bootstrap_test.go:43-47`).
- **Fuente mínima actual.** `minimalCatalogSource` (`tui_hosts_test.go`) usa `sync.OnceValues`, `os.MkdirTemp` y `EvalSymlinks` (porque `target.Safe` rechaza un ancestro enlazado), y `TestMain` la borra. La usan 51 llamadas a `minimalTestSource`.
- **La prueba sobre contenido real que se conserva.** `TestRepositoryCatalogueInstructionReferences` (`tooling/management/references_test.go:60`) planifica el catálogo real y comprueba sus referencias, pero no aplica.

## Prácticas de Go aplicadas (fuentes consultadas el 2026-09-30)

- **`testdata/`.** `go help test` (Go 1.27): «The go tool will ignore a directory named "testdata", making it available to hold ancillary data needed by the tests.» Por eso la fuente mínima pasa a archivos fijos en `tooling/cli/testdata/` (D2-A).
- **`sync.Once` y `TestMain`.** La [guía de estilo de Google](https://github.com/google/styleguide/blob/gh-pages/go/best-practices.md) acepta `sync.Once` solo si la preparación es cara, afecta a algunas pruebas y no necesita limpieza. Añade que un `TestMain` propio «should not be your first choice». La fuente generada incumplía la primera regla, porque necesitaba limpieza. Con `testdata` desaparecen el `Once` y la limpieza. `TestMain` sigue solo para `DisableDiskSyncForTests`, un ajuste de todo el proceso que debe aplicarse antes de cualquier prueba, que es el caso documentado en `go doc testing` («# Main»).
- **Ganchos exportados solo para pruebas.** La misma guía propone paquetes de ayuda aparte (`footest`) para los dobles de prueba. No habla de ganchos exportados en particular, así que esto es una lectura prudente, no una regla literal. `go doc testing.Testing` dice: «reports whether the current code is being run in a test». D3-A añade ese chequeo, para que el gancho no pueda desactivar fsync en el binario `hive`.
- **Ayudantes.** La [guía de comentarios de pruebas de Go](https://go.dev/wiki/TestComments) pide `t.Helper()` en los ayudantes que reciben `*testing.T`. El arreglo de #59 va en el ayudante `openMenuEntry` y no en una sola subprueba.
- **`-race`.** El [artículo del detector de carreras](https://go.dev/doc/articles/race_detector) indica un costo de 2 a 20 veces más tiempo y recomienda correr las pruebas con `-race`. Se mantiene en la CI; lo que se reduce es lo que esas pruebas escriben.
- **Context7** (`/golang/go`) solo devolvió fragmentos de código sin guía normativa. La guía se tomó de `go doc`, de `go help` y de las fuentes enlazadas.

## Enfoque

### Fuente mínima fija (D2-A)

- `tooling/cli/testdata/minimal-source/` contiene lo que genera hoy `minimalCatalogSource`: `content/guidance/global.md`, `content/voices/preamble.md` y `content/voices/testvoice.md`. Incluye además `integrations/agent-profiles.json` con un JSON sintético mínimo, porque la verificación del manifiesto de bootstrap lo exige (`tooling/distribution/manifest.go:123`). `BuildPlan` solo lee los perfiles si hay agentes (`plan.go:161`).
- Excepción conocida: `TestInstallSharedHostClosureRequiresConsent` (`install_test.go:255`) exige un recurso compartido `.agents`, que solo aparece si la fuente trae skills (`install.go:966`). Esa prueba copia la fuente mínima a `t.TempDir()` y le añade una skill de una línea, como hace `tui_voice_view_test.go:265`. No se añade una skill a la fuente compartida, porque cambiaría las 51 llamadas.
- `minimalTestSource(t)` devuelve la ruta absoluta de ese directorio, sin `EvalSymlinks`: la ruta del checkout no pasa por un enlace. Si alguna prueba modifica la fuente, copia el árbol a `t.TempDir()` con un ayudante propio, en vez de escribir en `testdata`.
- `TestMain` en `tui_hosts_test.go` queda con `DisableDiskSyncForTests()` y `m.Run()`, sin limpieza.

### Pruebas que se mueven

- `installArgs`, `interfaceTestOptions`, las llamadas de `provider_adapter_test.go`, `main_test.go:50`, `tui_test.go:1012` y las de `install_test.go` usan `minimalTestSource(t)`.
- Si una prueba afirma algo que depende de lo que trae la fuente, se ajusta la afirmación a la fuente mínima sin perder el comportamiento que prueba. Un ejemplo sería el número de archivos o una línea con un recurso concreto. Cada ajuste así queda registrado en `tasks.md`.
- Los dos fixtures de bootstrap empaquetan la fuente mínima en lugar de `content/` completo.
- `TestModelsViewFitsAndScrollsWithTheFullCatalogueOnSixCLIs` genera unos 22 roles sintéticos en `t.TempDir()`. El umbral de desplazamiento en 80×24 anda por 17, porque la vista resta `modelsFixedRows = 6` (`tui_models_view.go:19,83`). Entre esos roles va el nombre más largo real (`hive-design-architecture`, 24 caracteres) y la mezcla de perfiles razonamiento, ejecución y hereda. La prueba pasa a exigir que `v.box.scrollable()` sea verdadero en 80×24: hoy la parte de desplazamiento va dentro de un `if` (`tui_models_view_test.go:444`) y se saltaría sin avisar. Sigue leyendo `agent-profiles.json` real, que es un dato fijado que D1-A deja en #61. Deja de exigir «al menos 20 roles» y exige el mínimo sintético. Su nombre puede cambiar para no prometer el catálogo completo.

### Una sola prueba sobre el contenido real

`TestRepositoryCatalogueInstructionReferences` pasa a aplicar el plan del catálogo real en un home temporal de un solo CLI. El ayudante `setup(t)` instala en `codex` y `claude` (`management_test.go:37`), así que la prueba reduce `o.Hosts` a uno. Después comprueba, usando `PlanUnchanged` (`voice_fix_test.go:408`):

- que un segundo plan no tiene cambios;
- que existen la guía global, una skill y un agente instalados.

Así, una sola prueba cubre la validación del contenido y su instalación real. Se espera que con `-race` tarde unos 30 s, en `tooling/management`, que corre en paralelo con `tooling/cli`.

### Chequeo de binario de prueba (D3-A)

`DisableDiskSyncForTests` comprueba un predicado de paquete, `isTestBinary = testing.Testing`, y si es falso entra en pánico con «DisableDiskSyncForTests called outside a test binary». La prueba sustituye el predicado por uno falso, comprueba el pánico con `recover` y restaura el predicado y `skipDiskSync`.

Importar `testing` en código de producción no registra sus flags, porque desde Go 1.13 los registra `testing.Init` en el binario de prueba. El revisor del plan lo comprobó en Go 1.27: un binario que importa `testing` registra 0 flags y `testing.Testing()` devuelve `false`. T3 lo verifica otra vez con `hive --help`.

### #59

`openMenuEntry` termina con un dibujo de la vista, `d.screen()` o un `mustShow` del título de la vista abierta, antes de devolver el control. Así ninguna prueba escribe en una vista sin dibujar. Otros ayudantes que abren vistas (`env.openApp`) ya dibujan antes de escribir.
