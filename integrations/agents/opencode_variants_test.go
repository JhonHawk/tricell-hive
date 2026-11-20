package agents

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// openCodeGoProfiles is the release profiles as they stood when OpenCode roles
// ran on OpenCode Go models with declared variants. The repository profiles now
// inherit the session model on OpenCode, so the variant mapping, which still
// applies to model overrides and to frozen releases, is tested against this copy.
var openCodeGoProfiles = filepath.Join("testdata", "opencode-go-profiles.json")

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
	// An extra entry no profile uses, so the rejection can only come from the
	// guard the row names and not from a profile that depends on the table.
	withExtra := func(entry string) string { return strings.Replace(table, "}", ", "+entry+"}", 1) }
	for name, c := range map[string]struct{ from, to, want string }{
		"profile effort not offered":   {`"execution": {"model": "opencode-go/deepseek-v4.1-flash", "effort": "high"}`, `"execution": {"model": "opencode-go/deepseek-v4.1-flash", "effort": "medium"}`, "does not offer"},
		"embedded variant not offered": {`deepseek-v4.1-flash#max`, `deepseek-v4.1-flash#medium`, "does not offer"},
		"empty list":                   {table, withExtra(`"x/extra": []`), "declares no variants"},
		"unknown effort":               {table, withExtra(`"x/extra": ["fast"]`), "invalid or repeated variant"},
		"repeated effort":              {table, withExtra(`"x/extra": ["low", "low"]`), "invalid or repeated variant"},
		"model with variant":           {table, withExtra(`"x/extra#high": ["high"]`), "invalid model"},
		"empty table":                  {table, `"variants": {}`, "empty variants table"},
		"other host declares one":      {`"codex": {`, `"codex": {"variants": {"m": ["low"]}, `, "does not declare model variants"},
	} {
		bad := strings.Replace(string(base), c.from, c.to, 1)
		if bad == string(base) {
			t.Fatalf("%s: replacement did not apply", name)
		}
		if _, err := ReadProfiles([]byte(bad)); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: ReadProfiles = %v, want an error containing %q", name, err, c.want)
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

// TestOpenCodeModelOnlyOverrideMapsAgainstNewModel asserts the variant each
// effort maps to on the OpenCode Go profiles, with and without a model
// override. OpenCode 2.0.27 rejects a launch that names a variant the model
// does not offer with "Variant unavailable for <model>: <variant>" before any
// provider call.
func TestOpenCodeModelOnlyOverrideMapsAgainstNewModel(t *testing.T) {
	const flash, pro = "opencode-go/deepseek-v4.1-flash", "opencode-go/deepseek-v4-pro"
	for _, c := range []struct {
		name, roleEffort string
		o                *ModelOverride
		want             string
	}{
		// Without an override the role's effort maps to the release model's variants.
		{"no override medium", "medium", nil, flash + "#high"},
		{"no override xhigh", "xhigh", nil, flash + "#max"},
		// With a model override the role's effort maps against the rendered model,
		// not the release's: flash keeps low, pro offers high and max only.
		{"new model xhigh", "xhigh", &ModelOverride{Model: pro}, pro + "#max"},
		{"new model low", "low", &ModelOverride{Model: pro}, pro + "#high"},
		// An unlisted model keeps the role's effort unmapped.
		{"unlisted model", "medium", &ModelOverride{Model: "x/y"}, "x/y#medium"},
	} {
		t.Run(c.name, func(t *testing.T) {
			m, err := resolveOpenCode(t, "execution", c.roleEffort, c.o)
			if err != nil || m.Model != c.want {
				t.Fatalf("got %v, %v; want %s", m, err, c.want)
			}
		})
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
