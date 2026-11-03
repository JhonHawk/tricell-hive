package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"syscall"
	"tricell-hive/tooling/viewer"
	"unicode/utf8"
)

const workDocumentLimit = 8
const workDocumentByteLimit int64 = 256 * 1024
const workScanEntryLimit = 4096
const workScanDirectoryLimit = 256
const workScanDepthLimit = 32

type workDocument struct {
	Path   string
	Label  string
	SHA256 string
	Size   int64
	Bytes  []byte
}

type workScanNode struct {
	path string
	info os.FileInfo
	file *os.File
}

type workDocumentScan struct {
	root        *os.Root
	nodes       []workScanNode
	entries     int
	directories int
}

// collectWorkDocuments returns only a completely validated, frozen set. The root
// is canonicalized once; every subsequent lookup is anchored to its descriptor.
func collectWorkDocuments(root string, stage string, documents []string, explicit bool) ([]workDocument, error) {
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, workScanError("", "cannot resolve source root")
	}
	anchor, err := os.OpenRoot(canonical)
	if err != nil {
		return nil, workScanError("", "cannot open source root")
	}
	defer anchor.Close()
	scan := workDocumentScan{root: anchor}
	defer func() {
		for _, node := range scan.nodes {
			if node.file != nil {
				node.file.Close()
			}
		}
	}()
	if _, err := scan.rememberDirectory("."); err != nil {
		return nil, err
	}
	paths := []string{}
	if !explicit {
		if stage != "plan" {
			return nil, workScanError("", "non-plan publication requires explicit Markdown selection")
		}
		paths = append(paths, "proposal.md")
		if _, err := anchor.Lstat("design.md"); err == nil {
			paths = append(paths, "design.md")
		} else if !os.IsNotExist(err) {
			return nil, workScanError("design.md", "cannot inspect optional document")
		}
		paths = append(paths, "tasks.md")
		specs, err := scan.discoverSpecs()
		if err != nil {
			return nil, err
		}
		paths = append(paths, specs...)
	}
	paths = append(paths, documents...)
	if len(paths) == 0 || len(paths) > workDocumentLimit {
		return nil, workScanError("", fmt.Sprintf("selected %d documents; select 1 to %d explicitly, without silently omitting plan documents", len(paths), workDocumentLimit))
	}
	seen := map[string]bool{}
	labels := map[string]bool{}
	result := make([]workDocument, 0, len(paths))
	for _, input := range paths {
		path, err := workScanRelativePath(input)
		if err != nil {
			return nil, err
		}
		if seen[path] {
			return nil, workScanError(path, "duplicate selected path")
		}
		seen[path] = true
		label := path
		if !explicit && stage == "plan" {
			switch path {
			case "proposal.md":
				label = "Proposal"
			case "design.md":
				label = "Design"
			case "tasks.md":
				label = "Tasks"
			}
		}
		// Reuse the existing v3 tab validator before a create/upload can occur.
		labelRequest := viewer.TimelineDocumentSnapshotRequest{SchemaVersion: 3, OperationID: "preflight", Actor: viewer.TimelineActor{Kind: "native_producer", ProducerID: "preflight"}, ExpectedVersion: 1, ExpectedSessionVersion: 1, DocumentKey: "preflight", Status: "draft", Tabs: []viewer.TimelineDocumentTab{{Label: label, DocumentID: "preflight"}}}
		if labelRequest.Validate() != nil {
			return nil, workScanError(path, "document label violates v3 text rules (maximum 100 characters); select a shorter relative path")
		}
		if labels[label] {
			return nil, workScanError(path, "duplicate document label")
		}
		labels[label] = true
		data, err := scan.read(path)
		if err != nil {
			return nil, err
		}
		result = append(result, workDocument{Path: path, Label: label, SHA256: fmt.Sprintf("%x", sha256.Sum256(data)), Size: int64(len(data)), Bytes: data})
	}
	if err := scan.validate(); err != nil {
		return nil, err
	}
	// Recheck optional presence as well as selected descriptors. Directory identity
	// and metadata checks cover additions/removals in the discovered spec tree.
	if !explicit {
		_, err := anchor.Lstat("design.md")
		present := err == nil
		if present != seen["design.md"] || (err != nil && !os.IsNotExist(err)) {
			return nil, workScanError("design.md", "source selection changed during collection")
		}
	}
	return result, nil
}

func workScanError(path, cause string) error {
	if path != "" {
		return fmt.Errorf("document preflight %q: %s; preserve source files and retry with a stable, bounded Markdown selection", path, cause)
	}
	return fmt.Errorf("document preflight: %s; preserve source files and retry with a stable, bounded Markdown selection", cause)
}

func workScanRelativePath(input string) (string, error) {
	if input == "" || filepath.IsAbs(input) || strings.ContainsAny(input, "\\\x00") {
		return "", workScanError("", "document path must be source-relative")
	}
	for _, part := range strings.Split(filepath.ToSlash(input), "/") {
		if part == ".." {
			return "", workScanError("", "document path escapes source root")
		}
	}
	path := filepath.Clean(input)
	if path == "." || strings.ToLower(filepath.Ext(path)) != ".md" {
		return "", workScanError("", "selected document must be Markdown")
	}
	return filepath.ToSlash(path), nil
}

