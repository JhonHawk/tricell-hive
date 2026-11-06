package distribution

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"tricell-hive/tooling/version"
)

const (
	UpdateProtocolVersion = 1
	UpdateBootstrapName   = "update-bootstrap.json"
	UpdateInternalCommand = "__hive-update-v1"
	MaxUpdateMessageBytes = 32 << 20
	UpdateProcessTimeout  = 2 * time.Minute
)

// UpdateBootstrap is deliberately independent of the full manifest and core
// state schemas. Its file must be included in the manifest inventory. The
// launcher authenticates the archive and these identities before execution;
// the target executor then verifies the complete package.
type UpdateBootstrap struct {
	Protocol         int    `json:"protocol"`
	ProductVersion   string `json:"product_version"`
	Platform         string `json:"platform"`
	SourceID         string `json:"source_id"`
	ExecutableSHA256 string `json:"executable_sha256"`
	ManifestVersion  int    `json:"manifest_version"`
	MinStateVersion  int    `json:"min_state_version"`
	MaxStateVersion  int    `json:"max_state_version"`
}

func ValidSHA256(s string) bool {
	b, err := hex.DecodeString(s)
	return err == nil && len(b) == 32 && strings.ToLower(s) == s
}

func (b UpdateBootstrap) Validate(product, platform string, stateVersion int) error {
	if b.Protocol != UpdateProtocolVersion {
		return fmt.Errorf("unsupported update protocol; install this version manually")
	}
	if !version.Valid(b.ProductVersion) || b.ProductVersion == "dev" || b.ProductVersion != product || b.Platform != platform || !ValidSHA256(b.SourceID) || !ValidSHA256(b.ExecutableSHA256) {
		return fmt.Errorf("update package identity mismatch")
	}
	if b.ManifestVersion != 1 && b.ManifestVersion != 2 {
		return fmt.Errorf("unsupported update manifest; install this version manually")
	}
	if b.MinStateVersion < 1 || b.MaxStateVersion < b.MinStateVersion || stateVersion < b.MinStateVersion || stateVersion > b.MaxStateVersion {
		return fmt.Errorf("unsupported update state schema; install this version manually")
	}
	return nil
}

// ReadUpdateBootstrap authenticates the bootstrap against an already verified
// manifest. It does not replace archive, platform or full inventory validation.
func ReadUpdateBootstrap(root string, manifest Manifest, stateVersion int) (UpdateBootstrap, error) {
	var b UpdateBootstrap
	if manifest.UpdateProtocol != UpdateProtocolVersion {
		return b, fmt.Errorf("unsupported update protocol; install this version manually")
	}
	path := filepath.Join(root, UpdateBootstrapName)
	info, err := os.Lstat(path)
	if err != nil {
		return b, err
	}
	if !info.Mode().IsRegular() || info.Size() > 4096 {
		return b, fmt.Errorf("invalid update bootstrap file")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return b, err
	}
	if manifest.Files[UpdateBootstrapName] != Digest(data) {
		return b, fmt.Errorf("update bootstrap digest mismatch")
	}
	if err := DecodeUpdateMessage(bytes.NewReader(data), &b); err != nil {
		return b, err
	}
	if err := b.Validate(manifest.ProductVersion, manifest.Platform, stateVersion); err != nil {
		return b, err
	}
	if b.ManifestVersion != manifest.Version || b.SourceID != manifest.SourceID || b.ExecutableSHA256 != manifest.Files["bin/hive"] {
		return b, fmt.Errorf("update bootstrap does not match manifest")
	}
	return b, nil
}

type ExecutorIdentity struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// UpdateArtifact freezes the remote artifact independently of its content ID
// and exact manifest bytes. No URL or command in this structure is executable.
type UpdateArtifact struct {
	ProductVersion string          `json:"product_version"`
	Platform       string          `json:"platform"`
	Label          string          `json:"label"`
	ArchiveSHA256  string          `json:"archive_sha256"`
	ManifestSHA256 string          `json:"manifest_sha256"`
	SourceID       string          `json:"source_id"`
	Bootstrap      UpdateBootstrap `json:"bootstrap"`
}

