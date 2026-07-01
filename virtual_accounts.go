package crayfi

type VirtualAccountsService struct {
	client *Client
}

func (s *VirtualAccountsService) Create(data interface{}) (interface{}, error) {
	return s.client.post("/api/virtual-accounts/create", data)
}

func (s *VirtualAccountsService) Initiate(data interface{}) (interface{}, error) {
	return s.client.post("/api/virtual-accounts/initiate", data)
}

func (s *VirtualAccountsService) List() (interface{}, error) {
	return s.client.get("/api/virtual-accounts/list", nil)
}

func (s *VirtualAccountsService) Providers() (interface{}, error) {
	return s.client.get("/api/virtual-accounts/providers", nil)
}

func (s *VirtualAccountsService) GenerateWallet(data interface{}) (interface{}, error) {
	return s.client.post("/api/virtual-accounts/generate-wallet", data)
}

func (s *VirtualAccountsService) SubmitOtp(data interface{}) (interface{}, error) {
	return s.GenerateWallet(data)
}
