package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anuchito/replay/internal/git"
	"github.com/anuchito/replay/internal/navigator"
	"github.com/anuchito/replay/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: replay <start-commit>\n")
		os.Exit(1)
	}
	startCommit := os.Args[1]

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := git.NewClient(cwd)
	display := ui.New(os.Stdout)

	if err := run(client, display, startCommit); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(client git.GitClient, display *ui.UI, startCommit string) error {
	// 1. Ensure we are inside a Git repository
	isRepo, err := client.IsRepo()
	if err != nil {
		return err
	}
	if !isRepo {
		return fmt.Errorf("not a git repository")
	}

	// 2. Detect dirty working tree
	clean, err := client.IsClean()
	if err != nil {
		return err
	}
	if !clean {
		return fmt.Errorf("working tree is dirty, please commit or stash your changes")
	}

	// 3. Validate the provided commit exists
	if err := client.ValidateCommit(startCommit); err != nil {
		return fmt.Errorf("invalid commit: %s", startCommit)
	}

	// 4. Ensure the commit is an ancestor of HEAD
	isAnc, err := client.IsAncestor(startCommit, "HEAD")
	if err != nil {
		return err
	}
	if !isAnc {
		return fmt.Errorf("commit %s is not an ancestor of HEAD", startCommit)
	}

	// 5. Save original branch/state
	originalRef, err := client.CurrentBranch()
	if err != nil {
		return err
	}

	// 6. Collect commits from start to HEAD
	commits, err := client.CommitRange(startCommit, "HEAD")
	if err != nil {
		return err
	}
	if len(commits) == 0 {
		return fmt.Errorf("no commits in range")
	}

	// 7. Create navigator
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

	// 8. Checkout starting commit
	cur := nav.Current()
	if err := client.Checkout(cur.Hash); err != nil {
		return err
	}

	// 9. Enter interactive mode
	display.PrintBanner()
	pos, total := nav.Position()
	display.PrintCommit(cur, pos, total)

	// Raw terminal input
	buf := make([]byte, 1)
	// Set terminal to raw mode
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
