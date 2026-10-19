package main

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"tricell-hive/tooling/viewer"
)

const viewerTrackingOutputLimit = 8 << 20

var errViewerTrackingIncomplete = errors.New("Git tracking output exceeded its limit")

type viewerTrackingGitRunner func(context.Context, string, ...string) ([]byte, error)

// trackingOutput refuses overflow even when the last retained byte is NUL.
// A complete-looking prefix must never prove that a path is untracked.
type trackingOutput struct {
	bytes.Buffer
	limit    int
	overflow bool
}

func (b *trackingOutput) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		b.overflow = true
		return 0, errViewerTrackingIncomplete
	}
	return b.Buffer.Write(p)
}

func runViewerTrackingGit(ctx context.Context, dir string, args ...string) ([]byte, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	run := func(dir string, env []string, args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, git, append([]string{"-C", dir}, args...)...)
		cmd.Env = env
		cmd.WaitDelay = gitWaitDelay
		out := &trackingOutput{limit: viewerTrackingOutputLimit}
		cmd.Stdout = out
		err := cmd.Run()
		if out.overflow {
			return nil, errViewerTrackingIncomplete
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if err != nil {
			return nil, err
		}
		return out.Bytes(), nil
	}
	// Ask Git for the repository-local variables, as the manager's shared
	// runner does, but keep this setup process under the same cancellation.
	listed, err := run(os.TempDir(), os.Environ(), "rev-parse", "--local-env-vars")
	if err != nil {
		return nil, err
	}
	drop := map[string]bool{}
	for _, key := range strings.Fields(string(listed)) {
		drop[key] = true
	}
	env := []string{}
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if !drop[key] && key != "GIT_OPTIONAL_LOCKS" {
			env = append(env, kv)
		}
	}
	env = append(env, "GIT_OPTIONAL_LOCKS=0")
	return run(dir, env, args...)
}

// viewerTrackingRoot discovers the nearest containing Git marker, including
// worktree .git files. Only reaching the filesystem root proves no repository;
// access failures and malformed markers remain query failures.
func viewerTrackingRoot(path string, cache map[string]trackingRoot) trackingRoot {
	dir := filepath.Dir(path)
	if r, ok := cache[dir]; ok {
		return r
	}
	var visited []string
	var result trackingRoot
	for {
		if r, ok := cache[dir]; ok {
			result = r
			break
		}
		visited = append(visited, dir)
		info, err := os.Lstat(filepath.Join(dir, ".git"))
		if err == nil {
			if info.IsDir() || info.Mode().IsRegular() {
				result.root = dir
			} else {
				result.failed = true
			}
			break
		}
		if !errors.Is(err, fs.ErrNotExist) {
			result.failed = true
			break
		}
		// Missing source directories can still have a tracked path in an ancestor.
		if parent := filepath.Dir(dir); parent != dir {
			dir = parent
		} else {
			break
		}
	}
	for _, path := range visited {
		cache[path] = result
	}
	return result
}

type trackingRoot struct {
	root   string
	failed bool
}

func viewSourceTracking(run viewerTrackingGitRunner) viewer.SourceTrackingReader {
	return func(ctx context.Context, paths []string) map[string]viewer.SourceClassification {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		out := make(map[string]viewer.SourceClassification, len(paths))
		groups := map[string][]string{}
		cache := map[string]trackingRoot{}
		for _, path := range paths {
			result := viewerTrackingRoot(path, cache)
			switch {
			case result.failed || ctx.Err() != nil:
				out[path] = viewer.SourceClassification{Status: "unknown", Reason: "git_query_failed"}
			case result.root == "":
				out[path] = viewer.SourceClassification{Status: "local"}
			default:
				groups[result.root] = append(groups[result.root], path)
			}
		}
		roots := make([]string, 0, len(groups))
		for root := range groups {
			roots = append(roots, root)
		}
		sort.Strings(roots)
		for _, root := range roots {
			prefixes := map[string]bool{}
			for _, path := range groups[root] {
				prefix := filepath.Dir(filepath.Dir(path))
				rel, err := filepath.Rel(root, prefix)
				if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
					rel = "."
				}
				if rel == "." {
					prefixes[":(top,literal)"] = true
				} else {
					prefixes[":(top,literal)"+filepath.ToSlash(rel)+"/"] = true
				}
			}
			specs := make([]string, 0, len(prefixes))
			for prefix := range prefixes {
				specs = append(specs, prefix)
			}
			sort.Strings(specs)
			args := append([]string{"ls-files", "--cached", "--full-name", "-z", "--"}, specs...)
			raw, err := run(ctx, root, args...)
			reason := "git_query_failed"
			if errors.Is(err, errViewerTrackingIncomplete) || (len(raw) > 0 && raw[len(raw)-1] != 0) {
				reason = "git_output_incomplete"
				err = errViewerTrackingIncomplete
			}
			tracked := map[string]bool{}
			if err == nil {
				for _, rel := range bytes.Split(raw, []byte{0}) {
					if len(rel) > 0 {
						tracked[filepath.Join(root, filepath.FromSlash(string(rel)))] = true
					}
				}
			}
			for _, path := range groups[root] {
				c := viewer.SourceClassification{Status: "local"}
				if err != nil {
					c = viewer.SourceClassification{Status: "unknown", Reason: reason}
				} else if tracked[path] {
					c.Status = "tracked"
				}
				out[path] = c
			}
		}
		return out
	}
}
