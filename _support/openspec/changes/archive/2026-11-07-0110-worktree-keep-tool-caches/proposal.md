# worktree-keep-tool-caches — Tool caches of other stacks no longer keep a worktree

| Field | Value |
| --- | --- |
| Status | Completed |
| Tracker | None |

## Request

Finding H9 of worktree-keep-ignored-copies, proposed when the user typed `/flow-build` after the close listed it (2026-11-06); no real session has shown these caches keeping a worktree, and the user adopted it at once (D3 A): regenerable caches of other stacks still kept a merged worktree.

## Change

The close's keep conditions exempt a closed list of tool caches and generated files: `*.tsbuildinfo`, `next-env.d.ts`, `*.pyc`, `.turbo/`, `__pycache__/`, `.pytest_cache/`, `.mypy_cache/`, `.ruff_cache/` and `.DS_Store`. Untracked caches still keep a worktree, since `git worktree remove` refuses untracked files without `--force`.
