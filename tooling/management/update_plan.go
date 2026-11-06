package management

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"tricell-hive/integrations/pi"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/distribution"
)

func releasePiPreflight(p Plan) error {
	if !hostsIncludePi(p.Hosts) {
		return nil
	}
	d, err := readSubagents(p.Config.PiHome)
	if err != nil || d.Status != pi.Present {
		return fmt.Errorf("release update requires a loadable pi-subagents declaration; repair separately with hive install")
	}
	return nil
}

func checkPlanExecutable(p Plan) error {
	if !p.ReleaseUpdate && !p.PackageInstall {
		return checkHookExecutable(p.Executable)
	}
	if p.UpdatePackage == nil || p.Executable != p.UpdatePackage.TargetExecutable.Path {
		return fmt.Errorf("release hook executable is not manifest bound")
	}
	if p.previewUpdate {
		return distribution.VerifyPreparedPackage(*p.UpdatePackage, stateVersion)
	}
	return distribution.VerifyRetainedPackage(*p.UpdatePackage, stateVersion)
}

// PrepareReleaseUpdate renders exclusively for the registered user cohort. The
// verified temporary executor is never used as a hook destination.
func PrepareReleaseUpdate(q UpdatePreparePayload, operationID string) (UpdatePlan, error) {
	var out UpdatePlan
	if err := q.Package.Validate(stateVersion); err != nil {
		return out, err
	}
	if err := distribution.VerifyPreparedPackage(q.Package, stateVersion); err != nil {
		return out, err
	}
	o := Options{Scope: "user", Home: q.Home, StateDir: q.StateDir, Source: q.Package.PreparationDirectory, updatePackage: &q.Package}
	if q.NativeHome {
		o.Home = ""
	}
	cfg, dir, err := normalize(o)
	if err != nil {
		return out, err
	}
	if cfg.Home != q.Home {
		return out, fmt.Errorf("update preparation home does not match executor environment")
	}
	if err = checkUpdatePending(dir); err != nil {
		return out, err
	}
	state, _, err := readState(dir)
	if err != nil {
		return out, err
	}
	hosts := map[string]bool{}
	for _, r := range state.Records {
		for _, c := range r.Consumers {
			if c.Scope == "user" {
				hosts[c.Host] = true
			}
		}
	}
	for h := range hosts {
		o.Hosts = append(o.Hosts, h)
	}
	sort.Strings(o.Hosts)
	if len(o.Hosts) == 0 {
		return out, fmt.Errorf("no registered user hosts; run hive install first")
	}
	core, err := BuildPlan("install", o)
	if err != nil {
		return out, err
	}
	command, err := InspectUpdateCommand(filepath.Join(q.Home, ".local", "bin", "hive"), dir)
	if err != nil {
		return out, err
	}
	id, err := newUpdateID()
	if err != nil {
		return out, err
	}
	data, err := json.Marshal(core)
	if err != nil {
		return out, err
	}
	out = UpdatePlan{Kind: UpdatePlanKind, Version: 1, Protocol: 1, OperationID: operationID, Package: q.Package, StateDir: dir, CoreStateVersion: stateVersion, CorePlan: data, CorePlanID: core.ID, CoreTransactionID: id, Hosts: core.Hosts, Command: command}
	for _, ch := range core.Changes {
		out.Destinations = append(out.Destinations, UpdateDestination{Path: ch.Target.Path, Expected: ch.Expected})
	}
	return FreezeUpdatePlan(out)
}

func SaveUpdatePlan(path string, p UpdatePlan) error {
	if p.Installation {
		return fmt.Errorf("package installation must be confirmed through hive install")
	}
	var privacy struct {
		Legacy []json.RawMessage `json:"legacy"`
	}
	if err := json.Unmarshal(p.CorePlan, &privacy); err != nil {
		return err
	}
	if len(privacy.Legacy) > 0 {
		return fmt.Errorf("legacy migration plans contain private configuration; use hive install")
	}
	if err := p.Validate(); err != nil {
		return err
	}
	if err := target.Safe(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return writeJSON(path, p)
}

// BuildPackageInstallPlan retains normal install provider semantics, with a
// manifest-bound executable instead of a generic hook-path override.
func BuildPackageInstallPlan(o Options, b distribution.UpdatePackageBinding) (Plan, error) {
	if err := distribution.VerifyRetainedPackage(b, stateVersion); err != nil {
		return Plan{}, err
	}
	o.Source = b.PackageDirectory
	o.packageInstall = true
	o.updatePackage = &b
	return BuildPlan("install", o)
}

func PreparePackageInstallation(core Plan, b distribution.UpdatePackageBinding, command UpdateCommand, operationID string) (UpdatePlan, error) {
	if !core.PackageInstall || core.ReleaseUpdate || core.UpdatePackage == nil {
		return UpdatePlan{}, fmt.Errorf("expected a manifest-bound package installation plan")
	}
	id, err := newUpdateID()
	if err != nil {
		return UpdatePlan{}, err
	}
	data, err := json.Marshal(core)
	if err != nil {
		return UpdatePlan{}, err
	}
	p := UpdatePlan{Installation: true, Kind: UpdatePlanKind, Version: 1, Protocol: 1, OperationID: operationID, Package: b, StateDir: core.StateDir, CoreStateVersion: stateVersion, CorePlan: data, CorePlanID: core.ID, CoreTransactionID: id, Hosts: core.Hosts, Command: command}
	return FreezeUpdatePlan(p)
}

// BuildPackageInstallPreview verifies the disposable package but renders hooks
// for its final immutable destination, preserving install provider choices.
func BuildPackageInstallPreview(o Options, b distribution.UpdatePackageBinding) (Plan, error) {
	if err := distribution.VerifyPreparedPackage(b, stateVersion); err != nil {
		return Plan{}, err
	}
	o.Source = b.PreparationDirectory
	o.packageInstall = true
	o.packagePreview = true
	o.updatePackage = &b
	return BuildPlan("install", o)
}
