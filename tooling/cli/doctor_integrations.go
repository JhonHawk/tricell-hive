// doctor_integrations.go builds the Integrations section of the diagnostics
// (design.md "Reglas por sección"): for each optional capability, what Hive
// can see locally, what the last onboarding record says, and where the
// official instructions are. It only reads. It never runs an integration
// program (a program is only looked up in PATH, and a browser is looked for
// by file existence) and never reads a host's
// configuration: whether Pi has pi-subagents, or a host loads a skill, is not
// something Hive checks.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

// agentBrowserSource is where agent-browser's own instructions live; it equals
// the catalog entry's Source (Context7 resolves the library to
// /vercel-labs/agent-browser).
const agentBrowserSource = "github.com/vercel-labs/agent-browser"

// piCodemodeSource is Pi's own documentation of how to enable codemode.
const piCodemodeSource = "github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/cli.md#enable-codemode"

// context7Command is the one vendor command Hive has verified (see "Optional
// Context7 setup recommendation" in deployment-manager.md); hive setup prints
// the same one. Hive names no other install command.
const context7Command = "npx ctx7@latest setup --cli"

// Texts of the Record column that name a state rather than a step status.
const (
	recordNone       = "No onboarding record yet"
	recordUnreadable = "record unreadable"
	recordNotOffered = "not in the last record"
	recordNotTracked = "not part of onboarding"
	recordUnknown    = "unknown"
)

// recordDamagedText tells what failed and how to recover when the last
// onboarding record cannot be read; the technical error follows on its own line.
const recordDamagedText = "The last integrations record is damaged or unreadable; running hive install again writes a new one."

// integrationRow is one capability. Every string is already sanitized.
type integrationRow struct {
	ID           string   // providers.ID, or "agent-browser"
	Name         string   // shown in the list
	Found        string   // list column: detected, not detected, not checked, not verified, cli, skill or cli + skill
	Evidence     []string // one line per local finding, with its path
	Record       string   // list column: the step status in the last onboarding record, or why there is none
	RecordDetail string   // the same, with the record's phase and time when there is one
	Source       string   // official documentation
	Next         string   // what to do next, in plain words, for this row's local finding and record
	// BrowserMissing is set for agent-browser when the program is on PATH but
	// none of the browsers it uses was found.
	BrowserMissing bool
}

// detailLines is the text the view shows under the list for the selected row
// and the section prints under each row.
func (r integrationRow) detailLines() []string {
	lines := []string{"Local evidence:"}
	for _, e := range r.Evidence {
		lines = append(lines, "  "+e)
	}
	// RecordDetail may hold a technical detail after a line break; it goes on
	// its own, indented line.
	record, detail := errLines(r.RecordDetail)
	lines = append(lines, "Last onboarding: "+record)
	for _, l := range detail {
		lines = append(lines, "  "+l)
	}
	return append(lines,
		"Source: "+r.Source,
		"Next step: "+r.Next,
	)
}

// summaryLine is the row's line in the section text. The columns are the
// list's: name, local finding, record status.
func (r integrationRow) summaryLine() string {
	return integrationColumns(r.Name, r.Found, r.Record)
}

func integrationColumns(name, found, record string) string {
	return fmt.Sprintf("%-20s%-14s%s", name, found, record)
}

// collectIntegrations builds the Integrations section. deps may or may not be
// adapted to the options already; o carries Home, StateDir and Scope.
func collectIntegrations(o management.Options, deps doctorDeps) doctorSection {
	rows, errText := collectIntegrationRows(o, deps)
	return integrationsSection(rows, errText)
}

// integrationsSection turns the rows into the section, so `hive doctor`
// prints what the view shows.
func integrationsSection(rows []integrationRow, errText string) doctorSection {
	sec := doctorSection{Title: "Integrations", Err: errText}
	for _, r := range rows {
		sec.Lines = append(sec.Lines, r.summaryLine())
		for _, l := range r.detailLines() {
			sec.Lines = append(sec.Lines, "  "+l)
		}
	}
	return sec
}

