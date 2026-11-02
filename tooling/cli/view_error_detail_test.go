package main

import (
	"strings"
	"testing"
)

func TestViewErrorDetailNamesOnlyTheServerCode(t *testing.T) {
	for _, c := range []struct{ body, want string }{
		{`{"code":"stage_binding_mismatch","field":"segment_id"}`, ": stage_binding_mismatch on segment_id"},
		{`{"code":"version_conflict"}`, ": version_conflict"},
		{`{"code":"Bad Code <script>"}`, ""},
		{`{"code":"forbidden","field":"x y"}`, ": forbidden"},
		{`not json`, ""},
		{``, ""},
	} {
		if got := viewErrorDetail(strings.NewReader(c.body)); got != c.want {
			t.Errorf("viewErrorDetail(%q) = %q, want %q", c.body, got, c.want)
		}
	}
}
