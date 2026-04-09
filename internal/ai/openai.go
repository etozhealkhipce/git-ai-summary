package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIClient uses Chat Completions API (OpenAI-compatible shape).
type OpenAIClient struct {
	BaseURL    string // e.g. https://api.openai.com/v1
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func (c *OpenAIClient) Complete(ctx context.Context, system, user string) (string, error) {
	base := strings.TrimSuffix(c.BaseURL, "/")
	url := base + "/chat/completions"

	body := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	b, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("openai http %d: %s", res.StatusCode, truncate(string(b), 500))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty choices in response")
	}
	return parsed.Choices[0].Message.Content, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// DefaultOpenAITimeoutClient returns client with timeout.
func DefaultHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}
