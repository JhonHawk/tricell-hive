package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type criterionAssessment struct {
	Criterion string   `json:"criterion"`
	Status    string   `json:"status"`
	Evidence  []string `json:"evidence"`
}
type assessment struct {
	Host, Case, Delivery, Terminal string
	Status                         string
	Criteria                       []criterionAssessment
	SkillRead                      criterionAssessment
	Discovery                      criterionAssessment
	Limitations                    []string
}

func assessRun(dir, source string) error {
	b, err := os.ReadFile(filepath.Join(dir, "run.json"))
	if err != nil {
		return err
	}
	var r result
	if err = json.Unmarshal(b, &r); err != nil {
		return err
	}
	var f fixture
	if b, err = os.ReadFile(filepath.Join(dir, "fixture.json")); err == nil {
		if err = json.Unmarshal(b, &f); err != nil {
			return err
		}
	} else {
		b, err = os.ReadFile(filepath.Join(source, "tests/fixtures/workspace-conventions/cases.json"))
		if err != nil {
			return err
		}
		var corpus struct {
			Cases []fixture `json:"cases"`
		}
		if err = json.Unmarshal(b, &corpus); err != nil {
			return err
		}
		for _, c := range corpus.Cases {
			if c.ID == r.Case {
				f = c
			}
		}
	}
	if f.ID == "" {
		return fmt.Errorf("no fixture criteria for %q", r.Case)
	}
	// Reparse raw evidence, including historical runs, without changing run.json.
	trace, err := os.Open(filepath.Join(dir, "stdout.jsonl"))
	if err != nil {
		return err
	}
	r.Trace = parseTrace(r.Host, trace)
	trace.Close()
	r.Terminal = terminalState(r.Trace, r.ExitCode, r.Error)
	a := assessResult(r, f, filepath.Join(dir, "final-files"))
	b, err = json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(dir, "criterion-assessment.json")
	file, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = file.Write(append(b, '\n'))
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	fmt.Printf("ASSESS %s %s %s %s\n", r.Host, r.Case, a.Status, p)
	return nil
}

