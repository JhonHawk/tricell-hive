# Tasks

<!-- Template from flow-plan. Write the text in the session language; keep headings, field names, code and identifiers as they are. Delete these comments. -->

<The base commit of each repository when the plan was written, and the branch. Which tasks run in the main thread or in children, and why.>

### T1 — <what it delivers>

- [ ] <the observable result of this task>

**Closes:** <AC<n>, …>

**Depends on:** <T<n>, or none>

**Locations:** <files and packages it changes>

**Execution:** <main thread, or the role that runs it>

**Test approach:** <tdd · characterization · check, as the Tests section of flow-build defines them; omit for a mechanical change>

**Changes:** <what changes and what deliberately does not>

**Verification:** <the command or observation, and its expected result>

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | <command> | Covered by the implementation |
| UI check | Final candidate, locally | <viewports, themes, reviewer> | Covered by the implementation |
| Code review | Final candidate, before the push or the validation stop | <mechanism> | The review answer |

<The human walk, when the delivery stops for validation: the stack to serve, the account, and the steps.>

## Checkpoint

<Written by `scripts/tasks.py checkpoint`: Updated, Stage, Last commit, Open findings, Next, then `### Evidence` with the checks recorded so far.>

## Review status and progress

<Date — one line per task done: its commit and anything left unverified.>
