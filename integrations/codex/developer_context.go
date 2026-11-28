package codex

// DeveloperContextFile is the entry Hive's Codex SessionStart hook prints as
// developer context, relative to the skills directory Hive installs into. Codex
// sends AGENTS.md and skills as user messages, where the message formats they
// set lost to Codex's own formatting; text a SessionStart hook prints reaches
// the model as developer context instead.
const DeveloperContextFile = "hive-core/assets/codex-developer-context.md"

// DeveloperContextLimit caps the entry, after its skill: links are rewritten to
// absolute paths: Codex keeps about 2,500 tokens of each hook message the model
// sees, and 8,000 bytes is about 2,000 tokens.
const DeveloperContextLimit = 8000
