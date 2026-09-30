package agents

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repositoryProfiles returns the real profiles file. Use it only for
// invariants that must hold for any valid content, never for exact values.
func repositoryProfiles(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "agent-profiles.json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// syntheticProfiles returns the fixed profiles under testdata, whose values
// exist nowhere in the real profiles.
func syntheticProfiles(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "profiles.json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// syntheticRole returns the canonical source path and bytes of a role under
// testdata/roles.
func syntheticRole(t *testing.T, name string) (string, []byte) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "roles", name+".md"))
	if err != nil {
		t.Fatal(err)
	}
	return "content/agents/synthetic/" + name + ".md", data
}

// replaceOnce replaces the first occurrence of old and fails if it is absent,
// so a mutation of the synthetic profiles cannot silently do nothing.
func replaceOnce(t *testing.T, s, old, replacement string) string {
	t.Helper()
	if !strings.Contains(s, old) {
		t.Fatalf("synthetic profiles do not contain %q", old)
	}
	return strings.Replace(s, old, replacement, 1)
}

var allHosts = []string{"claude", "codex", "grok", "pi", "opencode", "cursor"}

// effortKeyPrefix is how each host's rendered effort line starts. Hosts that
// do not emit an effort are absent from the map.
var effortKeyPrefix = map[string]string{
	"claude": `effort: "`,
	"codex":  `model_reasoning_effort = "`,
	"pi":     `thinking: "`,
}

