package crayfi

type CryptoService struct {
	client *Client
}

func (s *CryptoService) SupportedAssets() (interface{}, error) {
	return s.client.get("/api/virtual-accounts/crypto/supported-assets", nil)
}

func (s *CryptoService) CreateVault(data interface{}) (interface{}, error) {
	return s.client.post("/api/accounts/crypto/vault", data)
}
