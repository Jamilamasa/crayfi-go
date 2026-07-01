package crayfi

type CryptoPayoutsService struct {
	client *Client
}

func (s *CryptoPayoutsService) SupportedAssets() (interface{}, error) {
	return s.client.get("/api/virtual-accounts/crypto/supported-assets", nil)
}

func (s *CryptoPayoutsService) AddBeneficiary(data interface{}) (interface{}, error) {
	return s.client.post("/api/payout/crypto/beneficiaries", data)
}

func (s *CryptoPayoutsService) InitiatePayout(data interface{}) (interface{}, error) {
	return s.client.post("/api/payout/crypto/initiate-payout", data)
}
