package main

import (
	"bytes"
	"encoding/json"
	"image/color"
	"os/exec"
	"strings"
	"testing"
	"tricell-hive/tooling/viewer"
)

// captureCLI drives `hive view work capture` as one native session against the
// fixture's viewer, in a repository with a commit, a branch and origin/HEAD.
type captureCLI struct {
	*queueCLI
	shots string
}

func captureGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", root, "-c", "user.email=t@example.test", "-c", "user.name=Test", "-c", "commit.gpgsign=false"}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func newCaptureCLI(t *testing.T) *captureCLI {
	q := newQueueCLI(t)
	captureGit(t, q.project, "symbolic-ref", "HEAD", "refs/heads/development")
	captureGit(t, q.project, "commit", "--allow-empty", "-m", "first")
	captureGit(t, q.project, "update-ref", "refs/remotes/origin/development", "HEAD")
	captureGit(t, q.project, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/development")
	return &captureCLI{queueCLI: q, shots: t.TempDir()}
}

func (c *captureCLI) shot(name string, rgba color.RGBA) string {
	return testPNG(c.t, c.shots, name, rgba)
}

func (c *captureCLI) capture(title string, extra ...string) viewWorkCaptureResult {
	c.t.Helper()
	out := c.must("", append([]string{"capture", "--project", c.project, "--view-title", title}, extra...)...)
	return decodeCapture(c.t, out)
}

func decodeCapture(t *testing.T, out string) viewWorkCaptureResult {
	t.Helper()
	var r viewWorkCaptureResult
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("receipt %q: %v", out, err)
	}
	if len(out) > 4096 {
		t.Fatalf("receipt of %d bytes exceeds the bounded output", len(out))
	}
	return r
}

// captureViews lists the views of a work whose title starts with the base, in
// the order they were published.
type publishedView struct{ Title, ViewID, Supersedes, CandidateID string }

func captureViews(s viewer.TimelineSession, base string) []publishedView {
	var views []publishedView
	for _, event := range s.Events {
		if event.Withdrawn {
			continue
		}
		for _, v := range event.Views {
			if strings.HasPrefix(v.Title, base) {
				pv := publishedView{Title: v.Title, ViewID: v.ViewID, CandidateID: event.CandidateID}
				if v.Supersedes != nil {
					pv.Supersedes = v.Supersedes.ViewID
				}
				views = append(views, pv)
			}
		}
	}
	return views
}

// AC1: with no open work, one command leaves the stage, the candidate and the
// numbered view in the viewer.
func TestWorkCaptureOpensWorkStageAndViewInOneCommand(t *testing.T) {
	c := newCaptureCLI(t)
	if got := c.workIDs(); len(got) != 0 {
		t.Fatalf("current works %v", got)
	}
	a := c.shot("a.png", color.RGBA{R: 200, A: 255})
	r := c.capture("Login", "--stage", "verify", "--image", a+":Login desktop", "--key", "capture-one", "--summary", "Login screen captures")
	if r.Result != "updated" || r.Iteration != 1 || r.Screen != "Login" || len(r.ViewIDs) != 1 || r.WorkID == "" {
		t.Fatalf("receipt %+v", r)
	}
	s := c.session(r.WorkID)
	if s.WorkKey == nil || *s.WorkKey != "capture-one" || s.Title != "capture-one" || s.Summary != "Login screen captures" {
		t.Fatalf("work %v %q %q", s.WorkKey, s.Title, s.Summary)
	}
	stage := workLifecycleStage(s, "verify")
	if stage == nil || stage.StageState != "in_progress" || stage.Reason != "captures published" {
		t.Fatalf("stage %+v", stage)
	}
	views := captureViews(s, "Login")
	if len(views) != 1 || views[0].Title != "Login · #1" || views[0].ViewID != r.ViewIDs[0] || views[0].Supersedes != "" {
		t.Fatalf("views %+v", views)
	}
	head := captureGit(t, c.project, "rev-parse", "HEAD")
	if len(s.Candidates) != 1 || s.Candidates[0].Commit != head || s.Candidates[0].Branch != "development" || s.Candidates[0].Base != "development" {
		t.Fatalf("candidates %+v, want HEAD %s", s.Candidates, head)
	}
	if got := c.workIDs(); len(got) != 1 || got[0] != r.WorkID {
		t.Fatalf("current works %v", got)
	}
	// Without --key the work is named after the current branch.
	other := newCaptureCLI(t)
	o := other.capture("Home", "--image", other.shot("h.png", color.RGBA{G: 200, A: 255}))
	if work := other.session(o.WorkID); work.WorkKey == nil || *work.WorkKey != "development" || workLifecycleStage(work, "verify") == nil {
		t.Fatalf("default key %v, stage %v", work.WorkKey, workLifecycleStage(work, "verify"))
	}
}

