package viewer

import (
	"os"
	"testing"

	"tricell-hive/tooling/durable"
)

// TestMain turns off fsync for the whole test binary.
func TestMain(m *testing.M) {
	durable.DisableForTests()
	os.Exit(m.Run())
}
