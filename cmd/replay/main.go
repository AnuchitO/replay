package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"golang.org/x/term"

	"github.com/anuchito/replay/internal/app"
	"github.com/anuchito/replay/internal/git"
	"github.com/anuchito/replay/internal/navigator"
	"github.com/anuchito/replay/internal/ui"
)

const defaultLogSize = 30

// version is set at build time via -ldflags "-X main.version=<tag>".
// Empty string means not set — getVersion() will fall back to build info.
var version = ""

func getVersion() string {
	if version != "" {
		return version // explicitly set via ldflags (e.g. "dev" or "v1.1.2")
	}
	// Not built with ldflags — try module version embedded by go install.
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := git.NewClient(cwd)
	display := ui.New(os.Stdout)

	// Handle help/version flags
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "-h", "--help", "help":
			printUsage()
			os.Exit(0)
		case "-v", "--version", "version":
			fmt.Printf("replay %s\n", getVersion())
			os.Exit(0)
		}
	}

	var opts app.RunOptions

	switch len(os.Args) {
	case 1:
		// No args — show interactive picker
		selected, err := pickStartCommit(client)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if selected == nil {
			os.Exit(0)
		}
		opts.StartCommit = selected.Hash
	case 2:
		opts.StartCommit = os.Args[1]
	default:
		opts.StartCommit = os.Args[1]
		opts.EndCommit = os.Args[2]
	}

	if err := run(client, display, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func pickStartCommit(client git.GitClient) (*navigator.Commit, error) {
	isRepo, err := client.IsRepo()
	if err != nil {
		return nil, err
	}
	if !isRepo {
		return nil, fmt.Errorf("not a git repository")
	}

	commits, err := client.Log(defaultLogSize)
	if err != nil {
		return nil, err
	}
	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found")
	}

	// Enter raw mode for the picker
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to set raw mode: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	return ui.PickCommit(commits, os.Stdin, os.Stdout, 20)
}

