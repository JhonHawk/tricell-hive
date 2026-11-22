# Tasks

Base: `development` en `d4f1eedb`; rama propuesta `refactor/leaf-restatements` (entrega `hold`: sin commit). T2 va primero y T1 después, en el hilo principal: T2 fija las hojas que `global.md` cita, y T1 ajusta los tests de contenido con el resultado de ambas. T3 y T4 van a hijos `hive-write-docs`, uno por grupo, después de T1, porque cada uno edita más de 3 archivos y sus escrituras no se cruzan.

### T1 — global.md cita en vez de resumir, con el índice por dominio

- [x] `global.md` no tiene los textos D o X de «Global», tiene el índice agrupado por dominio y mide 28.000 B o menos (AC7).

**Closes:** AC1, AC2, AC7

**Depends on:** T2

**Locations:** `content/guidance/global.md`, `tests/content/index_trigger_test.go`, `tests/content/single_home_test.go`

**Execution:** main thread

**Test approach:** check

**Changes:** como `design.md` «Global» e «Index», con subtítulos `###`. Los tests de contenido solo se ajustan donde registran frases que se mueven (las de T1 y las de T2) o el formato del índice; no se agrega ningún test nuevo (D13).

**Verification:** `wc -c content/guidance/global.md` ≤ 28000, medido con un borrador al empezar T1; lectura de la sección «Rule leaves» (un `###` por dominio, en orden, una hoja por línea); `go test ./tests/content/... -count=1` ok.

### T2 — Las hojas son la casa única, con D15–D17, D22

- [x] Los casos de «Leaves» de `design.md` quedan como puntero, y `foreign-commits`, `ticket-state`, `report-endings` y `merge-gate` dicen lo que deciden D15–D17, D22.

**Closes:** AC3, AC4, AC5, AC6

**Depends on:** none

**Locations:** `content/skills/hive-core/**` (las hojas de «Leaves», L18 y `close/close-report.md:30`), `content/skills/hive-core/plan/change-folder.md`, `content/skills/flow-close/references/batch-archive.md`; si D16 cambia la primera línea de `foreign-commits.md`, también su línea del índice en `global.md`

**Execution:** main thread

**Test approach:** check

**Changes:** como `design.md` «Leaves»; `python3 -I` en `change-folder.md:24` y `batch-archive.md`. Ninguna hoja pasa de su tope.

**Verification:** `go test ./tooling/management/ -run 'TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink|TestRealStagePacketsKeepAMarginUnderTheCap' -count=1` ok (anclas válidas y paquetes bajo su tope, incluido flow-close tras crecer `ticket-state.md`). Los tests de contenido que registran frases movidas los ajusta T1.

### T3 — Las skills y playbooks citan con ancla

- [x] Cada caso D o X de «Skills» de `design.md` queda como puntero con ancla o con su excepción local.

**Closes:** AC3, AC5, AC6

**Depends on:** T1

**Locations:** `content/skills/{create-design,hive-mode,flow-plan,flow-build,flow-verify,flow-close,flow-deploy,git-workflow,harness-audit}/**`

**Execution:** hijo `hive-write-docs`

**Test approach:** check

**Changes:** como `design.md` «Skills», cercanos incluidos.

**Verification:** el escaneo de `design.md` «Overlap scan» sobre `content/skills` (sin `hive-core`) solo devuelve pares cuya ubicación `design.md` clasifica P, E o I; `go test ./tests/content/... -count=1` ok.

### T4 — Los roles citan con ancla

- [x] Cada caso D o X de «Roles» de `design.md` queda como puntero con ancla.

**Closes:** AC3

**Depends on:** T1

**Locations:** `content/agents/**`

**Execution:** hijo `hive-write-docs`

**Test approach:** check

**Changes:** como `design.md` «Roles».

**Verification:** el escaneo de `design.md` «Overlap scan» sobre `content/agents` solo devuelve pares cuya ubicación `design.md` clasifica P, E o I; `go test ./tests/content/... -count=1` ok.

### T5 — La regla de citar queda escrita donde se mantiene la guía

- [x] `AGENTS.md` «Guidance structure» y `instruction-resources.md` dicen que una regla se cita fuera de su casa con su momento y un enlace con ancla, y que el índice va por dominio; `CHANGELOG.md` lo registra.

**Closes:** AC8

**Depends on:** T1

**Locations:** `AGENTS.md`, `README.md:476, 481`, `llms.txt:32`, `_support/docs/architecture/{instruction-resources.md:29, 33, tool-dependencies.md:52, agent-delivery.md:258}`, `_support/docs/harness-engineering/{rule-enforcers.md:38, 42, 56, session-review-rubric.md:51}`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:** una frase en la regla de casa única de `AGENTS.md`; las descripciones de «Questions and routes», «Evidence» y el índice en `instruction-resources.md`, README y llms.txt; los enlaces a `#questions-and-routes`; las casas que nombra `rule-enforcers.md`; el momento de foreign-commits en la rúbrica (D16); la condición de `Completion` en README (D22); una entrada en `[Unreleased]`.

**Verification:** el comando de AC8 sin resultados; `go test ./tests/content/... -count=1` ok.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./... -count=1` and `python3 -I -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Overlap scan | Final candidate, locally | Only P, E or I pairs remain | Covered by the implementation |
| UI check | n/a: no screen changes | — | — |
| Code review | Final candidate, before the validation stop | Claude Code `/code-review` | The session's review answer |

## Checkpoint

- Updated: 2026-11-22
- Stage: close · hold pause close
- Last commit: d4f1eedb
- Open findings: none
- Next: deliver: branch refactor/leaf-restatements, commit content + change folder, PR, merge; refresh install; archive
- Rules: flow-verify/stage-entry.md
- Principles: none

### Evidence

- go test ./tooling/management/ -run 'TestRealContent…|TestRealStagePackets…' · ok after T2
- wc -c global.md · 27981 B
- go test ./tests/content/... · ok after T1
- go test ./tooling/management/ real-content, packets, Antigravity, catalogue references · ok
- go test ./... -count=1 · ok
- tests/skills unittest · 412 OK
- overlap scan · 27 pairs, all on the keep list or coincidental (pr-reviews:8)
- global.md · 27906 B
- go test ./tests/content/... · ok after fixes
- management real-content, packets, catalogue, Antigravity · ok after fixes
- global.md · 27900 B
- go test ./... · ok final
- tests/skills · 412 OK final
- global.md · 28000 B

## Review status and progress

2026-11-22 — plan escrito; corregido tras dos lecturas de `hive-review-plan`.
2026-11-22 — T2 · uncommitted · unverified: single_home_test registers a ticket-state phrase that T2 removed; T1 moves that entry
2026-11-22 — T1 · uncommitted · unverified: none
2026-11-22 — T5 · uncommitted · unverified: none
2026-11-22 — T3 · uncommitted · unverified: the Reporting lines of bugfix, data-change and ticket-status playbooks stay as they are: TestLeafSkillInvariants pins their exact text (#588 scope)
2026-11-22 — T4 · uncommitted · unverified: none
