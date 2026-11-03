# branch-promotion

## ADDED Requirements

### Requirement: master accepts changes only by promotion from development
The GitHub repository SHALL protect `master` with an active ruleset that requires a pull request merged as a merge commit, requires the `source-branch` check emitted by GitHub Actions, and blocks direct pushes, force pushes and deletion. The `source-branch` check SHALL run from the default branch on every pull request targeting `master`, and SHALL pass only when the pull request's head is the `development` branch of this repository. Only the repository admin role MAY bypass the ruleset, and only by merging a pull request.

#### Scenario: Promotion from development
- **WHEN** a pull request from `development` targets `master`
- **THEN** the `source-branch` check succeeds and the pull request can be merged

#### Scenario: Pull request from another branch
- **WHEN** a pull request from any branch other than `development` of this repository, including a fork's `development`, targets `master`
- **THEN** the `source-branch` check fails and the pull request cannot be merged without an admin bypass

#### Scenario: Direct push
- **WHEN** anyone, an admin included, pushes, force-pushes or deletes `master` directly
- **THEN** GitHub rejects the update
