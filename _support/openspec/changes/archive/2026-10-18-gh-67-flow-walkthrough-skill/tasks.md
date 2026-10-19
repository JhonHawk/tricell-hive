# Tareas: skill `flow-walkthrough` y ajuste `UI verification`

Commit base: `b73bb16e` (`development`, 2026-10-18). Rama de trabajo: `feat/gh-67-flow-walkthrough-skill`.

## Tareas

### T1 — Skill `flow-walkthrough` y punteros

- [x] Existe `content/skills/flow-walkthrough/SKILL.md` con las siete secciones de DS1 (texto movido de `verification.md:65–79` más grabación DS2 y página DS3); la sección "Human UI handoff" de `verification.md` es el puntero de DS4 y la fila "Local candidate" nombra la skill; `flow-build/SKILL.md:34,75`, `flow-plan/SKILL.md:62` y `hive-review-ux.md` paso 3 apuntan a la skill; `hive-verify-change.md` tiene el paso de grabación.

**Closes:** AC1, AC2, AC3.

**Depends on:** T2 (la skill enlaza `../flow-report/assets/skeleton-walkthrough.html`, y la validación de enlaces de la release rechaza un destino inexistente).

**Locations:** `content/skills/flow-walkthrough/SKILL.md` (nuevo); `content/skills/flow-build/references/verification.md:33,65–79`; `content/skills/flow-build/SKILL.md:34,75`; `content/skills/flow-plan/SKILL.md:62`; `content/agents/review/hive-review-ux.md`; `content/agents/quality/hive-verify-change.md`.

**Execution:** hilo principal. El movimiento de texto y los punteros están acoplados entre seis archivos y el brief sería más largo que el cambio.

**Test approach:** check (`rg` por criterio, enlaces resueltos, grabación de prueba).

**Changes:** DS1, DS2, DS4 (primeros dos puntos). Mover el texto sin cambiar su significado; donde una frase movida nombraba "this reference", enlazar `verification.md`. Añadir la lectura de `browser-automation.md` y de `questions.md` donde la sección movida las daba por leídas desde la cabecera de `verification.md`. La línea 61 no se toca aquí: la reescribe T5.

**Verification:**
- `test -f content/skills/flow-walkthrough/SKILL.md` y `rg -n "record start|record stop|ffprobe|automatic|walkthrough.html|not recorded|stale" content/skills/flow-walkthrough/SKILL.md` devuelve cada término (AC1).
- `rg -n "correction round per request|first rendered version|ARK-123" content/skills/flow-build/references/verification.md` no devuelve nada; `rg -l "verification.md#human-ui-handoff" content/` no devuelve nada; `rg -n "flow-walkthrough" content/skills/flow-build/SKILL.md content/agents/review/hive-review-ux.md` devuelve la línea 75 y el paso 3 (AC2).
- `rg -n "not recorded" content/agents/quality/hive-verify-change.md` devuelve el paso con el tope de dos fallos (AC3).
- Enlaces: `go test ./tests/content/...` y `go test ./tooling/...` pasan.
- Comprobación acotada en esta máquina de los comandos que la skill nombra, con `T=timeout` o `gtimeout`, o el `perl -e 'alarm shift; exec @ARGV'` de `browser-automation.md:27`, `S=walkthrough-check` y `OUT=<scratch absoluto>/check.mp4`:
  ```
  $T 45 agent-browser --session $S open 'data:text/html,<h1>check</h1>'
  $T 45 agent-browser --session $S record start "$OUT" --cursor
  $T 45 agent-browser --session $S wait 3000
  $T 45 agent-browser --session $S record stop
  ffprobe -v error -show_entries format=duration,size -show_entries stream=codec_name -of default=noprint_wrappers=1 "$OUT"
  $T 45 agent-browser close --session $S && agent-browser session list
  ```
  Esperado: `duration` mayor que 1, `codec_name=h264`, y la sesión ausente de la lista. Registrar el resultado en la sección de progreso; si `record` falla bajo `--session`, ajustar DS2 antes de cerrar la tarea.

### T2 — Esqueleto HTML de recorrido en `flow-report`

- [x] Existe `content/skills/flow-report/assets/skeleton-walkthrough.html` conforme a DS3, pasa `self_check.py`, y `flow-report/SKILL.md` lo lista, acota la exclusión de handoffs y exceptúa los archivos hermanos del recorrido.

