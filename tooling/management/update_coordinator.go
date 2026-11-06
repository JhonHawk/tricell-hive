package management

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/distribution"
)

func newUpdateID() (string, error) {
	var b [16]byte
	_, err := rand.Read(b[:])
	return hex.EncodeToString(b[:]), err
}
func updateJournalPath(dir string) string { return filepath.Join(dir, "update", "pending.json") }

// PendingUpdate reads the parent without interpreting the core state schema.
func PendingUpdate(dir string) (*UpdateJournal, error) {
	canonical, err := target.Canonical(dir)
	if err != nil {
		return nil, err
	}
	dir = canonical
	s, err := read(updateJournalPath(dir))
	if err != nil {
		return nil, err
	}
	if !s.Exists {
		return nil, nil
	}
	j, err := ReadUpdateJournal(bytes.NewReader(s.Data))
	if err != nil {
		return nil, err
	}
	if j.Plan.StateDir != dir {
		return nil, fmt.Errorf("update parent state path mismatch")
	}
	return &j, nil
}
func checkUpdatePending(dir string) error {
	j, err := PendingUpdate(dir)
	if err != nil {
		return err
	}
	if j != nil {
		return fmt.Errorf("pending release update: recover with %s recover --state-dir %s", j.Plan.Package.TargetExecutable.Path, dir)
	}
	return nil
}
func saveUpdate(j *UpdateJournal, phase UpdatePhase) error {
	j.Phase = phase
	sealed, err := SealUpdateJournal(*j)
	if err != nil {
		return err
	}
	*j = sealed
	return writeJSON(updateJournalPath(j.Plan.StateDir), j)
}
func finishUpdate(j *UpdateJournal) error {
	// Archive the completed operation before removing its pending gate. Reapply
	// recognizes this exact operation and never duplicates the core transaction.
	if err := writeJSON(filepath.Join(j.Plan.StateDir, "update", "operation-"+j.Plan.OperationID+".json"), j); err != nil {
		return err
	}
	return write(updateJournalPath(j.Plan.StateDir), snapshot{})
}
func updateCore(p UpdatePlan) (Plan, error) {
	var core Plan
	if err := p.Validate(); err != nil {
		return core, err
	}
	if err := distribution.VerifyRetainedPackage(p.Package, p.CoreStateVersion); err != nil {
		return core, err
	}
	if err := distribution.DecodeUpdateMessage(bytes.NewReader(p.CorePlan), &core); err != nil {
		return core, err
	}
	if core.ID != planID(core) || core.ID != p.CorePlanID || core.StateDir != p.StateDir || core.ReleaseUpdate == p.Installation || core.PackageInstall != p.Installation || core.UpdatePackage == nil || !reflect.DeepEqual(*core.UpdatePackage, p.Package) || !reflect.DeepEqual(core.Hosts, p.Hosts) || core.Executable != p.Package.TargetExecutable.Path || p.Command.Path != filepath.Join(core.Config.Home, ".local", "bin", "hive") {
		return core, fmt.Errorf("update parent/core association mismatch")
	}
	expectedRelease, err := loadRelease(Options{Source: p.Package.PackageDirectory}, p.StateDir)
	if err != nil {
		return core, err
	}
	if core.Release == nil || !reflect.DeepEqual(*core.Release, expectedRelease) || core.Product == nil || core.Product.Version != p.Package.Artifact.ProductVersion || core.Product.ArtifactID != p.Package.Artifact.ManifestSHA256 {
		return core, fmt.Errorf("core payload is not bound to the retained package")
	}
	if !p.Installation && (core.PiPackage != nil || core.CodexWaitFloor != nil) {
		return core, fmt.Errorf("release update cannot install providers; repair separately with hive install")
	}
	return core, nil
}
func updateLocks(dir string) (func(), error) {
	// One order everywhere: update lock, then core lock. The target executor
	// owns both locally; launchers never hold either while waiting on it.
	u, err := lockNamed(dir, "update-lock")
	if err != nil {
		return nil, err
	}
	c, err := lockNamed(dir, "lock")
	if err != nil {
		u()
		return nil, err
	}
	return func() { c(); u() }, nil
}
func (e Engine) ApplyReleaseUpdate(p UpdatePlan) (UpdatePhase, error) {
	if raw, err := hex.DecodeString(p.CoreTransactionID); err != nil || len(raw) != 16 {
		return "", fmt.Errorf("invalid reserved core transaction ID")
	}
	core, err := updateCore(p)
	if err != nil {
		return "", err
	}
	unlock, err := updateLocks(p.StateDir)
	if err != nil {
		return "", err
	}
	defer unlock()
	if err = checkUpdatePending(p.StateDir); err != nil {
		return "", err
	}
	archived, err := read(filepath.Join(p.StateDir, "update", "operation-"+p.OperationID+".json"))
	if err != nil {
		return "", err
	}
	if archived.Exists {
		done, err := ReadUpdateJournal(bytes.NewReader(archived.Data))
		if err != nil || done.Plan.ID != p.ID {
			return "", fmt.Errorf("operation identity conflict")
		}
		if done.Phase != UpdateActivated {
			return done.Phase, fmt.Errorf("update operation was recovered; prepare a new plan")
		}
		command, err := InspectUpdateCommand(p.Command.Path, p.StateDir)
		if err != nil {
			return done.Phase, err
		}
		if command.Expected.LinkTarget != p.Package.TargetExecutable.Path {
			return done.Phase, fmt.Errorf("completed update command changed; preserved")
		}
		return done.Phase, nil
	}
	command, err := InspectUpdateCommand(p.Command.Path, p.StateDir)
	if err != nil {
		return "", err
	}
	if command != p.Command {
		return "", fmt.Errorf("stale stable command")
	}
	reserved := filepath.Join(p.StateDir, "transactions", p.CoreTransactionID+".json")
	if _, err := os.Lstat(reserved); err == nil {
		return "", fmt.Errorf("reserved core transaction already exists")
	} else if !os.IsNotExist(err) {
		return "", err
	}
	state, err := authoritativePreflight(core)
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(filepath.Join(p.StateDir, "update"), 0700); err != nil {
		return "", err
	}
	j := UpdateJournal{Version: 1, Plan: p, CoreJournalPath: filepath.Join(p.StateDir, "transactions", p.CoreTransactionID+".json")}
	if err = saveUpdate(&j, UpdatePrepared); err != nil {
		return "", err
	}
	if err = e.fail("update-prepared"); err != nil {
		return j.Phase, err
	}
	if err = saveUpdate(&j, UpdateCorePending); err != nil {
		return j.Phase, err
	}
	if err = e.fail("update-core-pending"); err != nil {
		return j.Phase, err
	}
	child := e
	child.nested = true
	child.presetState = state
	child.transactionID = p.CoreTransactionID
	_, err = child.Apply(core)
	if err != nil {
		return j.Phase, err
	}
	// A no-op still needs durable commit evidence reserved under the parent.
	if _, statErr := os.Stat(j.CoreJournalPath); os.IsNotExist(statErr) {
		return j.Phase, fmt.Errorf("update core produced no durable transaction")
	}
	if err = e.fail("update-child-response"); err != nil {
		return j.Phase, err
	}
	return e.completeUpdate(&j)
}
func (e Engine) completeUpdate(j *UpdateJournal) (UpdatePhase, error) {
	child, err := boundUpdateChild(*j)
	if err != nil {
		return j.Phase, err
	}
	if child.Phase != "committed" {
		return j.Phase, fmt.Errorf("update core has no durable commit")
	}

	if err := saveUpdate(j, UpdateCoreCommitted); err != nil {
		return j.Phase, err
	}
	if err := e.fail("update-core-committed"); err != nil {
		return j.Phase, err
	}
	if err := saveUpdate(j, UpdateActivationPending); err != nil {
		return j.Phase, err
	}
	if err := e.fail("update-activation-pending"); err != nil {
		return j.Phase, err
	}
	if err := distribution.VerifyRetainedPackage(j.Plan.Package, j.Plan.CoreStateVersion); err != nil {
		return j.Phase, err
	}
	if err := activateUpdateCommand(j.Plan, e.fail); err != nil {
		return j.Phase, err
	}
	if err := saveUpdate(j, UpdateActivated); err != nil {
		return j.Phase, err
	}
	if err := e.fail("update-activated"); err != nil {
		return j.Phase, err
	}
	return j.Phase, finishUpdate(j)
}
func (e Engine) RecoverReleaseUpdate(dir string) (UpdatePhase, error) {
	if err := target.Safe(dir); err != nil {
		return "", err
	}
	unlock, err := updateLocks(dir)
	if err != nil {
		return "", err
	}
	defer unlock()
	j, err := PendingUpdate(dir)
	if err != nil {
		return "", err
	}
	if j == nil {
		return "", nil
	}
	if _, err = updateCore(j.Plan); err != nil {
		return j.Phase, err
	}
	var child journal
	err = decodeFile(j.CoreJournalPath, &child)
	if err == nil {
		if _, err := boundUpdateChild(*j); err != nil {
			return j.Phase, err
		}
		var marker pending
		markerErr := decodeFile(filepath.Join(dir, "pending.json"), &marker)
		if markerErr != nil && !os.IsNotExist(markerErr) {
			return j.Phase, markerErr
		}
		if markerErr == nil && marker.ID != j.Plan.CoreTransactionID {
			return j.Phase, fmt.Errorf("pending core belongs to another operation; preserved")
		}
		if child.Phase == "committed" {
			if markerErr == nil {
				childEngine := e
				childEngine.nested = true
				if _, err := childEngine.Recover(dir); err != nil {
					return j.Phase, err
				}
			}
			return e.completeUpdate(j)
		}
		// The bound journal is durable before the marker and before any writes.
		// Reconstruct a missing marker, never select an unrelated transaction.
		if os.IsNotExist(markerErr) {
			if err := writeJSON(filepath.Join(dir, "pending.json"), pending{j.Plan.CoreTransactionID}); err != nil {
				return j.Phase, err
			}
		}

	} else if !os.IsNotExist(err) {
		return j.Phase, err
	}
	if j.Phase == UpdateCoreCommitted || j.Phase == UpdateActivationPending || j.Phase == UpdateActivated {
		return j.Phase, fmt.Errorf("committed update lost its durable child evidence")
	}
	if err == nil {
		childEngine := e
		childEngine.nested = true
		if _, err = childEngine.Recover(dir); err != nil {
			return j.Phase, err
		}
	}
	if err = saveUpdate(j, UpdateRecovered); err != nil {
		return j.Phase, err
	}
	return j.Phase, finishUpdate(j)
}

// RouteUpdate delegates to the exact retained executor, before core decoding.
func RouteUpdate(ctxRequest distribution.UpdateRequest, p UpdatePlan) (distribution.UpdateResponse, error) {
	return distribution.InvokeUpdate(context.Background(), p.Package.TargetExecutable, ctxRequest)
}

func boundUpdateChild(parent UpdateJournal) (journal, error) {
	var child journal
	if err := decodeFile(parent.CoreJournalPath, &child); err != nil {
		return child, err
	}
	if child.ID != parent.Plan.CoreTransactionID || child.Integrity != journalHash(child) || child.Plan.ID != parent.Plan.CorePlanID || child.Plan.ID != planID(child.Plan) || child.Version != child.Plan.Version || child.Plan.StateDir != parent.Plan.StateDir {
		return child, fmt.Errorf("invalid bound core journal")
	}
	return child, nil
}
