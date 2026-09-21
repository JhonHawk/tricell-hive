package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type memoryIsolationReport struct {
	Mode                      string `json:"mode"`
	DataDirectory             string `json:"data_directory"`
	EnvironmentRequested      bool   `json:"environment_requested"`
	PreflightCommand          string `json:"preflight_command"`
	PreflightVerifiedEmpty    bool   `json:"preflight_verified_empty"`
	CloudSyncDisabled         bool   `json:"cloud_sync_disabled"`
	HTTPStoreVerified         bool   `json:"http_store_verified"`
	RuntimeStatsVerified      bool   `json:"runtime_stats_verified"`
	RuntimeSessions           int    `json:"runtime_sessions"`
	RuntimeObservations       int    `json:"runtime_observations"`
	RuntimePrompts            int    `json:"runtime_prompts"`
	RuntimeMemoryToolReadSeen bool   `json:"runtime_memory_tool_read_seen"`
	Cleanup                   string `json:"cleanup"`
}

type memoryIsolation struct {
	outputDir string
	dataDir   string
	marker    string
	ownedInfo os.FileInfo
	server    *exec.Cmd
	report    memoryIsolationReport
}

const memoryOwnerMarker = "hive pilot isolated Engram store v1\n"

func isolatedMemoryEnvironment(parent []string, dataDir string) []string {
	blocked := map[string]bool{
		"ENGRAM_DATA_DIR":               true,
		"ENGRAM_PROJECT":                true,
		"ENGRAM_URL":                    true,
		"ENGRAM_PORT":                   true,
		"ENGRAM_HTTP_TOKEN":             true,
		"ENGRAM_CLOUD_AUTOSYNC":         true,
		"ENGRAM_CLOUD_SERVER":           true,
		"ENGRAM_CLOUD_TOKEN":            true,
		"ENGRAM_DATABASE_URL":           true,
		"ENGRAM_CLOUD_HOST":             true,
		"ENGRAM_CLOUD_MAX_PUSH_BYTES":   true,
		"ENGRAM_CLOUD_INSECURE_NO_AUTH": true,
	}
	env := make([]string, 0, len(parent)+2)
	for _, entry := range parent {
		key, _, ok := strings.Cut(entry, "=")
		if !ok || blocked[key] {
			continue
		}
		env = append(env, entry)
	}
	env = append(env, "ENGRAM_DATA_DIR="+dataDir, "ENGRAM_CLOUD_AUTOSYNC=0")
	return env
}

// Codex filters the environment inherited by stdio MCP servers. Forward only
// the test's Engram settings through native process-local configuration.
func codexMemoryOverrides(env []string, cwd string) []string {
	values := map[string]string{"ENGRAM_CLOUD_AUTOSYNC": "0", "ENGRAM_CLOUD_SERVER": "", "ENGRAM_CLOUD_TOKEN": "", "ENGRAM_DATABASE_URL": "", "ENGRAM_PROJECT": "", "ENGRAM_HTTP_TOKEN": ""}
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if ok && (key == "ENGRAM_DATA_DIR" || key == "ENGRAM_URL" || key == "ENGRAM_PORT") {
			values[key] = value
		}
	}
	// The installed Engram plugin bundles a second MCP entry. Disable only that
	// duplicate for this invocation; its runtime hooks and skills stay enabled.
	args := []string{"-c", `plugins."engram@engram".mcp_servers.engram.enabled=false`, "-c", "mcp_servers.engram.cwd=" + strconv.Quote(cwd)}
	for _, key := range []string{"ENGRAM_DATA_DIR", "ENGRAM_URL", "ENGRAM_PORT", "ENGRAM_CLOUD_AUTOSYNC", "ENGRAM_CLOUD_SERVER", "ENGRAM_CLOUD_TOKEN", "ENGRAM_DATABASE_URL", "ENGRAM_PROJECT", "ENGRAM_HTTP_TOKEN"} {
		args = append(args, "-c", "mcp_servers.engram.env."+key+"="+strconv.Quote(values[key]))
	}
	return args
}

