package crayfi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	// Test default initialization (fails without API Key)
	os.Unsetenv("CRAY_API_KEY")
	_, err := New("")
	if err == nil {
		t.Error("Expected error when API Key is missing")
	}

	// Test with API Key
	client, err := New("test_key")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if client.baseURL != SandboxURL {
		t.Errorf("Expected default Sandbox URL, got %s", client.baseURL)
	}

	// Test with Env Option
	client, _ = New("test_key", WithEnv("live"))
	if client.baseURL != LiveURL {
		t.Errorf("Expected Live URL, got %s", client.baseURL)
	}

	// Test with BaseURL Option (should override Env)
	customURL := "https://custom.api.com"
	client, _ = New("test_key", WithEnv("live"), WithBaseURL(customURL))
	if client.baseURL != customURL {
		t.Errorf("Expected Custom URL, got %s", client.baseURL)
	}
}

func TestCardsInitiate(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/initiate" {
			t.Errorf("Expected path /api/v2/initiate, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test_key" {
			t.Errorf("Expected Authorization header, got %s", r.Header.Get("Authorization"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success", "id": "123"}`))
	}))
	defer server.Close()

	client, _ := New("test_key", WithBaseURL(server.URL))

	resp, err := client.Cards.Initiate(map[string]interface{}{"amount": 100})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	result := resp.(map[string]interface{})
	if result["status"] != "success" {
		t.Errorf("Expected status success, got %v", result["status"])
	}
}

func TestPayoutsBanks(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/payout/banks" {
			t.Errorf("Expected path /api/payout/banks, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected method GET, got %s", r.Method)
		}
		if r.URL.Query().Get("countryCode") != "NG" {
			t.Errorf("Expected countryCode=NG, got %s", r.URL.Query().Get("countryCode"))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []string{"Bank A", "Bank B"}})
	}))
	defer server.Close()

	client, _ := New("test_key", WithBaseURL(server.URL))

	resp, err := client.Payouts.Banks("NG")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	result := resp.(map[string]interface{})
	data := result["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("Expected 2 banks, got %d", len(data))
	}
}

func TestFXEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		invoke     func(*Client) (interface{}, error)
		statusCode int
	}{
		{
			name:   "rates",
			method: "POST",
			path:   "/api/rates",
			invoke: func(client *Client) (interface{}, error) {
				return client.FX.Rates(map[string]interface{}{
					"source_currency":      "USD",
					"destination_currency": "NGN",
				})
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "rates by destination",
			method: "POST",
			path:   "/api/rates/destination",
			invoke: func(client *Client) (interface{}, error) {
				return client.FX.RatesByDestination(map[string]interface{}{
					"destination_currency": "NGN",
				})
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "quote",
			method: "POST",
			path:   "/api/quote",
			invoke: func(client *Client) (interface{}, error) {
				return client.FX.Quote(map[string]interface{}{
					"source_currency":      "USD",
					"destination_currency": "NGN",
					"source_amount":        100,
				})
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "convert",
			method: "POST",
			path:   "/api/conversions",
			invoke: func(client *Client) (interface{}, error) {
				return client.FX.Convert(map[string]interface{}{"quote_id": "quote_123"})
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "conversions",
			method: "GET",
			path:   "/api/conversions",
			invoke: func(client *Client) (interface{}, error) {
				return client.FX.Conversions()
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "dispute conversion",
			method: "POST",
			path:   "/api/conversions/conv_123/dispute",
			invoke: func(client *Client) (interface{}, error) {
				return client.FX.DisputeConversion("conv_123", map[string]interface{}{
					"reason": "settlement_mismatch",
				})
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("Expected path %s, got %s", tt.path, r.URL.Path)
				}
				if r.Method != tt.method {
					t.Errorf("Expected method %s, got %s", tt.method, r.Method)
				}

				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
			}))
			defer server.Close()

			client, err := New("test_key", WithBaseURL(server.URL))
			if err != nil {
				t.Fatalf("Unexpected client error: %v", err)
			}

			if _, err := tt.invoke(client); err != nil {
				t.Fatalf("Unexpected request error: %v", err)
			}
		})
	}
}

func TestCheckoutEndpoints(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		invoke func(*Client) (interface{}, error)
	}{
		{
			name:   "initialize",
			method: "POST",
			path:   "/api/checkout/initialize",
			invoke: func(client *Client) (interface{}, error) {
				return client.Checkout.Initialize(map[string]interface{}{
					"reference": "checkout_123",
					"amount":    100,
				})
			},
		},
		{
			name:   "query",
			method: "GET",
			path:   "/api/checkout/query/checkout_123",
			invoke: func(client *Client) (interface{}, error) {
				return client.Checkout.Query("checkout_123")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("Expected path %s, got %s", tt.path, r.URL.Path)
				}
				if r.Method != tt.method {
					t.Errorf("Expected method %s, got %s", tt.method, r.Method)
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
			}))
			defer server.Close()

			client, err := New("test_key", WithBaseURL(server.URL))
			if err != nil {
				t.Fatalf("Unexpected client error: %v", err)
			}

			if _, err := tt.invoke(client); err != nil {
				t.Fatalf("Unexpected request error: %v", err)
			}
		})
	}
}

