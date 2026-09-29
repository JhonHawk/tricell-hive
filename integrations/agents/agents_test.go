package agents

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	source := "content/agents/review/hive-research.md"
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
	source := "content/agents/review/hive-research.md"
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
	source := "content/agents/review/hive-research.md"
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

// emittedEffort matches an effort key at the start of a rendered line.
var emittedEffort = regexp.MustCompile(`(?m)^(effort|thinking|model_reasoning_effort)\b`)

func TestParseRejectsInvalidEffortValue(t *testing.T) {
	source := "content/agents/design/test-agent.md"
	data := []byte("---\nname: \"test-agent\"\ndescription: \"A role\"\nmodel_profile: \"execution\"\naccess_profile: \"observe\"\neffort: \"extreme\"\n---\nBody\n")
	if _, err := Parse(source, data); err == nil {
		t.Fatal("Parse accepted an invalid effort value")
	}
}

func TestParseRejectsClaudeEffortAsUnknownField(t *testing.T) {
	source := "content/agents/design/test-agent.md"
	data := []byte("---\nname: \"test-agent\"\ndescription: \"A role\"\nmodel_profile: \"execution\"\naccess_profile: \"observe\"\nclaude_effort: \"high\"\n---\nBody\n")
	_, err := Parse(source, data)
	if err == nil || !strings.Contains(err.Error(), `unsupported agent field "claude_effort"`) {
		t.Fatalf("Parse(claude_effort) = %v, want unsupported agent field error", err)
	}
}

func TestRoleEffortOverridesProfileEffortOnClaudeCodexAndPi(t *testing.T) {
	profiles := repositoryProfiles(t)
	source := "content/agents/design/test-agent.md"
	// execution profile carries effort "high" by default; the role declares "low".
	data := []byte("---\nname: \"test-agent\"\ndescription: \"A role\"\nmodel_profile: \"execution\"\naccess_profile: \"observe\"\neffort: \"low\"\n---\nBody\n")
	cases := []struct{ host, want string }{
		{"claude", "effort: \"low\"\n"},
		{"codex", "model_reasoning_effort = \"low\"\n"},
		{"pi", "thinking: \"low\"\n"},
	}
	for _, c := range cases {
		out, err := Render(source, data, profiles, c.host)
		if err != nil {
			t.Fatalf("Render(%s): %v", c.host, err)
		}
		if !strings.Contains(string(out), c.want) {
			t.Fatalf("Render(%s) missing role effort override %q:\n%s", c.host, c.want, out)
		}
	}
}

func TestGrokOpenCodeAndCursorEmitNoEffortEvenWhenRoleDeclaresOne(t *testing.T) {
	profiles := repositoryProfiles(t)
	source := "content/agents/design/test-agent.md"
	data := []byte("---\nname: \"test-agent\"\ndescription: \"A role\"\nmodel_profile: \"execution\"\naccess_profile: \"observe\"\neffort: \"max\"\n---\nBody\n")
	for _, host := range []string{"grok", "opencode", "cursor"} {
		out, err := Render(source, data, profiles, host)
		if err != nil {
			t.Fatalf("Render(%s): %v", host, err)
		}
		if emittedEffort.Match(out) {
			t.Fatalf("Render(%s) emitted effort despite the role declaring one:\n%s", host, out)
		}
	}
}

