package distribution

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func releaseArchiveFixture(t *testing.T) (ReleaseTarget, []byte) {
	t.Helper()
	root := packageFixture(t)
	m, err := ReadManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	m.ProductVersion = "1.2.3"
	m.UpdateProtocol = 1
	b := UpdateBootstrap{Protocol: 1, ProductVersion: m.ProductVersion, Platform: m.Platform, SourceID: m.SourceID, ExecutableSHA256: m.Files["bin/hive"], ManifestVersion: m.Version, MinStateVersion: 1, MaxStateVersion: 7}
	data, _ := json.Marshal(b)
	if err := os.WriteFile(filepath.Join(root, UpdateBootstrapName), data, 0600); err != nil {
		t.Fatal(err)
	}
	m.Files[UpdateBootstrapName] = Digest(data)
	data, _ = json.Marshal(m)
	if err := os.WriteFile(filepath.Join(root, ManifestName), data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(root, "bin/hive"), 0700); err != nil {
		t.Fatal(err)
	}
	target, err := (ReleaseClient{}).Resolve(context.Background(), "1.0.0", "1.2.3", runtime.GOOS+"/"+runtime.GOARCH)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	gz := gzip.NewWriter(&out)
	tw := tar.NewWriter(gz)
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, _ := filepath.Rel(root, p)
		name := target.Label
		if rel != "." {
			name += "/" + filepath.ToSlash(rel)
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		h, e := tar.FileInfoHeader(info, "")
		if e != nil {
			return e
		}
		h.Name = name
		if e = tw.WriteHeader(h); e != nil {
			return e
		}
		if !d.IsDir() {
			data, e := os.ReadFile(p)
			if e != nil {
				return e
			}
			_, e = tw.Write(data)
			return e
		}
		return nil
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
	return target, out.Bytes()
}
func releaseArchiveClient(target ReleaseTarget, archive []byte, checksum string) ReleaseClient {
	return ReleaseClient{Transport: releaseTransport(func(r *http.Request) (*http.Response, error) {
		switch r.URL.String() {
		case target.ArchiveURL:
			return releaseResponse(r, 200, string(archive)), nil
		case target.ArchiveURL + ".sha256":
			return releaseResponse(r, 200, checksum), nil
		default:
			return nil, fmt.Errorf("unexpected request")
		}
	})}
}
func canonicalTemp(t *testing.T) string {
	t.Helper()
	p, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestReleasePreparationPreviewAndOfflineRetention(t *testing.T) {
	target, archive := releaseArchiveFixture(t)
	client := releaseArchiveClient(target, archive, Digest(archive)+"  "+target.Label+".tar.gz\n")
	store := filepath.Join(canonicalTemp(t), "packages")
	prepared, err := client.Prepare(context.Background(), target, store, 7)
	if err != nil {
		t.Fatal(err)
	}
	defer prepared.Cleanup()
	if _, err := os.Stat(store); !os.IsNotExist(err) {
		t.Fatalf("preview wrote store: %v", err)
	}
	if prepared.Binding.PreparationExecutor.Path == prepared.Binding.TargetExecutable.Path {
		t.Fatal("temporary executor became hook target")
	}
	if err := prepared.Binding.Validate(7); err != nil {
		t.Fatal(err)
	}
	if err := VerifyRetainedPackage(prepared.Binding, 7); err == nil {
		t.Fatal("accepted cache miss")
	}
	if err := prepared.Retain(); err != nil {
		t.Fatal(err)
	}
	binding := prepared.Binding
	temporary := binding.PreparationDirectory
	if err := prepared.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(temporary); !os.IsNotExist(err) {
		t.Fatalf("preview not cleaned: %v", err)
	}
	if err := VerifyRetainedPackage(binding, 7); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binding.TargetExecutable.Path, []byte("tampered"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := VerifyRetainedPackage(binding, 7); err == nil {
		t.Fatal("accepted tampered retained executor")
	}
	other, err := client.Prepare(context.Background(), target, store, 7)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Cleanup()
	if err := other.Retain(); err == nil {
		t.Fatal("replaced changed retained package")
	}
}
func TestReleasePreparationRejectsRemoteCorruption(t *testing.T) {
	target, archive := releaseArchiveFixture(t)
	for _, tc := range []struct {
		name     string
		archive  []byte
		checksum string
	}{
		{"checksum mismatch", archive, strings.Repeat("a", 64) + "  " + target.Label + ".tar.gz\n"},
		{"wrong name", archive, Digest(archive) + "  other.tar.gz\n"},
		{"truncated gzip", archive[:len(archive)-4], Digest(archive[:len(archive)-4]) + "  " + target.Label + ".tar.gz\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := filepath.Join(canonicalTemp(t), "packages")
			p, err := releaseArchiveClient(target, tc.archive, tc.checksum).Prepare(context.Background(), target, store, 7)
			if p != nil {
				defer p.Cleanup()
			}
			if err == nil {
				t.Fatal("accepted corrupt package")
			}
			if _, err := os.Stat(store); !os.IsNotExist(err) {
				t.Fatalf("failure wrote store: %v", err)
			}
		})
	}
}

func replaceReleaseArchive(t *testing.T, original []byte, edit func(*tar.Header, []byte) (*tar.Header, []byte)) []byte {
	t.Helper()
	in, e := gzip.NewReader(bytes.NewReader(original))
	if e != nil {
		t.Fatal(e)
	}
	defer in.Close()
	reader := tar.NewReader(in)
	var buf bytes.Buffer
	out := gzip.NewWriter(&buf)
	writer := tar.NewWriter(out)
	for {
		h, e := reader.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		data, e := io.ReadAll(reader)
		if e != nil {
			t.Fatal(e)
		}
		h, data = edit(h, data)
		if e := writer.WriteHeader(h); e != nil {
			t.Fatal(e)
		}
		if h.Typeflag == tar.TypeReg || h.Typeflag == tar.TypeRegA {
			if _, e := writer.Write(data); e != nil {
				t.Fatal(e)
			}
		}
	}
	if e := writer.Close(); e != nil {
		t.Fatal(e)
	}
	if e := out.Close(); e != nil {
		t.Fatal(e)
	}
	return buf.Bytes()
}
func TestReleasePreparationRejectsUnsafePackages(t *testing.T) {
	target, archive := releaseArchiveFixture(t)
	for _, kind := range []string{"symlink", "hardlink", "traversal", "duplicate", "unexpected label", "non executable", "protocol", "schema", "platform", "extra file"} {
		t.Run(kind, func(t *testing.T) {
			changed := replaceReleaseArchive(t, archive, func(h *tar.Header, data []byte) (*tar.Header, []byte) {
				if strings.HasSuffix(h.Name, "/bin/hive") {
					switch kind {
					case "symlink":
						h.Typeflag = tar.TypeSymlink
						h.Linkname = "elsewhere"
						h.Size = 0
						data = nil
					case "hardlink":
						h.Typeflag = tar.TypeLink
						h.Linkname = target.Label + "/LICENSE"
						h.Size = 0
						data = nil
					case "traversal":
						h.Name = target.Label + "/../../escaped"
					case "duplicate":
						h.Name = target.Label + "/LICENSE"
					case "non executable":
						h.Mode = 0600
					case "extra file":
						h.Name = target.Label + "/unlisted"
					}
				}
				if kind == "unexpected label" {
					h.Name = strings.Replace(h.Name, target.Label, "other", 1)
				}
				if strings.HasSuffix(h.Name, "/"+ManifestName) {
					var m Manifest
					if e := json.Unmarshal(data, &m); e != nil {
						t.Fatal(e)
					}
					switch kind {
					case "protocol":
						m.UpdateProtocol = 99
					case "platform":
						m.Platform = "other/cpu"
					}
					data, _ = json.Marshal(m)
					h.Size = int64(len(data))
				}
				return h, data
			})
			store := filepath.Join(canonicalTemp(t), "store")
			state := 7
			if kind == "schema" {
				state = 8
			}
			p, err := releaseArchiveClient(target, changed, Digest(changed)+"  "+target.Label+".tar.gz\n").Prepare(context.Background(), target, store, state)
			if p != nil {
				defer p.Cleanup()
			}
			if err == nil {
				t.Fatal("accepted unsafe package")
			}
			if _, e := os.Stat(store); !os.IsNotExist(e) {
				t.Fatalf("wrote store: %v", e)
			}
		})
	}
}

func TestReleaseRetainReusesExactPackageAndRefusesSymlinkPaths(t *testing.T) {
	target, archive := releaseArchiveFixture(t)
	client := releaseArchiveClient(target, archive, Digest(archive)+"  "+target.Label+".tar.gz\n")
	base := canonicalTemp(t)
	store := filepath.Join(base, "packages")
	p, err := client.Prepare(context.Background(), target, store, 7)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Cleanup()
	if err := p.Retain(); err != nil {
		t.Fatal(err)
	}
	if err := p.Retain(); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "linked")
	if err := os.Symlink(store, link); err != nil {
		t.Fatal(err)
	}
	if p, err := client.Prepare(context.Background(), target, link, 7); err == nil {
		defer p.Cleanup()
		t.Fatal("accepted symlink store")
	}
}

func TestReleasePreparationArchiveLimits(t *testing.T) {
	target, _ := releaseArchiveFixture(t)
	for _, kind := range []string{"extracted bytes", "entries"} {
		t.Run(kind, func(t *testing.T) {
			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)
			tw := tar.NewWriter(gz)
			if err := tw.WriteHeader(&tar.Header{Name: target.Label, Mode: 0700, Typeflag: tar.TypeDir}); err != nil {
				t.Fatal(err)
			}
			if kind == "extracted bytes" {
				if err := tw.WriteHeader(&tar.Header{Name: target.Label + "/huge", Mode: 0600, Typeflag: tar.TypeReg, Size: MaxExtractedBytes + 1}); err != nil {
					t.Fatal(err)
				}
			} else {
				for i := 0; i < MaxArchiveEntries; i++ {
					if err := tw.WriteHeader(&tar.Header{Name: fmt.Sprintf("%s/d%d", target.Label, i), Mode: 0700, Typeflag: tar.TypeDir}); err != nil {
						t.Fatal(err)
					}
				}
				if err := tw.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if err := gz.Close(); err != nil {
				t.Fatal(err)
			}
			archive := buf.Bytes()
			store := filepath.Join(canonicalTemp(t), "store")
			_, err := releaseArchiveClient(target, archive, Digest(archive)+"  "+target.Label+".tar.gz\n").Prepare(context.Background(), target, store, 7)
			if err == nil {
				t.Fatal("accepted archive exceeding limits")
			}
			want := "extracted limit"
			if kind == "entries" {
				want = "too many entries"
			}
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("wrong boundary rejected: %v", err)
			}
		})
	}
}