func codexMemoryArgs(args, env []string, cwd string) ([]string, error) {
	for i, arg := range args {
		if arg == "exec" {
			// Codex 0.155.1 exec did not apply top-level -c overrides in the
			// observed launch. Bind these to the subcommand that starts MCP.
			out := append([]string{}, args[:i+1]...)
			out = append(out, codexMemoryOverrides(env, cwd)...)
			return append(out, args[i+1:]...), nil
		}
	}
	return nil, fmt.Errorf("memory overrides require Codex exec")
}

func prepareMemoryIsolation(output string, parentEnv []string) (*memoryIsolation, []string, error) {
	dataDir := filepath.Join(output, "engram-data")
	if err := os.Mkdir(dataDir, 0700); err != nil {
		return nil, nil, fmt.Errorf("create isolated Engram data directory: %w", err)
	}
	info, err := os.Lstat(dataDir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, nil, fmt.Errorf("isolated Engram data path is not a private directory")
	}
	if err := os.Chmod(dataDir, 0700); err != nil {
		return nil, nil, fmt.Errorf("secure isolated Engram data directory: %w", err)
	}
	if entries, err := os.ReadDir(dataDir); err != nil || len(entries) != 0 {
		return nil, nil, fmt.Errorf("isolated Engram data directory was not empty before preflight")
	}
	marker := filepath.Join(dataDir, ".pilot-owner")
	if err := os.WriteFile(marker, []byte(memoryOwnerMarker), 0600); err != nil {
		return nil, nil, fmt.Errorf("mark isolated Engram data directory")
	}
	bin, err := exec.LookPath("engram")
	if err != nil {
		return nil, nil, fmt.Errorf("Engram CLI is required for isolated pilot runs")
	}
	env := isolatedMemoryEnvironment(parentEnv, dataDir)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "projects", "list")
	cmd.Env = env
	stdout, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("Engram isolation preflight failed")
	}
	if !emptyEngramProjects(stdout) {
		return nil, nil, fmt.Errorf("Engram isolation preflight did not confirm an empty database")
	}
	return &memoryIsolation{
		outputDir: output,
		dataDir:   dataDir,
		marker:    marker,
		ownedInfo: info,
		report: memoryIsolationReport{
			Mode:                   "per_run_data_dir",
			DataDirectory:          "engram-data",
			EnvironmentRequested:   true,
			PreflightCommand:       "engram projects list",
			PreflightVerifiedEmpty: true,
			CloudSyncDisabled:      true,
		},
	}, env, nil
}

func emptyEngramProjects(stdout []byte) bool {
	return strings.TrimSpace(string(stdout)) == "No projects found."
}

