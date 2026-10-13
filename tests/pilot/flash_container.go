package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const disposableContainer = "disposable-container"
const deferredVolumeDestruction = "deferred_to_volume_destruction"

// Validate inherited routing before the runner generates its own environment.
func validateFlashContainer(mode, root string, env []string, marked bool) error {
	if mode == "" {
		if root != "" {
			return fmt.Errorf("native data root requires disposable container mode")
		}
		return nil
	}
	if mode != disposableContainer || !marked {
		return fmt.Errorf("disposable memory requires a container marker")
	}
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "ENGRAM_CLOUD_") || key == "ENGRAM_URL" || key == "ENGRAM_PORT" || key == "ENGRAM_DATA_DIR" {
			return fmt.Errorf("inherited memory routing is forbidden: %s", key)
		}
	}
	if !filepath.IsAbs(root) {
		return fmt.Errorf("native data root must be absolute")
	}
	real, err := filepath.EvalSymlinks(root)
	info, statErr := os.Stat(root)
	if err != nil || statErr != nil || !info.IsDir() || real != filepath.Clean(root) {
		return fmt.Errorf("native data root unavailable or symlinked")
	}
	return nil
}

func finishDisposableMemory(g *flashMemoryGuard, env []string, deadline time.Time) error {
	values := map[string]string{}
	for _, entry := range env {
		key, value, _ := strings.Cut(entry, "=")
		values[key] = value
	}
	real, err := filepath.EvalSymlinks(engramDataDir(g.output))
	root, rootErr := filepath.EvalSymlinks(g.output)
	if err != nil || rootErr != nil || !under(real, root) || values["ENGRAM_DATA_DIR"] != engramDataDir(g.output) || values["ENGRAM_CLOUD_AUTOSYNC"] != "0" || g.ownedMemory == nil || g.ownedMemory.verifyOwnedDataDirectory() != "" {
		return fmt.Errorf("disposable memory routing or confinement unverified")
	}
	g.records, err = g.export(env, g.project, deadline)
	if err != nil {
		return err
	}

	return nil
}

// Interrupted runs retain both the store and evidence without claiming success.
func finishInterruptedFlashMemory(g *flashMemoryGuard, memory *memoryIsolation, env []string) {
	memory.stopHTTP()
	err := g.finish(env, time.Now().Add(5*time.Second))
	status := "cleanup_withheld_unverified"
	if !g.disposable && err == nil {
		status = memory.cleanup()
	}
	save(filepath.Join(g.output, "interrupted-cleanup.json"), struct {
		Status  string
		Records flashMemoryRecords
	}{status, g.records})
}

func (g *guidanceVariant) prepareContainerNativeData(root string) error {
	if err := validateFlashContainer(disposableContainer, root, nil, true); err != nil {
		return err
	}
	realHome, err := filepath.EvalSymlinks(g.shadowHome)
	if err != nil || under(root, realHome) || under(realHome, root) {
		return fmt.Errorf("native data root overlaps the shadow home")
	}
	nativeHome, err := os.MkdirTemp(root, "native-")
	if err != nil {
		return fmt.Errorf("container native home unavailable")
	}
	g.containerNative = true
	g.shadowDataDir = filepath.Join(nativeHome, "opencode")
	if err := os.Mkdir(g.shadowDataDir, 0700); err != nil {
		return fmt.Errorf("container native data unavailable")
	}
	return nil
}
