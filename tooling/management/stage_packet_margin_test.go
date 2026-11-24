package management

import (
	"sort"
	"testing"
)

// stagePacketMargin is the room each real packet keeps under the manager's cap,
// so a leaf that grows fails this test before it makes every release invalid.
const stagePacketMargin = 500

func TestRealStagePacketsKeepAMarginUnderTheCap(t *testing.T) {
	r, err := loadRelease(Options{Source: "../.."}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	packets, err := renderStagePackets(&r)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) == 0 {
		t.Fatal("the repository content renders no stage packet")
	}
	names := make([]string, 0, len(packets))
	for name := range packets {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		size := len(packets[name])
		if size > stagePacketMaxBytes-stagePacketMargin {
			t.Errorf("%s: %d B, over %d B (cap %d B less a %d B margin): split a member leaf by responsibility or move a leaf the stage does not need out of its assets/stage.json; do not raise the cap", name, size, stagePacketMaxBytes-stagePacketMargin, stagePacketMaxBytes, stagePacketMargin)
		}
		t.Logf("%s: %d B", name, size)
	}
}
