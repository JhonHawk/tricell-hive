---
name: hive-rule-project-verification-format
description: "Hive rule leaf: read when writing or checking the JSON file of the project verification contract: targets, the five capabilities, features, and what makes a capability verified."
disable-model-invocation: true
---

# Project verification format

Read when you write or check the JSON of the project verification contract, after [project verification](../hive-rule-project-verification/SKILL.md).

The JSON object has exactly `schema_version: 1` and `targets`, a nonempty array. Each target has exactly:

- `id`: unique target ID; IDs use letters, digits, `_`, `-` and `.`, starting with a letter or digit.
- `cwd`: existing repo-relative working directory, with `.` for the root.
- `capabilities`: exactly `launch`, `doctor`, `drive`, `evidence`, `cleanup`.
- `features`: an array, initially empty, of the feature records below. IDs are unique within a target.

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

Reference IDs are unique across all five capabilities within their target. Reference paths name existing files. Commands are opaque data, interpreted by the flow only after its normal authorization checks. Every content path is repo-relative, resolved from `--repo` regardless of target `cwd`; absolute paths, `..` segments and symlinks escaping the root are rejected. Evidence paths also name existing files; keep raw runs in ignored support workspace and never embed credentials. Evidence result is one of `pass`, `failed`, `blocked`, `inconclusive`. A `verified` capability requires references, a success condition and passing evidence with a nonempty revision and existing file. A revision is declared identity, not a Git lookup or freshness guarantee.

Each feature has exactly `id`, `sources`, `drive`, `evidence`. `sources` is a nonempty array of existing repo-relative files. `drive` and `evidence` are arrays of reference IDs belonging to the matching capability in that same target. They may be empty while that capability is unverified, unsupported or not applicable; if verified, each feature needs at least one matching reference. Empty coverage never proves that feature works. Duplicate keys, IDs, array entries, unknown fields and unknown versions are errors.
