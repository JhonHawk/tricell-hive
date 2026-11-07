package management

import (
	"fmt"
	"path/filepath"
	"reflect"
)

// OnboardPackageProviders runs the install wizard's explicitly selected optional
// steps only after the package core and its stable command are committed. The
// rebuilt plan must be unchanged; this API cannot initiate a package install.
func (e Engine) OnboardPackageProviders(p Plan, steps []ExternalStep, runner ExternalRunner) (OnboardingResult, error) {
	if len(steps) == 0 {
		return OnboardingResult{Phase: "completed"}, nil
	}
	return e.onboard(p, steps, runner, true)
}
func packageProvidersPreflight(p Plan, state State) error {
	if !p.PackageInstall || p.ReleaseUpdate || p.UpdatePackage == nil {
		return fmt.Errorf("expected completed package installation")
	}
	command, err := InspectUpdateCommand(filepath.Join(p.Config.Home, ".local", "bin", "hive"), p.StateDir)
	if err != nil {
		return err
	}
	if command.Expected.LinkTarget != p.UpdatePackage.TargetExecutable.Path {
		return fmt.Errorf("package command is not activated")
	}
	j, next, _, _, err := prepareTransaction(p, state, "")
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(state, next) || p.PiPackage != nil && p.PiPackage.Action == PackageInstall {
		return fmt.Errorf("package core changed; prepare a new coordinated installation")
	}
	for _, en := range j.Entries {
		if !same(en.Before, en.After) {
			return fmt.Errorf("package core changed; prepare a new coordinated installation")
		}
	}
	if j.CodexLines != nil && !j.CodexLines.Skip && j.CodexLines.Before != j.CodexLines.After {
		return fmt.Errorf("package wait floor changed")
	}
	if j.PiConfig != nil && !j.PiConfig.Skip && j.PiConfig.Before != j.PiConfig.After {
		return fmt.Errorf("package pi subagent config changed")
	}
	for _, en := range j.HostHooks {
		if en.Before != en.After {
			return fmt.Errorf("package hooks changed")
		}
	}
	return nil
}
