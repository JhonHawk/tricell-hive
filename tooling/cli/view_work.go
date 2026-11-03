package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

// viewWorkCommand contains semantic intent only. Runtime dispatch resolves all
// producer/task/segment IDs, operation IDs and optimistic versions itself.
type viewWorkCommand struct {
	Action                                       string
	Options                                      management.Options
	Project, Change, WorkID, Key, Title, Summary string
	Host, NativeSessionID                        string
	Stage, State, Reason, PreviousStage          string
	DocumentKey, Status                          string
	Documents                                    []string
	ExplicitDocuments                            bool
	Outcome, Pending, ReconcileAction            string
	PlanRevision                                 string
	Cursor                                       string
	Limit                                        int
	IncludeArchived                              bool
	Block                                        *viewer.TimelineBlock
	Resolution, ReviewKind                       string
	Tasks                                        []viewWorkTaskInput
	Metadata                                     *viewWorkTaskMetadata
}

type viewWorkSource struct {
	Host            viewer.TimelineHost `json:"host"`
	NativeSessionID string              `json:"native_session_id"`
}
type viewWorkSegmentInput struct {
	Stage    viewer.TimelineStage `json:"stage"`
	Order    int64                `json:"order"`
	Assignee *viewWorkSource      `json:"assignee,omitempty"`
}
type viewWorkTaskInput struct {
	Key      string                 `json:"key"`
	Title    string                 `json:"title"`
	Summary  string                 `json:"summary"`
	Order    int64                  `json:"order"`
	Segments []viewWorkSegmentInput `json:"segments"`
}
type viewWorkTaskMetadata struct {
	Title             *string              `json:"title,omitempty"`
	Summary           *string              `json:"summary,omitempty"`
	Order             *int64               `json:"order,omitempty"`
	Stage             viewer.TimelineStage `json:"stage,omitempty"`
	SegmentOrder      *int64               `json:"segment_order,omitempty"`
	Assignee          *viewWorkSource      `json:"assignee,omitempty"`
	Withdrawn         *bool                `json:"withdrawn,omitempty"`
	PlanRevision      string               `json:"plan_revision,omitempty"`
	DecisionReference string               `json:"decision_reference,omitempty"`
}

type viewWorkResult struct {
	WorkID         string `json:"work_id"`
	Result         string `json:"result"`
	SessionVersion int64  `json:"session_version"`
	RevisionID     string `json:"revision_id,omitempty"`
	TaskID         string `json:"task_id,omitempty"`
	StageID        string `json:"stage_id,omitempty"`
	DocumentCount  *int   `json:"document_count,omitempty"`
	ChangedCount   *int   `json:"changed_count,omitempty"`
}

type viewWorkError struct{ Code, Field, Pending, Next string }

func (e *viewWorkError) Error() string {
	text := e.Code
	if e.Field != "" {
		text += ": " + e.Field
	}
	if e.Pending != "" {
		text += "; pending=" + e.Pending
	}
	if e.Next != "" {
		text += "; " + e.Next
	}
	return text
}
func workInput(field string) error {
	return &viewWorkError{Code: "invalid_input", Field: field, Next: "run `hive view work contract`"}
}

func runViewWork(args []string, w io.Writer, stdin io.Reader) error {
	if len(args) == 0 || len(args) == 1 && args[0] == "help" || len(args) > 0 && (args[len(args)-1] == "--help" || args[len(args)-1] == "-h") {
		_, err := io.WriteString(w, viewWorkContract)
		return err
	}
	if args[0] == "contract" {
		if len(args) != 1 {
			return workInput("contract accepts no flags")
		}
		_, err := io.WriteString(w, viewWorkContract)
		return err
	}
	c, err := parseViewWorkCommand(args, stdin)
	if err != nil {
		return err
	}
	return dispatchViewWork(c, w)
}

func dispatchViewWork(c viewWorkCommand, w io.Writer) error {
	return executeViewWork(c, w)
}

type viewWorkStrings []string

