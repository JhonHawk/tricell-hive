// Package agents renders canonical roles directly into native agent definitions.
// It does not run a host, change personal settings, or write generated trees.
package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"
	"tricell-hive/integrations/mdlinks"
	"unicode/utf8"
)

const Version = "1"
const ProfilesSource = "integrations/agent-profiles.json"

// ModelOverride replaces part of the model a host receives for one role. An
// empty field keeps the release's value. Only the part that is set is checked
// by ValidateOverride and applied by Resolve.
type ModelOverride struct {
	Model  string `json:"model,omitempty"`
	Effort string `json:"effort,omitempty"`
}

type Role struct {
	Name, Description, ModelProfile, AccessProfile, Body, Effort, ClaudeEffort string
}
type Model struct {
	Model  string `json:"model,omitempty"`
	Effort string `json:"effort,omitempty"`
}
type Host struct {
	Models map[string]Model                      `json:"models"`
	Access map[string]map[string]json.RawMessage `json:"access"`
}
type Profiles struct {
	Version int             `json:"version"`
	Hosts   map[string]Host `json:"hosts"`
}

var slug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func IsSource(source string) bool {
	p := strings.Split(source, "/")
	return len(p) == 4 && p[0] == "content" && p[1] == "agents" && slug.MatchString(p[2]) && strings.HasSuffix(p[3], ".md") && slug.MatchString(strings.TrimSuffix(p[3], ".md"))
}

// Parse accepts a deliberately small YAML subset: required single-line scalar keys and an optional effort.
// JSON-quoted scalars provide unambiguous Unicode/escaping without a YAML runtime.
func Parse(source string, data []byte) (Role, error) {
	var r Role
	if !IsSource(source) || !utf8.Valid(data) || bytes.Contains(data, []byte{0}) {
		return r, fmt.Errorf("invalid agent source %s", source)
	}
	s := string(data)
	header := "---\n"
	if strings.HasPrefix(s, "---\r\n") {
		header = "---\r\n"
	}
	if !strings.HasPrefix(s, header) {
		return r, fmt.Errorf("agent frontmatter missing: %s", source)
	}
	s = s[len(header):]
	end := regexp.MustCompile(`\r?\n---\r?\n`).FindStringIndex(s)
	if end == nil {
		return r, fmt.Errorf("agent frontmatter unclosed: %s", source)
	}
	front, body := strings.ReplaceAll(s[:end[0]], "\r\n", "\n"), s[end[1]:]
	fields := map[string]string{}
	for _, line := range strings.Split(front, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) != key || fields[key] != "" {
			return r, fmt.Errorf("invalid or duplicate agent field: %s", line)
		}
		switch key {
		case "name", "description", "model_profile", "access_profile", "effort", "effort_claude":
		default:
			return r, fmt.Errorf("unsupported agent field %q", key)
		}
		value = strings.TrimSpace(value)
		if strings.HasPrefix(value, `"`) {
			if err := json.Unmarshal([]byte(value), &value); err != nil {
				return r, fmt.Errorf("invalid agent scalar %s", key)
			}
		} else if strings.ContainsAny(value, "#\"'{}[]&*!|>") {
			return r, fmt.Errorf("quote agent scalar %s as JSON", key)
		}
		if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "\r\n\x00") {
			return r, fmt.Errorf("empty or multiline agent field %s", key)
		}
		fields[key] = value
	}
	for _, key := range []string{"name", "description", "model_profile", "access_profile"} {
		if fields[key] == "" {
			return r, fmt.Errorf("agent requires %s", key)
		}
	}
	r = Role{fields["name"], fields["description"], fields["model_profile"], fields["access_profile"], body, fields["effort"], fields["effort_claude"]}
	if (r.Effort != "" && !oneOf(r.Effort, "low", "medium", "high", "xhigh", "max")) ||
		(r.ClaudeEffort != "" && !oneOf(r.ClaudeEffort, "low", "medium", "high", "xhigh", "max")) {
		return r, fmt.Errorf("invalid agent effort")
	}
	if r.Name != strings.TrimSuffix(path.Base(source), ".md") || strings.TrimSpace(r.Body) == "" {
		return r, fmt.Errorf("agent name/body mismatch: %s", source)
	}
	if !oneOf(r.ModelProfile, "execution", "reasoning", "inherit", "verifier") || !oneOf(r.AccessProfile, "observe", "implement", "verify") {
		return r, fmt.Errorf("unknown agent profile: %s", source)
	}
	return r, nil
}

func oneOf(value string, allowed ...string) bool {
	for _, s := range allowed {
		if value == s {
			return true
		}
	}
	return false
}

