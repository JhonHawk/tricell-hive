package management

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/distribution"
)

// sha256Hex and encodeRelease reimplement production's hash/encode
// independently, so a change to those functions cannot silently change the
// identities the characterization test below expects.
func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func encodeRelease(v any) []byte {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(b, '\n')
}
func writeManifestBytes(t *testing.T, dir string, data []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, distribution.ManifestName), data, 0640); err != nil {
		t.Fatal(err)
	}
}

// makeUnreadableVersion makes VERSION a directory, so any attempt to read it
// fails and a branch that must not read VERSION is caught doing so.
func makeUnreadableVersion(t *testing.T, dir string) {
	t.Helper()
	if err := os.Mkdir(filepath.Join(dir, "VERSION"), 0750); err != nil {
		t.Fatal(err)
	}
}
func writeVersionFile(t *testing.T, dir, value string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte(value), 0640); err != nil {
		t.Fatal(err)
	}
}

// TestProductFromSourceCharacterizesIdentityAcrossManifestAndVersionSources
// fixes productFromSource's persisted identity for every source shape: the
// artifact ID is stored in installer state, so a changed value would stop an
// already published version from being recognized. Expected identities are
// computed with crypto/sha256 and a local reimplementation of encode, never
// with production's hash/encode.
func TestProductFromSourceCharacterizesIdentityAcrossManifestAndVersionSources(t *testing.T) {
	releaseID := strings.Repeat("a", 64)
	r := Release{ID: releaseID}

	t.Run("manifest present", func(t *testing.T) {
		dir := t.TempDir()
		manifest := []byte(`{"product_version":"1.2.3"}`)
		writeManifestBytes(t, dir, manifest)
		got, err := productFromSource(Options{Source: dir}, r, State{})
		if err != nil {
			t.Fatal(err)
		}
		want := ProductIdentity{Version: "1.2.3", ArtifactID: sha256Hex(manifest), ReleaseID: releaseID}
		if got == nil || *got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("manifest present with unreadable VERSION never read", func(t *testing.T) {
		dir := t.TempDir()
		manifest := []byte(`{"product_version":"4.5.6"}`)
		writeManifestBytes(t, dir, manifest)
		makeUnreadableVersion(t, dir)
		got, err := productFromSource(Options{Source: dir}, r, State{})
		if err != nil {
			t.Fatal(err)
		}
		want := ProductIdentity{Version: "4.5.6", ArtifactID: sha256Hex(manifest), ReleaseID: releaseID}
		if got == nil || *got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("legacy manifest without product_version yields nil without reading VERSION", func(t *testing.T) {
		dir := t.TempDir()
		writeManifestBytes(t, dir, []byte(`{"version":1,"platform":"linux/amd64"}`))
		makeUnreadableVersion(t, dir)
		got, err := productFromSource(Options{Source: dir}, r, State{})
		if err != nil || got != nil {
			t.Fatalf("got %+v, %v", got, err)
		}
	})

	t.Run("VERSION present without manifest", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "9.9.9\n")
		got, err := productFromSource(Options{Source: dir}, r, State{})
		if err != nil {
			t.Fatal(err)
		}
		artifact := sha256Hex(append([]byte("9.9.9\n"), encodeRelease(r)...))
		want := ProductIdentity{Version: "9.9.9", ArtifactID: artifact, ReleaseID: releaseID}
		if got == nil || *got != want {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("neither manifest nor VERSION", func(t *testing.T) {
		dir := t.TempDir()
		got, err := productFromSource(Options{Source: dir}, r, State{})
		if err != nil || got != nil {
			t.Fatalf("got %+v, %v", got, err)
		}
	})

	t.Run("empty VERSION", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "   \n")
		got, err := productFromSource(Options{Source: dir}, r, State{})
		if err != nil || got != nil {
			t.Fatalf("got %+v, %v", got, err)
		}
	})

	t.Run("invalid VERSION without manifest errors", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "not-a-version\n")
		_, err := productFromSource(Options{Source: dir}, r, State{})
		if err == nil || err.Error() != "invalid product identity" {
			t.Fatalf("got %v, want invalid product identity", err)
		}
	})
}

func TestProductVersionCreatesConsumerReceiptWithoutChangingPayloadIdentity(t *testing.T) {
	o := setup(t)
	put(t, filepath.Join(o.Source, "VERSION"), "1.2.0\n")
	first := plan(t, "install", o)
	apply(t, first)
	var state map[string]json.RawMessage
	if err := json.Unmarshal([]byte(get(t, filepath.Join(o.StateDir, "state.json"))), &state); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(state["installations"]), `"1.2.0"`) {
		t.Fatal("missing product version receipt")
	}
	put(t, filepath.Join(o.Source, "VERSION"), "1.2.1\n")
	second := plan(t, "install", o)
	if second.Release.ID != first.Release.ID {
		t.Fatal("product version altered payload identity")
	}
	if unchanged, err := PlanUnchanged(second); err != nil || unchanged {
		t.Fatalf("metadata update omitted: %v %v", unchanged, err)
	}
	apply(t, second)
	if !strings.Contains(get(t, filepath.Join(o.StateDir, "state.json")), `"1.2.1"`) {
		t.Fatal("new label not recorded")
	}
}

