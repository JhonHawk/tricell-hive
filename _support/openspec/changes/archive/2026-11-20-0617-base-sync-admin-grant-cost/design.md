# Design

## Approaches

Solo una forma razonable para cada parte, porque las decisiones del usuario ya fijaron el mecanismo:

- D1 eligió el script: se extiende `cleanup_branches.py`, que ya planea y aplica la limpieza por repositorio y ya tiene pruebas con `origin`. Una hoja en prosa fue la opción descartada.
- D2 cambia el texto de una sola hoja, `hive-rule-merge-gate`, la única casa de la regla de bypass (`single_home_test.go`), más las tres frases del global que la contradirían.
- D3 eligió el auditor: se agrega un bloque informativo a `--audit-session`, que la rúbrica corre primero en cada revisión y que ya lee las tres hosts. `measure-delegation.py` queda como referencia de cómo leer el uso de Claude y Codex; no se mueve ni se borra.

## Proposed shape

### Base al día en `cleanup_branches.py`

El plan gana una clave `base_update` (`base` ya nombra la rama):

```json
"base_update": {"action": "fast_forward | current | keep", "checkout": "<path>", "from": "<sha>", "to": "<sha>", "commits": 20, "reason": "...", "paths": ["..."]}
```

Una función `plan_base(repo, base, main_checkout)` calcula la entrada sin cambiar nada. En `main()`, solo con `--apply` y antes de planear la limpieza, el script corre `git fetch origin <base>`, calcula `plan_base` y, si da `fast_forward`, avanza la base; después planea y aplica la limpieza con la base ya al día, para que una rama mezclada solo en `origin` cuente como mezclada. Sin `--apply`, `plan_base` se calcula con las referencias que ya existen, sin fetch. Si el fetch o el avance fallan, `base` queda `keep` con la razón en `base_update.reason` y en `applied.errors` (salida 1), y la limpieza de ramas sigue igual.

`status`, `diff` y `merge` corren con `git -C <main_checkout>`, el primer worktree de `git worktree list`, aunque `--repo` apunte a un worktree enlazado. Reglas, en orden; la primera que aplica decide:

1. No hay checkout principal (repositorio desnudo) → `keep`, «no main checkout».
2. No hay remoto `origin` o no existe `origin/<base>` → `keep`, «no origin/<base>».
3. El checkout principal no está en `<base>` → `keep`, «main checkout on <branch>»; nunca cambia la rama de un checkout.
4. La base local tiene commits que no están en `origin/<base>` (`rev-list --count origin/<base>..<base>` > 0) → `keep`, «local <base> has N commits not on origin».
5. `origin/<base>` no tiene commits nuevos → `current`.
6. Algún archivo que cambia entre `HEAD` y `origin/<base>` (`git diff --name-only HEAD origin/<base>`) aparece en `git status --porcelain=v1 -uall --ignored` del checkout principal (modificado, preparado, sin seguimiento o ignorado) → `keep`, «incoming changes touch local files», con esos archivos en `paths`.
7. Si no → `fast_forward`; con `--apply` corre `git merge --ff-only --no-overwrite-ignore origin/<base>` en el checkout principal.

`merge --ff-only` no depende de `pull.rebase`, que en fleetsystem hizo fallar al `git pull`. Las pruebas llaman a `main(argv)` para los casos de fetch y de `--apply`, y a `plan_base` para las reglas. La referencia `branch-cleanup.md` pide reportar `base_update` con lo demás, su línea 3 nombra el avance como efecto que el procedimiento autoriza, y «return its checkout to the base» queda solo como volver a la base, no como actualizarla.

### Permiso de admin-merge por sesión

Texto propuesto para el último punto de `hive-rule-merge-gate`:

> Never bypass branch protection, required reviews or required checks, except an admin merge of an authorized pull request whose required checks passed where the repository does not enforce its rules on admins. Any other bypass needs the user's explicit grant, named in each report it covers. A grant to admin-merge holds for the session, unless withdrawn, for its pull requests in the work's repositories whose required checks passed or that have none, and lifts the hold of a reviewer whose latest review approves an earlier commit, named with it; never for the production branch (the last of `Environments` when it is not the base), a direct push, a promotion-path deviation, a requested reviewer, even one who approved before, or requested changes.

