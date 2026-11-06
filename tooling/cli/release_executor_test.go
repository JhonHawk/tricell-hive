package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
)

func TestReleaseExecutorRefusesWrongProductAndMalformedRequests(t *testing.T) {
	for _, data := range []string{`{"protocol":99}`, `{} {}`, `{"protocol":1,"operation_id":"test","action":"prepare","product_version":"1.2.3","payload":{}}`} {
		var out bytes.Buffer
		err := runReleaseExecutor(strings.NewReader(data), &out)
		if err == nil {
			var r distribution.UpdateResponse
			if json.Unmarshal(out.Bytes(), &r) != nil || r.Status != distribution.UpdateFailed {
				t.Fatalf("accepted request: %s", out.String())
			}
		}
	}
}

// This executor is an external process double. The launcher must route an
// opaque future plan before trying to read the deliberately unsupported core.
func TestRecoverRoutesRetainedExecutorBeforeUnsupportedCoreSchema(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(root, "state")
	pkg := filepath.Join(root, "packages", "target")
	if err = os.MkdirAll(filepath.Join(pkg, "bin"), 0700); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\n[ \"$1\" = \"__hive-update-v1\" ] || exit 2\ncat >/dev/null\nprintf '%s\\n' '{\"protocol\":1,\"operation_id\":\"route-test\",\"action\":\"recover\",\"product_version\":\"1.2.3\",\"status\":\"activated\",\"payload\":{}}'\n"
	exe := filepath.Join(pkg, "bin", "hive")
	if err = os.WriteFile(exe, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	h := strings.Repeat("a", 64)
	bootstrap := distribution.UpdateBootstrap{Protocol: 1, ProductVersion: "1.2.3", Platform: runtime.GOOS + "/" + runtime.GOARCH, SourceID: h, ExecutableSHA256: distribution.Digest([]byte(script)), ManifestVersion: 1, MinStateVersion: 1, MaxStateVersion: 999}
	binding := distribution.UpdatePackageBinding{Artifact: distribution.UpdateArtifact{ProductVersion: "1.2.3", Platform: bootstrap.Platform, Label: "target", SourceID: h, ArchiveSHA256: h, ManifestSHA256: h, Bootstrap: bootstrap}, PackageDirectory: pkg, PreparationDirectory: filepath.Join(root, "preparation"), TargetExecutable: distribution.ExecutorIdentity{Path: exe, SHA256: bootstrap.ExecutableSHA256}, PreparationExecutor: distribution.ExecutorIdentity{Path: filepath.Join(root, "preparation", "bin", "hive"), SHA256: bootstrap.ExecutableSHA256}}
	p, err := management.FreezeUpdatePlan(management.UpdatePlan{Kind: management.UpdatePlanKind, Version: 1, Protocol: 1, OperationID: "route-test", Package: binding, StateDir: state, CoreStateVersion: 999, CorePlan: json.RawMessage(`{"version":999,"future":true}`), CorePlanID: h, CoreTransactionID: "00112233445566778899aabbccddeeff", Hosts: []string{"codex"}, Command: management.UpdateCommand{Path: filepath.Join(root, ".local", "bin", "hive")}})
	if err != nil {
		t.Fatal(err)
	}
	journal, err := management.SealUpdateJournal(management.UpdateJournal{Version: 1, Plan: p, Phase: management.UpdateCorePending, CoreJournalPath: filepath.Join(state, "transactions", p.CoreTransactionID+".json")})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Join(state, "update"), 0700); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(journal)
	if err = os.WriteFile(filepath.Join(state, "update", "pending.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(state, "state.json"), []byte(`{"version":999,"future":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	output := captureStdout(t, func() {
		if err := run([]string{"recover", "--state-dir", state}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, `"status": "activated"`) {
		t.Fatalf("wrong routed result: %s", output)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(cwd, state)
	if err != nil {
		t.Fatal(err)
	}
	captureStdout(t, func() {
		if err := run([]string{"recover", "--state-dir", relative}); err != nil {
			t.Fatal("relative recovery route", err)
		}
	})
	// Corrupting the retained executable must stop routing before execution.
	if err = os.WriteFile(exe, []byte(script+"# altered\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err = run([]string{"recover", "--state-dir", state}); err == nil {
		t.Fatal("executed altered retained bytes")
	}
}