func TestCryptoCollectionEndpoints(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		invoke func(*Client) (interface{}, error)
	}{
		{
			name:   "supported assets",
			method: "GET",
			path:   "/api/virtual-accounts/crypto/supported-assets",
			invoke: func(client *Client) (interface{}, error) {
				return client.Crypto.SupportedAssets()
			},
		},
		{
			name:   "create vault",
			method: "POST",
			path:   "/api/accounts/crypto/vault",
			invoke: func(client *Client) (interface{}, error) {
				return client.Crypto.CreateVault(map[string]interface{}{
					"customer_reference": "customer_123",
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("Expected path %s, got %s", tt.path, r.URL.Path)
				}
				if r.Method != tt.method {
					t.Errorf("Expected method %s, got %s", tt.method, r.Method)
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
			}))
			defer server.Close()

			client, err := New("test_key", WithBaseURL(server.URL))
			if err != nil {
				t.Fatalf("Unexpected client error: %v", err)
			}

			if _, err := tt.invoke(client); err != nil {
				t.Fatalf("Unexpected request error: %v", err)
			}
		})
	}
}

func TestVirtualAccountOtpEndpoint(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Path != "/api/virtual-accounts/generate-wallet" {
			t.Errorf("Expected path /api/virtual-accounts/generate-wallet, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	client, err := New("test_key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("Unexpected client error: %v", err)
	}

	payload := map[string]interface{}{
		"otp":            "123456",
		"customer_email": "customer@example.com",
	}

	if _, err := client.VirtualAccounts.GenerateWallet(payload); err != nil {
		t.Fatalf("Unexpected GenerateWallet error: %v", err)
	}
	if _, err := client.VirtualAccounts.SubmitOtp(payload); err != nil {
		t.Fatalf("Unexpected SubmitOtp error: %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("Expected 2 requests, got %d", requestCount)
	}
}

func TestCryptoPayoutEndpoints(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		invoke func(*Client) (interface{}, error)
	}{
		{
			name:   "supported assets",
			method: "GET",
			path:   "/api/virtual-accounts/crypto/supported-assets",
			invoke: func(client *Client) (interface{}, error) {
				return client.CryptoPayouts.SupportedAssets()
			},
		},
		{
			name:   "add beneficiary",
			method: "POST",
			path:   "/api/payout/crypto/beneficiaries",
			invoke: func(client *Client) (interface{}, error) {
				return client.CryptoPayouts.AddBeneficiary(map[string]interface{}{
					"name":           "OMU",
					"asset":          "TRX_USDT_S2UZ",
					"wallet_address": "wallet_address",
				})
			},
		},
		{
			name:   "initiate payout",
			method: "POST",
			path:   "/api/payout/crypto/initiate-payout",
			invoke: func(client *Client) (interface{}, error) {
				return client.CryptoPayouts.InitiatePayout(map[string]interface{}{
					"amount":             "2",
					"currency":           "TRX_USDT_S2UZ",
					"address_reference":  "beneficiary_123",
					"customer_reference": "customer_ref_123",
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("Expected path %s, got %s", tt.path, r.URL.Path)
				}
				if r.Method != tt.method {
					t.Errorf("Expected method %s, got %s", tt.method, r.Method)
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
			}))
			defer server.Close()

			client, err := New("test_key", WithBaseURL(server.URL))
			if err != nil {
				t.Fatalf("Unexpected client error: %v", err)
			}

			if _, err := tt.invoke(client); err != nil {
				t.Fatalf("Unexpected request error: %v", err)
			}
		})
	}
}

func TestWebhookEndpoints(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		invoke func(*Client) (interface{}, error)
	}{
		{
			name:   "failed payout webhooks",
			method: "GET",
			path:   "/api/payout/failedWebhook",
			invoke: func(client *Client) (interface{}, error) {
				return client.Webhooks.FailedPayoutWebhooks()
			},
		},
		{
			name:   "retry failed payout webhook",
			method: "GET",
			path:   "/api/payout/failedWebhook/50",
			invoke: func(client *Client) (interface{}, error) {
				return client.Webhooks.RetryFailedPayoutWebhook("50")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("Expected path %s, got %s", tt.path, r.URL.Path)
				}
				if r.Method != tt.method {
					t.Errorf("Expected method %s, got %s", tt.method, r.Method)
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
			}))
			defer server.Close()

			client, err := New("test_key", WithBaseURL(server.URL))
			if err != nil {
				t.Fatalf("Unexpected client error: %v", err)
			}

			if _, err := tt.invoke(client); err != nil {
				t.Fatalf("Unexpected request error: %v", err)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Invalid Amount"}`))
	}))
	defer server.Close()

	client, _ := New("test_key", WithBaseURL(server.URL))

	_, err := client.Cards.Initiate(nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	apiErr, ok := err.(*APIException)
	if !ok {
		t.Fatalf("Expected APIException, got %T", err)
	}

	if apiErr.StatusCode != 400 {
		t.Errorf("Expected status 400, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Invalid Amount" {
		t.Errorf("Expected message 'Invalid Amount', got '%s'", apiErr.Message)
	}
}
