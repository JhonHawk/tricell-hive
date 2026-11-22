# Comment ratchet

Read at [comment audit](../SKILL.md) step 2, when choosing the inventory tool, and at step 8, when wiring the lint. It sets the template for JavaScript and TypeScript, the contract a port to another language keeps, and how the lint runs locally and in CI.

## The template

[check-comments.mjs](../assets/ratchet/check-comments.mjs) and its suite [check-comments.test.mjs](../assets/ratchet/check-comments.test.mjs) cover `.js`, `.jsx`, `.mjs`, `.cjs`, `.ts`, `.tsx`, `.mts` and `.cts`. Until step 8 they run from the temporary folder of the skill's step 2, which writes no repository file. At step 8, copy both into the project's CI scripts folder, such as `scripts/ci/`, and set the four constants at the top:

- `SOURCE_PATTERNS`: the Git pathspecs it checks.
- `EXCLUDED`: exact paths, or prefixes ending in `/`, that a tool generates.
- `ALLOWED`: the directive patterns that stay; add the project's own, found at step 3 of the audit. Anchor each pattern at the comment's start, so prose that only mentions a directive word is still a new comment: a license counts in its header forms: `@license` or `@preserve` at the start of a line of a `/**` or `/*!` block, or after `//`; a `/*!` banner with a year, `(c)`, `©`, `@license`, `licensed` or an SPDX identifier; and a comment that starts with `Copyright`, `(c)`, `©` or `SPDX-License-Identifier:`. The block-form workaround marker counts only inside a JSX expression. A lint suppression keeps its `--` description, which suppression review reads as the justification.
- `DECISIONS_PATH`: the decision record the error message names, which the [no comments](../../hive-core/review/no-comments.md) leaf reads to know where a Keep goes.

It parses with `@babel/parser`, choosing the plugins by extension and adding `decorators-legacy`. TypeScript 7 no longer exports the classic parser, so the template does not use `typescript`. Use `@babel/parser` 8 on Node `^22.18.0 || >=24.11.0`, and 7.29 on an older Node; both pass the suite. At step 8, add it as a devDependency after the package check [reuse before writing](../../hive-core/plan/reuse-before-writing.md) rung 5 sets.

Modes and exits:

- `node check-comments.mjs --base <rev>`: lint. Each changed source file's comments are compared with the same file at `<rev>`, a renamed file with its old path. A comment beyond those the base had, counted by its text with each line's indentation trimmed, is a finding: `file:line:column: no-new-comments`. Directives and the workaround marker do not count.
- `--inventory`: one JSON line per comment of every tracked and untracked source file: `file`, `line`, `column`, `start`, `end`, `fullLine`, `directive`, `text`.
- `--same-code --base <rev>`: confirms each changed file lost only comments, with the line breaks between its tokens unchanged, since a removed multi-line comment can change automatic semicolon insertion. A deleted, moved or renamed source file, one replaced by a symlink, or an added one with code, is a finding. A JSX `{}` that held only a comment and whitespace-only JSX text do not count as code.
- Exit 0 clean, 1 with findings, 2 on any error: an unknown base, a missing parser, a file that does not parse. Exit 2 means the check did not run, so fix the setup it names instead of removing comments.
- Symlinks and other non-regular files are skipped, so the script reads nothing outside the repository through them.

Run the suite once after copying, with the project's parser installed: `node --test <folder>/check-comments.test.mjs`. It needs Git, creates its repositories in the system's temporary folder and ignores the user's Git configuration.

The marker regex checks only the marker's form; the review checks that its ticket or record entry exists.

## Another language

Write the same contract with the language's own parser or tokenizer. Regular expressions over the source do not work for this, since they cannot tell a comment from a string. A port:

- lists each comment with its file, line, column, offsets, whether it fills its line, whether it is a directive, and its text;
- compares each changed file with its version at `--base`, following renames, by comment text with indentation trimmed;
- allows the language's directives and the workaround marker written with its line-comment token, such as `# workaround: <ticket>`;
- names `DECISIONS_PATH` in its error message;
- exits 0, 1 or 2 as above, and runs from any folder of the repository;
- has a suite with at least: no new comment, a new comment, a duplicate, a comment the base removed, a rename, comment-like text in a string, an excluded path, an allowed directive, an unknown base.

## Wiring

- Add a `lint:comments` command, or the project's equivalent, that runs the script, and call it from the project's main lint command.
- In CI, run it on pull requests with `--base` set to the pull request's base commit, and on pushes to the base branch with `--base` set to the commit before the push. Check out enough history to reach that commit, such as `fetch-depth: 0`. Without `--base`, the script uses the merge base with `origin/HEAD`, which a CI checkout often lacks.
- Run the suite in CI too, so a change to the script that breaks the ratchet fails.
- Run it on the `pull_request` event, not `pull_request_target`, which runs a fork's code with the base repository's secrets. Pass the base through an `env:` variable rather than an expression inside `run:`, so a ref name cannot become shell text.
- The pull request's author can edit the script, `EXCLUDED` and `ALLOWED`: protect the CI scripts folder with a required owner review where the host allows it.
