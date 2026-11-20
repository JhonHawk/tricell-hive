package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// globalGuidanceLimit is the size budget of the always-loaded global guidance.
// Each session and each role child loads it whole, and each prune before this
// budget existed grew back within weeks, so a new rule goes to a leaf and the
// core keeps only what must act unread.
const globalGuidanceLimit = 22000

// globalGuidanceProblem measures root/content/guidance/global.md and returns
// its size and, when it exceeds the budget, a message naming where a new rule
// goes.
func globalGuidanceProblem(root string) (int, string, error) {
	data, err := os.ReadFile(filepath.Join(root, "content", "guidance", "global.md"))
	if err != nil {
		return 0, "", err
	}
	if len(data) <= globalGuidanceLimit {
		return len(data), "", nil
	}
	return len(data), fmt.Sprintf("content/guidance/global.md is %d bytes, over its %d-byte budget: move the rule's procedure to a rule leaf or its skill and keep in the global core only its invariant, when it must act without the leaf being read (AGENTS.md «Guidance structure»)", len(data), globalGuidanceLimit), nil
}

// TestGlobalGuidanceSize enforces the budget and reports the size. Source size
// does not establish host loading or actual token use; scenario unions are
// reported separately by TestGuidanceReadingFootprint.
func TestGlobalGuidanceSize(t *testing.T) {
	size, problem, err := globalGuidanceProblem("../..")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile("../../content/guidance/global.md")
	t.Logf("content/guidance/global.md: %d bytes, %d characters", size, utf8.RuneCount(data))
	if problem != "" {
		t.Error(problem)
	}
}

func TestGlobalGuidanceBudgetBoundary(t *testing.T) {
	for _, tc := range []struct {
		size int
		over bool
	}{
		{globalGuidanceLimit, false},
		{globalGuidanceLimit + 1, true},
	} {
		root := t.TempDir()
		dir := filepath.Join(root, "content", "guidance")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "global.md"), []byte(strings.Repeat("a", tc.size)), 0o644); err != nil {
			t.Fatal(err)
		}
		_, problem, err := globalGuidanceProblem(root)
		if err != nil {
			t.Fatal(err)
		}
		if got := problem != ""; got != tc.over {
			t.Errorf("size %d: over budget = %v, want %v", tc.size, got, tc.over)
		}
		if tc.over && !strings.Contains(problem, "rule leaf") {
			t.Errorf("size %d: message %q does not name where the rule goes", tc.size, problem)
		}
	}
}
