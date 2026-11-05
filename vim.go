package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// editMode is the vim mode while the editor is focused.
type editMode int

const (
	modeInsert  editMode = iota // typing goes into the buffer
	modeNormal                  // keys are commands
	modeCommand                 // after ":", until enter or esc
)

const maxUndo = 200 // undo history depth per file

func (m editMode) label() string {
	switch m {
	case modeNormal:
		return "NORMAL"
	case modeCommand:
		return "COMMAND"
	}
	return "INSERT"
}

// snapshot is a point in undo history: buffer text plus cursor position.
type snapshot struct {
	value    string
	row, col int
}

// takeSnapshot records the buffer and cursor. LineInfo gives the rune column
// inside the logical line, which is what SetCursor expects.
func (m model) takeSnapshot() snapshot {
	li := m.editor.LineInfo()
	return snapshot{value: m.editor.Value(), row: m.editor.Line(), col: li.StartColumn + li.ColumnOffset}
}

// restore puts the editor back at a snapshot. SetValue leaves the cursor at the
// end, so walk up to the saved line; that always terminates since Line() only
// decreases.
func (m *model) restore(s snapshot) {
	m.editor.SetValue(s.value)
	row := clamp(s.row, 0, m.editor.LineCount()-1)
	for m.editor.Line() > row {
		m.editor.CursorUp()
	}
	m.editor.SetCursor(s.col)
}

// beginInsert starts an insert session and remembers where it started. One
// session maps to one undo step, like vim's insert-mode undo.
func (m *model) beginInsert() {
	m.mode = modeInsert
	m.pendingD = false
	m.status = ""
	if m.undoStart == nil {
		s := m.takeSnapshot()
		m.undoStart = &s
	}
}

// endInsert closes the session, pushing one undo step if anything changed.
func (m *model) endInsert() {
	if m.undoStart != nil && m.undoStart.value != m.editor.Value() {
		m.pushUndo(*m.undoStart)
	}
	m.undoStart = nil
}

func (m *model) pushUndo(s snapshot) {
	m.undoStack = append(m.undoStack, s)
	if len(m.undoStack) > maxUndo {
		m.undoStack = m.undoStack[len(m.undoStack)-maxUndo:]
	}
}

func (m *model) undo() {
	if len(m.undoStack) == 0 {
		m.status = "Already at oldest change."
		return
	}
	s := m.undoStack[len(m.undoStack)-1]
	m.undoStack = m.undoStack[:len(m.undoStack)-1]
	m.restore(s)
	m.status = "Undo"
	if m.showPrev {
		m.refreshPreview()
	}
}

// deleteLine removes the logical line under the cursor (vim's dd).
func (m *model) deleteLine() {
	m.pushUndo(m.takeSnapshot())
	lines := strings.Split(m.editor.Value(), "\n")
	row := clamp(m.editor.Line(), 0, len(lines)-1)
	lines = append(lines[:row], lines[row+1:]...)
	m.editor.SetValue(strings.Join(lines, "\n"))
	target := clamp(row, 0, m.editor.LineCount()-1)
	for m.editor.Line() > target {
		m.editor.CursorUp()
	}
	m.editor.CursorStart()
	m.status = "Deleted line"
}

// characterRight moves one rune right, without crossing the end of the line.
func (m *model) characterRight() {
	li := m.editor.LineInfo()
	m.editor.SetCursor(li.StartColumn + li.ColumnOffset + 1)
}

// updateInsert forwards the key to the textarea. The undo session was opened
// when insert mode began.
func (m *model) updateInsert(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return cmd
}

// updateNormal handles keys in normal mode. Anything not a command or a
// navigation key is swallowed so it cannot edit the buffer.
func (m *model) updateNormal(msg tea.KeyMsg) tea.Cmd {
	kd := m.pendingD
	m.pendingD = false
	if kd && msg.String() != "d" {
		m.status = "" // batalkan tampilan "d…" kalau bukan dd
	}
	switch msg.String() {
	case "i":
		m.beginInsert()
		return nil
	case "a":
		m.characterRight()
		m.beginInsert()
		return nil
	case "u":
		m.undo()
		return nil
	case "d":
		if kd {
			m.deleteLine()
		} else {
			m.pendingD = true
			m.status = "d…"
		}
		return nil
	case ":":
		m.mode = modeCommand
		m.cmdline = ""
		m.status = ""
		return nil
	case "up", "k":
		return m.nav(tea.KeyUp)
	case "down", "j":
		return m.nav(tea.KeyDown)
	case "left", "h":
		return m.nav(tea.KeyLeft)
	case "right", "l":
		return m.nav(tea.KeyRight)
	case "home":
		return m.nav(tea.KeyHome)
	case "end":
		return m.nav(tea.KeyEnd)
	case "pgup":
		moveCursor(&m.editor, -1, m.editor.Height())
	case "pgdown":
		moveCursor(&m.editor, 1, m.editor.Height())
	case "ctrl+home":
		moveToDocEdge(&m.editor, -1)
	case "ctrl+end":
		moveToDocEdge(&m.editor, 1)
	}
	return nil
}

// nav sends a synthetic navigation key to the textarea.
func (m *model) nav(t tea.KeyType) tea.Cmd {
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(tea.KeyMsg{Type: t})
	return cmd
}

// updateCmdline edits the ":" command line and runs it on enter.
func (m *model) updateCmdline(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeNormal
		m.cmdline = ""
		m.status = ""
	case tea.KeyEnter:
		cmd := strings.TrimSpace(m.cmdline)
		m.cmdline = ""
		m.mode = modeNormal
		m.runCommand(cmd)
	case tea.KeyBackspace:
		if r := []rune(m.cmdline); len(r) > 0 {
			m.cmdline = string(r[:len(r)-1])
		}
	case tea.KeySpace:
		m.cmdline += " "
	case tea.KeyRunes:
		m.cmdline += string(msg.Runes)
	}
}

func (m *model) runCommand(cmd string) {
	switch cmd {
	case "":
		return
	case "w":
		m.save()
	default:
		m.status = "Unknown command: :" + cmd
	}
}
