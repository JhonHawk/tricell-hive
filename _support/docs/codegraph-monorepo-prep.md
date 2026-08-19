# CodeGraph Monorepo Prep

Preparing a repo so CodeGraph's edges point at the real target, indexing it, keeping the index
in step with git, and retiring the indexes a monorepo cutover leaves behind. Hand this to the agent working in that repo — the
tsconfig part is a per-repo change, not a global one. Verified against CodeGraph v1.5.0 by
indexing a synthetic pnpm workspace and reading `.codegraph/codegraph.db` directly.

**Indexing is the user's decision.** Prepare the `paths` and stop; run `codegraph init` only
when the user asks for it.

## Does this apply to my repo?

| Repo shape | Action |
|---|---|
| Monorepo — `pnpm-workspace.yaml`, or `workspaces` in the root `package.json` | Apply the recipe below, then index the workspace root |
| Single app in its own repo | No prep. Its own `tsconfig.json` IS the index root, so its `paths` are already read — go straight to **Run it** |
| Monorepo with non-TS members | Prep covers the TS/JS members; the rest index without it (`paths` is a TypeScript mechanism) |

Never index per app inside a monorepo (`codegraph init apps/web`): `packages/` falls outside
the tree and workspace resolution never activates. One index, at the workspace root.

## Why the prep is needed

CodeGraph resolves an import by path in two steps, then falls back:

1. **tsconfig `paths`** — read from `tsconfig.json` (or `jsconfig.json`) **at the index root
   only**, and **`extends` is not followed**. Aliases declared in `apps/<x>/tsconfig.json`
   are invisible; so are `paths` that live in an extended base file.
2. **Workspace member packages** — read from `pnpm-workspace.yaml` / `workspaces`, resolved
   **by directory**: `@scope/lib/http` → `packages/lib/http`, then `.ts`/`.tsx`/`.js` and
   `/index.ts`. A member's `exports` and `main` are **ignored**, so a package that serves
   from `src/` or `dist/` does not resolve.
3. **Fallback: name matching** — when both fail, the edge is still created, to the
   same-named symbol nearest the importing file.

Step 3 is why an unprepared monorepo looks fine: most edges land correctly because the name
is unique. Where it is not — `Button`, `formatCurrency`, `AuthService`, the names a merge of
N repos multiplies — the edge points into the wrong app, and `callers` / `impact` /
`affected` report it with full confidence.

## Derive the map from disk

Do not guess the entry points. List every member's declared name and entry, then translate
each entry to the **source** file it is built from:

```bash
for f in packages/*/package.json apps/*/package.json; do
  python3 -c "
import json; d = json.load(open('$f'))
print(d.get('name'), '|', d.get('exports') or d.get('main') or d.get('types') or '(none)')"
done
```

An entry under `dist/`, `build/`, or `out/` is a build artifact — those directories are
excluded from the index by default, so map the package to its `src/` equivalent instead
(`./dist/http/backend.js` → `./src/http/backend.ts`, i.e. `"@scope/pkg/*": ["./packages/pkg/src/*"]`).

## The recipe

Add `paths` to a `tsconfig.json` **at the workspace root**, one entry per internal package:

```jsonc
{
  "compilerOptions": {
    "paths": {
      "@ark/contracts":              ["./packages/ark-contracts/src/index.ts"],
      "@ark/api-client":             ["./packages/api-client/src/index.ts"],
      "@ark/api-client/*":           ["./packages/api-client/src/*"],
      "@ark/auth-utils":             ["./packages/auth-utils/src/index.ts"],
      "@initech-mira/ark-contracts":   ["./packages/contracts/src/index.ts"],
      "@initech-mira/ark-contracts/*": ["./packages/contracts/src/*"]
    }
  }
}
```

Rules that follow from the mechanism:

- **Declare the bare name AND a `/*` subpath entry** for any package imported with subpaths
  (`@scope/pkg/http/backend`). The bare entry does not cover them.
- **Never declare `@/*` at the root.** It means a different directory in each app and a
  single map cannot disambiguate — the first target wins for every app, silently rewiring
  the others' imports. Leave per-app `@/` to the name-matching fallback, which resolves it
  correctly as long as no homonym exists elsewhere.
- **`paths` must be inline in the root file.** Putting them in a base config the root
  `extends` has no effect.
- **`baseUrl` is optional** (defaults to the root). Targets may be a file or a directory
  (`./packages/lib/src` → `/index.ts` is tried). Comments and trailing commas are tolerated.
- **If a root `tsconfig.json` already exists** (used by the build, an editor, or `tsc`), add
  the `paths` to it rather than creating a second file — CodeGraph only ever reads
  `tsconfig.json` / `jsconfig.json` at the root, under those exact names. If none exists,
  creating one is a real change to the repo: confirm no tool picks it up unintentionally
  (Next.js, `tsc`, IDE project resolution) before committing it.

## Run it

