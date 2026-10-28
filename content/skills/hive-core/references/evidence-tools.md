# Evidence tools

## When to read

Read before looking up version-sensitive external technical documentation, and before locating symbols, callers, or the impact of a change in a repository with a `.codegraph/` directory at its root. The global guidance keeps the invariants: a recommendation names the documentation it rests on, and you never create a CodeGraph index. Reading this reference does not invoke a workflow.

## External documentation

For version-sensitive external technical documentation, prefer Context7 as the first lookup through the installed documentation skill or integration, matching the relevant version; when access or coverage falls short, use current official documentation and state the verification limits. Inspect repository-local facts directly.

## CodeGraph

When a repository has a CodeGraph index, a `.codegraph/` directory at its root, and CodeGraph's MCP tool or `codegraph` command is available, use it to locate symbols, callers, and the impact of a change before text search; otherwise use text search. Never create an index, since indexing is the user's decision. Its MCP server keeps the index current while its session runs; the `codegraph` command does not, so before relying on callers, callees, or impact from that command, run `codegraph sync -q` first; a read-only or delegated role does not sync, and reports those relations as possibly stale. CodeGraph links some imports and calls by name when it cannot resolve them, without marking the guess: confirm with text search, or by reading the importing file, a result that crosses applications in a monorepo or comes from a workspace-package import or a path alias the root `tsconfig.json` does not declare.
