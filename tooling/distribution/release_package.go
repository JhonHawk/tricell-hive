package distribution

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// PreparedRelease owns disposable inputs until Cleanup. Binding's final path
// is usable for rendering a frozen plan before anything is retained there.
type PreparedRelease struct {
	Binding          UpdatePackageBinding
	privateDirectory string
	stateVersion     int
}

func (p *PreparedRelease) Cleanup() error {
	if p.privateDirectory == "" {
		return nil
	}
	err := removeReleasePrivate(p.privateDirectory)
	if err == nil {
		p.privateDirectory = ""
	}
	return err
}

// Archive permissions must not prevent disposal of our own private inputs.
func removeReleasePrivate(root string) error {
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.Chmod(path, 0700)
		}
		return nil
	})
	return os.RemoveAll(root)
}

// checkCanonicalAncestors also works for destinations that do not yet exist.
// It rejects symlinks in every existing ancestor rather than resolving them.
func checkCanonicalAncestors(path string) error {
	if !canonicalAbsolute(path) {
		return fmt.Errorf("package path must be canonical and absolute")
	}
	for current := path; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return fmt.Errorf("package path contains a link or non-directory")
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		if filepath.Dir(current) == current {
			return nil
		}
	}
}

func (c ReleaseClient) Prepare(ctx context.Context, target ReleaseTarget, store string, stateVersion int) (prepared *PreparedRelease, err error) {
	// Reconstruct every remote address and label; callers cannot inject URLs or
	// traversal through a forged target value.
	expected, e := (ReleaseClient{}).Resolve(ctx, target.ProductVersion, target.ProductVersion, target.Platform)
	if e != nil || target.Label != expected.Label || target.ArchiveURL != expected.ArchiveURL {
		return nil, fmt.Errorf("invalid release target")
	}
	if err = checkCanonicalAncestors(store); err != nil {
		return nil, err
	}
	checksum, err := c.fetch(ctx, target.ArchiveURL+".sha256", 4096)
	if err != nil {
		return nil, err
	}
	suffix := "  " + target.Label + ".tar.gz\n"
	if len(checksum) != 64+len(suffix) || !strings.HasSuffix(string(checksum), suffix) || !ValidSHA256(string(checksum[:64])) {
		return nil, fmt.Errorf("release checksum has invalid digest or archive name")
	}
	archive, err := c.fetch(ctx, target.ArchiveURL, MaxPackageBytes)
	if err != nil {
		return nil, err
	}
	archiveDigest := Digest(archive)
	if archiveDigest != string(checksum[:64]) {
		return nil, fmt.Errorf("release archive checksum mismatch")
	}
	// tar EOF can precede the gzip footer. Consume and bound the full compressed
	// stream before extraction so truncation and CRC errors cannot be ignored.
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("invalid compressed package: %w", err)
	}
	const streamLimit = MaxExtractedBytes + int64(MaxArchiveEntries)*1024 + 1<<20
	n, readErr := io.Copy(io.Discard, io.LimitReader(gz, streamLimit+1))
	closeErr := gz.Close()
	if readErr != nil || closeErr != nil || n > streamLimit {
		return nil, fmt.Errorf("invalid or oversized compressed package")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	private, err := os.MkdirTemp("", "hive-release-")
	if err != nil {
		return nil, err
	}
	// macOS's default temporary root may itself have a system symlink. Resolve
	// our private directory once; externally supplied store paths remain strict.
	resolved, err := filepath.EvalSymlinks(private)
	if err != nil {
		_ = removeReleasePrivate(private)
		return nil, err
	}
	private = resolved
	defer func() {
		if err != nil {
			_ = removeReleasePrivate(private)
		}
	}()
	root, err := Extract(archive, private)
	if err != nil {
		return nil, err
	}
	if filepath.Base(root) != target.Label {
		return nil, fmt.Errorf("release archive has unexpected package label")
	}
	artifact, err := verifyReleaseDirectory(root, target.ProductVersion, target.Platform, archiveDigest, stateVersion)
	if err != nil {
		return nil, err
	}
	final := filepath.Join(store, target.Label)
	binding := UpdatePackageBinding{Artifact: artifact, PreparationDirectory: root, PackageDirectory: final, PreparationExecutor: ExecutorIdentity{Path: filepath.Join(root, "bin", "hive"), SHA256: artifact.Bootstrap.ExecutableSHA256}, TargetExecutable: ExecutorIdentity{Path: filepath.Join(final, "bin", "hive"), SHA256: artifact.Bootstrap.ExecutableSHA256}}
	if err = binding.Validate(stateVersion); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &PreparedRelease{Binding: binding, privateDirectory: private, stateVersion: stateVersion}, nil
}

