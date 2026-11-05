package main

import (
	"os"
	"path/filepath"
	"testing"
)

// vimUpd mengirim beberapa tombol berurutan ke model.
func vimUpd(m model, keys ...string) model {
	for _, k := range keys {
		r, _ := m.Update(keyStr(k))
		m = r.(model)
	}
	return m
}

// openTemp menulis isi ke file sementara lalu membukanya (mode insert).
func openTemp(t *testing.T, content string) (model, string) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "a.md")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	m := newSized(t)
	m.openFile(p)
	return m, p
}

func TestVimEscEntersNormalAndTypingIsBlocked(t *testing.T) {
	m, _ := openTemp(t, "hello\n")
	m = vimUpd(m, "esc")
	if m.mode != modeNormal {
		t.Fatalf("esc: mode %v, want normal", m.mode)
	}
	m = vimUpd(m, "x", "y")
	if got := m.editor.Value(); got != "hello\n" {
		t.Fatalf("normal mode mengubah buffer: %q", got)
	}
	if m.mode != modeNormal {
		t.Fatalf("mode berubah jadi %v", m.mode)
	}
}

func TestVimInsertSessionUndo(t *testing.T) {
	m, _ := openTemp(t, "abc\n")
	m = vimUpd(m, "X", "Y") // insert mode, kursor di awal baris
	if got := m.editor.Value(); got != "XYabc\n" {
		t.Fatalf("isi setelah mengetik: %q", got)
	}
	m = vimUpd(m, "esc")
	if got := m.editor.Value(); got != "XYabc\n" {
		t.Fatalf("esc mengubah isi: %q", got)
	}
	m = vimUpd(m, "u")
	if got := m.editor.Value(); got != "abc\n" {
		t.Fatalf("undo satu sesi insert: %q, want %q", got, "abc\n")
	}
	m = vimUpd(m, "u")
	if got := m.editor.Value(); got != "abc\n" {
		t.Fatalf("undo saat riwayat habis mengubah isi: %q", got)
	}
}

func TestVimDdDeletesLineAndUndoRestores(t *testing.T) {
	m, _ := openTemp(t, "one\ntwo\nthree\n")
	m = vimUpd(m, "esc", "j") // baris kedua
	m = vimUpd(m, "d", "d")
	if got := m.editor.Value(); got != "one\nthree\n" {
		t.Fatalf("dd: %q, want %q", got, "one\nthree\n")
	}
	if got := m.editor.Line(); got != 1 {
		t.Fatalf("kursor setelah dd di baris %d, want 1", got)
	}
	m = vimUpd(m, "u")
	if got := m.editor.Value(); got != "one\ntwo\nthree\n" {
		t.Fatalf("undo dd: %q", got)
	}
}

func TestVimDdOnOnlyLine(t *testing.T) {
	m, _ := openTemp(t, "solo\n")
	m = vimUpd(m, "esc", "d", "d")
	if got := m.editor.Value(); got != "" {
		t.Fatalf("dd satu-satunya baris: %q, want kosong", got)
	}
	m = vimUpd(m, "u")
	if got := m.editor.Value(); got != "solo\n" {
		t.Fatalf("undo: %q", got)
	}
}

func TestVimSingleDDoesNothing(t *testing.T) {
	m, _ := openTemp(t, "one\ntwo\n")
	m = vimUpd(m, "esc", "d", "j") // d lalu gerakan, bukan dd
	if got := m.editor.Value(); got != "one\ntwo\n" {
		t.Fatalf("d tunggal menghapus: %q", got)
	}
	if m.pendingD {
		t.Fatal("pendingD tidak direset")
	}
}

func TestVimWriteCommand(t *testing.T) {
	m, p := openTemp(t, "abc\n")
	m = vimUpd(m, "Z", "esc", ":", "w", "enter")
	if got, _ := os.ReadFile(p); string(got) != "Zabc\n" {
		t.Fatalf("file setelah :w: %q", got)
	}
	if m.mode != modeNormal {
		t.Fatalf("mode setelah :w = %v, want normal", m.mode)
	}
	if m.status != "Saved: a.md" {
		t.Fatalf("status: %q", m.status)
	}
}

func TestVimCommandCancel(t *testing.T) {
	m, p := openTemp(t, "abc\n")
	m = vimUpd(m, "Z", "esc", ":", "w", "esc")
	if got, _ := os.ReadFile(p); string(got) != "abc\n" {
		t.Fatalf("esc di command line tetap menyimpan: %q", got)
	}
	if m.mode != modeNormal || m.cmdline != "" {
		t.Fatalf("setelah esc: mode %v cmdline %q", m.mode, m.cmdline)
	}
}

func TestVimUnknownCommand(t *testing.T) {
	m, _ := openTemp(t, "abc\n")
	m = vimUpd(m, "esc", ":", "q", "enter")
	if m.status != "Unknown command: :q" {
		t.Fatalf("status: %q", m.status)
	}
}

func TestVimAppendEntersInsert(t *testing.T) {
	m, _ := openTemp(t, "abc\n")
	m = vimUpd(m, "esc", "a", "X")
	if m.mode != modeInsert {
		t.Fatalf("a: mode %v, want insert", m.mode)
	}
	if got := m.editor.Value(); got != "aXbc\n" {
		t.Fatalf("a menyisipkan setelah kursor: %q", got)
	}
}

func TestOpenFileResetsUndo(t *testing.T) {
	m, _ := openTemp(t, "abc\n")
	m = vimUpd(m, "X", "esc")
	if len(m.undoStack) == 0 {
		t.Fatal("sesi insert tidak masuk riwayat")
	}
	p2 := filepath.Join(t.TempDir(), "b.md")
	if err := os.WriteFile(p2, []byte("other\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m.openFile(p2)
	if len(m.undoStack) != 0 {
		t.Fatalf("undo stack tidak direset: %d", len(m.undoStack))
	}
	if m.mode != modeInsert {
		t.Fatalf("mode setelah buka file: %v", m.mode)
	}
}
