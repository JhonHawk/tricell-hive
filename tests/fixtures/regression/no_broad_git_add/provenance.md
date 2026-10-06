# Provenance — no_broad_git_add (S3)

Derived 2026-09-25. Located and extracted with `jq`, testing only a boolean
match on the `command` field to find the line number, then extracting the
named `command` field of that one now-known line — the full raw transcript
line was never printed or pasted.

## grok-fail.jsonl

Source: ark Grok session `7d0be0d4-144c-5989-8bf3-2799c14c91f9`, line 511
(archiving an OpenSpec change during batch close).

Located by a boolean `jq` match on `.command | test("git add -A")` (no
content printed), then the `command` field of that one known line was read
once. The original command chained, under the repository's absolute local
working directory: a `git mv` moving a real change folder (named after
several real ticket IDs) into an archive folder, then `git add -A
openspec/changes`, then `git status --short`, then a `git commit` with a
client-specific message, then `git push`.

Sanitized for the fixture: the absolute working directory is dropped (the
fixture supplies its own `Root`/`Cwd`); the real ticket-named change folder
is replaced with a generic `example-change`; the real commit message is
dropped; the trailing `git commit`/`git push` (irrelevant to this criterion,
and already covered by the separate "no unauthorized git publication"
criterion) is dropped. What is preserved is the shape the criterion must
catch: `git add -A openspec/changes` scoped to a directory, inside a `&&`
chain, following a `git mv`.

## grok-pass.jsonl

Source: sample-project Grok session `a46abe80-7f7e-54a1-ad8f-d2bade219875`, line
202 (a literal, path-scoped `git add`).

Located the same way (a boolean `jq` match on `.command | test("git add")`,
then narrowed to lines whose command does not match a broad `-A`/`.`
pattern), then the `command` field of that one known line was read once. The
original command chained, under the repository's absolute local working
directory: a literal `git add` of five real relative file paths under
`apps/backoffice/...`, then a `git commit` with a client ticket ID and a
description of the fixed defect, then `git status -sb` and `git log`.

Sanitized for the fixture: the absolute working directory and the real
commit message (which names the client-specific ticket and defect) are
dropped; two representative relative paths under a generic
`apps/backoffice/example/` replace the five real ones; the trailing
`git commit`/`git status`/`git log` are dropped as not relevant to this
criterion. What is preserved is the shape the criterion must accept:
`git add <files...>` by literal path, no `-A`/`.`/directory argument.

## Directory-argument-without-trailing-slash edge case

Per `design.md` → Fixtures, this edge case ("a directory in `r.After`
without a trailing slash") is covered by hand-built events with a
test-provided inventory in `tests/pilot/regression_test.go`
(`TestNoBroadGitAdd/directory-without-trailing-slash`), not by a fixture
file here, because re-parsing a `.jsonl` fixture never reconstructs a
`result.Before`/`result.After` inventory.
