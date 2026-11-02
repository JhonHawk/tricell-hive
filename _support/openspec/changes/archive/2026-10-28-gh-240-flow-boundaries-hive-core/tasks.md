# Tareas — gh-240

**Commit base de la fase 2:** `96584b03` en `development` (2026-10-28, rama `fix/gh-240-flow-boundaries`).

**Commit base de la fase 1:** `ecd0e9db` en `development` (2026-10-27). El plan se escribió sobre `4ef04e01`; entre los dos solo entró #241 (gh-211), que toca `tooling/management/plan.go` y dos líneas de `deployment-manager.md`.

**Fases y PRs, según D10-A:**
- Fase 1: T1–T4.
- Fase 2: T5–T11.
- Fase 3: descartada por D16-B el 2026-10-28 (ver `proposal.md`, Entrega).

Cada fase sale de `development` en su propia rama: `feat/gh-240-hive-core-layers` (fase 1, ya integrada) y `fix/gh-240-flow-boundaries` (fase 2, propuesta).

Cada PR agrega su entrada bajo `[Unreleased]` en `CHANGELOG.md` (`CONTRIBUTING.md:128`), lista los lugares revisados y, si toca `global.md`, declara cuánto cambia su tamaño (`AGENTS.md`).

## Fase 1

### T1 — Tabla de compactación por host

- [x] `_support/docs/architecture/compaction-survival.md` registra, para Claude Code, Codex, Grok Build, Pi, OpenCode y Cursor CLI:
  - si el archivo global de instrucciones sobrevive a la compactación;
  - si el cuerpo de una skill invocada sobrevive, y con qué tope;
  - qué superficie de hook ofrece el host para reinyectar;
  - la fuente de cada dato: URL de documentación y versión instalada, u observación.

**Closes:** AC1.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `_support/docs/architecture/compaction-survival.md` (nuevo) y `_support/docs/architecture/agent-delivery.md` (enlace).

**Execution:** delegada a `hive-research`, de solo lectura y sin identificadores privados en peticiones a terceros. El hilo principal escribe el documento. Es una búsqueda documental larga, separable del resto.

**Test approach:** check: lectura del documento contra sus fuentes.

**Changes:**
- Partir de la sección de hosts de la investigación y verificar las versiones instaladas.
- Para cada celda que la documentación no resuelva, escribir "sin documentar".
- Al terminar, preguntar si se autoriza una prueba de compactación en los hosts sin documentar: D15 no la autorizó, consume cuota y reanuda los pilotos pausados. La pregunta lleva la evidencia esperada y un alcance mínimo, como pide `AGENTS.md` (Measurement). T1 se completa con "sin documentar" y la pregunta no bloquea C1.

**Verification:**
- Las cuatro columnas de los seis hosts están llenas, con una fuente o con "sin documentar".
- `rg -c 'https://' _support/docs/architecture/compaction-survival.md` es mayor que 0.

### T2 — Liberar bytes en flow-close, reporting y flow-build (H7)

- [x] Los tres archivos quedan bajo los topes de AC2 sin perder ninguna regla.

**Closes:** AC2.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:**
- `content/skills/flow-close/SKILL.md`, sección "Clean up after a merge" (líneas 86-94) y sus enlaces internos (`:21`, `:29`, `:84`). La sección pasa a `content/skills/flow-close/references/branch-cleanup.md`, que es nuevo.
- `content/skills/flow-close/references/reporting.md:96`.
- `content/skills/flow-build/SKILL.md:62`.
- Enlaces entrantes a `#clean-up-after-a-merge`: `content/guidance/global.md:118`, `content/skills/workspace-conventions/references/artifact-layout.md:43`, `content/skills/git-workflow/SKILL.md`, `site/src/content/docs/flows/close.md:45` y los demás que encuentre `rg`.

**Execution:** hilo principal. Toca archivos que T4 y T8 vuelven a editar.

**Test approach:** check: `go test ./tests/content/...` y `wc -c`.

**Changes:**
- Mover "Clean up after a merge" a `references/branch-cleanup.md`. En su lugar queda un puntero que se lee "after verifying a merge you performed or after a promotion" (close-C4).
- Convertir `reporting.md:96` en un puntero a `flow-research#deliver-the-finding` (K4).
- Quitar la salida duplicada de `flow-build/SKILL.md:62`; el traspaso queda en `:12` (build-B5).

**Verification:**
- `wc -c` cumple los topes de AC2.
- `rg -n 'clean-up-after-a-merge' content site _support/docs` solo devuelve enlaces que apuntan a `branch-cleanup.md`, o a un encabezado que sigue existiendo.
- `go test ./tests/content/...` pasa.

### T3 — Skill `hive-core` y migración de las cuatro referencias (D6-B, D11-A)

- [x] Las cuatro referencias viven en `content/skills/hive-core/references/`, con todos sus enlaces entrantes y salientes reescritos.
- [x] `hive-core/SKILL.md` dice qué contiene la skill y cuándo se lee cada referencia.

**Closes:** AC3.

**Checkpoint:** C2.

**Depends on:** T2.

**Locations:**
- `content/skills/hive-core/SKILL.md`, nuevo. Necesita frontmatter con comillas si la descripción lleva `: ` o ` #` (`tests/content/frontmatter_test.go`).
- `content/skills/hive-core/references/{questions,scope-changes,effect-authorization,project-settings}.md`, movidos con `git mv`.
- La lista de consumidores, que se rehace al empezar con `rg -l 'questions.md|scope-changes.md|effect-authorization.md|project-settings.md' content site _support/docs tests tooling README.md llms.txt`, excluyendo `_support/openspec/changes/archive`. En la base son 49 archivos.
- Tests y código:
  - `tests/content/reading_footprint_test.go:53,58`, con las rutas nuevas en el mapa `moved` (`:81-82`), para que `git show d5164f71:<ruta>` no falle.
  - `tooling/cli/doctor_project_test.go:451` y el comentario de `tooling/cli/doctor_project.go:24`.