func TestReleasePreparationCancellationAndPrivateCleanup(t *testing.T) {
	target, archive := releaseArchiveFixture(t)
	checksum := Digest(archive) + "  " + target.Label + ".tar.gz\n"
	temporaryHome := canonicalTemp(t)
	t.Setenv("TMPDIR", temporaryHome)
	store := filepath.Join(canonicalTemp(t), "store")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := releaseArchiveClient(target, archive, checksum)
	transport := c.Transport
	c.Transport = releaseTransport(func(r *http.Request) (*http.Response, error) {
		response, err := transport.RoundTrip(r)
		if r.URL.String() == target.ArchiveURL {
			cancel()
		}
		return response, err
	})
	if p, err := c.Prepare(ctx, target, store, 7); err == nil {
		defer p.Cleanup()
		t.Fatal("canceled preparation returned success")
	}
	entries, err := os.ReadDir(temporaryHome)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatal("canceled preparation left disposable inputs")
	}
	valid := releaseArchiveClient(target, archive, checksum)
	p, err := valid.Prepare(context.Background(), target, store, 7)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p.privateDirectory)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Fatalf("temporary permissions = %v", info.Mode())
	}
	// Even package directories that become unreadable remain disposable.
	if err := os.Chmod(filepath.Join(p.Binding.PreparationDirectory, "content"), 0000); err != nil {
		t.Fatal(err)
	}
	if err := p.Cleanup(); err != nil {
		t.Fatal(err)
	}
	entries, err = os.ReadDir(temporaryHome)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatal("cleanup left disposable inputs")
	}
}

