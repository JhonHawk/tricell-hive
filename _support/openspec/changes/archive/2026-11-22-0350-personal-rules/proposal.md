# personal-rules — reglas personales del usuario en todos los hosts, con un modo de harness-audit

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #577, 19f3b7be |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `e83104ef` · rama propuesta `feat/personal-rules` |
| Playbook | feature |

## Decisiones

2026-11-21
- **D7 A** — ¿Se escribe la regla «dónde trabajar» (mismo checkout, worktree o `sbx`) y se prueba? Sí; la prueba en cam (worktree temporal montado en `sbx`) confirmó que funciona con el `.git` del repo montado en solo lectura, `TURBO_CACHE_DIR` dentro del worktree y el store de pnpm fuera de él. El usuario pidió que sea compatible con todos los hosts, en el equivalente de la guía global, no en `~/.claude/rules`.
- **D8** — ¿Solo Claude Code o todos los hosts? Todos los hosts.
- **D9 A** — ¿Con qué enfoque llegan las reglas personales a todos los hosts? Un bloque del manager con el patrón del bloque de voz, desplegado por `hive update` desde un archivo del usuario; descartados la lectura al iniciar la sesión y solo documentarlo.
- **D10 A** — ¿Dónde va el bloque y qué dice sobre su prioridad? Al final, después del bloque de Hive y del de voz, con una línea inicial: son reglas propias del usuario y, donde contradigan la guía de Hive, mandan estas.
- **D11 B** — ¿Las reglas personales mandan también sobre las reglas de Secretos y de efectos críticos de Hive? Sí, mandan sobre toda la guía de Hive; la línea de prioridad lo dice.
- **D12 A** — ¿Qué manda cuando una regla personal choca con las instrucciones de un proyecto? Manda el proyecto, para ese proyecto.
- **D13 A** — ¿Cómo se evita que un agente cambie las reglas personales sin que el usuario se entere? El manager solo aplica un cambio del texto personal tras confirmarlo en una terminal interactiva con el diff a la vista.
- **Supuesto** — Confirmación solo interactiva, sin flag. **Retired:** pertenecía al enfoque del manager (D9 A), reemplazado por D14 A.
- **Supuesto** — Confirmación comparada con el último texto confirmado guardado en el estado. **Retired:** enfoque del manager, reemplazado por D14 A.
- **Supuesto** — Confirmación documentada como fricción, no candado. **Retired:** enfoque del manager, reemplazado por D14 A.
- **Supuesto** — `stateVersion` de 7 a 8. **Retired:** el nuevo enfoque no toca el estado del manager.
- **Supuesto** — Sin comando nuevo de `hive`. **Retired:** el nuevo enfoque no toca el manager.
- **Supuesto** — Archivos `hive-personal.*` en Cursor y Antigravity. **Retired:** el nuevo enfoque usa `personal-rules.*`, porque los archivos son del usuario y no del manager.
- **D14 A** — ¿Cambiamos a una skill que define la regla y la escribe en los archivos globales de cada host? Sí: un modo «reglas personales» de `harness-audit` con un script que la despliega a los siete hosts. **Replaces:** D9 A, D13 A
- **H1 A** — ¿Se ata el deploy al texto que el usuario aprobó? No: se queda la regla del modo, y una regla personal que choque con Secretos o autorización se reporta como hallazgo en las auditorías.
- **H2 A** — ¿Cómo se resuelve la contradicción con «Secrets» de `global.md`, que vale «in every mode»? La línea de prioridad excluye las reglas de Secretos de Hive, que siguen valiendo; autorización sigue bajo las reglas personales. **Replaces:** D11 B (en parte)
- **Supuesto** — Fuente única en `rules.md` de `$XDG_CONFIG_HOME/hive/` o `~/.config/hive/`; el script siempre despliega y compara desde ahí. Alternativa descartada: una fuente por host. Para revertirlo, se cambia la fuente.
- **Supuesto** — Solo a nivel de usuario. Alternativa descartada: también a nivel de proyecto. Para revertirlo, se amplía el alcance.
- **Supuesto** — Límite de 8,000 bytes y rechazo de líneas que `line_has_secret` de `harness-audit/scripts/local_instructions.py` marca (no el `secret_like` desnudo, que rechazaría identificadores de prosa). Alternativa descartada: sin límite. Para revertirlo, se cambia la constante.
- **Supuesto** — HA-HI-08 y HA-HI-11 (auditoría de memoria) no cambian en esta obra: no mandan todavía una preferencia de toda la máquina al modo nuevo. Alternativa descartada: enrutarlas ahora. Para revertirlo, se planea aparte.
- **Supuesto** — El modo no corre bajo delegación desatendida. Alternativa descartada: permitirlo con aprobación previa. Para revertirlo, se cambia la referencia.
- **Supuesto** — Hosts destino: los que tienen Hive instalado a nivel de usuario, detectados por el contenido administrado de Hive en cada destino. Alternativa descartada: los siete siempre. Para revertirlo, se agrega una opción de hosts.
- **Supuesto** — Al desinstalar Hive, las secciones personales se quedan, porque son del usuario; `personal_rules.py remove` las quita. Alternativa descartada: quitarlas con Hive. Para revertirlo, se agrega al desinstalador.
- **Supuesto** — Solo el modo escribe o sincroniza reglas personales, a pedido del usuario y tras su aprobación del texto en la conversación. Alternativa descartada: una confirmación en terminal del manager. Para revertirlo, se agrega una confirmación al script.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «para no obviar, lo que me interesa esa que yo siempre abra la sesión desde mi repositorio en esta mac, y que duran la sesión se razone si, se trabaja sobre el mismo o worktree o sbx […] eso va para mis reglas mías»
> Nota a D7: «ojo, para que sea compatible con todos los hosts, sabes que debe ir en lo equivalente a global.md, talvez necesitemos una skills que añada estas reglas globales»
> Respuesta al plan del manager: «y si hacemo algo más sencillo? Una skill que te ayude a definir la regla y despues solo la escriba a mano en los .md globales de todos los hosts que aplica»

