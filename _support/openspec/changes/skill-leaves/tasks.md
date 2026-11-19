# Tasks

El plan parte de `development` en `2bb96478`, con el trial en `6adebe35`. Las ramas son tres:
- **PR #491.** La integración del trial (T1).
- **`feat/skill-leaves`.** El PR 1, con T2 a T10. Se crea desde `development` después de integrar T1.
- **`feat/skill-leaves-principles`.** El PR 2, con T11, T15 y T12.

Cada rama va en su propio worktree del host.

Las tareas de escritura corren de a una, en el orden T5 → T6 → T7 → T8 → T13 → T14, y en el PR 2 T11 → T15, y cada una edita solo los archivos que lista.

Cada tarea escribe sus propias líneas en la sección «Rule leaves» de `global.md`. Los demás archivos que una tarea anterior ya tocó, los edita solo en las líneas que nombra.

El hilo principal fija el contrato: la lista de hojas, el formato del índice y los tests (T3 y T4). Después, un hijo escritor por tarea extrae su grupo, y el hilo principal integra y corre los tests después de cada una.

### T1 — El trial #491 se cierra con la evidencia medida y se integra

- [x] `question-card-alone` registra la medición del 2026-11-18 y un criterio de muestra nuevo que reemplaza el anterior. El PR #491, actualizado con `development`, está integrado. El párrafo del trial ya no está en `AGENTS.md`, `refs/hive/trial-join` está borrada y la instalación local está refrescada desde `development`.

**Closes:** AC1

**Depends on:** none

**Locations:** la rama `feat/question-card-alone` (en su worktree): `_support/openspec/changes/question-card-alone/` y `AGENTS.md`; GitHub PR #491

**Execution:** main thread

**Changes:**
- **En la rama del trial:**
  - Se integra `development`, que va 23 commits adelante.
  - Un criterio nuevo de esa carpeta reemplaza su criterio de muestra, con su marcador de reemplazo y la muestra medida: Claude 32/20, Codex 5 tarjetas y 7 rondas en texto, Grok 5/10, 1 tarjeta sin contestar, 0 rondas perdidas.
  - La tarea de medición de esa carpeta se marca como hecha, con su informe.
  - Se quita de `AGENTS.md` el párrafo «Trial branch (temporary)» y sus cuatro pasos, dentro del mismo PR #491, como pide ese párrafo.
- **Integración.** Se espera el CI y se integra el PR por D4 B.
- **Después del merge:**
  - `git update-ref -d refs/hive/trial-join`.
  - Se refresca desde `development` sin unión: `hive update --source … --out FILE` y `hive apply --plan FILE`.
  - Se cierra `question-card-alone` con `flow-close`.

**Verification:**
- `gh pr view 491 --json state` devuelve `MERGED`.
- `rg -n "Trial branch" AGENTS.md` en `development` no devuelve nada.
- `git show-ref refs/hive/trial-join` sale con código 1.
- `check_change.py` sobre `question-card-alone` sale con 0.
- La guía instalada de los 7 hosts contiene la regla del #491.

### T2 — El manager genera el `agents/openai.yaml` de Codex para las hojas ocultas

- [x] Una skill con `disable-model-invocation: true` recibe un `agents/openai.yaml` generado, con `policy.allow_implicit_invocation: false`, registrado como destino administrado. Las demás skills no cambian.

**Closes:** AC2

**Depends on:** none

**Locations:** `tooling/management/plan.go` (`loadRelease`, `validSource`, el plan de destinos), `tooling/management/generated_hooks.go` (el mecanismo de fuente generada que se reutiliza), sus tests

**Execution:** child `hive-build-backend`

**Test approach:** tdd

**Changes:**
- Se reutiliza el mecanismo `generatedSource` y `validRecordedHook`.
- Un `agents/openai.yaml` escrito a mano se rechaza explícitamente, también desde un checkout de fuente, donde hoy se omite en silencio.
- No cambia `SKILL.md`, ni el enlace de catálogo de ningún host, Antigravity incluido.

**Verification:** `go test ./tooling/... ./integrations/...` y `go vet ./...` pasan, con casos para:
- Una skill oculta.
- Una skill normal.
- Un archivo escrito a mano, desde una release guardada y desde un checkout.
- Una segunda actualización, que no escribe nada.
- Una skill que pierde la marca, cuyo yaml se quita.
- Codex desregistrado.
- Un yaml borrado, que se reinstala.
- La deriva detectada por `hive doctor`.

