---
name: engram-init-workspace
description: Establish one explicit Engram identity for a declared workspace or selected repositories. Use only when the user asks to initialize or repair that identity; do not use for ordinary memory access.
---

# Initialize an explicit Engram workspace identity

Engram identity is local state. Do not infer a workspace from a parent directory, a set of siblings, or an ambiguous detection result. This skill neither migrates existing memories nor changes global Git ignores.

Use [the helper](scripts/init_workspace.py) only after the user has declared the workspace root, identity, and target repositories. It is dry-run by default. Run `--apply` only for targets already covered by that authorization.

The helper resolves every target before writing. Existing matching `.engram/config.json` files are preserved; a differing or malformed identity blocks the whole apply before any target changes. A target must be the declared root or be contained by it after symlink resolution. Standalone repositories and worktrees are valid explicit targets, but do not make their parent or sibling repositories part of the request.

Report the declared root, selected targets, requested identity, dry-run or apply result, and any blocked existing configuration. Do not create a config when identity or targets remain ambiguous.
