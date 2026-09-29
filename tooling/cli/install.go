package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

var installerHosts = []string{"claude", "codex", "cursor", "grok", "opencode", "pi"}

type hostCandidate struct {
	Name       string
	Detected   bool
	Registered bool
	Legacy     bool
}

type installTerminal struct {
	reader      *bufio.Reader
	out         io.Writer
	interactive bool
}

// SelectHosts, SelectProviders and Confirm are thin methods
// over the plain-text functions below, which own the printed text.
func (t installTerminal) SelectHosts(candidates []hostCandidate) ([]string, bool, error) {
	return selectInstallerHosts(t, candidates)
}

func (t installTerminal) SelectProviders(offers []providerOffer) ([]providerRequest, bool, error) {
	return selectProviderRequests(t, offers)
}

func (t installTerminal) Confirm(prompt string, allowBack bool) (installDecision, error) {
	return confirmInstall(t, prompt, allowBack)
}

type installDecision int

const (
	installCancelled installDecision = iota
	installApply
	installBack
)

type onboardingAdapter interface {
	Detect(management.Options) ([]providerOffer, error)
	Plan(management.Plan, []providerRequest) (onboardingPreview, error)
	Runner() management.ExternalRunner
}

type providerOffer struct {
	ID     string
	Name   string
	Source string
	// ManualOnly marks a capability whose outcome is unconditionally manual
	// in this build (no native recipe validation available): the wizard
	// skips the exact-version prompt for it and labels it accordingly.
	ManualOnly bool
	Effects    []string
}

type providerRequest struct {
	ID      string
	Version string
}

type onboardingPreview struct {
	Steps   []management.ExternalStep
	Details []providerDetail
}

type providerDetail struct {
	ID, Version, Source string
	Effects             []string
}

type onboardingInput struct {
	Options management.Options
	// DryRun mirrors the invocation's --dry-run flag as a plain bool: an
	// adapter that needs to know must read this field, never re-scan raw CLI
	// arguments for it.
	DryRun bool
}

type onboardingAdapterFactory func(onboardingInput) (onboardingAdapter, error)

type installDependencies struct {
	DiscoverHosts func(management.Options) ([]hostCandidate, error)
	RequiredHosts func(management.Options) ([]string, error)
	// Pending reports which of the two independent journals (optional
	// onboarding, or a core transaction) blocks a new operation, so the flow
	// never reads pending.json by name to answer that question itself.
	Pending           func(string) (management.PendingKind, error)
	RecoverOnboarding func(string, onboardingAdapter) (management.OnboardingResult, error)
	RecoverCore       func(string) (string, error)
	AdapterFactory    onboardingAdapterFactory
	// BindRetainedInstaller is set only by the online bootstrap entry point.
	// It runs once, right after the operator confirms and before the plan
	// mutates anything: it retains the already-verified manager and package
	// privately, then binds that retained installer's identity into the
	// plan so a later `hive recover` (from another terminal, offline) can
	// find and reverify it. Ordinary local installs leave this nil.
	BindRetainedInstaller func(management.Options, management.Plan) (management.Plan, error)
}

func install(args []string, in io.Reader, out io.Writer, interactive bool) error {
	return installWithAdapterFactory(args, in, out, interactive, nativeProviderAdapterFactory)
}

// installWithAdapterFactory creates a fresh adapter for each invocation. The
// CLI retains no cross-run adapter state.
func installWithAdapterFactory(args []string, in io.Reader, out io.Writer, interactive bool, factory onboardingAdapterFactory) error {
	return installWithDependencies(args, in, out, interactive, defaultInstallDependencies(factory))
}

func defaultInstallDependencies(factory onboardingAdapterFactory) installDependencies {
	return installDependencies{
		DiscoverHosts:     detectInstallerHosts,
		RequiredHosts:     management.RequiredHosts,
		Pending:           management.Pending,
		RecoverOnboarding: recoverInstallOnboarding,
		RecoverCore:       (management.Engine{}).Recover,
		AdapterFactory:    factory,
	}
}

func installWithDependencies(args []string, in io.Reader, out io.Writer, interactive bool, dependencies installDependencies) error {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(out)
	o := management.Options{Scope: "user"}
	var hosts string
	var dry bool
	fs.StringVar(&hosts, "hosts", "", "selected CLIs, comma-separated; detected by default")
	fs.StringVar(&o.Home, "home", "", "synthetic home; ignores user paths and executables")
	fs.StringVar(&o.StateDir, "state-dir", "", "private state directory")
	fs.StringVar(&o.Source, "source", ".", "this distribution content")
	fs.BoolVar(&dry, "dry-run", false, "preview changes without applying them")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("install accepts no positional arguments")
	}
	if hosts != "" {
		for _, h := range strings.Split(hosts, ",") {
			o.Hosts = append(o.Hosts, strings.TrimSpace(h))
		}
	}
	return runInstallFlow(o, dry, in, out, interactive, dependencies)
}

