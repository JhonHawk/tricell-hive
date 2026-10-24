package viewer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// fixturesRoot is the shared executable contract written by T1 and also run
// by tests/skills/walkthrough_render_test.py.
const fixturesRoot = "../../tests/fixtures/walkthrough"

// invalidFields mirrors INVALID_FIELDS in tests/skills/walkthrough_render_test.py:
// the field each invalid fixture must name in its error.
var invalidFields = map[string]string{
	"unknown-field":       "bogus",
	"missing-required":    "title",
	"schema-version":      "schema_version",
	"duplicate-flow-id":   "flows[1].id",
	"bad-flow-id":         "flows[0].id",
	"absolute-path":       "flows[0].captures[0].path",
	"dotdot-path":         "flows[0].captures[0].path",
	"outside-folder":      "flows[0].captures[0].path",
	"bad-segment":         "flows[0].captures[0].path",
	"symlink":             "flows[0].captures[0].path",
	"capture-extension":   "flows[0].captures[0].path",
	"recording-extension": "flows[0].recording.path",
	"missing-file":        "flows[0].captures[0].path",
	"bad-viewport":        "flows[0].captures[0].viewport",
}

func subdirs(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

// copyFixture copies fixtures/<rel> into a fresh temporary folder. The
// "symlink" invalid fixture needs its link created, as the Python test does.
func copyFixture(t *testing.T, rel string) string {
	t.Helper()
	src := filepath.Join(fixturesRoot, filepath.FromSlash(rel))
	dst := filepath.Join(t.TempDir(), "case")
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		r, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, r)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	if rel == "invalid/symlink" {
		if err := os.Symlink("real.png", filepath.Join(dst, "images", "link.png")); err != nil {
			t.Fatal(err)
		}
	}
	return dst
}

func TestManifestFixtureSetsMatchTheContract(t *testing.T) {
	var want []string
	for name := range invalidFields {
		want = append(want, name)
	}
	sort.Strings(want)
	if got := subdirs(t, filepath.Join(fixturesRoot, "invalid")); !reflect.DeepEqual(got, want) {
		t.Fatalf("invalid fixtures = %v, table has %v", got, want)
	}
	if got := subdirs(t, filepath.Join(fixturesRoot, "valid")); len(got) < 5 {
		t.Fatalf("expected the five valid fixtures, got %v", got)
	}
}

func TestManifestAcceptsEveryValidFixture(t *testing.T) {
	for _, name := range subdirs(t, filepath.Join(fixturesRoot, "valid")) {
		t.Run(name, func(t *testing.T) {
			folder := copyFixture(t, "valid/"+name)
			m, err := ValidateManifest(folder)
			if err != nil {
				t.Fatalf("valid fixture rejected: %v", err)
			}
			if m.SessionID == "" || m.Title == "" || len(m.Raw) == 0 {
				t.Fatalf("manifest summary incomplete: %+v", m)
			}
			for _, f := range m.Files {
				if _, err := os.Lstat(filepath.Join(folder, filepath.FromSlash(f))); err != nil {
					t.Fatalf("listed file %q missing: %v", f, err)
				}
			}
		})
	}
}

func TestManifestFinalFixtureListsItsReferencedFiles(t *testing.T) {
	m, err := ValidateManifest(copyFixture(t, "valid/final"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"images/nav-1.png", "images/nav-2.png", "images/nav-3.png", "images/settings-1.png"}
	if !reflect.DeepEqual(m.Files, want) {
		t.Fatalf("Files = %v, want %v", m.Files, want)
	}
	if m.Lang != "es" || m.Kind != "final" || m.SessionHost != "claude-code" {
		t.Fatalf("unexpected summary: %+v", m)
	}
}

func TestManifestRejectsEveryInvalidFixtureNamingTheField(t *testing.T) {
	for name, field := range invalidFields {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateManifest(copyFixture(t, "invalid/"+name))
			if err == nil {
				t.Fatal("invalid fixture accepted")
			}
			var me *ManifestError
			if !errors.As(err, &me) {
				t.Fatalf("error is not a *ManifestError: %T %v", err, err)
			}
			if me.Field != field || !strings.Contains(err.Error(), field) {
				t.Fatalf("error names %q (%v), want field %q", me.Field, err, field)
			}
		})
	}
}

