package management

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Voice follows the same rule as the Hive block: a voice block deleted by
// hand is written again or dropped without writing, a file that lost its Hive
// block is skipped by "voice set", and edited voice text is still refused.

func removeVoiceBlock(t *testing.T, path string) {
	t.Helper()
	data := []byte(get(t, path))
	a, b, err := blockRange(data, voiceMarkers)
	if err != nil || a < 0 {
		t.Fatalf("no voice block in %s: %v", path, err)
	}
	put(t, path, string(append(append([]byte{}, data[:a]...), data[b:]...)))
}

func voiceInstalled(t *testing.T) Options {
	t.Helper()
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	return o
}

func TestVoiceSetSkipsFilesWithoutAHiveBlock(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	put(t, codexPath(o), "user text only\n")
	p := voicePlan(t, "set", o, jarvisSirSubtle)
	if len(p.Voice) != 1 || p.Voice[0].Path != claudePath(o) {
		t.Fatalf("want only the Claude file, got %+v", p.Voice)
	}
	apply(t, p)
	if get(t, codexPath(o)) != "user text only\n" {
		t.Fatal("the skipped file was touched")
	}
	if !strings.Contains(get(t, claudePath(o)), VoiceBegin) {
		t.Fatal("voice not written to the file that has a Hive block")
	}
}

func TestVoiceFailsOnlyWhenNoFileHasAHiveBlock(t *testing.T) {
	for _, action := range []string{"set", "off"} {
		t.Run(action, func(t *testing.T) {
			o := setup(t)
			voiceSource(t, o)
			apply(t, plan(t, "install", o))
			put(t, codexPath(o), "user text only\n")
			if err := os.Remove(claudePath(o)); err != nil {
				t.Fatal(err)
			}
			_, err := BuildVoicePlan(action, o, jarvisSirSubtle)
			if err == nil || !strings.Contains(err.Error(), "run hive install first") {
				t.Fatalf("want the run hive install first error, got %v", err)
			}
		})
	}
}

func TestVoiceOffRemovesAnOrphanedVoiceBlock(t *testing.T) {
	o := voiceInstalled(t)
	removeHiveBlock(t, codexPath(o))
	if !strings.Contains(get(t, codexPath(o)), VoiceBegin) {
		t.Fatal("setup: the voice block should remain")
	}
	p := voicePlan(t, "off", o, VoiceSetting{})
	apply(t, p)
	if strings.Contains(get(t, codexPath(o)), VoiceBegin) {
		t.Fatal("the orphaned voice block stayed")
	}
	if _, ok := stateFor(t, o).VoiceSpans[codexPath(o)]; ok {
		t.Fatal("the orphaned voice record stayed")
	}
}

func TestVoiceOffDropsTheStaleRecordOfAFileWithNeitherBlock(t *testing.T) {
	o := voiceInstalled(t)
	put(t, codexPath(o), "user text only\n")
	p := voicePlan(t, "off", o, VoiceSetting{})
	apply(t, p)
	if len(stateFor(t, o).VoiceSpans) != 0 {
		t.Fatal("voice records left behind")
	}
	entries, err := Status(o)
	if err != nil {
		t.Fatal(err)
	}
	for _, en := range entries {
		if en.Kind == "voice" && en.Status == "drift" {
			t.Fatalf("voice row still drifts: %+v", en)
		}
	}
	if get(t, codexPath(o)) != "user text only\n" {
		t.Fatal("the file was touched")
	}
	if strings.Contains(get(t, claudePath(o)), VoiceBegin) {
		t.Fatal("voice not removed from the file that still has it")
	}
}

func TestVoiceOffSkipsAFileWithNeitherBlockAndNoRecord(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	put(t, codexPath(o), "user text only\n")
	p := voicePlan(t, "off", o, VoiceSetting{})
	for _, vc := range p.Voice {
		if vc.Path == codexPath(o) {
			t.Fatalf("a file with no Hive block and no voice record must be skipped, got %+v", vc)
		}
	}
}

func TestVoiceSetRestoresADeletedVoiceBlock(t *testing.T) {
	o := voiceInstalled(t)
	want := get(t, codexPath(o))
	removeVoiceBlock(t, codexPath(o))
	p := voicePlan(t, "set", o, jarvisSirSubtle)
	gone := 0
	for _, vc := range p.Voice {
		if vc.Gone {
			gone++
		}
	}
	if gone != 1 {
		t.Fatalf("want one gone voice change, got %d", gone)
	}
	if unchanged, err := PlanUnchanged(p); err != nil || unchanged {
		t.Fatalf("restoring must not count as unchanged: %v %v", unchanged, err)
	}
	apply(t, p)
	if get(t, codexPath(o)) != want {
		t.Fatalf("voice block not restored byte for byte:\n%q\nwant\n%q", get(t, codexPath(o)), want)
	}
}

func TestVoiceOffOfADeletedVoiceBlockWritesNothing(t *testing.T) {
	o := voiceInstalled(t)
	removeVoiceBlock(t, codexPath(o))
	after := get(t, codexPath(o))
	apply(t, voicePlan(t, "off", o, VoiceSetting{}))
	if get(t, codexPath(o)) != after {
		t.Fatal("off wrote to a file whose voice block was already gone")
	}
	if len(stateFor(t, o).VoiceSpans) != 0 {
		t.Fatal("voice records left behind")
	}
}