// runInstallFlow is the interactive core shared by a local offline install
// and the online bootstrap hand-off: both resolve an Options value (hosts,
// home, state directory, and a source distribution tree) before reaching
// here, then walk the same detection/summary/consent/apply loop. Only
// bootstrap sets dependencies.BindRetainedInstaller. It runs package
// verification, options normalization, the pending-operation check and the
// onboarding wizard.
func runInstallFlow(o management.Options, dry bool, in io.Reader, out io.Writer, interactive bool, dependencies installDependencies) error {
	p := installTerminal{reader: bufio.NewReader(in), out: out, interactive: interactive}
	if err := validateInstallDependencies(dependencies); err != nil {
		return err
	}
	if err := distribution.VerifyIfPackaged(o.Source); err != nil {
		return err
	}
	// explicitStateDir is o's own, pre-normalization StateDir: once
	// NormalizeOptions resolves o.StateDir to some concrete path below, this
	// is the only point that can still tell whether the operator asked for a
	// specific one.
	explicitStateDir := o.StateDir != ""
	_, stateDir, err := management.NormalizeOptions(o)
	if err != nil {
		return err
	}
	o.StateDir = stateDir
	// online is set only by the bootstrap entry point (see bootstrap.go):
	// every recovery message below must name a concrete offline command
	// instead of "./install.sh", a file that bootstrap.sh's temporary
	// checkout never contains.
	online := dependencies.BindRetainedInstaller != nil
	// stillNeeded is ignored here: install's and bootstrap's own recovery
	// phrase names the concrete next command unconditionally
	// (bootstrap_test.go's TestInstallOfflinePendingRecoveryPointsToInstallScript
	// pins this for a successful recovery too). withRecoverySentence embeds
	// the phrase as its own standalone sentence.
	recoveryText := func(stateDir string, stillNeeded bool) string {
		return recoveryPhrase(online, stateDir)
	}
	if handled, err := handlePendingInstallOperation(o, dry, p, out, recoveryText, dependencies); handled {
		return err
	}
	return runOnboardingWizard(o, dry, p, out, online, explicitStateDir, dependencies)
}

func validateInstallDependencies(dependencies installDependencies) error {
	if dependencies.DiscoverHosts == nil || dependencies.RequiredHosts == nil || dependencies.Pending == nil || dependencies.RecoverOnboarding == nil || dependencies.RecoverCore == nil || dependencies.AdapterFactory == nil {
		return fmt.Errorf("incomplete install dependencies")
	}
	return nil
}

// withRecoverySentence appends recovery as its own trailing sentence onto
// base only when recovery is non-empty, so a caller (the recovery view's
// recoverPending) can omit the whole sentence once nothing more needs
// recovering, while install's and bootstrap's own closures, which never
// return "", keep their exact existing wording (matching format and trailing
// period) byte for byte.
func withRecoverySentence(base, recovery string) string {
	if recovery == "" {
		return base
	}
	return base + " " + recovery + "."
}

// handlePendingInstallOperation covers M5: a single Pending call (never a
// direct pending.json read by name) answers whether an optional-onboarding
// journal or a core transaction journal blocks a new operation; onboarding
// takes priority, matching management.Pending's own precedence. handled is
// true whenever the caller must return err as-is (including nil) instead of
// continuing into the host-selection wizard: either a pending recovery was
// resolved (or declined, or deferred by --dry-run) here, or resolving it
// itself failed. recoveryText's stillNeeded argument tells the caller
// whether this particular recovery attempt leaves anything unresolved
// (always false for PendingCore, which only reaches its own call site on
// success; result.Phase != "completed" for PendingOnboarding, which can
// finish "partial").
func handlePendingInstallOperation(o management.Options, dry bool, terminal installTerminal, out io.Writer, recoveryText func(stateDir string, stillNeeded bool) string, dependencies installDependencies) (handled bool, err error) {
	stateDir := o.StateDir
	kind, err := dependencies.Pending(stateDir)
	if err != nil {
		return true, err
	}
	switch kind {
	case management.PendingOnboarding:
		fmt.Fprintf(out, "Optional onboarding is pending. Recovery directory: %s\n", stateDir)
		if dry {
			return true, nil
		}
		decision, err := terminal.Confirm("Reconcile the pending optional steps?", false)
		if err != nil || decision != installApply {
			return true, err
		}
		adapter, err := dependencies.AdapterFactory(newOnboardingInput(o, dry))
		if err != nil {
			return true, err
		}
		result, err := dependencies.RecoverOnboarding(stateDir, adapter)
		if err != nil {
			return true, err
		}
		fmt.Fprintln(out, withRecoverySentence(fmt.Sprintf("Onboarding recovery: %s (%s).", result.ID, result.Phase), recoveryText(stateDir, result.Phase != "completed")))
		return true, nil
	case management.PendingCore:
		fmt.Fprintf(out, "An operation is pending. Recovery directory: %s\n", stateDir)
		if dry {
			return true, nil
		}
		decision, err := terminal.Confirm("Recover the pending operation?", false)
		if err != nil || decision != installApply {
			return true, err
		}
		result, err := dependencies.RecoverCore(stateDir)
		if err != nil {
			return true, err
		}
		fmt.Fprintln(out, withRecoverySentence(fmt.Sprintf("Recovery: %s.", result), recoveryText(stateDir, false)))
		return true, nil
	}
	return false, nil
}

