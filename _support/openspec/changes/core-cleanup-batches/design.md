# Design

## Approaches

Solo un enfoque tiene sentido para la herramienta: extender `tests/content` en Go con una prueba que resuelva cada «sección» citada junto a un enlace (un título `#` o una etiqueta en negrita `**Nombre.**` del archivo de destino) y cada «step N» citado junto a un enlace a un `SKILL.md` (un ítem numerado `N.` en el destino). Es el primer escalón de «Reuse before writing»: `tests/content` ya revisa que los enlaces y anclas existan (`skill_links_test.go`, `leaf_links_test.go`, `tooling/management/references.go`), y una herramienta aparte duplicaría ese parser. Límite: la prueba confirma que el destino existe, no que diga lo que la cita le atribuye; los casos semánticos (como B-1, «git-workflow step 5» que existe pero habla de hooks) los corrige el lote 1 a mano.

Prototipo medido en la base `8aa361cb` (2026-11-22): con títulos y etiquetas en negrita como destinos válidos, reporta 1 cita colgando (PL-1, `scope-changes.md:12`); sin aceptar etiquetas en negrita reportaría 6 falsos positivos de «Decide and list».

Para los lotes no hay alternativa: cada hallazgo se corrige en su casa y los demás lugares quedan como puntero con ancla, como fija «A rule is cited outside its home, never restated» (`_support/openspec/specs/context-delivery/spec.md`).

## Decisions

- **D4 A** — Evidencia para relajar o quitar una regla: la decisión del usuario más una comparación de sesiones reales con `go run ./tests/pilot --audit-session` antes y después, o la razón que la propia regla declara verificada como vencida en documentación o versión actual. Quitar una repetición, un puntero colgando o una referencia circular no es relajar. Va en `AGENTS.md` «Measurement» y en la spec «Recurrence evidence from floor candidates».
- **D5 A** — «never» y «always» solo para invariantes (autorización, secretos, efectos irreversibles, preservación); en lo demás, alcance y razón. Aplica al texto que se escribe o edita, incluidos estos lotes; sin reescritura masiva. Va en `AGENTS.md` «Guidance structure» y en la spec «Writing target and support floor».
- **D7 A** — `security-boundaries.md` lista los ocho temas (entrada no confiable, permisos, ejecución de procesos, exposición de datos, autenticación, secretos, criptografía, dependencias) y `hive-review-security` apunta a esa lista.
- **Entrega `hold`** — Sin commit: el árbol de trabajo verificado es el resultado. Si después se entrega por PR, un commit por lote (T1+T2, T3, T4+T5, T6, T7) deja revertir uno sin los demás.
- **supuesto** — La casa de «cuándo una captura cuenta como reemplazada» es `project-layout/references/layout.md#retain-evidence` (convención de retención); `pr-captures.md`, `support-material.md` y `viewer-close.md` apuntan a ella (UI-1, CL-4). Descartada: `support-material.md` como casa.
- **supuesto** — Las condiciones del `--force-with-lease` desatendido viven en `git-commit.md`; `unattended-effects.md` apunta a ellas (G-C7, DL-1). Descartada: moverlas a `unattended-delegation`.
- **supuesto** — `ticket-state.md` es la casa de «crear tickets solo en el tracker primario» e incluye la excepción que ya aplica `create-issue` («salvo que el usuario nombre otro tracker o repositorio»); delta MODIFIED en «Tickets follow the tracker their key declares» (TR-4). Descartada: quitar la excepción de `create-issue`, que quitaría un comportamiento que hoy existe.
- **supuesto** — El registro `Playbook` del checkpoint, que la spec de `hive-mode` exige, lo leen `flow-build` y `flow-verify` al entrar, vía `checkpoint.md` (PL-3). Descartada: borrar el registro, que exigiría un delta REMOVED.
- **supuesto** — Los «registros de cierre pendiente» (CL-1) y el «close record» de `time.md:19` (TR-2) son el registro del resultado que escribe `flow-close` paso 7, que pasa a listar los pasos del cierre que esperan respuesta; las menciones apuntan a ese paso. Descartada: `pending/<change-id>`, que guarda trabajo dejado abierto a propósito.
- **supuesto** — Una hoja que pase de 4.000 B al recibir texto lleva la parte nueva a una referencia propia `hive-core/<dominio>/<hoja>/<nombre>.md`, como `local/local-write/worktree.md`; no se pide excepción.
- **supuesto** — Si la unión plan/build de `reading_footprint_test.go` crece al cerrar T3 o T6, el build se detiene y lo pregunta.