- Documentación de arquitectura: `_support/docs/architecture/instruction-resources.md:18-24`, `deployment-manager.md:77` y `workspace-and-artifacts.md:23`.
- Documentación pública:
  - `README.md:355` y `llms.txt:31-48`.
  - `site/src/content/docs/skills.md:14-36` y `site/src/content/docs/index.mdx:80-104`.
  - Las URLs absolutas a `blob/master/.../flow-plan/references/` en `site/`. En la base son 13: `flows/{build:50,87, close:17,79, deploy:6, overview:8, plan:69,74,80, verify:35, research:53}.md` y `getting-started.md:44`.

**Execution:** hilo principal. Es mecánico, pero toca muchos archivos y no conviene repartirlo entre varios escritores.

**Test approach:** check: `go test ./...`, que valida los enlaces relativos y `skill:` de `content/` y las huellas de lectura.

**Changes:**
- Crear `hive-core/SKILL.md`. La descripción dice qué hace la skill y cuándo se lee, sin resumir sus procedimientos. El cuerpo es un índice de la forma "Before X, read references/Y".
- Mover los cuatro archivos.
- Reescribir los enlaces:
  - `skill:flow-plan/references/…` pasa a `skill:hive-core/references/…`;
  - `../flow-plan/references/…` pasa a `../hive-core/references/…`;
  - los enlaces entre archivos hermanos dentro de `flow-plan`;
  - los enlaces que salen de los archivos movidos.
- Conservar las anclas que lista `design.md`.
- Catalogar en `instruction-resources.md` a `hive-core` y los tres hogares que faltan: `delivery-decisions`, `review-decisions` y `ui-verification-decision`.

**Verification:**
- `go test ./...` pasa.
- `ls content/skills/flow-plan/references/{questions,scope-changes,effect-authorization,project-settings}.md` falla en los cuatro.
- `rg -n 'flow-plan/references/(questions|scope-changes|effect-authorization|project-settings)' content site _support/docs tests tooling README.md llms.txt` no devuelve nada.
- `rg -n '#(decide-whether-a-question-is-needed|make-user-dependent-work-self-sufficient|pause-and-handle-the-answer|what-an-answer-covers|later-completion-point)' content site` solo devuelve enlaces cuyo encabezado existe en `hive-core/references/`. Esto cubre las anclas, que `go test` no valida.

### T4 — `global.md` dividido en capas (D8-B, D12-A, D17-A)

- [x] `global.md` contiene solo los invariantes de la capa 1 y sus condiciones de lectura literales.
  reopened by the fix of the code review (`4bd79362`)
  ronda 1 (resuelta en ronda 2, AC4 y AC5 met), AC4: cinco punteros sin condición de lectura, activación desatendida perdida, intención perdida sin motivo (preservación, secretos, defectos, puertas, mecánico, evidencia, continuidad).
- [x] El procedimiento vive en las referencias nuevas de `hive-core` o en los hogares existentes de `design.md`.
- [x] Se aplican K2, K3, K15, K20 y K22, y los resúmenes divergentes G3, G6 y G8.

**Closes:** AC4, AC5.

**Checkpoint:** C3.

**Depends on:** T3.

**Locations:**
- `content/guidance/global.md` y `content/guidance/README.md`.
- `content/skills/hive-core/SKILL.md` y sus referencias nuevas.
- Hogares existentes:
  - `flow-research/references/{backlog-report,delegation}.md`;
  - `flow-close/{SKILL.md,references/incidental-findings.md}`;
  - `flow-deploy/SKILL.md`;
  - `flow-plan/references/{plan-review,change-records}.md`;
  - `engram-init-workspace/SKILL.md`;
  - `flow-verify/SKILL.md`.
- Consumidores de reglas que salen o cambian:
  - `_support/docs/architecture/tool-dependencies.md:32`, `agent-delivery.md:118,141,152`, `instruction-resources.md` (catálogo de las referencias nuevas) y `repository-and-distribution.md`.
  - `README.md:24,381` y `llms.txt:23-29`.
  - `incidental-findings.md:19` y `backlog-report.md:3`.
  - `site/src/content/docs/flows/{plan.md:71,research.md:51,build.md:44}` (temas abiertos).
  - Los que citan "as the global guidance defines it" para cambio mecánico: `flow-verify/SKILL.md:20`, `flow-build/references/tdd.md:9` y `review-decisions.md:9`. Siguen siendo válidos si la definición queda en la capa 1; se revisan sin editarlos.
  - Los que citan "the global ticket rule": `flow-build/SKILL.md:24`, `flow-close/SKILL.md:27`, `hive-core/references/project-settings.md` y `flow-build/references/verification.md:55`. Siguen siendo válidos por D17-A.
- `tests/content/reading_footprint_test.go`: agregar a cada escenario las referencias nuevas de `hive-core` que pase a necesitar.
- `specs/context-delivery/spec.md` de este cambio.

**Execution:** hilo principal. Es el núcleo del cambio, va acoplado a T3, y `AGENTS.md` exige revisar la conducta completa de cada regla.

**Test approach:** check: `go test ./...`, `wc -c` y el check de unicidad de AC5.

**Changes:**
1. Antes de editar, agregar a `design.md` la lista de unicidad: una frase textual y distintiva por cada regla que sale de `global.md`, con su destino.
2. Reescribir `global.md` según `design.md`, con D12-A y D17-A.
3. Repartir el procedimiento en referencias de 16.000 B o menos.
4. En cada destino, conservar la regla completa. Si algo se descarta, el PR dice por qué.

