package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"novella/backend/models"
)

type OpenAICompatibleProvider struct {
	name         string
	baseURL      string
	defaultModel string
}

func NewOpenAICompatibleProvider(name, baseURL, defaultModel string) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		name:         name,
		baseURL:      baseURL,
		defaultModel: defaultModel,
	}
}

func NewOpenAIProvider() *OpenAICompatibleProvider {
	return NewOpenAICompatibleProvider("openai", "https://api.openai.com/v1", "gpt-4o-mini")
}

func (p *OpenAICompatibleProvider) Name() string {
	return p.name
}

type openAIRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *OpenAICompatibleProvider) GenerateContent(ctx context.Context, prompt string, cfg GenerateConfig) (GenerateResult, error) {
	model := cfg.Model
	if model == "" {
		model = p.defaultModel
	}
	reqBody := openAIRequest{
		Model: model,
		Messages: []message{
			{Role: "system", Content: cfg.SystemPrompt},
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return GenerateResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return GenerateResult{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GenerateResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return GenerateResult{}, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var response openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return GenerateResult{}, err
	}

	if response.Error != nil {
		return GenerateResult{}, fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return GenerateResult{}, fmt.Errorf("no choices in response")
	}

	result := GenerateResult{
		Content: response.Choices[0].Message.Content,
	}

	if response.Usage != nil {
		result.TokenUsage = models.TokenUsage{
			InputTokens:  response.Usage.PromptTokens,
			OutputTokens: response.Usage.CompletionTokens,
			TotalTokens:  response.Usage.TotalTokens,
		}
	} else {
		result.TokenUsage = estimateTokens(len(cfg.SystemPrompt)+len(prompt), len(result.Content))
	}

	return result, nil
}

func (p *OpenAICompatibleProvider) StreamContent(ctx context.Context, prompt string, cfg GenerateConfig) (<-chan string, <-chan error) {
	ch := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		model := cfg.Model
		if model == "" {
			model = p.defaultModel
		}
		reqBody := openAIRequest{
			Model: model,
			Messages: []message{
				{Role: "system", Content: cfg.SystemPrompt},
				{Role: "user", Content: prompt},
			},
			Stream: true,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			errCh <- err
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			errCh <- err
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		req.Header.Set("Accept", "text/event-stream")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
			return
		}

		reader := io.Reader(resp.Body)
		buf := make([]byte, 4096)
		var leftover string

		for {
			n, readErr := reader.Read(buf)
			if n > 0 {
				leftover += string(buf[:n])
				for {
					idx := IndexOf(leftover, "\n\n")
					if idx == -1 {
						break
					}
					line := leftover[:idx]
					leftover = leftover[idx+2:]

					if len(line) > 6 && line[:6] == "data: " {
						data := line[6:]
						if data == "[DONE]" {
							return
						}
						var resp openAIResponse
						if err := json.Unmarshal([]byte(data), &resp); err == nil {
							if len(resp.Choices) > 0 {
								content := resp.Choices[0].Delta.Content
								if content != "" {
									select {
									case ch <- content:
									case <-ctx.Done():
										return
									}
								}
							}
						}
					}
				}
			}
			if readErr != nil {
				if readErr != io.EOF {
					errCh <- readErr
				}
				return
			}
		}
	}()

	return ch, errCh
}

func (p *OpenAICompatibleProvider) TestConnection(ctx context.Context, apiKey string, model string) error {
	reqBody := openAIRequest{
		Model: model,
		Messages: []message{
			{Role: "user", Content: "test"},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("connection failed: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}