El punto del revisor humano gana un final: «…, unless an admin-merge grant lifts it as the last bullet sets». Con ese texto la hoja pesa unos 4360 bytes: entra en `leafSizeExceptions` de `tests/content/leaf_skill_test.go` con el comentario «One moment, the merge: CI wait, current base, reviewer hold and bypass grants, read together before merging». La rama de producción es la última del encadenamiento `Environments` del `## Hive` cuando no es la base; un repositorio sin encadenamiento no tiene rama de producción para esta regla. Un revisor que aprobó un commit anterior y volvió a ser pedido sigue en `requested_reviewers` y retiene el merge.

Frases del global (`content/guidance/global.md`), con su tamaño antes y después en el pull request:

- línea 20: «repeated approvals are not standing permission» → «repeated approvals are not standing permission, except a grant its rule scopes to the session, such as an admin merge's».
- línea 140: «a human reviewer's pending approval holds the merge» → añade «unless an admin-merge grant lifts it».
- «Preservation», la lista de efectos del cierre: añade «fast-forwarding the base in the main checkout of each repository the work touched, as [flow-close](…branch-cleanup.md) «Branch cleanup» sets».

### Bloque `main_thread_cost` del auditor

```go
type mainThreadCost struct {
    ModelCalls   int            `json:"model_calls"`
    InputTokens  int64          `json:"input_tokens"`   // entrada procesada, con caché
    CachedTokens int64          `json:"cached_tokens"`
    PeakContext  int64          `json:"peak_context,omitempty"` // solo con uso por llamada
    UsageSource  string         `json:"usage_source"`   // per_call | per_turn | none
    Compactions  int            `json:"compactions"`
    ToolCalls    int            `json:"tool_calls"`
    HiveReads    int            `json:"hive_reads"`
    HiveScripts  int            `json:"hive_scripts"`
    ByTool       map[string]int `json:"by_tool"`
}
```

- Claude: una llamada por `message.id` del asistente, con `input_tokens + cache_read_input_tokens + cache_creation_input_tokens` de su `usage`, contada una vez por id (como en `measure-delegation.py`); solo líneas que no sean `isSidechain`; compactación = registro `system` cuyo subtipo dice `compact`.
- Codex: una llamada por evento `token_count` con `info.last_token_usage`, solo cuando `info.total_token_usage` cambió respecto del anterior; entrada = `input_tokens`, caché = `cached_input_tokens`; compactación = registro `compacted`.
- Grok: suma de `turn_completed.usage` (`modelCalls`, `inputTokens`, `cachedReadTokens`) menos el `usage.json` de cada hijo que nombra `subagents/*/meta.json` (`child_session_id`, buscado en todas las carpetas de proyecto de `~/.grok/sessions/`, porque un hijo puede vivir en la de otro directorio), `usage_source: per_turn`, sin `peak_context`; si falta el uso de algún hijo, no resta ninguno y marca `per_turn_with_children`. Compactación = `auto_compact_completed`. En `01a1f5fc`: 599 − 189 = 410 llamadas, 70,803,744 − 15,688,019 = 55,115,725 tokens (medido el 2026-11-20).
- Lectura de Hive: la ruta (o el texto del comando de lectura `cat`, `sed`, `head`, `tail`, `rg`, `grep`) contiene `/.<carpeta>/…/skills/<nombre>/SKILL.md` o `…/skills/<nombre>/references/` bajo una carpeta oculta (como `~/.agents/skills/`, `~/.claude/skills/`), no `content/skills/` de un repositorio. Script de Hive: el comando contiene `…/skills/<nombre>/scripts/<archivo>.py` bajo una carpeta oculta. Una llamada cuenta en una sola clase; `by_tool` cuenta todas por nombre de herramienta.

Los loaders guardan el uso en la sesión cargada (`auditSession`), no en cada `traceEvent`, así ningún criterio cambia.

## Impact

- `cleanup_branches.py` lo corre `flow-close` y una petición de higiene; el JSON gana una clave y `--apply` gana un fetch y un merge en el checkout principal. El código que lee la salida es el agente, que lee la referencia.
- La regla de bypass vive solo en `hive-rule-merge-gate`; el global la contradiría en tres frases (líneas 20 y 140, y la lista de efectos del cierre en «Preservation»), que cambian como se dijo arriba.
- `--audit-session` gana un campo; las pruebas que comparan el JSON completo, si las hay, se actualizan.
- Ninguna parte toca producción.

## Recovery

Revertir el pull request deshace todo. El avance de la base local es un fast-forward: se deshace con `git reset --keep <from>` usando el `from` del reporte.