// runOnboardingWizard walks the host-selection/summary/consent/apply loop
// once no pending operation blocks it.
func runOnboardingWizard(o management.Options, dry bool, terminal installTerminal, out io.Writer, online, explicitStateDir bool, dependencies installDependencies) error {
	stateDir := o.StateDir
	explicitHosts := len(o.Hosts) > 0

	for {
		if !explicitHosts {
			candidates, err := dependencies.DiscoverHosts(o)
			if err != nil {
				return err
			}
			selected, ok, err := terminal.SelectHosts(candidates)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Fprintln(out, "Cancelled. No changes applied.")
				return nil
			}
			o.Hosts = selected
		}

		expanded, ok, err := expandToRequiredHosts(terminal, out, o, dependencies)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(out, "Cancelled. No changes applied.")
			return nil
		}
		o.Hosts = expanded

		p, err := management.BuildPlan("install", o)
		if err != nil {
			return err
		}
		unchanged, err := management.PlanUnchanged(p)
		if err != nil {
			return err
		}
		adapter, err := dependencies.AdapterFactory(newOnboardingInput(o, dry))
		if err != nil {
			return err
		}
		preview, ok, err := previewInstallOnboarding(terminal, o, p, adapter)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(out, "Cancelled. No changes applied.")
			return nil
		}
		showInstallSummary(out, p, preview, dry, unchanged, true)
		if dry {
			fmt.Fprintln(out, "Preview: installation was not changed.")
			return nil
		}
		decision, err := terminal.Confirm("Apply these changes?", !explicitHosts)
		if err != nil {
			return err
		}
		switch decision {
		case installBack:
			o.Hosts = nil
			continue
		case installCancelled:
			fmt.Fprintln(out, "Cancelled. No changes applied.")
			return nil
		}
		// Consent is now in hand. Bootstrap's hook retains the already
		// verified manager and package and binds their identity into the
		// plan here, after consent and strictly before applyInstallOnboarding
		// mutates anything, so recovery from another terminal can find them.
		if dependencies.BindRetainedInstaller != nil {
			p, err = dependencies.BindRetainedInstaller(o, p)
			if err != nil {
				return err
			}
		}
		result, err := applyInstallOnboarding(p, preview, adapter)
		return finalizeInstallResult(out, result, err, online, explicitStateDir, stateDir, false)
	}
}

// describeRequiredHosts covers U8's notice: when shared resources force
// additional hosts into the selection, it names them and the affected
// resources, and the consequence of accepting, on out. needsConsent is false,
// with hosts o.Hosts unchanged and nothing printed, when nothing needs
// expanding; otherwise hosts is the full required set the operator must accept.
func describeRequiredHosts(out io.Writer, o management.Options, dependencies installDependencies) (hosts []string, needsConsent bool, err error) {
	required, err := dependencies.RequiredHosts(o)
	if err != nil {
		return nil, false, err
	}
	additional := additionalHosts(o.Hosts, required)
	if len(additional) == 0 {
		return o.Hosts, false, nil
	}
	fmt.Fprintf(out, "Shared resources require selecting: %s\n", strings.Join(additional, ", "))
	sharedHostsOptions := o
	sharedHostsOptions.Hosts = required
	if sharedPlan, err := management.BuildPlan("install", sharedHostsOptions); err == nil {
		if resources := sharedResourceRoots(sharedPlan, additional); len(resources) > 0 {
			fmt.Fprintf(out, "Affected shared resources: %s\n", strings.Join(resources, ", "))
		}
	}
	fmt.Fprintln(out, "Accepting will rewrite those resources too, on already-installed hosts that share them.")
	return required, true, nil
}

