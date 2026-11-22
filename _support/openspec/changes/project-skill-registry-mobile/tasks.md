# Tasks

Base `development` en `867f5463`, rama `feat/skill-registry-mobile`, worktree `.claude/worktrees/skill-registry-mobile`. T1 a T5 corren en el hilo principal: T2 a T5 editan hojas que se citan entre sí y no pueden correr en paralelo sin solaparse, y T1 fija el contrato que T2 describe. T6 (documentación para personas) va a `hive-write-docs`, porque sus fuentes quedan fijadas por T1 a T5.

### T1 — Script `skill_registry.py` con `check` y `propose`, y la tabla `stack-skills.json`

- [x] `check` compara `## Skills` con las skills instaladas en un repositorio y en un workspace declarado; `propose` imprime las fuentes oficiales del stack detectado.

**Closes:** AC1, AC2, AC3

**Depends on:** none

**Locations:** `content/skills/harness-audit/scripts/skill_registry.py` (nuevo), `content/skills/harness-audit/assets/stack-skills.json` (nuevo), `tests/skills/skill_registry_test.py` (nuevo), `content/skills/*/assets/hive-managed.json` (nuevos, 28, por D5 A), `tests/content/hive_managed_marker_test.go` (nuevo)

**Execution:** hilo principal

**Test approach:** check — pruebas unitarias con repositorios temporales, con los valores esperados escritos en la prueba y no calculados con la tabla:
- formato: el ejemplo exacto de HA-IF-19; un disparador con dos puntos; ruta absoluta, con `~` o con `..` rechazada; `## Skills` dentro de un bloque de código ignorado; CRLF; `### Skills` o `- Skills:` dentro de `## Hive` reportado como mal ubicado;
- listado: ruta inexistente; skill en `.agents/skills`, en `.claude/skills` como enlace por skill y como carpeta enlazada entera, y en `.grok/skills`; copias reales en dos carpetas reportadas una vez (la disposición de `reui` aquí); enlace roto; carpeta sin `SKILL.md`; `.DS_Store`; entrada de `skills-lock.json` sin carpeta; skills instaladas por Hive excluidas; todo coincide (sale 0);
- workspace: `## Repositories` y `## Repositorios`; línea con prefijo que falta; hijo con skills y sin `## Skills` reportado por su propia comprobación; hijo no clonado reportado y saltado; hallazgos y entrada ilegible en la misma corrida (sale 2 e imprime ambos);
- errores (sale 2 con el archivo nombrado y sin traceback): lock inválido, con marcadores de conflicto o versión distinta de 2; ruta objetivo inexistente o que no es carpeta;
- sin `AGENTS.md`: sin skills instaladas sale 0; con skills instaladas sale 1; entrada del lock sin carpeta reportada como «en el lock pero no instalada»;
- `propose`: por stack, el stack detectado y una fuente escrita en la prueba (por ejemplo `flutter/agent-plugins`); Flutter con `android/` e `ios/Runner.xcodeproj` propone solo Flutter; Expo tras prebuild propone solo Expo; Kotlin Multiplatform con `iosApp/*.xcodeproj`; plugin de Android por alias de `libs.versions.toml`; `next` solo en `devDependencies` no cuenta; monorepo con varios `package.json`; `.xcodeproj` dentro de `node_modules` o `Pods` no cuenta; iOS imprime que no hay skills oficiales; sin stack reconocido lo dice y sale 0;
- `stack-skills.json`: esquema (claves requeridas, fecha de revisión ISO, organización, repositorio y paquete instalador por fuente), sin comparar la fecha con hoy.

**Changes:** biblioteca estándar solamente; lectura sin escrituras ni red. Carga `workspace-repo/scripts/workspace_repo.py` por ruta con `importlib` para leer `## Repositories`, como `personal_rules.py:58` carga su dependencia. Formato, regla de lectura, qué cuenta como listada, códigos de salida y detección: los de `design.md` «Decisions». Mide primero, leyendo el código del manager y una instalación de prueba en un directorio temporal con `go run ./tooling/cli plan install --scope project` (vista previa de solo lectura), si una carpeta de skill instalada por Hive lleva una marca propia; si no la lleva, se detiene y vuelve como decisión, porque la exclusión solo valdría en máquinas con el manager.

