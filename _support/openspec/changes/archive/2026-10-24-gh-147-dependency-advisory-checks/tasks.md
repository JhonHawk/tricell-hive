# Tareas

Base: `5894fbf4` (2026-10-24). Rama `feat/gh-147-dependency-advisory-checks`, worktree `.claude/worktrees/gh-147`.

Convención de las comprobaciones: `rg -F -c "<frase>" <archivo>` debe dar `1` después del cambio; en la base no hay coincidencias (búsqueda del 2026-10-24). Si el implementador ajusta el estilo de una frase, actualiza aquí la frase de comprobación antes de la verificación.

## T1 — Comprobación previa a instalar, con su hogar canónico

- [x] `security-boundaries.md` tiene la sección «Check a dependency before installing» y la viñeta de la línea 8 remite a ella.
Ronda 1: AC1–AC4 cumplidos (commit `7c83c748`). Reabierta por la revisión D9-A: el paso 5 trata los avisos ya presentes como hallazgo incidental. Ronda 2: AC1–AC4 cumplidos (commit adicional `(T1)`). Ronda 3, por los hallazgos de `/code-review`: AC1–AC4 cumplidos (commit adicional `(T1)`).

**Closes:** AC1, AC2, AC3, AC4.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-build/references/security-boundaries.md:8` y final del archivo.
**Execution:** hilo principal; la redacción está fijada en [design.md](design.md#security-boundariesmd) y el encargo sería más largo que el cambio.
**Test approach:** check.
- En `security-boundaries.md`, una coincidencia de cada frase:
  - AC1: «## Check a dependency before installing», «including a one-off run such as `npx`», «Check that the exact name exists in its registry».
  - AC2: «Query `POST https://api.osv.dev/v1/query`», «never query OSV with a private or internal package», «osv-scanner --offline --download-offline-databases -r <dir>», «including the project's own name and its workspace packages, is public», «Use the name the ecosystem publishes», «scan before installing».
  - AC3: «install that version and report the change», «MEDIUM or LOW with no fixed version: proceed and report», «CRITICAL or HIGH with no safe path, or a `MAL-` record: stop and ask the user before installing, every time», «with none, treat it as CRITICAL or HIGH».
  - AC4: «also for a one-off run whose runner does not apply it», «Never lower or bypass the cooldown without asking the user», «never `@latest` or no version».
  - Hijos delegados: «a delegated child stops and reports to its parent instead».
  - Avisos ya presentes (D9-A): «advisories on versions this install does not change are an incidental finding», «as the global rule on incidental findings sets», «ask the user before installing when one is CRITICAL or HIGH or a `MAL-` record», «When the install only restores an existing lockfile, scan it before installing», «When the scan could not run in an install that only restores the lockfile, say so and proceed».
  - Alcance y privacidad: «A package you name on the command or add to a manifest: all five steps», «An update with no package argument», «A one-off run or a global install, which has no project lockfile: steps 1, 2, 3 and 5», «A system package manager such as Homebrew, apt or winget is outside this section», «Use only a scanner that keeps private package names off third-party services», «Do not look the private name up in a public registry», «its public dependencies still go through step 4».
  - Revisión de código: «When that later scan finds what step 5 would ask about, stop», «the user's package-manager configuration», «holds for that unchanged lockfile».
- `rg -F "consult current authoritative advisories when version-specific risk matters" content/skills/flow-build/references/security-boundaries.md`: sin coincidencias.

