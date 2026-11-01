package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"tricell-hive/tooling/viewer"
)

func TestViewTimelineCatalogNativeForms(t *testing.T) {
	id := strings.Repeat("a", 64)
	for _, c := range []struct {
		args                    []string
		operation, method, path string
	}{
		{[]string{"ticket", "create", id}, "ticket_create", http.MethodPost, "/tickets"},
		{[]string{"ticket", "update", id, "ticket-one"}, "ticket_patch", http.MethodPatch, "/tickets/ticket-one"},
		{[]string{"delivery", "create", id}, "delivery_create", http.MethodPost, "/deliveries"},
		{[]string{"delivery", "update", id, "delivery-one"}, "delivery_patch", http.MethodPatch, "/deliveries/delivery-one"},
		{[]string{"candidate", "create", id}, "candidate_create", http.MethodPost, "/candidates"},
	} {
		got, rest, err := parseViewTimelineCommand(c.args)
		if err != nil || len(rest) != 0 || got.operation != c.operation || got.method != c.method || got.path != "/api/timeline/sessions/"+id+c.path {
			t.Fatalf("native form %+v %v", got, err)
		}
		if !isTimelineViewCommand(c.args[0]) {
			t.Fatal("missing native command router")
		}
		contract, err := viewer.TimelineProtocolContract(c.operation)
		if err != nil || len(contract.Commands) != 1 {
			t.Fatal(err)
		}
		if err = validateViewTimelineInput(c.operation, contract.Commands[0].Example); err != nil {
			t.Fatal("offline example rejected", err)
		}
		if contract.Commands[0].Method != c.method || contract.Commands[0].Path == "" {
			t.Fatal("offline route not declared")
		}
	}
}
func TestViewTimelineCatalogOfflineSharedInputs(t *testing.T) {
	raw, err := os.ReadFile("../viewer/testdata/timeline_contract/catalog-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, Action, Kind, JSON string
		Valid                    bool
	}
	if err = json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		if c.Action != "shape" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			if err := validateViewTimelineInput(c.Kind, []byte(c.JSON)); (err == nil) != c.Valid {
				t.Fatalf("valid=%v got %v", c.Valid, err)
			}
		})
	}
}