func TestReleaseRetainedVerificationBoundsChangedFiles(t *testing.T) {
	target, archive := releaseArchiveFixture(t)
	client := releaseArchiveClient(target, archive, Digest(archive)+"  "+target.Label+".tar.gz\n")
	p, err := client.Prepare(context.Background(), target, filepath.Join(canonicalTemp(t), "packages"), 7)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Cleanup()
	if err := p.Retain(); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(p.Binding.TargetExecutable.Path, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = file.Truncate(MaxExtractedBytes + 1)
	closeErr := file.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	err = VerifyRetainedPackage(p.Binding, 7)
	if err == nil || !strings.Contains(err.Error(), "extracted limit") {
		t.Fatalf("changed oversized file was not bounded before hashing: %v", err)
	}
}

func TestLocalReleaseImportRetainsWithoutOwningSource(t *testing.T) {
	target, archive := releaseArchiveFixture(t)
	parent, _ := filepath.EvalSymlinks(t.TempDir())
	source, err := Extract(archive, parent)
	if err != nil {
		t.Fatal(err)
	}
	store := filepath.Join(parent, "store")
	p, err := PrepareLocalRelease(source, store, 7)
	if err != nil {
		t.Fatal(err)
	}
	if p.Binding.Artifact.SourceKind != "directory" || p.Binding.Artifact.ArchiveSHA256 != "" {
		t.Fatal("local directory claimed archive identity")
	}
	if p.Binding.PackageDirectory != filepath.Join(store, target.Label) {
		t.Fatal(p.Binding.PackageDirectory)
	}
	if _, err := os.Stat(store); !os.IsNotExist(err) {
		t.Fatal("preparation wrote store")
	}
	if err := p.Retain(); err != nil {
		t.Fatal(err)
	}
	if err := p.Retain(); err != nil {
		t.Fatal(err)
	}
	if err := p.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if err := VerifyIfPackaged(source); err != nil {
		t.Fatalf("cleanup removed source: %v", err)
	}
	if err := VerifyRetainedPackage(p.Binding, 7); err != nil {
		t.Fatal(err)
	}
	// Directory imports cannot weaken archive bindings or invent source kinds.
	for _, kind := range []string{"", "archive", "other"} {
		b := p.Binding
		b.Artifact.SourceKind = kind
		if err := b.Validate(7); err == nil {
			t.Fatalf("accepted missing archive digest for %q", kind)
		}
	}
	if err := os.WriteFile(filepath.Join(source, "LICENSE"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := p.Retain(); err == nil {
		t.Fatal("accepted changed source")
	}
}

func TestLocalReleaseRetentionPreservesPayloadPermissions(t *testing.T) {
	_, archive := releaseArchiveFixture(t)
	parent, _ := filepath.EvalSymlinks(t.TempDir())
	source, err := Extract(archive, parent)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(source, "LICENSE"), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := PrepareLocalRelease(source, filepath.Join(parent, "store"), 7)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Retain(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"LICENSE", "bin/hive"} {
		before, _ := os.Stat(filepath.Join(source, name))
		after, _ := os.Stat(filepath.Join(p.Binding.PackageDirectory, name))
		if before.Mode().Perm() != after.Mode().Perm() {
			t.Fatalf("%s changed mode: %v -> %v", name, before.Mode().Perm(), after.Mode().Perm())
		}
	}
}