**Changes:** redacción de [design.md](design.md#security-boundariesmd).
**Verification:**
- Los comandos del enfoque de prueba dan lo esperado.
- La lectura del archivo confirma tres cosas. La viñeta conserva los scripts de instalación, la exposición, los controles de lockfile, procedencia y avisos del proyecto, y la remediación automática. La frase «A clean advisory scan is not evidence of trusted origin» quedó cubierta por «An empty result means no known advisory, not a safe package». Y el alcance queda claro en los tres casos: una instalación con paquete nombrado pasa por los cinco pasos, una actualización sin argumento por los pasos 2, 4 y 5, y una restauración sin paquete nombrado solo por los pasos 4 y 5.

## T2 — Punteros desde `tools-and-dependencies.md` y `global.md` (D4-A)

- [x] Las dos secciones de `tools-and-dependencies.md` y una viñeta de `global.md` remiten a la comprobación sin repetir sus pasos.

**Closes:** AC5.
**Depends on:** T1 (el ancla y el nombre de la sección).
**Locations:** `content/skills/flow-build/references/tools-and-dependencies.md:9` y `:13`; `content/guidance/global.md:81`.
**Execution:** hilo principal; son tres frases.
**Test approach:** check.
- `rg -F -c "run [Check a dependency before installing](security-boundaries.md#check-a-dependency-before-installing)" content/skills/flow-build/references/tools-and-dependencies.md` da `2`.
- En `content/guidance/global.md`, una coincidencia de «read and run the section "Check a dependency before installing" in [security boundaries](skill:flow-build/references/security-boundaries.md)».
- `wc -c content/guidance/global.md`: anotar el tamaño antes (35 478 bytes) y después para el PR.

**Changes:** redacción de [design.md](design.md#tools-and-dependenciesmd).
**Verification:**
- Los comandos del enfoque de prueba dan lo esperado.
- La lectura de las tres frases confirma que solo nombran el disparador y la sección, sin repetir ninguno de sus pasos.
- `go test ./tests/content/... ./tooling/management/...` en verde (enlaces `skill:` de `global.md` y enlaces relativos de las referencias).

## T3 — Skill `security-advisory`

- [x] Existe `content/skills/security-advisory/SKILL.md` con el procedimiento de solo lectura.
Ronda 1: AC6 cumplido (commit `8f97f4ea`). Ronda 2, por el hallazgo 10 de `/code-review`: la skill remite a la regla global sobre otros proyectos en vez de repetirla. AC6 cumplido (commit adicional `(T3)`). Abrir tickets en otro proyecto queda cubierto por esa regla global y por la de efectos autorizados.

**Closes:** AC6.
**Depends on:** ninguna.
**Locations:** `content/skills/security-advisory/SKILL.md` (nuevo).
**Execution:** hilo principal; el texto está fijado en [design.md](design.md#contentskillssecurity-advisoryskillmd) y depende de reglas de esta conversación (datos a terceros, proyectos externos).
**Test approach:** check.
- `head -4 content/skills/security-advisory/SKILL.md` muestra `name: security-advisory` y una `description` con «Use when».
- En el archivo, una coincidencia de cada frase: «The sweep is read-only», «Read each advisory from a primary source», «delegate it to `hive-research` children», «not the manifest range», «launch a `hive-review-security` child», «the global rule on requests to third-party services», «web lookups carry only advisory IDs and public package names», «the evidence locator» y «under the global rule on changing only the current project».
- `wc -c content/skills/security-advisory/SKILL.md` da 16 000 o menos.

**Changes:** redacción de [design.md](design.md#contentskillssecurity-advisoryskillmd).
**Verification:**
- Los comandos del enfoque de prueba dan lo esperado.
- `go test ./tests/content/... ./tooling/management/...` en verde (tamaño de skills, validación de rutas y enlaces). `tooling/management/plan.go:230` toma toda `SKILL.md` nueva, así que no hace falta una prueba de despliegue aparte.

## T4 — Sitio y CHANGELOG

- [x] Las dos listas de skills de apoyo y el CHANGELOG nombran la skill y la comprobación.

**Closes:** AC7.
**Depends on:** T1, T3.
**Locations:** `site/src/content/docs/skills.md:17-26`, `site/src/content/docs/index.mdx:79-90`, `CHANGELOG.md:11-13`.
**Execution:** hilo principal; tres viñetas.
**Test approach:** check.
- `rg -c "security-advisory" site/src/content/docs/skills.md site/src/content/docs/index.mdx CHANGELOG.md`: al menos `1` en cada archivo.
- En `CHANGELOG.md`, una coincidencia de «Check a dependency before installing» y de «The `security-advisory` skill checks, read-only».

**Changes:** redacción de [design.md](design.md#sitio-y-changelogmd).
**Verification:**
- Los comandos del enfoque de prueba dan lo esperado.
- `pnpm build` en `site/` termina sin errores.

## Verificación y revisión humana

| Comprobación | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | Tras cada tarea, working tree de la rama | `hive-verify-task` con los criterios que cierra; veredicto por criterio | Parte de la implementación |
| Suite local | Candidato final | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` en verde | Parte de la implementación |
| Sitio | Candidato final | `pnpm build` en `site/` sin errores | Parte de la implementación |
| Revisión de código | Antes del PR, diff de la rama contra `development` | `/code-review` de Claude Code (D6-A); hallazgos atendidos o refutados con evidencia | Autorizada (D6-A) |
| Validación humana | Parada antes del push (D5-A) | El usuario lee el diff y, si quiere, prueba una instalación con la rama desplegada a pedido | Pendiente del usuario |

No hay efecto de UI ni en vivo: el cambio es guía en Markdown y dos páginas del sitio sin cambio de diseño.

## Estado de la revisión y progreso

Revisión del plan, ronda 1, sobre la revisión `fa65aec09aa3`, con dos `hive-review-plan` en paralelo:

- **Guía distribuida** (~134 k tokens, 154 s). Un bloqueante y seis menores, todos aceptados:
  - B1: AC2 no coincidía con la redacción.
  - N1: los hijos delegados no pueden preguntar.
  - N2: alcance de los pasos para una restauración del lockfile.
  - N3: se recuperan los «advisory controls» del proyecto.
  - N4: la skill remite a `delegation.md` en vez de repetirlo, y no delega una lectura corta.
  - N5: comprobaciones de AC5 y AC6, AC7 reformulado y texto del CHANGELOG fijado.
  - N6: se quita la prueba de despliegue de T3.
- **Seguridad de la regla** (~104 k tokens, 151 s). Dos bloqueantes y tres menores, todos aceptados:
  - 1: la auditoría del gestor solo se usa con todas las dependencias del registro por defecto.
  - 2: un aviso sin calificación cuenta como CRITICAL/HIGH.
  - 3: el lockfile se escanea antes de instalar cuando el gestor lo permite.
  - 4: ejecuciones puntuales con versión exacta y periodo de espera; una corrección rechazada por el periodo cuenta como sin camino seguro.
  - 5: «nuevo o con versión cambiada», y trato del paquete privado.
  - Notas: `<project-dir>`; los hijos de la skill reciben la línea de red.

Ronda de corrección (única), sobre `1742a28686d7`, retomando a los mismos revisores:

- **Guía distribuida** (~146 k tokens, 34 s). Sin bloqueantes y sin criterios huérfanos, sin cubrir o ya ciertos en la base. Quedaban dos huecos:
  - N2': una instalación con paquete nombrado parecía saltarse los pasos 4 y 5. Se aplicó su propia redacción («Every install takes steps 4 and 5»).
  - N2'': una actualización sin argumento no encajaba en ningún caso. Decisión del orquestador: pasa por los pasos 2, 4 y 5. No tiene revisión independiente y se presenta al usuario con el plan.
- **Seguridad** (~112 k tokens, 44 s). Sin bloqueantes. Dos hallazgos, corregidos con su propia propuesta:
  - El escáner del proyecto también debe mantener los nombres privados fuera de terceros.
  - Un paquete privado debe resolver desde el registro privado de su scope o índice, y su nombre no debe resolver desde el registro público (confusión de dependencias). Sus dependencias públicas siguen el paso 4.

Límites: nada se ejecutó. No se confirmó con fuente primaria que la API de OSV y sus bases offline incluyan registros `MAL-`, ni si `pnpm install --lockfile-only` descarga paquetes.

Build (2026-10-24):

- **Ejecución:** las cuatro tareas se hacen en el hilo principal. El implementador es el modelo de la sesión, `claude-opus-5-5`, según el entorno.
- **Verificadores:** `hive-verify-task` con su perfil configurado; su modelo no se observó.
- **Commits:** T1 `7c83c748`, T3 `8f97f4ea` y T2 `75db9be8`, cada uno tras su veredicto `met`.
- **T4, build del sitio (D8-A):** `pnpm audit` del lockfile del sitio dio 33 avisos: 11 altos, 16 moderados y 6 bajos, todos en dependencias transitivas de astro, wrangler/miniflare y starlight-ion-theme. `extract-zip` ≤2.0.1 no tiene corrección. Se restauró igual porque el checkout principal ya tiene instalado ese mismo lockfile. `pnpm build` terminó bien (10 páginas). Queda como hallazgo incidental, con un issue propuesto.
- **Revisión D9-A:** cambia AC3 y el paso 5 de T1. En una restauración, los avisos ya presentes son un hallazgo incidental y no se cambian versiones. T1 vuelve a `[/]`. El revisor de seguridad hizo una ronda sobre la corrección (~118 k tokens, 32 s). Encontró tres cosas y se aplicó su propia redacción:
  - La regla de hallazgo incidental cubre cualquier instalación cuando la versión con el aviso no cambia.
  - El caso de una restauración que no pudo escanearse pasa a esa frase.
  - Se agregan un puntero a la regla global de hallazgos incidentales y el escaneo antes de una restauración.
- **T4:** `met` (AC7); commit tras el veredicto.
- **Suite final:** `go vet ./...` y `go test ./...` en verde, y `unittest` con 64 pruebas OK, sobre el candidato final.
- **Hueco que deja la verificación de T1, sin corregir:** si el gestor no puede actualizar el lockfile sin instalar, el escaneo corre después de la instalación, y entonces «ask the user before installing» ya no se puede cumplir. La regla pide reportar que el escaneo vino después. Queda como propuesta para tu validación.
- **Sin revisión de seguridad dedicada:** el cambio es guía en texto y no cruza una frontera de `security-boundaries.md` en código.
- **Validación (D10-A, 2026-10-24):** aprobaste y publicar.
- **`/code-review` de Claude Code** sobre la rama (~120 k tokens, 210 s): 10 hallazgos, todos aceptados y corregidos en T1 y T3.
  1. El paso 1 enviaba nombres privados a un registro público.
  2. El escaneo después de instalar no tenía una acción que se pudiera cumplir.
  3. Ejecuciones puntuales sin lockfile.
  4. Gestores de sistema (brew, apt).
  5. `osv-scanner` necesita `-r` y el directorio del lockfile.
  6. Periodo de espera en la configuración global.
  7. La respuesta del usuario sobre avisos ya presentes vale para la sesión.
  8. Formato del nombre de paquete para OSV.
  9. La auditoría del gestor puede enviar el nombre del proyecto y de sus workspaces.
  10. La skill repetía la regla sobre otros proyectos.
  Se descartó mover la sección a otra referencia, una sugerencia de eficiencia que no está entre los 10.
- **T1 y T3 verificadas de nuevo:** `met`. La suite final y el build del sitio se corrieron otra vez sobre el candidato.
- **Entrega:** [PR #148](https://github.com/JhonHawk/tricell-hive/pull/148), con merge `b93af714` a `development` el 2026-10-24. No hay CI. `hive update` desplegó `b93af714` (`hive apply` con un plan guardado, porque `hive update` pide confirmar en una terminal interactiva).
