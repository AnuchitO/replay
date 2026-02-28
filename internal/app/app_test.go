package app

import (
	"fmt"
	"testing"

	"github.com/anuchito/replay/internal/navigator"
)

// mockGitClient implements git.GitClient for testing.
type mockGitClient struct {
	isRepo         bool
	isClean        bool
	validateErr    map[string]error
	isAncestor     bool
	commits        []navigator.Commit
	commitRangeErr error
	branch         string
	checkoutCalls  []string
}

func (m *mockGitClient) IsRepo() (bool, error)  { return m.isRepo, nil }
func (m *mockGitClient) IsClean() (bool, error)  { return m.isClean, nil }
func (m *mockGitClient) ValidateCommit(hash string) error {
	if m.validateErr != nil {
		if err, ok := m.validateErr[hash]; ok {
			return err
		}
	}
	return nil
}
func (m *mockGitClient) IsAncestor(_, _ string) (bool, error) { return m.isAncestor, nil }
func (m *mockGitClient) Log(_ int) ([]navigator.Commit, error) { return m.commits, nil }
func (m *mockGitClient) CommitRange(_, _ string) ([]navigator.Commit, error) {
	return m.commits, m.commitRangeErr
}
func (m *mockGitClient) CurrentBranch() (string, error) { return m.branch, nil }
func (m *mockGitClient) Checkout(ref string) error {
	m.checkoutCalls = append(m.checkoutCalls, ref)
	return nil
}

func TestValidate_WithEndCommit(t *testing.T) {
	mock := &mockGitClient{
		isRepo:     true,
		isClean:    true,
		isAncestor: true,
		commits: []navigator.Commit{
			{Hash: "abc1234", Message: "first"},
			{Hash: "def5678", Message: "second"},
		},
		branch: "main",
	}

	opts := RunOptions{
		StartCommit: "abc1234",
		EndCommit:   "def5678",
	}

	err := Validate(mock, opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidate_WithEndCommit_DefaultsToHEAD(t *testing.T) {
	mock := &mockGitClient{
		isRepo:     true,
		isClean:    true,
		isAncestor: true,
		commits: []navigator.Commit{
			{Hash: "abc1234", Message: "first"},
		},
		branch: "main",
	}

	opts := RunOptions{
		StartCommit: "abc1234",
		EndCommit:   "", // empty = HEAD
	}

	err := Validate(mock, opts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidate_InvalidEndCommit(t *testing.T) {
	mock := &mockGitClient{
		isRepo:  true,
		isClean: true,
		validateErr: map[string]error{
			"badcommit": fmt.Errorf("invalid commit: badcommit"),
		},
		branch: "main",
	}

	opts := RunOptions{
		StartCommit: "abc1234",
		EndCommit:   "badcommit",
	}

	err := Validate(mock, opts)
	if err == nil {
		t.Fatal("expected error for invalid end commit, got nil")
	}
}

func TestValidate_StartNotAncestorOfEnd(t *testing.T) {
	mock := &mockGitClient{
		isRepo:     true,
		isClean:    true,
		isAncestor: false,
		branch:     "main",
	}

	opts := RunOptions{
		StartCommit: "abc1234",
		EndCommit:   "def5678",
	}

	err := Validate(mock, opts)
	if err == nil {
		t.Fatal("expected error when start is not ancestor of end, got nil")
	}
}

func TestValidate_NotARepo(t *testing.T) {
	mock := &mockGitClient{
		isRepo: false,
	}

	opts := RunOptions{StartCommit: "abc1234"}

	err := Validate(mock, opts)
	if err == nil {
		t.Fatal("expected error for not a repo, got nil")
	}
}

func TestValidate_DirtyWorkingTree(t *testing.T) {
	mock := &mockGitClient{
		isRepo:  true,
		isClean: false,
	}

	opts := RunOptions{StartCommit: "abc1234"}

	err := Validate(mock, opts)
	if err == nil {
		t.Fatal("expected error for dirty working tree, got nil")
	}
}
