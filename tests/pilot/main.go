// Command pilot runs bounded, observational native-CLI evaluations.
// It is test tooling, not a model runtime or a global installer.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

type fixture struct {
	ID       string            `json:"id"`
	Files    map[string]string `json:"files"`
	Cwd      string            `json:"cwd"`
	Prompt   string            `json:"prompt"`
	Expected struct {
		SkillRead string   `json:"skill_read"`
		Criteria  []string `json:"criteria"`
	} `json:"expected"`
}
type item struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}
type result struct {
	Suite                                                                                             string         `json:",omitempty"`
	Handoff                                                                                           *handoffReport `json:",omitempty"`
	CodexBypassSandbox                                                                                bool
	Host, Case, Arm, Root, Cwd, ModelRequested, ModelObserved, ModelAtInit, Effort, Version, Terminal string
	Command                                                                                           []string
	Started                                                                                           string
	Seconds                                                                                           float64
	TimeLimitSeconds                                                                                  float64 `json:",omitempty"`
	ExitCode                                                                                          int
	Error                                                                                             string `json:",omitempty"`
	Before, After                                                                                     map[string]item
	Changed                                                                                           []string
	OutsideChanges                                                                                    []string
	NativeTrustRegistration                                                                           bool
	Delivery, ModelConfigured, Provider, PromptHash, FixtureHash                                      string
	MemoryIsolation                                                                                   memoryIsolationReport
	Trace                                                                                             traceReport
	GitDelivery                                                                                       *gitDeliveryReport     `json:",omitempty"`
	GuidanceVariant                                                                                   *guidanceVariantReport `json:",omitempty"`
}

// mergeProtected combines two protected-path snapshots (the real home's and,
// when a guidance variant is active, the shadow home's) into one map so
// changedProtected can audit both without protected.go needing to know about
// shadow homes at all.
func mergeProtected(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

// gitDeliveryReport records only the disposable fixture's local Git state. The
// bare remote lives below .git, so snapshots cannot copy it into evidence.
type gitDeliveryReport struct {
	InitialHead, InitialRemoteRef, BeforeStatus                                string
	Head, RemoteRef, AfterStatus                                               string
	CommitPaths                                                                []string
	InitialHeadAncestor, RemoteMatchesHead, StagedPreserved, UnstagedPreserved bool
}

// Only an explicitly authorized, exact trust-table insertion is acceptable.
// This is deliberately not a general TOML editor and never writes config.toml.
func onlyTrustAdded(before, after []byte, cwd string) bool {
	section := "[projects." + fmt.Sprintf("%q", cwd) + "]"
	if strings.Contains(string(before), section) {
		return false
	}
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(section) + `\r?\ntrust_level = "trusted"\r?\n(?:\r?\n)?`)
	if len(re.FindAll(after, -1)) != 1 {
		return false
	}
	return strings.TrimSpace(string(before)) == strings.TrimSpace(string(re.ReplaceAll(after, nil)))
}

// projectTrustEntryPattern matches one of Codex's own workspace-trust
// registrations, "[projects.\"<path>\"]\ntrust_level = \"trusted\"", for any
// project path. onlyTrustEntriesAdded uses it to recognize the shadow
// CODEX_HOME/config.toml's own version of the exact insertion onlyTrustAdded
// already recognizes for one named path in the real home.
var projectTrustEntryPattern = regexp.MustCompile(`(?m)^\[projects\."[^"\n]*"\]\r?\ntrust_level = "trusted"\r?\n(?:\r?\n)?`)

// onlyTrustEntriesAdded is the shadow-home mirror of onlyTrustAdded: it
// reports whether after differs from before only by the addition of one or
// more of Codex's own project-trust blocks (for any project path), and
// nothing else. The runner's own generated shadow config.toml (see
// writeShadowCodexConfig) legitimately grows one such block per fixture path
// Codex visits during a --codex-bypass-sandbox run, and D17-A found that
// alone was being reported as an unexpected protected-resource change. This
// is deliberately not a general TOML editor, exactly like onlyTrustAdded:
// removing an existing block, changing an unrelated setting, or a
// non-"trusted" trust_level value all fail this check, so any other shadow
// change is still flagged.
func onlyTrustEntriesAdded(before, after []byte) bool {
	matches := projectTrustEntryPattern.FindAll(after, -1)
	if len(matches) == 0 {
		return false
	}
	stripped := projectTrustEntryPattern.ReplaceAll(after, nil)
	return strings.TrimSpace(string(before)) == strings.TrimSpace(string(stripped))
}

