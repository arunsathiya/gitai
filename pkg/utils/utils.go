package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/eiannone/keyboard"
)

const maxAttempts = 5

func ConfirmCommitMessage(message string, attempt int) bool {
	fmt.Printf("\nGenerated commit message (Attempt %d/%d):\n%s\n", attempt, maxAttempts, message)
	fmt.Print("Do you want to use this commit message? [Y/n]: ")

	if err := keyboard.Open(); err != nil {
		fmt.Fprintln(os.Stderr, "Error opening keyboard:", err)
		return false
	}
	defer keyboard.Close()

	char, key, err := keyboard.GetSingleKey()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
		return false
	}

	fmt.Println() // Print newline for better formatting

	input := strings.ToLower(string(char))

	switch {
	case input == "y" || key == keyboard.KeyEnter:
		return true
	case input == "n":
		return false
	default:
		fmt.Println("Invalid input. Please enter 'y' or 'n'.")
		return ConfirmCommitMessage(message, attempt) // Recursively ask again
	}
}
