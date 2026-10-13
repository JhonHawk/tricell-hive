package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const flashSuite = "flash-class-guidance"

var flashModels = []string{"opencode-go/claude-haiku-5-5#high", "opencode-go/deepseek-v4.1-flash#high", "opencode-go/qwen3.8-flash#medium", "opencode-go/glm-5.3-flash#high"}

func flashRole(caseID string) (string, string) {
	switch caseID {
	case "plan-reference":
		return "hive-review-plan", "flow-plan/references/plan-review.md"
	case "harness-reference":
		return "hive-review-harness", "harness-audit/SKILL.md"
	}
	return "", ""
}

func flashProtectedPaths(g *guidanceVariant, resource string) []string {
	configDir := filepath.Join(g.shadowHome, ".config", "opencode")
	skills := filepath.Dir(filepath.Dir(g.report.FlowBuildSkillPath))
	return []string{filepath.Join(skills, filepath.FromSlash(resource)), filepath.Join(configDir, "hive", "pilot-observer", "index.js"), filepath.Join(configDir, "hive", "context-loader", "index.js"), filepath.Join(configDir, "hive", "context-loader", "context-loader.json")}
}

func validateFlashClassFlags(host, caseID, arm, source, model, delivery string, limit time.Duration) error {
	role, _ := flashRole(caseID)
	if host != "opencode" || role == "" || !slices.Contains([]string{"A", "B", "C"}, arm) || source == "" || !slices.Contains(flashModels, model) || delivery != "deployed-global" || limit <= 0 || limit > 3*time.Minute {
		return fmt.Errorf("flash-class-guidance requires OpenCode, a fixed case, arm A/B/C, guidance source, an approved model, deployed-global and timeout at most 3m")
	}
	return nil
}

type flashExecution struct {
	Model, Arm, Case string
	Repetition       int
}

// Rotate the arm order before any execution; do not retry or substitute models.
func flashSchedule() []flashExecution {
	var out []flashExecution
	for m, model := range flashModels {
		for repetition := 1; repetition <= 2; repetition++ {
			for c, caseID := range []string{"plan-reference", "harness-reference"} {
				for offset := 0; offset < 3; offset++ {
					out = append(out, flashExecution{model, []string{"A", "B", "C"}[(m+repetition-1+c+offset)%3], caseID, repetition})
				}
			}
		}
	}
	return out
}

func fixtureWorkingDirectory(root, relative string) (string, error) {
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("fixture cwd must be relative")
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("fixture root unavailable")
	}
	cwd, err := filepath.EvalSymlinks(filepath.Join(root, relative))
	if err != nil {
		return "", fmt.Errorf("fixture cwd unavailable")
	}
	if !under(cwd, realRoot) {
		return "", fmt.Errorf("fixture cwd escapes root")
	}
	return cwd, nil
}

type flashObservation struct {
	Phase          string   `json:"phase"`
	SessionID      string   `json:"sessionID"`
	ParentID       string   `json:"parentID"`
	Agent          string   `json:"agent"`
	AgentName      string   `json:"agentName"`
	Model          string   `json:"model"`
	Sequence       int      `json:"sequence"`
	Resource       string   `json:"resource"`
	Hash           string   `json:"hash"`
	Copies         int      `json:"copies"`
	InjectedCopies int      `json:"injectedCopies"`
	Canonical      bool     `json:"canonical"`
	Correlated     bool     `json:"correlated"`
	Protocol       string   `json:"protocol"`
	Code           string   `json:"code,omitempty"`
	Kind           string   `json:"kind"`
	Plugins        []string `json:"plugins"`
	LoaderActive   bool     `json:"loaderActive"`
}

type flashChild struct {
	SessionID, ParentID, Agent, AgentName, Model string
	FinalAssistantID                             string
	Completed                                    bool
	Cost                                         *float64         `json:",omitempty"`
	Tokens                                       *flashTokenUsage `json:",omitempty"`
	FinalText                                    []string         `json:",omitempty"`
	Operations                                   []traceEvent     `json:",omitempty"`
	EvidenceFile, EvidenceHash                   string           `json:",omitempty"`
}

