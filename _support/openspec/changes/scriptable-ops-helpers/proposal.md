# scriptable-ops-helpers — Estado de git y del PR, y progreso de `tasks.md`, con scripts

| Campo | Valor |
| --- | --- |
| Estado | Borrador · plan escrito el 2026-11-08 |
| Tracker | Sin ticket |
| Git | tricell-hive · base `development` @ `71c18043` · rama propuesta `feat/scriptable-ops-helpers` |

## Decisiones

2026-11-08
- **D6 A** — ¿Qué vehículo usan los helpers? Scripts Python incluidos en las skills (`git-workflow/scripts/git_state.py`, `flow-plan/scripts/tasks.py`), no subcomandos de `hive`: viajan en la misma versión que la skill que los invoca y usan el `python3` que `flow-plan` y `flow-close` ya asumen.
- **D7 A** — ¿Qué cubre el helper de git? Estado de git y del PR en una llamada de solo lectura: rama, cambios sin commit, adelante/atrás, commits con autor, si el head contiene la base (con fetch opcional), y el PR con sus checks y revisiones filtradas al head y a `--since`. La limpieza de ramas (`branch-cleanup.md`) queda fuera.
- **D8 A** — ¿Qué cubre el helper de `tasks.md`? Marcar las casillas de una tarea, agregar su línea de progreso con el commit, contar `n/total` y escribir un checkpoint con formato fijo, que se suma a la plantilla. El visor no cambia.

## Pedido

"B y C" (2026-11-08), sobre la opción "Retomar scriptable-ops H1/H2: es la lección del artículo que tiene medición propia", y luego `/flow-plan` sobre "Planificar B: script de estado git+PR y gestión de tasks.md".

Origen: `_support/sessions/2026-11-06-scriptable-ops/scriptable-ops.research.md`, candidatos H1 (estado de git y del PR, ~410k + 176k tokens por semana como cota superior, ~450 ventanas por semana) y H2 (`tasks.md`, ~77k tokens por semana; protege el checkpoint tras una compactación).

## Objetivo

Hoy el agente reconstruye el estado de git y del PR con 3 a 10 llamadas sueltas (`git fetch`, `git log origin/<base>..HEAD`, `git merge-base --is-ancestor`, `gh pr view`, `gh api --paginate` sobre revisiones y comentarios) y aplica a mano el filtro de revisiones por commit y por hora del push de git-workflow paso 7. En `tasks.md` marca casillas, escribe la línea de progreso y el checkpoint con ediciones a mano, y la plantilla no fija dónde va el checkpoint.

Al terminar: una llamada a `git_state.py` devuelve ese estado en JSON con el orden fetch → comparación ya aplicado y el filtro de revisiones hecho; `tasks.py` marca una tarea, escribe su línea, cuenta y mantiene una sección `## Checkpoint` con formato fijo. Las skills invocan los scripts y conservan los comandos manuales como alternativa.

## Alcance y aceptación

Dentro: los dos scripts con sus tests, la exención del checkpoint en `check_change.py`, la sección `## Checkpoint` en la plantilla, los punteros en `git-workflow` y `flow-build`, la documentación (README, `site/`, `instruction-resources.md`) y el delta de `workflow-lifecycle`.

Fuera: la limpieza de ramas (`branch-cleanup.md`), la espera de CI (`gh pr checks --watch` sigue siendo del agente), decidir qué commits son de este trabajo o cuándo fue el push (los aporta el agente), `hive view work tasks register` y cualquier subcomando de `hive`.

