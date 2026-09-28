// hive update refreshes an installed Hive from a Git commit of this checkout
// (default HEAD), without requiring an operator to run `git archive`, `plan
// install` and `apply` by hand. Its contract is documented in
// _support/docs/architecture/deployment-manager.md ("Update from a commit
// and list releases").
package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
)

type updateFlags struct {
	Rev, Source, Home, StateDir, Out string
	DryRun                           bool
}

func parseUpdateFlags(args []string) (updateFlags, error) {
	f := updateFlags{Rev: "HEAD", Source: "."}
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.StringVar(&f.Rev, "rev", f.Rev, "commit-ish to update from")
	fs.StringVar(&f.Source, "source", f.Source, "Git checkout to update from")
	fs.StringVar(&f.Home, "home", "", "explicit synthetic home; ignores host environment paths")
	fs.StringVar(&f.StateDir, "state-dir", "", "state directory (default: user Application Support/tricell-hive)")
	fs.BoolVar(&f.DryRun, "dry-run", false, "preview only; do not change anything")
	fs.StringVar(&f.Out, "out", "", "save the plan to FILE instead of applying")
	if err := fs.Parse(args); err != nil {
		return f, err
	}
	if fs.NArg() != 0 {
		return f, fmt.Errorf("unexpected positional arguments")
	}
	return f, nil
}

// update implements `hive update`. It is testable with injected stdin,
// stdout and an interactive flag, following how install (install.go) is
// structured. Its own core is updateWith, which takes a prompter instead of
// building its own installTerminal, so a future TUI can drive the same flow
// with its own huh-based prompter.
func update(args []string, in io.Reader, out io.Writer, interactive bool) error {
	f, err := parseUpdateFlags(args)
	if err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	terminal := installTerminal{reader: bufio.NewReader(in), out: out, interactive: interactive}
	// mentionDryRunFlag is true: hive update has a real --dry-run flag of its
	// own to suggest, unchanged from before this parameter existed.
	return updateWith(f, out, interactive, terminal, true)
}

// updateWith is update's core, separated from parsing and from the terminal
// implementation of prompter (design.md "Separar las preguntas de la
// lógica"). mentionDryRunFlag is threaded through to showInstallSummary
// (T4 fix round item 4): tui_screens.go's updateScreen is the only other
// caller, and passes false, since the interface has no --dry-run flag of
// its own to suggest — the same reasoning install.go's own
// runInstallFlowWith already applies for Install CLIs.
func updateWith(f updateFlags, out io.Writer, interactive bool, terminal prompter, mentionDryRunFlag bool) error {
	// 1. Choose the CLIs before running Git, so a host-less home fails
	// without extracting anything.
	o := management.Options{Scope: "user", Home: f.Home, StateDir: f.StateDir}
	hosts, err := management.RegisteredHosts(o)
	if err != nil {
		return err
	}
	if len(hosts) == 0 {
		return fmt.Errorf("no CLI hosts are registered with Hive for this home; run hive install first")
	}
	o.Hosts = hosts

	// 2. Validate the revision before touching Git: empty or dash-led values
	// could otherwise be read as an option by a later Git invocation.
	if f.Rev == "" || strings.HasPrefix(f.Rev, "-") {
		return fmt.Errorf("invalid --rev %q", f.Rev)
	}

	// 3. Run Git in a controlled way: resolved from PATH, no shell, and with
	// the ownership-affecting GIT_* variables stripped from its environment.
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git not found in PATH; hive update needs Git and a checkout; the offline package installs with install.sh")
	}
	env, err := filteredGitEnv(gitPath)
	if err != nil {
		return err
	}

	commit, err := resolveCommit(gitPath, env, f.Source, f.Rev)
	if err != nil {
		return err
	}

	// 4–5. Extract the commit and plan from it. The plan freezes the source
	// bytes, so the extraction is already gone before the summary and the
	// confirmation prompt.
	p, err := planFromCommit(gitPath, env, f.Source, commit, o)
	if err != nil {
		return err
	}

	unchanged, err := management.PlanUnchanged(p)
	if err != nil {
		return err
	}

	// 6. Summarize, reusing the install summary with no optional capabilities.
	showInstallSummary(out, p, onboardingPreview{}, f.DryRun, unchanged, mentionDryRunFlag)
	fmt.Fprintf(out, "Source commit %s (requested %s)\n", shortHash(commit), f.Rev)

	// 7. Apply or save.
	if f.DryRun {
		fmt.Fprintln(out, "Preview: installation was not changed.")
		return nil
	}
	if f.Out != "" {
		if err := management.SavePlan(f.Out, p); err != nil {
			return err
		}
		fmt.Fprintf(out, "Plan saved to %s; run hive apply --plan %s to apply it.\n", f.Out, f.Out)
		return nil
	}
	if unchanged {
		result, err := (management.Engine{}).Apply(p)
		if err != nil {
			return err
		}
		reportApplyResult(out, "Hive is already up to date", result)
		return nil
	}
	if !interactive {
		return fmt.Errorf("an interactive terminal is required to confirm; use --dry-run to preview or --out FILE to save a plan for hive apply")
	}
	decision, err := terminal.Confirm("Apply these changes?", false)
	if err != nil {
		return err
	}
	if decision != installApply {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return nil
	}
	result, err := (management.Engine{}).Apply(p)
	if err != nil {
		return err
	}
	reportApplyResult(out, "Hive updated", result)
	return nil
}

