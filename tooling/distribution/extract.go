package distribution

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	MaxPackageBytes   int64 = 128 << 20
	MaxExtractedBytes int64 = 256 << 20
	MaxArchiveEntries       = 10000
)

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
