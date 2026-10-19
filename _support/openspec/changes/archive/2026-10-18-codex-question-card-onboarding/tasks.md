# Tareas

Base del build: `63d93c3e` en `feat/gh-67-flow-walkthrough-skill` (2026-10-18, B1-A).

## Tareas

### T1 — Proveedor, paso y lector de configuración

- [x] `codex-question-card` existe en el catálogo con su paso ejecutable validado, y `integrations/codex.FeatureEnabled` lee la clave.

**Closes:** AC1.

**Depends on:** none.

**Locations:** `tooling/providers/types.go`, `tooling/providers/catalog.go`, `tooling/providers/run.go`, `tooling/providers/run_test.go`, `tooling/providers/providers_test.go`, `integrations/codex/features.go` (nuevo), `integrations/codex/features_test.go` (nuevo).

**Execution:** delegada a `hive-build-backend` (Go, interfaz fijada por DS1 y DS5; escrituras sin solape con T2).

**Test approach:** tdd.

**Changes:** DS1 y DS5. Tests primero: `ValidateStep` acepta el paso con argv exacto y rechaza argv distinto, método distinto, proveedor cruzado, un paso manual con otro ID y el paso manual con comando; `TestCatalogEndsWithAgentBrowserForAllHosts` reescrita como orden estable con agent-browser antes de la entrada nueva; `FeatureEnabled` con archivo ausente, clave ausente, `true`, `false`, forma punteada raíz, tabla duplicada, tabla inline y las dos formas con valores distintos. Comentarios de paquete y `README.md` del paquete actualizados.

**Verification:** `go test ./tooling/providers/ ./integrations/codex/` en verde con los casos nuevos fallando antes del cambio (RED observado en el primer run).

### T2 — Oferta, plan, ejecución y reconciliación

- [x] El instalador ofrece la capacidad solo con Codex, planifica el paso ejecutable o el manual, lo ejecuta por el ejecutor acotado y lo reconcilia leyendo `config.toml`.

**Closes:** AC2, AC4.

**Depends on:** T1.

**Locations:** `tooling/cli/provider_adapter.go`, `tooling/cli/agent_browser_runner.go`, `tooling/cli/install.go` (`providerOffer.DefaultOn`), `tooling/cli/provider_adapter_test.go`, `tooling/cli/agent_browser_test.go` (fixtures `abFake`, `useAgentBrowserFake`, `newTestAdapter`).

**Execution:** delegada a `hive-build-backend`.

**Test approach:** tdd.

**Changes:** DS2 y DS4. El adaptador resuelve `CodexHome` en la fábrica y lo pasa al runner; cachea el resultado de `PATH` para `codex`; `Reconcile` del paso nuevo lee `FeatureEnabled` antes del resultado almacenado; `stepResultDetail` reconoce el prefijo nuevo; mensaje para Codex sin subcomando `features`; la oferta manual no lleva `DefaultOn`.

**Verification:** `go test ./tooling/cli/ -run 'Provider|AgentBrowser|CodexQuestion|Install'`: oferta ejecutable con `codex` en hosts y en `PATH`, manual sin `PATH` o sintético, ausente sin `codex` en hosts; `Execute` invoca el fake con el argv exacto; `Reconcile` → `verified` con `[features]\ndefault_mode_request_user_input = true\n` en el home de prueba aunque el resultado esté vacío, `failed` sin la clave, `skipped` con runner sintético; las pruebas existentes `install_test.go:140,156,417,711,723` (hosts codex, home sintético, entrada `"\ny\n"`) siguen sin error.

### T3 — Selección marcada por defecto

- [x] Las ofertas `DefaultOn` aparecen marcadas en la TUI y el instalador de texto acepta la selección por defecto con Enter; el modo no interactivo no selecciona nada.

**Closes:** AC3.

**Depends on:** T2 (campo `DefaultOn`).