// collectIntegrationRows runs the local checks and reads the last onboarding
// record. errText is set when the record could not be read; the rows still
// carry the local findings.
func collectIntegrationRows(o management.Options, deps doctorDeps) (rows []integrationRow, errText string) {
	deps = deps.forOptions(o)
	if o.Scope == "" {
		o.Scope = "user"
	}
	synthetic := o.Home != ""
	rec, recErr := readLastOnboarding(o)
	if recErr != nil {
		errText = recordDamagedText + "\nDetail: cannot read the last onboarding record: " + sanitizeLine(recErr.Error())
	}

	skillRoots, homeErr := integrationSkillRoots(deps)

	engram := integrationRow{ID: string(providers.Engram), Name: "Engram"}
	engram.Found, engram.Evidence = programEvidence(deps, synthetic, "engram")

	ctx7 := integrationRow{ID: string(providers.Context7), Name: "Context7"}
	ctx7.Found, ctx7.Evidence = context7Evidence(deps, homeErr)

	pi := integrationRow{ID: string(providers.PiSubagents), Name: "pi-subagents", Found: "not checked",
		Evidence: []string{"Not checked: finding it would mean reading Pi's configuration, which Hive does not do."}}

	browser := integrationRow{ID: "agent-browser", Name: "agent-browser", Source: agentBrowserSource}
	browser.Found, browser.Evidence = agentBrowserEvidence(deps, synthetic, skillRoots, homeErr)
	browser.BrowserMissing = agentBrowserBrowserEvidence(deps, synthetic, browser.Found, &browser.Evidence)

	for _, r := range []*integrationRow{&engram, &ctx7, &pi, &browser} {
		fillSourceFromCatalog(r)
		rec.apply(r, recErr)
	}
	rows = []integrationRow{engram, ctx7, pi}
	// The Pi codemode row is built from constants only: it is neither in the
	// catalog nor in the onboarding record, and it never reads Pi's
	// configuration. Hive's own state decides whether Pi is registered; if
	// that cannot be read, the informative row still shows.
	hosts, hostsErr := management.RegisteredHosts(o)
	if hostsErr != nil || slices.Contains(hosts, "pi") {
		rows = append(rows, integrationRow{
			ID: "pi-codemode", Name: "Pi codemode", Found: "not checked",
			Evidence:     []string{"Not checked: finding it would mean reading Pi's configuration, which Hive does not do."},
			Record:       recordNotTracked,
			RecordDetail: "Not part of onboarding: Hive does not configure Pi's tools.",
			Source:       piCodemodeSource,
		})
	}
	rows = append(rows, browser)
	// The Codex question card row is built like the Pi codemode row: constants
	// and the onboarding record only. Finding the feature would mean reading
	// Codex's configuration, so no program is looked up and no file is read.
	if hostsErr != nil || slices.Contains(hosts, "codex") {
		card := integrationRow{
			ID: string(providers.CodexQuestionCard), Name: "Codex question card", Found: "not checked",
			Evidence: []string{"Not checked: finding it would mean reading Codex's configuration, which Hive does not do."},
		}
		fillSourceFromCatalog(&card)
		rec.apply(&card, recErr)
		rows = append(rows, card)
	}
	for i := range rows {
		rows[i].Next = nextStep(rows[i])
	}
	return rows, errText
}

// fillSourceFromCatalog copies the official source from providers.Catalog. The
// catalog's own reason is not shown: it says every install is pending
// validation, which tells a reader nothing about what to do.
func fillSourceFromCatalog(r *integrationRow) {
	for _, p := range providers.Catalog() {
		if string(p.ID) == r.ID {
			r.Source = sanitizeLine(p.Source)
			return
		}
	}
	r.Source = "-"
}

// stepStatusRank orders step statuses from best to worst for the Record column
// of a capability with several steps; an unlisted status ranks worst.
var stepStatusRank = map[string]int{
	management.StepVerified:    1,
	management.StepSkipped:     2,
	management.StepManual:      3,
	management.StepAuthPending: 4,
	management.StepPending:     5,
	management.StepRunning:     5,
	management.StepUnknown:     6,
	management.StepFailed:      7,
}

// recordSentences paraphrase a step status of the last onboarding record in
// plain words. Any other Record text (no record, unreadable, not offered) adds
// nothing to the next step.
var recordSentences = map[string]string{
	management.StepManual:      "The last install left this for you to finish by hand.",
	management.StepVerified:    "The last install recorded it as confirmed.",
	management.StepFailed:      "The last install could not finish it.",
	management.StepAuthPending: "The last install was waiting for you to sign in.",
	management.StepSkipped:     "The last install did not run it.",
	management.StepPending:     "The last install did not finish checking it.",
}

