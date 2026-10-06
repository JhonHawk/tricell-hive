package distribution

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDownloadRejectsUnsafeLocationsAndOversizeResponses(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("12345")), Header: make(http.Header), Request: r}, nil
	})}
	if _, err := download(context.Background(), client, "https://downloads.example", "../hive", 10); err == nil {
		t.Fatal("accepted traversal resource")
	}
	if _, err := download(context.Background(), client, "http://downloads.example", "hive", 10); err == nil {
		t.Fatal("accepted insecure origin")
	}
	if _, err := download(context.Background(), client, "https://downloads.example", "hive", 4); err == nil {
		t.Fatal("accepted oversized response")
	}
}

func TestDownloadRejectsRedirectOutsideOrigin(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "downloads.example" {
			return &http.Response{StatusCode: http.StatusFound, Header: http.Header{"Location": []string{"https://evil.example/hive"}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("payload")), Request: r}, nil
	})}
	if _, err := download(context.Background(), client, "https://downloads.example", "hive", 10); err == nil {
		t.Fatal("followed redirect outside origin")
	}
}

// TestDownloadAllowsLoopbackHTTPOnlyWhenTestSeamIsEnabled covers the Go-side
// loopback-HTTP test seam: same-package tests set allowLoopbackHTTP directly
// (a test-built manager binary could also set it via -ldflags; the shell-level
// tests that did were retired with bootstrap.sh in 0.1.0) so their own httptest fixtures
// can use plain HTTP, without weakening production, which must still refuse
// HTTP to a loopback host when the seam is off — the default in every
// shipped binary, since production never sets allowLoopbackHTTPValue.
func TestDownloadAllowsLoopbackHTTPOnlyWhenTestSeamIsEnabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("payload"))
	}))
	defer srv.Close()
	client := &http.Client{}

	if _, err := download(context.Background(), client, srv.URL, "hive", 1024); err == nil {
		t.Fatal("accepted a loopback HTTP origin with the test seam off")
	}

	previous := allowLoopbackHTTP
	allowLoopbackHTTP = true
	defer func() { allowLoopbackHTTP = previous }()

	data, err := download(context.Background(), client, srv.URL, "hive", 1024)
	if err != nil || string(data) != "payload" {
		t.Fatalf("download() = %q, %v; want it allowed once allowLoopbackHTTP is set", data, err)
	}
	if _, err := download(context.Background(), client, "https://evil.example", "hive", 1024); err == nil {
		t.Fatal("allowLoopbackHTTP must never widen acceptance of a non-loopback origin")
	}
}

// TestParseOriginRejectsHTTPWhenLoopbackSeamIsOff proves production's default
// posture directly: with the test seam off (as it always is in a shipped
// binary, since allowLoopbackHTTPValue is only ever set by a test -ldflags
// build), parseOrigin refuses plain HTTP even to a loopback host.
func TestParseOriginRejectsHTTPWhenLoopbackSeamIsOff(t *testing.T) {
	if allowLoopbackHTTP {
		t.Fatal("test seam must be off by default for this test; another test left it enabled")
	}
	for _, value := range []string{"http://127.0.0.1:8443", "http://localhost:8443", "http://example.com"} {
		if _, err := parseOrigin(value); err == nil {
			t.Fatalf("parseOrigin(%q) = nil error, want rejection with the loopback seam off", value)
		}
	}
}

func TestExtractRejectsUnsafeArchiveWithoutEscape(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "outside")
	archive := testArchive(t, []tar.Header{{Name: "hive-1.2.3/../../outside", Typeflag: tar.TypeReg, Mode: 0600, Size: 1}}, []string{"x"})
	if _, err := Extract(archive, t.TempDir()); err == nil {
		t.Fatal("accepted traversal")
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatalf("archive escaped destination: %v", err)
	}
}

func TestExtractRejectsLinksAndDuplicates(t *testing.T) {
	for name, headers := range map[string][]tar.Header{
		"link": {{Name: "hive-1.2.3/link", Typeflag: tar.TypeSymlink, Linkname: "elsewhere"}},
		"duplicate": {
			{Name: "hive-1.2.3/file", Typeflag: tar.TypeReg, Mode: 0600, Size: 1},
			{Name: "hive-1.2.3/file", Typeflag: tar.TypeReg, Mode: 0600, Size: 1},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Extract(testArchive(t, headers, []string{"x", "y"}), t.TempDir()); err == nil {
				t.Fatal("accepted unsafe archive")
			}
		})
	}
}

