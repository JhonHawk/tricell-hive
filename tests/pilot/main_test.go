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

// D17-A: Codex writes its own "[projects.\"<fixture>\"]\ntrust_level =
// \"trusted\"\n" registration into the shadow CODEX_HOME/config.toml the
// runner generates for a guidance-variant run (with --codex-bypass-sandbox),
// and the protected-resources check flagged that expected addition as an
// unrecognized change. onlyTrustEntriesAdded is the shadow-home mirror of
// onlyTrustAdded (which checks one named path in the real home): it accepts
// a shadow config.toml change consisting only of one or more of Codex's own
// project-trust blocks, for any path, and nothing else.
func TestOnlyTrustEntriesAddedAcceptsShadowTrustRegistration(t *testing.T) {
	before := []byte("[mcp_servers.engram]\ncommand = \"node\"\nargs = [\"server.js\"]\n")
	one := []byte("\n[projects.\"/repo/fixture\"]\ntrust_level = \"trusted\"\n")
	if !onlyTrustEntriesAdded(before, append(append([]byte{}, before...), one...)) {
		t.Fatal("expected a single shadow trust insertion to be allowed")
	}
	two := append(append([]byte{}, one...), []byte("\n[projects.\"/repo/other\"]\ntrust_level = \"trusted\"\n")...)
	if !onlyTrustEntriesAdded(before, append(append([]byte{}, before...), two...)) {
		t.Fatal("expected multiple shadow trust insertions (one per visited fixture path) to be allowed")
	}
}

func TestOnlyTrustEntriesAddedRejectsAnyOtherShadowChange(t *testing.T) {
	before := []byte("[mcp_servers.engram]\ncommand = \"node\"\nargs = [\"server.js\"]\n")
	trust := []byte("\n[projects.\"/repo/fixture\"]\ntrust_level = \"trusted\"\n")
	// A trust entry alongside an unrelated setting change is still flagged.
	otherEdit := append([]byte("[mcp_servers.engram]\ncommand = \"python\"\nargs = [\"server.py\"]\n"), trust...)
	if onlyTrustEntriesAdded(before, otherEdit) {
		t.Fatal("accepted an unrelated setting change alongside a trust entry")
	}
	// A non-"trusted" value is not the recognized registration.
	notTrusted := append(append([]byte{}, before...), []byte("\n[projects.\"/repo/fixture\"]\ntrust_level = \"workspace\"\n")...)
	if onlyTrustEntriesAdded(before, notTrusted) {
		t.Fatal("accepted a non-trusted trust_level value")
	}
	// Nothing added at all is not this exemption's concern (changedProtected
	// would not even call it in that case, but the function must not claim
	// a change occurred when none is present).
	if onlyTrustEntriesAdded(before, before) {
		t.Fatal("accepted an unchanged file as an addition")
	}
	// A removed block alongside an added one is still flagged.
	before2 := append(append([]byte{}, before...), []byte("\n[projects.\"/repo/old\"]\ntrust_level = \"trusted\"\n")...)
	swapped := append(append([]byte{}, before...), trust...)
	if onlyTrustEntriesAdded(before2, swapped) {
		t.Fatal("accepted a removed trust block alongside an added one")
	}
}
