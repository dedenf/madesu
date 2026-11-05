// madesu: markdown simple utility. File explorer (kiri) + editor dengan preview markdown (kanan).
// Usage: madesu [dir]
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultTreeWidth = 30 // lebar awal panel kiri, termasuk border
	minTreeWidth     = 16 // batas bawah saat garis pemisah digeser
	minRightWidth    = 30 // batas bawah lebar panel kanan
	gutterWidth      = 1  // jarak kosong antara dua panel
	panelPadX        = 1  // padding kiri-kanan di dalam border; isi panel = lebar - 2*(border+padding)
)

// innerWidth mengembalikan lebar isi yang bisa dipakai di dalam panel selebar w.
func innerWidth(w int) int { return w - 2 - 2*panelPadX }

type focus int

const (
	focusTree  focus = iota // panel kiri: explorer
	focusRight              // panel kanan: editor, atau preview kalau showPrev
)

// theme memusatkan semua gaya tampilan. Hanya satu warna aksen; sisanya netral.
type theme struct {
	panel, panelOn lipgloss.Style // border panel tidak aktif / aktif
	title, titleOn lipgloss.Style
	muted          lipgloss.Style
	dir            lipgloss.Style
	sel, selIdle   lipgloss.Style // baris terpilih saat fokus / tidak fokus
	badge          lipgloss.Style // label mode di status bar
	key            lipgloss.Style // nama tombol di hint
	errMsg         lipgloss.Style
}

func newTheme(dark bool) theme {
	pick := func(light, darkC string) lipgloss.Color {
		if dark {
			return lipgloss.Color(darkC)
		}
		return lipgloss.Color(light)
	}
	accent := pick("25", "117")
	muted := pick("243", "245")
	onAccent := pick("255", "234")
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, panelPadX)
	return theme{
		panel:   box.BorderForeground(pick("250", "238")),
		panelOn: box.BorderForeground(accent),
		title:   lipgloss.NewStyle().Bold(true),
		titleOn: lipgloss.NewStyle().Bold(true).Foreground(accent),
		muted:   lipgloss.NewStyle().Foreground(muted),
		dir:     lipgloss.NewStyle().Foreground(accent),
		sel:     lipgloss.NewStyle().Background(accent).Foreground(onAccent).Bold(true),
		selIdle: lipgloss.NewStyle().Underline(true).Foreground(muted),
		badge:   lipgloss.NewStyle().Background(accent).Foreground(onAccent).Bold(true),
		key:     lipgloss.NewStyle().Bold(true),
		errMsg:  lipgloss.NewStyle().Foreground(pick("160", "210")),
	}
}

type model struct {
	dir     string
	entries []entry
	cursor  int

	openPath string // file yang sedang dibuka; "" kalau belum ada
	editor   textarea.Model
	preview  viewport.Model
	dark     bool // tema gelap/terang untuk renderer markdown dan tema UI
	th       theme
	showPrev bool // panel kanan menampilkan preview (read-only, bisa di-scroll)
	focus    focus

	treeW    int  // lebar panel kiri saat ini, bisa diubah dengan drag mouse
	dragging bool // true selama tombol kiri ditahan di garis pemisah

	status        string
	width, height int
}

func newModel(dir string, dark bool) (model, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return model{}, err
	}
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.Blur()
	m := model{
		dir:    abs,
		editor: ta,
		dark:   dark,
		th:     newTheme(dark),
		focus:  focusTree,
		treeW:  defaultTreeWidth,
	}
	if err := m.loadDir(abs); err != nil {
		return model{}, err
	}
	return m, nil
}

