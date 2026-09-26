package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/version"
)

// bootstrapFixtureOrigin serves a download index plus a real, install-shaped
// package (this checkout's own content/ and integrations/agent-profiles.json,
// exactly as install_test.go's --source ../.. uses them, packaged the same
// way tooling/distribution's own fixtures build a package directory: plain
// bytes for bin/hive rather than a real compiled binary, since only Go-level
// bootstrap() is under test here — bootstrap_shell_test.go covers the real
// exec'd manager). It returns the httptest server and the manager bytes and
// checksum bootstrap.sh would have already verified before invoking bootstrap.
func bootstrapFixtureOrigin(t *testing.T, productVersion string) (srv *httptest.Server, managerBytes []byte, managerSHA256 string) {
	t.Helper()
	// ValidateBootstrapPackage binds the requested release to the running
	// manager's own compiled-in version (real release builds set this via
	// -ldflags; here the test process stands in for that already-verified
	// manager binary, so it must claim the same version being bootstrapped).
	previous := version.Current
	version.Current = productVersion
	t.Cleanup(func() { version.Current = previous })
	src, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, rel := range []string{"content", "integrations/agent-profiles.json"} {
		if err := copyTreeForBootstrapFixture(filepath.Join(src, rel), filepath.Join(dir, rel)); err != nil {
			t.Fatal(err)
		}
	}
	managerBytes = []byte("#!/bin/sh\nexit 0\n")
	managerSHA256 = distribution.Digest(managerBytes)
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bin", "hive"), managerBytes, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bin", "hive.sha256"), []byte(managerSHA256+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "install.sh"), []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	platform := runtime.GOOS + "/" + runtime.GOARCH
	if err := os.WriteFile(filepath.Join(dir, "platform"), []byte(platform+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	files, err := distribution.Files(dir)
	if err != nil {
		t.Fatal(err)
	}
	manifest := distribution.Manifest{Version: 1, Platform: platform, SourceID: strings.Repeat("a", 64), ProductVersion: productVersion, Files: files}
	metadata, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, distribution.ManifestName), append(metadata, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	archive := tarGzDirectory(t, dir)
	packageSHA256 := distribution.Digest(archive)
	index := distribution.DownloadIndex{Version: 1, ProductVersion: productVersion, Releases: []distribution.DownloadRelease{{
		Platform: platform, SourceID: manifest.SourceID,
		Package: "versions/" + productVersion + "/pkg.tar.gz", PackageSHA256: packageSHA256,
		RawBinary: "versions/" + productVersion + "/hive", RawBinarySHA256: managerSHA256,
	}}}
	indexData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/versions/"+productVersion+"/index.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(indexData)
	})
	mux.HandleFunc("/versions/"+productVersion+"/pkg.tar.gz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	})
	// distribution.Download refuses every non-HTTPS origin (production must
	// never fall back to plain HTTP), so this fixture serves real TLS. Its
	// self-signed certificate is trusted only by swapping the process-wide
	// default transport distribution.Download's client falls back to,
	// restored immediately after this test.
	srv = httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)
	previousTransport := http.DefaultTransport
	http.DefaultTransport = srv.Client().Transport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
	return srv, managerBytes, managerSHA256
}

// resolvedTempDir mirrors what management.NormalizeOptions resolves --home
// to internally (target.Canonical evaluates symlinks). On macOS t.TempDir()
// sits under /var, itself a symlink to /private/var, and target.Safe rejects
// any symlinked ancestor; comparing raw and resolved forms of the same
// directory would otherwise look like two different, conflicting paths.
func resolvedTempDir(t *testing.T) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func copyTreeForBootstrapFixture(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, info.Mode().Perm())
	})
}

