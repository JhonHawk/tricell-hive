package management

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"tricell-hive/tooling/distribution"
)

func TestReleasePiPreflightNeverSchedulesInstallation(t *testing.T) {
	for _, settings := range []string{"", `{"packages":[{"source":"npm:pi-subagents","extensions":[]}]}`, `{"packages":["npm:pi-subagents"]}`} {
		t.Run(settings, func(t *testing.T) {
			o := setup(t)
			o.Hosts = []string{"pi"}
			p := plan(t, "install", o)
			if settings != "" {
				put(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"), settings)
			}
			p.ReleaseUpdate = true
			err := planPiPackage(&p)
			valid := settings == `{"packages":["npm:pi-subagents"]}`
			if valid && (err != nil || p.PiPackage != nil) {
				t.Fatalf("valid declaration: %v %+v", err, p.PiPackage)
			}
			if !valid && (err == nil || !strings.Contains(err.Error(), "hive install")) {
				t.Fatalf("missing separate repair: %v", err)
			}
		})
	}
}

// The package and core engine are real; failpoints interrupt durable boundaries,
// rather than replacing manager behavior with in-process doubles.
func releaseUpdateFixture(t *testing.T, hosts []string) (Options, UpdatePlan) {
	t.Helper()
	o := setup(t)
	o.Hosts = hosts
	if hostsIncludePi(hosts) {
		put(t, filepath.Join(o.Home, ".pi", "agent", "settings.json"), `{"packages":["npm:pi-subagents"]}`)
	}
	apply(t, plan(t, "install", o))
	root := filepath.Join(filepath.Dir(o.Source), "preparation", "hive-1.2.3-test")
	files := map[string]string{GlobalSource: "# New release\n", SkillSource: get(t, filepath.Join(o.Source, SkillSource)), "bin/hive": "#!/bin/sh\nexit 1\n", "bin/hive.sha256": "fixture\n", "install.sh": "offline fixture\n", "platform": runtime.GOOS + "/" + runtime.GOARCH, "integrations/agent-profiles.json": "{}", "LICENSE": "fixture", "THIRD_PARTY_NOTICES.md": "fixture"}
	for path, data := range files {
		put(t, filepath.Join(root, path), data)
	}
	if err := os.Chmod(filepath.Join(root, "bin", "hive"), 0700); err != nil {
		t.Fatal(err)
	}
	boot := distribution.UpdateBootstrap{Protocol: 1, ProductVersion: "1.2.3", Platform: runtime.GOOS + "/" + runtime.GOARCH, SourceID: strings.Repeat("a", 64), ExecutableSHA256: distribution.Digest([]byte(files["bin/hive"])), ManifestVersion: 1, MinStateVersion: 1, MaxStateVersion: stateVersion}
	data, _ := json.Marshal(boot)
	put(t, filepath.Join(root, distribution.UpdateBootstrapName), string(data))
	inventory, err := distribution.Files(root)
	if err != nil {
		t.Fatal(err)
	}
	m := distribution.Manifest{UpdateProtocol: 1, Version: 1, Platform: boot.Platform, SourceID: boot.SourceID, ProductVersion: boot.ProductVersion, Files: inventory}
	data, _ = json.Marshal(m)
	put(t, filepath.Join(root, distribution.ManifestName), string(data))
	final := filepath.Join(filepath.Dir(o.Source), "packages", filepath.Base(root))
	b := distribution.UpdatePackageBinding{Artifact: distribution.UpdateArtifact{ProductVersion: boot.ProductVersion, Platform: boot.Platform, SourceID: boot.SourceID, Label: filepath.Base(root), ArchiveSHA256: strings.Repeat("b", 64), ManifestSHA256: distribution.Digest(data), Bootstrap: boot}, PreparationDirectory: root, PackageDirectory: final, PreparationExecutor: distribution.ExecutorIdentity{Path: filepath.Join(root, "bin", "hive"), SHA256: boot.ExecutableSHA256}, TargetExecutable: distribution.ExecutorIdentity{Path: filepath.Join(final, "bin", "hive"), SHA256: boot.ExecutableSHA256}}
	p, err := PrepareReleaseUpdate(UpdatePreparePayload{Package: b, Home: o.Home, StateDir: o.StateDir}, "test-operation")
	if err != nil {
		t.Fatal(err)
	}
	absent(t, final)
	// Simulate the distribution boundary retaining exactly its verified package.
	if err = os.MkdirAll(filepath.Dir(final), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.CopyFS(final, os.DirFS(root)); err != nil {
		t.Fatal(err)
	}
	if err = filepath.WalkDir(root, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, _ := filepath.Rel(root, path)
		info, e := d.Info()
		if e != nil {
			return e
		}
		return os.Chmod(filepath.Join(final, rel), info.Mode().Perm())
	}); err != nil {
		t.Fatal(err)
	}
	if err = os.Chmod(filepath.Join(final, "bin", "hive"), 0700); err != nil {
		t.Fatal(err)
	}
	return o, p
}

func TestReleaseUpdateCommitBoundaryRecovery(t *testing.T) {
	for _, stage := range []string{"update-prepared", "update-core-pending", "prepared", "write:0", "state", "update-child-response", "update-core-committed", "update-activation-pending", "command-before-rename", "command-renamed", "update-activated"} {
		t.Run(stage, func(t *testing.T) {
			o, p := releaseUpdateFixture(t, []string{"codex", "claude"})
			before := get(t, filepath.Join(o.Home, ".codex", "AGENTS.md"))
			injected := false
			e := Engine{failpoint: func(at string) error {
				if at == stage {
					injected = true
					return fmt.Errorf("interrupted at %s", at)
				}
				return nil
			}}
			if _, err := e.ApplyReleaseUpdate(p); err == nil || !injected {
				t.Fatalf("boundary not exercised: %v %v", injected, err)
			}
			j, err := PendingUpdate(o.StateDir)
			if err != nil || j == nil {
				t.Fatalf("missing parent: %v", err)
			}
			if _, err = lock(o.StateDir); err == nil {
				t.Fatal("competing mutator bypassed update gate")
			}
			phase, err := (Engine{}).RecoverReleaseUpdate(o.StateDir)
			if err != nil {
				t.Fatal(err)
			}
			committed := stage == "update-child-response" || stage == "update-core-committed" || stage == "update-activation-pending" || stage == "command-before-rename" || stage == "command-renamed" || stage == "update-activated"
			if committed {
				if phase != UpdateActivated {
					t.Fatal(phase)
				}
				link, err := os.Readlink(p.Command.Path)
				if err != nil || link != p.Package.TargetExecutable.Path {
					t.Fatalf("not activated: %s %v", link, err)
				}
				if get(t, filepath.Join(o.Home, ".codex", "AGENTS.md")) == before {
					t.Fatal("rolled back committed core")
				}
				phase, err = (Engine{}).ApplyReleaseUpdate(p)
				if err != nil || phase != UpdateActivated {
					t.Fatalf("reapply: %s %v", phase, err)
				}
			} else {
				if phase != UpdateRecovered {
					t.Fatal(phase)
				}
				absent(t, p.Command.Path)
				if get(t, filepath.Join(o.Home, ".codex", "AGENTS.md")) != before {
					t.Fatal("incomplete core not restored")
				}
			}
			j, err = PendingUpdate(o.StateDir)
			if err != nil || j != nil {
				t.Fatalf("gate still pending: %v", err)
			}
		})
	}
}

func TestReleaseUpdatePreservesCommandAndPackageConflicts(t *testing.T) {
	for _, drift := range []string{"command", "package", "saved Pi step", "stale state"} {
		t.Run(drift, func(t *testing.T) {
			o, p := releaseUpdateFixture(t, []string{"codex"})
			before := get(t, filepath.Join(o.Home, ".codex", "AGENTS.md"))
			switch drift {
			case "command":
				put(t, p.Command.Path, "user command")
			case "package":
				put(t, p.Package.TargetExecutable.Path, "altered executable")
			case "saved Pi step":
				var core Plan
				json.Unmarshal(p.CorePlan, &core)
				core.PiPackage = &PiPackageStep{Action: PackageInstall, Source: "npm:pi-subagents"}
				core.ID = planID(core)
				p.CorePlan, _ = json.Marshal(core)
				p.CorePlanID = core.ID
				p, _ = FreezeUpdatePlan(p)
			case "stale state":
				put(t, filepath.Join(o.StateDir, "state.json"), `{"version":999,"records":{}}`)
			}
			if _, err := (Engine{}).ApplyReleaseUpdate(p); err == nil {
				t.Fatal("accepted conflict")
			}
			if get(t, filepath.Join(o.Home, ".codex", "AGENTS.md")) != before {
				t.Fatal("conflict changed core")
			}
		})
	}
}

func TestReleaseUpdateHookPreviewAndApplyRequireFinalPackage(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"claude", "codex", "cursor", "grok", "opencode", "pi"})
	var core Plan
	if err := json.Unmarshal(p.CorePlan, &core); err != nil {
		t.Fatal(err)
	}
	for _, ch := range core.Changes {
		if ch.Target.Kind == "hook" && ch.After != nil {
			if !strings.Contains(string(ch.After.Managed), p.Package.TargetExecutable.Path) || strings.Contains(string(ch.After.Managed), p.Package.PreparationExecutor.Path) {
				t.Fatal("temporary hook destination")
			}
		}
	}
	if err := os.RemoveAll(p.Package.PackageDirectory); err != nil {
		t.Fatal(err)
	}
	if _, err := (Engine{}).ApplyReleaseUpdate(p); err == nil {
		t.Fatal("applied without final executable")
	}
	absent(t, p.Command.Path)
	if len(piCalls) != 0 {
		t.Fatal("release invoked Pi")
	}
	_ = o
}

