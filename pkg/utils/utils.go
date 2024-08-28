package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const maxAttempts = 5

func ConfirmCommitMessage(message string, attempt int) bool {
	fmt.Printf("\nGenerated commit message (Attempt %d/%d):\n%s\n", attempt, maxAttempts, message)
	fmt.Print("Do you want to use this commit message? (Y/n): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	return input == "" || input == "y" || input == "yes"
}
