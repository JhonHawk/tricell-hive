package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/version"
)

// bootstrap is the online entry point bootstrap.sh execs after it has
// already downloaded a raw manager binary and verified its checksum
// externally (see bootstrap.sh's final line and design.md's "Online entry and trust"
// (archived record heading, originally in Spanish)). This process IS that verified manager; --manager and
// --manager-sha256 name the same file bootstrap.sh just checked, so this
// function can re-derive and cross-check that identity without trusting the
// shell's verification alone.
func bootstrap(args []string, in io.Reader, out io.Writer, interactive bool) error {
	return bootstrapWithAdapterFactory(args, in, out, interactive, nativeProviderAdapterFactory)
}

// bootstrapWithAdapterFactory keeps bootstrap on the same per-invocation
// adapter-construction path as installWithAdapterFactory: no adapter state
// survives across runs.
func bootstrapWithAdapterFactory(args []string, in io.Reader, out io.Writer, interactive bool, factory onboardingAdapterFactory) error {
	return bootstrapWithDependencies(args, in, out, interactive, defaultInstallDependencies(factory))
}

// bootstrapFlags is bootstrap's parsed command line: the Options core install
// already understands, plus the pieces only the online entry point needs.
type bootstrapFlags struct {
	Options                                                     management.Options
	DryRun                                                      bool
	Origin, RequestedVersion, ManagerPath, ManagerSHA256, Hosts string
}

func parseBootstrapFlags(args []string, out io.Writer) (bootstrapFlags, error) {
	var f bootstrapFlags
	f.Options = management.Options{Scope: "user"}
	fs := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	fs.SetOutput(out)
	fs.StringVar(&f.Origin, "origin", "", "trusted HTTPS origin bootstrap.sh already verified a manager from (required)")
	fs.StringVar(&f.RequestedVersion, "version", "", "product version bootstrap.sh already resolved and verified (required)")
	fs.StringVar(&f.ManagerPath, "manager", "", "path to the manager binary bootstrap.sh already checksum-verified (required)")
	fs.StringVar(&f.ManagerSHA256, "manager-sha256", "", "checksum bootstrap.sh already verified for --manager (required)")
	fs.StringVar(&f.Hosts, "hosts", "", "selected CLIs, comma-separated; detected by default")
	fs.StringVar(&f.Options.Home, "home", "", "synthetic home; ignores user paths and executables")
	fs.StringVar(&f.Options.StateDir, "state-dir", "", "private state directory")
	fs.BoolVar(&f.DryRun, "dry-run", false, "preview changes without applying them")
	if err := fs.Parse(args); err != nil {
		return f, err
	}
	if fs.NArg() != 0 {
		return f, fmt.Errorf("bootstrap accepts no positional arguments")
	}
	if f.Origin == "" || f.RequestedVersion == "" || f.ManagerPath == "" || f.ManagerSHA256 == "" {
		return f, fmt.Errorf("bootstrap requires --origin, --version, --manager, and --manager-sha256")
	}
	return f, nil
}

// verifyRetainedManager re-derives and cross-checks the manager binary's
// identity instead of trusting bootstrap.sh's own shell verification alone.
func verifyRetainedManager(managerPath, managerSHA256 string) error {
	managerBytes, err := os.ReadFile(managerPath)
	if err != nil {
		return fmt.Errorf("reading the verified manager at --manager: %w", err)
	}
	if distribution.Digest(managerBytes) != managerSHA256 {
		return fmt.Errorf("the file at --manager no longer matches --manager-sha256")
	}
	return nil
}

// resolveBootstrapRelease downloads and validates the release index for
// requestedVersion, returning the entry for this platform once its raw
// binary checksum is bound to the manager bootstrap.sh already verified: a
// stale or tampered index must never serve package bytes for an unrelated
// build under the same trusted, already-running manager.
func resolveBootstrapRelease(ctx context.Context, origin, requestedVersion, managerSHA256 string) (*distribution.DownloadRelease, error) {
	indexData, err := distribution.Download(ctx, origin, "versions/"+requestedVersion+"/index.json", distribution.MaxIndexBytes)
	if err != nil {
		return nil, fmt.Errorf("downloading the release index: %w", err)
	}
	index, err := distribution.ParseDownloadIndex(indexData)
	if err != nil {
		return nil, err
	}
	if index.ProductVersion != requestedVersion {
		return nil, fmt.Errorf("release index does not match the requested version")
	}
	platform := runtime.GOOS + "/" + runtime.GOARCH
	var release *distribution.DownloadRelease
	for i := range index.Releases {
		if index.Releases[i].Platform == platform {
			release = &index.Releases[i]
			break
		}
	}
	if release == nil {
		return nil, fmt.Errorf("no published release for %s %s", requestedVersion, platform)
	}
	if release.RawBinarySHA256 != managerSHA256 {
		return nil, fmt.Errorf("release index does not match the verified manager")
	}
	return release, nil
}

