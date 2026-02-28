You are an expert Go systems engineer.

We are building a production-quality CLI tool called "replay".

The tool must be implemented in Go.

----------------------------------------
PROJECT GOAL
----------------------------------------

Create an interactive terminal CLI that replays local Git history from a given starting commit to HEAD.

The user should be able to navigate forward and backward through commits using keyboard controls.

----------------------------------------
CLI SPEC
----------------------------------------

Usage:
    replay <start-commit>

Behavior:
1. Ensure we are inside a Git repository.
2. Validate the provided commit exists.
3. Ensure the commit is an ancestor of HEAD.
4. Collect all commits from <start-commit> to HEAD in chronological order.
5. Save the user's original branch or detached HEAD state.
6. Checkout the starting commit.
7. Enter interactive navigation mode.`

----------------------------------------
INTERACTIVE MODE
----------------------------------------

Controls:
    n → checkout next commit
    p → checkout previous commit
    q → quit and restore original state

Requirements:
- Prevent navigating out of bounds.
- After each move, print:
    - Short commit hash
    - Commit message (first line)
    - Position indicator (e.g., 3/15)
- Handle Ctrl+C gracefully and restore original state.
- Clean, minimal terminal UI (no heavy frameworks).

----------------------------------------
SAFETY REQUIREMENTS
----------------------------------------

- Detect dirty working tree and fail safely (warn and exit).
- Always restore original branch/state on exit.
- Handle detached HEAD correctly.
- Handle invalid commit input.
- Handle empty commit range.

----------------------------------------
ARCHITECTURE REQUIREMENTS
----------------------------------------

Clean modular structure:

/cmd/replay          → CLI entrypoint
/internal/git        → Git interaction layer
/internal/navigator  → Commit navigation state machine
/internal/ui         → Terminal interaction logic

Rules:
- No business logic inside main().
- Git commands must be wrapped in a dedicated interface (mockable).
- Navigator must be pure logic (unit-testable).
- UI must be separated from navigation state.

----------------------------------------
TESTING REQUIREMENTS
----------------------------------------

1. Unit Tests:
   - Commit range resolution
   - Ancestor validation
   - Navigator bounds behavior
   - State transitions

2. Integration Tests:
   - Use temporary Git repositories
   - Programmatically create commits
   - Validate checkout behavior

3. Tests must be written BEFORE implementation.
4. Use Go's standard testing package.
5. Keep dependencies minimal.

----------------------------------------
TECHNICAL CONSTRAINTS
----------------------------------------

- Go 1.26+
- No heavy dependencies
- Prefer standard library
- Cross-platform (Linux, macOS)
- OK to use:
    - cobra (CLI parsing)
    - termios or minimal TTY library if needed (use golang.org/x/term for TTY handling cross-platform)
- Avoid large TUI frameworks unless absolutely necessary
- Produce a single binary

----------------------------------------
UX REQUIREMENTS
----------------------------------------

When interactive mode starts, print:

Replay Mode
-----------
n → next
p → previous
q → quit

Show current commit information after every action.

Keep output clean and developer-focused.

