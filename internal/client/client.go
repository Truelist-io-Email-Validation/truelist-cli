package client

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
	DefaultBaseURL = "https://api.truelist.io"
	rateLimit      = 10 // requests per second
)

// ValidationResult holds the response from the Truelist API.
type ValidationResult struct {
	Email      string `json:"email"`
	State      string `json:"state"`
	SubState   string `json:"sub_state"`
	FreeEmail  bool   `json:"free_email"`
	Role       bool   `json:"role"`
	Disposable bool   `json:"disposable"`
	Suggestion string `json:"suggestion,omitempty"`
}

// AccountInfo holds the response from the whoami/account endpoint.
type AccountInfo struct {
	Email   string `json:"email"`
	Plan    string `json:"plan"`
	Credits int    `json:"credits"`
}

// Client is the Truelist API client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client

	// Rate limiter fields.
	mu        sync.Mutex
	tokens    int
	lastReset time.Time
}

// New creates a new API client.
func New(apiKey string) *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tokens:    rateLimit,
		lastReset: time.Now(),
	}
}

// WithBaseURL overrides the default base URL (useful for testing).
func (c *Client) WithBaseURL(url string) *Client {
	c.baseURL = url
	return c
}

// waitForToken blocks until a rate limit token is available.
func (c *Client) waitForToken() {
	for {
		c.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(c.lastReset)

		if elapsed >= time.Second {
			c.tokens = rateLimit
			c.lastReset = now
		}

		if c.tokens > 0 {
			c.tokens--
			c.mu.Unlock()
			return
		}
		c.mu.Unlock()

		sleepDuration := time.Second - elapsed
		if sleepDuration < 10*time.Millisecond {
			sleepDuration = 10 * time.Millisecond
		}
		time.Sleep(sleepDuration)
	}
}

// doRequest performs an authenticated HTTP request.
func (c *Client) doRequest(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "truelist-cli")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// Validate verifies a single email address.
func (c *Client) Validate(ctx context.Context, email string) (*ValidationResult, error) {
	c.waitForToken()

	payload := map[string]string{"email": email}
	body, status, err := c.doRequest(ctx, http.MethodPost, "/api/v1/verify", payload)
	if err != nil {
		return nil, err
	}

	if status == 401 {
		return nil, fmt.Errorf("unauthorized — check your API key")
	}
	if status == 429 {
		return nil, fmt.Errorf("rate limited — too many requests")
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", status, string(body))
	}

	var result ValidationResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Email == "" {
		result.Email = email
	}

	return &result, nil
}

// Whoami checks the API key and returns account info.
func (c *Client) Whoami(ctx context.Context) (*AccountInfo, error) {
	c.waitForToken()

	body, status, err := c.doRequest(ctx, http.MethodGet, "/api/v1/account", nil)
	if err != nil {
		return nil, err
	}

	if status == 401 {
		return nil, fmt.Errorf("unauthorized — check your API key")
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", status, string(body))
	}

	var info AccountInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &info, nil
}