**Verification:** `python3 -m unittest discover -s tests/skills -p 'skill_registry_test.py'` pasa; `python3 -I content/skills/harness-audit/scripts/skill_registry.py check /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/skill-registry-mobile` sale con 1 y nombra `reui` una sola vez.

### T2 — HA-IF-19 y el procedimiento de onboarding en `harness-audit`

- [x] `harness-audit` revisa y escribe `## Skills`, propone las skills oficiales del stack y verifica con el script.

**Closes:** AC4

**Depends on:** T1

**Locations:** `content/skills/harness-audit/references/instruction-files.md`, `content/skills/harness-audit/references/stack-skills.md` (nuevo), `content/skills/harness-audit/SKILL.md`, `content/skills/harness-audit/references/apply.md`, `content/skills/workspace-repo/SKILL.md`, `tests/content/single_home_test.go`

**Execution:** hilo principal

**Test approach:** check — pruebas de contenido existentes (tamaños, enlaces) más un par nuevo en `single_home_test.go` para el formato de `## Skills` con su casa en `instruction-files.md`.

**Changes:** HA-IF-19 define el formato y la regla de lectura de `design.md`, la ubicación fuera de `## Hive` y de los bloques gestionados por instaladores, la repetición con prefijo en el workspace como excepción declarada a HA-HI-05 y HA-IF-08, y que una instrucción en prosa «usa la skill X» fuera de un bloque gestionado se reubica en la sección. `stack-skills.md` ordena: `propose`; preguntar qué fuentes instalar como una decisión; revisar el CLI `skills` con `check_package.py`; fijar cada fuente a un commit; leer cada `SKILL.md` y sus scripts; instalar con la única ruta permitida de `design.md` y `DISABLE_TELEMETRY=1`; escribir la sección; correr `check`. La tarea no corre ningún instalador: lo que escribe cada ruta se toma de la documentación ya leída. El párrafo de onboarding de `SKILL.md` agrega HA-IF-19; su `description` no cambia (1,022 de 1,024 caracteres). `apply.md` agrega `## Skills` como sección que se puede escribir. `workspace-repo/SKILL.md` menciona las líneas con prefijo. No toca `global.md` ni `project-settings.md`.

**Verification:** `go test ./tests/content/...` pasa; `grep -c 'HA-IF-19' content/skills/harness-audit/references/instruction-files.md` da 1 o más.

### T3 — Líneas del registro en los encargos y documentación de entrega

- [x] El hilo principal copia las líneas de `## Skills` que aplican en el encargo de cada hijo, y la arquitectura deja de negar el registro.

**Closes:** AC5

**Depends on:** T2

**Locations:** `content/skills/hive-core/references/child-brief.md`, `content/skills/hive-core/references/reviewer-brief.md`, `_support/docs/architecture/agent-delivery.md`, `_support/docs/architecture/instruction-resources.md`, `tests/content/single_home_test.go`

**Execution:** hilo principal

**Test approach:** check — par nuevo en `single_home_test.go` para la viñeta de `child-brief.md`.

**Changes:** una viñeta en «Every child» de `child-brief.md` y una en `reviewer-brief.md` que remite a ella (`reviewer-brief.md` está en el paquete de salida de `flow-verify`, con 3,080 bytes libres); en `agent-delivery.md` se corrigen las líneas 61, 63, 65 y 123 y se cita el registro como la fuente de convenciones del proyecto; `instruction-resources.md:68` nombra las líneas del registro. `delegation.md` no cambia, porque está en el paquete de etapa de `flow-build`. El commit nombra la observación del caso vigilado (cam, sesión `e2c8d39f`, Claude Code, opus-5-5, 2026-10-18) y la entrada `watch/child-brief-project-skills-and-browser-isolation` se actualiza al cerrar.