// hasLinePrefix reports whether any line of out starts with prefix.
func hasLinePrefix(out []byte, prefix string) bool {
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

// TestRepositoryRolesRenderOnEveryHostAndCarryEffortOnlyWhereEmitted renders
// every real role on all six hosts: each must render, Claude, Codex and Pi
// must declare an effort with their own key, and Grok, OpenCode and Cursor
// must start no line with any of those keys. The valid effort values are not
// repeated here; Parse and ReadProfiles enforce them.
func TestRepositoryRolesRenderOnEveryHostAndCarryEffortOnlyWhereEmitted(t *testing.T) {
	profiles := repositoryProfiles(t)
	sources, err := filepath.Glob(filepath.Join("..", "..", "content", "agents", "*", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) == 0 {
		t.Fatal("catalogue has no roles")
	}
	for _, source := range sources {
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		canonical := strings.TrimPrefix(filepath.ToSlash(source), "../../")
		for _, host := range allHosts {
			out, err := Render(canonical, data, profiles, host)
			if err != nil {
				t.Fatalf("Render(%s, %s): %v", host, canonical, err)
			}
			if want, emits := effortKeyPrefix[host]; emits {
				if !hasLinePrefix(out, want) {
					t.Fatalf("Render(%s, %s) has no line starting %q:\n%s", host, canonical, want, out)
				}
				continue
			}
			for _, prefix := range effortKeyPrefix {
				if hasLinePrefix(out, prefix) {
					t.Fatalf("Render(%s, %s) has a line starting %q, want no effort:\n%s", host, canonical, prefix, out)
				}
			}
		}
	}
}

func TestPiObserveUsesNativeSimpleToolList(t *testing.T) {
	profiles := syntheticProfiles(t)
	source, data := syntheticRole(t, "synthetic-observer")
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
	profiles := syntheticProfiles(t)
	source, data := syntheticRole(t, "synthetic-observer")
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

	// The unicode role puts double quotes, single quotes, backslashes, a tab
	// and non-ASCII text in both the description and the instruction body.
	source, data = syntheticRole(t, "synthetic-unicode")
	out, err = Render(source, data, profiles, "codex")
	if err != nil {
		t.Fatal(err)
	}
	got = string(out)
	for _, want := range []string{
		// Quotes and backslashes are escaped; non-ASCII text is kept verbatim.
		`description = "A \"quoted\" role at C:\\work with mañana."` + "\n",
		`\"double quotes\", 'single quotes', C:\\\\work, and `,
		"mañana, naïve café, 日本語, and a tab:\\tend.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Codex TOML missing escaped text %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "\t") {
		t.Fatalf("Codex TOML holds a raw tab instead of an escape:\n%s", got)
	}
}

func TestCodexTOMLRoundTripsQuotedUnicodeInstructions(t *testing.T) {
	profiles := syntheticProfiles(t)
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
	bad := replaceOnce(t, string(syntheticProfiles(t)), `"excludeTools": ["write", "edit"]`, `"excludeTools": "write,edit"`)
	if _, err := ReadProfiles([]byte(bad)); err == nil {
		t.Fatal("ReadProfiles accepted scalar Pi tool restriction")
	}
}

func TestCursorObserveRendersInheritModelAndReadonly(t *testing.T) {
	profiles := syntheticProfiles(t)
	source, data := syntheticRole(t, "synthetic-observer")
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
	bad := replaceOnce(t, string(syntheticProfiles(t)), `"execution": {"model": "inherit"}, "reasoning": {"model": "inherit"}, "inherit": {"model": "inherit"}`, `"execution": {"model": "inherit", "effort": "high"}, "reasoning": {"model": "inherit"}, "inherit": {"model": "inherit"}`)
	if _, err := ReadProfiles([]byte(bad)); err == nil || !strings.Contains(err.Error(), "inherited effort") {
		t.Fatalf("ReadProfiles accepted a Cursor model effort: %v", err)
	}
}

func TestProfilesRejectCursorReadonlyAsString(t *testing.T) {
	bad := replaceOnce(t, string(syntheticProfiles(t)), `"observe": {"readonly": true}`, `"observe": {"readonly": "true"}`)
	if _, err := ReadProfiles([]byte(bad)); err == nil {
		t.Fatal("ReadProfiles accepted a string Cursor readonly value")
	}
}

func TestProfilesRejectCursorReadonlyFalse(t *testing.T) {
	bad := replaceOnce(t, string(syntheticProfiles(t)), `"observe": {"readonly": true}`, `"observe": {"readonly": false}`)
	if _, err := ReadProfiles([]byte(bad)); err == nil {
		t.Fatal("ReadProfiles accepted a false Cursor readonly value")
	}
}

func TestProfilesRejectCursorUnsupportedAccessKey(t *testing.T) {
	bad := replaceOnce(t, string(syntheticProfiles(t)), `"observe": {"readonly": true}`, `"observe": {"permissionMode": "plan"}`)
	if _, err := ReadProfiles([]byte(bad)); err == nil {
		t.Fatal("ReadProfiles accepted an unsupported Cursor access key")
	}
}

func TestProfilesRejectUnknownHost(t *testing.T) {
	bad := replaceOnce(t, string(syntheticProfiles(t)), `"cursor": {`, `"cursorx": {`)
	if _, err := ReadProfiles([]byte(bad)); err == nil || !strings.Contains(err.Error(), "unsupported agent profile host") {
		t.Fatalf("ReadProfiles accepted an unknown host key: %v", err)
	}
}

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
	profiles := syntheticProfiles(t)
	source := "content/agents/design/test-agent.md"
	// the synthetic execution profile carries effort "medium"; the role declares "low".
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
	profiles := syntheticProfiles(t)
	source := "content/agents/design/test-agent.md"
	data := []byte("---\nname: \"test-agent\"\ndescription: \"A role\"\nmodel_profile: \"execution\"\naccess_profile: \"observe\"\neffort: \"max\"\n---\nBody\n")
	for _, host := range []string{"grok", "opencode", "cursor"} {
		out, err := Render(source, data, profiles, host)
		if err != nil {
			t.Fatalf("Render(%s): %v", host, err)
		}
		for _, prefix := range effortKeyPrefix {
			if hasLinePrefix(out, prefix) {
				t.Fatalf("Render(%s) emitted a line starting %q despite the role declaring an effort:\n%s", host, prefix, out)
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
	if err := json.Unmarshal(syntheticProfiles(t), &frozen); err != nil {
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
	source, body := syntheticRole(t, "synthetic-observer")
	for _, host := range []string{"claude", "codex", "grok", "pi", "opencode"} {
		if _, err := Render(source, body, data, host); err != nil {
			t.Fatalf("Render(%s) on five-host profile: %v", host, err)
		}
	}
	if _, err := Render(source, body, data, "cursor"); err == nil || err.Error() != `unsupported agent host "cursor"` {
		t.Fatalf("Render(cursor) on five-host profile = %v, want unsupported agent host error", err)
	}
}
