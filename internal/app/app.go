package app

import (
	"fmt"

	"github.com/anuchito/replay/internal/git"
)

type RunOptions struct {
	StartCommit string
	EndCommit   string // empty defaults to "HEAD"
}

func (o RunOptions) endRef() string {
	if o.EndCommit == "" {
		return "HEAD"
	}
	return o.EndCommit
}

// Validate checks all preconditions before entering interactive mode.
func Validate(client git.GitClient, opts RunOptions) error {
	isRepo, err := client.IsRepo()
	if err != nil {
		return err
	}
	if !isRepo {
		return fmt.Errorf("not a git repository")
	}

	clean, err := client.IsClean()
	if err != nil {
		return err
	}
	if !clean {
		return fmt.Errorf("working tree is dirty, please commit or stash your changes")
	}

	if err := client.ValidateCommit(opts.StartCommit); err != nil {
		return fmt.Errorf("invalid start commit: %s", opts.StartCommit)
	}

	endRef := opts.endRef()
	if opts.EndCommit != "" {
		if err := client.ValidateCommit(opts.EndCommit); err != nil {
			return fmt.Errorf("invalid end commit: %s", opts.EndCommit)
		}
	}

	isAnc, err := client.IsAncestor(opts.StartCommit, endRef)
	if err != nil {
		return err
	}
	if !isAnc {
		return fmt.Errorf("commit %s is not an ancestor of %s", opts.StartCommit, endRef)
	}

	return nil
}