### T3 — Los tests fijan los invariantes, el hogar único y los enlaces requeridos

- [x] `tests/content` tiene el test de hojas, el de frases de un solo hogar y el de enlaces requeridos.

**Closes:** AC3, AC4, AC7

**Depends on:** none

**Locations:** `tests/content/leaf_skill_test.go`, `single_home_test.go` y `leaf_links_test.go` (nuevos); `design.md` (frases y enlaces)

**Execution:** main thread

**Test approach:** tdd

**Changes:**
- **Qué cuenta como hoja.** Las carpetas `hive-rule-*`, `hive-principle-*` (salvo `hive-principle-index`) y `hive-typescript`.
- **Invariantes:**
  - `name` igual a la carpeta y `disable-model-invocation: true`. `hive-typescript` queda exenta (D13 A): visible, con `paths` y sin yaml de Codex.
  - Descripción de una línea de 200 caracteres o menos.
  - Tamaño de `SKILL.md` según D11.
  - Una entrada exacta en su índice, y toda entrada resuelve.
- **Frases de un solo hogar.** Se fijan antes de T5, al menos tres de ellas repetidas hoy, con su cuenta anotada en `design.md`.
- **Enlaces requeridos.** Son la columna «Enlazada desde» de `design.md`.
- **Qué queda en suspenso.** Cada frase, cada enlace y cada entrada del índice, hasta que su tarea de extracción los activa.
- **Excepciones de tamaño.** Una lista con nombre, de hasta 7,000 bytes cada una (D11).
- **Playbooks.** Un tope de 7,000 bytes.

**Verification:**
- `go test ./tests/content/...` pasa.
- Falla, nombrando el archivo, en cada uno de estos casos sembrados:
  - Una hoja sin la marca.
  - Una hoja duplicada en el índice.
  - Una frase activa repetida en dos archivos.
  - Un enlace activo ausente.

### T4 — El índice, la hoja canario y HA-SK-04

- [x] `global.md` tiene la sección «Rule leaves» con su formato y una sola hoja, `hive-rule-git-commit`, extraída de `git-workflow` y enlazada desde el paso de commit de `flow-build`. HA-SK-04 acepta las hojas de Hive.

**Closes:** AC9, AC13

**Depends on:** T2, T3

**Locations:** `content/guidance/global.md`, `content/skills/hive-rule-git-commit/`, `content/skills/git-workflow/SKILL.md`, `content/skills/flow-build/SKILL.md` (solo el enlace), `content/skills/harness-audit/references/skills-and-agents.md`

**Execution:** main thread

**Changes:**
- **Formato de cada línea del índice:** «- **<condición>:** lee el archivo [<hoja>](skill:<hoja>/SKILL.md) y síguelo; no la invoques como skill.»
- **HA-SK-04.** Exceptúa `disable-model-invocation` y el yaml de Codex en las hojas que Hive distribuye. Los demás campos exclusivos de un host siguen marcados.
- **Hoja canario.** Se instala desde el commit del worktree de `feat/skill-leaves` (D18 A), con `hive update --source <worktree> --rev <commit> --out FILE` y `hive apply --plan FILE`, y se observa (D12 A). El commit es local, en la rama de trabajo, y no se publica. Hasta el PR 1, la instalación local va adelante de `development`. Si una sesión no la lee como archivo, se revisa el plan antes de seguir: cambiaría el mecanismo de carga.

**Verification:**
- `go test ./...` pasa.
- La instalación aplicada muestra la hoja y su yaml.
- En las sesiones reales que hagan commit, el registro muestra la lectura del archivo de la hoja antes del primer `git add`, en Claude Code y en Codex.

### T5 — Se extraen las hojas de `global.md`

- [x] Existen `decision-format`, `question-channel`, `report-endings`, `delivery-ask`, `foreign-commits`, `local-record`, `test-accounts`, `ticket-state`, `scope-changes` y `project-settings`, partidas según D11. `global.md` conserva los invariantes de `design.md`.

**Closes:** AC4, AC5, AC7

**Depends on:** T1, T4

**Locations:**
- `content/guidance/global.md` y sus `content/skills/hive-rule-*/`.
- Las copias de esas reglas en otros archivos, que esta tarea es la única en editar: `git-workflow` 13/14/26, `flow-close` 10/18, `flow-build` 14 y `urgent-fixes` 7.
- Las citas en prosa de las secciones vaciadas, en `content/` y `site/`, roles incluidos.
- Las rutas afectadas en `reading_footprint_test.go` y `tests/pilot/`.

