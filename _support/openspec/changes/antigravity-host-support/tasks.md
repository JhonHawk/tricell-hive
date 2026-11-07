# Tareas

Base: `tricell-hive` `development` en `93ab65ee` (2026-11-06); rama `feat/antigravity-host-support`. T0 corre primero, en el hilo principal, porque confirma las rutas de reglas y agentes de las que dependen T2 y T3. T1 sigue porque fija el host y su raíz, que el resto consume. T2, T3 y T4 escriben archivos distintos (partes de la guía en el plan y la propiedad; perfiles y renderizado de roles; hooks y visor) y corren en hijos uno tras otro, porque las tres tocan `plan.go` o los fixtures compartidos de `testdata`. T5 corre en un hijo de documentación cuando T1–T4 fijaron los hechos. T6 es del hilo principal.

### T0 — Sonda de formato de reglas y agentes globales en `agy`

- [x] Confirmado que `agy` 1.5.2 carga una regla `always_on` de `~/.gemini/config/rules/` y lista un agente de `~/.gemini/config/agents/` con `model: "inherit"`; ambos archivos de prueba quedan borrados.

**Closes:** ninguno (sonda que confirma las rutas de las que dependen AC1 y AC5 antes de escribir código)

**Depends on:** ninguna

**Locations:** ninguna en el repositorio; escribe y borra `~/.gemini/config/rules/hive-probe.md` y `~/.gemini/config/agents/hive-probe.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** crea los dos archivos de prueba solo si no existen, corre `agy agents` (sin llamada al modelo) y una llamada `agy -p` con el modelo más barato que pide citar un marcador aleatorio de la regla sin herramientas, y los borra al terminar aunque falle. Usa una de las 3 llamadas autorizadas. Si alguna ruta no carga, el plan se revisa antes de T2 y T3.

**Verification:** `agy agents` muestra `hive-probe`; la salida `stream-json` cita el marcador sin pasos de herramienta; `ls` confirma que los archivos de prueba ya no existen.

### T1 — Antigravity registrado como host de scope usuario, con skills

- [x] `--hosts antigravity` se acepta en install, update, remove, status, recover y doctor; `--scope project` se rechaza; las skills quedan enlazadas en `<AntigravityHome>/skills/`.

**Closes:** AC1 (registro, skills y rechazo de scope), AC2 (skills ajenas preservadas)

**Depends on:** T0

**Locations:** `integrations/antigravity/` (nuevo), `integrations/target/target.go`, `tooling/management/{types.go,plan.go,update_contract.go}`, `tooling/cli/{install.go,main.go,doctor_sessions.go,view_work_contract.go,setup.go}`, `tooling/providers/catalog.go`, fixtures y pruebas que fijan la lista de hosts (incluidos `tooling/management/testdata/state_without_overrides.golden` y `providers_test.go`), `tooling/management/antigravity_test.go` (nuevo)

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** nuevo paquete con `Resolve` (raíz `~/.gemini/config`, enlace relativo por skill hacia `../../../.agents/skills/<skill>` como `claude.go:33-41`), el campo `AntigravityHome` y su resolución, la entrada en cada lista de hosts que nombra el diseño (sección Impacto), las listas de hosts de onboarding de proveedores y de raíces de Context7, y el binario `agy` para detectar la instalación. No toca la guía ni los roles todavía.

**Verification:** `go test ./tooling/... ./integrations/...` pasa; en un home sintético `go run ./tooling/cli plan install --hosts antigravity --scope user --home <tmp> --state-dir <tmp-state>` propone un enlace relativo por skill que resuelve a `<tmp>/.agents/skills/<skill>`, y `--scope project --root <dir>` falla con el error de scope no soportado; con un enlace ajeno `firecrawl` previo en `<tmp>/.gemini/config/skills/`, install y remove lo dejan con el mismo destino.

### T2 — Guía global partida en reglas siempre activas

- [x] El plan propone `rules/hive-NN.md` con `trigger: always_on`, de 24.000 bytes o menos cada una, cuya unión reproduce la guía reescrita; update agrega o quita partes cuando cambia su cantidad.

**Closes:** AC1 (partes de la guía), AC2 (reglas ajenas preservadas y partes sobrantes quitadas)

**Depends on:** T1

**Locations:** `integrations/antigravity/`, `tooling/management/{plan.go,ownership.go,apply.go,voice.go,types.go}`, pruebas en `tooling/management/`

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** el corte determinista por encabezados `##` fuera de bloques de código con límite de 22.000 bytes por parte, el error que nombra una sección que no cabe, el número de parte tomado de la ruta del registro, su validación de fuente y de `Leading`, y la voz solo en la primera parte (también en `addVoiceChangesForInstall` y en la acción `hive voice`), como fija `design.md`. Cursor, Claude y los demás hosts siguen con un solo bloque.