func TestReleaseUpdateRejectsUncoordinatedCoreAndReservedCollision(t *testing.T) {
	for _, mode := range []string{"direct core", "onboarding providers", "reserved collision"} {
		t.Run(mode, func(t *testing.T) {
			o, p := releaseUpdateFixture(t, []string{"codex"})
			before := get(t, filepath.Join(o.Home, ".codex", "AGENTS.md"))
			var core Plan
			json.Unmarshal(p.CorePlan, &core)
			var err error
			switch mode {
			case "direct core":
				_, err = (Engine{}).Apply(core)
			case "onboarding providers":
				_, err = (Engine{}).Onboard(core, []ExternalStep{{ID: "provider", Payload: json.RawMessage(`{}`)}}, &fakeExternal{})
			case "reserved collision":
				put(t, filepath.Join(o.StateDir, "transactions", p.CoreTransactionID+".json"), "prior journal")
				_, err = (Engine{}).ApplyReleaseUpdate(p)
			}
			if err == nil {
				t.Fatal("accepted uncoordinated update or journal overwrite")
			}
			if get(t, filepath.Join(o.Home, ".codex", "AGENTS.md")) != before {
				t.Fatal("changed core")
			}
			if mode == "reserved collision" && get(t, filepath.Join(o.StateDir, "transactions", p.CoreTransactionID+".json")) != "prior journal" {
				t.Fatal("overwrote evidence")
			}
		})
	}
}

