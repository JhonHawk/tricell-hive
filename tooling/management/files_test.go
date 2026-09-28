package management

// Characterization tests for the Hive block mechanics in files.go: blockRange,
// managedBlock, owned and transform. They pin current behavior for the Hive
// marker pair before those functions are parametrized by a marker pair (see
// design.md "Bloque de voz"), so the parametrization can be verified to leave
// Hive behavior unchanged.

import (
	"testing"
	"tricell-hive/integrations/target"
)

func blockRecord(path string, managed []byte) Record {
	return Record{Target: target.Target{Path: path, Kind: "block"}, Managed: managed}
}

// --- blockRange --------------------------------------------------------------

func TestBlockRangeNoMarkers(t *testing.T) {
	a, b, err := blockRange([]byte("plain text\n"), hiveMarkers)
	if err != nil || a != -1 || b != -1 {
		t.Fatalf("got %d %d %v", a, b, err)
	}
}

func TestBlockRangeCRLF(t *testing.T) {
	data := []byte("intro\r\n" + Begin + "\r\nline one\r\nline two\r\n" + End + "\r\ntrailer\r\n")
	a, b, err := blockRange(data, hiveMarkers)
	if err != nil {
		t.Fatal(err)
	}
	want := Begin + "\r\nline one\r\nline two\r\n" + End + "\r\n"
	if got := string(data[a:b]); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBlockRangeDuplicateMarkers(t *testing.T) {
	data := []byte(Begin + "\nbody\n" + End + "\n" + Begin + "\nbody\n" + End + "\n")
	if _, _, err := blockRange(data, hiveMarkers); err == nil {
		t.Fatal("expected an error for duplicate markers")
	}
}

func TestBlockRangeReversedMarkers(t *testing.T) {
	data := []byte(End + "\nbody\n" + Begin + "\n")
	if _, _, err := blockRange(data, hiveMarkers); err == nil {
		t.Fatal("expected an error for reversed markers")
	}
}

func TestBlockRangeMarkerNotFullLineLeading(t *testing.T) {
	data := []byte("prefix" + Begin + "\nbody\n" + End + "\n")
	if _, _, err := blockRange(data, hiveMarkers); err == nil {
		t.Fatal("expected an error: begin marker is not a full line")
	}
}

func TestBlockRangeMarkerNotFullLineTrailing(t *testing.T) {
	data := []byte(Begin + "\nbody\n" + End + "suffix\n")
	if _, _, err := blockRange(data, hiveMarkers); err == nil {
		t.Fatal("expected an error: end marker is not a full line")
	}
}

func TestBlockRangeAtEOFWithoutTrailingNewline(t *testing.T) {
	data := []byte(Begin + "\nbody\n" + End)
	a, b, err := blockRange(data, hiveMarkers)
	if err != nil {
		t.Fatal(err)
	}
	if string(data[a:b]) != string(data) {
		t.Fatalf("expected the whole buffer, got %q", data[a:b])
	}
}

// --- managedBlock --------------------------------------------------------------

func TestManagedBlockLF(t *testing.T) {
	got := managedBlock([]byte("hello\nworld\n\n"), []byte("current\nfile\n"), hiveMarkers)
	want := Begin + "\nhello\nworld\n" + End + "\n"
	if string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestManagedBlockCRLF(t *testing.T) {
	got := managedBlock([]byte("hello\nworld\n"), []byte("current\r\nfile\r\n"), hiveMarkers)
	want := Begin + "\r\nhello\r\nworld\r\n" + End + "\r\n"
	if string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestManagedBlockNoCurrentNewlineDefaultsToLF(t *testing.T) {
	got := managedBlock([]byte("hello\n"), nil, hiveMarkers)
	want := Begin + "\nhello\n" + End + "\n"
	if string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

// --- owned --------------------------------------------------------------

func TestOwnedHiveCreatedFileMatches(t *testing.T) {
	managed := managedBlock([]byte("body\n"), nil, hiveMarkers)
	r := blockRecord("/f", managed)
	r.CreatedFile = true
	s := snapshot{Exists: true, Data: managed}
	if err := owned(s, r, hiveMarkers); err != nil {
		t.Fatal(err)
	}
}

func TestOwnedMissingFile(t *testing.T) {
	r := blockRecord("/f", managedBlock([]byte("body\n"), nil, hiveMarkers))
	if err := owned(snapshot{}, r, hiveMarkers); err == nil {
		t.Fatal("expected an error for a missing managed file")
	}
}

func TestOwnedBlockMismatch(t *testing.T) {
	r := blockRecord("/f", managedBlock([]byte("body\n"), nil, hiveMarkers))
	tampered := managedBlock([]byte("tampered body\n"), nil, hiveMarkers)
	s := snapshot{Exists: true, Data: tampered}
	if err := owned(s, r, hiveMarkers); err == nil {
		t.Fatal("expected an error for a tampered block")
	}
}

func TestOwnedDuplicateMarkersIsAnError(t *testing.T) {
	managed := managedBlock([]byte("body\n"), nil, hiveMarkers)
	r := blockRecord("/f", managed)
	dup := append(append([]byte{}, managed...), managed...)
	if err := owned(snapshot{Exists: true, Data: dup}, r, hiveMarkers); err == nil {
		t.Fatal("expected an error for duplicate markers")
	}
}

// --- transform --------------------------------------------------------------

func TestTransformInstallAddsSeparatorAfterExistingContent(t *testing.T) {
	after := blockRecord("/f", managedBlock([]byte("body\n"), nil, hiveMarkers))
	after.Leading = "\n"
	s := snapshot{Exists: true, Data: []byte("existing content, no trailing newline")}
	out, err := transform(s, nil, &after, hiveMarkers)
	if err != nil {
		t.Fatal(err)
	}
	want := "existing content, no trailing newline\n" + string(after.Managed)
	if string(out.Data) != want {
		t.Fatalf("got %q want %q", out.Data, want)
	}
}

func TestTransformInstallIntoEmptyFileNoSeparator(t *testing.T) {
	after := blockRecord("/f", managedBlock([]byte("body\n"), nil, hiveMarkers))
	s := snapshot{}
	out, err := transform(s, nil, &after, hiveMarkers)
	if err != nil {
		t.Fatal(err)
	}
	if string(out.Data) != string(after.Managed) {
		t.Fatalf("got %q want %q", out.Data, after.Managed)
	}
}

func TestTransformInstallOntoExistingUnownedBlockConflicts(t *testing.T) {
	after := blockRecord("/f", managedBlock([]byte("body\n"), nil, hiveMarkers))
	s := snapshot{Exists: true, Data: managedBlock([]byte("someone else's block\n"), nil, hiveMarkers)}
	if _, err := transform(s, nil, &after, hiveMarkers); err == nil {
		t.Fatal("expected an unowned-block conflict")
	}
}

func TestTransformRemoveHiveCreatedFileLeavesNothing(t *testing.T) {
	managed := managedBlock([]byte("body\n"), nil, hiveMarkers)
	before := blockRecord("/f", managed)
	before.CreatedFile = true
	s := snapshot{Exists: true, Data: managed}
	out, err := transform(s, &before, nil, hiveMarkers)
	if err != nil {
		t.Fatal(err)
	}
	if out.Exists {
		t.Fatalf("expected the created file to be removed, got %q", out.Data)
	}
}

func TestTransformRemoveStripsRecordedSeparator(t *testing.T) {
	managed := managedBlock([]byte("body\n"), nil, hiveMarkers)
	before := blockRecord("/f", managed)
	before.Leading = "\n"
	s := snapshot{Exists: true, Data: append([]byte("kept content\n\n"), managed...)}
	out, err := transform(s, &before, nil, hiveMarkers)
	if err != nil {
		t.Fatal(err)
	}
	if string(out.Data) != "kept content\n" {
		t.Fatalf("got %q", out.Data)
	}
}

func TestTransformReplaceKeepsSurroundingBytes(t *testing.T) {
	before := blockRecord("/f", managedBlock([]byte("old body\n"), nil, hiveMarkers))
	after := blockRecord("/f", managedBlock([]byte("new body\n"), nil, hiveMarkers))
	s := snapshot{Exists: true, Data: append(append([]byte("before-text\n"), before.Managed...), []byte("after-text\n")...)}
	out, err := transform(s, &before, &after, hiveMarkers)
	if err != nil {
		t.Fatal(err)
	}
	want := "before-text\n" + string(after.Managed) + "after-text\n"
	if string(out.Data) != want {
		t.Fatalf("got %q want %q", out.Data, want)
	}
}

// --- marker error texts name which block failed --------------------------------

func TestBlockRangeErrorTextsNameTheBlock(t *testing.T) {
	dup := []byte(Begin + "\nbody\n" + End + "\n" + Begin + "\nbody\n" + End + "\n")
	_, _, err := blockRange(dup, hiveMarkers)
	if err == nil || err.Error() != "malformed or duplicate Hive markers" {
		t.Fatalf("got %v", err)
	}
	dupVoice := []byte(VoiceBegin + "\nbody\n" + VoiceEnd + "\n" + VoiceBegin + "\nbody\n" + VoiceEnd + "\n")
	_, _, err = blockRange(dupVoice, voiceMarkers)
	if err == nil || err.Error() != "malformed or duplicate voice markers" {
		t.Fatalf("got %v", err)
	}
	reversed := []byte(End + "\nbody\n" + Begin + "\n")
	_, _, err = blockRange(reversed, hiveMarkers)
	if err == nil || err.Error() != "reversed Hive markers" {
		t.Fatalf("got %v", err)
	}
	reversedVoice := []byte(VoiceEnd + "\nbody\n" + VoiceBegin + "\n")
	_, _, err = blockRange(reversedVoice, voiceMarkers)
	if err == nil || err.Error() != "reversed voice markers" {
		t.Fatalf("got %v", err)
	}
}

func TestTransformUnownedBlockErrorNamesTheBlock(t *testing.T) {
	after := blockRecord("/f", managedBlock([]byte("body\n"), nil, hiveMarkers))
	s := snapshot{Exists: true, Data: managedBlock([]byte("someone else's block\n"), nil, hiveMarkers)}
	if _, err := transform(s, nil, &after, hiveMarkers); err == nil || err.Error() != "unowned Hive block" {
		t.Fatalf("got %v", err)
	}
	afterVoice := blockRecord("/f", managedBlock([]byte("body\n"), nil, voiceMarkers))
	sv := snapshot{Exists: true, Data: managedBlock([]byte("someone else's block\n"), nil, voiceMarkers)}
	if _, err := transform(sv, nil, &afterVoice, voiceMarkers); err == nil || err.Error() != "unowned voice block" {
		t.Fatalf("got %v", err)
	}
}

func TestTransformOnTamperedBeforeConflicts(t *testing.T) {
	before := blockRecord("/f", managedBlock([]byte("old body\n"), nil, hiveMarkers))
	after := blockRecord("/f", managedBlock([]byte("new body\n"), nil, hiveMarkers))
	s := snapshot{Exists: true, Data: []byte("no block here at all")}
	if _, err := transform(s, &before, &after, hiveMarkers); err == nil {
		t.Fatal("expected a conflict when Before does not match the current file")
	}
}