## Proposed shape

Inventario de hallazgos de la sesión del 2026-11-22 (siete agentes de investigación más verificación en el hilo principal), por lote. Clase: X conflicto, D colgante o destino equivocado, C circular, R repetición parcial, U inalcanzable, M mal ubicada, P paquete. Las líneas son las de la base `8aa361cb`; el build las relee antes de editar.

### Lote 1 — Referencias colgantes, destinos equivocados, circulares y conflictos (T3)

- **G-C1** D — `git-workflow/SKILL.md:21` «run the checks of flow-verify step 5 under its condition»: flow-verify no tiene condición. Apuntar a `checks.md#local-ci`.
- **G-C2** X — Disparador «rama de entorno» redactado de tres formas (`promotion-path.md:3`, `git-workflow/SKILL.md:23`, `flow-deploy/SKILL.md:16`). Alinear a la hoja.
- **G-C3** X — `git-workflow/SKILL.md:15` pone el chequeo de commits ajenos en el análisis; `foreign-commits.md:3` lo pone antes de validación o push. Puntero al momento de la hoja.
- **G-C4** D — `ticket-state.md:8` y `git-workflow/SKILL.md:25` citan «la parada del paso 8», que el paso 8 no define. Apuntar a `merge-gate.md`.
- **G-C7 / DL-1** C — `git-commit.md:8` («que unattended delegation permite») ↔ `unattended-effects.md:5` («que git-workflow step 4 permite»). Condiciones en `git-commit.md`; `unattended-effects.md` apunta.
- **G-C8** M — `pr-review-deploy.md:11` (merge de una PR que el trabajo no escribió) no está en `merge-gate.md`. Mover a la hoja.
- **B-1** D — `flow-deploy/references/deployment-checks.md:11` cita «git-workflow step 5» (hooks) para leer workflows de CI. Apuntar a `checks.md#local-ci`.
- **B-2** D — «Tests section of flow-build» en `test-audit/SKILL.md:3,47` y `flow-plan/assets/change-folder/tasks.md:19`; `test-audit/SKILL.md:21` («the Tests leaf sets» para un repositorio sin runner) debe enlazar `tests-no-runner.md`. Apuntar a la hoja Tests.
- **PL-1b** D — `flow-plan/assets/change-folder/proposal.md:39` cita flow-plan «Revise a plan in progress» sin enlace; vive en `change-folder.md`. Nombrar la hoja (la prueba de T2 no lo ve porque no hay enlace).
- **B-3** D — `unattended-delegation/SKILL.md:18` enlaza Tests para la infraestructura de pruebas, que vive en `tests-no-runner.md:9`. Reenlazar.
- **B-11** X — `checks.md:19` vs `flow-verify/SKILL.md:24` sobre CI local. Alinear a la hoja.
- **RV-2 / CL-8** C — P1 bloqueante: `close-report.md:49` ↔ `review-findings.md:5`. Definir en la escala de `close-report.md`; `review-findings.md` apunta.
- **UI-1** D — `viewer-close.md:7` señala la casa equivocada para capturas reemplazadas. Apuntar a `layout.md#retain-evidence`.
- **UI-2** X — Dos hojas enlazadas con la etiqueta «viewer publication» (`walkthrough.md:5`, `first-view.md:5`, `flow-verify/SKILL.md:21`, `AGENTS.md`); `walkthrough.md:7` etiqueta `ui-plan.md` como «flow-plan». Reetiquetar.
- **UI-3** X — `hive-mode/SKILL.md:12` («no publica nada al visor») vs `:14` (demo por `walkthrough`). `:14` solo enruta a flow-verify.
- **UI-9** X — `design-stage.md:9` cuenta la memoria como fuente del archivo de Figma; `:10` la omite. Alinear.
- **LO-1** D/X — `test-accounts.md:5` cita «Local instructions» para el rechazo del host en un worktree; la regla está en `local-write/worktree.md:6`, y hay tres redacciones (`worktree.md:6`, `test-accounts.md:5`, `local-script.md:5`). `worktree.md` es la casa.
- **CL-1** D — Registros de cierre pendiente sin definición (`close-hygiene.md:15`, `close-modes.md:10,17`, `branch-cleanup.md:7,15`). `flow-close` paso 7 dice que el registro del resultado lista los pasos que esperan respuesta, y las menciones apuntan a ese paso (supuesto).
- **CL-2** X — «step 2» ambiguo en `close-hygiene.md:19`. «item 2 below».
- **CL-9** X — Un cierre por pausa en trabajo desatendido cae en dos filas de `close-modes.md`. Excluir las pausas de la fila 2.
- **TR-2** D — `time.md:19` «the work's close record in persistent memory» sin definir: apuntar al registro del resultado de `flow-close` paso 7 (supuesto).
- **TR-3** X — Orden de herramientas: `trackers.md:5` (CLI → MCP → REST) vs `trackers/linear.md:3` (MCP primero). Alinear Linear a la hoja.
- **TR-4** X — Crear tickets en otro tracker: `ticket-state.md:11` vs `create-issue/SKILL.md:17` y `references/trackers.md:7`. Ver supuesto.
- **TR-5** X — `ticket-state.md:5,8` y `flow-close/SKILL.md:18` incondicionales frente a `Completion` (`:9`, `project-settings.md:17`). Agregar «sin `Completion`».
- **BU-1** D — Cuatro punteros a `business-rules.md` para el formato de las reglas, que está en `business-record.md` (`create-prd/SKILL.md:16`, `create-design/references/example-record.md:7`, `rule-cards.md:3`, `create-design/SKILL.md:31`). Reenlazar.
- **MS-1** X — `dependency-security.md:7,45` termina en la fila de análisis y en un bloque que hive-mode paso 5 no define. Bloque de respuestas de la fila de análisis, con el arreglo como una respuesta.
- **MS-2** D — `hive-mode/references/code-audit.md:71` cita bloques del paso 5 que no existen. Apuntar a la fila de análisis de `report-endings.md`.
- **MS-3** C — `report-endings.md:10` ↔ `hive-mode/SKILL.md:27,73`; `report-endings.md:90` ↔ `plan-handoff.md:28`. Destino explícito en `report-endings`; los otros apuntan.
- **MS-7** X — `ticket-status.md:7` vs `:17`. Alinear a `:17`.
- **PL-1** D — `scope-changes.md:12` cita flow-plan «Revise a plan in progress», que vive en `change-folder.md`. Enlace con ancla; lo arregla T2 junto con la prueba que lo detecta, para que ninguna tarea deje la suite en rojo.
- **PL-2** D — `external-effects.md:9` «step 4» debe ser «step 6»; etiqueta «outside the repository» en `unattended-delegation/SKILL.md:17`.
- **PL-4** D — `create-design/references/approval.md:21`, `flow-plan/SKILL.md:27` «as flow-build sets» para el checkpoint; `flow-build/SKILL.md:10` agrega pasos de recuperación que no están en `checkpoint.md:9`. Apuntar a `checkpoint.md` y mover la cláusula a la hoja.
- **PL-5** X — `refactor.md:5` manda a flow-plan lo que `proportionality.md:5` trata como mecánico; `refactor.md:8-10` tiene una regla de tests que no está en `tests.md`. Usar el criterio de proportionality y mover la regla a `tests.md`.
- **PL-6** X — `flow-build/SKILL.md:22` (tarea operativa sin CI ni verify; parada de validación para un cambio mecánico) no está en `proportionality.md`. Agregar la cláusula a la hoja.