func verifyReleaseDirectory(root, product, platform, archiveDigest string, stateVersion int) (UpdateArtifact, error) {
	var result UpdateArtifact
	if err := checkCanonicalAncestors(root); err != nil {
		return result, err
	}
	// Revalidation of a retained tree has the same resource bounds as initial
	// extraction, including when a previously valid file has since grown.
	var entries int
	var total int64
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		entries++
		if entries > MaxArchiveEntries {
			return fmt.Errorf("package has too many entries")
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("package contains non-regular entry")
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Size() < 0 || info.Size() > MaxExtractedBytes-total {
			return fmt.Errorf("package exceeds extracted limit")
		}
		total += info.Size()
		return nil
	}); err != nil {
		return result, err
	}
	if err := VerifyIfPackaged(root); err != nil {
		return result, err
	}
	m, data, err := ReadManifestData(root)
	if err != nil {
		return result, err
	}
	if m.ProductVersion != product || m.Platform != platform {
		return result, fmt.Errorf("release package identity mismatch")
	}
	bootstrap, err := ReadUpdateBootstrap(root, m, stateVersion)
	if err != nil {
		return result, err
	}
	info, err := os.Lstat(filepath.Join(root, "bin", "hive"))
	if err != nil {
		return result, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		return result, fmt.Errorf("release executor is not an executable regular file")
	}
	return UpdateArtifact{ProductVersion: product, Platform: platform, Label: filepath.Base(root), ArchiveSHA256: archiveDigest, ManifestSHA256: Digest(data), SourceID: m.SourceID, Bootstrap: bootstrap}, nil
}

// PrepareLocalRelease verifies a complete offline package without taking
// ownership of the caller's directory. Retain copies and rechecks its pinned
// bytes; Cleanup never removes the imported source.
func PrepareLocalRelease(root, store string, stateVersion int) (*PreparedRelease, error) {
	if err := checkCanonicalAncestors(root); err != nil {
		return nil, err
	}
	if err := checkCanonicalAncestors(store); err != nil {
		return nil, err
	}
	m, err := ReadManifest(root)
	if err != nil {
		return nil, err
	}
	artifact, err := verifyReleaseDirectory(root, m.ProductVersion, m.Platform, "", stateVersion)
	if err != nil {
		return nil, err
	}
	artifact.SourceKind = "directory"
	// An offline extraction may be renamed by its owner; its immutable store
	// label comes from verified product/platform identity rather than that name.
	artifact.Label = "hive-" + m.ProductVersion + "-" + strings.ReplaceAll(m.Platform, "/", "-")
	final := filepath.Join(store, artifact.Label)
	b := UpdatePackageBinding{Artifact: artifact, PreparationDirectory: root, PackageDirectory: final,
		PreparationExecutor: ExecutorIdentity{Path: filepath.Join(root, "bin", "hive"), SHA256: artifact.Bootstrap.ExecutableSHA256},
		TargetExecutable:    ExecutorIdentity{Path: filepath.Join(final, "bin", "hive"), SHA256: artifact.Bootstrap.ExecutableSHA256}}
	if err := b.Validate(stateVersion); err != nil {
		return nil, err
	}
	return &PreparedRelease{Binding: b, stateVersion: stateVersion}, nil
}

func verifiedBoundDirectory(root string, artifact UpdateArtifact, stateVersion int) (UpdateArtifact, error) {
	got, err := verifyReleaseDirectory(root, artifact.ProductVersion, artifact.Platform, artifact.ArchiveSHA256, stateVersion)
	if err == nil && artifact.SourceKind == "archive" {
		got.SourceKind = "archive"
	}
	if err == nil && artifact.SourceKind == "directory" {
		got.SourceKind = "directory"
		got.Label = "hive-" + got.ProductVersion + "-" + strings.ReplaceAll(got.Platform, "/", "-")
	}
	return got, err
}

