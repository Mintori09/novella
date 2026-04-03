package provider

import (
	"context"

	"novella/backend/models"
)

type GenerateConfig struct {
	Temperature  float64
	MaxTokens    int
	SystemPrompt string
	Model        string
	APIKey       string
}

type GenerateResult struct {
	Content    string
	TokenUsage models.TokenUsage
}

type Provider interface {
	Name() string
	GenerateContent(ctx context.Context, prompt string, cfg GenerateConfig) (GenerateResult, error)
	StreamContent(ctx context.Context, prompt string, cfg GenerateConfig) (<-chan string, <-chan error)
	TestConnection(ctx context.Context, apiKey string) error
}
