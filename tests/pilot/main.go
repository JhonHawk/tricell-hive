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

func must(err error) {
	if err != nil {
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
	suite := flag.String("suite", "workspace-conventions", "workspace-conventions or flows")
	handoff := flag.String("handoff-from", "", "producer plan run directory; required for flows build")
	host := flag.String("host", "", "codex, claude, grok, pi, or opencode")
	caseID := flag.String("case", "", "fixture ID or smoke")
	arm := flag.String("arm", "", "A or B")
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
	if *delivery == "deployed-global" && *arm != "" {
		must(fmt.Errorf("deployed-global does not select or inject an A/B arm"))
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
	r := result{Suite: *suite, Handoff: imported, CodexBypassSandbox: *codexBypass, Host: *host, Case: f.ID, Arm: *arm, Delivery: *delivery, Root: root, Cwd: cwd, ModelRequested: *model, ModelConfigured: *configuredModel, Provider: *provider, Effort: *effort, Started: time.Now().UTC().Format(time.RFC3339Nano), Terminal: "not_verified", PromptHash: digest([]byte(f.Prompt)), FixtureHash: digest(b), TimeLimitSeconds: timeout.Seconds()}
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
	beforeProtected := protectedFor(*host, userHome)
	configPath := filepath.Join(envRoot("CODEX_HOME", filepath.Join(userHome, ".codex")), "config.toml")
	configBefore, _ := os.ReadFile(configPath)
	save(filepath.Join(output, "protected-before.json"), beforeProtected)
	args, e := launchArgs(r, output, f.Prompt)
	must(e)
	memory, childEnv, e := prepareMemoryIsolation(output, os.Environ())
	must(e)
	childEnv = environmentInDirectory(childEnv, cwd)
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
	save(filepath.Join(output, "events.json"), r.Trace)
	r.After, e = inventory(root)
	must(e)
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
