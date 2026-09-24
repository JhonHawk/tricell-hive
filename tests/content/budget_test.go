package content

import (
	"os"
	"testing"
)

// globalGuidanceBudget is the byte ceiling for the always-loaded global guidance.
// It only tightens: raise it in the same change as the growth it allows, and
// lower it when the file shrinks by more than the slack.
const (
	globalGuidanceBudget = 32518
	globalGuidanceSlack  = 1024
)

func TestGlobalGuidanceStaysWithinRatchetedBudget(t *testing.T) {
	data, err := os.ReadFile("../../content/guidance/global.md")
	if err != nil {
		t.Fatal(err)
	}
	size := len(data)
	if size > globalGuidanceBudget {
		t.Fatalf("content/guidance/global.md is %d bytes, over the %d-byte budget; shrink it or raise globalGuidanceBudget deliberately in this change", size, globalGuidanceBudget)
	}
	if size < globalGuidanceBudget-globalGuidanceSlack {
		t.Fatalf("content/guidance/global.md shrank to %d bytes; lower globalGuidanceBudget to %d to keep the reduction", size, size)
	}
}
