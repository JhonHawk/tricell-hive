package antigravity

import "testing"

func TestRulePartNumberOnlyAcceptsHiveFilesInTheRulesFolder(t *testing.T) {
	home := "/h/.gemini/config"
	if got := RulePath(home, 2); got != "/h/.gemini/config/rules/hive-02.md" {
		t.Fatalf("RulePath = %s", got)
	}
	for path, want := range map[string]int{RulePath(home, 1): 1, RulePath(home, 12): 12} {
		if n, ok := RulePartNumber(home, path); !ok || n != want {
			t.Fatalf("%s -> %d %v", path, n, ok)
		}
	}
	for _, path := range []string{
		"/h/.gemini/config/rules/other.md", "/h/.gemini/config/rules/hive-00.md", "/h/.gemini/config/rules/hive-1.md",
		"/h/.gemini/config/rules/hive-001.md", "/h/.gemini/config/rules/sub/hive-01.md", "/h/.gemini/config/agents/hive-01.md",
		"/h/.cursor/rules/hive-01.md",
	} {
		if _, ok := RulePartNumber(home, path); ok {
			t.Fatalf("%s accepted", path)
		}
	}
	if _, ok := RulePartNumber("", RulePath("", 1)); ok {
		t.Fatal("empty home accepted")
	}
}
