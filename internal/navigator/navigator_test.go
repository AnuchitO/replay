package navigator

import (
	"testing"
)

func TestNewNavigator_WithCommits(t *testing.T) {
	commits := []Commit{
		{Hash: "abc1234", Message: "first commit"},
		{Hash: "def5678", Message: "second commit"},
		{Hash: "ghi9012", Message: "third commit"},
	}

	nav, err := NewNavigator(commits)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if nav == nil {
		t.Fatal("expected navigator, got nil")
	}
}

func TestNewNavigator_EmptyCommits(t *testing.T) {
	_, err := NewNavigator([]Commit{})
	if err == nil {
		t.Fatal("expected error for empty commits, got nil")
	}
}

func TestNavigator_Current(t *testing.T) {
	commits := []Commit{
		{Hash: "abc1234", Message: "first commit"},
		{Hash: "def5678", Message: "second commit"},
	}

	nav, _ := NewNavigator(commits)
	cur := nav.Current()

	if cur.Hash != "abc1234" {
		t.Errorf("expected hash abc1234, got %s", cur.Hash)
	}
	if cur.Message != "first commit" {
		t.Errorf("expected message 'first commit', got %s", cur.Message)
	}
}

func TestNavigator_Next(t *testing.T) {
	commits := []Commit{
		{Hash: "abc1234", Message: "first commit"},
		{Hash: "def5678", Message: "second commit"},
		{Hash: "ghi9012", Message: "third commit"},
	}

	nav, _ := NewNavigator(commits)

	err := nav.Next()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cur := nav.Current()
	if cur.Hash != "def5678" {
		t.Errorf("expected hash def5678, got %s", cur.Hash)
	}
}

func TestNavigator_Next_AtEnd(t *testing.T) {
	commits := []Commit{
		{Hash: "abc1234", Message: "first commit"},
	}

	nav, _ := NewNavigator(commits)

	err := nav.Next()
	if err == nil {
		t.Fatal("expected error at end, got nil")
	}
}

func TestNavigator_Prev(t *testing.T) {
	commits := []Commit{
		{Hash: "abc1234", Message: "first commit"},
		{Hash: "def5678", Message: "second commit"},
	}

	nav, _ := NewNavigator(commits)
	nav.Next()

	err := nav.Prev()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cur := nav.Current()
	if cur.Hash != "abc1234" {
		t.Errorf("expected hash abc1234, got %s", cur.Hash)
	}
}

func TestNavigator_Prev_AtStart(t *testing.T) {
	commits := []Commit{
		{Hash: "abc1234", Message: "first commit"},
	}

	nav, _ := NewNavigator(commits)

	err := nav.Prev()
	if err == nil {
		t.Fatal("expected error at start, got nil")
	}
}

func TestNavigator_Position(t *testing.T) {
	commits := []Commit{
		{Hash: "abc1234", Message: "first commit"},
		{Hash: "def5678", Message: "second commit"},
		{Hash: "ghi9012", Message: "third commit"},
	}

	nav, _ := NewNavigator(commits)

	cur, total := nav.Position()
	if cur != 1 {
		t.Errorf("expected current 1, got %d", cur)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}

	nav.Next()
	cur, total = nav.Position()
	if cur != 2 {
		t.Errorf("expected current 2, got %d", cur)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
}