// planFromCommit extracts commit by its full hash into a private temporary
// directory, builds the install plan from it, and removes the directory on
// every return path before the caller shows or applies the plan.
func planFromCommit(gitPath string, env []string, source, commit string, o management.Options) (management.Plan, error) {
	prefix := "hive-" + shortHash(commit)
	tarBytes, err := archiveGitCommit(gitPath, env, source, commit, prefix)
	if err != nil {
		return management.Plan{}, err
	}
	gz, err := adaptGitArchive(tarBytes)
	if err != nil {
		return management.Plan{}, err
	}
	// os.TempDir() is resolved through target.Canonical before use: on
	// macOS it names a path under /var or /tmp, both symlinks, and
	// target.Safe (which BuildPlan applies while walking the extracted
	// source) rejects any symlink ancestor. Extracting under the
	// unresolved path would make every `hive update` fail on macOS, not
	// only in tests.
	extractDest, err := target.Canonical(os.TempDir())
	if err != nil {
		return management.Plan{}, err
	}
	root, err := distribution.Extract(gz, extractDest)
	if err != nil {
		return management.Plan{}, err
	}
	defer os.RemoveAll(filepath.Dir(root))

	// BindSourceCommit recomputes the plan ID: the ID is a hash over the
	// whole plan, so it must reflect SourceCommit or validatePlan,
	// PlanUnchanged and Apply reject the plan as tampered.
	o.Source = root
	p, err := management.BuildPlan("install", o)
	if err != nil {
		return management.Plan{}, err
	}
	return management.BindSourceCommit(p, commit)
}

// reportApplyResult renders Apply's result string. Engine.Apply may append
// "; warning: ..." to either "unchanged" or a transaction ID when recording
// the source commit failed after the installation itself already committed
// (see management.recordSourceCommit); that warning is surfaced, never
// swallowed, and the result is never compared with == "unchanged" since it
// may carry that suffix.
func reportApplyResult(out io.Writer, verb, result string) {
	id, warning, hasWarning := strings.Cut(result, "; ")
	fmt.Fprintf(out, "%s (%s).\n", verb, id)
	if hasWarning {
		fmt.Fprintln(out, warning)
	}
	fmt.Fprintln(out, "Open new CLI sessions.")
}

