# Before captures

Read before the first edit of an existing screen. It applies in a build that changes an existing screen, to the first edit of anything that screen renders: its components, styles, layout, the data contract it consumes or its seed data. Those captures feed the before and after table of the pull request, as [pull request captures](../../git-workflow/references/pr-captures.md) sets.

- The main thread serves the unchanged checkout with the data it already has and captures each region the plan changes, or the request changes when the work has no plan, at the widths, themes and states the verification will check, before it edits that surface or launches a child task that edits it. A child's brief asks for no capture.
- For a native app, the main thread installs the unchanged build on the device or emulator the verification will use and captures its screen as [Device tool](ui-verification/device-tool.md) sets.
- Save them in the work's `workspace/YYYY-MM-DD-<work>/images/before/`, each named so it pairs with the candidate capture of the same region, width, theme and state, and record their paths, or that there are none, as one `--check` line of the build's checkpoint, which later checkpoint writes keep. Do not publish them to the viewer: they would number as the screen's first iteration.
- When that data differs from the data the verification will seed, note the difference for the table. A new screen has no before capture, and a before state an earlier edit already changed is not rebuilt: the table says so.
