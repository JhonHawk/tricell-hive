# Tasks

**Base:** `development` en `6fa78bfd` (2026-11-21), que ya contiene `reporting-family-leaves` (#563, sin su T6) y su ronda de hallazgos (#564); el plan se escribió sobre `7f6277e0`. Cada unidad se construye en su propio worktree bajo `.claude/worktrees/`, porque una sesión paralela trabaja en el checkout principal. T1 rehace sobre esa base el inventario y la medida de `global.md` (28,459 B en `6fa78bfd`; la huella usa `planReviewAllowance` 182 + `reportingFamilyAllowance` 9,230 B).

**Quién ejecuta:**
- **Hijos.** Las tareas que editan más de 3 archivos con interfaz fija van a hijos:
  - `hive-build-infra`: T1, T4 y T6;
  - `hive-write-tests`: T2 y T3;
  - `hive-write-docs`: T5 y T9.
- **Hilo principal.** Corre los checks, hace los commits y escribe T7, T8 y T10.
- **Verificación aparte.** T4 y T6 son arriesgadas porque cambian el contrato de instalación, así que cada una lleva un hijo `hive-verify-task`.

### T1 — Mudar las hojas a `hive-core`, reescribir enlaces y menciones, y alinear el índice

- [x] La mudanza queda hecha y alineada:
  - cada hoja es `content/skills/hive-core/references/<nombre>.md`, sin frontmatter, y su primera línea de cuerpo después del título empieza por «Read»;
  - no queda ningún directorio `hive-rule-*`, y `hive-core/SKILL.md` está oculto;
  - las referencias y los scripts propios de cada hoja están en `references/<nombre>/` y `scripts/<nombre>/`;
  - los enlaces y las menciones de `content/` usan las rutas y los nombres nuevos;
  - cada línea del índice cita literalmente la primera línea de su hoja;
  - `global.md` define la hoja como un archivo de `hive-core`.

**Closes:** AC11, AC3

**Depends on:** none (dentro de E1; E1 espera el orden interno de la propuesta)

**Locations:**
- `content/skills/hive-rule-*/`, que pasa a `content/skills/hive-core/`;
- `content/guidance/global.md`;
- `content/skills/*/SKILL.md`, sus `references/` y `playbooks/`;
- `content/agents/**`;
- `content/skills/harness-audit/references/{skills-and-agents,harness-audit-rules}.md`.

**Execution:** `hive-build-infra`

**Test approach:** characterization. Las pruebas actuales, adaptadas en T2, deben pasar sin cambiar lo que afirman. El disparador literal tiene su prueba en T2.

**Changes:**
- Se mueve con `git mv`.
- La reescritura sigue el mapa de `design.md` «Rewrite map»: todas las formas de enlace, las referencias y scripts anidados, los marcadores en bloques de código como `<hive-rule-checks-dir>`, y los nombres sueltos `hive-rule-<nombre>` en prosa.
- El mapa viejo→nuevo queda en el cuerpo del commit.
- Cada línea del índice se reescribe hasta citar la primera línea «Read» de su hoja; si eso no cabe en la línea, se ajusta la línea «Read» de la hoja. Así el índice queda alineado sin cambiar el momento que nombra.
- Una línea que agrupa varias hojas, o una hoja con dos momentos, separa sus disparadores con «; », y la prueba de T2 compara cada uno con su hoja.
- `global.md:61` pasa a decir que una hoja es un archivo de `hive-core`.
- No cambia el texto de ninguna regla más allá de su primera línea y sus enlaces.
- `global.md` se mide después: no puede pasar de 30,000 B, y su tamaño va al cuerpo del PR.

**Verification:**
- `go -C <worktree> test -count=1 ./tests/content/...` pasa tras T2.
- El comando de AC16, limitado a `<worktree>/content`, no imprime nada.
- `wc -c <worktree>/content/guidance/global.md` da ≤ 30,000.

### T2 — Adaptar las pruebas de contenido y Python, y añadir la del disparador literal

- [x] Las pruebas de `tests/content/` enumeran las hojas en `hive-core/references/` y pasan. Una prueba nueva compara cada línea del índice con la primera línea de su hoja. Las pruebas Python usan las rutas nuevas de los scripts. `rule-enforcers.md` y `single_home_test.go` apuntan a los hogares nuevos.

**Closes:** AC11, AC3

**Depends on:** T1

**Locations:**
- `tests/content/`: `leaf_skill_test.go`, `single_home_test.go`, `leaf_links_test.go`, `skill_size_test.go`, `reading_footprint_test.go`, `budget_test.go`, `rule_enforcers_test.go` y `skill_links_test.go`;
- `tests/skills/`: `run_check_test.py`, `jira_rest_test.py` y `work_time_test.py`;
- `_support/docs/harness-engineering/rule-enforcers.md`.

**Execution:** `hive-write-tests`

**Test approach:**
- tdd para la prueba del disparador: se la ve fallar con una línea del índice alterada;
- tdd para el caso «queda un directorio `hive-rule-*`»;
- characterization para el resto.

**Changes:**
- `leaf_skill_test` deja de pedir frontmatter por hoja. Conserva el tope de 4,000 B y 7,000 B y la cuenta del índice, y exige `disable-model-invocation` solo a `hive-core/SKILL.md` y a los principios.
- `reading_footprint` conserva su comparación contra `2bb96478`, con las rutas nuevas como mapa `moved`.

**Verification:**
- `go -C <worktree> test -count=1 ./tests/content/...` termina con exit 0.
- `python3 -I -m unittest discover -s <worktree>/tests/skills -p '*_test.py'` termina con exit 0.
- La prueba nueva falla al quitar una palabra de un disparador del índice y vuelve a pasar al restaurarla.

### T3 — La auditoría de sesiones reconoce la ruta vieja y la nueva, por lectura y por shell

- [x] Los detectores reconocen la ruta nueva:
  - `leaf_read_before_action`, `reportEndingsRead` y el conteo de costo reconocen `hive-core/references/<nombre>.md` además de `hive-rule-<nombre>/SKILL.md`, tanto con la herramienta de lectura como por shell;
  - `leafFormerNames` acepta los nombres viejos;
  - los fixtures viejos siguen pasando, y hay fixtures nuevos de lectura y de shell con la ruta nueva.

**Closes:** AC12

**Depends on:** T1

**Locations:**
- `tests/pilot/session_audit.go` (`leafPaths`, `shellReadsLeaf:1706`, `namesLeaf:1728`, `leafFormerNames`);
- `session_audit_endings.go:14`;
- `session_audit_cost.go:35`;
- `tests/fixtures/sessions/leaf_read_before_action/`.

**Execution:** `hive-write-tests`

**Test approach:** tdd

**Changes:**
- La comparación es por ruta completa, no por el sufijo `/SKILL.md`.
- Los nombres viejos se arman como `"hive-rule-" + nombre`, sin escribirlos literales, para que la búsqueda de AC16 no los encuentre.
- No se reescriben los fixtures grabados.
- Los paquetes se añaden en T6.

**Verification:**
- `go -C <worktree> test -count=1 ./tests/pilot/...` termina con exit 0.
- Los fixtures nuevos fallan antes del cambio.

### T4 — El manager instala `hive-core`, retira las carpetas de hoja y limpia su `__pycache__`

- [x] Una prueba del manager sobre el contenido real, con un estado previo que registra las carpetas `hive-rule-*`, comprueba que el manager:
  - instala `hive-core` con sus `references/<nombre>/` y `scripts/<nombre>/`;
  - genera un solo `openai.yaml` de regla;
  - resuelve cada enlace `skill:` del bloque global;
  - retira en una actualización las carpetas `hive-rule-*`, sus enlaces de catálogo y los `__pycache__` que solo contienen `.pyc`.

**Closes:** AC2

**Depends on:** T1

**Locations:**
- `tooling/management/apply.go` (`cleanupDirs`);
- `tooling/management/*_test.go`;
- comentarios en `tooling/cli/doctor_project.go:27,41` y `tooling/management/voice.go:220`;
- `tooling/cli/doctor_project_test.go:451`.

**Execution:** `hive-build-infra`. Al cerrarla, el hilo principal lanza un `hive-verify-task`.

**Test approach:** tdd (contrato persistido de la instalación)

**Changes:**
- `plan.go` y `references.go` no cambian, porque `references/` anidado ya está admitido.
- El estado previo de la prueba se construye instalando una versión con una carpeta de hoja y aplicando después la nueva.
- La limpieza actúa solo en carpetas de skill de Hive que se retiran, y solo después de que la retirada quede registrada:
  - borra archivos regulares `*.pyc`, comprobados con `Lstat` y `target.Safe`;
  - rechaza un `__pycache__` que sea enlace;
  - conserva un directorio con cualquier otro archivo y lo reporta en la salida de `apply`.
- Las pruebas cubren un `.pyc` borrado, un archivo ajeno conservado y reportado, y un enlace rechazado.

**Verification:**
- `go -C <worktree> test -count=1 ./tooling/... ./integrations/...` termina con exit 0.
- `go -C <worktree> build -o <scratchpad>/bin/hive-e1 ./tooling/cli && <scratchpad>/bin/hive-e1 update --source <worktree> --dry-run` (el binario se niega a correr desde el directorio temporal del sistema). Antes se nombran los commits que aún no están desplegados. El resultado retira las carpetas `hive-rule-*` y añade `hive-core`, y no cambia nada fuera de las rutas que toca E1, de los roles renderizados ni de los bloques globales de cada host, salvo los archivos auxiliares que el manager genera a partir de ellas, como el cargador de OpenCode.

### T5 — Documentación, sitio y deltas de E1

- [x] Toda la documentación y el sitio usan las rutas y los nombres nuevos. `instruction-resources.md` describe `hive-core` como almacén y corrige su nota de `5e6e8907` y la de HA-HI-09. Los deltas de `context-delivery` y el MODIFIED del helper de checks en `workflow-lifecycle` coinciden con el comportamiento.

**Closes:** AC16

**Depends on:** T1, T2, T3, T4

**Locations:**
- `site/src/content/docs/**`;
- `README.md`, `llms.txt`, `AGENTS.md` y `CHANGELOG.md`;
- `_support/docs/architecture/{instruction-resources,deployment-manager,compaction-survival,agent-delivery}.md`;
- `specs/context-delivery/spec.md` y `specs/workflow-lifecycle/spec.md` de esta carpeta.

**Execution:** `hive-write-docs`

**Changes:**
- Solo rutas, nombres y la descripción de la estructura.
- El cuerpo del PR lleva la lista de lugares revisados y el cambio de tamaño de `global.md`.

**Verification:** el comando de AC16 sobre `<worktree>` no imprime nada. `rg -n 'hive-rule-' <worktree>/_support/openspec/specs` muestra solo requisitos que un delta de esta carpeta modifica. Si `reporting-family-leaves` se archivó antes y su delta añadió nombres de hoja a `workflow-lifecycle`, T5 añade los deltas MODIFIED de esos requisitos.

### T6 — El manager genera los paquetes por etapa y la auditoría los reconoce

- [x] El manager genera un paquete por clave de cada `<flujo>/assets/stage.json`, como fija el delta de `versioned-installation`:
  - lo registra y lo instala;
  - lo genera con la concatenación de sus miembros bajo `## <nombre>`;
  - reescribe sus enlaces y valida el paquete generado;
  - rechaza un manifiesto malo, un miembro faltante y un paquete por encima de 24,000 B o 1,000 líneas;
  - lo retira, lo reescribe o lo reinstala igual que cualquier contenido gestionado;
  - rechaza un `stage*.md` escrito a mano.

  Un binario anterior a E2 que se usa después de retirar los paquetes lee el estado sin error. La auditoría cuenta un paquete por sus encabezados en `leaf_read_before_action`, en `report-endings` y en el costo.

**Closes:** AC5, AC12

**Depends on:** E1 integrado

**Locations:**
- en `tooling/management/`:
  - `stage_packets.go` (nuevo);
  - `generated_hooks.go` (`generatedSource`) y `ownership.go` (`desiredResources`, `normalizeState`);
  - `plan.go` (`nextRecord`, `validateRelease`, `Status`) y `hidden_skills.go` (`rejectHandWrittenPolicies`);
  - las pruebas en `stage_packets_test.go`;
- en `tests/pilot/`: `session_audit.go`, `session_audit_endings.go`, `session_audit_cost.go` y los fixtures de paquete;
- `specs/versioned-installation/spec.md` de esta carpeta.

**Execution:** `hive-build-infra`. Al cerrarla, el hilo principal lanza un `hive-verify-task`.

**Test approach:** tdd

**Changes:**
- Sigue el patrón de `agents/openai.yaml`.
- Los enlaces se reescriben como fija «Decisions» de la propuesta, incluidas las definiciones de estilo referencia.
- Cada paquete generado entra en la lista de archivos y anclas que usa `validateInstructionReferences` antes de validar sus enlaces.
- La auditoría une cada lectura con el texto de su resultado por `ID`.
- Cada paquete termina con `<!-- hive packet end: <clave> · <miembro>,<miembro>,… -->`.
- El nombre de cada miembro es su ruta del manifiesto sin `.md`.
- La auditoría quita antes el prefijo de número de línea `^\s*\d+(\t|→)`, la forma de Claude Code.
- No se midieron las formas de Grok, Pi, OpenCode y Cursor. Si una antepone otro prefijo, sus sesiones dan `not_observed` en vez de pasar. Lo revisa el seguimiento de las primeras sesiones tras el despliegue, que añade el patrón y su fixture. Después lee los miembros de la línea de fin y da el paquete por leído solo si el resultado trae esa línea y un encabezado exacto `## <miembro>` por cada uno.
- Si falta la línea o algún encabezado, o el resultado no tiene texto, reporta `not_observed`.
- El orden de recuperación de E2 queda en `design.md` «Recovery».
- No cambia ningún archivo instalado fuera de la carpeta del flujo.

**Verification:**
- `go -C <worktree> test -count=1 ./tooling/... ./integrations/... ./tests/pilot/...` termina con exit 0.
- Las pruebas nuevas cubren estos casos:
  - alta de uno y de dos paquetes;
  - actualización sin cambios, que no escribe;
  - retirada por clave, por manifiesto y por flujo;
  - reescritura cuando cambia un miembro, y reinstalación;
  - rechazo de un archivo escrito a mano, de una clave desconocida, de un miembro faltante y de un tope excedido;
  - un enlace roto en el paquete generado;
  - tras retirar los paquetes, cada origen registrado cumple el patrón de `validSource` anterior a E2 (`plan.go:363-369` en `7f6277e0`);
  - un enlace con ancla a otro miembro y una definición de estilo referencia;
  - fixtures de auditoría:
    - lectura de un paquete por shell, y por la herramienta Read de Claude Code tomada de una transcripción real, con números de línea; las dos pasan;
    - un paquete completo que además trae la sección propia `## Change folder`: pasa, y esa sección no cuenta como el miembro `hive-core/references/change-folder`;
    - dos que dan `not_observed`: el resultado cortado al final, y cortado por la mitad con la línea de fin intacta.

### T7 — Manifiestos, línea de entrada de cada flujo y huella

- [x] Los ocho flujos y la huella quedan así:
  - cada uno tiene `assets/stage.json` con la membresía de `design.md`, recalculada sobre la base de E2;
  - la suma de miembros de cada paquete da ≤ 23,000 B, y el paquete ya generado ≤ 24,000 B en la prueba del manager;
  - cada flujo dice al entrar que lea su paquete entero, que no relea sus hojas y que, si falta, lea las hojas enlazadas;
  - la definición de hoja de `global.md` admite el paquete como forma de leerla;
  - `reading_footprint_test` declara el escenario plan/build con paquetes, ≤ 26 lecturas, y reporta las cuatro cifras.

**Closes:** AC6, AC13

**Depends on:** T6

**Locations:**
- `content/skills/{hive-mode,flow-plan,flow-build,flow-verify,flow-close,git-workflow,flow-deploy,unattended-delegation}/SKILL.md` y su `assets/stage.json`;
- `content/guidance/global.md`;
- `tests/content/reading_footprint_test.go` y una prueba nueva `TestStageManifests` en `tests/content/`.

**Execution:** hilo principal

**Test approach:** check, extendiendo `reading_footprint_test.go`, que es la prueba dueña de lo que lee una ruta. La prueba de manifiestos comprueba que cada clave es válida, que cada miembro existe y la suma.

**Changes:**
- Las hojas fuera del paquete conservan su enlace y su momento.
- El paquete de `flow-plan` deja fuera `reuse-before-writing` y conserva `plan-handoff`.
- La línea de entrada añade: si la última línea del paquete no es su línea de fin, leer el resto.
- Antes de fijar cada manifiesto se recuenta la ruta, para que no pase de 26 lecturas.
- `global.md` crece solo en la frase del paquete y sigue ≤ 30,000 B.

**Verification:**
- `go -C <worktree> test -count=1 ./tests/content -run 'TestRequiredProcedureReadingFootprint|TestStageManifests' -v` muestra ≤ 26 lecturas y las cuatro cifras.
- `./tests/content/...` y `./tooling/management/...` completos terminan con exit 0.

### T8 — La línea `Rules:` del checkpoint y el orden tras compactar

- [ ] `tasks.py checkpoint --rules` escribe y conserva `- Rules:`, y `check_change.py` la acepta bien formada y la rechaza mal formada. La hoja del checkpoint y una línea de `global.md` dicen el orden tras compactar: checkpoint, `HIVE.local.md`, paquete, extras y memoria. Las líneas de `global.md` que hoy nombran pasos tras compactar (`:46`, `:90`, `:126` en `7f6277e0`) quedan reconciliadas con ese orden, sin otro orden en `content/`.

**Closes:** AC8

**Depends on:** E2 integrado

**Locations:**
- `content/skills/flow-plan/scripts/{tasks.py,check_change.py}` y `assets/change-folder/tasks.md`;
- `content/skills/hive-core/references/checkpoint.md`;
- `content/guidance/global.md`;
- `tests/skills/{tasks_test.py,change_check_test.py}`;
- `specs/workflow-lifecycle/spec.md` de esta carpeta.

**Execution:** hilo principal

**Test approach:** tdd (los scripts guardan el estado de reanudación)

**Changes:**
- El flag sigue el patrón de `--worktree`.
- La ruta es relativa al almacén de skills.
- La línea del orden en `global.md` y la hoja `checkpoint` dicen que, si el paquete falta, se leen las hojas que enlaza el flujo.

**Verification:**
- `python3 -I -m unittest discover -s <worktree>/tests/skills -p '*_test.py'` termina con exit 0, con casos nuevos de escritura, conservación y rechazo.
- `go -C <worktree> test -count=1 ./tests/content/...` termina con exit 0.
- `rg -n 'compact' <worktree>/content` no muestra ningún otro orden.

### T9 — La línea `Reply:` de cada flujo

- [ ] `hive-mode`, `flow-plan`, `flow-build`, `flow-verify`, `flow-close`, `flow-deploy` y `git-workflow` tienen una línea `Reply:`. Esa línea nombra la fila de reporte que termina el flujo y solo lo que el flujo añade y la fila no tiene. Una prueba de contenido la exige.

**Closes:** AC9

**Depends on:** E2 integrado

**Locations:** los 7 `SKILL.md`; una prueba nueva en `tests/content/`; `site/src/content/docs/flows/*.md` que describan cómo termina un flujo

**Execution:** `hive-write-docs`; la prueba la escribe el hilo principal

**Test approach:** check

**Changes:**
- La línea no repite la fila.
- Las filas viven en las hojas de reporte que deja `reporting-family-leaves`.

**Verification:**
- `go -C <worktree> test -count=1 ./tests/content/...` termina con exit 0.
- La prueba nueva falla si se quita la línea de un flujo.

### T10 — Documentación de paquetes, checkpoint y cierre

- [ ] Estos documentos describen los paquetes, la línea `Rules:`, el orden tras compactar y la línea `Reply:`, cada uno en el PR que los introduce:
  - `deployment-manager.md`, `instruction-resources.md` y `compaction-survival.md`;
  - el sitio y `CHANGELOG.md`;
  - `README.md` y `llms.txt`, si cambia algo de lo que declaran.

**Closes:** AC5, AC8, AC9

**Depends on:** T6 y T7 (parte de E2); T8 y T9 (parte de E3)

**Locations:** los archivos citados

**Execution:** hilo principal

**Changes:**
- Describe el comportamiento y enlaza la regla, sin copiarla.
- El cuerpo de cada PR lleva la lista de lugares revisados y el cambio de tamaño de `global.md`.

**Verification:**
- `go -C <worktree> test -count=1 ./tests/content/...` termina con exit 0.
- Al leer el diff de documentación no aparece ninguna regla copiada.

## Delivery units

### E1 — PR `refactor/hive-core-sheets` (T1–T5) · depends on: none · files: content/skills/hive-core/, content/, tests/content/, tests/pilot/, tests/skills/, tooling/, site/, _support/docs/, README.md, llms.txt, AGENTS.md, CHANGELOG.md

- [x] unit: go -C <worktree> test -count=1 ./... y python3 -I -m unittest discover -s <worktree>/tests/skills -p '*_test.py'
- [x] live: hive update --source <worktree> --dry-run con el binario del worktree · pass: retira las carpetas hive-rule-* y añade hive-core, sin cambios fuera de las rutas de E1, sus roles renderizados y los bloques globales
- [x] perf: n/a: mudanza sin cambio de lectura

### E2 — PR `feat/stage-packets` (T6, T7, T10 parte paquetes) · depends on: E1 · files: tooling/management/, content/skills/*/assets/stage.json, content/skills/*/SKILL.md, content/guidance/global.md, tests/content/, tests/pilot/, docs

- [ ] unit: go -C <worktree> test -count=1 ./... y python3 -I -m unittest discover -s <worktree>/tests/skills -p '*_test.py'
- [ ] live: bin=$(mktemp) && go -C <worktree> build -o "$bin" ./tooling/cli && "$bin" update --source <worktree> --dry-run · pass: un stage.md por flujo con manifiesto y dos en flow-verify, cada uno ≤ 24,000 B
- [ ] perf: lecturas declaradas plan/build · base: 39 · fail: > 26

### E3 — PR `feat/stage-rules-reply` (T8, T9, T10 parte checkpoint y cierre) · depends on: E2 · files: content/skills/flow-plan/, content/skills/hive-core/references/checkpoint.md, content/guidance/global.md, 7 SKILL.md, tests/skills/, tests/content/, docs

- [ ] unit: go -C <worktree> test -count=1 ./... y python3 -I -m unittest discover -s <worktree>/tests/skills -p '*_test.py'
- [ ] live: tras el merge de E3 y tu aviso, binario reconstruido (archivo temporal + mv), luego hive update --source <checkout> --out <archivo> y hive apply --plan <archivo> (D9 A) · pass: hive status sin drift, hive-core y los paquetes instalados en los 7 hosts, ninguna carpeta hive-rule-* instalada
- [ ] perf: n/a: sin meta medida en esta unidad

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, in the unit's worktree | The task's command passes | Covered by the implementation |
| Full suite | Each unit's candidate, locally | `go -C <worktree> test -count=1 ./...` y `python3 -I -m unittest discover -s <worktree>/tests/skills -p '*_test.py'` | Covered by the implementation |
| Blast radius | Each unit's candidate | `hive-rule-blast-radius` | Covered by the implementation |
| Code review | Each unit, before the push | Claude Code `/code-review` | The review answer (2026-11-21, A) |

Tras mergear E3 hay un solo refresco local (D9 A). Antes, aviso y espero a que se cierren o reinicien las sesiones abiertas. `hive status` confirma el resultado, y con eso se observan AC2 y AC5. Después del despliegue se revisan las primeras sesiones reales para ver la mejora y si los paquetes se leen enteros. Es un seguimiento, no un criterio.

Si otra sesión refresca antes (D10 A), cada unidad funciona sola, y el reporte de cada merge avisa que hay que reiniciar las sesiones abiertas.

## Checkpoint

- Updated: 2026-11-21
- Stage: build · E2 T6
- Last commit: 89befa58
- Open findings: E1 /code-review nits a cierre: link text [flow-build]/[flow-close] apunta a tests.md / close-hygiene.md (unattended-delegation:16, ticket-status.md:14, pre-existing); leaf_links_test acepta /<leaf>.md de cualquier skill; leafPathPattern recompila regex por evento; removePycache reporta motivo incorrecto al conservar; comentario leafReadBeforeAction obsoleto · decisión: findings del cierre
- Next: E2/T6 manager stage packets (hive-build-infra + hive-verify-task). Al finalizar el build (tras E3), actualizar el artifact «Skills de Hive» https://claude.ai/artifact/77U7EN4pUTcEtiAijziLxE
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/hive-core-stage-packets · feat/stage-packets

### Evidence

- base: go test ./... y python unittest exit 0 en 7f6277e0 (2026-11-21); global.md 28,459 B en 6fa78bfd
- before captures: none (sin pantalla)
- suite E1 @000c57fe (dirty: tasks.md): go vet exit 0; python unittest exit 0 (71 s); go test -count=1 ./... exit 0 (333 s)
- AC16 sweep sobre el worktree: sin salida
- dry-run E1 con binario del worktree: 708 archivos, 60 hive-rule retirados, un openai.yaml en hive-core
- verify E1 paso 1: comentarios añadidos = doc comments (Keep); supresiones añadidas: ninguna; blast radius: hook de Grok cubre references/*.md; prueba real-content del manager resuelve cada enlace skill: (rung 4)
- E1 integrado: PR #565 merge 89befa58 (2026-11-21); sin CI de PR ni de base (solo master-source-branch.yml)

## Review status and progress

2026-11-21 — T1 · c12051a9 · unverified: go test ./tests/content/... pasa solo tras T2
2026-11-21 — T3 · be3462a1 · unverified: none
2026-11-21 — T2 · b9e17031 · unverified: none
2026-11-21 — T5 · 000c57fe · unverified: none
2026-11-21 — T4 · 70f07007 · unverified: AC2: se observa en la instalación local tras el refresco posterior a E3; hoy solo prueba del manager y dry-run (sustitutos)
2026-11-21 — T6 · f6a984cb · unverified: AC5: instalación tras el refresco posterior a E3; fixtures de lectura construidos con la forma real (1\t…), no copiados de una transcripción
2026-11-21 — T7 · 4ad914c9 · unverified: none
