package content

import (
	"os"
	"testing"
	"unicode/utf8"
)

// TestGlobalGuidanceSize reports the size of the always-loaded global
// guidance without imposing a quota. Source size does not establish host
// loading or actual token use. A guidance change states its size change in
// the pull request (AGENTS.md, Guidance structure); scenario unions are
// reported separately by TestGuidanceReadingFootprint.
func TestGlobalGuidanceSize(t *testing.T) {
	data, err := os.ReadFile("../../content/guidance/global.md")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("content/guidance/global.md: %d bytes, %d characters", len(data), utf8.RuneCount(data))
}
