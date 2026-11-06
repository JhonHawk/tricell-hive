package management

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/distribution"
)

type CommandReceipt struct {
	Path             string      `json:"path"`
	Fingerprint      Fingerprint `json:"fingerprint"`
	PackageDirectory string      `json:"package_directory"`
}

func commandReceiptPath(dir string) string { return filepath.Join(dir, "update-command.json") }
func commandFingerprint(path string) (Fingerprint, error) {
	if !updateAbsolute(path) {
		return Fingerprint{}, fmt.Errorf("invalid stable command path")
	}
	if err := target.Safe(filepath.Dir(path)); err != nil {
		return Fingerprint{}, err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return Fingerprint{}, nil
	}
	if err != nil {
		return Fingerprint{}, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return Fingerprint{}, fmt.Errorf("stable command is not a managed link: %s", path)
	}
	link, err := os.Readlink(path)
	if err != nil {
		return Fingerprint{}, err
	}
	return Fingerprint{Exists: true, Kind: "symlink", LinkTarget: link}, nil
}

// InspectUpdateCommand refuses adoption of any existing unowned command.
func InspectUpdateCommand(path, dir string) (UpdateCommand, error) {
	fp, err := commandFingerprint(path)
	if err != nil {
		return UpdateCommand{}, err
	}
	if fp.Exists {
		var receipt CommandReceipt
		if err = decodeFile(commandReceiptPath(dir), &receipt); err != nil || receipt.Path != path || receipt.Fingerprint != fp {
			return UpdateCommand{}, fmt.Errorf("stable command ownership conflict: %s; use the new package executable directly", path)
		}
	}
	return UpdateCommand{Path: path, Expected: fp}, nil
}
func activateUpdateCommand(p UpdatePlan, fail func(string) error) error {
	want := Fingerprint{Exists: true, Kind: "symlink", LinkTarget: p.Package.TargetExecutable.Path}
	current, err := commandFingerprint(p.Command.Path)
	if err != nil {
		return err
	}
	if current != p.Command.Expected && current != want {
		return fmt.Errorf("stable command changed; preserved: %s", p.Command.Path)
	}
	if current != want {
		if err = os.MkdirAll(filepath.Dir(p.Command.Path), 0700); err != nil {
			return err
		}
		temp := filepath.Join(filepath.Dir(p.Command.Path), ".hive-link-"+p.CoreTransactionID)
		staged, err := commandFingerprint(temp)
		if err != nil {
			return err
		}
		if staged.Exists && staged != want {
			return fmt.Errorf("command staging path changed; preserved")
		}
		if !staged.Exists {
			if err = os.Symlink(want.LinkTarget, temp); err != nil {
				return err
			}
		}
		defer os.Remove(temp)
		if err = fail("command-before-rename"); err != nil {
			return err
		}
		check, err := commandFingerprint(p.Command.Path)
		if err != nil {
			return err
		}
		if check != p.Command.Expected {
			return fmt.Errorf("stable command changed; preserved")
		}
		if err = os.Rename(temp, p.Command.Path); err != nil {
			return err
		}
		d, err := os.Open(filepath.Dir(p.Command.Path))
		if err != nil {
			return err
		}
		err = syncFile(d)
		d.Close()
		if err != nil {
			return err
		}
		if err = fail("command-renamed"); err != nil {
			return err
		}
	}
	staging := filepath.Join(filepath.Dir(p.Command.Path), ".hive-link-"+p.CoreTransactionID)
	if fp, err := commandFingerprint(staging); err != nil {
		return err
	} else if fp.Exists {
		if fp != want {
			return fmt.Errorf("command staging path changed; preserved")
		}
		if err = os.Remove(staging); err != nil {
			return err
		}
	}
	receipt := CommandReceipt{Path: p.Command.Path, Fingerprint: want, PackageDirectory: p.Package.PackageDirectory}
	return writeJSON(commandReceiptPath(p.StateDir), receipt)
}

// ActivateManagedCommand can finish activation only from an exact durable core
// journal. New installs should prepare the parent before core mutation with
// PreparePackageInstallation instead.
func ActivateManagedCommand(command UpdateCommand, binding distribution.UpdatePackageBinding, stateDir, coreTransactionID string) error {
	if err := distribution.VerifyRetainedPackage(binding, stateVersion); err != nil {
		return err
	}
	unlock, err := updateLocks(stateDir)
	if err != nil {
		return err
	}
	defer unlock()
	if err = checkUpdatePending(stateDir); err != nil {
		return err
	}
	if raw, err := hex.DecodeString(coreTransactionID); err != nil || len(raw) != 16 {
		return fmt.Errorf("invalid core transaction association")
	}
	var child journal
	if err = decodeFile(filepath.Join(stateDir, "transactions", coreTransactionID+".json"), &child); err != nil {
		return err
	}
	if child.ID != coreTransactionID || child.Phase != "committed" || child.Integrity != journalHash(child) || child.Plan.Executable != binding.TargetExecutable.Path || child.Plan.Product == nil || child.Plan.Product.ArtifactID != binding.Artifact.ManifestSHA256 || command.Path != filepath.Join(child.Plan.Config.Home, ".local", "bin", "hive") {
		return fmt.Errorf("command activation requires matching committed core")
	}
	p := UpdatePlan{Package: binding, Command: command, StateDir: stateDir, CoreTransactionID: coreTransactionID}
	return activateUpdateCommand(p, func(string) error { return nil })
}

// RemoveManagedCommand retires only the link whose exact fingerprint is owned.
// A user replacement survives; its stale ownership receipt is retired too.
func RemoveManagedCommand(stateDir string) error {
	unlock, err := updateLocks(stateDir)
	if err != nil {
		return err
	}
	defer unlock()
	if err = checkUpdatePending(stateDir); err != nil {
		return err
	}
	return removeManagedCommandLocked(stateDir)
}
func removeManagedCommandLocked(stateDir string) error {
	return removeManagedCommandWithFail(stateDir, func(string) error { return nil })
}
func removeManagedCommandWithFail(stateDir string, fail func(string) error) error {
	var receipt CommandReceipt
	if err := decodeFile(commandReceiptPath(stateDir), &receipt); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if !updateAbsolute(receipt.Path) || !receipt.Fingerprint.Exists || receipt.Fingerprint.Kind != "symlink" || receipt.Fingerprint.LinkTarget != filepath.Join(receipt.PackageDirectory, "bin", "hive") {
		return fmt.Errorf("invalid command ownership receipt")
	}
	var err error
	current, inspectErr := commandFingerprint(receipt.Path)
	if inspectErr == nil && !current.Exists {
		return write(commandReceiptPath(stateDir), snapshot{})
	}
	if inspectErr != nil || current != receipt.Fingerprint {
		if err = write(commandReceiptPath(stateDir), snapshot{}); err != nil {
			return err
		}
		return fmt.Errorf("stable command changed; preserved: %s", receipt.Path)
	}
	if err = os.Remove(receipt.Path); err != nil {
		return err
	}
	if err = fail("command-removed"); err != nil {
		return err
	}
	return write(commandReceiptPath(stateDir), snapshot{})
}

func finalUserRemoval(p Plan, state State) bool {
	if p.Action != "remove" || p.Config.Scope != "user" {
		return false
	}
	for _, r := range state.Records {
		for _, consumer := range r.Consumers {
			if consumer.Scope == "user" {
				return false
			}
		}
	}
	return true
}
