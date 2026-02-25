package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/anuchito/replay/internal/navigator"
)

func TestPrintBanner(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)
	u.PrintBanner()

	out := buf.String()
	if !strings.Contains(out, "Replay Mode") {
		t.Error("banner should contain 'Replay Mode'")
	}
	if !strings.Contains(out, "n") {
		t.Error("banner should mention 'n' key")
	}
	if !strings.Contains(out, "p") {
		t.Error("banner should mention 'p' key")
	}
	if !strings.Contains(out, "q") {
		t.Error("banner should mention 'q' key")
	}
}

func TestPrintCommit(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)

	commit := navigator.Commit{Hash: "abc1234", Message: "add feature"}
	u.PrintCommit(commit, 3, 15)

	out := buf.String()
	if !strings.Contains(out, "abc1234") {
		t.Error("should contain commit hash")
	}
	if !strings.Contains(out, "add feature") {
		t.Error("should contain commit message")
	}
	if !strings.Contains(out, "3/15") {
		t.Error("should contain position indicator")
	}
}

func TestPrintError(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf)
	u.PrintError("something went wrong")

	out := buf.String()
	if !strings.Contains(out, "something went wrong") {
		t.Error("should contain error message")
	}
}