// AC2: a later capture of the same screen supersedes the previous one and
// numbers the iteration; another screen starts at #1.
func TestWorkCaptureNumbersIterationsAndSupersedes(t *testing.T) {
	c := newCaptureCLI(t)
	first := c.capture("Login", "--image", c.shot("a.png", color.RGBA{R: 200, A: 255}), "--key", "iterations")
	second := c.capture("Login", "--image", c.shot("a.png", color.RGBA{R: 100, A: 255}))
	profile := c.capture("Profile", "--image", c.shot("p.png", color.RGBA{B: 200, A: 255}))
	if first.Iteration != 1 || second.Iteration != 2 || profile.Iteration != 1 || second.WorkID != first.WorkID || profile.WorkID != first.WorkID {
		t.Fatalf("receipts %+v %+v %+v", first, second, profile)
	}
	s := c.session(first.WorkID)
	login := captureViews(s, "Login")
	if len(login) != 2 || login[0].Title != "Login · #1" || login[1].Title != "Login · #2" || login[1].Supersedes != login[0].ViewID || login[0].Supersedes != "" {
		t.Fatalf("login views %+v", login)
	}
	if p := captureViews(s, "Profile"); len(p) != 1 || p[0].Title != "Profile · #1" || p[0].Supersedes != "" {
		t.Fatalf("profile views %+v", p)
	}
	if stageEvents(s, "verify") != 1 || len(s.Candidates) != 1 {
		t.Fatalf("%d verify stages, %d candidates", stageEvents(s, "verify"), len(s.Candidates))
	}
}

// AC2 and AC3: with the viewer stopped nothing starts it, and the queue sends
// the work, the stage and both iterations once, in order.
func TestWorkCaptureQueuedWhileTheViewerIsStopped(t *testing.T) {
	c := newCaptureCLI(t)
	c.f.mustView("stop")
	a := c.shot("a.png", color.RGBA{R: 200, A: 255})
	first := decodeQueued(t, c.must("", "capture", "--project", c.project, "--view-title", "Login", "--image", a, "--key", "offline-captures"))
	testPNG(t, c.shots, "a.png", color.RGBA{R: 50, A: 255})
	second := decodeQueued(t, c.must("", "capture", "--project", c.project, "--view-title", "Login", "--image", a))
	if first.WorkID != second.WorkID || first.Position != 1 || second.Position != 2 || first.OperationID == second.OperationID {
		t.Fatalf("queued %+v %+v", first, second)
	}
	c.assertViewerNotStarted()
	if got := c.workIDs(); len(got) != 1 || got[0] != first.WorkID {
		t.Fatalf("current works %v", got)
	}
	if err := validateWorkQueue(c.journal()); err != nil {
		t.Fatalf("journal: %v", err)
	}

	c.f.mustView("start")
	profile := c.capture("Profile", "--image", c.shot("p.png", color.RGBA{B: 200, A: 255}))
	if profile.Queue == nil || profile.Queue.Sent != 2 || len(profile.Queue.Rejected) != 0 || profile.Iteration != 1 {
		t.Fatalf("profile %+v", profile)
	}
	if n := c.workCount(); n != 1 {
		t.Fatalf("%d works, want 1", n)
	}
	s := c.session(profile.WorkID)
	login := captureViews(s, "Login")
	if len(login) != 2 || login[0].Title != "Login · #1" || login[1].Title != "Login · #2" || login[1].Supersedes != login[0].ViewID {
		t.Fatalf("login views %+v", login)
	}
	if stageEvents(s, "verify") != 1 || s.WorkKey == nil || *s.WorkKey != "offline-captures" {
		t.Fatalf("%d verify stages, key %v", stageEvents(s, "verify"), s.WorkKey)
	}
	// The two frozen captures kept their own bytes: two artifacts for Login.
	if len(s.Artifacts) != 3 {
		t.Fatalf("%d artifacts, want 3", len(s.Artifacts))
	}
	if j := c.journal(); len(j.Queue) != 0 || len(j.Pending) != 0 {
		t.Fatalf("journal queue %d pending %d", len(j.Queue), len(j.Pending))
	}
	if got := c.workIDs(); len(got) != 1 || got[0] != profile.WorkID {
		t.Fatalf("current works %v, want the viewer's ID", got)
	}
}