func (s *viewWorkStrings) String() string     { return strings.Join(*s, ",") }
func (s *viewWorkStrings) Set(v string) error { *s = append(*s, v); return nil }

func parseViewWorkCommand(args []string, stdin io.Reader) (viewWorkCommand, error) {
	var c viewWorkCommand
	if len(args) == 0 {
		return c, workInput("command required")
	}
	c.Action = args[0]
	args = args[1:]
	if c.Action == "tasks" || c.Action == "task" {
		if len(args) == 0 {
			return c, workInput("task action required")
		}
		c.Action += " " + args[0]
		args = args[1:]
	}
	allowed := map[string]string{
		"open":           "project change work-id key title summary",
		"publish":        "project change work-id key title summary stage reason document-key status document explicit-documents",
		"list":           "project include-archived limit cursor",
		"show":           "work-id",
		"stage":          "work-id stage state reason summary previous-stage",
		"tasks register": "work-id json plan-revision reason",
		"task update":    "work-id key stage state reason json block-reason block-owner block-step resume-when resolution review-kind",
		"close":          "work-id outcome reason",
		"resume":         "work-id",
		"reconcile":      "work-id pending action reason",
		"bind":           "work-id change reason",
	}
	flags, ok := allowed[c.Action]
	if !ok {
		return c, workInput("unknown work command")
	}
	fs := flag.NewFlagSet("view work "+c.Action, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	str := map[string]*string{"project": &c.Project, "change": &c.Change, "work-id": &c.WorkID, "key": &c.Key, "title": &c.Title, "summary": &c.Summary, "host": &c.Host, "native-session-id": &c.NativeSessionID, "stage": &c.Stage, "state": &c.State, "reason": &c.Reason, "previous-stage": &c.PreviousStage, "document-key": &c.DocumentKey, "status": &c.Status, "outcome": &c.Outcome, "pending": &c.Pending, "action": &c.ReconcileAction, "plan-revision": &c.PlanRevision, "cursor": &c.Cursor, "resolution": &c.Resolution, "review-kind": &c.ReviewKind}
	var jsonFile, blockReason, blockOwner, resumeWhen string
	str["json"] = &jsonFile
	str["block-reason"] = &blockReason
	str["block-owner"] = &blockOwner
	str["resume-when"] = &resumeWhen
	var docs, steps viewWorkStrings
	for _, name := range strings.Fields(flags + " host native-session-id") {
		switch name {
		case "include-archived":
			fs.BoolVar(&c.IncludeArchived, name, false, "")
		case "explicit-documents":
			fs.BoolVar(&c.ExplicitDocuments, name, false, "")
		case "limit":
			fs.IntVar(&c.Limit, name, 20, "")
		case "document":
			fs.Var(&docs, name, "")
		case "block-step":
			fs.Var(&steps, name, "")
		default:
			fs.StringVar(str[name], name, "", "")
		}
	}
	fs.StringVar(&c.Options.Home, "home", "", "")
	fs.StringVar(&c.Options.StateDir, "state-dir", "", "")
	c.Options.Scope = "user"
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		return c, workInput("unsupported flags or positional arguments")
	}
	seen := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	counts := map[string]int{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		name, _, _ := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		counts[name]++
		f := fs.Lookup(name)
		if f != nil && !strings.Contains(arg, "=") {
			boolean, ok := f.Value.(interface{ IsBoolFlag() bool })
			if !ok || !boolean.IsBoolFlag() {
				i++
			}
		}
	}
	for name, n := range counts {
		if n > 1 && name != "document" && name != "block-step" {
			return c, workInput("duplicate flag")
		}
	}
	if seen["json"] && jsonFile == "" {
		return c, workInput("json input path required")
	}
	c.Documents = docs
	if blockReason != "" || blockOwner != "" || resumeWhen != "" || len(steps) > 0 {
		c.Block = &viewer.TimelineBlock{Reason: blockReason, Owner: blockOwner, ResumeWhen: resumeWhen, Steps: steps}
	}
	if err := validateViewWorkCommand(&c, seen); err != nil {
		return c, err
	}
	if jsonFile != "" {
		raw, err := viewTimelineReadInput(jsonFile, stdin, viewer.TimelineMaxBody)
		if err != nil {
			return c, workInput("cannot read bounded JSON input")
		}
		if c.Action == "tasks register" {
			err = viewer.DecodeTimelineJSON(raw, &c.Tasks)
			if err == nil {
				err = validateViewWorkTasks(c.Tasks)
			}
		} else {
			c.Metadata = &viewWorkTaskMetadata{}
			err = viewer.DecodeTimelineJSON(raw, c.Metadata)
			if err == nil {
				err = validateViewWorkMetadata(*c.Metadata)
			}
		}
		if err != nil {
			return c, workInput("invalid semantic JSON")
		}
	}
	return c, nil
}