func under(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
func eventPath(path, cwd string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(cwd, path)
}
func nativeIdentity(path string) bool {
	return path == ".git/engram-project-identity.json" || strings.HasSuffix(path, "/.git/engram-project-identity.json")
}
func evidenceLine(e traceEvent) string {
	return fmt.Sprintf("stdout.jsonl:%d %s %s", e.Line, e.Kind, e.Path)
}

// Deliberately bounded recognition, never a shell interpreter. Other shell
// writes remain opaque and require manual review even if the final tree is clean.
var redirectPath = regexp.MustCompile(`(?:^|[\s;])(?:[12]?>>?)\s*["']?([^\s"'<>;&|]+)`)
var pythonWritePath = regexp.MustCompile(`(?:open|Path)\(\s*["']([^"']+)["']\s*(?:,\s*["'][wax][^"']*["']|\)\.(?:write_text|write_bytes)\()`)

func shellWritePaths(command string) []string {
	paths := []string{}
	for _, re := range []*regexp.Regexp{redirectPath, pythonWritePath} {
		for _, m := range re.FindAllStringSubmatch(command, -1) {
			if m[1] != "/dev/null" && !strings.ContainsAny(m[1], "$`*") {
				paths = append(paths, m[1])
			}
		}
	}
	return paths
}

func assessResult(r result, f fixture, files string) assessment {
	a := assessment{Host: r.Host, Case: r.Case, Delivery: r.Delivery, Terminal: r.Terminal, Status: "manual_review", SkillRead: criterionAssessment{Criterion: f.Expected.SkillRead, Status: "not_required"}, Limitations: []string{"Tool traces and bounded file inventories are not an OS-wide write audit.", "Opaque shell operations and semantic/language judgments require human review; model self-report is not source-read evidence.", "Daily plugins, MCP, and external memory remain active; this is an observational screen."}}
	a.Discovery = criterionAssessment{Criterion: "Native workspace-conventions advertisement", Status: "not_observed"}
	if r.Trace.DiscoveryPresent {
		a.Discovery.Status = "fail"
		for _, name := range r.Trace.Skills {
			if name == "workspace-conventions" || strings.HasSuffix(name, ":workspace-conventions") {
				a.Discovery.Status = "pass"
				a.Discovery.Evidence = []string{"Native system/init.skills: " + name}
			}
		}
	}
	if f.ID == "smoke" {
		a.Status = "unscored"
		a.SkillRead.Status = "not_observed"
		for _, e := range r.Trace.Events {
			if e.Kind == "read" && e.Success != nil && *e.Success && strings.HasSuffix(filepath.ToSlash(e.Path), "/workspace-conventions/SKILL.md") {
				a.SkillRead.Status = "pass"
				a.SkillRead.Evidence = append(a.SkillRead.Evidence, evidenceLine(e))
			}
		}
		return a
	}
	for _, c := range f.Expected.Criteria {
		a.Criteria = append(a.Criteria, criterionAssessment{Criterion: c, Status: "not_observed", Evidence: []string{"Manual review required."}})
	}
	set := func(i int, status string, evidence ...string) {
		if i < len(a.Criteria) {
			a.Criteria[i].Status = status
			a.Criteria[i].Evidence = evidence
		}
	}
	check := func(i int, yes bool, evidence ...string) {
		status := "fail"
		if yes {
			status = "pass"
		}
		set(i, status, evidence...)
	}
	same := func(p string) bool {
		b, exists := r.Before[p]
		after, present := r.After[p]
		return exists && present && b == after
	}
	exists := func(p string) bool { _, ok := r.After[p]; return ok }
	read := func(p string) string {
		if filepath.IsAbs(p) || !under(filepath.Join(files, p), files) {
			return ""
		}
		b, _ := os.ReadFile(filepath.Join(files, p))
		return string(b)
	}
	changed := []string{}
	for p, v := range r.After {
		if old, ok := r.Before[p]; !ok || old != v {
			changed = append(changed, p)
		}
	}
	for p := range r.Before {
		if _, ok := r.After[p]; !ok {
			changed = append(changed, p)
		}
	}
	sort.Strings(changed)
	opaque := false
	prohibitedGit := false
	firstMutation := 0
	skillRead := 0
	for _, e := range r.Trace.Events {
		if e.Kind == "shell" {
			opaque = true
			if regexp.MustCompile(`\bgit\s+(?:[^;&|\n]*\s)?(?:init|commit|push)\b`).MatchString(e.Command) {
				prohibitedGit = true
			}
		}
		if e.Kind == "write" || e.Kind == "delete" || e.Kind == "shell" {
			if firstMutation == 0 {
				firstMutation = e.Line
			}
		}
		if e.Kind == "read" && e.Success != nil && *e.Success && strings.HasSuffix(filepath.ToSlash(e.Path), "/workspace-conventions/SKILL.md") {
			if skillRead == 0 {
				skillRead = e.Line
			}
		}
	}
	if strings.Contains(f.Expected.SkillRead, "required_before") && (r.Delivery == "deployed-global" || r.Arm == "B") {
		a.SkillRead.Status = "not_observed"
		if skillRead > 0 && (firstMutation == 0 || skillRead < firstMutation) {
			a.SkillRead.Status = "pass"
			a.SkillRead.Evidence = []string{fmt.Sprintf("Successful SKILL.md read at line %d; first possible mutation line %d.", skillRead, firstMutation)}
		} else if skillRead > 0 {
			a.SkillRead.Status = "fail"
			a.SkillRead.Evidence = []string{"Observed skill read follows a possible mutation."}
		}
	}
	noUnexpected := func(i int, allowed func(string) bool) {
		bad := []string{}
		for _, p := range changed {
			if !nativeIdentity(p) && !allowed(p) {
				bad = append(bad, p)
			}
		}
		for _, e := range r.Trace.Events {
			if (e.Kind == "write" || e.Kind == "delete") && e.Path != "" && e.Success != nil && *e.Success {
				abs := eventPath(e.Path, r.Cwd)
				rel, _ := filepath.Rel(r.Root, abs)
				if !nativeIdentity(rel) && (!under(abs, r.Root) || !allowed(rel)) {
					bad = append(bad, evidenceLine(e))
				}
			}
		}
		if prohibitedGit {
			bad = append(bad, "trace contains git init/commit/push")
		}
		if len(bad) > 0 {
			set(i, "fail", bad...)
			return
		}
		if opaque {
			set(i, "not_observed", "Final changed paths are allowed; shell operations need manual review for transient/unrelated writes.")
			return
		}
		set(i, "pass", "All observed changed paths are allowed; no prohibited Git operation is visible.")
	}
	switch f.ID {
	case "placement":
		paths := []string{}
		for p := range r.After {
			if strings.HasPrefix(p, "frontend/_support/sessions/") && strings.HasSuffix(p, ".md") {
				paths = append(paths, p)
			}
		}
		sort.Strings(paths)
		if len(paths) == 0 {
			set(0, "fail", "No report under frontend/_support/sessions.")
		} else {
			set(0, "not_observed", append(paths, "Verify initial date/work folder and Spanish language manually.")...)
		}
		noUnexpected(2, func(p string) bool { return strings.HasPrefix(p, "frontend/") })
		noUnexpected(3, func(p string) bool { return strings.HasPrefix(p, "frontend/_support/sessions/") })
	case "continuation":
		record := "_support/sessions/2026-09-18-navigation/navigation-findings.md"
		if same(record) || !exists(record) {
			set(0, "fail", "Existing findings record was not updated.")
		}
		newSession := false
		for p := range r.After {
			if strings.HasPrefix(p, "_support/sessions/") && p != record {
				if _, had := r.Before[p]; !had {
					newSession = true
				}
			}
		}
		check(1, exists(record) && !newSession, "Original session path: "+record)
		check(2, same("src/navigation.json"), "Before/after hash of src/navigation.json")
		noUnexpected(3, func(p string) bool { return p == record })
	case "archive":
		old := "_support/sessions/2026-09-18-navigation/navigation-findings.md"
		moved := "_support/sessions/archived/2026-09-18-navigation/navigation-findings.md"
		open := "_support/sessions/2026-08-01-api/api-findings.md"
		evidence := "_support/evidence/2026-09-18-navigation/response.txt"
		index := "_support/sessions/index.md"
		check(0, !exists(old) && exists(moved), old+" -> "+moved)
		check(1, same(open) && same(evidence), "Open session and evidence path/content hashes")
		link := regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
		targets := func(p string) []string {
			out := []string{}
			for _, m := range link.FindAllStringSubmatch(read(p), -1) {
				out = append(out, filepath.ToSlash(filepath.Clean(filepath.Join(filepath.Dir(p), m[1]))))
			}
			sort.Strings(out)
			return out
		}
		idx, report := targets(index), targets(moved)
		want := []string{moved, open}
		sort.Strings(want)
		check(2, strings.Join(idx, "\n") == strings.Join(want, "\n") && len(report) == 1 && report[0] == evidence, "Resolved local Markdown targets in index and moved findings")
		normalize := func(s string) string {
			return regexp.MustCompile(`(\[[^\]]*\])\([^)]+\)`).ReplaceAllString(s, "${1}(link)")
		}
		check(3, exists(moved) && normalize(read(moved)) == normalize(f.Files[old]), "Findings text compared with only Markdown links normalized")
		noUnexpected(4, func(p string) bool { return p == old || p == moved || p == index })
	case "scratch":
		scratch := filepath.Join(r.Cwd, "_support/workspace/2026-09-18-navigation")
		created := map[string]bool{}
		misplaced := []string{}
		saw55 := false
		anyUse := false
		for _, e := range r.Trace.Events {
			if (e.Kind == "text" || e.Kind == "final" || e.Kind == "tool_result") && regexp.MustCompile(`\b55\b`).MatchString(e.Text) {
				saw55 = true
			}
			if e.Kind == "read" && e.Success != nil && *e.Success && created[eventPath(e.Path, r.Cwd)] {
				anyUse = true
			}
			paths := []string{}
			if e.Kind == "write" && e.Path != "" {
				paths = append(paths, e.Path)
			}
			if e.Kind == "shell" {
				paths = append(paths, shellWritePaths(e.Command)...)
			}
			if e.Success == nil || !*e.Success {
				continue
			}
			for _, p := range paths {
				abs := eventPath(p, r.Cwd)
				rel, _ := filepath.Rel(r.Root, abs)
				if _, prior := r.Before[rel]; prior {
					continue
				}
				created[abs] = true
				if !under(abs, scratch) {
					misplaced = append(misplaced, fmt.Sprintf("%s creates %s outside existing scratch", evidenceLine(e), abs))
				}
			}
		}
		if len(misplaced) > 0 {
			set(0, "fail", misplaced...)
		} else if len(created) == 0 {
			set(0, "not_observed", "No successful task-created temporary path is observable.")
		} else {
			set(0, "not_observed", fmt.Sprintf("Observed %d candidate temporary path(s), result55=%t, later read=%t; verify exact contents and computation/use in the trace.", len(created), saw55, anyUse))
		}
		remaining := []string{}
		for p := range created {
			rel, _ := filepath.Rel(r.Root, p)
			if exists(rel) {
				remaining = append(remaining, rel)
			}
		}
		if len(remaining) > 0 {
			set(1, "fail", remaining...)
		} else if len(created) == 0 || opaque {
			set(1, "not_observed", "Clean final inventory cannot establish cleanup of opaque or unobserved temporaries.")
		} else {
			set(1, "pass", "Every observed task-created temporary is absent in the terminal inventory.")
		}
		check(2, same("_support/workspace/2026-09-18-navigation/prior-notes.txt") && same("_support/workspace/2026-09-18-navigation/unique-failure.txt"), "Both preexisting scratch path/content hashes")
		noUnexpected(3, func(p string) bool {
			return strings.HasPrefix(p, "_support/workspace/2026-09-18-navigation/") || p == "_support/sessions/2026-09-18-navigation/navigation-findings.md"
		})
	}
	fail, unknown := false, false
	for _, c := range a.Criteria {
		fail = fail || c.Status == "fail"
		unknown = unknown || c.Status == "not_observed"
	}
	fail = fail || a.SkillRead.Status == "fail" || len(r.OutsideChanges) > 0
	unknown = unknown || a.SkillRead.Status == "not_observed" || r.Terminal != "completed"
	if len(r.OutsideChanges) > 0 {
		a.Limitations = append(a.Limitations, "Protected paths changed; attribution/reconciliation required: "+strings.Join(r.OutsideChanges, ", "))
	}
	if fail {
		a.Status = "fail"
	} else if unknown {
		a.Status = "manual_review"
	} else {
		a.Status = "pass"
	}
	return a
}