func TestRetainVerifiedInstallerIsPrivateAndNeverOverwrites(t *testing.T) {
	root := t.TempDir()
	manager := filepath.Join(root, "manager")
	pkg := filepath.Join(root, "package.tar.gz")
	if err := os.WriteFile(manager, []byte("manager"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pkg, []byte("package"), 0600); err != nil {
		t.Fatal(err)
	}
	id := strings.Repeat("a", 64)
	retained, err := RetainVerifiedInstaller(filepath.Join(root, "state"), id, VerifiedFile{Path: manager, SHA256: Digest([]byte("manager"))}, VerifiedFile{Path: pkg, SHA256: Digest([]byte("package"))})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RetainedInstallerAt(filepath.Join(root, "state"), id); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(retained.Manager); err != nil || info.Mode().Perm() != 0700 {
		t.Fatalf("manager permissions = %v, %v", info.Mode(), err)
	}
	again, err := RetainVerifiedInstaller(filepath.Join(root, "state"), id, VerifiedFile{Path: manager, SHA256: Digest([]byte("manager"))}, VerifiedFile{Path: pkg, SHA256: Digest([]byte("package"))})
	if err != nil || again.Directory != retained.Directory {
		t.Fatalf("did not reuse retained installer: %+v, %v", again, err)
	}
}

func TestRetainVerifiedInstallerNeverReplacesARecordedInstallerThatFailsValidation(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	manager := filepath.Join(root, "manager")
	pkg := filepath.Join(root, "package.tar.gz")
	if err := os.WriteFile(manager, []byte("manager"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pkg, []byte("package"), 0600); err != nil {
		t.Fatal(err)
	}
	id := strings.Repeat("e", 64)
	retained, err := RetainVerifiedInstaller(state, id, VerifiedFile{Path: manager, SHA256: Digest([]byte("manager"))}, VerifiedFile{Path: pkg, SHA256: Digest([]byte("package"))})
	if err != nil {
		t.Fatal(err)
	}
	// A published directory always carries its retention record, so a record
	// that fails validation is evidence to preserve, not a crash leftover.
	if err := os.WriteFile(retained.Package, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := RetainVerifiedInstaller(state, id, VerifiedFile{Path: manager, SHA256: Digest([]byte("manager"))}, VerifiedFile{Path: pkg, SHA256: Digest([]byte("package"))}); err == nil {
		t.Fatal("replaced a recorded retained installer that failed validation")
	}
	if data, _ := os.ReadFile(retained.Package); string(data) != "tampered" {
		t.Fatalf("recorded installer was rewritten: %q", data)
	}
}

func TestRetainVerifiedInstallerRejectsIdentityCollision(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	manager := filepath.Join(root, "manager")
	pkg := filepath.Join(root, "package.tar.gz")
	other := filepath.Join(root, "other.tar.gz")
	for path, data := range map[string]string{manager: "manager", pkg: "package", other: "other"} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(manager, 0700); err != nil {
		t.Fatal(err)
	}
	id := strings.Repeat("f", 64)
	retained, err := RetainVerifiedInstaller(state, id, VerifiedFile{Path: manager, SHA256: Digest([]byte("manager"))}, VerifiedFile{Path: pkg, SHA256: Digest([]byte("package"))})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RetainVerifiedInstaller(state, id, VerifiedFile{Path: manager, SHA256: Digest([]byte("manager"))}, VerifiedFile{Path: other, SHA256: Digest([]byte("other"))}); err == nil {
		t.Fatal("rebound an artifact ID to different content")
	}
	if data, _ := os.ReadFile(retained.Package); string(data) != "package" {
		t.Fatalf("retained package was replaced: %q", data)
	}
}

func TestRetainedInstallerAtRejectsTamperingAndLinkedState(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state")
	manager := filepath.Join(root, "manager")
	pkg := filepath.Join(root, "package.tar.gz")
	if err := os.WriteFile(manager, []byte("manager"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pkg, []byte("package"), 0600); err != nil {
		t.Fatal(err)
	}
	id := strings.Repeat("c", 64)
	retained, err := RetainVerifiedInstaller(state, id, VerifiedFile{Path: manager, SHA256: Digest([]byte("manager"))}, VerifiedFile{Path: pkg, SHA256: Digest([]byte("package"))})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(retained.Package, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := RetainedInstallerAt(state, id); err == nil {
		t.Fatal("accepted retained package content that changed")
	}
	linkedState := filepath.Join(root, "linked-state")
	if err := os.Symlink(t.TempDir(), linkedState); err != nil {
		t.Fatal(err)
	}
	if _, err := RetainVerifiedInstaller(linkedState, strings.Repeat("d", 64), VerifiedFile{Path: manager, SHA256: Digest([]byte("manager"))}, VerifiedFile{Path: pkg, SHA256: Digest([]byte("package"))}); err == nil {
		t.Fatal("accepted linked state directory")
	}
}

func TestValidateBootstrapPackageBindsVersionsAndManifestIdentity(t *testing.T) {
	root := packageFixture(t)
	data, err := os.ReadFile(filepath.Join(root, ManifestName))
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.ProductVersion = "1.2.3"
	data, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ManifestName), data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, artifactID, err := ValidateBootstrapPackage(root, "1.2.3", "1.2.3"); err != nil || artifactID != Digest(data) {
		t.Fatalf("ValidateBootstrapPackage() = %q, %v", artifactID, err)
	}
	if _, _, err := ValidateBootstrapPackage(root, "1.2.3", "1.2.4"); err == nil {
		t.Fatal("accepted manager version mismatch")
	}
}

func TestRetainVerifiedInstallerRejectsUnsafeOrChangedInput(t *testing.T) {
	root := t.TempDir()
	manager := filepath.Join(root, "manager")
	pkg := filepath.Join(root, "package.tar.gz")
	if err := os.WriteFile(manager, []byte("manager"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pkg, []byte("package"), 0600); err != nil {
		t.Fatal(err)
	}
	id := strings.Repeat("b", 64)
	if _, err := RetainVerifiedInstaller(filepath.Join(root, "state"), id, VerifiedFile{Path: manager, SHA256: Digest([]byte("different"))}, VerifiedFile{Path: pkg, SHA256: Digest([]byte("package"))}); err == nil {
		t.Fatal("accepted changed manager")
	}
	if err := os.Chmod(pkg, 0666); err != nil {
		t.Fatal(err)
	}
	if _, err := RetainVerifiedInstaller(filepath.Join(root, "state"), id, VerifiedFile{Path: manager, SHA256: Digest([]byte("manager"))}, VerifiedFile{Path: pkg, SHA256: Digest([]byte("package"))}); err == nil {
		t.Fatal("accepted world-writable package")
	}
}

// TestRetainVerifiedInstallerRecoversFromCrashedPriorAttempt covers H5: a
// process that crashes between os.Mkdir(dir) and the retention record being
// written must not leave every later retry permanently failing with
// "retained installer identity is missing or invalid".
func TestRetainVerifiedInstallerRecoversFromCrashedPriorAttempt(t *testing.T) {
	for name, seed := range map[string]func(t *testing.T, dir string){
		"bare_directory": func(t *testing.T, dir string) {
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
		},
		"partial_files_no_record": func(t *testing.T, dir string) {
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "manager"), []byte("stale"), 0700); err != nil {
				t.Fatal(err)
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			state := filepath.Join(root, "state")
			manager := filepath.Join(root, "manager")
			pkg := filepath.Join(root, "package.tar.gz")
			if err := os.WriteFile(manager, []byte("manager"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(pkg, []byte("package"), 0600); err != nil {
				t.Fatal(err)
			}
			id := strings.Repeat("e", 64)
			seed(t, filepath.Join(state, "installers", id))
			retained, err := RetainVerifiedInstaller(state, id, VerifiedFile{Path: manager, SHA256: Digest([]byte("manager"))}, VerifiedFile{Path: pkg, SHA256: Digest([]byte("package"))})
			if err != nil {
				t.Fatalf("did not recover from a crashed prior attempt: %v", err)
			}
			again, err := RetainedInstallerAt(state, id)
			if err != nil || again.ManagerDigest != retained.ManagerDigest || again.PackageDigest != retained.PackageDigest {
				t.Fatalf("recovered retention is not valid: %+v, %v", again, err)
			}
		})
	}
}

func testArchive(t *testing.T, headers []tar.Header, data []string) []byte {
	t.Helper()
	var result bytes.Buffer
	gz := gzip.NewWriter(&result)
	tw := tar.NewWriter(gz)
	for i, h := range headers {
		if err := tw.WriteHeader(&h); err != nil {
			t.Fatal(err)
		}
		if h.Size > 0 {
			if _, err := tw.Write([]byte(data[i])); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return result.Bytes()
}
