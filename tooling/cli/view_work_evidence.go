package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"tricell-hive/tooling/viewer"
)

// `work evidence` publishes one validation view: it uploads the captures,
// registers the candidate when needed and appends the validation (or, with
// --supersedes, the iteration) event with its view. Each step is idempotent
// against the viewer's state, so repeating a command that was interrupted reuses
// what already arrived and repeating a finished one changes nothing.

// The limit keeps one upload, base64 encoded in the journal, inside the room
// the journal reserves for an intent.
const viewWorkEvidenceImageLimit = 8 << 20

var (
	viewWorkDefaultStep     = "Open the captured screen"
	viewWorkDefaultExpected = "The screen matches the captures"
)

// viewWorkEvidenceImage is one --image FILE[:caption].
type viewWorkEvidenceImage struct {
	Path, Caption, Media, SHA string
	Data                      []byte
}

type viewWorkEvidenceResult struct {
	WorkID         string   `json:"work_id"`
	Result         string   `json:"result"`
	SessionVersion int64    `json:"session_version"`
	CandidateID    string   `json:"candidate_id,omitempty"`
	EventID        string   `json:"event_id,omitempty"`
	EvidenceID     string   `json:"evidence_id,omitempty"`
	ViewIDs        []string `json:"view_ids,omitempty"`
	ArtifactIDs    []string `json:"artifact_ids,omitempty"`
	// ArtifactsOmitted counts the capture IDs that do not fit the bounded reply.
	ArtifactsOmitted int                  `json:"artifacts_omitted,omitempty"`
	Queue            *viewWorkQueueReport `json:"queue,omitempty"`
}

// capViewWorkEvidenceResult lists the artifact IDs that fit the bounded
// mutation output and counts the rest, so a published evidence never fails on
// its own reply.
func capViewWorkEvidenceResult(r *viewWorkEvidenceResult) {
	for len(r.ArtifactIDs) > 0 && !fitsViewWorkOutput(*r) {
		r.ArtifactIDs = r.ArtifactIDs[:len(r.ArtifactIDs)-1]
		r.ArtifactsOmitted++
	}
}

// splitViewWorkImage separates FILE from :caption. A path that exists whole is
// a file; otherwise the text after the last colon is the caption.
func splitViewWorkImage(spec string) (path, caption string) {
	if st, err := os.Stat(spec); err == nil && st.Mode().IsRegular() {
		return spec, ""
	}
	if n := strings.LastIndex(spec, ":"); n > 0 {
		return spec[:n], spec[n+1:]
	}
	return spec, ""
}

// validateViewWorkEvidence checks everything about the command that does not
// need the viewer or the files, with the validators the viewer applies.
func validateViewWorkEvidence(c *viewWorkCommand) error {
	if len(c.Images) == 0 || len(c.Images) > viewer.TimelineMaxImageRefs {
		return workInput("at least one --image is required")
	}
	for _, spec := range c.Images {
		if path, _ := splitViewWorkImage(spec); path == "" {
			return workInput("invalid --image")
		}
	}
	if c.Viewport == "" {
		c.Viewport = "desktop"
	}
	if c.Theme == "" {
		c.Theme = "unspecified"
	}
	if c.Reason == "" {
		c.Reason = "Publish validation evidence"
	}
	if len(c.Steps) == 0 {
		c.Steps = []string{viewWorkDefaultStep}
	}
	if c.Expected == "" {
		c.Expected = viewWorkDefaultExpected
	}
	if c.Supersedes != "" && !viewTimelineObjectID.MatchString(c.Supersedes) {
		return workInput("invalid supersedes view")
	}
	if c.Candidate != "" && len(c.Candidate) > 256 || len(c.Branch) > 256 || len(c.Base) > 256 {
		return workInput("invalid candidate")
	}
	artifact := viewer.TimelineArtifactMetadata{Actor: viewer.TimelineActor{Kind: "native_producer", ProducerID: "validate"}, SchemaVersion: 3, OperationID: "validate", CapturedAt: "2026-01-01T00:00:00Z", CandidateID: "validate", Viewport: c.Viewport, Theme: c.Theme, MediaType: "image/png"}
	if artifact.Validate() != nil {
		return workInput("invalid viewport or theme")
	}
	refs := make([]viewer.TimelineImageRef, len(c.Images))
	for n := range refs {
		refs[n] = viewer.TimelineImageRef{ArtifactID: fmt.Sprintf("validate%d", n), Caption: c.ViewTitle}
	}
	request := viewWorkEvidenceEvent(*c, "validate", viewer.TimelineActor{Kind: "native_producer", ProducerID: "validate"}, refs, nil)
	if request.Validate() != nil {
		return workInput("invalid view")
	}
	return nil
}

