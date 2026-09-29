package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// statusRunner is an ExternalRunner whose steps all reconcile to one status.
type statusRunner struct{ status string }

func (statusRunner) Validate(management.ExternalStep) error { return nil }
func (statusRunner) Execute(management.ExternalStep) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (r statusRunner) Reconcile(management.ExternalStep, json.RawMessage) (string, error) {
	return r.status, nil
}

// writeOnboardingRecord finishes a real onboarding over the installed home in
// o, so the record carries the manager's own integrity hash. Each step ID is
// "<provider>-manual", as the native adapter names them.
func writeOnboardingRecord(t *testing.T, o management.Options, status string, providerIDs ...string) management.OnboardingResult {
	t.Helper()
	var steps []management.ExternalStep
	for _, id := range providerIDs {
		steps = append(steps, management.ExternalStep{ID: id + "-manual", Payload: json.RawMessage(`{}`)})
	}
	o.Source, o.Hosts = minimalTestSource(t), []string{"claude"}
	plan, err := management.BuildPlan("install", o)
	if err != nil {
		t.Fatal(err)
	}
	result, err := management.Engine{}.Onboard(plan, steps, statusRunner{status})
	// A step that ends short of verified finishes the record and still reports it.
	if err != nil && !strings.Contains(err.Error(), "optional capabilities incomplete") {
		t.Fatalf("onboard: %v", err)
	}
	return result
}

// tamperOnboardingRecord changes one byte of meaning in the newest record.
func tamperOnboardingRecord(t *testing.T, stateDir string) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(stateDir, "onboarding", "*.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("records = %v, %v", files, err)
	}
	raw, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(string(raw), `"manual"`, `"verified"`, 1)
	if changed == string(raw) {
		t.Fatal("nothing to tamper with")
	}
	if err := os.WriteFile(files[0], []byte(changed), 0o600); err != nil {
		t.Fatal(err)
	}
}

