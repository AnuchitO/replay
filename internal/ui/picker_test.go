package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/anuchito/replay/internal/navigator"
)

func sampleCommits() []navigator.Commit {
	return []navigator.Commit{
		{Hash: "abc1234", Message: "first commit"},
		{Hash: "def5678", Message: "second commit"},
		{Hash: "ghi9012", Message: "third commit"},
		{Hash: "jkl3456", Message: "fourth commit"},
		{Hash: "mno7890", Message: "fifth commit"},
	}
}

func TestNewPicker(t *testing.T) {
	p := NewPicker(sampleCommits(), 5)
	if p == nil {
		t.Fatal("expected picker, got nil")
	}
}

func TestPicker_MoveDown(t *testing.T) {
	p := NewPicker(sampleCommits(), 5)
	p.MoveDown()

	selected := p.Selected()
	if selected.Hash != "def5678" {
		t.Errorf("expected cursor at def5678, got %s", selected.Hash)
	}
}

func TestPicker_MoveUp(t *testing.T) {
	p := NewPicker(sampleCommits(), 5)
	p.MoveDown()
	p.MoveDown()
	p.MoveUp()

	selected := p.Selected()
	if selected.Hash != "def5678" {
		t.Errorf("expected cursor at def5678, got %s", selected.Hash)
	}
}

func TestPicker_MoveDown_AtBottom(t *testing.T) {
	commits := sampleCommits()
	p := NewPicker(commits, 5)
	// Move to the last commit
	for i := 0; i < len(commits); i++ {
		p.MoveDown()
	}

	selected := p.Selected()
	if selected.Hash != "mno7890" {
		t.Errorf("expected cursor at mno7890 (last), got %s", selected.Hash)
	}
}

func TestPicker_MoveUp_AtTop(t *testing.T) {
	p := NewPicker(sampleCommits(), 5)
	p.MoveUp() // should stay at top

	selected := p.Selected()
	if selected.Hash != "abc1234" {
		t.Errorf("expected cursor at abc1234 (first), got %s", selected.Hash)
	}
}

func TestPicker_Selected(t *testing.T) {
	p := NewPicker(sampleCommits(), 5)

	selected := p.Selected()
	if selected.Hash != "abc1234" {
		t.Errorf("expected abc1234, got %s", selected.Hash)
	}
	if selected.Message != "first commit" {
		t.Errorf("expected 'first commit', got %s", selected.Message)
	}
}

func TestPicker_Render(t *testing.T) {
	p := NewPicker(sampleCommits(), 5)
	var buf bytes.Buffer
	p.Render(&buf)

	out := buf.String()

	// Should contain header
	if !strings.Contains(out, "Select a commit") {
		t.Error("render should contain header text")
	}

	// Should contain all commit hashes
	for _, c := range sampleCommits() {
		if !strings.Contains(out, c.Hash) {
			t.Errorf("render should contain hash %s", c.Hash)
		}
	}

	// First commit should have cursor indicator
	lines := strings.Split(out, "\n")
	foundCursor := false
	for _, line := range lines {
		if strings.Contains(line, "abc1234") && strings.Contains(line, "▸") {
			foundCursor = true
			break
		}
	}
	if !foundCursor {
		t.Error("render should show cursor ▸ on first commit")
	}
}

func TestPicker_Render_AfterMove(t *testing.T) {
	p := NewPicker(sampleCommits(), 5)
	p.MoveDown()

	var buf bytes.Buffer
	p.Render(&buf)

	out := buf.String()
	lines := strings.Split(out, "\n")

	// Cursor should be on second commit
	foundCursor := false
	for _, line := range lines {
		if strings.Contains(line, "def5678") && strings.Contains(line, "▸") {
			foundCursor = true
			break
		}
	}
	if !foundCursor {
		t.Error("render should show cursor ▸ on second commit after MoveDown")
	}
}

func TestPicker_Render_Scrolling(t *testing.T) {
	// Page size 3, 5 commits — should scroll
	p := NewPicker(sampleCommits(), 3)

	// Move to 4th commit (index 3), beyond page size
	p.MoveDown()
	p.MoveDown()
	p.MoveDown()

	var buf bytes.Buffer
	p.Render(&buf)

	out := buf.String()

	// 4th commit should be visible
	if !strings.Contains(out, "jkl3456") {
		t.Error("after scrolling, jkl3456 should be visible")
	}
}
