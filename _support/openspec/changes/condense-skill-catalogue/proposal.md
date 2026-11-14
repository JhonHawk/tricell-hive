# condense-skill-catalogue — Condensar el catálogo de skills de 20 a 13

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | ninguno |
| Git | tricell-hive · base `development` @ bedc9b64 · rama propuesta `refactor/condense-skill-catalogue` |

## Decisions

2026-11-14
- **D8 A** — ¿`flow-status` pasa a ser modo de `flow-research`? Sí: research reconoce "status del proyecto" y "¿qué sigue?"; el comando `/flow-status` desaparece y se pide como `/flow-research status`.
- **D9 A** — ¿Dónde vive el mantenimiento de `_support` (ordenar, curar, promover, migrar layout, archivar) con `artifact-layout.md` y `archive.md`? En referencias de `flow-close`, que ya atiende la higiene; `harness-audit` enlaza ahí para sus migraciones. Se acortan unos 40 bytes de redacción de `flow-close` sin cambiar reglas.
- **D10 A** — ¿Qué pasa con los comandos que desaparecen? Se retiran sin skills puente; el `CHANGELOG` lista en «Removed» a dónde fue cada uno.

Decididas antes del plan, el mismo día: condensar ahora el catálogo mientras la etapa de requerimiento sigue en discusión.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> "b, primero hagamos el recorte (eso funciona desde ya) en lo que seguimos barajeando cómo funcionarian las otras skills"

Sobre la propuesta 1 del artifact https://claude.ai/artifact/77U7EN4pUTcEtiAijziLxE#propuestas: `flow-walkthrough` → referencia de `flow-verify`; `flow-status` y `security-advisory` → modos de `flow-research`; `local-instructions` y `engram-init-workspace` → modos de `harness-audit`; `workspace-conventions` → referencias de `flow-close`; `starlight-docs-site` → modo de `flow-report`.

Fuera de alcance: la etapa de requerimiento (propuesta 2), cambiar el comportamiento de cualquier procedimiento que se mueve, y `tooling/legacy/catalog.json`.

## Objective

Hoy el catálogo tiene 20 skills y cada una carga su descripción (7.956 caracteres en total) en cada sesión de cada host. Siete tuvieron 7 sesiones o menos, o nunca se escribieron a mano, en 553 sesiones interactivas de cuatro hosts (2026-09-29 a 2026-11-14). Al terminar, el catálogo tiene 13 skills; cada procedimiento movido sigue alcanzable desde la skill que ya se carga para ese trabajo, con su condición de lectura y su disparador en la descripción, y el instalador retira las skills viejas en cada host al actualizar.

## Scope and acceptance

Dentro: mover contenido, scripts y assets de las 7 skills; actualizar cada enlace y mención en `content/`, `AGENTS.md`, `site/`, `README.md`, `llms.txt`, `_support/docs/` y los deltas de los specs; los tests con rutas fijas y la lista de skills de los pilotos; el `CHANGELOG`. Fuera: `tooling/` e `integrations/`. `SkillSource` (`tooling/management/types.go:28`) sigue nombrando `workspace-conventions`, porque `ownership.go:57-62` lo asigna a los registros de estados anteriores a la versión 3 sin origen, y cambiarlo haría fallar `hive update` en esas instalaciones; queda como identificador histórico y su único efecto es que `hive status` en una máquina limpia nombre una skill retirada.

