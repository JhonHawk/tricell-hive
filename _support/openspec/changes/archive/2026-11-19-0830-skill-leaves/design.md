# Diseño

## Approaches

Se presentaron tres enfoques el 2026-11-18:

- **Hojas como referencias dentro de una skill índice.** Funciona igual en los 7 hosts y reutiliza lo que ya existe.
- **Hojas como skills sueltas, al estilo de pstack.** Requiere un interruptor por host y un cambio en el manager.
- **Reparación puntual sin estructura nueva.**

El usuario eligió las skills sueltas (D3 B), y luego un interruptor por host (D6 A). Lo validamos contra las referencias:
- **pstack.** Oculta 57 de sus 58 skills con `disable-model-invocation`. Su router lista cada una con «cuándo aplica» y la nombra por su ruta relativa (`skills/poteto-mode/SKILL.md:11,33-63`). El agente solo puede citar las hojas que leyó (`:15`).
- **gentle-ai.** Lo que no debe catalogarse va en una carpeta sin `SKILL.md` (`internal/components/skills/inject.go:140`). Sus descripciones tienen un tope de 160 caracteres, comprobado por un test (`internal/assets/skills_frontmatter_test.go:67-77`).
- **superpowers y openspec.** No ocultan nada.
- **mattpocock-skills.** Combina `disable-model-invocation` con `agents/openai.yaml` (`skills/productivity/grill-me/agents/openai.yaml`), el mismo par que adopta este plan.

## Decisions

- **Nombres.** Las hojas de proceso se llaman `hive-rule-<nombre>`, las de principios `hive-principle-<nombre>` y la de TypeScript `hive-typescript`. El prefijo evita chocar con skills propias del usuario, porque el manager aborta la instalación ante una carpeta ajena con el mismo nombre (`plan.go:603-627`). Es una decisión menor, tomada como supuesto.
- **Índice.** Vive en una sección nueva de `global.md`, «Rule leaves». Tiene una línea por hoja: la condición observable y el enlace `skill:` que el instalador convierte en ruta absoluta. Los flujos repiten solo el enlace, en el paso donde la hoja aplica.
  - Los principios tienen su propio índice, en la hoja `hive-principle-index`. La enlazan el paso de escritura de código de `flow-build`, los roles `hive-build-*` y `hive-review-code`.
- **Invariantes que se quedan enteros en `global.md`.** Autorización y efectos críticos, secretos, preservación, cambiar solo el proyecto actual, la lectura de `HIVE.local.md` al inicio de la sesión, los modos de entrega y lo que cada uno autoriza, y la línea mínima de cada hoja de preguntar (tarjeta sola, con texto antes).
- **Antigravity.** Se queda como D6 A lo eligió: la hoja se enlaza en su catálogo y queda listada, con una descripción de una línea. Probamos sin enlace y lo descartamos: un enlace relativo desde un flujo enlazado (`…/config/skills/flow-build/../hive-rule-x`) no resuelve si el host resuelve la ruta como texto. Hive rechaza el alcance de proyecto en Antigravity (`antigravity.go:71-72`).
- **Índice.** Dice «lee el archivo», nunca «invoca la skill». La documentación de Claude Code dice que bloquea la invocación de una skill oculta y le indica al modelo no reproducir sus pasos por otra vía. Por eso la hoja se lee como archivo y se observa en una sesión real antes de la extracción masiva (T5).
- **Invariantes que no salen de `global.md`.** Además de los ya listados, se quedan enteros:
  - La línea 20: escalar al momento un P0, un hallazgo de seguridad o un secreto expuesto.
  - Las autorizaciones de escritura al tracker de la línea 21.
  - La autorización de limpieza al cierre de la línea 32.
  - La línea de «Secrets» sobre el commit por `git_state.py`.
  - La colocación básica de artefactos («Support folder»).

  Solo se mueve su detalle de procedimiento. El «done when integrated» y `Completion` se mudan juntos a `hive-rule-ticket-state`, y en `global.md` queda la autorización con su enlace.
