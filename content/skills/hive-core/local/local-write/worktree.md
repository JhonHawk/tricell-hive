# `HIVE.local.md` in a worktree

Read when the session runs in a worktree, at session start for the files the global guidance's «Local instructions» reads, and from [local write](../local-write.md) before writing one of them.

- In a worktree, use the main checkout's file (the first `worktree` line of `git worktree list --porcelain`).
- If the host refuses editing it, only append a new entry through its permitted shell: never update or prune existing entries from that worktree. Report those changes for a session in the main checkout; if appending is also refused, report the unwritten entry.
- Never write a copy of the file inside the worktree.
- Check exclusions from the file's own repository, or the same relative path in the worktree when the host refuses the main checkout.
