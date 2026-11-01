package viewer

import (
	"encoding/json"
	"testing"
)

func TestTimelineV2SharedFixtures(t *testing.T) {
	s, err := DecodeTimelineSession(timelineFixture(t, "v2-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string `json:"name"`
		Action   string `json:"action"`
		Kind     string `json:"kind"`
		ObjectID string `json:"object_id"`
		Valid    bool   `json:"valid"`
		JSON     string `json:"json"`
	}
	if err := json.Unmarshal(timelineFixture(t, "v2-cases.json"), &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var err error
			switch c.Action {
			case "resolved_project":
				var context struct {
					Source      TimelineSource `json:"source"`
					ResolvedKey string         `json:"resolved_key"`
				}
				err = json.Unmarshal([]byte(c.JSON), &context)
				if err == nil {
					err = ValidateTimelineResolvedProject(s, context.Source, context.ResolvedKey)
				}
			case "graph":
				_, err = DecodeTimelineSession([]byte(c.JSON))
			case "shape":
				_, err = DecodeTimelineV2Operation(c.Kind, []byte(c.JSON))
			case "references":
				var value any
				value, err = DecodeTimelineV2Operation(c.Kind, []byte(c.JSON))
				if err == nil {
					err = ValidateTimelineV2OperationReferences(s, c.Kind, c.ObjectID, value)
				}
			case "wait":
				var w TimelineCIWait
				err = DecodeTimelineJSON([]byte(c.JSON), &w)
				if err == nil {
					err = ValidateTimelineCIWait(w)
				}
			case "authority":
				var a TimelineActor
				err = DecodeTimelineJSON([]byte(c.JSON), &a)
				if err == nil {
					err = ValidateTimelineAuthority(s, c.Kind, a, "producer-original")
				}
			default:
				t.Fatal("unknown fixture action")
			}
			if (err == nil) != c.Valid {
				t.Fatalf("valid=%v got %v", c.Valid, err)
			}
		})
	}
}
func TestTimelineV2TaskAggregateAndBounds(t *testing.T) {
	for _, c := range []struct {
		states []string
		want   string
	}{{[]string{"completed", "not_started"}, "in_progress"}, {[]string{"completed", "completed"}, "completed"}, {[]string{"blocked", "in_review"}, "blocked"}, {[]string{"completed", "in_review"}, "in_review"}, {[]string{"not_started"}, "not_started"}} {
		var segs []TimelineTaskSegment
		for _, s := range c.states {
			segs = append(segs, TimelineTaskSegment{State: s})
		}
		if got := AggregateTimelineTaskState(segs); got != c.want {
			t.Fatalf("aggregate %v got %s", c.states, got)
		}
	}
	for _, c := range []struct {
		count, add      int
		terminal, valid bool
	}{{499, 1, false, true}, {500, 1, false, false}, {500, 8, true, true}, {508, 1, true, false}} {
		if err := ValidateTimelineEventCapacity(c.count, c.add, c.terminal); (err == nil) != c.valid {
			t.Fatalf("capacity %v: %v", c, err)
		}
	}
}
func TestTimelineV2RejectsServerFieldsAndMixedActors(t *testing.T) {
	raw := []byte(`{"schema_version":2,"operation_id":"add","actor":{"kind":"native_producer","producer_id":"producer-original"},"kind":"request","title":"Request","summary":"Scope","ticket_refs":[],"stage_event_id":"injected"}`)
	if _, err := DecodeTimelineAppendEvent(raw); err == nil {
		t.Fatal("server association accepted")
	}
	if err := (TimelineActor{Kind: "local_viewer", ProducerID: "producer-original"}).Validate(); err == nil {
		t.Fatal("mixed actor accepted")
	}
}

func TestTimelineV2CanonicalProjectPreservesActualSources(t *testing.T) {
	s, err := DecodeTimelineSession(timelineFixture(t, "v2-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	before, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Producers[1].Source.ProjectRoot != "/fixture/project-worktree" || s.PendingJoinRequests[0].Source.ProjectRoot != "/fixture/project-alias" {
		t.Fatal("truthful worktree/alias source was replaced")
	}
	if err := ValidateTimelineResolvedProject(s, s.Producers[1].Source, s.ProjectKey); err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("project validation rewrote provenance")
	}
}
