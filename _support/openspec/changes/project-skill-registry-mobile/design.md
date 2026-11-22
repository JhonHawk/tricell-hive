# Design

## Approaches

Dónde vive el registro (D1):

1. **Sección `## Skills` en el `AGENTS.md` del proyecto** — elegida. Sigue lo que ya hacen el bloque gestionado de ReUI en este repositorio y el bloque que Next.js 16.3 escribe en `AGENTS.md`. Todos los hosts cargan `AGENTS.md` al arrancar en el repositorio, y funciona sin Hive. Costo: unas pocas líneas de contexto siempre cargadas.
2. **Clave `Skills:` en `## Hive`** — reutiliza `project-settings.md`, pero `## Hive` guarda ajustes y no instrucciones con condición, y solo serviría en proyectos con Hive.
3. **Archivo aparte enlazado desde `AGENTS.md`** (como el registro de gentle-ai) — agrega una lectura que el agente puede omitir, que es el fallo que el cambio busca evitar.

Quién lo mantiene (D2): el onboarding de `harness-audit`, que ya escribe `## Hive` (HA-IF-17), más un script. Se descartó un comando del manager (`hive skills`) por agregar maquinaria en Go sin una necesidad medida (AGENTS.md del repositorio, «Scope and preservation»).

Peldaño de reutilización: (2) el onboarding existente y la convención de scripts Python de `harness-audit` (`local_instructions.py`, `personal_rules.py`); (5) código nuevo solo para el script, porque ningún componente existente compara secciones de `AGENTS.md` con carpetas de skills.

## Decisions

- **Formato único**, el mismo en HA-IF-19, en el script, en el ejemplo y en el delta: `- <disparador observable>: `<ruta>/SKILL.md``. El disparador es una carpeta (`Before editing files under `apps/mobile/``) o un tipo de tarea (`Before an EAS build or a store release`) y puede contener dos puntos. La primera línea de la sección dice cómo leerla, así que un proyecto sin Hive la entiende igual.
- **Regla de lectura del script:** de cada elemento `- ` de la sección toma el token entre backticks que termina en `SKILL.md`; la ruta se resuelve desde la carpeta del `AGENTS.md`; rechaza rutas absolutas, con `~` o con `..`. Ignora un `## Skills` dentro de un bloque de código y acepta CRLF. Un `### Skills` o una clave `- Skills:` dentro de `## Hive` se reporta como mal ubicado.
- **Qué cuenta como listada:** una skill instalada está listada cuando alguna línea apunta a una carpeta con su nombre, comparando tras resolver enlaces simbólicos; cada nombre se reporta una sola vez aunque haya copias reales en varias carpetas (como `reui` en este repositorio). Una carpeta sin `SKILL.md` y las entradas ocultas o que no son carpetas no cuentan. Las skills que instaló Hive con `--scope project` no cuentan. Hive guarda su registro de propiedad fuera del repositorio (`tooling/management/ownership.go:108`), así que T1 busca primero una marca dentro de la carpeta instalada; si no la hay, la exclusión solo vale en máquinas con el manager y T1 se detiene para que el usuario decida. Una entrada de `skills-lock.json` sin carpeta se reporta como «en el lock pero no instalada» (sale 1).
- En un workspace declarado, el `AGENTS.md` del workspace repite cada línea de cada hijo con el prefijo del hijo en la ruta de la skill y en las rutas del disparador (`repoA/apps/mobile/`, `repoA/.agents/skills/...`), porque Codex, OpenCode, Pi, Grok y Antigravity no descubren skills de subcarpetas. Es una excepción declarada a HA-HI-05 («relocate, never by copying») y HA-IF-08 (duplicados entre niveles), que HA-IF-19 nombra.
- El script tiene dos subcomandos: `check <path>` (repositorio o workspace; en el workspace también corre la comprobación de cada hijo) y `propose <repo>`. Códigos de salida: 0 coincide o nada que reportar; 1 hallazgos; 2 entrada ilegible (la ruta objetivo no existe o no es carpeta; un `skills-lock.json` inválido, con marcadores de conflicto o con versión distinta de 2), con un mensaje que nombra el archivo y sin traceback, como `local_instructions.py`. Un repositorio sin `AGENTS.md` y sin skills instaladas sale 0; sin `AGENTS.md` pero con skills instaladas, es un hallazgo (sale 1). En un workspace, un hijo no clonado se reporta y se salta; si en una corrida hay a la vez hallazgos y una entrada ilegible, gana 2 y se imprimen ambos. `check` solo lee; escribir la sección lo hace el agente en `harness-audit` tras la aprobación del usuario, como ya hace con `## Hive`. Lee `## Repositories` cargando `workspace-repo/scripts/workspace_repo.py` por ruta con `importlib`, como `personal_rules.py`.
- **Detección del stack** (`propose`): profundidad máxima limitada y sin entrar en `node_modules`, `Pods`, `build` ni `.git`; Flutter, Expo/React Native y Kotlin Multiplatform tapan a Android e iOS de sus carpetas envoltorio; `next`, `expo` y `@nestjs/core` cuentan solo en `dependencies`; los plugins de Gradle cuentan por id literal o por alias de `libs.versions.toml`; cada stack se imprime con el archivo y el campo que lo mostró.
- `stack-skills.json` vive en `harness-audit/assets/` y guarda, por stack, las fuentes oficiales (organización y repositorio), la ruta de instalación permitida (el nombre del paquete instalador, no una versión fija; la versión la elige `check_package.py`) y la fecha de revisión de la tabla. Una prueba valida su esquema sin depender del reloj.
- **Instalación permitida** (`stack-skills.md`): una sola ruta, el CLI `skills` en alcance de proyecto (`npx skills@<versión revisada> add <org>/<repo>@<commit>`), porque es un paquete npm que `check_package.py` revisa, fija la fuente a un commit y escribe en `.agents/skills` con un enlace por skill en `.claude/skills`; las seis fuentes oficiales son repositorios de GitHub que ese CLI instala. Los CLIs de cada fabricante (`android skills add`, `@nestjs/skills`), los plugins de Claude Code y las skills de usuario quedan fuera, porque no se pueden revisar con `check_package.py` o no escriben en el proyecto. Antes de instalar: revisar el instalador, fijar el commit, leer cada `SKILL.md` y sus scripts, y correrlo con `DISABLE_TELEMETRY=1`. El plan no corre ningún instalador: la primera instalación real ocurre en el onboarding de un proyecto, con la aprobación del usuario.
- Los bloques gestionados por un instalador (como el de ReUI) no se reubican: la línea del registro se agrega aparte, porque la siguiente ejecución del instalador revertiría la reubicación.
- `hive-build-mobile` no detecta stacks para instalar nada: detecta el stack solo para elegir la herramienta de build y leer las líneas del registro que aplican. No es el «setup por proyecto» que excluye #553; la propuesta de skills sí lo es, y D3 la acepta.
- El dispositivo lo asigna el encargo por su serial; el rol usa `adb -s <serial>` y nunca un nombre fijo como `emulator-5554`. Así no necesita una excepción a la cláusula de servicios del pie de roles.