type flashTokenUsage struct {
	Input, Output, Reasoning float64
	Cache                    struct{ Read, Write float64 }
}

type flashClassReport struct {
	PluginActive                                                            bool
	ObserverHash, LoaderHash, CanonicalHash, Resource, Role, SourceRevision string
	Observations                                                            []flashObservation
	Children                                                                []flashChild
	Coverage, Completion, Memory, Error                                     string
	MemoryRecords                                                           flashMemoryRecords
	PhysicalHTTPRequests                                                    int
	MemoryRouting                                                           string
	EvidenceHash                                                            string
	LoaderMetadataHash                                                      string
	SourceGlobalHash, SourceReferenceHash                                   string
	RoleFileHashBefore, RoleFileHashAfter, OriginalRoleModel                string
	PrimaryCompletion                                                       *flashChild `json:",omitempty"`
	MemoryProject                                                           string
	ConfiguredPluginOrder                                                   []string
	OperationSafety                                                         string
	OperationEvidence                                                       []string
}

func flashCompletion(root, model, role string, children []flashChild) error {
	if root == "" || len(children) != 1 {
		return fmt.Errorf("exactly one observed native child required")
	}
	c := children[0]
	if c.SessionID == "" || c.SessionID == root || c.ParentID != root || c.Agent == "" || !flashAgentRoleMatches(c.Agent, role) || c.Model != model || !c.Completed || c.FinalAssistantID == "" {
		return fmt.Errorf("native child identity, model or completion unverified")
	}
	return nil
}

func readFlashObservations(filename string) ([]flashObservation, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("observer evidence unavailable")
	}
	var out []flashObservation
	for _, line := range bytes.Split(b, []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var record flashObservation
		if json.Unmarshal(line, &record) != nil {
			return nil, fmt.Errorf("observer evidence malformed")
		}
		out = append(out, record)
	}
	return out, nil
}

func modelReference(provider, id, variant string) string {
	if provider == "" || id == "" {
		return "not_observed"
	}
	s := provider + "/" + id
	if variant != "" {
		s += "#" + variant
	}
	return s
}

func flashNativeDispatches(data []byte) (int, error) {
	var doc struct {
		Messages []struct {
			Content []struct {
				Type, ID, Name string
				State          struct{ Status string }
			}
		}
	}
	if json.Unmarshal(data, &doc) != nil {
		return 0, fmt.Errorf("native dispatch metadata malformed")
	}
	seen := map[string]bool{}
	for _, message := range doc.Messages {
		for _, part := range message.Content {
			if part.Type == "tool" && part.Name == "subagent" && part.ID != "" {
				seen[part.ID] = true
			}
		}
	}
	return len(seen), nil
}

