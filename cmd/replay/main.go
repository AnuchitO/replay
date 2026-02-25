package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anuchito/replay/internal/app"
	"github.com/anuchito/replay/internal/git"
	"github.com/anuchito/replay/internal/navigator"
	"github.com/anuchito/replay/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: replay <start-commit> [end-commit]\n")
		os.Exit(1)
	}

	opts := app.RunOptions{
		StartCommit: os.Args[1],
	}
	if len(os.Args) >= 3 {
		opts.EndCommit = os.Args[2]
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := git.NewClient(cwd)
	display := ui.New(os.Stdout)

	if err := run(client, display, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
		fmt.Println("\nRestoring original state...")
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
	oldState, err := makeRaw(os.Stdin.Fd())
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %v", err)
	}
	defer restore(os.Stdin.Fd(), oldState)

	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			return err
		}

		switch buf[0] {
		case 'n':
			if err := nav.Next(); err != nil {
				display.PrintError(err.Error())
				continue
			}
			cur = nav.Current()
			if err := client.Checkout(cur.Hash); err != nil {
				return err
			}
			pos, total = nav.Position()
			display.PrintCommit(cur, pos, total)

		case 'p':
			if err := nav.Prev(); err != nil {
				display.PrintError(err.Error())
				continue
			}
			cur = nav.Current()
			if err := client.Checkout(cur.Hash); err != nil {
				return err
			}
			pos, total = nav.Position()
			display.PrintCommit(cur, pos, total)

		case 'q':
			fmt.Println("\nRestoring original state...")
			return nil

		case 3: // Ctrl+C
			fmt.Println("\nRestoring original state...")
			return nil
		}
	}
}
