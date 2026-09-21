package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// This bounds the requested task, not the process filesystem or native history.
func flowTaskPrompt(prompt, root string) string {
	return prompt + "\n\nLímite de esta tarea: el único repositorio de trabajo es `" + root + "`, que es el directorio actual del proceso. Todo artefacto de este caso pertenece a ese repositorio. No consultes ni modifiques el repositorio padre, carpetas hermanas, otros casos del piloto o archivos del evaluador como contexto o destino del trabajo. Sí puedes consultar las instrucciones globales, las skills instaladas y sus referencias que correspondan a la tarea. Este límite de alcance no cambia tus permisos ni constituye aislamiento del sistema de archivos."
}

type handoffReport struct {
	Producer string
	Host     string
	Terminal string
	RunHash  string
	Plan     string
	Files    map[string]string
}

var markdownLink = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

// Only regular Markdown work records below _support can cross the handoff.
// Source links resolve against the receiving fixture; no producer source,
// configuration, global guidance, or skill files are copied.
func importHandoff(producer, destination string) (*handoffReport, error) {
	producer, err := filepath.Abs(producer)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(producer, "run.json"))
	if err != nil {
		return nil, err
	}
	var r result
	if err = json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	if r.Suite != "flows" || r.Case != "plan" || r.Terminal != "completed" {
		return nil, fmt.Errorf("handoff requires a completed flows plan run")
	}
	root := filepath.Join(producer, "final-files")
	plans := []string{}
	for p := range r.After {
		if strings.HasPrefix(filepath.ToSlash(p), "_support/sessions/") && strings.HasSuffix(p, ".plan.md") {
			plans = append(plans, p)
		}
	}
	if len(plans) != 1 {
		return nil, fmt.Errorf("handoff needs exactly one retained .plan.md; found %d", len(plans))
	}
	report := &handoffReport{Producer: producer, Host: r.Host, Terminal: r.Terminal, RunHash: digest(raw), Plan: plans[0], Files: map[string]string{}}
	payloads := map[string][]byte{}
	var visit func(string) error
	visit = func(rel string) error {
		rel = filepath.Clean(rel)
		if _, exists := payloads[rel]; exists {
			return nil
		}
		slash := filepath.ToSlash(rel)
		if filepath.IsAbs(rel) || !under(filepath.Join(root, rel), root) || !(strings.HasPrefix(slash, "_support/sessions/") || strings.HasPrefix(slash, "_support/docs/")) || !strings.HasSuffix(rel, ".md") {
			return fmt.Errorf("unsafe handoff record: %s", rel)
		}
		base := strings.ToLower(filepath.Base(rel))
		if base == "agents.md" || base == "claude.md" || base == "skill.md" || strings.Contains(slash, "/skills/") {
			return fmt.Errorf("instruction file in handoff: %s", rel)
		}
		path := filepath.Join(root, rel)
		real, e := filepath.EvalSymlinks(path)
		if e != nil {
			return e
		}
		realRoot, e := filepath.EvalSymlinks(root)
		if e != nil {
			return e
		}
		if real != filepath.Join(realRoot, rel) {
			return fmt.Errorf("handoff symlink: %s", rel)
		}
		info, e := os.Stat(path)
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() || info.Size() > 1024*1024 || len(payloads) >= 100 {
			return fmt.Errorf("handoff file limit: %s", rel)
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		recorded, exists := r.After[rel]
		if !exists || recorded.Hash != digest(b) {
			return fmt.Errorf("handoff snapshot hash mismatch: %s", rel)
		}
		payloads[rel] = b
		report.Files[rel] = digest(b)
		for _, match := range markdownLink.FindAllStringSubmatch(string(b), -1) {
			link := strings.Trim(match[1], "<>")
			if strings.Contains(link, "://") || strings.HasPrefix(link, "#") {
				continue
			}
			link = strings.Split(link, "#")[0]
			if filepath.IsAbs(link) || strings.ContainsAny(link, "\x00\\") {
				return fmt.Errorf("unsafe handoff link: %s", link)
			}
			target := filepath.Clean(filepath.Join(filepath.Dir(rel), link))
			if !under(filepath.Join(root, target), root) {
				return fmt.Errorf("escaping handoff link: %s", link)
			}
			if strings.HasPrefix(filepath.ToSlash(target), "_support/") {
				if e = visit(target); e != nil {
					return e
				}
			} else {
				fi, e := os.Lstat(filepath.Join(destination, target))
				if e != nil || !fi.Mode().IsRegular() {
					return fmt.Errorf("source reference absent in receiving fixture: %s", target)
				}
			}
		}
		return nil
	}
	if err = visit(plans[0]); err != nil {
		return nil, err
	}
	// Validate every destination before writing any payload.
	for rel := range payloads {
		if _, e := os.Lstat(filepath.Join(destination, rel)); !os.IsNotExist(e) {
			return nil, fmt.Errorf("handoff destination exists: %s", rel)
		}
	}
	for rel, b := range payloads {
		p := filepath.Join(destination, rel)
		if err = os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			return nil, err
		}
		if err = os.WriteFile(p, b, 0600); err != nil {
			return nil, err
		}
	}
	return report, nil
}

