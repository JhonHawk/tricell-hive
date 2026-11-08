package main

import (
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

const repoRoot = "../.."

func TestQualificationArmFlags(t *testing.T) {
	for _, c := range []struct {
		name, host, source, arm, model string
		wantErr                        string
	}{
		{"deployed claude", "claude", "", "deployed", "opus", ""},
		{"deployed codex", "codex", "", "deployed", "gpt-6", ""},
		{"deployed grok", "grok", "", "deployed", "grok-5", ""},
		{"deployed opencode", "opencode", "", "deployed", "opencode-go/m#high", ""},
		{"deployed pi", "pi", "", "deployed", "m", ""},
		{"deployed rejects a guidance source", "codex", "/src", "deployed", "gpt-6", "takes no --guidance-source"},
		{"A codex", "codex", "/src", "A", "gpt-6", ""},
		{"A grok", "grok", "/src", "A", "grok-5", ""},
		{"A opencode", "opencode", "/src", "A", "opencode-go/m#high", ""},
		{"A needs a source", "codex", "", "A", "gpt-6", "requires --guidance-source"},
		{"A claude rejected", "claude", "/src", "A", "opus", "only supported"},
		{"A pi rejected", "pi", "/src", "A", "m", "only supported"},
		{"none codex", "codex", "", "none", "gpt-6", ""},
		{"none opencode", "opencode", "", "none", "opencode-go/m#high", ""},
		{"none claude", "claude", "", "none", "opus", "--arm none is only supported for --host codex or opencode"},
		{"none grok", "grok", "", "none", "grok-5", "--arm none is only supported for --host codex or opencode"},
		{"none pi", "pi", "", "none", "m", "--arm none is only supported for --host codex or opencode"},
		{"none rejects a guidance source", "codex", "/src", "none", "gpt-6", "takes no --guidance-source"},
		{"B rejected", "codex", "/src", "B", "gpt-6", "must be deployed, A or none"},
		{"opencode none needs provider/model", "opencode", "", "none", "m", "provider"},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := validateQualificationArmFlags(c.host, c.source, c.arm, c.model)
			if c.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr)) {
				t.Fatalf("want error containing %q, got %v", c.wantErr, err)
			}
		})
	}
}

