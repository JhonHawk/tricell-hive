package agents

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// openCodeGoProfiles is the release profiles as they stood when OpenCode roles
// ran on OpenCode Go models with declared variants. The repository profiles now
// inherit the session model on OpenCode, so the variant mapping, which still
// applies to model overrides and to frozen releases, is tested against this copy.
var openCodeGoProfiles = filepath.Join("testdata", "opencode-go-profiles.json")

var renderedOpenCodeModel = regexp.MustCompile(`(?m)^model: "?([^"\n#]+)(?:#([^"\n]+))?"?$`)

// TestOpenCodeRolesRequestOfferedVariants renders every shipped role for
// OpenCode with the OpenCode Go profiles and fails when the resulting variant is
// not one the model offers. OpenCode 2.0.27 rejects such a launch with
// "Variant unavailable for <model>: <variant>" before any provider call.
func TestOpenCodeRolesRequestOfferedVariants(t *testing.T) {
	profiles, err := os.ReadFile(openCodeGoProfiles)
	if err != nil {
		t.Fatal(err)
	}
	p, err := ReadProfiles(profiles)
	if err != nil {
		t.Fatal(err)
	}
	declared := p.Hosts["opencode"].Variants
	if len(declared) == 0 {
		t.Fatal("repository profiles declare no OpenCode variants")
	}
	sources, err := filepath.Glob(filepath.Join("..", "..", "content", "agents", "*", "*.md"))
	if err != nil || len(sources) == 0 {
		t.Fatalf("no roles found: %v", err)
	}
	for _, path := range sources {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		source := "content/agents/" + filepath.Base(filepath.Dir(path)) + "/" + filepath.Base(path)
		out, err := Render(source, data, profiles, "opencode", "/skills", nil)
		if err != nil {
			t.Fatalf("Render(%s): %v", source, err)
		}
		m := renderedOpenCodeModel.FindSubmatch(out)
		if m == nil {
			t.Fatalf("%s: no model line in the OpenCode render:\n%s", source, out)
		}
		model, variant := string(m[1]), string(m[2])
		offered, known := declared[model]
		if !known {
			t.Errorf("%s: model %q has no declared variants", source, model)
			continue
		}
		if variant == "" || !oneOf(variant, offered...) {
			t.Errorf("%s: %s requests variant %q, offered %v", source, model, variant, offered)
		}
	}
}

func TestNearestOffered(t *testing.T) {
	flash := []string{"low", "high", "max"}
	for _, c := range []struct {
		effort  string
		offered []string
		want    string
	}{
		{"medium", flash, "high"},
		{"low", flash, "low"},
		{"xhigh", flash, "max"},
		{"ultra", flash, "max"},
		{"medium", []string{"high", "max"}, "high"},
		{"low", []string{"high", "max"}, "high"},
	} {
		if got := nearestOffered(c.effort, c.offered); got != c.want {
			t.Errorf("nearestOffered(%s, %v) = %s, want %s", c.effort, c.offered, got, c.want)
		}
	}
}

func TestReadProfilesOpenCodeVariants(t *testing.T) {
	base, err := os.ReadFile(openCodeGoProfiles)
	if err != nil {
		t.Fatal(err)
	}
	table := `"variants": {"opencode-go/deepseek-v4.1-flash": ["low", "high", "max"], "opencode-go/deepseek-v4-pro": ["high", "max"], "opencode-go/glm-5.3": ["low", "high", "max"]}`
	for name, c := range map[string]struct{ from, to string }{
		"profile effort not offered":   {`"execution": {"model": "opencode-go/deepseek-v4.1-flash", "effort": "high"}`, `"execution": {"model": "opencode-go/deepseek-v4.1-flash", "effort": "medium"}`},
		"embedded variant not offered": {`deepseek-v4.1-flash#max`, `deepseek-v4.1-flash#medium`},
		"empty list":                   {table, `"variants": {"opencode-go/glm-5.3": []}`},
		"unknown effort":               {table, `"variants": {"opencode-go/glm-5.3": ["fast"]}`},
		"repeated effort":              {table, `"variants": {"opencode-go/glm-5.3": ["low", "low"]}`},
		"model with variant":           {table, `"variants": {"opencode-go/glm-5.3#high": ["high"]}`},
		"empty table":                  {table, `"variants": {}`},
		"other host declares one":      {`"codex": {`, `"codex": {"variants": {"m": ["low"]}, `},
	} {
		bad := strings.Replace(string(base), c.from, c.to, 1)
		if bad == string(base) {
			t.Fatalf("%s: replacement did not apply", name)
		}
		if _, err := ReadProfiles([]byte(bad)); err == nil {
			t.Errorf("%s: ReadProfiles accepted it", name)
		}
	}
}

func resolveOpenCode(t *testing.T, profile, roleEffort string, o *ModelOverride) (Model, error) {
	t.Helper()
	profiles, err := os.ReadFile(openCodeGoProfiles)
	if err != nil {
		t.Fatal(err)
	}
	_, m, err := Resolve(resolveSource, resolveRole(profile, roleEffort), profiles, "opencode", o)
	return m, err
}

func TestOpenCodeOverrideEffortNotOfferedIsRefused(t *testing.T) {
	_, err := resolveOpenCode(t, "execution", "", &ModelOverride{Effort: "medium"})
	var oe *OverrideError
	if !errors.As(err, &oe) || !strings.Contains(err.Error(), "deepseek-v4.1-flash") || !strings.Contains(err.Error(), "low, high, max") {
		t.Fatalf("want an OverrideError naming the model and variants, got %v", err)
	}
	m, err := resolveOpenCode(t, "execution", "", &ModelOverride{Effort: "max"})
	if err != nil || m.Model != "opencode-go/deepseek-v4.1-flash#max" {
		t.Fatalf("offered override effort: %v %v", m, err)
	}
	// A model with no table entry keeps accepting any valid effort.
	m, err = resolveOpenCode(t, "execution", "", &ModelOverride{Model: "x/y", Effort: "medium"})
	if err != nil || m.Model != "x/y#medium" {
		t.Fatalf("unlisted model: %v %v", m, err)
	}
}

func TestOpenCodeModelOnlyOverrideMapsAgainstNewModel(t *testing.T) {
	// A role's medium maps against the model that is rendered, not the release's.
	m, err := resolveOpenCode(t, "execution", "medium", &ModelOverride{Model: "opencode-go/deepseek-v4-pro"})
	if err != nil || m.Model != "opencode-go/deepseek-v4-pro#high" {
		t.Fatalf("got %v %v", m, err)
	}
	// An unlisted model keeps the role's effort unmapped.
	m, err = resolveOpenCode(t, "execution", "medium", &ModelOverride{Model: "x/y"})
	if err != nil || m.Model != "x/y#medium" {
		t.Fatalf("got %v %v", m, err)
	}
}

func TestReadProfilesRequiresEntryForEveryVariantModel(t *testing.T) {
	base, _ := os.ReadFile(openCodeGoProfiles)
	for name, c := range map[string]struct{ from, to string }{
		"typo in effort model": {`"model": "opencode-go/glm-5.3", "effort": "high"`, `"model": "opencode-go/glm-5-3", "effort": "high"`},
		"embedded variant":     {`"model": "opencode-go/deepseek-v4.1-flash#max"`, `"model": "opencode-go/other#max"`},
	} {
		bad := strings.Replace(string(base), c.from, c.to, 1)
		if bad == string(base) {
			t.Fatalf("%s: replacement did not apply", name)
		}
		if _, err := ReadProfiles([]byte(bad)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