// expandToRequiredHosts covers U8: when shared resources force additional
// hosts into the selection, name them, the affected resources, and the
// consequence of accepting, then ask for explicit consent. ok=false means
// the operator declined the expansion (already reported by the caller);
// hosts is o.Hosts unchanged when nothing needs expanding.
func expandToRequiredHosts(terminal installTerminal, out io.Writer, o management.Options, dependencies installDependencies) (hosts []string, ok bool, err error) {
	required, needsConsent, err := describeRequiredHosts(out, o, dependencies)
	if err != nil {
		return nil, false, err
	}
	if !needsConsent {
		return o.Hosts, true, nil
	}
	decision, err := terminal.Confirm("Select all required hosts?", false)
	if err != nil {
		return nil, false, err
	}
	if decision != installApply {
		return nil, false, nil
	}
	return required, true, nil
}

// finalizeInstallResult renders applyInstallOnboarding's outcome (success,
// no-op, or partial) and turns a partial or failed apply into the CLI's own
// error; the core mutation, if any, has already happened by the time this
// runs, so it never decides whether to retry. fromInterface/explicitStateDir
// route every recovery phrase here through recoveryPhraseFor: the CLIs view
// passes true, the install command false.
func finalizeInstallResult(out io.Writer, result management.OnboardingResult, err error, online, explicitStateDir bool, stateDir string, fromInterface bool) error {
	if err != nil {
		if result.Phase == "partial" {
			showPartialOnboardingDetail(out, result, online, explicitStateDir, stateDir, fromInterface)
			return fmt.Errorf("optional capabilities incomplete (%s)", result.ID)
		}
		return fmt.Errorf("installation did not complete: %w; %s to check recovery", err, recoveryPhraseFor(fromInterface, explicitStateDir, online, stateDir))
	}
	if result.ID == "unchanged" {
		// management.Engine.Apply's own literal sentinel ID (no exported
		// constant) for a no-op core with no provider steps: render a
		// distinct user-facing phrase instead of leaking the bare
		// English literal (U10).
		fmt.Fprintln(out, "Hive was already installed and verified (no changes). Open new CLI sessions.")
		return nil
	}
	fmt.Fprintf(out, "Hive installed and verified (%s).\nOpen new CLI sessions.\n", result.ID)
	return nil
}

// retainedManagerPath returns the absolute path to a manager binary retained
// under stateDir by a prior consented bootstrap, so an online-flow recovery
// message can name a concrete offline command instead of pointing at a
// script bootstrap.sh has already deleted. It reports ok=false when nothing
// is retained yet.
func retainedManagerPath(stateDir string) (path string, ok bool) {
	entries, err := os.ReadDir(filepath.Join(stateDir, "installers"))
	if err != nil {
		return "", false
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for i := len(names) - 1; i >= 0; i-- {
		candidate := filepath.Join(stateDir, "installers", names[i], "manager")
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate, true
		}
	}
	return "", false
}

// recoveryPhrase names the concrete next command for a pending or
// interrupted operation, as a lower-case clause fit for embedding mid
// sentence. Online (bootstrap) invocations must never point to
// ./install.sh: bootstrap.sh deletes its own temporary manager on exit, so
// the only thing left to run offline, from another terminal, is the
// manager already retained under the state directory.
func recoveryPhrase(online bool, stateDir string) string {
	if online {
		if path, ok := retainedManagerPath(stateDir); ok {
			return fmt.Sprintf("run the retained manager's recover (%s recover --state-dir %s)", path, stateDir)
		}
		return "run bootstrap.sh again once you have network access"
	}
	return "run ./install.sh again"
}

// recoveryPhraseFor is recoveryPhrase's own interface-aware wrapper: every
// recovery text reached while fromInterface is true (the CLIs view) must name
// hive recover, a subcommand the interface's own operator can always run
// directly, instead of ./install.sh (a script this process may not have been
// launched from at all) or a bootstrap-only retained-manager phrase, neither
// of which apply inside the full-screen interface. When the interface was
// opened against an explicit --state-dir, that phrase names it too, so the
// operator recovers the same, possibly synthetic, state they are looking at
// rather than the real user's default. fromInterface false defers to
// recoveryPhrase unchanged, for install's and bootstrap's own callers,
// ignoring explicitStateDir.
func recoveryPhraseFor(fromInterface, explicitStateDir, online bool, stateDir string) string {
	if fromInterface {
		if explicitStateDir {
			return "run hive recover --state-dir " + shellQuote(stateDir)
		}
		return "run hive recover"
	}
	return recoveryPhrase(online, stateDir)
}

