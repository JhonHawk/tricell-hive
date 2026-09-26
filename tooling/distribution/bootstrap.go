package distribution

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"tricell-hive/tooling/version"
)

const (
	MaxIndexBytes     int64 = 1 << 20
	MaxPackageBytes   int64 = 128 << 20
	MaxExtractedBytes int64 = 256 << 20
	MaxArchiveEntries       = 10000
)

// DownloadIndex is the trusted data contract between a verified raw manager
// and a versioned package. Its schema is independent from Manifest.Version.
type DownloadIndex struct {
	Version        int               `json:"version"`
	ProductVersion string            `json:"product_version"`
	Releases       []DownloadRelease `json:"releases"`
}

type DownloadRelease struct {
	Platform        string `json:"platform"`
	SourceID        string `json:"source_id"`
	Package         string `json:"package"`
	PackageSHA256   string `json:"package_sha256"`
	RawBinary       string `json:"raw_binary"`
	RawBinarySHA256 string `json:"raw_binary_sha256"`
}

// WriteDownloadIndex creates a new index without overwriting a prior artifact.
func WriteDownloadIndex(filename string, index DownloadIndex) error {
	if _, err := ParseDownloadIndex(mustJSON(index)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	data := append(mustJSON(index), '\n')
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func mustJSON(value any) []byte {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	return data
}

// ParseDownloadIndex validates download paths and digests before any request
// uses them. It accepts only the current download-index schema.
func ParseDownloadIndex(data []byte) (DownloadIndex, error) {
	if int64(len(data)) > MaxIndexBytes {
		return DownloadIndex{}, fmt.Errorf("download index exceeds limit")
	}
	var index DownloadIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return DownloadIndex{}, fmt.Errorf("invalid download index: %w", err)
	}
	if index.Version != 1 || !version.Valid(index.ProductVersion) || len(index.Releases) == 0 {
		return DownloadIndex{}, fmt.Errorf("incompatible download index")
	}
	seen := make(map[string]bool, len(index.Releases))
	for _, release := range index.Releases {
		if release.Platform == "" || seen[release.Platform] || !validDigest(release.SourceID) || !validDigest(release.PackageSHA256) || !validDigest(release.RawBinarySHA256) || !validDownloadPath(release.Package) || !validDownloadPath(release.RawBinary) {
			return DownloadIndex{}, fmt.Errorf("invalid download index release")
		}
		seen[release.Platform] = true
	}
	return index, nil
}

// ValidateBootstrapPackage verifies the archive selected by the index before
// a caller journals or retains it. Public bootstrap packages must identify the
// requested release; legacy manifests remain valid for offline compatibility.
func ValidateBootstrapPackage(root, requestedVersion, runningVersion string) (Manifest, string, error) {
	if requestedVersion == "dev" || !version.Valid(requestedVersion) || runningVersion != requestedVersion {
		return Manifest{}, "", fmt.Errorf("bootstrap version does not match its verified manager")
	}
	if err := VerifyIfPackaged(root); err != nil {
		return Manifest{}, "", err
	}
	manifest, err := ReadManifest(root)
	if err != nil {
		return Manifest{}, "", err
	}
	if manifest.ProductVersion != requestedVersion {
		return Manifest{}, "", fmt.Errorf("package product version does not match requested release")
	}
	data, err := os.ReadFile(filepath.Join(root, ManifestName))
	if err != nil {
		return Manifest{}, "", err
	}
	return manifest, Digest(data), nil
}

// Download fetches a fixed relative resource over HTTPS. Redirects may remain
// under origin, but no request may leave its scheme, host, or port. Production
// callers cannot substitute transport behavior; tests use the package-private
// download helper with a fake client.
func Download(ctx context.Context, origin, resource string, maxBytes int64) ([]byte, error) {
	return download(ctx, &http.Client{Timeout: time.Minute}, origin, resource, maxBytes)
}

func download(ctx context.Context, client *http.Client, origin, resource string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("download limit must be positive")
	}
	base, err := parseOrigin(origin)
	if err != nil {
		return nil, err
	}
	if !validDownloadPath(resource) {
		return nil, fmt.Errorf("invalid download resource")
	}
	requestURL := *base
	requestURL.Path = path.Join(base.Path, resource)
	if !strings.HasPrefix(requestURL.Path, "/") {
		requestURL.Path = "/" + requestURL.Path
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	if client == nil {
		client = http.DefaultClient
	}
	clone := *client
	clone.Timeout = boundedTimeout(clone.Timeout)
	clone.CheckRedirect = func(next *http.Request, via []*http.Request) error {
		if sameOrigin(base, next.URL) {
			return nil
		}
		return fmt.Errorf("redirect leaves trusted origin")
	}
	response, err := clone.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if !sameOrigin(base, response.Request.URL) {
		return nil, fmt.Errorf("response leaves trusted origin")
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: %s", response.Status)
	}
	if response.ContentLength > maxBytes {
		return nil, fmt.Errorf("download exceeds limit")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("download exceeds limit")
	}
	return data, nil
}

func boundedTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 || timeout > time.Minute {
		return time.Minute
	}
	return timeout
}

