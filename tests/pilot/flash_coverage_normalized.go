package main

import "strings"

// flashCanonicalReference is the canonical text the observer compares and
// hashes: the reference without trailing whitespace, so a read of the whole
// file, which ends at the last line, matches the file's final newline.
func flashCanonicalReference(text string) string {
	return strings.TrimRight(text, " \t\r\n\f\v")
}
