package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDirOrderAndHidden(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"b.md", "a.md", ".hidden"} {
		if err := os.WriteFile(filepath.Join(dir, f), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "zdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	es, err := readDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"..", "zdir", "a.md", "b.md"}
	if len(es) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(es), len(want), es)
	}
	for i, w := range want {
		if es[i].name != w {
			t.Errorf("entry %d = %q, want %q", i, es[i].name, w)
		}
	}
}

func TestReadDirMissing(t *testing.T) {
	if _, err := readDir(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want string
	}{
		{"abc", 5, "abc"},
		{"abcdef", 4, "abc…"},
		{"abc", 0, ""},
	}
	for _, c := range cases {
		if got := truncate(c.in, c.n); got != c.want {
			t.Errorf("truncate(%q,%d) = %q, want %q", c.in, c.n, got, c.want)
		}
	}
}