- AC1 — `content/skills/` contiene exactamente 13 skills: flow-build, flow-close, flow-deploy, flow-figma-design, flow-plan, flow-report, flow-research, flow-verify, git-workflow, harness-audit, hive-ui-criteria, test-audit, unattended-delegation · false today because `ls content/skills` lists 20 · Observed in: `ls content/skills` y la vista previa `hive update --source <checkout> --dry-run`
- AC2 — Cada procedimiento movido se alcanza desde su destino: la descripción del destino nombra su disparador y su `SKILL.md` enlaza la referencia con una condición de lectura literal: walkthrough o demo de una interfaz (flow-verify), estado del proyecto o backlog y exposición a un aviso de seguridad (flow-research), crear o curar `HIVE.local.md` y la identidad de Engram solo a petición explícita (harness-audit), ordenar, curar, promover, migrar el layout o archivar material de `_support` (flow-close), y un sitio Starlight (flow-report) · false today because ninguno de esos destinos nombra esos disparadores · Observed in: las descripciones y los `SKILL.md` de destino
- AC3 — Ningún enlace, localizador `skill:` ni mención en `content/`, `AGENTS.md`, `site/`, `README.md`, `llms.txt`, `_support/docs/` o los deltas de este cambio apunta a una de las 7 skills retiradas como skill. Excepciones permitidas: el `CHANGELOG`, los registros archivados, la capability `local-instructions` (carpeta, título y la mención de `test-accounts/spec.md:10`), el título del requisito «Bootstrap and curation through the local-instructions skill», la carpeta de fixtures `tests/fixtures/workspace-conventions/`, las fixtures sintéticas de `tooling/` e `integrations/` y la ruta por defecto `SkillSource`; los specs vigentes cambian al archivar el cambio. La validación de enlaces y anclas, el tamaño (`flow-close/SKILL.md` ≤ 16.000 bytes) y las descripciones (≤ 1.024 caracteres) pasan · false today because `rg` encuentra unos 30 enlaces a esas rutas · Observed in: `rg` sobre esos árboles, `go test ./...` y `hive update --source <checkout> --dry-run`
- AC4 — Los tests de scripts movidos y la lista de skills de los pilotos apuntan a las rutas nuevas y pasan · false today because 4 tests de `tests/skills/` y `tests/pilot/flows.go:550` nombran las rutas viejas · Observed in: `python3 -m unittest discover -s tests/skills -p '*_test.py'` y `go test ./tests/pilot/...`
- AC5 — Tras desplegar en la instalación local, las 7 skills ya no existen en `~/.agents/skills/`, `~/.claude/skills/`, `~/.gemini/config/skills/` ni `~/.gemini/antigravity/skills/`, y sí las referencias nuevas · false today because esas carpetas o enlaces existen · Observed in: la instalación local de esta máquina tras `hive apply`
- AC6 — El `CHANGELOG` tiene una sección «Removed» que nombra cada skill retirada y su nuevo hogar · false today because no existe · Observed in: `CHANGELOG.md`

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Uso por skill, sesiones interactivas 2026-09-29 a 2026-11-14: Claude Code 211, Codex 95, Grok 165, OpenCode 82. flow-walkthrough 88 (ninguna escrita a mano), flow-status 19 (5 escritas a mano en Grok, 1 en Claude Code), starlight-docs-site 7, engram-init-workspace 6, workspace-conventions 5, security-advisory 3, local-instructions 0. Scripts de medición en el scratchpad de la sesión 246ec5d9.
- El manager arma el catálogo con un glob de `content/skills/*/SKILL.md` (`tooling/management/plan.go:229`) y retira lo ausente de la nueva versión para los consumidores seleccionados (`tooling/management/ownership.go:237`); precedentes sin tocar `tooling/`: `adf076cf` (fusión de workspace-archive y flow-security) y `5e6e8907` (hive-core).
- Validaciones sobre el árbol real: enlaces de la versión (`tooling/management/references.go`), anclas (`reference_anchors_test.go`: hay que conservar `#close-and-input-hygiene` y `#open-the-viewer`), `skill:` de `global.md` (`tests/content/skill_links_test.go`), tamaño (`skill_size_test.go`) y frontmatter (`skill_frontmatter_test.go`, descripción ≤ 1.024).
- Tamaños: flow-close 15.897 B (103 de margen), flow-verify 11.874, flow-research 9.245, harness-audit 8.730, flow-report 2.744. El cargador de OpenCode inyecta `harness-audit/SKILL.md` entero en el rol `hive-review-harness` (`integrations/opencode/context-loader.js:27-30`), por eso los modos van en referencias.
- Rutas fijas: `tests/skills/{workspace_archive,engram_workspace,starlight_assets,local_instructions_extract}_test.py`; `tests/pilot/flows.go:550` y `flows_test.go:756`; `tooling/management/types.go:28`, `integrations/target/catalog.go:12` y el `resolveBase` de cada integración (cosmético: `hive status` en una máquina limpia).
- `flow-walkthrough/scripts/` solo tiene un `__pycache__` sin versionar.

## Effects

None: the tasks take no effect on an external system.

## Delivery

Interactiva: parada antes del push para tu validación; después push, PR a `development` y merge. Revisión de código: Claude Code `/code-review`. El refresco local solo despliega contenido, sin recompilar el binario. El cuerpo del PR declara el cambio de tamaño de `global.md` y lista los lugares revisados, como pide `AGENTS.md`. No autoriza una release ni cambiar instalaciones a nivel de proyecto en otros repos.

## Review status and progress

2026-11-14 — Plan escrito.
2026-11-14 — Revisión del plan (`hive-review-plan`): 9 hallazgos aplicados. Se quitó el cambio de `SkillSource` (rompía estados antiguos y fixtures); el recorte de `flow-close` se recalculó en unos 120 bytes, no 40; AC3 se mide contra los deltas con excepciones explícitas; las páginas de flujo siguen enlazando `SKILL.md`; T2 fija que estado y aviso reemplazan los pasos 0–5 de research; T6 nombra la lista de reemplazo de los pilotos; AC5 incluye los enlaces de Antigravity; T5 valida tras el commit.
2026-11-14 — Build T1–T7 (d373e77e … 0648ad6a). Validación del usuario: aprobado (D11 A). `/code-review`: 9 hallazgos, 7 corregidos en 78ef0945 (incluido mover «Find the next ticket» a la referencia de status); revisión limitada: 8 corregidos en d6fa6b3d con línea `Unreviewed:`. Quedan como hallazgos del cierre: `SkillSource` (deliberado) y la suite histórica de pilotos (P3).