// capitalize upper-cases a phrase's first byte for sentence-initial use,
// without otherwise altering it (every phrase here is plain ASCII).
func capitalize(phrase string) string {
	if phrase == "" {
		return phrase
	}
	return strings.ToUpper(phrase[:1]) + phrase[1:]
}

func newOnboardingInput(o management.Options, dryRun bool) onboardingInput {
	return onboardingInput{Options: o, DryRun: dryRun}
}

// previewInstallOnboarding's bool result mirrors selectInstallerHosts's own
// ok/cancel signal: false means the operator cancelled (including a genuine
// EOF while entering a capability version), which the caller must report the
// same way as any other cancellation — quietly, with no error and no writes.
func previewInstallOnboarding(terminal installTerminal, o management.Options, p management.Plan, adapter onboardingAdapter) (onboardingPreview, bool, error) {
	offers, err := adapter.Detect(o)
	if err != nil {
		return onboardingPreview{}, false, err
	}
	requests, ok, err := terminal.SelectProviders(offers)
	if err != nil || !ok {
		return onboardingPreview{}, ok, err
	}
	preview, err := adapter.Plan(p, requests)
	return preview, true, err
}

func applyInstallOnboarding(p management.Plan, preview onboardingPreview, adapter onboardingAdapter) (management.OnboardingResult, error) {
	return (management.Engine{}).Onboard(p, preview.Steps, adapter.Runner())
}

func recoverInstallOnboarding(stateDir string, adapter onboardingAdapter) (management.OnboardingResult, error) {
	return (management.Engine{}).RecoverOnboarding(stateDir, adapter.Runner())
}

// recoverWithAdapterFactory keeps command recovery on the same per-invocation
// adapter path as install.
func recoverWithAdapterFactory(o management.Options, factory onboardingAdapterFactory) (management.OnboardingResult, error) {
	adapter, err := factory(newOnboardingInput(o, false))
	if err != nil {
		return management.OnboardingResult{}, err
	}
	return recoverInstallOnboarding(o.StateDir, adapter)
}

