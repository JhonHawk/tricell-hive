package main

import (
	"archive/tar"
	"bytes"
	tea "charm.land/bubbletea/v2"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/version"
)

type updateHTTPBoundary func(*http.Request) (*http.Response, error)

func (f updateHTTPBoundary) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Exercise the launcher against a real locally built target and the real core.
// Only HTTPS is doubled; no public artifact is downloaded or executed.
func TestReleaseUpdateCLI(t *testing.T) {
	previousVersion, previousTransport := version.Current, http.DefaultTransport
	t.Cleanup(func() { version.Current = previousVersion; http.DefaultTransport = previousTransport })
	version.Current = "1.0.0"
	requests := 0
	http.DefaultTransport = updateHTTPBoundary(func(r *http.Request) (*http.Response, error) {
		requests++
		return nil, fmt.Errorf("unexpected network")
	})
	var out bytes.Buffer
	if err := update(nil, strings.NewReader(""), &out, false); err == nil || !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatalf("nonTTY: %v", err)
	}
	if requests != 0 {
		t.Fatal("nonTTY downloaded")
	}
	version.Current = "dev"
	if err := update([]string{"--dry-run"}, strings.NewReader(""), &out, false); err == nil || !strings.Contains(err.Error(), "--source") {
		t.Fatalf("dev: %v", err)
	}
	if requests != 0 {
		t.Fatal("dev downloaded")
	}
	version.Current = "1.0.0"
	e := newUpdateEnv(t)
	writeUpdateCatalog(t, e.repo, skillBody("Old release.\n"))
	installDirect(t, e.home, e.stateDir, e.repo, []string{"codex"})
	buildPackage := func(productVersion, body string) (string, distribution.UpdateBootstrap) {
		t.Helper()
		root, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		writeUpdateCatalog(t, root, skillBody(body))
		putCharacterization(t, filepath.Join(root, "install.sh"), "offline fixture\n")
		putCharacterization(t, filepath.Join(root, "platform"), runtime.GOOS+"/"+runtime.GOARCH)
		putCharacterization(t, filepath.Join(root, "LICENSE"), "fixture")
		putCharacterization(t, filepath.Join(root, "THIRD_PARTY_NOTICES.md"), "fixture")
		binary := filepath.Join(root, "bin", "hive")
		if err := os.MkdirAll(filepath.Dir(binary), 0700); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		build := exec.CommandContext(ctx, "go", "build", "-ldflags=-X tricell-hive/tooling/version.Current="+productVersion, "-o", binary, ".")
		if b, err := build.CombinedOutput(); err != nil {
			t.Fatalf("local target build: %v\n%s", err, b)
		}
		executable, err := os.ReadFile(binary)
		if err != nil {
			t.Fatal(err)
		}
		putCharacterization(t, filepath.Join(root, "bin", "hive.sha256"), distribution.Digest(executable)+"  hive\n")
		_, schema := management.UpdateStateVersions()
		boot := distribution.UpdateBootstrap{Protocol: 1, ProductVersion: productVersion, Platform: runtime.GOOS + "/" + runtime.GOARCH, SourceID: strings.Repeat("a", 64), ExecutableSHA256: distribution.Digest(executable), ManifestVersion: 1, MinStateVersion: 1, MaxStateVersion: schema}
		data, _ := json.Marshal(boot)
		putCharacterization(t, filepath.Join(root, distribution.UpdateBootstrapName), string(data))
		files, err := distribution.Files(root)
		if err != nil {
			t.Fatal(err)
		}
		manifest := distribution.Manifest{Version: 1, UpdateProtocol: 1, ProductVersion: boot.ProductVersion, Platform: boot.Platform, SourceID: boot.SourceID, Files: files}
		data, _ = json.Marshal(manifest)
		putCharacterization(t, filepath.Join(root, distribution.ManifestName), string(data))
		return root, boot
	}
	root, boot := buildPackage("1.2.3", "New release.\n")
	oldPackage, _ := buildPackage("1.0.0", "Old release.\n")
	var err error
	var data []byte
	label := "hive-1.2.3-" + strings.ReplaceAll(boot.Platform, "/", "-")
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, _ := filepath.Rel(root, path)
		name := label
		if rel != "." {
			name += "/" + filepath.ToSlash(rel)
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		h, e := tar.FileInfoHeader(info, "")
		if e != nil {
			return e
		}
		h.Name = name
		if e = tw.WriteHeader(h); e != nil {
			return e
		}
		if !d.IsDir() {
			b, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			_, e = tw.Write(b)
			return e
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err = gz.Close(); err != nil {
		t.Fatal(err)
	}
	releaseTransport := updateHTTPBoundary(func(r *http.Request) (*http.Response, error) {
		requests++
		var body []byte
		switch {
		case strings.HasSuffix(r.URL.Path, "/latest"):
			body = []byte(`{"tag_name":"v1.2.3"}`)
		case strings.HasSuffix(r.URL.Path, ".tar.gz.sha256"):
			body = []byte(distribution.Digest(archive.Bytes()) + "  " + label + ".tar.gz\n")
		case strings.HasSuffix(r.URL.Path, ".tar.gz"):
			body = archive.Bytes()
		default:
			return nil, fmt.Errorf("unexpected request %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body)), ContentLength: int64(len(body)), Header: make(http.Header), Request: r}, nil
	})
	http.DefaultTransport = releaseTransport
	base := []string{"--home", e.home, "--state-dir", e.stateDir}
	store, err := management.ReleasePackageStore(e.home)
	if err != nil {
		t.Fatal(err)
	}
	before := collectFiles(t, e.home)
	for _, tc := range []struct {
		name        string
		args        []string
		answer      string
		interactive bool
	}{{"preview", []string{"--dry-run"}, "", false}, {"cancel", nil, "n\n", true}} {
		t.Run(tc.name, func(t *testing.T) {
			out.Reset()
			if err := update(append(append([]string{}, base...), tc.args...), strings.NewReader(tc.answer), &out, tc.interactive); err != nil {
				t.Fatalf("%v\n%s", err, out.String())
			}
			if _, err := os.Stat(store); !os.IsNotExist(err) {
				t.Fatalf("preview/cancel retained package: %v", err)
			}
			if fmt.Sprint(before) != fmt.Sprint(collectFiles(t, e.home)) {
				t.Fatal("preview/cancel changed home")
			}
			if !strings.Contains(out.String(), "1.0.0 → 1.2.3") {
				t.Fatal(out.String())
			}
		})
	}

	t.Run("terminal confirmation", func(t *testing.T) {
		v := newUpdateView(appConfig{Options: management.Options{Scope: "user", Home: e.home, StateDir: e.stateDir}})
		cmd := v.resolve()
		msg := cmd().(updatePlannedMsg)
		_, act := v.onPlanned(msg)
		confirm, ok := act.push.(*confirmView)
		if !ok {
			t.Fatalf("release did not reach one confirmation: %+v %s", act, v.message)
		}
		th := newAppTheme(false, true)
		for _, width := range []int{80, 40} {
			confirm.Resize(width, 20)
			screen := confirm.View(viewCtx{Theme: &th, Width: width, Height: 20})
			if !strings.Contains(screen, "Update Hive release") || !strings.Contains(screen, "1.0.0 → 1.2.3") {
				t.Fatalf("missing complete summary: %s", screen)
			}
			t.Logf("confirmation %dx24:\n%s", width, screen)
		}
		temporary := msg.release.plan.Package.PreparationDirectory
		confirm.Update(keyMsg(t, "esc"))
		if _, err := os.Stat(temporary); !os.IsNotExist(err) {
			t.Fatalf("cancel retained temporary package: %v", err)
		}
	})

	for _, kind := range []string{"keyboard", "SIGINT", "QuitMsg"} {
		t.Run("quit confirmed preview/"+kind, func(t *testing.T) {
			cfg := appConfig{Options: management.Options{Scope: "user", Home: e.home, StateDir: e.stateDir}}
			v := newUpdateView(cfg)
			cmd := v.resolve()
			msg := cmd().(updatePlannedMsg)
			_, act := v.onPlanned(msg)
			if act.push == nil {
				t.Fatalf("no confirmation: %s", v.message)
			}
			temporary := msg.release.plan.Package.PreparationDirectory
			root := newAppModel(cfg)
			root.width = 80
			root.height = 24
			root.stack = append(root.stack, v, act.push)
			var quit tea.Msg = keyMsg(t, "ctrl+c")
			if kind == "SIGINT" {
				quit = appFilter(root, tea.InterruptMsg{})
			}
			if kind == "QuitMsg" {
				quit = appFilter(root, tea.QuitMsg{})
			}
			_, cmd = root.Update(quit)
			if cmd == nil {
				t.Fatal("quit did not schedule cleanup")
			}
			ready := cmd()
			if _, ok := ready.(tea.QuitMsg); ok {
				t.Fatal("quit skipped cleanup completion")
			}
			_, cmd = root.Update(ready)
			if cmd == nil {
				t.Fatal("cleanup did not permit quit")
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Fatal("cleanup did not finish quit")
			}
			if _, err := os.Stat(temporary); !os.IsNotExist(err) {
				t.Fatalf("quit retained prepared bytes: %v", err)
			}
			if _, err := os.Stat(store); !os.IsNotExist(err) {
				t.Fatalf("quit retained persistent package: %v", err)
			}
		})
	}
	t.Run("pending recovery view", func(t *testing.T) {
		prepared, err := prepareReleaseUpdate(updateFlags{Home: e.home, StateDir: e.stateDir})
		if err != nil {
			t.Fatal(err)
		}
		defer prepared.packageFiles.Cleanup()
		if err = prepared.packageFiles.Retain(); err != nil {
			t.Fatal(err)
		}
		journal, err := management.SealUpdateJournal(management.UpdateJournal{Version: 1, Plan: prepared.plan, Phase: management.UpdatePrepared, CoreJournalPath: filepath.Join(e.stateDir, "transactions", prepared.plan.CoreTransactionID+".json")})
		if err != nil {
			t.Fatal(err)
		}
		data, _ := json.Marshal(journal)
		putCharacterization(t, filepath.Join(e.stateDir, "update", "pending.json"), string(data))
		http.DefaultTransport = updateHTTPBoundary(func(*http.Request) (*http.Response, error) {
			t.Error("recovery view used network")
			return nil, fmt.Errorf("offline")
		})
		v := newUpdateView(appConfig{Options: management.Options{Scope: "user", Home: e.home, StateDir: e.stateDir}})
		cmd := v.resolve()
		v.onPlanned(cmd().(updatePlannedMsg))
		if v.pendingRelease == nil {
			t.Fatalf("no actionable recovery: %s", v.message)
		}
		th := newAppTheme(false, true)
		v.Resize(80, 20)
		screen := v.View(viewCtx{Theme: &th, Width: 80, Height: 20})
		if !strings.Contains(screen, "Enter recovers offline") {
			t.Fatal(screen)
		}
		t.Logf("pending 80x24:\n%s", screen)
		cmd, act := v.onKey(keyMsg(t, "enter"))
		if cmd == nil || !act.write {
			t.Fatal("recovery did not use protected write path")
		}
		msg := cmd().(updateAppliedMsg)
		v.onApplied(msg)
		if msg.err != nil {
			t.Fatalf("offline recovery: %v", msg.err)
		}
		if pending, err := management.PendingUpdate(e.stateDir); err != nil || pending != nil {
			t.Fatalf("recovery left pending: %+v %v", pending, err)
		}
		http.DefaultTransport = releaseTransport
	})
	planFile := filepath.Join(filepath.Dir(e.repo), "frozen.json")
	out.Reset()
	if err = update(append(base, "--out", planFile), strings.NewReader(""), &out, false); err != nil {
		t.Fatalf("save: %v\n%s", err, out.String())
	}
	data, err = os.ReadFile(planFile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = management.ReadUpdatePlan(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	http.DefaultTransport = updateHTTPBoundary(func(*http.Request) (*http.Response, error) {
		t.Error("saved plan used network")
		return nil, fmt.Errorf("offline")
	})
	if handled, err := routeSavedUpdate(data); !handled || err != nil {
		t.Fatalf("offline apply: %v %v", handled, err)
	}
	link, err := os.Readlink(filepath.Join(e.home, ".local", "bin", "hive"))
	if err != nil || link != filepath.Join(store, label, "bin", "hive") {
		t.Fatalf("activation: %s %v", link, err)
	}
	if b, err := os.ReadFile(e.sharedSkillPath()); err != nil || !strings.Contains(string(b), "New release") {
		t.Fatalf("content: %s %v", b, err)
	}

	t.Run("confirmed complete update", func(t *testing.T) {
		fresh := newUpdateEnv(t)
		hosts := []string{"codex", "claude"}
		installPackage := func(source string) {
			t.Helper()
			args := []string{"--home", fresh.home, "--state-dir", fresh.stateDir, "--hosts", strings.Join(hosts, ","), "--source", source}
			var log bytes.Buffer
			if err := installWithAdapterFactory(args, strings.NewReader("yes\n"), &log, true, coreOnlyAdapterFactory); err != nil {
				t.Fatalf("package installation: %v\n%s", err, log.String())
			}
		}
		installPackage(installPackageFixture(t, false))
		command := filepath.Join(fresh.home, ".local", "bin", "hive")
		if _, err := os.Lstat(command); !os.IsNotExist(err) {
			t.Fatalf("legacy install unexpectedly adopted stable command: %v", err)
		}
		installPackage(oldPackage)
		checkVersion := func(want string) {
			t.Helper()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			output, err := exec.CommandContext(ctx, command, "--version").CombinedOutput()
			if err != nil || strings.TrimSpace(string(output)) != want {
				t.Fatalf("stable command version: %q %v; want %s", output, err, want)
			}
		}
		checkVersion("1.0.0")
		t.Setenv("PATH", t.TempDir())
		for _, tool := range []string{"git", "go"} {
			if path, err := exec.LookPath(tool); err == nil {
				t.Fatalf("%s unexpectedly available at %s", tool, path)
			}
		}
		checkVersion("1.0.0")
		http.DefaultTransport = releaseTransport
		var result bytes.Buffer
		if err := update([]string{"--home", fresh.home, "--state-dir", fresh.stateDir}, strings.NewReader("y\n"), &result, true); err != nil {
			t.Fatalf("complete CLI: %v\n%s", err, result.String())
		}
		if strings.Count(result.String(), "Apply this complete update") != 1 || !strings.Contains(result.String(), "Hive updated (1.0.0 → 1.2.3)") {
			t.Fatalf("wrong confirmation/result: %s", result.String())
		}
		checkVersion("1.2.3")
		if body, err := os.ReadFile(fresh.sharedSkillPath()); err != nil || !strings.Contains(string(body), "New release.") {
			t.Fatalf("updated content: %q %v", body, err)
		}
		options := management.Options{Scope: "user", Home: fresh.home, StateDir: fresh.stateDir, Hosts: hosts}
		registered, err := management.RegisteredHosts(options)
		if err != nil || strings.Join(registered, ",") != "claude,codex" {
			t.Fatalf("registered cohort: %v %v", registered, err)
		}
		entries, err := management.Status(options)
		if err != nil {
			t.Fatal(err)
		}
		for _, host := range hosts {
			found := false
			for _, entry := range entries {
				if entry.Host == host && entry.ProductVersion != "" {
					found = true
					if entry.ProductVersion != "1.2.3" || entry.Status != "installed" {
						t.Fatalf("host did not reach release: %+v", entry)
					}
				}
			}
			if !found {
				t.Fatalf("no versioned resources for registered host %s", host)
			}
		}
		t.Log("legacy package migrated to managed 1.0.0; stable command and claude+codex reached 1.2.3 with Git/Go absent from PATH")

	})
	http.DefaultTransport = releaseTransport
	version.Current = "1.2.3"
	stateBeforeNoop := collectFiles(t, e.stateDir)
	out.Reset()
	if err = update(base, strings.NewReader(""), &out, true); err != nil {
		t.Fatalf("no-op: %v\n%s", err, out.String())
	}
	if fmt.Sprint(stateBeforeNoop) != fmt.Sprint(collectFiles(t, e.stateDir)) {
		t.Fatal("no-op wrote state or a redundant migration")
	}
	if !strings.Contains(out.String(), "already up to date") || strings.Contains(out.String(), "Apply this complete") {
		t.Fatalf("no-op unexpectedly confirmed or migrated: %s", out.String())
	}

}
