package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"tricell-hive/tooling/management"
)

// Tests of the CLI model catalog (#46, T7). The recorded outputs live in
// testdata/catalog; every executable here is a fake in a temporary PATH.

func readCatalogSample(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "catalog", name+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// recordingRunner returns each queued output in turn and records the calls.
type recordingRunner struct {
	outputs [][]byte
	calls   [][]string
}

func (r *recordingRunner) run(_ context.Context, bin string, args []string) ([]byte, error) {
	r.calls = append(r.calls, append([]string{bin}, args...))
	i := len(r.calls) - 1
	if i >= len(r.outputs) {
		i = len(r.outputs) - 1
	}
	return r.outputs[i], nil
}

func catalogIDs(models []catalogModel) []string {
	ids := make([]string, 0, len(models))
	for _, m := range models {
		ids = append(ids, m.ID)
	}
	return ids
}

func TestCatalogCodexSkipsHiddenModelsAndReadsEfforts(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{readCatalogSample(t, "codex")}}
	got, err := listHostModels(context.Background(), "codex", r.run)
	if err != nil {
		t.Fatal(err)
	}
	want := []catalogModel{
		{ID: "gpt-6.1-sol", Name: "GPT-6.1 Sol", Efforts: []string{"low", "medium", "high", "xhigh", "max", "ultra"}},
		{ID: "gpt-6-luna", Name: "GPT-6 Luna", Efforts: []string{"low", "medium", "high", "xhigh", "max"}},
		{ID: "gpt-5.5", Name: "GPT-5.5", Efforts: []string{"low", "medium", "high", "xhigh"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("models = %+v, want %+v", got, want)
	}
	if !reflect.DeepEqual(r.calls, [][]string{{"codex", "debug", "models"}}) {
		t.Fatalf("calls = %v", r.calls)
	}
}

func TestCatalogCodexDropsEffortsThatOverridesRefuse(t *testing.T) {
	out := `{"models":[{"slug":"m1","visibility":"list","supported_reasoning_levels":[{"effort":"low"},{"effort":"turbo"},{"effort":"high"}]}]}`
	r := &recordingRunner{outputs: [][]byte{[]byte(out)}}
	got, err := listHostModels(context.Background(), "codex", r.run)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"low", "high"}; !reflect.DeepEqual(got[0].Efforts, want) {
		t.Fatalf("efforts = %v, want %v", got[0].Efforts, want)
	}
}

func TestCatalogCodexUnreadableOutputIsAnError(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{[]byte("not json")}}
	if _, err := listHostModels(context.Background(), "codex", r.run); err == nil {
		t.Fatal("expected an error")
	}
}

func TestCatalogOpenCodeReadsOneModelPerLine(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{readCatalogSample(t, "opencode")}}
	got, err := listHostModels(context.Background(), "opencode", r.run)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"opencode/big-pickle", "openai/gpt-5.5", "github-copilot/claude-opus-4.7", "anthropic/claude-sonnet-4.5"}
	if !reflect.DeepEqual(catalogIDs(got), want) {
		t.Fatalf("ids = %v, want %v", catalogIDs(got), want)
	}
	if !reflect.DeepEqual(r.calls, [][]string{{"opencode", "models"}}) {
		t.Fatalf("calls = %v", r.calls)
	}
}

func TestCatalogOpenCodeRetriesOnceOnEmptyOutput(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{{}, []byte("a/b\n")}}
	got, err := listHostModels(context.Background(), "opencode", r.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.calls) != 2 || !reflect.DeepEqual(catalogIDs(got), []string{"a/b"}) {
		t.Fatalf("calls = %d, ids = %v", len(r.calls), catalogIDs(got))
	}
}

func TestCatalogOpenCodeFailsAfterTwoEmptyOutputs(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{{}, []byte("  \n")}}
	if _, err := listHostModels(context.Background(), "opencode", r.run); err == nil {
		t.Fatal("expected an error")
	}
	if len(r.calls) != 2 {
		t.Fatalf("calls = %d, want exactly 2", len(r.calls))
	}
}

func TestCatalogOnlyOpenCodeRetries(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{{}}}
	if _, err := listHostModels(context.Background(), "pi", r.run); err == nil {
		t.Fatal("expected an error")
	}
	if len(r.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(r.calls))
	}
}

