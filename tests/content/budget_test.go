package content

import (
	"os"
	"testing"
)

// globalGuidanceBudget is the byte ceiling for the always-loaded global
// guidance, set about 1 KiB above its size when last raised. Growth past it
// fails until the ceiling is raised deliberately in the same change; small
// growth under it and any shrink need no edit here.
const globalGuidanceBudget = 45535

func TestGlobalGuidanceStaysUnderBudget(t *testing.T) {
	data, err := os.ReadFile("../../content/guidance/global.md")
	if err != nil {
		t.Fatal(err)
	}
	if size := len(data); size > globalGuidanceBudget {
		t.Fatalf("content/guidance/global.md is %d bytes, over the %d-byte budget; shrink it or raise globalGuidanceBudget deliberately in this change", size, globalGuidanceBudget)
	}
}
