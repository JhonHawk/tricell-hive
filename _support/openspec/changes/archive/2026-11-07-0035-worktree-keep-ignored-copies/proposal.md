# worktree-keep-ignored-copies — Remove worktrees whose ignored files lose nothing

| Field | Value |
| --- | --- |
| Status | Completed |
| Tracker | None |

## Request

The user approved correcting the close's worktree keep condition (H7 A, 2026-11-06). Grok Build session 01a1b012 (2026-11-06) kept a merged worktree and both copies of its branch because it held an ignored `.env` identical to the main checkout's copy, two identical `next-env.d.ts` and three `*.tsbuildinfo` caches.

## Change

Ignored entries identical to the same path in the main checkout of the worktree's own repository, and TypeScript or Next.js generated files, no longer keep a worktree; the work's own processes inside it stop before it is removed.