func (scan *workDocumentScan) rememberDirectory(path string) (os.FileInfo, error) {
	if scan.directories >= workScanDirectoryLimit {
		return nil, workScanError(path, "directory traversal budget exceeded")
	}
	info, err := scan.root.Lstat(path)
	if err != nil || !info.IsDir() {
		return nil, workScanError(path, "directory must exist without symlink components")
	}
	scan.nodes = append(scan.nodes, workScanNode{path: path, info: info})
	scan.directories++
	return info, nil
}

func (scan *workDocumentScan) discoverSpecs() ([]string, error) {
	_, err := scan.root.Lstat("specs")
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, workScanError("specs", "cannot inspect specification directory")
	}
	paths := []string{}
	var walk func(string, int) error
	walk = func(path string, depth int) error {
		if depth > workScanDepthLimit {
			return workScanError(path, "directory depth budget exceeded")
		}
		before, err := scan.rememberDirectory(path)
		if err != nil {
			return err
		}
		dir, err := scan.root.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
		if err != nil {
			return workScanError(path, "cannot enumerate specification directory")
		}
		defer dir.Close()
		actual, err := dir.Stat()
		if err != nil || !workScanSameMetadata(before, actual) {
			return workScanError(path, "directory changed while opening")
		}
		entries, err := dir.ReadDir(workScanEntryLimit - scan.entries + 1)
		if err != nil && err != io.EOF {
			return workScanError(path, "cannot enumerate specification directory")
		}
		scan.entries += len(entries)
		if scan.entries > workScanEntryLimit {
			return workScanError(path, "directory entry budget exceeded")
		}
		for _, entry := range entries {
			child := filepath.ToSlash(filepath.Join(path, entry.Name()))
			if entry.Type()&os.ModeSymlink != 0 {
				return workScanError(child, "symlinks in specification selection are forbidden")
			}
			if entry.IsDir() {
				if err := walk(child, depth+1); err != nil {
					return err
				}
			} else if entry.Name() == "spec.md" {
				paths = append(paths, child)
			}
		}
		return nil
	}
	if err := walk("specs", 1); err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

func (scan *workDocumentScan) read(path string) ([]byte, error) {
	// Lstat each parent, including for explicit selections. Root.Open constrains a
	// concurrent path replacement to the canonical root, and final checks reject it.
	parent := filepath.Dir(path)
	for parent != "." {
		if _, err := scan.rememberDirectory(parent); err != nil {
			return nil, err
		}
		parent = filepath.Dir(parent)
	}
	before, err := scan.root.Lstat(path)
	if err != nil || !workScanRegular(before) || before.Size() > workDocumentByteLimit {
		return nil, workScanError(path, "document must be a single-link regular file of at most 256 KiB")
	}
	// Nonblocking open prevents a regular-file-to-FIFO race from hanging preflight.
	file, err := scan.root.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, workScanError(path, "cannot open selected document")
	}
	scan.nodes = append(scan.nodes, workScanNode{path: path, info: before, file: file})
	actual, err := file.Stat()
	if err != nil || !workScanSameMetadata(before, actual) || !workScanRegular(actual) {
		return nil, workScanError(path, "document identity changed while opening")
	}
	for _, node := range scan.nodes {
		if node.path != path && node.file != nil && os.SameFile(node.info, actual) {
			return nil, workScanError(path, "selected documents share an inode")
		}
	}
	data, err := io.ReadAll(io.LimitReader(file, workDocumentByteLimit+1))
	if err != nil || int64(len(data)) != before.Size() || int64(len(data)) > workDocumentByteLimit {
		return nil, workScanError(path, "document size changed or bounded read failed")
	}
	if !utf8.Valid(data) {
		return nil, workScanError(path, "document is not valid UTF-8")
	}
	return data, nil
}

func workScanRegular(info os.FileInfo) bool {
	if info == nil || !info.Mode().IsRegular() {
		return false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat != nil && stat.Nlink == 1
}

func (scan *workDocumentScan) validate() error {
	for _, node := range scan.nodes {
		current, err := scan.root.Lstat(node.path)
		if err != nil || !workScanSameMetadata(node.info, current) {
			return workScanError(node.path, "source changed during collection")
		}
		if node.file != nil {
			actual, err := node.file.Stat()
			if err != nil || !workScanRegular(actual) || !workScanSameMetadata(node.info, actual) {
				return workScanError(node.path, "open document changed during collection")
			}
		}
	}
	return nil
}

func workScanSameMetadata(a, b os.FileInfo) bool {
	if a == nil || b == nil || !os.SameFile(a, b) || a.Size() != b.Size() || a.Mode() != b.Mode() || !a.ModTime().Equal(b.ModTime()) {
		return false
	}
	// Darwin and Linux expose ctime under different Stat_t field names. Checking
	// it catches in-place rewrites even when a writer restores the mtime; atime is
	// intentionally excluded because the scanner's own reads may update it.
	av, bv := reflect.ValueOf(a.Sys()), reflect.ValueOf(b.Sys())
	if av.Kind() != reflect.Pointer || bv.Kind() != reflect.Pointer || av.IsNil() || bv.IsNil() {
		return false
	}
	av, bv = av.Elem(), bv.Elem()
	for _, name := range []string{"Ctimespec", "Ctim"} {
		x, y := av.FieldByName(name), bv.FieldByName(name)
		if x.IsValid() && y.IsValid() {
			return reflect.DeepEqual(x.Interface(), y.Interface())
		}
	}
	return false
}
