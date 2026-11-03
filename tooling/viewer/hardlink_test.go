package viewer

import (
	"os"
	"strings"
	"syscall"
	"testing"
)

// hardlinkFileInfo supplies only the synthetic metadata used by the link policy.
type hardlinkFileInfo struct {
	os.FileInfo
	metadata any
}

func (i hardlinkFileInfo) Sys() any { return i.metadata }

func TestValidateSingleLinkMetadata(t *testing.T) {
	for _, tc := range []struct {
		name      string
		metadata  any
		wantError string
	}{
		{"missing", nil, "cannot determine"},
		{"unsupported", struct{}{}, "cannot determine"},
		{"nil-stat", (*syscall.Stat_t)(nil), "cannot determine"},
		{"multiple-links", &syscall.Stat_t{Nlink: 2}, "more than one hard link"},
		{"single-link-other-owner", &syscall.Stat_t{Nlink: 1, Uid: ^uint32(0)}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSingleLink(hardlinkFileInfo{metadata: tc.metadata})
			if tc.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("got %v, want %q", err, tc.wantError)
			}
		})
	}
}