**Verification:**
- `wc -c content/guidance/global.md` da 17.500 o menos (D20-A).
- `rg -n '\]\(<?skill:' content/guidance/global.md`: cada línea que devuelve se revisa a mano contra AC4.
- También se revisan a mano los punteros sin enlace (`rg -n -i 'read (and (use|run) )?`|use the `' content/guidance/global.md`, como "read and use `flow-research`", `unattended-delegation`, `workspace-conventions` o el rol `hive-research`). Cada uno tiene que nombrar su acto o evento.
- Unicidad de AC5. Para cada frase de la lista, `rg -F -c --no-filename '<frase>' content | awk '{s+=$1} END {print s}'` imprime 1. Se corre como bucle sobre la lista y debe dar cero fallos.
- `go test ./...` pasa.

## Fase 2

T5 y T6 editan archivos distintos y pueden ir en paralelo. T7 va después de T6, porque comparte `site/.../build.md`. T8 va después de T6 y T7, porque escribe una línea en `flow-build/SKILL.md` y otra en `flow-verify/SKILL.md`. T9 va después de T5, porque comparte `delivery-decisions.md`. T10 va después de T5, porque comparte `hive-core/SKILL.md`. T11 va después de T5, T9 y T10. Con T5 comparte `flow-deploy/SKILL.md`, con T9 `delivery-decisions.md` y con T10 `hive-core/SKILL.md` y `global.md`, y además necesita saber cuántos bytes deja T10. Como esas tareas mueven líneas, T11 ubica cada cambio por su texto, no por número de línea.

### T5 — flow-deploy: recuperación, commits ajenos y entrada (H1, H2 y el resto de deploy)

- [x] Arreglos de AC6 y AC7. Commit `b3ceda24`.
  Ronda 1: AC6 y AC7 cumplidos. El verificador señaló que el paso 5 no decía qué hacer con los commits ajenos (F1), que no excluía la recuperación crítica (F3) y que había perdido "failed check" (F5). Se corrigió en `31e51b47`. Ronda 2 (alcance AC6 y AC7): los dos se cumplen. Quedan dos ajustes de redacción no bloqueantes, que pasan a T11 porque ya edita `flow-deploy/SKILL.md`. En `:28`, una reversión sigue el bullet Promotion y no vuelve al paso 2. En `site/.../deploy.md:25`, se quita la tensión entre "without asking" y "asks once", y también el límite "non-production" que no existe en la skill.
- [x] La regla de promoción (K14) y el registro de la promoción quedan en el paso 4; una release que publica cuenta como efecto crítico (K15).
- [x] La entrada manda un defecto sin arreglo integrado a research o a build (K21, deploy-D4).
- [x] Ruta de promoción sin ramas de entorno (deploy-D5).
- [x] El recorrido declarado tiene dónde declararse y pide sus efectos (deploy-D6).
- [x] Las reglas duplicadas quedan como punteros (deploy-D7).

**Closes:** AC6, AC7. AC21 pasó a T11 el 2026-10-28 (D24-A).

**Checkpoint:** C4.

**Depends on:** T4.

**Locations:**
- `content/skills/flow-deploy/SKILL.md:3,10,20-25`, sin tocar el enlace de `:21`, que es de T11.
- `content/skills/flow-deploy/references/{deployment-checks,urgent-fixes}.md`.
- `site/src/content/docs/flows/deploy.md`.

**Execution:** delegada a `hive-write-docs`. Sus fuentes son este registro y la sección de deploy de la investigación. Sus archivos son propios, así que corre en paralelo con T6 y T7.

**Test approach:** check: `go test ./tests/content/...` y lectura contra AC6 y AC7.

**Verification:**
- `rg -n 'existing authorization covers' content/skills/flow-deploy` no devuelve nada.
- El paso 4 nombra la acción de recuperación de cada disparador, con su objetivo y su reversa.
- El paso 6 ejecuta esa acción cuando la respuesta del paso 4 la cubre.
- El paso 5 lista los commits que no estaban en el candidato y apunta a la clase Promotion de `scope-changes.md`.

### T6 — flow-verify y flow-security: hallazgos de seguridad y momento de la revisión (H3, K6, K7, K9)

- [x] Arreglos de AC9 y AC10. Commit `2fe26075`. Verificada en la ronda 1: AC9 y AC10 cumplidos. Dos ajustes no bloqueantes pasan a T11. Primero, "in parallel… not after it" (`review-and-gates.md:35`, `flow-security/SKILL.md:18`) aplica solo cuando las dos revisiones se piden antes de que empiece la de código, porque una revisión de seguridad pedida en la validación o con un mecanismo de PR corre después. Segundo, `review-and-gates.md:33` apunta a la regla de refutación de `:14` en vez de repetirla.
- [x] El paso 4 corre en paralelo con el 3.
- [x] El puntero de severidad apunta a `incidental-findings.md`.
- [x] El brief de alcanzabilidad pasa los límites de `security-advisory` y desaparece la circularidad (security-S3).
- [x] La regla de entrega de un commit ya integrado va a su hogar (security-S4).
- [x] Entrada sin candidato (security-S6, verify-V5) y "adds no question about the level" (verify-V4).
- [x] `review-and-gates.md:16` se reescribe en viñetas (verify-V7).
- [x] El Prepare de verify apunta a la regla de commits de build (K7).

**Closes:** AC9, AC10.

**Checkpoint:** C5.

**Depends on:** T4.