func ReadProfiles(data []byte) (Profiles, error) {
	var p Profiles
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		return p, fmt.Errorf("agent profiles: %w", err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return p, fmt.Errorf("trailing agent profile data")
	}
	if p.Version != 1 {
		return p, fmt.Errorf("agent profiles require version 1 and the five base hosts")
	}
	for host := range p.Hosts {
		switch host {
		case "claude", "codex", "grok", "pi", "opencode", "cursor":
		default:
			return p, fmt.Errorf("unsupported agent profile host %q", host)
		}
	}
	required := []string{"claude", "codex", "grok", "pi", "opencode"}
	hosts := append([]string(nil), required...)
	if _, ok := p.Hosts["cursor"]; ok {
		hosts = append(hosts, "cursor")
	}
	if len(p.Hosts) != len(hosts) {
		return p, fmt.Errorf("agent profiles require version 1 and the five base hosts")
	}
	for _, host := range hosts {
		h, ok := p.Hosts[host]
		// verifier is optional so a frozen release with three model profiles stays readable.
		models := []string{"execution", "reasoning", "inherit"}
		if _, ok := h.Models["verifier"]; ok {
			models = append(models, "verifier")
		}
		if !ok || len(h.Models) != len(models) || len(h.Access) != 3 {
			return p, fmt.Errorf("incomplete profiles for %s", host)
		}
		for _, name := range models {
			m, ok := h.Models[name]
			if !ok || strings.ContainsAny(m.Model+m.Effort, "\r\n\x00") {
				return p, fmt.Errorf("invalid model profile %s/%s", host, name)
			}
			if m.Effort != "" && !oneOf(m.Effort, "low", "medium", "high", "xhigh", "max", "ultra") {
				return p, fmt.Errorf("invalid effort for %s", host)
			}
			if !acceptsProfileEffort(host) && m.Effort != "" {
				return p, fmt.Errorf("%s uses inherited effort or a model variant", host)
			}
			if host == "opencode" && m.Effort != "" && (m.Model == "" || strings.Contains(m.Model, "#")) {
				return p, fmt.Errorf("OpenCode effort requires a base model without a variant")
			}
		}
		for _, name := range []string{"observe", "implement", "verify"} {
			a, ok := h.Access[name]
			if !ok || a == nil {
				return p, fmt.Errorf("missing access profile %s/%s", host, name)
			}
			if err := validateAccess(host, a); err != nil {
				return p, err
			}
		}
	}
	return p, nil
}

func validateAccess(host string, values map[string]json.RawMessage) error {
	for key, raw := range values {
		switch {
		case host == "codex" && key == "sandbox_mode":
			var s string
			if json.Unmarshal(raw, &s) != nil || !oneOf(s, "read-only", "workspace-write") {
				return fmt.Errorf("invalid Codex agent sandbox")
			}
		case host == "cursor" && key == "readonly":
			var b bool
			if json.Unmarshal(raw, &b) != nil || !b {
				return fmt.Errorf("invalid Cursor readonly value")
			}
		case (host == "claude" || host == "grok") && key == "permissionMode":
			var s string
			if json.Unmarshal(raw, &s) != nil || !oneOf(s, "default", "plan") {
				return fmt.Errorf("invalid %s permissionMode", host)
			}
		case (host == "claude" || host == "grok") && key == "disallowedTools" || host == "pi" && key == "excludeTools":
			var list []string
			if json.Unmarshal(raw, &list) != nil || len(list) == 0 {
				return fmt.Errorf("invalid %s tool restrictions", host)
			}
			for _, s := range list {
				if s == "" || strings.ContainsAny(s, "\r\n\x00") {
					return fmt.Errorf("invalid tool name")
				}
			}
		case host == "opencode" && key == "permissions":
			var rules []struct {
				Action   string `json:"action"`
				Resource string `json:"resource"`
				Effect   string `json:"effect"`
			}
			d := json.NewDecoder(bytes.NewReader(raw))
			d.DisallowUnknownFields()
			if d.Decode(&rules) != nil || len(rules) == 0 {
				return fmt.Errorf("invalid OpenCode permissions")
			}
			for _, r := range rules {
				if !oneOf(r.Action, "edit", "subagent") || r.Resource != "*" || r.Effect != "deny" {
					return fmt.Errorf("unsupported OpenCode restriction")
				}
			}
		default:
			return fmt.Errorf("unsupported agent access key %s/%s", host, key)
		}
	}
	return nil
}

// acceptsProfileEffort reports whether a host allows an effort in its profile.
// OpenCode encodes it into the model variant rather than emitting a field.
func acceptsProfileEffort(host string) bool {
	return host == "claude" || host == "codex" || host == "pi" || host == "opencode"
}

// acceptsRoleEffort reports whether a role can replace the selected profile's
// effort. OpenCode only supports this when the profile declares a base model
// plus effort; legacy profiles with a fixed model variant remain unchanged.
func acceptsRoleEffort(host string, m Model) bool {
	return host == "claude" || host == "codex" || host == "pi" ||
		(host == "opencode" && m.Model != "" && m.Effort != "")
}

