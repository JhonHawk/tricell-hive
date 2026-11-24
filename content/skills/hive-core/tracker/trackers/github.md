# GitHub Issues

Read from [trackers](../trackers.md) for a ticket whose key the `## Hive` assigns to GitHub Issues. Use the `gh` CLI; read `gh <command> --help` before the first use in a session.

- An issue is open or closed; a project board adds a `Status` single-select field whose options are the states (`gh project field-list <project> --owner <owner> --format json`). Move on a board with `gh project item-edit`; otherwise close or reopen the issue.
- Attach a capture with `gh issue artifact create <number> -R <owner>/<repo> --attach <file>#<alt text>`, or text such as test output with `--body-file <file>`; where the command or issue artifacts are not available, fall back as [trackers](../trackers.md) «Comment and attach» sets.
- Categories: a closed issue is done and an open one to do, unless the board's status option says in progress; record the board's option names when the project uses one.

Time: [time](time.md) sets it.
