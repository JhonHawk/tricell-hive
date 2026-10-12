package content

import (
	"os"
	"testing"
	"unicode/utf8"
)

// TestGlobalGuidanceSize reports the size of the always-loaded global
// guidance without failing. No supported host truncates or skips the
// user-level instruction file near this size (checked 2026-10-11), so a byte
// ceiling protects no technical limit; a change that grows the file states
// the size change in its pull request instead (AGENTS.md, Guidance structure).
func TestGlobalGuidanceSize(t *testing.T) {
	data, err := os.ReadFile("../../content/guidance/global.md")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("content/guidance/global.md: %d bytes, %d characters", len(data), utf8.RuneCount(data))
}
