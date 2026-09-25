package agents

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func repositoryProfiles(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "agent-profiles.json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCatalogueRendersAllRolesForEveryHost(t *testing.T) {
	profiles := repositoryProfiles(t)
	sources, err := filepath.Glob(filepath.Join("..", "..", "content", "agents", "*", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 20 {
		t.Fatalf("catalogue has %d roles, want 20", len(sources))
	}
	for _, source := range sources {
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		canonical := filepath.ToSlash(source)
		canonical = strings.TrimPrefix(canonical, "../../")
		for _, host := range []string{"claude", "codex", "grok", "pi", "opencode", "cursor"} {
			if _, err := Render(canonical, data, profiles, host); err != nil {
				t.Fatalf("Render(%s, %s): %v", host, canonical, err)
			}
		}
	}
}

func TestPiObserveUsesNativeSimpleToolList(t *testing.T) {
	profiles := repositoryProfiles(t)
	source := "content/agents/review/sdd-explore.md"
	data, err := os.ReadFile(filepath.Join("..", "..", source))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(source, data, profiles, "pi")
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	// pi-subagents v0.67.0 parses list fields by splitting a scalar on commas;
	// JSON array syntax would make the two effective names `[\"write\"` and
	// `\"edit\"]`, neither of which matches Pi's built-in tools.
	if !strings.Contains(got, "excludeTools: write,edit\n") {
		t.Fatalf("Pi excludeTools must be a native comma-separated scalar:\n%s", got)
	}
	for _, field := range []string{
		"systemPromptMode: \"append\"\n",
		"inheritProjectContext: true\n",
		"inheritGlobalContext: true\n",
		"inheritSkills: true\n",
	} {
		if !strings.Contains(got, field) {
			t.Fatalf("Pi output missing %q:\n%s", strings.TrimSpace(field), got)
		}
	}
}

func TestCodexTOMLIsFlatAndEscapesInstructionBody(t *testing.T) {
	profiles := repositoryProfiles(t)
	source := "content/agents/review/sdd-explore.md"
	data, err := os.ReadFile(filepath.Join("..", "..", source))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(source, data, profiles, "codex")
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	for _, prefix := range []string{
		"description = ",
		"name = ",
		"sandbox_mode = \"read-only\"\n",
		"developer_instructions = ",
	} {
		if !strings.Contains(got, prefix) {
			t.Fatalf("Codex TOML missing %q:\n%s", prefix, got)
		}
	}
	if strings.Contains(got, "\n---\n") || strings.Contains(got, "\\x") {
		t.Fatalf("Codex TOML contains unsupported frontmatter or escape:\n%s", got)
	}
}

func TestCodexTOMLRoundTripsQuotedUnicodeInstructions(t *testing.T) {
	profiles := repositoryProfiles(t)
	source := "content/agents/design/test-agent.md"
	body := "Use \"quoted\" values, C:\\\\work, and mañana.\n"
	data := []byte("---\nname: test-agent\ndescription: \"A \\\"quoted\\\" role at C:\\\\work\"\nmodel_profile: inherit\naccess_profile: observe\n---\n" + body)
	out, err := Render(source, data, profiles, "codex")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("python3", "-c", "import json, sys, tomllib; print(json.dumps(tomllib.loads(sys.stdin.read())))")
	command.Stdin = strings.NewReader(string(out))
	decoded, err := command.Output()
	if err != nil {
		t.Fatalf("tomllib rejected rendered Codex TOML: %v\n%s", err, out)
	}
	var fields map[string]string
	if err := json.Unmarshal(decoded, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["description"] != `A "quoted" role at C:\work` || fields["developer_instructions"] != body {
		t.Fatalf("TOML round trip changed fields: %#v", fields)
	}
}

func TestProfilesRejectWrongPiToolRestrictionType(t *testing.T) {
	bad := strings.Replace(string(repositoryProfiles(t)), `"excludeTools": ["write", "edit"]`, `"excludeTools": "write,edit"`, 1)
	if _, err := ReadProfiles([]byte(bad)); err == nil {
		t.Fatal("ReadProfiles accepted scalar Pi tool restriction")
	}
}

func TestCursorObserveRendersInheritModelAndReadonly(t *testing.T) {
	profiles := repositoryProfiles(t)
	source := "content/agents/review/sdd-explore.md"
	data, err := os.ReadFile(filepath.Join("..", "..", source))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(source, data, profiles, "cursor")
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	for _, field := range []string{`model: "inherit"` + "\n", "readonly: true\n"} {
		if !strings.Contains(got, field) {
			t.Fatalf("Cursor output missing %q:\n%s", strings.TrimSpace(field), got)
		}
	}
	if strings.Contains(got, "effort") {
		t.Fatalf("Cursor output must not declare effort:\n%s", got)
	}
}

func TestProfilesRejectCursorEffort(t *testing.T) {
	bad := strings.Replace(string(repositoryProfiles(t)), `"execution": {"model": "inherit"}, "reasoning": {"model": "inherit"}, "inherit": {"model": "inherit"}`, `"execution": {"model": "inherit", "effort": "high"}, "reasoning": {"model": "inherit"}, "inherit": {"model": "inherit"}`, 1)
	if _, err := ReadProfiles([]byte(bad)); err == nil || !strings.Contains(err.Error(), "inherited effort") {
		t.Fatalf("ReadProfiles accepted a Cursor model effort: %v", err)
	}
}

func TestProfilesRejectCursorReadonlyAsString(t *testing.T) {
	bad := strings.Replace(string(repositoryProfiles(t)), `"observe": {"readonly": true}`, `"observe": {"readonly": "true"}`, 1)
	if _, err := ReadProfiles([]byte(bad)); err == nil {
		t.Fatal("ReadProfiles accepted a string Cursor readonly value")
	}
}

func TestProfilesRejectCursorReadonlyFalse(t *testing.T) {
	bad := strings.Replace(string(repositoryProfiles(t)), `"observe": {"readonly": true}`, `"observe": {"readonly": false}`, 1)
	if _, err := ReadProfiles([]byte(bad)); err == nil {
		t.Fatal("ReadProfiles accepted a false Cursor readonly value")
	}
}

func TestProfilesRejectCursorUnsupportedAccessKey(t *testing.T) {
	bad := strings.Replace(string(repositoryProfiles(t)), `"observe": {"readonly": true}`, `"observe": {"permissionMode": "plan"}`, 1)
	if _, err := ReadProfiles([]byte(bad)); err == nil {
		t.Fatal("ReadProfiles accepted an unsupported Cursor access key")
	}
}

func TestProfilesRejectUnknownHost(t *testing.T) {
	bad := strings.Replace(string(repositoryProfiles(t)), `"cursor": {`, `"cursorx": {`, 1)
	if _, err := ReadProfiles([]byte(bad)); err == nil || !strings.Contains(err.Error(), "unsupported agent profile host") {
		t.Fatalf("ReadProfiles accepted an unknown host key: %v", err)
	}
}

// A release frozen before Cursor existed has exactly five host profiles. It
// must keep validating and rendering for those hosts so a saved plan can
// still be reinstalled with --release, and Cursor on it fails with a clear,
// stable error rather than a validation error over a missing host.
func TestFiveHostProfilesStillValidateAndCursorFailsCleanly(t *testing.T) {
	var frozen map[string]any
	if err := json.Unmarshal(repositoryProfiles(t), &frozen); err != nil {
		t.Fatal(err)
	}
	hosts := frozen["hosts"].(map[string]any)
	delete(hosts, "cursor")
	data, err := json.Marshal(frozen)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ReadProfiles(data); err != nil {
		t.Fatalf("five-host profile rejected: %v", err)
	}
	source := "content/agents/review/sdd-explore.md"
	body, err := os.ReadFile(filepath.Join("..", "..", source))
	if err != nil {
		t.Fatal(err)
	}
	for _, host := range []string{"claude", "codex", "grok", "pi", "opencode"} {
		if _, err := Render(source, body, data, host); err != nil {
			t.Fatalf("Render(%s) on five-host profile: %v", host, err)
		}
	}
	if _, err := Render(source, body, data, "cursor"); err == nil || err.Error() != `unsupported agent host "cursor"` {
		t.Fatalf("Render(cursor) on five-host profile = %v, want unsupported agent host error", err)
	}
}
