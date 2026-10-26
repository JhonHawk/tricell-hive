# Contributing to Hive

Thank you for considering a contribution. Hive is a small guidance layer and
manager for coding-agent CLIs; changes are kept proportionate to what the project
needs.

## Set up

Install **Go 1.27.0 or later**, as declared in [go.mod](go.mod), plus Python 3 for
the skill tests. Clone the repository and work from its root. The first build
needs access to the Go modules unless they are already cached.

## Run the full local test suite

The repository has no CI, so contributors run the suite locally before proposing
a change:

```sh
go vet ./...
go test ./...
python3 -m unittest discover -s tests/skills -p '*_test.py'
```

Also run `go test -race ./...` when a change touches concurrency: goroutines,
locks, state shared across goroutines, or the manager's concurrent-write tests.
The [deployment manager reference](_support/docs/architecture/deployment-manager.md)
describes what the tests cover.

## Viewer frontend

The standalone React package in `tooling/viewer/frontend/` uses public MIT
shadcn/ui and ReUI components, with no premium blocks. Development requires
**Node 24 or later** and **pnpm 12.8.1**. From the repository root, run:

```sh
pnpm --dir tooling/viewer/frontend install --frozen-lockfile
pnpm --dir tooling/viewer/frontend check
pnpm --dir tooling/viewer/frontend lint
pnpm --dir tooling/viewer/frontend test
pnpm --dir tooling/viewer/frontend build
pnpm --dir tooling/viewer/frontend verify:dist
```

Commit the generated `tooling/viewer/static/app/` bundle and manifest alongside
frontend changes. Rebuild a second time with the frozen lockfile and check that
the output has no diff. The source digest includes the frontend README, tests,
configuration and lockfile, so changes to those files also require regeneration.
`verify:dist` checks freshness without changing output; the Go viewer tests check
the source digest without Node. This detects stale output, not authenticity.
Ordinary `go build ./tooling/cli` and package building consume checked-in output
and do not require Node or pnpm.

`pnpm --dir tooling/viewer/frontend dev` uses the real same-origin API by default.
The generated development composition requires `?preview`; it is absent from the
production bundle. A development preview does not verify the Go server's API,
authentication or CSP. Check the compiled bundle served by Go for those surfaces,
and preserve the shared offline renderer byte for byte. See the
[frontend README](tooling/viewer/frontend/README.md) for the bundle contract and
[third-party notices](THIRD_PARTY_NOTICES.md) for the distributed license notices.

## Branches and commits

- Branch from `development`, the base branch, and open pull requests against it.
  `master` is the release branch and receives changes only by promotion.
- Use [Conventional Commits](https://www.conventionalcommits.org/), such as
  `fix(cli): ...`, `feat(guidance): ...`, or `docs: ...`.
- Keep a change focused, and do not mix unrelated fixes into it.
- Keep `README.md`, `llms.txt`, and the documentation site in `site/` in step
  with your change, in the same pull request, as the
  [README and llms.txt](AGENTS.md#readme-and-llmstxt) section of `AGENTS.md`
  describes.

## Documentation site

The site in `site/` uses Astro Starlight with the Ion theme and Hive's brand
accents. Its navigation is Home, Getting started, and Changelog, followed by a
separator, a collapsible Skills group, and an Inspiration page. Skills links to its catalog
and the guides of the main flows (research, plan, build, verify, security, close and deploy). Inspiration lists OpenSpec, gentle-ai,
Superpowers, and Skills For Real Engineers, in that order. Keep its existing routes, logo assets, and
metadata when updating its presentation. Flow pages explain the user experience and link to the canonical
skill; they do not duplicate the procedural rules in `content/`.

Update the affected flow page in the same pull request as its behavior changes.
Check claims against the canonical skill and its references on `development`,
and distinguish supported behavior from verified host behavior. Update
`README.md` and `llms.txt` too when their claims change.

`CHANGELOG.md` at the repository root is the only editorial source for the
site's release history. Edit it rather than a second site copy. Keep
`[Unreleased]` separate from published versions, preserve the recorded release
dates, and verify historical additions against release tags and release notes.
An entry under `[Unreleased]` is not a published release.

Use the site's declared Node and pnpm versions. From `site/`, run:

```sh
pnpm install --frozen-lockfile
pnpm check
pnpm lint
pnpm build
pnpm preview --host 127.0.0.1
```

Check the affected pages, navigation, search, keyboard access, mobile layout,
and both color themes in the preview. When changing colors or the theme,
remeasure text, links, and button contrast against their actual backgrounds.
Publishing follows the release process below.

## Releases and version numbers

The version number tracks what the installer delivers: the `hive` binary (its
commands and flags), the installed content (shared guidance, skills, and roles),
the supported hosts, and the package format. Changes elsewhere, such as the
documentation site in `site/`, the README, `_support/`, or tests, ship with the
next release but do not change the number by themselves.

While the version is below 1.0:

- Raise the minor version (`0.x.0`) for what users add, remove, rename, or must
  act on: a skill, role, flow, supported host, command, flag, or installation
  route; a setting of the `## Hive` section; or a file layout the guidance
  creates in their projects. Any incompatible change also raises it and needs a
  migration note in the changelog.
- Raise the patch version (`0.x.y`) for everything else in what the installer
  delivers: defect fixes, guidance changes to how agents carry out existing
  flows, and dependency or template version updates.

A change does not release by itself. Each pull request adds its entry under
`[Unreleased]` in `CHANGELOG.md` and leaves `VERSION` at `dev`. A release is cut
only when the maintainer asks for it: a release pull request sets the number and
moves the `[Unreleased]` entries under it, the release is promoted to `master`
and tagged, and a follow-up sets `VERSION` back to `dev`. Agents propose a
release, without cutting it, when `[Unreleased]` holds an entry more than seven
days old, or a fix for a defect in the latest published version.

The documentation site is published from `master`: each release deploys `site/`
from its tag, building it first with `pnpm --dir site build` and then running
`pnpm exec wrangler deploy`; nothing else deploys it. A
change that touches only the site therefore waits for the next release. The one exception: a site fix that cannot
wait may go out as a patch release of its own.

## Language

Write everything that goes into source files in English: identifiers, file and
directory names, comments, test names, log and error messages, and distributed
guidance (rules, skills, references, and contracts).

## Repository guidance

The root [AGENTS.md](AGENTS.md) is maintainer guidance for this repository. It is
not a template to copy into another project, and contributors do not need to
follow its maintainer workflows (such as local installation refreshes).

## Model behavior pilots

Pilots that measure how a model follows the guidance are paused. Do not add or
run them in a contribution. Ordinary non-model tests, as listed above, are
unaffected.

## Reporting security issues

Do not open a public issue for a vulnerability. See [SECURITY.md](SECURITY.md).
