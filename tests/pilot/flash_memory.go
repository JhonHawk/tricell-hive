package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type flashMemoryCounts struct {
	Sessions     int `json:"sessions"`
	Observations int `json:"observations"`
	Prompts      int `json:"prompts"`
}

type flashMemoryRecords struct {
	ExportedCounts     *flashMemoryCounts `json:"exported_counts,omitempty"`
	UnexpectedProjects []string           `json:"unexpected_projects,omitempty"`
	Sessions           []string           `json:"sessions"`
	Observations       []string           `json:"observations"`
	Prompts            []string           `json:"prompts"`
}

type flashMemoryGuard struct {
	project, output, temporary string
	everydayEnv                []string
	records                    flashMemoryRecords
	mu                         sync.Mutex
	finished                   bool
	disposable                 bool
	finishError                error
	ownedMemory                *memoryIsolation
}

// Scope every everyday-store lookup to this fresh fixture project. Never export
// other projects, dump human history, or delete records in the everyday store.
func prepareFlashMemoryGuard(project, output string, parent []string) (*flashMemoryGuard, error) {
	temporary, err := os.MkdirTemp("", "hive-flash-memory-")
	if err != nil {
		return nil, err
	}
	g := &flashMemoryGuard{project: project, output: output, temporary: temporary}
	for _, entry := range parent {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "ENGRAM_CLOUD_") || key == "ENGRAM_PROJECT" || key == "ENGRAM_URL" || key == "ENGRAM_PORT" || key == "ENGRAM_HTTP_TOKEN" {
			continue
		}
		g.everydayEnv = append(g.everydayEnv, entry)
	}
	g.everydayEnv = append(g.everydayEnv, "ENGRAM_CLOUD_AUTOSYNC=0")
	before, err := g.export(g.everydayEnv, project, time.Now().Add(10*time.Second))
	if err != nil {
		g.removeTemporaryExports()
		return nil, err
	}
	if !before.empty() {
		g.removeTemporaryExports()
		return nil, fmt.Errorf("fixture project already exists in everyday memory")
	}
	return g, nil
}

func (r flashMemoryRecords) empty() bool {
	return len(r.Sessions)+len(r.Observations)+len(r.Prompts) == 0
}

// Exports are private, transient inputs. Only IDs with exact project ownership
// are returned. A missing ownership field is unverified, never permission.
func parseFlashMemoryRecords(raw []byte, project string) (flashMemoryRecords, error) {
	return parseFlashMemoryExport(raw, project, false)
}

func parseFlashMemoryExport(raw []byte, project string, fullStore bool) (flashMemoryRecords, error) {
	var doc struct {
		Version  string `json:"version"`
		Sessions []struct {
			ID      string `json:"id"`
			Project string `json:"project"`
		} `json:"sessions"`
		Observations []struct {
			ID      json.Number `json:"id"`
			Project string      `json:"project"`
		} `json:"observations"`
		Prompts []struct {
			ID      json.Number `json:"id"`
			Project string      `json:"project"`
		} `json:"prompts"`
	}
	if json.Unmarshal(raw, &doc) != nil || doc.Version == "" {
		return flashMemoryRecords{}, fmt.Errorf("memory export schema unverified")
	}
	r := flashMemoryRecords{}
	if fullStore {
		// Count every exported record, including unexpected projects. Null arrays
		// decode as empty slices in Engram's empty-store export.
		r.ExportedCounts = &flashMemoryCounts{Sessions: len(doc.Sessions), Observations: len(doc.Observations), Prompts: len(doc.Prompts)}
		owners := map[string]bool{}
		noteOwner := func(owner string) {
			if owner != project && !owners[owner] {
				owners[owner] = true
				r.UnexpectedProjects = append(r.UnexpectedProjects, owner)
			}
		}
		for _, record := range doc.Sessions {
			noteOwner(record.Project)
		}
		for _, record := range doc.Observations {
			noteOwner(record.Project)
		}
		for _, record := range doc.Prompts {
			noteOwner(record.Project)
		}
	}
	seen := map[string]bool{}
	add := func(kind, id, owner string, target *[]string) error {
		if id == "" || owner != project || seen[kind+":"+id] {
			return fmt.Errorf("memory record provenance unverified")
		}
		seen[kind+":"+id] = true
		*target = append(*target, id)
		return nil
	}
	for _, s := range doc.Sessions {
		if err := add("session", s.ID, s.Project, &r.Sessions); err != nil {
			return r, err
		}
	}
	for _, o := range doc.Observations {
		if n, err := strconv.ParseInt(o.ID.String(), 10, 64); err != nil || n <= 0 {
			return r, fmt.Errorf("memory observation ID invalid")
		}
		if err := add("observation", o.ID.String(), o.Project, &r.Observations); err != nil {
			return r, err
		}
	}
	for _, p := range doc.Prompts {
		if n, err := strconv.ParseInt(p.ID.String(), 10, 64); err != nil || n <= 0 {
			return r, fmt.Errorf("memory prompt ID invalid")
		}
		if err := add("prompt", p.ID.String(), p.Project, &r.Prompts); err != nil {
			return r, err
		}
	}
	return r, nil
}

