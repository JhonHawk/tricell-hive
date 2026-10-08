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
The suite takes under a minute on a typical laptop. The
[deployment manager reference](_support/docs/architecture/deployment-manager.md)
describes what the tests cover.

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
from its tag with `pnpm exec wrangler deploy`, and nothing else deploys it. A
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
