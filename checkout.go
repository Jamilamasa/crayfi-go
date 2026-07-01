package crayfi

import "fmt"

type CheckoutService struct {
	client *Client
}

func (s *CheckoutService) Initialize(data interface{}) (interface{}, error) {
	return s.client.post("/api/checkout/initialize", data)
}

func (s *CheckoutService) Query(reference string) (interface{}, error) {
	return s.client.get(fmt.Sprintf("/api/checkout/query/%s", reference), nil)
}
