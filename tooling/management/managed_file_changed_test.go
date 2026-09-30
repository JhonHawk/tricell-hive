package management

import (
	"errors"
	"strings"
	"testing"
	"tricell-hive/integrations/target"
)

// Every failure owned() reports about a managed file must be a
// *ManagedFileChangedError whose text carries the path exactly once, in words
// that say what the user can do.

func skillRecord(path, kind string, managed []byte, mode uint32) Record {
	return Record{Target: target.Target{Path: path, Kind: kind}, Managed: managed, Mode: mode}
}

func TestOwnedReturnsTypedErrorWithPlainWords(t *testing.T) {
	const path = "/synthetic/home/.agents/skills/demo/SKILL.md"
	const block = "/synthetic/home/CLAUDE.md"
	const changedFile = " differs from what Hive expects there; undo the change, restore it from a backup, or move your own file elsewhere"
	symlink := Record{Target: target.Target{Path: path, Kind: "symlink", LinkTarget: "/elsewhere"}}
	cases := []struct {
		name string
		snap snapshot
		rec  Record
		want string
		path string
	}{
		{"skill bytes changed", snapshot{Exists: true, Data: []byte("edited"), Mode: 0600}, skillRecord(path, "skill", []byte("original"), 0600), path + changedFile, path},
		{"agent bytes changed", snapshot{Exists: true, Data: []byte("edited"), Mode: 0600}, skillRecord(path, "agent", []byte("original"), 0600), path + changedFile, path},
		{"skill mode changed", snapshot{Exists: true, Data: []byte("same"), Mode: 0777}, skillRecord(path, "skill", []byte("same"), 0600), path + " differs from what Hive expects there: its permissions changed; restore them or restore the file from a backup", path},
		{"symlink retargeted", snapshot{Exists: true, Kind: "symlink", LinkTarget: "/other"}, symlink, path + changedFile, path},
		{"symlink replaced by a file", snapshot{Exists: true, Data: []byte("x")}, symlink, path + changedFile, path},
		{"resource type changed", snapshot{Exists: true, Kind: "symlink", LinkTarget: "/x"}, skillRecord(path, "skill", []byte("a"), 0600), path + changedFile, path},
		{"block differs", snapshot{Exists: true, Data: []byte(Begin + "\nedited\n" + End + "\n")}, blockRecord(block, []byte(Begin+"\noriginal\n"+End+"\n")), "the Hive block in " + block + " differs from what Hive wrote; undo the change inside the block", block},
		{"block missing", snapshot{Exists: true, Data: []byte("user text only\n")}, blockRecord(block, []byte(Begin+"\noriginal\n"+End+"\n")), "the Hive block in " + block + " is gone; restore the file from a backup", block},
		{"block file missing", snapshot{}, blockRecord(block, []byte("x")), block + ", which held the Hive block, is no longer there; restore it from a backup", block},
		{"file missing", snapshot{}, skillRecord(path, "skill", []byte("a"), 0600), "Hive installed " + path + " and it is no longer there; restore it from a backup", path},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := owned(tc.snap, tc.rec, hiveMarkers)
			var changed *ManagedFileChangedError
			if !errors.As(err, &changed) {
				t.Fatalf("owned returned %T (%v), want *ManagedFileChangedError", err, err)
			}
			if changed.Path != tc.path {
				t.Errorf("Path = %q, want %q", changed.Path, tc.path)
			}
			if err.Error() != tc.want {
				t.Errorf("text = %q, want %q", err.Error(), tc.want)
			}
			if n := strings.Count(err.Error(), tc.path); n != 1 {
				t.Errorf("the path appears %d times in %q", n, err.Error())
			}
		})
	}
}

// Errors that are not about a changed file still name it, because callers
// return them as they are.
func TestMarkerAndCollisionErrorsNameThePath(t *testing.T) {
	const block = "/synthetic/home/CLAUDE.md"
	bad := snapshot{Exists: true, Data: []byte(Begin + "\nno end marker\n")}
	after := blockRecord(block, []byte("x"))
	_, transformErr := transform(bad, nil, &after, hiveMarkers)
	skill := skillRecord(block, "skill", []byte("a"), 0600)
	_, collisionErr := transform(snapshot{Exists: true, Data: []byte("mine")}, nil, &skill, hiveMarkers)
	_, migrationErr := transformResource(snapshot{}, Change{Target: target.Target{Path: block}, Replaces: &skill}, "install")
	for name, err := range map[string]error{
		"owned":     owned(bad, blockRecord(block, []byte("x")), hiveMarkers),
		"transform": transformErr, "collision": collisionErr, "migration": migrationErr,
	} {
		if err == nil || strings.Count(err.Error(), block) != 1 {
			t.Errorf("%s: want one mention of the path, got %v", name, err)
		}
	}
}

func TestOwnedBlockMarkerErrorStaysUntyped(t *testing.T) {
	snap := snapshot{Exists: true, Data: []byte(Begin + "\nno end marker\n")}
	err := owned(snap, blockRecord("/synthetic/CLAUDE.md", []byte("x")), hiveMarkers)
	var changed *ManagedFileChangedError
	if err == nil || errors.As(err, &changed) {
		t.Fatalf("a marker error must stay a plain error, got %T %v", err, err)
	}
}

func TestVoiceConflictDoesNotRepeatThePath(t *testing.T) {
	const path = "/synthetic/home/.codex/AGENTS.md"
	span := VoiceSpan{Managed: []byte(VoiceBegin + "\noriginal\n" + VoiceEnd + "\n")}
	edited := snapshot{Exists: true, Data: []byte(VoiceBegin + "\nedited\n" + VoiceEnd + "\n")}
	_, err := checkVoiceConflict(path, edited, true, span)
	var changed *ManagedFileChangedError
	if !errors.As(err, &changed) {
		t.Fatalf("errors.As failed on %T (%v)", err, err)
	}
	if n := strings.Count(err.Error(), path); n != 1 {
		t.Errorf("the path appears %d times in %q", n, err.Error())
	}
	if want := "the voice block in " + path + " differs from what Hive wrote; undo the change inside the block"; err.Error() != want {
		t.Errorf("voice conflict = %q, want %q", err.Error(), want)
	}
}
