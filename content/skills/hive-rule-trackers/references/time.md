# Time worked on a ticket

Read from [trackers](../SKILL.md) when [ticket state](../../hive-rule-ticket-state/SKILL.md) says to write or update a ticket's time. Only a tracker with native time tracking gets it, such as Jira; Linear and GitHub Issues get nothing about time.

## Measure

1. **Window.** Each ticket's time runs from the first message of the session that works it, research and plan included, to the delivery point or the close. Tickets worked together each get the whole window; a ticket started after another in the same session starts at its own first message.
2. **This session.** Run `hive time --since <that message's time>`, which prints `active_seconds`, `user_wait_seconds`, `ci_wait_seconds` and `converted_seconds`. With `"result":"partial"` at the delivery point and no marks, write nothing and report it.
3. **No transcript.** Run `hive time` once when work on a ticket starts; when it prints anything but `"result":"ok"`, keep marks for this session instead with [work_time.py](../scripts/work_time.py) `mark <file> <event>` (`python3 -I`), in a file named after the work in the host's temporary location: `start` at the first message on the ticket, `wait` before each message that waits for the user, `resume` at the user's reply, `ci-start` and `ci-end` around a remote CI or deployment wait. `sum <file>` prints the same fields.
4. **Earlier sessions.** The work state in persistent memory lists, for each session that worked the ticket, its host, its session ID and the time of its first message on the ticket, or the result of `sum` for a session that kept marks: the plan-ready write adds the planning session, and a pause or a close adds its own. Measure each listed ID again with `hive time --host <host> --session <id> --since <time>`, and add every `converted_seconds`.

User waits are the only time subtracted; builds, tests, remote CI and children count. The factor is 2.0 for the agent's work and 1 for remote CI waits: `hive time` recognizes `gh pr checks --watch` and `gh run watch`, and marks recognize each `ci-start` and `ci-end` pair, which may also enclose a remote deployment wait; `--factor` changes it only after a recalibration the user approved, never above 3.

## Write

1. Read the worklogs your account already has on the ticket, as its tracker's reference sets.
2. At the delivery point, add one worklog with the total `converted_seconds` and nothing else: no comment, no description. At the close, update that same worklog to the new total; when a worklog this work wrote exists, update it instead of adding one.
3. Never mention the time, how it was measured or the factor anywhere a ticket's readers see: its comments, its evidence or the pull request.
4. Record in the work's close record in persistent memory, per ticket: the worklog ID, `active_seconds`, `ci_wait_seconds`, the factor, the ticket type and the sessions measured.