func validateViewWorkCommand(c *viewWorkCommand, seen map[string]bool) error {
	required := func(values ...string) error {
		for _, v := range values {
			if strings.TrimSpace(v) == "" {
				return workInput("required semantic field missing")
			}
		}
		return nil
	}
	if c.Action != "list" && c.Action != "show" || c.Host != "" || c.NativeSessionID != "" {
		if err := required(c.Host, c.NativeSessionID); err != nil {
			return err
		}
		if err := validateViewWorkSource(viewWorkSource{Host: viewer.TimelineHost(c.Host), NativeSessionID: c.NativeSessionID}); err != nil {
			return workInput("invalid native source")
		}
	}
	if c.Project != "" {
		p, err := filepath.Abs(c.Project)
		if err != nil {
			return workInput("invalid project")
		}
		c.Project = p
	}
	if c.WorkID != "" && !viewTimelineSessionID.MatchString(c.WorkID) {
		return workInput("invalid work-id")
	}
	switch c.Action {
	case "open", "publish":
		if err := required(c.Change); err != nil {
			return err
		}
		if c.WorkID != "" {
			if seen["key"] || seen["title"] || seen["summary"] {
				return workInput("creation fields conflict with work-id")
			}
		} else {
			if err := required(c.Project, c.Key, c.Title, c.Summary); err != nil {
				return err
			}
			r := viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "validate", Tickets: []viewer.TimelineTicketInput{}, Source: viewer.TimelineSource{Host: viewer.TimelineHost(c.Host), NativeSessionID: c.NativeSessionID, ProjectRoot: c.Project}, WorkKey: &c.Key, Title: c.Title, Summary: c.Summary}
			if r.Validate() != nil {
				return workInput("invalid creation fields")
			}
		}
		if c.Action == "publish" {
			if c.WorkID == "" {
				if err := required(c.Reason); err != nil {
					return err
				}
			}
			if c.Status == "" {
				c.Status = "draft"
			}
			if c.Status != "draft" && c.Status != "complete" {
				return workInput("invalid snapshot status")
			}
			if c.DocumentKey == "" && c.Stage == "plan" {
				c.DocumentKey = "plan"
			}
			if err := required(c.Stage, c.DocumentKey); err != nil {
				return err
			}
			if !validWorkStage(c.Stage) {
				return workInput("invalid stage")
			}
			// The existing snapshot validator owns document-key syntax.
			req := viewer.TimelineDocumentSnapshotRequest{SchemaVersion: 3, OperationID: "validate", Actor: viewer.TimelineActor{Kind: "native_producer", ProducerID: "validate"}, ExpectedVersion: 1, ExpectedSessionVersion: 1, DocumentKey: c.DocumentKey, Status: c.Status, Tabs: []viewer.TimelineDocumentTab{{Label: "Validate", DocumentID: "validate"}}}
			if req.Validate() != nil {
				return workInput("invalid document key or status")
			}
			if (c.ExplicitDocuments || c.Stage != "plan") && len(c.Documents) == 0 {
				return workInput("explicit selection requires document")
			}
			for _, d := range c.Documents {
				if d == "" || filepath.IsAbs(d) || filepath.Clean(d) == ".." || strings.HasPrefix(filepath.Clean(d), ".."+string(filepath.Separator)) || !strings.EqualFold(filepath.Ext(d), ".md") {
					return workInput("document must be relative Markdown within change")
				}
			}
		}
	case "list":
		if err := required(c.Project); err != nil {
			return err
		}
		if c.Limit < 1 || c.Limit > 100 {
			return workInput("limit must be 1..100")
		}
	default:
		if err := required(c.WorkID); err != nil {
			return err
		}
	}
	switch c.Action {
	case "stage":
		if err := required(c.Stage, c.State, c.Reason); err != nil {
			return err
		}
		state := viewer.TimelineStageState(c.State)
		n := int64(1)
		r := viewer.TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "validate", Actor: viewer.TimelineActor{Kind: "native_producer", ProducerID: "validate"}, ExpectedVersion: 1, ExpectedSessionVersion: &n, StageState: &state, Reason: c.Reason}
		if c.Summary != "" {
			r.Summary = &c.Summary
		}
		if !validWorkStage(c.Stage) || r.Validate() != nil {
			return workInput("invalid stage transition")
		}
		if c.PreviousStage != "" && !viewTimelineObjectID.MatchString(c.PreviousStage) {
			return workInput("invalid previous-stage")
		}
	case "tasks register":
		if err := required(c.PlanRevision); err != nil {
			return err
		}
		if !seen["json"] || !viewTimelineObjectID.MatchString(c.PlanRevision) {
			return workInput("registration requires json and confirmed plan-revision")
		}
	case "task update":
		if err := required(c.Key, c.Reason); err != nil {
			return err
		}
		keyCheck := viewer.TimelineTaskCreateRequest{TimelineNativeOperation: workValidationOperation("Validate"), PlanOrder: 1, TaskKey: c.Key, Title: "Validate", Segments: []viewer.TimelineTaskSegmentInput{{PlannedStage: "build", Order: 1}}}
		if viewer.ValidateTimelineV3Operation(&keyCheck) != nil {
			return workInput("invalid task key")
		}
		if seen["json"] {
			for _, f := range []string{"stage", "state", "block-reason", "block-owner", "block-step", "resume-when", "resolution", "review-kind"} {
				if seen[f] {
					return workInput("metadata and transition modes are exclusive")
				}
			}
		} else {
			if !validWorkStage(c.Stage) {
				return workInput("invalid stage")
			}
			r := viewer.TimelineTaskTransitionRequest{TimelineNativeOperation: workValidationOperation(c.Reason), ExpectedTaskVersion: 1, SegmentID: "validate", State: c.State, Block: c.Block, Resolution: c.Resolution, ReviewKind: c.ReviewKind}
			if viewer.ValidateTimelineV3Operation(&r) != nil {
				return workInput("invalid task transition or incomplete block")
			}
		}
	case "close":
		if err := required(c.Reason); err != nil {
			return err
		}
		if c.Outcome != "completed" && c.Outcome != "abandoned" {
			return workInput("outcome must be completed or abandoned")
		}
	case "bind":
		if err := required(c.Change, c.Reason); err != nil {
			return err
		}
	case "reconcile":
		if err := required(c.Pending, c.Reason); err != nil {
			return err
		}
		if c.ReconcileAction != "archive-rejected" || !viewTimelineObjectID.MatchString(c.Pending) {
			return workInput("invalid reconciliation action or pending identifier")
		}
	}
	if c.Reason != "" && viewer.ValidateTimelineV3Operation(&viewer.TimelineTaskTransitionRequest{TimelineNativeOperation: workValidationOperation(c.Reason), ExpectedTaskVersion: 1, SegmentID: "validate", State: "in_progress"}) != nil {
		return workInput("invalid reason")
	}
	return nil
}

