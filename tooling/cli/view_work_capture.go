package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"tricell-hive/tooling/viewer"
)

// `work capture` publishes one batch of captures of one screen in a single
// command. It is a thin layer over the evidence, lifecycle and queue code: it
// opens the work, starts the stage and numbers the iteration from the viewer's
// own state when the declaration is sent, then publishes the view with
// `evidence`. The screen is the view title without its iteration suffix.

const (
	viewWorkCaptureReason = "captures published"
	viewWorkCaptureNoWork = "UI captures for "
)

type viewWorkCaptureResult struct {
	viewWorkEvidenceResult
	Iteration int    `json:"iteration"`
	Screen    string `json:"screen"`
}

// viewWorkIterationSuffix is the trailing " · #N" capture appends to the screen.
var viewWorkIterationSuffix = regexp.MustCompile(` · #([0-9]{1,6})$`)

// viewWorkCaptureScreen is the title without its iteration suffix.
func viewWorkCaptureScreen(title string) string {
	title = strings.TrimSpace(title)
	return strings.TrimSpace(viewWorkIterationSuffix.ReplaceAllString(title, ""))
}

// viewWorkCaptureOpens is a queued capture that also opens the work: it is
// addressed by the local handle of its key until it is sent.
func viewWorkCaptureOpens(c viewWorkCommand) bool { return c.Action == "capture" && c.Key != "" }

// validateViewWorkCapture checks everything about the command that needs
// neither the viewer nor the files, and resolves the candidate from the
// repository the command runs in, so a queued capture keeps the commit it was
// made at.
func validateViewWorkCapture(c *viewWorkCommand) error {
	c.ViewTitle = viewWorkCaptureScreen(c.ViewTitle)
	if c.ViewTitle == "" {
		return workInput("required semantic field missing")
	}
	if c.Stage == "" {
		c.Stage = "verify"
	}
	if c.Stage != "build" && c.Stage != "verify" {
		return workInput("capture stage must be build or verify")
	}
	if c.Project == "" {
		p, err := defaultViewWorkProject()
		if err != nil {
			return workInput("project: run inside the repository or pass --project")
		}
		c.Project = p
	}
	if c.WorkID != "" {
		// The work exists: the creation fields only matter to the capture that opens it.
		c.Key, c.Title, c.Summary = "", "", ""
	}
	if c.Reason == "" {
		c.Reason = "Publish UI captures"
	}
	// The title the viewer receives carries the suffix; check it with room for the number.
	screen := c.ViewTitle
	c.ViewTitle = screen + " · #999999"
	err := validateViewWorkEvidence(c)
	c.ViewTitle = screen
	if err != nil {
		return err
	}
	if err = resolveViewWorkCaptureCandidate(c); err != nil {
		return err
	}
	if c.WorkID != "" {
		return nil
	}
	if c.Key == "" {
		out, gitErr := newGitRunner()(c.Project, "rev-parse", "--abbrev-ref", "HEAD")
		if branch := strings.TrimSpace(out); gitErr == nil && branch != "" && branch != "HEAD" {
			c.Key = branch
		} else {
			return &viewWorkError{Code: "key_required", Next: "pass --key with the ticket or the change folder name: the current checkout has no branch to name the work"}
		}
	}
	if c.Title == "" {
		c.Title = c.Key
	}
	if c.Summary == "" {
		c.Summary = viewWorkCaptureNoWork + c.Key
	}
	r := viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "validate", Tickets: []viewer.TimelineTicketInput{}, Source: viewer.TimelineSource{Host: viewer.TimelineHost(c.Host), NativeSessionID: c.NativeSessionID, ProjectRoot: c.Project}, WorkKey: &c.Key, Title: c.Title, Summary: c.Summary}
	if r.Validate() != nil {
		return workInput("invalid creation fields")
	}
	return nil
}