### Lote 2a — Repeticiones parciales a punteros (T4)

- **G-C5** — `git-workflow` paso 9 repite la espera de CI de `merge-gate.md:5`; el paso 5 es un puntero que el 4 ya cubre (fusionar). **G-C9**: `flow-verify/SKILL.md:25` alinea con `merge-gate`.
- **B-7** — `bugfix.md:8,10`, `hotfix.md:8`, `infra-ci.md:9` (queda solo la mitad propia), `prose.md:3` (amend), `hive-typescript/SKILL.md:11`.
- **RV-1** — `agents/review/hive-review-code.md:12`, `agents/quality/hive-verify-task.md:21` repiten 3 de los criterios de rechazo de `tests.md` punto 4.
- **UI-4** — `walkthrough.md:5` repite el disparador de `viewer.md`; agregar «un hijo nunca publica».
- **UI-8** — `ui-plan.md:7` repite el disparador de `ui-example.md:5`.
- **UI-11** — `ui-change.md:18` lista hojas del visor que hive-mode no lee; `:11` parafrasea first-view.
- **UI-12** — `layout.md:48` y los demás lugares que repiten el orden readback/borrado.
- **LO-2** — `local-record.md:6` repite `child-brief.md:7`.
- **LO-3** — `local-write.md:5-6` repite exclusiones de `local-record.md:5` y el secreto de `global.md`.
- **LO-4** — La autorización de escribir `HIVE.local.md` repetida en `hive-mode/SKILL.md:12`, `ticket-status.md:9`, `dependency-security.md:11`, `harness-audit/SKILL.md:12`, `harness-audit/references/local-instructions.md:16,37,40`.
- **LO-6** — `data-change.md:11` repite media regla de `test-accounts.md:8`.
- **CL-3** — Mover material a la checkout principal antes de quitar un worktree, en `close-hygiene.md:11`, `support-material.md:11`, `branch-cleanup.md:9`. Casa: `support-material.md`.
- **CL-4** — Capturas reemplazadas en `support-material.md:7` (casa: `layout.md`, supuesto).
- **CL-5** — Retiro de registros de memoria en `close-hygiene.md:19` y `batch-archive.md:7`.
- **CL-7** — `flow-close/SKILL.md:10,12` «read X first» de hojas que su paquete ya carga.
- **CL-10** — `close-report.md:18` / `support-material.md:11`; `close-modes.md:15` repite `viewer-close.md:7`.
- **CL-11** — `tasks.py checkpoint --out` del cierre: mover a `checkpoint.md`.
- **TR-6** — Regla de tickets de tarea operativa en cuatro redacciones (casa: `global.md`).
- **TR-7** — `flow-deploy/SKILL.md:34-35` repite `Completion` y el cierre de despliegue.
- **TR-8** — `hive-mode/SKILL.md:22` y `ticket-status.md:12` (tope de 50 tickets).
- **BU-2** — `business-record.md:12` repite «Decide and list» sin su cláusula de consecuencia.
- **MS-4** — Pregunta abierta en `questions.md:18` y `hive-mode/SKILL.md:19`: disparador en hive-mode, formato solo en la hoja.
- **MS-5** — «Solo la respuesta del usuario arranca el siguiente flujo» en `hive-mode/SKILL.md:28`, `flow-plan/SKILL.md:17`, `plan-handoff.md:21`.
- **PL-7** — `investigate.md:8-10,17` repite hive-mode paso 4 y el cambio de playbook.
- **PL-8** — `flow-plan/SKILL.md:16` vs `plan-review.md:7`: alinear la redacción.
- **DL-4** — `delegation.md:18` repite `external-facts.md:11`; `child-brief.md:11` repite git-workflow paso 3.

