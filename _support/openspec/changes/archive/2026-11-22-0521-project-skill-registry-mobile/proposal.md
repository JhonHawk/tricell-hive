# project-skill-registry-mobile — Registro de skills por proyecto, rol móvil neutral y verificación en dispositivo

| Field | Value |
| --- | --- |
| Estado | Completado · PR #581, merge 846198ea |
| Tracker | none (sin ticket; relacionado con #553, que excluye «un setup por proyecto que detecte el stack») |
| Git | tricell-hive · base `development` en `867f5463` · rama `feat/skill-registry-mobile` · worktree `.claude/worktrees/skill-registry-mobile` |
| Playbook | none (cruza feature de guía, tooling y verificación) |

## Decisions

2026-11-21
- **Investigación A** — ¿Qué hacemos tras el research? Planear en un solo cambio el registro de skills en AGENTS.md, `hive-build-mobile` en lugar de `hive-build-kmp` y la verificación en dispositivo.
- **Dónde A** — ¿Dónde se trabaja? En un worktree propio, porque otras sesiones hacen commits en el checkout principal.
- **D1 A** — ¿Dónde vive el registro? En una sección `## Skills` del `AGENTS.md` del repositorio (y del workspace, con el prefijo del repositorio), fuera de `## Hive`.
- **D2 A** — ¿Quién lo escribe y mantiene? El onboarding de `harness-audit`, más un script con pruebas que compara el registro con las skills instaladas; Hive no instala skills por sí mismo.
- **D3 A** — ¿El onboarding propone las skills oficiales del stack detectado? Sí: detecta el stack por archivos, propone las fuentes oficiales de una tabla, instala solo las que el usuario aprueba tras la revisión de paquetes de 7 días. Revierte la exclusión de #553 para este caso.
- **D4 A** — ¿Se adopta ya que el encargo de cada subagente copie las líneas de `## Skills` que aplican, con un solo caso observado? Sí; el commit nombra el caso de project-g y la entrada vigilada se actualiza al cerrar.
- **D5 A** (build, 2026-11-21) — ¿Cómo reconoce `check` las skills que Hive instaló en el proyecto, si su propiedad vive en el estado del manager de cada máquina? Con una marca dentro de la carpeta de cada skill: cada skill de `content/skills/` lleva `assets/hive-managed.json` (`{"managed_by": "tricell-hive"}`), que el manager ya despliega con `assets/`; da el mismo resultado en cualquier máquina. Se implementó como archivo fijo y no generado por el manager, sin código Go nuevo; una prueba de contenido exige la marca en cada skill.
- **D6 A** (build, 2026-11-21) — ¿Se acepta que el ciclo de plan y build lea 353 bytes más por la viñeta del registro en `child-brief.md` y `reviewer-brief.md`? Sí: `reading_footprint_test.go` declara `skillRegistryAllowance = 353`.
- **D7 A** (verify, 2026-11-21) — ¿Con qué dispositivo se observa AC7, si adb ve un WDY-LX3 y no el Pixel? Con el WDY-LX3 (serial `AXLKCP4403411088`, Android 14): la prueba comprueba el procedimiento de adb, que no depende del modelo.
- **Supuesto** — La verificación en dispositivo es una sección «Device tool» de `ui-verification.md`, no una hoja nueva; descartado: una hoja nueva, que necesitaría una línea de índice en `global.md` (le quedan 383 bytes). Se revierte moviendo la sección a una hoja.
- **Supuesto** — El script es Python de biblioteca estándar bajo `harness-audit/scripts/` con la tabla en un JSON junto a él; descartado: un chequeo de `hive doctor` en Go. Se revierte portando la lógica a `doctor_project.go`.
- **Supuesto** — Los revisores (`hive-review-code`, `hive-review-ux`) también reciben las líneas del registro que aplican, porque son convenciones del proyecto y no argumentos a favor del trabajo; descartado: solo los constructores. Se revierte quitando la viñeta de `reviewer-brief.md`.
- **Supuesto** — Hive no instala ni diagnostica herramientas de dispositivo (adb, simuladores, mobile-mcp, Maestro): se usan si están instaladas y el encargo asigna el dispositivo por su serial; descartado: dependencia gestionada como `agent-browser`. Se revierte agregando su chequeo a `hive doctor`.
- **Supuesto** — Para iOS la tabla no propone skills: Apple no publica skills oficiales y D3 pide fuentes oficiales; descartado: incluir skills de la comunidad. Se revierte agregando una fila marcada como comunidad.
- **Supuesto** (revisión del plan) — `check` no cuenta las skills que instaló Hive con `--scope project`, identificadas por el registro de propiedad del manager; descartado: listarlas en `## Skills`, que llenaría el registro de decenas de skills de Hive. Se revierte quitando el filtro.
- **Supuesto** (revisión del plan) — En el workspace, también las rutas del disparador llevan el prefijo del hijo (`repoA/apps/mobile/`), porque desde la raíz del workspace los archivos están bajo esa carpeta; descartado: prefijar solo la ruta de la skill.
- **Supuesto** (revisión del plan) — La sección «Device tool» prohíbe desinstalar o borrar datos de un dispositivo físico sin preguntar, obliga a restaurar los ajustes que cambie (como el modo oscuro) y prefiere un emulador cuando existe; descartado: dejarlo al criterio del rol.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «talvez necesitemos un agente neutral para mobile, y que use la skill (si se tiene) para la tecnologia con la que se implementa»
>
> «podriamos hacer que las rutas de las skills relevantes este en el agents.md como registry para que no haya duda de donde leerlas y que no dependamos de que el agente le recuerde»
>
> «A, solo toma en cuenta que hay otras sesiones activas en este proyecto»

Sin ticket ni criterios previos. Fuera de alcance: el aislamiento de sesiones de navegador del caso vigilado `watch/child-brief-project-skills-and-browser-isolation`, la auditoría de accesibilidad nativa y la gestión de herramientas de dispositivo.

## Objective

Hoy Hive no tiene forma de que un proyecto declare qué skills de su stack deben leerse ni cuándo; depende de que la skill se active sola (no se invocó en el 56% de los casos en el eval de Vercel) o de que el hilo principal la mencione en el encargo (falló una vez con HeroUI). Desde un workspace, cinco de siete hosts no descubren las skills de los repositorios hijos. El único rol móvil es `hive-build-kmp` (0 usos) y la verificación de interfaz asume un navegador.

Al terminar: cada proyecto puede declarar en `## Skills` qué `SKILL.md` leer y en qué condición, `harness-audit` lo propone, lo escribe y un script lo verifica; el hilo principal copia esas líneas en cada encargo; `hive-build-mobile` cubre cualquier stack móvil leyendo esas skills; y la verificación de una app nativa usa el emulador o el dispositivo.

## Scope and acceptance

Dentro: script y tabla de fuentes, regla HA-IF-19 y procedimiento de onboarding, líneas del registro en los encargos, rol `hive-build-mobile` (retira `hive-build-kmp`), sección «Device tool» y sus menciones en roles y hojas de verificación, documentación de arquitectura, sitio y README. Fuera: instalar skills en este repositorio, cambiar `global.md`, gestionar herramientas de dispositivo, accesibilidad nativa.

- AC1 — `python3 -I content/skills/harness-audit/scripts/skill_registry.py check <repo>` sale con 1 y nombra, una vez cada uno, cada ruta de `## Skills` que no existe y cada skill instalada (carpetas con `SKILL.md` en `.agents/skills`, `.claude/skills` y `.grok/skills`, y entradas de `skills-lock.json`) cuyo nombre ninguna línea lista, sin contar las skills que instaló Hive; sale con 0 cuando coinciden y con 2, nombrando el archivo, ante una entrada que no puede leer · false today because el script no existe · Observed in: pruebas con repositorios temporales, y este worktree, donde nombra `reui` (dos copias reales, sin `## Skills`)
- AC2 — Sobre un workspace declarado (`AGENTS.md` con `## Repositories` o `## Repositorios`), `check` corre la comprobación de cada repositorio hijo y además falla nombrando cada línea de un hijo que falta en el `## Skills` del workspace con el prefijo de ese hijo en sus rutas · false today because el script no existe · Observed in: pruebas con un workspace temporal de dos repositorios
- AC3 — `skill_registry.py propose <repo>` imprime cada stack detectado con el archivo y el campo que lo mostró, y sus fuentes oficiales de `stack-skills.json`: Flutter, Expo y React Native, y Kotlin Multiplatform tapan a Android e iOS de sus carpetas envoltorio; `next`, `expo` y `@nestjs/core` cuentan solo en `dependencies`; los plugins de Gradle cuentan también por alias de `libs.versions.toml`; para iOS imprime que no hay skills oficiales, y sin stack reconocido lo dice y sale con 0 · false today because el script y la tabla no existen · Observed in: pruebas con un repositorio temporal por stack y con disposiciones reales (Flutter con `android/` e `ios/`, Expo tras prebuild, monorepo con varios `package.json`)
- AC4 — `harness-audit` define HA-IF-19 y su procedimiento de onboarding: correr `propose`, preguntar qué fuentes instalar, revisar el instalador con `check_package.py`, fijar cada fuente a un commit y leer cada `SKILL.md` antes de instalarla, instalar solo por rutas que escriben en las carpetas del proyecto, escribir `## Skills` fuera de `## Hive` y de los bloques gestionados por instaladores, repetir las líneas con prefijo en el workspace como excepción declarada a HA-HI-05 y HA-IF-08, y correr `check` · false today because `grep -c 'HA-IF-19' content/skills/harness-audit/references/instruction-files.md` da 0 · Observed in: el texto de `content/skills/harness-audit/` del candidato y las pruebas de contenido
- AC5 — `child-brief.md` exige copiar en el encargo de cada hijo, constructor o revisor, las líneas de `## Skills` cuyo disparador aplica a su tarea, y `agent-delivery.md` ya no afirma que no hace falta un registro por stack · false today because `child-brief.md` no menciona `## Skills` y `agent-delivery.md:123` dice «No per-language backend roles or additional stack routing registry are needed» · Observed in: el texto del candidato y el par de `single_home_test.go`
- AC6 — Tras el refresco local del mantenedor posterior al merge, cada host registrado de esta Mac tiene `hive-build-mobile` y ninguno tiene `hive-build-kmp` · false today because `ls ~/.claude/agents/hive-build-kmp.md ~/.codex/agents/hive-build-kmp.toml` los encuentra y `hive-build-mobile` no existe · Observed in: las carpetas de agentes de los hosts en esta Mac tras el refresco
- AC7 — Siguiendo la sección «Device tool» de `ui-verification.md`, una captura tomada del dispositivo asignado por su serial con `adb -s <serial> exec-out screencap -p` se publica con `hive view work capture --viewport mobile` y aparece en el visor, sin desinstalar ni borrar datos del dispositivo · false today because `ui-verification.md` abre con «Drive the browser» y no tiene sección de dispositivo · Observed in: el Pixel 10 Pro XL conectado por USB y el visor local
- AC8 — `ui-verification.md`, `before-captures.md`, `walkthrough.md`, `hive-verify-change`, `hive-review-ux` y `hive-design-ui` indican el camino de dispositivo para una app nativa; `hive-review-ux` y `a11y-audit.md` indican que la auditoría de accesibilidad de una app nativa se reporta como no ejecutada · false today because `ui-verification.md:5` solo indica el navegador, `hive-review-ux.md:12` y `hive-design-ui.md:15` dicen «Drive the browser», `hive-verify-change.md:12` dice «When the check drives a browser», `before-captures.md:5` y `walkthrough.md:7` asumen servir la app localmente, y `a11y-audit.md` solo cubre `agent-browser a11y` · Observed in: el texto del candidato

## Prerequisites

- [ ] Un dispositivo Android conectado y autorizado por adb · owner: el usuario (Pixel 10 Pro XL, autorizado el 2026-11-21 y desconectado al escribir el plan) · needed by: T5

## Verified context

- Fuentes oficiales por stack, existencia y actividad · `gh api repos/<org>/<repo>` · `nestjs/agent-skills` creado 2026-11-20; `android/skills` empujado 2026-11-17; `expo/skills` 2026-11-21; `flutter/agent-plugins` 2026-11-19; `vercel/next.js` con `skills/` (6 skills); `Kotlin/kotlin-agent-skills` 2026-10-30 · 2026-11-21
- CLI `skills` (vercel-labs) · `npm view skills time` · 1.9.0 del 2026-11-16 no pasa la regla de 7 días; 1.8.1 del 2026-11-11 sí · 2026-11-21
- Descubrimiento de skills por host desde un workspace · documentación oficial de los siete hosts y issues anthropics/claude-code #93403 y #102083 · solo Claude Code (perezoso, con Read/Edit/Write) y Cursor (sin verificar) descubren skills de subcarpetas · 2026-11-21
- Subagentes de Claude Code y skills · code.claude.com/docs/en/sub-agents · ven el catálogo y la herramienta de skills; `skills:` solo precarga · 2026-11-21
- Herramientas de dispositivo en esta Mac · `command -v adb xcrun; adb version; xcrun simctl list` · adb 1.0.41 presente; `simctl` no disponible (sin Xcode completo), así que el camino iOS queda sin observar · 2026-11-21
- `adb -s <serial> exec-out screencap -p` · not measured: el Pixel se desconectó antes de correrlo (`adb devices -l` vacío); lo cumple T5, que necesita el dispositivo conectado · 2026-11-21
- Formato de `skills-lock.json` del CLI `skills` · `gh api` sobre `vercel-labs/skills` `src/local-lock.ts` (lectura del revisor del plan) · v2 en la raíz del proyecto: `{version, skills: {<nombre instalado>: {source, sourceType, skillPath?, skillName?, computedHash, installMode?}}}`; el CLI también lee y migra la versión 1, así que `check` acepta 1 y 2 · 2026-11-21
- Carpetas que escribe el CLI `skills` en alcance de proyecto · `docs/commands.md` y `docs/agents.md` de `vercel-labs/skills` · `.agents/skills/<nombre>` y un enlace simbólico por skill en `.claude/skills/<nombre>`; Grok en `.grok/skills` · 2026-11-21
- Telemetría del CLI `skills` · README de `vercel-labs/skills` (research del 2026-11-21) · activa por defecto, se apaga con `DISABLE_TELEMETRY=1`; el procedimiento de instalación la apaga · 2026-11-21
- Rutas de instalación oficiales · docs.flutter.dev/ai/agent-skills, kotlinlang.org/docs/kotlin-ai-skills.html, docs.expo.dev/skills, README de `android/skills` y de `nestjs/agent-skills` (research) · Flutter, Kotlin, Expo y Next.js documentan `npx skills add <org>/<repo>`; Android documenta `android skills add --project` (escribe en las carpetas de cada agente detectado, sin `skills-lock.json`); NestJS documenta `npx @nestjs/skills add`. Las seis fuentes son repositorios de GitHub que el CLI `skills` instala; el plan permite solo esa ruta · 2026-11-21
- Fijar una fuente a un commit con el CLI `skills` · README de `vercel-labs/skills` (research del 2026-11-21) · acepta `owner/repo@<tag o sha>`; documental, no ejecutado: el plan no corre instaladores · 2026-11-21
- `check_package.py` · `check_package.py:5` · cubre npm, pypi, go y cargo; el CLI `skills` es npm, el CLI `android` no · 2026-11-21
- Margen del paquete de etapa de `flow-build` · `stage_manifest_test.go` (lectura del revisor del plan) · 22,471 de 22,500 bytes; `verification-floor.md` es miembro, así que el plan no lo edita · 2026-11-21
- Cómo identificar las skills que instaló Hive con `--scope project` desde el propio repositorio · not measured: el registro de propiedad vive fuera del repositorio (`tooling/management/ownership.go:108`, lectura del revisor); T1 busca una marca dentro de la carpeta instalada y, si no la hay, vuelve como decisión · 2026-11-21
- Descripción de `harness-audit` · HA-SK-01 · 1,022 de 1,024 caracteres: no cabe un disparador nuevo; el onboarding ya la activa · 2026-11-21
- `go vet ./... && go test ./...` en la base · pasa · ~260 s · 2026-11-21
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` en la base · pasa, 347 pruebas · ~83 s · 2026-11-21
- `python3 -m unittest discover -s tests/skills -p 'skill_registry_test.py'` · forma más cercana en la base: el mismo runner con `personal_rules_test.py` pasa; el archivo lo crea T1 · 2026-11-21
- `python3 -I content/skills/harness-audit/scripts/skill_registry.py check <repo>` · not measured: el script lo crea T1
- `go test ./tests/content/...` en la base · pasa (incluido en el suite) · 2026-11-21
- `go test ./tests/pilot/...` y `./tooling/management/...` en la base · pasan, ~87 s y ~148 s (incluidos en el suite) · 2026-11-21
- `grep -c 'HA-IF-19' …/instruction-files.md` en la base · 0 · 2026-11-21
- `grep -n 'stack routing registry are needed' _support/docs/architecture/agent-delivery.md` en la base · línea 123 · 2026-11-21
- `hive update --source <worktree> --dry-run` en la base · sale 0 en ~1 s; 0 menciones de `hive-build-mobile`; lista `hive-build-kmp` en Claude, Codex y OpenCode · 2026-11-21
- `hive view work capture --image FILE --viewport mobile` · la ayuda acepta PNG, JPEG o WebP y `--viewport mobile`; la publicación no se midió: la cumple T5 en la caminata · 2026-11-21

## Effects

None: the tasks take no effect on an external system.

## Delivery

`automatic` (2026-11-21): revisión, push, PR y merge sin parada de validación; el merge espera la aprobación pendiente de un revisor humano si la hay. Revisión de código: nativa, `/code-review` de Claude Code. Revisión de seguridad: sí, `hive-review-security` sobre la frontera de dependencias (instalación de skills de terceros). No autoriza: promover a `master`, instalar skills en este repositorio ni comentar en #553.

## Review status and progress

2026-11-21 — plan escrito; revisión por dominio: guía (2 P1, 8 P2, 8 P3; ~172k tokens, ~5 min) y servidor (2 P1, 5 P2, 4 P3; ~140k tokens, ~3.5 min).
2026-11-21 — aplicados: formato único y regla de lectura del registro (ambos revisores), qué cuenta como listada y exclusión de las skills de Hive, códigos de salida, detección con precedencia, rutas de instalación permitidas con commit fijo y lectura previa, excepción declarada a HA-HI-05/HA-IF-08, bloques gestionados intactos, protección de datos del dispositivo, UI rules en el rol, T5 sin `verification-floor.md` (tope del paquete), AC6 observado tras el refresco, AC8 con las frases reales, páginas del sitio que faltaban, prueba del detector de sesiones. 2026-11-21 — **D4 A**: el usuario aprueba ya la viñeta de `child-brief.md` con un solo caso observado (project-g, `e2c8d39f`, 2026-10-18), como permite «Measurement».
2026-11-21 — relectura de ambos revisores (servidor ~41 s, guía ~68 s): resueltos todos los P1 y P2 salvo P2-6 en parte; nuevos N1 a N3 y dos de guía (instaladores sin revisar o sin commit fijo, T2 corriendo instaladores, delta atrasado), aplicados: códigos de salida del workspace, exclusión de las skills de Hive con marca en la carpeta o decisión, una sola ruta de instalación (CLI `skills`), T2 sin instaladores, delta al día.