// resolveViewWorkCaptureCandidate fills the commit, branch and base the capture
// belongs to: the project's HEAD unless --candidate says another commit.
func resolveViewWorkCaptureCandidate(c *viewWorkCommand) error {
	git := newGitRunner()
	if c.Branch == "" {
		if out, err := git(c.Project, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
			c.Branch = strings.TrimSpace(out)
		}
	}
	if c.Base == "" {
		if out, err := git(c.Project, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
			c.Base = strings.TrimPrefix(strings.TrimSpace(out), "origin/")
		}
	}
	if c.Branch == "" || c.Branch == "HEAD" || c.Base == "" {
		return &viewWorkError{Code: "candidate_branch_required", Next: "check out a branch and set origin/HEAD (git remote set-head origin -a), or pass --branch and --base"}
	}
	ref := c.Candidate
	if ref == "" {
		ref = "HEAD"
	}
	out, err := git(c.Project, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	if sha := strings.TrimSpace(out); err != nil || sha == "" {
		return workInput("candidate")
	} else {
		c.Candidate = sha
	}
	return nil
}

// captureIteration is the latest published iteration of a screen.
type captureIteration struct {
	N     int
	Event viewer.TimelineEvent
	View  viewer.TimelineView
}

// latestCaptureIteration reads the work's views and returns the one of the
// screen with the highest number, or nil when the screen has none.
func latestCaptureIteration(s viewer.TimelineSession, screen string) *captureIteration {
	var latest *captureIteration
	for _, event := range s.Events {
		if event.Withdrawn || event.EvidenceID == "" || event.Kind != "validation" && event.Kind != "iteration" {
			continue
		}
		for _, v := range event.Views {
			m := viewWorkIterationSuffix.FindStringSubmatchIndex(v.Title)
			if m == nil || v.Title[:m[0]] != screen {
				continue
			}
			n, err := strconv.Atoi(v.Title[m[2]:m[3]])
			if err != nil {
				continue
			}
			if latest == nil || n >= latest.N {
				latest = &captureIteration{N: n, Event: event, View: v}
			}
		}
	}
	return latest
}

// captureRepeats reports whether the iteration already holds these captures,
// with the same captions, for this candidate.
func captureRepeats(s viewer.TimelineSession, it *captureIteration, candidateID string, images []viewWorkEvidenceImage, c viewWorkCommand) ([]string, bool) {
	if candidateID == "" || it.Event.CandidateID != candidateID || len(it.View.ImageRefs) != len(images) {
		return nil, false
	}
	artifacts := map[string]viewer.TimelineArtifact{}
	for _, a := range s.Artifacts {
		artifacts[a.ArtifactID] = a
	}
	ids := make([]string, len(images))
	for n, ref := range it.View.ImageRefs {
		a, ok := artifacts[ref.ArtifactID]
		if !ok || a.SHA256 != images[n].SHA || a.Viewport != c.Viewport || a.Theme != c.Theme || ref.Caption != images[n].Caption {
			return nil, false
		}
		ids[n] = ref.ArtifactID
	}
	return ids, true
}

// captureWork returns the ID of the work the capture belongs to, opening it
// first when this capture is the one that creates it.
func (e *viewWorkEngine) captureWork(c viewWorkCommand) (string, error) {
	if !viewWorkCaptureOpens(c) {
		return c.WorkID, nil
	}
	project, err := canonicalWorkDir(c.Project)
	if err != nil {
		return "", err
	}
	handle := queuedWorkHandle(project, c.Host, c.NativeSessionID, c.Key)
	// A work already addressed by a real ID was opened by an earlier attempt.
	if c.WorkID != "" && c.WorkID != handle {
		return c.WorkID, nil
	}
	if _, err = viewer.RegisterRepository(e.state, project); err != nil && !errors.Is(err, viewer.ErrNotGitRepository) {
		return "", err
	}
	open := c
	open.Action, open.WorkID, open.Project, open.Change = "open", "", project, ""
	s, err := e.open(open)
	if err != nil {
		return "", err
	}
	real := s.ViewerSessionID
	if err = addCurrentWork(e.state, c.Host, c.NativeSessionID, real); err != nil {
		return "", err
	}
	open.Project = project
	if err = e.noteHandle(open, real); err != nil {
		return "", err
	}
	if handle != real {
		if err = releaseCurrentWork(e.state, c.Host, c.NativeSessionID, handle); err != nil {
			return "", err
		}
	}
	return real, nil
}

// captureStage makes the stage ready for the capture: started when it has no
// occurrence, resumed when paused, and a linked occurrence when it ended.
func (e *viewWorkEngine) captureStage(c viewWorkCommand, s viewer.TimelineSession) error {
	latest := workLifecycleStage(s, c.Stage)
	stage := viewWorkCommand{Action: "stage", Options: c.Options, Host: c.Host, NativeSessionID: c.NativeSessionID, WorkID: c.WorkID, Stage: c.Stage, State: "in_progress", Reason: viewWorkCaptureReason}
	switch {
	case latest == nil:
	case latest.StageState == "completed" || latest.StageState == "failed":
		stage.PreviousStage = latest.EventID
	case latest.StageState == "paused":
	default:
		return nil
	}
	_, err := e.lifecycle(stage)
	return err
}

func (e *viewWorkEngine) capture(c viewWorkCommand) (viewWorkCaptureResult, error) {
	id, err := e.captureWork(c)
	if err != nil {
		return viewWorkCaptureResult{}, err
	}
	// The actor is matched by native identity from here on, as `evidence` does.
	c.WorkID, c.Project = id, ""
	images, err := readViewWorkImages(c)
	if err != nil {
		return viewWorkCaptureResult{}, err
	}
	s, err := e.session(id)
	if err != nil {
		return viewWorkCaptureResult{}, err
	}
	actor, err := e.actor(s, c)
	if err != nil {
		return viewWorkCaptureResult{}, err
	}
	if err = e.resumeEvidenceIntents(c, workLifecycleSource(s, actor)); err != nil {
		return viewWorkCaptureResult{}, err
	}
	if s, err = e.session(id); err != nil {
		return viewWorkCaptureResult{}, err
	}
	if s.State == viewer.TimelineClosed || s.State == viewer.TimelineFailed {
		return viewWorkCaptureResult{}, workStateError("session_closed")
	}
	if err = e.captureStage(c, s); err != nil {
		return viewWorkCaptureResult{}, err
	}
	if s, err = e.session(id); err != nil {
		return viewWorkCaptureResult{}, err
	}
	screen := c.ViewTitle
	latest := latestCaptureIteration(s, screen)
	next := 1
	if latest != nil {
		candidateID, err := e.evidenceCandidate(s, actor, c)
		if err != nil {
			return viewWorkCaptureResult{}, err
		}
		if artifacts, same := captureRepeats(s, latest, candidateID, images, c); same {
			r := viewWorkEvidenceResult{WorkID: s.ViewerSessionID, Result: "unchanged", SessionVersion: s.Version, CandidateID: candidateID, EventID: latest.Event.EventID, EvidenceID: latest.Event.EvidenceID, ViewIDs: []string{latest.View.ViewID}, ArtifactIDs: artifacts}
			return viewWorkCaptureResult{viewWorkEvidenceResult: r, Iteration: latest.N, Screen: screen}, nil
		}
		next = latest.N + 1
	}
	publish := c
	publish.Action = "evidence"
	publish.ViewTitle = fmt.Sprintf("%s · #%d", screen, next)
	if latest != nil {
		publish.Supersedes = latest.View.ViewID
	}
	r, err := e.evidence(publish)
	if err != nil {
		return viewWorkCaptureResult{}, err
	}
	return viewWorkCaptureResult{viewWorkEvidenceResult: r, Iteration: next, Screen: screen}, nil
}

// capViewWorkCaptureResult lists the artifact IDs that fit the bounded reply.
func capViewWorkCaptureResult(r *viewWorkCaptureResult) {
	for len(r.ArtifactIDs) > 0 && !fitsViewWorkOutput(*r) {
		r.ArtifactIDs = r.ArtifactIDs[:len(r.ArtifactIDs)-1]
		r.ArtifactsOmitted++
	}
}

const viewWorkCaptureHelp = `hive view work capture — publish a batch of UI captures of one screen in one command

hive view work capture --view-title TEXT --image FILE[:caption] ... [--stage build|verify] [--key KEY] [--title TEXT] [--summary TEXT] [--candidate COMMIT] [--work-id ID] [--branch B --base B] [--viewport desktop|mobile] [--theme light|dark|unspecified] [--project ROOT]

--view-title names the screen, without a number. Repeat --image once per capture of the screen (1..8 files: PNG, JPEG or WebP).

What one call does, in order, when the command is sent:
1. Work: acts on the work this session has open; with several open, pass --work-id. With none it opens one: --key is the ticket or the change folder name (default: the current branch; in a detached HEAD, --key is required), --title defaults to the key and --summary (one paragraph on what the work is about; give it on the first capture) to "UI captures for KEY".
2. Stage (default verify): started when it has no occurrence, resumed when paused, and a linked new occurrence when it ended. Nothing else is declared.
3. Candidate: --candidate COMMIT, default HEAD, resolved when you run the command, on the current branch against origin/HEAD. Without a branch or a base the command stops with candidate_branch_required and publishes nothing: check out a branch, run git remote set-head origin -a, or pass --branch and --base.
4. Iteration: the view is titled "<screen> · #N". N is one more than the highest of the work's views with that screen, and the new view supersedes it; the first is "· #1". Another screen starts at "· #1".
5. Repeat: the same images, captions and commit as the latest iteration reply unchanged and create nothing; another commit or another capture creates "· #N+1".

The reply (at most 4096 bytes) is the receipt: work_id, result updated|unchanged, iteration, screen, candidate_id, event_id, view_ids and artifact_ids. A reply of result updated or unchanged confirms the write: do not read the work back.

Viewer stopped or silent: the command replies queued (work_id is a local handle), never starts the viewer and freezes the images. The next work command that finds the viewer running (capture, sync, close) sends the queue in order: the work, the stage and each capture, numbered when it is sent, so two captures of one screen made while it was stopped come out as "· #1" and "· #2". Keep the image files until a later reply is not queued.

Close the work at its completion with: hive view work close --outcome completed|abandoned --reason TEXT (or work sync --stage STAGE --state in_review|paused|awaiting_validation --reason TEXT while it waits). hive view work contract lists every command.
`