func (g *flashMemoryGuard) export(env []string, project string, deadline time.Time) (flashMemoryRecords, error) {
	path := filepath.Join(g.temporary, "export.json")
	if g.disposable {
		path = filepath.Join(g.output, "memory-export.json")
	}
	if err := os.WriteFile(path, nil, 0600); err != nil {
		return flashMemoryRecords{}, fmt.Errorf("private memory export could not be prepared")
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	args := []string{"export", path, "--project", project}
	if g.disposable {
		args = []string{"export", path, "--all"}
	}
	cmd := exec.CommandContext(ctx, "engram", args...)
	cmd.Env, cmd.Dir = env, g.temporary
	if code, _ := runProcess(cmd, time.Until(deadline)); code != 0 {
		return flashMemoryRecords{}, fmt.Errorf("scoped memory export failed")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return flashMemoryRecords{}, fmt.Errorf("scoped memory export unavailable")
	}
	return parseFlashMemoryExport(raw, project, g.disposable)
}

func (g *flashMemoryGuard) finish(env []string, deadline time.Time) (result error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.finished {
		return g.finishError
	}
	defer func() { g.finished = true; g.finishError = result }()
	defer func() {
		if !g.disposable || result == nil {
			g.removeTemporaryExports()
		}
	}()
	if g.disposable {
		return finishDisposableMemory(g, env, deadline)
	}
	values := map[string]string{}
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			values[key] = value
		}
	}
	if values["ENGRAM_DATA_DIR"] != engramDataDir(g.output) || values["ENGRAM_CLOUD_AUTOSYNC"] != "0" {
		return fmt.Errorf("memory cleanup refused unverified runtime routing")
	}
	if g.ownedMemory == nil || g.ownedMemory.verifyOwnedDataDirectory() != "" {
		return fmt.Errorf("memory directory ownership or confinement unverified")
	}
	everyday, err := g.export(g.everydayEnv, g.project, deadline)
	if err != nil {
		return err
	}
	if !everyday.empty() {
		return fmt.Errorf("fixture records reached everyday memory; cleanup withheld")
	}
	g.records, err = g.export(env, g.project, deadline)
	if err != nil {
		return err
	}
	// Inspect the entire already-verified-empty isolated store to detect writes
	// routed under another project. Export --all is never used on everyday data.
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	path := filepath.Join(g.temporary, "isolated-all.json")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		return fmt.Errorf("private routing export could not be prepared")
	}
	cmd := exec.CommandContext(ctx, "engram", "export", path, "--all")
	cmd.Env, cmd.Dir = env, g.temporary
	if code, _ := runProcess(cmd, time.Until(deadline)); code != 0 {
		return fmt.Errorf("isolated memory routing export failed")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("isolated routing evidence unavailable")
	}
	all, err := parseFlashMemoryRecords(raw, g.project)
	if err != nil {
		return err
	}
	if len(all.Sessions) != len(g.records.Sessions) || len(all.Observations) != len(g.records.Observations) || len(all.Prompts) != len(g.records.Prompts) {
		return fmt.Errorf("isolated memory routing mismatched scoped evidence")
	}
	// A fresh CLI invocation captures whatever database is currently at the
	// pathname, not the generation from which these IDs were exported. Engram
	// 3.3.2's HTTP Store has generation fencing, but its public API has no atomic
	// exported-generation capability covering observation, prompt and session
	// deletion. Do not treat path rechecks or a raw directory removal as the
	// required supported exact-ID cleanup. Preserve the store and block the gate.
	return fmt.Errorf("atomic memory store binding unavailable; exact-ID cleanup withheld")
}

func (g *flashMemoryGuard) removeTemporaryExports() {
	if g != nil && g.temporary != "" {
		_ = os.RemoveAll(g.temporary)
	}
}
