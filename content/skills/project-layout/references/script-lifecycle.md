# Script lifecycle

Read when a work writes or runs a script, when it decides where a script goes, when a close turns a candidate into a mechanism, and at [project-layout](../SKILL.md) step 3 when the inventory holds scripts. It sets the level of a script and what moves it up; [layout](layout.md) sets where each level lives.

A script climbs one level at a time, and only on the evidence each level names. Its level is set by what it did and who runs it again, never by its age or folder; its file name decides only one case, a script nothing calls, between run once (level 2) and a question to the user.

1. **Disposable.** A probe, a one-off query, a generated payload, such as a design tool's batch, a log parser or a measurement harness whose result the report already states. It lives in the host's scratch location and is removed at the close; it never goes in `workspace/` to be kept "just in case".
2. **Run once against shared data.** A script that changed data the work does not own alone, such as a backfill, a repair, a data normalization or a cutover against a shared or production database. It is the record of what changed and how to undo it: keep it in the work's `sessions/YYYY-MM-DD-<work>/` folder with its output, its revert and a report that states what it changed, when and against which environment. It is not a tool, even when it could be run again; running it again is new work.
3. **Candidate.** A purpose that repeats, as flow-close's [Candidates](../../hive-rule-close-report/SKILL.md#candidates) sets: its record counts the sessions that saw it, and the close footer offers it once it counts two.
4. **Internal tool.** A candidate the user chose to turn into a mechanism, written into the code of the repository whose code or data it uses, as [layout](layout.md) sets, with a line in the project's index. Before it moves there, it meets all of these:
   - no ticket key, work name or absolute path in its code, file name or defaults;
   - its inputs as arguments or configuration, and its secrets read from their location, never written in it;
   - a dry run by default, with an explicit flag to apply, and safe to run twice;
   - a test in the repository's test runner, or, without one, the check the global guidance sets for a project without tests;
   - its usage on one line in the index: what it is for, the command and its test.
5. **Mechanism.** A tool that always runs the same way becomes part of a check, such as a CI job, a hook or a package script, chosen as [defect fix](../../hive-rule-defect-fix/SKILL.md) point 5 sets for a repeatable error.

When an inventory finds a script under `_support/`, its level and its place come from the [placement planner](../scripts/plan_layout.py): who calls it decides between an internal tool of a package, of CI, of a hook or of a runbook, and a script nothing calls goes to level 2 when its name marks a one-off run (a ticket key, or a word such as backfill, fix, cutover, migrate, seed or repair), otherwise to the user. A tracked script under `_support/` that CI or every clone runs is an internal tool in the wrong place; the planner's `apply` updates every file that calls it.