// TestRepositorySourcesMatchExpectedEffortLevels renders the canonical roles
// for all six hosts and checks the exact Claude/Codex/Pi effort level: every
// Claude-rendered role must declare an effort line, and Grok/OpenCode/Cursor
// must never declare one.
func TestRepositorySourcesMatchExpectedEffortLevels(t *testing.T) {
	profiles := repositoryProfiles(t)
	// {Claude, Codex, Pi} expected effort level per role.
	expected := map[string][3]string{
		"hive-build-backend":       {"high", "high", "high"},
		"hive-build-frontend":      {"high", "high", "high"},
		"hive-build-kmp":           {"high", "high", "high"},
		"hive-build-infra":         {"high", "high", "high"},
		"hive-write-tests":         {"high", "high", "high"},
		"hive-verify-change":       {"high", "high", "high"},
		"hive-review-ux":           {"high", "high", "high"},
		"hive-research":            {"high", "high", "high"},
		"hive-write-spec":          {"medium", "medium", "medium"},
		"hive-read-state":          {"low", "low", "low"},
		"hive-build-data":          {"high", "medium", "medium"},
		"hive-tune-performance":    {"high", "medium", "medium"},
		"hive-review-code":         {"high", "medium", "medium"},
		"hive-verify-task":         {"high", "medium", "medium"},
		"hive-review-harness":      {"high", "medium", "medium"},
		"hive-review-plan":         {"medium", "medium", "medium"},
		"hive-design-architecture": {"high", "high", "high"},
		"hive-design-ui":           {"high", "high", "high"},
		"hive-refute-claim":        {"high", "high", "high"},
		"hive-review-security":     {"max", "max", "max"},
	}
	sources, err := filepath.Glob(filepath.Join("..", "..", "content", "agents", "*", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != len(expected) {
		t.Fatalf("catalogue has %d roles, want %d", len(sources), len(expected))
	}
	claudeEffort := regexp.MustCompile(`(?m)^effort: "([a-z]+)"$`)
	codexEffort := regexp.MustCompile(`(?m)^model_reasoning_effort = "([a-z]+)"$`)
	piEffort := regexp.MustCompile(`(?m)^thinking: "([a-z]+)"$`)
	for _, source := range sources {
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		name := strings.TrimSuffix(filepath.Base(source), ".md")
		want, ok := expected[name]
		if !ok {
			t.Fatalf("no expected effort levels declared for role %s", name)
		}
		canonical := filepath.ToSlash(source)
		canonical = strings.TrimPrefix(canonical, "../../")

		claudeOut, err := Render(canonical, data, profiles, "claude")
		if err != nil {
			t.Fatalf("Render(claude, %s): %v", name, err)
		}
		m := claudeEffort.FindSubmatch(claudeOut)
		if m == nil {
			t.Fatalf("Render(claude, %s) has no effort line:\n%s", name, claudeOut)
		}
		if got := string(m[1]); got != want[0] {
			t.Fatalf("Render(claude, %s) effort = %q, want %q", name, got, want[0])
		}

		codexOut, err := Render(canonical, data, profiles, "codex")
		if err != nil {
			t.Fatalf("Render(codex, %s): %v", name, err)
		}
		m = codexEffort.FindSubmatch(codexOut)
		if m == nil {
			t.Fatalf("Render(codex, %s) has no model_reasoning_effort line:\n%s", name, codexOut)
		}
		if got := string(m[1]); got != want[1] {
			t.Fatalf("Render(codex, %s) effort = %q, want %q", name, got, want[1])
		}

		piOut, err := Render(canonical, data, profiles, "pi")
		if err != nil {
			t.Fatalf("Render(pi, %s): %v", name, err)
		}
		m = piEffort.FindSubmatch(piOut)
		if m == nil {
			t.Fatalf("Render(pi, %s) has no thinking line:\n%s", name, piOut)
		}
		if got := string(m[1]); got != want[2] {
			t.Fatalf("Render(pi, %s) effort = %q, want %q", name, got, want[2])
		}

		for _, host := range []string{"grok", "opencode", "cursor"} {
			out, err := Render(canonical, data, profiles, host)
			if err != nil {
				t.Fatalf("Render(%s, %s): %v", host, name, err)
			}
			if emittedEffort.Match(out) {
				t.Fatalf("Render(%s, %s) emitted effort:\n%s", host, name, out)
			}
		}
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
	source := "content/agents/review/hive-research.md"
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
