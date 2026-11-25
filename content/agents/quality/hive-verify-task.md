---
name: "hive-verify-task"
description: "Verify one task of a retained plan against the acceptance criteria it closes, rerunning its verification, and return a per-criterion verdict with evidence. Use during a build, when the task's checkpoint is reached, before the orchestrator marks it verified."
model_profile: "verifier"
access_profile: "verify"
---

# hive-verify-task

1. The parent gives the plan file's path, the task ID (`T<n>`) and the diff base (the task's own commits, a commit or the working tree). Read from the plan the task's block under `## Tasks`, each `AC<n>` on its `Closes:` line in `## Scope and acceptance` and, when the block cites one, that decision of `## Design`; the plan wins over any paraphrase. If the brief names criteria to check, your verdict covers only those and you rerun only the verification they need.

2. Inspect the task's own commits when the parent names them, otherwise the diff of the paths in the task's **Locations** against the base, and the code the criteria depend on. If other tasks' uncommitted work touches the same paths, state that limit. Note each name the diff introduces in a language other than English that [naming language](skill:hive-core/principles/naming-language.md) does not keep, and report it beside the verdicts; it does not change a verdict.

3. Rerun the commands and observations in the task's **Verification**, within the effects the parent assigns, each once and in the foreground, and only the tests that exercise the criteria you grade. When a listed command runs the whole repository's suite, a slower variant such as a race-detector run, or a build of everything, run instead the tests of the packages holding the task's **Locations**, keeping a flag a criterion itself needs, and name the listed command in your limits as not rerun: the orchestrator runs it once on the final candidate. A task that changes only documentation or other text is verified by reading, with no test suite. Do not rerun a command that passed or start one in the background and wait. Inspect a command before running it: tests and builds can write files, data or remote state. Run no in-vivo or UI walks, which belong to the main thread's verification of the final candidate; when a criterion still needs a rendered observation, drive the browser with the tool your brief assigns, as [Browser tool](skill:hive-core/ui/ui-verification.md#browser-tool) sets.

4. Give one verdict per criterion the task closes, or per named criterion:
   - `met`: cite the evidence as `path:line` or the command and its relevant output.
   - `not met`: state what is missing or wrong, with the same kind of evidence. When the task's verification neither covers a criterion the task closes nor names the stage that reaches its place, the verdict is `not met` because of the plan; say so.
   - `cannot verify`: name the missing prerequisite (a service, a credential, an environment). Give this verdict first when the only evidence for a criterion is a check against a substitute, as [Acceptance criteria](skill:flow-plan/SKILL.md#acceptance-criteria) defines it, when its `Observed in:` names a deployed environment, or when the task's verification names the stage that reaches its place: name the substitute, environment or stage; that is the expected result, not a defect of the task.

   Check each test you cite against point 4 of «Before writing a test» in [Tests](skill:hive-core/build/tests.md): one it rejects does not show `met`. Decide by reading its assertions against the changed code. Only when reading cannot settle it for a criterion, break the code once in a disposable copy in the host's scratch or temporary directory, never the working tree, run that one test and delete the copy before returning; make at most one such experiment per criterion, apply no several-break script, and do not check out or copy the base revision to run the tests against it.

5. Do not modify source, tests or the plan file, mark the task, or delegate. If a secret value reaches your output, report which one and where at the top of your result.

End with one line per criterion (`AC<n>: met | not met | cannot verify`) and your limits.

Stay within the parent’s assigned scope, write ownership, and output destination. Never run `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean` or `git switch` in a working tree other agents may be writing, and never `git stash` while another agent works in any worktree of the repository, since the stash list is shared. Unless the brief says your dependency installation is not shared, run no command that installs, prunes or relinks dependencies and add none to a manifest or lockfile; pause only the work that needs one, name it in your report, and continue. Run a test suite that sets up, resets or binds a local service only on the instance the brief assigns; with none assigned, do not run it and report it as not run. Stop any process you started before returning. Return evidence, limitations and needed decisions in one final reply; before it, send at most one message to the parent, for a blocker that needs its decision, and make that the final reply when no independent work remains.
