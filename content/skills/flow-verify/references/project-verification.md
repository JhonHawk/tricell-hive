# Project verification contract

Read when E2E needs a project's verification capabilities, or when a change adds or changes a command, prerequisite, observable success condition or feature journey used to verify that project. This index describes existing application checks; installing Hive does not implement those checks.

## Reuse and maintain

Inspect the project's maintained verification documentation and scripts first. Reuse an equivalent contract and record its location in the checkpoint. Do not create a second active source. Convert an existing source only explicitly, preserving its intent and callers. Keep application scripts in their code locations; do not install a test framework or move executable code into the support folder to fill this index.

When no equivalent source exists, initialize the [template](../assets/project-verification.json) at `_support/docs/verification/project.json` under the project's existing support home, using the [contract helper](../scripts/check_contract.py):

```sh
python3 -I <flow-verify-dir>/scripts/check_contract.py init --repo /absolute/project --path _support/docs/verification/project.json
python3 -I <flow-verify-dir>/scripts/check_contract.py check --repo /absolute/project --path _support/docs/verification/project.json
```

Both arguments are required. `--repo` is an absolute root; `--path` is relative to that root or absolute within it. `init` creates parent folders and refuses an existing file or symlink, including a dangling symlink. It generates one `project` target, five `unverified` capabilities and no invented commands or features. `check` reads only. Exit 0 means initialized or structurally valid; exit 1 means rejected; argument errors exit 2. Neither action executes commands, discovers executables or proves behavior.

Populate entries from actual project sources and commands. Select target boundaries from the project's services, not an invented shared layout. Cover the journeys affected by the objective, with no minimum feature count. On later changes, update only affected targets, capabilities and features; preserve unrelated entries. Invalidate affected `verified` statuses to `unverified` when code, commands, requirements or success conditions change; obtain new evidence before restoring them.

## Format v1

The JSON object has exactly `schema_version: 1` and `targets`, a nonempty array. Each target has exactly:

- `id`: unique target ID; IDs use letters, digits, `_`, `-` and `.`, starting with a letter or digit.
- `cwd`: existing repo-relative working directory, with `.` for the root.
- `capabilities`: exactly `launch`, `doctor`, `drive`, `evidence`, `cleanup`.
- `features`: an array, initially empty, of the feature records below. IDs are unique within a target.

Each capability has exactly these fields:

| Field | Contract |
| --- | --- |
| `status` | `unverified`, `verified`, `unsupported`, `not_applicable` |
| `reason` | Nonempty text for unsupported or not applicable; otherwise text or null |
| `references` | Array of `{ "id": "check-id", "path": "scripts/check.py", "command": "python3 scripts/check.py" }`; command is text or null for a document |
| `requirements` | Array of nonempty descriptions of tools, environment, accounts and resources required; no secrets |
| `effects` | Array of nonempty descriptions of writes, cost or other effects; empty means none declared, not authorization |
| `success` | Observable success condition, or null while unverified |
| `evidence` | Null, or `{ "revision": "checked revision", "result": "pass", "path": "_support/workspace/run/output.txt" }` |

Reference IDs are unique across all five capabilities within their target. Reference paths name existing files. Commands are opaque data, interpreted by the flow only after its normal authorization checks. Every content path is repo-relative, resolved from `--repo` regardless of target `cwd`; absolute paths, `..` segments and symlinks escaping the root are rejected. Evidence paths also name existing files; keep raw runs in ignored support workspace and never embed credentials. Evidence result is one of `pass`, `failed`, `blocked`, `inconclusive`. A `verified` capability requires references, a success condition and passing evidence with a nonempty revision and existing file. A revision is declared identity, not a Git lookup or freshness guarantee.

Each feature has exactly `id`, `sources`, `drive`, `evidence`. `sources` is a nonempty array of existing repo-relative files. `drive` and `evidence` are arrays of reference IDs belonging to the matching capability in that same target. They may be empty while that capability is unverified, unsupported or not applicable; if verified, each feature needs at least one matching reference. Empty coverage never proves that feature works. Duplicate keys, IDs, array entries, unknown fields and unknown versions are errors.

## Execute through the real path

Launch starts the required application resources. Doctor checks prerequisites and readiness. Drive exercises the feature journeys. Evidence records reproducible outputs and their revision. Cleanup stops or removes only resources this check owns. Mark an unsupported capability or a capability that does not apply with its reason; do not invent a successful command to make the shape pass.

The flow inspects requirements, effects and authorization before running the project's normal commands from the target's declared directory. Use the project's account index before checks that sign in or use existing records, and isolate resources assigned to concurrent actors. Inspect exit results, outputs, checked revision and cleanup. Structural validity and a declared `pass` are not execution evidence. Before reusing a verified entry, contrast its evidence with the current candidate and affected feature journeys; stale, missing or uninspected evidence remains unverified. Report capability and feature gaps explicitly rather than inferring overall readiness from `check` exit 0.