- **Comandos.** Las hojas ocultas aparecen como `/hive-rule-*` en los hosts que respetan la marca, igual que las 57 de pstack. Lo tomamos como supuesto aceptable.
- **Una tarea por archivo.** Cada archivo lo edita una sola tarea de extracción, y las tareas de escritura corren en secuencia (T5 → T6 → T7 → T8). Cada tarea:
  - Mueve las rutas de `reading_footprint_test.go` y `tests/pilot/` de los archivos que toca.
  - Reescribe las citas en prosa de las secciones que vacía, como «Questions and routes», «Local instructions», «Delivery», «Test accounts», «Project settings» y «Scope changes».
- **Fusiones con pstack (fase 2).**
  - `fix-root-causes` y `attack-the-premise` entran en `hive-rule-defect-fix`.
  - `test-behavior-not-implementation` entra en `hive-rule-tests`.
  - `subtract-before-you-add` y `laziness-protocol` entran en `hive-rule-reuse-before-writing`.
- **Principios descartados de pstack.** Chocan con reglas de Hive:
  - `never-block-on-the-human`, con la autorización.
  - `outcome-oriented-execution`, con los planes de compatibilidad.
  - `exhaust-the-design-space`, con el ejemplo único de Figma.
  - `guard-the-context-window`, que ya cubre la delegación.
  - `build-the-lever`, `encode-lessons-in-structure`, `experience-first`, `sequence-verifiable-units`, `foundational-thinking` y `redesign-from-first-principles` no se adoptan en esta fase.
- **Redacción.** Las hojas se escriben con nuestras palabras. El crédito a pstack, que es MIT, va en un `references/notice.md` no enlazado, como en `code-audit.md`.

## Proposed shape

### Hojas de proceso (fase 1)

Las fuentes vienen de los dos inventarios del 2026-11-18. Los bytes son estimados.