func parseOrigin(value string) (*url.URL, error) {
	origin, err := url.Parse(value)
	if err != nil || origin.Host == "" || origin.RawQuery != "" || origin.Fragment != "" || origin.User != nil {
		return nil, fmt.Errorf("trusted origin must be an absolute HTTPS origin")
	}
	if origin.Scheme == "https" {
		return origin, nil
	}
	if origin.Scheme == "http" && allowLoopbackHTTP && isLoopbackHost(origin.Hostname()) {
		return origin, nil
	}
	return nil, fmt.Errorf("trusted origin must be an absolute HTTPS origin")
}

// allowLoopbackHTTPValue is a build-time-only test seam: a test-built manager
// binary sets it via
// "-ldflags -X tricell-hive/tooling/distribution.allowLoopbackHTTPValue=true"
// (see bootstrap_shell_test.go's buildRealManager). It must be a string
// because -X can only overwrite a package-level string variable at link
// time. Production binaries built by tooling/package never pass that flag,
// so it stays empty and allowLoopbackHTTP stays false in every shipped
// artifact.
var allowLoopbackHTTPValue string

// allowLoopbackHTTP reports whether this build accepts a loopback HTTP
// origin. Same-package Go tests may instead set this variable directly
// (see TestDownloadAllowsLoopbackHTTPOnlyWhenTestSeamIsEnabled) rather than
// rebuilding the binary. It is never derived from an environment variable,
// so nothing at runtime can widen production's HTTPS-only acceptance.
var allowLoopbackHTTP = allowLoopbackHTTPValue == "true"

func isLoopbackHost(hostname string) bool {
	switch hostname {
	case "127.0.0.1", "localhost":
		return true
	default:
		return false
	}
}

func sameOrigin(origin, candidate *url.URL) bool {
	return candidate != nil && candidate.Scheme == origin.Scheme && candidate.Host == origin.Host
}

func validDownloadPath(value string) bool {
	return value != "" && !strings.HasPrefix(value, "/") && !strings.Contains(value, "\\") && path.Clean(value) == value && value != "." && !strings.HasPrefix(value, "../")
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

// Extract expands a verified archive under a new private directory. The
// returned root is the archive's one allowed top-level directory.
func Extract(archive []byte, destination string) (root string, err error) {
	if int64(len(archive)) > MaxPackageBytes {
		return "", fmt.Errorf("package exceeds compressed limit")
	}
	parent, err := os.MkdirTemp(destination, ".hive-extract-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(parent)
		}
	}()
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return "", fmt.Errorf("invalid compressed package: %w", err)
	}
	defer gz.Close()
	reader := tar.NewReader(gz)
	seen := map[string]bool{}
	var prefix string
	var count int
	var extracted int64
	for {
		header, nextErr := reader.Next()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return "", fmt.Errorf("invalid package archive: %w", nextErr)
		}
		count++
		if count > MaxArchiveEntries {
			return "", fmt.Errorf("package has too many entries")
		}
		name, entryPrefix, err := archivePath(header.Name)
		if err != nil {
			return "", err
		}
		if prefix == "" {
			if name != entryPrefix || header.Typeflag != tar.TypeDir {
				return "", fmt.Errorf("package must begin with one top-level directory")
			}
			prefix = entryPrefix
		} else if prefix != entryPrefix || (name != prefix && !strings.HasPrefix(name, prefix+"/")) {
			return "", fmt.Errorf("package has inconsistent top-level paths")
		}
		if seen[name] {
			return "", fmt.Errorf("package has duplicate path: %s", name)
		}
		seen[name] = true
		if header.Typeflag != tar.TypeDir && header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			return "", fmt.Errorf("package contains unsupported archive entry")
		}
		if header.Size < 0 || extracted+header.Size > MaxExtractedBytes {
			return "", fmt.Errorf("package exceeds extracted limit")
		}
		target := filepath.Join(parent, filepath.FromSlash(name))
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(target, os.FileMode(header.Mode)&0777); err != nil {
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return "", err
		}
		file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.FileMode(header.Mode)&0777)
		if err != nil {
			return "", err
		}
		written, copyErr := io.Copy(file, io.LimitReader(reader, header.Size))
		closeErr := file.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		if written != header.Size {
			return "", fmt.Errorf("truncated package entry: %s", name)
		}
		extracted += written
	}
	if prefix == "" {
		return "", fmt.Errorf("package is empty")
	}
	return filepath.Join(parent, filepath.FromSlash(prefix)), nil
}

