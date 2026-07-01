package crayfi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	SandboxURL = "https://dev-gateman.v3.connectramp.com"
	LiveURL    = "https://pay.connectramp.com"
)

// Client is the main Cray client
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	retries    int
	timeout    time.Duration

	Cards           *CardsService
	MoMo            *MoMoService
	Wallets         *WalletsService
	FX              *FXService
	Payouts         *PayoutsService
	Refunds         *RefundsService
	VirtualAccounts *VirtualAccountsService
	Checkout        *CheckoutService
	Crypto          *CryptoService
	CryptoPayouts   *CryptoPayoutsService
	Webhooks        *WebhooksService
}

// Option allows configuring the client
type Option func(*Client)

func WithEnv(env string) Option {
	return func(c *Client) {
		if env == "live" {
			c.baseURL = LiveURL
		} else {
			c.baseURL = SandboxURL
		}
	}
}

func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

func WithTimeout(seconds int) Option {
	return func(c *Client) {
		c.timeout = time.Duration(seconds) * time.Second
		c.httpClient.Timeout = c.timeout
	}
}

func WithRetries(retries int) Option {
	return func(c *Client) {
		c.retries = retries
	}
}

// New creates a new Cray client
func New(apiKey string, opts ...Option) (*Client, error) {
	// Try loading .env, ignore error if not found
	_ = godotenv.Load()

	// Default values
	key := apiKey
	if key == "" {
		key = os.Getenv("CRAY_API_KEY")
	}
	if key == "" {
		return nil, NewAuthenticationException("API Key is required.")
	}

	env := os.Getenv("CRAY_ENV")
	if env == "" {
		env = "sandbox"
	}

	timeoutStr := os.Getenv("CRAY_TIMEOUT")
	timeout := 30
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = t
		}
	}

	retriesStr := os.Getenv("CRAY_RETRIES")
	retries := 2
	if retriesStr != "" {
		if r, err := strconv.Atoi(retriesStr); err == nil {
			retries = r
		}
	}

	baseURL := os.Getenv("CRAY_BASE_URL")
	if baseURL == "" {
		if env == "live" {
			baseURL = LiveURL
		} else {
			baseURL = SandboxURL
		}
	}

	c := &Client{
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		apiKey:  key,
		baseURL: baseURL,
		retries: retries,
		timeout: time.Duration(timeout) * time.Second,
	}

	// Apply options (overriding env vars)
	for _, opt := range opts {
		opt(c)
	}

	c.Cards = &CardsService{client: c}
	c.MoMo = &MoMoService{client: c}
	c.Wallets = &WalletsService{client: c}
	c.FX = &FXService{client: c}
	c.Payouts = &PayoutsService{client: c}
	c.Refunds = &RefundsService{client: c}
	c.VirtualAccounts = &VirtualAccountsService{client: c}
	c.Checkout = &CheckoutService{client: c}
	c.Crypto = &CryptoService{client: c}
	c.CryptoPayouts = &CryptoPayoutsService{client: c}
	c.Webhooks = &WebhooksService{client: c}

	return c, nil
}

func (c *Client) request(method, path string, body interface{}) (interface{}, error) {
	url := fmt.Sprintf("%s%s", strings.TrimRight(c.baseURL, "/"), path)

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, NewValidationException("Invalid request body: " + err.Error())
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	var resp *http.Response
	var err error

	// Retry logic
	for i := 0; i <= c.retries; i++ {
		var req *http.Request
		req, err = http.NewRequest(method, url, reqBody)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err = c.httpClient.Do(req)

		// If success or non-retriable error (like 4xx), break
		// If network error or 5xx, continue if retries left
		if err == nil && resp.StatusCode < 500 {
			break
		}

		if i < c.retries {
			// Reset body reader for retry if needed
			if body != nil {
				jsonBytes, _ := json.Marshal(body)
				reqBody = bytes.NewBuffer(jsonBytes)
			}
			time.Sleep(1 * time.Second)
		}
	}

	if err != nil {
		return nil, NewAPIException("Request failed: "+err.Error(), 0, nil)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			// Fallback if not JSON
			result = map[string]string{"raw": string(respBody)}
		}
	}

	if resp.StatusCode >= 400 {
		msg := "Unknown API Error"
		if m, ok := result.(map[string]interface{}); ok {
			if v, exists := m["message"]; exists {
				msg = fmt.Sprintf("%v", v)
			}
		}
		return nil, NewAPIException(msg, resp.StatusCode, result)
	}

	return result, nil
}

func (c *Client) post(path string, data interface{}) (interface{}, error) {
	return c.request("POST", path, data)
}

func (c *Client) get(path string, params map[string]string) (interface{}, error) {
	fullPath := path
	if len(params) > 0 {
		fullPath += "?"
		qs := []string{}
		for k, v := range params {
			qs = append(qs, fmt.Sprintf("%s=%s", k, v))
		}
		fullPath += strings.Join(qs, "&")
	}
	return c.request("GET", fullPath, nil)
}
