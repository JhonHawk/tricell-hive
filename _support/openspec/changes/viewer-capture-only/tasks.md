# Tasks

Base: `development` @ `7ae26b0e`; rama `feat/viewer-capture-only` en `.claude/worktrees/`. T1 la hace un hijo `hive-build-backend`, porque edita más de 3 archivos de `tooling/cli` y necesita más de 5 lecturas grandes. T2 y T3 se quedan en el hilo principal: son reglas distribuidas, y AGENTS.md exige revisar todo el comportamiento que gobiernan. T4 la hace un hijo `hive-write-docs`, desde las fuentes ya cambiadas. T2 y T4 dependen de T1 solo por la sintaxis final del comando.

### T1 — Comando `hive view work capture`

- [x] `capture` publica un lote de capturas en una llamada, como fija «Proposed shape» de `design.md`: abre el trabajo, arranca, reanuda o vuelve a entrar en la etapa, usa `HEAD` como candidato, y calcula al enviar `supersedes` y «· #N» desde los títulos del servidor. Se encola como acción propia. `--help` es autosuficiente, y `capture` aparece en los flujos `build` y `verify` del contrato.

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** none

**Locations:** `tooling/cli/view_work.go`, `view_work_evidence.go`, `view_work_lifecycle.go`, `view_work_queue.go`, `view_work_journal.go`, `view_work_contract.go`, tests nuevos `tooling/cli/view_work_capture_test.go`; ajustes en `view_work_contract_test.go`

**Execution:** hijo `hive-build-backend`

**Test approach:** tdd — tests escritos antes del código, contra el servidor de prueba que ya usan `view_work_sync_test.go` y `view_work_queue_test.go`, uno por criterio y estos casos:
- dos capturas offline de la misma pantalla salen como «#1» y «#2»;
- una captura de `build` después de una de `verify` reanuda la etapa `build` en pausa;
- una captura sobre una etapa terminada abre una ocurrencia enlazada;
- `HEAD` separado o sin `origin/HEAD` responde `candidate_branch_required` y no publica nada;
- repetir con el mismo candidato da `unchanged`, y con otro commit da «#N+1».

El repositorio de prueba define `refs/remotes/origin/HEAD`.

**Changes:** el verbo nuevo, la acción encolable, su ayuda y su lugar en el contrato. No cambian `open`, `sync`, `evidence`, `close`, el servidor (`tooling/viewer/`) ni la interfaz.

**Verification:**
- `go test ./tooling/cli/ -run 'Capture|Contract|Queue' -count=1` pasa, y `go test ./...` pasa.
- In vivo: un binario temporal y un visor con `--state-dir` temporal en el directorio temporal de la sesión, en otro puerto. Es un proceso de la tarea y se detiene por su PID.
  - Dos `capture` de la misma pantalla y uno de otra: `work show` muestra «· #1» y «· #2», con `supersedes`, y la otra pantalla en «· #1».
  - Se repite con el visor detenido y luego arrancado: hay una vista por captura, numerada igual.

### T2 — Guía distribuida: solo capturas

- [x] La referencia `viewer-publication.md` queda en ≤ 6.000 bytes y describe solo:
  - cuándo hay sesión: la primera captura de un flujo, con `--key` y un `--summary` de un párrafo;
  - `work capture` por lote;
  - `queued` y qué conservar;
  - la opción de abrir el visor, solo en las paradas de validación de interfaz y en forma condensada;
  - en flow-close, `work close` con su lectura de vuelta, o un `work sync` a `in_review`, `paused` o `awaiting_validation` cuando el trabajo espera;
  - la eliminación como efecto crítico y la higiene de entradas.

  Estos archivos apuntan a esa regla sin repetirla:
  - `flow-walkthrough/SKILL.md`, incluida la `description` del frontmatter: hoy anuncia etapas y documentos y «Use from plan»;
  - `flow-plan`: SKILL l. 8 y `plan-handoff.md` l. 21;
  - `flow-build`: l. 8, l. 15 (`#open-the-viewer`) y l. 20 (declaraciones de tareas por `#tasks`);
  - `flow-verify`: l. 8, 15 y 17;
  - `flow-close`: l. 10, 12 y 21;
  - `ui-verification.md` l. 25;
  - `workspace-conventions/references/artifact-layout.md` l. 27 (`#close-and-input-hygiene`, «Publish API evidence»).

**Closes:** AC5, AC6, AC7

**Depends on:** T1 (sintaxis final)

**Locations:** `content/skills/flow-walkthrough/SKILL.md`, `content/skills/flow-walkthrough/references/viewer-publication.md`, `content/skills/flow-walkthrough/references/host-session-identity.md` (solo si deja de tener consumidor), `content/skills/flow-plan/SKILL.md`, `content/skills/flow-plan/references/plan-handoff.md`, `content/skills/flow-build/SKILL.md`, `content/skills/flow-verify/SKILL.md`, `content/skills/flow-verify/references/ui-verification.md`, `content/skills/flow-close/SKILL.md`, `content/skills/workspace-conventions/references/artifact-layout.md`, `content/agents/review/hive-review-ux.md` (solo si cambia su sentido)

**Execution:** hilo principal

**Test approach:** check

**Changes:** se quitan de la guía las declaraciones de etapa y tarea, el registro de tareas, los snapshots de documentos, la lectura del contrato, `session get`, `resume`/`reconcile`, la gestión compartida, la entrega y CI, y la opción de abrir el visor en la parada del plan. Se conservan la regla de identidad (sin `--host` ni `--native-session-id`), la cola, la eliminación como efecto crítico y la higiene de entradas. Antes de editar se buscan todos los lugares que dependen de este comportamiento (guía global, skills, roles, `site/`, `_support/docs/architecture/`, `tests/pilot/`) y se listan en el PR, como manda AGENTS.md.