**Locations:** `tooling/cli/install.go` (`selectProviderRequests`, `providerOfferLabel`), `tooling/cli/tui_hosts_view.go` (`capsView`), `tooling/cli/install_test.go` o el test existente de `selectProviderRequests`, `tooling/cli/tui_hosts_view_test.go`, `tooling/cli/agent_browser_tui_test.go`.

**Execution:** delegada a `hive-build-backend` (tras T2, mismo hijo o uno nuevo; sin solape con T4).

**Test approach:** tdd.

**Changes:** DS3: `DefaultOn` y `Effect` en `providerOffer`, etiqueta propia, texto de ayuda de la TUI (`:1124`) y prompt de texto (`:655`), `none`.

**Verification:** `go test ./tooling/cli/ -run 'SelectProvider|CapsView|Onboarding|Install'`: Enter con oferta `DefaultOn` devuelve esa petición; `none` devuelve vacío; Enter sin ofertas `DefaultOn` devuelve vacío (comportamiento actual); sesión no interactiva devuelve vacío; `capsView` inicial muestra `[x]` para la oferta `DefaultOn` y `[ ]` para agent-browser y para la oferta manual; la etiqueta de la oferta dice «enables Codex's question card in Default mode» y no «installs the latest version»; las pruebas de `install_test.go` que asumen que Enter omite todo se ajustan solo donde haya una oferta `DefaultOn`.

### T4 — Fila en doctor

- [x] `hive doctor` muestra la fila `Codex question card` con registro y siguiente paso, sin programas ni lectura de configuración.

**Closes:** AC5.

**Depends on:** T1 (ID del proveedor).

**Locations:** `tooling/cli/doctor_integrations.go` (`collectIntegrationRows`, `nextStep` con caso propio, `fillSourceFromCatalog`), `tooling/cli/doctor_integrations_test.go` (fixtures `fakePrograms`, `integrationsFixtureHosts`, `writeStepRecord`; fixture nueva con `codex` registrado), `tooling/cli/tui_integrations_view_test.go:69-85` (ajuste a 80x24 con seis filas).

**Execution:** delegada a `hive-build-backend` (paralelo a T2/T3: archivos distintos).

**Test approach:** tdd.

**Changes:** DS6.

**Verification:** `go test ./tooling/cli/ -run 'Doctor|Integration'`: fila presente con `codex` registrado y ausente sin él; `Record` refleja `verified`/`failed` del registro escrito; el siguiente paso contiene `codex features enable default_mode_request_user_input`; ningún `lookPath` para `codex`; con `[features]\ndefault_mode_request_user_input = true\n` en el home de prueba la fila sigue en «not checked»; la vista de integraciones cabe en 80x24 con seis filas.

### T5 — Documentación, changelog y tabla de Codex

- [x] La documentación describe la capacidad y su reversión, y la tabla de Codex en `agent-delivery.md` refleja la tarjeta en Default con la feature.

**Closes:** AC6.

**Depends on:** T1–T4 (nombres finales).

**Locations:** `README.md:157-160`, `llms.txt:39`, `CHANGELOG.md` (Unreleased/Added), `_support/docs/architecture/tool-dependencies.md:10-17` y § Boundaries, `_support/docs/architecture/installer.md:98,102,106,117,119-123`, `_support/docs/architecture/deployment-manager.md:76,79`, `_support/docs/architecture/agent-delivery.md:191,196,208,211`.

**Execution:** hilo principal (texto corto que depende de las decisiones de este registro).

**Test approach:** check: el `rg` de AC6 (positivo por `codex-question-card` en cada archivo; negativo por las frases «the one recipe», «Only `agent-browser` has a provider recipe», «Two steps are automated») y `rg -n 'rust-v0.163.0.*default\.md|60 s' _support/docs/architecture/agent-delivery.md` devuelve las filas editadas.

**Changes:** DS7.

**Verification:** los `rg` anteriores; `go test ./tests/content/` (enlaces) en verde.