func archivePath(name string) (string, string, error) {
	if name == "" || strings.HasPrefix(name, "/") || strings.Contains(name, "\\") || path.Clean(name) != name || strings.HasPrefix(name, "../") {
		return "", "", fmt.Errorf("unsafe package path: %q", name)
	}
	prefix, _, _ := strings.Cut(name, "/")
	if prefix == "." || prefix == ".." || prefix == "" {
		return "", "", fmt.Errorf("unsafe package path: %q", name)
	}
	return name, prefix, nil
}

type RetainedInstaller struct {
	ArtifactID    string
	ManagerDigest string
	PackageDigest string
	Directory     string
	Manager       string
	Package       string
}

const retentionName = "retention.json"

type retentionRecord struct {
	ArtifactID    string `json:"artifact_id"`
	ManagerDigest string `json:"manager_digest"`
	PackageDigest string `json:"package_digest"`
}

// VerifiedFile names a private, regular file and the digest verified before
// consent. Retention checks this digest again while copying it.
type VerifiedFile struct {
	Path   string
	SHA256 string
}

// RetainVerifiedInstaller copies already verified inputs into private state
// after consent. It never replaces a previously retained artifact.
func RetainVerifiedInstaller(stateDir, artifactID string, manager, packageArchive VerifiedFile) (RetainedInstaller, error) {
	if !filepath.IsAbs(stateDir) || !validDigest(artifactID) || !validDigest(manager.SHA256) || !validDigest(packageArchive.SHA256) {
		return RetainedInstaller{}, fmt.Errorf("invalid artifact identity")
	}
	if err := ensurePrivateDir(stateDir); err != nil {
		return RetainedInstaller{}, err
	}
	base := filepath.Join(stateDir, "installers")
	if err := ensurePrivateDir(base); err != nil {
		return RetainedInstaller{}, err
	}
	dir := filepath.Join(base, artifactID)
	if retained, ok, err := reuseRetainedInstaller(stateDir, dir, artifactID, manager, packageArchive); err != nil {
		return RetainedInstaller{}, err
	} else if ok {
		return retained, nil
	}
	// Stage every write in a sibling temporary directory and publish it with a
	// single atomic rename. A crash before that rename leaves only an orphaned
	// staging directory next to dir, never a dir that looks retained but holds
	// partial content or no retention.json; the next call's reuse check above
	// then safely replaces such a leftover instead of failing forever.
	staging, err := os.MkdirTemp(base, ".staging-"+artifactID+"-")
	if err != nil {
		return RetainedInstaller{}, err
	}
	defer os.RemoveAll(staging) // no-op once the rename below succeeds
	if err := checkPrivateDir(staging); err != nil {
		return RetainedInstaller{}, err
	}
	result := RetainedInstaller{ArtifactID: artifactID, ManagerDigest: manager.SHA256, PackageDigest: packageArchive.SHA256, Directory: dir, Manager: filepath.Join(dir, "manager"), Package: filepath.Join(dir, "package.tar.gz")}
	if err := copyRetainedFile(manager, filepath.Join(staging, "manager"), 0700); err != nil {
		return RetainedInstaller{}, err
	}
	if err := copyRetainedFile(packageArchive, filepath.Join(staging, "package.tar.gz"), 0600); err != nil {
		return RetainedInstaller{}, err
	}
	if err := writeRetentionRecord(filepath.Join(staging, retentionName), retentionRecord{ArtifactID: artifactID, ManagerDigest: manager.SHA256, PackageDigest: packageArchive.SHA256}); err != nil {
		return RetainedInstaller{}, err
	}
	if err := os.Rename(staging, dir); err != nil {
		// A concurrent caller may have published dir first; accept its result
		// only if it verifiably retains the same identity we were staging.
		if retained, ok, readErr := reuseRetainedInstaller(stateDir, dir, artifactID, manager, packageArchive); readErr == nil && ok {
			return retained, nil
		}
		return RetainedInstaller{}, err
	}
	return result, nil
}

