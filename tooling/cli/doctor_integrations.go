// doctor_integrations.go builds the Integrations section of the diagnostics
// (design.md "Reglas por sección"): for each optional capability, what Hive
// can see locally, what the last onboarding record says, and where the
// official instructions are. It only reads. It never runs an integration
// program (a program is only looked up in PATH) and never reads a host's
// configuration: whether Pi has pi-subagents, or a host loads a skill, is not
// something Hive checks.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

// agentBrowserSource is where agent-browser's own instructions live. The tool
// is not in providers.Catalog, because Hive never offers to install it, so its
// source is fixed here (Context7 resolves the library to /vercel-labs/agent-browser).
const (
	agentBrowserSource = "github.com/vercel-labs/agent-browser"
	agentBrowserNext   = "Optional and not installed by Hive; follow the official instructions to install the CLI and its skill."
)

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
	Next         string   // reason or next step
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
	return fmt.Sprintf("%-14s%-14s%s", name, found, record)
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

	browser := integrationRow{ID: "agent-browser", Name: "agent-browser", Source: agentBrowserSource, Next: agentBrowserNext}
	browser.Found, browser.Evidence = agentBrowserEvidence(deps, synthetic, skillRoots, homeErr)

	for _, r := range []*integrationRow{&engram, &ctx7, &pi} {
		fillFromCatalog(r)
		rec.apply(r, recErr)
	}
	browser.Record, browser.RecordDetail = recordNotTracked, "Not part of onboarding: Hive does not offer this tool."
	return []integrationRow{engram, ctx7, pi, browser}, errText
}

// fillFromCatalog copies the official source and the reason from
// providers.Catalog.
func fillFromCatalog(r *integrationRow) {
	for _, p := range providers.Catalog() {
		if string(p.ID) == r.ID {
			r.Source, r.Next = sanitizeLine(p.Source), sanitizeLine(p.Reason)
			return
		}
	}
	r.Source, r.Next = "-", "-"
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

// apply sets the row's record columns. A provider's step is the one whose ID
// starts with "<provider>-", the name the native adapter gives it.
func (rec onboardingRecord) apply(r *integrationRow, err error) {
	switch {
	case err != nil:
		r.Record = recordUnreadable
		r.RecordDetail = "The record is damaged or unreadable; running hive install again writes a new one.\nDetail: cannot read the last onboarding record: " + sanitizeLine(err.Error())
	case !rec.found:
		r.Record, r.RecordDetail = recordNone, recordNone
	default:
		r.Record = recordNotOffered
		for _, s := range rec.result.Steps {
			if strings.HasPrefix(s.Step.ID, r.ID+"-") {
				r.Record = sanitizeLine(s.Status)
				break
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
