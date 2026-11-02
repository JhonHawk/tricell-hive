## MODIFIED Requirements
### Requirement: Published documentation site
The repository SHALL contain a static documentation site in `site/`, built with the pinned toolchain in `site/package.json`, and it SHALL be published from `master` at `https://hive.tricell.tech/`, as a Cloudflare Worker reached through the route `hive.tricell.tech/*`. The site SHALL have a landing page, a flows overview page at `flows/overview/`, and one page for each of these flow skills: `flows/research/`, `flows/plan/`, `flows/build/`, `flows/verify/`, `flows/deploy/`, `flows/close/`, and `flows/security/`. The `flow-report` and `flow-walkthrough` skills have no flow page. Its text SHALL be in English, the product language.

#### Scenario: Site served after a release
- **WHEN** a release promotes a commit containing `site/` to `master`, builds `site/`, and deploys it from its tag
- **THEN** `https://hive.tricell.tech/` and each flow page respond `200` with that commit's site

#### Scenario: Other branches do not publish
- **WHEN** a commit lands on `development` or a work branch
- **THEN** the published site does not change until a release moves `master`

## ADDED Requirements
### Requirement: Flows overview page
The flows overview page SHALL describe for people, per flow, where work iterates, where the user can request a change, and where the work returns after it, and SHALL link to the skill or reference that holds each rule on `master`. It SHALL NOT reproduce rule text from the flow skills or `content/guidance/global.md`, and it is not a flow page, so the sections that flow pages require do not apply to it.

#### Scenario: A reader wants to change scope mid-flow
- **WHEN** a reader opens `flows/overview/` to learn how to ask for a change during verify
- **THEN** the page says where that request returns the work and links to the rule that classifies it