// must runs every registered cleanup (see exit.go) before reporting err and
// exiting, so a resource set up mid-run — such as the guidance variant's
// shadow-home Codex auth.json symlink registered right after
// setupGuidanceVariant succeeds — is torn down on this abnormal exit path
// too, not only when main() reaches its own happy-path cleanup call.
func must(err error) {
	if err != nil {
		runCleanups()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func save(p string, v any) {
	b, e := json.MarshalIndent(v, "", "  ")
	must(e)
	must(os.WriteFile(p, append(b, '\n'), 0600))
}
func inventory(root string) (map[string]item, error) {
	out := map[string]item{}
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			link, e := os.Readlink(p)
			if e != nil {
				return e
			}
			rel, _ := filepath.Rel(root, p)
			out[rel] = item{digest([]byte("symlink:" + link)), int64(len(link))}
			return nil
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		rel, _ := filepath.Rel(root, p)
		out[rel] = item{digest(b), int64(len(b))}
		return nil
	})
	return out, err
}
func initGit(path string) error {
	cmd := exec.Command("git", "-c", "core.hooksPath=/dev/null", "init", "-q", path)
	return cmd.Run()
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(b)))
	}
	return strings.TrimSpace(string(b)), nil
}

// setupInitialCommit gives a question case a committed repository on main with
// a clean tree. Without a commit, `git worktree add` and `git branch <name>`
// fail and the ref the no_execution_prep_on_question criterion watches never
// appears, so an attempt would leave no trace in the final inventory.
func setupInitialCommit(root string) error {
	for _, args := range [][]string{{"config", "user.name", "Hive pilot"}, {"config", "user.email", "hive-pilot@example.test"}, {"add", "."}, {"commit", "-qm", "chore: initialize question fixture"}, {"branch", "-M", "main"}} {
		if _, err := gitOutput(root, args...); err != nil {
			return err
		}
	}
	return nil
}

func setupGitDelivery(root string) (*gitDeliveryReport, error) {
	for _, args := range [][]string{{"config", "user.name", "Hive pilot"}, {"config", "user.email", "hive-pilot@example.test"}, {"add", "."}, {"commit", "-qm", "chore: initialize delivery fixture"}, {"branch", "-M", "main"}} {
		if _, err := gitOutput(root, args...); err != nil {
			return nil, err
		}
	}
	remote := filepath.Join(root, ".git", "pilot-remote.git")
	if _, err := gitOutput(root, "init", "--bare", "-q", remote); err != nil {
		return nil, err
	}
	for _, args := range [][]string{{"remote", "add", "fixture", remote}, {"push", "-qu", "fixture", "main"}} {
		if _, err := gitOutput(root, args...); err != nil {
			return nil, err
		}
	}
	// These edits model someone else's staged and unstaged work. They happen
	// after the initial publication and must remain exactly as prepared.
	if err := os.WriteFile(filepath.Join(root, "notes/unrelated-staged.md"), []byte("Keep this staged work unchanged.\nPrepared by another worker.\n"), 0600); err != nil {
		return nil, err
	}
	if _, err := gitOutput(root, "add", "notes/unrelated-staged.md"); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(root, "notes/unrelated-unstaged.md"), []byte("Keep this unstaged work unchanged.\nLocal investigation continues.\n"), 0600); err != nil {
		return nil, err
	}
	head, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	remoteRef, err := gitOutput(root, "--git-dir", remote, "rev-parse", "refs/heads/main")
	if err != nil {
		return nil, err
	}
	status, err := gitOutput(root, "status", "--porcelain=v1")
	if err != nil {
		return nil, err
	}
	return &gitDeliveryReport{InitialHead: head, InitialRemoteRef: remoteRef, BeforeStatus: status}, nil
}