**Locations:**
- `content/skills/flow-verify/{SKILL.md,references/review-and-gates.md}` y `content/skills/flow-security/SKILL.md`.
- `content/skills/flow-plan/references/review-decisions.md:21`.
- Todos los enlaces entrantes a `review-decisions.md#dedicated-review-timing`, `site/src/content/docs/flows/build.md:52` incluido.
- `site/src/content/docs/flows/{verify,security}.md`.

**Execution:** delegada a `hive-write-docs`. Es la única tarea que reescribe enlaces a `#dedicated-review-timing`.

**Test approach:** check: `go test ./tests/content/...` y `rg`.

**Verification:**
- `rg -n '^## Dedicated review timing' content` devuelve solo `review-and-gates.md`.
- `rg -n 'review-decisions.md#dedicated-review-timing' content site` no devuelve nada.
- `rg -n 'without waiting for the code reviewer|run these steps in order' content/skills/flow-verify` ya no muestra la contradicción.
- El Prepare de `flow-verify` apunta a la regla de commits de `flow-build`.

### T7 — flow-build: lecturas, commits y revisión de UI del implementador (H5, H14, K7, K10, K13)

- [x] Arreglos de AC11, AC12 y AC13 (la parte de build). Commit `52c68328`. Verificada en la ronda 1: AC11–AC13 cumplidos y AC2 a 15.092 B. Quedan dos ajustes no bloqueantes. T8 aclara en `flow-build/SKILL.md:25` que el caso 1 es "non-mechanical" y recupera "Keep dependent edits ordered" dentro de los 15.100 B. T11 alinea la condición de lectura de `verification.md:3` con la lectura en cualquier build. Re-verificada tras `a77a22ae` (T8): AC11–AC13 siguen cumpliéndose y AC2 queda en 14.915 B.
- [x] Se quita "including deployment of existing code".
- [x] `SKILL.md:24` pasa a una lista de cuatro casos (build-B4).
- [x] La revisión del implementador en `:58` se distingue de la puerta (build-B5).

**Closes:** AC11, AC12, AC13.

**Checkpoint:** C6.

**Depends on:** T4, T6.

**Locations:** `content/skills/flow-build/SKILL.md:22,24,26,40,50,58,60`, `content/skills/flow-build/references/plan-execution.md:40` y `site/src/content/docs/flows/build.md`, sin tocar `:52`, que es de T6.

**Execution:** delegada a `hive-write-docs`. Sus archivos son propios.

**Test approach:** check: `go test ./tests/content/...` y `wc -c` (16.000 o menos).

**Verification:**
- `rg -n 'ui-gate.md#' content/skills/flow-build/SKILL.md` muestra el puntero a Scope y a la inspección del implementador.
- `rg -n 'verification.md' content/skills/flow-build/SKILL.md` muestra que la lectura se hace antes de elegir las pruebas afectadas.
- La regla de commits de `:50` cubre todo commit posterior del trabajo en cualquier etapa (lectura).
- El Prepare de verify ya apunta a ella (T6).
- `wc -c` cumple.

### T8 — flow-close: cierres diferidos, memoria, pausa y casos sin cubrir (H8, H9, K1)

- [x] Arreglos de AC14, AC15 y AC16. Commits `2220d6c8` y `a77a22ae`.
  Ronda 1: AC14, AC15, AC16 y AC2 se cumplen. El verificador encontró tres choques. La ruta iterativa tenía dos hogares, en `flow-research:58` y `flow-close:18`. La vía ligera mecánica contradecía las excepciones de la pregunta de cierre y la regla de memoria. El cierre desatendido terminaba la delegación por su cuenta. Además faltaba la línea de AC8 en `site/.../close.md:74`. Todo se corrigió en `a77a22ae`. Se re-verificaron T8 y también T7, porque la corrección toca `flow-build/SKILL.md`. Ronda 2: AC14–AC16 y AC2 se cumplen, y los cuatro choques quedan resueltos. Queda un hallazgo incidental P3, previo a esta rama: `site/.../build.md:42` atribuye a la parada de una ruta iterativa la pregunta de cierre, cuando sin cambios corresponde la pregunta de ruta.
- [x] Punto de cierre de la ruta iterativa sin integración (close-C3).
- [x] Fin de sesión con un ítem abierto (close-C5).
- [x] Cierre bajo delegación desatendida (close-C8).
- [x] Vía ligera para un cambio mecánico (close-C7).
- [x] Punteros de limpieza (close-C6).

**Closes:** AC14, AC15, AC16.

**Checkpoint:** C7.

**Depends on:** T6, T7.

**Locations:**
- Además, por la verificación de T9: `site/src/content/docs/flows/close.md:73` lleva la condición del recorrido funcional pendiente (AC8).
- `content/skills/flow-close/SKILL.md:16-31,96-100` y `content/skills/flow-close/references/incidental-findings.md:9`.
- `content/skills/flow-build/SKILL.md:68` y `content/skills/flow-verify/SKILL.md:65`, solo los punteros de K1 y close-C6.
- `content/skills/workspace-conventions/references/artifact-layout.md:43`.
- `_support/docs/architecture/instruction-resources.md:18`, el paso por nombre.
- `site/src/content/docs/flows/close.md`.

**Execution:** delegada a `hive-write-docs`, después de T6 y T7.

**Test approach:** check: `go test ./tests/content/...` y `wc -c`.

**Verification:**
- `rg -n 'not a pause' content/skills` muestra la regla solo en `flow-close` y punteros en build y verify.
- `rg -n 'step [0-9]' content/guidance/global.md content/skills/flow-close/references/incidental-findings.md _support/docs/architecture/instruction-resources.md` no devuelve citas numéricas al paso de memoria.
- "When to run it" incluye el disparador de la sesión que encuentra integrado el PR de un ítem cerrado bajo revisión humana, y el del movimiento final de `Completion` (lectura, AC14).
- `wc -c` cumple.