// HTTP-backed native extensions must reach the same isolated store as stdio MCP.
// The pilot owns this native Engram process and never reuses the everyday server.
func (m *memoryIsolation) startHTTP(env []string, cwd string) ([]string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("reserve isolated Engram port: %w", err)
	}
	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	listener.Close()
	env = append(env, "ENGRAM_PORT="+port, "ENGRAM_URL=http://127.0.0.1:"+port)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	idCmd := exec.CommandContext(ctx, "engram", "instance-id")
	idCmd.Env = env
	idCmd.Dir = cwd
	id, err := idCmd.Output()
	if err != nil || len(strings.TrimSpace(string(id))) != 32 {
		return nil, fmt.Errorf("isolated Engram identity unavailable")
	}
	m.server = exec.Command("engram", "serve", port)
	m.server.Env = env
	m.server.Dir = cwd
	if err := m.server.Start(); err != nil {
		m.server = nil
		return nil, fmt.Errorf("isolated Engram HTTP startup failed")
	}
	client := &http.Client{Timeout: 250 * time.Millisecond}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get("http://127.0.0.1:" + port + "/health")
		if err == nil {
			var health struct {
				InstanceID string `json:"instance_id"`
			}
			decodeErr := json.NewDecoder(http.MaxBytesReader(nil, response.Body, 4096)).Decode(&health)
			response.Body.Close()
			if response.StatusCode == http.StatusOK && decodeErr == nil && health.InstanceID == strings.TrimSpace(string(id)) {
				m.report.HTTPStoreVerified = true
				return env, nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	m.stopHTTP()
	return nil, fmt.Errorf("isolated Engram HTTP ownership was not verified")
}

func (m *memoryIsolation) stopHTTP() {
	if m != nil && m.server != nil {
		_ = m.server.Process.Kill()
		_ = m.server.Wait()
		m.server = nil
	}
}

func (m *memoryIsolation) captureStats() {
	if m == nil || m.dataDir == "" {
		return
	}
	bin, err := exec.LookPath("engram")
	if err != nil {
		return
	}
	env := isolatedMemoryEnvironment(os.Environ(), m.dataDir)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "stats", "--all")
	cmd.Env = env
	stdout, err := cmd.Output()
	if err != nil {
		return
	}
	counts := map[string]int{}
	for _, line := range strings.Split(string(stdout), "\n") {
		for label, key := range map[string]string{"Sessions:": "sessions", "Observations:": "observations", "Prompts:": "prompts"} {
			if strings.HasPrefix(strings.TrimSpace(line), label) {
				value := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), label))
				count, parseErr := strconv.Atoi(value)
				if parseErr == nil {
					counts[key] = count
				}
			}
		}
	}
	if len(counts) != 3 {
		return
	}
	m.report.RuntimeSessions = counts["sessions"]
	m.report.RuntimeObservations = counts["observations"]
	m.report.RuntimePrompts = counts["prompts"]
	m.report.RuntimeStatsVerified = true
}

func (m *memoryIsolation) cleanup() string {
	if m == nil || m.dataDir == "" {
		return "not_created"
	}
	m.stopHTTP()
	info, err := os.Lstat(m.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "already_absent"
		}
		return "cleanup_failed"
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "cleanup_refused_unexpected_path_type"
	}
	if m.ownedInfo == nil || !os.SameFile(m.ownedInfo, info) {
		return "cleanup_refused_replaced_directory"
	}
	rel, err := filepath.Rel(m.outputDir, m.dataDir)
	if err != nil || rel != "engram-data" {
		return "cleanup_refused_unexpected_path"
	}
	marker, err := os.ReadFile(m.marker)
	if err != nil || string(marker) != memoryOwnerMarker {
		return "cleanup_refused_missing_owner_marker"
	}
	if err := os.RemoveAll(m.dataDir); err != nil {
		return "cleanup_failed"
	}
	if _, err := os.Lstat(m.dataDir); !os.IsNotExist(err) {
		return "cleanup_failed"
	}
	return "isolated_database_removed"
}

func observedMemoryRead(events []traceEvent) bool {
	for _, event := range events {
		if !event.Memory || event.Success == nil || !*event.Success {
			continue
		}
		name := strings.ToLower(event.Tool)
		for _, readTool := range []string{"mem_search", "mem_context", "mem_get_observation", "mem_timeline"} {
			if strings.Contains(name, readTool) {
				for _, result := range events {
					if result.ID == event.ID && result.Kind == "tool_result" && result.Success != nil && *result.Success && memoryReadSucceeded(result.Text) {
						return true
					}
				}
			}
		}
	}
	return false
}

func memoryReadSucceeded(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" || strings.HasPrefix(text, "gentle-engram could not initialize") {
		return false
	}
	var response map[string]any
	if json.Unmarshal([]byte(text), &response) == nil {
		if code, ok := response["error_code"].(string); ok && code != "" {
			return false
		}
	}
	return true
}