// fakePrograms puts engram and agent-browser scripts in a directory that is the
// whole PATH. Each script writes marker when it runs.
func fakePrograms(t *testing.T, names ...string) (pathDir, marker string) {
	t.Helper()
	pathDir = t.TempDir()
	marker = filepath.Join(t.TempDir(), "executed")
	for _, name := range names {
		script := "#!/bin/sh\necho " + name + " >> '" + marker + "'\n"
		if err := os.WriteFile(filepath.Join(pathDir, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", pathDir)
	return pathDir, marker
}

// integrationsFixture is an installed home with nothing detected.
func integrationsFixture(t *testing.T) (o management.Options, f *doctorFake, home, stateDir string) {
	t.Helper()
	o, home, stateDir = doctorHome(t, "claude")
	o = nonSyntheticOptions(t, o, home)
	return o, newDoctorFake(home), home, stateDir
}

func integrationsText(o management.Options, deps doctorDeps) string {
	return sectionText(collectIntegrations(o, deps))
}

func mustNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("%s exists (%v)", path, err)
	}
}

func rowByID(t *testing.T, rows []integrationRow, id string) integrationRow {
	t.Helper()
	for _, r := range rows {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("no row %q in %+v", id, rows)
	return integrationRow{}
}

func TestIntegrationsSectionListsTheFourRowsInOrder(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	sec := collectIntegrations(o, f.deps())
	if sec.Title != "Integrations" || sec.Err != "" {
		t.Fatalf("section = %+v", sec)
	}
	text := sectionText(sec)
	last := -1
	for _, name := range []string{"Engram", "Context7", "pi-subagents", "agent-browser"} {
		i := strings.Index(text, name)
		if i < 0 || i < last {
			t.Fatalf("%s missing or out of order in:\n%s", name, text)
		}
		last = i
	}
}

func TestIntegrationsDetectsProgramsWithoutRunningThem(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	pathDir, marker := fakePrograms(t, "engram", "agent-browser", "pi")
	deps := f.deps()
	var looked []string
	deps.lookPath = func(name string) (string, error) {
		looked = append(looked, name)
		return exec.LookPath(name)
	}
	text := integrationsText(o, deps)
	mustContain(t, text, filepath.Join(pathDir, "engram"), filepath.Join(pathDir, "agent-browser"))
	mustNotExist(t, marker)
	if len(f.ran) != 0 {
		t.Fatalf("a program was run: %v", f.ran)
	}
	if fmt.Sprint(looked) != "[engram agent-browser]" {
		t.Fatalf("looked up %v; only engram and agent-browser are detected (pi-subagents needs Pi's configuration)", looked)
	}
}

func TestIntegrationsReportsProgramsNotFound(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	fakePrograms(t) // an empty PATH
	deps := f.deps()
	deps.lookPath = exec.LookPath
	rows, _ := collectIntegrationRows(o, deps)
	for _, id := range []string{"engram", "agent-browser"} {
		if r := rowByID(t, rows, id); r.Found != "not detected" {
			t.Fatalf("%s found = %q", id, r.Found)
		}
	}
}

func TestIntegrationsUnderASyntheticHomeChecksNoProgram(t *testing.T) {
	o, home, _ := doctorHome(t, "claude") // synthetic: Home is set
	f := newDoctorFake(home)
	_, marker := fakePrograms(t, "engram", "agent-browser")
	deps := f.deps()
	deps.lookPath = func(name string) (string, error) {
		t.Fatalf("lookPath(%q) under a synthetic home", name)
		return "", nil
	}
	rows, _ := collectIntegrationRows(o, deps)
	mustNotExist(t, marker)
	for _, id := range []string{"engram", "agent-browser"} {
		r := rowByID(t, rows, id)
		if r.Found != "not checked" && !(id == "agent-browser" && r.Found == "not detected") {
			t.Fatalf("%s found = %q under a synthetic home", id, r.Found)
		}
		mustContain(t, strings.Join(r.Evidence, "\n"), "synthetic")
	}
}

func TestIntegrationsShowsTheContext7SkillWithItsPath(t *testing.T) {
	o, f, home, _ := integrationsFixture(t)
	skill := filepath.Join(home, ".agents", "skills", "find-docs", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skill), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill, []byte("---\nname: find-docs\n---\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rows, _ := collectIntegrationRows(o, f.deps())
	r := rowByID(t, rows, "context7")
	if r.Found != "detected" {
		t.Fatalf("found = %q", r.Found)
	}
	mustContain(t, strings.Join(r.Evidence, "\n"), skill)

	// An empty file, and a folder of that name, are not evidence.
	other := filepath.Join(home, ".claude", "skills", "context7-mcp", "SKILL.md")
	if err := os.MkdirAll(other, 0o700); err != nil {
		t.Fatal(err)
	}
	rows, _ = collectIntegrationRows(o, f.deps())
	text := strings.Join(rowByID(t, rows, "context7").Evidence, "\n")
	mustContain(t, text, "Not verified: "+other)
}

func TestIntegrationsContext7AbsentSaysSo(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	rows, _ := collectIntegrationRows(o, f.deps())
	r := rowByID(t, rows, "context7")
	if r.Found != "not detected" {
		t.Fatalf("found = %q", r.Found)
	}
	mustContain(t, strings.Join(r.Evidence, "\n"), "No find-docs or context7-mcp SKILL.md")
}

func TestIntegrationsContext7UsesTheEnvironmentRoots(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	claudeDir := filepath.Join(t.TempDir(), "custom-claude")
	f.env["CLAUDE_CONFIG_DIR"] = claudeDir
	skill := filepath.Join(claudeDir, "skills", "find-docs", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skill), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	mustContain(t, integrationsText(o, f.deps()), skill)
}

func TestIntegrationsAgentBrowserSkillUsesTheSameRoots(t *testing.T) {
	o, f, home, _ := integrationsFixture(t)
	skill := filepath.Join(home, ".codex", "skills", "agent-browser", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skill), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	rows, _ := collectIntegrationRows(o, f.deps())
	r := rowByID(t, rows, "agent-browser")
	if r.Found != "skill" {
		t.Fatalf("found = %q", r.Found)
	}
	mustContain(t, strings.Join(r.Evidence, "\n"), skill)
}

func TestIntegrationsRecordStatusReachesTheMatchingRows(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	writeOnboardingRecord(t, o, "manual", "engram", "pi-subagents")
	rows, _ := collectIntegrationRows(o, f.deps())
	if r := rowByID(t, rows, "engram"); r.Record != "manual" {
		t.Fatalf("engram record = %q", r.Record)
	}
	pi := rowByID(t, rows, "pi-subagents")
	if pi.Record != "manual" {
		t.Fatalf("pi-subagents record = %q", pi.Record)
	}
	if pi.Found != "not checked" {
		t.Fatalf("pi-subagents found = %q; Hive cannot detect it", pi.Found)
	}
	// A provider the last onboarding never offered says so; agent-browser is not part of onboarding.
	if r := rowByID(t, rows, "context7"); r.Record != "not in the last record" {
		t.Fatalf("context7 record = %q", r.Record)
	}
	if r := rowByID(t, rows, "agent-browser"); r.Record != "not part of onboarding" {
		t.Fatalf("agent-browser record = %q", r.Record)
	}
	mustContain(t, rowByID(t, rows, "engram").RecordDetail, "manual (onboarding partial, ")
}

func TestIntegrationsRecordStatusIsTheStepStatusAsWritten(t *testing.T) {
	for _, status := range []string{"verified", "failed", "auth_pending", "skipped"} {
		t.Run(status, func(t *testing.T) {
			o, f, _, _ := integrationsFixture(t)
			writeOnboardingRecord(t, o, status, "engram")
			rows, _ := collectIntegrationRows(o, f.deps())
			if r := rowByID(t, rows, "engram"); r.Record != status {
				t.Fatalf("record = %q, want %q", r.Record, status)
			}
		})
	}
}

func TestIntegrationsEveryRowShowsItsSourceAndNextStep(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	rows, _ := collectIntegrationRows(o, f.deps())
	want := map[string]string{
		"engram":        "github.com/Gentleman-Programming/engram",
		"context7":      "context7.com",
		"pi-subagents":  "github.com/nicobailon/pi-subagents",
		"agent-browser": agentBrowserSource,
	}
	for id, source := range want {
		r := rowByID(t, rows, id)
		if r.Source != source || r.Next == "" {
			t.Fatalf("%s: source %q next %q", id, r.Source, r.Next)
		}
	}
	if next := rowByID(t, rows, "engram").Next; !strings.Contains(next, "Install it from github.com/Gentleman-Programming/engram") {
		t.Fatalf("engram next step = %q; want a step for a program that was not found", next)
	}
	text := integrationsText(o, f.deps())
	for _, source := range want {
		mustContain(t, text, "Source: "+source)
	}
	mustContain(t, text, "Next step:")
}

// installSkill writes a nonempty skill file, the evidence Hive looks for.
func installSkill(t *testing.T, home, root, name string) {
	t.Helper()
	path := filepath.Join(home, root, "skills", name, "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestIntegrationsNextStepIsConcretePerCase covers M1: each row says what to do
// in its own case (found, not found, or not checkable), with the one command
// Hive knows (Context7's) and the official source, and never the catalog's
// "validation is pending" reason.
func TestIntegrationsNextStepIsConcretePerCase(t *testing.T) {
	const ctx7 = "npx ctx7@latest setup --cli"
	for _, tc := range []struct {
		name  string
		setup func(t *testing.T, f *doctorFake, home string)
		want  map[string][]string // row ID -> substrings its Next must hold
	}{
		{"nothing found", func(t *testing.T, f *doctorFake, home string) {}, map[string][]string{
			"context7":      {"To install it, run " + ctx7, "sign in"},
			"engram":        {"Install it from github.com/Gentleman-Programming/engram", "Hive does not install it"},
			"pi-subagents":  {"cannot check it without reading Pi's configuration", "If you use Pi, install it with Pi's own package manager", "github.com/nicobailon/pi-subagents"},
			"agent-browser": {"Optional", "Hive does not install it", "follow the official instructions at " + agentBrowserSource},
		}},
		{"everything found", func(t *testing.T, f *doctorFake, home string) {
			f.install("engram", "")
			f.install("agent-browser", "")
			installSkill(t, home, ".agents", "find-docs")
			installSkill(t, home, ".agents", "agent-browser")
		}, map[string][]string{
			"context7":      {"To refresh it, run " + ctx7, "sign in"},
			"engram":        {"Hive does not configure Engram", "each CLI", "github.com/Gentleman-Programming/engram"},
			"pi-subagents":  {"cannot check it"},
			"agent-browser": {"Hive does not install or update it", "follow the official instructions at " + agentBrowserSource},
		}},
		{"agent-browser only as a program", func(t *testing.T, f *doctorFake, home string) {
			f.install("agent-browser", "")
		}, map[string][]string{
			"agent-browser": {"program was found but no skill file", "follow the official instructions at " + agentBrowserSource},
		}},
		{"agent-browser only as a skill", func(t *testing.T, f *doctorFake, home string) {
			installSkill(t, home, ".codex", "agent-browser")
		}, map[string][]string{
			"agent-browser": {"skill file was found but not the program", "follow the official instructions at " + agentBrowserSource},
		}},
		{"context7 file that cannot be verified", func(t *testing.T, f *doctorFake, home string) {
			if err := os.MkdirAll(filepath.Join(home, ".agents", "skills", "find-docs", "SKILL.md"), 0o700); err != nil {
				t.Fatal(err)
			}
		}, map[string][]string{
			"context7": {"could not be verified", "run " + ctx7, "sign in"},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o, f, home, _ := integrationsFixture(t)
			tc.setup(t, f, home)
			rows, _ := collectIntegrationRows(o, f.deps())
			for id, wants := range tc.want {
				next := rowByID(t, rows, id).Next
				mustContain(t, next, wants...)
			}
			assertNoCatalogReasonOrInventedCommand(t, rows, ctx7)
			mustNotContain(t, integrationsText(o, f.deps()), "validation is pending")
		})
	}
}

// assertNoCatalogReasonOrInventedCommand fails when a next step repeats the
// catalog's pending-validation reason, or names an install command Hive has not
// verified: only Context7's is known.
func assertNoCatalogReasonOrInventedCommand(t *testing.T, rows []integrationRow, known string) {
	t.Helper()
	for _, r := range rows {
		mustNotContain(t, r.Next, "validation is pending", "Native install")
		rest := strings.ReplaceAll(r.Next, known, "")
		mustNotContain(t, rest, "npx ", "npm ", "brew ", "go install", "pip ", "curl ", "cargo ")
		if r.ID != "context7" && strings.Contains(r.Next, known) {
			t.Errorf("%s names Context7's command: %q", r.ID, r.Next)
		}
	}
}

func TestIntegrationsNextStepWhenNothingCanBeChecked(t *testing.T) {
	o, home, _ := doctorHome(t, "claude") // synthetic: no program is looked up
	rows, _ := collectIntegrationRows(o, newDoctorFake(home).deps())
	engram := rowByID(t, rows, "engram").Next
	mustContain(t, engram, "could not look for it here", "github.com/Gentleman-Programming/engram", "Hive does not install")
	mustNotContain(t, engram, "Install it from")
	browser := rowByID(t, rows, "agent-browser").Next
	mustContain(t, browser, "could not look for the program here", "follow the official instructions at "+agentBrowserSource)
	assertNoCatalogReasonOrInventedCommand(t, rows, "npx ctx7@latest setup --cli")
}

// TestIntegrationsNextStepSaysWhatTheLastRecordLeftInPlainWords keeps the
// record's status in its own column and only paraphrases it in the next step.
func TestIntegrationsNextStepSaysWhatTheLastRecordLeftInPlainWords(t *testing.T) {
	for status, words := range map[string]string{
		"manual":       "left this for you to finish by hand",
		"verified":     "recorded it as confirmed",
		"failed":       "could not finish it",
		"auth_pending": "was waiting for you to sign in",
		"skipped":      "did not run it",
	} {
		t.Run(status, func(t *testing.T) {
			o, f, _, _ := integrationsFixture(t)
			writeOnboardingRecord(t, o, status, "engram", "context7")
			rows, _ := collectIntegrationRows(o, f.deps())
			for _, id := range []string{"engram", "context7"} {
				r := rowByID(t, rows, id)
				if r.Record != status {
					t.Fatalf("%s record = %q, want %q", id, r.Record, status)
				}
				mustContain(t, r.Next, "The last install "+words)
			}
			assertNoCatalogReasonOrInventedCommand(t, rows, "npx ctx7@latest setup --cli")
		})
	}
	o, f, _, _ := integrationsFixture(t)
	writeOnboardingRecord(t, o, "manual", "engram")
	rows, _ := collectIntegrationRows(o, f.deps())
	for _, id := range []string{"context7", "pi-subagents", "agent-browser"} {
		mustNotContain(t, rowByID(t, rows, id).Next, "The last install")
	}
}

func TestIntegrationsWithoutARecordSaysSo(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	rows, errText := collectIntegrationRows(o, f.deps())
	if errText != "" {
		t.Fatalf("err = %q", errText)
	}
	for _, id := range []string{"engram", "context7", "pi-subagents"} {
		if r := rowByID(t, rows, id); r.Record != "No onboarding record yet" {
			t.Fatalf("%s record = %q", id, r.Record)
		}
	}
	mustContain(t, integrationsText(o, f.deps()), "No onboarding record yet")
}

func TestIntegrationsTamperedRecordIsAnErrorNotAStatus(t *testing.T) {
	o, f, _, stateDir := integrationsFixture(t)
	writeOnboardingRecord(t, o, "manual", "engram")
	tamperOnboardingRecord(t, stateDir)
	sec := collectIntegrations(o, f.deps())
	if sec.Err == "" || !strings.Contains(sec.Err, "invalid onboarding journal") {
		t.Fatalf("err = %q", sec.Err)
	}
	text := sectionText(sec)
	mustContain(t, text, "Could not check everything", "record unreadable", "invalid onboarding journal")
	mustNotContain(t, text, "No onboarding record yet")
	// M2: plain words and the way out first, the technical error on its own line.
	if !strings.HasPrefix(sec.Err, "The last integrations record is damaged or unreadable; running hive install again writes a new one.\nDetail: ") {
		t.Fatalf("err = %q", sec.Err)
	}
	mustContain(t, text, "Last onboarding: The record is damaged or unreadable; running hive install again writes a new one.\n", "    Detail: cannot read the last onboarding record: ", "invalid onboarding journal")
	mustNotContain(t, text, "Last onboarding: Cannot read")
	rows, _ := collectIntegrationRows(o, f.deps())
	if r := rowByID(t, rows, "engram"); r.Record != "record unreadable" || r.Source == "" {
		t.Fatalf("engram = %+v; the local checks still show", r)
	}
}

func TestIntegrationsLeavesStateAndHomeUnchanged(t *testing.T) {
	o, f, home, stateDir := integrationsFixture(t)
	writeOnboardingRecord(t, o, "manual", "engram")
	fakePrograms(t, "engram", "agent-browser")
	deps := f.deps()
	deps.lookPath = exec.LookPath
	beforeHome, beforeState := collectFiles(t, home), collectFiles(t, stateDir)
	stateJSON, err := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	collectIntegrations(o, deps)
	collectDoctor(o, t.TempDir(), deps)
	afterState, _ := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if string(stateJSON) != string(afterState) {
		t.Fatal("state.json changed")
	}
	if fmt.Sprint(beforeHome) != fmt.Sprint(collectFiles(t, home)) || fmt.Sprint(beforeState) != fmt.Sprint(collectFiles(t, stateDir)) {
		t.Fatal("files changed")
	}
}

func TestIntegrationsSanitizesTextFromOutside(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	evil := filepath.Join(t.TempDir(), "h\x1b[2Jx")
	skill := filepath.Join(evil, ".agents", "skills", "find-docs", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skill), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	f.home = evil
	deps := f.deps()
	deps.lookPath = func(name string) (string, error) { return "/bin/en\x1b[31mgram\x07", nil }
	sec := collectIntegrations(o, deps)
	text := sectionText(sec)
	if strings.ContainsAny(text, "\x1b\x07") {
		t.Fatalf("control characters in %q", text)
	}
	mustContain(t, text, "hx", "/bin/engram")
}

func TestIntegrationsSectionLinesComeFromTheRows(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	rows, errText := collectIntegrationRows(o, f.deps())
	want := integrationsSection(rows, errText)
	got := collectIntegrations(o, f.deps())
	if fmt.Sprint(want) != fmt.Sprint(got) {
		t.Fatalf("section differs from the rows:\n%v\n%v", want, got)
	}
	for _, r := range rows {
		mustContain(t, sectionText(got), r.Name)
		for _, l := range r.detailLines() {
			mustContain(t, sectionText(got), l)
		}
	}
}

func TestCollectDoctorFillsTheIntegrationsSection(t *testing.T) {
	o, f, _, _ := integrationsFixture(t)
	report := collectDoctor(o, t.TempDir(), f.deps())
	if report.Integrations.Title != "Integrations" {
		t.Fatalf("section = %+v", report.Integrations)
	}
	var b strings.Builder
	renderDoctorText(report, &b)
	mustContain(t, b.String(), "Integrations", "Engram")
}
