package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

var viewTimelineSessionID = regexp.MustCompile(`^[a-f0-9]{64}$`)
var viewTimelineObjectID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)

type viewTimelineCommand struct {
	operation, method, path string
	sessionID, objectID     string
	jsonFile, uploadFile    string
}

func isTimelineViewCommand(name string) bool {
	switch name {
	case "contract", "session", "event", "stage", "artifact", "document":
		return true
	}
	return false
}

// runViewTimeline owns the machine interface; it never launches a viewer or
// opens a browser. Input operation IDs belong to the producer, including retries.
func runViewTimeline(args []string, w io.Writer, stdin io.Reader) error {
	if len(args) == 0 {
		return errors.New("use `hive view contract` for timeline usage")
	}
	if args[0] == "contract" {
		fs := flag.NewFlagSet("view contract", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		operation := fs.String("operation", "", "narrow to one protocol operation")
		if err := fs.Parse(args[1:]); err != nil || fs.NArg() != 0 {
			return errors.New("use `hive view contract [--operation NAME]`")
		}
		contract, err := viewer.TimelineProtocolContract(*operation)
		if err != nil {
			return errors.New("unknown timeline operation; run `hive view contract`")
		}
		return viewTimelineOutput(w, contract)
	}
	command, rest, err := parseViewTimelineCommand(args)
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("view timeline", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&command.jsonFile, "json", "", "request JSON file, or - for stdin")
	fs.StringVar(&command.uploadFile, "file", "", "local image or Markdown file")
	options, ok, err := readOnlyOptions(fs, rest)
	if err != nil || !ok {
		if err == nil {
			return errors.New("run `hive view contract` for operation usage")
		}
		return errors.New("invalid timeline flags; run `hive view contract`")
	}
	if command.method == http.MethodGet {
		if command.jsonFile != "" || command.uploadFile != "" {
			return errors.New("timeline reads accept neither --json nor --file")
		}
	} else if command.jsonFile == "" {
		return errors.New("timeline mutations require --json FILE or --json -")
	}
	upload := command.operation == "artifact_add" || command.operation == "document_add"
	if upload != (command.uploadFile != "") {
		return errors.New("only artifact/document add requires --file")
	}
	var input []byte
	if command.method != http.MethodGet {
		input, err = viewTimelineReadInput(command.jsonFile, stdin, viewer.TimelineMaxBody)
		if err != nil {
			return err
		}
		if err = validateViewTimelineInput(command.operation, input); err != nil {
			return fmt.Errorf("invalid timeline input: %s", viewer.TimelineContractCode(err))
		}
	}
	var body io.Reader = bytes.NewReader(input)
	contentType := "application/json"
	if upload {
		limit := int64(20 << 20)
		if command.operation == "document_add" {
			limit = viewer.TimelineMaxDocumentBytes
		}
		data, err := viewTimelineReadFile(command.uploadFile, limit)
		if err != nil {
			return err
		}
		if command.operation == "document_add" {
			if err := viewer.ValidateTimelineDocumentBytes(data); err != nil {
				return errors.New("invalid Markdown bytes; preserve the input and correct its encoding or size")
			}
		}
		var buffer bytes.Buffer
		writer := multipart.NewWriter(&buffer)
		if err := writer.WriteField("metadata", string(input)); err != nil {
			return errors.New("cannot prepare viewer upload")
		}
		// The source path and basename are not server metadata.
		part, err := writer.CreateFormFile("file", "upload")
		if err != nil {
			return errors.New("cannot prepare viewer upload")
		}
		if _, err = part.Write(data); err != nil || writer.Close() != nil {
			return errors.New("cannot prepare viewer upload")
		}
		body, contentType = &buffer, writer.FormDataContentType()
	}
	_, state, err := management.NormalizeOptions(options)
	if err != nil {
		return errors.New("cannot resolve viewer state directory")
	}
	state, err = filepath.Abs(state)
	if err != nil {
		return errors.New("cannot resolve viewer state directory")
	}
	info, err := viewTimelineRunning(newViewPaths(state))
	if err != nil {
		return err
	}
	data, err := viewTimelineHTTP(info, command.method, command.path, contentType, body, viewer.TimelineMaxBody)
	if err != nil {
		return err
	}
	return viewTimelineDecodeOutput(w, command, input, data)
}

func viewTimelineRunning(paths viewPaths) (serverInfo, error) {
	held, lockErr := viewLockHeld(paths)
	info, infoErr := readServerInfo(paths.info)
	if lockErr != nil || !held || infoErr != nil || info.PID <= 0 || info.Port < 1 || info.Port > 65535 || len(info.Token) < 16 {
		return serverInfo{}, errors.New("the viewer is unavailable; start it with `hive view start` and retry the unchanged operation; local inputs were preserved")
	}
	if err := viewTimelineProtocolReady(info, paths.state); err != nil {
		return serverInfo{}, err
	}
	return info, nil
}

func parseViewTimelineCommand(args []string) (viewTimelineCommand, []string, error) {
	var c viewTimelineCommand
	bad := errors.New("invalid timeline command; run `hive view contract`")
	if len(args) < 2 {
		return c, nil, bad
	}
	key := args[0] + "/" + args[1]
	positionals := 1
	switch key {
	case "session/create":
		c.operation, c.method, positionals = "create_session", http.MethodPost, 0
	case "session/get":
		c.operation, c.method = "session_get", http.MethodGet
	case "session/update":
		c.operation, c.method = "session_patch", http.MethodPatch
	case "session/delete":
		c.operation, c.method = "session_delete", http.MethodDelete
	case "event/add":
		c.operation, c.method = "append_event", http.MethodPost
	case "event/update":
		c.operation, c.method, positionals = "event_patch", http.MethodPatch, 2
	case "event/withdraw":
		c.operation, c.method, positionals = "event_withdraw", http.MethodDelete, 2
	case "artifact/add":
		c.operation, c.method = "artifact_add", http.MethodPost
	case "document/add":
		c.operation, c.method = "document_add", http.MethodPost
	case "document/get":
		c.operation, c.method, positionals = "document_get", http.MethodGet, 2
	case "stage/start":
		c.operation, c.method = "stage_start", http.MethodPost
	case "stage/update":
		c.operation, c.method, positionals = "stage_update", http.MethodPatch, 2
	case "stage/snapshot":
		c.operation, c.method, positionals = "document_snapshot", http.MethodPost, 2
	default:
		return c, nil, bad
	}
	if len(args) < 2+positionals {
		return c, nil, bad
	}
	c.path = "/api/timeline/sessions"
	if positionals > 0 {
		c.sessionID = args[2]
		if !viewTimelineSessionID.MatchString(c.sessionID) {
			return c, nil, errors.New("invalid viewer session ID")
		}
		c.path += "/" + c.sessionID
	}
	if positionals > 1 {
		c.objectID = args[3]
		if !viewTimelineObjectID.MatchString(c.objectID) {
			return c, nil, errors.New("invalid viewer event/document ID")
		}
	}
	switch c.operation {
	case "append_event", "stage_start":
		c.path += "/events"
	case "event_patch", "event_withdraw", "stage_update", "document_snapshot":
		c.path += "/events/" + c.objectID
		if c.operation == "document_snapshot" {
			c.path += "/document-snapshots"
		}
	case "artifact_add":
		c.path += "/artifacts"
	case "document_add":
		c.path += "/documents"
	case "document_get":
		c.path += "/documents/" + c.objectID
	}
	return c, args[2+positionals:], nil
}

func validateViewTimelineInput(operation string, raw []byte) error {
	switch operation {
	case "create_session":
		_, err := viewer.DecodeTimelineCreateSession(raw)
		return err
	case "session_patch":
		_, err := viewer.DecodeTimelineSessionPatch(raw)
		return err
	case "session_delete", "event_withdraw":
		_, err := viewer.DecodeTimelineMutation(raw)
		return err
	case "event_patch", "stage_update":
		r, err := viewer.DecodeTimelineEventPatch(raw)
		if err == nil && operation == "stage_update" && (r.StageState == nil || r.ExpectedSessionVersion == nil) {
			return errors.New("invalid_stage_update")
		}
		return err
	case "append_event", "stage_start":
		r, err := viewer.DecodeTimelineAppendEvent(raw)
		if err == nil && operation == "stage_start" && r.Kind != "stage" {
			return errors.New("invalid_stage_start")
		}
		return err
	case "artifact_add":
		_, err := viewer.DecodeTimelineArtifactMetadata(raw)
		return err
	case "document_add":
		_, err := viewer.DecodeTimelineDocumentMetadata(raw)
		return err
	case "document_snapshot":
		_, err := viewer.DecodeTimelineDocumentSnapshot(raw)
		return err
	}
	return errors.New("unknown_operation")
}

func viewTimelineReadInput(name string, stdin io.Reader, limit int) ([]byte, error) {
	if name != "-" {
		return viewTimelineReadFile(name, int64(limit))
	}
	if stdin == nil {
		return nil, errors.New("timeline JSON stdin is unavailable")
	}
	data, err := io.ReadAll(io.LimitReader(stdin, int64(limit)+1))
	if err != nil || len(data) == 0 || len(data) > limit {
		return nil, errors.New("cannot read bounded timeline JSON input")
	}
	return data, nil
}

func viewTimelineReadFile(name string, limit int64) ([]byte, error) {
	before, err := os.Lstat(name)
	if err != nil || !before.Mode().IsRegular() || before.Size() > limit {
		return nil, errors.New("viewer input must be a bounded regular file, not a symlink; local inputs were preserved")
	}
	file, err := os.Open(name)
	if err != nil {
		return nil, errors.New("cannot read viewer input; local inputs were preserved")
	}
	defer file.Close()
	after, err := file.Stat()
	if err != nil || !os.SameFile(before, after) {
		return nil, errors.New("viewer input changed while opening; local inputs were preserved")
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil || int64(len(data)) > limit {
		return nil, errors.New("cannot read bounded viewer input; local inputs were preserved")
	}
	return data, nil
}

// Redirects and transport/body diagnostics must not reveal the access token.
func viewTimelineHTTP(info serverInfo, method, path, contentType string, body io.Reader, limit int) ([]byte, error) {
	if !strings.HasPrefix(path, "/api/") || strings.ContainsAny(path, "?#\\") {
		return nil, errors.New("invalid local viewer route")
	}
	req, err := http.NewRequest(method, info.url()+path, body)
	if err != nil {
		return nil, errors.New("cannot prepare viewer request")
	}
	req.Header.Set("X-Hive-Token", info.Token)
	if method != http.MethodGet {
		req.Header.Set("Origin", info.url())
		req.Header.Set("X-Hive-Action-Scope", "session")
		req.Header.Set("Content-Type", contentType)
	}
	client := *viewClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.New("viewer request outcome is uncertain; preserve inputs and retry the unchanged operation ID")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("viewer request failed (HTTP %d); preserve inputs, read current state and retry only with the appropriate original operation ID", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, int64(limit)+1))
	if err != nil || len(data) > limit {
		return nil, errors.New("viewer response is incomplete or oversized; preserve inputs and reconcile the unchanged operation ID")
	}
	return data, nil
}

func viewTimelineProtocolReady(info serverInfo, state string) error {
	raw, err := viewTimelineHTTP(info, http.MethodGet, "/api/health", "", nil, viewInfoLimit)
	var health struct {
		TimelineProtocolVersion int    `json:"timeline_protocol_version"`
		PID                     int    `json:"pid"`
		StateDir                string `json:"state_dir"`
	}
	if err != nil || json.Unmarshal(raw, &health) != nil || health.PID != info.PID || filepath.Clean(health.StateDir) != filepath.Clean(state) {
		return errors.New("cannot confirm viewer health and identity; preserve local inputs and check the running viewer")
	}
	if health.TimelineProtocolVersion != viewer.TimelineSchemaVersion {
		return errors.New("the running viewer lacks the supported timeline protocol; update its binary and restart it explicitly, then retry the unchanged operation")
	}
	return nil
}

func viewTimelineDecodeOutput(w io.Writer, command viewTimelineCommand, input, raw []byte) error {
	invalid := errors.New("invalid or mismatched viewer response; preserve inputs and reconcile the unchanged operation ID")
	switch command.operation {
	case "create_session", "session_get":
		response, err := viewer.DecodeTimelineSessionResponse(raw)
		if err != nil || (command.sessionID != "" && response.Session.ViewerSessionID != command.sessionID) {
			return invalid
		}
		if command.operation == "create_session" {
			request, err := viewer.DecodeTimelineCreateSession(input)
			if err != nil || response.Session.Source.Host != request.Source.Host || response.Session.Source.NativeSessionID != request.Source.NativeSessionID {
				return invalid
			}
		}
		// Browser image capabilities are credentials, not durable CLI receipts.
		response.ArtifactURLs = nil
		return viewTimelineOutput(w, response)
	case "document_get":
		response, err := viewer.DecodeTimelineDocumentResponse(raw)
		if err != nil || response.Document.SessionID != command.sessionID || response.Document.DocumentID != command.objectID {
			return invalid
		}
		return viewTimelineOutput(w, response)
	default:
		result, err := viewer.DecodeTimelineOperationResult(raw)
		var request struct {
			OperationID string `json:"operation_id"`
		}
		if err != nil || json.Unmarshal(input, &request) != nil || result.ViewerSessionID != command.sessionID || result.OperationID != request.OperationID {
			return invalid
		}
		kind := command.operation
		if kind == "stage_start" {
			kind = "append_event"
		}
		if kind == "stage_update" {
			kind = "event_patch"
		}
		if string(result.Kind) != kind || (command.objectID != "" && result.EventID != command.objectID) {
			return invalid
		}
		return viewTimelineOutput(w, result)
	}
}

func viewTimelineOutput(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
