package management

import (
	"fmt"
	"reflect"
	"tricell-hive/tooling/distribution"
)

// BindInstaller reads an installer retained after consent and binds its
// authenticated identity to a plan before that plan is journaled. It produces
// a new plan ID because the retained manager and package are durable inputs.
func BindInstaller(plan Plan, artifactID string) (Plan, error) {
	if plan.StateDir == "" {
		return Plan{}, fmt.Errorf("plan state directory is required for retained installer")
	}
	installer, err := distribution.RetainedInstallerAt(plan.StateDir, artifactID)
	if err != nil {
		return Plan{}, err
	}
	if plan.Installer != nil && !reflect.DeepEqual(*plan.Installer, installer) {
		return Plan{}, fmt.Errorf("plan already binds a different retained installer")
	}
	plan.Installer = &installer
	plan.ID = planID(plan)
	return plan, nil
}

// ValidateInstallerBinding re-reads a retained installer before a plan writes
// state or launches a recovery path. Legacy plans with no installer remain
// valid because they predate online bootstrap retention.
func ValidateInstallerBinding(plan Plan) error {
	if plan.Installer == nil {
		return nil
	}
	installer, err := distribution.RetainedInstallerAt(plan.StateDir, plan.Installer.ArtifactID)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(*plan.Installer, installer) {
		return fmt.Errorf("retained installer no longer matches the plan")
	}
	return nil
}
