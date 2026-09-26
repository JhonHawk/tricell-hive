package management

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/distribution"
)

func TestBindInstallerBindsVerifiedIdentityAndRehashesPlan(t *testing.T) {
	state := t.TempDir()
	manager, packageFile := retainedFixture(t, state)
	artifactID := strings.Repeat("a", 64)
	retained, err := distribution.RetainVerifiedInstaller(state, artifactID, manager, packageFile)
	if err != nil {
		t.Fatal(err)
	}
	plan := Plan{Version: stateVersion, Action: "install", StateDir: state}
	bound, err := BindInstaller(plan, artifactID)
	if err != nil {
		t.Fatal(err)
	}
	if bound.Installer == nil || bound.Installer.ArtifactID != retained.ArtifactID || bound.ID != planID(bound) {
		t.Fatalf("invalid bound plan: %+v", bound)
	}
	if err := ValidateInstallerBinding(bound); err != nil {
		t.Fatal(err)
	}
}

func TestValidateInstallerBindingRejectsChangedRetainedPackage(t *testing.T) {
	state := t.TempDir()
	manager, packageFile := retainedFixture(t, state)
	artifactID := strings.Repeat("b", 64)
	if _, err := distribution.RetainVerifiedInstaller(state, artifactID, manager, packageFile); err != nil {
		t.Fatal(err)
	}
	bound, err := BindInstaller(Plan{Version: stateVersion, Action: "install", StateDir: state}, artifactID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bound.Installer.Package, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateInstallerBinding(bound); err == nil {
		t.Fatal("accepted changed retained package")
	}
}

func TestValidateInstallerBindingAcceptsLegacyPlan(t *testing.T) {
	if err := ValidateInstallerBinding(Plan{Version: stateVersion}); err != nil {
		t.Fatal(err)
	}
}

func boundInstallPlan(t *testing.T, o Options, artifactID string) Plan {
	t.Helper()
	if err := os.MkdirAll(o.StateDir, 0700); err != nil {
		t.Fatal(err)
	}
	manager, packageFile := retainedFixture(t, t.TempDir())
	if _, err := distribution.RetainVerifiedInstaller(o.StateDir, artifactID, manager, packageFile); err != nil {
		t.Fatal(err)
	}
	bound, err := BindInstaller(plan(t, "install", o), artifactID)
	if err != nil {
		t.Fatal(err)
	}
	return bound
}

func TestApplyRejectsChangedRetainedInstallerBeforeWriting(t *testing.T) {
	o := setup(t)
	bound := boundInstallPlan(t, o, strings.Repeat("c", 64))
	if err := os.WriteFile(bound.Installer.Package, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := (Engine{}).Apply(bound); err == nil {
		t.Fatal("applied a plan whose retained installer changed")
	}
	absent(t, bound.Changes[0].Target.Path)
}

func TestRecoverRollsBackCoreEvenIfRetainedInstallerWasRemoved(t *testing.T) {
	o := setup(t)
	bound := boundInstallPlan(t, o, strings.Repeat("d", 64))
	crash := Engine{failpoint: func(s string) error {
		if s == "write:0" {
			return os.ErrDeadlineExceeded
		}
		return nil
	}}
	if _, err := crash.Apply(bound); err == nil {
		t.Fatal("missing crash")
	}
	if err := os.RemoveAll(filepath.Dir(bound.Installer.Package)); err != nil {
		t.Fatal(err)
	}
	// Rolling back Hive's own bytes never needs the retained copy; blocking on it
	// would leave pending.json with no way out after a restore or cleanup.
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatalf("recovery blocked on a missing retained installer: %v", err)
	}
	absent(t, bound.Changes[0].Target.Path)
}

func TestRecoverRemovesTemporaryWritesLeftByAKilledProcess(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	crash := Engine{failpoint: func(s string) error {
		if s == "write:0" {
			return os.ErrDeadlineExceeded
		}
		return nil
	}}
	if _, err := crash.Apply(p); err == nil {
		t.Fatal("missing crash")
	}
	// A SIGKILL skips deferred cleanup, leaving write temporaries behind.
	stateOrphan := filepath.Join(o.StateDir, "transactions", ".hive-write-111")
	targetOrphan := filepath.Join(filepath.Dir(p.Changes[0].Target.Path), ".hive-write-222")
	unrelated := filepath.Join(filepath.Dir(p.Changes[0].Target.Path), "user-notes.txt")
	for _, path := range []string{stateOrphan, targetOrphan, unrelated} {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{stateOrphan, targetOrphan} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("recovery left a temporary write: %s", path)
		}
	}
	if _, err := os.Stat(unrelated); err != nil {
		t.Fatalf("recovery removed an unrelated file: %v", err)
	}
}

func retainedFixture(t *testing.T, state string) (distribution.VerifiedFile, distribution.VerifiedFile) {
	t.Helper()
	managerPath := filepath.Join(state, "source-manager")
	packagePath := filepath.Join(state, "source-package.tar.gz")
	if err := os.WriteFile(managerPath, []byte("manager"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(packagePath, []byte("package"), 0600); err != nil {
		t.Fatal(err)
	}
	return distribution.VerifiedFile{Path: managerPath, SHA256: distribution.Digest([]byte("manager"))}, distribution.VerifiedFile{Path: packagePath, SHA256: distribution.Digest([]byte("package"))}
}