func Validate(source string, data, profiles []byte) error {
	if _, err := Parse(source, data); err != nil {
		return err
	}
	_, err := ReadProfiles(profiles)
	return err
}

// Resolve returns a role's model profile and the model and effort a host
// receives for it. It is the single place that decides the effective effort:
// effort_claude replaces the profile's effort on Claude Code only, and a
// role's effort replaces it on every host that can represent it.
func Resolve(source string, data, profiles []byte, host string, override *ModelOverride) (profile string, m Model, err error) {
	r, err := Parse(source, data)
	if err != nil {
		return "", m, err
	}
	p, err := ReadProfiles(profiles)
	if err != nil {
		return "", m, err
	}
	h, ok := p.Hosts[host]
	if !ok {
		return "", m, fmt.Errorf("unsupported agent host %q", host)
	}
	m, ok = h.Models[r.ModelProfile]
	if !ok {
		return "", m, fmt.Errorf("agent profiles for %s have no %s model profile", host, r.ModelProfile)
	}
	if host == "claude" && r.ClaudeEffort != "" {
		m.Effort = r.ClaudeEffort
	} else if r.Effort != "" && acceptsRoleEffort(host, m) {
		m.Effort = r.Effort
	}
	// OpenCode receives the effort as a model variant, not a separate field.
	if host == "opencode" && m.Effort != "" {
		m.Model += "#" + m.Effort
		m.Effort = ""
	}
	return r.ModelProfile, m, nil
}

// Render builds a host's native role file. skillsDir is the directory the
// installer places skills in for the target scope; Render stores it in the
// role body in place of each skill: locator (see mdlinks.RewriteSkillLinks).
func Render(source string, data, profiles []byte, host, skillsDir string, override *ModelOverride) ([]byte, error) {
	_, m, err := Resolve(source, data, profiles, host, override)
	if err != nil {
		return nil, err
	}
	r, err := Parse(source, data)
	if err != nil {
		return nil, err
	}
	// Rewrite before encoding: Codex's TOML string is one escaped line, which can
	// no longer be scanned for fenced code.
	r.Body = mdlinks.RewriteSkillLinks(r.Body, skillsDir)
	p, err := ReadProfiles(profiles)
	if err != nil {
		return nil, err
	}
	h := p.Hosts[host]
	fields := map[string]any{"name": r.Name, "description": r.Description}
	for key, raw := range h.Access[r.AccessProfile] {
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		fields[key] = value
	}
	if m.Model != "" {
		fields["model"] = m.Model
	}
	if m.Effort != "" {
		key := "effort"
		if host == "codex" {
			key = "model_reasoning_effort"
		}
		if host == "pi" {
			key = "thinking"
		}
		fields[key] = m.Effort
	}
	switch host {
	case "grok":
		fields["promptMode"] = "extend"
		fields["agentsMd"] = true
	case "opencode":
		delete(fields, "name")
		fields["mode"] = "subagent"
	case "pi":
		fields["systemPromptMode"] = "append"
		fields["inheritProjectContext"] = true
		fields["inheritGlobalContext"] = true
		fields["inheritSkills"] = true
	}
	var out strings.Builder
	if host != "codex" {
		out.WriteString("---\n")
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if host == "codex" {
			s, ok := fields[k].(string)
			if !ok {
				return nil, fmt.Errorf("non-string Codex field %s", k)
			}
			fmt.Fprintf(&out, "%s = %s\n", k, tomlQuote(s))
		} else if host == "pi" && k == "excludeTools" {
			// pi-subagents uses a simple comma-separated scalar, not a YAML array.
			list := fields[k].([]any)
			names := make([]string, len(list))
			for i, item := range list {
				name := item.(string)
				if strings.ContainsAny(name, ",\"'[]# \t") {
					return nil, fmt.Errorf("unsupported Pi tool name")
				}
				names[i] = name
			}
			fmt.Fprintf(&out, "%s: %s\n", k, strings.Join(names, ","))
		} else if host == "pi" && k == "description" {
			// Literal blocks preserve quotes/backslashes with Pi's lightweight parser.
			fmt.Fprintf(&out, "%s: |-\n  %s\n", k, fields[k])
		} else {
			b, err := json.Marshal(fields[k])
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(&out, "%s: %s\n", k, b)
		}
	}
	if host == "codex" {
		fmt.Fprintf(&out, "developer_instructions = %s\n", tomlQuote(r.Body))
	} else {
		out.WriteString("---\n")
		out.WriteString(r.Body)
	}
	return []byte(out.String()), nil
}

// Go's Quote uses non-TOML \x escapes for some controls. JSON uses only escapes
// accepted by TOML basic strings, with DEL explicitly escaped as well.
func tomlQuote(s string) string {
	b, _ := json.Marshal(s)
	return strings.ReplaceAll(string(b), string(rune(127)), `\u007f`)
}
