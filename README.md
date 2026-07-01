# Cray Go SDK

A Go library for integrating with the Cray Finance API.

## Installation

```bash
go get github.com/noibilism/crayfi-go
```

## Usage

### Initialization

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/noibilism/crayfi-go" // Imported as crayfi
)

func main() {
    // Initialize with API Key (defaults to Sandbox environment)
    client, err := crayfi.New("your-api-key")
    if err != nil {
        log.Fatal(err)
    }
    
    // OR with Options
    client, err = crayfi.New("your-api-key", 
        crayfi.WithEnv("live"),
        crayfi.WithTimeout(60),
    )
    
    // OR rely on Environment Variables
    // CRAY_API_KEY=your-api-key
    // CRAY_ENV=sandbox
    // CRAY_BASE_URL=https://custom-url.com
    client, err = crayfi.New("") 
}
```

### Modules

#### Cards

```go
// Initiate Transaction
// Prepare the transaction data including amount, currency, and customer details
data := map[string]interface{}{
    "amount": 1000,
    "email": "customer@example.com",
    "currency": "NGN",
    // Add other required fields like card_data if applicable for your use case
}
resp, err := client.Cards.Initiate(data)

// Charge
// Complete the transaction using the token or transaction ID received
resp, err = client.Cards.Charge(map[string]interface{}{"token": "..."})

// Query
// Check the status of a transaction using your reference
resp, err = client.Cards.Query("cust_ref_123")
```

#### MoMo (Mobile Money)

```go
// Initiate
// Start a mobile money payment request
resp, err := client.MoMo.Initiate(map[string]interface{}{
    "amount": 500,
    "phone_no": "2348012345678",
    "provider": "MTN",
})

// Requery
// Verify the status of a mobile money transaction
resp, err := client.MoMo.Requery("cust_ref_123")
```

#### Wallets

```go
// Get Balances
// Retrieve current balances for all currencies in your merchant wallet
resp, err := client.Wallets.Balances()

// Get Subaccounts
// List all subaccounts associated with your merchant account
resp, err := client.Wallets.Subaccounts()
```

#### FX (Foreign Exchange)

```go
// Get Rates
// Check current exchange rate between two currencies
resp, err := client.FX.Rates(map[string]interface{}{
    "source_currency": "USD",
    "destination_currency": "NGN",
})

// Convert
// Execute a currency conversion based on a quote or direct parameters
resp, err := client.FX.Convert(map[string]interface{}{
    "quote_id": "quote:98a5d6d3-7cbc-4c7d-b4f6-d3bbbbe340b6",
})
```

#### Payouts

```go
// Get Banks
// List supported banks for a specific country (e.g., "NG" for Nigeria)
resp, err := client.Payouts.Banks("NG")

// Disburse
// Send money to a bank account
resp, err := client.Payouts.Disburse(map[string]interface{}{
    "account_number": "1234567890",
    "bank_code": "058",
    "amount": 5000,
    "currency": "NGN",
    "narration": "Payment for services",
})
```

#### Crypto Payouts

```go
// Get supported crypto payout assets
resp, err := client.CryptoPayouts.SupportedAssets()

// Add a wallet beneficiary
resp, err = client.CryptoPayouts.AddBeneficiary(map[string]interface{}{
    "name": "OMU",
    "asset": "TRX_USDT_S2UZ",
    "wallet_address": "wallet_address",
})

// Initiate a crypto payout
resp, err = client.CryptoPayouts.InitiatePayout(map[string]interface{}{
    "amount": "2",
    "currency": "TRX_USDT_S2UZ",
    "address_reference": "beneficiary_reference",
    "customer_reference": "customer_ref_123",
    "narration": "Stablecoin payout",
})
```

#### Refunds

```go
// Initiate Refund
// Refund a transaction (partial or full)
resp, err := client.Refunds.Initiate(map[string]interface{}{
    "transaction_ref": "ref_123",
    "amount": 500, // Optional if full refund
})
```

#### Failed Payout Webhooks

```go
// List failed payout webhooks
resp, err := client.Webhooks.FailedPayoutWebhooks()

// Retry a failed payout webhook
resp, err = client.Webhooks.RetryFailedPayoutWebhook("50")
```

#### Virtual Accounts

```go
// Create a virtual account
resp, err := client.VirtualAccounts.Create(map[string]interface{}{
    "bvn":                  "22192474887",
    "type":                 "Corporate", // "Corporate" or "Individual"
    "nin":                  "11111122221",
    "virtual_account_type": "Permanent", // "Permanent" or "Onetime"
    "account_name":         "BOlaOla",
    "rc_number":            "99988828822", // Required for Corporate type
    "currency":             "NGN",
    "reference":            "cbf0d060-1544-4a53-a00b-7cb75a3eb59d",
    "customer_email":       "hello@gmail.com",
    "provider":             "monnify",
})

// Initiate a virtual account request (pre-create step)
resp, err = client.VirtualAccounts.Initiate(map[string]interface{}{
    "provider": "wema",
    "bvn":      "22192474887",
})

// List all virtual accounts for the merchant
resp, err = client.VirtualAccounts.List()

// Get available virtual account providers
resp, err = client.VirtualAccounts.Providers()

// Submit OTP to generate the Wema wallet
resp, err = client.VirtualAccounts.GenerateWallet(map[string]interface{}{
    "otp":            "768238",
    "customer_email": "hello@gmail.com",
})
```

## Error Handling

```go
_, err := client.Cards.Initiate(nil)
if err != nil {
    if apiErr, ok := err.(*crayfi.APIException); ok {
        fmt.Printf("API Error: %d - %s\n", apiErr.StatusCode, apiErr.Message)
    } else {
        fmt.Printf("Error: %v\n", err)
    }
}
```
