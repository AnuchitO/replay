package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/anuchito/replay/internal/navigator"
)

// setupTestRepo creates a temporary git repo with n commits and returns
// the repo path and a cleanup function.
func setupTestRepo(t *testing.T, numCommits int) (string, []string) {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %s\n%s", args, err, out)
		}
		return string(out)
	}

	run("init")
	run("checkout", "-b", "main")

	var hashes []string
	for i := 1; i <= numCommits; i++ {
		file := filepath.Join(dir, "file.txt")
		content := []byte("commit " + string(rune('0'+i)))
		if err := os.WriteFile(file, content, 0644); err != nil {
			t.Fatal(err)
		}
		run("add", ".")
		run("commit", "-m", "commit "+string(rune('0'+i)))
		hash := run("rev-parse", "HEAD")
		hashes = append(hashes, trimNewline(hash))
	}

	return dir, hashes
}

func trimNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}

func TestIsRepo_InRepo(t *testing.T) {
	dir, _ := setupTestRepo(t, 1)
	client := NewClient(dir)

	isRepo, err := client.IsRepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isRepo {
		t.Error("expected true, got false")
	}
}

func TestIsRepo_NotInRepo(t *testing.T) {
	dir := t.TempDir()
	client := NewClient(dir)

	isRepo, err := client.IsRepo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isRepo {
		t.Error("expected false, got true")
	}
}

func TestIsClean_CleanRepo(t *testing.T) {
	dir, _ := setupTestRepo(t, 1)
	client := NewClient(dir)

	clean, err := client.IsClean()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !clean {
		t.Error("expected clean repo, got dirty")
	}
}

func TestIsClean_DirtyRepo(t *testing.T) {
	dir, _ := setupTestRepo(t, 1)
	client := NewClient(dir)

	// Create an untracked modification
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty"), 0644); err != nil {
		t.Fatal(err)
	}

	clean, err := client.IsClean()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clean {
		t.Error("expected dirty repo, got clean")
	}
}

func TestValidateCommit_Valid(t *testing.T) {
	dir, hashes := setupTestRepo(t, 1)
	client := NewClient(dir)

	err := client.ValidateCommit(hashes[0])
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateCommit_Invalid(t *testing.T) {
	dir, _ := setupTestRepo(t, 1)
	client := NewClient(dir)

	err := client.ValidateCommit("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if err == nil {
		t.Error("expected error for invalid commit, got nil")
	}
}

func TestIsAncestor_True(t *testing.T) {
	dir, hashes := setupTestRepo(t, 3)
	client := NewClient(dir)

	isAnc, err := client.IsAncestor(hashes[0], hashes[2])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isAnc {
		t.Error("expected true, got false")
	}
}

func TestIsAncestor_False(t *testing.T) {
	dir, hashes := setupTestRepo(t, 3)
	client := NewClient(dir)

	isAnc, err := client.IsAncestor(hashes[2], hashes[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isAnc {
		t.Error("expected false, got true")
	}
}

func TestCommitRange(t *testing.T) {
	dir, hashes := setupTestRepo(t, 3)
	client := NewClient(dir)

	commits, err := client.CommitRange(hashes[0], hashes[2])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 3 {
		t.Fatalf("expected 3 commits, got %d", len(commits))
	}
	// Should be in chronological order (oldest first)
	if commits[0].Hash != hashes[0][:7] && commits[0].Hash != hashes[0] {
		t.Errorf("expected first commit hash to match, got %s", commits[0].Hash)
	}
}

func TestCurrentBranch(t *testing.T) {
	dir, _ := setupTestRepo(t, 1)
	client := NewClient(dir)

	branch, err := client.CurrentBranch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "main" {
		t.Errorf("expected 'main', got '%s'", branch)
	}
}

func TestCheckout(t *testing.T) {
	dir, hashes := setupTestRepo(t, 3)
	client := NewClient(dir)

	err := client.Checkout(hashes[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify HEAD is at the first commit
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	if trimNewline(string(out)) != hashes[0] {
		t.Errorf("expected HEAD at %s, got %s", hashes[0], trimNewline(string(out)))
	}
}

// Ensure CommitRange returns navigator.Commit type
func TestCommitRange_ReturnsNavigatorCommits(t *testing.T) {
	dir, hashes := setupTestRepo(t, 2)
	client := NewClient(dir)

	commits, err := client.CommitRange(hashes[0], hashes[1])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Type assertion to verify it returns navigator.Commit
	var _ []navigator.Commit = commits
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
}