func downloadVerifiedPackage(ctx context.Context, origin string, release *distribution.DownloadRelease) ([]byte, error) {
	packageData, err := distribution.Download(ctx, origin, release.Package, distribution.MaxPackageBytes)
	if err != nil {
		return nil, fmt.Errorf("downloading the package: %w", err)
	}
	if distribution.Digest(packageData) != release.PackageSHA256 {
		return nil, fmt.Errorf("downloaded package does not match its verified checksum")
	}
	return packageData, nil
}

// extractedBootstrapPackage names what bootstrapWithDependencies needs after
// extraction: Root is the source checkout runInstallFlow reads, ArtifactID
// identifies the release for BindRetainedInstaller, and PackageTemp is the
// already-verified package bytes restaged as a plain file for
// RetainVerifiedInstaller (which copies from a path and independently
// reverifies this digest before ever retaining it).
type extractedBootstrapPackage struct {
	Root        string
	ArtifactID  string
	PackageTemp string
}

// extractBootstrapPackage safely extracts packageData into a fresh private
// scratch directory and validates it as requestedVersion's package. The
// returned cleanup func removes that scratch directory; the caller must
// defer it regardless of the returned error. Extraction is disposable:
// BuildPlan reads every needed file's bytes into the plan/release before the
// operator is ever asked to confirm, so nothing downstream needs this
// directory to outlive the call — cancelling before consent leaves no
// persistent change, only this already-removed temporary download.
func extractBootstrapPackage(packageData []byte, requestedVersion string) (extractedBootstrapPackage, func(), error) {
	extractParent, err := os.MkdirTemp("", "hive-bootstrap-extract-")
	if err != nil {
		return extractedBootstrapPackage{}, func() {}, err
	}
	cleanup := func() { os.RemoveAll(extractParent) }
	// Resolved once so every path built under it (Root, and everything
	// runInstallFlow reads under that root) names the same directory
	// management itself resolves internally: target.Safe rejects any
	// symlinked ancestor, and a system temp directory is commonly one on
	// both macOS (/tmp -> /private/tmp) and other platforms. cleanup still
	// targets the original (pre-resolution) path, which names the same
	// directory either way.
	extractParent, err = filepath.EvalSymlinks(extractParent)
	if err != nil {
		return extractedBootstrapPackage{}, cleanup, err
	}
	root, err := distribution.Extract(packageData, extractParent)
	if err != nil {
		return extractedBootstrapPackage{}, cleanup, err
	}
	_, artifactID, err := distribution.ValidateBootstrapPackage(root, requestedVersion, version.Current)
	if err != nil {
		return extractedBootstrapPackage{}, cleanup, err
	}
	packageTemp := filepath.Join(extractParent, "downloaded-package.tar.gz")
	if err := os.WriteFile(packageTemp, packageData, 0600); err != nil {
		return extractedBootstrapPackage{}, cleanup, err
	}
	return extractedBootstrapPackage{Root: root, ArtifactID: artifactID, PackageTemp: packageTemp}, cleanup, nil
}

// bootstrapWithDependencies resolves the exact requested release from
// origin, verifies and safely extracts its package, then hands off to the
// same interactive install flow (runInstallFlow) the offline package uses.
// Only after the operator consents does it retain the verified manager and
// package and bind them into the plan, via dependencies.BindRetainedInstaller
// — download and extraction alone never authorize installation (design.md, archived record quoted in translation from Spanish:
// "Invoking the online entry requests a temporary download ... It does not
// authorize installation").
func bootstrapWithDependencies(args []string, in io.Reader, out io.Writer, interactive bool, dependencies installDependencies) error {
	flags, err := parseBootstrapFlags(args, out)
	if err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if err := verifyRetainedManager(flags.ManagerPath, flags.ManagerSHA256); err != nil {
		return err
	}

	fmt.Fprintf(out, "Temporarily downloading Hive %s's verified package from %s to show the installer. This does not authorize installation.\n", flags.RequestedVersion, flags.Origin)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	release, err := resolveBootstrapRelease(ctx, flags.Origin, flags.RequestedVersion, flags.ManagerSHA256)
	if err != nil {
		return err
	}
	packageData, err := downloadVerifiedPackage(ctx, flags.Origin, release)
	if err != nil {
		return err
	}
	pkg, cleanup, err := extractBootstrapPackage(packageData, flags.RequestedVersion)
	defer cleanup()
	if err != nil {
		return err
	}

	managerFile := distribution.VerifiedFile{Path: flags.ManagerPath, SHA256: flags.ManagerSHA256}
	packageFile := distribution.VerifiedFile{Path: pkg.PackageTemp, SHA256: release.PackageSHA256}
	dependencies.BindRetainedInstaller = func(current management.Options, p management.Plan) (management.Plan, error) {
		if _, err := distribution.RetainVerifiedInstaller(current.StateDir, pkg.ArtifactID, managerFile, packageFile); err != nil {
			return management.Plan{}, err
		}
		return management.BindInstaller(p, pkg.ArtifactID)
	}

	o := flags.Options
	o.Source = pkg.Root
	if flags.Hosts != "" {
		for _, h := range strings.Split(flags.Hosts, ",") {
			o.Hosts = append(o.Hosts, strings.TrimSpace(h))
		}
	}
	return runInstallFlow(o, flags.DryRun, in, out, interactive, dependencies)
}