func TestPublishedVersionCannotBeReboundToDifferentSource(t *testing.T) {
	o := setup(t)
	put(t, filepath.Join(o.Source, "VERSION"), "2.0.0\n")
	apply(t, plan(t, "install", o))
	put(t, filepath.Join(o.Source, GlobalSource), "Different content.\n")
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("version rebound to different source")
	}
}

func TestVersionReceiptRecoveryPreservesCommittedPredecessor(t *testing.T) {
	o := setup(t)
	put(t, filepath.Join(o.Source, "VERSION"), "3.0.0\n")
	apply(t, plan(t, "install", o))
	before := get(t, filepath.Join(o.StateDir, "state.json"))
	put(t, filepath.Join(o.Source, "VERSION"), "3.0.1\n")
	p := plan(t, "install", o)
	_, err := (Engine{failpoint: func(stage string) error {
		if stage == "state" {
			return &versionTestError{}
		}
		return nil
	}}).Apply(p)
	if err == nil {
		t.Fatal("metadata-only operation did not reach journal")
	}
	if _, err = (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if got := get(t, filepath.Join(o.StateDir, "state.json")); got != before {
		t.Fatal("recovery changed previous receipt")
	}
}

type versionTestError struct{}

func (*versionTestError) Error() string { return "test interruption" }

func TestRequiredHostsExplainsSharedChangesWithoutSelectingThem(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	o.Hosts = []string{"codex"}
	put(t, filepath.Join(o.Source, SkillSource), "---\nname: workspace-conventions\ndescription: Organize records.\n---\nUpdated shared content.\n")
	hosts, err := RequiredHosts(o)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(hosts, ",") != "claude,codex" {
		t.Fatalf("wrong cohort %v", hosts)
	}
	if len(o.Hosts) != 1 {
		t.Fatal("selection mutated")
	}
	if _, err = BuildPlan("install", o); err == nil {
		t.Fatal("unapproved cohort accepted")
	}
}

func TestConsumerVersionIsIndependentOfSharedRecordProvenance(t *testing.T) {
	o := setup(t)
	put(t, filepath.Join(o.Source, "VERSION"), "4.0.0\n")
	apply(t, plan(t, "install", o))
	o.Hosts = []string{"codex"}
	put(t, filepath.Join(o.Source, "VERSION"), "4.0.1\n")
	apply(t, plan(t, "install", o))
	s := stateFor(t, o)
	for _, item := range []struct{ host, version string }{{"codex", "4.0.1"}, {"claude", "4.0.0"}} {
		got, status := installationVersion(s, Consumer{item.host, "user", o.Home})
		if got != item.version || status != "verified" {
			t.Fatalf("%s: %s %s", item.host, got, status)
		}
	}
}

func TestV5StateIsNotMigratedDuringPlanning(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	apply(t, p)
	old := stateFor(t, o)
	old.Version = 5
	put(t, filepath.Join(o.StateDir, "state.json"), string(encode(old)))
	before := get(t, filepath.Join(o.StateDir, "state.json"))
	next := plan(t, "install", o)
	if get(t, filepath.Join(o.StateDir, "state.json")) != before {
		t.Fatal("planning migrated state")
	}
	if next.Release.ID != p.Release.ID {
		t.Fatal("migration changed payload identity")
	}
	apply(t, next)
	if stateFor(t, o).Version != stateVersion {
		t.Fatal("state not migrated transactionally")
	}
}
