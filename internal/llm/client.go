package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Config holds LLM client configuration.
type Config struct {
	APIURL string
	APIKey string
	Model  string
}

// Source represents a piece of source material for man page generation.
type Source struct {
	Type    string // "help", "readme", "stdin"
	Content string
}

// GenerateRequest holds the parameters for a man page generation request.
type GenerateRequest struct {
	Sources []Source
	Name    string
	Section int
}

// Usage holds token usage statistics from the API response.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GenerateResult holds the output from a generation request.
type GenerateResult struct {
	Content string
	Usage   Usage
}

// Client is an OpenAI-compatible API client.
type Client struct {
	cfg  Config
	http *http.Client
}

// NewClient creates a new LLM client.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{},
	}
}

// chatRequest is the OpenAI chat completions request format.
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is the OpenAI chat completions response format.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage *Usage `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Generate produces an mdoc(7) man page from the given sources.
func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	systemPrompt := SystemPrompt()
	userPrompt := BuildUserPrompt(req)

	body := chatRequest{
		Model: c.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := strings.TrimRight(c.cfg.APIURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		// Show the raw response so the user can diagnose (e.g. wrong base URL)
		body := string(respBody)
		if len(body) > 200 {
			body = body[:200] + "..."
		}
		return nil, fmt.Errorf("unexpected API response: %s", body)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("API returned no choices")
	}

	content := chatResp.Choices[0].Message.Content
	content = stripCodeFences(content)

	result := &GenerateResult{Content: content}
	if chatResp.Usage != nil {
		result.Usage = *chatResp.Usage
	}

	return result, nil
}

// stripCodeFences removes ```mdoc or ``` wrapping from LLM output.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx >= 0 {
			s = s[idx+1:]
		}
		if strings.HasSuffix(s, "```") {
			s = strings.TrimSuffix(s, "```")
		}
		s = strings.TrimSpace(s)
	}
	return s
}
