---
name: hive-rule-git-commit
description: "Hive rule leaf: read before any git add or commit. Stage by name, commit through the state helper, let hooks run, keep history intact."
disable-model-invocation: true
---

# Stage and commit

Read before the first `git add` or commit of a task, in any flow. The authorization to commit comes from the delivery answer in force or a rule of the global guidance; this leaf sets only how.

- Stage only the paths and portions that belong to the result, naming each file. Never run `git add -A`, `git add .`, or `git add` on a directory, even `openspec/changes/`, since it takes other work in it. A `git mv` stages its source and destination; name both.
- Commit whole authorized files with `python3 -I <git-workflow-dir>/scripts/git_state.py commit --repo <dir> -m <msg> [--unreviewed-fix <reason>] -- <paths>` ([state](../git-workflow/scripts/git_state.py)), which preserves unrelated staged paths, scans `<paths>` for secrets, and runs `git commit --only -- <paths>`; for partial-file staging or without `python3`, record the pre-existing index, commit only the authorized portions, and verify the remaining index and diff afterward.
- Write commit messages and the pull request body as [prose](../hive-rule-prose/SKILL.md) sets.
- Make each commit one reviewable, revertable concern. Never rewrite, amend, force-push, clean or alter existing history or unrelated work without explicit authorization, except the `--force-with-lease` on the exact remote commit that [unattended delegation](../unattended-delegation/SKILL.md) allows (recording the previous commit to revert; a failed lease stops the push).
- **Let hooks run.** Never skip a Git hook, such as with `--no-verify`, `git commit -n`, `HUSKY=0`, `LEFTHOOK=0` or an overridden `core.hooksPath`, without the user's approval for that repository; the approval holds for the rest of the session unless withdrawn, and each commit it covers names it in the report. When a hook fails or skips checks because a tool is missing here, read the hook and run its checks another way with tools already available; installing a tool needs its own authorization. Report a check that still cannot run as not run, name the missing tool, propose installing it, and ask before pushing. When no one can answer, push only when the workflow files show a job triggered by that push or pull request whose steps run the same commands, since a job name alone does not prove it; otherwise keep the commit local and report the push as pending.
