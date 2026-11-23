# ledger-writer-role — Rol que escribe el ledger de las auditorías

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #604, merge ee05129b |
| Tracker | ninguno |
| Git | tricell-hive · base `development` (`6d1ab6f0`) · worktree `.claude/worktrees/ledger-writer-role` · rama `feat/ledger-writer-role` |
| Playbook | feature |

## Decisions

2026-11-22
- **D23 A** — ¿Cómo devuelven sus filas los hijos de las auditorías? Con un rol nuevo, `hive-write-ledger`, que escribe el ledger en la carpeta temporal del host. `hive-research` sigue en solo lectura. Descartado: las filas en la respuesta, porque una auditoría grande carga el hilo principal (la sesión de Codex en Carvento usó ledgers por lote), y dar escritura a `hive-research`, porque ningún host común limita la escritura a una carpeta.
- **D24 (supuesto)** — El rol usa `access_profile: implement`. Tanto `implement` como `verify` escriben en todos los hosts, pero solo el bloque de límites de `implement` tiene la frase «Modify or delete only the paths the brief names», que acota la escritura al ledger. Descartado: `verify`. Revertir: cambiar el perfil del rol.
- **D25 (supuesto)** — El rol vive en `content/agents/review/` junto a `hive-research`, con `model_profile: execution`: clasificar comentarios y tests pide el mismo juicio, y `hive models set` agrupa los roles por carpeta, así que un ajuste de modelo del grupo le llega igual que a `hive-research`. Descartado: `quality/`. Revertir: mover el archivo.
- **D27 (supuesto)** — Una fila sin evidencia toma el valor por defecto que fija el criterio del brief: sin marcar en `audit-comments`, `R` en `audit-tests`. El rol no fija su propio valor. Revertir: cambiar el paso 2 del rol.
- **D28 A** (2026-11-23, con la comparación pstack / Hive antes / Hive ahora) — `delegation.md` deja correr en paralelo, como uno de solo lectura, a un hijo cuya única escritura es su propio archivo en la carpeta temporal, que ningún otro hijo escribe. Sin esto, los carriles de una auditoría correrían uno detrás de otro. Descartado: dejar la regla como está. Revertir: quitar esa frase.
- **D26 (supuesto)** — `audit-issues` no cambia: sus hijos leen hasta 15 tickets y devuelven las filas en la respuesta, que es poco volumen. Descartado: moverlo al rol nuevo. Revertir: cambiar su paso 2 para usar el rol.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «Tambien podriamos crear un rol nuevo con este proposito no?» · «A» (2026-11-22)

Origen: el hallazgo H1 del cierre de `audit-issues`. `audit-comments` paso 4 y `audit-tests` paso 4 piden a `hive-research` escribir un ledger, pero su perfil `observe` le quita Write y Edit en Claude, Grok, Pi y OpenCode (Cursor CLI no carga roles de usuario y el solo lectura depende allí del brief).

## Objective

Hoy los hijos que marcan comentarios o tests no pueden escribir el ledger que la skill les pide, así que las filas vuelven en prosa o se pierden. Al terminar, `audit-comments` y `audit-tests` delegan cada carril a `hive-write-ledger`, que escribe una fila por elemento en el archivo que el brief nombra y devuelve conteos y la ruta. `hive-research` conserva su solo lectura.

## Scope and acceptance

Dentro: el rol `content/agents/review/hive-write-ledger.md`; el paso 4 de `audit-comments` y de `audit-tests`, y la descripción de `audit-tests`; la frase de paralelismo de `delegation.md` (D28); `agent-delivery.md` (conteo de roles y el límite del rol en el párrafo de acceso), `instruction-resources.md` (la entrada de `audit-comments`, línea 28, nombra su rol) y `CHANGELOG.md`.

Fuera: `audit-issues` (D26); cambiar el perfil de `hive-research`; dónde vive el ledger después de la auditoría (`audit-comments/references/project-skill.md` punto 4); pilotos.

