# Monorepo cutover playbook — Turborepo + pnpm lane

> Generalized from the ARK cutover (2026-08; the audited case lives in
> `ark-specs/sessions/2026-08-17-monorepo-cutover/`). Placeholders: `<org>` (GitHub org),
> `<project>` (product name), `<scope>` (npm scope), `<svc>` (a runtime service).
> This is the **JS/TS lane**. A polyglot monorepo re-derives the tool table at intake
> (orchestrator decision); the section ORDER still applies.

## 1 · Why do it — and when not to

A monorepo resolves three failures of one product split across remotes:

1. **Drifting contracts.** A wire-types package published to a registry, each consumer
   pinning a version, local `node_modules` patched by overlay — until the overlay misses
   Docker or the registry denies the token. `workspace:*` in one tree removes the round-trip.
2. **Changes that cross the split.** A new API field + its consumer = two PRs, two CIs,
   two promotions. A shared SHA lands contract and both sides in the same merge to `qa`.
3. **CI/CD lying about the graph.** Per-repo Dockerfiles with stale lockfiles install
   something other than what the workspace runs. The Turbo format is `turbo prune
   --docker`, not "copy the old Dockerfile and pray".

NOT sufficient reasons: "looks modern"; one pipeline redeploying everything per push;
renaming images/repos to match the monorepo name (image names are
`<project>-<component>`, decoupled from the git remote).

## 2 · What unifies, what stays out

| Into the monorepo | Stays out |
|---|---|
| Runtime services (API, worker, apps) | Specs / épicas (`<project>-specs`) |
| Contracts package and facades | Visual mocks (separate baseline) |
| CI, image CD, frontend-host config | Account/VPS Terraform if it lives in specs/infra (copied for apply, never re-applied during cutover) |
| Canonical QA compose | Real secrets (workspace `_support/secrets/`) |
| Quality layer: ONE lint/format config, ONE hook runner, ONE name per equivalent task script | The per-repo variants of it — they do not survive the hoist as-is |

Old remotes are **archived**, never deleted — history, issues, and registry packages tied
to them survive. No new work lands there.

## 3 · Decisions to close before moving code

Without these, the hoist produces a repo that cannot deploy.

| Decision | Recommended default | Cost of choosing wrong |
|---|---|---|
| Contracts resolution | `workspace:*` in-repo; registry publish optional; overlay/pin recovery-only, then deleted | Docker/CI keep talking to the registry; auth failures return |
| Image names | **Do not rename**: `ghcr.io/<org>/<project>-<component>` | VPS compose, rollback tags, pull secrets break |
| Registry push auth | Monorepo `GITHUB_TOKEN` + **Manage Actions access → Write** per package | Source + Inherit alone do NOT grant it (§6) |
| CD | Path-filtered per slice; never "deploy all" | A docs change redeploys every service |
| CI per stage | PR→`development` = quick checks · push `development` = full suite · PR `development→qa` = branch policy only · deploy builds the image | Slow promotions, hung checks on docs PRs |
| Local compose | ONE cwd; the Postgres bind mount is pointed or symlinked, never blindly moved | An `up` from the new path creates an empty data dir |
| Quality stack, when the sources disagree | Pick ONE before the hoist (whichever the larger surface already uses) and converge in the same PR | Each app keeps its own; "temporary" becomes permanent and no gate is enforceable across the repo |
| Production | Out of the cutover; QA green first | Go-live mixed with repo migration |

Per-project (re-decide each time): stack versions from the DESTINATION lockfile; domain
patterns (encryption blob shape, WS auth) if not already closed there.

## 4 · Tool anchors (validated versions, re-anchor to the destination)

