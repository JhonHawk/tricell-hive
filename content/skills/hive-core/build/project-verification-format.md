# Project verification format

Read when writing or checking the JSON of the project verification contract. Read it after [project verification](project-verification.md). `check_contract.py check` enforces the structure and prints each violation; this leaf holds the meaning of the fields.

The object has `schema_version: 1` and a nonempty `targets` array. Each target has a unique `id`, a repo-relative `cwd` (`.` for the root), `capabilities` (exactly `launch`, `doctor`, `drive`, `evidence`, `cleanup`) and `features`, initially empty.

Each capability has exactly these fields:

| Field | Contract |
| --- | --- |
| `status` | `unverified`, `verified`, `unsupported`, `not_applicable` |
| `reason` | Nonempty text for unsupported or not applicable; otherwise text or null |
| `references` | Array of `{ "id": "check-id", "path": "scripts/check.py", "command": "python3 scripts/check.py" }`; command is text or null for a document |
| `requirements` | Array of nonempty descriptions of tools, environment, accounts and resources required; no secrets |
| `effects` | Array of nonempty descriptions of writes, cost or other effects; empty means none declared, not authorization |
| `success` | Observable success condition, or null while unverified |
| `evidence` | Null, or `{ "revision": "checked revision", "result": "pass", "path": "_support/workspace/run/output.txt" }` |

Commands are opaque data, interpreted by the flow only after its normal authorization checks. Keep raw runs in the ignored support workspace and never embed credentials. Evidence `result` is `pass`, `failed`, `blocked` or `inconclusive`. A `verified` capability requires references, a success condition and passing evidence with a nonempty revision and an existing file. A revision is declared identity, not a Git lookup or freshness guarantee.

Each feature has `id`, `sources` (existing repo-relative files), `drive` and `evidence` (reference IDs of the matching capability in the same target). They may be empty while that capability is unverified, unsupported or not applicable; if it is verified, each feature needs at least one matching reference. Empty coverage never proves that feature works. Duplicate keys, IDs, array entries, unknown fields and unknown versions are errors.