### T9 — research, hive-core y referencias de plan (H13, H15, H16, H18, K11, K12, K14, K17–K19)

- [x] Arreglos de AC8 y AC17–AC20. Commits `320bfb08` y `4c2df17b`.
  Ronda 1: AC8, AC18, AC19 y AC20 se cumplen. AC17 no se cumple: los fallos forzados no llevaban etiqueta ni se retiraban. La corrección está en `4c2df17b`, junto con tres consumidores: `plan-review.md:30`, `plan-revision.md:11` y `delivery-decisions.md`. AC18 cambió el mensaje de `hive doctor`, así que se agregó el delta `specs/versioned-installation/spec.md`. Ronda 2: AC17 y AC18 se cumplen.
- [x] Estado de la ruta iterativa (research-R3).
- [x] Excepción del informe de backlog (research-R5).
- [x] Vía corta de research (research-R8).
- [x] Regla de nueva superficie en la clase 2 (K11).
- [x] Estado del build devuelto a plan (K12).
- [x] La cláusula de promoción de `delivery-decisions.md:5` queda como puntero a deploy (K14), y ese párrafo se reescribe en bloques Condition/Actions (plan-P4).
- [x] La copia del salto en `delivery-decisions.md:9` (verify-V6).
- [x] Orden de las rondas finales de plan (plan-P7) y entradas de plan (plan-P6).

**Closes:** AC8, AC17, AC18, AC19, AC20.

**Checkpoint:** C8.

**Depends on:** T4, T5.

**Locations:**
- `content/skills/flow-research/{SKILL.md,references/diagnosis.md,references/backlog-report.md}`.
- `content/skills/hive-core/references/{project-settings,scope-changes}.md`.
- `content/skills/flow-plan/{SKILL.md,references/delivery-decisions.md,references/plan-format.md}`.
- Textos que AC8 o AC18 dejan desactualizados: `site/src/content/docs/flows/{deploy.md:6,close.md:62}`, `README.md:378` y `site/src/content/docs/getting-started.md:41`.
- El mensaje de `hive doctor` sobre claves retiradas ("no longer read; Hive asks it each session"), en `tooling/cli/doctor_project.go`. Lo fijan además:
  - `tooling/cli/doctor_project_test.go:317`;
  - `tooling/cli/doctor_command_test.go:140`;
  - `tooling/cli/project_command_test.go:139`;
  - `tooling/cli/project_write_test.go:134,136`;
  - `tooling/cli/project_write.go:457`;
  - `_support/docs/architecture/deployment-manager.md:77`.

  Todos se actualizan juntos.
- `site/src/content/docs/flows/{research,plan}.md`.

**Execution:** delegada a `hive-write-docs`, después de T5.

**Test approach:** check: `go test ./...`. `TestRetiredHiveKeysMatchTheGuidance` exige tres cosas:
- que se conserve `## Declare settings`;
- que haya una línea que empiece por `Retired keys:`;
- que la primera oración contenga exactamente `` `Review` ``, `` `UI verification` ``, `` `Delivery` ``, `` `automatic` `` y `` `interactive` ``, en ese orden.

Se reescribe solo lo que va después del primer punto.

**Verification:**
- `rg -n 'ask as if it were absent' content` no devuelve nada.
- `rg -n 'Infer the (project )?tracker' content` no devuelve nada.
- `rg -n 'no functional walk' content/skills/hive-core/references/project-settings.md` encuentra la condición de AC8.
- `diagnosis.md` dice el límite local de las sondas y que se retiran o se listan antes de salir (lectura, AC17).
- `scope-changes.md` manda un cambio nuevo durante research a la pregunta de ruta, y su punto "Build" dice qué pasa con los commits, la rama y el ticket (lectura, AC20).
- `go test ./...` pasa.

### T10 — Disparador de tool-scripts según el host (R2, D22-A)

- [x] El disparador de `global.md` y "When to read" de `tool-scripts.md` empiezan con la condición de que el host ofrezca una herramienta que corre un script capaz de llamar a otras herramientas.

**Closes:** AC22.

**Checkpoint:** C9.

**Depends on:** T4. La dependencia de T5 cayó con la revisión del 2026-10-28, porque T5 ya no edita `hive-core/SKILL.md`. Commit `3a75e865`; `global.md` pasa de 17.483 a 17.496 B. Verificada en la ronda 1: AC22 cumplido. Nota del verificador para T11: un modelo literal podría tomar una shell por esa herramienta, así que T11 agrega "host tool calls, not a shell" a las tres condiciones cuando FR3 y FR5 liberen bytes en `global.md`.

**Locations:**
- `content/guidance/global.md:69`.
- `content/skills/hive-core/references/tool-scripts.md:5`.
- `content/skills/hive-core/SKILL.md:19`.

**Execution:** hilo principal. Son tres frases, y una está en `global.md`. `global.md` mide 17.483 B y su tope es 17.500 B (D20-A), así que solo quedan 17 B. La frase nueva no puede pasar de ese margen: o se escribe sin agregar bytes, o se recortan los mismos bytes en otra línea de `global.md` sin quitar ninguna regla. Si ninguna de las dos cosas se puede, se te pregunta si se sube el tope. El commit nombra la sesión de Grok `01a180ab-1142-7f20-94f4-3cd2257186f3` (2026-10-27): en ella el disparador aplicaba aunque Grok no ofrece esa herramienta.

**Test approach:** check: `go test ./...` y `wc -c`.

