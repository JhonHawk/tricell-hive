---
paths:
  - "**/*.sh"
  - "**/*.bash"
---

## Shell Scripts

> **Apply proportionally.** Scripts >20 lines OR persistent (committed beyond the current session, deployed to CI/production, or shared with others) follow all rules. Ephemeral session scripts under `_support/scripts/` need only `set -euo pipefail` plus quoted variables — the rest is overkill for code that gets deleted today.

- **Shebang:** `#!/usr/bin/env bash`.
- **`set -euo pipefail`** at the top of every non-trivial script. Fail on errors (`-e`), undefined variables (`-u`), and pipe failures (`-o pipefail`).
- **Quote all variables:** `"$var"` not `$var`. Unquoted variables cause word splitting and glob expansion bugs.
- **`[[ ]]` over `[ ]`** for conditionals — supports regex, no word splitting, and safe with empty strings.
- **`readonly`** for constants: `readonly DB_NAME="mydb"`.
- **`local`** for function variables: prevents leaking into global scope.
- **Logging to stderr, data to stdout:** `echo "Error: ..." >&2`. Callers pipe stdout; log messages must not pollute it.
- **`trap cleanup EXIT`** for temporary files or background processes. Clean up on any exit path.
- **`shellcheck`** as linter when available in the project. Address all warnings — they catch real bugs (SC2086, SC2046, SC2155).