func TestQualificationFixtureResolvesCaseAndInlineTasks(t *testing.T) {
	cases, err := loadFlowCaseIndex(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	f, task, raw, err := loadQualifyFixture(repoRoot, "research-loads-skill")
	if err != nil {
		t.Fatal(err)
	}
	if f.ID != "research" || task.Case != "research" || f.Prompt != cases["research"].Prompt || len(raw) == 0 {
		t.Fatalf("a case task must keep the case ID and its prompt: %+v", f)
	}
	// The runner wraps a case task's prompt with flowTaskPrompt.
	if got := flowTaskPrompt(f.Prompt, "/root"); !strings.HasPrefix(got, f.Prompt) || got == f.Prompt {
		t.Fatal("flowTaskPrompt must extend the case prompt")
	}
	inline, _, _, err := loadQualifyFixture(repoRoot, "question-no-edit")
	if err != nil || inline.ID != "question-no-edit" || inline.Cwd != "." || len(inline.Files) == 0 {
		t.Fatalf("inline task: %+v %v", inline, err)
	}
	if _, _, _, err := loadQualifyFixture(repoRoot, "nope"); err == nil {
		t.Fatal("an unknown task must fail")
	}
}

func writeFixtureFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for p, c := range files {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestQualificationFixtureGitSetupDerivesFromTheCaseID(t *testing.T) {
	for _, c := range []struct {
		task       string
		wantCommit bool
		wantReport bool
	}{
		{"question-worktree-no-prep", true, false},
		{"git-delivery-loads-skill", true, true},
		{"question-no-edit", true, false},
		{"env-file-not-read", false, false},
		{"research-loads-skill", false, false},
	} {
		t.Run(c.task, func(t *testing.T) {
			f, task, _, err := loadQualifyFixture(repoRoot, c.task)
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			writeFixtureFiles(t, root, f.Files)
			if err := initGit(root); err != nil {
				t.Fatal(err)
			}
			report, err := setupQualifyGit(root, task, f)
			if err != nil {
				t.Fatal(err)
			}
			if (report != nil) != c.wantReport {
				t.Fatalf("delivery report = %v", report)
			}
			_, headErr := gitOutput(root, "rev-parse", "HEAD")
			if (headErr == nil) != c.wantCommit {
				t.Fatalf("commit present = %v, want %v", headErr == nil, c.wantCommit)
			}
			if c.wantCommit {
				if branch, _ := gitOutput(root, "rev-parse", "--abbrev-ref", "HEAD"); branch != "main" {
					t.Fatalf("branch %q", branch)
				}
			}
		})
	}
}

func TestQualificationFixtureRootIsOutsideSourceAndOut(t *testing.T) {
	src, out := "/work/checkout", "/work/checkout/_support/workspace/out"
	if err := checkQualifyFixtureRoot("/tmp/hive-pilot-x", src, out); err != nil {
		t.Fatal(err)
	}
	for _, root := range []string{"/work/checkout/tmp/hive-pilot-x", "/work/checkout/_support/workspace/out/hive-pilot-x"} {
		if err := checkQualifyFixtureRoot(root, src, out); err == nil {
			t.Fatalf("%s must be refused", root)
		}
	}
}

// qualifyUserHome writes the real-home files the isolated-home preparation
// reads: Codex's Engram config and auth, and OpenCode's config and catalog.
func qualifyUserHome(t *testing.T) string {
	t.Helper()
	home, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFiles(t, home, map[string]string{
		".codex/config.toml":             "[mcp_servers.engram]\ncommand = \"engram\"\nargs = [\"mcp\"]\n",
		".codex/auth.json":               `{"token":"fake"}`,
		".config/opencode/opencode.json": `{"mcp":{"servers":{"engram":{"command":["engram","mcp"],"type":"local"}}}}`,
		".cache/opencode/models.json":    `{}`,
	})
	for _, k := range []string{"XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME", "CODEX_HOME", "GROK_HOME"} {
		t.Setenv(k, "")
	}
	return home
}

func filesOf(t *testing.T, root string) (paths []string, contents map[string]string) {
	t.Helper()
	contents = map[string]string{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		paths = append(paths, rel)
		if d.Type()&os.ModeSymlink == 0 {
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			contents[rel] = string(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return paths, contents
}

func TestQualificationArmNoneHomeHasCredentialsAndMemoryButNoGuidance(t *testing.T) {
	logPath := fakeOpenCodeOnPath(t)
	for _, host := range []string{"codex", "opencode"} {
		t.Run(host, func(t *testing.T) {
			userHome := qualifyUserHome(t)
			out, err := target.Canonical(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			g, err := setupIsolatedHomeWithoutGuidance("none", out, host, userHome, "opencode-go/deepseek-v4.1-flash#high", filepath.Join(out, "engram-data"))
			if err != nil {
				t.Fatal(err)
			}
			if host == "opencode" {
				if err := g.importCredential(); err != nil {
					t.Fatal(err)
				}
			}
			shadow := filepath.Join(out, "shadow-home")
			paths, contents := filesOf(t, shadow)
			for _, p := range paths {
				if filepath.Base(p) == ".hive-native-owner" {
					continue // the runner's own cleanup marker, not guidance
				}
				low := strings.ToLower(p)
				if strings.Contains(low, "skills") || strings.Contains(low, "agents") || strings.Contains(low, "hive") || strings.Contains(low, "tricell") {
					t.Fatalf("the none home holds a guidance path: %s", p)
				}
				if strings.Contains(contents[p], management.Begin) || strings.Contains(contents[p], "TRICELL HIVE") {
					t.Fatalf("the none home holds the managed block marker in %s", p)
				}
			}
			if g.report.GuidanceBlockHash != "absent" || g.report.FlowBuildSkillHash != "absent" || g.report.ReleaseID != "" || g.report.Arm != "none" {
				t.Fatalf("report: %+v", g.report)
			}
			switch host {
			case "codex":
				if !strings.Contains(contents[".codex/config.toml"], "[mcp_servers.engram]") {
					t.Fatal("memory configuration missing")
				}
				if info, err := os.Lstat(filepath.Join(shadow, ".codex/auth.json")); err != nil || info.Mode()&os.ModeSymlink == 0 {
					t.Fatal("credential link missing")
				}
			case "opencode":
				if !strings.Contains(contents[".config/opencode/opencode.json"], filepath.Join(out, "engram-data")) {
					t.Fatal("memory configuration missing")
				}
				if _, ok := contents[".local/share/opencode/opencode.db"]; !ok {
					t.Fatal("credential import missing")
				}
				if calls, _ := os.ReadFile(logPath); !strings.Contains(string(calls), "import-home="+shadow) {
					t.Fatalf("credential import did not run in the isolated home: %s", calls)
				}
			}
			g.cleanup()
		})
	}
}

func TestQualificationArmAHomeCarriesTheMarkerTheNoneHomeLacks(t *testing.T) {
	userHome := qualifyUserHome(t)
	out, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	g, err := setupGuidanceVariant(newFixtureCheckout(t, "# Rules\n"), "A", out, "codex", userHome, "gpt-6", filepath.Join(out, "engram-data"))
	if err != nil {
		t.Fatal(err)
	}
	defer g.cleanup()
	_, contents := filesOf(t, filepath.Join(out, "shadow-home"))
	found := false
	for _, c := range contents {
		found = found || strings.Contains(c, management.Begin)
	}
	if !found {
		t.Fatal("the marker scan would not detect installed guidance")
	}
}

func installFor(t *testing.T, source, home, host string) {
	t.Helper()
	if err := os.MkdirAll(home, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := installGuidanceVariant(source, home, host); err != nil {
		t.Fatal(err)
	}
}

func mustHash(t *testing.T, host, home string) string {
	t.Helper()
	h, err := guidanceHash(host, home)
	if err != nil || len(h) != 64 {
		t.Fatalf("guidanceHash(%s, %q) = %q, %v", host, home, h, err)
	}
	return h
}

func TestGuidanceHashIsStableAcrossHomesAndTracksManagedContent(t *testing.T) {
	for _, k := range []string{"CODEX_HOME", "GROK_HOME", "CLAUDE_CONFIG_DIR"} {
		t.Setenv(k, "")
	}
	source := newFixtureCheckout(t, "# Rules\nBody.\n")
	homeA, _ := target.Canonical(t.TempDir())
	homeB, _ := target.Canonical(t.TempDir())
	installFor(t, source, homeA, "codex")
	installFor(t, source, homeB, "codex")
	base := mustHash(t, "codex", homeA)
	if again := mustHash(t, "codex", homeA); again != base {
		t.Fatal("same content hashed differently")
	}
	if other := mustHash(t, "codex", homeB); other != base {
		t.Fatalf("two installations of the same checkout differ: %s %s", base, other)
	}
	// The deployed reading (no explicit home) of the same installation agrees.
	t.Setenv("HOME", homeA)
	if deployed := mustHash(t, "codex", ""); deployed != base {
		t.Fatalf("deployed and isolated readings differ: %s %s", deployed, base)
	}
	t.Setenv("HOME", os.Getenv("HOME"))

	block, _, err := guidanceReadPaths("codex", homeB)
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(block)
	if err != nil {
		t.Fatal(err)
	}
	// Text outside the markers does not change the hash.
	if err := os.WriteFile(block, append([]byte("My own notes.\n"), append(original, []byte("\nMore notes.\n")...)...), 0600); err != nil {
		t.Fatal(err)
	}
	if got := mustHash(t, "codex", homeB); got != base {
		t.Fatal("text outside the managed block changed the hash")
	}
	// An edit inside the block does.
	if err := os.WriteFile(block, []byte(strings.Replace(string(original), "Body.", "Edited.", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	if got := mustHash(t, "codex", homeB); got == base {
		t.Fatal("editing the managed block did not change the hash")
	}
	if err := os.WriteFile(block, original, 0600); err != nil {
		t.Fatal(err)
	}
	if got := mustHash(t, "codex", homeB); got != base {
		t.Fatal("restoring the block did not restore the hash")
	}
}

func TestGuidanceHashTracksSymlinkTargetsAndSkillFiles(t *testing.T) {
	t.Setenv("CODEX_HOME", "")
	source := newFixtureCheckout(t, "# Rules\n")
	home, _ := target.Canonical(t.TempDir())
	installFor(t, source, home, "codex")
	base := mustHash(t, "codex", home)
	entries, err := management.Status(management.Options{Scope: "user", Home: home, Hosts: []string{"codex"}})
	if err != nil {
		t.Fatal(err)
	}
	var link, skill string
	for _, en := range entries {
		if en.Kind == "symlink" && link == "" && en.Status == "installed" {
			link = en.Path
		}
		if en.Kind == "skill" && skill == "" && en.Status == "installed" {
			skill = en.Path
		}
	}
	if link == "" && skill == "" {
		t.Fatalf("the installation has neither a symlink nor a skill record: %+v", entries)
	}
	if link != "" {
		dest, err := os.Readlink(link)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(link); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(dest+"-moved", link); err != nil {
			t.Fatal(err)
		}
		if got := mustHash(t, "codex", home); got == base {
			t.Fatal("changing a symlink destination did not change the hash")
		}
		if err := os.Remove(link); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(dest, link); err != nil {
			t.Fatal(err)
		}
		if got := mustHash(t, "codex", home); got != base {
			t.Fatal("restoring the symlink did not restore the hash")
		}
	}
	if skill != "" {
		data, err := os.ReadFile(skill)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(skill, append(data, []byte("\nedited\n")...), 0600); err != nil {
			t.Fatal(err)
		}
		if got := mustHash(t, "codex", home); got == base {
			t.Fatal("editing a skill did not change the hash")
		}
	}
}

func TestGuidanceHashIgnoresRecordsOfOtherHostsAndFailsWithoutInstallation(t *testing.T) {
	for _, k := range []string{"CODEX_HOME", "GROK_HOME"} {
		t.Setenv(k, "")
	}
	source := newFixtureCheckout(t, "# Rules\n")
	home, _ := target.Canonical(t.TempDir())
	installFor(t, source, home, "codex")
	base := mustHash(t, "codex", home)
	// Another host installed in the same home, then edited: codex is unchanged.
	installFor(t, source, home, "opencode")
	if got := mustHash(t, "codex", home); got != base {
		t.Fatal("installing another host changed the codex hash")
	}
	block, _, err := guidanceReadPaths("opencode", home)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(block)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(block, []byte(strings.Replace(string(data), "Rules", "Changed", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	if got := mustHash(t, "codex", home); got != base {
		t.Fatal("editing another host's record changed the codex hash")
	}
	if got := mustHash(t, "opencode", home); got == "" {
		t.Fatal("opencode hash missing")
	}
	if _, err := guidanceHash("codex", t.TempDir()); err == nil {
		t.Fatal("a home with no Hive installation must not hash to a value")
	}
}

// ---- the runner, end to end, with fake hosts -------------------------------

const fakeHostScript = `#!/bin/sh
echo "$(basename "$0") $*" >> "$FAKE_HOST_LOG"
case "$1" in
  --version) echo "fake 1.0"; exit 0 ;;
  auth)
    if [ "$2" = "export" ]; then printf '{"marker":"%s"}' "$3"; exit 0; fi
    if [ "$2" = "import" ]; then mkdir -p "$HOME/.local/share/opencode"; cat > "$HOME/.local/share/opencode/opencode.db"; exit 0; fi ;;
  session) exit 1 ;;
esac
echo '{"type":"noop"}'
exit 0
`

const fakeEngramScript = `#!/usr/bin/env python3
import http.server,sys
a=sys.argv[1:]
if a==['projects','list']: print('No projects found.');sys.exit(0)
if a==['instance-id']: print('0123456789abcdef0123456789abcdef');sys.exit(0)
if a[0]=='serve':
 class H(http.server.BaseHTTPRequestHandler):
  def do_GET(self):
   self.send_response(200);self.end_headers();self.wfile.write(b'{"instance_id":"0123456789abcdef0123456789abcdef"}')
  def log_message(self,*_): pass
 http.server.HTTPServer(('127.0.0.1',int(a[1])),H).serve_forever()
if a[0]=='stats': print('Sessions: 0\nObservations: 0\nPrompts: 0');sys.exit(0)
sys.exit(1)
`

// qualifyRun is one simulated runner invocation environment.
type qualifyRun struct {
	t      *testing.T
	home   string
	bin    string
	source string
	out    string
	log    string
}

func newQualifyRun(t *testing.T) *qualifyRun {
	t.Helper()
	q := &qualifyRun{t: t, home: qualifyUserHome(t), bin: t.TempDir()}
	q.source = newFixtureCheckout(t, "# Rules\nBody.\n")
	for _, f := range []string{"tests/fixtures/qualification/suite.json", "tests/fixtures/flows/cases.json"} {
		data, err := os.ReadFile(filepath.Join(repoRoot, f))
		if err != nil {
			t.Fatal(err)
		}
		writeFixtureFiles(t, q.source, map[string]string{f: string(data)})
	}
	if err := os.MkdirAll(filepath.Join(q.source, "_support/workspace"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, host := range []string{"codex", "claude", "grok", "pi", "opencode"} {
		writeFixtureFiles(t, q.bin, map[string]string{host: fakeHostScript})
		if err := os.Chmod(filepath.Join(q.bin, host), 0700); err != nil {
			t.Fatal(err)
		}
	}
	writeFixtureFiles(t, q.bin, map[string]string{"engram": fakeEngramScript})
	if err := os.Chmod(filepath.Join(q.bin, "engram"), 0700); err != nil {
		t.Fatal(err)
	}
	q.log = filepath.Join(t.TempDir(), "host.log")
	return q
}

// run launches the test binary as the pilot runner and returns its output,
// exit error and the run directory.
func (q *qualifyRun) run(name string, tmpdir string, args ...string) (string, error, string) {
	q.t.Helper()
	out := filepath.Join(q.source, "_support/workspace", name)
	full := append([]string{"--suite", qualificationSuite, "--source", q.source, "--configured-model", "configured", "--out", out, "--timeout", "1m"}, args...)
	cmd := exec.Command(os.Args[0], "-test.run=^TestQualificationRunnerSubprocess$")
	env := append(os.Environ(), "QUALIFY_RUNNER_SUBPROCESS=1", "QUALIFY_RUNNER_ARGS="+strings.Join(full, "\n"),
		"HOME="+q.home, "PATH="+q.bin+string(os.PathListSeparator)+os.Getenv("PATH"), "FAKE_HOST_LOG="+q.log,
		"ENGRAM_DATA_DIR="+filepath.Join(q.home, "everyday"), "CODEX_HOME=", "CLAUDE_CONFIG_DIR=", "GROK_HOME=", "XDG_CONFIG_HOME=", "XDG_CACHE_HOME=", "XDG_DATA_HOME=")
	if tmpdir != "" {
		env = append(env, "TMPDIR="+tmpdir)
	}
	cmd.Env = env
	b, err := cmd.CombinedOutput()
	return string(b), err, out
}

func (q *qualifyRun) result(out string) result {
	q.t.Helper()
	data, err := os.ReadFile(filepath.Join(out, "run.json"))
	if err != nil {
		q.t.Fatal(err)
	}
	var r result
	if err := json.Unmarshal(data, &r); err != nil {
		q.t.Fatal(err)
	}
	return r
}

func (q *qualifyRun) launches() string {
	b, _ := os.ReadFile(q.log)
	return string(b)
}

func TestQualificationRunnerSubprocess(t *testing.T) {
	if os.Getenv("QUALIFY_RUNNER_SUBPROCESS") != "1" {
		t.Skip("entry point of the re-executed runner")
	}
	flag.CommandLine = flag.NewFlagSet("pilot", flag.ExitOnError)
	os.Args = append([]string{"pilot"}, strings.Split(os.Getenv("QUALIFY_RUNNER_ARGS"), "\n")...)
	main()
}

func TestQualificationArmRunnerDeployedRecordsTaskSplitArmEffortAndHash(t *testing.T) {
	q := newQualifyRun(t)
	installFor(t, q.source, q.home, "claude")
	outText, err, out := q.run("deployed", "", "--host", "claude", "--task", "research-loads-skill", "--arm", "deployed", "--model", "opus", "--effort", "high")
	if err != nil {
		t.Fatalf("%v\n%s", err, outText)
	}
	r := q.result(out)
	want := mustHash(t, "claude", q.home)
	if r.Suite != qualificationSuite || r.Task != "research-loads-skill" || r.Split != "tuning" || r.Arm != "deployed" || r.Effort != "high" || r.GuidanceHash != want || r.GuidanceVariant != nil {
		t.Fatalf("run.json: task=%q split=%q arm=%q effort=%q hash=%q (want %q) variant=%v", r.Task, r.Split, r.Arm, r.Effort, r.GuidanceHash, want, r.GuidanceVariant)
	}
	if r.Case != "research" {
		t.Fatalf("a case task must keep the case ID: %q", r.Case)
	}
	if !strings.Contains(q.launches(), "claude --version") {
		t.Fatalf("host was not launched: %q", q.launches())
	}
	// The fixture root is outside --source and --out.
	if under(r.Root, q.source) || under(r.Root, out) {
		t.Fatalf("fixture root %s is under the source or the output", r.Root)
	}
	// A case task receives the flows prompt wrapper (the research case).
	fixtureJSON, err := os.ReadFile(filepath.Join(out, "fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(fixtureJSON), "Límite de esta tarea") {
		t.Fatal("the case:research task did not receive flowTaskPrompt")
	}
}

func TestQualificationArmRunnerDeployedWithoutInstallationRefuses(t *testing.T) {
	q := newQualifyRun(t)
	outText, err, _ := q.run("deployed-empty", "", "--host", "claude", "--task", "question-no-edit", "--arm", "deployed", "--model", "opus")
	if err == nil || !strings.Contains(outText, "no installed Hive guidance") {
		t.Fatalf("a deployed run without an installation must not record a hash: %v\n%s", err, outText)
	}
	if strings.Contains(q.launches(), "claude -p") {
		t.Fatal("the host was launched")
	}
}

func TestQualificationArmRunnerNoneRefusedWithoutLaunchingOnClaudeGrokPi(t *testing.T) {
	for _, host := range []string{"claude", "grok", "pi"} {
		t.Run(host, func(t *testing.T) {
			q := newQualifyRun(t)
			outText, err, out := q.run("none-"+host, "", "--host", host, "--task", "question-no-edit", "--arm", "none", "--model", "m")
			if err == nil || !strings.Contains(outText, "--arm none is only supported for --host codex or opencode (model-floor.md «Known limits»)") {
				t.Fatalf("expected the limitation error: %v\n%s", err, outText)
			}
			if q.launches() != "" {
				t.Fatalf("the host or any CLI was invoked: %q", q.launches())
			}
			if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
				t.Fatal("a refused run left its output behind")
			}
		})
	}
}

func TestQualificationArmRunnerNoneAndAOnCodexRecordHashes(t *testing.T) {
	q := newQualifyRun(t)
	outText, err, out := q.run("none-codex", "", "--host", "codex", "--task", "question-no-edit", "--arm", "none", "--model", "gpt-6", "--effort", "high")
	if err != nil {
		t.Fatalf("%v\n%s", err, outText)
	}
	r := q.result(out)
	if r.GuidanceHash != "none" || r.Arm != "none" || r.GuidanceVariant == nil || r.GuidanceVariant.Arm != "none" || r.Task != "question-no-edit" || r.Split != "heldback" {
		t.Fatalf("none run.json: %+v", r)
	}
	for _, p := range []string{"AGENTS.md", ".agents"} {
		if _, err := os.Stat(filepath.Join(out, "shadow-home", ".codex", p)); err == nil {
			t.Fatalf("none home has %s", p)
		}
	}

	outText, err, outA := q.run("a-codex", "", "--host", "codex", "--task", "question-no-edit", "--arm", "A", "--guidance-source", q.source, "--model", "gpt-6", "--effort", "high")
	if err != nil {
		t.Fatalf("%v\n%s", err, outText)
	}
	a := q.result(outA)
	if len(a.GuidanceHash) != 64 || a.Arm != "A" || a.GuidanceVariant == nil || a.GuidanceVariant.ReleaseID == "" {
		t.Fatalf("A run.json: hash=%q", a.GuidanceHash)
	}
	// The hash is the one function over the isolated home, independent of the
	// per-attempt home path.
	if again := mustHash(t, "codex", filepath.Join(outA, "shadow-home")); again != a.GuidanceHash {
		t.Fatalf("recorded hash differs from the function's: %s %s", a.GuidanceHash, again)
	}
	outText, err, outA2 := q.run("a-codex-2", "", "--host", "codex", "--task", "question-no-edit", "--arm", "A", "--guidance-source", q.source, "--model", "gpt-6", "--effort", "high")
	if err != nil {
		t.Fatalf("%v\n%s", err, outText)
	}
	if a2 := q.result(outA2); a2.GuidanceHash != a.GuidanceHash {
		t.Fatalf("two attempts on the same checkout hash differently: %s %s", a.GuidanceHash, a2.GuidanceHash)
	}
}

func TestQualificationArmRunnerNoneOpenCodeRejectsWorkingDirectoryUnderRealHome(t *testing.T) {
	q := newQualifyRun(t)
	under := filepath.Join(q.home, "tmp")
	if err := os.MkdirAll(under, 0700); err != nil {
		t.Fatal(err)
	}
	outText, err, out := q.run("none-opencode-home", under, "--host", "opencode", "--task", "question-no-edit", "--arm", "none", "--model", "opencode-go/deepseek-v4.1-flash#high")
	if err == nil || !strings.Contains(outText, "under the real home") {
		t.Fatalf("expected the home check to refuse: %v\n%s", err, outText)
	}
	if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
		t.Fatal("a refused run left its output behind")
	}

	outText, err, out = q.run("none-opencode", "", "--host", "opencode", "--task", "question-no-edit", "--arm", "none", "--model", "opencode-go/deepseek-v4.1-flash#high")
	if err != nil {
		t.Fatalf("%v\n%s", err, outText)
	}
	if r := q.result(out); r.GuidanceHash != "none" || r.Arm != "none" {
		t.Fatalf("%+v", r)
	}
}

func TestQualificationArmRunnerRefusesAFixtureRootUnderSource(t *testing.T) {
	q := newQualifyRun(t)
	installFor(t, q.source, q.home, "claude")
	inside := filepath.Join(q.source, "tmp")
	if err := os.MkdirAll(inside, 0700); err != nil {
		t.Fatal(err)
	}
	outText, err, _ := q.run("root-under-source", inside, "--host", "claude", "--task", "question-no-edit", "--arm", "deployed", "--model", "opus")
	if err == nil || !strings.Contains(outText, "fixture root") {
		t.Fatalf("expected the root separation check: %v\n%s", err, outText)
	}
	if strings.Contains(q.launches(), "claude -p") {
		t.Fatal("the host was launched")
	}
}

func TestQualificationArmRunnerRefusesOutputOutsideWorkspaceAndCaseFlag(t *testing.T) {
	q := newQualifyRun(t)
	installFor(t, q.source, q.home, "claude")
	cmd := exec.Command(os.Args[0], "-test.run=^TestQualificationRunnerSubprocess$")
	args := []string{"--suite", qualificationSuite, "--source", q.source, "--configured-model", "c", "--out", filepath.Join(t.TempDir(), "out"), "--host", "claude", "--task", "question-no-edit", "--arm", "deployed", "--model", "opus"}
	cmd.Env = append(os.Environ(), "QUALIFY_RUNNER_SUBPROCESS=1", "QUALIFY_RUNNER_ARGS="+strings.Join(args, "\n"), "HOME="+q.home, "PATH="+q.bin+string(os.PathListSeparator)+os.Getenv("PATH"), "FAKE_HOST_LOG="+q.log)
	if b, err := cmd.CombinedOutput(); err == nil || !strings.Contains(string(b), "must be below repository _support/workspace") {
		t.Fatalf("expected the output rule: %v\n%s", err, b)
	}
	outText, err, _ := q.run("with-case", "", "--host", "claude", "--task", "question-no-edit", "--case", "plan", "--arm", "deployed", "--model", "opus")
	if err == nil || !strings.Contains(outText, "no --case") {
		t.Fatalf("expected --case to be refused: %v\n%s", err, outText)
	}
}

// fakeOpenCodeScript emulates the OpenCode CLI: a JSONL run and a session
// export whose body, delay and visibility of the isolated data directory the
// test controls through the environment.
const fakeOpenCodeScript = `#!/bin/sh
echo "$(basename "$0") $*" >> "$FAKE_HOST_LOG"
case "$1" in
  --version) echo "fake 1.0"; exit 0 ;;
  auth)
    if [ "$2" = "export" ]; then printf '{"marker":"%s"}' "$3"; exit 0; fi
    if [ "$2" = "import" ]; then mkdir -p "$HOME/.local/share/opencode"; cat > "$HOME/.local/share/opencode/opencode.db"; exit 0; fi ;;
  session)
    if [ "$2" = "export" ]; then
      if [ -f "$HOME/.local/share/opencode/opencode.db" ]; then echo "EXPORT_DATA_DIR_PRESENT" >> "$FAKE_HOST_LOG"; fi
      if [ -n "$FAKE_EXPORT_SLEEP" ]; then sleep "$FAKE_EXPORT_SLEEP"; fi
      if [ -z "$FAKE_EXPORT_JSON" ]; then exit 1; fi
      printf '%s' "$FAKE_EXPORT_JSON"; exit 0
    fi
    exit 1 ;;
esac
if [ -n "$FAKE_RUN_STREAM_STOP" ]; then
  echo '{"type":"text","sessionID":"ses-1","part":{"type":"text","messageID":"msg-1","text":"done"}}'
  echo '{"type":"step_finish","sessionID":"ses-1","part":{"type":"step-finish","messageID":"msg-1","reason":"stop"}}'
else
  echo '{"type":"text","sessionID":"ses-1","part":{"type":"text","messageID":"msg-1","text":"done"}}'
fi
exit 0
`

func (q *qualifyRun) useFakeOpenCode() {
	q.t.Helper()
	writeFixtureFiles(q.t, q.bin, map[string]string{"opencode": fakeOpenCodeScript})
	if err := os.Chmod(filepath.Join(q.bin, "opencode"), 0700); err != nil {
		q.t.Fatal(err)
	}
}

func (q *qualifyRun) exports() int {
	return strings.Count(q.launches(), "opencode session export")
}

const servedModelDeepseek = `{"id":"deepseek-v4.1-flash","providerID":"opencode-go","variant":"high"}`

func exportWithModels(infoModel, assistantModel string) string {
	msg := `{"id":"msg-1","type":"assistant","finish":"stop"`
	if assistantModel != "" {
		msg += `,"model":` + assistantModel
	}
	msg += `}`
	return `{"info":{"id":"ses-1","model":` + infoModel + `},"messages":[` + msg + `]}`
}

func runOpenCodeQualification(t *testing.T, q *qualifyRun, name string, env ...string) result {
	t.Helper()
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		t.Setenv(k, v)
	}
	outText, err, out := q.run(name, "", "--host", "opencode", "--task", "question-no-edit", "--arm", "none", "--model", "opencode-go/deepseek-v4.1-flash#high")
	if err != nil {
		t.Fatalf("%v\n%s", err, outText)
	}
	return q.result(out)
}

func TestQualificationServedModelComesFromTheExportOnce(t *testing.T) {
	for name, stream := range map[string]string{"stream-terminal": "1", "no-terminal-reuses-verify-export": ""} {
		t.Run(name, func(t *testing.T) {
			q := newQualifyRun(t)
			q.useFakeOpenCode()
			r := runOpenCodeQualification(t, q, "served", "FAKE_EXPORT_JSON="+exportWithModels(`{}`, servedModelDeepseek), "FAKE_RUN_STREAM_STOP="+stream)
			if r.ModelObserved != "opencode-go/deepseek-v4.1-flash#high" {
				t.Fatalf("ModelObserved = %q", r.ModelObserved)
			}
			if n := q.exports(); n != 1 {
				t.Fatalf("export launched %d times", n)
			}
		})
	}
}

func TestQualificationServedModelPrefersTheLastAssistantMessage(t *testing.T) {
	q := newQualifyRun(t)
	q.useFakeOpenCode()
	info := `{"id":"other-model","providerID":"other","variant":"low"}`
	r := runOpenCodeQualification(t, q, "assistant-wins", "FAKE_EXPORT_JSON="+exportWithModels(info, servedModelDeepseek), "FAKE_RUN_STREAM_STOP=1")
	if r.ModelObserved != "opencode-go/deepseek-v4.1-flash#high" {
		t.Fatalf("ModelObserved = %q", r.ModelObserved)
	}
}

func TestQualificationServedModelRunsBeforeTheIsolatedDataDirIsRemoved(t *testing.T) {
	q := newQualifyRun(t)
	q.useFakeOpenCode()
	runOpenCodeQualification(t, q, "before-cleanup", "FAKE_EXPORT_JSON="+exportWithModels(`{}`, servedModelDeepseek), "FAKE_RUN_STREAM_STOP=1")
	if !strings.Contains(q.launches(), "EXPORT_DATA_DIR_PRESENT") {
		t.Fatalf("the export did not see the isolated data directory:\n%s", q.launches())
	}
}

func TestQualificationServedModelIsNotObservedWithoutAModelInTheExport(t *testing.T) {
	for name, body := range map[string]string{
		"no model":      exportWithModels(`{}`, ""),
		"export fails":  "",
		"invalid JSON":  `{nope`,
		"empty message": `{"info":{"id":"ses-1"},"messages":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			q := newQualifyRun(t)
			q.useFakeOpenCode()
			r := runOpenCodeQualification(t, q, "not-observed", "FAKE_EXPORT_JSON="+body, "FAKE_RUN_STREAM_STOP=1")
			if r.ModelObserved != "not_observed" {
				t.Fatalf("ModelObserved = %q", r.ModelObserved)
			}
		})
	}
}

func TestQualificationServedModelExportUsesItsOwnTimeoutNotTheRunDeadline(t *testing.T) {
	q := newQualifyRun(t)
	q.useFakeOpenCode()
	t.Setenv("FAKE_EXPORT_JSON", exportWithModels(`{}`, servedModelDeepseek))
	t.Setenv("FAKE_RUN_STREAM_STOP", "1")
	t.Setenv("FAKE_EXPORT_SLEEP", "2")
	// The run limit is 3s and the export takes 2s: a run-deadline export would
	// have no time left, the dedicated 10s timeout does.
	outText, err, out := q.run("own-timeout", "", "--host", "opencode", "--task", "question-no-edit", "--arm", "none", "--model", "opencode-go/deepseek-v4.1-flash#high", "--timeout", "3s")
	if err != nil {
		t.Fatalf("%v\n%s", err, outText)
	}
	if r := q.result(out); r.ModelObserved != "opencode-go/deepseek-v4.1-flash#high" {
		t.Fatalf("ModelObserved = %q", r.ModelObserved)
	}
}

func TestQualificationServedModelCodexStaysNotObserved(t *testing.T) {
	q := newQualifyRun(t)
	outText, err, out := q.run("codex-served", "", "--host", "codex", "--task", "question-no-edit", "--arm", "none", "--model", "gpt-6", "--effort", "high")
	if err != nil {
		t.Fatalf("%v\n%s", err, outText)
	}
	r := q.result(out)
	if r.ModelObserved != "not_observed" || r.ModelRequested != "gpt-6" {
		t.Fatalf("observed=%q requested=%q", r.ModelObserved, r.ModelRequested)
	}
}

func TestQualificationFixtureRootResolvesSymlinksOfSourceAndOut(t *testing.T) {
	real := t.TempDir()
	realRoot, _ := filepath.EvalSymlinks(real)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(realRoot, link); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(realRoot, "fixture")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := checkQualifyFixtureRoot(root, link, ""); err == nil {
		t.Fatal("a root under a symlinked --source was accepted")
	}
	if err := checkQualifyFixtureRoot(root, "", link); err == nil {
		t.Fatal("a root under a symlinked --out was accepted")
	}
}

func TestGuidanceHashSurvivesAMissingManagedFile(t *testing.T) {
	t.Setenv("CODEX_HOME", "")
	source := newFixtureCheckout(t, "# Rules\n")
	home, _ := target.Canonical(t.TempDir())
	installFor(t, source, home, "codex")
	base := mustHash(t, "codex", home)
	entries, err := management.Status(management.Options{Scope: "user", Home: home, Hosts: []string{"codex"}})
	if err != nil {
		t.Fatal(err)
	}
	removed := false
	for _, en := range entries {
		if en.Host == "codex" && (en.Kind == "skill" || en.Kind == "symlink") && en.Status == "installed" {
			if err := os.RemoveAll(en.Path); err != nil {
				t.Fatal(err)
			}
			removed = true
			break
		}
	}
	if !removed {
		t.Fatal("no managed file to remove")
	}
	if got := mustHash(t, "codex", home); got == base {
		t.Fatal("a missing managed file left the hash unchanged")
	}
}
