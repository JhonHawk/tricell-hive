# Tareas

Base: `development` en `e83104ef`; rama propuesta `feat/personal-rules`. Todas las tareas en el hilo principal: cada una edita tres archivos o menos con lecturas acotadas, y T4 escribe en la configuración del usuario.

### T1 — Script `personal_rules.py` con `deploy`, `check` y `remove`

- [x] El script despliega `rules.md` a los hosts con Hive instalado, reemplaza en su lugar, detecta copias que faltan o difieren, quita todo con `remove`, y rechaza una fuente inválida antes de escribir.

**Closes:** AC8, AC9, AC10, AC11, AC13

**Depends on:** none

**Locations:** `content/skills/harness-audit/scripts/personal_rules.py` (nuevo), `tests/skills/personal_rules_test.py` (nuevo).

**Execution:** hilo principal.

**Test approach:** tdd — escribe archivos de instrucciones del usuario. Casos: la línea de prioridad exacta del delta; `CLAUDE_CONFIG_DIR` para Claude Code y `~/.claude/CLAUDE.md` siempre para Grok; un archivo que cambia entre la lectura y la escritura no se escribe, se nombra y la corrida sale con código distinto de 0; `check` reporta una sección fuera de posición y una copia en un host sin Hive, y `deploy` mueve la sección a su lugar; un identificador de prosa como `autoMemoryDirectory` no se rechaza; inserción justo después del final del bloque de Hive y, cuando existe, del de voz; reemplazo en su lugar sin duplicar; bytes fuera de los marcadores iguales (incluidos finales de línea y permisos); frontmatter `alwaysApply: true` en Cursor y `trigger: always_on` en Antigravity; solo los hosts con contenido administrado de Hive; `check` con código distinto de 0 que nombra el host que falta o difiere; `remove` quita secciones y archivos y deja el bloque de Hive; rechazo sin escribir por más de 8,000 bytes, por marcadores propios o de Hive, y por una línea que `line_has_secret` marca; resolución de `CODEX_HOME`, `PI_CODING_AGENT_DIR` y `XDG_CONFIG_HOME`. Cada prueba usa un `HOME` sintético en un directorio temporal.

**Changes:** reutiliza `line_has_secret` de `local_instructions.py`, sin copiarla. No toca el manager ni los archivos de Hive fuera de las secciones y archivos propios.

**Verification:** `python3 -m unittest discover -s tests/skills -p 'personal_rules_test.py'` pasa.

### T2 — Modo «reglas personales» en `harness-audit`

- [x] `harness-audit` enruta un pedido de definir, cambiar, sincronizar, revisar o quitar reglas personales a su referencia nueva, que guía la entrevista, muestra el texto y el diff, escribe `rules.md` solo tras la aprobación del usuario y corre el script.

**Closes:** AC8, AC10

**Depends on:** T1

**Locations:** `content/skills/harness-audit/references/personal-rules.md` (nuevo), `content/skills/harness-audit/SKILL.md` (línea de entrada del modo, su mención en `description`, la excepción del modo en el alcance de escritura de la línea 12 y su exclusión del trabajo desatendido en la línea 79), `content/skills/harness-audit/references/apply.md` (la línea 5 nombra la excepción del modo). Antes de editar, leer `_support/docs/architecture/instruction-resources.md`, como pide `AGENTS.md` para recursos nuevos de una skill.

**Execution:** hilo principal.

**Test approach:** check

**Changes:** la referencia dice quién puede escribir (solo el modo, a pedido del usuario y tras su aprobación), dónde cae en cada host, la línea de prioridad, el límite, `check` y `remove`, que el pedido del usuario al modo es la instrucción explícita que `global.md` exige para un cambio de configuración global, que el modo no corre bajo delegación desatendida, que el contenido de la sección sobrevive a `hive update` (su posición tras reinstalar a mano el bloque de Hive no está garantizada y `check` la reporta) y se queda al desinstalar Hive, que los workspaces de Cursor fuera de la carpeta personal no cargan `personal-rules.mdc`, y que una auditoría de la capa global no la cuente como duplicado (cada host carga solo su copia). Antes de editar, revisar todo lo que gobierna esa conducta como pide `AGENTS.md` «Guidance structure» y listar los lugares revisados en el PR; `content/guidance/global.md` no cambia.