func TestCatalogPiSkipsTheHeaderAndJoinsProviderAndModel(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{readCatalogSample(t, "pi")}}
	got, err := listHostModels(context.Background(), "pi", r.run)
	if err != nil {
		t.Fatal(err)
	}
	ids := catalogIDs(got)
	if len(ids) != 8 || ids[0] != "github-copilot/claude-fable-5" || ids[7] != "xai/grok-4.5" {
		t.Fatalf("ids = %v", ids)
	}
	for _, id := range ids {
		if strings.HasPrefix(id, "provider/") {
			t.Fatalf("header was read as a model: %v", ids)
		}
	}
	if !reflect.DeepEqual(r.calls, [][]string{{"pi", "--list-models"}}) {
		t.Fatalf("calls = %v", r.calls)
	}
}

func TestCatalogGrokSkipsSessionLinesAndDefaultSuffix(t *testing.T) {
	// A session line like the real CLI prints before the list.
	out := append([]byte("You are logged in with an account.\n"), readCatalogSample(t, "grok")...)
	r := &recordingRunner{outputs: [][]byte{out}}
	got, err := listHostModels(context.Background(), "grok", r.run)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"grok-4.7", "grok-4.7-build-fast", "grok-4.6"}
	if !reflect.DeepEqual(catalogIDs(got), want) {
		t.Fatalf("ids = %v, want %v", catalogIDs(got), want)
	}
	if !reflect.DeepEqual(r.calls, [][]string{{"grok", "models"}}) {
		t.Fatalf("calls = %v", r.calls)
	}
}

func TestCatalogCursorSkipsHeaderAndTip(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{readCatalogSample(t, "cursor")}}
	got, err := listHostModels(context.Background(), "cursor", r.run)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"auto", "gpt-5.3-codex-low", "gpt-5.3-codex-low-fast", "gpt-5.3-codex", "gpt-5.3-codex-fast", "gpt-5.3-codex-high", "gpt-5.3-codex-high-fast"}
	if !reflect.DeepEqual(catalogIDs(got), want) {
		t.Fatalf("ids = %v, want %v", catalogIDs(got), want)
	}
	if !reflect.DeepEqual(r.calls, [][]string{{"cursor-agent", "models"}}) {
		t.Fatalf("calls = %v", r.calls)
	}
}

func TestCatalogDoesNotOfferInvalidIDs(t *testing.T) {
	out := "good/one\nbad;rm -rf\n$(touch x)\nspace id/x\n" + strings.Repeat("x", 300) + "\ngood/two\n"
	r := &recordingRunner{outputs: [][]byte{[]byte(out)}}
	got, err := listHostModels(context.Background(), "opencode", r.run)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"good/one", "good/two"}; !reflect.DeepEqual(catalogIDs(got), want) {
		t.Fatalf("ids = %v, want %v", catalogIDs(got), want)
	}
}

func TestCatalogErrorsWhenNoValidIDRemainsAndCountsDiscards(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{[]byte("bad;one\nbad|two\n")}}
	_, err := listHostModels(context.Background(), "opencode", r.run)
	if err == nil || !strings.Contains(err.Error(), "2") {
		t.Fatalf("err = %v, want one that counts the 2 discarded ids", err)
	}
}

func TestCatalogDoesNotRepeatAnID(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{[]byte("a/b\na/b\nc/d\n")}}
	got, _ := listHostModels(context.Background(), "opencode", r.run)
	if want := []string{"a/b", "c/d"}; !reflect.DeepEqual(catalogIDs(got), want) {
		t.Fatalf("ids = %v, want %v", catalogIDs(got), want)
	}
}

func TestCatalogClaudeReturnsFixedAliasesWithoutExecuting(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{{}}}
	got, err := listHostModels(context.Background(), "claude", r.run)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"fable", "opus", "sonnet", "haiku"}; !reflect.DeepEqual(catalogIDs(got), want) {
		t.Fatalf("ids = %v, want %v", catalogIDs(got), want)
	}
	if len(r.calls) != 0 {
		t.Fatalf("claude executed %v", r.calls)
	}
	if got, err = listHostModels(context.Background(), "claude", nil); err != nil || len(got) != 4 {
		t.Fatalf("claude with a nil runner = %v, %v", got, err)
	}
}

func TestCatalogUnknownHostIsAnError(t *testing.T) {
	if _, err := listHostModels(context.Background(), "emacs", (&recordingRunner{outputs: [][]byte{{}}}).run); err == nil {
		t.Fatal("expected an error")
	}
}

// fakeCLI writes an executable shell script named bin into a fresh directory
// and puts that directory alone on PATH.
func fakeCLI(t *testing.T, bin, script string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, bin), []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	return dir
}

