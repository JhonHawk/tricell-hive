// hive update refreshes executable and content from a verified release, or
// content alone from an explicitly selected committed source.
package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
	"tricell-hive/tooling/version"
)

type updateFlags struct {
	Rev, Source, Home, StateDir, Out, Version string
	SourceMode                                bool
	DryRun                                    bool
	// WaitFloor asks the plan to set the Codex wait floor. It is not a command
	// flag: update sets it after the operator accepts the offer.
	WaitFloor bool
}

// updateAdapterFactory builds the adapter that offers optional capabilities in
// update, for the Codex home the plan resolved. Tests replace it.
var updateAdapterFactory = func(f updateFlags, interactive bool, codexHome string) onboardingAdapter {
	a := newNativeProviderAdapter(f.Home != "", f.DryRun, interactive)
	a.codexHome = codexHome
	return a
}

func parseUpdateFlags(args []string) (updateFlags, error) {
	f := updateFlags{Rev: "HEAD", Source: "."}
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.StringVar(&f.Version, "version", "", "exact published product version")
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
	var selectedVersion bool
	fs.Visit(func(v *flag.Flag) {
		if v.Name == "source" || v.Name == "rev" {
			f.SourceMode = true
		}
		if v.Name == "version" {
			selectedVersion = true
		}
	})
	if selectedVersion && f.SourceMode {
		return f, fmt.Errorf("--version conflicts with --source and --rev")
	}
	if selectedVersion && (f.Version == "" || f.Version == "dev" || !version.Valid(f.Version)) {
		return f, fmt.Errorf("invalid --version")
	}
	if f.SourceMode && (f.Source == "" || f.Rev == "" || strings.HasPrefix(f.Rev, "-")) {
		return f, fmt.Errorf("invalid --source or --rev")
	}
	return f, nil
}

// update implements `hive update`. It is testable with injected stdin,
// stdout and an interactive flag, following how install (install.go) is
// structured.
func update(args []string, in io.Reader, out io.Writer, interactive bool) error {
	f, err := parseUpdateFlags(args)
	if err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if !f.SourceMode {
		return updateRelease(f, in, out, interactive)
	}
	terminal := installTerminal{reader: bufio.NewReader(in), out: out, interactive: interactive}
	p, sourceLine, unchanged, err := planUpdate(f)
	if err != nil {
		return err
	}

	// Offer the optional capabilities that change what update writes (the
	// Codex wait floor), only to a person; the choice rebuilds the plan.
	adapter, offers, err := detectUpdateOffers(p, f, interactive)
	if err != nil {
		return err
	}
	var preview onboardingPreview
	if len(offers) > 0 {
		requests, ok, err := terminal.SelectProviders(offers)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(out, "Cancelled. No changes applied.")
			return nil
		}
		if p, unchanged, preview, err = finishUpdateChoice(f, p, unchanged, adapter, offers, requests); err != nil {
			return err
		}
	}

	// 6. Summarize, reusing the install summary with the chosen capabilities.
	showInstallSummary(out, p, preview, f.DryRun, unchanged, true)
	fmt.Fprintln(out, sourceLine)

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
	// The notes are computed before Apply, while the entries are not yet in
	// place; Apply's result stays a transaction ID.
	notes, err := management.HostHookNotes(p)
	if err != nil {
		return err
	}
	notes = append(notes, management.HostHookWarnings(p)...)
	if unchanged {
		result, err := (management.Engine{}).Apply(p)
		if err != nil {
			return err
		}
		reportApplyResult(out, "Hive is already up to date", result)
		printHookNotes(out, notes)
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
	printHookNotes(out, notes)
	reportUpdateViewer(out, p.StateDir, p.SourceCommit)
	return nil
}

// printHookNotes prints what the user must still do or know about the hooks.
func printHookNotes(out io.Writer, notes []string) {
	for _, n := range notes {
		for _, line := range wrapText(n, terminalWrapWidth) {
			fmt.Fprintln(out, line)
		}
	}
}

// detectUpdateOffers asks the adapter which optional capabilities update can
// offer for p. Only an interactive session is offered anything, and only the Codex wait floor; the others
// return no adapter and no offers.
func detectUpdateOffers(p management.Plan, f updateFlags, interactive bool) (onboardingAdapter, []providerOffer, error) {
	if !interactive {
		return nil, nil, nil
	}
	adapter := updateAdapterFactory(f, interactive, p.Config.CodexHome)
	o := management.Options{Scope: "user", Home: f.Home, StateDir: f.StateDir, Hosts: p.Hosts}
	// The native adapter detects the wait floor alone, so an update runs none
	// of the other capabilities' probes. Any other adapter lists everything and
	// update keeps the wait floor, the only capability that changes a file it
	// applies itself, because it runs no onboarding steps.
	if native, ok := adapter.(*nativeProviderAdapter); ok {
		var offers []providerOffer
		if offer, ok := native.waitFloorOffer(o); ok {
			offers = append(offers, offer)
		}
		return adapter, offers, nil
	}
	all, err := adapter.Detect(o)
	var offers []providerOffer
	for _, offer := range all {
		if offer.ID == string(providers.CodexWaitFloor) {
			offers = append(offers, offer)
		}
	}
	return adapter, offers, err
}