**Verification:** `go test ./tests/content/...` pasa; `grep -n 'stack routing registry are needed' _support/docs/architecture/agent-delivery.md` no devuelve nada.

### T4 — Rol `hive-build-mobile` en lugar de `hive-build-kmp`

- [x] Existe `hive-build-mobile`, `hive-build-kmp` se retira, y las listas de roles y el auditor de sesiones lo reflejan.

**Closes:** AC6

**Depends on:** T3

**Locations:** `content/agents/development/hive-build-mobile.md` (nuevo), `content/agents/development/hive-build-kmp.md` (borrado), `content/skills/hive-core/references/child-brief.md`, `_support/docs/architecture/agent-delivery.md`, `tests/pilot/session_audit_trackers.go`, `tests/fixtures/sessions/try_it_block_in_report/` (casos nuevos), `_support/docs/harness-engineering/session-review-rubric.md`, `CHANGELOG.md`

**Execution:** hilo principal

**Test approach:** check — `agent_boundary_test.go` exige el pie de límites; `catalog_test.go` prueba que borrar un rol borra su destino; `TestAuditNewCriteria` con un caso que falla y otro que pasa lanzando `hive-build-mobile`.

**Changes:** el rol detecta el stack por archivos; lee las líneas de `## Skills` del encargo y del `AGENTS.md`; lee la precedencia de UI criteria, `ui-plan.md` y el `DESIGN.md` del proyecto, y aplica reuse-before-writing y los principios de código como `hive-build-frontend`; construye con la herramienta oficial del stack; verifica en el dispositivo o emulador que el encargo asigna por serial, siguiendo «Device tool»; no edita `.pbxproj` a mano; usa `adb reverse` para servidores locales; no prueba builds de release de Expo; conserva los puntos de Kotlin Multiplatform (código común, cancelación de coroutines, frontera con Swift, targets no ejercitados). Perfil `execution`, acceso `implement`, `effort_claude: medium`, pie de límites sin excepciones. `child-brief.md:14` y `agent-delivery.md:41,123` cambian el nombre; `launchesRole` acepta también `hive-build-mobile`; la fila 49 de la rúbrica anota que `uiSource` no cubre `.kt`, `.swift`, `.dart` ni layouts `.xml`, a propósito; el CHANGELOG anota el retiro y que un override de modelo sobre `hive-build-kmp` queda huérfano.

**Verification:** `go test ./tests/content/... ./tests/pilot/... ./tooling/management/...` pasa. AC6 se observa tras el refresco local del mantenedor posterior al merge.

### T5 — Camino de dispositivo en la verificación de interfaz

- [ ] `ui-verification.md` tiene «Device tool», y los roles y hojas que asumen navegador remiten a él para una app nativa.

**Closes:** AC7, AC8

**Depends on:** T4

**Locations:** `content/skills/hive-core/references/ui-verification.md`, `content/skills/hive-core/references/ui-verification/device-tool.md` (nuevo: la hoja tiene tope de 4,000 bytes), `content/skills/hive-core/references/before-captures.md`, `content/skills/hive-core/references/walkthrough.md`, `content/skills/hive-core/references/a11y-audit.md`, `content/agents/quality/hive-verify-change.md`, `content/agents/review/hive-review-ux.md`, `content/agents/design/hive-design-ui.md`

**Execution:** hilo principal

**Test approach:** check — pruebas de contenido, y la caminata real en el Pixel conectado.

