# Tareas

Base del cambio: `dc1a749` (`origin/development`, reconciliado al empezar el build el 2026-09-30). Worktree `.claude/worktrees/session-findings`, rama `fix/session-findings-2026-09-30`. Otra sesión entrega en paralelo sobre el checkout principal.

## T1 — Enlaces `skill:` resueltos al instalar

- [x] (`hive-verify-task`, opus: AC8 cumplido en código, con 5 mutaciones que hacen fallar las pruebas. Huecos sin bloqueo: H2, el cambio de `versioning.go:206` no tiene prueba; H3, un comentario engañoso en `skill_links_test.go:126`. AC8 en el home real se comprueba en T4.) Los roles y el bloque global desplegados llevan la ruta del directorio de skills en lugar de `skill:owner/path`: absoluta en ámbito de usuario y relativa a la raíz en ámbito de proyecto. Diseño: «Enlaces `skill:` resueltos al instalar».

**Closes:** AC8.

**Depends on:** ninguna.

**Locations:**
- `integrations/<host>/*.go`: exponer el directorio de skills por `target.Config` y ámbito;
- un paquete nuevo, importable desde `integrations/agents` (por ejemplo `integrations/mdlinks`), con `RewriteSkillLinks` y la detección de bloques de código que hoy está en `outsideFences` (`tooling/management/references.go:23-52`); el validador pasa a reutilizarla;
- `integrations/agents/agents.go`: `Render` recibe el directorio de skills y reescribe `r.Body` antes de codificar (`:376-377`);
- `tooling/management/plan.go`: la llamada a `Render` (`:480`), la reescritura del bloque (`:468`) y la comprobación de igualdad entre hosts (`:483`), extendida al bloque;
- pruebas en `tooling/management/*_test.go` y en `integrations/agents/render_golden_test.go`;
- documentación en `_support/docs/architecture/instruction-resources.md` (sección «Runtime resolution») y en `_support/docs/architecture/deployment-manager.md:133`.

**Execution:** delegada a `hive-build-backend` en el worktree. No comparte archivos con T2 ni con T3; la frase de `global.md:88` queda para T2. El brief debe incluir:
- enlaces a `design.md` y a la spec delta;
- `_support/docs/architecture/deployment-manager.md#verification`;
- `AGENTS.md` del repo;
- la regla de no tocar `content/`;
- no subir `agents.Version`;
- no crear un ciclo de importación entre `management` y `agents`.

**Test approach:** tdd. Las pruebas se escriben primero y fallan en `dc1a749`, todas con contenido sintético, como hace `render_golden_test.go`:
- **Usuario, por host:** un rol sintético con `[x](skill:flow-build/references/browser-automation.md)` sale como `<home>/.agents/skills/flow-build/references/browser-automation.md` entre `<...>`, sin `](skill:`, en cada host. En Codex se comprueba dentro de la cadena TOML decodificada.
- **Proyecto:** en Claude sale `.claude/skills/...` y en Codex `.agents/skills/...`, ambos relativos, sin el `home`.
- **Exclusiones:** un `skill:` dentro de un bloque de código y un enlace de referencia `[x]: skill:...` quedan como estaban.
- **Bloque global:** el sintético no contiene `](skill:` y es igual para Claude y Grok cuando comparten archivo. Si se fuerzan directorios distintos, el plan falla antes de escribir.
- **Idempotencia:** dos planes seguidos sobre la misma revisión; el segundo no propone escrituras.
- **Versiones anteriores:** `plan install --release <anterior>` sigue validando, porque `Version` no cambia.

**Verification:** `go vet ./... && go test ./...` en verde (unos 50 s), con las pruebas nuevas fallando en `dc1a749`. `/code-review` (D5-A) sobre el diff Go antes del merge.

## T2 — Reglas en la guía global

- [x] (`hive-verify-task`, sonnet: AC1, AC5, AC6, AC7, AC9, AC10 y AC11 cumplidos. `global.md` queda en 44511 bytes; el presupuesto sube a 45535, el tamaño final más 1024.) `content/guidance/global.md` incorpora H1, H4, H5, H6 (relevo), H7, H8, H9, y ajusta la frase de `:88` sobre resolver `skill:` según D3-A, sin superar el presupuesto. Diseño: «Texto de la guía».

**Closes:** AC1, AC5, AC6, AC7, AC9, AC10, AC11.

**Depends on:** ninguna. La frase de `:88` usa el contrato de T1.

**Locations:** `content/guidance/global.md`, líneas 8, 16, 22, 32, 88, 90 y 108; y `tests/content/budget_test.go:12`, solo si hace falta.

**Execution:** hilo principal. Son ediciones de redacción cuyo brief sería más largo que el cambio.