func viewWorkEvidenceEvent(c viewWorkCommand, candidateID string, actor viewer.TimelineActor, refs []viewer.TimelineImageRef, supersedes *viewWorkSupersedes) viewer.TimelineAppendEventRequest {
	view := viewer.TimelineViewInput{Title: c.ViewTitle, TicketIDs: []string{}, Steps: c.Steps, ExpectedResult: c.Expected, ImageRefs: refs}
	kind := viewer.TimelineEventKind("validation")
	related := ""
	if supersedes != nil {
		kind, related = "iteration", supersedes.EventID
		view.ViewID = supersedes.ViewID
		view.Supersedes = &viewer.TimelineSupersedes{EventID: supersedes.EventID, EvidenceID: supersedes.EvidenceID, ViewID: supersedes.ViewID}
	}
	return viewer.TimelineAppendEventRequest{SchemaVersion: 3, OperationID: workID("op-"), Actor: actor, Reason: c.Reason, Kind: kind, Title: c.ViewTitle, Summary: c.Reason, TicketIDs: []string{}, CandidateID: candidateID, RelatedEventID: related, Views: []viewer.TimelineViewInput{view}}
}

type viewWorkSupersedes struct{ EventID, EvidenceID, ViewID string }

// evidenceStep sends one request of the evidence as a journal intent: its exact
// bytes are recorded before they leave the process, so an interrupted step is
// replayed unchanged by the next evidence command or by `work resume`.
func (e *viewWorkEngine) evidenceStep(c viewWorkCommand, source viewer.TimelineSource, step, method, path, contentType string, body []byte) (viewer.TimelineOperationResult, error) {
	next := "run the same command again: it reuses what already arrived"
	context := viewWorkLifecycleContext{Command: viewWorkCommand{Action: "evidence", Host: c.Host, NativeSessionID: c.NativeSessionID, WorkID: c.WorkID}}
	i, err := e.prepareIntent("evidence", c.WorkID, source, context)
	if err != nil {
		return viewer.TimelineOperationResult{}, err
	}
	if _, err = e.appendStep(i, method, path, contentType, body); err != nil {
		return viewer.TimelineOperationResult{}, err
	}
	raw, err := e.executeStep(i, 0)
	if err != nil {
		var failure *viewWorkError
		if errors.As(err, &failure) {
			named := *failure
			named.Field = step
			if named.Next == "" {
				named.Next = next
			}
			return viewer.TimelineOperationResult{}, &named
		}
		return viewer.TimelineOperationResult{}, err
	}
	r, err := viewer.DecodeTimelineOperationResult(raw)
	if err != nil {
		return r, &viewWorkError{Code: "outcome_unknown", Field: step, Next: next}
	}
	return r, e.finishIntent(i)
}

