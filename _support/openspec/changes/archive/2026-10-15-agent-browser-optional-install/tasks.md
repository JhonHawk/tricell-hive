# Tareas

Base del cambio: `d8ee7a60` (2026-10-13), la punta de `feat/flash-class-guidance` tras fusionar `feat/pi-subagents-unversioned-install` (P1-A del usuario, que reemplaza el orden de O1-A y D-12; `feat/flash-class-guidance` ya contenía `development` actual). Comprobado al iniciar: `providers.Catalog()` sin `agent-browser`, `ValidateStep` rechaza todo paso no manual (`tooling/providers/run.go:25`), `showPartialOnboardingDetail` solo muestra `ManualReason`, `doctor_integrations.go:236-253` dice «Hive does not install it» y `SECURITY.md:32` «Hive does not install CLI executables»: AC1 y AC3–AC7 son falsos en la base.

T1–T3 tocan código Go acoplado (tipos → adaptador y ejecutor → doctor) y van juntas a un hijo `hive-build-backend`. T4 es texto y va en el hilo principal en paralelo, sin compartir archivos. T5 es el gate nativo, solo con autorización.

### T1 — Receta validable en el catálogo de proveedores

- [x] `agent-browser` al final del catálogo, con campos de paso `omitempty` y validación de `argv` ligada al ID.

**Closes:** AC7.

**Depends on:** ninguna.

**Locations:** `tooling/providers/{types,catalog,run}.go`, `providers_test.go` (`:17`), `run_test.go` (`:28`).

**Execution:** delegada a `hive-build-backend` junto con T2 y T3.

**Test approach:** tdd.

**Changes:** D-1 y D-3 (paso `agent-browser-manual`).

**Verification:** `go test ./tooling/providers/...` pasa; pruebas nuevas: el catálogo termina en `agent-browser` para los seis hosts; `ValidateStep` acepta `agent-browser-install` con `["brew","install","agent-browser"]` o `["npm","install","-g","agent-browser"]`, `agent-browser-chrome` con `["agent-browser","install"]` y `agent-browser-manual`, y rechaza `argv` en el ID equivocado, `["sh","-c",…]`, `["brew","install","otra"]` y campos desconocidos; un payload manual antiguo sigue validando.

### T2 — Oferta sin versión, ejecutor, reconciliación y detalle de fallo

- [x] `hive install` ofrece la capacidad según D-3 y D-5 sin pedir versión, ejecuta los pasos con los límites de D-6, reconcilia cada paso según D-7, reintenta Chrome según D-8 y muestra el mensaje de fallo.
  Ronda 1, AC4: con `agent-browser` ya en el PATH y sin navegador, `planAgentBrowser` vuelve antes del filtro de plataforma (`tooling/cli/agent_browser.go:184-192`) y ofrece `agent-browser install` como ejecutable en plataformas no probadas; D-3 pide manual.

**Closes:** AC1, AC3, AC4.

**Depends on:** T1.

**Locations:** `tooling/cli/provider_adapter.go` (`Detect`, `Plan`, `Runner`, comentarios `:9-16`, `:86-89`), ejecutor nuevo en `tooling/cli` (patrón de `pi_package.go` para la variable sustituible y de `model_catalog.go:285-300` para la ejecución), `install.go` (`:643`, `:688`, `selectProviderRequests`, resumen `:784-796`, `showPartialOnboardingDetail` `:862-876`), `tui_hosts_view.go` (`:1024`, `:1105-1111`, `:1167`); pruebas en `provider_adapter_test.go` (`:50-80`, `:118-133`), `install_test.go` y archivos nuevos.

**Execution:** el mismo hijo que T1.

**Test approach:** tdd, sin procesos reales: la variable de ejecución registra `argv` y devuelve resultados simulados; `Detect` usa funciones sustituibles (D-9).

**Changes:** D-2 a D-9.

