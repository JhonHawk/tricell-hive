# Project verification contract

Read when adding or changing a project verification command. It applies when E2E needs a project's verification capabilities, and to a change to a command, prerequisite, observable success condition or feature journey used to verify that project. This index describes existing application checks; installing Hive does not implement those checks.

## Reuse and maintain

Inspect the project's maintained verification documentation and scripts first. Reuse an equivalent contract and record its location in the checkpoint. Do not create a second active source. Convert an existing source only explicitly, preserving its intent and callers. Keep application scripts in their code locations.

When no equivalent source exists, initialize the [template](../../flow-verify/assets/project-verification.json) at `_support/docs/verification/project.json` under the project's existing support home with the [contract helper](../../flow-verify/scripts/check_contract.py) (`init`, then `check`, each with `--repo` and `--path`; flags and exit codes in `--help`). `init` generates one `project` target with five `unverified` capabilities and no invented commands or features. Neither action executes commands or proves behavior. When writing or checking the contract's JSON, read [verification format](project-verification-format.md).

Populate entries from actual project sources and commands, with target boundaries taken from the project's services and the journeys the objective affects. On later changes, update only affected targets, capabilities and features, and set an affected `verified` status back to `unverified` when code, commands, requirements or success conditions change; obtain new evidence before restoring it.

## Execute through the real path

Mark an unsupported capability or one that does not apply with its reason; do not invent a successful command to make the shape pass. The flow inspects requirements, effects and authorization before running the project's normal commands from the target's declared directory, uses the project's account index before checks that sign in or use existing records, and isolates resources assigned to concurrent actors. Cleanup stops or removes only resources this check owns. Inspect exit results, outputs, checked revision and cleanup. Structural validity and a declared `pass` are not execution evidence: before reusing a verified entry, contrast its evidence with the current candidate and affected journeys; stale, missing or uninspected evidence remains unverified. Report capability and feature gaps explicitly rather than inferring readiness from `check` exit 0.
