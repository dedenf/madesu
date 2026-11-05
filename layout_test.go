package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Setiap frame harus tepat H baris, dan setiap baris tepat W kolom.
func TestViewGeometry(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {80, 24}, {60, 20}} {
		w, h := size[0], size[1]
		m, err := newModel("testdata", true)
		if err != nil {
			t.Fatal(err)
		}
		var tm tea.Model = m
		tm, _ = tm.Update(tea.WindowSizeMsg{Width: w, Height: h})
		for _, s := range []string{"tree", "file", "preview"} {
			mm := tm.(model)
			switch s {
			case "file":
				mm.openFile(mm.dir + "/sample.md")
			case "preview":
				mm.togglePreview()
			}
			tm = mm
			out := tm.View()
			lines := strings.Split(out, "\n")
			if len(lines) != h {
				t.Errorf("%dx%d %s: %d baris, want %d", w, h, s, len(lines), h)
			}
			for i, l := range lines {
				if lw := lipgloss.Width(l); lw != w {
					t.Errorf("%dx%d %s: baris %d lebar %d, want %d: %q", w, h, s, i, lw, w, l)
					break
				}
			}
		}
	}
}