// finishUpdateChoice applies the operator's choice among offers. When it
// selects the Codex wait floor the plan is built again from the same commit
// with the option set and unchanged is recomputed: the plan now writes Codex's
// config, so update asks for confirmation even if Hive's files are current.
// Update applies the plan alone, so the preview keeps the capability's details
// and drops any step.
func finishUpdateChoice(f updateFlags, p management.Plan, unchanged bool, adapter onboardingAdapter, offers []providerOffer, requests []providerRequest) (management.Plan, bool, onboardingPreview, error) {
	if waitFloorChosen(offers, requests) {
		f.WaitFloor, f.Rev = true, p.SourceCommit
		rebuilt, _, same, err := planUpdate(f)
		if err != nil {
			return p, unchanged, onboardingPreview{}, err
		}
		p, unchanged = rebuilt, same
	}
	preview, err := adapter.Plan(p, requests)
	preview.Steps = nil
	return p, unchanged, preview, err
}

// planUpdate is update's planning half, shared with the Update view: it
// chooses the registered CLIs, validates the revision, runs Git, resolves the
// commit, extracts it and builds the install plan. It returns the plan, the
// "Source commit" line the summary ends with, and whether the plan changes
// nothing. It writes nothing.
func planUpdate(f updateFlags) (p management.Plan, sourceLine string, unchanged bool, err error) {
	// 1. Choose the CLIs before running Git, so a host-less home fails
	// without extracting anything.
	o := management.Options{Scope: "user", Home: f.Home, StateDir: f.StateDir, CodexWaitFloor: f.WaitFloor}
	hosts, err := management.RegisteredHosts(o)
	if err != nil {
		return management.Plan{}, "", false, err
	}
	if len(hosts) == 0 {
		return management.Plan{}, "", false, fmt.Errorf("no CLI hosts are registered with Hive for this home; run hive install first")
	}
	o.Hosts = hosts

	// 2. Validate the revision before touching Git: empty or dash-led values
	// could otherwise be read as an option by a later Git invocation.
	if f.Rev == "" || strings.HasPrefix(f.Rev, "-") {
		return management.Plan{}, "", false, fmt.Errorf("invalid --rev %q", f.Rev)
	}

	// 3. Run Git in a controlled way: resolved from PATH, no shell, and with
	// the ownership-affecting GIT_* variables stripped from its environment.
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return management.Plan{}, "", false, fmt.Errorf("git not found in PATH; hive update needs Git and a checkout; the offline package installs with install.sh")
	}
	env, err := filteredGitEnv(gitPath)
	if err != nil {
		return management.Plan{}, "", false, err
	}

	commit, err := resolveCommit(gitPath, env, f.Source, f.Rev)
	if err != nil {
		return management.Plan{}, "", false, err
	}

	// 4–5. Extract the commit and plan from it. The plan freezes the source
	// bytes, so the extraction is already gone before the summary and the
	// confirmation prompt.
	p, err = planFromCommit(gitPath, env, f.Source, commit, o)
	if err != nil {
		return management.Plan{}, "", false, err
	}

	unchanged, err = management.PlanUnchanged(p)
	if err != nil {
		return management.Plan{}, "", false, err
	}

	return p, fmt.Sprintf("Source commit %s (requested %s)", shortHash(commit), f.Rev), unchanged, nil
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
	reportApplyOutcome(out, verb, result, true)
}