// tarGzDirectory archives dir's own contents under one top-level directory
// named after dir's base, matching the single-top-level-directory shape
// distribution.Extract requires.
func tarGzDirectory(t *testing.T, dir string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	base := filepath.Dir(dir)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestBootstrapReachesInstallSummaryAndConsentRetainsInstaller(t *testing.T) {
	version := "9.1.1"
	srv, managerBytes, managerSHA256 := bootstrapFixtureOrigin(t, version)
	managerFile := filepath.Join(t.TempDir(), "hive-manager")
	if err := os.WriteFile(managerFile, managerBytes, 0700); err != nil {
		t.Fatal(err)
	}
	home := resolvedTempDir(t)
	stateDir := filepath.Join(home, "state")
	args := []string{
		"--origin", srv.URL, "--version", version,
		"--manager", managerFile, "--manager-sha256", managerSHA256,
		"--home", home, "--state-dir", stateDir, "--hosts", "codex",
	}
	var out bytes.Buffer
	if err := bootstrapWithAdapterFactory(args, strings.NewReader("y\n"), &out, true, coreOnlyAdapterFactory); err != nil {
		t.Fatalf("bootstrap did not complete: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Apply these changes?") {
		t.Fatalf("bootstrap did not reach the install summary/consent prompt: %s", out.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "AGENTS.md")); err != nil {
		t.Fatalf("consented bootstrap did not install: %v", err)
	}
	// The retained installer must be discoverable offline afterward, exactly
	// as a second, disconnected terminal running `hive recover` would need.
	entries, err := os.ReadDir(filepath.Join(stateDir, "installers"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("installer was not retained after consent: %v %v", entries, err)
	}
	if _, err := distribution.RetainedInstallerAt(stateDir, entries[0].Name()); err != nil {
		t.Fatalf("retained installer failed offline reverification: %v", err)
	}
}

func TestBootstrapCancelLeavesNoPersistentChangeAndCleansItsScratch(t *testing.T) {
	version := "9.1.2"
	srv, managerBytes, managerSHA256 := bootstrapFixtureOrigin(t, version)
	managerFile := filepath.Join(t.TempDir(), "hive-manager")
	if err := os.WriteFile(managerFile, managerBytes, 0700); err != nil {
		t.Fatal(err)
	}
	home := resolvedTempDir(t)
	stateDir := filepath.Join(home, "state")
	args := []string{
		"--origin", srv.URL, "--version", version,
		"--manager", managerFile, "--manager-sha256", managerSHA256,
		"--home", home, "--state-dir", stateDir, "--hosts", "codex",
	}
	var out bytes.Buffer
	if err := bootstrapWithAdapterFactory(args, strings.NewReader("n\n"), &out, true, coreOnlyAdapterFactory); err != nil {
		t.Fatalf("cancelling bootstrap returned an error: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Apply these changes?") {
		t.Fatalf("bootstrap did not reach the install summary/consent prompt before cancelling: %s", out.String())
	}
	if entries, err := os.ReadDir(home); err != nil || len(entries) != 0 {
		t.Fatalf("cancelled bootstrap wrote into home: %v %v", entries, err)
	}
	if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
		t.Fatalf("cancelled bootstrap created persistent state: %v", err)
	}
}

func TestBootstrapRejectsTamperedManagerFile(t *testing.T) {
	version := "9.1.3"
	srv, managerBytes, managerSHA256 := bootstrapFixtureOrigin(t, version)
	managerFile := filepath.Join(t.TempDir(), "hive-manager")
	// Simulate the file at --manager having changed since bootstrap.sh's own
	// checksum verification: it must never be trusted a second time.
	if err := os.WriteFile(managerFile, append(managerBytes, '\n'), 0700); err != nil {
		t.Fatal(err)
	}
	home := resolvedTempDir(t)
	args := []string{
		"--origin", srv.URL, "--version", version,
		"--manager", managerFile, "--manager-sha256", managerSHA256,
		"--home", home, "--hosts", "codex",
	}
	var out bytes.Buffer
	err := bootstrapWithAdapterFactory(args, strings.NewReader("y\n"), &out, true, coreOnlyAdapterFactory)
	if err == nil || !strings.Contains(err.Error(), "manager-sha256") {
		t.Fatalf("accepted a manager file that no longer matches its verified checksum: %v", err)
	}
}

// TestBootstrapRejectsUnpublishedVersion covers the plain case the old
// TestBootstrapRejectsWrongPlatformOrTamperedPackage actually exercised: a
// requested version with no published index at all (the origin's index.json
// route was never registered for it, so the download itself 404s). Wrong
// platform and a tampered package are each their own, more targeted test
// below, since neither one is this: both start from a genuinely published,
// otherwise-valid index and package.
func TestBootstrapRejectsUnpublishedVersion(t *testing.T) {
	version := "9.1.4"
	srv, managerBytes, managerSHA256 := bootstrapFixtureOrigin(t, version)
	managerFile := filepath.Join(t.TempDir(), "hive-manager")
	if err := os.WriteFile(managerFile, managerBytes, 0700); err != nil {
		t.Fatal(err)
	}
	home := resolvedTempDir(t)
	args := []string{
		"--origin", srv.URL, "--version", "0.0.0-nonexistent",
		"--manager", managerFile, "--manager-sha256", managerSHA256,
		"--home", home, "--hosts", "codex",
	}
	var out bytes.Buffer
	if err := bootstrapWithAdapterFactory(args, strings.NewReader("y\n"), &out, true, coreOnlyAdapterFactory); err == nil {
		t.Fatal("accepted a version with no published index")
	}
}

// bootstrapFixtureConfig lets a test start from one fully valid,
// install-shaped package/index/manager triple (the same shape
// bootstrapFixtureOrigin builds) and corrupt exactly one input, isolating
// which of bootstrap.go's own verification lines is the one that rejects it.
// The empty value of every field keeps that input genuinely valid.
type bootstrapFixtureConfig struct {
	indexProductVersion    string // overrides the index.json top-level version
	releasePlatform        string // overrides the one release entry's platform
	releaseRawBinarySHA256 string // overrides the release's bound manager digest
	corruptPackageBytes    bool   // serves the archive with a stray trailing byte
}

// newBootstrapFixture mirrors bootstrapFixtureOrigin's own construction (a
// real install-shaped package, a manager binary, and a served index.json),
// applying cfg's overrides before marshaling the index or serving the
// package so every field except the one under test stays genuinely valid:
// each checksum is computed from the same bytes actually served, except
// where cfg deliberately says otherwise.
func newBootstrapFixture(t *testing.T, productVersion string, cfg bootstrapFixtureConfig) (srv *httptest.Server, managerBytes []byte, managerSHA256 string) {
	t.Helper()
	src, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, rel := range []string{"content", "integrations/agent-profiles.json"} {
		if err := copyTreeForBootstrapFixture(filepath.Join(src, rel), filepath.Join(dir, rel)); err != nil {
			t.Fatal(err)
		}
	}
	managerBytes = []byte("#!/bin/sh\nexit 0\n")
	managerSHA256 = distribution.Digest(managerBytes)
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bin", "hive"), managerBytes, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bin", "hive.sha256"), []byte(managerSHA256+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "install.sh"), []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	platform := runtime.GOOS + "/" + runtime.GOARCH
	if err := os.WriteFile(filepath.Join(dir, "platform"), []byte(platform+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	files, err := distribution.Files(dir)
	if err != nil {
		t.Fatal(err)
	}
	manifest := distribution.Manifest{Version: 1, Platform: platform, SourceID: strings.Repeat("a", 64), ProductVersion: productVersion, Files: files}
	metadata, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, distribution.ManifestName), append(metadata, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	archive := tarGzDirectory(t, dir)
	packageSHA256 := distribution.Digest(archive)
	served := archive
	if cfg.corruptPackageBytes {
		// Still a fully valid tar.gz (tar's own end-of-archive marker is
		// read well before this trailing byte), but no longer matching the
		// digest recorded in the index below: undetectable by structural
		// extraction alone, caught only by the explicit digest comparison.
		served = append(append([]byte{}, archive...), 0)
	}

	releasePlatform := platform
	if cfg.releasePlatform != "" {
		releasePlatform = cfg.releasePlatform
	}
	rawBinarySHA256 := managerSHA256
	if cfg.releaseRawBinarySHA256 != "" {
		rawBinarySHA256 = cfg.releaseRawBinarySHA256
	}
	indexProductVersion := productVersion
	if cfg.indexProductVersion != "" {
		indexProductVersion = cfg.indexProductVersion
	}
	index := distribution.DownloadIndex{Version: 1, ProductVersion: indexProductVersion, Releases: []distribution.DownloadRelease{{
		Platform: releasePlatform, SourceID: manifest.SourceID,
		Package: "versions/" + productVersion + "/pkg.tar.gz", PackageSHA256: packageSHA256,
		RawBinary: "versions/" + productVersion + "/hive", RawBinarySHA256: rawBinarySHA256,
	}}}
	indexData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/versions/"+productVersion+"/index.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(indexData)
	})
	mux.HandleFunc("/versions/"+productVersion+"/pkg.tar.gz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(served)
	})
	srv = httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)
	previousTransport := http.DefaultTransport
	http.DefaultTransport = srv.Client().Transport
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
	return srv, managerBytes, managerSHA256
}

