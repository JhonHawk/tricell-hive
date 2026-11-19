# skill-leaves — Reglas en hojas pequeñas, cargadas en su momento

| Campo | Valor |
| --- | --- |
| Estado | Borrador · plan escrito, pendiente de revisión, entrega y aprobación |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `2bb96478` · rama de prueba `feat/question-card-alone` en `6adebe35` (PR #491) · rama propuesta: `feat/skill-leaves` (PR 1, proceso) y `feat/skill-leaves-principles` (PR 2, principios) |

## Decisions

2026-11-18
- **D3 B** — ¿Con qué mecanismo cargamos las hojas pequeñas? Como skills sueltas, al estilo de pstack: cada hoja es una skill que el modelo no carga por su cuenta y que un índice nombra con su condición de lectura.
- **D4 B** — ¿Qué hacemos con las reglas de preguntar mientras el trial del PR #491 sigue abierto? Cerrar el trial con la evidencia medida, integrar el #491 y construir la hoja de preguntar sobre esa regla. El criterio de muestra del trial se rebaja a lo medido.
- **D5 A** — ¿Entran ya los principios de código y `typescript-best-practices`? Sí, como fase 2 del mismo plan, en un segundo pull request sobre la estructura ya integrada.
- **D6 A** — ¿Cómo cubrimos los hosts que no pueden ocultar las hojas? Un interruptor por host: `disable-model-invocation` en el frontmatter, un `agents/openai.yaml` generado por el manager para Codex y, en Antigravity, que el manager no registre las hojas en su catálogo. El usuario pidió validar contra las referencias: ninguna oculta por host (ver `design.md`).
- **D7** — ¿Cuánto sacamos de `global.md`? El usuario amplió el alcance: también se sacan hojas de las skills `flow-*` y de `git-workflow`, y se adoptan los principios de pstack que se recomienden.
- **D8 A** — ¿Cómo se entregan los PR? En modo `interactive`: parada antes de cada push para que el usuario valide, y luego push, PR y merge.
- **D9 A** — ¿Qué revisión de código corre? La nativa de Claude Code, `/code-review`, antes de la validación.
- **D10 A** — ¿Se nombra el principio usado? Sí, en una línea al final del informe de build y de la revisión de código, y solo los principios leídos que cambiaron una decisión.
- **D11** — ¿Qué tamaño máximo tiene una hoja? El usuario adopta los playbooks de pstack y convierte `flow-research` en el router `hive-mode`. Tamaños, como supuesto a confirmar en el resumen del plan: hojas de 4 KB como máximo, con excepciones de hasta 7 KB nombradas en el test, y playbooks de 7 KB como máximo.
- **D12 A** — ¿Esperamos una hoja canario? Sí. `hive-rule-git-commit` se publica sola, y la extracción masiva espera a que una sesión real de Claude Code y otra de Codex la lean.
- **D13 A** — ¿Cómo se carga `hive-typescript`? Visible, con `paths: ["**/*.ts", "**/*.tsx"]`.
- **D14 A, B, C y D** — ¿Qué suma `flow-verify` de pstack? Los cuatro complementos: `no-comments`, `blast-radius`, `unslop` con `technical-writing`, e `interrogate`. Este último gasta cuota de otro proveedor, así que entra como opción de la pregunta de revisión.
- **D15 A** — ¿`hive-mode` pasa a `flow-build` en el mismo turno ante un cambio corto y claro? Sí, como hoy research, sin que `hive-mode` edite nada.
- **D16 A** — ¿`flow-verify` se mantiene como etapa? Sí, delgada, como puerta común: la suite, el recorrido de criterios, los complementos, las revisiones, tu validación, el PR y el merge. Corre la prueba que dicta cada playbook.
- **D17 A** — ¿Cómo se limita `interrogate`?
  - Se pregunta en cada trabajo y su respuesta no se reutiliza.
  - Una corrida por candidato y otra tras la ronda de correcciones, como máximo.
  - Corre además de `/code-review`.
  - En trabajo desatendido, solo con un monto autorizado antes.
- **D19 A** — ¿Los efectos externos del trabajo desatendido van en el mismo registro que las decisiones? Sí. `decisions.tsv` lleva una fila por decisión y una por efecto externo: el objetivo y el cambio en la evidencia, y cómo revertirlo en el resultado. El registro de la tarea queda solo con el avance.
- **Compatibilidad (2026-11-18)** — El usuario pide considerar la compatibilidad con los modos de entrega. Se agregan la matriz de `design.md` y AC17.
- **Tarea añadida (2026-11-18)** — Pasan a la fase 2 tres prácticas de traspaso de pstack: el registro TSV de decisiones, las casillas por PR y lo que no está escrito al pausar. El usuario respondió «va» a la propuesta.
- **D18 A** — ¿Cómo se publica la hoja canario? Se instala localmente desde el commit del PR 1 en su worktree, sin merge. `development` no cambia hasta el PR 1 completo.
- **Supuestos.**
  - Se suman los playbooks `refactor` y `perf`, adaptados de pstack.
  - Las etapas siguen siendo comandos y el playbook solo las ordena.
  - Los 5 principios fundidos se integran como líneas en sus hojas, no como sección.
  - El flujo queda así: `hive-mode` → [`flow-figma-design`] → [`flow-plan`] → `flow-build` → `flow-verify` (con el PR) → [`flow-deploy`] → `flow-close`.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «lo que haremos con los flujos es apegarnos un poco más a lo que hace pstack, pero no en la verificación en sí, sino en la estructura de las skills […] ahorita, nuestras skills son ENOOOOOORMES, y siento que algunas veces tienen conflictos por condensar tantas reglas y variaciones» (2026-11-18)

> «no solo es sacar de global, sino también de las skills de flow-*, y adoptar las que me recomiendes de pstack» (respuesta a D7, 2026-11-18)

## Objective

Hoy `global.md` pesa 52.8 KB y se carga en cada sesión. Tres skills rozan el tope de 16 KB, y cada regla que falló en las sesiones reales se repite en 5 a 15 archivos. El análisis de 79 sesiones del 2026-11-11 al 18 mostró cuatro fallas recurrentes:

1. La sección de preguntar, densa y en conflicto con el host.
2. Reglas que no están cargadas cuando hacen falta: `git add -A` en 8 sesiones, porque `flow-build` no enlaza `git-workflow`.
3. Copias de una regla que se contradicen: unas 10 por cambio de guía.
4. Huecos entre líneas lejanas de un mismo archivo: el cierre en `flow-close` (líneas 12, 23 y 70) y el ticket terminado en `global.md` (líneas 21 y 115).

Al terminar, cada regla de momento vive en una sola hoja pequeña, oculta al modelo en los hosts que lo permiten. `global.md` conserva los invariantes y un índice de una línea por hoja. Cada flujo enlaza la hoja en el paso donde aplica, y los principios de código se consultan al escribir código.

## Scope and acceptance

Dentro:
- Cierre del trial #491.
- Soporte del manager para hojas ocultas: Codex y Antigravity.
- Tests de invariantes de hojas.
- Extracción de `global.md`, `flow-plan`, `flow-build`, `flow-verify`, `flow-close`, `git-workflow`, `flow-research`, `flow-figma-design`, `flow-deploy` y `unattended-delegation`.
- Detectores, documentación y sitio.
- Fase 2: 13 principios, `hive-typescript` y tres prácticas de traspaso tomadas de pstack (AC16).

Fuera:
- Cambiar el significado de una regla, salvo AC16, que reemplaza la regla de no llevar un registro aparte en el trabajo desatendido. Por lo demás, solo se mueve, se fusiona con su copia y se reescribe su condición de lectura.
- Las skills `create-issue`, `create-prd`, `project-layout`, `harness-audit`, `test-audit`, `flow-report`, `hive-ui-criteria` y `workspace-repo`, salvo sus enlaces.
- Los roles de agente, salvo sus enlaces.
- Pilotos de modelo.

- AC1 — El trial está cerrado. El PR #491 está integrado en `development`, y `AGENTS.md` ya no tiene el párrafo «Trial branch (temporary)». No existe `refs/hive/trial-join`. `question-card-alone` registra un criterio de muestra nuevo, con la evidencia medida el 2026-11-18, que reemplaza al suyo. · false today because el PR #491 está abierto y el párrafo y la referencia existen · Observed in: GitHub (`gh pr view 491`), el repositorio en `development` y `git show-ref`
- AC2 — Una skill con `disable-model-invocation: true` tiene en su directorio instalado un `agents/openai.yaml` generado, con `policy.allow_implicit_invocation: false`. Ese archivo se quita cuando la skill pierde la marca, y se rechaza si viene escrito a mano. Las skills sin esa marca se instalan como hoy. · false today because el manager no genera ese archivo y omite o rechaza `agents/` (`plan.go:229-243`, `validSource`) · Observed in: la instalación aplicada en esta máquina (`~/.agents/skills/<hoja>/agents/openai.yaml`) tras `hive update` y `hive apply`
- AC3 — Un test de contenido falla cuando una hoja rompe sus invariantes y pasa cuando las cumple. Los invariantes son: `name` igual a la carpeta, `disable-model-invocation: true`, descripción de una línea de 200 caracteres o menos, el tamaño de `SKILL.md` que fije D11, exactamente una entrada en su índice, y una entrada del índice que resuelve a una hoja existente. · false today because no existe test de hojas · Observed in: `go test ./tests/content/...` con una violación sembrada y sin ella
- AC4 — Cada regla extraída tiene un solo hogar. Para cada frase clave de la tabla de `design.md`, `rg` sobre `content/` encuentra un solo archivo, que es su hoja, y un test lo comprueba. · false today because, por ejemplo, la prohibición de `git add -A` y la regla de ticket terminado aparecen en varios archivos · Observed in: `go test ./tests/content/...`
- AC5 — `global.md` pesa 30,000 bytes o menos y conserva enteras las secciones de autorización, secretos, preservación y alcance, la lectura de `HIVE.local.md` al inicio de la sesión y los modos de entrega. · false today because pesa 52,769 bytes · Observed in: `wc -c content/guidance/global.md` y la revisión del diff
- AC6 — `flow-build`, `flow-plan`, `flow-verify`, `flow-close`, `git-workflow` y `flow-figma-design` pesan 9,000 bytes o menos cada uno. · false today because pesan entre 12,711 y 15,985 bytes · Observed in: `wc -c` de cada `SKILL.md`
- AC7 — Cada hoja de momento está enlazada en el paso del flujo donde aplica: preparar el commit en `flow-build`, ticket terminado en `flow-build`, `flow-close` y `flow-deploy`, y modos de cierre al inicio de `flow-close`. La tabla de enlaces de `design.md` está completa. · false today because `flow-build` tiene 0 enlaces a `git-workflow` · Observed in: `rg` de los enlaces por paso
- AC8 — En 10 sesiones reales o más con la guía nueva, `--audit-session` da 0 fallas de `no_broad_git_add` y del criterio nuevo `leaf_read_before_action`. · false today because 8 sesiones usaron `git add -A` sin la regla cargada y el criterio no existe · Observed in: sesiones reales en esta máquina después del despliegue del PR 1
- AC14 — `flow-research` pasa a ser `hive-mode`, el router de entrada.
  - **Estructura.** Su `SKILL.md` pesa 11,000 bytes o menos y contiene 13 playbooks en `playbooks/<tipo>.md`, de 7,000 bytes o menos cada uno. Cada playbook nombra el orden de las etapas, lo propio de su tipo y sus hojas.
  - **Comportamiento.** El router no edita archivos del repositorio, y ningún texto de `content/`, `site/`, `README.md` o `llms.txt` sigue ofreciendo `/flow-research`.
  - false today because `flow-research` no tiene playbooks y se ofrece en los bloques de siguiente paso
  - Observed in: los archivos, `go test ./tests/content/...` y una sesión real que entra por `/hive-mode`.
- AC15 — Antes de la revisión, `flow-verify` lee `hive-rule-no-comments`, `hive-rule-blast-radius` y `hive-rule-prose`. La pregunta de revisión ofrece `interrogate` como opción y dice que gasta cuota de otro proveedor.
  - false today because verify no tiene esos pasos y la pregunta no ofrece esa opción
  - Observed in: `flow-verify/SKILL.md`, la hoja de la pregunta de entrega y una corrida real de verify.
- AC16 — Fase 2: tres prácticas de traspaso tomadas de pstack.
  - **Registro de decisiones.** El trabajo desatendido y la coordinación E2E llevan un registro TSV, `decisions.tsv`, con una fila por decisión y una por efecto externo (hora, fase, decisión, por qué, evidencia, resultado). La evidencia nombra el objetivo y el cambio, y el resultado dice cómo revertirlo.
  - **Casillas por unidad de entrega.** La plantilla de `tasks.md` de un plan de varias entregas agrupa, por PR (o por commit en `direct-base`, o por candidato en `hold`), casillas de prueba unitaria, en vivo y de rendimiento. `check_change.py` rechaza una casilla en vivo sin predicado y una de rendimiento sin base ni umbral, salvo con `n/a:`.
  - **Lo que no está escrito.** El checkpoint escrito al pausar nombra lo que todavía no está escrito en ningún archivo.
  - false today because no existe el registro, la plantilla no agrupa por PR y el checkpoint no tiene ese campo
  - Observed in: la plantilla, la salida de `tasks.py checkpoint`, `check_change.py` sobre un plan de ejemplo y el registro de una corrida desatendida real.
- AC17 — Ningún playbook ni hoja fija o sustituye el modo de entrega. Cada parada, push, PR, merge y promoción que nombran remite al modo en vigor. La matriz de «Compatibilidad con los modos de entrega» de `design.md` se cumple en los 13 playbooks y en las piezas de AC15 y AC16.
  - false today because no existen los playbooks ni las piezas
  - Observed in: un test de contenido y la revisión de código de T10 y T12 contra la matriz. El test rechaza los nombres de modo entre comillas invertidas (`interactive`, `automatic`, `ready`, `direct-base`, `hold`) en los playbooks, salvo en la frase fija, escrita en inglés en los playbooks: «under the delivery mode in force».
- AC13 — La hoja canario `hive-rule-git-commit`, instalada oculta, se lee como archivo antes del primer `git add` en una sesión real de Claude Code y en una de Codex. · false today because la hoja no existe · Observed in: el registro de esas sesiones en esta máquina
- AC9 — El chequeo HA-SK-04 de `harness-audit` acepta `disable-model-invocation` y el `agents/openai.yaml` de Codex en las skills que Hive distribuye como hojas, y sigue marcando otros campos exclusivos de un host. · false today because HA-SK-04 marca `disable-model-invocation` y `paths` · Observed in: `content/skills/harness-audit/references/skills-and-agents.md` y una auditoría de `content/`
- AC10 — `README.md`, `llms.txt`, la página de skills del sitio e `instruction-resources.md` describen las hojas, el índice y su soporte por host. · false today because no existen las hojas · Observed in: el sitio construido localmente (`npm --prefix site run build`) y los dos archivos
- AC11 — Fase 2. Las hojas de principios están instaladas ocultas. `hive-typescript` está visible, con `paths: ["**/*.ts", "**/*.tsx"]` y sin yaml de Codex, y el índice de principios la enlaza. Un índice de principios está enlazado desde el paso de escritura de código de `flow-build` y desde `hive-review-code`. `fix-root-causes`, `test-behavior-not-implementation` y `subtract-before-you-add`/`laziness-protocol` quedan fundidos en sus hojas de Hive, sin duplicado. · false today because no existen · Observed in: la instalación aplicada y `go test ./tests/content/...`
- AC12 — La lectura obligatoria del escenario plan/build de `reading_footprint_test.go` baja respecto de la línea base `2bb96478`, con las rutas nuevas reportadas aparte. · false today because es la línea base · Observed in: `go test ./tests/content -run Footprint -count=1 -v`

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- **Sesiones.** El análisis de 79 sesiones reales (35 de Codex y 44 de Claude Code, del 2026-11-11 al 18) lo hicieron 6 hijos `hive-research`, más una verificación propia. El hallazgo de que `flow-build` no tiene enlaces a `git-workflow` es un `rg` con 0 resultados. La sección «Questions and routes» pesa 14,354 bytes.
- **Trial #491.** Medido el 2026-11-18 con sus scripts, desde el corte del 2026-11-17 09:30:37 UTC:
  - Claude Code: 32 rondas, 0 duplicadas.
  - Codex: 5 tarjetas y 7 rondas en texto.
  - Grok: 5 rondas.
  - Tarjetas sin contestar: 1. Rondas perdidas: 0.
- **Manager.**
  - Copia cada skill tal cual (`tooling/management/plan.go:762-763`).
  - Valida solo el nombre, las rutas permitidas y el `SKILL.md` (`plan.go:229-243`, `:367`, `:431-436`).
  - Las únicas pruebas de frontmatter son `tests/content/skill_frontmatter_test.go` y `frontmatter_test.go`.
  - Los tests de ruta fija que se rompen al mover texto son `reading_footprint_test.go:61-190` y `tests/pilot/` (`flows.go`, `session_audit.go`).
- **Versiones locales observadas el 2026-11-18:** Claude Code 2.1.327, Codex 0.170.0, Grok 1.0.89, Pi 1.5.2, OpenCode 2.3.2, Cursor 2026.11.13 y `agy` 1.5.11.
- **Lo que dice la documentación, no observado en una sesión:**
  - `disable-model-invocation` lo respetan Claude, Grok, Pi, OpenCode y Cursor.
  - Codex necesita `agents/openai.yaml` con `policy.allow_implicit_invocation: false`.
  - Antigravity no tiene interruptor.
  - `paths` solo existe en Claude y Cursor, y no actúa sobre una skill oculta.
  - En Claude y Codex, una skill oculta sale del catálogo.
  - Claude bloquea la invocación de una skill oculta y le pide al modelo no reproducir sus pasos por otra vía. Leerla como archivo no está observado, y lo cubre AC13.
- **Trial #491, el 2026-11-18:**
  - Está `OPEN`, `MERGEABLE` y `CLEAN` (`gh pr view 491 --json state,mergeable,mergeStateStatus`).
  - Su rama va 23 commits detrás de `development` (`git log origin/feat/question-card-alone..origin/development`, medido por el revisor del plan).
  - Existe `refs/hive/trial-join`.
  - La sesión está en `development` (`git branch --show-current`).
- **Referencias.**
  - pstack oculta 57 de sus 58 skills con un solo campo y enruta por ruta relativa desde su router.
  - gentle-ai oculta material en una carpeta sin `SKILL.md`.
  - superpowers y openspec no ocultan nada.

## Effects

- integrar `development` en la rama del trial y hacer merge del PR #491 · target: GitHub `JhonHawk/tricell-hive` · ceiling: 0 · tasks: T1 · answer: approved (D4 B)

## Delivery

- **Modo `interactive` (D8 A).** El trabajo para antes de cada push, el del PR 1 y el del PR 2, para tu validación. Después siguen el push, el PR y el merge en `development`.
- **Revisión.** Claude Code `/code-review` sobre cada candidato antes de la validación (D9 A).
- **Refresco local.** Tras cada integración, la instalación local se refresca por la marca de mantenedor.
- **Lo que esto no autoriza:** una release ni la promoción a `master`.

## Review status and progress

2026-11-18 — Plan escrito.
