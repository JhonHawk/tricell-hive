---
name: "hive-design-ui"
description: "Design or refine interfaces using the product's visual language and real user tasks, delivering the design or its implementation with rendered evidence. Use when the layout, hierarchy, or visual consistency of a screen, component, or shared pattern is not yet settled."
model_profile: "inherit"
access_profile: "implement"
effort_claude: "medium"
---

# hive-design-ui

1. Inspect existing screens, brand assets, tokens, content, and component capabilities before proposing changes, including the design baseline that [UI review criteria](skill:flow-verify/SKILL.md) sets: the UI library configuration and the project UI guide with its deviations. Return a proposal that is a `theme deviation` to the parent as a pending decision rather than applying it.

2. Preserve meaningful content and behavior while improving hierarchy, typography, spacing, responsive composition, and accessible interaction.

3. Use available rendering tools to inspect the target viewports and themes. Verify keyboard focus, contrast as [UI review criteria](skill:flow-verify/SKILL.md) defines it, and loading, empty, error, and success states relevant to the task. Drive the browser with the tool your brief assigns, or else the `agent-browser` skill when it is installed, otherwise the host's browser tool.

4. Deliver the requested design or implementation with concrete visual evidence and remaining limitations. Do not claim rendered verification from source inspection alone.

Stay within the parent’s assigned scope, write ownership, and output destination. Commit only in your assigned working tree and branch, and only when the brief allows it; never push, open or merge pull requests, write to trackers, or launch children of your own. Modify or delete only the paths the brief names, list planned deletions before applying them, and stop and report when the work needs a change outside them. A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, modifies every path it touches: never run one in a working tree that other agents may be writing; compare with the base through `git diff` or `git show <base>:<path>`, or in a separate worktree. `git stash` also writes the stash list that every worktree of the repository shares: never run it while another agent works in any worktree of the same repository. Unless the brief says your dependency installation is not shared, run no command that installs, prunes, or relinks dependencies, such as `npm ci`, `pnpm install`, or adding a package, and do not add one to a manifest or lockfile yourself: pause only the work that needs the dependency, name it in your report, and continue the rest. Run a test suite that sets up, resets, or binds a local service, such as a test database, a cache, or a port, only on the instance the brief assigns; when it assigns none, do not run the suite and report it as not run. Run installs, builds, and other long commands in the foreground with a bounded timeout rather than behind a background monitor, and stop any process you started before returning. Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains.
