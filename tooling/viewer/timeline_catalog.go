package viewer

// Protocol 3 canonical registries and closed mutation DTOs. Persistence writers
// are deliberately separate from these pure boundary and graph validators.
import (
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

const TimelineMaxCatalogRecords = 100

type TimelineTicketCreateRequest struct {
	TimelineNativeOperation
	TimelineTicketInput
}
type TimelineTicketPatchRequest struct {
	TimelineNativeOperation
	Key   *string `json:"key,omitempty"`
	Title *string `json:"title,omitempty"`
	URL   *string `json:"url,omitempty"`
}
type TimelineDeliveryCreateRequest struct {
	TimelineNativeOperation
	RepositoryURL string `json:"repository_url"`
	ExternalID    string `json:"external_id"`
	TimelineDeliverySnapshot
}
type TimelineDeliveryPatchRequest struct {
	TimelineNativeOperation
	Label  *string `json:"label,omitempty"`
	URL    *string `json:"url,omitempty"`
	Status *string `json:"status,omitempty"`
}
type TimelineCandidateCreateRequest struct {
	TimelineNativeOperation
	TimelineCandidate
}

// CanonicalTimelineIdentityURL normalizes only the stable namespace locator.
// Display locators use their own validator and keep query/fragment/path context.
func CanonicalTimelineIdentityURL(raw string) (string, error) {
	if err := timelineHTTPS(raw, "identity_url"); err != nil {
		return "", err
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", timelineError("identity_url", "invalid_https_url")
	}
	if u.RawQuery != "" || u.ForceQuery || strings.Contains(raw, "#") {
		return "", timelineError("identity_url", "invalid_identity_url")
	}
	for _, segment := range strings.Split(u.EscapedPath(), "/") {
		v, e := url.PathUnescape(segment)
		if e != nil || v == "." || v == ".." {
			return "", timelineError("identity_url", "invalid_identity_url")
		}
	}
	return strings.TrimRight(timelineLocator(raw), "/"), nil
}

// Preserve the locator suffix verbatim: URL parsing can otherwise collapse dot
// segments, IPv6 forms, empty paths, escape spelling or Unicode display paths.
func timelineLocator(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	start := strings.Index(raw, "://")
	if start < 0 {
		return raw
	}
	start += 3
	end := len(raw)
	if n := strings.IndexAny(raw[start:], "/?#"); n >= 0 {
		end = start + n
	}
	host := strings.ToLower(raw[start:end])
	port, e := strconv.Atoi(u.Port())
	if e == nil && port == 443 {
		host = strings.TrimSuffix(host, ":"+u.Port())
	}
	return "https://" + host + raw[end:]
}
func timelineExternalID(s string) error {
	if err := timelineText(s, "external_id", 256, true); err != nil {
		return err
	}
	for _, r := range s {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return timelineError("external_id", "invalid_external_id")
		}
	}
	return nil
}
func timelineTicketInput(t TimelineTicketInput) error {
	if _, e := CanonicalTimelineIdentityURL(t.ExternalScope); e != nil {
		return e
	}
	if e := timelineExternalID(t.ExternalID); e != nil {
		return e
	}
	if e := timelineText(t.Key, "key", 100, true); e != nil {
		return e
	}
	if e := timelineText(t.Title, "title", 300, false); e != nil {
		return e
	}
	if t.URL != "" {
		return timelineHTTPS(t.URL, "url")
	}
	return nil
}
func timelineDeliverySnapshot(d TimelineDeliverySnapshot) error {
	if !contains([]string{"pull_request", "commit"}, d.Kind) || !contains([]string{"open", "merged", "local"}, d.Status) {
		return timelineError("delivery", "invalid_kind_or_status")
	}
	if e := timelineText(d.Label, "label", 200, true); e != nil {
		return e
	}
	return timelineHTTPS(d.URL, "url")
}
func timelineIDs(ids []string, field string, required bool) error {
	if ids == nil || len(ids) > 100 || required && len(ids) == 0 {
		return timelineError(field, "array_required_or_limit")
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if e := timelineID(id, field); e != nil {
			return e
		}
		if seen[id] {
			return timelineError(field, "duplicate_id")
		}
		seen[id] = true
	}
	return nil
}
func timelineTicketIDs(refs []TimelineTicketRef) []string {
	ids := make([]string, 0, len(refs))
	for _, r := range refs {
		ids = append(ids, r.TicketID)
	}
	return ids
}
func timelineDeliveryIDs(refs []TimelineDeliveryReference) []string {
	if refs == nil {
		return nil
	}
	ids := make([]string, 0, len(refs))
	for _, r := range refs {
		ids = append(ids, r.DeliveryID)
	}
	return ids
}
func TimelineCandidateByID(s TimelineSession, id string) (TimelineCandidateRecord, error) {
	for _, c := range s.Candidates {
		if c.CandidateID == id {
			return c, nil
		}
	}
	return TimelineCandidateRecord{}, timelineError("candidate_id", "unknown_candidate")
}
func TimelineDeliveryByID(s TimelineSession, id string) (TimelineDelivery, error) {
	for _, d := range s.Deliveries {
		if d.DeliveryID == id {
			return d, nil
		}
	}
	return TimelineDelivery{}, timelineError("delivery_id", "unknown_delivery")
}
func CaptureTimelineTicketReferences(s TimelineSession, ids []string) ([]TimelineTicketRef, error) {
	if e := timelineIDs(ids, "ticket_ids", false); e != nil {
		return nil, e
	}
	refs := make([]TimelineTicketRef, 0, len(ids))
	for _, id := range ids {
		found := false
		for _, t := range s.Tickets {
			if t.TicketID == id {
				refs = append(refs, TimelineTicketRef{TicketID: id, Snapshot: TimelineTicketSnapshot{Key: t.Key, Title: t.Title, URL: t.URL}})
				found = true
				break
			}
		}
		if !found {
			return nil, timelineError("ticket_ids", "unknown_ticket")
		}
	}
	return refs, nil
}
func CaptureTimelineDeliveryReferences(s TimelineSession, ids []string) ([]TimelineDeliveryReference, error) {
	if e := timelineIDs(ids, "delivery_ids", true); e != nil {
		return nil, e
	}
	refs := make([]TimelineDeliveryReference, 0, len(ids))
	for _, id := range ids {
		d, e := TimelineDeliveryByID(s, id)
		if e != nil {
			return nil, e
		}
		refs = append(refs, TimelineDeliveryReference{DeliveryID: id, Snapshot: d.TimelineDeliverySnapshot})
	}
	return refs, nil
}
func validateTimelineCatalogOperation(value any) (TimelineNativeOperation, bool, error) {
	var base TimelineNativeOperation
	switch r := value.(type) {
	case *TimelineTicketCreateRequest:
		base = r.TimelineNativeOperation
		if e := timelineTicketInput(r.TimelineTicketInput); e != nil {
			return base, true, e
		}
	case *TimelineTicketPatchRequest:
		base = r.TimelineNativeOperation
		if r.Key == nil && r.Title == nil && r.URL == nil {
			return base, true, timelineError("body", "empty_patch")
		}
		if r.Key != nil {
			if e := timelineText(*r.Key, "key", 100, true); e != nil {
				return base, true, e
			}
		}
		if r.Title != nil {
			if e := timelineText(*r.Title, "title", 300, false); e != nil {
				return base, true, e
			}
		}
		if r.URL != nil && *r.URL != "" {
			if e := timelineHTTPS(*r.URL, "url"); e != nil {
				return base, true, e
			}
		}
	case *TimelineDeliveryCreateRequest:
		base = r.TimelineNativeOperation
		if _, e := CanonicalTimelineIdentityURL(r.RepositoryURL); e != nil {
			return base, true, e
		}
		if e := timelineExternalID(r.ExternalID); e != nil {
			return base, true, e
		}
		if e := timelineDeliverySnapshot(r.TimelineDeliverySnapshot); e != nil {
			return base, true, e
		}
	case *TimelineDeliveryPatchRequest:
		base = r.TimelineNativeOperation
		if r.Label == nil && r.URL == nil && r.Status == nil {
			return base, true, timelineError("body", "empty_patch")
		}
		if r.Label != nil {
			if e := timelineText(*r.Label, "label", 200, true); e != nil {
				return base, true, e
			}
		}
		if r.URL != nil {
			if e := timelineHTTPS(*r.URL, "url"); e != nil {
				return base, true, e
			}
		}
		if r.Status != nil && !contains([]string{"open", "merged", "local"}, *r.Status) {
			return base, true, timelineError("status", "invalid_kind_or_status")
		}
	case *TimelineCandidateCreateRequest:
		base = r.TimelineNativeOperation
		if e := timelineCandidate(r.TimelineCandidate); e != nil {
			return base, true, e
		}
	default:
		return base, false, nil
	}
	return base, true, base.validate(false)
}
func validateTimelineCatalog(s TimelineSession) error {
	if s.Deliveries == nil || s.Candidates == nil || len(s.Deliveries) > 100 || len(s.Candidates) > 100 {
		return timelineError("catalog", "array_required_or_limit")
	}
	ids := map[string]bool{}
	natural := map[string]bool{}
	urls := map[string]bool{}
	add := func(id, key, locator string) error {
		if e := timelineID(id, "catalog.id"); e != nil {
			return e
		}
		if ids[id] || natural[key] {
			return timelineError("catalog", "duplicate_identity")
		}
		ids[id] = true
		natural[key] = true
		if locator != "" {
			if urls[locator] {
				return timelineError("catalog", "entity_conflict")
			}
			urls[locator] = true
		}
		return nil
	}
	for _, t := range s.Tickets {
		if e := timelineTicketInput(t.TimelineTicketInput); e != nil {
			return e
		}
		scope, _ := CanonicalTimelineIdentityURL(t.ExternalScope)
		if e := add(t.TicketID, "ticket\x00"+scope+"\x00"+t.ExternalID, func() string {
			if t.URL == "" {
				return ""
			}
			return "ticket\x00" + timelineLocator(t.URL)
		}()); e != nil {
			return e
		}
	}
	for _, d := range s.Deliveries {
		if e := timelineDeliverySnapshot(d.TimelineDeliverySnapshot); e != nil {
			return e
		}
		scope, e := CanonicalTimelineIdentityURL(d.RepositoryURL)
		if e != nil {
			return e
		}
		if e = timelineExternalID(d.ExternalID); e != nil {
			return e
		}
		if e = add(d.DeliveryID, "delivery\x00"+scope+"\x00"+d.Kind+"\x00"+d.ExternalID, "delivery\x00"+timelineLocator(d.URL)); e != nil {
			return e
		}
	}
	for _, c := range s.Candidates {
		if e := timelineCandidate(c.TimelineCandidate); e != nil {
			return e
		}
		if e := add(c.CandidateID, "candidate\x00"+c.Branch+"\x00"+c.Commit+"\x00"+c.Base, ""); e != nil {
			return e
		}
	}
	return nil
}
func validateTimelineCatalogResult(r TimelineOperationResult) error {
	field, id := "", ""
	switch r.Kind {
	case "ticket_create", "ticket_patch":
		field, id = "ticket_id", r.TicketID
	case "delivery_create", "delivery_patch":
		field, id = "delivery_id", r.DeliveryID
	case "candidate_create":
		field, id = "candidate_id", r.CandidateID
	}
	if e := timelineID(id, field); e != nil {
		return e
	}
	if strings.HasSuffix(string(r.Kind), "_create") && r.Created == nil {
		return timelineError("created", "required_field")
	}
	return nil
}
func timelineCatalogProtocolDefinitions(actor TimelineActor) []timelineProtocolDefinition {
	base := TimelineNativeOperation{SchemaVersion: 3, OperationID: "catalog-example", Actor: actor, ExpectedSessionVersion: 1, Reason: "Verified resource registration"}
	return []timelineProtocolDefinition{
		{"ticket_create", "hive view ticket create <session-id> --json <file|->", TimelineTicketCreateRequest{}, TimelineOperationResult{}, TimelineTicketCreateRequest{TimelineNativeOperation: base, TimelineTicketInput: TimelineTicketInput{ExternalScope: "https://example.test/tracker", ExternalID: "resource-1", Key: "EX-1"}}},
		{"ticket_patch", "hive view ticket update <session-id> <ticket-id> --json <file|->", TimelineTicketPatchRequest{}, TimelineOperationResult{}, TimelineTicketPatchRequest{TimelineNativeOperation: base, Key: timelineStringPointer("EX-2")}},
		{"delivery_create", "hive view delivery create <session-id> --json <file|->", TimelineDeliveryCreateRequest{}, TimelineOperationResult{}, TimelineDeliveryCreateRequest{TimelineNativeOperation: base, RepositoryURL: "https://example.test/repository", ExternalID: "42", TimelineDeliverySnapshot: TimelineDeliverySnapshot{Kind: "pull_request", Label: "PR #42", URL: "https://example.test/pull/42", Status: "open"}}},
		{"delivery_patch", "hive view delivery update <session-id> <delivery-id> --json <file|->", TimelineDeliveryPatchRequest{}, TimelineOperationResult{}, TimelineDeliveryPatchRequest{TimelineNativeOperation: base, Status: timelineStringPointer("merged")}},
		{"candidate_create", "hive view candidate create <session-id> --json <file|->", TimelineCandidateCreateRequest{}, TimelineOperationResult{}, TimelineCandidateCreateRequest{TimelineNativeOperation: base, TimelineCandidate: TimelineCandidate{Branch: "feature/example", Commit: "abc1234", Base: "development"}}},
	}
}