// Native export is completion metadata only, never evidence of outgoing context.
func flashChildFromExport(data []byte, root, role, requested string) (flashChild, error) {
	var exported struct {
		Info struct {
			ID, ParentID, Agent, Outcome string
			Cost                         *float64
			Tokens                       *flashTokenUsage
			Model                        struct{ ProviderID, ID, Variant string }
			Time                         struct{ Idle float64 }
		}
		Messages []struct {
			ID, Type, Agent, Finish string
			Model                   struct{ ProviderID, ID, Variant string }
			Time                    struct{ Completed float64 }
			Content                 []struct {
				Type, ID, Name, Text string
				State                struct {
					Status string
					Input  struct {
						Command, Cmd, Code string
						FilePath           string `json:"filePath"`
						Path               string `json:"path"`
						FilePathLegacy     string `json:"file_path"`
						TargetFile         string `json:"target_file"`
					}
				}
			}
		}
	}
	if json.Unmarshal(data, &exported) != nil {
		return flashChild{}, fmt.Errorf("child export malformed")
	}
	c := flashChild{SessionID: exported.Info.ID, ParentID: exported.Info.ParentID, Agent: exported.Info.Agent, AgentName: role, Model: modelReference(exported.Info.Model.ProviderID, exported.Info.Model.ID, exported.Info.Model.Variant)}
	c.Cost = exported.Info.Cost
	c.Tokens = exported.Info.Tokens
	var lastAssistantID string
	validFinal := false
	for _, m := range exported.Messages {
		if m.Type == "assistant" {
			lastAssistantID = m.ID
			validFinal = m.Finish == "stop" && m.Time.Completed > 0 && m.Agent == c.Agent && modelReference(m.Model.ProviderID, m.Model.ID, m.Model.Variant) == requested
			c.FinalText = nil
			for _, part := range m.Content {
				if part.Type == "text" {
					c.FinalText = append(c.FinalText, part.Text)
				}
				if part.Type == "tool" {
					kind := toolKind(part.Name)
					if part.Name == "subagent" {
						kind = "dispatch"
					}
					path := part.State.Input.FilePath
					if path == "" {
						path = part.State.Input.Path
					}
					if path == "" {
						path = part.State.Input.FilePathLegacy
					}
					if path == "" {
						path = part.State.Input.TargetFile
					}
					command := ""
					commandWithheld := false
					if kind == "shell" || strings.EqualFold(part.Name, "execute") {
						command = part.State.Input.Command
						if command == "" {
							command = part.State.Input.Cmd
						}
						if command == "" {
							command = part.State.Input.Code
						}
						originalCommandPresent := command != ""
						if kind == "shell" {
							command = flashOperationCommand(command)
						} else {
							command = ""
						}
						commandWithheld = strings.EqualFold(part.Name, "execute") || (originalCommandPresent && command == "")
					}
					success := part.State.Status == "completed"
					c.Operations = append(c.Operations, traceEvent{SessionID: c.SessionID, Agent: m.Agent, Model: modelReference(m.Model.ProviderID, m.Model.ID, m.Model.Variant), Kind: kind, Tool: part.Name, ID: part.ID, Message: m.ID, Path: path, Command: command, CommandWithheld: commandWithheld, Success: &success})
				}
			}
		} else if m.Type != "idle" {
			validFinal = false
		}
	}
	c.FinalAssistantID = lastAssistantID
	c.Completed = c.SessionID != "" && c.ParentID == root && c.Model == requested && exported.Info.Outcome == "succeeded" && exported.Info.Time.Idle > 0 && validFinal
	if !c.Completed {
		c.FinalText = nil
		return c, fmt.Errorf("child completion or effective model unverified")
	}
	return c, nil
}

func saveFlashProjection(root *os.Root, c *flashChild) error {
	if !c.Completed {
		return fmt.Errorf("native projection lacks completed provenance")
	}
	name := "native-review-" + digest([]byte(c.SessionID))[:24] + ".json"
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("native projection encoding failed")
	}
	file, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("private native projection destination unavailable")
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		return fmt.Errorf("private native projection write failed")
	}
	c.EvidenceFile, c.EvidenceHash = name, digest(data)
	return nil
}

func gateFlashOperations(r *result) {
	if r.FlashClass == nil {
		return
	}
	combined := *r
	combined.Trace.Events = append([]traceEvent(nil), r.Trace.Events...)
	if primary := r.FlashClass.PrimaryCompletion; primary != nil {
		combined.Trace.Events = append(combined.Trace.Events, primary.Operations...)
	}
	for _, child := range r.FlashClass.Children {
		combined.Trace.Events = append(combined.Trace.Events, child.Operations...)
	}
	status, evidence := observedPathSafety(combined, r.Changed, func(string) bool { return false }, true)
	r.FlashClass.OperationSafety, r.FlashClass.OperationEvidence = status, evidence
	if status == "fail" {
		r.Terminal = "unauthorized_write"
	} else if status != "pass" {
		r.Terminal = "operation_inspection_required"
	}
}

func exportFlashChild(sessionID string, deadline time.Time, cwd string, env []string) ([]byte, error) {
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	cmd := exec.CommandContext(ctx, "opencode", "session", "export", "--standalone", sessionID)
	cmd.Dir, cmd.Env = cwd, env
	var b limitedBuffer
	b.limit = openCodeExportLimit
	cmd.Stdout = &b
	if code, _ := runProcess(cmd, time.Until(deadline)); code != 0 {
		return nil, fmt.Errorf("native child export unavailable")
	}
	return b.Bytes(), nil
}

