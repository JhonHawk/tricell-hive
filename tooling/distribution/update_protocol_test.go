package distribution

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUpdateBootstrapCompatibility(t *testing.T) {
	good := UpdateBootstrap{Protocol: UpdateProtocolVersion, ProductVersion: "1.2.3", Platform: runtime.GOOS + "/" + runtime.GOARCH, SourceID: strings.Repeat("a", 64), ExecutableSHA256: strings.Repeat("b", 64), ManifestVersion: 2, MinStateVersion: 1, MaxStateVersion: 7}
	for _, tc := range []struct {
		name   string
		change func(*UpdateBootstrap)
		bad    bool
	}{
		{"compatible", func(*UpdateBootstrap) {}, false},
		{"absent", func(b *UpdateBootstrap) { b.Protocol = 0 }, true},
		{"unknown", func(b *UpdateBootstrap) { b.Protocol = 99 }, true},
		{"schema", func(b *UpdateBootstrap) { b.MaxStateVersion = 6 }, true},
		{"manifest", func(b *UpdateBootstrap) { b.ManifestVersion = 99 }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := good
			tc.change(&b)
			if err := b.Validate("1.2.3", good.Platform, 7); (err != nil) != tc.bad {
				t.Fatalf("Validate = %v", err)
			}
		})
	}
	if err := good.Validate("1.2.4", good.Platform, 7); err == nil {
		t.Fatal("accepted a different target version")
	}
}

func TestUpdateBootstrapBoundToManifest(t *testing.T) {
	root := packageFixture(t)
	m, err := ReadManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	m.ProductVersion = "1.2.3"
	if _, err := ReadUpdateBootstrap(root, m, 7); err == nil {
		t.Fatal("accepted legacy package without protocol")
	}
	b := UpdateBootstrap{Protocol: 1, ProductVersion: m.ProductVersion, Platform: m.Platform, SourceID: m.SourceID, ExecutableSHA256: m.Files["bin/hive"], ManifestVersion: m.Version, MinStateVersion: 1, MaxStateVersion: 7}
	data, _ := json.Marshal(b)
	if err := os.WriteFile(filepath.Join(root, UpdateBootstrapName), data, 0600); err != nil {
		t.Fatal(err)
	}
	m.UpdateProtocol = 1
	m.Files[UpdateBootstrapName] = Digest(data)
	if _, err := ReadUpdateBootstrap(root, m, 7); err != nil {
		t.Fatal(err)
	}
	m.Files["bin/hive"] = strings.Repeat("e", 64)
	if _, err := ReadUpdateBootstrap(root, m, 7); err == nil {
		t.Fatal("accepted bootstrap from different executable")
	}
	m.Files["bin/hive"] = b.ExecutableSHA256
	if err := os.WriteFile(filepath.Join(root, UpdateBootstrapName), append(data, ' '), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadUpdateBootstrap(root, m, 7); err == nil {
		t.Fatal("accepted altered bootstrap bytes")
	}
}

func TestUpdateWireRejectsMalformedAndOversizedInput(t *testing.T) {
	for _, input := range []string{`{`, `{} {}`, `{"protocol":99}`, strings.Repeat(" ", MaxUpdateMessageBytes+1)} {
		if _, err := ReadUpdateRequest(strings.NewReader(input)); err == nil {
			t.Fatal("accepted invalid protocol input")
		}
	}
}

// The executor is an external process, not an in-process replacement of our
// modules. Both versions receive exactly the same fixed command and envelope.
func TestInvokeUpdateAcrossTwoExecutors(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	source := `package main
import ("encoding/json"; "os"; "strings")
var product string
func main() {
    if len(os.Args) != 2 || os.Args[1] != "__hive-update-v1" { os.Exit(3) }
    var q struct {
        OperationID string ` + "`json:\"operation_id\"`" + `
        Action string ` + "`json:\"action\"`" + `
    }
    if json.NewDecoder(os.Stdin).Decode(&q) != nil { os.Exit(4) }
    switch q.OperationID {
    case "malformed": os.Stdout.WriteString("{"); return
    case "oversize": os.Stdout.WriteString(strings.Repeat("x", 33554433)); return
    }
    status := "prepared"
    switch q.OperationID {
    case "core-only": status = "core_committed"
    case "activated": status = "activated"
    case "recovered": status = "recovered"
    }
    json.NewEncoder(os.Stdout).Encode(map[string]any{
        "protocol": 1, "operation_id": q.OperationID, "action": q.Action,
        "product_version": product, "status": status,
        "payload": map[string]any{"version": 999},
    })
}`
	src := filepath.Join(dir, "executor.go")
	if err := os.WriteFile(src, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	for _, v := range []string{"1.2.3", "1.3.0"} {
		binary := filepath.Join(dir, "hive-"+v)
		cmd := exec.Command("go", "build", "-ldflags", "-X main.product="+v, "-o", binary, src)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build fixture: %s: %v", out, err)
		}
		b, err := os.ReadFile(binary)
		if err != nil {
			t.Fatal(err)
		}
		executor := ExecutorIdentity{Path: binary, SHA256: Digest(b)}
		for _, tc := range []struct {
			action    UpdateAction
			operation string
			status    UpdateStatus
			refused   bool
		}{
			{UpdatePrepare, "operation", UpdatePrepared, false},
			{UpdateApply, "core-only", UpdateCoreCommitted, false},
			{UpdateApply, "activated", UpdateActivated, false},
			{UpdateRecover, "core-only", UpdateCoreCommitted, false},
			{UpdateRecover, "activated", UpdateActivated, false},
			{UpdateRecover, "recovered", UpdateRecovered, false},
			{UpdatePrepare, "activated", UpdateActivated, true},
			{UpdateApply, "operation", UpdatePrepared, true},
			{UpdateRecover, "operation", UpdatePrepared, true},
		} {
			t.Run(v+"/"+string(tc.action)+"/"+string(tc.status), func(t *testing.T) {
				req := UpdateRequest{Protocol: 1, OperationID: tc.operation, Action: tc.action, ProductVersion: v, Payload: json.RawMessage(`{}`)}
				response, err := InvokeUpdate(context.Background(), executor, req)
				if tc.refused {
					if err == nil {
						t.Fatalf("accepted %s response for %s", tc.status, tc.action)
					}
					return
				}
				if err != nil || response.Status != tc.status {
					t.Fatalf("response = %+v, error = %v; want status %s", response, err, tc.status)
				}
				if string(response.Payload) != `{"version":999}` {
					t.Fatalf("opaque response payload changed: %s", response.Payload)
				}
			})
		}
		req := UpdateRequest{Protocol: 1, OperationID: "operation", Action: UpdatePrepare, ProductVersion: v, Payload: json.RawMessage(`{}`)}
		req.ProductVersion = "9.9.9"
		if _, err := InvokeUpdate(context.Background(), executor, req); err == nil {
			t.Fatal("accepted different version")
		}
		req.ProductVersion = v
		req.OperationID = "oversize"
		if _, err := InvokeUpdate(context.Background(), executor, req); err == nil {
			t.Fatal("accepted oversized response")
		}
		req.OperationID = "malformed"
		if _, err := InvokeUpdate(context.Background(), executor, req); err == nil {
			t.Fatal("accepted malformed response")
		}
		executor.SHA256 = Digest(bytes.Repeat([]byte("wrong"), 5))
		if _, err := InvokeUpdate(context.Background(), executor, req); err == nil {
			t.Fatal("executed changed bytes")
		}
	}
}
