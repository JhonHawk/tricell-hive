# Paso de onboarding: tarjeta de preguntas de Codex en modo Default

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · fusionado en `development` por [#73](https://github.com/JhonHawk/tricell-hive/pull/73) (`bb4fd503`) el 2026-10-18, a decisión del usuario antes de las pruebas en sesiones reales previstas; desplegado en local |
| Tracker · GitHub Issues (JhonHawk/tricell-hive) | Sin issue: el trabajo nace de la sesión del 2026-10-18; no se crea uno sin pedirlo |
| Git | Rama de prueba actual `feat/gh-67-flow-walkthrough-skill` · commits por tarea · recompilar el binario y `hive apply` local · sin push ni PR hasta que el usuario lo pida |
| Verificación | `go test ./...` · `go vet ./...` · `hive install --dry-run --home <tmp>` · revisión `hive-review-code` + `hive-refute-claim` y `hive-review-security` |
| Siguiente paso | Ninguno: registro archivado |

## Petición

> activemos la bandera y hagamos que hive la habilite por defecto al instalarse en codex, research al respecto

Respuestas del 2026-10-18: D1-A (activar ahora el flag en `~/.codex/config.toml`, hecho con `codex features enable default_mode_request_user_input`), D2-A (mecanismo: paso de onboarding ejecutable en `tooling/providers`), Q1-A (solo en `hive install`), Q2-A (marcado por defecto en instalaciones interactivas; el modo no interactivo no lo corre), D-G («nos mantenemos en la rama actual, commit y despliegue»), D-R (`hive-review-code` + `hive-refute-claim`, más `hive-review-security`).

## Objetivo

Codex solo muestra su tarjeta síncrona de preguntas en modo Plan; en modo Default la guía de Hive pregunta en texto y el usuario responde tecleando IDs. La feature experimental `default_mode_request_user_input` (etapa *under development*, apagada por defecto) registra la tarjeta también en Default: el núcleo espera la respuesta, y la TUI la autorresuelve vacía tras 60 s ocultos más 60 s de cuenta visible, salvo que el usuario toque el teclado en la tarjeta. Hive debe ofrecer activarla al instalar con Codex como host, marcada por defecto y confirmada por el usuario en «Apply these changes?», usando el escritor propio de Codex (`codex features enable …`). Invariante que este cambio fija: Hive no edita por sí mismo ningún archivo de configuración de un host; con consentimiento puede ejecutar el comando propio del host que lo escribe, como ya hace con `pi install` para Pi (`_support/docs/architecture/installer.md:119-123`). Las frases existentes que hablan de "nunca escribe" están acotadas a Grok (`deployment-manager.md:42`, `agent-delivery.md:24,46`); la única general, `tool-dependencies.md:17`, se reescribe en T5.

Fallo observado (host Codex CLI 0.163.0, TUI, sesión `01a15146-7905-7c70-a160-e56cb2dd534e`, proyecto carvento, 2026-10-18): en modo Default sin la feature, las preguntas D5 (autorizar Bugbot) y D6 (retirar un commit accidental) salieron en texto, la tarjeta asíncrona cerró al terminar el turno y el usuario tuvo que contestar tecleando («no me dan opciones realmente»). Esta opción quedó excluida como D4-A el 2026-10-08 (cambio `gh-9-communication-decision-flow`) y reiterada el 2026-10-11 (`question-flow-checkpoints`) por la autorresolución en un minuto; el usuario la reabre por decisión explícita el 2026-10-18, sin la evidencia de dos sesiones que pide la regla de medición, con la lectura de `rust-v0.163.0` (obs (src), sin documentación oficial de la feature): el handler espera la respuesta (`session.request_user_input` aguarda `rx_response`), la TUI autorresuelve a los 60 s ocultos + 60 s visibles y cualquier tecla en la tarjeta cancela la cuenta, el resultado vacío llega al modelo como `{"answers":{}}`, y una pregunta vencida ya se repite en texto por la regla vigente de `questions.md`. El mensaje del commit nombra la sesión.

## Alcance y aceptación

Incluido: proveedor nuevo en el catálogo de capacidades opcionales, su paso ejecutable con argv fijado, oferta condicionada a Codex, selección marcada por defecto en el instalador de texto y en la TUI, ejecución y reconciliación, fila en `hive doctor`, documentación y changelog, actualización de la tabla de Codex en `agent-delivery.md`.

Excluido, con motivo:
- `hive update` no reofrece ni repite pasos de onboarding (Q1-A); las instalaciones existentes activan el flag con el comando, que doctor nombra.
- Ningún cambio en `content/`: `questions.md` ya usa la tarjeta que espera cuando el host la tiene y repite en texto una pregunta vencida.
- Reversión al desinstalar: ningún paso de onboarding se revierte hoy (`tooling/management/onboarding.go:20-21`); la documentación nombra `codex features disable default_mode_request_user_input`.
- Release: refrescar la instalación local no es un release; se propone al cierre solo si se cumple el criterio de `CONTRIBUTING.md`. Una capacidad nueva marcada por defecto sube la versión menor del siguiente release (`CONTRIBUTING.md:52-56`); Unreleased ya contiene entradas de ese nivel.
- Verificación de extremo a extremo del paso real: no corre en esta máquina (`hive update` no ofrece onboarding y el flag ya está activo). Se verifica el escritor de Codex aislado con un `CODEX_HOME` temporal (puerta compartida) y el paso con un ejecutor falso en las pruebas.

Restricciones: `go test ./...` y `go vet ./...` siguen en verde; el orden del catálogo conserva los índices existentes (`catalog.go:5`, `providers_test.go:26`); `hive doctor` no ejecuta programas ni lee configuración de hosts (`doctor_integrations.go:3-8`); el modo no interactivo y el `--home` sintético no corren el paso.

- AC1. `providers.ValidateStep` acepta un paso `planned` con ID `codex-question-card-enable`, proveedor `codex-question-card`, método `codex` y argv exacto `codex features enable default_mode_request_user_input`, y rechaza el mismo paso con otro argv, otro método o como paso de otro proveedor. *Falso en la base cuando* `rg -n 'only agent-browser has an executable recipe' tooling/providers/run.go` devuelve la línea 50 y ningún test nombra `codex-question-card`.
- AC2. `hive install` con `codex` entre los hosts seleccionados ofrece la capacidad `codex-question-card` como paso ejecutable sin versión cuando `codex` está en `PATH`, y como paso manual, con la razón de instalar Codex y correr el comando, cuando no lo está; sin `codex` entre los hosts no la ofrece. *Falso en la base cuando* `rg -n 'codex-question-card' tooling/providers/catalog.go` no devuelve nada.
- AC3. En el instalador de texto y en la vista de capacidades de la TUI la oferta ejecutable aparece marcada por defecto con una etiqueta propia (no «installs the latest version»), Enter o continuar la acepta, el usuario puede desmarcarla, la oferta manual (sin `codex` en `PATH` o bajo `--home`) no viene marcada, y la sesión no interactiva sigue sin seleccionar nada. *Falso en la base cuando* `capsView` arranca con todas las casillas en falso (`tui_hosts_view.go:1015`) y `selectProviderRequests` devuelve vacío ante Enter (`install.go:663-665`).
- AC4. `providerRunner.Execute` corre el argv del paso por el ejecutor acotado existente; `Reconcile` devuelve `verified` cuando `<CodexHome>/config.toml` tiene `default_mode_request_user_input = true` (tabla `[features]` o clave punteada raíz `features.default_mode_request_user_input`), incluso sin resultado almacenado tras una interrupción; `failed` cuando no; y `skipped` cuando el runner es sintético. *Falso en la base cuando* `Reconcile` cae en `StepFailed` para todo ID distinto de los de agent-browser (`agent_browser_runner.go:276-286`).
- AC5. `hive doctor` muestra una fila `Codex question card` cuando `codex` está registrado, con el estado del último registro de onboarding y un siguiente paso que nombra `codex features enable default_mode_request_user_input`, sin ejecutar programas ni leer `config.toml` (con la clave activa en el home de prueba sigue diciendo «not checked»). *Falso en la base cuando* `rg -n 'question card' tooling/cli/doctor_integrations.go` no devuelve nada.
- AC6. `README.md`, `llms.txt`, `_support/docs/architecture/tool-dependencies.md`, `installer.md`, `deployment-manager.md` y `agent-delivery.md` describen la capacidad, el invariante de configuración de host, la autorresolución y el comando de reversión, sin frases que digan que agent-browser es la única receta ejecutable o que la configuración del host queda intacta; `CHANGELOG.md` la registra en Unreleased/Added. *Falso en la base cuando* `rg -n 'codex-question-card' README.md llms.txt CHANGELOG.md _support/docs/architecture/tool-dependencies.md _support/docs/architecture/installer.md _support/docs/architecture/deployment-manager.md _support/docs/architecture/agent-delivery.md` no devuelve nada y `rg -n 'the one recipe|Only `agent-browser` has a provider recipe|Two steps are automated' _support/docs/architecture/installer.md _support/docs/architecture/tool-dependencies.md README.md` sí devuelve líneas.

## Entrega

- Repositorio `tricell-hive`, rama `feat/gh-67-flow-walkthrough-skill` (rama de prueba única del usuario), base de este cambio: el `HEAD` al iniciar el build, registrado en `tasks.md`.
- Commits por tarea verificada, solo con las rutas de la tarea; la carpeta de cambio queda en disco sin versionar hasta el cierre.
- Tras el build: `go build -o "$(command -v hive)" ./tooling/cli` y despliegue local con `hive update --out` + `hive apply --plan` (el contenido no cambia; el binario sí).
- Sin push, PR ni merge: el usuario prueba en sesiones reales y decide el merge de la rama; entonces `flow-close` cierra este registro con el de gh-67.
- Revisión: `hive-review-code` con `hive-refute-claim` sobre sus hallazgos bloqueantes y `hive-review-security` (ejecución de proceso con allowlist), sobre el candidato final antes de reportarlo completo.
