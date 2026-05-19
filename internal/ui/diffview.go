package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/anuchito/replay/internal/navigator"
)

const (
	colorReset = "\x1b[0m"
	colorRed   = "\x1b[31m"
	colorGreen = "\x1b[32m"
	colorCyan  = "\x1b[36m"
	colorBold  = "\x1b[1m"
	colorDim   = "\x1b[2m"
)

// DiffView renders a full-screen view of the next commit's diff.
// It is toggled on/off with Toggle() and renders via Render().
type DiffView struct {
	Active       bool
	scrollOffset int
	diffLines    []string
}

func NewDiffView() *DiffView {
	return &DiffView{}
}

func (dv *DiffView) Toggle() {
	dv.Active = !dv.Active
	dv.scrollOffset = 0
}

// SetDiff replaces the cached diff and resets scroll to top.
func (dv *DiffView) SetDiff(lines []string) {
	dv.diffLines = lines
	dv.scrollOffset = 0
}

func (dv *DiffView) visibleLines(termH int) int {
	// reserved: line 1 (current), line 2 (next header), line H-1 (separator), line H (controls)
	v := termH - 4
	if v < 0 {
		return 0
	}
	return v
}

func (dv *DiffView) ScrollDown(termH int) {
	va := dv.visibleLines(termH)
	max := len(dv.diffLines) - va
	if max < 0 {
		max = 0
	}
	if dv.scrollOffset < max {
		dv.scrollOffset++
	}
}

func (dv *DiffView) ScrollUp() {
	if dv.scrollOffset > 0 {
		dv.scrollOffset--
	}
}

// Render clears the screen and draws the full-screen detail view.
// termW and termH are the current terminal dimensions.
func (dv *DiffView) Render(out io.Writer, termW, termH int, cur, next navigator.Commit, hasNext bool, pos, total int) {
	// Clear screen, cursor home
	fmt.Fprint(out, "\x1b[2J\x1b[H")

	// Line 1: current commit
	curLine := fmt.Sprintf("[%d/%d] %s  %s", pos, total, cur.Hash, cur.Message)
	fmt.Fprintf(out, "%s\r\n", limitWidth(curLine, termW))

	// Line 2: next commit header
	if hasNext {
		label := fmt.Sprintf(" NEXT [%d/%d] %s  %s ", pos+1, total, next.Hash, next.Message)
		pad := termW - len(label) - 2 // 2 for leading "──"
		if pad < 0 {
			pad = 0
		}
		header := "──" + label + strings.Repeat("─", pad)
		fmt.Fprintf(out, "%s%s%s\r\n", colorCyan+colorBold, limitWidth(header, termW), colorReset)
	} else {
		label := "── NEXT ── (end of range)"
		pad := termW - len(label)
		if pad > 0 {
			label += strings.Repeat("─", pad)
		}
		fmt.Fprintf(out, "%s%s%s\r\n", colorDim, limitWidth(label, termW), colorReset)
	}

	// Diff area
	va := dv.visibleLines(termH)
	end := dv.scrollOffset + va
	if end > len(dv.diffLines) {
		end = len(dv.diffLines)
	}

	rendered := 0
	for i := dv.scrollOffset; i < end; i++ {
		raw := limitWidth(dv.diffLines[i], termW)
		fmt.Fprintf(out, "\x1b[2K%s\r\n", colorizeDiffLine(raw))
		rendered++
	}
	// Fill any remaining lines in the diff area with blank cleared lines
	for rendered < va {
		fmt.Fprint(out, "\x1b[2K\r\n")
		rendered++
	}

	// Separator
	fmt.Fprintf(out, "%s\r\n", strings.Repeat("─", termW))

	// Controls / status bar
	scrollInfo := ""
	if len(dv.diffLines) > va && va > 0 {
		shown := dv.scrollOffset + va
		if shown > len(dv.diffLines) {
			shown = len(dv.diffLines)
		}
		scrollInfo = fmt.Sprintf("(%d/%d) ", shown, len(dv.diffLines))
	}
	controls := fmt.Sprintf("j↓ k↑ scroll %s n next  p prev  d details:off  q quit", scrollInfo)
	fmt.Fprintf(out, "%s\r", limitWidth(controls, termW))
}

// colorizeDiffLine applies ANSI color to a raw diff line.
// The line must already be width-limited so color codes don't interfere with truncation.
func colorizeDiffLine(line string) string {
	switch {
	case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"):
		return colorBold + line + colorReset
	case strings.HasPrefix(line, "+"):
		return colorGreen + line + colorReset
	case strings.HasPrefix(line, "-"):
		return colorRed + line + colorReset
	case strings.HasPrefix(line, "@@"):
		return colorCyan + line + colorReset
	case strings.HasPrefix(line, "diff "), strings.HasPrefix(line, "index "):
		return colorBold + line + colorReset
	default:
		return line
	}
}

// limitWidth truncates s to at most maxW bytes (not runes — good enough for ASCII diff output).
func limitWidth(s string, maxW int) string {
	if maxW <= 0 || len(s) <= maxW {
		return s
	}
	return s[:maxW]
}