## Verificación y revisión humana

| Puerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad / condición pendiente |
| --- | --- | --- | --- |
| Suite completa | Candidato final, local | `go test ./...` y `go vet ./...` en verde | Implementación |
| Instalación en seco | Candidato final, home sintético | `hive install --dry-run --home <tmp>` con Codex: la capacidad aparece como manual (sintético) sin ejecutar nada; en home real `hive install --dry-run` lista la oferta marcada por defecto con su etiqueta y el efecto en el resumen | Implementación |
| Escritor de Codex aislado | Candidato final, `CODEX_HOME` temporal | `CODEX_HOME=<tmp> codex features enable default_mode_request_user_input` y `codex features list` en ese home: la clave queda en `[features]` de `<tmp>/config.toml` y `FeatureEnabled` la lee como `true`; el home real no se toca. El paso real de extremo a extremo queda sin correr aquí (ver `proposal.md`) | Implementación |
| Revisión de código | Candidato final, antes de reportar completo | `hive-review-code` + `hive-refute-claim` sobre hallazgos bloqueantes; `hive-review-security` sobre allowlist, ejecución y lector TOML | D-R del 2026-10-18 |
| Binario y despliegue local | Tras la revisión | `go build -o "$(command -v hive)" ./tooling/cli`; `hive update --out` + `hive apply --plan`; `hive doctor` muestra la fila nueva | D-G del 2026-10-18 |

Sin efecto de UI renderizada: no aplica la puerta de UI.

## Entrega

Ver `proposal.md` § Entrega: rama actual, commits por tarea, sin push; revisión y despliegue local al final; merge cuando el usuario lo decida tras sus pruebas.

## Estado de revisión y progreso