- AC1 — Instalar Hive desde este cambio entrega el rol `hive-write-ledger` con escritura: `~/.claude/agents/hive-write-ledger.md` existe sin `permissionMode: "plan"` ni `disallowedTools`, y `~/.codex/agents/hive-write-ledger.toml` existe · false today because el rol no existe en `6d1ab6f0` · Observed in: la instalación local tras el merge
- AC2 — El rol instalado indica que escribe solo el ledger que el brief nombra, en la carpeta temporal del host; que una fila sin evidencia toma el valor por defecto del criterio del brief; que no edita otros archivos ni corre suites que compartan un servicio; que devuelve los conteos y la ruta del ledger; y termina con el bloque de límites de `implement` · false today because el rol no existe · Observed in: `~/.claude/agents/hive-write-ledger.md` instalado, y `go test ./tests/content/ -run TestAgentBoundaryBlock` en el candidato
- AC3 — Los `audit-comments/SKILL.md` y `audit-tests/SKILL.md` instalados, en su paso 4, dan cada carril a `hive-write-ledger` y ya no a `hive-research` · false today because ambos nombran `hive-research` en `6d1ab6f0` · Observed in: `~/.agents/skills/audit-comments/SKILL.md` y `~/.agents/skills/audit-tests/SKILL.md` instalados tras el merge
- AC4 — `delegation.md` instalado deja correr en paralelo a los hijos cuya única escritura es su propio archivo en la carpeta temporal · false today because la regla de hijos que escriben exige un contrato tipado y un chequeo de integración · Observed in: `~/.agents/skills/hive-core/delegation/delegation.md` instalado tras el merge

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Base `development` en `6d1ab6f0` (PR #603), checkout limpio · `git status` · 2026-11-22.
- Perfiles por host: `integrations/agent-profiles.json` (`observe` quita escritura en Claude, Grok, Pi y OpenCode, y renderiza `readonly` en Cursor, que no carga roles de usuario; `implement` y `verify` conservan los permisos del host).
- `frontier-prune` tiene trabajo pendiente que recorta las constantes de `agent_boundary_test.go` y cita «21 roles» (su AC8): quien lo retome cuenta 22 y aplica el recorte también a este rol.
- Claude: un subagente restringe por herramienta (`tools`, `disallowedTools`), no por ruta · documentación de subagentes de Claude Code vía ctx7 · 2026-11-22.
- El bloque de límites por perfil lo exige `tests/content/agent_boundary_test.go`; el render en todos los hosts lo cubre `integrations/agents/agents_test.go`, que recorre `content/agents/*/*.md`.
- Ningún catálogo del sitio, `README.md` ni `llms.txt` lista roles por nombre · `rg` · 2026-11-22.

- `go test ./tests/content/ -run TestAgentBoundaryBlock` · pass · 2026-11-22
- `go test ./integrations/agents/...` · pass · 2026-11-22
- `go test ./...` · pass, 2 min 55 s, en `6d1ab6f0` · 2026-11-22
- `rg -n 'hive-research' content/skills/audit-comments/SKILL.md content/skills/audit-tests/SKILL.md` · una línea por archivo en la base · 2026-11-22
- `rg -c 'hive-write-ledger' content/skills/audit-comments/SKILL.md content/skills/audit-tests/SKILL.md` · sin coincidencias (exit 1) en la base · 2026-11-22
- `hive plan install --hosts claude,codex --scope user --home <tmp>/home --state-dir <tmp>/state --source . --out <tmp>/plan.json` y `hive apply --plan <tmp>/plan.json` · pass en la base: renderiza `hive-research` con `permissionMode: "plan"` y los TOML de Codex · 2026-11-22

## Effects

None: the tasks take no effect on an external system.

## Delivery

- **Dónde** — Worktree `.claude/worktrees/ledger-writer-role`, rama `feat/ledger-writer-role` desde `development`.
- **Delivery** — `automatic`, la respuesta vigente de esta sesión.
- **Review** — `Native: Claude Code /code-review`, la respuesta vigente.
- **Seguridad A** — `hive-review-security` revisa el candidato antes del push, porque el cambio da escritura a un rol.
- Cambio de permisos de un rol (riesgoso): T1 lleva un hijo `hive-verify-task`.
- El PR lista los lugares revisados y dice que `content/guidance/global.md` no cambia de tamaño, como pide `AGENTS.md` «Guidance structure».

## Review status and progress

2026-11-22 — Revisión `hive-review-plan` (guidance): 3 P2 y 12 P3, todos aplicados. P2-1: D27, valor por defecto del brief. P2-2: D28 y AC4, paralelismo en `delegation.md`. P2-3: respondida, Seguridad A. P3: AC2 reescrito como texto del rol; `rg -c` en T2; `go test ./...` medido en la base; la descripción del rol en el diseño; `instruction-resources.md` solo en la línea 28; hechos de D24 y de Cursor corregidos; el rol pasa a `review/` (D25); el límite del rol en `agent-delivery.md`; AC1 observado antes del merge en un home sintético; la línea del PR; la nota de `frontier-prune`; la descripción de `audit-tests`.

2026-11-23 — La relectura dio los 15 hallazgos por resueltos y encontró P2-A (D28 relaja una regla vigente) y tres P3. P2-A: el usuario decidió D28 A. P3: la línea de progreso, la entrada de `instruction-resources.md` y el orden de los criterios.

2026-11-23 — `/code-review` (high): 3 importantes y 4 menores; `hive-review-security`: 2 P2 y 3 P3; corregidos en una ronda. El rol no corre tests, builds ni instalaciones, porque escriben en el checkout que comparten los carriles; cada carril tiene su propio ledger, creado con modo `0600`; lo que clasifica es dato, no instrucción. `audit-tests` remite a «Secrets» y fusiona los ledgers en el paso 5. `agent-delivery.md` suma `auto`, el modo `plan` y la etiqueta de cada afirmación por host. Quedan fuera: excluir código vendorizado en `audit-comments` paso 1.