func TestManagedCommandLifecyclePreservesUnownedPaths(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"codex"})
	if _, err := (Engine{}).ApplyReleaseUpdate(p); err != nil {
		t.Fatal(err)
	}
	p.Command, _ = InspectUpdateCommand(p.Command.Path, o.StateDir)
	if err := ActivateManagedCommand(p.Command, p.Package, o.StateDir, p.CoreTransactionID); err != nil {
		t.Fatal(err)
	}
	command, err := InspectUpdateCommand(p.Command.Path, o.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	if err = ActivateManagedCommand(command, p.Package, o.StateDir, p.CoreTransactionID); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(command.Path); err != nil {
		t.Fatal(err)
	}
	put(t, command.Path, "replacement owned by user")
	if err = RemoveManagedCommand(o.StateDir); err == nil {
		t.Fatal("unowned replacement not reported")
	}
	if get(t, command.Path) != "replacement owned by user" {
		t.Fatal("removed replacement")
	}
	absent(t, commandReceiptPath(o.StateDir))
	if _, err = InspectUpdateCommand(command.Path, o.StateDir); err == nil {
		t.Fatal("adopted replacement")
	}
}

func TestPackageInstallationSharesDurableCoordinator(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"codex"})
	core, err := BuildPackageInstallPlan(o, p.Package)
	if err != nil {
		t.Fatal(err)
	}
	parent, err := PreparePackageInstallation(core, p.Package, p.Command, "install-operation")
	if err != nil {
		t.Fatal(err)
	}
	failed := Engine{failpoint: func(stage string) error {
		if stage == "update-child-response" {
			return fmt.Errorf("lost response")
		}
		return nil
	}}
	if _, err = failed.ApplyReleaseUpdate(parent); err == nil {
		t.Fatal("expected interruption")
	}
	if phase, err := (Engine{}).RecoverReleaseUpdate(o.StateDir); err != nil || phase != UpdateActivated {
		t.Fatalf("install recovery: %s %v", phase, err)
	}
	o.Source = p.Package.PackageDirectory
	apply(t, plan(t, "remove", o))
	absent(t, p.Command.Path)
	absent(t, commandReceiptPath(o.StateDir))
	if _, err = os.Stat(p.Package.PackageDirectory); err != nil {
		t.Fatal("removed retained package")
	}
}

