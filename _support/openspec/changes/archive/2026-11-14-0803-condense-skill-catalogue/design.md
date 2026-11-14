# Design

## Approaches

Una sola forma sensata para el conjunto: mover cada procedimiento a una referencia de la skill que ya se carga para ese trabajo, como se hizo con `code-audit.md` en `flow-research` y con `security-review.md` en `flow-verify` (`adf076cf`). Mantener las skills con descripciones más cortas no quita su coste por sesión, y un mecanismo por host para ocultarlas no es portable entre los siete hosts. Las elecciones por skill están en Decisions.

## Decisions

- `flow-status` → modo de `flow-research` (D8 A), pese a su uso escrito a mano en Grok.
- `workspace-conventions` → referencias de `flow-close` (D9 A), porque cuatro de sus cinco operaciones son higiene; las opciones descartadas fueron harness-audit (encaje raro para "ordena `_support`") y mantenerla.
- Sin skills puente (D10 A).
- Supuestos menores: los demás destinos de la propuesta; scripts y assets viajan con su procedimiento; los nombres de anclas se conservan; las fixtures sintéticas de `tooling/*_test.go` que crean su propio `workspace-conventions` se dejan como están; `tooling/legacy/catalog.json` no se toca; `SkillSource` y las plantillas `resolveBase` de las integraciones se quedan como identificadores históricos, porque `ownership.go:57-62` asigna `SkillSource` a registros antiguos sin origen y cambiarlo haría que `hive update` los tome como reubicados.
- Presupuesto de `flow-close/SKILL.md` (15.897 B, límite 16.000): los enlaces acortados de las líneas 12, 29, 32 y 33 ahorran unos 60 B; la línea 85 (estado como modo de research) suma unos 15 B; el disparador en la descripción, unos 60 B; la línea de lectura del mantenimiento de `_support`, unos 150 B. Hay que recortar unos 120 B de redacción sin quitar ninguna regla, comparando el diff línea a línea. D9 se aprobó sobre una estimación de 40 B.

## Proposed shape

| Origen | Destino |
| --- | --- |
| `flow-walkthrough/SKILL.md` | `flow-verify/references/walkthrough.md` |
| `flow-walkthrough/references/{viewer-publication,host-session-identity}.md` | `flow-verify/references/` |
| `flow-status/SKILL.md` | `flow-research/references/status.md` |
| `security-advisory/SKILL.md` | `flow-research/references/security-advisory.md` |
| `local-instructions/SKILL.md`, `references/transcript-sources.md`, `scripts/extract_candidates.py` | `harness-audit/references/local-instructions.md`, `references/transcript-sources.md`, `scripts/extract_candidates.py` |
| `engram-init-workspace/SKILL.md`, `scripts/init_workspace.py` | `harness-audit/references/engram-workspace.md`, `scripts/init_workspace.py` |
| `workspace-conventions/SKILL.md`, `references/{artifact-layout,archive}.md`, `scripts/archive_sessions.py` | `flow-close/references/support-maintenance.md`, `references/{artifact-layout,archive}.md`, `scripts/archive_sessions.py` |
| `starlight-docs-site/SKILL.md`, `assets/{spec-site,user-manual}`, `scripts/contrast-check.py` | `flow-report/references/docs-site.md`, `assets/starlight/{spec-site,user-manual}`, `scripts/contrast-check.py` |

En `flow-research`, la línea de lectura del estado y la del aviso de seguridad dicen que esa referencia reemplaza los pasos 0 a 5 y usa su propio cierre, porque ambos procedimientos son de solo lectura y no reproducen la pregunta abierta ni la lectura del ticket. Las páginas de flujo del sitio siguen enlazando el `SKILL.md` de su skill (Status enlaza `flow-research/SKILL.md`), así que el requisito «Flow pages describe without copying rules» no cambia. Cada `SKILL.md` de destino gana una línea con su condición de lectura, como la de la auditoría de código, y su descripción gana el disparador. Cada referencia nueva abre diciendo cuándo se lee y desde qué skill. La regla global de identidad de Engram pasa a nombrar el modo de harness-audit y conserva "solo a petición explícita".

## Impact

- Consumidores: unos 30 enlaces en 20 archivos de `content/` (incluida `global.md` líneas 31, 57, 71 y 76, los roles `hive-review-ux` y `hive-write-docs`), `AGENTS.md` líneas 53 y 76, `content/guidance/README.md`, `site/` (skills, index, flows/status, flows/security, viewer, getting-started si aplica), `README.md`, `llms.txt`, `_support/docs/architecture/` y `_support/docs/harness-engineering/`, tres specs.
- Instalaciones: el siguiente `hive update` retira las 7 skills en los hosts y el alcance seleccionados; las instalaciones a nivel de proyecto las conservan hasta reinstalar ese proyecto. Un archivo instalado editado a mano bloquea la actualización, como ya ocurre hoy.
- Comandos: desaparecen `/flow-status`, `/flow-walkthrough`, `/security-advisory`, `/local-instructions`, `/engram-init-workspace`, `/workspace-conventions`, `/starlight-docs-site` y sus equivalentes en cada host.

## Recovery

Revertir el merge restaura las 7 skills; el siguiente `hive update` las reinstala. No hay datos que migrar.