func readViewWorkImages(c viewWorkCommand) ([]viewWorkEvidenceImage, error) {
	images := make([]viewWorkEvidenceImage, 0, len(c.Images))
	for _, spec := range c.Images {
		path, caption := splitViewWorkImage(spec)
		data, err := viewTimelineReadFile(path, viewWorkEvidenceImageLimit)
		if err != nil {
			return nil, &viewWorkError{Code: "image_unreadable", Field: filepath.Base(path), Next: "pass a regular image file, not a link, up to 20 MiB"}
		}
		media := http.DetectContentType(data)
		if media != "image/png" && media != "image/jpeg" && media != "image/webp" {
			return nil, &viewWorkError{Code: "unsupported_media_type", Field: filepath.Base(path), Next: "capture as PNG, JPEG or WebP"}
		}
		if caption == "" {
			caption = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		}
		sum := sha256.Sum256(data)
		images = append(images, viewWorkEvidenceImage{Path: path, Caption: caption, Media: media, SHA: hex.EncodeToString(sum[:]), Data: data})
	}
	return images, nil
}

func (e *viewWorkEngine) evidence(c viewWorkCommand) (viewWorkEvidenceResult, error) {
	images, err := readViewWorkImages(c)
	if err != nil {
		return viewWorkEvidenceResult{}, err
	}
	s, err := e.session(c.WorkID)
	if err != nil {
		return viewWorkEvidenceResult{}, err
	}
	actor, err := e.actor(s, c)
	if err != nil {
		return viewWorkEvidenceResult{}, err
	}
	if err = e.resumeEvidenceIntents(c, workLifecycleSource(s, actor)); err != nil {
		return viewWorkEvidenceResult{}, err
	}
	if s, err = e.session(c.WorkID); err != nil {
		return viewWorkEvidenceResult{}, err
	}
	if s.State == viewer.TimelineClosed || s.State == viewer.TimelineFailed {
		return viewWorkEvidenceResult{}, workStateError("session_closed")
	}
	if workLifecycleStage(s, c.Stage) == nil {
		return viewWorkEvidenceResult{}, workStateError("stage_not_started")
	}
	var supersedes *viewWorkSupersedes
	if c.Supersedes != "" {
		for _, event := range s.Events {
			if event.Withdrawn || event.EvidenceID == "" {
				continue
			}
			for _, v := range event.Views {
				if v.ViewID == c.Supersedes {
					supersedes = &viewWorkSupersedes{EventID: event.EventID, EvidenceID: event.EvidenceID, ViewID: v.ViewID}
				}
			}
		}
		if supersedes == nil {
			return viewWorkEvidenceResult{}, workStateError("supersedes_view_not_found")
		}
	}
	candidateID, err := e.evidenceCandidate(s, actor, c)
	if err != nil {
		return viewWorkEvidenceResult{}, err
	}
	// The same view already published for this candidate: nothing to do.
	if r, ok := evidencePublished(s, c, candidateID, images, supersedes); ok {
		r.SessionVersion = s.Version
		return r, nil
	}
	basePath := "/api/timeline/sessions/" + s.ViewerSessionID
	if candidateID == "" {
		created, err := e.evidenceCreateCandidate(s, actor, c)
		if err != nil {
			return viewWorkEvidenceResult{}, err
		}
		candidateID = created.CandidateID
		if s, err = e.session(c.WorkID); err != nil {
			return viewWorkEvidenceResult{}, err
		}
	}
	refs := make([]viewer.TimelineImageRef, len(images))
	artifactIDs := make([]string, len(images))
	for n, image := range images {
		id := ""
		for _, a := range s.Artifacts {
			if a.SHA256 == image.SHA && a.CandidateID == candidateID && a.Viewport == c.Viewport && a.Theme == c.Theme && a.MediaType == image.Media {
				id = a.ArtifactID
			}
		}
		if id == "" {
			meta := viewer.TimelineArtifactMetadata{Actor: actor, SchemaVersion: 3, OperationID: workID("op-"), CapturedAt: time.Now().UTC().Format(time.RFC3339), CandidateID: candidateID, Viewport: c.Viewport, Theme: c.Theme, MediaType: image.Media}
			if err = meta.Validate(); err != nil {
				return viewWorkEvidenceResult{}, err
			}
			rawMeta, err := json.Marshal(meta)
			if err != nil {
				return viewWorkEvidenceResult{}, err
			}
			var buffer bytes.Buffer
			writer := multipart.NewWriter(&buffer)
			err = writer.WriteField("metadata", string(rawMeta))
			if err == nil {
				var part io.Writer
				if part, err = writer.CreateFormFile("file", "upload"); err == nil {
					_, err = part.Write(image.Data)
				}
			}
			if err == nil {
				err = writer.Close()
			}
			if err != nil {
				return viewWorkEvidenceResult{}, workStateError("upload_unavailable")
			}
			uploaded, err := e.evidenceStep(c, workLifecycleSource(s, actor), "upload "+filepath.Base(image.Path), "POST", basePath+"/artifacts", writer.FormDataContentType(), buffer.Bytes())
			if err != nil {
				return viewWorkEvidenceResult{}, err
			}
			id = uploaded.ArtifactID
		}
		artifactIDs[n] = id
		refs[n] = viewer.TimelineImageRef{ArtifactID: id, Caption: image.Caption}
	}
	if s, err = e.session(c.WorkID); err != nil {
		return viewWorkEvidenceResult{}, err
	}
	request := viewWorkEvidenceEvent(c, candidateID, actor, refs, supersedes)
	if err = request.Validate(); err == nil {
		err = viewer.ValidateTimelineEventReferences(s, request)
	}
	if err != nil {
		return viewWorkEvidenceResult{}, err
	}
	raw, err := json.Marshal(request)
	if err != nil {
		return viewWorkEvidenceResult{}, err
	}
	receipt, err := e.evidenceStep(c, workLifecycleSource(s, actor), "event", "POST", basePath+"/events", "application/json", raw)
	if err != nil {
		return viewWorkEvidenceResult{}, err
	}
	// Read the work back: the event must be there with its view and images.
	if s, err = e.session(c.WorkID); err != nil {
		return viewWorkEvidenceResult{}, err
	}
	if r, ok := evidencePublished(s, c, candidateID, images, supersedes); ok {
		r.Result = "updated"
		r.SessionVersion = s.Version
		return r, nil
	}
	return viewWorkEvidenceResult{WorkID: s.ViewerSessionID, Result: "updated", SessionVersion: receipt.SessionVersion}, &viewWorkError{Code: "readback_mismatch", Field: "evidence"}
}