### Lote 2b — Paquetes (T5)

- **PK-1 (B-4)** — `flow-verify` precarga `project-verification.md` (3,5 KB) siempre; su paso 10 lo lee sin condición. Condicionar la línea (E2E o comandos de verificación cambiados), sacarlo de `stage-entry` y dejar la repetición de `suppression-review` como puntero.
- **PK-2 (B-10)** — `flow-build` precarga `stored-data.md` y `security-boundaries.md`; ya los enlaza con condición en `SKILL.md:28`. Sacarlos del paquete.
- **PK-3 (BU-4)** — El paquete de `flow-close` trae `business-rules.md`; el cierre necesita `business-ticket.md`. Agregarlo o cambiarlo.
- **PK-4 (DL-2)** — `reviewer-brief.md:11,13` mezcla tres cosas y se carga en cada verify. Queda en la hoja: cómo llegar a los estados (cargando, vacío, error, valores más largos) y los estados de `criteria.md`, que toda revisión UX de verify necesita (spec «UX review in one pass»). Sale a una referencia nueva de create-design que leen `example-record.md:23` y `rule-cards.md:29`: el brief de marcos dibujados. Sale a `create-design/references/review.md`: solo la línea 13 (diseño de otra persona). La hoja deja un puntero a cada destino.

### Lote 3 — Alcance: punteros donde la regla aplica (T6)

