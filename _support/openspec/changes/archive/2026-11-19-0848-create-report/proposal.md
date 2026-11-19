# create-report

## Request

Rename the `flow-report` skill to `create-report`, so every skill that makes a deliverable starts with `create-`, next to `create-issue` and `create-prd`.

## Decisions

- `/flow-report` is retired in the same release, with no alias.
- The Starlight docs-site mode stays inside `create-report`.
- The `flow-report-assets` meta tag keeps its name: reports already written with it keep passing the checker.

## Scope

A mechanical rename: the skill folder, every link and literal use in `content/`, `tests/`, `site/`, `llms.txt` and `_support/docs/`, the changelog, and the documentation-site requirement that names the skill. No behavior changes.

Playbook: refactor