func (m *model) loadDir(dir string) error {
	es, err := readDir(dir)
	if err != nil {
		return err
	}
	m.dir = dir
	m.entries = es
	m.cursor = 0
	return nil
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
		return m, nil

	case tea.MouseMsg:
		ev := tea.MouseEvent(msg)
		if isWheel(ev) && ev.X > m.treeW {
			m.wheelRight(msg)
			return m, nil
		}
		m.handleMouse(ev)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.toggleFocus()
			return m, nil
		case "ctrl+s":
			m.save()
			return m, nil
		case "ctrl+p":
			m.togglePreview()
			return m, nil
		}
		if m.focus == focusTree {
			m.updateTree(msg.String())
			return m, nil
		}
		if m.showPrev {
			// Preview read-only: up/down/j/k/pgup/pgdn ditangani viewport.
			// home/end ke awal dan akhir dokumen.
			switch msg.String() {
			case "home":
				m.preview.GotoTop()
				return m, nil
			case "end":
				m.preview.GotoBottom()
				return m, nil
			}
			var cmd tea.Cmd
			m.preview, cmd = m.preview.Update(msg)
			return m, cmd
		}
		// Editor: home/end tetap per baris (bawaan textarea); ctrl+home/ctrl+end ke dokumen.
		switch msg.String() {
		case "pgup":
			moveCursor(&m.editor, -1, m.editor.Height())
			return m, nil
		case "pgdown":
			moveCursor(&m.editor, 1, m.editor.Height())
			return m, nil
		case "ctrl+home":
			moveToDocEdge(&m.editor, -1)
			return m, nil
		case "ctrl+end":
			moveToDocEdge(&m.editor, 1)
			return m, nil
		}
		var cmd tea.Cmd
		m.editor, cmd = m.editor.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) toggleFocus() {
	if m.focus == focusTree {
		if m.openPath == "" {
			m.status = "No file open. Pick a file with enter."
			return
		}
		m.focus = focusRight
		if !m.showPrev {
			m.editor.Focus()
		}
		return
	}
	m.focus = focusTree
	m.editor.Blur()
}

func (m *model) togglePreview() {
	if m.openPath == "" {
		m.status = "Open a file first to preview it."
		return
	}
	m.showPrev = !m.showPrev
	m.focus = focusRight
	if m.showPrev {
		m.editor.Blur()
		m.refreshPreview()
		m.status = ""
	} else {
		m.editor.Focus()
		m.status = ""
	}
}

// moveCursor memindahkan kursor editor n baris visual ke atas (dir < 0) atau ke bawah (dir > 0).
func moveCursor(ta *textarea.Model, dir, n int) {
	for i := 0; i < n; i++ {
		if dir < 0 {
			ta.CursorUp()
		} else {
			ta.CursorDown()
		}
	}
}

// moveToDocEdge memindahkan kursor ke baris pertama (dir < 0) atau terakhir (dir > 0)
// dokumen. Berhenti saat posisi tidak berubah lagi, jadi aman untuk baris yang ter-wrap.
func moveToDocEdge(ta *textarea.Model, dir int) {
	for {
		line, row := ta.Line(), ta.LineInfo().RowOffset
		if dir < 0 {
			ta.CursorUp()
		} else {
			ta.CursorDown()
		}
		if ta.Line() == line && ta.LineInfo().RowOffset == row {
			break
		}
	}
	if dir < 0 {
		ta.CursorStart()
	} else {
		ta.CursorEnd()
	}
}

func (m *model) updateTree(key string) {
	switch key {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.entries) == 0 {
			return
		}
		e := m.entries[m.cursor]
		target := filepath.Join(m.dir, e.name)
		if e.isDir {
			if err := m.loadDir(filepath.Clean(target)); err != nil {
				m.status = "error: " + err.Error()
			}
			return
		}
		m.openFile(target)
	}
}

func (m *model) openFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		m.status = "error reading file: " + err.Error()
		return
	}
	m.openPath = path
	m.editor.SetValue(string(data))
	// SetValue menaruh kursor di akhir teks; file dibuka dari baris pertama.
	moveToDocEdge(&m.editor, -1)
	m.focus = focusRight
	if m.showPrev {
		m.refreshPreview()
	} else {
		m.editor.Focus()
	}
	m.status = "Opened: " + filepath.Base(path)
}

func (m *model) save() {
	if m.openPath == "" {
		m.status = "No file is open."
		return
	}
	if err := os.WriteFile(m.openPath, []byte(m.editor.Value()), 0o644); err != nil {
		m.status = "error saving: " + err.Error()
		return
	}
	m.status = "Saved: " + filepath.Base(m.openPath)
	if m.showPrev {
		m.refreshPreview()
	}
}

