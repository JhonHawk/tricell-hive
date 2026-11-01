package main

import (
	"runtime/debug"
	"testing"
)

func TestViewBuildIdentityDoesNotInventARevision(t *testing.T) {
	if revision, modified := viewBuildSettings(nil); revision != "unknown" || modified {
		t.Fatal("missing build identity was fabricated")
	}
	settings := []debug.BuildSetting{{Key: "vcs.revision", Value: "fixture-revision"}, {Key: "vcs.modified", Value: "true"}}
	if revision, modified := viewBuildSettings(settings); revision != "fixture-revision" || !modified {
		t.Fatal("build identity lost revision or dirty state")
	}
}