func bootstrapFixtureArgs(srv *httptest.Server, version, managerFile, managerSHA256, home string) []string {
	return []string{
		"--origin", srv.URL, "--version", version,
		"--manager", managerFile, "--manager-sha256", managerSHA256,
		"--home", home, "--hosts", "codex",
	}
}

// TestBootstrapRejectsIndexProductVersionMismatch covers bootstrap.go's own
// "index.ProductVersion != requestedVersion" check: a served index whose
// top-level version does not match the version bootstrap.sh already resolved
// and requested must be rejected, even though its single release entry (and
// the package it names) are otherwise completely valid.
func TestBootstrapRejectsIndexProductVersionMismatch(t *testing.T) {
	version := "9.2.1"
	srv, managerBytes, managerSHA256 := newBootstrapFixture(t, version, bootstrapFixtureConfig{
		indexProductVersion: version + "-evil",
	})
	managerFile := filepath.Join(t.TempDir(), "hive-manager")
	if err := os.WriteFile(managerFile, managerBytes, 0700); err != nil {
		t.Fatal(err)
	}
	home := resolvedTempDir(t)
	var out bytes.Buffer
	err := bootstrapWithAdapterFactory(bootstrapFixtureArgs(srv, version, managerFile, managerSHA256, home), strings.NewReader("y\n"), &out, true, coreOnlyAdapterFactory)
	if err == nil || !strings.Contains(err.Error(), "release index does not match the requested version") {
		t.Fatalf("accepted a release index whose own version does not match the request: %v", err)
	}
	if entries, statErr := os.ReadDir(home); statErr != nil || len(entries) != 0 {
		t.Fatalf("rejected index still wrote into home: %v %v", entries, statErr)
	}
}