**Changes:** la línea 5 de `ui-verification.md` se reescribe para separar app web (navegador, anchos) de app nativa (dispositivo, su pantalla), sin contradecir `draw.md:9`. «Device tool»: el dispositivo lo asigna el encargo por serial; prefiere un emulador cuando existe; nunca desinstala ni borra datos de un dispositivo físico sin preguntar y restaura los ajustes que cambie; Android con `adb -s <serial>` (instalar, lanzar, `exec-out screencap -p`, `reverse` para servidores locales); iOS con `xcrun simctl` cuando hay Xcode completo; `mobile-mcp` o Maestro si están instalados; las capturas se publican con `--viewport mobile`. `before-captures.md` y `walkthrough.md` agregan instalar y lanzar en el dispositivo como forma de servir una app nativa. `hive-review-ux.md:12` y `hive-design-ui.md:15` («Drive the browser») y `hive-verify-change.md:12` («When the check drives a browser») remiten a «Device tool» para una app nativa; el paso 2 de `hive-review-ux` y `a11y-audit.md` reportan la auditoría de accesibilidad de una app nativa como no ejecutada. `verification-floor.md` no cambia: está en el paquete de `flow-build` con 29 bytes libres y su «Try it» ya acepta «the URL or command to open it». `hive-build-frontend.md:17` no cambia: es solo web.

**Verification:** `go test ./tests/content/...` pasa; caminata, con el dispositivo conectado y su serial de `adb devices -l` (hoy `58240DLCQ003EW`): `adb -s <serial> exec-out screencap -p > <tmp>/device.png` y `hive view work capture --view-title "Pixel" --image <tmp>/device.png --viewport mobile --stage verify` publican la captura y el visor la muestra.

### T6 — Sitio, README y llms.txt

- [ ] La documentación para personas explica `## Skills`, el script, el rol móvil y la verificación en dispositivo.

**Closes:** none — documentación para personas de lo que cierran T1 a T5

**Depends on:** T5

**Locations:** `site/src/content/docs/getting-started.md`, `site/src/content/docs/skills.md`, `site/src/content/docs/index.mdx`, `site/src/content/docs/flows/verify.md`, `site/src/content/docs/flows/build.md`, `README.md`, `llms.txt`, `_support/docs/harness-engineering/rule-enforcers.md`

**Execution:** `hive-write-docs`, con las fuentes de T1 a T5 y esta carpeta

**Test approach:** check — `rule_enforcers_test.go` verifica que la fila nueva apunte a archivos existentes.

**Changes:** describe el comportamiento y enlaza la skill, sin copiar sus reglas; `flows/verify.md:26` y `flows/build.md:20` agregan el caso de app nativa; agrega la fila del script en `rule-enforcers.md`.

**Verification:** `go test ./tests/content/...` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, en el worktree | El comando de la tarea pasa | Covered by the implementation |
| Full suite | Candidato final, en el worktree | `go vet ./... && go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Caminata de AC7 | Candidato final, con el Pixel conectado | Captura del Pixel publicada en el visor | Covered by the implementation |
| Code review | Candidato final, antes del push | `/code-review` de Claude Code | The review answer |
| Security review | Candidato final, antes del push | `hive-review-security` sobre la instalación de skills de terceros | The review answer |
| AC6 | Tras el merge, con el refresco local del mantenedor | Carpetas de agentes de los hosts en esta Mac | Marcador `maintainer.local` presente |

Sin parada de validación humana (`automatic`).

## Checkpoint

- Updated: 2026-11-21
- Stage: build · T1
- Last commit: 867f5463
- Open findings: none
- Next: T1: measure Hive-owned skill marker, then write skill_registry.py with tests
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/skill-registry-mobile · feat/skill-registry-mobile
- Rules: flow-build/stage.md

### Evidence

- before captures: none (no screen of an app changes)

## Review status and progress

2026-11-21 — T1 · ab99cc4f · unverified: none
2026-11-21 — T2 · 5add19bd · unverified: AC4 se observa en el texto; ninguna instalación real corrió (el plan no corre instaladores)
2026-11-21 — T3 · 10582c10 · unverified: none
2026-11-21 — T4 · 5911a832 · unverified: AC6 se observa tras el refresco del mantenedor posterior al merge
