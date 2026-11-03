package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"tricell-hive/tooling/viewer"
)

// The complete semantic intent and selected bytes are persisted before any
// remote mutation. Phase/Step advance only after a validated receipt is saved;
// an interruption between those writes safely reuses the preceding exact step.
type workPublicationIntent struct {
	Command        viewWorkCommand `json:"command"`
	Documents      []workDocument  `json:"documents"`
	Phase          string          `json:"phase"`
	Step           int             `json:"step"`
	StageID        string          `json:"stage_id"`
	StageVersion   int64           `json:"stage_version"`
	SessionVersion int64           `json:"session_version"`
	Predecessor    string          `json:"predecessor"`
	DocumentIDs    []string        `json:"document_ids"`
	DocumentIndex  int             `json:"document_index"`
	Changed        int             `json:"changed"`
	RevisionID     string          `json:"revision_id"`
	OperationIDs   []string        `json:"operation_ids,omitempty"`
}

func (e *viewWorkEngine) publish(c viewWorkCommand, w io.Writer) error {
	// The caller may still use its original creation selectors after the
	// work ID was saved internally. Recover that exact frozen intent before
	// reading changed or deleted source files.
	if c.WorkID == "" {
		// Canonicalize directory aliases without rereading any selected document.
		// The source directory may survive while its files changed or disappeared.
		if canonical, err := canonicalWorkDir(c.Project); err == nil {
			c.Project = canonical
		}
		if canonical, err := canonicalWorkDir(c.Change); err == nil {
			c.Change = canonical
		}
		absoluteChange, changeErr := filepath.Abs(c.Change)
		for _, pending := range e.journal.Pending {
			if pending.Action != "publish" || string(pending.Source.Host) != c.Host || pending.Source.NativeSessionID != c.NativeSessionID {
				continue
			}
			var original workPublicationIntent
			if json.Unmarshal(pending.Context, &original) != nil {
				return workStateError("invalid_publication_intent")
			}
			if changeErr == nil && filepath.Clean(absoluteChange) == original.Command.Change && original.Command.Key == c.Key && c.Project == original.Command.Project {
				return e.resumePublish(c, pending, w)
			}
		}
	}
	change, err := canonicalWorkDir(c.Change)
	if err != nil {
		return err
	}
	c.Change = change
	// Collection is a preflight: no create, upload or stage mutation precedes it.
	docs, err := collectWorkDocuments(change, c.Stage, c.Documents, c.ExplicitDocuments || c.Stage != "plan")
	if err != nil {
		return &viewWorkError{Code: "document_preflight", Field: err.Error(), Next: "correct the selected source files before retrying"}
	}
	s, found, err := e.resolvePublication(c)
	if err != nil {
		return err
	}
	if found {
		c.WorkID = s.ViewerSessionID
		if err = e.bind(c, s, false); err != nil {
			return err
		}
		b := e.binding(c.WorkID, c.Host, c.NativeSessionID)
		if b == nil {
			return workStateError("binding_required")
		}
		c.Project = b.Project
	} else {
		c.Project, err = canonicalWorkDir(c.Project)
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(c.Project, c.Change)
		if relErr != nil || rel == ".." || strings.HasPrefix(rel, "../") {
			return workInput("change must be within project")
		}
	}
	for _, pending := range e.journal.Pending {
		if pending.WorkID == c.WorkID && c.WorkID != "" {
			return &viewWorkError{Code: "pending_operation", Pending: pending.ID, Next: "resume or reconcile the original pending publication"}
		}
		if pending.Source.Host == viewer.TimelineHost(c.Host) && pending.Source.NativeSessionID == c.NativeSessionID {
			var original workPublicationIntent
			if pending.Action == "publish" && json.Unmarshal(pending.Context, &original) == nil && original.Command.Change == c.Change {
				if original.Command.Key != "" && c.Key != "" && original.Command.Key != c.Key {
					return workStateError("binding_mismatch")
				}
				// Creation can have committed without returning an ID. Reissuing its
				// explicit semantic selector resumes the frozen intent, never a new create.
				if pending.WorkID == "" {
					return e.resumePublish(c, pending, w)
				}
				return &viewWorkError{Code: "pending_operation", Pending: pending.ID, Next: "resume the original pending publication"}
			}
		}
	}
	p := workPublicationIntent{Command: c, Documents: docs, Phase: "create", DocumentIDs: make([]string, len(docs))}
	if found {
		actor, err := e.actor(s, c)
		if err != nil {
			return err
		}
		if err = viewer.ValidateTimelineAuthority(s, "document_snapshot", actor, ""); err != nil {
			return workStateError("publication_forbidden")
		}
		stage, err := publicationStage(s, c.Stage)
		if err != nil {
			return err
		}
		if s.State == viewer.TimelineClosed || s.State == viewer.TimelineFailed {
			return workStateError("work_closed")
		}
		p.StageID, p.StageVersion, p.SessionVersion = stage.EventID, stage.Version, s.Version
		previous, previousStage := publicationTip(s, c.DocumentKey)
		if previous != nil {
			p.Predecessor = previous.DocumentRevisionID
			if previousStage.EventID != stage.EventID && (previousStage.Stage != stage.Stage || stage.RelatedEventID != previousStage.EventID) {
				return workStateError("stage_chain_mismatch")
			}
			if samePublication(s, *previous, previousStage.EventID, stage.EventID, c, docs) {
				if err = e.verifyPublicationDocuments(s, previous.Tabs, docs); err != nil {
					return err
				}
				count, changed := len(docs), 0
				return writeViewWorkResult(w, viewWorkResult{WorkID: s.ViewerSessionID, Result: "unchanged", SessionVersion: s.Version, RevisionID: previous.DocumentRevisionID, StageID: stage.EventID, DocumentCount: &count, ChangedCount: &changed})
			}
		}
		if err = publicationCapacity(s, docs); err != nil {
			return err
		}
		p.Changed = publicationChanged(s, previous, docs)
		p.Phase = "upload"
	}
	source := viewer.TimelineSource{ProjectRoot: c.Project, Host: viewer.TimelineHost(c.Host), NativeSessionID: c.NativeSessionID}
	intent, err := e.prepareIntent("publish", c.WorkID, source, p)
	if err != nil {
		return err
	}
	return publicationPendingError(e.continuePublication(intent, &p, w), intent.ID)
}