`flow-deploy` no se edita aquí: sus copias de estas reglas pasan a T8.

**Execution:** child `hive-write-docs`, brief con el contrato de T3/T4 y la tabla de `design.md`

**Changes:** se mueve el texto sin cambiar su significado. Se fusiona cada copia con su hoja y la copia se quita. Las autorizaciones de las líneas 20, 21 y 32 y la línea de commit de «Secrets» se quedan enteras.

**Verification:**
- `wc -c content/guidance/global.md` da 30,000 bytes o menos.
- `rg -n "«(Questions and routes|Local instructions|Delivery|Test accounts|Project settings|Scope changes)»" content site` solo devuelve citas que siguen siendo verdad.
- `go test ./...` pasa, con las frases y los enlaces de estas hojas activos.

### T6 — Se extraen las hojas de `git-workflow` y `flow-close`

- [x] Existen `merge-gate`, `promotion-path`, `close-modes`, `close-hygiene` y `close-report`, y `flow-close` abre con `hive-rule-close-modes`.

**Closes:** AC4, AC6, AC7

**Depends on:** T5

**Locations:**
- `content/skills/git-workflow/` y `content/skills/flow-close/`.
- `flow-verify/SKILL.md` línea 21, que pasa a `merge-gate`. Esta tarea es la única en editar `flow-verify` antes de T7.
- Las rutas afectadas en los tests.

**Execution:** child `hive-write-docs`

**Changes:** la tabla de modos de cierre reemplaza las líneas 12, 14, 23, 25, 27, 29 y 70. El orden de pasos queda en `flow-close`.

**Verification:** `wc -c` de `git-workflow` y `flow-close` da 9,000 bytes o menos cada uno. `go test ./...` pasa.

### T7 — Se extraen las hojas de `flow-plan`, `flow-build` y `flow-verify`

- [x] Existen las hojas de la tabla de `design.md` cuyas fuentes son plan, build, verify y sus referencias, y cada flujo las enlaza en su paso.

**Closes:** AC4, AC6, AC7

**Depends on:** T6

**Locations:**
- **Skills y sus referencias:** `content/skills/flow-plan/`, `flow-build/` y `flow-verify/`.
- **Consumidores de las referencias convertidas:**
  - Los roles `hive-build-*`, `hive-review-code`, `hive-review-security`, `hive-write-tests` y `hive-design-architecture`.
  - `create-prd` (`SKILL.md`, `technical-review.md`).
  - `project-layout/references/script-lifecycle.md`.
  - `harness-audit/references/instruction-files.md`.
  - `flow-research/references/{security-advisory,status,code-audit}.md`.
  - `unattended-delegation` y `test-audit` (el ancla `#tests`).
  - `site/src/content/docs/flows/{build,verify}.md`.
- **Tests:** las rutas afectadas.

`flow-deploy` y `flow-figma-design` no se editan aquí: sus enlaces a estas hojas pasan a T8.

**Execution:** child `hive-write-docs`

**Changes:** las referencias convertidas en hoja dejan de existir, y sus consumidores pasan a la hoja. `plan-handoff.md`, `execution-record.md` y `e2e-coordination.md` se quedan como referencias, partidas por responsabilidad si pasan del tope.

**Verification:**
- `wc -c` de los tres `SKILL.md` da 9,000 bytes o menos.
- `go test ./...` pasa.
- `rg -n "references/(defect-fix|delegation|business-rules|suppression-review|project-verification)\.md" content site` solo devuelve lo que T8 todavía tiene que cambiar.

### T8 — Se extraen las hojas de figma, deploy y unattended, y se cierran sus enlaces

- [x] Existen `design-stage` y `unattended-effects`. `flow-figma-design` pesa 9,000 bytes o menos. Deploy, figma y unattended enlazan las hojas comunes en lugar de repetirlas.

**Closes:** AC4, AC6, AC7, AC17

**Depends on:** T7

**Locations:**
- `content/skills/flow-figma-design/`, `flow-deploy/` (con `urgent-fixes.md` y `deployment-checks.md`) y `unattended-delegation/`.
- `flow-figma-design/references/rule-cards.md`.
- Las rutas afectadas en los tests.

**Execution:** child `hive-write-docs`