Sin ticket.

## Objective

Hoy Hive solo administra su propio bloque en el archivo global de cada host; una regla propia del usuario hay que copiarla a mano en cinco archivos y en dos carpetas de reglas, y nada la mantiene igual. Con el cambio, el usuario pide a `harness-audit` definir o cambiar una regla, aprueba el texto, y un script lo escribe en los siete hosts como una sección o un archivo que cada host carga siempre, después del bloque de Hive y con prioridad declarada.

## Scope and acceptance

Dentro: el modo «reglas personales» de `harness-audit`, su script con `deploy`, `check` y `remove`, sus pruebas y su documentación, y la regla «dónde trabajar» como primer contenido en esta Mac. Fuera: cambios en el manager (`tooling/`, `integrations/`), en `content/guidance/global.md` o en otras reglas de Hive; el nivel de proyecto.

- AC1 — Bloque personal del manager tras `hive update`. **Retired:** enfoque del manager, reemplazado por D14 A.
- AC2 — Archivos `hive-personal.*` creados por `hive update`. **Retired:** enfoque del manager, reemplazado por D14 A.
- AC3 — `hive update` aplica una edición solo de `rules.md`. **Retired:** enfoque del manager, reemplazado por D14 A.
- AC4 — El plan del manager rechaza un `rules.md` inválido. **Retired:** enfoque del manager; el rechazo pasa a AC11.
- AC5 — Conflicto, reinstalación y recuperación del bloque del manager. **Retired:** enfoque del manager, reemplazado por D14 A.
- AC6 — Fila personal en `hive status`. **Retired:** enfoque del manager, reemplazado por D14 A.
- AC7 — Confirmación en terminal de un texto personal nuevo. **Retired:** enfoque del manager, reemplazado por D14 A.
- AC8 — Con `rules.md` aprobado, `personal_rules.py deploy` deja en `~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md`, `~/.pi/agent/AGENTS.md` y `~/.config/opencode/AGENTS.md` una sección entre los marcadores `PERSONAL RULES` justo después del bloque de Hive (o del de voz), que abre con la línea de prioridad del delta y contiene el texto · false today because el script no existe y `grep -c "PERSONAL RULES:BEGIN"` da 0 en esos archivos · Observed in: los archivos globales reales de esta Mac en T3, y pruebas con `HOME` sintético
- AC9 — El mismo deploy crea `~/.cursor/rules/personal-rules.mdc`, que abre con `alwaysApply: true`, y `~/.gemini/config/rules/personal-rules.md`, que abre con `trigger: always_on`, ambos con la línea de prioridad y el texto · false today because ninguno de los dos archivos existe · Observed in: `~/.cursor/rules/` y `~/.gemini/config/rules/` de esta Mac en T3
- AC10 — Correr `deploy` otra vez tras editar `rules.md` reemplaza cada copia en su lugar, sin duplicarla, y deja iguales los bytes fuera de los marcadores; `check` sale con código distinto de 0 y nombra cada host cuya copia falta, difiere de `rules.md` o no está justo después del bloque de Hive (o de voz), y cada copia en un host que ya no tiene Hive; un archivo que cambió entre la lectura y la escritura no se escribe, se nombra y `deploy` sale con código distinto de 0 · false today because no hay script · Observed in: pruebas con `HOME` sintético y una edición real de `rules.md` en esta Mac en T3
- AC11 — Un `rules.md` de más de 8,000 bytes, con los marcadores propios o los del bloque de Hive, o con una línea que `line_has_secret` marca, hace que `deploy` salga con código distinto de 0 nombrando `rules.md` y la causa, sin cambiar ningún archivo · false today because no hay script · Observed in: pruebas con `HOME` sintético
- AC12 — Tras un `hive update` real posterior al deploy, cada sección personal sigue presente e igual · false today because no hay secciones · Observed in: los archivos globales reales de esta Mac en T3, con el refresco local del mantenedor
- AC13 — `personal_rules.py remove` quita las cuatro secciones y los dos archivos, y deja intacto el bloque de Hive · false today because no hay script · Observed in: pruebas con `HOME` sintético

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `harness-audit` ya conoce los archivos de instrucciones de cada host (`content/skills/harness-audit/references/instruction-files.md`, `hierarchy.md`) y ya tiene un modo para un archivo de la máquina (`references/local-instructions.md`); su `SKILL.md` pesa 9,794 bytes (tope 16,000, `tests/content/skill_size_test.go`).
- Revisión de secretos en Python: `line_has_secret` de `content/skills/harness-audit/scripts/local_instructions.py:81-95`, que envuelve `secret_like` de `extract_candidates.py`.
- Alcance de escritura de `harness-audit`: `SKILL.md:12` prohíbe editar configuración fuera del proyecto salvo el archivo de la máquina; `references/apply.md:5` edita solo el proyecto actual; `SKILL.md:79` lista lo que el trabajo desatendido deja al usuario.
- Claude Code resuelve `CLAUDE_CONFIG_DIR` (`tooling/management/types.go:444`); Grok usa siempre `~/.claude/CLAUDE.md` (`integrations/grok/grok.go:37`).
- El manager preserva los bytes fuera de su bloque y reemplaza el bloque en su lugar: `_support/docs/architecture/repository-and-distribution.md` «Installation and update contract» pasos 2–3; `tooling/management/files.go` `transform` (leído por un hijo `hive-research`, 2026-11-21, `e83104ef`). La voz se inserta justo después del final del bloque de Hive (`voice.go:313`).
- Destinos por host: `integrations/claude/claude.go:22-26`, `codex/codex.go:25-35`, `grok/grok.go:39` (comparte `~/.claude/CLAUDE.md`), `pi/pi.go:26`, `opencode/opencode.go:29`, `cursor/cursor.go:36` (`~/.cursor/rules/hive.mdc`, `alwaysApply`), `antigravity/antigravity.go:27-29,~44` (`hive-NN.md`, `trigger: always_on`).
- Hosts con Hive instalado en esta Mac: los siete (`hive status --scope user --hosts codex,claude,grok,pi,opencode,cursor,antigravity`, 2026-11-21).
- Grok da prioridad al texto posterior · fuente: `repository-and-distribution.md:92` · not measured contra Grok; no cambia ninguna tarea · 2026-11-21
- `python3 -m unittest discover -s tests/skills -p 'personal_rules_test.py'` · not measured: la prueba la agrega T1; forma más cercana: `python3 -m unittest discover -s tests/skills -p 'local_instructions_test.py'` exit 0 en 1.6 s, y una prueba desechable que solo importa `extract_candidates.py` exit 0 · 2026-11-21 en `e83104ef`
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` · exit 0, 101 s · 2026-11-21 en `e83104ef`
- `go test -count=1 ./tests/content/` · exit 0, 2.3 s · 2026-11-21 en `e83104ef`
- `python3 -m unittest discover -s tests/skills -p 'harness_audit_rules_test.py'` · exit 0, 2 pruebas · 2026-11-21 en `e83104ef`
- `go vet ./...`, `go test ./...` · exit 0 (258 s) · 2026-11-21 en `e83104ef`
- `personal_rules.py check` · not measured: T1 agrega el script; T4 lo corre primero en esta Mac · 2026-11-21
- `grep -c "PERSONAL RULES:BEGIN"` en los cuatro archivos compartidos · da 0 hoy; `~/.cursor/rules/personal-rules.mdc` y `~/.gemini/config/rules/personal-rules.md` no existen · 2026-11-21

## Effects

None: the tasks take no effect on an external system.

## Delivery

Modo `automatic` (respuesta de la sesión, 2026-11-21): revisión, push, PR e integración en `development` sin detenerse para validar. Revisión de código: nativa, `/code-review` de Claude Code. Revisión de seguridad: sí, un hijo `hive-review-security`. El refresco de la instalación local tras la integración lo autoriza el marcador de mantenedor (`AGENTS.md` «Refreshing a local installation»); escribir `rules.md` y desplegarlo en los archivos globales lo autoriza el pedido del usuario (D7 A, D14 A), tras su aprobación del texto en T3. No autoriza un release ni la promoción a `master`.

## Review status and progress

2026-11-21 — plan escrito con el enfoque del manager (D9 A) y revisado en cuatro rondas por un hijo `hive-review-plan`; la última, sin hallazgos.
2026-11-21 — revisión: D14 A reemplaza D9 A y D13 A por un modo de `harness-audit` con script. Retirados AC1–AC7 y los supuestos del manager; agregados AC8–AC13; carpeta renombrada de `personal-rules-block` a `personal-rules`; delta movido de `versioned-installation` a una capacidad nueva `personal-rules`. Invalida T1–T3 anteriores (reescritos).
2026-11-21 — revisión del plan nuevo (mismo hijo `hive-review-plan`, dominio guía; ~262k tokens, 111 s): 2 P2, 7 P3, sin P0/P1. Aplicado: excepción del modo en el alcance de escritura de `harness-audit` (`SKILL.md:12`, `apply.md:5`) y su exclusión del trabajo desatendido; HA-HI-08/11 sin cambios como supuesto; `line_has_secret` en vez de `secret_like`; `CLAUDE_CONFIG_DIR`; relectura antes de escribir; copias fuera de los hosts y posición en `check`; `instruction-resources.md` y el límite del puntero de Cursor en T2/T3; línea de `personal_rules.py check` en el contexto verificado; el delta ya no promete la posición tras una reinstalación a mano.
2026-11-21 — un P3 más (código de salida cuando `deploy` se salta un archivo cambiado) aplicado; relectura final sin hallazgos.
2026-11-21 — verify: `/code-review` (9 hallazgos, 2 importantes) y `hive-review-security` (2 medios, 4 bajos); ronda de correcciones en `bc06dcf1`; H1 A y H2 A del usuario.
