---
order: 180
targets: [claude]
---

### Deferred Tools

- **Deferred tools surface only by name at session start** when tool search is enabled — most MCP server tools and several built-ins. The session's own listing is authoritative; never work from a remembered list. Calling a deferred tool directly returns `InputValidationError`.
- **Load via `ToolSearch` before first use.** Use `ToolSearch({query: "select:Name1,Name2"})` for direct selection, or a keyword query to discover relevant tools. When you'll use several tools from one MCP server, batch the load into a single call.
