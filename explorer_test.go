package main

import (
	"os"
	"path/filepath"
	"testing"
)

func names(es []entry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.name
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Semua entry tampil: dotfile, dotfolder, file, folder, symlink folder, symlink rusak.
func TestReadDirListsEverything(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"b.md", "a.md", ".hidden", ".env"} {
		if err := os.WriteFile(filepath.Join(dir, f), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range []string{"zdir", ".git"} {
		if err := os.Mkdir(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(dir, "zdir"), filepath.Join(dir, "linkdir")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "nowhere"), filepath.Join(dir, "broken")); err != nil {
		t.Fatal(err)
	}

	es, err := readDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"..", ".git", "linkdir", "zdir", ".env", ".hidden", "a.md", "b.md", "broken"}
	if got := names(es); !equal(got, want) {
		t.Fatalf("entries:\n got  %v\n want %v", got, want)
	}
	isDir := map[string]bool{}
	for _, e := range es {
		isDir[e.name] = e.isDir
	}
	if !isDir["linkdir"] {
		t.Error("symlink ke folder harus dianggap folder")
	}
	link := map[string]bool{}
	for _, e := range es {
		link[e.name] = e.isLink
	}
	if !link["linkdir"] || !link["broken"] {
		t.Error("symlink harus ditandai isLink")
	}
	if link["zdir"] || link["a.md"] {
		t.Error("file/folder biasa tidak boleh ditandai isLink")
	}
	if isDir["broken"] {
		t.Error("symlink rusak harus tampil sebagai file")
	}
}

func TestReadDirRootHasNoParent(t *testing.T) {
	es, err := readDir("/")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range es {
		if e.name == ".." {
			t.Fatal("root tidak boleh punya entry ..")
		}
	}
}

func TestReadDirMissing(t *testing.T) {
	if _, err := readDir(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing dir")
	}
}