**Changes:** además de mover, agrega una frase a `urgent-fixes.md`. En `ready` la promoción espera tu merge y pide su propia autorización. En `hold` el arreglo verificado se reporta como bloqueo, con la pregunta de cambiar el modo. Es la única adición de significado de esta tarea, porque la matriz de modos (AC17) la pide.

**Verification:**
- `wc -c` de `flow-figma-design/SKILL.md` da 9,000 bytes o menos.
- La búsqueda de T7 no devuelve nada.
- `go test ./...` pasa.

### T13 — `flow-research` pasa a `hive-mode`, con sus 13 playbooks

- [x] Existe `content/skills/hive-mode/`: su `SKILL.md` (11,000 bytes o menos) y `playbooks/<tipo>.md` para los 13 tipos de `design.md`, de 7,000 bytes o menos cada uno. Ya no existe `flow-research`.

**Closes:** AC14, AC17

**Depends on:** T8

**Locations:**
- **La skill.** `content/skills/flow-research/` → `content/skills/hive-mode/`:
  - `status.md` y `security-advisory.md` pasan a playbooks.
  - `code-audit.md` se queda como referencia de `hive-mode`, por debajo de 16 KB, y la enlaza el playbook `investigate`.
  - Se quitan de su paso 5 las copias de `report-endings` y `design-stage`, y en su lugar se enlazan esas hojas, cuyas frases y enlaces se activan aquí.
- **Consumidores de `flow-research` y de `/flow-research`.** En `content/`, `site/`, `README.md` y `llms.txt`. También en `_support/docs/architecture/agent-delivery.md` e `instruction-resources.md`.
- **Tests:**
  - `tests/pilot/`: `flows.go`, `flows_test.go`, `protected.go`, `e2e_case.go`, `guidance_variant.go`, `session_audit.go`, `qualify_score_test.go`, `qualify_suite_test.go` y `shell_reads_test.go`.
  - `tests/fixtures/flows/cases.json` y su `README.md`.
  - `content/skills/flow-plan/assets/change-folder/proposal.md` (campo `Playbook`).
  - `tests/content/reading_footprint_test.go` y `leaf_skill_test.go`, que también rechaza los nombres de modo entre comillas invertidas en los playbooks, salvo en la frase fija «under the delivery mode in force» (AC17). Lleva un caso sembrado: un playbook que nombra un modo fuera de esa frase, y el test falla.
- **Bloques MODIFIED, solo con el cambio de nombre:**
  - `workflow-lifecycle` «Tickets are created through one procedure».
  - `workflow-lifecycle` «Requirement documents are written and validated through one procedure».
  - `documentation-site` «Published documentation site»: la página `flows/research/` pasa a `flows/hive-mode/`.
- **Tu línea en «Rule leaves».**

`tests/fixtures/qualification/suite.json` se queda como histórico, salvo que un test exija lo contrario. `tooling/legacy/catalog.json` también es histórico.

**Execution:** child `hive-write-docs` para los playbooks; main thread para el cambio de nombre y los tests

**Changes:**
- **Lo que no cambia.** El procedimiento de research: leer, devolver el pedido, preguntar y cerrar el qué.
- **Lo nuevo.** La clasificación por tipo y la apertura del playbook.
- **Elección del playbook.** Se abre a lo sumo uno. Si ninguno encaja, se sigue la ruta general. Cambiar de playbook es un traspaso explícito, tratado como cambio de alcance.
- **Registro.** El playbook elegido queda en el campo `Playbook` de `proposal.md` cuando hay plan (la plantilla gana el campo), o en una línea `- Playbook:` del archivo de checkpoint, a mano, cuando no lo hay.
- **`hotfix`** enlaza `urgent-fixes.md` en vez de repetirlo.
- **Cada playbook nombra:**
  - El orden de las etapas.
  - Lo propio de su tipo: por ejemplo, el repro que falla y luego pasa en `bugfix`.
  - Sus hojas y principios.
- **`refactor` y `perf`** se redactan con nuestras palabras, a partir de los playbooks de pstack.

**Verification:**
- `rg -n "flow-research" content site README.md llms.txt tests` solo devuelve el registro histórico.
- `go test ./...` pasa, con el test de los 13 playbooks.
- `wc -c` de cada playbook da 7,000 bytes o menos.

### T14 — `flow-verify` suma los complementos de pstack

- [ ] Existen `hive-rule-no-comments`, `hive-rule-blast-radius`, `hive-rule-prose` y `hive-rule-interrogate`. `flow-verify` lee las tres primeras en el paso previo a la revisión. La hoja de la pregunta de entrega ofrece «Adversarial: interrogate» y dice que gasta cuota de otro proveedor.