**Verification:** `go test ./tooling/cli/... ./tooling/management/...` pasa; pruebas nuevas, cada una con plataforma y PATH simulados:
- (a) macOS arm64 con `brew` y sin `agent-browser` → oferta sin «Exact version for», resumen con `brew install agent-browser`, «última versión» y aviso de Node;
- (b) Linux amd64 con Node 24 y prefijo escribible → `npm install -g agent-browser`;
- (c) navegador presente en una ruta de D-5 o `AGENT_BROWSER_EXECUTABLE_PATH` a un archivo → sin paso de Chrome; sin navegador → con paso de Chrome;
- (d) sin `brew`, Node 22, prefijo no escribible o Linux arm64 → capacidad manual con su motivo;
- (e) `agent-browser` presente y sin navegador → solo el paso de Chrome;
- (f) primer paso falla → segundo `skipped`; comando que falla o expira → `failed` con mensaje visible en el detalle del resultado, núcleo conservado;
- (h) interrupción sin resultado guardado → `RecoverOnboarding` deja `failed`, nunca `unknown`;
- (i) sesión no interactiva, `--dry-run` y `--home` sintético → ningún `argv` registrado y ningún programa ejecutado en `Detect` (tampoco `node --version` ni `npm config get prefix`);
- (k) salida de más de 64 KiB → el proceso no se cancela; se guarda recortada;
- (l) paso de Chrome con `agent-browser` ausente del PATH → `skipped` con motivo;
- (j) prueba del ejecutor real contra un programa de prueba que duerme e hereda tuberías → el grupo de procesos termina al expirar.

### T3 — Doctor y vista Integrations

- [x] Con `agent-browser` ausente o sin navegador, el doctor y la vista nombran `hive install`; Record muestra el peor estado de sus pasos; el doctor no ejecuta nada nuevo.

**Closes:** AC5.

**Depends on:** T1.

**Locations:** `tooling/cli/doctor_integrations.go` (`:23-26`, `:50`, `:118`, `:139-143`, `:241-254`, `:287`) con las funciones sustituibles de D-9, `doctor_integrations_test.go` (`:173`, `:341`, `:410-441`, `:465-480`, `:489`), `tui_integrations_view_test.go` (`:114`, `:121`, `:255`).

**Execution:** el mismo hijo que T1.

**Test approach:** tdd: las pruebas que fijan «Hive does not install it», «not part of onboarding» y la exclusión de comandos de instalación cambian primero a la nueva expectativa.

**Changes:** D-10 y D-11.

**Verification:** `go test ./tooling/cli/ -run 'Doctor|Integrations'` pasa; pruebas nuevas: fila ausente y fila sin navegador nombran `hive install`; con `-install` verificado y `-chrome` fallido, Record muestra el fallo.

### T4 — Especificación, seguridad y documentación

- [x] La especificación, `SECURITY.md` y los documentos describen la capacidad y su receta.

**Closes:** AC6.

**Depends on:** la integración de la rama de pi-subagents (O1-A), para escribir el delta sobre su texto.

**Locations:** delta `specs/versioned-installation/spec.md` en esta carpeta (MODIFIED: requisito del catálogo, `:62-66`; «Verified release package installation» con su escenario, `:76`, `:88`, sobre el texto de la rama de pi-subagents; requisito de diagnóstico, `:155`), `SECURITY.md:27-32`, `_support/docs/architecture/{installer,tool-dependencies,deployment-manager}.md`, `README.md` (`:127`, `:153-158`), `llms.txt:39`, `CHANGELOG.md` (`### Added`, con nota de versión menor según `CONTRIBUTING.md:52`).

**Execution:** hilo principal.

**Test approach:** check.

**Changes:** D-3, D-4, D-6, D-12 y D-13 en lenguaje de requisitos; `tool-dependencies.md` anota la versión sin fijar y el binario sin suma del postinstall de npm.

**Verification:** `rg -n -i -U "does not install CLI executables|No CLI executable is installed by Hive|limited to Engram, Context7,\s+and pi-subagents|no other host step downloads|only the Pi host's pi-subagents step may download|Hive does not install (it|them)|not part of the optional capability catalog|no\s+(automated\s+)?provider\s+recipes|never detects, downloads, installs|The one network\s+step|network-free except for this step|makes no network requests of its own" SECURITY.md README.md llms.txt _support/docs _support/openspec/changes/agent-browser-optional-install/specs` no devuelve frases vigentes (cada coincidencia restante se justifica en el progreso); `rg -n "agent-browser" SECURITY.md README.md llms.txt _support/docs/architecture/tool-dependencies.md _support/openspec/changes/agent-browser-optional-install/specs` devuelve las menciones nuevas; `go test ./tests/content/` pasa.