// reportApplyOutcome is reportApplyResult; sessionNote says whether to end with
// the advice to open new CLI sessions, which only makes sense when a file the
// CLIs read was written.
func reportApplyOutcome(out io.Writer, verb, result string, sessionNote bool) {
	id, warning, hasWarning := strings.Cut(result, "; ")
	fmt.Fprintf(out, "%s (%s).\n", verb, id)
	if hasWarning {
		fmt.Fprintln(out, warning)
	}
	if sessionNote {
		fmt.Fprintln(out, "Open new CLI sessions.")
	}
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
// archive-extraction security boundary and is left unchanged; this adapter only normalizes the two ways
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

// releasePreparation owns disposable verified bytes. Only confirmation or --out
// retains them; dry-run, cancellation and errors dispose of the preparation.
type pendingReleaseUpdateError struct{ journal management.UpdateJournal }

func (e *pendingReleaseUpdateError) Error() string {
	return "pending release update: run " + releaseRecoveryCommand(e.journal.Plan)
}

type releasePreparation struct {
	packageFiles *distribution.PreparedRelease
	plan         management.UpdatePlan
	target       distribution.ReleaseTarget
	unchanged    bool
}

func prepareReleaseUpdate(f updateFlags) (*releasePreparation, error) {
	return prepareReleaseUpdateContext(context.Background(), f)
}

func prepareReleaseUpdateContext(ctx context.Context, f updateFlags) (*releasePreparation, error) {
	if version.Current == "dev" {
		return nil, fmt.Errorf("development build: use hive update --source <checkout> (or --rev REV)")
	}
	cfg, state, err := management.NormalizeOptions(management.Options{Scope: "user", Home: f.Home, StateDir: f.StateDir})
	if err != nil {
		return nil, err
	}
	if pending, err := management.PendingUpdate(state); err != nil {
		return nil, err
	} else if pending != nil {
		return nil, &pendingReleaseUpdateError{journal: *pending}
	}
	hosts, err := management.RegisteredHosts(management.Options{Scope: "user", Home: f.Home, StateDir: f.StateDir})
	if err != nil {
		return nil, err
	}
	if len(hosts) == 0 {
		return nil, fmt.Errorf("no CLI hosts are registered with Hive for this home; run hive install first")
	}
	client := distribution.ReleaseClient{}
	targetRelease, err := client.Resolve(ctx, version.Current, f.Version, runtime.GOOS+"/"+runtime.GOARCH)
	if err != nil {
		return nil, err
	}
	store, err := management.ReleasePackageStore(f.Home)
	if err != nil {
		return nil, err
	}
	_, schema := management.UpdateStateVersions()
	files, err := client.Prepare(ctx, targetRelease, store, schema)
	if err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			_ = files.Cleanup()
		}
	}()
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(management.UpdatePreparePayload{Package: files.Binding, Home: cfg.Home, StateDir: state, NativeHome: !cfg.Synthetic})
	if err != nil {
		return nil, err
	}
	q := distribution.UpdateRequest{Protocol: distribution.UpdateProtocolVersion, OperationID: hex.EncodeToString(nonce[:]), Action: distribution.UpdatePrepare, ProductVersion: targetRelease.ProductVersion, Payload: payload}
	response, err := distribution.InvokeUpdate(ctx, files.Binding.PreparationExecutor, q)
	if err != nil {
		return nil, err
	}
	var result management.UpdateResultPayload
	if err = distribution.DecodeUpdateMessage(bytes.NewReader(response.Payload), &result); err != nil {
		return nil, err
	}
	if result.Plan == nil {
		return nil, fmt.Errorf("target executor returned no update plan")
	}
	p := *result.Plan
	if err = p.Validate(); err != nil {
		return nil, err
	}
	if p.OperationID != q.OperationID || p.Package != files.Binding || p.StateDir != state {
		return nil, fmt.Errorf("target executor changed update binding")
	}
	// A known-schema core can prove an actual no-op. Never decode a future
	// schema just to optimize: that plan still routes to its target executor.
	unchanged := false
	if p.CoreStateVersion == schema && version.Current == targetRelease.ProductVersion && p.Command.Expected.Exists && p.Command.Expected.LinkTarget == p.Package.TargetExecutable.Path {
		var core management.Plan
		if json.Unmarshal(p.CorePlan, &core) == nil {
			unchanged, _ = management.PlanUnchanged(core)
		}
	}
	success = true
	return &releasePreparation{packageFiles: files, plan: p, target: targetRelease, unchanged: unchanged}, nil
}

func releaseUpdateSummary(out io.Writer, r *releasePreparation) {
	fmt.Fprintf(out, "Release: %s → %s\n", version.Current, r.target.ProductVersion)
	if r.target.Downgrade {
		fmt.Fprintln(out, "Explicit downgrade selected.")
	}
	fmt.Fprintf(out, "Hosts: %s\nPackage: %s\nCommand: %s → %s\n", strings.Join(r.plan.Hosts, ", "), r.plan.Package.PackageDirectory, r.plan.Command.Path, r.plan.Package.TargetExecutable.Path)
	_, schema := management.UpdateStateVersions()
	if r.plan.CoreStateVersion == schema {
		var core management.Plan
		if json.Unmarshal(r.plan.CorePlan, &core) == nil {
			showInstallSummary(out, core, onboardingPreview{}, false, r.unchanged, false)
		}
	}
	fmt.Fprintf(out, "Frozen core plan: %s (%d destinations)\n", r.plan.CorePlanID, len(r.plan.Destinations))
	for _, d := range r.plan.Destinations {
		fmt.Fprintln(out, "  "+d.Path)
	}
}