// VerifyPreparedPackage checks the preparation bytes against the original
// frozen binding; its directory may have been renamed during offline extraction.
func VerifyPreparedPackage(binding UpdatePackageBinding, stateVersion int) error {
	if err := binding.Validate(stateVersion); err != nil {
		return err
	}
	got, err := verifiedBoundDirectory(binding.PreparationDirectory, binding.Artifact, stateVersion)
	if err != nil {
		return err
	}
	if got != binding.Artifact {
		return fmt.Errorf("prepared package differs from frozen artifact")
	}
	return nil
}

// VerifyRetainedPackage authenticates pinned bytes without network access.
// It does not require the now-disposable preparation path.
func VerifyRetainedPackage(binding UpdatePackageBinding, stateVersion int) error {
	if err := binding.Validate(stateVersion); err != nil {
		return err
	}
	got, err := verifiedBoundDirectory(binding.PackageDirectory, binding.Artifact, stateVersion)
	if err != nil {
		return err
	}
	if got != binding.Artifact {
		return fmt.Errorf("retained package differs from frozen artifact")
	}
	return nil
}

// Retain materializes a verified immutable copy and never overwrites an existing
// package. A private store lock excludes other aware retainers during placement.
// Existing packages must have exactly the frozen identity before reuse.
func (p *PreparedRelease) Retain() error {
	b := p.Binding
	if err := b.Validate(p.stateVersion); err != nil {
		return err
	}
	got, err := verifiedBoundDirectory(b.PreparationDirectory, b.Artifact, p.stateVersion)
	if err != nil {
		return err
	}
	if got != b.Artifact {
		return fmt.Errorf("prepared package changed")
	}
	store := filepath.Dir(b.PackageDirectory)
	if err := checkCanonicalAncestors(store); err != nil {
		return err
	}
	if err := os.MkdirAll(store, 0700); err != nil {
		return err
	}
	lock := filepath.Join(store, ".retain-lock")
	lockFile, err := os.OpenFile(lock, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return fmt.Errorf("package store lock unavailable: %w", err)
	}
	defer lockFile.Close()
	lockInfo, err := lockFile.Stat()
	if err != nil {
		return err
	}
	if !lockInfo.Mode().IsRegular() || lockInfo.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("package store lock is not private regular file")
	}
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return fmt.Errorf("package store busy: %w", err)
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
	// Keep the lock inode: removing it could let a third process lock a different
	// inode while another waiter still holds this one. The OS unlocks on a crash.
	if _, err := os.Lstat(b.PackageDirectory); err == nil {
		return VerifyRetainedPackage(b, p.stateVersion)
	} else if !os.IsNotExist(err) {
		return err
	}
	stage, err := os.MkdirTemp(store, ".release-stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	root := filepath.Join(stage, b.Artifact.Label)
	err = filepath.WalkDir(b.PreparationDirectory, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(b.PreparationDirectory, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(root, relative)
		if d.IsDir() {
			return os.Mkdir(dest, 0700)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("prepared package contains non-regular entry")
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		source, err := os.Open(path)
		if err != nil {
			return err
		}
		defer source.Close()
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
		if err != nil {
			return err
		}
		copyErr := out.Chmod(info.Mode().Perm())
		if copyErr == nil {
			_, copyErr = io.Copy(out, source)
		}
		if copyErr == nil {
			copyErr = out.Sync()
		}
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if err != nil {
		return err
	}
	staged := b
	staged.PackageDirectory = root
	staged.TargetExecutable.Path = filepath.Join(root, "bin", "hive")
	if err := VerifyRetainedPackage(staged, p.stateVersion); err != nil {
		return err
	}
	if err := syncReleaseDirectories(root); err != nil {
		return err
	}
	// Recheck immediately before rename; a populated raced destination is never
	// replaced by rename, and aware retainers share the store lock above.
	if _, err := os.Lstat(b.PackageDirectory); err == nil {
		return VerifyRetainedPackage(b, p.stateVersion)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(root, b.PackageDirectory); err != nil {
		return err
	}
	if err := syncReleaseDirectory(store); err != nil {
		return err
	}
	return VerifyRetainedPackage(b, p.stateVersion)
}

func syncReleaseDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	syncErr := dir.Sync()
	closeErr := dir.Close()
	if syncErr != nil {
		return syncErr
	}
	return closeErr
}
func syncReleaseDirectories(root string) error {
	var dirs []string
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	}); err != nil {
		return err
	}
	for i := len(dirs) - 1; i >= 0; i-- {
		if err := syncReleaseDirectory(dirs[i]); err != nil {
			return err
		}
	}
	return nil
}