func TestReleasePendingPinsChildAgainstPruningAndPreservesChangedLink(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"codex"})
	e := Engine{failpoint: func(stage string) error {
		if stage == "update-child-response" {
			return fmt.Errorf("lost response")
		}
		return nil
	}}
	if _, err := e.ApplyReleaseUpdate(p); err == nil {
		t.Fatal("expected interruption")
	}
	old := time.Now().Add(-48 * time.Hour)
	childPath := filepath.Join(o.StateDir, "transactions", p.CoreTransactionID+".json")
	if err := os.Chtimes(childPath, old, old); err != nil {
		t.Fatal(err)
	}
	seedJournals(t, o.StateDir, "committed", RetainedJournals+5)
	if warning := pruneJournals(o.StateDir, ""); warning != "" {
		t.Fatal(warning)
	}
	if _, err := os.Stat(childPath); err != nil {
		t.Fatal("pruned update commit evidence")
	}
	put(t, p.Command.Path, "foreign replacement")
	if _, err := (Engine{}).RecoverReleaseUpdate(o.StateDir); err == nil {
		t.Fatal("recovery overwrote command conflict")
	}
	if get(t, p.Command.Path) != "foreign replacement" {
		t.Fatal("foreign command changed")
	}
	if !strings.Contains(get(t, filepath.Join(o.Home, ".codex", "AGENTS.md")), "New release") {
		t.Fatal("committed core undone")
	}
	j, err := PendingUpdate(o.StateDir)
	if err != nil || j == nil {
		t.Fatalf("lost recovery gate: %v", err)
	}
}

func TestReleaseUpdateLockOrderRejectsCompetingActors(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"codex"})
	for _, name := range []string{"update-lock", "lock"} {
		unlock, err := lockNamed(o.StateDir, name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = (Engine{}).ApplyReleaseUpdate(p); err == nil {
			unlock()
			t.Fatal("ignored held lock")
		}
		unlock()
		// Failure acquiring the second lock must release the first one.
		unlock, err = lockNamed(o.StateDir, "update-lock")
		if err != nil {
			t.Fatal("leaked update lock", err)
		}
		unlock()
	}
	if _, err := (Engine{}).ApplyReleaseUpdate(p); err != nil {
		t.Fatal(err)
	}
}

func TestPackageProvidersRequireActivatedUnchangedCore(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"codex"})
	core, err := BuildPackageInstallPlan(o, p.Package)
	if err != nil {
		t.Fatal(err)
	}
	steps := []ExternalStep{{ID: "explicit-provider", Payload: json.RawMessage(`{}`)}}
	if _, err = (Engine{}).OnboardPackageProviders(core, steps, &fakeExternal{}); err == nil {
		t.Fatal("providers ran before activation")
	}
	parent, err := PreparePackageInstallation(core, p.Package, p.Command, "install-providers")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = (Engine{}).ApplyReleaseUpdate(parent); err != nil {
		t.Fatal(err)
	}
	core, err = BuildPackageInstallPlan(o, p.Package)
	if err != nil {
		t.Fatal(err)
	}
	if result, err := (Engine{}).OnboardPackageProviders(core, steps, &fakeExternal{}); err != nil || result.Phase != "completed" {
		t.Fatalf("provider continuation: %+v %v", result, err)
	}
}

