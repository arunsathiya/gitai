package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const groqAPIURL = "https://gateway.ai.cloudflare.com/v1/75d17a47b6c80ac40b0e7e44a4a8517d/gitai/groq/openai/v1/chat/completions"

func GenerateCommitMessage(diff string, apiKey string) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY environment variable is not set")
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `You are a highly skilled developer tasked with generating precise and meaningful git commit messages. Follow these guidelines:

1. Use the Conventional Commits format: <type>(<scope>): <description>
2. Choose the most appropriate type (feat, update, fix, refactor, style, docs, test, chore)
3. Identify the specific scope of the changes
4. Write a concise but informative description of the changes, but limit to one line
5. Aim for clarity and specificity in your message
6. Analyze the entire diff to understand the full context of the changes
7. Focus on the most significant changes if there are multiple modifications
8. Avoid generic messages like "Update file" or "Fix bug"
9. Do not mention "using AI" or "automatic commit" in the message

Respond only with the commit message, without any additional text or explanations.`,
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