func flashCoverage(arm, root, model, role, canonicalHash string, records []flashObservation) string {
	matched := false
	rootObserved := false
	for _, r := range records {
		if r.Code != "" || r.Phase == "websocket" {
			return "blocked"
		}
		if r.Phase == "http" && r.LoaderActive != (arm == "C") {
			return "blocked"
		}
		if r.SessionID == root && r.Phase == "http" {
			if r.Model != model || !r.Correlated || r.Protocol == "unsupported" {
				return "blocked"
			}
			rootObserved = true
		}
		if r.ParentID != root || !flashAgentRoleMatches(r.Agent, role) || r.Phase != "http" {
			continue
		}
		if r.Model != model || !r.Correlated || r.Protocol == "unsupported" {
			return "blocked"
		}
		if r.Canonical && r.Copies >= 1 && (arm != "C" || r.InjectedCopies == 1) && r.Hash == canonicalHash && canonicalHash != "" {
			matched = true
		} else if arm == "C" {
			return "unverified"
		}
	}
	if matched && rootObserved {
		return "observed"
	}
	return "unverified"
}

// Metadata records provenance only; it grants no runtime file-reading capability.
func validateFlashSnapshotMetadata(metadata []byte, source, skillsDir string) error {
	if !utf8.Valid(metadata) {
		return fmt.Errorf("delivered loader snapshot metadata invalid")
	}
	if err := flashSnapshotJSONKeys(json.NewDecoder(bytes.NewReader(metadata))); err != nil {
		return fmt.Errorf("delivered loader snapshot metadata invalid")
	}
	var derived struct {
		Schema     int `json:"schema"`
		References []struct {
			Role     string `json:"role"`
			Resource string `json:"resource"`
			SHA256   string `json:"sha256"`
		} `json:"references"`
	}
	decoder := json.NewDecoder(bytes.NewReader(metadata))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&derived) != nil || derived.Schema != 2 || len(derived.References) != 2 {
		return fmt.Errorf("delivered loader snapshot metadata invalid")
	}
	if decoder.Decode(new(any)) != io.EOF {
		return fmt.Errorf("delivered loader snapshot metadata invalid")
	}
	for i, caseID := range []string{"harness-reference", "plan-reference"} {
		role, resource := flashRole(caseID)
		reference := derived.References[i]
		if reference.Role != role || reference.Resource != resource {
			return fmt.Errorf("delivered loader snapshot identities or order invalid")
		}
		original, err := os.ReadFile(filepath.Join(source, "content/skills", filepath.FromSlash(resource)))
		if err != nil || !utf8.Valid(original) {
			return fmt.Errorf("arm source snapshot reference unavailable or invalid")
		}
		installed, err := os.ReadFile(filepath.Join(skillsDir, filepath.FromSlash(resource)))
		if err != nil || !bytes.Equal(installed, original) || reference.SHA256 != digest(original) {
			return fmt.Errorf("delivered loader snapshot reference does not match the arm")
		}
	}
	return nil
}

