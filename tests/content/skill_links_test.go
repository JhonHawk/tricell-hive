package content

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// skillLinkTarget matches the destination of a Markdown link that uses the
// skill:owner/path locator. This test checks that each file the global
// guidance links exists; tooling/management/references.go validates their
// anchors when a release is installed.
var skillLinkTarget = regexp.MustCompile(`\]\(skill:([^)\s]+)\)`)

func TestGlobalGuidanceSkillLinksResolve(t *testing.T) {
	data, err := os.ReadFile("../../content/guidance/global.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range skillLinkTarget.FindAllStringSubmatch(string(data), -1) {
		target, _, _ := strings.Cut(m[1], "#")
		path := filepath.Join("../../content/skills", filepath.FromSlash(target))
		if _, err := os.Stat(path); err != nil {
			t.Errorf("global.md links skill:%s, but %s does not exist", m[1], path)
		}
	}
}