// TestBootstrapRejectsReleaseNotBoundToVerifiedManager covers bootstrap.go's
// own "release.RawBinarySHA256 != managerSHA256" check: a release entry whose
// recorded manager checksum does not match the manager bootstrap.sh already
// verified must be rejected, even though its package checksum is genuinely
// correct — otherwise a stale or tampered index could serve unrelated
// package bytes under an already-trusted, already-running manager.
func TestBootstrapRejectsReleaseNotBoundToVerifiedManager(t *testing.T) {
	version := "9.2.2"
	srv, managerBytes, managerSHA256 := newBootstrapFixture(t, version, bootstrapFixtureConfig{
		releaseRawBinarySHA256: strings.Repeat("b", 64),
	})
	managerFile := filepath.Join(t.TempDir(), "hive-manager")
	if err := os.WriteFile(managerFile, managerBytes, 0700); err != nil {
		t.Fatal(err)
	}
	home := resolvedTempDir(t)
	var out bytes.Buffer
	err := bootstrapWithAdapterFactory(bootstrapFixtureArgs(srv, version, managerFile, managerSHA256, home), strings.NewReader("y\n"), &out, true, coreOnlyAdapterFactory)
	if err == nil || !strings.Contains(err.Error(), "release index does not match the verified manager") {
		t.Fatalf("accepted a release not bound to the verified manager: %v", err)
	}
	if entries, statErr := os.ReadDir(home); statErr != nil || len(entries) != 0 {
		t.Fatalf("unbound release still wrote into home: %v %v", entries, statErr)
	}
}

