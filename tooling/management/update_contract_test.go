package management

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/distribution"
)

func frozenUpdateFixture() UpdatePlan {
	h := strings.Repeat("a", 64)
	b := distribution.UpdateBootstrap{Protocol: 1, ProductVersion: "1.2.3", Platform: "linux/amd64", SourceID: h, ExecutableSHA256: h, ManifestVersion: 2, MinStateVersion: 1, MaxStateVersion: 999}
	p := UpdatePlan{Kind: UpdatePlanKind, Version: 1, Protocol: 1, OperationID: "operation-1", StateDir: "/home/test/state", CoreStateVersion: 999, CorePlan: json.RawMessage(`{"version":999,"future_schema":{"anything":true}}`), CorePlanID: h, CoreTransactionID: "reserved-child", Hosts: []string{"codex"}, Command: UpdateCommand{Path: "/home/test/.local/bin/hive", Expected: Fingerprint{Exists: false}}}
	p.Package = distribution.UpdatePackageBinding{Artifact: distribution.UpdateArtifact{ProductVersion: b.ProductVersion, Platform: b.Platform, Label: "hive-1.2.3-linux-amd64", ArchiveSHA256: h, ManifestSHA256: h, SourceID: h, Bootstrap: b}, PreparationDirectory: "/private/tmp/preparation", PackageDirectory: "/home/test/packages/hive-1.2.3-linux-amd64", PreparationExecutor: distribution.ExecutorIdentity{Path: "/private/tmp/preparation/bin/hive", SHA256: h}, TargetExecutable: distribution.ExecutorIdentity{Path: "/home/test/packages/hive-1.2.3-linux-amd64/bin/hive", SHA256: h}}
	p.CorePlanSHA256 = distribution.Digest(p.CorePlan)
	p, err := FreezeUpdatePlan(p)
	if err != nil {
		panic(err)
	}
	return p
}

func TestUpdateParentReadableWithoutCoreSchema(t *testing.T) {
	p := frozenUpdateFixture()
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if !IsUpdatePlan(data) {
		t.Fatal("parent not discriminated")
	}
	got, err := ReadUpdatePlan(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.CorePlan, p.CorePlan) || got.CoreStateVersion != 999 {
		t.Fatal("decoded or changed opaque core plan")
	}
	if IsUpdatePlan([]byte(`{"version":7,"action":"install"}`)) {
		t.Fatal("core plan treated as parent")
	}
	if _, err := ReadUpdatePlan(strings.NewReader(`{"kind":"release_update","version":99}`)); err == nil {
		t.Fatal("accepted unknown parent schema")
	}
}

func TestUpdateParentRejectsChangedBinding(t *testing.T) {
	for _, tc := range []struct {
		name           string
		change         func(*UpdatePlan)
		invalidBinding bool
	}{
		{"changed plan", func(p *UpdatePlan) { p.CorePlan = json.RawMessage(`{"version":1000}`) }, false},
		{"temporary hook target", func(p *UpdatePlan) { p.Package.TargetExecutable = p.Package.PreparationExecutor }, true},
		{"different manifest", func(p *UpdatePlan) { p.Package.Artifact.SourceID = strings.Repeat("b", 64) }, true},
		{"relative command", func(p *UpdatePlan) { p.Command.Path = "bin/hive" }, true},
		{"missing transaction", func(p *UpdatePlan) { p.CoreTransactionID = "" }, true},
		{"duplicate host", func(p *UpdatePlan) { p.Hosts = append(p.Hosts, "codex") }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := frozenUpdateFixture()
			tc.change(&p)
			t.Run("tampered", func(t *testing.T) {
				if err := p.Validate(); err == nil {
					t.Fatal("accepted changed parent")
				}
			})
			t.Run("fresh", func(t *testing.T) {
				// Recompute the envelope's integrity, so invalid bindings must
				// fail semantic validation rather than the tampering check.
				p.ID = ""
				p.CorePlanSHA256 = ""
				_, err := FreezeUpdatePlan(p)
				if tc.invalidBinding && err == nil {
					t.Fatal("froze a semantically invalid parent")
				}
				if !tc.invalidBinding && err != nil {
					t.Fatalf("could not freeze a new opaque core plan: %v", err)
				}
			})
		})
	}
}

func TestUpdateJournalBindsDurableChild(t *testing.T) {
	p := frozenUpdateFixture()
	j := UpdateJournal{Version: 1, Plan: p, Phase: UpdateCorePending, CoreJournalPath: filepath.Join(p.StateDir, "transactions", p.CoreTransactionID+".json")}
	j, sealErr := SealUpdateJournal(j)
	if sealErr != nil {
		t.Fatal(sealErr)
	}
	if err := j.Validate(); err != nil {
		t.Fatal(err)
	}
	j.CoreJournalPath = filepath.Join(p.StateDir, "transactions", "different.json")
	if err := j.Validate(); err == nil {
		t.Fatal("accepted unrelated durable evidence")
	}
	j.CoreJournalPath = filepath.Join(p.StateDir, "transactions", p.CoreTransactionID+".json")
	j.Phase = "unknown"
	if err := j.Validate(); err == nil {
		t.Fatal("accepted unknown recovery phase")
	}
}

func TestUpdateEnvelopeIntegritySurvivesSavedJSON(t *testing.T) {
	p := frozenUpdateFixture()
	p.CorePlan = json.RawMessage("{\n  \"version\": 999, \"future_path\": \"/home/<operator>/hive\"\n}")
	p, err := FreezeUpdatePlan(p)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ReadUpdatePlan(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	j, err := SealUpdateJournal(UpdateJournal{Version: 1, Plan: p, Phase: UpdateCorePending, CoreJournalPath: filepath.Join(p.StateDir, "transactions", p.CoreTransactionID+".json")})
	if err != nil {
		t.Fatal(err)
	}
	j.Phase = UpdateActivated
	if err := j.Validate(); err == nil {
		t.Fatal("accepted altered durable phase")
	}
	p.Command.Expected = Fingerprint{Exists: true, Kind: "symlink", LinkTarget: "/foreign"}
	if err := p.Validate(); err == nil {
		t.Fatal("accepted altered command expectation")
	}
}