**Verification:** pruebas que comprueban: cada parte de 24.000 bytes o menos con su frontmatter; la unión de los textos gestionados igual a la guía reescrita; un `##` dentro de un bloque de código no corta; una sección sintética más grande que el límite hace fallar el plan nombrándola; una guía sintética que crece y otra que se achica cambian la cantidad de partes y update quita solo las sobrantes de Hive; install con la voz más larga y `hive voice set` cambian solo `hive-01.md`, que queda en 24.000 bytes o menos; un `rules/otra.md` ajeno queda byte a byte igual tras install, update y remove. `go test ./tooling/management/...` pasa.

### T3 — Roles de Hive renderizados para Antigravity con modelo heredado

- [x] Cada rol queda en `agents/<rol>.md` con `name`, `description` y `model: "inherit"`, sin `tools` ni esfuerzo; `hive models set` ofrece `inherit`, `flash_lite`, `flash` y `pro` para Antigravity y rechaza un esfuerzo.

**Closes:** AC1 (roles)

**Depends on:** T0, T1

**Locations:** `integrations/agent-profiles.json`, `integrations/agents/{agents.go,testdata/*}`, `tooling/cli/{model_catalog.go,tui_models_view.go,tui_models_picker.go}`, `tooling/management/models.go`, fixtures `agent-profiles.json` de `tooling/management/testdata` y `tooling/cli/testdata`

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** entrada `antigravity` opcional en los perfiles (como Cursor), `observe` sin campos, rechazo de esfuerzo, `render.golden` con los roles de Antigravity, y el catálogo fijo de niveles de modelo. No cambia los roles de otros hosts.

**Verification:** `go test ./integrations/agents/... ./tooling/cli/...` pasa con el golden nuevo; un rol `observe` renderizado no tiene `tools` ni esfuerzo y conserva su instrucción de solo lectura; un release congelado sin la entrada `antigravity` sigue leyendo sus perfiles.

### T4 — Hook de sesión y visor para Antigravity

- [x] `hooks.json` recibe el hook con nombre `hive` sin tocar otros hooks; `hive hook` traduce la carga de `agy` al evento canónico con host `antigravity` e imprime JSON válido; el visor y la identidad de sesión aceptan el host.

**Closes:** AC2 (hooks ajenos preservados), AC3

**Depends on:** T1

**Locations:** `integrations/antigravity/{hook_register.go,hook_event.go}`, `integrations/hostevent/`, `tooling/management/{host_hooks.go,json_hooks.go}`, `tooling/cli/{hook_command.go,hook_map.go,view_identity_resolve.go}`, `tooling/viewer/{timeline_contract.go,timeline_protocol.go}`, `tooling/viewer/frontend/src/timeline-contract.ts`, bundle en `tooling/viewer/static/app/assets/`

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** la forma de hook con nombre en la raíz de `hooks.json` (el editor reemplaza solo la clave `hive` y guarda el respaldo), los eventos `SessionStart`, `PreInvocation` y `Stop` con el evento por argumento y su traducción como fija `design.md` (`Stop` es fin de turno, nunca fin de sesión; sin `PreToolUse`), el directorio de trabajo tomado de `workspacePaths[0]`, la salida mínima aceptada por evento en todo camino de `runHook`, `agy` → `antigravity` en la resolución de identidad, y el host en la línea de tiempo con su bundle reconstruido. No lee el historial nativo de `agy`.

**Verification:** con un `hooks.json` que contiene el hook `herdr` de hoy, install, update y remove dejan `herdr` con los mismos bytes y dejan un respaldo; para cada evento, una carga `agy` de ejemplo por stdin produce el evento canónico esperado con host `antigravity` y el directorio de `workspacePaths[0]`, y `Stop` nunca produce un fin de sesión; con argumentos desconocidos, un pánico forzado y un tiempo agotado, `hive hook` sale con código 0 e imprime la salida aceptada del evento; una ejecución sin un proceso `agy` como ancestro (la App) también; las pruebas del contrato de la línea de tiempo y del frontend aceptan `antigravity`. `go test ./tooling/...` y las pruebas del frontend pasan.

### T5 — Contenido distribuido y documentación con Antigravity