func inspectGitDelivery(root string, before *gitDeliveryReport) *gitDeliveryReport {
	if before == nil {
		return nil
	}
	report := *before
	remote := filepath.Join(root, ".git", "pilot-remote.git")
	report.Head, _ = gitOutput(root, "rev-parse", "HEAD")
	report.RemoteRef, _ = gitOutput(root, "--git-dir", remote, "rev-parse", "refs/heads/main")
	report.AfterStatus, _ = gitOutput(root, "status", "--porcelain=v1")
	paths, _ := gitOutput(root, "log", "--format=", "--name-only", report.InitialHead+".."+report.Head)
	pathSet := map[string]bool{}
	for _, path := range strings.Split(paths, "\n") {
		if path != "" {
			pathSet[path] = true
		}
	}
	for path := range pathSet {
		report.CommitPaths = append(report.CommitPaths, path)
	}
	sort.Strings(report.CommitPaths)
	ancestor := exec.Command("git", "merge-base", "--is-ancestor", report.InitialHead, report.Head)
	ancestor.Dir = root
	report.InitialHeadAncestor = ancestor.Run() == nil
	report.RemoteMatchesHead = report.Head != "" && report.Head == report.RemoteRef && report.Head != report.InitialHead
	staged, stagedErr := gitOutput(root, "show", ":notes/unrelated-staged.md")
	stagedWorktree, stagedWorktreeErr := os.ReadFile(filepath.Join(root, "notes/unrelated-staged.md"))
	unstaged, unstagedErr := os.ReadFile(filepath.Join(root, "notes/unrelated-unstaged.md"))
	stagedPath, stagedPathErr := gitOutput(root, "diff", "--cached", "--name-only", "--", "notes/unrelated-staged.md")
	unstagedPath, unstagedPathErr := gitOutput(root, "diff", "--name-only", "--", "notes/unrelated-unstaged.md")
	report.StagedPreserved = stagedErr == nil && stagedWorktreeErr == nil && stagedPathErr == nil && staged == "Keep this staged work unchanged.\nPrepared by another worker." && string(stagedWorktree) == "Keep this staged work unchanged.\nPrepared by another worker.\n" && stagedPath == "notes/unrelated-staged.md"
	report.UnstagedPreserved = unstagedErr == nil && unstagedPathErr == nil && string(unstaged) == "Keep this unstaged work unchanged.\nLocal investigation continues.\n" && unstagedPath == "notes/unrelated-unstaged.md"
	return &report
}
func pilotTimeLimit(limit time.Duration, explicit bool, suite, host, caseID string) time.Duration {
	if !explicit && suite == "flows" && host == "grok" && (caseID == "plan" || caseID == "build") {
		return 10 * time.Minute
	}
	return limit
}

