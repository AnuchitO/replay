package ui

import (
	"bytes"
	"testing"

	"github.com/anuchito/replay/internal/navigator"
)

func TestPickCommit_Enter(t *testing.T) {
	commits := []navigator.Commit{
		{Hash: "abc1234", Message: "first commit"},
		{Hash: "def5678", Message: "second commit"},
		{Hash: "ghi9012", Message: "third commit"},
	}

	// Simulate: press 'j' (down), then Enter
	input := bytes.NewReader([]byte{'j', '\r'})
	var output bytes.Buffer

	commit, err := PickCommit(commits, input, &output, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if commit == nil {
		t.Fatal("expected a commit, got nil")
	}
	if commit.Hash != "def5678" {
		t.Errorf("expected def5678, got %s", commit.Hash)
	}
}

func TestPickCommit_Quit(t *testing.T) {
	commits := []navigator.Commit{
		{Hash: "abc1234", Message: "first commit"},
	}

	input := bytes.NewReader([]byte{'q'})
	var output bytes.Buffer

	commit, err := PickCommit(commits, input, &output, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if commit != nil {
		t.Error("expected nil commit on quit")
	}
}

func TestPickCommit_VimKeys(t *testing.T) {
	commits := []navigator.Commit{
		{Hash: "abc1234", Message: "first commit"},
		{Hash: "def5678", Message: "second commit"},
		{Hash: "ghi9012", Message: "third commit"},
	}

	// Press 'j' twice (down to third), 'k' once (back to second), Enter
	input := bytes.NewReader([]byte{'j', 'j', 'k', '\r'})
	var output bytes.Buffer

	commit, err := PickCommit(commits, input, &output, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if commit.Hash != "def5678" {
		t.Errorf("expected def5678, got %s", commit.Hash)
	}
}

func TestPickCommit_ArrowKeys(t *testing.T) {
	commits := []navigator.Commit{
		{Hash: "abc1234", Message: "first commit"},
		{Hash: "def5678", Message: "second commit"},
		{Hash: "ghi9012", Message: "third commit"},
	}

	// Arrow down is ESC [ B (0x1b, 0x5b, 0x42)
	// Arrow down twice, arrow up once, Enter
	input := bytes.NewReader([]byte{
		0x1b, '[', 'B', // down
		0x1b, '[', 'B', // down
		0x1b, '[', 'A', // up
		'\r',           // enter
	})
	var output bytes.Buffer

	commit, err := PickCommit(commits, input, &output, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if commit.Hash != "def5678" {
		t.Errorf("expected def5678, got %s", commit.Hash)
	}
}