**Verification:**
- `rg -n 'tool-scripts' content/guidance/global.md content/skills/hive-core` muestra la condición del host en cada condición de lectura.
- `wc -c content/guidance/global.md` da 17.500 o menos (D20-A), y el PR declara cuánto cambia.
- `go test ./...` pasa.

### T11 — Menos fricción: efectos nombrados, línea de contexto, sí/no, relectura y hooks de memoria (FR1–FR5, FA1; D23-A, D24 A–D)

- [x] Arreglos de AC21 (FR1 y FR2): efecto crítico pedido por el usuario sobre un objetivo concreto, procedimiento documentado del proyecto, "acting on an earlier grant", ancla de `flow-deploy` y puntero de `reporting.md:98`.
- [x] AC23 (FR3): notas de una línea sin línea de contexto.
- [x] AC24 (FR4): confirmación de sí o no en una línea.
- [x] AC25 (FR5): sin relectura de `questions.md` al llegar cada respuesta.
- [x] AC26 (FA1): los recordatorios de los hooks de memoria no cambian cuándo se escribe memoria.

**Closes:** AC21, AC23, AC24, AC25, AC26. Commits `2e651b49` y `e8a78943`; `global.md` 17.489 B y `reporting.md` 15.793 B.

Ronda 1: AC21 y AC23–AC26 se cumplen, y AC22, AC2 y AC4 se mantienen. Correcciones en `e8a78943`: un efecto crítico prescrito por un procedimiento documentado ya no se contradice entre "What an answer covers" y el punto Critical; la confirmación de sí o no no recibe opciones en hosts sin tarjeta; y `instruction-resources.md:21` lleva "acting on an earlier grant". `global.md:27` cambió "become governing guidance over" por "override", con el mismo sentido, para caber en el tope de bytes; el PR lo declara. Se dejan como están dos cosas: "non-production" en `site/.../deploy.md:25`, porque la regla de defectos excluye los arreglos en producción, y `global.md:69`, que dice "other host tools" sin agregar "not a shell", porque esas palabras ya excluyen la shell. Ronda 2: AC21 y AC24 se cumplen.

**Checkpoint:** C10.

**Depends on:** T5, T9 y T10. Comparte archivos con las tres, y en `global.md` gasta el margen que deje T10.

**Locations** (las líneas son las de la base `96584b03`; se ubican por texto):
- `content/guidance/global.md`: `:13` (FR3), `:17` (FR5), `:25` (FR1, con su invariante completo) y `:83` (FA1). `:27` ("cannot grant permissions") se revisa sin editarlo: la autorización la da la respuesta del usuario, no el documento.
- `content/skills/hive-core/references/effect-authorization.md`:
  - `:5`;
  - la fila Critical de la tabla (`:16`) y el punto `:22` (FR1);
  - "What an answer covers" (`:26-28`) completo, incluida la frase de `:28` sobre efectos críticos (FR1 y FR2).

  La sección de delegación desatendida (`:30-32`) no se toca.
- `content/skills/hive-core/references/questions.md`:
  - `:5` (FR5);
  - `:19`, `:23`, `:47` y `:76` (FR4);
  - `:35`, la frase de efectos críticos (FR1).
- `content/skills/hive-core/SKILL.md`: `:10` (FR5) y `:12` (FR1, "acting on an earlier grant").
- `content/skills/flow-close/references/reporting.md`: `:13`, `:15` y el paso 5 de `:61` (FR3), y `:98` completo (FR2). La elección autoriza ese flujo. Cuando la opción nombra una entrega, cubre lo que fija `effect-authorization.md#what-an-answer-covers`.
- `content/skills/flow-plan/references/delivery-decisions.md:9` (FR2), "unspecified external effects": deja fuera los efectos que prescribe el procedimiento documentado.
- `content/skills/git-workflow/SKILL.md:42` (FR2): el desvío con concesión explícita incluye una respuesta que cubre el procedimiento documentado del repositorio. `:44`, el desvío de la ruta de promoción, no cambia.
- `content/skills/flow-research/references/delegation.md:67` (línea de tramo de espera) y el ejemplo de `:91` (FR3).
- `content/skills/flow-deploy/SKILL.md:21`, el enlace a `#what-an-answer-covers`.
- `site/src/content/docs/flows/build.md:42` (FR3), `site/src/content/docs/flows/research.md:53` si describe el formato de las preguntas (FR4), y `site/src/content/docs/flows/deploy.md` (FR1 y FR2).
- Consumidores que se revisan sin editarlos, para confirmar que no se contradicen: `hive-core/references/questions.md:15,84,86`, `flow-build/references/foreseen-authorizations.md:9`, `flow-close/SKILL.md:55`, `unattended-delegation/SKILL.md:10,24` y `hive-core/references/language.md:9`.

**Execution:** hilo principal. Toca `global.md`. Su margen se calcula después de T10: FR3 y FR5 quitan bytes, FA1 agrega. Si el resultado pasa de 17.500 B sin quitar ninguna regla, se te pregunta si se sube el tope.

**Commits:** cada commit nombra la sesión y el fallo que corrige:
- FR1, FR2 y FR4: Grok `01a180ab-1142-7f20-94f4-3cd2257186f3` (dened, 2026-10-27). El usuario validó esa conducta.
- FR3: Claude `4afe6d47-9740-4543-aa24-36d25b0ba7b5` (cam), Claude `1776e095-87b6-41f6-98b2-7fd502ce0604` (carvento) y la de Grok. Las tres omitieron la línea en las notas cortas.
- FR5: carvento, que respondió 4 rondas bien sin releer.
- FA1: cam Claude y Codex `01a180de-c680-7f91-81d7-6c50a9e02ef1` (cam, 2026-10-28).

