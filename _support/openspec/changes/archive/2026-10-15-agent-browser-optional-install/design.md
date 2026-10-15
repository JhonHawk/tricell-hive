# Diseño

## Contexto verificado

Inspeccionado el 2026-10-13 en `feat/flash-class-guidance` (`5d15446f`) por un hijo `hive-research` y por dos revisores del plan; las afirmaciones decisivas se comprobaron en sus ubicaciones.

- Catálogo: `tooling/providers/catalog.go` devuelve Engram, Context7 y pi-subagents, todos manuales: «Hive never installs or configures them».
- Tipos: `tooling/providers/types.go:23-25`, `Manual` es el único estado; `Step` solo tiene `ID`, `Provider`, `Status`, `ManualReason`. `ValidateStep` decodifica con `DisallowUnknownFields`.
- Validación: `tooling/providers/run.go:19-29` rechaza todo proveedor fuera de los tres y todo estado no manual (`:25`).
- Adaptador: `tooling/cli/provider_adapter.go:16-31` (`Detect`, `ManualOnly: true`, sin puntos de sustitución, se llama también en `--dry-run`: `install.go:352-358`), `:64-82` (`Plan`), `:95` (`manualProviderRunner.Execute` no ejecuta nada).
- Consentimiento: `install.go:497-508`; `selectProviderRequests` (`:634-700`, vacío si no es interactivo, `:638`); toda oferta no `ManualOnly` pasa por `readProviderVersion` («Exact version for…», `:643`, `:688`) y por la TUI (`tui_hosts_view.go:1024`, `:1105-1111`, `:1167`); resumen con `Effects` (`:784-796`); detalle de un `partial` en `showPartialOnboardingDetail` (`:862-876`), que hoy solo lee `ManualReason`.
- Diario: `management.ExternalRunner` (`onboarding.go:22-26`); `runProviders` (`:347-379`) ya acepta un `failed` devuelto sin error de Go y sigue con el paso siguiente; si `Execute` devuelve error, marca `unknown` y aborta; si el proceso se interrumpe antes de guardar `Result` (`:354`), `RecoverOnboarding` (`:480`) solo llama a `Reconcile`, y un `unknown` bloquea operaciones nuevas. `onboardingHash` (`:110`) recodifica los structs; no hay golden de onboarding.
- Ejecución con límite: `tooling/cli/model_catalog.go:285-300` usa `CommandContext` y `WaitDelay`. `pi_package.go:13-37` no tiene límite ni control de stdin; solo sirve de modelo para la variable sustituible y el falso en `--home` sintético.
- Doctor: `doctor_integrations.go:23-26`, `:50`, `:114-164`, `:143` (`recordNotTracked`), `:241-254` (`agentBrowserNextStep`), `:244-245` («Found» con cli + skill), `:287` (`apply` toma el primer paso con el prefijo).
- Especificación (`_support/openspec/specs/versioned-installation/spec.md`): `:62` limita el catálogo y exige gate nativo por host y plataforma, y que los runtimes no se instalen implícitamente; `:76` y el escenario `:88` dicen que solo el paso de pi-subagents descarga de npm; `:155` dice que detectar `agent-browser` no lo añade al catálogo y que el diagnóstico solo ejecuta `--version`.
- Otros textos que hoy niegan la instalación: `SECURITY.md:27-32` («Hive does not install CLI executables»), `_support/docs/architecture/installer.md:98`, `:101-113`, `tool-dependencies.md:30-34` («exact versions resolved before consent»), `deployment-manager.md:76`, `README.md:127`, `:153-158`, `llms.txt:39`.
- Rama sin integrar `feat/pi-subagents-unversioned-install`: `57d2fcab` quita la versión de pi-subagents; `9670e99e` añade un delta que modifica en su totalidad el requisito «Verified release package installation».
- `agent-browser` 0.39.0 (README del Cellar y `scripts/postinstall.js`): `brew install agent-browser` (la fórmula depende de `node` y de unas 25 fórmulas más, `INSTALL_RECEIPT.json`); `npm install -g agent-browser` (`engines.node >= 24.0.0`; el postinstall descarga el binario nativo de GitHub sin comprobar suma); el README (`:85`) dice que detecta automáticamente Chrome, Brave, Playwright y Puppeteer instalados; `findSystemChrome` (`postinstall.js:166-190`) revisa en macOS `/Applications/Google Chrome.app`, `Google Chrome Canary.app` y `Chromium.app`, y en Linux `google-chrome`, `google-chrome-stable`, `chromium-browser` y `chromium` en el PATH; `agent-browser install` descarga Chrome for Testing en `$HOME/.agent-browser/browsers` (indicio: carpeta existente; no documentado si es idempotente).
- Imagen bb1 (`tests/pilot/container/Dockerfile`, `d345158a`): Debian slim amd64, UID 1000, sin Node ni npm ni las bibliotecas de Chrome.

## Diseño elegido

D-1. **Primera receta ejecutable.** `agent-browser` entra al catálogo al final de `Catalog()` (los índices de las pruebas existentes siguen valiendo), para los seis hosts, con paso instalable. `providers.Step` gana campos `omitempty` para método y `argv` (los pasos manuales antiguos siguen validando; `ExternalStep` y `StepReceipt` no cambian). `ValidateStep` liga cada `argv` a su ID: `agent-browser-install` acepta solo `["brew","install","agent-browser"]` o `["npm","install","-g","agent-browser"]`; `agent-browser-chrome` acepta solo `["agent-browser","install"]`. Engram, Context7 y pi-subagents siguen manuales.