**Test approach:** check. Por cada criterio, su búsqueda de texto; más la prueba de presupuesto.

**Verification:**
- Cada búsqueda da al menos 1 en el worktree:
  - `grep -c "reasoning is not shown"` (AC1)
  - `grep -c "verified in this session"` (AC5)
  - `grep -c "cite the ID without a link"` (AC6)
  - `grep -c "every defect a child reports"` (AC7)
  - `grep -c "recommendation for that question"` (AC9)
  - `grep -c "option labels"` (AC10)
  - `grep -c "asks the user to name"` (AC11)
- `go test ./tests/content/` en verde.
- `wc -c content/guidance/global.md` queda en 44336 o menos, o el presupuesto sube al tamaño final más ~1 KiB, con la razón en el PR y un enlace a #47.

## T3 — Recorrido, comprobaciones fallidas y umbral de delegación

- [x] (`hive-verify-task`, sonnet: AC2, AC3 y AC4 cumplidos.) `flow-build/SKILL.md` asigna el recorrido funcional a hijos y trata la comprobación fallida (H2). `flow-research/SKILL.md` fija el umbral, y `flow-plan/SKILL.md` remite a él (H3).

**Closes:** AC2, AC3, AC4.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-build/SKILL.md:63`, `content/skills/flow-research/SKILL.md:18` y `content/skills/flow-plan/SKILL.md` («Ground the decisions»).

**Execution:** hilo principal. Son tres frases.

**Test approach:** check.

**Verification:**
- Cada búsqueda da al menos 1:
  - `grep -c "does not script" content/skills/flow-build/SKILL.md` (AC2)
  - `grep -c "never as healthy" content/skills/flow-build/SKILL.md` (AC3)
  - `grep -c "sixth search" content/skills/flow-research/SKILL.md` (AC4)
  - `grep -c "flow-research" content/skills/flow-plan/SKILL.md` (AC4)
- `go test ./tests/content/` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` en verde, para comprobar enlaces y formato de las skills.

## T4 — Entrega y despliegue

