# Diseño — master-promotion-guard

## Check de origen

Archivo `.github/workflows/master-source-branch.yml`:

```yaml
name: master-source-branch
on:
  pull_request_target:
    branches: [master]
    types: [opened, synchronize, reopened, edited]
permissions: {}
jobs:
  source-branch:
    name: source-branch
    runs-on: ubuntu-latest
    timeout-minutes: 2
    steps:
      - name: Require development of this repository as the source
        env:
          HEAD_REF: ${{ github.event.pull_request.head.ref }}
          HEAD_REPO: ${{ github.event.pull_request.head.repo.full_name }}
          BASE_REPO: ${{ github.repository }}
        run: |
          if [ "$HEAD_REPO" = "$BASE_REPO" ] && [ "$HEAD_REF" = "development" ]; then
            echo "Source is $HEAD_REPO:$HEAD_REF"
          else
            echo "::error::master accepts pull requests only from development of $BASE_REPO; this one comes from $HEAD_REPO:$HEAD_REF"
            exit 1
          fi
```

- **`pull_request_target`, no `pull_request`:** el workflow sale de la rama por defecto (`development`), así que un PR no puede reescribir el check para que pase. No hay checkout ni secretos, y `permissions: {}` deja el token sin permisos. Con eso se cumplen las condiciones de uso seguro que da la documentación de GitHub.
- **Sin `edited`** (cambio por la revisión de código): si alguien cambia a `master` la base de un PR que ya existe, el check no corre, falta el check obligatorio y el merge queda bloqueado hasta que un push lo dispare. Un filtro `if` habría dejado un job saltado, y un job saltado cuenta como éxito.
- **El ruleset reconoce el check solo por su nombre:** cualquier job llamado `source-branch` en cualquier workflow cuenta, y fijar `integration_id` no los distingue. `CONTRIBUTING.md` lo dice, y el riesgo de los forks se trata aparte (decisión D6).
- **Variables de entorno:** el nombre de la rama llega por `env`, nunca interpolado dentro de `run`. Así un nombre de rama no puede inyectar comandos.
- **Repositorio de origen:** el check también exige que el origen sea este repositorio, porque un fork puede llamar `development` a una de sus ramas.

## Ruleset

Se crea con `gh api -X POST repos/JhonHawk/tricell-hive/rulesets --input ruleset.json`:

```json
{
  "name": "master: promotion from development",
  "target": "branch",
  "enforcement": "active",
  "conditions": { "ref_name": { "include": ["refs/heads/master"], "exclude": [] } },
  "bypass_actors": [
    { "actor_id": 5, "actor_type": "RepositoryRole", "bypass_mode": "pull_request" }
  ],
  "rules": [
    { "type": "deletion" },
    { "type": "non_fast_forward" },
    { "type": "pull_request", "parameters": {
        "required_approving_review_count": 0,
        "dismiss_stale_reviews_on_push": false,
        "require_code_owner_review": false,
        "require_last_push_approval": false,
        "required_review_thread_resolution": false,
        "allowed_merge_methods": ["merge"] } },
    { "type": "required_status_checks", "parameters": {
        "strict_required_status_checks_policy": false,
        "required_status_checks": [ { "context": "source-branch", "integration_id": 15368 } ] } }
  ]
}
```

- **Bypass:** `actor_id: 5` con `RepositoryRole` es el rol admin; con `bypass_mode: pull_request` el admin solo puede saltarse las reglas mergeando un PR (D2 B). La documentación revisada no confirma que el id 5 sea el rol admin; es conocimiento de la comunidad. Por eso, después de crear el ruleset, T3 lee `current_user_can_bypass` como owner y espera `pull_requests_only`. Si da otro valor, corrige el actor antes de seguir.
- **Cero aprobaciones:** el único maintainer no puede aprobar su propio PR.
- **Solo merge commit:** las promociones usan merge commits. Un squash o un rebase hacia `master` generaría commits que `development` no tiene, y la siguiente promoción entraría en conflicto.
- **`strict: false`:** `master` tiene merges que `development` no tiene. Con `strict` activado, cada promoción obligaría primero a traer `master` a `development`.
- **`integration_id: 15368`:** es la app GitHub Actions. Así el check no se puede satisfacer con un status que publique otra fuente. T3 confirma el id en el `app.id` del check-run real antes de crear el ruleset.

Para revertirlo: `gh api -X DELETE repos/JhonHawk/tricell-hive/rulesets/<id>`. El id se anota en `tasks.md` cuando se crea, y llega al repositorio con el commit de cierre del change record (flow-close), no con un commit aparte a `development`.

## Hotfix

Después de este cambio, un PR de hotfix hacia `master` siempre falla `source-branch`. Solo entra con el bypass de admin, y solo cuando el maintainer lo decide de forma explícita para ese PR. `CONTRIBUTING.md` lo dice así, para que un agente no lea el check fallido como "la política bloquea el hotfix" según `urgent-fixes.md:7`.