func run(client git.GitClient, display *ui.UI, opts app.RunOptions) error {
	// Validate preconditions
	if err := app.Validate(client, opts); err != nil {
		return err
	}

	// Save original branch/state
	originalRef, err := client.CurrentBranch()
	if err != nil {
		return err
	}

	// Collect commits
	commits, err := client.CommitRange(opts.StartCommit, opts.EndRef())
	if err != nil {
		return err
	}
	if len(commits) == 0 {
		return fmt.Errorf("no commits in range")
	}

	// Create navigator
	nav, err := navigator.NewNavigator(commits)
	if err != nil {
		return err
	}

	// Setup Ctrl+C handler to restore state
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Print("\r\nRestoring original state...\r\n")
		client.Checkout(originalRef)
		os.Exit(0)
	}()

	// Ensure we restore state on exit
	defer func() {
		client.Checkout(originalRef)
	}()

	// Checkout starting commit
	cur := nav.Current()
	if err := client.Checkout(cur.Hash); err != nil {
		return err
	}

	// Enter interactive mode
	display.PrintBanner()
	pos, total := nav.Position()
	display.PrintCommit(cur, pos, total)

	// Raw terminal input
	buf := make([]byte, 1)
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	dv := ui.NewDiffView()

	// loadNextDiff fetches the diff for the next commit and caches it in dv.
	loadNextDiff := func() {
		next, ok := nav.Peek()
		if !ok {
			dv.SetDiff(nil)
			return
		}
		lines, err := client.ShowDiff(next.Hash)
		if err != nil {
			dv.SetDiff(nil)
			return
		}
		dv.SetDiff(lines)
	}

	// renderDetail performs a full-screen redraw of the detail view.
	renderDetail := func() {
		termW, termH, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			termW, termH = 80, 24
		}
		cur := nav.Current()
		pos, total := nav.Position()
		next, hasNext := nav.Peek()
		dv.Render(os.Stdout, termW, termH, cur, next, hasNext, pos, total)
	}

	// exitDetail clears the screen and returns to append-style output.
	exitDetail := func() {
		fmt.Print("\x1b[2J\x1b[H")
		display.PrintBanner()
		cur := nav.Current()
		pos, total := nav.Position()
		display.PrintCommit(cur, pos, total)
	}

	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			return err
		}

		switch buf[0] {
		case 'n':
			if err := nav.Next(); err != nil {
				if !dv.Active {
					display.PrintError(err.Error())
				}
				continue
			}
			cur = nav.Current()
			if err := client.Checkout(cur.Hash); err != nil {
				return err
			}
			if dv.Active {
				loadNextDiff()
				renderDetail()
			} else {
				pos, total = nav.Position()
				display.PrintCommit(cur, pos, total)
			}

		case 'p':
			if err := nav.Prev(); err != nil {
				if !dv.Active {
					display.PrintError(err.Error())
				}
				continue
			}
			cur = nav.Current()
			if err := client.Checkout(cur.Hash); err != nil {
				return err
			}
			if dv.Active {
				loadNextDiff()
				renderDetail()
			} else {
				pos, total = nav.Position()
				display.PrintCommit(cur, pos, total)
			}

		case 'd':
			dv.Toggle()
			if dv.Active {
				loadNextDiff()
				renderDetail()
			} else {
				exitDetail()
			}

		case 'j':
			if dv.Active {
				_, termH, _ := term.GetSize(int(os.Stdout.Fd()))
				dv.ScrollDown(termH)
				renderDetail()
			}

		case 'k':
			if dv.Active {
				_, termH, _ := term.GetSize(int(os.Stdout.Fd()))
				dv.ScrollUp(termH)
				renderDetail()
			}

		case 4: // Ctrl+D — half page down
			if dv.Active {
				_, termH, _ := term.GetSize(int(os.Stdout.Fd()))
				dv.ScrollHalfDown(termH)
				renderDetail()
			}

		case 21: // Ctrl+U — half page up
			if dv.Active {
				_, termH, _ := term.GetSize(int(os.Stdout.Fd()))
				dv.ScrollHalfUp(termH)
				renderDetail()
			}

		case ' ': // Space — full page down
			if dv.Active {
				_, termH, _ := term.GetSize(int(os.Stdout.Fd()))
				dv.ScrollPageDown(termH)
				renderDetail()
			}

		case 0x1b: // escape sequence (arrow keys)
			seq := make([]byte, 2)
			if _, err := io.ReadFull(os.Stdin, seq); err != nil {
				continue
			}
			if seq[0] != '[' {
				continue
			}
			switch seq[1] {
			case 'A': // arrow up
				if dv.Active {
					_, termH, _ := term.GetSize(int(os.Stdout.Fd()))
					dv.ScrollUp(termH)
					renderDetail()
				}
			case 'B': // arrow down
				if dv.Active {
					_, termH, _ := term.GetSize(int(os.Stdout.Fd()))
					dv.ScrollDown(termH)
					renderDetail()
				}
			}

		case 'q':
			if dv.Active {
				fmt.Print("\x1b[2J\x1b[H")
			}
			fmt.Print("\r\nRestoring original state...\r\n")
			return nil

		case 3: // Ctrl+C
			if dv.Active {
				fmt.Print("\x1b[2J\x1b[H")
			}
			fmt.Print("\r\nRestoring original state...\r\n")
			return nil
		}
	}
}

func printUsage() {
	fmt.Printf("replay %s - interactively navigate Git commit history\n", getVersion())
	fmt.Print(`
Usage:
  replay                          Select a commit interactively
  replay <start-commit>           Replay from commit to HEAD
  replay <start-commit> <end>     Replay from commit to end commit
  replay -h, --help               Show this help
  replay -v, --version            Show version

Interactive picker controls:
  j / ↓      Move down
  k / ↑      Move up
  Enter      Select commit
  q          Quit

Replay mode controls:
  n          Next commit
  p          Previous commit
  d          Toggle next-commit diff preview (on/off)
  j / ↓      Scroll diff down            (detail mode)
  k / ↑      Scroll diff up              (detail mode)
  Ctrl+D     Scroll half page down       (detail mode)
  Ctrl+U     Scroll half page up         (detail mode)
  Space      Scroll full page down       (detail mode)
  q          Quit and restore original state
  Ctrl+C     Quit and restore original state

Examples:
  replay                          Browse and pick a commit
  replay abc1234                  Replay from abc1234 to HEAD
  replay abc1234 def5678          Replay from abc1234 to def5678
`)
}