D-2. **Dos pasos y su dependencia.** `agent-browser-install` y `agent-browser-chrome`. `Execute` solo recibe su propio paso (`onboarding.go:347`), así que el de Chrome no lee estado compartido: busca `agent-browser` en el PATH y, si no está, devuelve `skipped` con motivo.

D-3. **Gestor por plataforma, solo en pares probados.** Se ofrece como instalable solo en macOS arm64 (brew) y Linux amd64 (npm con Node ≥ 24 y prefijo global escribible). Si falta el gestor, Node es antiguo, el prefijo de npm no es escribible o la plataforma no es una de esas dos, la capacidad aparece como manual con el motivo y el comando oficial; `ValidateStep` acepta para eso un paso `agent-browser-manual`. El resumen previo al consentimiento advierte que `brew` instala Node y sus dependencias.

D-4. **Sin versión fija** (D3 del usuario). La oferta es de un tipo nuevo «sin versión» en `providerOffer`, para que ni el texto ni la TUI pidan una versión exacta. El resumen dice que se instala la última versión del gestor. `tool-dependencies.md` cambia su regla de versiones para estas recetas y anota que el postinstall de npm descarga un binario sin suma.

D-5. **Chrome solo si no hay navegador** (D4b-A del usuario). El paso de Chrome se planifica solo si no existe ninguno de los navegadores que revisa `findSystemChrome` en esa plataforma, ni `AGENT_BROWSER_EXECUTABLE_PATH` apunta a un archivo, ni hay un navegador en `$HOME/.agent-browser/browsers`.

D-6. **Ejecutor.** `agentBrowserRunner` en `tooling/cli`: variable de paquete sustituible en pruebas y falso en `--home` sintético, como `pi_package.go`; ejecución como `model_catalog.go` (`exec.CommandContext`, sin shell, `WaitDelay`), en su propio grupo de procesos que se mata entero al expirar, `Stdin` vacío, `HOMEBREW_NO_AUTO_UPDATE=1`, límite de 15 min por paso y salida guardada recortada a 64 KiB: lo que exceda se descarta sin cancelar el proceso (a diferencia de `overflowBuffer` en `model_catalog.go:290-295`, que cancela). La salida guardada queda en el estado local y el detalle del resultado muestra solo sus últimas líneas. Un fallo, expiración o cancelación devuelve un resultado con código de salida y mensaje, nunca un error de Go. Mensajes específicos para `EACCES` de npm y para un binario instalado que no quedó en el PATH.

D-7. **Reconciliación por paso.** `agent-browser-install`: verificado si `agent-browser` está en el PATH. `agent-browser-chrome`: verificado si hay un navegador según D-5. En ambos, si no hay resultado guardado (interrupción), devuelve `failed` con motivo «interrumpido», nunca `unknown`.

D-8. **Reintento.** `Detect` ofrece la capacidad completa cuando falta `agent-browser`, y ofrece solo el paso de Chrome cuando `agent-browser` está pero no hay navegador según D-5. Así un Chrome fallido se reintenta en la siguiente ejecución.

D-9. **Sin efectos en modos seguros.** El adaptador recibe si la sesión es interactiva (por `onboardingInput`, `install.go:124-128`, `:497-503`). Con `--home` sintético, `--dry-run` o sesión no interactiva, `Detect` no ejecuta ningún programa (ni `node --version` ni `npm config get prefix`). En sesión interactiva real, obtiene la versión de Node con `node --version` y el prefijo global con `npm config get prefix` (Linux), ambos con límite de tiempo. Todas las comprobaciones usan funciones sustituibles para `lookPath`, `GOOS`/`GOARCH`, entorno, existencia de archivos, versión de Node, prefijo de npm y la lista de rutas candidatas de navegador de D-5.

D-10. **Doctor.** Sigue de solo lectura: solo lee el PATH y archivos (las rutas de D-5), sin ejecutar programas, con las mismas funciones sustituibles de D-9. Con `agent-browser` ausente, o presente sin navegador, el siguiente paso nombra `hive install`. La columna Record muestra el peor estado de los pasos con prefijo `agent-browser-`.

D-11. **Sin archivo de skill.** `browser-automation.md` ya usa `agent-browser skills get core`.

D-12. **Orden de integración** (O1-A del usuario). La rama `feat/pi-subagents-unversioned-install` (cortada de `bba2fd48`) se integra antes en `development`; después `feat/flash-class-guidance` incorpora `development`, y esa punta es la base del build. El delta se escribe sobre el texto resultante de «Verified release package installation».

D-13. **Gate por plataforma en la especificación.** El requisito MODIFIED del catálogo dice que el gate de una herramienta que no depende del host es por plataforma, y nombra lo que pasó: Linux amd64 para los dos pasos y macOS arm64 solo para el paso de Chrome, con el riesgo de `brew` aceptado por el usuario (G1-A). También dice que la instalación de Node por `brew` se declara antes del consentimiento, así que no es implícita.

Decisiones menores del orquestador: D-2, D-6 (límites), D-7, D-10, D-11.
