package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// LiteLLMClient implements AIClient for litellm proxies (OpenAI-compatible API).
type LiteLLMClient struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// NewLiteLLMClient creates a new litellm client.
func NewLiteLLMClient(baseURL, apiKey, model string) *LiteLLMClient {
	return &LiteLLMClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{Timeout: 120 * 1e9}, // 120 seconds
	}
}

// Complete sends a completion request to the litellm proxy.
func (c *LiteLLMClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	url := c.baseURL + "/v1/chat/completions"
	if req.Debug {
		fmt.Fprintf(os.Stderr, "[debug] litellm request url=%s model=%s max_tokens=%d temperature=%.2f messages=%d\n",
			url, req.Model, req.MaxTokens, req.Temperature, len(req.Messages))
	}

	// Convert messages to OpenAI format
	messages := make([]map[string]string, 0, len(req.Messages)+1)
	if req.System != "" {
		messages = append(messages, map[string]string{"role": "system", "content": req.System})
	}
	for _, m := range req.Messages {
		messages = append(messages, map[string]string{"role": m.Role, "content": m.Content})
	}

	payload := map[string]interface{}{
		"model":       req.Model,
		"messages":    messages,
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("litellm request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if req.Debug {
			fmt.Fprintf(os.Stderr, "[debug] litellm response status=%s body=%s\n", resp.Status, strings.TrimSpace(string(respBody)))
		}
		return nil, fmt.Errorf("litellm API error (%s): %s", resp.Status, string(respBody))
	}
	if req.Debug {
		fmt.Fprintf(os.Stderr, "[debug] litellm response status=%s\n", resp.Status)
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens     int `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse litellm response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from litellm")
	}

	return &CompletionResponse{
		Content:      openAIResp.Choices[0].Message.Content,
		Model:        openAIResp.Model,
		TokensIn:     openAIResp.Usage.PromptTokens,
		TokensOut:    openAIResp.Usage.CompletionTokens,
		FinishReason: openAIResp.Choices[0].FinishReason,
	}, nil
}

// Name returns the provider name.
func (c *LiteLLMClient) Name() string {
	return "litellm"
}
