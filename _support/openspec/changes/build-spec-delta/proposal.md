# build-spec-delta — Carpeta de cambio para un cambio sin plan que altera una spec

| Campo | Valor |
| --- | --- |
| Estado | En curso · entrega `automatic` |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `d3d2b46c` · rama `fix/build-spec-delta` |
| Playbook | feature (hallazgo H10 de #537, cambio sin plan) |

## Request

> «a» — resolver ahora el hallazgo H10 de #537: `flow-build` no dice qué hacer cuando un cambio pequeño sin plan altera un comportamiento descrito en una spec.

## Decisions

2026-11-20
- **D3 A** — Cuando un cambio pequeño sin plan altera un comportamiento descrito en una spec, `flow-build` escribe la carpeta de cambio con las plantillas de `flow-plan`, sin revisión del plan, y el cierre la archiva con `apply_deltas.py`. Descartado: B, mandarlo a `flow-plan`, porque cada cambio de una línea a la guía pagaría un plan y un revisor. Engram: `decision/spec-change-needs-change-folder`.
- **Supuesto** — El procedimiento vive en una referencia situacional, `flow-build/references/spec-delta.md`, que `flow-build` enlaza con su condición, porque el presupuesto de lectura de plan y build (`TestRequiredProcedureReadingFootprint`) tenía 27 B libres y los arreglos de la revisión pedían unos 300. Descartado: recortar texto no relacionado. Se revierte moviendo los cuatro pasos a `flow-build`.
- **Supuesto** — Para entrar en el presupuesto, el enlace a `Tests` de `flow-build` paso 2 absorbe el puntero aparte de «when a check fails», y salen dos punteros a «Proportionality» en términos que la guía global define. Se revierte restaurándolos.

## Acceptance criteria

- AC1 — Un cambio sin plan que altera un comportamiento descrito en `<Specs>/specs/<capability>/spec.md` tiene en `flow-build` un puntero condicional a `references/spec-delta.md`, que pide proposal, tasks y delta, sin `hive-review-plan`, con `check_change.py` en 0 antes de editar · falso hoy porque `flow-build` no lo menciona · Observed in: `content/skills/flow-build/SKILL.md` y `content/skills/flow-build/references/spec-delta.md`
- AC2 — `flow-build` paso 2 registra las tareas de cualquier trabajo con carpeta de cambio, no solo de un plan · falso hoy porque dice «When the plan has a change folder» · Observed in: `content/skills/flow-build/SKILL.md`
- AC3 — `hive-mode` y la spec «Entry router with playbooks» registran el playbook en el `proposal.md` lo escriba quien lo escriba · falso hoy porque nombran solo a `flow-plan` · Observed in: `content/skills/hive-mode/SKILL.md` y el delta de esta carpeta
- AC4 — La lectura requerida de plan y build no supera la línea base · Observed in: `go test ./tests/content -run TestRequiredProcedureReadingFootprint`