**Closes:** AC15, AC17

**Depends on:** T13

**Locations:** `content/skills/hive-rule-{no-comments,blast-radius,prose,interrogate}/`, `content/skills/flow-verify/SKILL.md`, `content/skills/hive-rule-delivery-ask/SKILL.md` (las opciones de revisión), `content/skills/hive-rule-git-commit/SKILL.md` (enlace a `prose` para el mensaje de commit), las líneas de estas hojas en «Rule leaves», y `references/notice.md` de cada hoja adaptada.

`interrogate` sigue a D17 A:
- Se pregunta en cada trabajo y no se reutiliza.
- Corre como máximo una vez por candidato y otra tras la ronda de correcciones.
- Corre además de la revisión elegida.
- En trabajo desatendido, solo con un monto autorizado antes. Nombra la capacidad, el agente de línea de comandos de otro proveedor (Codex desde Claude Code, Claude desde Codex), y su respaldo: si no está instalado o no tiene sesión, lo informa como no disponible y corre solo la revisión elegida.

**Execution:** child `hive-write-docs`

**Changes:**
- Las hojas se redactan con nuestras palabras.
- `interrogate` lanza revisores de otro proveedor y reejecuta cada hallazgo. Corre solo cuando la respuesta de revisión lo elige.

**Verification:**
- `go test ./...` pasa.
- `flow-verify/SKILL.md` sigue pesando 9,000 bytes o menos.
- En la verificación de este mismo PR 1 se leen las tres hojas antes de `/code-review`.

### T9 — Auditoría, footprint, documentación y sitio siguen a las hojas

- [ ] `--audit-session` tiene un criterio de lectura de la hoja antes de la acción. El footprint mide las hojas, y la documentación y el sitio las describen.

**Closes:** AC8, AC10, AC12

**Depends on:** T14

**Locations:** `tests/pilot/session_audit.go` y sus fixtures, `tests/content/reading_footprint_test.go`, `_support/docs/harness-engineering/rule-enforcers.md` y `session-review-rubric.md`, `_support/docs/architecture/instruction-resources.md`, `README.md`, `llms.txt`, `site/src/content/docs/skills.md` y las páginas de flujo restantes

**Execution:** main thread para los tests; child `hive-write-docs` para la documentación y el sitio

**Changes:**
- **Criterio nuevo de auditoría, `leaf_read_before_action`.** Antes del primer `git add` o commit busca la lectura de `hive-rule-git-commit`. Antes de mover un ticket, la de `hive-rule-ticket-state`. Antes de preguntar una ronda, la de `hive-rule-question-channel`.
- **Footprint.** El escenario plan/build incluye cada hoja que sus condiciones seleccionan, entre ellas `delivery-ask`, `decision-format`, `question-channel` y `report-endings`.
- **`instruction-resources.md`.** Deja de decir que no hay skill de núcleo compartida.

**Verification:**
- `go test ./...`, `go vet ./...` y `npm --prefix site run build` terminan con código 0.
- El criterio nuevo pasa y falla en dos fixtures.
- `go test ./tests/content -run Footprint -count=1 -v` reporta el plan/build con las hojas por debajo de la línea base.

### T10 — El PR 1 se verifica, se revisa, se valida, se entrega y se despliega

- [ ] El candidato pasa:
  - La suite completa.
  - Una instalación de prueba en los 7 hosts.
  - Una prueba de reversión: el contenido revertido se aplica con el binario actual y luego se recompila.
  - `/code-review`.
  - Tu validación.

  Después se publica, se integra y se refresca la instalación local.

**Closes:** AC2, AC5, AC6, AC10, AC14, AC15, AC17

**Depends on:** T9

**Locations:** la rama `feat/skill-leaves`; la instalación local

**Execution:** main thread

**Changes:** el PR lista los lugares revisados y el cambio de tamaño de `global.md`, como pide `AGENTS.md`.

**Verification:**
- `hive update --source <worktree> --dry-run` muestra las hojas y sus yaml.
- Tras aplicar, se leen `~/.agents/skills/hive-rule-question-channel/SKILL.md` y su `agents/openai.yaml`.
- La reversión de prueba termina sin error.
- `/code-review` se contrasta con la matriz de modos de entrega de `design.md` (AC17).
- El CI del PR queda en verde.

### T11 — Fase 2: principios, índice de principios y `hive-typescript`

