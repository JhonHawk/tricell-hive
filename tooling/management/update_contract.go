package management

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"tricell-hive/tooling/distribution"
)

const (
	UpdatePlanKind      = "release_update"
	UpdateParentVersion = 1
)

// UpdateStateVersions is the core-schema range this executor can read. It is
// recorded in package bootstrap metadata; it is not the parent format version.
func UpdateStateVersions() (minimum, maximum int) { return 1, stateVersion }

type UpdateCommand struct {
	Path     string      `json:"path"`
	Expected Fingerprint `json:"expected"`
}

// UpdatePlan is independent of Plan and State: older launchers can recognize
// and route it before attempting to decode a future core schema. CorePlan is
// frozen JSON interpreted only by the verified target executor. Its hash binds
// the exact bytes; CorePlanID and CoreTransactionID bind durable child evidence
// reserved before the first mutation. Neither a process exit nor this envelope
// alone proves the child's commit.
type UpdatePlan struct {
	Installation      bool                              `json:"installation,omitempty"`
	ID                string                            `json:"id"`
	Kind              string                            `json:"kind"`
	Version           int                               `json:"version"`
	Protocol          int                               `json:"protocol"`
	OperationID       string                            `json:"operation_id"`
	Package           distribution.UpdatePackageBinding `json:"package"`
	StateDir          string                            `json:"state_dir"`
	CoreStateVersion  int                               `json:"core_state_version"`
	CorePlan          json.RawMessage                   `json:"core_plan"`
	CorePlanSHA256    string                            `json:"core_plan_sha256"`
	CorePlanID        string                            `json:"core_plan_id"`
	CoreTransactionID string                            `json:"core_transaction_id"`
	Hosts             []string                          `json:"hosts"`
	Command           UpdateCommand                     `json:"command"`
	Destinations      []UpdateDestination               `json:"destinations"`
}

type UpdateDestination struct {
	Path     string      `json:"path"`
	Expected Fingerprint `json:"expected"`
}

var updateIdentifier = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)

func updateAbsolute(path string) bool { return filepath.IsAbs(path) && filepath.Clean(path) == path }

func (p UpdatePlan) Validate() error {
	if p.ID != updatePlanDigest(p) {
		return fmt.Errorf("release update plan integrity mismatch")
	}
	if p.Kind != UpdatePlanKind || p.Version != UpdateParentVersion || p.Protocol != distribution.UpdateProtocolVersion {
		return fmt.Errorf("unsupported release update plan; use its retained target executor")
	}
	if !updateIdentifier.MatchString(p.OperationID) || !updateIdentifier.MatchString(p.CoreTransactionID) || !distribution.ValidSHA256(p.CorePlanID) || !json.Valid(p.CorePlan) || distribution.Digest(p.CorePlan) != p.CorePlanSHA256 {
		return fmt.Errorf("invalid frozen update plan or child association")
	}
	if err := p.Package.Validate(p.CoreStateVersion); err != nil {
		return err
	}
	if !updateAbsolute(p.StateDir) || !updateAbsolute(p.Command.Path) || len(p.Hosts) == 0 {
		return fmt.Errorf("invalid update destinations or host cohort")
	}
	seen := map[string]bool{}
	for _, host := range p.Hosts {
		switch host {
		case "claude", "codex", "grok", "pi", "opencode", "cursor":
		default:
			return fmt.Errorf("unknown update host")
		}
		if seen[host] {
			return fmt.Errorf("repeated update host")
		}
		seen[host] = true
	}
	paths := map[string]bool{}
	for _, d := range p.Destinations {
		if !updateAbsolute(d.Path) || paths[d.Path] {
			return fmt.Errorf("invalid or repeated update destination")
		}
		paths[d.Path] = true
	}
	return nil
}

func compactUpdateCore(data json.RawMessage) (json.RawMessage, error) {
	// Marshal the opaque RawMessage instead of decoding a schema. Besides
	// whitespace it normalizes HTML escaping, which Marshal also applies when
	// this payload is embedded in a saved parent envelope.
	return json.Marshal(data)
}
func updatePlanDigest(p UpdatePlan) string {
	p.ID = ""
	data, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return distribution.Digest(data)
}

