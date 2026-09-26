package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// ValidateStep admits only a manual step for a known catalog provider: no
// recipe in this build has passed its own native validation gate, so any
// other status is rejected rather than silently accepted.
func ValidateStep(payload json.RawMessage) error {
	var s Step
	d := json.NewDecoder(bytes.NewReader(payload))
	d.DisallowUnknownFields()
	if err := d.Decode(&s); err != nil {
		return fmt.Errorf("invalid provider payload")
	}
	switch s.Provider {
	case Engram, Context7, PiSubagents:
	default:
		return fmt.Errorf("unknown provider")
	}
	if s.Status != Manual {
		return fmt.Errorf("only a manual provider step is supported in this build")
	}
	if s.ManualReason == "" {
		return fmt.Errorf("manual provider step requires a reason")
	}
	return nil
}