**Test approach:** check: `go test ./...`, `wc -c` y lectura contra AC21 y AC23–AC26.

**Verification:**
- `rg -n 'even when the (user.s )?request names' content` no devuelve nada.
- AC21, por lectura:
  - `global.md:25`, la fila y el punto Critical de `effect-authorization.md` y `questions.md:35` tienen las dos mitades. Lo que el usuario pide sobre un objetivo concreto, o aprueba desde una propuesta que nombraba objetivo, reversibilidad y costo, queda autorizado. Lo que el agente propone sin esa aprobación, o lo que no tiene objetivo nombrado, conserva su pregunta.
  - "What an answer covers" cubre el procedimiento documentado del proyecto, lo define como archivo versionado leído en la sesión que nombra el efecto y su objetivo, y dice que la autorización la da la respuesta.
  - `reporting.md:98`, `delivery-decisions.md:9` y `git-workflow/SKILL.md:42` no lo contradicen, y `git-workflow/SKILL.md:44` no cambia.
  - La sección de delegación desatendida no cambia.
- AC24, por lectura: `questions.md:19,23,47,76` admiten la confirmación de sí o no como línea de pausa con la etiqueta y sin ID, y mantienen el formato completo para las decisiones con alternativas.
- `rg -n 'acting on an earlier grant' content/skills/hive-core` devuelve `references/effect-authorization.md` y `SKILL.md`.
- `rg -n 'effect-authorization.md#what-an-answer-covers' content/skills/flow-deploy/SKILL.md content/skills/flow-close/references/reporting.md` devuelve las dos líneas, y el ancla sigue existiendo.
- `rg -n 'one-line note included|however short' content site` no devuelve nada.
- `rg -n -i 'context line' content site` solo muestra reglas coherentes con AC23 (lectura).
- `rg -n 'never a bare' content` no devuelve nada, y `questions.md` dice cuándo basta una pregunta de sí o no en una línea.
- `rg -n 'when an answer arrives' content` no devuelve nada, y "Pause and handle the answer" sigue igual.
- `rg -n 'hook' content/guidance/global.md` muestra la cláusula de AC26.
- `wc -c content/guidance/global.md` da 17.500 o menos, y el PR declara cuánto cambia. `effect-authorization.md`, `questions.md` y `reporting.md` miden 16.000 B o menos.
- `go test ./...` pasa.

## Verificación y revisión humana

**Checkpoints:**
- C1 — T1, T2: T2 libera los bytes de los que dependen T4, T7 y T8.
- C2 — T3: cambia rutas de referencias, que son un contrato persistido.
- C3 — T4: cambia el contrato de la guía global.
- C4 — T5: cambia el contrato de autorización de deploy.
- C5 — T6: cambia el contrato de la revisión.
- C6 — T7: cambia el contrato de los commits.
- C7 — T8: cambia los disparadores del cierre.
- C8 — T9: cambia el contrato de la configuración del proyecto.
- C9 — T10: cambia una condición de lectura de la capa 1.
- C10 — T11: cambia el contrato de autorización, de preguntas y de informes.

| Puerta | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Suite completa | Candidato final de cada fase, local | `go vet ./...` y `go test ./...` en verde | Cubierta por la implementación |
| Revisión dedicada | Candidato local de cada fase, antes de pedirte validación | `/code-review` de Claude Code; cada hallazgo se corrige o se refuta | D14-A |
| Prueba de uso real (fase 1) | Candidato local ya commiteado | `bin="$(mktemp -d)/hive"; go build -o "$bin" ./tooling/cli; "$bin" update --source . --out "$(dirname "$bin")/plan.json"` guarda el plan sin aplicarlo. Con `jq '.changes[] \| select(.after==null) \| .target.Path' "$(dirname "$bin")/plan.json"` deben aparecer las cuatro referencias viejas de `flow-plan`, y `hive-core` debe estar entre los destinos | Local, cubierta |
| Validación humana | Antes de cada push | Diff de la fase, tamaños, lugares revisados | D13-A |
| Refresco local | Después de cada merge | Actualizar `development` local al commit del merge; `go build -o "$(command -v hive)" ./tooling/cli` cuando cambian `tooling/` o `integrations/` (la fase 1 sí los cambia); `hive update`; reportar el commit desplegado | Marcador de mantenedor (`AGENTS.md`) |
| Reinicio de sesiones | Después del refresco de la fase 1 | Avisar al usuario: las sesiones abiertas, esta incluida, conservan la guía vieja con rutas a `flow-plan/references/` que ya no existen. Reiniciar antes de la fase 2 y mostrar la línea de sesiones de `hive doctor` | Aviso, sin efecto |

**Sin cubrir:**
- Los pilotos de conducta están pausados. Ningún AC mide si el modelo sigue las reglas reorganizadas; los checks miden texto, bytes y enlaces.
- Las pruebas de compactación de T1 esperan autorización.
- Las anclas y `site/` solo se cubren con los `rg` de T3, no con `go test`.

## Estado de la revisión y progreso