// evidenceCandidate returns the candidate the evidence belongs to, without
// creating it: the one with the requested commit, or the latest when no commit
// was requested. An empty ID means the requested commit is not registered yet.
func (e *viewWorkEngine) evidenceCandidate(s viewer.TimelineSession, actor viewer.TimelineActor, c viewWorkCommand) (string, error) {
	if c.Candidate == "" {
		if len(s.Candidates) == 0 {
			return "", &viewWorkError{Code: "candidate_required", Next: "pass --candidate COMMIT"}
		}
		return s.Candidates[len(s.Candidates)-1].CandidateID, nil
	}
	commit := evidenceCommit(s, actor, c)
	for _, candidate := range slices.Backward(s.Candidates) {
		if candidate.Commit == commit {
			return candidate.CandidateID, nil
		}
	}
	return "", nil
}

// evidenceCommit resolves --candidate to a full commit when the project is a
// repository that knows it, and otherwise keeps the text given.
func evidenceCommit(s viewer.TimelineSession, actor viewer.TimelineActor, c viewWorkCommand) string {
	root := workLifecycleSource(s, actor).ProjectRoot
	if root != "" {
		if out, err := newGitRunner()(root, "rev-parse", "--verify", "--quiet", c.Candidate+"^{commit}"); err == nil && strings.TrimSpace(out) != "" {
			return strings.TrimSpace(out)
		}
	}
	return c.Candidate
}

