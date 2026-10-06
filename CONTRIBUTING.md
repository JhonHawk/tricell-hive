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
- When a change alters something that `README.md`, `llms.txt`, or a page of the
  documentation site in `site/` states, update it in the same pull request. The
  [README and llms.txt](AGENTS.md#readme-and-llmstxt) section lists what each one
  covers.

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