// A capture of a stage that another stage pushed to the background resumes it.
func TestWorkCaptureResumesAPausedStage(t *testing.T) {
	c := newCaptureCLI(t)
	build := c.capture("Form", "--stage", "build", "--image", c.shot("b.png", color.RGBA{R: 200, A: 255}), "--key", "stages")
	verify := c.capture("Form check", "--stage", "verify", "--image", c.shot("v.png", color.RGBA{G: 200, A: 255}))
	s := c.session(build.WorkID)
	if b := workLifecycleStage(s, "build"); b == nil || b.StageState != "paused" {
		t.Fatalf("build stage %+v after a verify capture", b)
	}
	again := c.capture("Form", "--stage", "build", "--image", c.shot("b.png", color.RGBA{R: 90, A: 255}))
	if again.Iteration != 2 || again.WorkID != verify.WorkID {
		t.Fatalf("receipt %+v", again)
	}
	s = c.session(build.WorkID)
	if b := workLifecycleStage(s, "build"); b == nil || b.StageState != "in_progress" || stageEvents(s, "build") != 1 {
		t.Fatalf("build stage %+v, %d occurrences", b, stageEvents(s, "build"))
	}
}

// A capture on an ended stage opens a linked occurrence.
func TestWorkCaptureReentersAnEndedStage(t *testing.T) {
	c := newCaptureCLI(t)
	first := c.capture("Login", "--image", c.shot("a.png", color.RGBA{R: 200, A: 255}), "--key", "reentry")
	ended := workLifecycleStage(c.session(first.WorkID), "verify")
	c.must("", "stage", "--stage", "verify", "--state", "completed", "--reason", "Verified")
	second := c.capture("Login", "--image", c.shot("a.png", color.RGBA{R: 90, A: 255}))
	if second.Iteration != 2 {
		t.Fatalf("receipt %+v", second)
	}
	s := c.session(first.WorkID)
	latest := workLifecycleStage(s, "verify")
	if stageEvents(s, "verify") != 2 || latest.EventID == ended.EventID || latest.StageState != "in_progress" || latest.RelatedEventID != ended.EventID {
		t.Fatalf("%d occurrences, latest %+v after %s", stageEvents(s, "verify"), latest, ended.EventID)
	}
}

// A candidate that cannot be resolved publishes nothing, queued or not.
func TestWorkCaptureRefusesAnUnresolvableCandidate(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(c *captureCLI)
	}{
		{"detached HEAD", func(c *captureCLI) { captureGit(t, c.project, "checkout", "-q", "--detach") }},
		{"no origin HEAD", func(c *captureCLI) { captureGit(t, c.project, "symbolic-ref", "--delete", "refs/remotes/origin/HEAD") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newCaptureCLI(t)
			tc.setup(c)
			args := []string{"capture", "--project", c.project, "--view-title", "Login", "--image", c.shot("a.png", color.RGBA{R: 200, A: 255}), "--key", "no-candidate"}
			if _, err := c.run("", args...); err == nil || !strings.Contains(err.Error(), "candidate_branch_required") || !strings.Contains(err.Error(), "--branch") {
				t.Fatalf("error %v", err)
			}
			if n := c.workCount(); n != 0 || len(c.workIDs()) != 0 {
				t.Fatalf("%d works, current %v", n, c.workIDs())
			}
			c.f.mustView("stop")
			if _, err := c.run("", args...); err == nil || !strings.Contains(err.Error(), "candidate_branch_required") {
				t.Fatalf("queued error %v", err)
			}
			if len(c.journal().Queue) != 0 {
				t.Fatal("a refused capture was queued")
			}
		})
	}
}

// AC4: the same images, captions and commit are not a new iteration; another
// commit is.
func TestWorkCaptureRepeatIsUnchangedUntilTheCommitMoves(t *testing.T) {
	c := newCaptureCLI(t)
	a := c.shot("a.png", color.RGBA{R: 200, A: 255})
	first := c.capture("Login", "--image", a+":Desktop", "--key", "repeat-capture")
	again := c.capture("Login", "--image", a+":Desktop")
	if again.Result != "unchanged" || again.Iteration != 1 || len(again.ViewIDs) != 1 || again.ViewIDs[0] != first.ViewIDs[0] {
		t.Fatalf("repeat %+v after %+v", again, first)
	}
	if got := captureViews(c.session(first.WorkID), "Login"); len(got) != 1 {
		t.Fatalf("views %+v", got)
	}
	// Another caption is another capture.
	captioned := c.capture("Login", "--image", a+":Mobile")
	if captioned.Result != "updated" || captioned.Iteration != 2 {
		t.Fatalf("captioned %+v", captioned)
	}
	captureGit(t, c.project, "commit", "--allow-empty", "-m", "second")
	next := c.capture("Login", "--image", a+":Mobile")
	if next.Result != "updated" || next.Iteration != 3 {
		t.Fatalf("after a new commit %+v", next)
	}
	s := c.session(first.WorkID)
	login := captureViews(s, "Login")
	if len(login) != 3 || login[2].Title != "Login · #3" || login[2].Supersedes != login[1].ViewID || len(s.Candidates) != 2 || login[2].CandidateID == login[1].CandidateID {
		t.Fatalf("views %+v candidates %+v", login, s.Candidates)
	}
}

