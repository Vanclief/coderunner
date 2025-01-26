package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	TokensPerMinute        = 80000
	EstimatedTokensPerChar = 0.5
	RefillInterval         = time.Minute
)

type API struct {
	APIKey    string
	Model     string
	BaseURL   string
	MaxTokens int
	client    *http.Client

	remainingTokens float64
	nextRefill      time.Time
	mutex           sync.Mutex
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type request struct {
	Model     string    `json:"model"`
	Messages  []message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type response struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []contentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type APIError struct {
	StatusCode int
	Status     string
	Body       string
	Headers    http.Header
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Claude API Error:\nStatus: %s (%d)\nHeaders: %v\nResponse Body: %s",
		e.Status,
		e.StatusCode,
		e.Headers,
		e.Body,
	)
}

func NewAPI(apiKey, model string, maxTokens int) *API {
	return &API{
		APIKey:          apiKey,
		Model:           model,
		BaseURL:         "https://api.anthropic.com/v1/messages",
		MaxTokens:       maxTokens,
		client:          &http.Client{Timeout: 30 * time.Second},
		remainingTokens: float64(TokensPerMinute),
		nextRefill:      time.Now().Add(RefillInterval),
	}
}

func (a *API) waitForTokens(requiredTokens float64) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for {
		now := time.Now()

		if now.After(a.nextRefill) {
			a.remainingTokens = float64(TokensPerMinute)
			a.nextRefill = now.Add(RefillInterval)
		}

		if a.remainingTokens >= requiredTokens {
			a.remainingTokens -= requiredTokens
			return
		}

		waitTime := a.nextRefill.Sub(now)
		if waitTime <= 0 {
			continue
		}

		a.mutex.Unlock()
		time.Sleep(waitTime)
		a.mutex.Lock()
	}
}

func (a *API) estimateTokens(prompt string) float64 {
	return float64(len(prompt)) * EstimatedTokensPerChar
}

func (a *API) Prompt(prompt string) (string, error) {
	requiredTokens := a.estimateTokens(prompt)
	a.waitForTokens(requiredTokens)

	req := request{
		Model: a.Model,
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: a.MaxTokens,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("x-api-key", a.APIKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			// Force remaining tokens to zero and retry once
			a.mutex.Lock()
			a.remainingTokens = 0
			a.mutex.Unlock()
			return a.Prompt(prompt)
		}

		return "", &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(bodyBytes),
			Headers:    resp.Header,
		}
	}

	var apiResponse response
	if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		return "", fmt.Errorf("error decoding response: %w\nResponse body: %s", err, string(bodyBytes))
	}

	if len(apiResponse.Content) > 0 {
		return apiResponse.Content[0].Text, nil
	}

	return "", fmt.Errorf("no content in response")
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