// ValidateTimelineCatalogReferences checks the operation against one locked work
// snapshot. Writers still own atomic receipts and persistence.
func ValidateTimelineCatalogReferences(s TimelineSession, kind, objectID string, value any) error {
	base, ok, err := validateTimelineCatalogOperation(value)
	if !ok {
		return timelineError("kind", "unsupported_operation")
	}
	if err != nil {
		return err
	}
	if base.ExpectedSessionVersion != s.Version {
		return timelineError("expected_session_version", "version_conflict")
	}
	if s.State == "closed" || s.State == "failed" {
		return timelineError("state", "session_terminal")
	}
	if err = ValidateTimelineAuthority(s, kind, base.Actor, ""); err != nil {
		return err
	}
	switch r := value.(type) {
	case *TimelineTicketCreateRequest:
		scope, _ := CanonicalTimelineIdentityURL(r.ExternalScope)
		for _, t := range s.Tickets {
			existing, _ := CanonicalTimelineIdentityURL(t.ExternalScope)
			if scope == existing && t.ExternalID == r.ExternalID {
				input := r.TimelineTicketInput
				input.ExternalScope = scope
				old := t.TimelineTicketInput
				old.ExternalScope = existing
				if input != old {
					return timelineError("ticket_id", "entity_exists")
				}
				return nil
			}
			if r.URL != "" && t.URL != "" && timelineLocator(r.URL) == timelineLocator(t.URL) {
				return timelineError("url", "entity_conflict")
			}
		}
		if len(s.Tickets) >= 100 {
			return timelineError("tickets", "limit")
		}
	case *TimelineTicketPatchRequest:
		found := false
		for _, t := range s.Tickets {
			if t.TicketID == objectID {
				found = true
			}
		}
		if !found {
			return timelineError("ticket_id", "unknown_ticket")
		}
		if r.URL != nil && *r.URL != "" {
			for _, t := range s.Tickets {
				if t.TicketID != objectID && t.URL != "" && timelineLocator(t.URL) == timelineLocator(*r.URL) {
					return timelineError("url", "entity_conflict")
				}
			}
		}
	case *TimelineDeliveryCreateRequest:
		scope, _ := CanonicalTimelineIdentityURL(r.RepositoryURL)
		for _, d := range s.Deliveries {
			existing, _ := CanonicalTimelineIdentityURL(d.RepositoryURL)
			if scope == existing && d.Kind == r.Kind && d.ExternalID == r.ExternalID {
				if d.TimelineDeliverySnapshot != r.TimelineDeliverySnapshot {
					return timelineError("delivery_id", "entity_exists")
				}
				return nil
			}
			if timelineLocator(d.URL) == timelineLocator(r.URL) {
				return timelineError("url", "entity_conflict")
			}
		}
		if len(s.Deliveries) >= 100 {
			return timelineError("deliveries", "limit")
		}
	case *TimelineDeliveryPatchRequest:
		if _, err := TimelineDeliveryByID(s, objectID); err != nil {
			return err
		}
		if r.URL != nil {
			for _, d := range s.Deliveries {
				if d.DeliveryID != objectID && timelineLocator(d.URL) == timelineLocator(*r.URL) {
					return timelineError("url", "entity_conflict")
				}
			}
		}
	case *TimelineCandidateCreateRequest:
		for _, c := range s.Candidates {
			if c.TimelineCandidate == r.TimelineCandidate {
				return nil
			}
		}
		if len(s.Candidates) >= 100 {
			return timelineError("candidates", "limit")
		}
	}
	return nil
}

