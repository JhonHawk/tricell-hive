# documentation-site

## Purpose

A documentation site for people, published at https://hive.tricell.tech/, that describes each flow and links to its skill instead of copying rules.

## Requirements

### Requirement: Published documentation site
The repository SHALL contain a static documentation site in `site/`, built with the pinned toolchain in `site/package.json`, and it SHALL be published from `master` at `https://hive.tricell.tech/`, as a Cloudflare Worker reached through the route `hive.tricell.tech/*`. The site SHALL have a landing page and one page per flow skill: `flows/research/`, `flows/plan/`, `flows/build/`, and `flows/close/`. Its text SHALL be in English, the product language.

#### Scenario: Site served after a release
- **WHEN** a release promotes a commit containing `site/` to `master` and deploys `site/` from its tag
- **THEN** `https://hive.tricell.tech/` and each flow page respond `200` with that commit's site

#### Scenario: Other branches do not publish
- **WHEN** a commit lands on `development` or a work branch
- **THEN** the published site does not change until a release moves `master`

### Requirement: Flow pages describe without copying rules
Each flow page SHALL contain the sections `When to use`, `What it produces`, `What it does not do`, `Related flows`, and `FAQ`, and SHALL link to its skill's `SKILL.md` on `master`. A page SHALL describe the flow's behavior for people and SHALL NOT reproduce rule text from the flow skills (`content/skills/flow-*/`) or the shared guidance (`content/guidance/global.md`), so each rule keeps one canonical home.

#### Scenario: A reader needs the exact rule
- **WHEN** a reader wants the rule behind a page's description
- **THEN** the page's link opens the skill's `SKILL.md` on `master`

### Requirement: Installation redirect unaffected by the site
Serving the site on `hive.tricell.tech` SHALL NOT change the `/install.sh` redirect that the `versioned-installation` capability defines.

#### Scenario: One-line installation after the site is live
- **WHEN** a client requests `https://hive.tricell.tech/install.sh` after the site is published
- **THEN** the response is still `302` to `get-hive.sh` on `master`

### Requirement: Site kept out of distribution and Go tooling
`site/` SHALL NOT be part of the release package or of the content that `hive update` deploys, and its installed dependencies SHALL NOT affect `go vet ./...` or `go test ./...` at the repository root.

#### Scenario: Dependencies installed with a flat layout
- **WHEN** `site/node_modules` contains a Go file with a broken import
- **THEN** `go vet ./...` and `go test ./...` at the repository root still succeed