func TestInstallRestoresDeletedHiveAndVoiceBlocks(t *testing.T) {
	cases := map[string]func(t *testing.T, o Options){
		"voice block only": func(t *testing.T, o Options) { removeVoiceBlock(t, codexPath(o)) },
		"both blocks": func(t *testing.T, o Options) {
			removeVoiceBlock(t, codexPath(o))
			removeHiveBlock(t, codexPath(o))
		},
		"file deleted": func(t *testing.T, o Options) {
			if err := os.Remove(codexPath(o)); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, damage := range cases {
		t.Run(name, func(t *testing.T) {
			o := voiceInstalled(t)
			want := get(t, codexPath(o))
			damage(t, o)
			apply(t, plan(t, "install", o))
			got := get(t, codexPath(o))
			if !strings.Contains(got, Begin) || !strings.Contains(got, VoiceBegin) {
				t.Fatalf("blocks not restored: %q", got)
			}
			if name != "both blocks" && got != want {
				t.Fatalf("restored file differs:\n%q\nwant\n%q", got, want)
			}
			if apply(t, plan(t, "install", o)) != "unchanged" {
				t.Fatal("second install must be unchanged")
			}
			noDrift(t, o)
		})
	}
}

func TestRemoveWithADeletedVoiceBlockKeepsUserText(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	put(t, codexPath(o), "user text\n")
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	removeVoiceBlock(t, codexPath(o))
	apply(t, plan(t, "remove", o))
	if get(t, codexPath(o)) != "user text\n" {
		t.Fatalf("user text not preserved: %q", get(t, codexPath(o)))
	}
	if len(stateFor(t, o).VoiceSpans) != 0 {
		t.Fatal("voice records left behind")
	}
}

func TestRemoveOfOneHostWithADeletedVoiceBlockInASharedFileWritesNothing(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	o.Hosts = []string{"claude", "grok"}
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	if len(stateFor(t, o).Records[claudePath(o)].Consumers) != 2 {
		t.Fatal("setup: claude and grok must share the file")
	}
	want := get(t, claudePath(o))
	removeVoiceBlock(t, claudePath(o))
	damaged := get(t, claudePath(o))
	partial := o
	partial.Hosts = []string{"claude"}
	apply(t, plan(t, "remove", partial))
	if get(t, claudePath(o)) != damaged {
		t.Fatalf("remove wrote to the shared file: %q", get(t, claudePath(o)))
	}
	if len(stateFor(t, o).Records[claudePath(o)].Consumers) != 1 {
		t.Fatal("grok must keep the shared Hive block record")
	}
	// The next install brings the voice block back for the host that stays.
	grok := o
	grok.Hosts = []string{"grok"}
	apply(t, plan(t, "install", grok))
	if get(t, claudePath(o)) != want {
		t.Fatalf("voice block not restored:\n%q\nwant\n%q", get(t, claudePath(o)), want)
	}
	noDrift(t, grok)
}

func TestInterruptedVoiceRestoreRecoversToTheDeletedState(t *testing.T) {
	for _, stage := range []string{"prepared", "first write", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := voiceInstalled(t)
			removeVoiceBlock(t, codexPath(o))
			damaged := get(t, codexPath(o))
			stateBefore := get(t, filepath.Join(o.StateDir, "state.json"))
			p := plan(t, "install", o)
			e := Engine{failpoint: func(s string) error {
				if s == stage || (stage == "first write" && strings.HasPrefix(s, "write:")) {
					return errors.New("injected")
				}
				return nil
			}}
			if _, err := e.Apply(p); err == nil {
				t.Fatal("did not fail")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if get(t, codexPath(o)) != damaged || get(t, filepath.Join(o.StateDir, "state.json")) != stateBefore {
				t.Fatal("recovery did not restore the deleted state")
			}
		})
	}
}

func TestForgedVoiceGoneFailsClosedWhenTheBlockIsStillThere(t *testing.T) {
	o := voiceInstalled(t)
	before := get(t, codexPath(o))
	p := voicePlan(t, "off", o, VoiceSetting{})
	for i := range p.Voice {
		p.Voice[i].Gone = true
	}
	p.ID = planID(p)
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("a forged voice Gone was accepted")
	}
	if get(t, codexPath(o)) != before || len(stateFor(t, o).VoiceSpans) == 0 {
		t.Fatal("a forged voice Gone changed the file or the records")
	}
}

func TestEditedVoiceBlockIsStillRefused(t *testing.T) {
	o := voiceInstalled(t)
	data := get(t, codexPath(o))
	put(t, codexPath(o), strings.Replace(data, "Warm, formal", "Warm, edited", 1))
	for name, build := range map[string]func() error{
		"set":     func() error { _, err := BuildVoicePlan("set", o, jarvisSirSubtle); return err },
		"off":     func() error { _, err := BuildVoicePlan("off", o, VoiceSetting{}); return err },
		"install": func() error { _, err := BuildPlan("install", o); return err },
		"remove":  func() error { _, err := BuildPlan("remove", o); return err },
	} {
		if build() == nil {
			t.Fatalf("%s accepted an edited voice block", name)
		}
	}
}

func TestVoiceSetReportsTheFilesItSkipped(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	put(t, codexPath(o), "user text only\n")
	p := voicePlan(t, "set", o, jarvisSirSubtle)
	if len(p.VoiceSkipped) != 1 || p.VoiceSkipped[0] != codexPath(o) {
		t.Fatalf("want the codex file reported as skipped, got %v", p.VoiceSkipped)
	}
	if p2 := voicePlan(t, "set", setupFresh(t), jarvisSirSubtle); len(p2.VoiceSkipped) != 0 {
		t.Fatalf("nothing skipped expected, got %v", p2.VoiceSkipped)
	}
}

func setupFresh(t *testing.T) Options {
	t.Helper()
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	return o
}