func detectInstallerHosts(o management.Options) ([]hostCandidate, error) {
	registered, err := management.RegisteredHosts(o)
	if err != nil {
		return nil, err
	}
	legacy, err := management.DetectLegacyHosts(o)
	if err != nil {
		return nil, err
	}
	seenRegistered := hostSet(registered)
	seenLegacy := hostSet(legacy)
	candidates := make([]hostCandidate, 0, len(installerHosts))
	for _, host := range installerHosts {
		candidate := hostCandidate{Name: host, Registered: seenRegistered[host], Legacy: seenLegacy[host]}
		if o.Home == "" {
			if _, err := exec.LookPath(host); err == nil {
				candidate.Detected = true
			}
		}
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func selectInstallerHosts(terminal installTerminal, candidates []hostCandidate) ([]string, bool, error) {
	if !terminal.interactive {
		return nil, false, fmt.Errorf("an interactive terminal is required to select hosts; use --hosts to select them explicitly")
	}
	fmt.Fprintln(terminal.out, "Select CLI hosts")
	for i, candidate := range candidates {
		status := "not detected"
		if candidate.Detected {
			status = "executable detected"
		} else if candidate.Registered {
			status = "registered by Hive"
		} else if candidate.Legacy {
			status = "legacy installation detected"
		}
		fmt.Fprintf(terminal.out, "%d. %s (%s)\n", i+1, candidate.Name, status)
	}
	for {
		fmt.Fprint(terminal.out, "Enter host numbers separated by commas, or press Enter to cancel: ")
		line, complete, err := terminal.readLine()
		if err != nil {
			return nil, false, err
		}
		if !complete || strings.TrimSpace(line) == "" {
			return nil, false, nil
		}
		indices := strings.Split(line, ",")
		selected := map[string]bool{}
		var invalid error
		for _, value := range indices {
			var index int
			if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &index); err != nil || index < 1 || index > len(candidates) {
				invalid = fmt.Errorf("select host numbers from 1 to %d", len(candidates))
				break
			}
			if selected[candidates[index-1].Name] {
				invalid = fmt.Errorf("duplicate host selection %q", candidates[index-1].Name)
				break
			}
			selected[candidates[index-1].Name] = true
		}
		if invalid != nil {
			// Re-prompt in place instead of ending the whole install: prior
			// answers (there are none yet at this step) are never lost.
			fmt.Fprintln(terminal.out, invalid.Error())
			continue
		}
		result := make([]string, 0, len(selected))
		for host := range selected {
			result = append(result, host)
		}
		sort.Strings(result)
		return result, true, nil
	}
}

// selectProviderRequests returns ok=false only when the operator cancels the
// whole install: a genuine EOF (input ends outright, not merely an empty
// line) while selecting capabilities or entering one's version. Any other
// invalid entry re-prompts in place, showing the same error text the CLI
// always used for it, keeping every earlier capability's already-recorded
// version.
func selectProviderRequests(terminal installTerminal, offers []providerOffer) ([]providerRequest, bool, error) {
	if len(offers) == 0 {
		return nil, true, nil
	}
	if !terminal.interactive {
		return nil, true, nil
	}
	fmt.Fprintln(terminal.out, "Select optional capabilities")
	for i, offer := range offers {
		version := "exact version required"
		if offer.ManualOnly {
			version = "manual instructions only"
		}
		fmt.Fprintf(terminal.out, "%d. %s (%s; %s)\n", i+1, offer.Name, offer.Source, version)
	}
	for {
		fmt.Fprint(terminal.out, "Enter capability numbers separated by commas, or press Enter to skip: ")
		line, complete, err := terminal.readLine()
		if err != nil {
			return nil, false, err
		}
		if !complete {
			return nil, false, nil
		}
		if line == "" {
			return nil, true, nil
		}
		indices := strings.Split(line, ",")
		chosen := make([]int, 0, len(indices))
		seen := map[int]bool{}
		var invalid error
		for _, value := range indices {
			var index int
			if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &index); err != nil || index < 1 || index > len(offers) || seen[index] {
				invalid = fmt.Errorf("select capability numbers from 1 to %d without duplicates", len(offers))
				break
			}
			seen[index] = true
			chosen = append(chosen, index)
		}
		if invalid != nil {
			fmt.Fprintln(terminal.out, invalid.Error())
			continue
		}
		requests := make([]providerRequest, 0, len(chosen))
		cancelled := false
		for _, index := range chosen {
			offer := offers[index-1]
			if offer.ManualOnly {
				// An unconditionally manual outcome needs no version: it
				// never reaches the automatic recipe that would validate one.
				requests = append(requests, providerRequest{ID: offer.ID})
				continue
			}
			version, ok, err := readProviderVersion(terminal, offer)
			if err != nil {
				return nil, false, err
			}
			if !ok {
				cancelled = true
				break
			}
			requests = append(requests, providerRequest{ID: offer.ID, Version: version})
		}
		if cancelled {
			return nil, false, nil
		}
		return requests, true, nil
	}
}

func readProviderVersion(terminal installTerminal, offer providerOffer) (string, bool, error) {
	prompt := fmt.Sprintf("Exact version for %s", offer.Name)
	for {
		fmt.Fprint(terminal.out, prompt+": ")
		version, complete, err := terminal.readLine()
		if err != nil {
			return "", false, err
		}
		if !complete {
			return "", false, nil
		}
		if version == "" {
			fmt.Fprintf(terminal.out, "exact version required for %s\n", offer.Name)
			continue
		}
		return version, true, nil
	}
}

// changedChangeCount counts only the Changes whose Before and After differ,
// excluding the no-op entries BuildPlan always includes for an already
// up-to-date resource.
func changedChangeCount(changes []management.Change) int {
	n := 0
	for _, ch := range changes {
		if !reflect.DeepEqual(ch.Before, ch.After) {
			n++
		}
	}
	return n
}