// filteredGitEnv copies the process environment without the variables Git
// itself clears when it switches repositories (`git rev-parse
// --local-env-vars`: GIT_DIR, GIT_OBJECT_DIRECTORY, GIT_CONFIG_PARAMETERS
// and the rest). Inherited from a hook or another repository's script, they
// would outrank the explicit -C source directory of every Git call here.
func filteredGitEnv(gitPath string) ([]string, error) {
	cmd := exec.Command(gitPath, "rev-parse", "--local-env-vars")
	cmd.Dir = os.TempDir()
	listed, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git rev-parse --local-env-vars: %w", err)
	}
	drop := map[string]bool{}
	for _, key := range strings.Fields(string(listed)) {
		drop[key] = true
	}
	var out []string
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if !drop[key] {
			out = append(out, kv)
		}
	}
	return out, nil
}

// resolveCommit resolves rev to its full commit hash inside source, without
// ever passing the operator's revision to anything but rev-parse. The
// ^{commit} suffix, together with the leading-dash rejection in update,
// keeps rev from being read as an option even before Git parses it.
func resolveCommit(gitPath string, env []string, source, rev string) (string, error) {
	cmd := exec.Command(gitPath, "-C", source, "rev-parse", "--verify", "--quiet", rev+"^{commit}")
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := fmt.Sprintf("%s is not a Git checkout, or %q is not a known commit", source, rev)
		if detail := strings.TrimSpace(stderr.String()); detail != "" {
			msg += ": " + detail
		}
		return "", fmt.Errorf("%s", msg)
	}
	// BindSourceCommit validates the hash format.
	return strings.TrimSpace(stdout.String()), nil
}

// shortHash is the short form used for the extraction prefix and the
// operator-facing summary line; it does not need to match `git rev-parse
// --short`; a stable prefix length avoids one more Git invocation.
func shortHash(commit string) string {
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}

// archiveSizeLimit is the most bytes read from git archive; tests lower it.
var archiveSizeLimit = distribution.MaxPackageBytes

// archiveGitCommit runs `git archive` for the resolved commit hash (never
// the operator's own revision text) and reads it with a limit one byte past
// distribution.MaxPackageBytes, the same ceiling Extract enforces, so an
// oversized archive is rejected before it is fully buffered. tar, not
// tar.gz, is requested so no repository-configured compressor
// (tar.tar.gz.command) ever runs.
func archiveGitCommit(gitPath string, env []string, source, commit, prefix string) ([]byte, error) {
	cmd := exec.Command(gitPath, "-C", source, "archive", "--format=tar", "--prefix="+prefix+"/", commit)
	cmd.Env = env
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(stdout, archiveSizeLimit+1))
	if readErr != nil || int64(len(data)) > archiveSizeLimit {
		// Git may still be writing into a pipe nobody reads any more; stop it
		// so Wait returns instead of blocking forever.
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		if readErr != nil {
			return nil, readErr
		}
		return nil, fmt.Errorf("archived commit exceeds the package size limit")
	}
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("git archive %s: %s", commit, strings.TrimSpace(stderr.String()))
	}
	return data, nil
}

// adaptGitArchive turns raw `git archive --format=tar` output into the
// gzip-compressed form distribution.Extract accepts. Extract itself is the
// bootstrap's security boundary and is left unchanged (it also serves the
// network-facing downloader); this adapter only normalizes the two ways
// Git's own tar output differs from what Extract requires:
//   - it drops the pax_global_header entry (tar.TypeXGlobalHeader) Git
//     writes with the commit hash in its comment;
//   - it strips the trailing slash Git gives every directory name, which
//     Extract's archivePath rejects.
//
// Every other entry is copied through unchanged.
func adaptGitArchive(raw []byte) ([]byte, error) {
	reader := tar.NewReader(bytes.NewReader(raw))
	var buf bytes.Buffer
	// Extract decompresses this right away, so spend as little as possible
	// on compression.
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	tw := tar.NewWriter(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid archive from git archive: %w", err)
		}
		if header.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		header.Name = strings.TrimSuffix(header.Name, "/")
		if err := tw.WriteHeader(header); err != nil {
			return nil, err
		}
		if _, err := io.Copy(tw, reader); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