// TestBootstrapRejectsWrongPlatformOrTamperedPackage covers two independent
// gates a genuinely published, otherwise-valid release can still fail: no
// release entry for the running platform, and a downloaded package whose
// bytes no longer match their own recorded checksum (bootstrap.go's
// "distribution.Digest(packageData) != release.PackageSHA256"). Neither
// subtest ever reaches extraction or writes into home.
func TestBootstrapRejectsWrongPlatformOrTamperedPackage(t *testing.T) {
	t.Run("wrong platform", func(t *testing.T) {
		version := "9.2.3"
		srv, managerBytes, managerSHA256 := newBootstrapFixture(t, version, bootstrapFixtureConfig{
			releasePlatform: "plan9/386",
		})
		managerFile := filepath.Join(t.TempDir(), "hive-manager")
		if err := os.WriteFile(managerFile, managerBytes, 0700); err != nil {
			t.Fatal(err)
		}
		home := resolvedTempDir(t)
		var out bytes.Buffer
		err := bootstrapWithAdapterFactory(bootstrapFixtureArgs(srv, version, managerFile, managerSHA256, home), strings.NewReader("y\n"), &out, true, coreOnlyAdapterFactory)
		if err == nil || !strings.Contains(err.Error(), "no published release for") {
			t.Fatalf("accepted an index without a release for the running platform: %v", err)
		}
		if entries, statErr := os.ReadDir(home); statErr != nil || len(entries) != 0 {
			t.Fatalf("wrong-platform release still wrote into home: %v %v", entries, statErr)
		}
	})
	t.Run("tampered package", func(t *testing.T) {
		version := "9.2.4"
		srv, managerBytes, managerSHA256 := newBootstrapFixture(t, version, bootstrapFixtureConfig{
			corruptPackageBytes: true,
		})
		managerFile := filepath.Join(t.TempDir(), "hive-manager")
		if err := os.WriteFile(managerFile, managerBytes, 0700); err != nil {
			t.Fatal(err)
		}
		home := resolvedTempDir(t)
		var out bytes.Buffer
		err := bootstrapWithAdapterFactory(bootstrapFixtureArgs(srv, version, managerFile, managerSHA256, home), strings.NewReader("y\n"), &out, true, coreOnlyAdapterFactory)
		if err == nil || !strings.Contains(err.Error(), "downloaded package does not match its verified checksum") {
			t.Fatalf("accepted a package that no longer matches its verified checksum: %v", err)
		}
		if entries, statErr := os.ReadDir(home); statErr != nil || len(entries) != 0 {
			t.Fatalf("tampered package still wrote into home: %v %v", entries, statErr)
		}
	})
}