**Verification:** `go test -count=1 ./tests/content/` pasa (tamaño de skill, frontmatter, enlaces); `python3 -m unittest discover -s tests/skills -p 'harness_audit_rules_test.py'` pasa.

### T3 — Documentación del modo

- [x] La página de skills del sitio, `llms.txt` y el README nombran el modo «reglas personales» de `harness-audit` y dónde vive `rules.md`.

**Closes:** AC8

**Depends on:** T2

**Locations:** `site/src/content/docs/skills.md`, `llms.txt`, `README.md` (junto a la mención de `harness-audit` de la línea 495 y al puntero de Cursor de la línea 270).

**Execution:** hilo principal.

**Test approach:** check

**Changes:** describe el modo y enlaza su referencia; no copia sus reglas. Junto al puntero de Cursor del README (línea 270), nombra el límite de los workspaces fuera de la carpeta personal.

**Verification:** `go test -count=1 ./tests/content/` pasa; lectura de cada línea contra el delta.

### T4 — Regla «dónde trabajar» como primer contenido

- [ ] `~/.config/hive/rules.md` tiene la regla «dónde trabajar» aprobada por el usuario, desplegada en los siete hosts, y sigue intacta tras el refresco local del mantenedor.

**Closes:** AC8, AC9, AC10, AC12

**Depends on:** T1, T2, T3, y la integración en `development`

**Locations:** `~/.config/hive/rules.md` y los archivos globales de los hosts (fuera del repositorio).

**Execution:** hilo principal, siguiendo el modo nuevo.

**Test approach:** check

**Changes:** redacta la regla en inglés desde el borrador de `engram:tricell-hive/decision/personal-where-to-work-rule`, ajustado a lo que verificó la prueba en cam (`.git:ro`, `TURBO_CACHE_DIR`, store de pnpm fuera del worktree), y **se la muestra al usuario para que apruebe el texto** antes de escribir `rules.md`. Corre `deploy` y `check`. Copia los cuatro archivos compartidos a la carpeta temporal de la sesión, edita una línea de `rules.md` con la aprobación del usuario, corre `deploy` otra vez y compara cada archivo con su copia. Después del refresco local del mantenedor (`hive update`), corre `check` otra vez.

**Verification:** `grep -c "PERSONAL RULES:BEGIN"` da 1 en los cuatro archivos compartidos; `~/.cursor/rules/personal-rules.mdc` abre con `alwaysApply: true` y `~/.gemini/config/rules/personal-rules.md` con `trigger: always_on`, ambos con la línea de prioridad; tras la edición, el `diff` de cada archivo contra su copia solo toca líneas dentro de los marcadores; `personal_rules.py check` sale con 0 antes y después del refresco local.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar la tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Cubierto por la implementación |
| Revisión de código | Candidato final, antes del push | `/code-review` de Claude Code | La respuesta de revisión |
| Revisión de seguridad | Candidato final, antes del push | hijo `hive-review-security`, con la escritura de archivos de instrucciones, el detector de secretos y la línea de prioridad (D11 B) en su brief | La respuesta de seguridad |

Recorrido humano: abrir una sesión nueva en Codex y otra en Claude Code y ver que la regla «dónde trabajar» está en su contexto.

## Checkpoint

- Updated: 2026-11-21
- Stage: build · 5 local CI
- Last commit: 6b2ee737
- Open findings: none
- Next: flow-verify: reviews (/code-review, hive-review-security), commit change folder, push, PR, merge; then T4
- Rules: flow-build/stage.md
- Principles: none

### Evidence

- before captures: none (no UI)
- go vet ./... · exit 0 · 6b2ee737
- go test -count=1 ./... · exit 0, 22 ok, 264 s · 6b2ee737
- python3 -m unittest discover -s tests/skills -p '*_test.py' · exit 0, 119 s · 6b2ee737
- CI-only: .github/workflows/master-source-branch.yml (source-branch policy on master PRs, no tests)

## Review status and progress

2026-11-21 — T1 · 2ec3e343 · unverified: AC8-AC10 en archivos reales: los observa T4
2026-11-21 — T2 · 0ff323ab · unverified: none
2026-11-21 — T3 · 2db70972 · unverified: none