// UpdatePackageBinding separates the temporary preparation process from the
// real final hook destination. This is data, not authority to override hooks.
// The verified preparation path and retained apply path must each be checked
// against package bytes by their caller before supplying this binding.
type UpdatePackageBinding struct {
	Artifact             UpdateArtifact   `json:"artifact"`
	PreparationDirectory string           `json:"preparation_directory"`
	PackageDirectory     string           `json:"package_directory"`
	PreparationExecutor  ExecutorIdentity `json:"preparation_executor"`
	TargetExecutable     ExecutorIdentity `json:"target_executable"`
}

func canonicalAbsolute(path string) bool { return filepath.IsAbs(path) && filepath.Clean(path) == path }

func (b UpdatePackageBinding) Validate(stateVersion int) error {
	a := b.Artifact
	if err := a.Bootstrap.Validate(a.ProductVersion, a.Platform, stateVersion); err != nil {
		return err
	}
	if !ValidSHA256(a.ArchiveSHA256) || !ValidSHA256(a.ManifestSHA256) || a.SourceID != a.Bootstrap.SourceID || a.Label == "" || a.Label == "." || strings.ContainsAny(a.Label, "/\\") {
		return fmt.Errorf("invalid frozen update artifact")
	}
	if !canonicalAbsolute(b.PreparationDirectory) || !canonicalAbsolute(b.PackageDirectory) || filepath.Base(b.PackageDirectory) != a.Label {
		return fmt.Errorf("invalid update package directories")
	}
	for _, item := range []struct {
		root     string
		identity ExecutorIdentity
	}{{b.PreparationDirectory, b.PreparationExecutor}, {b.PackageDirectory, b.TargetExecutable}} {
		if item.identity.Path != filepath.Join(item.root, "bin", "hive") || item.identity.SHA256 != a.Bootstrap.ExecutableSHA256 {
			return fmt.Errorf("executable is not bound to update package")
		}
	}
	return nil
}

type UpdateAction string

const (
	UpdatePrepare UpdateAction = "prepare"
	UpdateApply   UpdateAction = "apply"
	UpdateRecover UpdateAction = "recover"
)

type UpdateStatus string

const (
	UpdatePrepared      UpdateStatus = "prepared"
	UpdateCoreCommitted UpdateStatus = "core_committed"
	UpdateActivated     UpdateStatus = "activated"
	UpdateRecovered     UpdateStatus = "recovered"
	UpdateFailed        UpdateStatus = "failed"
)

// Payload remains opaque to the launcher: a future core schema is interpreted
// only by its verified executor. OperationID binds every reply to one parent.
type UpdateRequest struct {
	Protocol       int             `json:"protocol"`
	OperationID    string          `json:"operation_id"`
	Action         UpdateAction    `json:"action"`
	ProductVersion string          `json:"product_version"`
	Payload        json.RawMessage `json:"payload"`
}
type UpdateResponse struct {
	Protocol       int             `json:"protocol"`
	OperationID    string          `json:"operation_id"`
	Action         UpdateAction    `json:"action"`
	ProductVersion string          `json:"product_version"`
	Status         UpdateStatus    `json:"status"`
	Payload        json.RawMessage `json:"payload,omitempty"`
	Error          string          `json:"error,omitempty"`
}

func (q UpdateRequest) Validate() error {
	if q.Protocol != UpdateProtocolVersion || q.OperationID == "" || len(q.OperationID) > 128 || !version.Valid(q.ProductVersion) || q.ProductVersion == "dev" || !json.Valid(q.Payload) {
		return fmt.Errorf("invalid update request")
	}
	switch q.Action {
	case UpdatePrepare, UpdateApply, UpdateRecover:
		return nil
	default:
		return fmt.Errorf("unknown update action")
	}
}