- [x] La línea de sintaxis de comandos de `global.md` incluye Antigravity con `/flow-x`; las referencias, la arquitectura, `README.md`, `llms.txt` y el sitio describen el host y sus límites.

**Closes:** AC4

**Depends on:** T1, T2, T3, T4

**Locations:** `content/guidance/global.md:85`, `content/skills/flow-walkthrough/references/host-session-identity.md`, `content/skills/harness-audit/{SKILL.md,references/hierarchy.md,references/memory.md,references/native-tools.md,references/instruction-files.md,references/project-layout.md}`, `_support/docs/architecture/{deployment-manager.md,agent-delivery.md,host-session-events.md,compaction-survival.md,repository-and-distribution.md,installer.md,tool-dependencies.md,workspace-and-artifacts.md}`, `_support/openspec/specs/` solo mediante los deltas de este cambio, `AGENTS.md`, `README.md` (incluido "Six host adapters"), `llms.txt`, `site/src/content/docs/{host-limits.md,viewer.md,index.mdx}`

**Execution:** `hive-write-docs`

**Changes:** cada lugar que lista hosts gana a Antigravity con hechos marcados como documentados u observados (tope de 24 KB verificado el 2026-11-06, memoria Knowledge, historial no legible). Antes de editar `global.md`, revisa todo lo que gobierna esa línea como pide `AGENTS.md` y registra su tamaño nuevo para el pull request. No cambia ninguna otra regla.

**Verification:** `go test ./tests/...` pasa (enlaces y tamaños de skills); búsquedas de `Cursor` y de los conteos de hosts (`six`, `seis`, `6 hosts`) en `content/`, `_support/docs/`, `site/`, `README.md`, `llms.txt` y `AGENTS.md` no dejan una lista ni un conteo sin Antigravity; el sitio compila (`npm run build` en `site/`).

### T6 — Suite completa y comprobación en vivo en `agy`

- [ ] Con el candidato instalado solo para Antigravity en el `~/.gemini/config` real, `agy` carga la guía completa, lista los roles una vez y su sesión llega al visor.

**Closes:** AC5

**Depends on:** T1, T2, T3, T4, T5

**Locations:** ninguna en el repositorio; escribe `~/.gemini/config/{rules,skills,agents,hooks.json}` mediante `hive install`

**Execution:** hilo principal

**Test approach:** check

**Changes:** antes de empezar, lee `_support/docs/harness-engineering/pilot-procedure.md` y registra la versión de `agy`, el modelo, la memoria cotidiana (Engram activo por el bloque de `~/.gemini/GEMINI.md`; la llamada no pide escribir memoria y se revisa que no lo haga) y el directorio de estado real de Hive. Construye el candidato en un binario temporal, corre `hive install --hosts antigravity` con ese binario, y hace hasta 2 llamadas a `agy` con el modelo más barato (la tercera la usó T0). Si `agy` no cargó una parte, o si `view_file` corta la referencia más grande de Hive (`[Truncated at line …]`), lo reporta como bloqueo antes del push. Reversión: `hive remove --hosts antigravity` con el mismo binario temporal; tras el merge, el refresco local del mantenedor reinstala con el binario integrado.

**Verification:** `agy -p` cita el primer y el último encabezado `##` de la guía sin herramientas (salida `stream-json` sin pasos de herramienta); `agy agents` lista cada rol de Hive una vez; `hive view work list` o el visor muestra la sesión de `agy` con host `antigravity`; Engram no recibe registros de la prueba. Prerrequisito: la autorización D5 A.

## Verificación y revisión humana

| Comprobación | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, local | El comando de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, local | `go test ./...`, `go test -race ./...`, `go vet ./...`, pruebas del frontend del visor | Cubierta por la implementación |
| Comprobación en vivo | Candidato final, `~/.gemini/config` real | T6 | La respuesta de la ronda de entrega |
| Revisión de código | Candidato final, antes del push | `/code-review` de Claude Code | La respuesta de revisión (D7 A) |

Sin parada de validación humana: el modo es `automatic` (D6 A).

## Estado de la revisión y progreso

2026-11-06 — Plan escrito.
2026-11-07 — Revisión del plan aplicada: T0 nuevo, T1–T6 ajustadas (detalle en `proposal.md`).

## Checkpoint

Etapa: build · worktree `.claude/worktrees/antigravity-host-support` · rama `feat/antigravity-host-support`. Siguiente: T6.

## Progreso