// TestBootstrapBindsRetainedInstallerIntoPlanJournal covers
// "management.BindInstaller(p, artifactID)": after consent, the plan
// actually applied must carry the just-retained installer's identity, not
// only leave that installer sitting under <state-dir>/installers/. hive
// recover (from another, offline terminal) resolves the retained manager
// through this binding, so it must be read back here from the persisted
// transaction journal, not merely from Go values already in this process.
func TestBootstrapBindsRetainedInstallerIntoPlanJournal(t *testing.T) {
	version := "9.2.5"
	srv, managerBytes, managerSHA256 := bootstrapFixtureOrigin(t, version)
	managerFile := filepath.Join(t.TempDir(), "hive-manager")
	if err := os.WriteFile(managerFile, managerBytes, 0700); err != nil {
		t.Fatal(err)
	}
	home := resolvedTempDir(t)
	stateDir := filepath.Join(home, "state")
	args := append(bootstrapFixtureArgs(srv, version, managerFile, managerSHA256, home), "--state-dir", stateDir)
	var out bytes.Buffer
	if err := bootstrapWithAdapterFactory(args, strings.NewReader("y\n"), &out, true, coreOnlyAdapterFactory); err != nil {
		t.Fatalf("bootstrap did not complete: %v\n%s", err, out.String())
	}
	installers, err := os.ReadDir(filepath.Join(stateDir, "installers"))
	if err != nil || len(installers) != 1 {
		t.Fatalf("installer was not retained after consent: %v %v", installers, err)
	}
	artifactID := installers[0].Name()

	txns, err := os.ReadDir(filepath.Join(stateDir, "transactions"))
	if err != nil || len(txns) != 1 {
		t.Fatalf("expected exactly one applied transaction: %v %v", txns, err)
	}
	data, err := os.ReadFile(filepath.Join(stateDir, "transactions", txns[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	var applied struct {
		Plan struct {
			Installer *struct{ ArtifactID string } `json:"installer"`
		} `json:"Plan"`
	}
	if err := json.Unmarshal(data, &applied); err != nil {
		t.Fatal(err)
	}
	if applied.Plan.Installer == nil || applied.Plan.Installer.ArtifactID != artifactID {
		t.Fatalf("applied plan's journal does not carry the retained installer binding: %+v", applied)
	}
}

// TestInstallOfflinePendingRecoveryPointsToInstallScript covers L3's offline
// side: a plain (non-bootstrap) install invocation that finds a pending core
// operation must still say to run ./install.sh again, exactly as before —
// that script exists in this flow, and re-running it is the correct offline
// recovery step.
func TestInstallOfflinePendingRecoveryPointsToInstallScript(t *testing.T) {
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "pending.json"), []byte(`{"id":"core"}`), 0600); err != nil {
		t.Fatal(err)
	}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.RecoverCore = func(string) (string, error) { return "core-id", nil }
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := installWithDependencies([]string{"--home", home, "--state-dir", stateDir, "--source", source}, strings.NewReader("y\n"), &out, true, dependencies); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "run ./install.sh again") {
		t.Fatalf("offline pending recovery lost its ./install.sh instruction: %s", out.String())
	}
}

// TestBootstrapOnlinePendingRecoveryPointsToRetainedManager covers L3's
// online side: bootstrap.sh deletes its own temporary manager on exit, so a
// pending core operation found under `hive bootstrap` must never tell the
// operator to run ./install.sh (a file that does not exist in this flow).
// Instead it must name the retained manager's absolute path and the state
// directory, so `hive recover` can finish the install later, offline, from
// another terminal.
func TestBootstrapOnlinePendingRecoveryPointsToRetainedManager(t *testing.T) {
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	retainedDir := filepath.Join(stateDir, "installers", strings.Repeat("c", 64))
	if err := os.MkdirAll(retainedDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(retainedDir, "manager"), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "pending.json"), []byte(`{"id":"core"}`), 0600); err != nil {
		t.Fatal(err)
	}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.RecoverCore = func(string) (string, error) { return "core-id", nil }
	// Only bootstrap.go ever sets this hook; runInstallFlow uses its mere
	// presence to detect the online flow (see install.go's "online" flag).
	dependencies.BindRetainedInstaller = func(o management.Options, p management.Plan) (management.Plan, error) {
		t.Fatal("BindRetainedInstaller must not run for an already-pending recovery")
		return p, nil
	}
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := runInstallFlow(management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source, Hosts: []string{"codex"}}, false, strings.NewReader("y\n"), &out, true, dependencies); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "./install.sh") {
		t.Fatalf("online pending recovery still pointed to ./install.sh: %s", out.String())
	}
	wantManager := filepath.Join(retainedDir, "manager")
	if !strings.Contains(out.String(), wantManager) || !strings.Contains(out.String(), "recover") || !strings.Contains(out.String(), stateDir) {
		t.Fatalf("online pending recovery did not name the retained manager and state dir: %s", out.String())
	}
}
