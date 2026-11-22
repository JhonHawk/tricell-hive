# Design

## Approaches

Solo hay una forma razonable para la estructura: carpetas de dominio en el primer nivel de `hive-core`, en lugar de `references/`.

- **Elegida — `hive-core/<dominio>/<hoja>.md`.** Reusa el almacén oculto que ya existe (`hive-core`, peldaño 2 de «Reuse before writing»), y el retiro genérico del manager. Las rutas quedan igual de largas o más cortas: `global.md` baja de 29.617 a 29.210 bytes, medido aplicando la tabla de abajo a sus 85 rutas.
- **Descartada — `hive-core/references/<dominio>/<hoja>.md`.** Suma unos 9 bytes por ruta, ~750 B en `global.md`, que solo tiene 383 libres.
- **Descartada — una skill oculta por dominio.** Suma entradas de catálogo en Antigravity y quizá en otros hosts.

Para el manager (D6) se presentaron tres formas: un manifiesto de dominios (elegida), nombres fijos en Go y cualquier subcarpeta. El manifiesto sigue el precedente de `assets/stage.json`: el manager lee un JSON de la skill y solo actúa sobre lo declarado.

## Decisions

Las decisiones D3 a D7 y los supuestos están en `proposal.md`.

## Proposed shape

| Dominio | Hojas |
| --- | --- |
| `messages` | questions, reasons-and-blockers, report-endings |
| `git` | delivery-ask, foreign-commits, git-commit, promotion-path, merge-gate, pr-reviews |
| `plan` | change-folder, checkpoint, scope-changes, project-settings*, reuse-before-writing*, external-facts, external-effects, loops* |
| `build` | tests, tests-no-runner, defect-fix, stored-data*, security-boundaries*, checks, background-wait, verification-floor, prose*, project-verification, project-verification-format |
| `delegation` | delegation, child-brief, reviewer-brief, unattended-effects* |
| `ui` | design-stage, ui-plan, ui-example, ui-verification, before-captures, a11y-audit, walkthrough, viewer, viewer-open, viewer-close |
| `review` | no-comments, blast-radius, suppression-review, review-evidence, review-findings, interrogate |
| `business` | business-rules, business-record, business-ticket |
| `close` | close-modes, close-hygiene, close-report |
| `local` | local-record, local-write, local-script, test-accounts |
| `tracker` | ticket-state, trackers |
| `principles` | index, boundary-discipline, explain-the-number, make-operations-idempotent, migrate-callers-then-delete-legacy-apis, minimize-reader-load, model-the-domain, separate-before-serializing-shared-state, type-system-discipline, verify-current-state |

`*` = encaja en dos dominios; va donde se lee primero: `project-settings`, `reuse-before-writing` y `loops` se leen al planear; `stored-data`, `security-boundaries` y `prose` antes de escribir el cambio; `unattended-effects` es trabajo delegado.

```
content/skills/hive-core/
  SKILL.md        # description and body name the domain folders; no links to leaves
  assets/
    hive-managed.json
    domains.json  # {"domains": ["messages", "git", …, "principles"]}
  scripts/        # unchanged: scripts/<leaf>/ (checks, trackers)
  messages/ git/ plan/ build/ delegation/ ui/ review/ business/ close/ local/ tracker/ principles/
  build/checks/run-check.md        # a leaf's own references move with it
  tracker/trackers/{github,jira,linear,time}.md
  local/local-write/worktree.md
  messages/questions/unanswered.md
  ui/ui-verification/device-tool.md
```

### Manager

- `assets/domains.json` is a JSON object with one key, `domains`: an array of kebab-case names, unique, none of them `references`, `scripts`, `assets`, `playbooks` or `agents` (Codex's generated `agents/openai.yaml` lives there), each an existing, non-empty folder of the skill. Any skill may hold one; only `hive-core` does.
- Collection from a checkout (`plan.go:239-252`) adds the folders the manifest lists to the four fixed ones. A folder it does not list is not collected.
- `validateRelease` (`plan.go:420-479`) reads and checks the manifest from the release's own files, for a checkout and for a stored release (`hive install --release <id>`) alike, as it does for `stage.json`: every file in a non-fixed first-level folder belongs to a listed domain, and each listed domain holds a file. A malformed manifest, or a file outside the listed domains, rejects the release before any write, naming the manifest and the reason.
- The source shape check (`validSource`), which also guards recorded state (`ownership.go:64`), accepts a path under a kebab-case first-level folder of a skill; which folders ship is decided by collection and `validateRelease`, so the check stays a pure function of the path.
- Link validation (`validateInstructionReferences`) receives the release's domains, so `instructionMarkdown` also covers Markdown under a listed folder.

### Content move

- One local script reads the table and builds a map from every old file path to its new one: `hive-core/references/<leaf>.md` → `hive-core/<domain>/<leaf>.md`, `hive-core/references/<leaf>/<f>` → `hive-core/<domain>/<leaf>/<f>`, and `hive-principle-<n>/SKILL.md` → `hive-core/principles/<n>.md` (`hive-principle-index` → `principles/index.md`). It runs `git mv` by that map, then rewrites each link by resolving it, not by string rules: it resolves the link's old target from the source file's old location (for a `skill:` locator, from `content/skills/`), maps the target through the map, and writes the new link from the source file's new location, keeping the locator kind and the anchor. A target the map does not hold, such as `../scripts/…` or another skill, keeps its file and only gets a new relative path when its source moved. This covers the principles, which sit one level deeper than their old `SKILL.md`, and their links into leaves.
- The same map rewrites the paths written as text, outside links, in `content/`, `site/`, `README.md`, `llms.txt`, `AGENTS.md` and `_support/docs/`, such as `stage.json` entries and code spans; prose that describes the structure is rewritten by hand in T6.
- The principles lose their frontmatter and their `hive-principle-` prefix. Their text changes only in link targets.
- `assets/stage.json` of each flow lists the new paths, so packet headings become `## hive-core/<domain>/<leaf>`.
- The session audit (`tests/pilot/session_audit*.go`) recognizes a leaf read at its two old paths (`hive-rule-<n>/SKILL.md`, `hive-core/references/<n>.md`) and at the new one, since it audits sessions recorded before the change.
- `tests/content/leaf_links_test.go` registers the principle index by its path `principles/index.md`, not the bare name `index`, which would match any `index.md` link.

## Impact

`rg` finds `hive-core/references` or `hive-principle-` 659 times in 95 files of `content/`, 240 times in 27 test files, 50 in 4 `tooling/` files (comments and tests), 51 in 12 `site/` pages, 48 in 6 `_support/docs` files, plus `README.md`, `llms.txt`, `AGENTS.md` and 2 specs; and 149 relative links between leaves. The installed paths change for every host: a session opened before the update keeps the old paths in its loaded guidance and must restart. The `- Rules:` and `- Principles:` checkpoint lines name leaves without a path, so they do not change. Removing 10 skills is a minor version change (`CONTRIBUTING.md` «Releases and version numbers»).

## Recovery

Revert the commit, or under `hold` discard the working tree. After an installation with the new binary, the state records the domain paths, and a binary older than this change fails reading that state (`ownership.go:64`, "invalid recorded source") for every command, `status` and `remove` included, not only `update`. A rollback therefore keeps the new binary and reinstalls the previous release with it (`hive install --release <id>`, or `hive update --source <checkout>` at the reverted commit), which installs what that release holds and retires the rest. The binary is rebuilt before content is deployed, as `AGENTS.md` «Refreshing a local installation» sets for changes to both.