- [ ] Existen `hive-principle-index`, las 8 hojas de principio nuevas y `hive-typescript`, y las tres fusiones están hechas. El informe final de build y la revisión de código nombran, en una línea, cada principio leído que cambió una decisión.

**Closes:** AC11

**Depends on:** T10 integrado

**Locations:** `content/skills/hive-principle-*/`, `content/skills/hive-typescript/`, las hojas `defect-fix`, `tests` y `reuse-before-writing`, `flow-build` (paso de código e informe final), los roles `hive-build-*` y `hive-review-code`

**Execution:** child `hive-write-docs`

**Changes:** las hojas se redactan con nuestras palabras. `references/notice.md` da el crédito a pstack y no se enlaza. `hive-typescript` sigue a D13.

**Verification:**
- `go test ./...` pasa.
- La instalación aplicada muestra las hojas.
- `rg` de las frases de los principios fundidos encuentra un solo hogar.

### T15 — Fase 2: registro de decisiones, casillas por unidad de entrega y lo que no está escrito

- [ ] Existen las tres piezas:
  - **Registro de decisiones.** El registro TSV del trabajo desatendido y de la coordinación E2E, con su formato en un solo hogar.
  - **Casillas por unidad de entrega.** La sección `## Delivery units` con la gramática de `design.md`, que `check_change.py` valida.
  - **Lo que no está escrito.** El campo opcional `--unwritten` del checkpoint, que se pide al pausar.

**Closes:** AC16, AC17

**Depends on:** T11

**Locations:**
- **Registro de decisiones:**
  - El formato vive en una sección de `content/skills/unattended-delegation/SKILL.md`, junto con la plantilla `unattended-delegation/assets/decision-log.tsv`.
  - `content/skills/flow-build/references/e2e-coordination.md` lo enlaza. Ahí el coordinador es el único que escribe.
  - Las líneas que hoy mandan decisiones y efectos al registro de la tarea: `unattended-delegation`, paso 5 (línea 16, con su prohibición de «no additional ledger») y paso 8 (línea 23), y la hoja `hive-rule-unattended-effects`, adonde T8 movió el paso 6. Se reemplazan por D19 A: decisiones y efectos del trabajo desatendido van al mismo `decisions.tsv`, y el registro de la tarea queda solo con el avance.
  - Los efectos del trabajo atendido siguen en el registro de la tarea (`hive-rule-external-effects`), sin cambios.
  - `e2e-coordination.md` (7.4 KB) no se parte, porque está por debajo del tope.
- **Casillas:** la sección `## Delivery units` en `content/skills/flow-plan/assets/change-folder/tasks.md`, `flow-plan/scripts/check_change.py` y la hoja `hive-rule-change-folder` (cuándo se usa la sección).
- **Pausa:**
  - `flow-plan/scripts/tasks.py`, con `--unwritten` opcional. Se conserva en las llamadas siguientes hasta que otro valor lo reemplaza, o `--unwritten none` lo borra.
  - La hoja `hive-rule-checkpoint`, que también se lee al pausar.
  - La hoja `hive-rule-close-modes`, donde la pausa pide el campo.
- **Tests:** `tests/skills/tasks_test.py` y `tests/skills/change_check_test.py`.
- **Sitio:** `site/src/content/docs/flows/plan.md` y las páginas del sitio sobre trabajo desatendido y checkpoint.
- **Créditos:** `unattended-delegation/references/notice.md` y `flow-plan/references/notice.md`, con el crédito a pstack (`show-me-your-work`, `multi-phase-plan`, `pause-safely`).

**Execution:** child `hive-build-backend` para los scripts; child `hive-write-docs` para el texto

**Test approach:** tdd para `check_change.py` y `tasks.py`

**Changes:**
- Una fila por decisión, no por acción. La evidencia es un enlace o una ruta, nunca un párrafo.
- El registro vive en `_support/workspace/YYYY-MM-DD-<trabajo>/decisions.tsv`. No entra en commits ni en el PR salvo que lo pidas, y el informe final lo cita por ruta.
- Se cumple la matriz de modos de entrega.

