package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AnthropicClient uses Messages API.
type AnthropicClient struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func (c *AnthropicClient) Complete(ctx context.Context, system, user string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"

	body := map[string]any{
		"model":      c.Model,
		"max_tokens": 8192,
		"system":     system,
		"messages": []map[string]string{
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
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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
		return "", fmt.Errorf("anthropic http %d: %s", res.StatusCode, truncate(string(b), 500))
	}

	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	var out string
	for _, block := range parsed.Content {
		if block.Type == "text" {
			out += block.Text
		}
	}
	if out == "" {
		return "", fmt.Errorf("empty content in anthropic response")
	}
	return out, nil
}