- [x] (PR #86 integrado como `3988d0b`; binario recompilado; `hive update` aplicado en los 6 hosts. Ningún rol ni bloque desplegado conserva `](skill:`: Claude, Grok, Codex (TOML), Cursor, OpenCode, Pi y los bloques de `~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md` y `~/.config/opencode/AGENTS.md`. Los bloques enlazan `/path/to/home/.agents/skills/flow-research/references/backlog-report.md`. El segundo `hive update --dry-run` da «Hive files checked: 188; none change».) PR a `development` con T1–T3, merge, binario recompilado, `hive update` aplicado y AC8 comprobado en lo desplegado.

**Closes:** AC8. Es la comprobación desplegada; T1 la cubre en pruebas.

**Depends on:** T1, T2 y T3.

**Execution:** hilo principal, que es el dueño de los efectos de entrega.

**Changes:**
1. Ejecutar la suite completa una vez sobre el candidato final.
2. Ejecutar `/code-review` sobre el diff Go y atender sus hallazgos.
3. Hacer push y abrir el PR.
4. Hacer merge.
5. Actualizar el checkout principal.
6. Recompilar con `go build -o "$(command -v hive)" ./tooling/cli`.
7. Ejecutar `hive update`, con `--dry-run` primero.

**Verification:**
- `grep -l "](skill:"` no devuelve nada sobre `~/.claude/agents/*.md`, `~/.grok/agents/*.md`, `~/.codex/agents/*.toml`, los directorios de roles de Cursor, OpenCode y Pi que muestre `hive status`, y los archivos de bloque de cada host registrado (`~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md`, `~/.config/opencode/AGENTS.md` y los demás que liste `hive status`).
- El bloque de `~/.claude/CLAUDE.md` contiene `<home>/.agents/skills/flow-research/references/backlog-report.md`.
- Un segundo `hive update --dry-run` no propone escrituras.

## Verificación compartida

| Gate | Momento | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Pruebas por tarea | Durante T1–T3 | `hive-verify-task` por cada tarea `tdd`/`check`, con un modelo distinto al del implementador | Parte de la implementación |
| Suite completa | Candidato final | `go vet ./... && go test ./...` | Parte de la implementación |
| Revisión de código | Antes del merge | `/code-review` de Claude Code sobre el diff Go (D5-A) | Acordada (D5-A) |
| Despliegue local | Tras el merge | AC8 sobre todos los hosts registrados y un segundo `--dry-run` sin escrituras | Acordada (D4-A) |
| Efecto en sesiones reales | Sesiones posteriores de ark, sample-project y globex | Vigilancia en la memoria `project_next_session_openspec_monitoring.md` | Sin piloto: los pilotos siguen pausados |

No aplican recorrido en vivo ni revisión de UI: el cambio no tiene superficie que se pueda ejecutar o renderizar.

## Revisión del plan y progreso

**Ronda 1** sobre la revisión `96744d565910`, con dos `hive-review-plan` nativos en paralelo (Claude Code; modelo según el perfil del rol, no verificado).

- **Guía distribuida (T2, T3).**
  - Se aceptan B1–B4:
    - la escala de severidad, que cubre todas las escalas de los hijos;
    - la URL tomada del tracker, sin chocar con el recorte del slug;
    - el presupuesto, que sube al tamaño final más ~1 KiB;
    - H2 como frase condicional que remite a `verification.md:97`.
  - Se aceptan N1, N2, N3 (redacción ampliada a cualquier comprobación fallida, que sigue en `flow-build/SKILL.md` porque es lo que está cargado) y N4.
  - N5 se acepta: el "0 commits" de OpenCode queda cubierto por Evidence, no por H4.
  - N6 se acepta como límite: las búsquedas de texto comprueban una parte de cada criterio, y `hive-verify-task` lee el texto completo.
  - Todo se aplicó como lo propuso el revisor, así que el orquestador lo concilió sin otra ronda.
- **Instalador (T1, T4).**
  - B1, la ruta en ámbito de proyecto, llevó a la decisión del usuario D6-A (ruta relativa a la raíz).
  - Se aceptan B2 (todos los hosts y Codex en TOML), I1 (no subir `agents.Version`), I2 (bloque por host con comprobación de igualdad), I3 (interfaz fijada) y S1–S4.
  - Esto cambiaba el contrato, así que hubo una re-revisión.

**Re-revisión del instalador:** con el mismo revisor, reanudado sobre el plan corregido, sin hash nuevo.
- Quedan resueltos B2, I1, I2, S1, S2 y S4.
- **R1, bloqueante:** la reescritura de Codex no podía ocurrir después de `agents.Render`, y la función no podía vivir en `management` sin crear un ciclo de importación. Se aplicó tal como la propuso el revisor: `Render` recibe el directorio de skills y reescribe antes de codificar, y la detección de bloques de código pasa a un paquete compartido. El orquestador la concilió sin otra ronda.
- **N1:** la ruta relativa a la raíz se documenta como limitación.
- **N2:** el escape `\u003c` en TOML queda anotado en el diseño.

**Límites que quedan:** las búsquedas de texto solo comprueban una parte de cada criterio de la guía, y `hive-verify-task` lee el texto completo. El ciclo de importación lo dedujo el revisor de las importaciones de `plan.go`, sin revisar el grafo completo, así que el implementador lo confirma.

**Estado:** cerrado.

**Suite final:** `go vet ./...`, `go test -count=1 ./...` y las 23 pruebas de skills, en verde sobre el candidato final.

**`/code-review` (D5-A), nivel medium, sobre el diff Go:** dos hallazgos de severidad baja y ningún error de corrección. Ambos se descartan con su razón:
- **`agents.Version` no sube.** La consecuencia es que un plan guardado con el binario anterior falla al aplicarse con «invalid ownership or release payload». No se sube a propósito, por I1: subirlo rompe `--release` de las versiones anteriores. La entrega de T4 recompila antes de planificar, así que no lo encuentra.
- **En ámbito de proyecto, la ruta es relativa a la raíz.** Es la limitación aceptada con D6-A y ya está documentada en `instruction-resources.md` y en `global.md:89`.

Ronda de corrección, delegada al mismo `hive-build-backend`: una prueba para `versioning.go:206` (H2) y el comentario de `skill_links_test.go:126` (H3).

**Hallazgo incidental, corregido dentro del cambio:** lo señaló el verificador de T3. La fila «Postdeploy» de `flow-build/references/verification.md:34` decía «Offer the functional in-vivo/UI gate…», y eso contradecía `:97` («you do not offer it»). La contradicción ya estaba en `dc1a749`. La fila ahora remite a las comprobaciones de despliegue y no ofrece el recorrido. Es una frase y la cubren las pruebas de contenido.

**Modelos de verificación por tarea:** el implementador de T2 y T3 es el hilo principal, `claude-opus-5-5`, observado. El verificador es `hive-verify-task`, que por perfil usa `opus` (el mismo modelo), así que se lanza con `sonnet` por la opción de modelo del lanzamiento; está configurado, no observado. T1 lo implementa `hive-build-backend`, con `sonnet` configurado y no observado, y lo verifica `hive-verify-task`, con `opus` configurado y no observado: son modelos distintos.

**Siguiente paso:** `flow-build` con esta carpeta.