- AC1 — En un repo cuya rama va 2 commits adelante y 1 atrás de `origin/development`, `git_state.py --base development --fetch` imprime JSON con `ahead: 2`, `behind: 1`, los 2 commits con sha, asunto y autor, y `base_in_head: false`; tras integrar la base en local sin push, `base_in_head: true` y `base_in_pushed_head: false`, y tras el push, ambos `true`; si el fetch falla, esos campos salen en `null` · falso hoy porque el script no existe (`ls content/skills/git-workflow/scripts/` falla).
- AC2 — Con un PR cuyo head (`head_ref_oid`) tiene una revisión de bot sobre un commit viejo y otra sobre el head, un comentario de revisión humano sobre un commit viejo, una respuesta en un hilo y comentarios anteriores y posteriores a `--since`, cada elemento sale con `user_type`, `on_head` y `after_since` correctos (la respuesta en hilo por `created_at`, los demás comentarios por `updated_at`), y ninguno trae un veredicto que descarte el comentario humano · falso hoy: no existe el script y el filtro lo hace el agente a mano.
- AC3 — Cuando `gh` no está instalado o falla, la salida trae igual los campos de git, `pr.available: false` con el motivo y sale con código 1; contra un remoto HTTP que responde 401 no queda esperando credenciales, y un `gh` que tarda más que su límite termina con `error` dentro de ese límite · falso hoy: no existe el script.
- AC4 — Correr `git_state.py` no cambia el índice, el árbol de trabajo, las ramas locales ni la configuración; con `--fetch` solo cambian las refs de seguimiento remoto. No usa shell · falso hoy: no existe el script.
- AC5 — `tasks.py <carpeta> done T2 --commit <sha>` marca solo las casillas de T2, agrega bajo `## Review status and progress` una línea con fecha, T2, el sha y lo que quedó sin verificar, e imprime `✓ T2 · 2/4 · <título>`; correrlo otra vez no duplica la línea · falso hoy: no existe el script.
- AC6 — `tasks.py <carpeta> checkpoint` con etapa, paso, último commit, hallazgos abiertos, siguiente acción y un `--check` escribe la sección `## Checkpoint` con formato fijo y su `### Evidence`; una segunda llamada reemplaza las líneas fijas y conserva la evidencia anterior; y `check_change.py` sale con 0 aunque esa sección nombre una decisión reemplazada · falso hoy: no existe la sección ni el script, y `check_change.py` solo exime la sección de progreso (`check_change.py:75`).
- AC7 — `git-workflow/SKILL.md` (pasos 2, 7, 8 y 9), `flow-build/SKILL.md` (pasos 2 y 5 y el checkpoint) y `flow-verify/SKILL.md` (el checkpoint) invocan los scripts con la alternativa manual, `git-workflow` mantiene el filtro de revisiones solo para bots y los hilos humanos abiertos, la plantilla `tasks.md` trae `## Checkpoint`, y `go test ./tests/content/` pasa (tamaños ≤ 16.000 B) · falso hoy: ninguna skill nombra estos scripts.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- Comandos y salidas medidos en `_support/sessions/2026-11-06-scriptable-ops/` (199 sesiones de Claude Code y 39 de Codex, 2026-10-16 → 2026-11-06).
- Hijo `hive-research` (2026-11-08) sobre `origin/development` @ `71c18043`; verificado en el hilo principal: `main.go` despacha con un `switch`, `newGitRunner` en `doctor_project.go:165` fija `GIT_OPTIONAL_LOCKS=0`, el README dice que el instalador no edita `PATH`, ningún código Go llama a `gh`, `viewer-publication.md:36` dice que el visor no parsea casillas.
- Los 8 scripts de `content/skills/*/scripts/` son de biblioteca estándar y ninguno importa `subprocess`; cada uno tiene su test en `tests/skills/` corrido con `python3 -m unittest discover -s tests/skills`.
- `check_change.py:50,75` exime solo la sección cuyo título contiene progress/progreso/avance; no lee títulos `### T<n>` ni casillas marcadas.
- Tamaños actuales: `git-workflow/SKILL.md` 14.440 B, `flow-build/SKILL.md` 13.108 B (tope 16.000).

## Entrega

Respuestas vigentes en la sesión para tricell-hive (2026-11-08): **automatic** (D3 A), con revisión, push, PR y merge sin parar para validación; revisión **Native: Claude Code `/code-review`** (D4 A). Este cambio ejecuta procesos (`git`, `gh`): `hive-review-security` corre solo si el usuario la pide. La entrega no autoriza un release ni tocar el binario instalado más allá del refresco local del mantenedor.

## Estado de la revisión y progreso

2026-11-08 — Plan escrito.
2026-11-08 — Revisión `hive-review-plan`: 0 P0, 2 P1, 6 P2, varios P3, todos aplicados. El helper informa hechos por revisión (`on_head`, `after_since`) en vez de un veredicto que descartaría hilos humanos; el checkpoint conserva `### Evidence` para la evidencia de build y verify; comparaciones en `null` si el fetch falla; `base_in_pushed_head` en AC1; PR y checks por REST; tests reales de no-prompt y timeout; prueba real sobre el PR #426 ya fusionado; `--slurp`, `--unverified` obligatorio, `flow-verify` entre los consumidores.
