# Design

## Approaches

1. **Scripts Python incluidos en las skills** — elegido (D6 A). Viajan con la skill que los invoca, así que la instrucción y el script siempre están en la misma versión; usan el `python3` que `flow-plan` y `flow-close` ya asumen; siguen el patrón de `tests/skills`. Costo: es el primer script que ejecuta procesos (`git`, `gh`), así que se blinda como `newGitRunner`.
2. **Subcomandos de `hive` en Go** — descartado: el binario no está garantizado en el `PATH` (README: el instalador no edita `PATH`), `git-workflow` y `flow-build` pasarían a depender del gestor, y el binario se actualiza aparte del contenido, de modo que una skill podría pedir un subcomando que el binario instalado no tiene.
3. **Mixto** (H2 en Python, H1 en `hive`) — descartado: suma los costos de ambos.

## Decisions

- **D7 A** — `git_state.py` es de solo lectura y no barre la limpieza: los candidatos de `branch-cleanup.md` necesitan `lsof`, comparaciones `cmp` contra el checkout principal y entradas de Engram que git no da.
- **D8 A** — `tasks.py` no alimenta `hive view work tasks register`: `walkthrough-viewer/spec.md:502` prohíbe inferir estado del Markdown y `tasks.md` no tiene etapas ni la revisión del plan.
- **Decisiones menores tomadas sin preguntar** (no cambian alcance, contratos, datos guardados, costo ni reversibilidad):
  - `tasks.py done` y `checkpoint` escriben directo, sin `--apply`: son ediciones pequeñas de un archivo versionado que el diff muestra; `apply_deltas.py` usa `--apply` porque mueve carpetas y fusiona specs.
  - La salida de `git_state.py` es JSON compacto en stdout; los errores por fuente van dentro del JSON, no en stderr.
  - Los scripts se invocan como `python3 -I <git-workflow-dir>/scripts/git_state.py` y `python3 -I <flow-plan-dir>/scripts/tasks.py`, cada uno con su link Markdown, como `flow-plan/SKILL.md:54`. Con `-I`, `tasks.py` no puede importar `check_change.py`: cada script es autocontenido.
  - El script informa hechos; las reglas de juicio quedan en la skill: qué commits son del trabajo, qué cuenta como bot además de `user.type == "Bot"` (la cuenta de máquina de la automatización que lee el paso 6) y que un hilo humano abierto sigue vigente.

## Proposed shape

### `content/skills/git-workflow/scripts/git_state.py`

```
python3 -I git_state.py --base <branch> [--repo DIR] [--remote origin] [--fetch]
                        [--pr N | --no-pr] [--since ISO-8601]
```

- Orden fijo: con `--fetch`, `git fetch <remote>` corre antes de cualquier comparación (git-workflow pasos 2 y 8). Si el fetch falla, `ahead`, `behind`, `commits_ahead`, `base_in_head` y `base_in_pushed_head` salen en `null`: nunca se compara contra refs viejas cuando se pidió refrescarlas.
- Campos de git: `repo`, `branch`, `head`, `upstream`, `dirty {staged, unstaged, untracked}`, `base`, `fetch {ran, ok, error}`, `ahead`, `behind`, `commits_ahead [{sha, subject, author, date}]` (los commits de `<remote>/<base>..HEAD`; el agente decide cuáles son de su trabajo), `base_in_head` (`merge-base --is-ancestor <remote>/<base> HEAD`), `pushed_head` (sha de `<remote>/<branch>` o null) y `base_in_pushed_head` (`merge-base --is-ancestor <remote>/<base> <remote>/<branch>`, el criterio del paso 8).
- Campo `pr` (salvo con `--no-pr`), todo por REST con `gh api`, sin GraphQL (git-workflow paso 9 lo evita bajo límite de tasa):
  - El repo `<owner>/<repo>` sale de la URL del remoto. Sin `--pr`, el PR abierto de la rama: `repos/<owner>/<repo>/pulls?head=<owner>:<branch>&state=open`; con `--pr N`, `repos/<owner>/<repo>/pulls/N`.
  - `available`, `reason`, `number`, `state`, `merged`, `head_ref_oid`, `head_matches_local`, `head_matches_pushed`.
  - `checks`: `commits/<head_ref_oid>/check-runs` (`name`, `status`, `conclusion`) y `commits/<head_ref_oid>/status` (estados heredados).
  - `reviews`, `review_comments`, `comments`: `pulls/<n>/reviews`, `pulls/<n>/comments`, `issues/<n>/comments`, cada uno con `gh api --paginate --slurp` (sin `--slurp` las páginas se concatenan en JSON inválido; verificado en la ayuda de `gh` 2.103.3).
  - Cada elemento lleva hechos, no veredictos: `user`, `user_type`, `created_at`, `updated_at`, `commit_id`/`original_commit_id`, `in_reply_to_id`, `on_head` (su commit es `head_ref_oid`, el del PR, no el HEAD local) y `after_since` (para respuestas en un hilo compara `created_at`; para los demás comentarios, `updated_at`; `null` sin `--since`). No hay campo `current`: la skill aplica el filtro del paso 7 solo a bots y conserva los hilos humanos abiertos.