- **G-R1** U — `flow-verify` y `flow-deploy` hacen push y abren PRs sin enlazar `git-workflow`: el cuerpo de la PR, la lectura de la automatización y la pregunta de auto-merge en `ready` no les llegan. Su paso de push enlaza `git-workflow` desde el paso 6. Resuelve también **G-C10** (`pr-reviews.md:7` depende del paso 6).
- **B-5** M — `suppression-review.md`: agregar la segunda frase para el momento de agregar una supresión.
- **B-6 (D7)** X — Lista de ocho temas en la primera oración de `security-boundaries.md`, y por eso también en su línea del índice de `global.md:96` (`index_trigger_test`); `flow-build/SKILL.md:28`, el paso 1 de `hive-review-security.md`, `hive-mode/playbooks/infra-ci.md` y el enlace nuevo de B-10 apuntan a la hoja en vez de repetir temas.
- **B-8** U — `defect-fix.md` no llega a los hijos de build: puntero condicional en `child-brief.md` «Build children».
- **B-9** U — `background-wait.md` en `flow-deploy` paso 5; `project-verification.md` enlaza `project-verification-format.md`.
- **B-10** U — `flow-plan` paso 2 enlaza con condición `stored-data.md` y `security-boundaries.md`.
- **RV-3** U — `pr-review-deploy.md` agrega `review-findings.md` a su línea de hojas y a la revisión.
- **UI-5** M — `flow-verify/references/host-session-identity.md` lo usa también flow-build vía `viewer.md`: mover a `hive-core/ui/viewer/host-session-identity.md`, referencia propia de `viewer.md` (una hoja directa en `hive-core/ui/` exigiría su línea del índice); actualizar el enlace de `site/src/content/docs/viewer.md:68`.
- **UI-6** M/U — Ampliar el disparador de «Browser tool» en `ui-verification.md`; punteros en `deployment-checks.md:32` y `child-brief.md:7` (herramienta o serial del dispositivo).
- **UI-7** U — `a11y-audit` en la fila «pantalla agregada o cambiada» de `verification-floor.md`.
- **UI-10** M — Encabezado de `ui-plan.md` ampliado para quien lee la sección UI.
- **LO-5** U — `local-script.md` en `flow-build/SKILL.md:24`.
- **TR-1** U — Actualizar el worklog al cerrar: paso de `flow-close` enlaza `time.md` (cumple la spec de tiempo).
- **TR-9** U — `hive-read-state` enlaza `trackers.md`.
- **BU-3** U — `child-brief.md` «Build children»: el hijo reporta reglas supuestas, como pide `business-rules.md:17`.
- **BU-5** D — `deliverable.md:15` registra decisiones de un PRD «as business rules sets» sin lugar definido: apuntar al propio PRD, que `create-prd` ya escribe.
- **MS-6** U — `create-issue/SKILL.md:9` y `create-prd/SKILL.md:22` enlazan el bloque de respuestas o el de siguiente paso.
- **PL-3** U — Registro `Playbook`: `checkpoint.md` lo nombra y `flow-build`/`flow-verify` lo leen al entrar (supuesto).
- **DL-3a–d** U — Enlaces condicionales: `delegation.md#when-children-run-at-the-same-time` en `flow-plan` y `flow-verify`; `scope-changes.md` en `flow-plan`, `flow-build`, `flow-verify`; `project-settings.md` en `flow-plan/SKILL.md:12` y `ticket-status.md:9`; `reasons-and-blockers.md` en `flow-build/SKILL.md:26` y `flow-verify`.

### Descartados al verificar

- **CL-6** — El párrafo «Reporting» de los playbooks es intencional: `tests/content/leaf_skill_test.go:394` lo exige, y la cláusula «deployment row» solo para playbooks que pasan por flow-deploy.
- **UI-5 (cita)** — `host-session-identity.md:5` nombra el repositorio de Hive como procedencia; no es un enlace roto. Solo queda la ubicación.
- **RV-4, B-12** — Notas sin acción.
- Plegar `close-cleanup.md` en `close-hygiene.md` — inferido, sin evidencia.

## Impact

Detectores de `tests/pilot`: la revisión del plan buscó las rutas y frases que se mueven y no aparecen en ellos. `content/guidance/global.md` cambia en T3 (índice, si cambia una primera oración) y en T6 (D7): el build registra su tamaño antes y después en el progreso, para la PR que siga si el usuario la pide.

Solo guía distribuida (`content/`), `AGENTS.md`, specs, pruebas de contenido y documentación; nada de `tooling/` salvo que una prueba de `tooling/management` dependa de un paquete o una hoja movida. Las sesiones abiertas cargan la guía nueva tras reiniciar, después de `hive update`. Cada edición que toque una línea con «never» que no sea invariante la reescribe con alcance y razón (D5).

## Recovery

Bajo `hold` no hay commit: deshacer es `git restore` de los archivos modificados del lote, borrar los archivos nuevos (`tests/content/cited_targets_test.go`, la referencia nueva de create-design y cualquier referencia creada por el supuesto de tope) y devolver `host-session-identity.md` a `flow-verify/references/` con `git mv`; nada sale del repositorio. Si luego se entrega con un commit por lote, `git revert` de uno deshace ese lote, y `hive update` vuelve a desplegar la guía anterior.