| Hoja | Fuentes que fusiona | Condición de lectura | Enlazada desde |
| --- | --- | --- | --- |
| hive-rule-decision-format | global «Questions and routes»: formato de decisión, hallazgos, bloqueos y validación | Al escribir una decisión, un hallazgo, un bloqueo o una petición de validación | global (índice), todas las flows que preguntan |
| hive-rule-question-channel | global (regla del #491): tarjeta, texto antes, tarjeta sin contestar, preguntar en texto | Al preguntar una ronda, o cuando una tarjeta vuelve sin respuesta | global |
| hive-rule-report-endings | global: bloque de siguiente paso, finales de análisis y de ticket, acciones sueltas, compactar. Más el final del paso 5 de `flow-research` y la regla de encabezado de `flow-close` | Al terminar un informe, un análisis o un flujo | research, plan, close |
| hive-rule-delivery-ask | global «Delivery»: las preguntas de entrega y revisión, y su recomendación | Antes de la primera edición, si no hay respuestas de entrega en vigor | plan paso 4, build |
| hive-rule-foreign-commits | global «Delivery», última regla. Más `git-workflow` 13/14/26, `flow-deploy` 28 y `urgent-fixes` 7 | Antes de la validación o del push, si la rama tiene commits ajenos | verify, git-workflow, deploy |
| hive-rule-local-record | global «Local instructions»: escritura, detalle, helper y scripts | Al añadir, actualizar o marcar una entrada de `HIVE.local.md`, o al nombrar un script local | global |
| hive-rule-test-accounts | global «Test accounts» | Antes de un chequeo que inicia sesión o usa registros existentes | verify, deploy |
| hive-rule-ticket-state | global 21 (el punto de terminado, sin las autorizaciones) y 115 (`Completion`, `Other trackers`). Más `flow-close` 10 y 18, `flow-deploy` 32 y `flow-build` 14 | Al mover, comentar o dar por terminado un ticket | build paso 1, close, deploy |
| hive-rule-scope-changes | global «Scope changes» y la regla de cuándo se propone un ticket | Cuando el usuario cambia el alcance o se propone un ticket | global |
| hive-rule-project-settings | global «Project settings» | Al leer o pedir una clave de `## Hive` | global |
| hive-rule-git-commit | `git-workflow` 15-17, 20-21: staging por ruta, hooks, escáner. Más global «Secrets», commit por `git_state.py` | Antes de cualquier `git add` o commit | build (paso de commit), git-workflow |
| hive-rule-merge-gate | `git-workflow` 23-33 (revisiones del PR, espera de CI, puerta de merge). Más `flow-verify` 21 | Después del push y antes del merge | verify paso 6, git-workflow |
| hive-rule-promotion-path | `git-workflow` 34, `flow-deploy` 14 y 27, `urgent-fixes` 7 | Ante un branch, merge o push que toque un entorno posterior a la base | git-workflow, deploy |
| hive-rule-close-modes | `flow-close` 12, 14, 23, 25, 27, 29, 70, en una sola tabla de decisión | Al empezar un cierre | close (primera línea), deploy, unattended, git-workflow, figma |
| hive-rule-close-hygiene | global 32, `flow-close` 19-20 y 31-38 (archivar, limpieza, trabajo pendiente) | En los pasos 5 a 7 del cierre | close |
| hive-rule-close-report | `flow-close` 40-57 y 65-85 (informe final, hallazgos, pie). Más global 20 (P0) | Al escribir un informe final o clasificar un hallazgo | close, unattended, git-workflow |
| hive-rule-reuse-before-writing | `flow-build` 28-38 | Antes de escribir código nuevo o de comparar enfoques | build paso 2, plan paso 2 |
| hive-rule-tests | `flow-build` 40-67 (filtro de tests, TDD, sin runner) | Antes de escribir, cambiar o borrar un test, o cuando falla un chequeo | build paso 2, plan paso 2, roles |
| hive-rule-defect-fix | `defect-fix.md` y las cláusulas de build 33 y 44 y de verify 18 | Antes de la primera edición de una corrección de defecto | build, verify, deploy, `hive-review-code` |
| hive-rule-delegation | `delegation.md` (decisión de hijo, hijos en paralelo, espera) | Antes de decidir si un hijo hace un trabajo | global, todas las flows |
| hive-rule-reviewer-brief | `delegation.md` 37-45 y el seguimiento de UX de verify 18 | Al preparar el brief de un hijo revisor | plan paso 5, build, verify |
| hive-rule-business-rules | `business-rules.md` (registro y escritura al ticket) y las cláusulas de global 21, build, plan y verify | Cuando se decide, asume, reemplaza o cancela una regla de negocio | plan, build, verify, close, figma |
| hive-rule-external-effects | `outside-the-repository.md` y las cláusulas de plan 11 y 13 y de build 15 | Cuando el enfoque lee o cambia un sistema externo, o antes de un efecto externo | plan paso 2, build paso 2 |
| hive-rule-ui-plan | plan 17-27 (decisiones de UI, ejemplo aprobado, orden de tareas de pantalla) | Al planear un cambio de pantalla | plan |
| hive-rule-ui-verification | `ui-verification.md` | Al capturar UI, antes de editar una pantalla existente o al auditar UX | build, verify |
| hive-rule-viewer | `viewer-publication.md` y `walkthrough.md` | En la primera captura de UI, en una parada de revisión y en el cierre | build, verify, close |
| hive-rule-change-folder | plan 39-60 (carpeta de cambio y revisión de un plan en curso) | Cuando el plan necesita carpeta, o cuando un plan escrito cambia | plan |
| hive-rule-checkpoint | build 20 y 22, verify 23 | Al empezar, después de una compactación y al pausar | build y verify (primeras líneas) |
| hive-rule-checks | verify 25-31 y 20, build 18 (reutilizar chequeos, CI local) | Antes de correr o repetir chequeos, o con todas las tareas hechas | build, verify |
| hive-rule-review-findings | verify 18 y 33-35 (evidencia de hallazgos, ronda de correcciones) | Al juzgar un hallazgo o aplicar correcciones tras las revisiones | verify |
| hive-rule-project-verification | `project-verification.md` | Al añadir o cambiar comandos de verificación del proyecto | plan, build, verify |
| hive-rule-stored-data, hive-rule-security-boundaries, hive-rule-suppression-review | Las referencias actuales del mismo nombre | Las condiciones de lectura actuales | Los enlaces actuales |
| hive-rule-design-stage | research 35, figma 16 y 18/30 (cuántos ejemplos) | En el paso 5 de research ante un cambio de pantalla, o al elegir cuántos ejemplos dibujar | research, figma |
| hive-rule-unattended-effects | `unattended-delegation` 17-21 | Al tomar un efecto en trabajo desatendido | unattended |

Las referencias que se quedan como referencias, porque se leen siempre en su paso o tienen un solo padre:
- `plan-handoff.md`.
- `host-session-identity.md`.
- `code-audit.md`, que se parte por sección si supera el tope.
- `execution-record.md` y `e2e-coordination.md`, que se parten por responsabilidad.
- Las de `project-layout`, `create-issue` y `create-prd`.

### Router y playbooks

`flow-research` pasa a llamarse `hive-mode`, el equivalente de `poteto-mode`. Conserva lo que research hace hoy: leer antes de preguntar, devolver el pedido con nuestras palabras, preguntar lo que falta y elegir la ruta. Nunca edita archivos del repositorio. Ante un cambio corto y claro, pasa el trabajo a `flow-build` en el mismo turno (D15 A).

Además, clasifica el trabajo y abre su playbook en `playbooks/<tipo>.md`. Las etapas siguientes leen de ese playbook lo propio del tipo. Los playbooks son:
- `feature`, `ui-change`, `bugfix`, `hotfix`, `data-change`, `dependency-security`, `infra-ci`, `investigate`, `ticket-status`, `deliverable` y `pr-review-deploy`.
- Dos adaptados de pstack: `refactor` y `perf`.

Lo que hoy vive como referencias de research pasa a sus playbooks: `status.md` y `security-advisory.md`. `code-audit.md` (13.5 KB) se queda como referencia de `hive-mode` y lo enlaza el playbook `investigate`.

El cambio de nombre toca a todos los que ofrecen `/flow-research`: los bloques de siguiente paso, `site/`, `README.md`, `llms.txt`, `tests/pilot/flows.go` y la instalación. El manager quita la carpeta vieja porque la administra.

**Cómo se elige el playbook.** Se sigue a pstack (`poteto-mode/SKILL.md:108`): un trabajo usa un solo playbook.
- **Sin playbook que encaje.** Si ninguno encaja, o el trabajo es grande y transversal, `hive-mode` sigue la ruta general de hoy, de research a plan o build. Es lo que en pstack hace `figure-it-out`.
- **Cambiar de playbook es un traspaso explícito,** nunca un salto silencioso:
  - `investigate` → `bugfix`, cuando se confirma un defecto y pides el arreglo.
  - `refactor` que encuentra un bug: se separa como trabajo aparte, como pide pstack (`refactoring.md:5`).
- **A mitad del trabajo, un cambio de tipo se trata como cambio de alcance,** con su clase y su destino en una línea.

**El playbook elegido queda registrado, con un solo hogar por caso.**
- **Trabajo con plan.** Va en el campo `Playbook` de `proposal.md`: lo escribe `flow-plan`, y la plantilla gana ese campo.
- **Trabajo sin carpeta de cambio.** Va en una línea `- Playbook:` del archivo de checkpoint, escrita a mano por `flow-build`.

Cada etapa lo lee de ahí, también en otra sesión o después de compactar. `tasks.py` no gana una opción `--playbook`.

**Sin copias de lo que ya existe.** `hotfix` enlaza `flow-deploy/references/urgent-fixes.md` en vez de repetirlo.

### Casillas por unidad de entrega (AC16)

Un plan que se entrega en dos o más unidades (PR, commits a la base en `direct-base`, o candidatos en `hold`) lleva en `tasks.md` una sección `## Delivery units` (encabezado fijo en inglés, como `## Checkpoint`), fuera de las tareas `### T<n>`, para que `tasks.py done` no las marque:

- **Encabezado por unidad:** `### E<n> — <PR, commit o candidato> · depende de: <E<m>, o ninguna> · archivos: <rutas>`.
- **Casillas:**
  - `- [ ] unit: <comando>`
  - `- [ ] live: <escenario> · pass: <predicado>`
  - `- [ ] perf: <métrica> · base: <valor medido> · falla: <umbral>`
- **Marcador de «no aplica».** Reemplaza al predicado: `- [ ] live: n/a: <motivo>`. `check_change.py --archive` la cuenta como resuelta.
- **Quién marca.** `flow-verify` marca a mano la casilla cuyo chequeo pasó, y deja el resultado en `### Evidence`.
- **Qué rechaza `check_change.py`:** una casilla `live:` sin `pass:` y una `perf:` sin `base:` y `falla:`, salvo con `n/a:`. Hoy no las rechaza.
- **Dónde van los resultados.** En el `### Evidence` del checkpoint, nunca en el texto de la casilla, para que editar un resultado no cuente como cambio del plan en la coordinación E2E.

La condición para usar esta sección («dos o más unidades de entrega») vive en `hive-rule-change-folder`.

### Compatibilidad con los modos de entrega

Ningún playbook, hoja ni práctica nueva fija o sustituye el modo de entrega. Todos usan el modo en vigor (`interactive`, `automatic`, `ready`, `direct-base` o `hold`) y lo que cada uno autoriza según la sección «Delivery» de `global.md`.

| Pieza | interactive | automatic | ready | direct-base | hold |
| --- | --- | --- | --- | --- | --- |
| Playbook `hotfix` | Para antes del push; la promoción, por `flow-deploy` y su autorización por entorno | Push, PR y merge sin parar; la promoción sigue pidiendo su autorización por entorno | PR abierto, sin merge; nada se promueve | Push a la base; la promoción sigue a `flow-deploy` | Sin commit; el árbol verificado se entrega como bloqueo, porque la urgencia necesita publicar |
| Playbook `pr-review-deploy` | Revisa. El merge de un PR que este trabajo no escribió pide tu instrucción en todos los modos; el modo solo cubre las correcciones que el trabajo hace en su propio commit. La promoción sigue la autorización por entorno de `flow-deploy` | Igual | Igual; además nunca hace merge de lo propio | Igual | Igual; sin commit de lo propio |
| Complementos de verify (no-comments, blast-radius, prosa) | Antes de la revisión, en todos los modos; sus correcciones entran en el mismo candidato | Igual | Igual | Igual | Igual, sin commit |
| `interrogate` | Por trabajo, si lo eliges | Igual; en trabajo desatendido, solo con monto autorizado | Igual | Igual | Igual; revisa el árbol sin commit |
| Casillas por unidad de entrega (AC16) | Por PR; la casilla en vivo de un cambio de interacción se cierra en tu validación | Por PR; el agente cierra cada casilla antes del merge | Por PR; se cierran antes de dejar el PR abierto | Por commit a la base, porque no hay PR | Por candidato, sin PR ni commit |
| Registro TSV de decisiones (AC16) | Local; no entra en commits ni en el PR salvo que lo pidas | Igual | Igual | Igual | Igual |
| Pausa con lo no escrito (AC16) | Sin commit `wip:`: solo se hace commit de la carpeta de cambio, como en la pausa de hoy (`git-workflow`, línea 19); la nota nombra los archivos sin commit y lo no escrito | Igual | Igual | Igual (la carpeta de cambio se publica como hoy en `direct-base`) | Igual, sin commit |

### Complementos de verify (D14)

Las hojas se leen en el paso de `flow-verify` previo a la revisión:
- **`hive-rule-no-comments`.** Quita los comentarios que narran o repiten el código y propone convertir en chequeo una restricción escrita como comentario.
- **`hive-rule-blast-radius`.** Busca qué rompe el cambio fuera del diff y prueba con código real el hecho que lo hace seguro.
- **`hive-rule-prose`.** Junta `unslop` y `technical-writing` para el texto del PR, los commits y la documentación. También la enlaza `hive-rule-git-commit`, así que se lee al escribir el mensaje de commit en build y no solo en verify.
- **`hive-typescript`.** El índice de principios la enlaza con la condición «al editar `.ts` o `.tsx`». Así no depende de que el modelo la elija en los hosts sin `paths`.
- **`hive-rule-interrogate`.** Se lee solo cuando el usuario dijo que sí a la revisión adversarial de ese trabajo. La hoja de la pregunta de entrega la pregunta en la misma ronda que la revisión, como una decisión aparte de sí o no por trabajo (la revisión de código encontró que, como cuarta opción de la pregunta de revisión, se reutilizaba en toda la sesión). Dice que gasta cuota de otro proveedor, así que decir que sí autoriza ese gasto para ese trabajo.

### Fase 2

- **Índice:** `hive-principle-index`.
- **Principios:** `hive-principle-type-system-discipline`, `-boundary-discipline`, `-model-the-domain`, `-make-operations-idempotent`, `-migrate-callers-then-delete-legacy-apis`, `-minimize-reader-load`, `-separate-before-serializing-shared-state` y `-explain-the-number`.
- **`hive-typescript`**, con `paths`.
- **Fusiones:** las tres de «Decisions».

Contando los fundidos, suman 12 principios.

### Frases clave para el test de un solo hogar (AC4)

T3 fija, antes de T5, una frase literal de varias palabras por hoja. Las líneas mínimas que se quedan en `global.md` no cuentan como copia. Al menos tres de esas frases deben estar hoy en dos o más archivos, para que el test falle sobre la base. Se miden con `rg -l` y se anotan aquí con su cuenta actual. Los nombres de una sola palabra, como `Completion`, no sirven de frase.

Fijadas en T3 (`tests/content/single_home_test.go`), con su cuenta de archivos en la base `2bb96478`:

| Frase | Hoja | Archivos hoy |
| --- | --- | --- |
| «Never run `git add -A`» | git-commit | 1 |
| «Before the validation, or before the push when the mode» | foreign-commits | 2 (`global.md`, `git-workflow`) |
| «is done when its changes are integrated» | ticket-state | 1 |
| «adds, replaces or cancels a business rule, record it in» | business-rules | 2 (`flow-build`, `flow-verify`) |
| «Name the cause before you edit» | defect-fix | 1 |
| «yourself that the host has a connection that can write» | design-stage | 2 (`flow-research`, `flow-figma-design`) |

Cada tarea de extracción activa sus frases y puede sumar más, siempre que no sean líneas mínimas que se quedan en `global.md`.

### Enlaces requeridos (AC7)

T3 guarda la columna «Enlazada desde» de la tabla como datos del test. El test comprueba que cada flujo listado contiene el enlace a su hoja.

## Impact

- **Archivos.** Unos 90: ~38 carpetas de hoja nuevas, 10 skills de flujo, `global.md`, 8 roles (solo enlaces), `tooling/management` (Codex y Antigravity), `tests/content`, `tests/pilot`, `README.md`, `llms.txt`, `site/`, `instruction-resources.md` y `rule-enforcers.md`.
- **Dónde se nota.** La guía instalada en los 7 hosts de esta máquina. No toca producción de ningún proyecto.
- **Riesgo principal.** Que una hoja no se lea en su momento. Lo mitigan el índice siempre cargado, el enlace en el paso, la línea mínima en `global.md` para las reglas de autorización y la observación en sesiones reales (AC8).

## Recovery

Revertir los merges de los PR 1 y 2 en `development` y refrescar la instalación devuelve la guía anterior, en este orden:
1. Aplicar el contenido revertido con el binario actual. Este quita las carpetas de hoja y los `agents/openai.yaml` generados, que administra.
2. Recompilar el binario revertido.

Un binario viejo podría rechazar un estado que registra destinos generados que no conoce. Por eso T10 prueba una vez ese orden antes del merge. La integración del #491 se revierte aparte. No hay datos que se pierdan.