func TestCatalogRunnerRunsTheFixedCommandInANeutralDirectoryWithoutATerminal(t *testing.T) {
	record := filepath.Join(t.TempDir(), "record")
	fakeCLI(t, "codex", `PATH=/usr/bin:/bin; { echo "argv=$*"; echo "cwd=$(pwd -P)"; if [ -t 0 ]; then echo stdin=tty; else echo stdin=none; fi; echo "tty=$(ps -o tty= -p $$ | tr -d ' ')"; } > "`+record+`"
echo '{"models":[]}'
`)
	repo := t.TempDir()
	t.Chdir(repo)
	out, err := newCatalogRunner(5*time.Second, 1<<20)(context.Background(), "codex", []string{"debug", "models"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != `{"models":[]}` {
		t.Fatalf("stdout = %q", out)
	}
	data, _ := os.ReadFile(record)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	tmp, _ := filepath.EvalSymlinks(os.TempDir())
	if len(lines) != 4 || lines[0] != "argv=debug models" || lines[1] != "cwd="+tmp || lines[2] != "stdin=none" {
		t.Fatalf("record = %q (want cwd %q, not the repository %q)", lines, tmp, repo)
	}
	// ps prints "??" (macOS) or "?" (Linux) for a process with no controlling terminal.
	if tty := strings.TrimPrefix(lines[3], "tty="); tty != "??" && tty != "?" {
		t.Fatalf("the CLI has a controlling terminal: %q", lines[3])
	}
}

func TestCatalogRunnerIgnoresStderr(t *testing.T) {
	fakeCLI(t, "grok", "echo noise >&2\necho out\n")
	out, err := newCatalogRunner(5*time.Second, 1<<20)(context.Background(), "grok", []string{"models"})
	if err != nil || strings.TrimSpace(string(out)) != "out" {
		t.Fatalf("out = %q, err = %v", out, err)
	}
}

func TestCatalogRunnerKillsTheWholeProcessGroupOnTimeout(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "grandchild.pid")
	fakeCLI(t, "pi", `/bin/sleep 30 &
echo $! > "`+pidFile+`"
/bin/sleep 30
`)
	start := time.Now()
	_, err := newCatalogRunner(500*time.Millisecond, 1<<20)(context.Background(), "pi", []string{"--list-models"})
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("err = %v, want a timeout", err)
	}
	if time.Since(start) > 5*time.Second {
		t.Fatalf("took %s", time.Since(start))
	}
	raw, rerr := os.ReadFile(pidFile)
	if rerr != nil {
		t.Fatal(rerr)
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(raw)))
	deadline := time.Now().Add(3 * time.Second)
	for processAlive(pid) && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if processAlive(pid) {
		_ = syscall.Kill(pid, syscall.SIGKILL)
		t.Fatalf("grandchild %d survived the timeout", pid)
	}
}

func TestCatalogRunnerFailsWhenOutputExceedsTheCap(t *testing.T) {
	fakeCLI(t, "opencode", "/usr/bin/yes line | /usr/bin/head -c 200000\n")
	out, err := newCatalogRunner(5*time.Second, 1000)(context.Background(), "opencode", []string{"models"})
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("err = %v, out length %d; want an overflow error", err, len(out))
	}
	if out != nil {
		t.Fatalf("a partial output of %d bytes was returned", len(out))
	}
}

func TestCatalogRunnerReportsAMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, err := newCatalogRunner(time.Second, 1000)(context.Background(), "codex", []string{"debug", "models"}); err == nil {
		t.Fatal("expected an error")
	}
}

func TestCatalogRunnerReportsANonZeroExit(t *testing.T) {
	fakeCLI(t, "codex", "exit 3\n")
	if _, err := newCatalogRunner(time.Second, 1000)(context.Background(), "codex", []string{"debug", "models"}); err == nil {
		t.Fatal("expected an error")
	}
}

func TestCatalogRunnerForSyntheticHomeExecutesNothing(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	fakeCLI(t, "codex", `echo ran > "`+marker+`"`+"\n")
	run := catalogRunnerFor(management.Options{Home: t.TempDir()})
	if _, err := listHostModels(context.Background(), "codex", run); err == nil {
		t.Fatal("expected an error under a synthetic home")
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the fake CLI ran under a synthetic home")
	}
	// Without --home the real runner is used and does run it.
	if _, err := catalogRunnerFor(management.Options{})(context.Background(), "codex", []string{"debug", "models"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("the real runner did not execute the fake CLI")
	}
}

// countCatalogQueries counts every attempt to list a model list, through
// whatever runner, until the test ends. Counting at listHostModels, not at the
// real runner, keeps the guard honest under --home, where catalogRunnerFor
// returns a refusing runner that a runner-level counter never sees.
func countCatalogQueries(t *testing.T) *int {
	t.Helper()
	calls := 0
	old := catalogQueryHook
	catalogQueryHook = func(string) { calls++ }
	t.Cleanup(func() { catalogQueryHook = old })
	return &calls
}

// TestCatalogIsNeverQueriedByModelsOrDoctor is AC13: the listing commands do
// not run for `hive models` or `hive doctor`. Browsing the Models view without
// opening a panel is covered in the view's tests.
func TestCatalogIsNeverQueriedByModelsOrDoctor(t *testing.T) {
	calls := countCatalogQueries(t)
	e := newModelsEnv(t)
	var out bytes.Buffer
	if err := runModels(e.common(), &out); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"models", "--home", e.home, "--state-dir", e.stateDir},
		{"doctor", "--home", e.home, "--state-dir", e.stateDir, "--project", t.TempDir()},
	} {
		if err := run(args); err != nil {
			t.Fatalf("hive %s: %v", args[0], err)
		}
	}
	if *calls != 0 {
		t.Fatalf("a model list was queried %d times", *calls)
	}
}