func applyPreparedRelease(r *releasePreparation, out io.Writer) error {
	if r.unchanged {
		fmt.Fprintf(out, "Hive is already up to date (%s).\n", r.target.ProductVersion)
		return nil
	}
	if err := r.packageFiles.Retain(); err != nil {
		return err
	}
	payload, err := json.Marshal(management.UpdateApplyPayload{Plan: r.plan})
	if err != nil {
		return err
	}
	q := distribution.UpdateRequest{Protocol: r.plan.Protocol, OperationID: r.plan.OperationID, Action: distribution.UpdateApply, ProductVersion: r.target.ProductVersion, Payload: payload}
	response, err := management.RouteUpdate(q, r.plan)
	if err != nil {
		return fmt.Errorf("complete update did not finish: %w; inspect hive doctor and recover with %s", err, releaseRecoveryCommand(r.plan))
	}
	if response.Status != distribution.UpdateActivated {
		return fmt.Errorf("core committed; command activation is pending; run %s", releaseRecoveryCommand(r.plan))
	}
	fmt.Fprintf(out, "Hive updated (%s → %s).\nOpen new CLI sessions.\n", version.Current, r.target.ProductVersion)
	reportUpdateViewer(out, r.plan.StateDir, r.plan.Package.Artifact.ProductVersion)
	return nil
}

func releaseRecoveryCommand(p management.UpdatePlan) string {
	return shellQuote(p.Package.TargetExecutable.Path) + " recover --state-dir " + shellQuote(p.StateDir)
}

// targetIdentity is a product version for a release, or the resolved VCS
// commit for source update. Package inventory digests are not VCS revisions.
func reportUpdateViewer(out io.Writer, state, targetIdentity string) {
	info, running, err := viewRunning(newViewPaths(state))
	if err != nil || !running {
		return
	}
	known := false
	if version.Valid(targetIdentity) && targetIdentity != "dev" {
		known = version.Valid(info.Version) && info.Version != "dev" && !info.Modified
		if known && info.Version == targetIdentity {
			return
		}
	} else if targetIdentity != "" && targetIdentity != "dev" && targetIdentity != "unknown" {
		known = info.Revision != "" && info.Revision != "unknown" && !info.Modified
		if known && info.Revision == targetIdentity {
			return
		}
	}
	if info.Version == "" {
		info.Version = "unknown"
	}
	if info.Revision == "" {
		info.Revision = "unknown"
	}
	if !known {
		fmt.Fprintf(out, "The viewer is running version %s, revision %s; cannot compare this runtime identity with the update. Restart it separately if needed.\n", info.Version, info.Revision)
		return
	}
	fmt.Fprintf(out, "The viewer is still running version %s, revision %s; restart it separately to load the update.\n", info.Version, info.Revision)
}

func updateRelease(f updateFlags, in io.Reader, out io.Writer, interactive bool) error {
	if version.Current == "dev" {
		return fmt.Errorf("development build: use hive update --source <checkout> (or --rev REV)")
	}
	if !interactive && !f.DryRun && f.Out == "" {
		return fmt.Errorf("an interactive terminal is required to confirm; use --dry-run to preview or --out FILE to save a plan for hive apply")
	}
	r, err := prepareReleaseUpdate(f)
	if err != nil {
		return err
	}
	defer r.packageFiles.Cleanup()
	releaseUpdateSummary(out, r)
	if f.DryRun {
		fmt.Fprintln(out, "Preview: installation and package storage were not changed.")
		return nil
	}
	if f.Out != "" {
		if err = r.packageFiles.Retain(); err != nil {
			return err
		}
		if err = management.SaveUpdatePlan(f.Out, r.plan); err != nil {
			return err
		}
		fmt.Fprintf(out, "Complete update plan saved to %s; run hive apply --plan %s (offline).\n", f.Out, f.Out)
		return nil
	}
	if r.unchanged {
		return applyPreparedRelease(r, out)
	}
	terminal := installTerminal{reader: bufio.NewReader(in), out: out, interactive: interactive}
	decision, err := terminal.Confirm("Apply this complete update (core and command)?", false)
	if err != nil {
		return err
	}
	if decision != installApply {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return nil
	}
	return applyPreparedRelease(r, out)
}