- T0 hecha (2026-11-07, sin commit de código): `agy agents` listó `hive-probe` desde `~/.gemini/config/agents/`, y `agy -p` (`gemini-3.8-flash-low`, sin pasos de herramienta) citó el marcador de la regla `always_on` de `~/.gemini/config/rules/`; archivos y carpetas de prueba borrados. Llamadas usadas: 1 de 3.
- T1 hecha (9080486c): host registrado, enlaces de skills relativos y detección de `agy`; `doctor` sin la línea en blanco entre dos secciones «Not checked» para caber en 80x24 con siete hosts. `go build`, `go vet` y `go test ./...` pasan (`tooling/cli` 206,7 s, una corrida previa del hijo falló sin nombre y no se repitió). Pendiente: el esfuerzo de la versión de `agy` en `doctor` con cadenas largas, sin verificar.
- Verificación de T1 (`hive-verify-task`): AC1 y AC2 cumplidos en su parte. Arreglado: la columna de `doctor` se alinea al nombre de host más largo (antes `antigravity` desalineaba la fila). Queda menor: `TestAntigravityState…IsCreatedAndCleanedUp` no comprueba que `.gemini/config` desaparezca (la conducta sí lo hace).
- T2 hecha (commit «feat(antigravity): deliver the guidance as size-bounded always-on rules»): cada parte es un registro `block` con `GlobalSource` y su número en la ruta (`rules/hive-NN.md`), sin campo nuevo en el estado; corte por `##` fuera de bloques de código con 22.000 bytes y reserva fija de 2.500 bytes de voz en la parte 1; un `hive-NN.md` ajeno no vacío detiene la instalación. Guía real en `~/`: 2 partes (16.975 y 21.374 bytes); con rutas de home largas salen 3, porque los enlaces de skills se reescriben a rutas absolutas. La unión se cumple como `strings.Join(cuerpos, "\n\n")` porque el bloque gestionado quita los saltos finales. `TestModelsViewFitsAndScrollsWithManyRolesOnSixCLIs` saltaba los roles de Antigravity hasta T3.
- Verificación de T2 (`hive-verify-task`): AC1 y AC2 cumplidos en su parte; probados crecer a 4 partes, achicar a 1, voz con 40 caracteres de 4 bytes, un estado previo al cambio y `hive recover` tras un fallo de escritura. Menor pendiente: si la guía vuelve a crecer hasta una parte sobrante que el usuario editó, la instalación se detiene con «unowned Antigravity rule file» sin decir que se renombre.
- T3 hecha (commit «feat(antigravity): render Hive roles with the inherited model»): perfil `antigravity` opcional con `inherit` en cada rol, sin esfuerzo ni campos de acceso; 19 roles en `agents/`; lista fija `inherit`, `flash_lite`, `flash`, `pro` sin comando de listado; la prueba de la vista de modelos pasa a `…OnEveryCLI`. `go test ./... -count=1`: 22 paquetes `ok`.
- T4 hecha (commit «feat(antigravity): report agy sessions to the viewer»): hook con nombre `hive` (`SessionStart`, `PreInvocation`, `Stop`, timeout 5 s) mediante el mismo editor JSON con contenedor configurable; `Stop` responde `{"decision":"stop"}` y los demás `{}` en todo camino; `Stop` nunca termina la sesión; `agy` → `antigravity` en identidad y línea de tiempo; bundle del visor regenerado (`index-DTSo08Dd.js`). El brief del flujo delivery pasó de ~9.999 a 10.013 bytes: la prueba mantiene el límite de 10.000 de la spec para build (9.037) y da 10 KiB a los demás. Sin verificar en vivo: la respuesta de `Stop` y el nombre de proceso `agy` (T6). `go test ./... -count=1`: 22 paquetes `ok`; frontend 773 pruebas, `tsc`, `biome` y `dist verify` limpios.
- T5 hecha (commit «docs: document Antigravity as the seventh host»): 19 archivos; `global.md` pasa de 37.738 a 37.751 bytes (+13, solo la línea de sintaxis de comandos; `/flow-x` no se repite por host en otro lugar del contenido distribuido); lo no verificado para Antigravity (compactación, tarjeta de preguntas, lanzamiento de roles, alcance fuera del home) queda escrito como tal. Sin cambios a propósito: HA-IF-18 (`instruction-files.md`, `project-layout.md`) es solo de Cursor, e `index.mdx` no lista hosts. `go test ./tests/... -count=1` ok; `npm run build` en `site/` ok (17 páginas, con el `node_modules` del checkout principal enlazado y retirado).