**Verification:**
- `wc -c content/skills/flow-walkthrough/references/viewer-publication.md` ≤ 6000.
- El `rg` de AC5 sobre `content/` solo devuelve los casos que permite ese criterio.
- `rg -n -i 'viewer' content/skills/flow-plan/` no abre sesión ni publica snapshots.
- Cada ancla `viewer-publication.md#…` que aparece en `content/` existe como encabezado en la referencia nueva. Se comprueba listando las anclas con `rg -o 'viewer-publication\.md#[a-z-]+' content/` contra los encabezados, porque `tests/content` no revisa anclas.
- `go test ./tests/content/...` pasa (tamaños y enlaces).

### T3 — Delta del spec

- [x] El delta `specs/walkthrough-viewer/spec.md` modifica tres requisitos: «Viewer-first walkthrough publication» (l. 133), «Proportional identity-aware publication procedure» (l. 235) y «Declared flow stages and immutable documents» (l. 243). Añade el requisito del comando `capture`, y `check_change.py` sale con 0. `workflow-lifecycle` no cambia.

**Closes:** AC6

**Depends on:** none

**Locations:** `_support/openspec/changes/viewer-capture-only/specs/walkthrough-viewer/spec.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:**
- El procedimiento publica solo capturas.
- La sesión empieza con la primera captura, y la opción de abrir el visor queda solo en las paradas con capturas.
- «User questions SHALL publish awaiting_validation first» se aplica solo a un trabajo que existe en el visor.
- Se añade «Capture publication in one command».
- No cambian «Declarative stage publication without a change folder» ni la cola, porque los verbos siguen existiendo.

**Verification:** `python3 ~/.claude/skills/flow-plan/scripts/check_change.py _support/openspec/changes/viewer-capture-only` sale con 0.

### T4 — Sitio y documentos de arquitectura

- [x] Las páginas del sitio describen la publicación solo de capturas y enlazan la skill sin copiar sus reglas: `site/src/content/docs/viewer.md`, `flows/{overview,plan,build,verify,close,research}.md`, `skills.md` (l. 34) e `index.mdx` (l. 104). `README.md` (l. 480-486) y `llms.txt` (l. 57) dejan de describir declaraciones de etapa y proyecciones de tareas. `deployment-manager.md` (l. 176 y 195) e `instruction-resources.md` (l. 17) nombran `work capture`.

**Closes:** AC8

**Depends on:** T1, T2

**Locations:** `site/src/content/docs/viewer.md`, `site/src/content/docs/flows/*.md`, `site/src/content/docs/skills.md`, `site/src/content/docs/index.mdx`, `README.md`, `llms.txt`, `_support/docs/architecture/deployment-manager.md`, `_support/docs/architecture/instruction-resources.md`

**Execution:** hijo `hive-write-docs`

**Test approach:** check

**Changes:** solo descripción del comportamiento; no cambian las secciones de arranque, idioma, grupos ni gestión manual.

**Verification:** `rg -n 'One declaration per stage change|work sync|tasks register|stage/task projections|session milestones' site/ README.md llms.txt _support/docs/architecture/` no devuelve descripciones de publicación por agentes, salvo el `work sync` de espera en el cierre. La compilación del sitio pasa con el comando que registre `HIVE.local.md`, o, si no hay, el de `site/package.json`.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar la tarea, en local | El comando de su verificación pasa | Cubierto por la implementación |
| Suite completa | Candidato final, en local | `go test ./...` | Cubierto por la implementación |
| Prueba in vivo de `capture` | Candidato final, visor temporal propio | Salida de `work show` con «· #1», «· #2» y `supersedes`; caso en cola | Cubierto por la implementación (proceso propio, detenido por PID) |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` (D5 A) | La respuesta de revisión |

Sin recorrido de interfaz: la interfaz no cambia (D3 A).

## Checkpoint

- Updated: 2026-11-08
- Stage: build · 5
- Last commit: d878426e
- Open findings: none
- Next: hive-verify-task de T1, luego flow-verify

### Evidence

- go test ./... -count=1|pass|26e9e671
- go test ./tests/content/... -count=1|pass|e9a17e4c
- pnpm build (site/)|pass, 17 páginas|de3ea595
- CI: master-source-branch.yml solo aplica a PR hacia master|n/a|d878426e

## Review status and progress

2026-11-08 — Plan escrito.
2026-11-08 — Revisión `hive-review-plan`: 5 bloqueantes y 6 no bloqueantes, todos aplicados.
- Estados de espera en flow-close sin tocar `workflow-lifecycle`.
- Tres requisitos modificados en el delta.
- La iteración se calcula al enviar, desde el servidor, y `capture` se encola como acción propia.
- Lugares de la guía y del sitio añadidos a T2 y T4.
- AC5 pasa a una lista permitida y AC7 a ≤ 6.000 bytes.
- Etapas en pausa o terminadas, y el candidato en `HEAD` separado, cubiertos en T1.
2026-11-08 — T2 · e9a17e4c · unverified: sintaxis de capture contra el binario final de T1
2026-11-08 — T3 · f638c16b · unverified: none
2026-11-08 — T4 · de3ea595 · unverified: sintaxis contra el --help final de T1
2026-11-08 — T1 · 26e9e671 · unverified: rechazo de una primera captura en cola (limpieza del handle) sin test
