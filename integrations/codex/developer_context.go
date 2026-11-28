package codex

import (
	"os"
	"path/filepath"

	"tricell-hive/integrations/hostevent"
	"tricell-hive/integrations/mdlinks"
)

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

// DeveloperContext is what `hive hook codex <event>` prints for payload: the
// entry under skillsDir with its skill: links pointing into skillsDir, for a
// SessionStart of the main session only. A subagent's event (agent_id), any
// other event, a missing or unreadable entry and one over the limit print
// nothing.
func DeveloperContext(event string, payload map[string]any, skillsDir string) string {
	if hostevent.Normalize(event) != "sessionstart" || hostevent.ID(payload["agent_id"]) != "" || skillsDir == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(skillsDir, DeveloperContextFile))
	if err != nil {
		return ""
	}
	text := mdlinks.RewriteSkillLinks(string(data), skillsDir)
	if len(text) > DeveloperContextLimit {
		return ""
	}
	return text
}