// mentionDryRunFlag governs the closing "Use --dry-run..." hint: true for
// the commands (install's and update's own --dry-run flag really exists
// there), false when a full-screen view calls this (it has no such flag to
// suggest).
func showInstallSummary(out io.Writer, p management.Plan, preview onboardingPreview, dry, unchanged, mentionDryRunFlag bool) {
	// A --source checkout without VERSION or release.json (for example, a
	// partial development tree) leaves p.Product nil; that identifies a
	// development build rather than a defect, and must never be dereferenced.
	productVersion := "development build"
	if p.Product != nil {
		productVersion = p.Product.Version
	}
	fmt.Fprintf(out, "Hive %s · %s\n", productVersion, strings.Join(p.Hosts, ", "))
	// hiveChanged is about the Hive files themselves, not the plan as a
	// whole: a voice-only update leaves every p.Changes entry a no-op
	// (Before == After) while the plan overall is not unchanged (its voice
	// changes are real), and the header and the file count below must both
	// reflect that distinction rather than count every no-op entry.
	hiveChanged := changedChangeCount(p.Changes) > 0 || len(p.Legacy) > 0
	if unchanged {
		msg := "Hive's core is already up to date."
		if len(preview.Details) > 0 {
			msg += "\nThe selected optional capabilities still require confirmation."
		}
		fmt.Fprintln(out, msg)
	} else if !hiveChanged {
		fmt.Fprintln(out, "Hive's core is already up to date.")
	} else if len(p.Legacy) > 0 {
		fmt.Fprintln(out, "Migrate legacy Hive and install the rebuild")
	} else {
		fmt.Fprintln(out, "Install / update")
	}
	fmt.Fprintf(out, "Private backups: %s\n", filepath.Join(p.StateDir, "transactions"))
	if !hiveChanged {
		fmt.Fprintf(out, "Hive files checked: %d; none change.\n", len(p.Changes))
	} else {
		fmt.Fprintf(out, "Hive files to install or update: %d; legacy changes: %d\n", changedChangeCount(p.Changes), len(p.Legacy))
	}
	if len(p.Voice) > 0 {
		fmt.Fprintf(out, "Voice files to regenerate: %d\n", len(p.Voice))
	}
	if p.VoiceWarning != "" {
		fmt.Fprintln(out, p.VoiceWarning)
	}
	if len(preview.Details) == 0 {
		fmt.Fprintln(out, "Optional capabilities: none selected.")
	}
	for _, detail := range preview.Details {
		if detail.Version != "" {
			fmt.Fprintf(out, "Optional capability: %s %s from %s\n", detail.ID, detail.Version, detail.Source)
		} else {
			fmt.Fprintf(out, "Optional capability: %s from %s\n", detail.ID, detail.Source)
		}
		if len(detail.Effects) > 0 {
			fmt.Fprintln(out, "  Effects:")
			for _, effect := range detail.Effects {
				for _, line := range wrapText(effect, terminalWrapWidth) {
					fmt.Fprintf(out, "    %s\n", line)
				}
			}
		}
	}
	destinations := map[string]bool{}
	for _, ch := range p.Changes {
		destinations[ch.Target.Path] = true
	}
	for _, ch := range p.Legacy {
		destinations[ch.Path] = true
	}
	var paths []string
	for path := range destinations {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	shown := paths
	if !dry {
		shown = installRoots(paths, p.Config)
	}
	for _, path := range shown {
		fmt.Fprintf(out, "  %s\n", path)
	}
	if !dry && mentionDryRunFlag {
		fmt.Fprintln(out, "Use --dry-run to see the full file list before applying.")
	}
	fmt.Fprintln(out, "Close these CLI sessions before continuing.")
}

// showPartialOnboardingDetail lists every provider step's terminal status so a
// partial outcome is never reported as a single opaque message: the core is
// installed, but the operator must see exactly which optional capability
// needs manual follow-up. online/stateDir let nextOnboardingStepAction name a
// concrete recovery command instead of a hard-coded ./install.sh (see
// recoveryPhrase).
func showPartialOnboardingDetail(out io.Writer, result management.OnboardingResult, online, explicitStateDir bool, stateDir string, fromInterface bool) {
	fmt.Fprintf(out, "Partial installation (%s).\n", result.ID)
	fmt.Fprintln(out, "The core was installed; optional capabilities pending:")
	for _, step := range result.Steps {
		fmt.Fprintf(out, "  %s: %s\n", step.Step.ID, step.Status)
		// Repeat the human reason and next action here too, not only in the
		// pre-confirmation summary: this is the last thing the operator sees.
		var decoded providers.Step
		if json.Unmarshal(step.Step.Payload, &decoded) == nil && decoded.ManualReason != "" {
			printLabeled(out, "    Reason: ", decoded.ManualReason)
		}
		if action := nextOnboardingStepAction(step.Status, online, explicitStateDir, stateDir, fromInterface); action != "" {
			printLabeled(out, "    Next action: ", action)
		}
	}
}

// printLabeled writes label+text wrapped to 80 columns, indenting continuation
// lines under the text so the label is printed once.
func printLabeled(out io.Writer, label, text string) {
	indent := strings.Repeat(" ", len(label))
	for i, line := range wrapText(text, 80-len(label)) {
		if i == 0 {
			fmt.Fprintf(out, "%s%s\n", label, line)
		} else {
			fmt.Fprintf(out, "%s%s\n", indent, line)
		}
	}
}

// nextOnboardingStepAction turns a provider step's terminal status into the
// concrete next action for the operator, matching design.md's per-provider
// states (pending -> running -> verified|failed|unknown|skipped|auth_pending).
func nextOnboardingStepAction(status string, online, explicitStateDir bool, stateDir string, fromInterface bool) string {
	recovery := recoveryPhraseFor(fromInterface, explicitStateDir, online, stateDir)
	switch status {
	case management.StepManual:
		return "Complete the installation following the official instructions, then " + recovery + " to confirm it."
	case management.StepAuthPending:
		return "Complete the pending sign-in, then " + recovery + " to confirm it."
	case management.StepFailed:
		return "Review the error reported by the provider, then " + recovery + " to retry."
	case management.StepUnknown:
		return capitalize(recovery) + " to start recovery; it will not repeat without reconciling the outcome."
	case management.StepSkipped:
		return "This capability was not installed; you may select it again on a future run."
	default:
		return ""
	}
}

func hostSet(hosts []string) map[string]bool {
	set := make(map[string]bool, len(hosts))
	for _, host := range hosts {
		set[host] = true
	}
	return set
}

func additionalHosts(selected, required []string) []string {
	selectedSet := hostSet(selected)
	additional := make([]string, 0, len(required))
	for _, host := range required {
		if !selectedSet[host] {
			additional = append(additional, host)
		}
	}
	sort.Strings(additional)
	return additional
}

// sharedResourceRoots names the shared destinations that pull in a required
// host: every change whose consumers include one of the additional hosts,
// collapsed to the same root directories showInstallSummary already uses so
// the operator sees ".agents" once instead of dozens of individual files.
func sharedResourceRoots(p management.Plan, additional []string) []string {
	additionalSet := hostSet(additional)
	seen := map[string]bool{}
	for _, ch := range p.Changes {
		if ch.After == nil {
			continue
		}
		for _, c := range ch.After.Consumers {
			if additionalSet[c.Host] {
				seen[ch.Target.Path] = true
				break
			}
		}
	}
	paths := make([]string, 0, len(seen))
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return installRoots(paths, p.Config)
}

func installRoots(paths []string, c target.Config) []string {
	roots := []string{c.CodexHome, c.ClaudeHome, c.GrokHome, c.PiHome, c.OpenCodeHome, c.CursorHome, filepath.Join(c.Home, ".agents"), filepath.Join(c.Home, ".codex"), filepath.Join(c.Home, ".claude"), filepath.Join(c.Home, ".grok"), filepath.Join(c.Home, ".pi", "agent"), filepath.Join(c.Home, ".config", "opencode")}
	sort.Slice(roots, func(i, j int) bool { return len(roots[i]) > len(roots[j]) })
	seen := map[string]bool{}
	for _, path := range paths {
		displayed := path
		for _, root := range roots {
			if root != "" && (path == root || strings.HasPrefix(path, root+string(filepath.Separator))) {
				displayed = root
				break
			}
		}
		seen[displayed] = true
	}
	result := make([]string, 0, len(seen))
	for path := range seen {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

// terminalWrapWidth keeps every wrapped detail line at or under 80 columns
// once combined with its indent, for a plain terminal without line wrapping.
const terminalWrapWidth = 74

// wrapText breaks text into lines of at most width columns, splitting only at
// spaces so a source label or sentence is never cut mid-word.
func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	lines := make([]string, 0, 1)
	line := words[0]
	for _, word := range words[1:] {
		if len(line)+1+len(word) > width {
			lines = append(lines, line)
			line = word
			continue
		}
		line += " " + word
	}
	lines = append(lines, line)
	return lines
}

func (terminal installTerminal) readLine() (string, bool, error) {
	line, err := terminal.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", false, err
	}
	if err == io.EOF {
		return "", false, nil
	}
	return strings.TrimSpace(line), true, nil
}

func confirmInstall(terminal installTerminal, prompt string, allowBack bool) (installDecision, error) {
	if !terminal.interactive {
		return installCancelled, fmt.Errorf("an interactive terminal is required to confirm; use --dry-run to inspect")
	}
	if allowBack {
		fmt.Fprintf(terminal.out, "%s [y] apply, [b] back, [N] cancel ", prompt)
	} else {
		fmt.Fprintf(terminal.out, "%s [y/N] ", prompt)
	}
	answer, complete, err := terminal.readLine()
	if err != nil {
		return installCancelled, err
	}
	if !complete {
		return installCancelled, nil
	}
	switch strings.ToLower(answer) {
	case "y", "yes":
		return installApply, nil
	case "b", "back":
		if allowBack {
			return installBack, nil
		}
	}
	return installCancelled, nil
}