// Resolve without creating: the frozen publish intent must precede creation.
func (e *viewWorkEngine) resolvePublication(c viewWorkCommand) (viewer.TimelineSession, bool, error) {
	if c.WorkID != "" {
		s, err := e.session(c.WorkID)
		return s, err == nil, err
	}
	project, err := canonicalWorkDir(c.Project)
	if err != nil {
		return viewer.TimelineSession{}, false, err
	}
	source := viewer.TimelineSource{ProjectRoot: project, Host: viewer.TimelineHost(c.Host), NativeSessionID: c.NativeSessionID}
	for _, b := range e.journal.Bindings {
		if b.Source == source && b.Change == c.Change {
			if b.Key != c.Key {
				return viewer.TimelineSession{}, false, workStateError("binding_mismatch")
			}
			s, err := e.session(b.WorkID)
			return s, err == nil, err
		}
	}
	index, err := viewIndex(e.info)
	if err != nil {
		return viewer.TimelineSession{}, false, err
	}
	for _, entry := range index.TimelineSessions {
		if entry.Source == source {
			s, err := e.session(entry.ViewerSessionID)
			if err != nil {
				return s, false, err
			}
			if s.WorkKey != nil && *s.WorkKey == c.Key {
				return s, true, nil
			}
		}
	}
	return viewer.TimelineSession{}, false, nil
}
func publicationStage(s viewer.TimelineSession, stage string) (viewer.TimelineEvent, error) {
	var selected viewer.TimelineEvent
	for _, event := range s.Events {
		if event.Kind == "stage" && !event.Withdrawn && string(event.Stage) == stage {
			selected = event
		}
	}
	if selected.EventID == "" {
		return selected, workStateError("stage_required")
	}
	if selected.StageState == "completed" || selected.StageState == "failed" {
		return selected, workStateError("stage_reentry_required")
	}
	return selected, nil
}
func publicationTip(s viewer.TimelineSession, key string) (*viewer.TimelineDocumentSnapshot, viewer.TimelineEvent) {
	var tip *viewer.TimelineDocumentSnapshot
	var stage viewer.TimelineEvent
	for _, event := range s.Events {
		for _, snapshot := range event.DocumentSnapshots {
			if snapshot.DocumentKey == key {
				copy := snapshot
				tip = &copy
				stage = event
			}
		}
	}
	return tip, stage
}
func samePublication(s viewer.TimelineSession, tip viewer.TimelineDocumentSnapshot, tipStage, stage string, c viewWorkCommand, docs []workDocument) bool {
	if tipStage != stage || tip.DocumentKey != c.DocumentKey || tip.Status != c.Status || tip.TaskID != "" || tip.SegmentID != "" || len(tip.Tabs) != len(docs) {
		return false
	}
	for n, tab := range tip.Tabs {
		if tab.Label != docs[n].Label {
			return false
		}
		found := false
		for _, stored := range s.Documents {
			if stored.DocumentID == tab.DocumentID && stored.SHA256 == docs[n].SHA256 && stored.Size == docs[n].Size {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func publicationChanged(s viewer.TimelineSession, tip *viewer.TimelineDocumentSnapshot, docs []workDocument) int {
	if tip == nil {
		return len(docs)
	}
	changed := 0
	for n, doc := range docs {
		same := false
		if n < len(tip.Tabs) && tip.Tabs[n].Label == doc.Label {
			for _, stored := range s.Documents {
				if stored.DocumentID == tip.Tabs[n].DocumentID && stored.SHA256 == doc.SHA256 && stored.Size == doc.Size {
					same = true
				}
			}
		}
		if !same {
			changed++
		}
	}
	if len(tip.Tabs) > len(docs) {
		changed += len(tip.Tabs) - len(docs)
	}
	return changed
}
func publicationCapacity(s viewer.TimelineSession, docs []workDocument) error {
	count, total := len(s.Documents), int64(0)
	hashes := map[string]bool{}
	for _, doc := range s.Documents {
		hashes[doc.SHA256] = true
		total += doc.Size
	}
	for _, doc := range docs {
		if !hashes[doc.SHA256] {
			count++
			total += doc.Size
			hashes[doc.SHA256] = true
		}
	}
	snapshots := 0
	for _, event := range s.Events {
		snapshots += len(event.DocumentSnapshots)
	}
	if count > viewer.TimelineMaxDocuments || total > viewer.TimelineMaxAggregateDocumentBytes || snapshots >= viewer.TimelineMaxDocumentSnapshots {
		return workStateError("publication_capacity")
	}
	return nil
}
func (e *viewWorkEngine) resumePublish(c viewWorkCommand, i *viewWorkIntent, w io.Writer) error {
	var p workPublicationIntent
	decoder := json.NewDecoder(bytes.NewReader(i.Context))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&p) != nil || decoder.Decode(new(any)) != io.EOF || len(p.Documents) < 1 || len(p.Documents) > workDocumentLimit || len(p.DocumentIDs) != len(p.Documents) || p.Command.Host != c.Host || p.Command.NativeSessionID != c.NativeSessionID || p.Command.Action != "publish" || !validWorkStage(p.Command.Stage) || (p.Command.Status != "draft" && p.Command.Status != "complete") || p.Step < 0 || p.Step > len(i.Steps) || p.DocumentIndex < 0 || p.DocumentIndex > len(p.Documents) {
		return workStateError("invalid_publication_intent")
	}
	for _, doc := range p.Documents {
		digest := sha256.Sum256(doc.Bytes)
		if viewer.ValidateTimelineDocumentBytes(doc.Bytes) != nil || int64(len(doc.Bytes)) != doc.Size || hex.EncodeToString(digest[:]) != doc.SHA256 {
			return workStateError("invalid_publication_intent")
		}
	}
	return publicationPendingError(e.continuePublication(i, &p, w), i.ID)
}
func (e *viewWorkEngine) savePublication(i *viewWorkIntent, p *workPublicationIntent) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	i.Context = raw
	return e.save()
}

// A step may have been appended immediately before a crash. Its saved body is
// authoritative; rebuilding the semantic request must never replace those bytes.
func (e *viewWorkEngine) publicationStep(i *viewWorkIntent, p *workPublicationIntent, method, path, ct string, body []byte) ([]byte, error) {
	if len(i.Steps) == p.Step {
		if _, err := e.appendStep(i, method, path, ct, body); err != nil {
			return nil, err
		}
	}
	if p.Step >= len(i.Steps) || i.Steps[p.Step].Method != method || i.Steps[p.Step].Path != path {
		return nil, workStateError("invalid_publication_step")
	}
	return e.executeStep(i, p.Step)
}
func (e *viewWorkEngine) continuePublication(i *viewWorkIntent, p *workPublicationIntent, w io.Writer) error {
	c := p.Command
	for {
		// Resolve an already prepared request before re-evaluating mutable remote
		// state. A committed upload/stage/snapshot still has its original receipt
		// even when another actor changes lifecycle or authority after the response.
		if p.Step < len(i.Steps) && (p.Phase == "stage" || p.Phase == "upload" || p.Phase == "snapshot") {
			path := "/api/timeline/sessions/" + i.WorkID
			kind := ""
			switch p.Phase {
			case "stage":
				path += "/events"
				kind = "append_event"
			case "upload":
				path += "/documents"
				kind = "document_add"
			case "snapshot":
				path += "/events/" + p.StageID + "/document-snapshots"
				kind = "document_snapshot"
			}
			if i.Steps[p.Step].Method != "POST" || i.Steps[p.Step].Path != path {
				return workStateError("invalid_publication_step")
			}
			raw, err := e.executeStep(i, p.Step)
			if err != nil {
				return err
			}
			result, err := publicationOperationResult(raw, kind)
			if err != nil {
				return err
			}
			p.SessionVersion = result.SessionVersion
			switch p.Phase {
			case "stage":
				p.StageID, p.StageVersion = result.EventID, result.EventVersion
				p.Phase = "upload"
			case "upload":
				if p.DocumentIndex >= len(p.DocumentIDs) {
					return workStateError("invalid_publication_intent")
				}
				p.DocumentIDs[p.DocumentIndex] = result.DocumentID
				p.DocumentIndex++
			case "snapshot":
				p.RevisionID = result.DocumentRevisionID
				p.Phase = "verify"
			}
			p.Step++
			if err = e.savePublication(i, p); err != nil {
				return err
			}
			continue
		}
		switch p.Phase {
		case "create":
			req := viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: workID("op-"), Source: i.Source, WorkKey: &c.Key, Title: c.Title, Summary: c.Summary, Tickets: []viewer.TimelineTicketInput{}}
			raw, err := json.Marshal(req)
			if err != nil {
				return err
			}
			response, err := e.publicationStep(i, p, "POST", "/api/timeline/sessions", "application/json", raw)
			if err != nil {
				return err
			}
			created, err := viewer.DecodeTimelineSessionResponse(response)
			if err != nil {
				return workStateError("invalid_response")
			}
			c.WorkID = created.Session.ViewerSessionID
			p.Command.WorkID = c.WorkID
			p.SessionVersion = created.Session.Version
			if err = e.bind(c, created.Session, false); err != nil {
				return err
			}
			// A reused remote work must already have the requested stage: only an
			// actually empty first creation is allowed to start one implicitly.
			if len(created.Session.Events) > 0 {
				stage, err := publicationStage(created.Session, c.Stage)
				if err != nil {
					return err
				}
				p.StageID, p.StageVersion = stage.EventID, stage.Version
				tip, _ := publicationTip(created.Session, c.DocumentKey)
				if tip != nil {
					p.Predecessor = tip.DocumentRevisionID
				}
				p.Phase = "upload"
			} else {
				p.Phase = "stage"
			}
			p.Changed = len(p.Documents)
			p.Step++
			if err = e.savePublication(i, p); err != nil {
				return err
			}
		case "stage":
			s, err := e.session(i.WorkID)
			if err != nil {
				return err
			}
			actor, err := e.actor(s, c)
			if err != nil {
				return err
			}
			req := viewer.TimelineAppendEventRequest{SchemaVersion: 3, OperationID: workID("op-"), Actor: actor, Reason: c.Reason, Kind: "stage", Title: "", Summary: c.Summary, TicketIDs: []string{}, Stage: viewer.TimelineStage(c.Stage), StageState: "in_progress", ExpectedSessionVersion: &p.SessionVersion}
			raw, err := json.Marshal(req)
			if err != nil {
				return err
			}
			response, err := e.publicationStep(i, p, "POST", "/api/timeline/sessions/"+i.WorkID+"/events", "application/json", raw)
			if err != nil {
				return err
			}
			result, err := publicationOperationResult(response, "append_event")
			if err != nil {
				return workStateError("invalid_response")
			}
			p.StageID, p.StageVersion, p.SessionVersion = result.EventID, result.EventVersion, result.SessionVersion
			p.Phase = "upload"
			p.Step++
			if err = e.savePublication(i, p); err != nil {
				return err
			}
		case "upload":
			if p.DocumentIndex == len(p.Documents) {
				p.Phase = "snapshot"
				if err := e.savePublication(i, p); err != nil {
					return err
				}
				continue
			}
			doc := p.Documents[p.DocumentIndex]
			s, err := e.session(i.WorkID)
			if err != nil {
				return err
			}
			actor, err := e.actor(s, c)
			if err != nil {
				return err
			}
			if s.State == viewer.TimelineClosed || s.State == viewer.TimelineFailed {
				return &viewWorkError{Code: "work_closed", Pending: i.ID, Next: "preserve the pending publication; closed work cannot be republished"}
			}
			reused := ""
			// If a request already exists, replay it even when GET now shows the blob.
			if len(i.Steps) == p.Step {
				for _, stored := range s.Documents {
					if stored.SHA256 == doc.SHA256 && stored.Size == doc.Size {
						reused = stored.DocumentID
						break
					}
				}
			}
			if reused != "" {
				if err = e.verifyPublicationDocuments(s, []viewer.TimelineDocumentTab{{Label: doc.Label, DocumentID: reused}}, []workDocument{doc}); err != nil {
					return err
				}
				p.DocumentIDs[p.DocumentIndex] = reused
				p.DocumentIndex++
				if err = e.savePublication(i, p); err != nil {
					return err
				}
				continue
			}
			metadata := viewer.TimelineDocumentMetadata{SchemaVersion: 3, OperationID: workID("op-"), Actor: actor, MediaType: "text/markdown"}
			if err = viewer.ValidateTimelineDocumentUpload(s, metadata, doc.Bytes); err != nil {
				return &viewWorkError{Code: "upload_precondition", Pending: i.ID, Next: "preserve the pending publication and inspect the current work"}
			}
			var body bytes.Buffer
			form := multipart.NewWriter(&body)
			raw, _ := json.Marshal(metadata)
			if err = form.WriteField("metadata", string(raw)); err != nil {
				return err
			}
			part, err := form.CreateFormFile("file", "upload.md")
			if err != nil {
				return err
			}
			if _, err = part.Write(doc.Bytes); err != nil {
				return err
			}
			if err = form.Close(); err != nil {
				return err
			}
			response, err := e.publicationStep(i, p, "POST", "/api/timeline/sessions/"+i.WorkID+"/documents", form.FormDataContentType(), body.Bytes())
			if err != nil {
				return err
			}
			result, err := publicationOperationResult(response, "document_add")
			if err != nil {
				return workStateError("invalid_response")
			}
			p.DocumentIDs[p.DocumentIndex] = result.DocumentID
			p.SessionVersion = result.SessionVersion
			p.DocumentIndex++
			p.Step++
			if err = e.savePublication(i, p); err != nil {
				return err
			}
		case "snapshot":
			s, err := e.session(i.WorkID)
			if err != nil {
				return err
			}
			actor, err := e.actor(s, c)
			if err != nil {
				return err
			}
			tabs := make([]viewer.TimelineDocumentTab, len(p.Documents))
			for n, doc := range p.Documents {
				tabs[n] = viewer.TimelineDocumentTab{Label: doc.Label, DocumentID: p.DocumentIDs[n]}
			}
			req := viewer.TimelineDocumentSnapshotRequest{SchemaVersion: 3, OperationID: workID("op-"), Actor: actor, ExpectedVersion: p.StageVersion, ExpectedSessionVersion: p.SessionVersion, PredecessorRevisionID: p.Predecessor, DocumentKey: c.DocumentKey, Status: c.Status, Tabs: tabs}
			raw, err := json.Marshal(req)
			if err != nil {
				return err
			}
			response, err := e.publicationStep(i, p, "POST", "/api/timeline/sessions/"+i.WorkID+"/events/"+p.StageID+"/document-snapshots", "application/json", raw)
			if err != nil {
				return err
			}
			result, err := publicationOperationResult(response, "document_snapshot")
			if err != nil {
				return workStateError("invalid_response")
			}
			p.RevisionID = result.DocumentRevisionID
			p.SessionVersion = result.SessionVersion
			p.Phase = "verify"
			p.Step++
			if err = e.savePublication(i, p); err != nil {
				return err
			}
		case "verify":
			s, err := e.session(i.WorkID)
			if err != nil {
				return err
			}
			var snapshot *viewer.TimelineDocumentSnapshot
			for _, event := range s.Events {
				if event.EventID == p.StageID {
					for _, candidate := range event.DocumentSnapshots {
						if candidate.DocumentRevisionID == p.RevisionID {
							copy := candidate
							snapshot = &copy
						}
					}
				}
			}
			if snapshot == nil || snapshot.PredecessorRevisionID != p.Predecessor || !samePublication(s, *snapshot, p.StageID, p.StageID, c, p.Documents) {
				return &viewWorkError{Code: "readback_mismatch", Pending: i.ID, Next: "resume the frozen publication to verify its exact revision"}
			}
			if err = e.verifyPublicationDocuments(s, snapshot.Tabs, p.Documents); err != nil {
				return err
			}
			// Only verified bytes are reproducible scratch. Commit the compact
			// confirmed history in one journal write, so a crash never leaves a
			// pending intent whose frozen inputs were already reclaimed.
			confirmed := *i
			retained := *p
			retained.Documents = append([]workDocument(nil), p.Documents...)
			retained.OperationIDs = []string{}
			confirmed.Steps = append([]viewWorkStep(nil), i.Steps...)
			for n := range confirmed.Steps {
				retained.OperationIDs = append(retained.OperationIDs, workStepOperationID(confirmed.Steps[n]))
				confirmed.Steps[n].Body = nil
			}
			for n := range retained.Documents {
				retained.Documents[n].Bytes = nil
			}
			confirmed.Context, err = json.Marshal(retained)
			if err != nil {
				return err
			}
			if err = e.finishIntent(&confirmed); err != nil {
				return err
			}
			count := len(p.Documents)
			return writeViewWorkResult(w, viewWorkResult{WorkID: i.WorkID, Result: "published", SessionVersion: p.SessionVersion, RevisionID: p.RevisionID, StageID: p.StageID, DocumentCount: &count, ChangedCount: &p.Changed})
		default:
			return workStateError("invalid_publication_phase")
		}
	}
}
func (e *viewWorkEngine) verifyPublicationDocuments(s viewer.TimelineSession, tabs []viewer.TimelineDocumentTab, docs []workDocument) error {
	if len(tabs) != len(docs) {
		return workStateError("readback_mismatch")
	}
	for n, tab := range tabs {
		raw, err := viewTimelineHTTP(e.info, "GET", "/api/timeline/sessions/"+s.ViewerSessionID+"/documents/"+tab.DocumentID, "", nil, viewer.TimelineMaxBody)
		if err != nil {
			return workStateError("document_readback_unavailable")
		}
		response, err := viewer.DecodeTimelineDocumentResponse(raw)
		if err != nil || response.Document.DocumentID != tab.DocumentID || response.Document.SessionID != s.ViewerSessionID || response.Document.SHA256 != docs[n].SHA256 || response.Document.Size != docs[n].Size || !bytes.Equal([]byte(response.Content), docs[n].Bytes) {
			return workStateError("readback_mismatch")
		}
	}
	return nil
}

func publicationPendingError(err error, pending string) error {
	if err == nil {
		return nil
	}
	if failure, ok := err.(*viewWorkError); ok {
		copy := *failure
		if copy.Pending == "" {
			copy.Pending = pending
		}
		return &copy
	}
	return &viewWorkError{Code: "publication_failed", Pending: pending, Next: "preserve publisher state; resume the frozen publication"}
}

func publicationOperationResult(raw []byte, kind string) (viewer.TimelineOperationResult, error) {
	result, err := viewer.DecodeTimelineOperationResult(raw)
	if err != nil || string(result.Kind) != kind {
		return result, workStateError("invalid_response")
	}
	return result, nil
}
