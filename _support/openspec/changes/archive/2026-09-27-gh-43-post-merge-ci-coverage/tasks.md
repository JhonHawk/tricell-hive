# Tareas

**Commit base:** `ec20d91` (fijado al empezar el build, 2026-09-27).

### T1 — Regla en `verification.md` y referencia en `git-workflow`

- [x] Las dos viñetas de [design.md](design.md#regla-nueva-en-verificationmd) en `verification.md`, después de la viñeta "At integration or promotion…".
- [x] La frase de [design.md](design.md#referencia-en-git-workflow) en `git-workflow/SKILL.md:30`.

**Evidencia (T1):** `review-task` nativo dio AC1, AC2 y AC3 como `met`: `verification.md:76-77` coincide palabra por palabra con este diseño, `git-workflow/SKILL.md:30` también, y la lectura contra los dos casos da el resultado de la tabla. Las búsquedas fallan en `ec20d91` y pasan con el cambio.

**Closes:** AC1, AC2, AC3.
**Depende de:** nada.
**Ubicaciones:** `content/skills/flow-build/references/verification.md` (sección "Frequency and coverage"), `content/skills/git-workflow/SKILL.md`.
**Ejecución:** hilo principal. Son dos inserciones de texto ya redactado, y el brief sería más largo que el cambio. Se edita por reemplazo exacto, sin reescribir el archivo.
**Enfoque de prueba:** `check`: las búsquedas de abajo devuelven cada parte de la regla, y la comprobación con los dos casos da el resultado de la tabla de [design.md](design.md#regla-nueva-en-verificationmd).
**Verificación:**
- En `content/skills/flow-build/references/verification.md`, cada búsqueda devuelve una línea de las viñetas nuevas:
  - AC1: `rg -n "merge or push you performed to the base branch"`, `rg -n "filter that triggers it"`, `rg -n "does not show what a job tests"`, `rg -n "no run listed yet"`, `rg -n "give the run's link as a reminder"`, `rg -n "closing of the change folder wait"`.
  - AC2: `rg -n "failed jobs"`, `rg -n "previous run already had"`, `rg -n "this rule authorizes that rerun"`.
- AC3: `rg -n "also check the CI it triggered" content/skills/git-workflow/SKILL.md` devuelve una línea.
- Leer el texto final contra los dos casos de la tabla. ARK-730 debe dar "esperar, relanzar una vez ante el timeout, el ticket espera". ARK-734 debe dar "no esperar, recordatorio, el ticket se mueve". Si algún caso da otro resultado, se corrige el texto antes de seguir.

### T2 — Comentario en #43

- [x] Comentario en #43 con D1-A, D2-A, D5-A y D6-A, el caso de ark ARK-734 como segunda evidencia ([design.md](design.md#contexto-verificado)) y el commit de T3.

**Evidencia (T2):** [comentario en #43](https://github.com/JhonHawk/tricell-hive-private/issues/43#issuecomment-5854590714). `review-task` nativo dio AC4 como `met`: el comentario contiene `71fcebe`, ARK-734, D1-A, D2-A, D5-A y D6-A.

**Closes:** AC4.
**Depende de:** T3 (el comentario cita el commit integrado).
**Ubicaciones:** [#43](https://github.com/JhonHawk/tricell-hive-private/issues/43).
**Ejecución:** hilo principal; es una escritura en el tracker, que se queda en el hilo principal.
**Enfoque de prueba:** `check`: el último comentario de #43 contiene cada elemento de AC4.
**Verificación:** `gh issue view 43 --json comments -q '.comments[-1].body'` contiene el SHA del commit de T3, "ARK-734", "D1-A", "D2-A", "D5-A" y "D6-A".

### T3 — Verificación conjunta y entrega

- [x] `go vet ./...` y `go test -race -count=1 ./...` en verde, en un worktree limpio de `ec20d91` con solo los dos archivos de T1. En el árbol principal, `go vet` fallaba en `tests/pilot/flows_test.go:163` (`undefined: skillReadIndex`) por ediciones sin commit de la otra sesión. Esa sesión las integró en `6c3a78d`, y sobre esa cabeza `go vet ./...` y `go test -race ./tests/...` pasan con este cambio.
- [x] Revisar `git status` y `git diff`: no había hunks ajenos en esos archivos (la otra sesión integró lo suyo en `6c3a78d`); no hizo falta el stage parcial. Si hay hunks ajenos sin commit en los dos archivos de T1, se hace stage solo de los propios con un parche aplicado por `git apply --cached`.
- [x] Stage por ruta explícita, nunca por directorio (`git-workflow/SKILL.md:20`): `content/skills/flow-build/references/verification.md`, `content/skills/git-workflow/SKILL.md` y `_support/openspec/changes/gh-43-post-merge-ci-coverage/{proposal,design,tasks}.md`. `git diff --cached --stat` debe listar solo esas cinco rutas. Después, gitleaks sobre lo que está en stage. **Evidencia:** cinco rutas en stage, gitleaks sin hallazgos.
- [x] Commit cuyo mensaje nombra las dos sesiones de [design.md](design.md#procedencia-en-el-commit) con su falla. `git pull --ff-only`, push a `rebuild/harness-engineering`, y confirmar que la cabeza remota es el commit. **Evidencia:** `71fcebe` sobre `6c3a78d`; la cabeza remota es `71fcebe`.
- [x] Release desde un worktree limpio del commit, con el plan en el directorio temporal de la sesión:
  - `go run ./tooling/cli plan install --hosts codex,claude,grok,pi,opencode,cursor --scope user --out <plan absoluto>`
  - `go run ./tooling/cli apply --plan <plan absoluto>`
  - `go run ./tooling/cli status --hosts codex,claude,grok,pi,opencode,cursor --scope user` reporta los recursos `installed`.
  - `rg -l "also check the CI it triggered"` encuentra el `git-workflow/SKILL.md` desplegado de cada host.
  - Después, `git worktree remove` del worktree y borrado del plan.
  - **Evidencia:**
    - Release `112fed74f1db` desplegada con el plan `a1671811edca`: 437 recursos `installed` y 0 en otro estado.
    - Los hosts leen `git-workflow/SKILL.md` y `verification.md` de una sola copia en `~/.agents/skills/`, cuyos consumidores son claude, codex, cursor, grok, opencode y pi. Las dos frases están en esa copia.
    - Worktree y plan borrados.
- [x] T2. Después, el cierre:
  - estado cerrado en `proposal.md` con el commit;
  - `git mv` a `changes/archive/2026-09-27-gh-43-post-merge-ci-coverage`;
  - commit de cierre por rutas explícitas y push;
  - cierre de #43.

**Depende de:** T1.
**Ubicaciones:** el árbol de trabajo, `_support/openspec/changes/gh-43-post-merge-ci-coverage/`.
**Ejecución:** hilo principal; dueño de los efectos de Git, release y tracker.
**Verificación:**
- `go test` en verde.
- `git show --stat <commit>` lista solo las cinco rutas.
- `status` reporta `installed`.
- `rg -l` encuentra la frase en cada host.
- `git ls-remote origin rebuild/harness-engineering` coincide con el commit de cierre.

## Verificación compartida

No hay superficie de UI ni runtime. La comprobación de comportamiento es la lectura del texto contra los dos casos reales (T1). El efecto en sesiones reales no se mide mientras los pilotos sigan pausados.

## Estado de revisión y progreso

**Revisión del plan:** dos revisores `review-plan` nativos, en paralelo y solo lectura, sobre la revisión `221de4fa8713` (SHA-256 de los tres archivos concatenados, truncado a 12). Un dominio fue infraestructura y entrega; el otro, calidad de guía para los modelos de ejecución.

**Hallazgos aceptados y aplicados** (propuestos por los propios revisores, conciliados sin nueva ronda):
- **Relanzamiento autorizado.** El relanzamiento lo autoriza la regla, solo para jobs de test. Antes colgaba del modo de entrega, que no lo cubre.
- **Corrida que aún no aparece.** Se trata como duda y se espera.
- **Turno que termina con la corrida abierta.** Queda como pendiente.
- **Retención explícita.** La regla nombra el movimiento del ticket y el cierre del change folder.
- **Jobs de despliegue.** Quedan bajo la verificación de despliegue.
- **T2 cubre todo AC4.** Su verificación comprueba ahora todos los elementos.
- **AC1 mejor comprobado.** Se agregan búsquedas de sus frases clave.
- **Stage por ruta explícita.**
- **Comandos de estado y de borrado del worktree.** Se escriben completos.
- **Commit base.** Pasa a `ec20d91`, y el procedimiento de stage parcial queda como contingencia.

**Decisiones del usuario tras la revisión:**
- D5-A (incluir los push directos a la rama base).
- D6-A (un fallo previo de la base es un hallazgo incidental).

**Sin ronda adicional:** todo cambio posterior a la revisión aplica la propuesta de un revisor tal como la hizo, incluidos D5-A y D6-A, que el usuario eligió entre las opciones que los revisores plantearon. Por eso no hace falta otra ronda.

**Límites:** los revisores no reverificaron los datos de sample-project ni de ark; los de ark los observé yo con `gh`.

**Implementación:** completa. T1 y T2 verificados por `review-task`; T3 entregado en `71fcebe`, con la release `112fed74f1db`. El efecto en sesiones reales no se midió porque los pilotos están pausados.
