package management

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	wfHeader  = "[features.multi_agent_v2]"
	wfMin     = "min_wait_timeout_ms = 300000"
	wfDefault = "default_wait_timeout_ms = 600000"
)

func wfOptions(t *testing.T) Options {
	t.Helper()
	o := setup(t)
	o.Hosts = []string{"codex"}
	o.CodexWaitFloor = true
	return o
}

func wfConfig(o Options) string { return filepath.Join(o.Home, ".codex", "config.toml") }

func TestPlanCodexWaitFloorStep(t *testing.T) {
	t.Run("off produces nothing", func(t *testing.T) {
		o := wfOptions(t)
		o.CodexWaitFloor = false
		put(t, wfConfig(o), "model = \"x\"\n")
		if p := plan(t, "install", o); p.CodexWaitFloor != nil {
			t.Fatalf("unexpected step %#v", p.CodexWaitFloor)
		}
	})
	t.Run("on plans the keys and the fingerprint", func(t *testing.T) {
		o := wfOptions(t)
		put(t, wfConfig(o), "model = \"x\"\n")
		p := plan(t, "install", o)
		st := p.CodexWaitFloor
		if st == nil || st.Path != wfConfig(o) {
			t.Fatalf("step = %#v", st)
		}
		want := []CodexWaitFloorKey{{"min_wait_timeout_ms", 300000}, {"default_wait_timeout_ms", 600000}}
		if !reflect.DeepEqual(st.Keys, want) {
			t.Fatalf("keys = %#v", st.Keys)
		}
		cur, _ := read(wfConfig(o))
		if st.Expected != finger(cur) {
			t.Fatalf("expected fingerprint %#v, want %#v", st.Expected, finger(cur))
		}
	})
	t.Run("missing file plans an absent fingerprint", func(t *testing.T) {
		o := wfOptions(t)
		p := plan(t, "install", o)
		if p.CodexWaitFloor == nil || p.CodexWaitFloor.Expected.Exists {
			t.Fatalf("step = %#v", p.CodexWaitFloor)
		}
	})
	t.Run("the plan ID covers the step", func(t *testing.T) {
		o := wfOptions(t)
		a := plan(t, "install", o)
		o.CodexWaitFloor = false
		b := plan(t, "install", o)
		if a.ID == b.ID {
			t.Fatal("plan ID ignores the wait floor step")
		}
	})
	t.Run("the plan holds no config bytes", func(t *testing.T) {
		o := wfOptions(t)
		put(t, wfConfig(o), "[mcp_servers.x.env]\nTOKEN = \"SUPERSECRET\"\n")
		p := plan(t, "install", o)
		if strings.Contains(string(encode(p)), "SUPERSECRET") {
			t.Fatal("plan holds config.toml bytes")
		}
	})
	t.Run("already applied still plans, apply is a no-op", func(t *testing.T) {
		o := wfOptions(t)
		put(t, wfConfig(o), wfHeader+"\n"+wfMin+"\n"+wfDefault+"\n")
		if p := plan(t, "install", o); p.CodexWaitFloor == nil {
			t.Fatal("expected a step")
		}
	})
	refusals := map[string]func(t *testing.T, o Options){
		"conflicting value": func(t *testing.T, o Options) { put(t, wfConfig(o), wfHeader+"\nmin_wait_timeout_ms = 1000\n") },
		"ambiguous form":    func(t *testing.T, o Options) { put(t, wfConfig(o), "features = { multi_agent_v2 = true }\n") },
		"symbolic link": func(t *testing.T, o Options) {
			real := filepath.Join(filepath.Dir(o.Home), "real.toml")
			put(t, real, "x = 1\n")
			if err := os.MkdirAll(filepath.Dir(wfConfig(o)), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(real, wfConfig(o)); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, arrange := range refusals {
		t.Run("refuses "+name, func(t *testing.T) {
			o := wfOptions(t)
			arrange(t, o)
			_, err := BuildPlan("install", o)
			if err == nil || !strings.Contains(err.Error(), "wait floor") {
				t.Fatalf("want a wait floor refusal, got %v", err)
			}
		})
	}
	t.Run("needs a user-scope install with codex", func(t *testing.T) {
		o := wfOptions(t)
		o.Hosts = []string{"claude"}
		if _, err := BuildPlan("install", o); err == nil {
			t.Fatal("codex is not a selected host")
		}
		o = wfOptions(t)
		if _, err := BuildPlan("remove", o); err == nil {
			t.Fatal("remove cannot carry the option")
		}
	})
}