// nextStep says what to do about one row, from what Hive found locally and what
// the last record says. The row's Record is already set. Hive installs only
// agent-browser, through `hive install`; the other steps point to the official
// source, and the only vendor command is Context7's. The Pi codemode step
// names a line for Pi's own settings file and Pi's /reload, which are not Hive
// commands. The Source line is printed right above the step, so the step says
// "see Source above" instead of repeating the address.
func nextStep(r integrationRow) string {
	var step string
	switch r.ID {
	case string(providers.Context7):
		step = context7NextStep(r.Found)
	case string(providers.Engram):
		step = engramNextStep(r.Found)
	case string(providers.PiSubagents):
		step = "Hive cannot check it without reading Pi's configuration, which it does not do. If you use Pi, install it with Pi's own package manager; see Source above for the official instructions."
	case "pi-codemode":
		step = `Pi turns codemode on by itself when an MCP server uses codemode exposure, its default. To keep it on in every session, add "defaultTools": ["+codemode"] to Pi's settings.json (by default ~/.pi/agent/settings.json) and run /reload in Pi, or ask Pi to enable codemode; see Source above.`
	case string(providers.CodexQuestionCard):
		step = "Run `codex features enable " + providers.CodexQuestionCardFeature + "` to show Codex's question card in Default mode, or `codex features disable " + providers.CodexQuestionCardFeature + "` to turn it off; `codex features list` shows the current state."
	default:
		step = agentBrowserNextStep(r.Found, r.BrowserMissing)
	}
	if s, ok := recordSentences[r.Record]; ok {
		step += " " + s
	}
	return step
}

func context7NextStep(found string) string {
	const signIn = " (it may ask you to sign in)."
	switch found {
	case "detected":
		return "To refresh it, run " + context7Command + signIn
	case "not detected":
		return "To install it, run " + context7Command + signIn + " Context7 is optional."
	case "not verified":
		return "A skill file exists but could not be verified. To reinstall it, run " + context7Command + signIn
	}
	return "Hive could not look for it here. To install or refresh it, run " + context7Command + signIn
}

func engramNextStep(found string) string {
	switch found {
	case "detected":
		return "Hive does not configure Engram. For each CLI, follow Engram's official setup instructions (see Source above)."
	case "not detected":
		return "Install it from the official source (see Source above); Hive does not install it."
	}
	return "Hive could not look for it here. Hive does not install or configure it; see Source above for the official instructions."
}

func agentBrowserNextStep(found string, browserMissing bool) string {
	const official = "follow the official instructions (see Source above)."
	const noBrowser = " No browser was found: run hive install to download Chrome for Testing, or set AGENT_BROWSER_EXECUTABLE_PATH to an existing browser."
	switch found {
	case "cli + skill":
		if browserMissing {
			return "The program and its skill were found." + noBrowser + " Hive does not update the program; to update it, " + official
		}
		return "Found. Hive does not update it; to update it, " + official
	case "cli":
		if browserMissing {
			return "The program was found but no skill file." + noBrowser + " To add the skill, " + official
		}
		return "The program was found but no skill file. Optional. To add the skill, " + official
	case "skill":
		return "The skill file was found but not the program. Optional. To install the program, run hive install, or " + official
	case "not detected":
		return "Not found. Optional. To install the program, run hive install; it offers to install agent-browser and, if no browser is found, Chrome for Testing. Or " + official
	}
	return "Hive could not look for the program here. Optional. To install the program, run hive install, or " + official
}

// onboardingRecord is the last finished record, or why there is none.
type onboardingRecord struct {
	found  bool
	result management.OnboardingResult
	at     time.Time
}

func readLastOnboarding(o management.Options) (onboardingRecord, error) {
	_, stateDir, err := management.NormalizeOptions(o)
	if err != nil {
		return onboardingRecord{}, err
	}
	result, at, found, err := management.LastOnboarding(stateDir)
	if err != nil {
		return onboardingRecord{}, err
	}
	return onboardingRecord{found: found, result: result, at: at}, nil
}

// apply sets the row's record columns. A provider's steps are the ones whose ID
// starts with "<provider>-", the name the native adapter gives them; with
// several, the column shows the worst status.
func (rec onboardingRecord) apply(r *integrationRow, err error) {
	switch {
	case err != nil:
		r.Record = recordUnreadable
		r.RecordDetail = "The record is damaged or unreadable; running hive install again writes a new one.\nDetail: cannot read the last onboarding record: " + sanitizeLine(err.Error())
	case !rec.found:
		r.Record, r.RecordDetail = recordNone, recordNone
	default:
		r.Record = recordNotOffered
		worst := 0
		for _, s := range rec.result.Steps {
			if !strings.HasPrefix(s.Step.ID, r.ID+"-") {
				continue
			}
			rank := stepStatusRank[s.Status]
			if rank == 0 {
				rank = len(stepStatusRank) + 1
			}
			if rank > worst {
				worst = rank
				r.Record = sanitizeLine(s.Status)
			}
		}
		r.RecordDetail = fmt.Sprintf("%s (onboarding %s, %s)", r.Record, sanitizeLine(rec.result.Phase), rec.at.UTC().Format("2006-01-02 15:04 UTC"))
	}
}

// ---------------------------------------------------------------------------
// Local evidence.
// ---------------------------------------------------------------------------