func (e *viewWorkEngine) evidenceCreateCandidate(s viewer.TimelineSession, actor viewer.TimelineActor, c viewWorkCommand) (viewer.TimelineOperationResult, error) {
	root := workLifecycleSource(s, actor).ProjectRoot
	branch, base := c.Branch, c.Base
	git := newGitRunner()
	if branch == "" && root != "" {
		if out, err := git(root, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
			branch = strings.TrimSpace(out)
		}
	}
	if base == "" && root != "" {
		if out, err := git(root, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
			base = strings.TrimPrefix(strings.TrimSpace(out), "origin/")
		}
	}
	if branch == "" || branch == "HEAD" || base == "" {
		return viewer.TimelineOperationResult{}, &viewWorkError{Code: "candidate_branch_required", Next: "pass --branch and --base for --candidate"}
	}
	r := &viewer.TimelineCandidateCreateRequest{TimelineNativeOperation: workLifecycleBase(s, actor, "Register the candidate of the evidence"), TimelineCandidate: viewer.TimelineCandidate{Branch: branch, Commit: evidenceCommit(s, actor, c), Base: base}}
	method, path, body, err := e.lifecycleRequest(s, "candidate_create", "", "POST", "/api/timeline/sessions/"+s.ViewerSessionID+"/candidates", r)
	if err != nil {
		return viewer.TimelineOperationResult{}, err
	}
	return e.evidenceStep(c, workLifecycleSource(s, actor), "candidate", method, path, "application/json", body)
}

// evidencePublished finds a validation or iteration event of the candidate
// with a view of this title whose captures are exactly these bytes.
func evidencePublished(s viewer.TimelineSession, c viewWorkCommand, candidateID string, images []viewWorkEvidenceImage, supersedes *viewWorkSupersedes) (viewWorkEvidenceResult, bool) {
	if candidateID == "" {
		return viewWorkEvidenceResult{}, false
	}
	sums := map[string]string{}
	for _, a := range s.Artifacts {
		sums[a.ArtifactID] = a.SHA256
	}
	for _, event := range slices.Backward(s.Events) {
		if event.Withdrawn || event.CandidateID != candidateID || event.Kind != "validation" && event.Kind != "iteration" {
			continue
		}
		for _, v := range event.Views {
			if v.Title != c.ViewTitle || len(v.ImageRefs) != len(images) || supersedes != nil != (v.Supersedes != nil) || supersedes != nil && v.Supersedes.ViewID != supersedes.ViewID {
				continue
			}
			same := true
			var artifacts []string
			for n, ref := range v.ImageRefs {
				same = same && sums[ref.ArtifactID] == images[n].SHA
				artifacts = append(artifacts, ref.ArtifactID)
			}
			if same {
				return viewWorkEvidenceResult{WorkID: s.ViewerSessionID, Result: "unchanged", CandidateID: candidateID, EventID: event.EventID, EvidenceID: event.EvidenceID, ViewIDs: []string{v.ViewID}, ArtifactIDs: artifacts}, true
			}
		}
	}
	return viewWorkEvidenceResult{}, false
}

// resumeEvidenceIntents replays the interrupted steps of this producer's earlier
// evidence for the work before a new one is built.
func (e *viewWorkEngine) resumeEvidenceIntents(c viewWorkCommand, source viewer.TimelineSource) error {
	for _, i := range append([]*viewWorkIntent{}, e.journal.Pending...) {
		if i.WorkID != c.WorkID || i.Action != "evidence" {
			continue
		}
		if i.Source != source {
			return &viewWorkError{Code: "pending_operation", Pending: i.ID, Next: "run hive view work resume with the original native source"}
		}
		if i.State == "rejected_conflict" {
			return &viewWorkError{Code: "pending_operation", Pending: i.ID, Next: "inspect and reconcile the rejected intent"}
		}
		if err := e.resumeEvidenceIntent(i); err != nil {
			return err
		}
	}
	return nil
}
