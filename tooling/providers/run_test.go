package providers

import (
	"encoding/json"
	"testing"
)

func TestValidateStepAcceptsOnlyManualStepForKnownProvider(t *testing.T) {
	valid := Step{ID: "context7-manual", Provider: Context7, Status: Manual, ManualReason: "fixture reason"}
	payload, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateStep(payload); err != nil {
		t.Fatalf("valid manual step rejected: %v", err)
	}

	unknownProvider := valid
	unknownProvider.Provider = "not-a-provider"
	payload, _ = json.Marshal(unknownProvider)
	if ValidateStep(payload) == nil {
		t.Fatal("unknown provider accepted")
	}

	notManual := valid
	notManual.Status = "verified"
	payload, _ = json.Marshal(notManual)
	if ValidateStep(payload) == nil {
		t.Fatal("non-manual status accepted although this build has no recipe")
	}

	noReason := valid
	noReason.ManualReason = ""
	payload, _ = json.Marshal(noReason)
	if ValidateStep(payload) == nil {
		t.Fatal("manual step accepted without a reason")
	}

	if ValidateStep(json.RawMessage(`{"id":"x","provider":"context7","status":"manual","manual_reason":"r","extra":true}`)) == nil {
		t.Fatal("payload with unknown field accepted")
	}
}
