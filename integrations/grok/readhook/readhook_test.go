package readhook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fixture struct {
	skills string // resolved skills dir
	ref    string // a valid reference inside it
	root   string
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	skills := filepath.Join(root, "skills")
	refDir := filepath.Join(skills, "flow-build", "references")
	if err := os.MkdirAll(refDir, 0o755); err != nil {
		t.Fatal(err)
	}
	ref := filepath.Join(refDir, "verification.md")
	writeLines(t, ref, 300)
	return fixture{skills: skills, ref: ref, root: root}
}

func writeLines(t *testing.T, path string, n int) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.Repeat("line\n", n)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func event(t *testing.T, fields map[string]any) []byte {
	t.Helper()
	b, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func readEvent(t *testing.T, input map[string]any, extra map[string]any) []byte {
	t.Helper()
	ev := map[string]any{"hookEventName": "pre_tool_use", "toolName": "read_file", "toolInput": input}
	for k, v := range extra {
		ev[k] = v
	}
	return event(t, ev)
}

func updatedInput(t *testing.T, out []byte) map[string]any {
	t.Helper()
	var r struct {
		HookSpecificOutput struct {
			HookEventName string         `json:"hookEventName"`
			UpdatedInput  map[string]any `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("output is not JSON: %q: %v", out, err)
	}
	if r.HookSpecificOutput.HookEventName != "PreToolUse" {
		t.Fatalf("hookEventName = %q", r.HookSpecificOutput.HookEventName)
	}
	return r.HookSpecificOutput.UpdatedInput
}

func TestReadWholeRewritesTargetFile(t *testing.T) {
	f := newFixture(t)
	out := ReadWhole(readEvent(t, map[string]any{"target_file": f.ref, "offset": 5, "limit": 40, "extra": "keep"}, nil), f.skills)
	in := updatedInput(t, out)
	if _, ok := in["offset"]; ok {
		t.Error("offset kept")
	}
	if _, ok := in["limit"]; ok {
		t.Error("limit kept")
	}
	if in["target_file"] != f.ref || in["extra"] != "keep" || len(in) != 2 {
		t.Errorf("updatedInput = %v", in)
	}
}

func TestReadWholeRewritesPathKey(t *testing.T) {
	f := newFixture(t)
	out := ReadWhole(readEvent(t, map[string]any{"path": f.ref, "limit": 40}, nil), f.skills)
	in := updatedInput(t, out)
	if in["path"] != f.ref || len(in) != 1 {
		t.Errorf("updatedInput = %v", in)
	}
}

func TestReadWholeOffsetOnly(t *testing.T) {
	f := newFixture(t)
	out := ReadWhole(readEvent(t, map[string]any{"target_file": f.ref, "offset": 3}, nil), f.skills)
	if len(updatedInput(t, out)) != 1 {
		t.Error("offset-only read not rewritten")
	}
}

func TestReadWholeRelativePathUsesWorkspaceRoot(t *testing.T) {
	f := newFixture(t)
	rel, _ := filepath.Rel(f.root, f.ref)
	other := t.TempDir()
	out := ReadWhole(readEvent(t, map[string]any{"target_file": rel, "limit": 10}, map[string]any{"workspaceRoot": f.root, "cwd": other}), f.skills)
	if len(out) == 0 {
		t.Fatal("relative path against workspaceRoot not rewritten")
	}
	if got := updatedInput(t, out)["target_file"]; got != rel {
		t.Errorf("path changed to %v; must be kept as sent", got)
	}
}

func TestReadWholeRelativePathFallsBackToCwd(t *testing.T) {
	f := newFixture(t)
	rel, _ := filepath.Rel(f.root, f.ref)
	out := ReadWhole(readEvent(t, map[string]any{"target_file": rel, "limit": 10}, map[string]any{"cwd": f.root}), f.skills)
	if len(out) == 0 {
		t.Fatal("relative path against cwd not rewritten")
	}
}

func TestReadWholeRelativePathWithoutBaseIsIgnored(t *testing.T) {
	f := newFixture(t)
	if out := ReadWhole(readEvent(t, map[string]any{"target_file": "flow-build/references/verification.md", "limit": 10}, nil), f.skills); len(out) != 0 {
		t.Errorf("output = %q", out)
	}
}

func TestReadWholePassesThrough(t *testing.T) {
	f := newFixture(t)
	outside := filepath.Join(f.root, "outside.md")
	writeLines(t, outside, 10)
	notMD := filepath.Join(f.skills, "script.sh")
	writeLines(t, notMD, 10)
	tooLong := filepath.Join(f.skills, "long.md")
	writeLines(t, tooLong, 1001)
	tooBig := filepath.Join(f.skills, "big.md")
	if err := os.WriteFile(tooBig, []byte(strings.Repeat("x", 64*1024+1)), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(f.skills, "escape.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	mdToSh := filepath.Join(f.skills, "tosh.md")
	if err := os.Symlink(notMD, mdToSh); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(f.skills, "dir.md")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	sibling := filepath.Join(f.root, "skills-evil")
	if err := os.MkdirAll(sibling, 0o755); err != nil {
		t.Fatal(err)
	}
	writeLines(t, filepath.Join(sibling, "x.md"), 5)

	cases := map[string]string{
		"outside the folder":         outside,
		"symlink leaving folder":     link,
		"not markdown":               notMD,
		"md symlink to non-md":       mdToSh,
		"more than 1000 lines":       tooLong,
		"more than 64 KiB":           tooBig,
		"nonexistent":                filepath.Join(f.skills, "missing.md"),
		"directory":                  dir,
		"dotdot escape":              filepath.Join(f.skills, "..", "outside.md"),
		"sibling with shared prefix": filepath.Join(sibling, "x.md"),
	}
	for name, p := range cases {
		t.Run(name, func(t *testing.T) {
			if out := ReadWhole(readEvent(t, map[string]any{"target_file": p, "limit": 40}, nil), f.skills); len(out) != 0 {
				t.Errorf("output = %q", out)
			}
		})
	}
}

func TestReadWholeAtLimitsIsRewritten(t *testing.T) {
	f := newFixture(t)
	p := filepath.Join(f.skills, "edge.md")
	writeLines(t, p, 1000)
	if out := ReadWhole(readEvent(t, map[string]any{"target_file": p, "limit": 40}, nil), f.skills); len(out) == 0 {
		t.Error("1000 lines should pass")
	}
	q := filepath.Join(f.skills, "edge64.md")
	if err := os.WriteFile(q, []byte(strings.Repeat("x", 64*1024)), 0o644); err != nil {
		t.Fatal(err)
	}
	if out := ReadWhole(readEvent(t, map[string]any{"target_file": q, "limit": 40}, nil), f.skills); len(out) == 0 {
		t.Error("64 KiB should pass")
	}
}

func TestReadWholeWithoutOffsetOrLimitIsIgnored(t *testing.T) {
	f := newFixture(t)
	if out := ReadWhole(readEvent(t, map[string]any{"target_file": f.ref}, nil), f.skills); len(out) != 0 {
		t.Errorf("output = %q", out)
	}
}

func TestReadWholeInvalidInputIsIgnored(t *testing.T) {
	f := newFixture(t)
	for name, in := range map[string][]byte{
		"invalid json":      []byte("{not json"),
		"empty":             nil,
		"toolInput array":   []byte(`{"toolInput":[1]}`),
		"no toolInput":      []byte(`{"toolName":"read_file"}`),
		"path not a string": []byte(`{"toolInput":{"target_file":7,"limit":1}}`),
		"truncated input":   readEvent(t, map[string]any{"target_file": f.ref, "limit": 1}, map[string]any{"toolInputTruncated": true}),
	} {
		t.Run(name, func(t *testing.T) {
			if out := ReadWhole(in, f.skills); len(out) != 0 {
				t.Errorf("output = %q", out)
			}
		})
	}
}

func TestReadWholeEmptySkillsDirIsIgnored(t *testing.T) {
	f := newFixture(t)
	for _, dir := range []string{"", "relative/skills", filepath.Join(f.root, "nope")} {
		if out := ReadWhole(readEvent(t, map[string]any{"target_file": f.ref, "limit": 1}, nil), dir); len(out) != 0 {
			t.Errorf("dir %q: output = %q", dir, out)
		}
	}
}