// Reject duplicate keys and case-insensitive aliases accepted by encoding/json.
func flashSnapshotJSONKeys(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	seen := map[string]bool{}
	for decoder.More() {
		if delimiter == '{' {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] || !slices.Contains([]string{"schema", "references", "role", "resource", "sha256"}, name) {
				return fmt.Errorf("invalid snapshot metadata key")
			}
			seen[name] = true
		}
		if err := flashSnapshotJSONKeys(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}

func setupFlashClass(source, caseID, arm, output, model string, g *guidanceVariant, project string) (*flashClassReport, error) {
	g.disposableNativeRoots = []string{filepath.Join(g.shadowHome, ".local"), filepath.Join(g.shadowHome, ".cache")}
	role, resource := flashRole(caseID)
	configDir := filepath.Join(g.shadowHome, ".config", "opencode")
	configPath := filepath.Join(configDir, "opencode.json")
	b, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg map[string]any
	if json.Unmarshal(b, &cfg) != nil {
		return nil, fmt.Errorf("shadow native configuration invalid")
	}
	// Process-local role overrides preserve the installed role's instructions.
	cfg["agents"] = map[string]any{role: map[string]any{"model": model}}
	server := object(object(object(cfg["mcp"])["servers"])["engram"])
	if server == nil {
		return nil, fmt.Errorf("shadow memory server unavailable")
	}
	server["codemode"] = false
	object(server["environment"])["ENGRAM_PROJECT"] = project
	observerSource := filepath.Join(source, "tests/fixtures", flashSuite, "observer.js")
	observer, err := os.ReadFile(observerSource)
	if err != nil {
		return nil, fmt.Errorf("independent observer source unavailable")
	}
	observerPath := filepath.Join(configDir, "hive", "pilot-observer", "index.js")
	skillsDir := filepath.Dir(filepath.Dir(g.report.FlowBuildSkillPath))
	canonicalPath := filepath.Join(skillsDir, filepath.FromSlash(resource))
	canonical, err := os.ReadFile(canonicalPath)
	if err != nil {
		return nil, fmt.Errorf("arm canonical reference unavailable")
	}
	r := &flashClassReport{PluginActive: arm == "C", ObserverHash: digest(observer), Resource: resource, Role: role, CanonicalHash: digest(canonical), Coverage: "unverified", Completion: "unverified", Memory: "unverified", MemoryProject: project}
	r.SourceGlobalHash = hashOrAbsent(filepath.Join(g.report.Source, "content/guidance/global.md"))
	r.SourceReferenceHash = hashOrAbsent(filepath.Join(g.report.Source, "content/skills", filepath.FromSlash(resource)))
	if r.SourceGlobalHash == "absent" || r.SourceReferenceHash == "absent" {
		return nil, fmt.Errorf("arm source guidance or reference unavailable")
	}
	r.SourceRevision, _ = gitOutput(g.report.Source, "rev-parse", "HEAD")
	if arm == "A" && !strings.HasPrefix(r.SourceRevision, "d5164f71") {
		return nil, fmt.Errorf("arm A must use baseline d5164f71")
	}
	if arm == "A" {
		dirty, err := gitOutput(g.report.Source, "status", "--porcelain", "--", "content", "integrations/opencode/context-loader.js", "integrations/agent-profiles.json")
		if err != nil || dirty != "" {
			return nil, fmt.Errorf("arm A baseline payload must match its recorded commit")
		}
	}
	loaderPath := filepath.Join(configDir, "hive", "context-loader", "index.js")
	r.LoaderHash = hashOrAbsent(loaderPath)
	r.LoaderMetadataHash = hashOrAbsent(filepath.Join(configDir, "hive", "context-loader", "context-loader.json"))
	plugins := []any{}
	if arm == "C" {
		loader, err := os.ReadFile(loaderPath)
		if err != nil {
			return nil, fmt.Errorf("delivered inactive loader unavailable")
		}
		r.LoaderHash = digest(loader)
		metadata, err := os.ReadFile(filepath.Join(configDir, "hive", "context-loader", "context-loader.json"))
		if err != nil {
			return nil, fmt.Errorf("delivered loader metadata unavailable")
		}
		if err := validateFlashSnapshotMetadata(metadata, g.report.Source, skillsDir); err != nil {
			return nil, err
		}
		r.LoaderMetadataHash = digest(metadata)
		plugins = append(plugins, filepath.Dir(loaderPath))
		r.ConfiguredPluginOrder = append(r.ConfiguredPluginOrder, "hive.context-loader")
	}
	r.ConfiguredPluginOrder = append(r.ConfiguredPluginOrder, "hive.pilot-observer")
	plugins = append(plugins, map[string]any{"package": filepath.Dir(observerPath), "options": map[string]any{"output": filepath.Join(output, "observer.jsonl"), "canonical": canonicalPath, "resource": resource, "role": role}})
	cfg["plugins"] = plugins
	b, err = json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	r.RoleFileHashBefore, r.RoleFileHashAfter, r.OriginalRoleModel, err = overrideFlashRoleModel(g.shadowHome, role, model)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(observerPath), 0700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(observerPath, observer, 0600); err != nil {
		return nil, err
	}
	if err := os.WriteFile(configPath, b, 0600); err != nil {
		return nil, err
	}
	return r, nil
}

// Agent IDs identify roles; display names are labels and do not grant identity.
func flashAgentRoleMatches(agent, role string) bool {
	return agent == role
}
