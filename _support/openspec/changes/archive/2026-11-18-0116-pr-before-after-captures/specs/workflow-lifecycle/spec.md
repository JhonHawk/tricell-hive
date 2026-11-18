## ADDED Requirements

### Requirement: Pull requests carry before and after captures of changed screens

Before its first edit of anything an existing screen renders, a build SHALL have its main thread serve the unchanged checkout and capture each region its plan, or the request when it has no plan, changes, with the widths, themes and states the verification will check, before any child task edits that surface; those captures SHALL NOT be published to the viewer, and the checkpoint SHALL record their paths or their absence. When the verification produced captures, the remote is github.com and the installed `gh` offers `--attach`, the pull request body SHALL hold a before and after table per region, width, theme and state, with the images uploaded through `--attach`, none added to a commit, and a cell that names a new screen or a missing before capture. In a public repository or on GitHub Enterprise Server, it SHALL attach only captures that show seeded data and test accounts, without internal URLs or real email addresses, and only with the user's one-time authorization for that work. Otherwise the body SHALL keep its text-only form and name the captures as kept locally. At the close, a capture SHALL be treated as replaced by the pull request only when its uploaded URL was read back in the merged body; other captures SHALL be kept.

#### Scenario: Existing screen changed in a private repository
- **WHEN** a build changes an existing screen in a private github.com repository and `--attach` is available
- **THEN** the pull request body shows each changed region's before and after images, uploaded and not committed

#### Scenario: Child edits the screen
- **WHEN** the plan delegates the screen's task to a child
- **THEN** the main thread captures the before state before launching it, and the child's brief asks for no capture

#### Scenario: New screen
- **WHEN** the change adds a screen that did not exist
- **THEN** its row has no before image and says the screen is new

#### Scenario: No attach option
- **WHEN** the installed `gh` lacks `--attach` or the remote is neither github.com nor GitHub Enterprise Server
- **THEN** the body keeps its text-only form, names the captures as kept locally, and the close keeps them

#### Scenario: Public repository without authorization
- **WHEN** the repository is public and the user has not authorized the public upload for this work
- **THEN** the captures stay local, the body keeps its text-only form, and the close keeps them

#### Scenario: Partial upload
- **WHEN** some attachments fail while the pull request is created
- **THEN** the agent attaches the missing ones with `gh pr edit --attach` and edits the body only from the published body

#### Scenario: Close after the merge
- **WHEN** the work is integrated and the agent read back each uploaded image URL in the merged body
- **THEN** the close proposes deleting those local captures and keeps any capture without a read-back URL