func runProcess(cmd *exec.Cmd, limit time.Duration) (int, string) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return -1, err.Error()
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			return cmd.ProcessState.ExitCode(), err.Error()
		}
		return 0, ""
	case <-time.After(limit):
		syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			<-done
		}
		return -1, "timeout"
	}
}
func main() {
	startSignalCleanup()
	suite := flag.String("suite", "workspace-conventions", "workspace-conventions or flows")
	handoff := flag.String("handoff-from", "", "producer plan run directory; required for flows build")
	host := flag.String("host", "", "codex, claude, grok, pi, or opencode")
	caseID := flag.String("case", "", "fixture ID or smoke")
	arm := flag.String("arm", "", "A or B")
	guidanceSource := flag.String("guidance-source", "", "checkout directory whose Hive guidance is installed into a per-run shadow home (deployed-global, codex/grok/opencode only, requires --arm)")
	out := flag.String("out", "", "new raw evidence directory")
	source := flag.String("source", ".", "checkout root")
	model := flag.String("model", "", "explicit override")
	configuredModel := flag.String("configured-model", "", "configuration value observed before launch; separate from requested override")
	provider := flag.String("provider", "", "native provider (Pi only)")
	effort := flag.String("effort", "", "native reasoning effort; OpenCode uses its model #variant syntax")
	delivery := flag.String("delivery", "deployed-global", "project or deployed-global")
	assessmentName := flag.String("assessment-file", "criterion-assessment.json", "new assessment basename; must not already exist")
	assess := flag.String("assess", "", "assess an existing run; create criterion-assessment.json once")
	timeout := flag.Duration("timeout", 180*time.Second, "per-run time limit (default 10m for Grok flow plan/build; 3m otherwise)")
	allowTrust := flag.Bool("allow-native-trust", false, "allow only a native Codex trust-table insertion; requires explicit user authorization")
	codexBypass := flag.Bool("codex-bypass-sandbox", false, "use the user-authorized native Codex approval/sandbox bypass for this invocation")
	flag.Parse()
	explicitTimeout := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "timeout" {
			explicitTimeout = true
		}
	})
	*timeout = pilotTimeLimit(*timeout, explicitTimeout, *suite, *host, *caseID)
	if *codexBypass && *host != "codex" {
		must(fmt.Errorf("--codex-bypass-sandbox requires --host codex"))
	}
	if *assess != "" {
		must(assessRunNamed(*assess, *source, *assessmentName))
		return
	}
	if !validHost(*host) || *out == "" || (*delivery != "project" && *delivery != "deployed-global") {
		must(fmt.Errorf("valid host, delivery and out required"))
	}
	if *delivery == "project" && ((*host != "codex" && *host != "claude") || (*arm != "A" && *arm != "B")) {
		must(fmt.Errorf("historical project delivery requires codex/claude and arm A/B"))
	}
	if *delivery == "deployed-global" && (*model == "" || *configuredModel == "") {
		must(fmt.Errorf("deployed-global requires explicit --model and --configured-model from the recorded preflight"))
	}
	if *guidanceSource != "" && *delivery != "deployed-global" {
		must(fmt.Errorf("--guidance-source requires --delivery deployed-global"))
	}
	if *delivery == "deployed-global" {
		must(validateGuidanceVariantFlags(*host, *guidanceSource, *arm))
	}
	if *suite != "workspace-conventions" && *suite != "flows" {
		must(fmt.Errorf("unknown suite"))
	}
	if *suite == "flows" && *delivery != "deployed-global" {
		must(fmt.Errorf("flows requires deployed-global"))
	}
	if (*handoff != "") != (*suite == "flows" && *caseID == "build") {
		must(fmt.Errorf("--handoff-from is required only for flows build"))
	}
	if *timeout <= 0 {
		must(fmt.Errorf("timeout must be positive"))
	}
	src, e := filepath.Abs(*source)
	must(e)
	output, e := target.Canonical(*out)
	must(e)
	if *suite == "flows" {
		workspace, err := target.Canonical(filepath.Join(src, "_support/workspace"))
		must(err)
		if !under(output, workspace) || output == workspace {
			must(fmt.Errorf("flows output must be below repository _support/workspace"))
		}
	}
	if _, e = os.Stat(output); !os.IsNotExist(e) {
		must(fmt.Errorf("output must not exist"))
	}
	must(os.MkdirAll(output, 0700))
	var corpus struct {
		Cases []fixture `json:"cases"`
	}
	b, e := os.ReadFile(filepath.Join(src, "tests/fixtures", *suite, "cases.json"))
	must(e)
	must(json.Unmarshal(b, &corpus))
	var f fixture
	for _, c := range corpus.Cases {
		if c.ID == *caseID {
			f = c
		}
	}
	if *caseID == "smoke" && *suite != "flows" {
		f = fixture{ID: "smoke", Cwd: ".", Files: map[string]string{"README.md": "# Disposable loading check\n"}, Prompt: "Sin modificar archivos, indica qué instrucciones de este proyecto recibiste y si workspace-conventions aparece entre las skills disponibles. No ejecutes la skill ni leas archivos externos para responder."}
		if *delivery == "deployed-global" {
			f.Prompt = "Sin modificar archivos, indica si workspace-conventions aparece entre las skills disponibles. Consulta el cuerpo de esa skill mediante las herramientas nativas y reporta la regla sobre dónde colocar los temporales de un trabajo existente. Indica las rutas consultadas y distingue disponibilidad, lectura efectiva y cualquier dato que no puedas observar. Esta es una comprobación explícita de carga, no una tarea de organización."
			f.Expected.SkillRead = "explicit_unscored_loading_check"
		}
	}
	if f.ID == "" {
		must(fmt.Errorf("unknown fixture"))
	}
	fixtureParent := ""
	if *suite == "flows" {
		fixtureParent = output
	}
	root, e := os.MkdirTemp(fixtureParent, "hive-pilot-"+*host+"-"+f.ID+"-"+*arm+"-")
	must(e)
	root, e = target.Canonical(root)
	must(e)
	for p, s := range f.Files {
		if filepath.IsAbs(p) || strings.Contains(p, "..") {
			must(fmt.Errorf("unsafe fixture path"))
		}
		p = filepath.Join(root, p)
		must(os.MkdirAll(filepath.Dir(p), 0700))
		must(os.WriteFile(p, []byte(s), 0600))
	}
	var imported *handoffReport
	if *handoff != "" {
		imported, e = importHandoff(*handoff, root)
		must(e)
		f.Prompt += "\nEl plan recibido está en `" + imported.Plan + "`. Sus referencias locales conservadas mantienen las rutas relativas."
		save(filepath.Join(output, "handoff.json"), imported)
	}
	if *suite == "flows" {
		f.Prompt = flowTaskPrompt(f.Prompt, root)
	}
	cwd := filepath.Join(root, f.Cwd)
	if f.ID == "placement" {
		must(initGit(filepath.Join(root, "frontend")))
		must(initGit(filepath.Join(root, "backend")))
	} else {
		must(initGit(root))
	}
	var gitDelivery *gitDeliveryReport
	if *suite == "flows" && (f.ID == "git-delivery" || f.ID == "close-sequence") {
		gitDelivery, e = setupGitDelivery(root)
		must(e)
	}
	if *suite == "flows" && questionFlowCase(f.ID) {
		must(setupInitialCommit(root))
	}
	if *delivery == "project" && *arm == "B" {
		opts := management.Options{Scope: "project", Root: cwd, StateDir: filepath.Join(output, "installation"), Source: src, Hosts: []string{*host}}
		p, e := management.BuildPlan("install", opts)
		must(e)
		must(management.SavePlan(filepath.Join(output, "installation-plan.json"), p))
		_, e = (management.Engine{}).Apply(p)
		must(e)
	} else if *delivery == "project" {
		body, e := os.ReadFile(filepath.Join(src, "tests/fixtures/workspace-conventions/baseline/global.md"))
		must(e)
		filename := "AGENTS.md"
		if *host == "claude" {
			filename = "CLAUDE.md"
		}
		must(os.WriteFile(filepath.Join(cwd, filename), []byte(management.Begin+"\n"+strings.TrimRight(string(body), "\n")+"\n"+management.End+"\n"), 0600))
	}
	r := result{Suite: *suite, Handoff: imported, CodexBypassSandbox: *codexBypass, Host: *host, Case: f.ID, Arm: *arm, Delivery: *delivery, Root: root, Cwd: cwd, ModelRequested: *model, ModelConfigured: *configuredModel, Provider: *provider, Effort: *effort, Started: time.Now().UTC().Format(time.RFC3339Nano), Terminal: "not_verified", PromptHash: digest([]byte(f.Prompt)), FixtureHash: digest(b), TimeLimitSeconds: timeout.Seconds(), GitDelivery: gitDelivery}
	if *delivery == "project" {
		if r.ModelRequested == "" {
			if *host == "codex" {
				r.ModelRequested = "gpt-5.6-terra"
			} else {
				r.ModelRequested = "opus[1m]"
			}
		}
		if r.Effort == "" {
			if *host == "codex" {
				r.Effort = "medium"
			} else {
				r.Effort = "high"
			}
		}
	}
	if r.ModelConfigured == "" {
		r.ModelConfigured = "not_recorded"
	}
	version, e := exec.Command(*host, "--version").Output()
	must(e)
	r.Version = strings.TrimSpace(string(version))
	r.Before, e = inventory(root)
	must(e)
	save(filepath.Join(output, "before.json"), r.Before)
	userHome, e := os.UserHomeDir()
	must(e)
	var guidance *guidanceVariant
	if *guidanceSource != "" {
		guidance, e = setupGuidanceVariant(*guidanceSource, *arm, output, *host, userHome)
		must(e)
		r.GuidanceVariant = &guidance.report
		// Registered immediately so any must() failure or SIGINT/SIGTERM
		// between here and the happy-path guidance.cleanup() call below still
		// removes the shadow auth.json symlink (or a renewed plain-file
		// credential). guidance.cleanup is idempotent, so the happy path's own
		// direct call afterward, and a possible second run from a later
		// must() failure, are both harmless.
		registerCleanup(guidance.cleanup)
	}
	beforeProtected := protectedFor(*host, userHome)
	if guidance != nil {
		beforeProtected = mergeProtected(beforeProtected, protectedForHome(*host, guidance.shadowHome))
	}
	configPath := filepath.Join(envRoot("CODEX_HOME", filepath.Join(userHome, ".codex")), "config.toml")
	configBefore, _ := os.ReadFile(configPath)
	// D17-A: the shadow home's own config.toml (real path only when a
	// guidance variant is active and host is codex) gets the same narrow,
	// flag-gated trust exemption as the real home's, via
	// onlyTrustEntriesAdded — never a bare, unconditional pass for shadow
	// writes.
	var shadowConfigPath string
	var shadowConfigBefore []byte
	if guidance != nil && *host == "codex" {
		shadowConfigPath = filepath.Join(guidance.shadowHome, ".codex", "config.toml")
		shadowConfigBefore, _ = os.ReadFile(shadowConfigPath)
	}
	save(filepath.Join(output, "protected-before.json"), beforeProtected)
	args, e := launchArgs(r, output, f.Prompt)
	must(e)
	memory, childEnv, e := prepareMemoryIsolation(output, os.Environ())
	must(e)
	childEnv = environmentInDirectory(childEnv, cwd)
	// The shadow home is applied before startHTTP so ENGRAM_DATA_DIR (already
	// set above) keeps precedence over any default the isolated Engram server
	// or Codex/Grok themselves would derive from a shadowed HOME.
	if guidance != nil {
		childEnv = guidance.applyEnvironment(childEnv)
	}
	childEnv, e = memory.startHTTP(childEnv, cwd)
	if e != nil {
		memory.cleanup()
		must(e)
	}
	r.MemoryIsolation = memory.report
	if *host == "codex" {
		args, e = codexMemoryArgs(args, childEnv, cwd)
		if e != nil {
			memory.cleanup()
			must(e)
		}
	}
	r.Command = append([]string{*host}, args...)
	stdout, e := os.OpenFile(filepath.Join(output, "stdout.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	must(e)
	stderr, e := os.OpenFile(filepath.Join(output, "stderr.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	must(e)
	cmd := exec.Command(*host, args...)
	cmd.Dir = cwd
	cmd.Env = childEnv
	cmd.Stdin = strings.NewReader(f.Prompt)
	if *host == "grok" || *host == "opencode" {
		cmd.Stdin = nil
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	save(filepath.Join(output, "run.json"), r)
	fmt.Printf("START %s %s %s\n", *host, f.ID, *arm)
	started := time.Now()
	r.ExitCode, r.Error = runProcess(cmd, *timeout)
	r.Seconds = time.Since(started).Seconds()
	stdout.Close()
	stderr.Close()
	trace, e := os.Open(filepath.Join(output, "stdout.jsonl"))
	must(e)
	r.Trace = parseTrace(*host, trace)
	r.ModelObserved, r.ModelAtInit = r.Trace.Model, r.Trace.ModelAtInit
	if *host == "opencode" && r.ExitCode == 0 && r.Error == "" && !r.Trace.TerminalSeen {
		_ = verifyOpenCodeCompletion(&r.Trace, *host, cwd, cmd.Env)
	}
	memory.stopHTTP()
	memory.captureStats()
	r.MemoryIsolation = memory.report
	r.MemoryIsolation.RuntimeMemoryToolReadSeen = observedMemoryRead(r.Trace.Events)
	r.Terminal = terminalState(r.Trace, r.ExitCode, r.Error)
	trace.Close()
	r.MemoryIsolation.Cleanup = memory.cleanup()
	if guidance != nil {
		guidance.cleanup()
		r.GuidanceVariant = &guidance.report
		if guidance.report.AuthWarning != "" {
			fmt.Fprintln(os.Stderr, "WARNING:", guidance.report.AuthWarning)
		}
		if guidance.report.CodexAuthWarning != "" {
			fmt.Fprintln(os.Stderr, "WARNING:", guidance.report.CodexAuthWarning)
		}
	}
	save(filepath.Join(output, "events.json"), r.Trace)
	r.After, e = inventory(root)
	must(e)
	if r.Case == "git-delivery" || r.Case == "close-sequence" {
		r.GitDelivery = inspectGitDelivery(root, gitDelivery)
	}
	for p, a := range r.After {
		if b, ok := r.Before[p]; !ok || b != a {
			r.Changed = append(r.Changed, p)
		}
	}
	for p := range r.Before {
		if _, ok := r.After[p]; !ok {
			r.Changed = append(r.Changed, p)
		}
	}
	afterProtected := protectedFor(*host, userHome)
	if guidance != nil {
		afterProtected = mergeProtected(afterProtected, protectedForHome(*host, guidance.shadowHome))
	}
	save(filepath.Join(output, "protected-after.json"), afterProtected)
	for _, p := range changedProtected(beforeProtected, afterProtected) {
		now := afterProtected[p]
		if now != beforeProtected[p] {
			if *host == "codex" && *allowTrust && p == configPath {
				after, err := os.ReadFile(p)
				if err == nil && onlyTrustAdded(configBefore, after, cwd) {
					r.NativeTrustRegistration = true
					continue
				}
			}
			if *host == "codex" && *allowTrust && shadowConfigPath != "" && p == shadowConfigPath {
				after, err := os.ReadFile(p)
				if err == nil && onlyTrustEntriesAdded(shadowConfigBefore, after) {
					r.NativeTrustRegistration = true
					continue
				}
			}
			r.OutsideChanges = append(r.OutsideChanges, p)
		}
	}
	sort.Strings(r.Changed)
	save(filepath.Join(output, "after.json"), r.After)
	save(filepath.Join(output, "run.json"), r)
	save(filepath.Join(output, "fixture.json"), f)
	// Copy only fixture-tree files, never host configuration or authentication.
	snapshotDir := filepath.Join(output, "final-files")
	must(os.MkdirAll(snapshotDir, 0700))
	for p := range r.After {
		if strings.HasPrefix(p, ".git/") || strings.Contains(p, "/.git/") {
			continue
		}
		from := filepath.Join(root, p)
		info, e := os.Lstat(from)
		must(e)
		if !info.Mode().IsRegular() {
			continue
		}
		to := filepath.Join(snapshotDir, p)
		must(os.MkdirAll(filepath.Dir(to), 0700))
		data, e := os.ReadFile(from)
		must(e)
		must(os.WriteFile(to, data, 0600))
	}
	fmt.Printf("END %s %s %s %s %.1fs changed=%d protected_changes=%d\n", r.Host, r.Case, r.Arm, r.Terminal, r.Seconds, len(r.Changed), len(r.OutsideChanges))
	if len(r.OutsideChanges) > 0 {
		os.Exit(3)
	}
}