- Una lectura por endpoint: no espera CI ni consulta en bucle.
- Procesos: listas de argumentos sin shell, `timeout` por comando (fetch 60 s, el resto 20 s; un test puede bajarlos con una variable de entorno que solo lo reduce), salida acotada, entorno con `GIT_OPTIONAL_LOCKS=0`, `GIT_TERMINAL_PROMPT=0`, `GCM_INTERACTIVE=never`, `GH_PROMPT_DISABLED=1`, `GH_NO_UPDATE_NOTIFIER=1`, `LC_ALL=C`.
- Fechas ISO-8601 con `Z` parseadas reemplazando `Z` por `+00:00` antes de `datetime.fromisoformat`, válido desde Python 3.7.
- Salida: 0 si todas las fuentes respondieron; 1 si alguna falló (su campo trae `error`, los demás siguen con la regla del fetch de arriba); 2 si el uso es inválido o `--repo` no es un repo de git.

### `content/skills/flow-plan/scripts/tasks.py`

```
python3 -I tasks.py <change-folder> status
python3 -I tasks.py <change-folder> done T<n> --commit <sha> --unverified "<text>|none"
python3 -I tasks.py <change-folder> checkpoint --stage <stage> --step <step> --commit <sha>
        --next "<action>" [--finding "<ID: decision>"]... [--check "<command> · <result> · <commit>"]...
```

- Solo para carpetas de cambio. El checkpoint de un plan guardado en la conversación (archivo en la ubicación temporal del host) sigue editándose a mano.
- Tareas: títulos `### T<n> — <título>`; sus casillas son las líneas `- [ ]`/`- [x]` hasta el siguiente título `##` o `###`.
- `status`: una línea por tarea (`T1 [x] 2/2 · título`) y el total `n/total tasks`.
- `done`: marca las casillas de esa tarea, agrega bajo `## Review status and progress` la línea `<fecha> — T<n> · <sha> · unverified: <texto>` si no existe ya una línea para esa tarea y ese sha, e imprime `✓ T<n> · <hechas>/<total> · <título>`. `--unverified` es obligatorio (`none` cuando todo quedó verificado), para no escribir texto en un idioma fijo. Sale con 2 si la tarea no existe o si falta el título de progreso, sin crearlo.
- `checkpoint`: escribe o reemplaza `## Checkpoint`, justo antes de `## Review status and progress`, con líneas fijas `- Updated:`, `- Stage:`, `- Last commit:`, `- Open findings:` (o `none`), `- Next:`, seguidas de `### Evidence`. Cada `--check` agrega una línea a `### Evidence`; al reemplazar, las líneas existentes de `### Evidence` se conservan, para que la evidencia que guardan `flow-build` paso 5 y `flow-verify` (comandos, salida, commit revisado, checks reutilizados) sobreviva a la siguiente actualización.
- `check_change.py` exime `## Checkpoint` y su `### Evidence` igual que la sección de progreso.

## Impact

- Consumidores: `git-workflow/SKILL.md` pasos 2, 7, 8 y 9; `flow-build/SKILL.md` paso 2, paso 5 y el párrafo del checkpoint; `flow-verify/SKILL.md` líneas 21 y 25 (dónde guarda el commit revisado, las decisiones de hallazgos y los checks reutilizados); la plantilla `flow-plan/assets/change-folder/tasks.md`; `check_change.py`. La regla de commits ajenos de «Delivery» en `global.md` no cambia: ya remite a git-workflow.
- Documentación: README (requisito `python3` para los scripts de las skills); `site/src/content/docs/flows/build.md` y `verify.md:30` revisados (lectura de revisiones y merge); `instruction-resources.md`.
- Presupuesto: `git-workflow/SKILL.md` tiene 1.560 B libres. Los punteros reemplazan los comandos manuales por una frase "si el script falta o sale con 2, usa los comandos de <paso>" y no los repiten; si aun así pasa de 16.000 B, la alternativa manual se mueve a una referencia `git-workflow/references/manual-state.md` con su condición de lectura.
- No toca producción ni datos guardados; no agrega dependencias.

## Recovery

Revertir el merge quita los scripts y los punteros; las skills vuelven a los comandos manuales. Las carpetas de cambio que ya tengan `## Checkpoint` siguen siendo válidas para un `check_change.py` revertido solo si su contenido no nombra decisiones reemplazadas (sin la exención, esas líneas se marcarían).
