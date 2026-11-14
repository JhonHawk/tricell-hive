# Initialize an explicit Engram workspace identity

Read from [harness-audit](../SKILL.md) only when the user asks to initialize one explicit Engram identity for a declared workspace or selected repositories, or to diagnose why it is blocked; never for ordinary memory access. This mode replaces the audit procedure.

Engram identity is local state governed by the global guidance's «Preservation» section, which allows creating or changing it only through this mode; an ambiguous detection result does not declare a root. This mode does not rewrite an existing identity: report a blocked configuration for the user to resolve. An identity is created only for a declared root, an explicit identity, and explicit repository targets; do not infer sibling scope, modify global Git ignores, or migrate memory.

Use [the helper](../scripts/init_workspace.py), run as `python3 <harness-audit-dir>/scripts/init_workspace.py --root <root> --name <project_name> [--target <path>]... [--apply]`, only after the user has declared the workspace root, identity, and target repositories. It is dry-run by default. Run `--apply` only for targets already covered by that authorization.

The helper resolves every target before writing. Existing matching `.engram/config.json` files are preserved; a differing or malformed identity blocks the whole apply before any target changes. A target must be the declared root or be contained by it after symlink resolution. Standalone repositories and worktrees are valid explicit targets, but do not make their parent or sibling repositories part of the request.

Report the declared root, selected targets, requested identity, dry-run or apply result, and any blocked existing configuration. Do not create a config when identity or targets remain ambiguous.