func TestWorkCaptureRequiresAWorkIDWhenSeveralWorksAreOpen(t *testing.T) {
	c := newCaptureCLI(t)
	one := decodeSync(t, c.open("work-one"))
	decodeSync(t, c.open("work-two"))
	args := []string{"capture", "--project", c.project, "--view-title", "Login", "--image", c.shot("a.png", color.RGBA{R: 200, A: 255})}
	if _, err := c.run("", args...); err == nil || !strings.Contains(err.Error(), "work-id required: several works are open in this session") {
		t.Fatalf("error %v", err)
	}
	r := decodeCapture(t, c.must("", append(args, "--work-id", one.WorkID)...))
	if r.WorkID != one.WorkID || r.Iteration != 1 {
		t.Fatalf("receipt %+v", r)
	}
}

func TestWorkCaptureValidatesItsInputBeforeAnything(t *testing.T) {
	c := newCaptureCLI(t)
	image := c.shot("a.png", color.RGBA{R: 200, A: 255})
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"no title", []string{"capture", "--project", c.project, "--image", image}},
		{"no image", []string{"capture", "--project", c.project, "--view-title", "Login"}},
		{"bad stage", []string{"capture", "--project", c.project, "--view-title", "Login", "--image", image, "--stage", "deploy"}},
		{"planning stage", []string{"capture", "--project", c.project, "--view-title", "Login", "--image", image, "--stage", "plan"}},
		{"not an evidence flag", []string{"capture", "--project", c.project, "--view-title", "Login", "--image", image, "--supersedes", strings.Repeat("a", 64)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := c.run("", tc.args...); err == nil || !strings.HasPrefix(err.Error(), "invalid_input") {
				t.Fatalf("error %v", err)
			}
		})
	}
	if c.workCount() != 0 {
		t.Fatal("a refused capture created a work")
	}
}

// The contract lists capture in the build and verify flows, and --help says
// everything a flow needs without reading the contract.
func TestWorkCaptureContractAndHelp(t *testing.T) {
	for _, want := range []string{"capture --view-title", "build or verify", "· #N", "unchanged"} {
		if !strings.Contains(viewWorkContract, want) {
			t.Errorf("contract lacks %q", want)
		}
	}
	var out bytes.Buffer
	if err := runViewWork([]string{"capture", "--help"}, &out, strings.NewReader("")); err != nil {
		t.Fatal(err)
	}
	help := out.String()
	for _, want := range []string{
		"--view-title", "--image FILE[:caption]", "--stage build|verify", "--key", "--summary", "--candidate", "--work-id",
		"· #N", "supersedes", "queued", "unchanged", "candidate_branch_required", "HEAD", "work close",
	} {
		if !strings.Contains(help, want) {
			t.Errorf("capture help lacks %q", want)
		}
	}
	if strings.Contains(help, "tasks register") {
		t.Error("capture help sends the reader to the task commands")
	}
}

// A screen captured in build and again in verify is one chain of iterations,
// and a capture queued for a work that is already open numbers after it.
func TestWorkCaptureChainsAcrossStagesAndQueuesForAnOpenWork(t *testing.T) {
	c := newCaptureCLI(t)
	a := c.shot("a.png", color.RGBA{R: 200, A: 255})
	build := c.capture("Form", "--stage", "build", "--image", a, "--key", "chain")
	verify := c.capture("Form", "--stage", "verify", "--image", c.shot("b.png", color.RGBA{G: 200, A: 255}))
	if verify.Iteration != 2 {
		t.Fatalf("verify %+v", verify)
	}
	c.f.mustView("stop")
	queued := decodeQueued(t, c.must("", "capture", "--project", c.project, "--view-title", "Form", "--stage", "verify", "--image", c.shot("c.png", color.RGBA{B: 200, A: 255})))
	if queued.WorkID != build.WorkID || queued.Position != 1 {
		t.Fatalf("queued %+v, want the open work %s", queued, build.WorkID)
	}
	c.f.mustView("start")
	last := c.capture("Form", "--stage", "verify", "--image", c.shot("d.png", color.RGBA{R: 9, G: 9, A: 255}))
	if last.Iteration != 4 || last.Queue == nil || last.Queue.Sent != 1 {
		t.Fatalf("last %+v", last)
	}
	forms := captureViews(c.session(build.WorkID), "Form")
	if len(forms) != 4 || forms[1].Supersedes != forms[0].ViewID || forms[2].Title != "Form · #3" || forms[3].Supersedes != forms[2].ViewID {
		t.Fatalf("views %+v", forms)
	}
}
