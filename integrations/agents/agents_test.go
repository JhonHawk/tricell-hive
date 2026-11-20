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

var allHosts = []string{"claude", "codex", "grok", "pi", "opencode", "cursor", "antigravity"}

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
// every real role on all seven hosts: each must render, Claude, Codex and Pi
// must declare an effort with their own key, and Grok, OpenCode, Cursor and Antigravity
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
			out, err := Render(canonical, data, profiles, host, "/skills", nil)
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

// Grok removes a tool from a role only by its tool ID: `write` removes the
// write tool and `search_replace` the edit tool, while the daemon display
// name `Write` removes nothing (Grok 1.0.55 user guide, headless mode).
func TestGrokObserveDeniesWriteAndEditToolIDs(t *testing.T) {
	source, data := syntheticRole(t, "synthetic-observer")
	out, err := Render(source, data, repositoryProfiles(t), "grok", "/skills", nil)
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	for _, id := range []string{`"write"`, `"search_replace"`} {
		if !strings.Contains(got, id) {
			t.Fatalf("Grok observe role must deny tool ID %s:\n%s", id, got)
		}
	}
}

func TestCodexTOMLRoundTripsQuotedUnicodeInstructions(t *testing.T) {
	profiles := syntheticProfiles(t)
	source := "content/agents/design/test-agent.md"
	// DEL and a control character are not valid raw in a TOML basic string, so
	// tomllib rejects the output unless tomlQuote escapes both.
	body := "Use \"quoted\" values, C:\\\\work, and mañana.\x7f\x01\n"
	data := []byte("---\nname: test-agent\ndescription: \"A \\\"quoted\\\" role at C:\\\\work\"\nmodel_profile: inherit\naccess_profile: observe\n---\n" + body)
	out, err := Render(source, data, profiles, "codex", "/skills", nil)
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

// Cursor and Antigravity inherit the session's effort, so a profile that names
// one is rejected on either host.
func TestProfilesRejectEffortOnHostsThatInheritIt(t *testing.T) {
	cases := []struct{ name, old, new string }{
		{"cursor", `"execution": {"model": "inherit"}, "reasoning": {"model": "inherit"}, "inherit": {"model": "inherit"}`,
			`"execution": {"model": "inherit", "effort": "high"}, "reasoning": {"model": "inherit"}, "inherit": {"model": "inherit"}`},
		{"antigravity", `"antigravity": {
      "models": {"execution": {"model": "inherit"}`, `"antigravity": {
      "models": {"execution": {"model": "inherit", "effort": "high"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bad := replaceOnce(t, string(syntheticProfiles(t)), c.old, c.new)
			if _, err := ReadProfiles([]byte(bad)); err == nil || !strings.Contains(err.Error(), "inherited effort") {
				t.Fatalf("ReadProfiles accepted a %s model effort: %v", c.name, err)
			}
		})
	}
}

func TestProfilesRejectAntigravityAccessKeys(t *testing.T) {
	for _, key := range []string{`"readonly": true`, `"tools": ["read_file"]`, `"permissionMode": "plan"`} {
		bad := replaceOnce(t, string(syntheticProfiles(t)), `"antigravity": {
      "models": {"execution": {"model": "inherit"}, "reasoning": {"model": "inherit"}, "inherit": {"model": "inherit"}, "verifier": {"model": "inherit"}},
      "access": {"observe": {}`, `"antigravity": {
      "models": {"execution": {"model": "inherit"}, "reasoning": {"model": "inherit"}, "inherit": {"model": "inherit"}, "verifier": {"model": "inherit"}},
      "access": {"observe": {`+key+`}`)
		if _, err := ReadProfiles([]byte(bad)); err == nil || !strings.Contains(err.Error(), "unsupported agent access key") {
			t.Fatalf("ReadProfiles accepted Antigravity access %s: %v", key, err)
		}
	}
}

// Cursor accepts only the boolean `readonly: true` as an observe restriction.
func TestProfilesRejectCursorReadonlyThatIsNotTrue(t *testing.T) {
	for _, c := range []struct{ name, value string }{
		{"false", `false`},
		{"string", `"true"`},
	} {
		t.Run(c.name, func(t *testing.T) {
			bad := replaceOnce(t, string(syntheticProfiles(t)), `"observe": {"readonly": true}`, `"observe": {"readonly": `+c.value+`}`)
			if _, err := ReadProfiles([]byte(bad)); err == nil || !strings.Contains(err.Error(), "invalid Cursor readonly value") {
				t.Fatalf("ReadProfiles accepted Cursor readonly %s: %v", c.value, err)
			}
		})
	}
}

func TestProfilesRejectCursorUnsupportedAccessKey(t *testing.T) {
	bad := replaceOnce(t, string(syntheticProfiles(t)), `"observe": {"readonly": true}`, `"observe": {"permissionMode": "plan"}`)
	if _, err := ReadProfiles([]byte(bad)); err == nil {
		t.Fatal("ReadProfiles accepted an unsupported Cursor access key")
	}
}

// Releases frozen before 2026-10-16 embed sandbox_mode in the Codex observe
// profile. They must stay readable, and the key must not reach the TOML.
func TestCodexLegacySandboxModeIsReadableAndNotRendered(t *testing.T) {
	profiles := string(syntheticProfiles(t))
	codex := strings.Index(profiles, `"codex"`)
	if codex < 0 {
		t.Fatal("synthetic profiles lack a codex host")
	}
	for _, mode := range []string{"read-only", "workspace-write"} {
		legacy := profiles[:codex] + strings.Replace(profiles[codex:], `"observe": {}`, `"observe": {"sandbox_mode": "`+mode+`"}`, 1)
		if legacy == profiles {
			t.Fatal("codex observe block not found")
		}
		if _, err := ReadProfiles([]byte(legacy)); err != nil {
			t.Fatalf("ReadProfiles rejected legacy Codex sandbox_mode %q: %v", mode, err)
		}
		source, data := syntheticRole(t, "synthetic-observer")
		out, err := Render(source, data, []byte(legacy), "codex", "/skills", nil)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(out), "sandbox_mode") {
			t.Fatalf("Codex TOML rendered the legacy sandbox_mode %q:\n%s", mode, out)
		}
	}
	bad := profiles[:codex] + strings.Replace(profiles[codex:], `"observe": {}`, `"observe": {"sandbox_mode": "danger-full-access"}`, 1)
	if _, err := ReadProfiles([]byte(bad)); err == nil || !strings.Contains(err.Error(), "Codex agent sandbox") {
		t.Fatalf("ReadProfiles accepted an unknown Codex sandbox value: %v", err)
	}
}

func TestProfilesRejectUnknownHost(t *testing.T) {
	bad := replaceOnce(t, string(syntheticProfiles(t)), `"cursor": {`, `"cursorx": {`)
	if _, err := ReadProfiles([]byte(bad)); err == nil || !strings.Contains(err.Error(), "unsupported agent profile host") {
		t.Fatalf("ReadProfiles accepted an unknown host key: %v", err)
	}
}

// parseRoleSource is a role source whose frontmatter is front, with a body.
func parseRoleSource(front string) []byte {
	return []byte("---\n" + front + "---\nBody\n")
}

const parseValidFront = "name: \"test-agent\"\ndescription: \"A role\"\nmodel_profile: \"execution\"\naccess_profile: \"observe\"\n"

// Each row breaks exactly one rule of Parse, so its message identifies the
// guard that rejected it.
func TestParseRejectsInvalidRoleSources(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want string
	}{
		{"frontmatter missing", []byte("name: \"test-agent\"\nBody\n"), "agent frontmatter missing"},
		{"frontmatter unclosed", []byte("---\n" + parseValidFront + "Body\n"), "agent frontmatter unclosed"},
		{"invalid scalar", parseRoleSource(strings.Replace(parseValidFront, `"A role"`, `"A role`, 1)), "invalid agent scalar description"},
		{"scalar not JSON quoted", parseRoleSource(strings.Replace(parseValidFront, `"A role"`, `it's a role`, 1)), "quote agent scalar description as JSON"},
		{"empty field", parseRoleSource(strings.Replace(parseValidFront, `"A role"`, `""`, 1)), "empty or multiline agent field description"},
		{"multiline field", parseRoleSource(strings.Replace(parseValidFront, `"A role"`, `"A\nrole"`, 1)), "empty or multiline agent field description"},
		{"duplicate key", parseRoleSource(parseValidFront + "name: \"other\"\n"), "invalid or duplicate agent field"},
		{"required field absent", parseRoleSource(strings.Replace(parseValidFront, "description: \"A role\"\n", "", 1)), "agent requires description"},
		{"invalid effort", parseRoleSource(parseValidFront + "effort: \"extreme\"\n"), "invalid agent effort"},
		{"invalid effort_claude", parseRoleSource(parseValidFront + "effort_claude: \"extreme\"\n"), "invalid agent effort"},
		{"name differs from file", parseRoleSource(strings.Replace(parseValidFront, `"test-agent"`, `"other-agent"`, 1)), "agent name/body mismatch"},
		{"unknown model profile", parseRoleSource(strings.Replace(parseValidFront, `"execution"`, `"bogus"`, 1)), "unknown agent profile"},
	}
	const source = "content/agents/design/test-agent.md"
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Parse(source, c.data); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Parse = %v, want an error containing %q", err, c.want)
			}
		})
	}
}

// Parse accepts CRLF line endings in the frontmatter and reads the same role
// as with LF.
func TestParseAcceptsCRLFFrontmatter(t *testing.T) {
	const source = "content/agents/design/test-agent.md"
	front := parseValidFront + "effort: \"low\"\n"
	want, err := Parse(source, parseRoleSource(front))
	if err != nil {
		t.Fatal(err)
	}
	crlf := []byte("---\r\n" + strings.ReplaceAll(front, "\n", "\r\n") + "---\r\nBody\n")
	got, err := Parse(source, crlf)
	if err != nil {
		t.Fatalf("Parse rejected a CRLF frontmatter: %v", err)
	}
	if got != want {
		t.Fatalf("Parse(CRLF) = %+v, want %+v", got, want)
	}
}

func TestParseRejectsUnknownFrontmatterKey(t *testing.T) {
	source := "content/agents/design/test-agent.md"
	data := []byte("---\nname: \"test-agent\"\ndescription: \"A role\"\nmodel_profile: \"execution\"\naccess_profile: \"observe\"\nclaude_effort: \"high\"\n---\nBody\n")
	_, err := Parse(source, data)
	if err == nil || !strings.Contains(err.Error(), `unsupported agent field "claude_effort"`) {
		t.Fatalf("Parse(claude_effort) = %v, want unsupported agent field error", err)
	}
}

// A release frozen before Cursor or before Antigravity has fewer host
// profiles. It must keep validating and rendering for the hosts it has so a
// saved plan can still be reinstalled with --release, and a host it lacks
// fails with a clear, stable error rather than a validation error over a
// missing host.
func TestFrozenReleasesWithFewerHostsStillValidateAndMissingHostFailsCleanly(t *testing.T) {
	cases := []struct {
		name    string
		drop    []string
		present []string
		missing []string
	}{
		{"five hosts", []string{"cursor", "antigravity"}, []string{"claude", "codex", "grok", "pi", "opencode"}, []string{"cursor", "antigravity"}},
		{"six hosts", []string{"antigravity"}, []string{"claude", "codex", "grok", "pi", "opencode", "cursor"}, []string{"antigravity"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var frozen map[string]any
			if err := json.Unmarshal(syntheticProfiles(t), &frozen); err != nil {
				t.Fatal(err)
			}
			for _, host := range c.drop {
				delete(frozen["hosts"].(map[string]any), host)
			}
			data, err := json.Marshal(frozen)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ReadProfiles(data); err != nil {
				t.Fatalf("profiles without %v rejected: %v", c.drop, err)
			}
			source, body := syntheticRole(t, "synthetic-observer")
			for _, host := range c.present {
				if _, err := Render(source, body, data, host, "/skills", nil); err != nil {
					t.Fatalf("Render(%s) on profiles without %v: %v", host, c.drop, err)
				}
			}
			for _, host := range c.missing {
				if _, err := Render(source, body, data, host, "/skills", nil); err == nil || err.Error() != `unsupported agent host "`+host+`"` {
					t.Fatalf("Render(%s) without a profile = %v, want unsupported agent host error", host, err)
				}
			}
		})
	}
}