// FreezeUpdatePlan canonicalizes opaque JSON whitespace before hashing, so a
// saved indented envelope retains the same identity without a core decoder.
func FreezeUpdatePlan(p UpdatePlan) (UpdatePlan, error) {
	var err error
	p.CorePlan, err = compactUpdateCore(p.CorePlan)
	if err != nil {
		return p, err
	}
	p.CorePlanSHA256 = distribution.Digest(p.CorePlan)
	p.ID = updatePlanDigest(p)
	return p, p.Validate()
}

// IsUpdatePlan only dispatches; callers must then validate with ReadUpdatePlan.
// It deliberately does not decode the core plan or check a core schema version.
func IsUpdatePlan(data []byte) bool {
	var tag struct {
		Kind string `json:"kind"`
	}
	return len(data) <= distribution.MaxUpdateMessageBytes && json.Unmarshal(data, &tag) == nil && tag.Kind == UpdatePlanKind
}
func ReadUpdatePlan(r io.Reader) (UpdatePlan, error) {
	var p UpdatePlan
	if err := distribution.DecodeUpdateMessage(r, &p); err != nil {
		return p, err
	}
	var err error
	p.CorePlan, err = compactUpdateCore(p.CorePlan)
	if err != nil {
		return p, err
	}
	return p, p.Validate()
}

type UpdatePhase string

const (
	UpdatePrepared          UpdatePhase = "prepared"
	UpdateCorePending       UpdatePhase = "core_pending"
	UpdateCoreCommitted     UpdatePhase = "core_committed"
	UpdateActivationPending UpdatePhase = "activation_pending"
	UpdateActivated         UpdatePhase = "activated"
	UpdateRecovered         UpdatePhase = "recovered"
)

// UpdateJournal is the durable parent record, written before core mutation.
// Lock order is update then core; a launcher must release locks before waiting
// on a child that needs them. Pending records pin both the retained package and
// CoreJournalPath. T3 supplies persistence, lock enforcement and commit checks.
type UpdateJournal struct {
	Version         int         `json:"version"`
	Plan            UpdatePlan  `json:"plan"`
	Phase           UpdatePhase `json:"phase"`
	CoreJournalPath string      `json:"core_journal_path"`
	Integrity       string      `json:"integrity"`
}

func (j UpdateJournal) Validate() error {
	if j.Integrity != updateJournalDigest(j) {
		return fmt.Errorf("release update journal integrity mismatch")
	}
	if j.Version != UpdateParentVersion {
		return fmt.Errorf("unsupported update journal")
	}
	if err := j.Plan.Validate(); err != nil {
		return err
	}
	if j.CoreJournalPath != filepath.Join(j.Plan.StateDir, "transactions", j.Plan.CoreTransactionID+".json") {
		return fmt.Errorf("update child journal association mismatch")
	}
	switch j.Phase {
	case UpdatePrepared, UpdateCorePending, UpdateCoreCommitted, UpdateActivationPending, UpdateActivated, UpdateRecovered:
		return nil
	default:
		return fmt.Errorf("unknown update journal phase")
	}
}

func updateJournalDigest(j UpdateJournal) string {
	j.Integrity = ""
	data, err := json.Marshal(j)
	if err != nil {
		return ""
	}
	return distribution.Digest(data)
}
func SealUpdateJournal(j UpdateJournal) (UpdateJournal, error) {
	j.Integrity = updateJournalDigest(j)
	return j, j.Validate()
}
func ReadUpdateJournal(r io.Reader) (UpdateJournal, error) {
	var j UpdateJournal
	if err := distribution.DecodeUpdateMessage(r, &j); err != nil {
		return j, err
	}
	var err error
	j.Plan.CorePlan, err = compactUpdateCore(j.Plan.CorePlan)
	if err != nil {
		return j, err
	}
	return j, j.Validate()
}

// Action payloads share the fixed envelope and remain opaque to old launchers.
// Prepare includes the desired final destination even when it does not exist.
type UpdatePreparePayload struct {
	NativeHome bool                              `json:"native_home,omitempty"`
	Package    distribution.UpdatePackageBinding `json:"package"`
	Home       string                            `json:"home"`
	StateDir   string                            `json:"state_dir"`
}
type UpdateApplyPayload struct {
	Plan UpdatePlan `json:"plan"`
}
type UpdateRecoverPayload struct {
	Journal UpdateJournal `json:"journal"`
}
type UpdateResultPayload struct {
	Plan              *UpdatePlan `json:"plan,omitempty"`
	CoreTransactionID string      `json:"core_transaction_id,omitempty"`
	CoreJournalPath   string      `json:"core_journal_path,omitempty"`
}
