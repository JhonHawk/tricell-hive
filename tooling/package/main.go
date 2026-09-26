// Command package builds complete, offline installer archives. It never publishes.
package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"tricell-hive/tooling/distribution"
)

var platforms = []string{"darwin/arm64", "darwin/amd64", "linux/arm64", "linux/amd64"}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("package", flag.ContinueOnError)
	source := flags.String("source", ".", "source checkout")
	output := flags.String("out", "", "new artifacts directory (required)")
	target := flags.String("platforms", strings.Join(platforms, ","), "target OS/architecture pairs")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *output == "" {
		return fmt.Errorf("use --out DIRECTORY; archives are local and never published")
	}
	src, err := filepath.Abs(*source)
	if err != nil {
		return err
	}
	out, err := filepath.Abs(*output)
	if err != nil {
		return err
	}
	targets := strings.Split(*target, ",")
	seen := map[string]bool{}
	for _, p := range targets {
		supported := false
		for _, known := range platforms {
			if p == known {
				supported = true
			}
		}
		if !supported || seen[p] {
			return fmt.Errorf("unsupported or repeated platform: %s", p)
		}
		seen[p] = true
	}
	if err := os.MkdirAll(out, 0700); err != nil {
		return err
	}
	scratch, err := os.MkdirTemp(out, ".build-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch) // This invocation owns this entire temporary tree.
	frozen := filepath.Join(scratch, "source")
	inputs := []string{"go.mod", "install.sh", "content", "integrations", "tooling/cli", "tooling/management", "tooling/legacy", "tooling/distribution"}
	for _, name := range inputs {
		if err := copyTree(filepath.Join(src, name), filepath.Join(frozen, name)); err != nil {
			return err
		}
	}
	inventory, err := distribution.Files(frozen)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(inventory))
	for name := range inventory {
		names = append(names, name)
	}
	sort.Strings(names)
	var identity strings.Builder
	for _, name := range names {
		info, err := os.Stat(filepath.Join(frozen, name))
		if err != nil {
			return err
		}
		fmt.Fprintf(&identity, "%s\x00%o\x00%s\n", name, info.Mode().Perm(), inventory[name])
	}
	sourceID := distribution.Digest([]byte(identity.String()))
	for _, p := range targets {
		pair := strings.Split(p, "/")
		label := "hive-" + sourceID[:12] + "-" + pair[0] + "-" + pair[1]
		dir := filepath.Join(scratch, label)
		for _, name := range []string{"content", "integrations/agent-profiles.json", "install.sh"} {
			if err := copyTree(filepath.Join(frozen, name), filepath.Join(dir, name)); err != nil {
				return err
			}
		}
		if err := os.MkdirAll(filepath.Join(dir, "bin"), 0755); err != nil {
			return err
		}
		binary := filepath.Join(dir, "bin", "hive")
		command := exec.Command("go", "build", "-trimpath", "-buildvcs=false", "-o", binary, "./tooling/cli")
		command.Dir = frozen
		command.Env = buildEnv(os.Environ(), pair[0], pair[1])
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			return fmt.Errorf("build %s: %w", p, err)
		}
		b, err := os.ReadFile(binary)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "bin/hive.sha256"), []byte(distribution.Digest(b)+"\n"), 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "platform"), []byte(p+"\n"), 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "README-install.txt"), []byte("Hive — offline installation without Go\n\nClose your CLI sessions and run ./install.sh in a terminal.\nUse ./install.sh --dry-run to inspect changes. The installer asks for confirmation.\nOpen new CLI sessions after installation. Keep this package for recovery.\nTo update, download a new complete package and run the same command.\nGitHub automatic Source code archives do not include binaries.\n"), 0644); err != nil {
			return err
		}
		files, err := distribution.Files(dir)
		if err != nil {
			return err
		}
		manifest := distribution.Manifest{Version: 1, Platform: p, SourceID: sourceID, Files: files}
		metadata, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, distribution.ManifestName), append(metadata, '\n'), 0644); err != nil {
			return err
		}
		archive := filepath.Join(out, label+".tar.gz")
		if err := writeArchive(dir, archive); err != nil {
			return err
		}
		compressed, err := os.ReadFile(archive)
		if err != nil {
			return err
		}
		checkfile, err := os.OpenFile(archive+".sha256", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		_, writeErr := fmt.Fprintf(checkfile, "%s  %s\n", distribution.Digest(compressed), filepath.Base(archive))
		closeErr := checkfile.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		fmt.Println(archive)
	}
	return nil
}

func buildEnv(env []string, osName, arch string) []string {
	var result []string
	overrides := map[string]string{"GOOS": osName, "GOARCH": arch, "CGO_ENABLED": "0", "GOTOOLCHAIN": "local"}
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if _, ok := overrides[key]; !ok {
			result = append(result, entry)
		}
	}
	for _, key := range []string{"GOOS", "GOARCH", "CGO_ENABLED", "GOTOOLCHAIN"} {
		result = append(result, key+"="+overrides[key])
	}
	return result
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("linked source rejected: %s", path)
		}
		if d.IsDir() {
			switch d.Name() {
			case "__pycache__", "node_modules", "dist", ".astro":
				return filepath.SkipDir
			}
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("special source rejected: %s", path)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, bytes, info.Mode().Perm())
	})
}

func writeArchive(root, dest string) (err error) {
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.Remove(dest)
		}
	}()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !d.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("archive contains nonregular file")
		}
		rel, err := filepath.Rel(filepath.Dir(root), path)
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		header.ModTime = time.Unix(0, 0)
		header.AccessTime = time.Time{}
		header.ChangeTime = time.Time{}
		header.Uid = 0
		header.Gid = 0
		header.Uname = ""
		header.Gname = ""
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(tw, file)
		return err
	})
	tarErr := tw.Close()
	gzipErr := gz.Close()
	fileErr := f.Close()
	for _, e := range []error{walkErr, tarErr, gzipErr, fileErr} {
		if e != nil {
			return e
		}
	}
	return nil
}
