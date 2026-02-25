package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/anuchito/replay/internal/navigator"
)

type Client struct {
	dir string
}

func NewClient(dir string) *Client {
	return &Client{dir: dir}
}

func (c *Client) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = c.dir
	out, err := cmd.CombinedOutput()
	return strings.TrimRight(string(out), "\n"), err
}

func (c *Client) IsRepo() (bool, error) {
	_, err := c.run("rev-parse", "--git-dir")
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (c *Client) IsClean() (bool, error) {
	out, err := c.run("status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return out == "", nil
}

func (c *Client) ValidateCommit(hash string) error {
	_, err := c.run("cat-file", "-t", hash)
	if err != nil {
		return fmt.Errorf("invalid commit: %s", hash)
	}
	return nil
}

func (c *Client) IsAncestor(commit, of string) (bool, error) {
	_, err := c.run("merge-base", "--is-ancestor", commit, of)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (c *Client) CommitRange(from, to string) ([]navigator.Commit, error) {
	out, err := c.run("log", "--reverse", "--format=%H %s", from+"^.."+to)
	if err != nil {
		// Try without ^ (if from is the root commit)
		out, err = c.run("log", "--reverse", "--format=%H %s", from+".."+to)
		if err != nil {
			return nil, fmt.Errorf("git log: %w", err)
		}
		// Prepend the from commit itself
		fromOut, err := c.run("log", "--format=%H %s", "-1", from)
		if err != nil {
			return nil, fmt.Errorf("git log: %w", err)
		}
		out = fromOut + "\n" + out
	}

	lines := strings.Split(out, "\n")
	var commits []navigator.Commit
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		hash := parts[0][:7]
		msg := ""
		if len(parts) > 1 {
			msg = parts[1]
		}
		commits = append(commits, navigator.Commit{Hash: hash, Message: msg})
	}
	return commits, nil
}

func (c *Client) CurrentBranch() (string, error) {
	out, err := c.run("symbolic-ref", "--short", "HEAD")
	if err != nil {
		// Detached HEAD — return the commit hash
		out, err = c.run("rev-parse", "--short", "HEAD")
		if err != nil {
			return "", fmt.Errorf("git rev-parse: %w", err)
		}
		return out, nil
	}
	return out, nil
}

func (c *Client) Checkout(ref string) error {
	_, err := c.run("checkout", ref)
	if err != nil {
		return fmt.Errorf("git checkout %s: %w", ref, err)
	}
	return nil
}
