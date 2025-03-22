package scanner

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

// Load .env file automatically when package is imported
func init() {
	// Try to load .env file but don't fail if missing
	if err := godotenv.Load(); err != nil {
		fmt.Println("Note: No .env file found - using environment variables")
	}
}

// Analyze code with OpenAI
func AnalyzeWithOpenAI(code string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not found in .env file or environment variables")
	}

	client := openai.NewClient(apiKey)
  
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(
						"Analyze this code for security vulnerabilities. Reply with 'SAFE' or 'UNSAFE: <reason>'.\n\n%s", 
						code,
					),
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}