### T5 — Gate nativo

- [x] Las recetas funcionan en las plataformas probadas.

**Closes:** AC3, AC4 (comprobación real, además de las pruebas con dobles).

**Depends on:** T1–T3.

**Locations:** binario construido desde la rama. Linux: imagen de gate derivada de `tests/pilot/container/Dockerfile` (presente en `feat/flash-class-guidance`, `d345158a`) con Node 24, prefijo de npm escribible por UID 1000 y las bibliotecas de Chrome instaladas sin el paquete `chromium` (un `chromium` en el PATH suprimiría el paso de Chrome por D-5); la variante con Node 22 se construye con un argumento de construcción. macOS (G1-A): `HOME` temporal y, como este Mac tiene `/Applications/Google Chrome.app`, un binario de gate que sustituye solo la lista de rutas candidatas de D-5 por una vacía; ese límite se registra con la evidencia.

**Execution:** hilo principal, solo con autorización explícita (red, descarga de Chrome, imagen nueva).

**Test approach:** check.

**Changes:** la imagen del gate, si se crea, vive en `tests/pilot/container/` y entra en la entrega.

**Verification:**
- Linux: `hive install` interactivo ofrece la capacidad; `npm install -g agent-browser` y, sin navegador, `agent-browser install` terminan; `agent-browser open about:blank` abre y cierra; los dos pasos quedan `verified`. Con Node 22 la capacidad queda manual con el motivo.
- macOS (G1-A): con `HOME` temporal, la lista de rutas candidatas vacía en el binario de gate y `agent-browser` real, el paso de Chrome descarga en el `HOME` temporal, `agent-browser open about:blank` funciona y el paso queda `verified`. La ruta de `brew` queda cubierta solo por las pruebas de T2 (riesgo residual aceptado por el usuario).

## Verificación compartida

| Control | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Pruebas Go | Al final, en el árbol de trabajo | `go vet ./...` y `go test ./...` pasan | Cubierto por la implementación |
| Gate nativo (T5) | Tras T1–T3 | Lo descrito en T5 | Pendiente de autorización explícita |
| Revisión de código | Antes de entregar | Pendiente de la pregunta de revisión | Pendiente |

## Revisión y estado

