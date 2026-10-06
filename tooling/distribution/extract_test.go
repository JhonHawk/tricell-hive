package distribution

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractRejectsUnsafeArchiveWithoutEscape(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "outside")
	archive := testArchive(t, []tar.Header{{Name: "hive-1.2.3/../../outside", Typeflag: tar.TypeReg, Mode: 0600, Size: 1}}, []string{"x"})
	if _, err := Extract(archive, t.TempDir()); err == nil {
		t.Fatal("accepted traversal")
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Fatalf("archive escaped destination: %v", err)
	}
}

func TestExtractRejectsLinksAndDuplicates(t *testing.T) {
	for name, headers := range map[string][]tar.Header{
		"link": {{Name: "hive-1.2.3/link", Typeflag: tar.TypeSymlink, Linkname: "elsewhere"}},
		"duplicate": {
			{Name: "hive-1.2.3/file", Typeflag: tar.TypeReg, Mode: 0600, Size: 1},
			{Name: "hive-1.2.3/file", Typeflag: tar.TypeReg, Mode: 0600, Size: 1},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Extract(testArchive(t, headers, []string{"x", "y"}), t.TempDir()); err == nil {
				t.Fatal("accepted unsafe archive")
			}
		})
	}
}

func testArchive(t *testing.T, headers []tar.Header, data []string) []byte {
	t.Helper()
	var result bytes.Buffer
	gz := gzip.NewWriter(&result)
	tw := tar.NewWriter(gz)
	for i, h := range headers {
		if err := tw.WriteHeader(&h); err != nil {
			t.Fatal(err)
		}
		if h.Size > 0 {
			if _, err := tw.Write([]byte(data[i])); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return result.Bytes()
}
