package atomicfile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/my-rv/hensu/internal/atomicfile"
)

func TestWrite_replacesAtomically(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.env")
	if err := os.WriteFile(path, []byte("OLD=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := atomicfile.Write(path, []byte("NEW=2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "NEW=2\n" {
		t.Fatalf("got %q", data)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", info.Mode().Perm())
	}
}

func TestWrite_createsParentDirs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", ".env")
	if err := atomicfile.Write(path, []byte("X=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "X=1\n" {
		t.Fatalf("got %q", data)
	}
}