// handleMouse menangani drag garis pemisah antara panel kiri dan kanan.
// Tekan di kolom border kiri atau gutter (±1 kolom) untuk mulai, geser untuk mengubah lebar,
// lepas untuk selesai. Preview di-render ulang saat lepas, bukan selama drag.
func (m *model) handleMouse(ev tea.MouseEvent) {
	switch {
	case ev.Action == tea.MouseActionPress && ev.Button == tea.MouseButtonLeft:
		if d := ev.X - m.treeW; d >= -1 && d <= 1 {
			m.dragging = true
		}
	case ev.Action == tea.MouseActionMotion && m.dragging:
		m.setTreeWidth(ev.X)
	case ev.Action == tea.MouseActionRelease && m.dragging:
		m.dragging = false
		m.refreshPreview()
	}
}

const wheelLines = 3 // baris per notch scroll wheel / trackpad

func isWheel(ev tea.MouseEvent) bool {
	return ev.Action == tea.MouseActionPress &&
		(ev.Button == tea.MouseButtonWheelUp || ev.Button == tea.MouseButtonWheelDown)
}

// wheelRight menggulir panel kanan. Preview digulir langsung lewat viewport.
// Editor: textarea menggulir viewport internalnya lewat event mouse, tapi lalu
// mengembalikan view ke kursor. Karena itu kursor digeser dengan jumlah baris yang sama
// dulu, sehingga kursor tetap di dalam view yang sudah digulir.
func (m *model) wheelRight(msg tea.MouseMsg) {
	dir := 1
	if msg.Button == tea.MouseButtonWheelUp {
		dir = -1
	}
	if m.showPrev {
		if dir < 0 {
			m.preview.ScrollUp(wheelLines)
		} else {
			m.preview.ScrollDown(wheelLines)
		}
		return
	}
	moveCursor(&m.editor, dir, wheelLines)
	m.editor, _ = m.editor.Update(msg)
}

// setTreeWidth mengubah lebar panel kiri dengan batas agar kedua panel tetap terbaca.
func (m *model) setTreeWidth(w int) {
	m.treeW = clamp(w, minTreeWidth, m.width-gutterWidth-minRightWidth)
	m.resizePanes()
}

func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

// layout dipanggil saat terminal di-resize. Geometri: body = height-1 baris (sisanya status bar).
// Setiap panel punya border 1 karakter di tiap sisi dan padding horizontal, lalu 1 baris judul.
func (m *model) layout() {
	m.treeW = clamp(m.treeW, minTreeWidth, m.width-gutterWidth-minRightWidth)
	m.resizePanes()
	m.refreshPreview()
}

// resizePanes mengatur ukuran editor dan viewport preview sesuai lebar panel kiri saat ini.
// Tidak me-render ulang markdown; itu tugas refreshPreview.
func (m *model) resizePanes() {
	paneW := innerWidth(m.width - m.treeW - gutterWidth) // lebar isi panel kanan
	paneH := m.height - 4                                // tinggi body dikurangi border (2), judul (1), status (1)
	if paneW < 10 || paneH < 3 {
		return
	}
	m.editor.SetWidth(paneW)
	m.editor.SetHeight(paneH)
	m.preview.Width = paneW
	m.preview.Height = paneH
}

// refreshPreview render markdown dari isi editor. Dipanggil saat preview dibuka,
// file dibuka/disimpan, atau layout berubah; tidak dipanggil per tombol.
func (m *model) refreshPreview() {
	if !m.showPrev || m.openPath == "" || m.preview.Width < 10 {
		return
	}
	out, err := renderMarkdown([]byte(m.editor.Value()), m.preview.Width, m.dark)
	if err != nil {
		m.status = "error render: " + err.Error()
		return
	}
	m.preview.SetContent(out)
}

// panel membungkus isi dengan border rounded. w dan h adalah ukuran total termasuk border.
func (m model) panel(content string, w, h int, on bool) string {
	st := m.th.panel
	if on {
		st = m.th.panelOn
	}
	return st.Width(w - 2).Height(h - 2).MaxWidth(w).MaxHeight(h).Render(content)
}

