package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Linux procps -g selects sessions, not the reserved process group. A failed
// session lookup must not conceal a group member still in uninterruptible wait.
func TestOpenCodeCredentialGroupQueryUsesExplicitPGID(t *testing.T) {
	for _, mode := range []string{"owned D", "unrelated live", "foreign transient status", "owned transient status", "empty", "malformed", "exit one", "timeout", "missing tool"} {
		t.Run(mode, func(t *testing.T) {
			bin := t.TempDir()
			log := filepath.Join(bin, "query.args")
			script := `#!/bin/sh
printf '%s\n' "$*" >> "$FAKE_QUERY_LOG"
case "$*" in
  "-o stat= -g "*) exit 1 ;;
  "-A -o pgid=,stat=") ;;
  *) exit 2 ;;
esac
case "$FAKE_QUERY_MODE" in
  "owned D") printf '%s D\n999999 R\n' "$FAKE_GROUP" ;;
  "unrelated live") printf '%s Z\n999999 R\n' "$FAKE_GROUP" ;;
  "foreign transient status") printf '%s Z\n999999 ?\n' "$FAKE_GROUP" ;;
  "owned transient status") printf '%s ?\n999999 R\n' "$FAKE_GROUP" ;;
  "empty") exit 0 ;;
  "malformed") printf '%s Z\nmalformed\n' "$FAKE_GROUP" ;;
  "exit one") exit 1 ;;
  "timeout") exec /bin/sleep 3 ;;
esac
`
			if mode != "missing tool" {
				if err := os.WriteFile(filepath.Join(bin, "ps"), []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
			}
			p, err := superviseCredentialProcess(context.Background(), exec.Command("/bin/sh", "-c", "exit 0"))
			if err != nil {
				t.Fatal(err)
			}
			if err := p.start(); err != nil {
				p.close()
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if !p.finished {
					_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
					_ = p.cmd.Wait()
					p.close()
				}
				if p.cmd.ProcessState == nil {
					t.Error("fixture supervisor was not reaped")
				}
			})
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if status, err := p.nativeStatus(ctx); err != nil || status != 0 {
				t.Fatal("fixture supervisor did not finish native command", status, err)
			}
			t.Setenv("PATH", bin)
			t.Setenv("FAKE_QUERY_LOG", log)
			t.Setenv("FAKE_QUERY_MODE", mode)
			t.Setenv("FAKE_GROUP", fmt.Sprint(p.cmd.Process.Pid))
			stopped := p.stop()
			if stopped != (mode == "unrelated live" || mode == "foreign transient status") {
				t.Fatalf("group cleanup incorrectly verified: mode %s, stopped %t", mode, stopped)
			}
			if p.finished != stopped {
				t.Fatal("uncertain group status released supervisor ownership")
			}
			if mode != "missing tool" {
				queries, err := os.ReadFile(log)
				if err != nil {
					t.Fatal(err)
				}
				for _, query := range strings.Split(strings.TrimSpace(string(queries)), "\n") {
					if query != "-A -o pgid=,stat=" {
						t.Fatalf("query did not enumerate explicit PGID/state columns: %s", query)
					}
				}
			}
		})
	}
}

func TestOpenCodeCredentialGroupEnumeration(t *testing.T) {
	for _, tt := range []struct {
		name, data    string
		live, invalid bool
	}{
		{"owned uninterruptible", "42 D\n17 R\n", true, false},
		{"owned runnable", "42 R+\n", true, false},
		{"owned sleeping", "42 Ss\n", true, false},
		{"owned zombie unrelated live", "42 Z\n17 R\n", false, false},
		{"owned zombie foreign transient status", "42 Z\n17 ?\n", false, false},
		{"owned live foreign transient status", "42 D\n17 ?\n", true, false},
		{"owned transient status", "42 ?\n17 R\n", false, true},
		{"only unrelated live", "17 Sl\n19 R+\n", false, false},
		{"owned zombie and live", "42 Z\n42 U\n", true, false},
		{"empty", "", false, true},
		{"whitespace", " \n\t", false, true},
		{"missing group", "D\n", false, true},
		{"extra column", "42 D 17\n", false, true},
		{"non numeric group", "bad D\n", false, true},
		{"signed group", "+42 D\n", false, true},
		{"unrelated kernel group", "0 D\n42 Z\n", false, false},
		{"overflow group", "9999999999999999999999999 D\n", false, true},
		{"unknown state", "42 unknown\n", false, true},
		{"malformed state suffix", "42 Zgarbage\n", false, true},
		{"malformed unrelated row", "42 Z\nother malformed\n", false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			live, err := credentialGroupLive([]byte(tt.data), 42)
			if (err != nil) != tt.invalid || live != tt.live {
				t.Fatalf("live=%t error=%v, expected live=%t invalid=%t", live, err, tt.live, tt.invalid)
			}
		})
	}
}

// Regression: a successful import parent can exit while a descendant still
// owns the credential namespace and recreates a copy after cleanup.
func TestOpenCodeCredentialImportStopsSurvivingWriter(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			bin, home, control := t.TempDir(), t.TempDir(), t.TempDir()
			script := `#!/bin/sh
if [ "$1" = auth ] && [ "$2" = export ]; then printf '{}'; exit 0; fi
if [ "$1" = auth ] && [ "$2" = import ]; then
 cat >/dev/null
 mkdir -p "$HOME/.local/share/opencode"
 printf 'FAKE_COPY' > "$HOME/.local/share/opencode/opencode.db"
 sh -c 'while [ ! -f "$FAKE_TRIGGER" ]; do sleep 0.01; done; mkdir -p "$HOME/.local/share/opencode"; printf "FAKE_RECREATED" > "$HOME/.local/share/opencode/opencode.db"; : > "$FAKE_ACK"' &
 echo $! > "$FAKE_PID"
 exit "$FAKE_EXIT"
fi
exit 1
`
			if err := os.WriteFile(filepath.Join(bin, "opencode"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("FAKE_PID", filepath.Join(control, "pid"))
			t.Setenv("FAKE_TRIGGER", filepath.Join(control, "trigger"))
			t.Setenv("FAKE_ACK", filepath.Join(control, "ack"))
			t.Setenv("FAKE_EXIT", "0")
			if fail {
				t.Setenv("FAKE_EXIT", "3")
			}
			g := &guidanceVariant{shadowHome: home, host: "opencode", integration: "opencode"}
			err := g.importCredential()
			if (err != nil) != fail {
				t.Fatal("unexpected fake import result", err)
			}
			b, err := os.ReadFile(filepath.Join(control, "pid"))
			if err != nil {
				t.Fatal(err)
			}
			var pid int
			if _, err := fmt.Sscanf(strings.TrimSpace(string(b)), "%d", &pid); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL); g.cleanup() })
			g.cleanup()
			if err := os.WriteFile(filepath.Join(control, "trigger"), nil, 0600); err != nil {
				t.Fatal(err)
			}
			time.Sleep(100 * time.Millisecond)
			if _, err := os.Stat(filepath.Join(control, "ack")); !os.IsNotExist(err) {
				t.Fatal("survivor recreated a credential copy after cleanup")
			}
			state, _ := exec.Command("ps", "-o", "stat=", "-p", fmt.Sprint(pid)).Output()
			s := strings.TrimSpace(string(state))
			if s != "" && !strings.HasPrefix(s, "Z") {
				t.Fatal("import descendant still live")
			}
		})
	}
}
