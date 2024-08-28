package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	groqAPIURL = "https://api.groq.com/openai/v1/chat/completions"
)

func main() {
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
	diff, err := getDiff(worktree, commit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting diff: %v\n", err)
		os.Exit(1)
	}

	// Print the diff
	if diff == "" {
		fmt.Println("No changes detected.")
		return
	}

	fmt.Println("Changes detected:")
	fmt.Println(diff)

	// Generate commit message using Groq API
	commitMessage, err := generateCommitMessage(diff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating commit message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nGenerated commit message:\n%s\n", commitMessage)
}

func getDiff(worktree *git.Worktree, commit *object.Commit) (string, error) {
	var diff string

	// Get the changes between the working tree and the commit
	status, err := worktree.Status()
	if err != nil {
		return "", err
	}

	for path, fileStatus := range status {
		if fileStatus.Worktree != git.Unmodified {
			// Get the diff for this file
			fileDiff, err := getFileDiff(worktree, commit, path)
			if err != nil {
				return "", err
			}
			diff += fileDiff
		}
	}

	return diff, nil
}

func getFileDiff(worktree *git.Worktree, commit *object.Commit, path string) (string, error) {
	// Get the file from the commit
	file, err := commit.File(path)
	if err != nil {
		if err == plumbing.ErrObjectNotFound {
			// If the file doesn't exist in the commit, it's a new file
			content, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("diff --git a/%s b/%s\nnew file mode 100644\nindex 0000000..%x\n--- /dev/null\n+++ b/%s\n@@ -0,0 +1,%d @@\n%s",
				path, path, plumbing.ComputeHash(plumbing.BlobObject, content), path, len(content), string(content)), nil
		}
		return "", err
	}

	// Get the contents of the file in the commit
	commitContent, err := file.Contents()
	if err != nil {
		return "", err
	}

	// Get the contents of the file in the working tree
	worktreeContent, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// If the contents are the same, return an empty string
	if string(worktreeContent) == commitContent {
		return "", nil
	}

	// Create a unified diff
	return fmt.Sprintf("diff --git a/%s b/%s\nindex %x..%x 100644\n--- a/%s\n+++ b/%s\n@@ -1,%d +1,%d @@\n%s",
		path, path,
		plumbing.ComputeHash(plumbing.BlobObject, []byte(commitContent)),
		plumbing.ComputeHash(plumbing.BlobObject, worktreeContent),
		path, path,
		len(commitContent), len(worktreeContent),
		generateUnifiedDiff(commitContent, string(worktreeContent))), nil
}

func generateUnifiedDiff(oldContent, newContent string) string {
	var diff string
	oldLines := splitLines(oldContent)
	newLines := splitLines(newContent)

	for i := 0; i < len(oldLines) || i < len(newLines); i++ {
		if i >= len(oldLines) {
			diff += fmt.Sprintf("+%s\n", newLines[i])
		} else if i >= len(newLines) {
			diff += fmt.Sprintf("-%s\n", oldLines[i])
		} else if oldLines[i] != newLines[i] {
			diff += fmt.Sprintf("-%s\n", oldLines[i])
			diff += fmt.Sprintf("+%s\n", newLines[i])
		} else {
			diff += fmt.Sprintf(" %s\n", oldLines[i])
		}
	}

	return diff
}

func splitLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

func generateCommitMessage(diff string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY environment variable is not set")
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Generate a concise git commit message for the following diff. Use the Conventional Commits format: <type>(<scope>): <description>. Respond only with the commit message, without any additional text or explanations.",
			},
			{
				"role":    "user",
				"content": diff,
			},
		},
		"model":       "llama-3.1-70b-versatile",
		"temperature": 1,
		"max_tokens":  8000,
		"top_p":       1,
		"stream":      false,
		"stop":        nil,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", groqAPIURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("unexpected API response format")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected API response format")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected API response format")
	}

	return strings.TrimSpace(content), nil
}
