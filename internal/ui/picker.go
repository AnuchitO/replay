package ui

import (
	"fmt"
	"io"

	"github.com/anuchito/replay/internal/navigator"
)

type Picker struct {
	commits  []navigator.Commit
	cursor   int
	offset   int
	pageSize int
}

func NewPicker(commits []navigator.Commit, pageSize int) *Picker {
	return &Picker{
		commits:  commits,
		cursor:   0,
		offset:   0,
		pageSize: pageSize,
	}
}

func (p *Picker) MoveDown() {
	if p.cursor < len(p.commits)-1 {
		p.cursor++
		if p.cursor >= p.offset+p.pageSize {
			p.offset = p.cursor - p.pageSize + 1
		}
	}
}

func (p *Picker) MoveUp() {
	if p.cursor > 0 {
		p.cursor--
		if p.cursor < p.offset {
			p.offset = p.cursor
		}
	}
}

func (p *Picker) Selected() navigator.Commit {
	return p.commits[p.cursor]
}

func (p *Picker) Render(w io.Writer) {
	fmt.Fprintln(w, "Select a commit to replay from:")
	fmt.Fprintln(w)

	end := p.offset + p.pageSize
	if end > len(p.commits) {
		end = len(p.commits)
	}

	for i := p.offset; i < end; i++ {
		c := p.commits[i]
		if i == p.cursor {
			fmt.Fprintf(w, "  ▸ %s  %s\n", c.Hash, c.Message)
		} else {
			fmt.Fprintf(w, "    %s  %s\n", c.Hash, c.Message)
		}
	}
}
