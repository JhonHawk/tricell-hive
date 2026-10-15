package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFlashClassGoSchedule(t *testing.T) {
	if len(flashSchedule()) != 48 {
		t.Fatal("expected 48 roots")
	}
	for _, run := range flashSchedule() {
		if !strings.HasPrefix(run.Model, "opencode-go/") {
			t.Errorf("unapproved provider: %s", run.Model)
		}
	}
}

func TestFlashMemoryContainerValidation(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := validateFlashContainer("disposable-container", root, nil, true); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"ENGRAM_CLOUD_AUTOSYNC", "ENGRAM_CLOUD_UNKNOWN", "ENGRAM_URL", "ENGRAM_PORT", "ENGRAM_DATA_DIR"} {
		if validateFlashContainer("disposable-container", root, []string{key + "="}, true) == nil {
			t.Errorf("inherited %s accepted", key)
		}
	}
	if validateFlashContainer("disposable-container", root, nil, false) == nil {
		t.Error("missing container marker accepted")
	}
	if validateFlashContainer("disposable-container", filepath.Join(root, "missing"), nil, true) == nil {
		t.Error("missing tmpfs root accepted")
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	if validateFlashContainer("disposable-container", link, nil, true) == nil {
		t.Error("symlink root accepted")
	}
}

func TestFlashMemoryDisposableFinish(t *testing.T) {
	_, log := fakeFlashMemoryCLI(t)
	out := t.TempDir()
	dir := engramDataDir(out)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(dir)
	marker := filepath.Join(dir, ".pilot-owner")
	if err := os.WriteFile(marker, []byte(memoryOwnerMarker), 0600); err != nil {
		t.Fatal(err)
	}
	g := &flashMemoryGuard{project: "fixture", output: out, temporary: t.TempDir(), disposable: true, ownedMemory: &memoryIsolation{outputDir: out, dataDir: dir, marker: marker, ownedInfo: info}}
	// Successful export retains the disposable store for volume destruction.
	err := finishDisposableMemory(g, isolatedMemoryEnvironment(os.Environ(), dir), time.Now().Add(5*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "memory-export.json")); err != nil {
		t.Fatal("export evidence missing", err)
	}
	calls, _ := os.ReadFile(log)
	if strings.Contains(string(calls), "everyday") || strings.Contains(string(calls), "delete") {
		t.Fatal("everyday lookup or deletion")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal("store removed")
	}
}

func TestFlashMemoryDisposableInterruptedCleanup(t *testing.T) {
	_, log := fakeFlashMemoryCLI(t)
	out := t.TempDir()
	dir := engramDataDir(out)
	os.Mkdir(dir, 0700)
	info, _ := os.Stat(dir)
	marker := filepath.Join(dir, ".pilot-owner")
	os.WriteFile(marker, []byte(memoryOwnerMarker), 0600)
	memory := &memoryIsolation{outputDir: out, dataDir: dir, marker: marker, ownedInfo: info}
	g := &flashMemoryGuard{project: "fixture", output: out, temporary: t.TempDir(), disposable: true, ownedMemory: memory}
	registerCleanup(func() { finishInterruptedFlashMemory(g, memory, isolatedMemoryEnvironment(os.Environ(), dir)) })
	runCleanups()
	b, err := os.ReadFile(filepath.Join(out, "interrupted-cleanup.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "cleanup_withheld_unverified") || strings.Contains(string(b), "verified_clean") {
		t.Fatal(string(b))
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal("interrupted store removed")
	}
	calls, _ := os.ReadFile(log)
	if strings.Contains(string(calls), "everyday") || strings.Contains(string(calls), "delete") {
		t.Fatal("unsafe cleanup")
	}
}

func TestFlashMemoryDisposableExportFailure(t *testing.T) {
	fakeFlashMemoryCLI(t)
	out := t.TempDir()
	dir := engramDataDir(out)
	os.Mkdir(dir, 0700)
	info, _ := os.Stat(dir)
	marker := filepath.Join(dir, ".pilot-owner")
	os.WriteFile(marker, []byte(memoryOwnerMarker), 0600)
	g := &flashMemoryGuard{project: "fixture", output: out, temporary: t.TempDir(), disposable: true, ownedMemory: &memoryIsolation{outputDir: out, dataDir: dir, marker: marker, ownedInfo: info}}
	os.WriteFile(os.Getenv("FAKE_MEMORY_STATE"), []byte(`{"invalid":true}`), 0600)
	if g.finish(isolatedMemoryEnvironment(os.Environ(), dir), time.Now().Add(time.Second)) == nil {
		t.Fatal("failed export accepted")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal("store removed after failed export")
	}
}

func TestFlashRunnerContainerNativeRoutingAndCatalog(t *testing.T) {
	out := t.TempDir()
	home := filepath.Join(out, "shadow-home")
	os.Mkdir(home, 0700)
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	g := &guidanceVariant{host: "opencode", shadowHome: home}
	if err := g.prepareContainerNativeData(root); err != nil {
		t.Fatal(err)
	}
	if !under(g.shadowDataDir, root) || under(g.shadowDataDir, out) {
		t.Fatal("native data escaped tmpfs")
	}
	env := g.applyEnvironment([]string{"XDG_DATA_HOME=/outside"})
	joined := strings.Join(env, "\n")
	if !strings.Contains(joined, "XDG_DATA_HOME="+filepath.Dir(g.shadowDataDir)) || strings.Contains(joined, "/outside") {
		t.Fatal("native routing missing")
	}
	other := &guidanceVariant{host: "opencode", shadowHome: home}
	if err := other.prepareContainerNativeData(root); err != nil {
		t.Fatal(err)
	}
	if g.shadowDataDir == other.shadowDataDir {
		t.Fatal("native homes shared")
	}
	cache := filepath.Join(t.TempDir(), "container-cache")
	os.Mkdir(cache, 0700)
	shadowCache := filepath.Join(home, ".cache/opencode")
	if copyOpenCodeModelCatalog(shadowCache, cache) == nil {
		t.Fatal("missing catalog accepted")
	}
	os.WriteFile(filepath.Join(cache, "models.json"), []byte(`{"container":true}`), 0600)
	if err := copyOpenCodeModelCatalog(shadowCache, cache); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(shadowCache, "models.json"))
	if string(b) != `{"container":true}` {
		t.Fatal("wrong catalog")
	}
}
