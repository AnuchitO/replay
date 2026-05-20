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
	fmt.Fprint(u.out, "Replay Mode\r\n")
	fmt.Fprint(u.out, "-----------\r\n")
	fmt.Fprint(u.out, "n → next\r\n")
	fmt.Fprint(u.out, "p → previous\r\n")
	fmt.Fprint(u.out, "d → toggle next commit diff\r\n")
	fmt.Fprint(u.out, "q → quit\r\n")
	fmt.Fprint(u.out, "\r\n")
}

func (u *UI) PrintCommit(commit navigator.Commit, current, total int) {
	fmt.Fprintf(u.out, "[%d/%d] %s %s\r\n", current, total, commit.Hash, commit.Message)
}

func (u *UI) PrintError(msg string) {
	fmt.Fprintf(u.out, "Error: %s\r\n", msg)
}