// programEvidence looks name up in PATH. A synthetic home detects nothing, so
// it says the program was not checked rather than not found.
func programEvidence(deps doctorDeps, synthetic bool, name string) (found string, evidence []string) {
	if synthetic {
		return "not checked", []string{"Not checked: a synthetic --home never looks programs up in PATH."}
	}
	path, err := deps.lookPath(name)
	if err != nil {
		return "not detected", []string{name + " not found in PATH"}
	}
	return "detected", []string{"Program: " + sanitizeLine(path)}
}

// integrationSkillRoots are the roots `hive setup` checks for skills, taken
// from context7Candidates so the two never disagree: each candidate is
// <root>/skills/<name>/SKILL.md.
func integrationSkillRoots(deps doctorDeps) ([]string, error) {
	home, err := deps.userHome()
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var roots []string
	for _, p := range context7Candidates(home, deps.getenv) {
		root := filepath.Dir(filepath.Dir(filepath.Dir(p)))
		if !seen[root] {
			seen[root] = true
			roots = append(roots, root)
		}
	}
	sort.Strings(roots)
	return roots, nil
}

// skillEvidence checks each skill file. found are the usable ones and
// unverified describes the ones that exist but cannot be trusted as evidence.
func skillEvidence(paths []string) (found, unverified []string) {
	for _, p := range paths {
		info, err := os.Stat(p) // follows the native skill aliases without changing them
		if os.IsNotExist(err) {
			continue
		}
		switch {
		case err != nil:
			unverified = append(unverified, "Not verified: "+sanitizeLine(p)+" ("+sanitizeLine(err.Error())+")")
		case !info.Mode().IsRegular() || info.Size() == 0:
			unverified = append(unverified, "Not verified: "+sanitizeLine(p)+" (not a nonempty regular file)")
		default:
			found = append(found, "Skill file: "+sanitizeLine(p))
		}
	}
	return found, unverified
}

func context7Evidence(deps doctorDeps, homeErr error) (string, []string) {
	if homeErr != nil {
		return "not checked", []string{"Not checked: the home directory is unknown (" + sanitizeLine(homeErr.Error()) + ")."}
	}
	home, _ := deps.userHome()
	candidates := context7Candidates(home, deps.getenv)
	found, unverified := skillEvidence(candidates)
	evidence := append(found, unverified...)
	switch {
	case len(found) > 0:
		return "detected", evidence
	case len(unverified) > 0:
		return "not verified", evidence
	}
	return "not detected", []string{fmt.Sprintf("No find-docs or context7-mcp SKILL.md in the %d checked locations.", len(candidates))}
}

func agentBrowserEvidence(deps doctorDeps, synthetic bool, roots []string, homeErr error) (string, []string) {
	programFound, evidence := programEvidence(deps, synthetic, "agent-browser")
	cli := programFound == "detected"
	if homeErr != nil {
		return programFound, append(evidence, "Skill not checked: the home directory is unknown ("+sanitizeLine(homeErr.Error())+").")
	}
	paths := make([]string, len(roots))
	for i, root := range roots {
		paths[i] = filepath.Join(root, "skills", "agent-browser", "SKILL.md")
	}
	found, unverified := skillEvidence(paths)
	evidence = append(evidence, found...)
	evidence = append(evidence, unverified...)
	skill := len(found) > 0
	switch {
	case cli && skill:
		return "cli + skill", evidence
	case cli:
		if len(unverified) == 0 {
			evidence = append(evidence, fmt.Sprintf("No agent-browser SKILL.md in the %d checked locations.", len(paths)))
		}
		return "cli", evidence
	case skill:
		return "skill", evidence
	}
	if len(unverified) == 0 {
		evidence = append(evidence, fmt.Sprintf("No agent-browser SKILL.md in the %d checked locations.", len(paths)))
	}
	if programFound == "not checked" {
		return "not checked", evidence
	}
	return "not detected", evidence
}

// agentBrowserBrowserEvidence adds the browser finding to the agent-browser
// evidence and reports whether a browser is missing. It only reads PATH,
// environment and files, through the doctor's own dependencies, and only when
// the program was found and the home is real: without the program there is
// nothing to run a browser, and under a synthetic home nothing is checked.
func agentBrowserBrowserEvidence(deps doctorDeps, synthetic bool, found string, evidence *[]string) bool {
	if synthetic || (found != "cli" && found != "cli + skill") {
		return false
	}
	sys := agentBrowserSys
	sys.lookPath, sys.getenv, sys.userHome = deps.lookPath, deps.getenv, deps.userHome
	if ok, path := sys.browserFound(); ok {
		*evidence = append(*evidence, "Browser: "+sanitizeLine(path))
		return false
	}
	*evidence = append(*evidence, "No browser found; checked "+sanitizeLine(sys.browserCheckedPlaces())+".")
	return true
}