func catalogByID(models []catalogModel) map[string]catalogModel {
	out := map[string]catalogModel{}
	for _, m := range models {
		out[m.ID] = m
	}
	return out
}

func TestCatalogCodexReadsTheDisplayName(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{readCatalogSample(t, "codex")}}
	models, err := listHostModels(context.Background(), "codex", r.run)
	if err != nil {
		t.Fatal(err)
	}
	by := catalogByID(models)
	if by["gpt-6.1-sol"].Name != "GPT-6.1 Sol" || by["gpt-5.5"].Name != "GPT-5.5" || by["gpt-5.5"].Provider != "" {
		t.Fatalf("models = %+v", models)
	}
}

func TestCatalogCursorReadsTheNameWithoutItsMarkers(t *testing.T) {
	r := &recordingRunner{outputs: [][]byte{readCatalogSample(t, "cursor")}}
	models, err := listHostModels(context.Background(), "cursor", r.run)
	if err != nil {
		t.Fatal(err)
	}
	by := catalogByID(models)
	if by["auto"].Name != "Auto" || by["gpt-5.3-codex-low"].Name != "Codex 5.3 Low" {
		t.Fatalf("models = %+v", models)
	}
	out := []byte("Available models\n\nm-1 - Model One (current)\nm-2 - Model Two (default) (current)\n")
	r = &recordingRunner{outputs: [][]byte{out}}
	models, _ = listHostModels(context.Background(), "cursor", r.run)
	if by = catalogByID(models); by["m-1"].Name != "Model One" || by["m-2"].Name != "Model Two" {
		t.Fatalf("markers kept: %+v", models)
	}
}

func TestCatalogOpenCodeAndPiGiveTheProviderAndNoName(t *testing.T) {
	for host, ids := range map[string]map[string]string{
		"opencode": {"opencode/big-pickle": "opencode", "openai/gpt-5.5": "openai", "github-copilot/claude-opus-4.7": "github-copilot"},
		"pi":       {"openai/gpt-4": "openai", "xai/grok-4.3": "xai"},
	} {
		r := &recordingRunner{outputs: [][]byte{readCatalogSample(t, host)}}
		models, err := listHostModels(context.Background(), host, r.run)
		if err != nil {
			t.Fatal(err)
		}
		by := catalogByID(models)
		for id, provider := range ids {
			if by[id].Provider != provider || by[id].Name != "" {
				t.Errorf("%s %s = %+v, want provider %q and no name", host, id, by[id], provider)
			}
		}
	}
	for _, host := range []string{"grok", "claude"} {
		r := &recordingRunner{outputs: [][]byte{readCatalogSample(t, "grok")}}
		models, _ := listHostModels(context.Background(), host, r.run)
		for _, m := range models {
			if m.Name != "" || m.Provider != "" {
				t.Errorf("%s gave %+v", host, m)
			}
		}
	}
}

func TestCatalogNamesAreCleanedAndCapped(t *testing.T) {
	long := strings.Repeat("n", 90)
	out := []byte("{\"models\":[{\"slug\":\"m-1\",\"display_name\":\"Bad\\u001b[31m\\tName\\u202e\",\"visibility\":\"list\"},{\"slug\":\"m-2\",\"display_name\":\"" + long + "\",\"visibility\":\"list\"}]}")
	r := &recordingRunner{outputs: [][]byte{out}}
	models, err := listHostModels(context.Background(), "codex", r.run)
	if err != nil {
		t.Fatal(err)
	}
	by := catalogByID(models)
	if strings.ContainsAny(by["m-1"].Name, "\x1b\t‮") || len([]rune(by["m-2"].Name)) != 60 {
		t.Fatalf("names = %q, %q", by["m-1"].Name, by["m-2"].Name)
	}
}
