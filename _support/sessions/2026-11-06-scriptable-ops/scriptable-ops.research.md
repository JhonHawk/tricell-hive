# Operaciones de las skills que podrían pasar a scripts (2026-11-06)

Pregunta: ¿qué operaciones que piden las skills (u otros procesos) podrían hacerse con un script determinista (Python, shell o un subcomando de `hive`), y cuánto ahorrarían en contexto y en tiempo, medido en sesiones reales?

- Scripts: `scriptable_ops.py` (carga los transcripts y escribe `$TMPDIR/scriptable-ops-calls.pkl`), `report.py` (ranking), `inline_purpose.py` (etiqueta los scripts inline). Uso: `python3 -I scriptable_ops.py build <session-id-a-excluir>`, luego `python3 -I scriptable_ops.py report [seq N | fam N | famall | bundle N]`. El pickle guarda hasta 4 000 caracteres del texto de cada comando: bórrenlo al terminar y nunca lo guarden aquí.
- Datos: hilos principales, sin subagentes. Claude Code: 199 sesiones, 37 200 llamadas, 16 repos, 2026-10-16 → 2026-11-06. Codex: 39 sesiones, 8 041 llamadas (24 411 operaciones al separar los lotes `exec`), 6 repos, 2026-10-19 → 2026-11-06. Son 3,01 semanas de un usuario en una máquina.
- Método: el de `../2026-11-05-archive-scripts-cost/measure-archive-cost.py`. El tamaño es (texto de la llamada + resultado) / 4. Los segundos son exclusivos y cuentan desde el último evento del usuario o resultado de herramienta, con tope de 600 s. Cada llamada se atribuye a la última skill cargada. Un candidato es la familia de llamadas dentro de una ventana de skill, que corta al cambiar de skill o tras 30 min de pausa. El costo con script se estima en 168 tokens y `max(4 s, tiempo de ejecución)`.
- Inventario de las skills: un child `hive-research` recorrió `content/guidance/global.md`, `content/skills/*` y `content/agents/*` (las referencias `file:line` siguientes vienen de ahí).

## Resultados

Escala: unos 11,5 M tokens de resultados de herramientas por semana. Las cifras semanales son cotas superiores (un script por ventana), se solapan y no se suman. Entre paréntesis va el subconjunto conservador: ventanas con 3 o más llamadas.

| | Candidato | Skills | Contexto/sem | Tiempo/sem | Ventana: contexto mediana/p90 | Fallo |
|---|---|---|---|---|---|---|
| H1 | Estado de git (status, log, branch, fetch, worktree, rev-parse, merge-base) | git-workflow, flow-close, flow-build, flow-verify | 410k (308k) | 5,0k s | 405 / 2,9k | 14 % |
| H1 | Estado del PR y el CI (`gh pr view/checks`, `gh run`, `gh api`) | flow-close, git-workflow, flow-research | 176k (140k) | 2,3k s | 333 / 2,5k | 11 % |
| H2 | `tasks.md`: marcar y contar casillas, checkpoint | flow-build, flow-close, flow-walkthrough | 77k (53k) | 1,8k s | 471 / 2,0k | 3 % |
| H3 | Publicación en el visor (`hive view *`) | flow-walkthrough, flow-close, flow-build | 217k (189k) | 2,6k s | 1,0k / 4,9k | 23 % (22 % reintentos) |
| H4 | Lecturas de openspec (`tasks.md`, `proposal.md`, `design.md`, specs) | flow-plan, flow-build | 177k (68k) | 0,6k s | 830 / 3,9k | 8 % |

Total realista: unos 0,4–0,6 M tokens por semana (~4–5 % de los tokens de herramientas) y 1,5–2,5 h por semana de llamadas. A eso se suma que cada token ahorrado temprano deja de releerse en las llamadas siguientes (sobre todo lectura de caché); no se cuantificó por candidato.

### Lectura

- **H1 — Un script que entregue el estado de git y del PR en una llamada:** rama, commits por delante o por detrás de la base, commits ajenos, worktrees con su estado, ramas fusionadas con `headRefOid`, checks y reviews filtradas al head.
  - Es lo más frecuente (unas 450 ventanas por semana). La mediana ahorra poco (~240 tokens y ~4 s por ventana); el valor está en el p90.
  - También da fiabilidad en reglas de orden exacto: `git-workflow/SKILL.md:11,15,21-27,32`, `flow-close/references/branch-cleanup.md:14-39` y `global.md` «Delivery» (commits ajenos).