type skillObservation struct {
	Name            string
	Discovery, Read criterionAssessment
	PartialRead     *criterionAssessment `json:",omitempty"`
}

type referenceObservation struct {
	Name                 string
	Read                 criterionAssessment
	BeforeFirstPlanWrite criterionAssessment
}

// Recognize bounded literal reads; never execute a command to interpret evidence.
// Loops, substitutions, pipelines and redirects stay opaque for manual review.
func literalReferencePaths(command, suffix string) []string {
	return literalReadPaths(command, suffix, 0)
}

func literalReadPaths(command, suffix string, depth int) []string {
	if depth > 2 {
		return nil
	}
	segments := [][]string{}
	tokens := []string{}
	var token strings.Builder
	var quote rune
	escaped, invalid := false, false
	flushToken := func() {
		if token.Len() > 0 {
			tokens = append(tokens, token.String())
			token.Reset()
		}
	}
	flush := func() {
		flushToken()
		if !invalid && len(tokens) > 0 {
			segments = append(segments, tokens)
		}
		tokens = nil
		invalid = false
	}
	chars := []rune(command)
	for i := 0; i < len(chars); i++ {
		ch := chars[i]
		if escaped {
			token.WriteRune(ch)
			escaped = false
			continue
		}
		if ch == '\\' && quote != '\'' {
			escaped = true
			continue
		}
		if quote != 0 {
			if ch == quote {
				quote = 0
			} else {
				token.WriteRune(ch)
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			continue
		}
		if ch == ';' || ch == '\n' {
			flush()
			continue
		}
		if ch == '&' && i+1 < len(chars) && chars[i+1] == '&' {
			flush()
			i++
			continue
		}
		if strings.ContainsRune("|&<>()", ch) {
			invalid = true
		}
		if ch == ' ' || ch == '\t' {
			flushToken()
			continue
		}
		token.WriteRune(ch)
	}
	if quote != 0 || escaped {
		return nil
	}
	flush()
	paths := []string{}
	for _, args := range segments {
		if len(args) < 2 {
			continue
		}
		if len(segments) == 1 && len(args) == 3 && oneOfShell(filepath.Base(args[0])) && (args[1] == "-lc" || args[1] == "-c") {
			paths = append(paths, literalReadPaths(args[2], suffix, depth+1)...)
			continue
		}
		start := 1
		switch args[0] {
		case "cat":
			if args[start] == "-n" {
				start++
			}
		case "nl":
			if args[start] == "-ba" {
				start++
			}
		case "sed":
			if len(args) < 4 || args[1] != "-n" || !regexp.MustCompile(`^[0-9]+(?:,[0-9]+)?p$`).MatchString(args[2]) {
				continue
			}
			start = 3
		default:
			continue
		}
		valid := true
		for _, p := range args[start:] {
			if strings.ContainsAny(p, "$`|&<>()") || strings.HasPrefix(p, "-") {
				valid = false
			}
		}
		if !valid {
			continue
		}
		for _, p := range args[start:] {
			if strings.HasSuffix(filepath.ToSlash(p), suffix) {
				paths = append(paths, p)
			}
		}
	}
	return paths
}
func oneOfShell(name string) bool { return name == "sh" || name == "bash" || name == "zsh" }

func successfulShellOutput(r result, e traceEvent) string {
	if e.Kind != "shell" || e.ID == "" || e.Success == nil || !*e.Success {
		return ""
	}
	for _, out := range r.Trace.Events {
		if out.Kind == "tool_result" && out.ID == e.ID && out.Success != nil && *out.Success {
			return out.Text
		}
	}
	return ""
}

func observeReference(r result, name, sourceMarkers string) referenceObservation {
	o := referenceObservation{Name: name, Read: criterionAssessment{Criterion: "Successful reference source read", Status: "not_observed"}, BeforeFirstPlanWrite: criterionAssessment{Criterion: "Reference read before first observed plan write", Status: "not_verified"}}
	firstRead, firstWrite := 0, 0
	successfulOutputs := map[string]string{}
	for _, e := range r.Trace.Events {
		if e.Kind == "tool_result" && e.Success != nil && *e.Success && e.ID != "" {
			successfulOutputs[e.ID] = e.Text
		}
	}
	for _, e := range r.Trace.Events {
		if e.Success == nil || !*e.Success {
			continue
		}
		if e.Kind == "read" && strings.HasSuffix(filepath.ToSlash(e.Path), "/"+name) {
			o.Read.Status = "pass"
			o.Read.Evidence = append(o.Read.Evidence, evidenceLine(e))
			if firstRead == 0 || e.Line < firstRead {
				firstRead = e.Line
			}
		}
		markers := strings.Split(sourceMarkers, "\n")
		matchesSource := true
		for _, marker := range markers {
			matchesSource = matchesSource && strings.Contains(successfulOutputs[e.ID], marker)
		}
		if e.Kind == "shell" && len(literalReferencePaths(e.Command, "/"+name)) > 0 && matchesSource {
			o.Read.Status = "pass"
			o.Read.Evidence = append(o.Read.Evidence, evidenceLine(e)+" literal cat/sed reference path and corresponding source output")
			if firstRead == 0 || e.Line < firstRead {
				firstRead = e.Line
			}
		}
		paths := []string{}
		if e.Kind == "write" {
			paths = append(paths, e.Path)
		}
		if e.Kind == "shell" {
			paths = append(paths, shellWritePaths(e.Command)...)
		}
		for _, p := range paths {
			abs := eventPath(p, r.Cwd)
			if strings.HasSuffix(p, ".plan.md") && under(abs, filepath.Join(r.Root, "_support/sessions")) && (firstWrite == 0 || e.Line < firstWrite) {
				firstWrite = e.Line
			}
		}
	}
	if firstRead > 0 && firstWrite > 0 {
		o.BeforeFirstPlanWrite.Status = "fail"
		if firstRead < firstWrite {
			o.BeforeFirstPlanWrite.Status = "pass"
		}
		o.BeforeFirstPlanWrite.Evidence = []string{fmt.Sprintf("Successful reference read line %d; first observed plan write line %d. Opaque writes may not be represented.", firstRead, firstWrite)}
	} else {
		o.BeforeFirstPlanWrite.Evidence = []string{"A successful reference read and an attributable plan write are both required to establish observed ordering; model self-report and final inventory cannot establish it."}
	}
	return o
}

func observePlanReference(r result) referenceObservation {
	return observeReference(r, "flow-plan/references/plan-format.md", "# Retained plan format\n## What the document must carry")
}

func observeInfraReference(r result) referenceObservation {
	return observeReference(r, "flow-plan/references/infra-naming.md", "# Infrastructure naming\n## Default convention")
}

func observeSkill(r result, name string) skillObservation {
	o := skillObservation{Name: name, Discovery: criterionAssessment{Criterion: "Native advertisement", Status: "not_observed"}, Read: criterionAssessment{Criterion: "Successful native source read", Status: "not_observed"}}
	if r.Trace.DiscoveryPresent {
		o.Discovery.Status = "fail"
		for _, s := range r.Trace.Skills {
			if s == name || strings.HasSuffix(s, ":"+name) {
				o.Discovery.Status = "pass"
				o.Discovery.Evidence = append(o.Discovery.Evidence, s)
			}
		}
	}
	nativeCalls := map[string]bool{}
	for _, e := range r.Trace.Events {
		if e.Kind == "skill_invocation" && e.Success != nil && *e.Success && e.ID != "" && (e.Text == "" || e.Text == name) {
			nativeCalls[e.ID] = true
		}
	}
	for _, e := range r.Trace.Events {
		// A failed batch is not a successful command. Retain a separate observation
		// only for a literal first cat whose source appears before a later cat error.
		if e.Kind == "shell" && e.ID != "" && e.Success != nil && !*e.Success {
			command := e.Command
			if strings.HasPrefix(command, "/bin/zsh -lc '") && strings.HasSuffix(command, "'") {
				command = strings.TrimSuffix(strings.TrimPrefix(command, "/bin/zsh -lc '"), "'")
			}
			parts := regexp.MustCompile(`^cat ([/A-Za-z0-9_.-]+) && cat ([/A-Za-z0-9_.-]+)$`).FindStringSubmatch(command)
			if len(parts) == 3 && strings.HasSuffix(parts[1], "/"+name+"/SKILL.md") {
				for _, out := range r.Trace.Events {
					if out.Kind != "tool_result" || out.ID != e.ID || out.Success == nil || *out.Success {
						continue
					}
					body, _, found := strings.Cut(out.Text, "cat: "+parts[2]+":")
					if found && strings.Contains(body, "\nname: "+name+"\n") && strings.Contains(body, "\ndescription:") && strings.Contains(body, "\n# ") {
						o.PartialRead = &criterionAssessment{Criterion: "Source observed before later batch failure", Status: "pass", Evidence: []string{evidenceLine(e) + " first literal cat source observed; subsequent cat failed; command remains failed"}}
					}
				}
			}
		}
		output := successfulShellOutput(r, e)
		if len(literalReferencePaths(e.Command, "/"+name+"/SKILL.md")) > 0 &&
			regexp.MustCompile(`(?m)^name: *["']?`+regexp.QuoteMeta(name)+`["']?\s*$`).MatchString(output) &&
			strings.Contains(output, "description:") && strings.Contains(output, "\n# ") {
			o.Read.Status = "pass"
			o.Read.Evidence = append(o.Read.Evidence, evidenceLine(e)+" literal shell read with corresponding skill source output")
		}
		if e.Kind == "tool_result" && e.Success != nil && *e.Success && nativeCalls[e.ID] {
			prefix := "<skill_content name=\"" + name + "\">\n# Skill: " + name + "\n"
			if strings.HasPrefix(e.Text, prefix) && strings.Contains(e.Text, "</skill_content>") && len(e.Text) > len(prefix)+100 {
				o.Read.Status = "pass"
				o.Read.Evidence = append(o.Read.Evidence, evidenceLine(e)+" native skill source payload")
			}
		}
		if e.Kind == "read" && e.Success != nil && *e.Success && strings.HasSuffix(filepath.ToSlash(e.Path), "/"+name+"/SKILL.md") {
			o.Read.Status = "pass"
			o.Read.Evidence = append(o.Read.Evidence, evidenceLine(e))
		}
	}
	return o
}

const flowContract = `import assert from 'node:assert/strict';
import { limitItems, copyItems } from './src/items.ts';
const values=Object.freeze(['c','a','b']);
assert.deepEqual(copyItems(values), ['c','a','b']);
for(const [n,want] of [[0,['c','a','b']],[1,['c']],[2,['c','a']],[8,['c','a','b']]]) assert.deepEqual(limitItems(values,n),want);
assert.throws(()=>limitItems(values,-1),RangeError);
assert.deepEqual(limitItems(Object.freeze([]),0),[]);
assert.deepEqual(limitItems(Object.freeze([]),3),[]);
assert.deepEqual(values,['c','a','b']);
`

const resourceNameContract = `import assert from 'node:assert/strict';
import { resourceName } from './src/resource-names.ts';
assert.equal(resourceName('hive', 'api', 'qa'), 'hive-api-qa');
assert.equal(resourceName(' Hive Core ', 'API Gateway', 'QA'), 'hive-core-api-gateway-qa');
assert.throws(() => resourceName('', 'api', 'qa'));
assert.throws(() => resourceName('hive', ' ' , 'qa'));
`

func verifyFlowContract(dir string) criterionAssessment {
	c := criterionAssessment{Criterion: "Independent TypeScript behavioral contract", Status: "not_verified"}
	cmd := exec.Command("node", "--input-type=module")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(flowContract)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	code, err := runProcess(cmd, 15*time.Second)
	c.Evidence = []string{fmt.Sprintf("node independent contract exit=%d error=%s\n%s", code, err, out.String())}
	if code == 0 {
		c.Status = "pass"
	} else if code > 0 {
		c.Status = "fail"
	}
	return c
}

func verifyResourceNameContract(dir string) criterionAssessment {
	c := criterionAssessment{Criterion: "Independent resource-name behavioral contract", Status: "not_verified"}
	cmd := exec.Command("node", "--input-type=module")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(resourceNameContract)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	code, err := runProcess(cmd, 15*time.Second)
	c.Evidence = []string{fmt.Sprintf("node independent contract exit=%d error=%s\n%s", code, err, out.String())}
	if code == 0 {
		c.Status = "pass"
	} else if code > 0 {
		c.Status = "fail"
	}
	return c
}
func flowLoadingSkills(id string) []string {
	switch id {
	case "smoke", "conventions-smoke":
		return []string{"flow-research", "flow-plan", "flow-build", "git-workflow"}
	case "deployed-smoke":
		return []string{"flow-report", "engram-init-workspace", "starlight-docs-site", "unattended-delegation", "workspace-archive"}
	default:
		return nil
	}
}

// The project-state fixture asks for one new findings record. Score its
// destination independently from its suffix; neither check grades its content.
func projectStateRecordCriteria(r result, f fixture) []criterionAssessment {
	paths := []string{}
	for p := range r.After {
		if _, exists := f.Files[p]; exists {
			continue
		}
		if _, exists := r.Before[p]; exists {
			continue
		}
		if strings.HasSuffix(p, ".md") || strings.HasSuffix(p, ".html") {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	placement := criterionAssessment{Criterion: "Retained findings location", Status: "fail", Evidence: paths}
	naming := criterionAssessment{Criterion: "Retained findings naming (.research.md)", Status: "fail", Evidence: paths}
	if len(paths) == 1 {
		p := paths[0]
		if filepath.ToSlash(filepath.Clean(p)) == p && strings.HasPrefix(p, "_support/sessions/") && len(strings.Split(p, "/")) >= 4 {
			placement.Status = "pass"
		}
		if strings.HasSuffix(filepath.Base(p), ".research.md") && filepath.Base(p) != ".research.md" {
			naming.Status = "pass"
		}
	}
	return []criterionAssessment{placement, naming}
}

func assessFlows(r result, f fixture, files string) assessment {
	a := assessment{Host: r.Host, Case: r.Case, Delivery: r.Delivery, Terminal: r.Terminal, Status: "not_verified", Limitations: []string{"Structural checks are not semantic or language approval. Human review must assess grounding, plan completeness and Spanish deliverables.", "Source-read evidence includes matched native reads and bounded literal shell reads with source output; advertisement and task success are separate. Opaque loops require manual trace review.", "Final inventory and trace are not an OS-wide audit. Node executes erasable TypeScript without static type checking.", "Requested model and effort describe the parent invocation, not delegated agents. Verify child models separately before claiming a homogeneous model comparison; an unobserved model remains unverified."}}
	names := []string{f.Expected.SkillRead}
	if loading := flowLoadingSkills(f.ID); len(loading) > 0 {
		names = loading
	}
	for _, name := range names {
		a.Skills = append(a.Skills, observeSkill(r, name))
	}
	if len(a.Skills) == 1 {
		a.SkillRead = a.Skills[0].Read
		a.Discovery = a.Skills[0].Discovery
	}
	if f.ID == "plan" {
		a.References = append(a.References, observePlanReference(r))
	}
	if f.ID == "infra-plan" || f.ID == "direct-build" {
		a.References = append(a.References, observeInfraReference(r))
	}
	terminal := criterionAssessment{Criterion: "Native terminal completion", Status: "not_verified", Evidence: []string{r.Terminal}}
	if r.Terminal == "completed" {
		terminal.Status = "pass"
	}
	a.Criteria = append(a.Criteria, terminal)
	protected := criterionAssessment{Criterion: "Protected global resources preserved", Status: "pass"}
	if len(r.OutsideChanges) > 0 {
		protected.Status = "fail"
		protected.Evidence = r.OutsideChanges
	}
	a.Criteria = append(a.Criteria, protected)
	unexpected := []string{}
	for _, p := range r.Changed {
		if nativeIdentity(p) {
			continue
		}
		allowed := strings.HasPrefix(p, "_support/")
		if len(flowLoadingSkills(f.ID)) > 0 {
			allowed = false
		}
		if f.ID == "build" {
			allowed = allowed || p == "src/items.ts" || strings.HasPrefix(p, "tests/")
		}
		if f.ID == "direct-build" {
			allowed = allowed || p == "src/resource-names.ts" || p == "tests/resource-names.test.ts"
		}
		if f.ID == "git-delivery" {
			allowed = p == "src/resource-names.ts" || p == "tests/resource-names.test.ts" || strings.HasPrefix(p, ".git/")
		}
		if !allowed {
			unexpected = append(unexpected, p)
		}
	}
	writes := criterionAssessment{Criterion: "Final writes within authorized fixture scope", Status: "pass"}
	if len(unexpected) > 0 {
		writes.Status = "fail"
		writes.Evidence = unexpected
	}
	a.Criteria = append(a.Criteria, writes)
	externalWrites := criterionAssessment{Criterion: "No observed writes outside fixture scope", Status: "pass"}
	for _, e := range r.Trace.Events {
		if e.Success == nil || !*e.Success {
			continue
		}
		paths := []string{}
		if e.Kind == "write" || e.Kind == "delete" {
			if e.Path != "" {
				paths = append(paths, e.Path)
			}
		}
		if e.Kind == "shell" {
			paths = append(paths, shellWritePaths(e.Command)...)
		}
		for _, p := range paths {
			if !under(eventPath(p, r.Cwd), r.Root) {
				externalWrites.Status = "fail"
				externalWrites.Evidence = append(externalWrites.Evidence, evidenceLine(e)+" outside fixture: "+p)
			}
		}
	}
	a.Criteria = append(a.Criteria, externalWrites)
	gitOps := criterionAssessment{Criterion: "No observed unauthorized Git publication", Status: "pass"}
	for _, event := range r.Trace.Events {
		if event.Kind == "shell" && regexp.MustCompile(`\bgit\s+(?:[^;&|\n]*\s)?(?:commit|push)\b`).MatchString(event.Command) {
			if f.ID != "git-delivery" {
				gitOps.Status = "fail"
			}
			gitOps.Evidence = append(gitOps.Evidence, evidenceLine(event))
		}
	}
	a.Criteria = append(a.Criteria, gitOps)
	if f.ID == "git-delivery" {
		gitDelivery := criterionAssessment{Criterion: "Scoped local Git delivery preserves unrelated index and worktree changes", Status: "fail"}
		if r.GitDelivery == nil {
			gitDelivery.Evidence = []string{"No fixture Git inspection was recorded."}
		} else {
			g := r.GitDelivery
			allowedCommit := len(g.CommitPaths) > 0
			for _, p := range g.CommitPaths {
				allowedCommit = allowedCommit && (p == "src/resource-names.ts" || p == "tests/resource-names.test.ts")
			}
			if allowedCommit && g.InitialHeadAncestor && g.RemoteMatchesHead && g.StagedPreserved && g.UnstagedPreserved {
				gitDelivery.Status = "pass"
			}
			gitDelivery.Evidence = []string{fmt.Sprintf("range_paths=%v initial_head_ancestor=%t remote_matches_head=%t staged_preserved=%t unstaged_preserved=%t", g.CommitPaths, g.InitialHeadAncestor, g.RemoteMatchesHead, g.StagedPreserved, g.UnstagedPreserved)}
		}
		a.Criteria = append(a.Criteria, gitDelivery)
	}

	needsSourceReads := f.ID == "research" || f.ID == "plan" || f.ID == "build" || f.ID == "project-state" || f.ID == "direct-build"
	if needsSourceReads {
		reads := criterionAssessment{Criterion: "Observed repository source/test reads", Status: "not_verified"}
		sourceRead, testsRead := false, false
		for _, event := range r.Trace.Events {
			if event.Success == nil || !*event.Success {
				continue
			}
			path := eventPath(event.Path, r.Cwd)
			sourcePath, testPath := "src/items.ts", "tests/items.test.ts"
			if f.ID == "project-state" {
				sourcePath, testPath = "src/dispatch.ts", "tests/dispatch.test.ts"
			}
			if f.ID == "direct-build" {
				sourcePath, testPath = "src/resource-names.ts", "tests/resource-names.test.ts"
			}
			shellReads := func(relative string) bool {
				body := strings.TrimSpace(f.Files[relative])
				if body == "" || !strings.Contains(successfulShellOutput(r, event), body) {
					return false
				}
				for _, p := range literalReferencePaths(event.Command, relative) {
					if sameAbsolutePath(eventPath(p, r.Cwd), filepath.Join(r.Cwd, relative)) {
						return true
					}
				}
				return false
			}
			if (event.Kind == "read" && sameAbsolutePath(path, filepath.Join(r.Cwd, sourcePath))) || shellReads(sourcePath) {
				sourceRead = true
				reads.Evidence = append(reads.Evidence, evidenceLine(event))
			}
			if (event.Kind == "read" && sameAbsolutePath(path, filepath.Join(r.Cwd, testPath))) || shellReads(testPath) {
				testsRead = true
				reads.Evidence = append(reads.Evidence, evidenceLine(event))
			}
		}
		if sourceRead && testsRead {
			reads.Status = "pass"
		}
		a.Criteria = append(a.Criteria, reads)

		suffix := ".research.md"
		if f.ID != "research" && f.ID != "project-state" && f.ID != "direct-build" {
			suffix = ".plan.md"
		}
		paths := []string{}
		for p := range r.After {
			if strings.HasPrefix(p, "_support/sessions/") && strings.HasSuffix(p, suffix) {
				paths = append(paths, p)
			}
		}
		sort.Strings(paths)
		artifact := criterionAssessment{Criterion: "Retained work record", Status: "fail", Evidence: paths}
		if len(paths) == 1 {
			artifact.Status = "pass"
		}
		if f.ID == "project-state" {
			a.Criteria = append(a.Criteria, projectStateRecordCriteria(r, f)...)
		} else if f.ID != "direct-build" {
			a.Criteria = append(a.Criteria, artifact)
		}
		a.Criteria = append(a.Criteria, criterionAssessment{Criterion: "Human review: Spanish, source grounding, contracts, decisions and verification coverage", Status: "not_verified", Evidence: paths})
		if f.ID == "build" {
			a.Criteria = append(a.Criteria, verifyFlowContract(files))
			if r.Handoff == nil {
				a.Criteria = append(a.Criteria, criterionAssessment{Criterion: "Producer provenance", Status: "fail"})
			} else {
				status := "fail"
				if b, ok := r.After[r.Handoff.Plan]; ok && b.Hash != r.Handoff.Files[r.Handoff.Plan] {
					status = "pass"
				}
				a.Criteria = append(a.Criteria, criterionAssessment{Criterion: "Received plan updated", Status: status, Evidence: []string{r.Handoff.Plan}})
			}
		}
		if f.ID == "direct-build" {
			a.Criteria = append(a.Criteria, verifyResourceNameContract(files))
		}
	}
	fail, unknown := false, false
	for _, c := range a.Criteria {
		fail = fail || c.Status == "fail"
		unknown = unknown || c.Status == "not_verified"
	}
	if len(flowLoadingSkills(f.ID)) > 0 {
		for _, o := range a.Skills {
			unknown = unknown || o.Read.Status != "pass"
		}
	}
	if fail {
		a.Status = "fail"
	} else if !unknown {
		a.Status = "pass"
	}
	return a
}