func readManifestMap(t *testing.T, folder string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(folder, "walkthrough.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func writeManifestMap(t *testing.T, folder string, m map[string]any) {
	t.Helper()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "walkthrough.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func firstFlow(m map[string]any) map[string]any {
	return m["flows"].([]any)[0].(map[string]any)
}

func firstCapture(m map[string]any) map[string]any {
	return firstFlow(m)["captures"].([]any)[0].(map[string]any)
}

// mutations edit a copy of valid/final. Both validators must agree on each.
var mutations = []struct {
	name    string
	invalid bool
	apply   func(m map[string]any)
}{
	{"baseline", false, func(m map[string]any) {}},
	{"lang-too-long", true, func(m map[string]any) { m["lang"] = "english" }},
	{"lang-region-tag", false, func(m map[string]any) { m["lang"] = "es-MX" }},
	{"kind-unknown", true, func(m map[string]any) { m["kind"] = "draft" }},
	{"kind-number", true, func(m map[string]any) { m["kind"] = 3 }},
	{"title-empty", true, func(m map[string]any) { m["title"] = "" }},
	{"title-blank", true, func(m map[string]any) { m["title"] = "   " }},
	{"title-number", true, func(m map[string]any) { m["title"] = 7 }},
	{"version-bool", true, func(m map[string]any) { m["schema_version"] = true }},
	{"version-float-one", false, func(m map[string]any) { m["schema_version"] = 1.0 }},
	{"version-two", true, func(m map[string]any) { m["schema_version"] = 2 }},
	{"session-no-id", true, func(m map[string]any) { m["session"] = map[string]any{"host": "x"} }},
	{"session-extra", true, func(m map[string]any) { m["session"] = map[string]any{"id": "a", "pid": 1} }},
	{"session-no-host", false, func(m map[string]any) { m["session"] = map[string]any{"id": "a"} }},
	{"candidate-no-commit", true, func(m map[string]any) { m["candidate"] = map[string]any{"branch": "b"} }},
	{"candidate-no-base", false, func(m map[string]any) { m["candidate"] = map[string]any{"branch": "b", "commit": "c"} }},
	{"created-garbage", true, func(m map[string]any) { m["created_at"] = "yesterday" }},
	{"created-zulu", false, func(m map[string]any) { m["created_at"] = "2026-10-18T22:00:00Z" }},
	{"created-naive", false, func(m map[string]any) { m["created_at"] = "2026-10-18T22:00:00" }},
	{"created-date-only", false, func(m map[string]any) { m["created_at"] = "2026-10-18" }},
	{"created-fraction", false, func(m map[string]any) { m["created_at"] = "2026-10-18T22:00:00.123456-06:00" }},
	{"created-lowercase-z", false, func(m map[string]any) { m["created_at"] = "2026-10-18T22:00:00z" }},
	{"duration-overflows-float", true, func(m map[string]any) {
		firstFlow(m)["recording"] = map[string]any{
			"status": "recorded", "path": "images/clip.mp4", "duration_seconds": json.Number("1e999"), "candidate": "c"}
	}},
	{"duration-huge-integer", true, func(m map[string]any) {
		firstFlow(m)["recording"] = map[string]any{
			"status": "recorded", "path": "images/clip.mp4", "duration_seconds": json.Number("1" + strings.Repeat("0", 400)), "candidate": "c"}
	}},
	{"ticket-number", true, func(m map[string]any) { m["ticket"] = 5 }},
	{"result-empty", true, func(m map[string]any) { m["result"] = "" }},
	{"check-status-unknown", true, func(m map[string]any) {
		m["checks"] = []any{map[string]any{"name": "n", "status": "ok"}}
	}},
	{"check-status-number", true, func(m map[string]any) {
		m["checks"] = []any{map[string]any{"name": "n", "status": 1}}
	}},
	{"check-extra-field", true, func(m map[string]any) {
		m["checks"] = []any{map[string]any{"name": "n", "status": "pass", "x": 1}}
	}},
	{"checks-not-array", true, func(m map[string]any) { m["checks"] = "none" }},
	{"checks-empty", false, func(m map[string]any) { m["checks"] = []any{} }},
	{"action-required-not-list", true, func(m map[string]any) { m["action_required"] = "x" }},
	{"provenance-non-string", true, func(m map[string]any) { m["provenance"] = []any{1} }},
	{"environment-extra", true, func(m map[string]any) {
		m["environment"] = []any{map[string]any{"label": "a", "value": "b", "z": 1}}
	}},
	{"flows-empty", true, func(m map[string]any) { m["flows"] = []any{} }},
	{"flows-not-array", true, func(m map[string]any) { m["flows"] = map[string]any{} }},
	{"flow-extra", true, func(m map[string]any) { firstFlow(m)["x"] = 1 }},
	{"flow-id-uppercase", true, func(m map[string]any) { firstFlow(m)["id"] = "Upper" }},
	{"flow-id-65-chars", true, func(m map[string]any) { firstFlow(m)["id"] = strings.Repeat("a", 65) }},
	{"flow-steps-non-string", true, func(m map[string]any) { firstFlow(m)["steps"] = []any{1} }},
	{"flow-captures-empty", false, func(m map[string]any) { firstFlow(m)["captures"] = []any{} }},
	{"capture-viewport-mobile", false, func(m map[string]any) { firstCapture(m)["viewport"] = "mobile" }},
	{"capture-viewport-number", true, func(m map[string]any) { firstCapture(m)["viewport"] = 1 }},
	{"capture-alt-empty", true, func(m map[string]any) { firstCapture(m)["alt"] = "" }},
	{"capture-caption-empty", true, func(m map[string]any) { firstCapture(m)["caption"] = "" }},
	{"capture-extra", true, func(m map[string]any) { firstCapture(m)["x"] = 1 }},
	{"path-backslash-start", true, func(m map[string]any) { firstCapture(m)["path"] = "\\images\\nav-1.png" }},
	{"path-drive-letter", true, func(m map[string]any) { firstCapture(m)["path"] = "C:images/nav-1.png" }},
	{"path-empty-segment", true, func(m map[string]any) { firstCapture(m)["path"] = "images//nav-1.png" }},
	{"path-dot-segment", true, func(m map[string]any) { firstCapture(m)["path"] = "images/./nav-1.png" }},
	{"path-trailing-slash", true, func(m map[string]any) { firstCapture(m)["path"] = "images/nav-1.png/" }},
	{"path-space", true, func(m map[string]any) { firstCapture(m)["path"] = "images/nav 1.png" }},
	{"path-non-string", true, func(m map[string]any) { firstCapture(m)["path"] = 3 }},
	{"path-extension-only", true, func(m map[string]any) { firstCapture(m)["path"] = "images/.png" }},
	{"path-trailing-dot", true, func(m map[string]any) { firstCapture(m)["path"] = "images/nav-1.png." }},
	{"recording-status-unknown", true, func(m map[string]any) {
		firstFlow(m)["recording"] = map[string]any{"status": "paused"}
	}},
	{"recording-status-number", true, func(m map[string]any) {
		firstFlow(m)["recording"] = map[string]any{"status": 2}
	}},
	{"recording-none-with-path", true, func(m map[string]any) {
		firstFlow(m)["recording"] = map[string]any{"status": "none", "path": "images/nav-1.png"}
	}},
	{"recording-not-recorded-no-reason", true, func(m map[string]any) {
		firstFlow(m)["recording"] = map[string]any{"status": "not_recorded"}
	}},
	{"recording-not-recorded-reason", false, func(m map[string]any) {
		firstFlow(m)["recording"] = map[string]any{"status": "not_recorded", "reason": "no"}
	}},
	{"recording-recorded-no-duration", true, func(m map[string]any) {
		firstFlow(m)["recording"] = map[string]any{"status": "recorded", "path": "images/nav-1.mp4", "candidate": "c"}
	}},
	{"recording-duration-negative", true, func(m map[string]any) {
		firstFlow(m)["recording"] = map[string]any{
			"status": "recorded", "path": "images/nav-1.mp4", "duration_seconds": -1, "candidate": "c"}
	}},
	{"recording-duration-bool", true, func(m map[string]any) {
		firstFlow(m)["recording"] = map[string]any{
			"status": "recorded", "path": "images/nav-1.mp4", "duration_seconds": true, "candidate": "c"}
	}},
	{"recording-recorded-ok", false, func(m map[string]any) {
		firstFlow(m)["recording"] = map[string]any{
			"status": "recorded", "path": "images/clip.mp4", "duration_seconds": 12.5, "candidate": "c"}
	}},
	{"flows-over-50", true, func(m map[string]any) {
		var flows []any
		for i := 0; i < 51; i++ {
			flows = append(flows, map[string]any{"id": fmt.Sprintf("f%d", i), "title": "t", "captures": []any{}})
		}
		m["flows"] = flows
	}},
	{"captures-over-100", true, func(m map[string]any) {
		var caps []any
		for i := 0; i < 101; i++ {
			caps = append(caps, map[string]any{"path": "images/nav-1.png", "caption": "c"})
		}
		firstFlow(m)["captures"] = caps
	}},
}

func TestManifestMutationsAreClassifiedAsExpected(t *testing.T) {
	for _, mu := range mutations {
		t.Run(mu.name, func(t *testing.T) {
			folder := copyFixture(t, "valid/final")
			if err := os.WriteFile(filepath.Join(folder, "images", "clip.mp4"), []byte("mp4"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(folder, "images", "nav-1.mp4"), []byte("mp4"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(folder, "images", ".png"), []byte("png"), 0o644); err != nil {
				t.Fatal(err)
			}
			m := readManifestMap(t, folder)
			mu.apply(m)
			writeManifestMap(t, folder, m)
			_, err := ValidateManifest(folder)
			if mu.invalid && err == nil {
				t.Fatal("accepted, want rejected")
			}
			if !mu.invalid && err != nil {
				t.Fatalf("rejected, want accepted: %v", err)
			}
		})
	}
}

// TestManifestAgreesWithPythonValidator runs the reference validator
// (render_walkthrough.py) over every fixture and every mutation and requires
// the same accept/reject decision, so the two validators cannot drift.
func TestManifestAgreesWithPythonValidator(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not available")
	}
	script, err := filepath.Abs("../../content/skills/flow-walkthrough/scripts/render_walkthrough.py")
	if err != nil {
		t.Fatal(err)
	}
	pyAccepts := func(folder string) bool {
		return exec.Command(python, "-I", script, folder).Run() == nil
	}
	check := func(t *testing.T, folder string) {
		t.Helper()
		_, goErr := ValidateManifest(folder)
		if goAccepts, py := goErr == nil, pyAccepts(folder); goAccepts != py {
			t.Fatalf("Go accepts=%v (%v), Python accepts=%v", goAccepts, goErr, py)
		}
	}
	for _, kind := range []string{"valid", "invalid"} {
		for _, name := range subdirs(t, filepath.Join(fixturesRoot, kind)) {
			t.Run(kind+"/"+name, func(t *testing.T) { check(t, copyFixture(t, kind+"/"+name)) })
		}
	}
	for _, mu := range mutations {
		t.Run("mutation/"+mu.name, func(t *testing.T) {
			folder := copyFixture(t, "valid/final")
			for _, f := range []string{"clip.mp4", "nav-1.mp4", ".png"} {
				if err := os.WriteFile(filepath.Join(folder, "images", f), []byte("x"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			m := readManifestMap(t, folder)
			mu.apply(m)
			writeManifestMap(t, folder, m)
			check(t, folder)
		})
	}
}

func TestManifestRejectsUnreadableOrOversizedDocuments(t *testing.T) {
	cases := map[string]func(folder string){
		"missing": func(folder string) { os.Remove(filepath.Join(folder, "walkthrough.json")) },
		"not-json": func(folder string) {
			os.WriteFile(filepath.Join(folder, "walkthrough.json"), []byte("{nope"), 0o644)
		},
		"trailing-data": func(folder string) {
			data, _ := os.ReadFile(filepath.Join(folder, "walkthrough.json"))
			os.WriteFile(filepath.Join(folder, "walkthrough.json"), append(data, []byte(" {}")...), 0o644)
		},
		"invalid-utf8": func(folder string) {
			os.WriteFile(filepath.Join(folder, "walkthrough.json"), []byte("{\"title\":\"\xff\"}"), 0o644)
		},
		"array-root": func(folder string) {
			os.WriteFile(filepath.Join(folder, "walkthrough.json"), []byte("[]"), 0o644)
		},
		"over-1-mib": func(folder string) {
			os.WriteFile(filepath.Join(folder, "walkthrough.json"), []byte(strings.Repeat(" ", 1<<20+1)), 0o644)
		},
		"manifest-is-symlink": func(folder string) {
			real := filepath.Join(folder, "real.json")
			os.Rename(filepath.Join(folder, "walkthrough.json"), real)
			os.Symlink("real.json", filepath.Join(folder, "walkthrough.json"))
		},
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			folder := copyFixture(t, "valid/final")
			setup(folder)
			_, err := ValidateManifest(folder)
			var me *ManifestError
			if !errors.As(err, &me) || me.Field != "walkthrough.json" {
				t.Fatalf("want a walkthrough.json ManifestError, got %v", err)
			}
		})
	}
	if _, err := ValidateManifest(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("missing folder accepted")
	}
}

func TestManifestRejectsOversizedCaptureAndRecording(t *testing.T) {
	folder := copyFixture(t, "valid/final")
	big := filepath.Join(folder, "images", "nav-1.png")
	if err := os.Truncate(big, 20<<20+1); err != nil {
		t.Fatal(err)
	}
	_, err := ValidateManifest(folder)
	var me *ManifestError
	if !errors.As(err, &me) || me.Field != "flows[0].captures[0].path" {
		t.Fatalf("oversized capture: %v", err)
	}
	if err := os.Truncate(big, 20<<20); err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateManifest(folder); err != nil {
		t.Fatalf("capture at the limit rejected: %v", err)
	}
}

func TestManifestRejectsSymlinkedDirectoryInPath(t *testing.T) {
	folder := copyFixture(t, "valid/final")
	outside := t.TempDir()
	data, _ := os.ReadFile(filepath.Join(folder, "images", "nav-1.png"))
	os.WriteFile(filepath.Join(outside, "nav-1.png"), data, 0o644)
	m := readManifestMap(t, folder)
	firstCapture(m)["path"] = "linked/nav-1.png"
	writeManifestMap(t, folder, m)
	if err := os.Symlink(outside, filepath.Join(folder, "linked")); err != nil {
		t.Fatal(err)
	}
	_, err := ValidateManifest(folder)
	var me *ManifestError
	if !errors.As(err, &me) || me.Field != "flows[0].captures[0].path" {
		t.Fatalf("symlinked directory accepted: %v", err)
	}
}

// I1: the cases the Python validator once accepted must stay rejected here
// too (trailing newline in patterns, non-finite numbers, a linked manifest).
func TestManifestParityWithPythonStrictCases(t *testing.T) {
	t.Run("trailing newline in a flow id", func(t *testing.T) {
		folder := copyFixture(t, "valid/first-view")
		m := readManifestMap(t, folder)
		firstFlow(m)["id"] = "one\n"
		writeManifestMap(t, folder, m)
		if _, err := ValidateManifest(folder); err == nil {
			t.Fatal("accepted")
		}
	})
	t.Run("trailing newline in lang", func(t *testing.T) {
		folder := copyFixture(t, "valid/first-view")
		m := readManifestMap(t, folder)
		m["lang"] = "es\n"
		writeManifestMap(t, folder, m)
		if _, err := ValidateManifest(folder); err == nil {
			t.Fatal("accepted")
		}
	})
	t.Run("non-finite number", func(t *testing.T) {
		folder := copyFixture(t, "valid/first-view")
		raw := readFile(t, filepath.Join(folder, "walkthrough.json"))
		bad := strings.Replace(string(raw), `"schema_version": 1`, `"schema_version": NaN`, 1)
		if bad == string(raw) {
			t.Skip("fixture formatting differs")
		}
		os.WriteFile(filepath.Join(folder, "walkthrough.json"), []byte(bad), 0o644)
		if _, err := ValidateManifest(folder); err == nil {
			t.Fatal("accepted")
		}
	})
	t.Run("manifest that is a link", func(t *testing.T) {
		folder := copyFixture(t, "valid/first-view")
		real := filepath.Join(t.TempDir(), "elsewhere.json")
		os.WriteFile(real, readFile(t, filepath.Join(folder, "walkthrough.json")), 0o644)
		os.Remove(filepath.Join(folder, "walkthrough.json"))
		if err := os.Symlink(real, filepath.Join(folder, "walkthrough.json")); err != nil {
			t.Fatal(err)
		}
		if _, err := ValidateManifest(folder); err == nil {
			t.Fatal("accepted")
		}
	})
}

// Both readers execute the same data to detect URL-parser differences.
func TestManifestDeliveryContract(t *testing.T) {
	var cases []struct {
		Name     string
		Omit     bool
		Delivery any
		Field    string
	}
	if err := json.Unmarshal(readFile(t, filepath.Join(fixturesRoot, "delivery-contract.json")), &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			folder := copyFixture(t, "valid/first-view")
			doc := readManifestMap(t, folder)
			if !tc.Omit {
				doc["delivery"] = tc.Delivery
			}
			writeManifestMap(t, folder, doc)
			_, err := ValidateManifest(folder)
			if tc.Field == "" {
				if err != nil {
					t.Fatalf("valid delivery rejected: %v", err)
				}
				return
			}
			var me *ManifestError
			if !errors.As(err, &me) || me.Field != tc.Field {
				t.Fatalf("error = %v, want field %s", err, tc.Field)
			}
		})
	}
}
