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
	fmt.Fprintln(w, "j/↓ down  k/↑ up  Enter select  q quit")
	fmt.Fprintln(w)

	end := p.offset + p.pageSize
	if end > len(p.commits) {
		end = len(p.commits)
	}

	for i := p.offset; i < end; i++ {
		c := p.commits[i]
		if i == p.cursor {
			fmt.Fprintf(w, "> %s %s\n", c.Hash, c.Message)
		} else {
			fmt.Fprintf(w, "  %s %s\n", c.Hash, c.Message)
		}
	}
}

// visibleLines returns how many commit lines are currently displayed.
func (p *Picker) visibleLines() int {
	end := p.offset + p.pageSize
	if end > len(p.commits) {
		end = len(p.commits)
	}
	return end - p.offset
}

// renderRaw writes the picker list for raw terminal mode using \r\n.
func (p *Picker) renderRaw(w io.Writer) {
	end := p.offset + p.pageSize
	if end > len(p.commits) {
		end = len(p.commits)
	}

	for i := p.offset; i < end; i++ {
		c := p.commits[i]
		if i == p.cursor {
			fmt.Fprintf(w, "\x1b[2K> %s %s\r\n", c.Hash, c.Message)
		} else {
			fmt.Fprintf(w, "\x1b[2K  %s %s\r\n", c.Hash, c.Message)
		}
	}
}

// PickCommit runs an interactive picker loop.
// Returns the selected commit, or nil if the user quits.
func PickCommit(commits []navigator.Commit, in io.Reader, out io.Writer, pageSize int) (*navigator.Commit, error) {
	p := NewPicker(commits, pageSize)

	// Print header (stays fixed)
	fmt.Fprint(out, "Select a commit to replay from:\r\n")
	fmt.Fprint(out, "j/↓ down  k/↑ up  Enter select  q quit\r\n")
	fmt.Fprint(out, "\r\n")

	// Initial render
	p.renderRaw(out)

	oneByte := make([]byte, 1)
	for {
		_, err := in.Read(oneByte)
		if err != nil {
			return nil, err
		}

		switch oneByte[0] {
		case 'j':
			p.MoveDown()
		case 'k':
			p.MoveUp()
		case '\r', '\n':
			selected := p.Selected()
			return &selected, nil
		case 'q':
			return nil, nil
		case 3: // Ctrl+C
			return nil, nil
		case 0x1b: // escape sequence (arrow keys)
			seq := make([]byte, 2)
			_, err := io.ReadFull(in, seq)
			if err != nil {
				continue
			}
			if seq[0] == '[' {
				switch seq[1] {
				case 'A': // arrow up
					p.MoveUp()
				case 'B': // arrow down
					p.MoveDown()
				}
			}
		default:
			continue
		}

		// Move cursor up to beginning of list and re-render in place
		lines := p.visibleLines()
		fmt.Fprintf(out, "\x1b[%dA", lines) // move up N lines
		p.renderRaw(out)
	}
}
