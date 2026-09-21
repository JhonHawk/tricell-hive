package main

import (
	"testing"
	"time"
)

func TestGrokTimeBudgetPreservesExplicitLimitsAndOtherCases(t *testing.T) {
	for _, c := range []struct {
		name, suite, host, task string
		explicit                bool
		limit, want             time.Duration
	}{
		{"plan", "flows", "grok", "plan", false, 3 * time.Minute, 10 * time.Minute},
		{"build", "flows", "grok", "build", false, 3 * time.Minute, 10 * time.Minute},
		{"explicit screening budget", "flows", "grok", "plan", true, 3 * time.Minute, 3 * time.Minute},
		{"explicit zero remains invalid", "flows", "grok", "plan", true, 0, 0},
		{"research unchanged", "flows", "grok", "research", false, 3 * time.Minute, 3 * time.Minute},
		{"other host unchanged", "flows", "codex", "plan", false, 3 * time.Minute, 3 * time.Minute},
		{"historical suite unchanged", "workspace-conventions", "grok", "plan", false, 3 * time.Minute, 3 * time.Minute},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := pilotTimeLimit(c.limit, c.explicit, c.suite, c.host, c.task); got != c.want {
				t.Fatalf("time limit = %v, want %v", got, c.want)
			}
		})
	}
}

func TestNativeTrustExceptionIsNarrow(t *testing.T) {
	before := []byte("model = \"gpt-6-astra\"\n")
	table := []byte("\n[projects.\"/temporary/repo\"]\ntrust_level = \"trusted\"\n")
	if !onlyTrustAdded(before, append(append([]byte{}, before...), table...), "/temporary/repo") {
		t.Fatal("expected exact insertion")
	}
	bad := append([]byte("model = \"other\"\n"), table...)
	if onlyTrustAdded(before, bad, "/temporary/repo") {
		t.Fatal("accepted other setting change")
	}
	if onlyTrustAdded(before, append(append([]byte{}, before...), table...), "/different/repo") {
		t.Fatal("accepted different target")
	}
	if onlyTrustAdded(before, append(append(append([]byte{}, before...), table...), table...), "/temporary/repo") {
		t.Fatal("accepted duplicates")
	}
}
func TestTrustInsertedBetweenExistingTables(t *testing.T) {
	before := []byte("[projects.\"/old\"]\ntrust_level = \"trusted\"\n\n[notice]\nflag = true\n")
	after := []byte("[projects.\"/old\"]\ntrust_level = \"trusted\"\n\n[projects.\"/temporary/repo\"]\ntrust_level = \"trusted\"\n\n[notice]\nflag = true\n")
	if !onlyTrustAdded(before, after, "/temporary/repo") {
		t.Fatal("table separator falsely classified as unrelated change")
	}
}
