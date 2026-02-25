package ui

import (
	"fmt"
	"io"

	"github.com/anuchito/replay/internal/navigator"
)

type UI struct {
	out io.Writer
}

func New(out io.Writer) *UI {
	return &UI{out: out}
}

func (u *UI) PrintBanner() {
	fmt.Fprintln(u.out, "Replay Mode")
	fmt.Fprintln(u.out, "-----------")
	fmt.Fprintln(u.out, "n → next")
	fmt.Fprintln(u.out, "p → previous")
	fmt.Fprintln(u.out, "q → quit")
	fmt.Fprintln(u.out)
}

func (u *UI) PrintCommit(commit navigator.Commit, current, total int) {
	fmt.Fprintf(u.out, "[%d/%d] %s %s\n", current, total, commit.Hash, commit.Message)
}

func (u *UI) PrintError(msg string) {
	fmt.Fprintf(u.out, "Error: %s\n", msg)
}
