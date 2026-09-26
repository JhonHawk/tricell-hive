package content

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// skillLinkTarget matches the destination of a Markdown link that uses the
// skill:owner/path locator. Release link validation covers skills and roles
// but not the global guidance, so this test covers its pointers.
var skillLinkTarget = regexp.MustCompile(`\]\(skill:([^)\s]+)\)`)

func TestGlobalGuidanceSkillLinksResolve(t *testing.T) {
	data, err := os.ReadFile("../../content/guidance/global.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range skillLinkTarget.FindAllStringSubmatch(string(data), -1) {
		path := filepath.Join("../../content/skills", filepath.FromSlash(m[1]))
		if _, err := os.Stat(path); err != nil {
			t.Errorf("global.md links skill:%s, but %s does not exist", m[1], path)
		}
	}
}
