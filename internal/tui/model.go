// Package tui renders an interactive terminal UI over a slice of parsed
// log entries. It's intentionally small: a vertical list with arrow-key
// navigation, level-filter cycling, and a free-text grep box.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alive-worker/logsift/internal/parser"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(lipgloss.Color("231")).Background(lipgloss.Color("63"))
	statusStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Padding(0, 1)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(lipgloss.Color("236"))
	levelStyles   = map[string]lipgloss.Style{
		"error": lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
		"warn":  lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		"info":  lipgloss.NewStyle().Foreground(lipgloss.Color("39")),
		"debug": lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
	}
)

const helpLine = "↑/↓ navigate  •  l: cycle level  •  /: grep  •  esc: clear  •  q: quit"

// Model holds the TUI state. It is exposed so cli_test can exercise the
// pure-update logic without launching a terminal.
type Model struct {
	all       []*parser.Entry
	visible   []*parser.Entry
	cursor    int
	height    int
	width     int
	levelFilt string
	grepMode  bool
	grepBuf   string
}

// New builds a Model over the given entries. The caller passes the same
// pre-filtered slice the CLI would have written to stdout.
func New(entries []*parser.Entry) Model {
	m := Model{all: entries, height: 24, width: 80}
	m.recompute()
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.grepMode {
			return m.handleGrep(msg)
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.visible)-1 {
				m.cursor++
			}
		case "l":
			m.cycleLevel()
		case "/":
			m.grepMode = true
		case "esc":
			m.levelFilt = ""
			m.grepBuf = ""
			m.recompute()
		}
	}
	return m, nil
}

func (m *Model) handleGrep(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.grepMode = false
		m.grepBuf = ""
		m.recompute()
	case tea.KeyEnter:
		m.grepMode = false
		m.recompute()
	case tea.KeyBackspace:
		if len(m.grepBuf) > 0 {
			m.grepBuf = m.grepBuf[:len(m.grepBuf)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.grepBuf += msg.String()
		}
	}
	return *m, nil
}

func (m *Model) cycleLevel() {
	order := []string{"", "error", "warn", "info", "debug"}
	for i, lvl := range order {
		if lvl == m.levelFilt {
			m.levelFilt = order[(i+1)%len(order)]
			break
		}
	}
	m.recompute()
}

func (m *Model) recompute() {
	out := make([]*parser.Entry, 0, len(m.all))
	for _, e := range m.all {
		if m.levelFilt != "" && !strings.EqualFold(e.Level, m.levelFilt) {
			continue
		}
		if m.grepBuf != "" && !strings.Contains(strings.ToLower(e.Message), strings.ToLower(m.grepBuf)) {
			continue
		}
		out = append(out, e)
	}
	m.visible = out
	if m.cursor >= len(out) {
		m.cursor = max(0, len(out)-1)
	}
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("logsift TUI"))
	b.WriteString("  ")
	b.WriteString(statusStyle.Render(fmt.Sprintf(
		"showing %d/%d  level=%s  grep=%q%s",
		len(m.visible), len(m.all),
		fallback(m.levelFilt, "*"),
		m.grepBuf,
		ifs(m.grepMode, "  (typing…)", ""),
	)))
	b.WriteString("\n")

	listHeight := max(4, m.height-4)
	start := 0
	if m.cursor >= listHeight {
		start = m.cursor - listHeight + 1
	}
	end := min(len(m.visible), start+listHeight)
	for i := start; i < end; i++ {
		line := formatLine(m.visible[i])
		if i == m.cursor {
			line = selectedStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	for i := end - start; i < listHeight; i++ {
		b.WriteString("\n")
	}
	b.WriteString(statusStyle.Render(helpLine))
	return b.String()
}

func formatLine(e *parser.Entry) string {
	ts := "????"
	if !e.Timestamp.IsZero() {
		ts = e.Timestamp.Format("15:04:05")
	}
	level := strings.ToLower(e.Level)
	style, ok := levelStyles[level]
	if !ok {
		style = lipgloss.NewStyle()
	}
	return fmt.Sprintf("%s %s %-12s %s",
		ts,
		style.Render(fmt.Sprintf("%-5s", strings.ToUpper(level))),
		e.Service,
		e.Message,
	)
}

// Visible exposes the currently-filtered slice for tests.
func (m Model) Visible() []*parser.Entry { return m.visible }

// Cursor exposes the cursor index for tests.
func (m Model) Cursor() int { return m.cursor }

// LevelFilter exposes the level filter for tests.
func (m Model) LevelFilter() string { return m.levelFilt }

func fallback(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func ifs(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
