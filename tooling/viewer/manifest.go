package viewer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Limits of the walkthrough.json contract (design DS1). They match the
// constants of content/skills/flow-walkthrough/scripts/render_walkthrough.py.
const (
	manifestFile  = "walkthrough.json"
	maxManifest   = 1 << 20
	maxCapture    = 20 << 20
	maxRecording  = 200 << 20
	maxFlows      = 50
	maxCaptures   = 100
	manifestField = "walkthrough.json"
)

var (
	flowIDPattern  = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)
	segmentPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	langPattern    = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{1,8})*$`)
	drivePattern   = regexp.MustCompile(`^[A-Za-z]:`)

	captureExtensions   = []string{".jpeg", ".jpg", ".png", ".webp"}
	recordingExtensions = []string{".mp4"}
	checkStatuses       = []string{"fail", "info", "not_run", "pass", "warn"}
	viewports           = []string{"desktop", "mobile"}
	kinds               = []string{"final", "first_view"}
)

// recordingRules maps a recording status to its allowed and required keys.
var recordingRules = map[string]struct{ allowed, required []string }{
	"none":         {[]string{"status"}, nil},
	"recorded":     {[]string{"status", "path", "duration_seconds", "candidate"}, []string{"path", "duration_seconds", "candidate"}},
	"not_recorded": {[]string{"status", "reason"}, []string{"reason"}},
	"stale":        {[]string{"status", "path", "candidate"}, []string{"path", "candidate"}},
}

// ManifestError is a validation failure that names the offending field.
type ManifestError struct {
	Field   string
	Message string
}

func (e *ManifestError) Error() string { return e.Field + ": " + e.Message }

func fail(field, format string, args ...any) *ManifestError {
	return &ManifestError{Field: cleanText(field, maxErrorText), Message: cleanText(fmt.Sprintf(format, args...), maxErrorText)}
}

// maxErrorText bounds the field name and the reason of an error, which can
// echo attacker-chosen manifest content (H4).
const maxErrorText = 200

// cleanText drops control and invalid characters and cuts s to max runes,
// marking the cut with an ellipsis.
func cleanText(s string, max int) string {
	var b strings.Builder
	n := 0
	for _, r := range s {
		if r == utf8.RuneError || unicode.IsControl(r) {
			continue
		}
		if n == max {
			b.WriteString("…")
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

// Manifest is a validated walkthrough.json together with the facts the
// store needs. Raw holds the exact bytes that were validated, so a copy and
// its content hash never disagree with what was checked.
type Manifest struct {
	Raw           []byte
	Lang          string
	Kind          string
	Title         string
	Ticket        string
	CreatedAt     string
	SessionID     string
	SessionHost   string
	SessionStatus string
	// Files lists every capture and recording path, slash-separated,
	// de-duplicated and sorted. All of them were checked to be regular files
	// inside the folder, with no symbolic link on the way.
	Files []string
}

// ValidateManifest reads <folder>/walkthrough.json and validates it against
// the version 1 contract, exactly as render_walkthrough.py does: strict
// fields and types, safe relative paths (no absolute path, "..", odd
// segment or symbolic link), allowed extensions, size limits and existing
// referenced files. Every filesystem access goes through an os.Root on
// folder, so a link cannot lead outside it. The error is a *ManifestError
// for any contract violation.
func ValidateManifest(folder string) (*Manifest, error) {
	root, err := os.OpenRoot(folder)
	if err != nil {
		return nil, fail("folder", "cannot open: %v", err)
	}
	defer root.Close()
	return validateRoot(root)
}

// validateRoot is ValidateManifest on a folder that is already open. The
// ingestion opens each folder once and passes that same root to validation,
// hashing and copying, so a folder swapped for a link in between changes
// nothing (H1).
func validateRoot(root *os.Root) (*Manifest, error) {
	info, err := root.Lstat(manifestFile)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, fail(manifestField, "file not found in the folder")
	case err != nil:
		return nil, fail(manifestField, "cannot be read: %v", err)
	case !info.Mode().IsRegular():
		return nil, fail(manifestField, "must be a regular file, not a link")
	case info.Size() > maxManifest:
		return nil, fail(manifestField, "exceeds 1 MiB")
	}
	f, err := root.Open(manifestFile)
	if err != nil {
		return nil, fail(manifestField, "cannot be read: %v", err)
	}
	defer f.Close()
	info, err = f.Stat()
	if err != nil {
		return nil, fail(manifestField, "cannot inspect opened file: %v", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fail(manifestField, "must be a regular file, not a link")
	}
	if info.Size() > maxManifest {
		return nil, fail(manifestField, "exceeds 1 MiB")
	}
	if err := validateSingleLink(info); err != nil {
		return nil, fail(manifestField, "%v", err)
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxManifest+1))
	if err != nil {
		return nil, fail(manifestField, "cannot be read: %v", err)
	}
	if len(raw) > maxManifest {
		return nil, fail(manifestField, "exceeds 1 MiB")
	}
	if !utf8.Valid(raw) {
		return nil, fail(manifestField, "is not valid UTF-8")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return nil, fail(manifestField, "is not valid JSON: %v", err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fail(manifestField, "is not valid JSON: trailing data")
	}
	if _, ok := doc.(map[string]any); !ok {
		return nil, fail(manifestField, "must contain a JSON object")
	}
	v := &validator{root: root, files: map[string]bool{}}
	m, err := v.manifest(doc)
	if err != nil {
		return nil, err
	}
	m.Raw = raw
	for p := range v.files {
		m.Files = append(m.Files, p)
	}
	sort.Strings(m.Files)
	return m, nil
}

type validator struct {
	root  *os.Root
	files map[string]bool
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func isOneOf(value any, allowed []string) bool {
	s, ok := value.(string)
	return ok && contains(allowed, s)
}

func qualify(field, key string) string {
	if field == "" {
		return key
	}
	return field + "." + key
}

func object(value any, field string, required, optional []string) (map[string]any, *ManifestError) {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, fail(field, "must be an object")
	}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !contains(required, k) && !contains(optional, k) {
			return nil, fail(qualify(field, k), "unknown field")
		}
	}
	sorted := append([]string(nil), required...)
	sort.Strings(sorted)
	for _, k := range sorted {
		if _, ok := obj[k]; !ok {
			return nil, fail(qualify(field, k), "required field is missing")
		}
	}
	return obj, nil
}

func text(value any, field string) (string, *ManifestError) {
	s, ok := value.(string)
	if !ok {
		return "", fail(field, "must be a string")
	}
	if strings.TrimSpace(s) == "" {
		return "", fail(field, "must not be empty")
	}
	return s, nil
}

func textList(value any, field string) *ManifestError {
	list, ok := value.([]any)
	if !ok {
		return fail(field, "must be an array of strings")
	}
	for i, item := range list {
		if _, err := text(item, fmt.Sprintf("%s[%d]", field, i)); err != nil {
			return err
		}
	}
	return nil
}

// isoLayouts approximates Python's datetime.fromisoformat for the shapes an
// agent realistically writes; anything else is rejected by both sides or, at
// worst, only by this one (a stricter result is safe).
var isoLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02T15:04Z07:00",
	"2006-01-02T15:04",
	"2006-01-02",
}

func isISODateTime(s string) bool {
	if strings.HasSuffix(s, "z") { // Python accepts a lowercase UTC designator too
		s = s[:len(s)-1] + "Z"
	}
	for _, layout := range isoLayouts {
		if _, err := time.Parse(layout, s); err == nil {
			return true
		}
	}
	return false
}

func (v *validator) manifest(doc any) (*Manifest, error) {
	obj, err := object(doc, "",
		[]string{"schema_version", "lang", "kind", "title", "session", "candidate", "created_at", "flows"},
		[]string{"ticket", "result", "checks", "action_required", "environment", "provenance"})
	if err != nil {
		return nil, err
	}
	m := &Manifest{}
	if n, ok := obj["schema_version"].(float64); !ok || n != 1 {
		return nil, fail("schema_version", "must be 1")
	}
	lang, e := text(obj["lang"], "lang")
	if e != nil {
		return nil, e
	}
	if !langPattern.MatchString(lang) {
		return nil, fail("lang", "must be a language tag such as es or en")
	}
	m.Lang = lang
	if !isOneOf(obj["kind"], kinds) {
		return nil, fail("kind", "must be first_view or final")
	}
	m.Kind = obj["kind"].(string)
	if m.Title, e = text(obj["title"], "title"); e != nil {
		return nil, e
	}
	session, err := object(obj["session"], "session", []string{"id"}, []string{"host", "status"})
	if err != nil {
		return nil, err
	}
	if m.SessionID, e = text(session["id"], "session.id"); e != nil {
		return nil, e
	}
	m.SessionStatus = "unknown"
	if status, ok := session["status"]; ok {
		if !isOneOf(status, []string{"in_progress", "closed"}) {
			return nil, fail("session.status", "must be in_progress or closed")
		}
		m.SessionStatus = status.(string)
	}
	if h, ok := session["host"]; ok {
		if m.SessionHost, e = text(h, "session.host"); e != nil {
			return nil, e
		}
	}
	candidate, err := object(obj["candidate"], "candidate", []string{"branch", "commit"}, []string{"base"})
	if err != nil {
		return nil, err
	}
	for _, k := range []string{"base", "branch", "commit"} {
		if val, ok := candidate[k]; ok {
			if _, e := text(val, "candidate."+k); e != nil {
				return nil, e
			}
		}
	}
	if m.CreatedAt, e = text(obj["created_at"], "created_at"); e != nil {
		return nil, e
	}
	if !isISODateTime(m.CreatedAt) {
		return nil, fail("created_at", "must be an ISO 8601 date-time")
	}
	for _, k := range []string{"ticket", "result"} {
		if val, ok := obj[k]; ok {
			s, e := text(val, k)
			if e != nil {
				return nil, e
			}
			if k == "ticket" {
				m.Ticket = strings.TrimSpace(s)
			}
		}
	}
	if e := v.checks(obj); e != nil {
		return nil, e
	}
	for _, k := range []string{"action_required", "provenance"} {
		if val, ok := obj[k]; ok {
			if e := textList(val, k); e != nil {
				return nil, e
			}
		}
	}
	if val, ok := obj["environment"]; ok {
		list, isList := val.([]any)
		if !isList {
			return nil, fail("environment", "must be an array")
		}
		for i, item := range list {
			field := fmt.Sprintf("environment[%d]", i)
			entry, err := object(item, field, []string{"label", "value"}, nil)
			if err != nil {
				return nil, err
			}
			for _, k := range []string{"label", "value"} {
				if _, e := text(entry[k], field+"."+k); e != nil {
					return nil, e
				}
			}
		}
	}
	flows, ok := obj["flows"].([]any)
	if !ok || len(flows) == 0 {
		return nil, fail("flows", "must be a non-empty array")
	}
	if len(flows) > maxFlows {
		return nil, fail("flows", "at most %d flows", maxFlows)
	}
	seen := map[string]bool{}
	for i, flow := range flows {
		if e := v.flow(flow, i, seen); e != nil {
			return nil, e
		}
	}
	return m, nil
}

func (v *validator) checks(obj map[string]any) *ManifestError {
	val, ok := obj["checks"]
	if !ok {
		return nil
	}
	list, isList := val.([]any)
	if !isList {
		return fail("checks", "must be an array")
	}
	for i, item := range list {
		field := fmt.Sprintf("checks[%d]", i)
		check, err := object(item, field, []string{"name", "status"}, []string{"evidence"})
		if err != nil {
			return err
		}
		if _, e := text(check["name"], field+".name"); e != nil {
			return e
		}
		if !isOneOf(check["status"], checkStatuses) {
			return fail(field+".status", "must be one of %s", strings.Join(checkStatuses, ", "))
		}
		if ev, ok := check["evidence"]; ok {
			if _, e := text(ev, field+".evidence"); e != nil {
				return e
			}
		}
	}
	return nil
}

func (v *validator) flow(value any, index int, seen map[string]bool) *ManifestError {
	field := fmt.Sprintf("flows[%d]", index)
	flow, err := object(value, field, []string{"id", "title", "captures"}, []string{"steps", "recording"})
	if err != nil {
		return err
	}
	id, e := text(flow["id"], field+".id")
	if e != nil {
		return e
	}
	if !flowIDPattern.MatchString(id) {
		return fail(field+".id", "must match %s", flowIDPattern.String())
	}
	if seen[id] {
		return fail(field+".id", "duplicate flow id %q", id)
	}
	seen[id] = true
	if _, e := text(flow["title"], field+".title"); e != nil {
		return e
	}
	if steps, ok := flow["steps"]; ok {
		if e := textList(steps, field+".steps"); e != nil {
			return e
		}
	}
	captures, ok := flow["captures"].([]any)
	if !ok {
		return fail(field+".captures", "must be an array")
	}
	if len(captures) > maxCaptures {
		return fail(field+".captures", "at most %d captures per flow", maxCaptures)
	}
	for n, c := range captures {
		item := fmt.Sprintf("%s.captures[%d]", field, n)
		capture, err := object(c, item, []string{"path", "caption"}, []string{"alt", "viewport"})
		if err != nil {
			return err
		}
		if e := v.path(capture["path"], item+".path", captureExtensions, maxCapture); e != nil {
			return e
		}
		if _, e := text(capture["caption"], item+".caption"); e != nil {
			return e
		}
		if alt, ok := capture["alt"]; ok {
			if _, e := text(alt, item+".alt"); e != nil {
				return e
			}
		}
		if vp, ok := capture["viewport"]; ok && !isOneOf(vp, viewports) {
			return fail(item+".viewport", "must be desktop or mobile")
		}
	}
	if rec, ok := flow["recording"]; ok {
		return v.recording(rec, field+".recording")
	}
	return nil
}

func (v *validator) recording(value any, field string) *ManifestError {
	rec, ok := value.(map[string]any)
	if !ok {
		return fail(field, "must be an object")
	}
	status := rec["status"]
	if s, isStr := status.(string); !isStr || recordingRules[s].allowed == nil {
		return fail(field+".status", "must be one of none, recorded, not_recorded, stale")
	}
	rules := recordingRules[status.(string)]
	keys := make([]string, 0, len(rec))
	for k := range rec {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !contains(rules.allowed, k) {
			return fail(field+"."+k, "unknown field for status %s", status)
		}
	}
	required := append([]string(nil), rules.required...)
	sort.Strings(required)
	for _, k := range required {
		if _, ok := rec[k]; !ok {
			return fail(field+"."+k, "required for status %s", status)
		}
	}
	if p, ok := rec["path"]; ok {
		if e := v.path(p, field+".path", recordingExtensions, maxRecording); e != nil {
			return e
		}
	}
	if d, ok := rec["duration_seconds"]; ok {
		if n, isNum := d.(float64); !isNum || n < 0 {
			return fail(field+".duration_seconds", "must be a non-negative number")
		}
	}
	for _, k := range []string{"candidate", "reason"} {
		if val, ok := rec[k]; ok {
			if _, e := text(val, field+"."+k); e != nil {
				return e
			}
		}
	}
	return nil
}

// extension mirrors Python's pathlib suffix: the last dot counts only when
// it is neither the first nor the last character of the name.
func extension(name string) string {
	i := strings.LastIndex(name, ".")
	if i <= 0 || i == len(name)-1 {
		return ""
	}
	return strings.ToLower(name[i:])
}

func (v *validator) path(value any, field string, extensions []string, limit int64) *ManifestError {
	p, e := text(value, field)
	if e != nil {
		return e
	}
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, "\\") || drivePattern.MatchString(p) {
		return fail(field, "path must be relative to the folder")
	}
	segments := strings.Split(p, "/")
	for _, seg := range segments {
		if seg == "" || seg == "." || seg == ".." {
			return fail(field, "path must not contain empty, '.' or '..' segments: %q", p)
		}
		if !segmentPattern.MatchString(seg) {
			return fail(field, "path segment %q must match %s", seg, segmentPattern.String())
		}
	}
	if ext := extension(segments[len(segments)-1]); !contains(extensions, ext) {
		if ext == "" {
			ext = "(none)"
		}
		return fail(field, "extension %s is not allowed; use one of %s", ext, strings.Join(extensions, ", "))
	}
	for i := range segments {
		prefix := filepath.FromSlash(strings.Join(segments[:i+1], "/"))
		info, err := v.root.Lstat(prefix)
		if errors.Is(err, fs.ErrNotExist) {
			return fail(field, "referenced file does not exist: %q", p)
		}
		if err != nil {
			return fail(field, "cannot inspect %q: %v", p, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return fail(field, "path crosses a symbolic link: %q", p)
		}
		if i == len(segments)-1 {
			if !info.Mode().IsRegular() {
				return fail(field, "referenced file does not exist: %q", p)
			}
			if info.Size() > limit {
				return fail(field, "referenced file exceeds %d MiB: %q", limit>>20, p)
			}
			if err := validateSingleLink(info); err != nil {
				return fail(field, "%q: %v", p, err)
			}
		} else if !info.IsDir() {
			return fail(field, "referenced file does not exist: %q", p)
		}
	}
	v.files[p] = true
	return nil
}
