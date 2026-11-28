package content

import (
	"os"
	"testing"

	"tricell-hive/integrations/codex"
	"tricell-hive/integrations/mdlinks"
)

// The Codex developer-context entry fits the size the hook prints once its
// skill: links point at a long install path, so it never goes silent in Codex.
func TestCodexDeveloperContextFitsTheHookLimit(t *testing.T) {
	data, err := os.ReadFile("../../content/skills/" + codex.DeveloperContextFile)
	if err != nil {
		t.Fatal(err)
	}
	printed := mdlinks.RewriteSkillLinks(string(data), "/Users/me/Library/Application Support/Example Host/.agents/skills")
	if len(printed) > codex.DeveloperContextLimit {
		t.Fatalf("%s prints %d bytes with its links rewritten, over the %d-byte limit of the Codex hook: shorten the entry; the limit rests on Codex's per-message cap", codex.DeveloperContextFile, len(printed), codex.DeveloperContextLimit)
	}
}