- 2026-10-18: plan escrito tras el research (hilo `hive-research` sobre el mecanismo de onboarding, unos 174 k tokens, 84 s) y las respuestas Q1-A, Q2-A, D-G, D-R.
- 2026-10-18, revisión 1 (hashes `1f9e7a97…`, `a470a562…`, `e572244b…`): dos `hive-review-plan`, tooling Go (unos 153 k tokens, 175 s) y frontera de host, documentación y guía (unos 118 k tokens, 119 s). Bloqueantes aplicados: la prueba `TestCatalogEndsWithAgentBrowserForAllHosts` que DS1 rompía (T1 la reescribe); la oferta manual sin `DefaultOn`, con las pruebas de `install_test.go` como guardas (DS2, T2, T3); `installer.md`, `tool-dependencies.md:10-17`, `README.md:157-160` y `deployment-manager.md:79` añadidos a T5 con sus frases a reescribir (DS7, AC6); el invariante de configuración de host reformulado y con casa en `tool-dependencies.md` § Boundaries (proposal, DS7); la sesión observada `01a15146` y la reapertura de D4-A por decisión explícita citadas en `proposal.md`. No bloqueantes aplicados: AC4 sin «en seco» y `Reconcile` lee la configuración aunque falte resultado (DS4); `CodexHome` resuelto en la fábrica del adaptador (DS4); forma punteada raíz en `FeatureEnabled` (DS5); etiqueta propia de la oferta y textos de ayuda y prompt (DS3, AC3 con la cita `:1015`); caché del resultado de `PATH` (DS2); caso propio en `nextStep`, fixture con `codex`, ajuste 80x24 y prueba «not checked» con la clave activa (DS6, T4, AC5); comprobación de T5 con frases propias de la edición y `llms.txt` por nombre; `agent-delivery.md:191,196` y el payload `{"answers":{}}` (DS7); nivel menor del changelog y puerta del escritor aislado con `CODEX_HOME` temporal (proposal, puertas). Las correcciones aplican propuestas de los propios revisores, sin contrato ni criterio nuevo que exija re-revisión. Límite declarado por ambos: no ejecutaron `codex`; la escritura bajo `[profiles.*]` con perfil activo queda sin verificar.
- 2026-10-18, B1-A, build iniciado en `63d93c3e`; T1 delegada a `hive-build-backend`. Puerta «escritor de Codex aislado» pasada antes del código: con `CODEX_HOME` temporal vacío, `codex features enable default_mode_request_user_input` (0.163.0) crea `config.toml` con `[features]\ndefault_mode_request_user_input = true` y un directorio `tmp/`; repetirlo es idempotente; `disable` deja la tabla vacía; una feature desconocida sale con código 1; `codex features list` refleja el estado. Sin perfil activo en esa prueba.
- 2026-10-18, construcción: T1 `de85d21a`, T2 `c394efc8`, T4 `b11beb6b`, T3 `dab65647` (una ronda de corrección: el prompt superaba 80 columnas; ahora dos líneas y prueba de anchura), T5 `0fb53c1c`; cada una verificada por `hive-verify-task` (AC1–AC5 cumplidos; implementador `hive-build-backend` y verificador en sus perfiles configurados, modelos no observados por el host). Puertas compartidas sobre `0fb53c1c`: `go vet ./...` limpio; `go test -count=1 ./...` en verde; instalador de texto en seco por pseudoterminal con el binario del candidato: home sintético ofrece `codex-question-card` como manual y sin marcar; home real lo ofrece con la etiqueta «enables Codex's question card in Default mode» y «(selected by default)», Enter lo acepta y el resumen muestra el comando, el efecto («Codex writes the file, not Hive») y el `disable` antes de confirmar; la sesión no interactiva muestra «Optional capabilities: none selected». Revisión `hive-review-code` y `hive-review-security` en curso.
- 2026-10-18, revisión de código (`hive-review-code`, unos 144 k tokens, 319 s): sin bloqueantes. P3 confirmado: `FeatureEnabled` no reconoce la cabecera `["features"]` (Codex escribe bajo ella si existe) y un paso con salida 0 sin confirmación queda `failed` sin mensaje. P2 no confirmado: con un Codex anterior a `features enable` la oferta marcada por defecto dejaría la instalación parcial. Ronda R1 delegada a `hive-build-backend`: comillas, BOM y líneas largas en el lector; mensaje «Codex reported success but Hive could not confirm…»; la oferta es ejecutable solo cuando `codex features list` (sonda única bajo `runsPrograms()`) reporta la feature, y manual con la razón de actualizar Codex en otro caso. DS2, DS4 y DS5 se actualizan con el resultado; el verificador de la ronda es el hilo principal con la suite completa, al ser propuestas del propio revisor.
- 2026-10-18, revisión de seguridad (`hive-review-security`, unos 185 k tokens, 376 s): sin bloqueantes; allowlist, consentimiento previo en texto y TUI, y fila de doctor confirmados con `file:line`. F1 (la salida de un error de parseo de Codex podría copiar una línea de `config.toml` al diario) y F2 (= P3) se suman a R1.
- 2026-10-18, R1 en commit `8c15f59a` (implementador `hive-build-backend`, hook `codexFeaturesProbe`; verificación del hilo principal: `go vet ./...` limpio y `go test -count=1` de `integrations/codex`, `tooling/providers`, `tooling/cli`, `tooling/management` y `tests/content` en verde; diff revisado). DS2, DS4 y DS5 y la documentación (`installer.md`, `tool-dependencies.md`, README, changelog) actualizados en el mismo commit. Binario reinstalado con `go build -o "$(command -v hive)" ./tooling/cli`; `hive update --out` + `hive apply` desde `8c15f59a` responde «unchanged» (el contenido no cambia); `hive doctor` muestra la fila «Codex question card · not checked · No onboarding record yet». Sin verificación del paso real de extremo a extremo en esta máquina (ver `proposal.md`). Trabajo entregado en la rama de prueba; pendiente tu prueba en sesiones reales y tu decisión de merge.
