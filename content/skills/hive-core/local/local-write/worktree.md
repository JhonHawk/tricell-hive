# `HIVE.local.md` in a worktree

Read when the session runs in a worktree, at session start for the files the global guidance's «Local instructions» reads, and from [local write](../local-write.md) before writing one of them.

- In a worktree, use the main checkout's file (the first `worktree` line of `git worktree list --porcelain`); never write a copy of the file inside the worktree.
- If the host refuses editing it, only append a new entry through its permitted shell: never update or prune existing entries from that worktree. Report those changes for a session in the main checkout, or the unwritten entry when appending is also refused. The same rule covers a row of the account index of [test accounts](../test-accounts.md) and a local script of [local scripts](../local-script.md): append a new row, or write a script or a second factor's key file only where no file exists at its path, and report a change to an existing row, script or key file. A saved browser state stays where [Browser tool](../../ui/ui-verification.md#browser-tool) saves it.
- Check exclusions from the file's own repository, or the same relative path in the worktree when the host refuses the main checkout.