func TestReleaseRecoveryUsesOnlyItsBoundChild(t *testing.T) {
	for _, marker := range []string{"missing before commit", "different child", "left after commit"} {
		t.Run(marker, func(t *testing.T) {
			o, p := releaseUpdateFixture(t, []string{"codex"})
			before := get(t, filepath.Join(o.Home, ".codex", "AGENTS.md"))
			stage := "state"
			if marker == "left after commit" {
				stage = "update-child-response"
			}
			e := Engine{failpoint: func(at string) error {
				if at == stage {
					return fmt.Errorf("interrupted")
				}
				return nil
			}}
			if _, err := e.ApplyReleaseUpdate(p); err == nil {
				t.Fatal("not interrupted")
			}
			path := filepath.Join(o.StateDir, "pending.json")
			switch marker {
			case "missing before commit":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			case "different child":
				put(t, path, `{"ID":"00000000000000000000000000000000"}`)
			case "left after commit":
				put(t, path, fmt.Sprintf(`{"ID":%q}`, p.CoreTransactionID))
			}
			phase, err := (Engine{}).RecoverReleaseUpdate(o.StateDir)
			if marker == "different child" {
				if err == nil {
					t.Fatal("accepted unrelated pending child")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			absent(t, path)
			if marker == "missing before commit" && (phase != UpdateRecovered || get(t, filepath.Join(o.Home, ".codex", "AGENTS.md")) != before) {
				t.Fatal("did not recover bound incomplete writes")
			}
		})
	}
}

func TestReleaseUpdatePiSettingsRecheckedBeforeWrites(t *testing.T) {
	for _, settings := range []string{`{"packages":[]}`, `{"packages":[{"source":"npm:pi-subagents","extensions":[]}]}`} {
		t.Run(settings, func(t *testing.T) {
			o, p := releaseUpdateFixture(t, []string{"pi"})
			before := get(t, filepath.Join(o.Home, ".pi", "agent", "AGENTS.md"))
			path := filepath.Join(o.Home, ".pi", "agent", "settings.json")
			put(t, path, settings)
			if _, err := (Engine{}).ApplyReleaseUpdate(p); err == nil || !strings.Contains(err.Error(), "hive install") {
				t.Fatalf("repair not required: %v", err)
			}
			if get(t, path) != settings || get(t, filepath.Join(o.Home, ".pi", "agent", "AGENTS.md")) != before {
				t.Fatal("preflight changed settings/core")
			}
			if len(piCalls) != 0 {
				t.Fatal("invoked Pi during release update")
			}
			absent(t, p.Command.Path)
		})
	}
}

func TestReleaseRecoveryPreservesPreviousOwnedCommand(t *testing.T) {
	for _, stage := range []string{"state", "update-child-response"} {
		t.Run(stage, func(t *testing.T) {
			o, p := releaseUpdateFixture(t, []string{"codex"})
			old := filepath.Join(o.Home, "packages", "previous", "bin", "hive")
			put(t, old, "previous executable")
			if err := os.MkdirAll(filepath.Dir(p.Command.Path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(old, p.Command.Path); err != nil {
				t.Fatal(err)
			}
			fp, err := commandFingerprint(p.Command.Path)
			if err != nil {
				t.Fatal(err)
			}
			if err = writeJSON(commandReceiptPath(o.StateDir), CommandReceipt{Path: p.Command.Path, Fingerprint: fp, PackageDirectory: filepath.Dir(filepath.Dir(old))}); err != nil {
				t.Fatal(err)
			}
			p.Command, err = InspectUpdateCommand(p.Command.Path, o.StateDir)
			if err != nil {
				t.Fatal(err)
			}
			p, err = FreezeUpdatePlan(p)
			if err != nil {
				t.Fatal(err)
			}
			e := Engine{failpoint: func(at string) error {
				if at == stage {
					return fmt.Errorf("interrupted")
				}
				return nil
			}}
			if _, err = e.ApplyReleaseUpdate(p); err == nil {
				t.Fatal("not interrupted")
			}
			if _, err = (Engine{}).RecoverReleaseUpdate(o.StateDir); err != nil {
				t.Fatal(err)
			}
			link, err := os.Readlink(p.Command.Path)
			if err != nil {
				t.Fatal(err)
			}
			if stage == "state" && link != old {
				t.Fatal("previous command not preserved")
			}
			if stage == "update-child-response" && link != p.Package.TargetExecutable.Path {
				t.Fatal("committed update not activated")
			}
			if get(t, old) != "previous executable" {
				t.Fatal("old package modified")
			}
		})
	}
}

func TestReleaseRecoveryReusesRecordedCommandStagingPath(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"codex"})
	e := Engine{failpoint: func(stage string) error {
		if stage == "update-activation-pending" {
			return fmt.Errorf("interrupted")
		}
		return nil
	}}
	if _, err := e.ApplyReleaseUpdate(p); err == nil {
		t.Fatal("not interrupted")
	}
	if err := os.MkdirAll(filepath.Dir(p.Command.Path), 0700); err != nil {
		t.Fatal(err)
	}
	staged := filepath.Join(filepath.Dir(p.Command.Path), ".hive-link-"+p.CoreTransactionID)
	if err := os.Symlink(p.Package.TargetExecutable.Path, staged); err != nil {
		t.Fatal(err)
	}
	if _, err := (Engine{}).RecoverReleaseUpdate(o.StateDir); err != nil {
		t.Fatal(err)
	}
	absent(t, staged)
}

func TestReleasePendingStatusBeforeFutureCoreDecode(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"codex"})
	e := Engine{failpoint: func(stage string) error {
		if stage == "update-child-response" {
			return fmt.Errorf("lost response")
		}
		return nil
	}}
	if _, err := e.ApplyReleaseUpdate(p); err == nil {
		t.Fatal("not interrupted")
	}
	put(t, filepath.Join(o.StateDir, "state.json"), `{"version":999,"future":true}`)
	if _, err := Pending(o.StateDir); err == nil || !strings.Contains(err.Error(), p.Package.TargetExecutable.Path) {
		t.Fatalf("mutator pending API lost bootstrap route: %v", err)
	}
	rows, err := Status(o)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range rows {
		if row.Kind == "update" && row.Status == "recovery_required" && strings.Contains(row.Reason, p.Package.TargetExecutable.Path) {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing pending bootstrap status: %+v", rows)
	}
}

func TestCommittedFinalRemovalRecoversCommandCleanup(t *testing.T) {
	for _, stage := range []string{"command-remove-pending", "command-removed"} {
		t.Run(stage, func(t *testing.T) {
			o, p := releaseUpdateFixture(t, []string{"codex"})
			if _, err := (Engine{}).ApplyReleaseUpdate(p); err != nil {
				t.Fatal(err)
			}
			remove := plan(t, "remove", o)
			hit := false
			e := Engine{failpoint: func(at string) error {
				if at == stage {
					hit = true
					return fmt.Errorf("interrupted cleanup")
				}
				return nil
			}}
			if _, err := e.Apply(remove); err == nil || !hit {
				t.Fatalf("cleanup boundary not exercised: %v", err)
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			absent(t, p.Command.Path)
			absent(t, commandReceiptPath(o.StateDir))
			absent(t, filepath.Join(o.StateDir, "pending.json"))
		})
	}
}

func TestReleaseRecoveryDoesNotRequireRepairingUntouchedPiSettings(t *testing.T) {
	for _, stage := range []string{"state", "update-child-response"} {
		t.Run(stage, func(t *testing.T) {
			o, p := releaseUpdateFixture(t, []string{"pi"})
			e := Engine{failpoint: func(at string) error {
				if at == stage {
					return fmt.Errorf("interrupted")
				}
				return nil
			}}
			if _, err := e.ApplyReleaseUpdate(p); err == nil {
				t.Fatal("not interrupted")
			}
			path := filepath.Join(o.Home, ".pi", "agent", "settings.json")
			put(t, path, `{"packages":[]}`)
			if _, err := (Engine{}).RecoverReleaseUpdate(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if get(t, path) != `{"packages":[]}` || len(piCalls) != 0 {
				t.Fatal("recovery changed Pi settings or invoked Pi")
			}
		})
	}
}

func TestCompletedReleasePlanRefusesLaterCommandDrift(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"codex"})
	if _, err := (Engine{}).ApplyReleaseUpdate(p); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(p.Command.Path); err != nil {
		t.Fatal(err)
	}
	put(t, p.Command.Path, "user replacement")
	if _, err := (Engine{}).ApplyReleaseUpdate(p); err == nil {
		t.Fatal("reported activation after command drift")
	}
	if get(t, p.Command.Path) != "user replacement" {
		t.Fatal("overwrote replacement")
	}
	_ = o
}

func TestReleasePreparationPreservesNativeHostOverrides(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"codex"})
	apply(t, plan(t, "remove", o))
	custom := filepath.Join(o.Home, "custom-codex")
	t.Setenv("HOME", o.Home)
	t.Setenv("CODEX_HOME", custom)
	native := o
	native.Home = ""
	apply(t, plan(t, "install", native))
	prepared, err := PrepareReleaseUpdate(UpdatePreparePayload{Package: p.Package, Home: o.Home, StateDir: o.StateDir, NativeHome: true}, "native-operation")
	if err != nil {
		t.Fatal(err)
	}
	var core Plan
	if err = json.Unmarshal(prepared.CorePlan, &core); err != nil {
		t.Fatal(err)
	}
	if core.Config.Synthetic || core.Config.CodexHome != custom {
		t.Fatalf("lost native home configuration: %+v", core.Config)
	}
}