- Ronda 1 (2026-10-13), dos `hive-review-plan` en paralelo (código Go del instalador; instalación y cadena de suministro): 11 bloqueantes y varias sugerencias. Corregidos aplicando sus propuestas: reintento solo de Chrome, reconciliación por paso, `failed` tras interrupción, oferta sin versión, detalle visible y salida acotada, `ValidateStep` ligado al ID, campos `omitempty`, `Detect` sin procesos en modos seguros, grupo de procesos y `Stdin` vacío, requisitos `:76`/`:88`, `SECURITY.md` e `installer.md`, imagen Linux con Node 24, plataformas acotadas, aviso de Node con `brew`, binario sin suma, versión menor. Decisiones del usuario tras la ronda: D4b-A, G1-A, O1-A.
- Re-revisión (única ronda) sobre la versión corregida: el dominio Go confirmó B1–B4 y dejó dos huecos (doctor que lea archivos; `Detect` sin procesos en sesión no interactiva) y tres sugerencias (prefijo de npm, salida sin cancelar, salto de Chrome sin estado compartido); el dominio de instalación dejó la búsqueda de T4 incompleta, el gate de macOS con Chrome presente, la cobertura del gate en el requisito y dos sugerencias (una sola base, imagen sin `chromium` y Node 22). Todas se aplicaron tal como las propusieron los revisores, sin otra ronda (D-2, D-6, D-9, D-10, D-12, D-13, T2, T3, T4, T5).
- Límite: ninguna ejecución; las rutas de `findSystemChrome` se leyeron en el `postinstall.js` instalado, no en el repositorio de agent-browser.
- Progreso: build autorizado el 2026-10-13 («hagámoslo de una vez»). Modelos: implementador `hive-build-backend` = sonnet, verificador `hive-verify-task` = opus (ambos configurados, no observados). T1–T3 delegadas a un hijo; T4 en el hilo principal. Revisión: `/code-review` (E2-A).
- T4: hecha (2026-10-13). Delta `specs/versioned-installation/spec.md` con tres MODIFIED (catálogo con `agent-browser` y su gate por plataforma, «Verified release package installation» sobre el texto de pi-subagents, diagnóstico de solo lectura); `SECURITY.md`, `installer.md`, `tool-dependencies.md`, `deployment-manager.md`, `README.md`, `llms.txt`, `CHANGELOG.md` (nota de versión menor). Búsqueda de frases vigentes: quedan seis coincidencias, todas correctas: `SECURITY.md:27` (ahora «with two optional exceptions»), `README.md:156` e `installer.md:104` (Engram y Context7 siguen sin receta), `installer.md:124` (red solo en ese paso y los de `agent-browser`), `deployment-manager.md:76` («Hive does not install them» se refiere a los demás) y el escenario de Pi del delta («no other host step downloads»: `agent-browser` no es un paso de host). `go test ./tests/content/` pasa. La frase del delta sobre qué plataformas pasaron el gate se confirma o corrige con T5.
- T1–T3: implementadas por `hive-build-backend` (archivos: `tooling/providers/{types,catalog,run}.go` y pruebas, `tooling/providers/README.md`; `tooling/cli/agent_browser.go`, `agent_browser_runner.go`, `provider_adapter.go`, `install.go`, `tui.go`, `tui_hosts_view.go`, `doctor_integrations.go` y pruebas). RED: T1 por compilación (símbolos nuevos); T2 por rupturas deliberadas sobre pruebas escritas antes (h, j, i, a) y, en la ronda de corrección, salida con cabecera y razón de interrupción ausente; T3 cinco pruebas en rojo antes. Desvíos aceptados: la salida guardada conserva los últimos 64 KiB; la razón «interrupted before its result was recorded» se deriva al mostrar (la recuperación ahora imprime el detalle de pasos parciales); en modos seguros en Linux la oferta no comprueba Node ni el prefijo (el resumen lo dice); la lista de capacidades se ajusta a 80 columnas; el doctor sin navegador también nombra `AGENT_BROWSER_EXECUTABLE_PATH`. Se conserva la constante `agentBrowserSource` del doctor, fijada igual al catálogo por una prueba. En verificación.
- T1 verificada por `hive-verify-task` (Opus 5.5, observado): AC7 met; abrí `tooling/providers/run.go:13-18` y `:48-68`. Huecos menores sin defecto: sin caso propio para un manual de `agent-browser` con otro ID, un manual con `Method` sin `argv`, y `agent-browser-chrome` con otro método.
- T3 verificada por `hive-verify-task` (Opus 5.5, observado): AC5 met; abrí `tooling/cli/doctor_integrations.go:273-277` y `:308-323`. Límites anotados: que el doctor no llame a `node`/`npm` se comprobó leyendo, no con una prueba; ninguna prueba de la vista exige el texto «hive install» (comparte `detailLines` con el doctor).
- T2 verificada por `hive-verify-task` (Opus 5.5, observado), ronda 1: AC1 y AC3 met, AC4 not met (filtro de plataforma ausente en el caso de solo Chrome). Ronda de corrección enviada al implementador.
- T2 ronda de corrección 1: `planAgentBrowser` aplica el filtro de plataforma antes del caso con `agent-browser` presente; prueba nueva `TestAgentBrowserPresentWithoutBrowserOnUntestedPlatformIsManual` (roja antes, verde después); comentarios obsoletos de `provider_adapter_test.go` corregidos. Re-verificación de AC4 en curso.
- T2 re-verificada (mismo verificador, Opus 5.5): AC4 met; abrí `tooling/cli/agent_browser.go:184-195`. AC1, AC3 y AC4 met. Observación abierta, no incumplimiento: el aviso de interrupción solo aparece al recuperar con `hive install`, no con `hive recover` ni la TUI.
- T5 (autorizado P3-A, 2026-10-13): imagen `tests/pilot/container/agent-browser-gate.Dockerfile` (Debian bookworm amd64, Node v24.21.0 o v22.23.3 por argumento, prefijo npm de UID 1000, librerías de Chrome sin `chromium`), corrida en Docker 29.8.2 con emulación amd64 en este Mac arm64; binario `linux/amd64` del árbol de trabajo; `hive install --hosts codex` interactivo bajo `script`.
  - Linux amd64, Node 24: el resumen mostró `npm install -g agent-browser` (última versión, postinstall sin suma) y `agent-browser install`; tras `y`, «Hive installed and verified»; `agent-browser` 0.39.1 en `~/.npm-global/bin`, Chrome 155.0.8059.79 en `~/.agent-browser/browsers`, `agent-browser open about:blank` abrió y cerró, `hive doctor` muestra el registro `verified`.
  - Linux amd64, Node 22: capacidad manual con «Node.js v22.23.3 is older than 24…» y los comandos oficiales; nada instalado; núcleo instalado; salida distinta de cero (parcial), como pide la especificación.
  - macOS arm64 (G1-A): binario de gate con la lista de rutas candidatas de D-5 vacía, `HOME` temporal y entorno limpio; el resumen ofreció solo `agent-browser install`; tras `y`, «Hive installed and verified»; Chrome 155.0.8059.79 descargado en el `HOME` temporal; `agent-browser open about:blank` funcionó con `AGENT_BROWSER_SOCKET_DIR` corto (la ruta del socket bajo el `HOME` temporal excedía 103 bytes: límite del entorno de prueba, no del cambio). La ruta de `brew` queda cubierta solo por pruebas con dobles (riesgo aceptado, G1-A).
  - AC3 y AC4: comprobación real hecha. La frase del delta sobre las plataformas del gate coincide con estos resultados. Limpieza: `HOME` temporal y copia del código del gate borrados; las imágenes Docker del gate se quitan al cerrar.