- **H2 — Un script para la gestión de cada tarea:** marcar la casilla, escribir la línea de commit y el checkpoint (`flow-build/SKILL.md:13,18`). Añadiría `hive view work tasks register` que lea `tasks.md` directamente en vez del JSON que hoy arma el agente (`flow-walkthrough/references/viewer-publication.md:32`). Ahorra poco y protege el checkpoint tras una compactación.
- **H3 — El visor ya es un CLI.** El costo son los errores y reintentos, y las lecturas de `hive view contract` (p90 1,65k).
  - La fila «Python inline alrededor del visor» (199k por semana) es en su mayoría ruido del modo código de Codex: una celda mezcla un `python3` que copia la carpeta de cambio con `hive view work bind` (muestra revisada: Codex, 2026-11-04, sesión `01a1a55e`).
  - El PR #402 (2026-11-06) corrigió los mensajes de error de `view work`. Hay que medir de nuevo antes de añadir nada.
- **H4 — Casi todo es lectura que el modelo necesita para decidir.** No compensa.
- **H5 — Descartados.**

  | Cadena | Por semana | Ahorro neto por vez |
  |---|---|---|
  | `git add` → `commit` → `log` | 241 | 33 tokens, 0,6 s |
  | `git status` → `log` | 85 | 3 tokens |
  | `git push` → `gh pr create` | 62 | 6 tokens, 2,2 s |
  | `go vet` → `go test` | 57 | 1,5 s (limitada por la ejecución) |

- **H6 — Lo más caro no se resuelve con scripts.**
  - Leer `SKILL.md` y sus referencias cuesta ~1,8 M tokens por semana (16 %).
  - El ciclo de Engram cuesta ~930k tokens y ~8,8k s por semana: 432 `mem_save` por semana, y en 92 casos por semana `mem_save` va seguido de tres `mem_judge` sueltos, aunque existe `mem_judge_many`.

### Costo de lo que ya es un script (referencia)

Mediana por llamada cuando la llamada ejecuta solo esa operación.

| Operación | Llamadas | Contexto | Tiempo | Fallo |
|---|---|---|---|---|
| `hive view work` | 100 | 168 tokens | 4,6 s | 3 % |
| `hive view contract` | 100 | 336 tokens | 3,1 s | 3 % |
| `hive update` | 50 | 130 tokens | 3,7 s | 0 % |
| `apply_deltas.py` | 23 | 378 tokens | 5,5 s | 17 % |
| `check_change.py` | 50 | 611 tokens | 19,7 s | 6 % |
| `self_check.py` | 17 | 316 tokens | 5,4 s | 18 % |

Un paso con script cuesta 130–340 tokens y 3–5 s, lo que confirma la estimación de 168 tokens y 4 s. Los de `apply_deltas.py` y `check_change.py` salen más caros porque iban encadenados con otros comandos. `archive_sessions.py` y `contrast-check.py` no aparecen como llamadas propias.

### Inventario de las skills: candidatos que la medición no aisló

- **Instalación local.** Nombrar los commits sin desplegar, `go build` a `.new` y luego `mv`, y `hive update --out` seguido de `hive apply` (`AGENTS.md` «Refreshing a local installation»).
- **Vigencia de `DESIGN.md`.** Comparar los hashes de las fuentes, la versión en el lockfile y los globs (`hive-ui-criteria/references/design-record.md:70-76`).
- **Comprobación de paquetes.** Antigüedad de 7 días, publicador y advisory (`global.md` «Scope and authorization»).
- **Inventario de material de soporte al cerrar** (`flow-close/SKILL.md:29-37`).
- **Copia del índice de cuentas a memoria al cerrar** (`flow-close/SKILL.md:43-46`).
- **Huella del candidato.** Reutilizar un check solo si las entradas no cambiaron (`flow-verify/SKILL.md:12,18`).

## Límites

- Las cifras son cotas superiores: agrupar en un script por ventana colapsa también las salidas grandes. La cifra conservadora es la de ventanas con 3 o más llamadas.
- El tiempo incluye esperas de permiso (hasta 600 s) y el razonamiento de Codex (mediana 12,7 s por llamada, frente a 4,8 s en Claude).
- La atribución a skills arrastra: dura hasta la siguiente carga o hasta 30 min de pausa, y leer un `SKILL.md` de este repo cuenta como carga.
- En Codex, los lotes `exec` se separan con regex y reparten tokens y segundos por partes iguales: confianza baja.
- Las sesiones de medición del mantenedor inflan las filas del visor y de `hive`.
- No mide razonamiento, subagentes ni resultados con imágenes (cada uno cuenta como 1,5k tokens). Sin Grok, Cursor, OpenCode ni Pi.

## Siguiente medición

- Medir de nuevo H3 (visor) con sesiones posteriores al PR #402.
- Si se implementan H1 o H2, comparar sus ventanas contra esta línea base con el mismo script.
