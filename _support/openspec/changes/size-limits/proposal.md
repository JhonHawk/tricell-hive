# size-limits — Topes de tamaño por concepto, no por archivo

| Campo | Valor |
| --- | --- |
| Estado | In build |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `013a9725` · rama `refactor/size-limits` |
| Playbook | refactor |

## Decisiones

2026-11-22
- **D2 (supuesto)** — `SKILL.md` se queda en 16.000 B, medido en bytes (descartado: medir tokens; la equivalencia es estimada y 16.000 B cabe en los 5.000 tokens que Claude Code conserva tras compactar).
- **D3 (supuesto)** — Se quedan `description` (1.024 caracteres), stage packets y partes de Antigravity (24.000 B) y el conteo de 26 lecturas de la ruta plan/build (descartado: quitarlos; son cortes reales de host o miden la ruta entera).
- **D4 (supuesto)** — Se quedan el global (30.000 B) y la voz (3.072 B) como presupuesto de lo que se carga siempre (descartado: quitarlos; sin presupuesto el global llegó a 52.508 B).
- **D5 A** — ¿Qué pasa con los topes de hojas (4.000/7.000/9.000 B y sus excepciones) y de playbooks (7.000 B)? Pasan a un tope único de lectura entera de 24.000 B y desaparece el mapa de excepciones.
- **D6 A** — ¿Las referencias pasan de 16.000 a 24.000 B? Sí, el mismo tope de lectura entera.
- **D7 A** — ¿Dónde vive qué hacer cuando un tope falla? En el mensaje de cada test; cada constante cita su fuente (host y versión, o decisión del usuario). Ninguna constante sube en el mismo PR que necesita el espacio.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> Los test que mide el limite de tamaños realmente no estan sirviendo poruqe se tienen que ajustar cada vez que movemos algo, esto requiere research para ver cuál seria un tope idual, bajo que concepto, si hoja, skill, etc, y qué hacer ante cada situación

## Objective

Hoy `tests/content` tiene topes por archivo sin apoyo en un host (hojas 4.000/7.000/9.000 B con excepciones, playbooks 7.000 B, referencias 16.000 B) que dejaron archivos a 99,9 % del tope y forzaron recortes de contenido. Al terminar quedan tres conceptos: el tope de host de `SKILL.md` (16.000 B), un tope de lectura entera de 24.000 B para todo otro archivo que el modelo lee de una vez, y el presupuesto de lo que se carga siempre (global y voz). Cada test dice qué hacer cuando falla.

## Scope and acceptance

Dentro: los tests de tamaño de `tests/content`, sus mensajes y comentarios, y cada texto del repositorio que repite esos números. Fuera: los topes de `tooling/` (packets, Antigravity, reglas personales), el contenido distribuido y el texto de los cambios archivados.

- AC1 — Una hoja, principio o playbook de hasta 24.000 B pasa `go test ./tests/content/...`, y uno de 24.001 B falla con un mensaje que dice qué hacer · false today because el tope es 4.000 B salvo 3 excepciones (`leaf_skill_test.go`) · Observed in: `go test ./tests/content/...`
- AC2 — Una referencia de skill de hasta 24.000 B pasa y una de 24.001 B falla con un mensaje que dice qué hacer · false today because `skillReferenceLimit` es 16.000 · Observed in: `go test ./tests/content/...`
- AC3 — Un `SKILL.md` de 16.001 B falla con un mensaje que nombra el corte de Claude Code tras compactar y qué mover · false today because el mensaje solo da el tamaño · Observed in: `go test ./tests/content/...`
- AC4 — No queda en `tests/content` ningún mapa de excepciones ni tope propio por hoja · false today because existen `leafSizeExceptions` y `leafReportEndingsSizeLimit` · Observed in: `rg` sobre `tests/content`
- AC5 — `AGENTS.md`, HA-SK-02 (`harness-audit`), `harness-audit-rules.md`, `rule-enforcers.md` e `instruction-resources.md` dan los topes nuevos y no repiten 16.000 B para referencias · false today because los cinco dicen 16.000 B · Observed in: `rg -n '16,000' ` sobre esos archivos

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Host limits (2026-11-22, research child): Claude Code keeps the first 5,000 tokens per invoked skill after compaction, 25,000 total (code.claude.com/docs/en/skills); Antigravity truncates a rule file past 24,000 B; Codex does not count or truncate the global AGENTS.md in `project_doc_max_bytes` (codex-rs main, `agents_md_manager.rs`), not checked on 0.171.0.
- Repository (2026-11-22, `013a9725`): files sat at 15,993–15,996 of 16,000 B and 6,989 of 7,000 B; at least 3 commits dropped content to fit; the global budget moved 22,000 → 27,500 → 30,000 on 2026-11-20; stage packets measure 19,838–23,346 B.
- `go test ./tests/content/...` · passes on the base · 2026-11-22

## Effects

None: the tasks take no effect on an external system.

## Delivery

`automatic`: revisión, push, PR y merge sin parar; una aprobación humana pendiente sigue bloqueando el merge. Revisión: Claude Code `/code-review`. Trabajo en el worktree `.claude/worktrees/size-limits`. No autoriza release ni refresco de otra cosa que la instalación local que AGENTS.md prevé.

## Review status and progress

2026-11-22 — Sin plan: carpeta escrita por flow-build para el delta de spec; sin revisión de plan.
2026-11-22 — /code-review (medium): H1 delta requirement under ADDED moved to MODIFIED; H2 language leaf keeps its 4,000-byte budget as a loading budget; H3 a packet member also fits its packet; H4 wholeReadLimit no longer cites Antigravity. H5 (other active change folders state 4,000 B) and H6 (main specs change at archive) not fixed: other works' records and the hygiene batch merge.
2026-11-22 — second /code-review (fix diff): H7 the ADDED requirement said no file has a limit of its own while the language leaf has a budget; scoped to whole-read files. H8-H12 (nits and one pre-existing line) go to the close findings.
2026-11-22 — close findings fixed on `fix/size-limits-findings`: H8 the leaf fixture now checks the language-leaf budget; H9 AGENTS.md names that budget; H10 the delta states the packet's 500-byte margin; H11 the budget message names `content/skills/`; H12 the vendored-references line says the 16,000-byte limit is the former one; an oversize leaf is reported once, by the leaf test. H5 needed no change: no other active delta states a leaf limit, only a dated progress line of comment-audit.
2026-11-22 — /code-review of the findings fix: the principle index and non-Markdown files of a domain folder lost their size check with the leaf skip, and an oversize leaf was named by its bare name; fixed in 166101db. Its review: playbooks now named by path and the principle index listed as a whole-read file (unreviewed fix); a non-Markdown file of a domain folder stays sized, as at the base.