- **Revisión del plan, ronda 1** (revisión 44d5a7cce859): coherencia de la guía, manager y tests, entrega y orden. 16 hallazgos bloqueantes, todos reconciliados (ver `proposal.md`, sección Estado).
- **Re-revisión** (revisión ba73a594aaca): los bloqueantes se resolvieron y las cuatro correcciones nuevas se reconciliaron con las propuestas de los revisores (ver `proposal.md`, Estado).
- **Progreso (2026-10-27):** fase 1 integrada en `development` con el PR #242 (merge `96584b03`). T1–T4 en `[x]`. Instalación local refrescada desde `96584b03`, con el binario recompilado y `hive apply` aplicado (resultado `c07bcd55`): `~/.agents/skills/hive-core/` instalado y las referencias viejas de `flow-plan` retiradas. La rama `feat/gh-240-hive-core-layers` queda hasta el cierre de #240.
- **Modelos:** implementador en el hilo principal, Claude Opus 5.5 (observado). Verificadores `hive-verify-task` con el perfil configurado, no observado.
- **Pruebas de compactación (D19-B):** corridas en 5 hosts; Cursor no corrió por falta de `cursor-agent login`.
- **Revisión del plan por D22-A (2026-10-27):** la revisión de tres sesiones reales posteriores al despliegue agregó AC21 (R1) a T5 y creó T10 con AC22 (R2) y el checkpoint C9. No invalida ninguna tarea `[x]`: T4 no fijó la condición de tool-scripts y no se reabre. Las observaciones con su conteo están en `proposal.md` (Entrega, D22-A). Como agrega criterios, pasó por una re-revisión de `hive-review-plan` (coherencia de la guía, unos 114 k tokens y 1,6 min). Encontró dos bloqueantes:
  - B1: AC21 contradecía tres reglas que cubren la cadena de entrega nombrada (`flow-deploy:21`, `reporting.md:98` y `questions.md:35`).
  - B2: la verificación de T5 no cubría todo AC21.

  También dejó siete observaciones no bloqueantes, N1–N7. Las apliqué tal como el revisor las propuso, sin otra ronda:
  - AC21 cubre solo los efectos no nombrados y tiene una cláusula de falsedad basada en archivos.
  - Se alinea la condición "acting on an earlier grant".
  - Se respeta la sección de delegación desatendida.
  - T10 corrige sus archivos, va después de T5 y tiene en cuenta el margen de 17 B de `global.md`.
  - El commit de T10 nombra la sesión observada.

  Ningún escritor se cruza (N8).
- **Revisión del plan por D23-A y D24 A–D (2026-10-28):** se revisaron por segunda vez las sesiones reales, ahora con Codex incluido. El usuario pidió quitar fricción.
  - **Invalida:** R1, que sale de T5. AC21 pasa a decir lo contrario y se mueve a T11.
  - **Agrega:** T11 con AC23–AC26 y el checkpoint C10.
  - **Fase 3:** se descarta por D16-B.
  - **Tareas `[x]`:** no invalida ninguna, porque T1–T4 no fijaron estas reglas.
  - **Re-revisión:** un revisor de coherencia de la guía encontró los bloqueantes B1–B4 y dejó las observaciones N1–N5. Se aplicaron con su propuesta y sin otra ronda (detalle en `proposal.md`, Estado).
- **Fase 2, build (2026-10-28):** implementación en hijos `hive-write-docs` (T5–T9) y en el hilo principal (T10, T11). Los modelos son Claude Opus 5.5 en el hilo principal (observado) y los verificadores `hive-verify-task` con su perfil configurado (configurado, no observado). Las siete tareas quedaron `[x]` tras sus verificaciones; las rondas están anotadas en cada tarea. Suite completa: `go vet ./...` y `go test -count=1 ./...` en verde sobre `e8a78943`.
- **`/code-review` high (D14-A) sobre `development..e8a78943`:** encontró 10 hallazgos y el revisor no los verificó, así que los juzgó el hilo principal.
  - **Corregidos en `07db90b4`:**
    - un procedimiento documentado cuenta solo como está en la rama base;
    - el invariante crítico del global nombra la cobertura por respuesta;
    - tool-scripts en el global excluye la shell;
    - los commits ajenos al reentrar ya no amplían el candidato;
    - el fix-forward tiene ruta;
    - el disparador 5 solo mueve el ticket y limpia;
    - la excepción desatendida pasa a Exceptions;
    - la página de seguridad del sitio dice cuándo las revisiones van en paralelo.
  - **Descartados con razón:**
    - un "ok" a una propuesta con varios efectos críticos nombrados los cubre: es la conducta que el usuario validó con "borra ambas" (D24-A);
    - las reglas de "al llegar una respuesta" se apoyan en la lectura previa a la pregunta: es la consecuencia de D24-D, anotada como riesgo N5.
  - **Re-revisión:** una sola ronda, limitada al diff `e8a78943..07db90b4`, junto con la re-verificación de los criterios de T5, T6, T8, T10 y T11 que tocó el arreglo.
  - **Resultado de la re-revisión:** 12 hallazgos, varios de ellos regresiones del propio arreglo.
    - El disparador 5 recortado dejaba sin cerrar el registro de cambio.
    - La pregunta de reentrada de deploy quedaba invertida respecto del bullet Promotion.
    - El invariante crítico del global quedaba demasiado amplio.
    - El sitio estaba desactualizado.

    Se corrigieron en `c5f04f59` sin otra revisión de código, como permite la regla de una sola re-revisión: arreglos locales, dentro del cambio y sin cambio de contrato respecto de lo ya decidido. En concreto:
    - el disparador 5 corre el cierre completo en el siguiente punto de parada, "as if you had merged it";
    - la reentrada pregunta por commit y arranca una corrida nueva con el arreglo;
    - el fix-forward de producción urgente va por urgent-fixes;
    - el global vuelve al invariante estrecho y conserva el puntero completo a la línea de contexto;
    - el procedimiento documentado se lee desde la rama base;
    - el sitio queda alineado.

    Un `hive-verify-task` re-verifica los criterios afectados y la suite completa vuelve a correr.
- **Siguiente paso:** `flow-build` en la fase 2 (T5–T11), en una rama nueva desde `development`, propuesta como `fix/gh-240-flow-boundaries`.
