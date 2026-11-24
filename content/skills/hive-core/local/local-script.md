# Local scripts

Read before naming, writing or running a local script. It covers changing one too, and the script is one that a project's `HIVE.local.md` names.

- A project's `HIVE.local.md` may name a local script, a helper for this machine that the agent wrote, such as one that creates a database per worktree; the machine file names none. The script lives at `_support/scripts/<name>` under the root of the file that names it, kept out of Git as [local write](local-write.md) sets, and the entry names it by that path; make a new script executable. From a worktree, run a repository's script by the same path under the main checkout, with the worktree as working directory; when that path is the worktree itself, a standalone copy or a bare repository, ask the user for the main checkout.
- The main thread writes or changes a script without asking only in `flow-build` (including the fixes `flow-verify` makes) or in an operational task. Research (its status report included), plan, a review and a child never do: they report the script they need instead. When the host refuses the edit from a worktree, follow [worktree](local-write/worktree.md).
- A script holds no secret value: it reads the secret location's files. The final report names each script the work created or changed. A script every clone needs is not local: name it as a candidate for the repository's tracked code.
