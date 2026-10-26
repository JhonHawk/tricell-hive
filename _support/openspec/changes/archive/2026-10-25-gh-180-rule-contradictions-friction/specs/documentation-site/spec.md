## MODIFIED Requirements
### Requirement: Published documentation site
The repository SHALL contain a static documentation site in `site/`, built with the pinned toolchain in `site/package.json`, and it SHALL be published from `master` at `https://hive.tricell.tech/`, as a Cloudflare Worker reached through the route `hive.tricell.tech/*`. The site SHALL have a landing page and one page per flow skill: `flows/research/`, `flows/plan/`, `flows/build/`, `flows/verify/`, `flows/deploy/`, `flows/close/`, and `flows/security/`. Its text SHALL be in English, the product language.

#### Scenario: Site served after a release
- **WHEN** a release promotes a commit containing `site/` to `master`, builds `site/`, and deploys it from its tag
- **THEN** `https://hive.tricell.tech/` and each flow page respond `200` with that commit's site

#### Scenario: Other branches do not publish
- **WHEN** a commit lands on `development` or a work branch
- **THEN** the published site does not change until a release moves `master`