## Proposed shape

```
content/skills/harness-audit/
  scripts/skill_registry.py      check | propose
  assets/stack-skills.json       tabla de fuentes oficiales con fecha de revisión
  references/instruction-files.md  + HA-IF-19 — Skills registry
  references/stack-skills.md     procedimiento: propose → preguntar → check_package → instalar → escribir → check
  SKILL.md                       onboarding incluye HA-IF-19
content/skills/hive-core/references/
  child-brief.md                 + viñeta: copiar las líneas de ## Skills que aplican
  reviewer-brief.md              + lo mismo para revisores
  ui-verification.md             + ## Device tool
content/agents/development/hive-build-mobile.md   reemplaza hive-build-kmp.md
```

Ejemplo de sección:

```markdown
## Skills

Before the work a line names, read that skill's SKILL.md whole.

- Before editing files under `apps/mobile/`: `.agents/skills/expo-router/SKILL.md`
- Before an EAS build or a store release: `.agents/skills/eas-app-stores/SKILL.md`
```

La misma sección en el `AGENTS.md` del workspace, para el hijo `app`:

```markdown
- Before editing files under `app/apps/mobile/`: `app/.agents/skills/expo-router/SKILL.md`
```

## Impact

- `hive-build-kmp` lo nombran solo `child-brief.md:14` y `agent-delivery.md:41`; el manager descubre los roles recorriendo `content/agents`, así que retirarlo borra su destino en cada host (`tooling/management/catalog_test.go:64-100`). Un override de modelo del usuario sobre `hive-build-kmp` queda huérfano: se anota en el CHANGELOG.
- `global.md` no cambia (29,617 de 30,000 bytes).
- `hive doctor` deja de leer `## Hive` en el siguiente encabezado, así que `## Skills` no lo afecta (`doctor_project.go:358-404`).
- No toca producción.

## Recovery

Revertir el merge deshace todo; un `hive update` posterior restaura `hive-build-kmp` en los hosts. Las secciones `## Skills` que el onboarding haya escrito en otros proyectos quedan como texto inerte para Hive y siguen sirviendo como instrucción del proyecto.
