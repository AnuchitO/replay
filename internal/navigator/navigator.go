package navigator

import "errors"

var (
	ErrEmptyCommits = errors.New("commits list is empty")
	ErrAtEnd        = errors.New("already at last commit")
	ErrAtStart      = errors.New("already at first commit")
)

type Commit struct {
	Hash    string
	Message string
}

type Navigator struct {
	commits []Commit
	current int
}

func NewNavigator(commits []Commit) (*Navigator, error) {
	if len(commits) == 0 {
		return nil, ErrEmptyCommits
	}
	return &Navigator{commits: commits, current: 0}, nil
}

func (n *Navigator) Current() Commit {
	return n.commits[n.current]
}

func (n *Navigator) Next() error {
	if n.current >= len(n.commits)-1 {
		return ErrAtEnd
	}
	n.current++
	return nil
}

func (n *Navigator) Prev() error {
	if n.current <= 0 {
		return ErrAtStart
	}
	n.current--
	return nil
}

func (n *Navigator) Position() (int, int) {
	return n.current + 1, len(n.commits)
}