| Piece | Anchor / note |
|---|---|
| pnpm | `packageManager`-pinned; `nodeLinker: hoisted` (Amplify SSR rejects pnpm's symlinked `node_modules`) |
| turbo | 2.10.x; MIT CLI; Vercel Remote Cache (free on all plans, apps need not host there) + HMAC signature |
| Node | Match engines; alpine build images + distroless runtime, digest-pinned |
| `turbo prune --docker` | Emits `out/json` (manifests+lock) and `out/full` (graph sources) — THAT is the install/build context |
| Actions | Quick checks hosted; full suite self-hosted; `pnpm/setup` with `dest: ${{ runner.temp }}` (shared `$HOME`) |
| GHCR | Org packages, granular; CD login = `docker/login-action` + `secrets.GITHUB_TOKEN`; VPS pull = member PAT (a different credential) |
| Amplify (if used) | One `amplify.yml`, `appRoot` per app, `cd ../..` ONCE in preBuild (cwd persists within a phase) |
| `eslint-plugin-turbo` | Only `turbo/no-undeclared-env-vars`; mind release-age cooldowns |
| sherif | Version drift across `package.json`s; ignore pre-existing drift explicitly |

Remote Cache is NOT required for prune or `workspace:*` — the cutover closes without it.

## 5 · Execution sequence (do not compact steps)

### 5.1 Hoist

1. Create `<org>/<project>-monorepo`, branches `development` → `qa` → `production`.
2. **Import histories** (never recreate); multiple `initial commit`s in the log is normal.
3. Layout: `apps/<svc>/`, `packages/contracts/` (`@<scope>/<project>-contracts`),
   `packages/<facades>/`, `infra/qa/` (compose+proxy), `infra/terraform/` (copy for
   apply), `.github/workflows/`, `turbo.json`, `pnpm-workspace.yaml`.
4. Move CI/CD/frontend config in. Docker build context is the **root**, not `apps/<svc>`.
5. Promote to QA **before** archiving remotes. Verify health + one real login.

### 5.2 Contracts by workspace

1. Each consumer: `"@<scope>/contracts": "workspace:*"`.
2. Regenerated lockfile must show `version: link:../../packages/contracts` — **no**
   registry tarball.
3. Drop the scope's `registry=` line from `.npmrc` once nothing else needs it.
4. `publishConfig.registry` stays only in the contracts `package.json` (publishing optional).
5. **Delete** overlay / pin / assert-published from the daily path — kept "just in case",
   they become the default again.

### 5.3 Images: `turbo prune`, never the leftover Dockerfile

Reusing the split-repo Dockerfile fails three ways: install against the old lockfile
(registry 403), `ERR_PNPM_IGNORED_BUILDS` (allowBuilds incomplete), runtime missing
`contracts/dist` (`ERR_MODULE_NOT_FOUND`). The Turbo-format skeleton:

```dockerfile
FROM node:<ver>-alpine@<digest> AS base
WORKDIR /app
RUN npm i -g pnpm@<lock> turbo@<lock>

FROM base AS prepare
COPY . .
RUN turbo prune <package-name> --docker

FROM base AS deps
COPY --from=prepare /app/out/json/ .
RUN pnpm install --frozen-lockfile

FROM base AS build
COPY --from=deps /app/node_modules ./node_modules
COPY --from=prepare /app/out/json/pnpm-lock.yaml ./pnpm-lock.yaml
COPY --from=prepare /app/out/full/ .
RUN pnpm exec turbo run build --filter=<package-name>...

FROM base AS prune
COPY --from=prepare /app/out/json/ .
RUN pnpm install --frozen-lockfile --prod
# nodeLinker: hoisted — Node may not resolve the workspace link from the app WORKDIR;
# fix the link at the root:
RUN mkdir -p node_modules/@<scope> \
 && ln -sfn ../../packages/contracts node_modules/@<scope>/<contracts-pkg>

FROM gcr.io/distroless/nodejs<ver>-debian12:nonroot@<digest>
WORKDIR /app/apps/<svc>
# COPY node_modules, app dist, contracts dist
```

Migrations: distroless runtime does NOT run them — a `deps-<sha>` image + mounted
`migrations/`, or `prisma migrate deploy`. Backup BEFORE migrate. Post-deploy assert:

```bash
RUNNING=$(docker inspect --format '{{.Config.Image}}' "$(docker compose ps -q <svc>)")
test "$RUNNING" = "$IMAGE:$TAG"
```

A healthz 200 can be the OLD container. The assert is not optional.

### 5.4 CI

- `fetch-depth: 0` when turbo filters against a base ref.
- Self-hosted runners: never write to `$HOME`.
- Full suite skips `next build` if the frontend host builds the apps (missing
  `NEXT_PUBLIC_*` otherwise).
- Import boundaries script: turbo has no Nx-style tags; an `apps/a → apps/b` import makes
  prune lie.

### 5.5 Promotion

feature → `development` (checks, merge) → full suite on push → PR `development → qa`
(head branch name must match policy) → branch policy only, no CI re-run → path-filtered
CD builds and deploys → verify VPS (tags + `docker inspect`) AND public health AND a real
UI login.

### 5.6 Turbo hygiene (separate PR, after QA green)

| Piece | Why |
|---|---|
| `--affected` | Replaces hand-rolled git filters; reads `GITHUB_BASE_REF` in PRs |
| `eslint-plugin-turbo` (`no-undeclared-env-vars` only, `--no-config-lookup --no-inline-config`) | A task green on miss and red on hit — undeclared env is the classic cache poison; declare env in `turbo.json` |
| sherif | Catches the next framework split-version; ignore pre-existing drift explicitly |
| Remote Cache | Ephemeral runners lose local cache. `TURBO_TOKEN` + `TURBO_REMOTE_CACHE_SIGNATURE_KEY` as Actions secrets, `TURBO_TEAM` as variable, `"remoteCache": {"signature": true}`. HMAC key is `openssl rand -hex 32`, never the Vercel token itself |

Vercel token: `vercel tokens add` from the CLI can 403 — create it in the UI
(vercel.com/account/tokens, team scope). Store per `security.md` (secret manager; `op`
else keychain). Seed repos via `gh secret set TURBO_TOKEN` from stdin. Without the token
the hygiene PR still merges — the cache is opt-in.

NOT in this batch: pnpm catalogs, Changesets, Knip, CODEOWNERS, Dependabot, Docker layer
cache, provenance, Nx.

## 6 · Registry (GHCR): the residue that hurts most

- **Never rename the package.** The git remote may be `<project>-monorepo`; images stay
  `<project>-<component>`. Different axes.
- **Re-link, don't recreate.** After the hoist the package still points at the archived
  remote → monorepo `GITHUB_TOKEN` push = 403. A member PAT works but is a bridge, not
  the end state. The REST API does not expose "Manage Actions access" — UI only.
- Three separate controls: **Repository source** (ownership label) ≠ **Inherit access**
  (people/teams) ≠ **Manage Actions access → Write** (what lets the workflow push). Only
  the third grants the push.
- UI path: package settings → remove archived source → Add Repository with an **empty
  search box** (typing filters to "No repositories found" even when the repo exists) →
  mark the monorepo → **Write** (default is Read) → reconnect source + Inherit. Repeat
  per pushed image. `workflow_dispatch` of the deploy is the proof: login 200 + push 403
  = missing Write, not a YAML permissions problem.
- VPS pull stays a member PAT; don't delete it until another pull mechanism exists.

## 7 · Local

Canonical compose in the app that owns the DB. If the old data volume lives in an
archived clone with a relative bind mount: symlink it (`ln -sfn <old-abs-path>
apps/<svc>/.docker-data/postgres`), gitignored. Never `docker compose down -v` to "fix"
an empty-volume symptom — it's a wrong-cwd symptom.

## 8 · Gotchas (each one happened)

| Symptom | Real cause | Do NOT |
|---|---|---|
| Docker 403 to the npm registry | Leftover lockfile / scope `.npmrc` | Reintroduce the overlay |
| `ERR_PNPM_IGNORED_BUILDS` | Incomplete `allowBuilds` in the leftover | Blanket `allowBuilds: true` |
| `ERR_MODULE_NOT_FOUND` contracts at runtime | `nodeLinker: hoisted` + app WORKDIR | Hand-copy `dist` "just this once" |
| Healthz 200 after deploy | The PREVIOUS container | Skip the image assert |
| `turbo ...[ref]` fails in CI | `fetch-depth: 1` | Widen the filter to "everything" |
| Full suite OOM / missing env | `next build` on self-hosted | Silence the job |
| Registry push 403 with workflow token | Actions access ≠ Inherit | Fall back to the PAT and forget |
| "No repositories found" adding the repo | The search box; or source = same repo | Create a new package / rename |
| Actions role Read | The default when adding | Leave it "because it's listed" |
| Empty local DB | Different cwd, different bind mount | `docker compose down -v` |
| Remote cache never hits | Missing `TURBO_TOKEN`/`TURBO_TEAM` | Rip out the YAML — cache is opt-in |
| Cutover "done", tooling half-consolidated | Verify proved the new tree deploys, never that the old one stopped existing | Run the §9 residue sweep before closing |

Deploy SSH: never feed the remote via `ssh … bash -s <<'EOF'` when an inner command reads
stdin (`docker compose exec -T` eats the script and exits 0). Use `bash -c "$(cat)"`.

## 9 · How to know the migration is done

- [ ] Lockfile: contracts as `link:`, zero registry tarballs.
- [ ] `turbo prune <app> --docker` in the Dockerfile that QA actually runs.
- [ ] CD logs in with the workflow token and the push is NOT 403.
- [ ] Deployed container `IMAGE:TAG` = SHA of the merge to `qa` (inspect, not healthz).
- [ ] Health endpoints 200 AND a real UI login.
- [ ] Old remotes archived; docs and `AGENTS.md` point at the monorepo.
- [ ] Overlay / pins / legacy trees deleted, not parked as "recovery".
- [ ] Production explicitly out, or a separate go-live plan exists.
- [ ] Quick checks use `--affected`; sherif + env-hash checks in CI.
- [ ] `turbo.json` declares `globalEnv`/test env and `remoteCache.signature`.
- [ ] (Optional) remote cache secrets set; one CI run shows a cache HIT.

### Residue sweep — did the consolidation actually consolidate?

QA green proves the new tree deploys. It does not prove the old one stopped existing. Every
category below was found in a monorepo that had passed its own cutover months earlier —
sweep all four before closing:

- [ ] **No dangling reference.** Nothing copies, excludes, or points at a path the hoist
      moved: `Dockerfile*` COPY sources, `.dockerignore`, CI path filters, script cwd
      assumptions. A dev image nobody builds in CI rots in silence.
- [ ] **No surviving duplicate.** What is shareable now exists once: per-app scripts
      differing only in constants, byte-identical configs, install/rebuild steps repeated
      across image stages, a second formatter still installed in one app.
- [ ] **The adopted capability is ON, not merely configured.** CI invokes tasks THROUGH the
      orchestrator (never the package manager directly), every task declares its `inputs`,
      and the checkout does not wipe the cache dir — a re-run of an untouched task must hit.
      Paying for the cutover and leaving the cache off is the most expensive outcome available.
- [ ] **Docs match the tree.** `AGENTS.md`/`CLAUDE.md`/READMEs declare nothing pending that
      is already done, nor done what is not. A ledger that lies is the costliest residue:
      it misroutes every agent and person who reads it next.

## 10 · What not to copy from the source case

- Its admin-merge standing authorizations — workspace-specific, never a default.
- The member-PAT push as a design (bridge only).
- Registry publishing of contracts as a daily step (optional in-monorepo).
- Moving the DB volume "so the path looks nice" inside the hoist change.
- `terraform apply` as a cutover side effect.
