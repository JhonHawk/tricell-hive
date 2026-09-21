package main

import "testing"

func TestNativeTrustExceptionIsNarrow(t *testing.T) {
	before := []byte("model = \"gpt-6-astra\"\n")
	table := []byte("\n[projects.\"/temporary/repo\"]\ntrust_level = \"trusted\"\n")
	if !onlyTrustAdded(before, append(append([]byte{}, before...), table...), "/temporary/repo") {
		t.Fatal("expected exact insertion")
	}
	bad := append([]byte("model = \"other\"\n"), table...)
	if onlyTrustAdded(before, bad, "/temporary/repo") {
		t.Fatal("accepted other setting change")
	}
	if onlyTrustAdded(before, append(append([]byte{}, before...), table...), "/different/repo") {
		t.Fatal("accepted different target")
	}
	if onlyTrustAdded(before, append(append(append([]byte{}, before...), table...), table...), "/temporary/repo") {
		t.Fatal("accepted duplicates")
	}
}
func TestTrustInsertedBetweenExistingTables(t *testing.T) {
	before := []byte("[projects.\"/old\"]\ntrust_level = \"trusted\"\n\n[notice]\nflag = true\n")
	after := []byte("[projects.\"/old\"]\ntrust_level = \"trusted\"\n\n[projects.\"/temporary/repo\"]\ntrust_level = \"trusted\"\n\n[notice]\nflag = true\n")
	if !onlyTrustAdded(before, after, "/temporary/repo") {
		t.Fatal("table separator falsely classified as unrelated change")
	}
}