func (m model) treeView(innerW, innerH int) string {
	var b strings.Builder
	titleStyle, selStyle := m.th.title, m.th.selIdle
	if m.focus == focusTree {
		titleStyle, selStyle = m.th.titleOn, m.th.sel
	}
	b.WriteString(titleStyle.Render(" Files") + "\n")
	b.WriteString(m.th.muted.Render(" "+truncate(m.dir, innerW-1)) + "\n")

	visible := innerH - 2
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	for i := start; i < len(m.entries) && i < start+visible; i++ {
		e := m.entries[i]
		name := e.name
		if e.isDir && name != ".." {
			name += "/"
		}
		name = truncate(" "+name, innerW)
		var line string
		switch {
		case i == m.cursor:
			line = selStyle.Width(innerW).Render(name)
		case e.isDir:
			line = m.th.dir.Render(name)
		default:
			line = name
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (m model) rightView(innerW, innerH int) string {
	titleName := "Editor"
	if m.showPrev {
		titleName = "Preview"
	}
	titleStyle := m.th.title
	if m.focus == focusRight {
		titleStyle = m.th.titleOn
	}
	head := titleStyle.Render(" " + titleName)
	if m.openPath != "" {
		head += m.th.muted.Render("  " + truncate(filepath.Base(m.openPath), innerW-len(titleName)-4))
	}

	var body string
	switch {
	case m.openPath == "":
		empty := m.th.title.Render("No file open") + "\n" +
			m.th.muted.Render("Pick a file on the left, then press enter.")
		body = lipgloss.Place(innerW, innerH-1, lipgloss.Center, lipgloss.Center, empty)
	case m.showPrev:
		body = m.preview.View()
	default:
		body = m.editor.View()
	}
	return head + "\n" + body
}

// hints mengembalikan daftar tombol untuk fokus yang sedang aktif.
func (m model) hints() string {
	pairs := [][2]string{{"tab", "switch pane"}}
	switch {
	case m.focus == focusTree:
		pairs = append([][2]string{{"enter", "open"}, {"↑↓", "select"}}, pairs...)
		pairs = append(pairs, [2]string{"ctrl+c", "quit"})
	case m.showPrev:
		pairs = append([][2]string{{"↑↓", "scroll"}, {"ctrl+p", "editor"}}, pairs...)
	default:
		pairs = append([][2]string{{"ctrl+s", "save"}, {"ctrl+p", "preview"}}, pairs...)
	}
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, m.th.key.Render(p[0])+" "+m.th.muted.Render(p[1]))
	}
	return strings.Join(parts, "   ")
}

// statusBar: badge mode + pesan di kiri, hint tombol di kanan. Selalu tepat satu baris.
func (m model) statusBar() string {
	label := "Files"
	switch {
	case m.focus == focusRight && m.showPrev:
		label = "Preview"
	case m.focus == focusRight:
		label = "Editor"
	}
	badge := m.th.badge.Render(" " + label + " ")

	hints := m.hints()
	hintW := lipgloss.Width(hints)
	msgRoom := m.width - lipgloss.Width(badge) - 2 - hintW - 2
	if msgRoom < 10 {
		hints, hintW = "", 0
		msgRoom = m.width - lipgloss.Width(badge) - 2
	}
	msgStyle := m.th.muted
	if strings.HasPrefix(m.status, "error") {
		msgStyle = m.th.errMsg
	}
	left := badge + "  " + msgStyle.Render(truncate(m.status, msgRoom))
	gap := m.width - lipgloss.Width(left) - hintW
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + hints
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}
	bodyH := m.height - 1
	innerH := bodyH - 2

	left := m.panel(m.treeView(innerWidth(m.treeW), innerH), m.treeW, bodyH, m.focus == focusTree)

	rightW := m.width - m.treeW - gutterWidth
	right := m.panel(m.rightView(innerWidth(rightW), innerH), rightW, bodyH, m.focus == focusRight)

	gutter := lipgloss.NewStyle().Width(gutterWidth).Height(bodyH).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, left, gutter, right) + "\n" + m.statusBar()
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	// Tema gelap default. Set MADESU_THEME=light untuk background terang.
	// (Bubbles/lipgloss tetap meng-query background saat init; itu di luar kendali kode ini.)
	dark := os.Getenv("MADESU_THEME") != "light"
	m, err := newModel(dir, dark)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