**Verification:**
- `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasan.
- `check_change.py` rechaza una casilla `live:` sin `pass:`, y `--archive` acepta `n/a:`.
- `tasks.py checkpoint --unwritten "…"` escribe el campo, una llamada posterior sin la opción lo conserva y `--unwritten none` lo borra.
- Cada pieza se comprueba contra la matriz de modos de `design.md`. Por ejemplo, en `hold` la pausa no hace commit, y en `direct-base` las casillas van por commit.

### T12 — El PR 2 se entrega y se observan sesiones reales

- [ ] Se entrega el PR 2, con `/code-review` y tu validación. Un informe de sesiones reales posteriores al despliegue del PR 1 cubre AC8.

**Closes:** AC8, AC11, AC14, AC16, AC17

**Depends on:** T11, T15

**Locations:** la rama `feat/skill-leaves-principles`; historiales de sesión, solo lectura

**Execution:** main thread; hijos `hive-research` para la auditoría

**Verification:**
- El CI está en verde.
- `--audit-session` sobre 10 sesiones reales o más da 0 fallas de `no_broad_git_add` y de `leaf_read_before_action`.
- Al menos una de esas sesiones entra por `/hive-mode`, abre un playbook y no edita archivos antes de entregar a `flow-build`.
- La revisión de código del PR 2 se contrasta con la matriz de modos (AC17).
- AC16: una corrida desatendida real tras el PR 2 deja su `decisions.tsv` con una fila por decisión y una por cada efecto externo, con cómo revertirlo. La mitad E2E se observa cuando corra una coordinación E2E real. AC16 queda sin verificar hasta entonces.
- AC8 queda sin verificar mientras la muestra no llegue a ese número.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar cada tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final de cada PR, local | `go test ./...`, `go vet ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`, `npm --prefix site run build` | Cubierto por la implementación |
| Instalación de prueba y reversión | Candidato final del PR 1, esta máquina | `hive update --dry-run`, los archivos instalados y la reversión de prueba | Marca de mantenedor |
| Revisión de código | Candidato final, antes de tu validación | Claude Code `/code-review` | D9 A |
| Tu validación | Antes de cada push | Tu respuesta | D8 A |
| Canario y sesiones reales | Tras T4 y tras el PR 1 | Lectura de la hoja en el registro de la sesión, e informe de T12 | Solo lectura |

Tu validación antes de cada push: lee `hive-rule-question-channel`, `hive-rule-ticket-state` y `hive-rule-close-modes` y compáralas con el texto que reemplazan. El diff de `global.md` y de `flow-close` está en el PR.

## Checkpoint

- Updated: 2026-11-19
- Stage: build · T8 next
- Last commit: f5f68b6c
- Open findings: global.md 30,957 bytes > AC5 30,000 after 55 index lines; trim conditions or ask the user to revise AC5 before PR 1
- Next: T8 in child
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/skill-leaves · feat/skill-leaves

### Evidence

- go test ./tests/content (leaf, home, links) · pass · e9561ec4
- trial branch: go test ./... · 1 flaky pilot failure (TestE2ESetupDeadlinesOwnDescendantsAndPipes/github-token), passes 3/3 alone on trial and base · 4ba449e5
- before captures: none (no UI)
- go test ./tests/content and tooling/management release checks · pass · c33681da
- hive rebuilt from worktree (backup: scratchpad/hive.backup-before-canary) and applied c33681da · leaf, agents/openai.yaml and Claude link installed · c33681da
- PR #491 merged addf3256; trial ref, branch and worktree removed; question-card-alone archived 0a1aab75
- canary Codex cam 01a1f20a · pass · c33681da
- canary Claude Code cam a9f82c64 · pass: cat of leaf 05:06:19Z before first commit 05:11:39Z via git_state.py, no broad git add · c33681da

## Review status and progress

2026-11-18 — Plan escrito.
2026-11-18 — Revisión del plan (`hive-review-plan`): 6 P1, 11 P2 y 7 P3.
- **Corregidos en la carpeta:**
  - Invariantes de autorización que se quedan en global (H2).
  - Archivos de un solo dueño y tareas en secuencia (H3).
  - Antigravity vuelve a enlazar las hojas (H4 y H7).
  - Citas en prosa (H5) y rutas de los tests en cada tarea (H6).
  - Colocación básica de artefactos (H8).
  - Ciclo de vida del yaml (H9) y orden de reversión (H10).
  - Lectura como archivo y canario (H11).
  - Frases y enlaces fijos (H13 y H23).
  - Criterio de auditoría (H14), efectos y rama del trial (H15).
  - Footprint con hojas (H16), consumidores (H17).
  - T1 dentro del #491 (H18), rama actual (H19), hechos por host (H20).
  - Comandos visibles como supuesto (H21) e índice de principios excluido (H22).
- **Van al usuario:** tamaño de las hojas (H1), canario (H11) y `paths` en `hive-typescript` (H12).

2026-11-18 — Revisión por las decisiones D11 a D15. Se agregan AC14 (`hive-mode` y sus 13 playbooks) y AC15 (complementos de verify), y las tareas T13 y T14. T8 deja `flow-research` a T13, y T9 pasa a depender de T14. No invalida tareas hechas, porque ninguna está hecha.
2026-11-18 — Segunda revisión del plan: 4 P1, 6 P2 y 3 P3.
- **Corregidos:**
  - Líneas del índice por tarea (H24).
  - Copias de research quitadas en T13 (H25).
  - `hive-typescript` exenta de la marca (H26).
  - Consumidores y bloques MODIFIED del cambio de nombre (H29 y H30).
  - Hogar de `code-audit.md` (H31) y enlace de `hive-typescript` desde el índice de principios (H32).
  - Sesión de AC14, tope con excepciones y respaldo de interrogate (H33).
  - Orden de tareas y checkpoint (H34), comandos del trial (H35), etapa de diseño y prosa en el commit (H36).
- **Van al usuario:** el límite de interrogate (H27, D17) y la ruta del canario (H28, D18).

2026-11-18 — Revisión: tarea añadida, AC16 y T15 en la fase 2 (registro TSV de decisiones, casillas por PR y lo que no está escrito al pausar). T12 pasa a depender de T15. No invalida tareas.
2026-11-18 — Revisión: compatibilidad con los modos de entrega. Se agregan una matriz en `design.md` y AC17, cerrado por T13, T14 y T15, con su test. AC16 agrupa las casillas por unidad de entrega.
2026-11-18 — Tercera revisión del plan (AC16 y T15): 3 P1, 6 P2 y 4 P3. Corregidos: observación de AC16 en T12 (H38); pausa con `--unwritten` opcional y lectura del checkpoint al pausar (H39); deltas de spec (H40); gramática de casillas y rechazo que hoy es falso (H41); condición de uso (H42); ruta y escritor del registro (H43); playbook a lo sumo uno, con ruta general (H44); registro del playbook elegido (H45); listas y comandos (H46 a H48); resultados fuera del texto de la casilla (H49). Va al usuario: si los efectos comparten el registro (H37, D19).
2026-11-18 — D19 A: decisiones y efectos externos comparten `decisions.tsv`. Se agrega al delta de `workflow-lifecycle` el requisito «Decision log for unattended and coordinated work».
2026-11-18 — Cuarta revisión (matriz, AC17 y D19): 1 P1, 6 P2 y 3 P3, todos aplicados. El merge de un PR ajeno pide tu instrucción en todos los modos (H50). Pausa sin commit `wip:` (H51). Test de AC17 con comillas invertidas y frase fija (H52). `ready` y `hold` en `urgent-fixes.md`, por T8 (H53). Un solo hogar para el playbook elegido, sin `--playbook` (H54). Filas de efectos en AC16 y T12 (H55). Ubicaciones del registro (H56). Matriz en T10 y T12 (H57). Gramática y encabezado `## Delivery units` (H58 y H59).
2026-11-18 — T3 · e9561ec4 · unverified: AC4 and AC7 stay inactive until each extraction task activates its phrases and links
2026-11-18 — T1 · addf3256 · unverified: AC4 of question-card-alone (rule survives a refresh) is moot: the trial paragraph is gone
2026-11-18 — T2 · 549b1583 · unverified: hive doctor end to end and a rollback to an older stored release not run; covered through Status and desiredResources
2026-11-18 — T4 · c33681da · unverified: none
2026-11-19 — T5 · dff232ad · unverified: flow-deploy citations and its ticket-state link wait for T8; README, llms.txt and _support/docs wait for T9
2026-11-19 — T6 · 9eda7fa9 · unverified: close-modes is not linked from git-workflow step 10 (would exceed 9,000 bytes); flow-close still reached there by name
2026-11-19 — T7 · f5f68b6c · unverified: global.md is 30,957 bytes, over AC5's 30,000 with more index lines to come; settled before PR 1
2026-11-19 — T8 · eb134916 · unverified: promotion commit listing stays in flow-deploy (different rule than foreign-commits); site pages for figma/deploy wait for T9
2026-11-19 — T13 · 8de4a4a5 · unverified: AC14's real hive-mode session waits for T12; tests/fixtures/qualification/suite.json kept as history