// reuseRetainedInstaller reports whether dir already holds a validly
// retained installer for artifactID. A dir that exists but has no valid
// retention record (an interrupted prior attempt) is treated as absent: it is
// removed so the caller can safely stage a fresh one in its place.
func reuseRetainedInstaller(stateDir, dir, artifactID string, manager, packageArchive VerifiedFile) (RetainedInstaller, bool, error) {
	info, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return RetainedInstaller{}, false, nil
	}
	if err != nil {
		return RetainedInstaller{}, false, err
	}
	if !info.IsDir() {
		return RetainedInstaller{}, false, fmt.Errorf("retained installer path is not a directory")
	}
	retained, readErr := RetainedInstallerAt(stateDir, artifactID)
	if readErr == nil {
		if retained.ManagerDigest != manager.SHA256 || retained.PackageDigest != packageArchive.SHA256 {
			return RetainedInstaller{}, false, fmt.Errorf("retained artifact identity collision")
		}
		return retained, true, nil
	}
	// Publication renames a complete staging directory, so only a directory
	// without its retention record is an interrupted attempt. A recorded one
	// that fails validation (tampering, permissions, I/O) is preserved.
	if _, err := os.Lstat(filepath.Join(dir, retentionName)); !os.IsNotExist(err) {
		if err != nil {
			return RetainedInstaller{}, false, err
		}
		return RetainedInstaller{}, false, fmt.Errorf("retained installer failed validation; preserved: %w", readErr)
	}
	if err := os.RemoveAll(dir); err != nil {
		return RetainedInstaller{}, false, err
	}
	return RetainedInstaller{}, false, nil
}

// RetainedInstallerAt validates the retained files before an offline recovery
// consumes them.
func RetainedInstallerAt(stateDir, artifactID string) (RetainedInstaller, error) {
	if !validDigest(artifactID) {
		return RetainedInstaller{}, fmt.Errorf("invalid artifact identity")
	}
	if !filepath.IsAbs(stateDir) || !validDigest(artifactID) {
		return RetainedInstaller{}, fmt.Errorf("invalid artifact identity")
	}
	if err := checkPrivateDir(stateDir); err != nil {
		return RetainedInstaller{}, err
	}
	dir := filepath.Join(stateDir, "installers", artifactID)
	if err := checkPrivateDir(filepath.Join(stateDir, "installers")); err != nil {
		return RetainedInstaller{}, err
	}
	if err := checkPrivateDir(dir); err != nil {
		return RetainedInstaller{}, err
	}
	result := RetainedInstaller{ArtifactID: artifactID, Directory: dir, Manager: filepath.Join(dir, "manager"), Package: filepath.Join(dir, "package.tar.gz")}
	record, err := readRetentionRecord(filepath.Join(dir, retentionName))
	if err != nil || record.ArtifactID != artifactID || !validDigest(record.ManagerDigest) || !validDigest(record.PackageDigest) {
		return RetainedInstaller{}, fmt.Errorf("retained installer identity is missing or invalid")
	}
	result.ManagerDigest = record.ManagerDigest
	result.PackageDigest = record.PackageDigest
	for _, item := range []struct {
		path string
		mode os.FileMode
		want string
	}{{result.Manager, 0700, record.ManagerDigest}, {result.Package, 0600, record.PackageDigest}} {
		info, err := os.Lstat(item.path)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != item.mode {
			return RetainedInstaller{}, fmt.Errorf("retained installer is missing or has unsafe permissions")
		}
		data, err := os.ReadFile(item.path)
		if err != nil {
			return RetainedInstaller{}, err
		}
		if Digest(data) != item.want {
			return RetainedInstaller{}, fmt.Errorf("retained installer content does not match its identity")
		}
	}
	return result, nil
}

func writeRetentionRecord(filename string, record retentionRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(append(data, '\n'))
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func readRetentionRecord(filename string) (retentionRecord, error) {
	info, err := os.Lstat(filename)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0600 {
		return retentionRecord{}, fmt.Errorf("retained installer identity is missing or unsafe")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return retentionRecord{}, err
	}
	var record retentionRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return retentionRecord{}, err
	}
	return record, nil
}

func ensurePrivateDir(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return checkPrivateDir(dir)
}

func checkPrivateDir(dir string) error {
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0022 != 0 {
		return fmt.Errorf("installer directory is missing or unsafe")
	}
	return nil
}

func copyRetainedFile(source VerifiedFile, destination string, mode os.FileMode) error {
	info, err := os.Lstat(source.Path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0022 != 0 || info.Size() > MaxPackageBytes {
		return fmt.Errorf("verified input is not a regular file")
	}
	input, err := os.Open(source.Path)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	data, copyErr := io.ReadAll(io.LimitReader(input, MaxPackageBytes+1))
	if int64(len(data)) > MaxPackageBytes || Digest(data) != source.SHA256 {
		copyErr = fmt.Errorf("verified input digest changed")
	}
	var writeErr error
	if copyErr == nil {
		_, writeErr = output.Write(data)
	}
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	if writeErr != nil {
		return writeErr
	}
	retained, err := os.ReadFile(destination)
	if err != nil || Digest(retained) != source.SHA256 {
		return fmt.Errorf("retained installer digest mismatch")
	}
	return closeErr
}
