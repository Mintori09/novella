package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"novella/backend/models"
)

type CustomProvider struct {
	baseURL string
	method  string
}

func NewCustomProvider(baseURL, method string) *CustomProvider {
	if method == "" {
		method = "POST"
	}
	return &CustomProvider{
		baseURL: baseURL,
		method:  method,
	}
}

func (p *CustomProvider) Name() string {
	return "custom"
}

func (p *CustomProvider) buildURL(model string, prompt string, systemPrompt string) string {
	if p.method == "GET" {
		params := url.Values{}
		params.Set("model", model)
		params.Set("prompt", prompt)
		if systemPrompt != "" {
			params.Set("system", systemPrompt)
		}
		base := p.baseURL
		if len(params) > 0 {
			if base[len(base)-1:] == "/" {
				base = base[:len(base)-1]
			}
			base += "?" + params.Encode()
		}
		return base
	}
	return p.baseURL
}

func (p *CustomProvider) buildRequest(ctx context.Context, method string, url string, body []byte) (*http.Request, error) {
	if method == "GET" {
		return http.NewRequestWithContext(ctx, "GET", url, nil)
	}
	return http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
}

func (p *CustomProvider) GenerateContent(ctx context.Context, prompt string, cfg GenerateConfig) (GenerateResult, error) {
	model := cfg.Model
	if model == "" {
		model = "default"
	}

	var reqBody []byte
	var reqURL string
	var err error

	if p.method == "GET" {
		reqURL = p.buildURL(model, prompt, cfg.SystemPrompt)
		reqBody = nil
	} else {
		reqURL = p.baseURL
		body := map[string]any{
			"model": model,
			"messages": []map[string]string{
				{"role": "system", "content": cfg.SystemPrompt},
				{"role": "user", "content": prompt},
			},
		}
		reqBody, err = json.Marshal(body)
		if err != nil {
			return GenerateResult{}, err
		}
	}

	req, err := p.buildRequest(ctx, p.method, reqURL, reqBody)
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

func (p *CustomProvider) StreamContent(ctx context.Context, prompt string, cfg GenerateConfig) (<-chan string, <-chan error) {
	ch := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		model := cfg.Model
		if model == "" {
			model = "default"
		}

		var reqBody []byte
		var reqURL string
		var err error

		if p.method == "GET" {
			reqURL = p.buildURL(model, prompt, cfg.SystemPrompt)
			reqBody = nil
		} else {
			reqURL = p.baseURL
			body := map[string]any{
				"model": model,
				"messages": []map[string]string{
					{"role": "system", "content": cfg.SystemPrompt},
					{"role": "user", "content": prompt},
				},
				"stream": true,
			}
			reqBody, err = json.Marshal(body)
			if err != nil {
				errCh <- err
				return
			}
		}

		req, err := p.buildRequest(ctx, p.method, reqURL, reqBody)
		if err != nil {
			errCh <- err
			return
		}

		req.Header.Set("Content-Type", "application/json")
		if cfg.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		}
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

func (p *CustomProvider) TestConnection(ctx context.Context, apiKey string, model string) error {
	var reqBody []byte
	var reqURL string
	var err error

	if p.method == "GET" {
		reqURL = p.buildURL(model, "test", "")
		reqBody = nil
	} else {
		reqURL = p.baseURL
		body := map[string]any{
			"model": model,
			"messages": []map[string]string{
				{"role": "user", "content": "test"},
			},
		}
		reqBody, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	req, err := p.buildRequest(ctx, p.method, reqURL, reqBody)
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
