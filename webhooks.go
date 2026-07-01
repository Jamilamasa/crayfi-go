package crayfi

import "fmt"

type WebhooksService struct {
	client *Client
}

func (s *WebhooksService) FailedPayoutWebhooks() (interface{}, error) {
	return s.client.get("/api/payout/failedWebhook", nil)
}

func (s *WebhooksService) RetryFailedPayoutWebhook(webhookID string) (interface{}, error) {
	return s.client.get(fmt.Sprintf("/api/payout/failedWebhook/%s", webhookID), nil)
}
