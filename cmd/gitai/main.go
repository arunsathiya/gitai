package main

import (
	"fmt"
	"os"

	"github.com/arunsathiya/gitai/internal/ai"
	"github.com/arunsathiya/gitai/internal/config"
	gitops "github.com/arunsathiya/gitai/internal/git"
	"github.com/arunsathiya/gitai/pkg/utils"
	"github.com/go-git/go-git/v5"
)

const maxAttempts = 5

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Open the repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening repository: %v\n", err)
		os.Exit(1)
	}

	// Get the working tree
	worktree, err := repo.Worktree()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting worktree: %v\n", err)
		os.Exit(1)
	}

	// Get the current HEAD commit
	head, err := repo.Head()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting HEAD: %v\n", err)
		os.Exit(1)
	}

	// Get the commit object for HEAD
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting commit object: %v\n", err)
		os.Exit(1)
	}

	// Get the diff
	diff, err := gitops.GetDiff(worktree, commit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting diff: %v\n", err)
		os.Exit(1)
	}

	// Print the diff
	if diff == "" {
		fmt.Println("No changes detected.")
		return
	}

	// Generate and confirm commit message
	var commitMessage string
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		commitMessage, err = ai.GenerateCommitMessage(diff, cfg.GroqAPIKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating commit message: %v\n", err)
			os.Exit(1)
		}

		confirmed := utils.ConfirmCommitMessage(commitMessage, attempt)
		if confirmed {
			break
		}

		if attempt == maxAttempts {
			fmt.Println("Maximum attempts reached. Exiting without committing.")
			return
		}
	}

	// Commit the changes
	err = gitops.CommitChanges(worktree, commitMessage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error committing changes: %v\n", err)
		os.Exit(1)
	}
}