**Closes:** AC4.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-report/assets/skeleton-walkthrough.html` (nuevo); `content/skills/flow-report/SKILL.md:10,25` y su tabla de activos; escritura permitida además en `<scratch>/sample/` (página de muestra) y en `_support/workspace/2026-10-17-gh-67-flow-walkthrough-skill/images/` (una captura).

**Execution:** delegada a `hive-build-frontend`, con DS3, `skeleton-review.html` como base de estilo, las reglas de `self_check.py` (sin referencias remotas, sin atributos `on*`, rutas relativas permitidas) y la convención de marcadores de los esqueletos existentes.

**Test approach:** check (`self_check.py` sobre el esqueleto y una página de muestra; comprobaciones estáticas).

**Changes:** DS3. Solo un bloque de tokens, `color-scheme: dark`, fondo explícito en `body`, sin `prefers-color-scheme` ni `data-theme`; bloque `@media print` en claro conservado; `<video controls preload="metadata">` y `<img>` con `src` y `poster` relativos a `images/`; ampliación de captura con `addEventListener`.

**Verification:**
- `python3 content/skills/flow-report/scripts/self_check.py content/skills/flow-report/assets/skeleton-walkthrough.html` sale con 0; `rg -c "data-theme|prefers-color-scheme" content/skills/flow-report/assets/skeleton-walkthrough.html` da 0; `rg -n "color-scheme: *dark|@media print|<video|<img" <archivo>` devuelve cada patrón; `rg -n '\bon[a-z]+=' <archivo>` no devuelve nada (la forma sin borde de palabra coincide con `content=` del meta viewport) (AC4).
- Página de muestra: copiar el esqueleto a `<scratch>/sample/walkthrough.html`, generar `<scratch>/sample/images/login.png` (`ffmpeg -f lavfi -i color=c=gray:s=640x360 -frames:v 1 <scratch>/sample/images/login.png`) y `<scratch>/sample/images/login.mp4` (`ffmpeg -f lavfi -i color=c=black:s=640x360:d=3 -c:v libx264 -pix_fmt yuv420p <scratch>/sample/images/login.mp4`), sustituir los marcadores, correr `self_check.py` sobre ella (sale con 0) y abrirla con `agent-browser` en una sesión propia para comprobar con `eval` que `document.querySelector('video').readyState >= 1` tras `load` y que `getComputedStyle(document.documentElement).colorScheme` es `dark`; una captura de la página de muestra a 1440 px va a `_support/workspace/2026-10-17-gh-67-flow-walkthrough-skill/images/`; cerrar la sesión.
- `rg -n "walkthrough" content/skills/flow-report/SKILL.md` devuelve la fila, la línea 10 acotada y la excepción de la línea 25 (AC4).

### T3 — Ciclo de vida de grabaciones y comprobación de carga

- [x] `artifact-layout.md` y `flow-close/SKILL.md` tratan página, grabaciones e imágenes como unidad de retención; `tests/pilot/flows.go`, `flows_test.go`, `cases.json` y el README de fixtures nombran `flow-walkthrough`.

**Closes:** AC5, AC7.

**Depends on:** ninguna.

**Locations:** `content/skills/workspace-conventions/references/artifact-layout.md:23–25`; `content/skills/flow-close/SKILL.md:78`; `tests/pilot/flows.go:550`; `tests/pilot/flows_test.go:756`; `tests/fixtures/flows/cases.json:33`; `tests/fixtures/flows/README.md:7`.

**Execution:** hilo principal: seis ediciones de una o dos líneas.

**Test approach:** check.

**Changes:** DS3 (retención) y DS4 (puntos tres y cinco). Añadir `flow-walkthrough` al final de la lista de Go, de la prueba, del prompt y de la frase del README.

**Verification:**
- `rg -n -i "recording" content/skills/workspace-conventions/references/artifact-layout.md content/skills/flow-close/SKILL.md` devuelve al menos una línea en cada archivo con la unidad de retención (AC5).
- `go test ./tests/pilot/...` pasa y `rg -l "flow-walkthrough" tests/` devuelve los cuatro archivos (AC7).

### T4 — Documentación del sitio

- [x] Las FAQ de `flows/build.md` y `flows/close.md` describen la grabación y la página; el párrafo de ajustes declarados de `flows/build.md` nombra `UI verification`; `index.mdx` lista la skill.

**Closes:** AC6, AC11 (parte del sitio).

**Depends on:** T1, T3, T5 (describe su comportamiento final).

**Locations:** `site/src/content/docs/flows/build.md:41,73`; `site/src/content/docs/flows/close.md:54`; `site/src/content/docs/index.mdx` ("Supporting skills").

**Execution:** delegada a `hive-write-docs`, con las fuentes T1, T3, T5 y DS1–DS3, DS6, la regla del sitio de describir sin copiar reglas y enlazar `SKILL.md` en `master`, y el estilo de las entradas existentes.

**Test approach:** check.

**Changes:** DS4 (punto cuatro). Una entrada por lista; cada FAQ en una o dos frases.

**Verification:**
- `rg -n "flow-walkthrough" site/src/content/docs` devuelve `build.md` e `index.mdx`; `rg -n -i "recording" site/src/content/docs/flows/close.md` devuelve la FAQ de qué borra; `rg -n "UI verification" site/src/content/docs/flows/build.md` devuelve el párrafo de ajustes (AC6, AC11).
- `pnpm -C site build` sale con 0 (`node_modules` ya está instalado; si faltara, `pnpm -C site install --frozen-lockfile`).

### T5 — Ajuste `UI verification` en la guía

- [x] `verification.md:61` (desde T7, el párrafo de `ui-gate.md`) y `flow-build/SKILL.md:36` seleccionan los hijos por valor; `delivery-decisions.md` y `flow-build/SKILL.md:20` preguntan el valor en la ronda de revisión cuando falta y reutilizan la respuesta; `flow-plan/SKILL.md:18` lo nombra en el traspaso; `plan-format.md:121` y la puerta del candidato aceptado de la skill siguen el valor.

**Closes:** AC9.

reopened by revision R1-A (2026-10-18): AC9 cambia de archivo el párrafo de la puerta; se re-verifica tras T7, sin implementación nueva.

**Depends on:** T1 (misma escritora en `verification.md`, `flow-build/SKILL.md` y la skill; evita escrituras solapadas) y T6 (la clave queda definida en `project-settings.md` antes de que las puertas la nombren).

**Locations:** `content/skills/flow-build/references/verification.md:61`; `content/skills/flow-build/SKILL.md:20,36`; `content/skills/flow-plan/references/delivery-decisions.md` (párrafo de la pregunta de revisión y su ronda); `content/skills/flow-plan/SKILL.md:18`; `content/skills/flow-plan/references/plan-format.md:121`; `content/skills/flow-walkthrough/SKILL.md` (sección del bucle de validación, puerta del candidato aceptado).

**Execution:** hilo principal, por el acoplamiento con T1.

**Test approach:** check.

**Changes:** DS6. En `verification.md:61`, conservar la exención de texto o estilo, el brief de cada hijo, la parada de dos fallos y el reporte de sandbox; sustituir "requires two independent children … need no separate approval" por la selección por valor con un puntero a `delivery-decisions.md` para la pregunta. En `flow-build/SKILL.md:36`, sustituir "are required, not proposed" por la selección por valor. En `plan-format.md:121`, "include the independent … gates" pasa a incluir los hijos que el valor seleccione. En la skill, la puerta del candidato aceptado lanza solo los hijos del valor. En `delivery-decisions.md`, la pregunta de `UI verification` va en el mismo párrafo que la de revisión: una vez por repositorio y sesión, opciones con efecto y costo, Recomendada por superficie, la opción dice que la respuesta vale el resto de la sesión en ese repositorio y se puede cambiar en cualquier petición, declarada en la línea única.

**Verification:**
- ``rg -n '`UI verification`' content/skills/flow-build/references/verification.md content/skills/flow-build/SKILL.md content/skills/flow-plan/references/delivery-decisions.md content/skills/flow-plan/SKILL.md content/skills/flow-plan/references/plan-format.md content/skills/flow-walkthrough/SKILL.md content/skills/flow-walkthrough/references/ui-gate.md`` devuelve al menos una línea por archivo (desde T7; `ui-gate.md` con la selección por valor); `rg -n "requires two independent children" content/skills/flow-build/references/verification.md content/skills/flow-walkthrough/references/ui-gate.md`, `rg -n "are required, not proposed" content/skills/flow-build/SKILL.md` y `rg -n "not a pending choice" content/skills/flow-plan/references/plan-format.md` no devuelven nada (AC9).
- `go test ./tooling/cli` pasa.

### T6 — Clave `UI verification`: referencia de ajustes y manager

- [x] `project-settings.md` documenta la clave y la añade al ejemplo; `hive doctor` y el formulario de proyecto aceptan `UI verification` con `full`, `functional` y `none` y avisan ante otro valor; `deployment-manager.md` lo describe.

**Closes:** AC8, AC10, AC11 (parte del manager).

**Depends on:** ninguna. La referencia y el manager van en la misma tarea porque dos pruebas los atan en ambos sentidos: `TestHiveSettingKeysAppearInTheGuidance` (`doctor_project_test.go:421–440`) exige que cada clave de `hiveSettingKeys` aparezca entre acentos graves en la sección "Declare settings" de `project-settings.md`, y `TestHiveSettingKeysCoverTheGuidanceExample` (`:441–460`) exige que el ejemplo use solo claves conocidas.

**Locations:** `content/skills/flow-plan/references/project-settings.md:9–21`; `tooling/cli/doctor_project.go:27–38` (tabla, `UI verification` al final para no mover los conteos de teclas del TUI), `:354–360` (`case` nuevo en `hiveFindings` con mensaje vía `quotedList`), `:472` (`hiveUIVerificationValues`); `tooling/cli/tui_project_view.go:267` (`projectLabelCol`, hoy 16, dimensionada para "Hive guidance"; "UI verification" tiene 15 caracteres), `:272–274` (comentario de `projectForm`), `:313` (`case "UI verification"` propio con sus valores), `:552` (`padRight`); `tooling/cli/doctor_project_test.go` (tabla de valores; lista de opcionales de `TestHiveSettingKeysRequiredSet` en `:470`); `tooling/cli/tui_project_view_test.go` (ciclo de opciones del campo nuevo y las aserciones de relleno `"Delivery      ‹ none ›"` en `:356,364,395`, que cambian al ensanchar la columna); `tooling/cli/project_write_test.go` (caso de `checkHiveText` con `full`); `_support/docs/architecture/deployment-manager.md:77` (viñeta Project).

**Execution:** delegada a `hive-build-backend`, con DS6 (texto de la clave para `project-settings.md`: valores, efecto, regla sin clave, aviso sobre `automatic`, línea del ejemplo) y el commit `9ec88c4e` como patrón.

**Test approach:** tdd (prueba de tabla para los tres valores válidos y uno inválido en `doctor_project_test.go`, el ciclo de opciones del campo nuevo en `tui_project_view_test.go` y el caso de `checkHiveText`, en rojo antes del cambio).

**Changes:** DS6 "Clave y valores", "Sin clave" (solo la definición; la pregunta la escribe T5) y "Manager".

**Verification:**
- ``rg -n '`UI verification`' content/skills/flow-plan/references/project-settings.md`` devuelve la definición, el aviso sobre `automatic` y la línea del ejemplo (AC8).
- `go test ./tooling/cli` pasa con las pruebas nuevas, incluidas las dos que atan la tabla con la referencia; `rg -n "UI verification" tooling/ _support/docs/architecture/deployment-manager.md` devuelve la tabla, los valores, las pruebas y la documentación (AC10, AC11).
- `go build -o /dev/null ./tooling/cli` sale con 0 (compila el binario `main`; `go vet ./...` va en la puerta compartida).

### T7 — Casa única de la puerta de UI (revisión R1-A)

- [x] Existe `content/skills/flow-walkthrough/references/ui-gate.md` conforme a DS7; `ui-review-criteria.md` vive en `flow-walkthrough/references/`; `verification.md` conserva solo el puntero, `browser-automation.md` y las puertas que no son de UI; la skill remite a la referencia para la puerta del candidato aceptado; los consumidores enlazan la nueva casa.

**Closes:** AC12.

**Depends on:** T1 y T5 (mueve texto que ellas escribieron).

**Locations:** `content/skills/flow-walkthrough/references/ui-gate.md` (nuevo); `content/skills/flow-walkthrough/references/ui-review-criteria.md` (movido con `git mv` desde `content/skills/flow-build/references/ui-review-criteria.md`); `content/skills/flow-walkthrough/SKILL.md:3,8,12–14,50`; `content/skills/flow-build/references/verification.md:33,57–61`; `content/skills/flow-build/SKILL.md:36,71`; `content/skills/flow-plan/SKILL.md:48`; `content/skills/flow-plan/references/plan-format.md:121`; `content/skills/flow-plan/references/ui-planning.md:32`; `content/skills/flow-build/references/browser-automation.md:3,20`; `content/skills/flow-plan/references/delivery-decisions.md:36`; `content/agents/review/hive-review-ux.md` pasos 2 y 3 (locators `skill:flow-walkthrough/references/...`).

**Execution:** hilo principal: movimiento acoplado entre doce archivos; el brief sería más largo que el cambio.

**Test approach:** check.

**Changes:** DS7. Mover el texto sin cambiar su significado; donde una frase movida decía "this reference" o "below", enlazar o nombrar la casa que corresponda. La referencia dice en su primera línea cuándo leerla.

**Verification:**
- `test -f content/skills/flow-walkthrough/references/ui-gate.md && test -f content/skills/flow-walkthrough/references/ui-review-criteria.md && test ! -e content/skills/flow-build/references/ui-review-criteria.md` (AC12).
- `rg -n "Brief the reviewer|scoped to the diff|pattern conformity" content/skills/flow-build/references/verification.md content/skills/flow-walkthrough/SKILL.md` no devuelve nada; `rg -n "Brief the reviewer|scoped to the diff|every delivery mode" content/skills/flow-walkthrough/references/ui-gate.md` devuelve cada término (AC12).
- `rg -n "flow-build/references/ui-review-criteria" content/ site/ _support/docs/architecture` no devuelve nada; `rg -n 'ui-review-criteria\.md' content/` devuelve solo líneas de `flow-walkthrough/references/` (relativas) o con `../../flow-walkthrough/references/` o `skill:flow-walkthrough/references/`; `rg -l "ui-gate.md" content/` devuelve `verification.md`, `flow-build/SKILL.md`, `flow-plan/SKILL.md`, `plan-format.md`, `ui-planning.md`, `browser-automation.md`, `delivery-decisions.md`, `hive-review-ux.md` y `flow-walkthrough/SKILL.md` (AC12).
- `go test ./tests/content/... ./tests/pilot/... ./tooling/...` pasa; `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa. Ninguna prueba valida los enlaces de la release sobre el contenido real: tras el commit, `hive update --dry-run` desde la rama sale con 0 y su plan incluye `flow-walkthrough/references/ui-gate.md` y `ui-review-criteria.md`.
- Tamaños: `wc -w` da `verification.md` por debajo de 3.250, `ui-gate.md` entre 600 y 850 y la skill por debajo de 1.800.
- Después, `hive-verify-task` re-verifica T5 (AC9) con el párrafo en su nueva casa.