```bash
codegraph init <workspace-root>     # first index (indexing runs by default; -i is deprecated)
codegraph index <workspace-root>    # full rebuild — required after changing paths
codegraph status <workspace-root>   # files / nodes / edges / DB size
codegraph sync <workspace-root>     # incremental catch-up (automate it — see Keep it fresh)
```

Maintenance commands (`init`, `index`, `sync`, `status`, `uninit`) take the path
**positionally**; query commands (`query`, `explore`, `callers`, `impact`, `affected`,
`files`) take `-p <root>` instead. Run from the root and both are optional.

`init` prints `Indexed N files` and a `N nodes, N edges` line — a node count near zero on a
real codebase means the scan found nothing to parse, not a healthy empty graph.

**Git:** nothing to add to the repo's `.gitignore`. CodeGraph writes `.codegraph/.gitignore`
itself, ignoring everything in that directory except that file.

## Verify after indexing

Do **not** verify with the `unresolved_refs` table — it records the path-resolution attempt
and still says `failed` for imports the alias map later resolved. Read the edges:

```bash
sqlite3 <workspace-root>/.codegraph/codegraph.db "
  select e.source, n.name, n.file_path
  from edges e join nodes n on n.id = e.target
  where e.kind = 'imports' and e.source like 'file:apps/%'
  order by n.file_path"
```

Every import made from an app, and where it landed. Read the third column: a symbol you know
lives in a package must show a `packages/…` path there. One showing a path back inside the
importing app is the failure this prep exists to remove — fix that package's `paths` entry
and rebuild with `codegraph index`. Then spot-check one symbol whose name exists in more than
one app: `codegraph callers <Name>` should list only its real callers.

## Keep it fresh — the git hook

The index does not follow git on its own. Verified on a synthetic repo: right after a merge,
`codegraph status` still showed the pre-merge counts and `codegraph query <newSymbol>`
answered **"No results found"** — indistinguishable from "that symbol does not exist".
A `codegraph_explore` MCP call in that state is worse, not better: it re-reads the files from
disk, so the new symbol appears in the printed source while the graph beside it (blast
radius, callers) is still the old one.

`codegraph sync` fixes it (43 ms for one changed file), so the last step is making git call
it. These repos run **lefthook 2.x**, whose `jobs` syntax is:

```yaml
# lefthook.yml — at the repo root
output: [failure]           # no banner on every checkout

post-merge:
  jobs:
    - name: codegraph-sync
      run: codegraph sync -q . || true

post-checkout:
  jobs:
    - name: codegraph-sync
      run: codegraph sync -q . || true

post-rewrite:               # rebase, amend
  jobs:
    - name: codegraph-sync
      run: codegraph sync -q . || true
```

Run `lefthook install` after adding it. Without lefthook, the same three files under
`.git/hooks/` (`chmod +x`), each guarded so a checkout never fails on a machine without
CodeGraph:

```bash
#!/usr/bin/env bash
[ "$3" = "1" ] || exit 0          # post-checkout ONLY: skip file checkouts, run on branch switches
command -v codegraph >/dev/null 2>&1 || exit 0
[ -d .codegraph ] || exit 0
codegraph sync -q . || true
```

- **`|| true` is not decorative.** `codegraph sync` exits **1** in a repo with no
  `.codegraph/` — which is every teammate who has not indexed. Without the guard, lefthook
  reports a failed job on each of their checkouts. (Native `.git/hooks/` are not committed,
  so they only affect you; a committed `lefthook.yml` affects everyone.)
- **Cost:** ~0.2 s per merge/checkout on a small tree; it scales with the number of changed
  files, not repo size.
- The index follows the active branch: a symbol that exists only on a feature branch
  disappears from the index on checkout back to the base branch, and returns on the way back.

### Function test

After installing the hook, prove it end to end — a hook that silently no-ops is the same as
no hook:

```bash
git checkout -b codegraph-hook-check
echo 'export function codegraphHookCheck(): string { return "ok" }' >> packages/<pkg>/src/index.ts
git commit -am "probe" && git checkout - && git merge codegraph-hook-check
codegraph query codegraphHookCheck        # must print the symbol WITHOUT running sync by hand
git checkout codegraph-hook-check && git checkout -   # branch switch, both directions
codegraph status                          # node count moves with the branch
git branch -d codegraph-hook-check
```

Run the probe from a throwaway branch based on your working branch, never from `development`
or an environment branch — the merge leaves a real commit behind, and undoing it means
rewriting history.

The query returning "No results found" means the hook never ran: check `lefthook install` was
executed (or the hook file is executable), and run the hook's command by hand to see its error.

## Retire the indexes the cutover left behind

A repo the monorepo replaced still answers queries from its own stale `.codegraph/`, with
code that no longer ships. Remove it once the repo is archived:

```bash
codegraph uninit <old-repo>   # deletes that repo's .codegraph/ directory
```
