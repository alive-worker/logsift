package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alive-worker/logsift/internal/parser"
)

func sample() []*parser.Entry {
	ts := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	return []*parser.Entry{
		{Level: "info", Service: "api", Message: "started", Timestamp: ts},
		{Level: "warn", Service: "api", Message: "slow query", Timestamp: ts.Add(time.Second)},
		{Level: "error", Service: "worker", Message: "timeout", Timestamp: ts.Add(2 * time.Second)},
		{Level: "error", Service: "api", Message: "upstream gone", Timestamp: ts.Add(3 * time.Second)},
	}
}

func TestNew_showsAllEntries(t *testing.T) {
	m := New(sample())
	if len(m.Visible()) != 4 {
		t.Fatalf("want 4 visible, got %d", len(m.Visible()))
	}
}

func TestCursorMoves(t *testing.T) {
	m := New(sample())
	got, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	mm := got.(Model)
	if mm.Cursor() != 1 {
		t.Fatalf("cursor want 1, got %d", mm.Cursor())
	}
	got, _ = mm.Update(tea.KeyMsg{Type: tea.KeyUp})
	if got.(Model).Cursor() != 0 {
		t.Fatalf("cursor should return to 0")
	}
}

func TestCursorClampsAtBottom(t *testing.T) {
	m := New(sample())
	for i := 0; i < 10; i++ {
		got, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = got.(Model)
	}
	if m.Cursor() != 3 {
		t.Fatalf("cursor should stop at last index 3, got %d", m.Cursor())
	}
}

func TestCycleLevelFiltersInPlace(t *testing.T) {
	m := New(sample())
	// One 'l' press → level=error
	got, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = got.(Model)
	if m.LevelFilter() != "error" {
		t.Fatalf("first cycle want error, got %q", m.LevelFilter())
	}
	if len(m.Visible()) != 2 {
		t.Fatalf("error filter should leave 2 entries, got %d", len(m.Visible()))
	}
}

func TestEscClearsFilters(t *testing.T) {
	m := New(sample())
	got, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = got.(Model)
	got, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = got.(Model)
	if m.LevelFilter() != "" || len(m.Visible()) != 4 {
		t.Fatalf("esc should reset; level=%q visible=%d", m.LevelFilter(), len(m.Visible()))
	}
}

func TestViewRenders(t *testing.T) {
	m := New(sample())
	v := m.View()
	if len(v) == 0 {
		t.Fatalf("view empty")
	}
	// Sanity: status line should report 4/4.
	if !contains(v, "showing 4/4") {
		t.Fatalf("status line missing in view:\n%s", v)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