- Gates finales sobre T1–T5 (2026-10-13): `gofmt -l` vacío; `go vet ./...`, `go test ./...` y las pruebas de skills en verde. Commits por tarea en `feat/flash-class-guidance`: `b0fb62e0` (T1), `f5721dd3` (T2), `e4e03978` (T3), `34bcc3df` (T4), `673cbe4e` (T5), sin push.
- `/code-review` medium de Claude Code (E2-A), 10 hallazgos sin pasada de verificación propia; confirmados leyendo el código los que se corrigen. P2 en corrección (implementador de T2, sin re-revisión por ser P2 locales): interrupción con Ctrl-C deja brew/npm huérfanos (`agent_browser_runner.go:62`, contexto sin señales y `Setsid`); el preflight mezcla stderr en el prefijo o la versión (`agent_browser.go:95-104`); descendientes que sobreviven a una salida normal; directorio de trabajo en `/tmp` compartido en Linux; carrera del límite de tiempo; home sintético que aún lee el sistema real. Solo reportados (P3): instalación de la última versión sin pin ni suma (decisión D3 del usuario, declarada antes del consentimiento); la TUI ignora señales durante una escritura de hasta 30 min; ejecución acotada duplicada con `model_catalog.go`; el Dockerfile del gate verifica Node con SHASUMS sin firma GPG.
- Correcciones P2 de la revisión, hechas sin re-revisión (locales, sin cambio de contrato), cada una con prueba que falla al revertirla: interrupción por señal con «interrupted» y grupo de procesos terminado; grupo terminado tras toda ejecución; directorio de trabajo privado 0700; preflight solo con stdout; salida limpia antes que el plazo; home sintético sin lecturas reales. Commit `067d0453`. El gate T5 corrió antes de estas correcciones; tocan el manejo de señales, el directorio de trabajo y el análisis del preflight, no las recetas, y las cubren sus pruebas con el ejecutor real.
- Gate de Linux amd64 con Node 24 repetido tras `067d0453`: «Hive installed and verified», `agent-browser` en el prefijo de npm, Chrome 155.0.8059.79 descargado, `open about:blank` abrió y cerró.
- Cierre de la sesión de build (2026-10-13): binario reconstruido y `hive update` aplicado desde `f8116af5`; imágenes Docker del gate y binarios temporales borrados. El registro queda sin commit hasta el merge, como pide el flujo de la rama de prueba.