type timelineCatalogEndpoint struct{ method, path string }

func timelineCatalogRoute(kind string) timelineCatalogEndpoint {
	switch kind {
	case "ticket_create":
		return timelineCatalogEndpoint{"POST", "/api/timeline/sessions/:id/tickets"}
	case "ticket_patch":
		return timelineCatalogEndpoint{"PATCH", "/api/timeline/sessions/:id/tickets/:ticket-id"}
	case "delivery_create":
		return timelineCatalogEndpoint{"POST", "/api/timeline/sessions/:id/deliveries"}
	case "delivery_patch":
		return timelineCatalogEndpoint{"PATCH", "/api/timeline/sessions/:id/deliveries/:delivery-id"}
	case "candidate_create":
		return timelineCatalogEndpoint{"POST", "/api/timeline/sessions/:id/candidates"}
	}
	return timelineCatalogEndpoint{}
}

// A browser may reinterpret numeric-looking DNS names as noncanonical IPv4,
// or reject them altogether. The v3 wire accepts only four canonical decimal
// IPv4 octets in that branch, preserving the locator instead of rewriting it.
func timelineURLHost(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if strings.Contains(host, ":") {
		return true
	}
	labels := strings.Split(strings.TrimSuffix(host, "."), ".")
	last := labels[len(labels)-1]
	numeric := last != ""
	for _, r := range last {
		if r < '0' || r > '9' {
			numeric = false
			break
		}
	}
	hex := strings.HasPrefix(last, "0x")
	if hex {
		for _, r := range last[2:] {
			if !isHex(byte(r)) || r > 127 {
				hex = false
				break
			}
		}
	}
	if !numeric && !hex {
		return true
	}
	octets := strings.Split(host, ".")
	if len(octets) != 4 {
		return false
	}
	for _, part := range octets {
		if part == "" || len(part) > 1 && part[0] == '0' {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
		n, e := strconv.Atoi(part)
		if e != nil || n > 255 {
			return false
		}
	}
	return true
}