// DecodeUpdateMessage accepts exactly one bounded JSON value. Unknown envelope
// fields require a new protocol; unknown fields inside opaque payloads do not.
func DecodeUpdateMessage(r io.Reader, value any) error {
	data, err := io.ReadAll(io.LimitReader(r, MaxUpdateMessageBytes+1))
	if err != nil {
		return err
	}
	if len(data) > MaxUpdateMessageBytes {
		return fmt.Errorf("update message exceeds limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("invalid update message: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("update message has trailing data")
	}
	return nil
}
func ReadUpdateRequest(r io.Reader) (UpdateRequest, error) {
	var q UpdateRequest
	if err := DecodeUpdateMessage(r, &q); err != nil {
		return q, err
	}
	return q, q.Validate()
}
func WriteUpdateMessage(w io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(data)+1 > MaxUpdateMessageBytes {
		return fmt.Errorf("update message exceeds limit")
	}
	_, err = w.Write(append(data, '\n'))
	return err
}

func (r UpdateResponse) Validate(q UpdateRequest) error {
	if r.Protocol != q.Protocol || r.OperationID != q.OperationID || r.Action != q.Action || r.ProductVersion != q.ProductVersion {
		return fmt.Errorf("update response identity mismatch")
	}
	if r.Status == UpdateFailed {
		if r.Error == "" {
			return fmt.Errorf("invalid update failure")
		}
		return fmt.Errorf("target executor refused update: %s", r.Error)
	}
	if r.Error != "" || !json.Valid(r.Payload) {
		return fmt.Errorf("invalid update response")
	}
	switch q.Action {
	case UpdatePrepare:
		if r.Status == UpdatePrepared {
			return nil
		}
	case UpdateApply:
		if r.Status == UpdateCoreCommitted || r.Status == UpdateActivated {
			return nil
		}
	case UpdateRecover:
		if r.Status == UpdateRecovered || r.Status == UpdateCoreCommitted || r.Status == UpdateActivated {
			return nil
		}
	}
	return fmt.Errorf("unexpected update response status")
}

type boundedOutput struct {
	bytes.Buffer
	limit int
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, fmt.Errorf("update process output exceeds limit")
	}
	return b.Buffer.Write(p)
}

// InvokeUpdate never uses a shell, PATH resolution, or manifest-supplied args.
// A zero exit status is insufficient: the bound structured response is required.
// Durable commit evidence is the parent's responsibility, even on process error.
func InvokeUpdate(ctx context.Context, executor ExecutorIdentity, q UpdateRequest) (UpdateResponse, error) {
	var response UpdateResponse
	if err := q.Validate(); err != nil {
		return response, err
	}
	if !canonicalAbsolute(executor.Path) || !ValidSHA256(executor.SHA256) {
		return response, fmt.Errorf("invalid update executor")
	}
	resolved, err := filepath.EvalSymlinks(executor.Path)
	if err != nil {
		return response, err
	}
	if resolved != executor.Path {
		return response, fmt.Errorf("update executor path contains symlinks")
	}
	info, err := os.Lstat(executor.Path)
	if err != nil {
		return response, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		return response, fmt.Errorf("update executor is not executable regular file")
	}
	f, err := os.Open(executor.Path)
	if err != nil {
		return response, err
	}
	data, readErr := io.ReadAll(io.LimitReader(f, MaxExtractedBytes+1))
	closeErr := f.Close()
	if readErr != nil {
		return response, readErr
	}
	if closeErr != nil {
		return response, closeErr
	}
	if int64(len(data)) > MaxExtractedBytes || Digest(data) != executor.SHA256 {
		return response, fmt.Errorf("update executor digest mismatch")
	}
	var input bytes.Buffer
	if err := WriteUpdateMessage(&input, q); err != nil {
		return response, err
	}
	ctx, cancel := context.WithTimeout(ctx, UpdateProcessTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, executor.Path, UpdateInternalCommand)
	cmd.Stdin = &input
	stdout := &boundedOutput{limit: MaxUpdateMessageBytes}
	stderr := &boundedOutput{limit: 64 << 10}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.WaitDelay = time.Second
	if err := cmd.Run(); err != nil {
		return response, fmt.Errorf("update executor did not complete: %w", err)
	}
	if err := DecodeUpdateMessage(&stdout.Buffer, &response); err != nil {
		return response, err
	}
	return response, response.Validate(q)
}
