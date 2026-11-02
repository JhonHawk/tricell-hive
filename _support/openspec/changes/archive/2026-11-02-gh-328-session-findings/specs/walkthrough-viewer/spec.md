## MODIFIED Requirements

### Requirement: Self-describing protocol and confirmed-input hygiene
Offline contract usage/schemas/limits SHALL describe protocol/storage v3 exclusively. New CLI and viewer SHALL require compatible v3 peers for publication. No v1/v2 reader, endpoint, migration or imported evidence reconstruction SHALL be part of the v3 runtime. v3 receipts, immutable blobs and atomic state transitions SHALL remain durable after restart or uncertain commit.

The offline contract SHALL be readable by part: the full contract; one operation in full; a brief per flow (research, plan, build, verify, delivery, close, collab, local) with each operation's usage, required and optional input fields, enum values and example, without output schemas; a list of every operation with its flows and usage; and the shared rules block. Every operation SHALL belong to at least one flow. Contract and command output SHALL be compact JSON with the same content as the indented form. A rejected input SHALL name its field and code.

One canonical publication procedure SHALL own collaboration/native identity, meaningful snapshots, states before user stops, explicit resumption and readable summaries. Consumers SHALL link to it. Children SHALL return evidence to the main producer. Viewer identity SHALL NOT change Engram runtime identity. Hygiene SHALL retain authoritative plans, current review inputs, uncertain uploads and unique evidence.

#### Scenario: v3 restart
- WHEN a v3 publication commits and the service restarts
- THEN IDs, hashes, actor audit and operation receipts remain resolvable without a legacy adapter.

#### Scenario: Old writer
- WHEN a v1 or provisional v2 publisher sends a request to a v1 endpoint of v3
- THEN the endpoint does not exist and nothing is mutated.

#### Scenario: Brief contract for one flow
- WHEN a producer runs `hive view contract --for build`
- THEN it receives only the build operations, each with usage, required and optional fields, enum values and an example, plus the common session, stage and ticket operations, in under 10,000 bytes of compact JSON.

#### Scenario: Operation without a flow
- WHEN a new operation is added without a flow
- THEN the contract test fails.

#### Scenario: Rejected input
- WHEN a stage start omits its reason
- THEN the CLI reports `invalid timeline input: invalid_text on reason` and nothing is sent.
