package distribution

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteDownloadIndex(t *testing.T) {
	path := filepath.Join(t.TempDir(), "versions", "1.2.3", "index.json")
	want := DownloadIndex{Version: 1, ProductVersion: "1.2.3", Releases: []DownloadRelease{{Platform: "linux/arm64", SourceID: Digest([]byte("source")), Package: "versions/1.2.3/linux-arm64/hive-1.2.3-linux-arm64.tar.gz", PackageSHA256: Digest([]byte("package")), RawBinary: "versions/1.2.3/linux-arm64/hive", RawBinarySHA256: Digest([]byte("binary"))}}}
	if err := WriteDownloadIndex(path, want); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseDownloadIndex(data)
	if err != nil || got.ProductVersion != want.ProductVersion || len(got.Releases) != 1 {
		t.Fatalf("ParseDownloadIndex() = %+v, %v", got, err)
	}
}