## Verificación y revisión humana

| Puerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad / condición pendiente |
| --- | --- | --- | --- |
| Pruebas del repositorio | Candidato final, local | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`, `pnpm -C site build` pasan | Implementación |
| Tamaños | Candidato final | `wc -w content/guidance/global.md content/skills/flow-build/references/verification.md content/skills/flow-walkthrough/SKILL.md`: `global.md` sin cambio (4.563 palabras en la base), `verification.md` por debajo de 3.750, la skill entre 1.300 y 1.800 (revisado en la construcción: el texto movido ya pesa 798 palabras y la grabación, la página y la puerta por valor añaden el resto; el objetivo original de 1.000 a 1.300 no cabía sin recortar significado); `go test ./tests/content/... -run 'Size|Footprint' -v` registra la unión del ciclo plan-build antes y después; tras T7, `verification.md` por debajo de 3.250 y `ui-gate.md` entre 600 y 850 | Implementación |
| Verificación por tarea | Al reportar cada tarea | `hive-verify-task` con la carpeta del cambio, el ID y la base de diff, como fija `flow-build` | Implementación |
| Verificación de conjunto | Candidato final | `hive-verify-change` sobre AC1–AC11 con los comandos de cada criterio; sin UI de aplicación: no aplican `hive-review-ux` ni recorrido en vivo | Implementación |
| Despliegue local | Solo cuando lo pidas | `go build -o "$(command -v hive)" ./tooling/cli` y después `hive update` desde la rama de prueba; nueva sesión para cargar la skill | Pendiente tu petición |
| Uso real | Después del despliegue | Dos sesiones reales con un cambio de UI: respuesta o clave de `UI verification`, tiempo de la ronda del verificador con grabación, si abriste el video y la página, y qué faltó en la tabla | Pendiente tu uso |

No hay interfaz de aplicación en este cambio: la página de muestra de T2 es un activo de la skill y su comprobación va en T2.

## Entrega

Pendiente E1 y E2 (ver `proposal.md`, sección Entrega). Esta carpeta se versiona en el cierre.

## Estado de revisión y progreso

- 2026-10-18, revisión 1 (hashes `1de995c9…`, `353e43d7…`, `afb25258…`): dos `hive-review-plan` (guía y punteros; activo HTML y evidencia), unos 126 k y 127 k tokens, 136 s y 172 s. Aplicado: A-B1 (puntero sin los términos de AC2 y búsqueda acotada), A-B2 (AC2 distingue enlaces con y sin ancla), A-B3 (la pasada extra queda como excepción explícita decidida en D2-A; V1-A reetiquetada como inferencia), A-B4 (tope de `verification.md` a 3.750 con el recuento real de 798 palabras), A-N1 (T1 depende de T2), A-N2 (Deliver apunta a `questions.md:54`; lugares revisados listados en DS4), A-N3 (obsolescencia de grabaciones), A-N4 (primera vista en la descripción), B-B1 (unidad de retención página + archivos), B-B2 (línea 25 de `flow-report`), B-N1 (impresión en claro y comprobación estática), B-N2 (`cases.json`), B-N3 (`llms.txt` fuera), B-N4 (pnpm), B-N5 (comprobación de T1 reescrita), B-N6 (precondiciones), B-N7 (motivo de `hive-review-ux`), B-N8 (T2 sin la suite de `tests/skills`, sin `on*`, rutas de escritura ampliadas). Menores: AC3 busca "not recorded"; "Carvento" fuera; T4 depende de T1, T3 y T5; enlaces sin ancla.
- 2026-10-18, revisión 2 (hashes `62188b14…`, `2a6fe19b…`, `d3c5aacd…`): P1-A y P2-A añaden DS6, T5, T6 y AC8–AC11. Re-revisión de guía (unos 140 k tokens, 42 s): R1 (falsificador de AC9 verdadero en la base: ahora busca la clave con acentos graves y las tres frases que desaparecen), R2 (`plan-format.md:121` añadido a T5), R3 (puerta del candidato aceptado por valor, en DS6 y T5), R5 (la opción dice que la respuesta vale la sesión) aplicados como los propuso el revisor; R4 (quién graba con `none`) queda como decisión Q3 para el usuario. Confirmó bien aplicadas las correcciones de la revisión 1. Re-revisión del manager (unos 155 k tokens, 95 s): B1 (T5 y T6 se exigían mutuamente por dos pruebas: la definición de la clave y el ejemplo pasan a T6 con el manager; AC8 pasa a T6; filtro de pruebas `go test ./tooling/cli`; lista de opcionales de `TestHiveSettingKeysRequiredSet`) y B2 (columna del TUI, `case` propio, comentario, posición de la clave, aserciones de relleno, `checkHiveText`) aplicados como los propuso; rutas absolutas en los `ffmpeg` de T2. Confirmó bien aplicadas las correcciones de la revisión 1. Plan listo salvo Q3 (quién graba con `none`), que va al usuario con la entrega.
- 2026-10-18, construcción: E1 (merge previo de la rama de prueba: PR #68 y cierre #69), E2-A, Q3-A, R1-A. Base `b73bb16e`. T3 verificada por `hive-verify-task` (AC5, AC7 cumplidos) y en commit `ec732b46`. T2 entregada por `hive-build-frontend` (la comprobación de atributos `on*` usa `\bon[a-z]+=` porque `content=` del meta viewport coincide con la forma sin borde). T1: la skill ya escribe la puerta del candidato aceptado por valor de `UI verification` (T5 la verifica en lugar de reescribirla); `verification.md` queda en 3.630 palabras y la skill en 1.722; comprobación de grabación en esta máquina: `record start` bajo `--session` y envoltorio `perl alarm 45` produce `check.mp4` con `codec_name=h264`, `duration=3.23`, `size=7241`; la sesión cerró. Modelos: implementador Claude (fable-5-1, observado por el host); verificadores y constructores en sus perfiles configurados, no observados.
- 2026-10-18, verificaciones: T2 cumplida (AC4; encabezado de flujo normalizado a guion simple; el comando `on*` del plan corregido con `\b`); T1 cumplida (AC1–AC3; se restauraron dos matices del texto movido: ofrecer capturas sin que las pidan y "pila verificada"); T6 cumplida (AC8, AC10, AC11 manager; el formulario muestra la opción vacía como `unset` porque `none` es un valor real; la referencia aclara que con `none` corre un hijo solo para grabar, Q3-A). Commits: T3 `ec732b46`, T2 `6a1f0b86`, T1 `1ffa41c3`, changelog `d70e0ad8`, T6 `247523fc`, T5 `ee537a66` (AC9 cumplido; tres retoques de coherencia aplicados), T4 `955a204f` (AC6 y AC11 sitio cumplidos; tres frases acotadas a lo que dicen las fuentes).
- 2026-10-18, puertas compartidas sobre el candidato final: `go vet ./...` limpio; `go test ./...` 18 paquetes en verde; pruebas Python de skills 23 en verde; `pnpm -C site build` sale con 0 (solo el aviso previo de la colección `i18n`); tamaños: `global.md` 4.563 (sin cambio), `verification.md` 3.690, skill 1.741. Verificación de conjunto por `hive-verify-change`: AC1–AC11 cumplidos, cada falsificador comprobado también contra la base; 167 enlaces resuelven; changelog presente. Defecto menor corregido después (commit siguiente): `plan-review.md:16` y `flow-build/SKILL.md:71` todavía llamaban "required" a los dos hijos. Hallazgo incidental: `TestGetHiveRestoresThePreviousPackageOnSignals` y dos pruebas hermanas de `tooling/distribution` fallaron una vez bajo carga (`exit 1, want 129`) y pasaron aisladas; el paquete no cambia en esta rama; salida guardada en `_support/workspace/2026-10-17-gh-67-flow-walkthrough-skill/go-test-distribution-signals-flaky-2026-10-18.txt`; sin duplicado en el tracker.
- 2026-10-18, pausa: trabajo entregado en la rama a la espera de tu prueba en sesiones reales y tu decisión de merge. X1-A: binario recompilado y despliegue local desde `c94072f4` aplicado (resultado `9372e51d0be2`); las sesiones nuevas cargan la skill. X2-A: las pruebas intermitentes de `tooling/distribution` quedan como vigilancia en memoria con conteo 1. La lista `session list` de `agent-browser` tarda unos segundos en reflejar un `close`: comprobarla tras una pausa breve.
- 2026-10-18, D1-A: la frase heredada "offer them even when the user did not ask" (skill, línea 30) chocaba con la regla de ofertas de `d54f3a75` (PR #64, ya en la base): ahora dice "Show or link them in the handoff even when the user did not ask". Commit `f70d2353`; pruebas de contenido, pilot y management en verde; despliegue local desde `f70d2353` aplicado (resultado `9a15eb745ad0`). La nota de T1 sobre "ofrecer capturas sin que las pidan" queda superada por esta línea.
- 2026-10-18, revisión R1-A: el usuario preguntó si toda la revisión de UX/UI se había movido a la skill; no: la puerta independiente seguía en `verification.md:57–61` y la skill restataba la selección por valor en su puerta del candidato aceptado. Eligió R1-A (consolidar en este cambio). Añade DS7, AC12 y T7; T5 vuelve a `[?]` porque AC9 cambia de archivo. Invalida la frase de DS1 "las comprobaciones independientes se quedan en `verification.md`" y la casa `verification.md:61` de DS6. Re-revisión de guía (`hive-review-plan`, unos 120 k tokens, 148 s): un bloqueante, el enlace relativo de `browser-automation.md:20` a los criterios, que las listas no recogían y que ninguna prueba de Go detecta porque la validación de enlaces de la release no corre sobre el contenido real; cuatro no bloqueantes: forma `skill:` de los locators de `hive-review-ux`, la re-verificación de AC9 debía incluir `ui-gate.md`, tres frases de la línea 61 que DS7 no nombraba, y el alcance "todo modo de entrega" limitado a las partes 1–3. Los cinco aplicados como los propuso el revisor (sin ronda nueva): DS7, AC12, T5 y T7 actualizados; la comprobación pasa a ser `rg` de toda forma del nombre más `hive update --dry-run` tras el commit. Confirmó AC12 falsable en la base y sin cláusula ya cierta, la carpeta `references/` nueva aceptada por `validateRelease`, y sin consumidores en `tests/pilot`, el sitio ni `agent-delivery.md`.
- 2026-10-18, T7 construida en el hilo principal y en commit `13c1ba91`: `ui-gate.md` 843 palabras (cuatro partes de DS7, con las tres frases nombradas), `ui-review-criteria.md` movido con `git mv` y cabecera repuntada, `verification.md` 3.209 palabras con el puntero, la skill 1.718 con la puerta del candidato aceptado como puntero, nueve consumidores repuntados (incluido `browser-automation.md:20`). Comprobaciones de T7 en verde salvo las delegadas: `go vet`, suites de `tests/content`, `tests/pilot` y `tooling/...` (las de `tooling/distribution` pasaron esta vez), 23 pruebas Python; `hive update --dry-run` desde `13c1ba91` salió con 0 ("Preview: installation was not changed"), así que la validación de enlaces de la release acepta la referencia nueva y el archivo movido. Verificación independiente de T7 (AC12) y re-verificación de T5 (AC9) lanzadas en paralelo a `hive-verify-task`.
- 2026-10-18, veredictos: T5 (AC9) cumplido con el párrafo en `ui-gate.md` (unos 84 k tokens, 126 s); T7 (AC12) cumplido, ninguna frase perdida en la comparación frase por frase (unos 102 k tokens, 197 s). Dos observaciones menores aplicadas: la línea 50 de la skill resumía tres de los cinco pasos de la puerta del candidato aceptado y queda como puntero puro (commit `f63abf50`); el plan decía 800 en dos sitios y 850 en otros dos para `ui-gate.md`, unificado en 850 (843 reales). Observación del verificador: `plan-format.md:121` repetía el mapeo valor → hijos; no venía de `development` (allí la línea nombra los dos hijos sin valores) sino de T5 en esta misma rama, así que se corrige aquí: la plantilla nombra los hijos que el valor selecciona, resueltos como fija `delivery-decisions.md`, sin el paréntesis. `flow-plan/SKILL.md:48` solo nombra los dos roles que el valor puede seleccionar, sin mapeo. Sin ramas huérfanas: ninguna rama local ni remota sin fusionar aparte de esta, sin stash, sin worktrees, sin PR abierto.
- 2026-10-18, fuera del alcance de gh-67 pero en la misma rama de prueba (regla de rama única): commit `76408ee7`, regla de idioma de `global.md` con "URL paths and route segments including those of public pages" y "a URL path is an identifier, not user-facing text" (4.583 palabras, 29.928 bytes, +20 palabras), con entrada en el changelog y despliegue local aplicado (resultado `dc28266103`). Origen: sesión de cam `ba5737f3` (Claude Code 2.1.302, fable-5-1, 08:32Z) que eligió slugs públicos en español; decisión R1-A del usuario. El PR de la rama lo lista como cambio independiente con su lugar revisado: `global.md` es la única casa de la regla (ningún otro archivo de `content/`, `site/` ni `_support/docs` la restata).
- 2026-10-18, también fuera del alcance de gh-67, tras la auditoría de la sesión de cam y el research completo sobre convenciones: `ae610f1f` (A3: `global.md` lee `browser-automation.md` antes de conducir un navegador con cualquier fin; 4.603 palabras) y `eb2a40c1` (R2-B: `hive-review-code` reporta identificadores, nombres de archivo, rutas o segmentos en otro idioma como hallazgo; `hive-verify-task` los anota junto a los veredictos), con changelog; despliegue local desde `eb2a40c1` aplicado (resultado `77fc485c0114`). Issue [#70](https://github.com/JhonHawk/tricell-hive/issues/70) abierto (A1: tarjeta sin mensaje previo, conteo 3). A2 (forma de los secretos) declinado: queda como vigilancia con conteo 1. El PR lista estos commits como cambios independientes; lugares revisados: `global.md` (único puntero nuevo), los dos roles (sin otra casa de la comprobación), `site/` sin cambio porque no describe los roles.
- 2026-10-18, también independiente (L1): la regla de idioma de `global.md` incluye los rótulos que el host muestra durante el trabajo (descripción de comando, asunto de tarea, etiqueta de hijo) y, para contribuciones a repos externos u open source, manda comprobar primero el idioma del repo para lo que llega a él. Evidencia: pnpm `fac7fdbe` (opus-5-5, 86/86 descripciones en inglés), cam `ba5737f3` (233/234) y cam `52e62938` (35/36). Commit y despliegue local anotados en la memoria de estado.
- 2026-10-18, independiente (W1-B, tras auditar la sesión de pnpm `fac7fdbe`): `df6e550e`, la skill `unattended-delegation` nombra cuatro fuentes de activación, incluida el silencio del usuario tras responder ruta, entrega y revisión; admite force-push con lease exacto y ediciones de texto de PR con su vuelta atrás; obliga a declarar una activación inferida como inferida. `global.md` (4.699 palabras) restata el invariante, `index.mdx` lo describe, changelog en "Changed". Despliegue local desde `df6e550e` (resultado `80f2781dff91`). Lugares revisados: `git-workflow/SKILL.md:24,42` (la excepción vive en la skill y la nombra), `delivery-decisions.md` (sin cambio: la pregunta de entrega sigue siendo previa), `tests/pilot` (solo nombra la skill). Changelog actualizado en `dc8f2237`. Despliegue local desde `f63abf50` aplicado (resultado `372ba5fd2f6c`); `~/.agents/skills/flow-walkthrough/references/ui-gate.md` y `ui-review-criteria.md` instalados y el antiguo `flow-build/references/ui-review-criteria.md` retirado. Vuelve la pausa: tu prueba en dos sesiones reales y tu decisión de merge.
- 2026-10-18, independiente, en la misma rama de prueba: `236c7c75`, `questions.md` elimina la exención de las preguntas sí/no (toda autorización lista dos opciones con ID y efecto) y dice que una herramienta que devuelve antes de la respuesta, como la tarjeta asíncrona de Codex, nunca sustituye a las opciones en texto; origen: sesión Codex de carvento `01a15146` (D5 Bugbot y D6 commit fixture sin opciones).
- 2026-10-18, revisión D3-A (sesión Claude Code de cam `e2c8d39f`: V1 en texto repetida cuatro veces mientras corrían hijos): la primera vista rendida (skill, línea 24) y la entrega del recorrido (línea 42) usan la tarjeta nativa que espera también con hijos corriendo; sus resultados esperan la respuesta y no se repite nada por ellos. `design.md` paso 5 actualizado. Ninguna AC ni tarea cambia de estado: T1 y T3 ya verificadas conservan su evidencia, el cambio es una frase de cada párrafo.