func workValidationOperation(reason string) viewer.TimelineNativeOperation {
	return viewer.TimelineNativeOperation{SchemaVersion: 3, OperationID: "validate", Actor: viewer.TimelineActor{Kind: "native_producer", ProducerID: "validate"}, ExpectedSessionVersion: 1, Reason: reason}
}
func validateViewWorkSource(s viewWorkSource) error {
	return (viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "validate", Tickets: []viewer.TimelineTicketInput{}, Source: viewer.TimelineSource{Host: s.Host, NativeSessionID: s.NativeSessionID, ProjectRoot: string(filepath.Separator)}, Title: "Validate", Summary: "Validate"}).Validate()
}
func validWorkStage(s string) bool {
	r := viewer.TimelineTaskCreateRequest{TimelineNativeOperation: workValidationOperation("Validate"), PlanOrder: 1, TaskKey: "T1", Title: "Validate", Segments: []viewer.TimelineTaskSegmentInput{{PlannedStage: viewer.TimelineStage(s), Order: 1}}}
	return viewer.ValidateTimelineV3Operation(&r) == nil
}
func validateViewWorkTasks(tasks []viewWorkTaskInput) error {
	if len(tasks) == 0 || len(tasks) > viewer.TimelineMaxTasks {
		return errors.New("invalid task count")
	}
	seen := map[string]bool{}
	orders := map[int64]bool{}
	segmentOrders := map[viewer.TimelineStage]map[int64]bool{}
	for _, t := range tasks {
		if seen[t.Key] {
			return errors.New("duplicate task key")
		}
		seen[t.Key] = true
		if orders[t.Order] {
			return errors.New("duplicate plan order")
		}
		orders[t.Order] = true
		r := viewer.TimelineTaskCreateRequest{TimelineNativeOperation: workValidationOperation("Register tasks"), PlanOrder: t.Order, TaskKey: t.Key, Title: t.Title, Summary: t.Summary}
		for _, s := range t.Segments {
			if segmentOrders[s.Stage] == nil {
				segmentOrders[s.Stage] = map[int64]bool{}
			}
			if segmentOrders[s.Stage][s.Order] {
				return errors.New("duplicate stage order")
			}
			segmentOrders[s.Stage][s.Order] = true
			if s.Assignee != nil {
				if err := validateViewWorkSource(*s.Assignee); err != nil {
					return err
				}
			}
			r.Segments = append(r.Segments, viewer.TimelineTaskSegmentInput{PlannedStage: s.Stage, Order: s.Order})
		}
		if err := viewer.ValidateTimelineV3Operation(&r); err != nil {
			return err
		}
	}
	return nil
}
func validateViewWorkMetadata(m viewWorkTaskMetadata) error {
	r := viewer.TimelineTaskPatchRequest{TimelineNativeOperation: workValidationOperation("Update metadata"), ExpectedTaskVersion: 1, PlanOrder: m.Order, Title: m.Title, Summary: m.Summary, Withdrawn: m.Withdrawn, PlanRevisionID: m.PlanRevision, Order: m.SegmentOrder}
	if m.Stage != "" {
		if !validWorkStage(string(m.Stage)) {
			return errors.New("invalid stage")
		}
		r.SegmentID = "validate"
	}
	if m.Assignee != nil {
		if err := validateViewWorkSource(*m.Assignee); err != nil {
			return err
		}
		r.AssignedProducerID = "validate"
	}
	if m.PlanRevision != "" && !viewTimelineObjectID.MatchString(m.PlanRevision) {
		return errors.New("invalid plan reference")
	}
	if m.DecisionReference != "" && !viewTimelineObjectID.MatchString(m.DecisionReference) {
		return errors.New("invalid decision reference")
	}
	if m.Order != nil && m.PlanRevision == "" && m.DecisionReference == "" {
		return errors.New("plan order update requires confirmed plan or decision reference")
	}
	return viewer.ValidateTimelineV3Operation(&r)
}

func writeViewWorkResult(w io.Writer, result viewWorkResult) error {
	if result.WorkID == "" || result.Result == "" || result.SessionVersion < 1 {
		return fmt.Errorf("invalid_result: confirmation missing")
	}
	return writeViewWorkBoundedJSON(w, result)
}
