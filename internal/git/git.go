package gitops

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func GetDiff(worktree *git.Worktree, commit *object.Commit) (string, error) {
	var fullDiff strings.Builder

	// Get diff of tracked changes (both staged and unstaged)
	cmd := exec.Command("git", "diff", "HEAD")
	trackedDiff, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running git diff: %v", err)
	}
	fullDiff.Write(trackedDiff)

	// Get status to check for untracked files
	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("error getting worktree status: %v", err)
	}

	// Handle untracked files and folders
	for path, fileStatus := range status {
		if fileStatus.Staging == git.Untracked {
			content, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("error reading untracked file %s: %v", path, err)
			}
			untrackedDiff := fmt.Sprintf("\ndiff --git a/%s b/%s\nnew file mode 100644\nindex 0000000..%x\n--- /dev/null\n+++ b/%s\n@@ -0,0 +1,%d @@\n%s",
				path, path, plumbing.ComputeHash(plumbing.BlobObject, content), path, len(strings.Split(string(content), "\n")), string(content))
			fullDiff.WriteString(untrackedDiff)
		}
	}

	return fullDiff.String(), nil
}

func CommitChanges(worktree *git.Worktree, message string) error {
	// Get the status of the working tree
	status, err := worktree.Status()
	if err != nil {
		return err
	}

	// Add all changes to the staging area
	for filepath, fileStatus := range status {
		if fileStatus.Worktree != git.Unmodified {
			_, err := worktree.Add(filepath)
			if err != nil {
				return fmt.Errorf("error adding file %s to staging area: %v", filepath, err)
			}
		}
	}

	err = GitAdd(".")
	if err != nil {
		return err
	}

	err = GitCommit(message)
	if err != nil {
		return err
	}

	return nil
}

func GitAdd(path string) error {
	cmd := exec.Command("git", "add", path)
	return cmd.Run()
}

func GitCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %v\nOutput: %s", err, output)
	}
	return nil
}

// EditorAmendCommit opens the last commit message in the default Git editor for amending
func EditorAmendCommit() error {
	// Prepare Git command to amend the commit using the configured editor
	cmd := exec.Command("git", "commit", "--amend")

	// Set up the command to use the terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to amend commit: %v", err)
	}

	// Check if the commit was actually amended
	newMessage, err := GetLastCommitMessage()
	if err != nil {
		return fmt.Errorf("failed to get new commit message: %v", err)
	}

	if strings.TrimSpace(newMessage) == "" {
		return fmt.Errorf("commit amendment was aborted")
	}

	fmt.Println("Commit amended successfully.")
	return nil
}

// GetLastCommitMessage retrieves the message of the last commit
func GetLastCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get last commit message: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
